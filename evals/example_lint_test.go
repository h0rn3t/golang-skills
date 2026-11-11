package evals_test

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
)

// The example tests compile the blocks a skill shows; this file lints them
// with the config the gate uses where a repository has none. A model copies
// the example, the edit hook lints the copy with that config, and
// go-style-core says a finding in touched code is fixed: an endorsed form
// that draws a finding sends every session that follows it into a loop.

// extractedBlock is one Go block exampleBlock returned to a test.
type extractedBlock struct {
	source string // "skills/x/references/Y.md § heading"
	lines  map[string]bool
}

var (
	exampleBlocksMu sync.Mutex
	exampleBlocks   = map[string][]extractedBlock{} // top-level test name -> blocks
)

func rememberExampleBlock(t *testing.T, path, heading, code string) {
	t.Helper()
	block := extractedBlock{source: path + " § " + strings.TrimLeft(heading, "# "), lines: map[string]bool{}}
	for line := range strings.Lines(code) {
		if line = strings.TrimSpace(line); line != "" {
			block.lines[line] = true
		}
	}
	name, _, _ := strings.Cut(t.Name(), "/")
	exampleBlocksMu.Lock()
	defer exampleBlocksMu.Unlock()
	exampleBlocks[name] = append(exampleBlocks[name], block)
}

// wrapperLinters report what the test's wrapper causes, not the example: a
// helper the wrapper never calls, an exported name whose comment the block
// elides, and the _test.go file the block is compiled in.
var wrapperLinters = []string{"unused", "usetesting", "godot", "revive", "unparam"}

// exampleLintFindings pins the findings inside example blocks that are meant
// to be there: the block shows the wrong form, or the finding is the lesson.
// The key is "<source> | <linter> | <source line>". An endorsed form with a
// finding is fixed in the skill, not listed here.
var exampleLintFindings = map[string]bool{
	"skills/go-logging/SKILL.md § Structured Logging | sloglint | slog.Info(fmt.Sprintf(\"order %d placed for $%.2f\", orderID, total))": true,
}

// The defended forms of go-security draw gosec's taint findings, and the skill
// shows the suppression for each. Linted in a handler, a suppression that
// stopped working is a gosec finding and one that is no longer needed is a
// nolintlint finding, both on an example line.
func TestSecurityExampleTaintSuppressions(t *testing.T) {
	client := exampleBlock(t, "skills/go-security/references/INJECTION.md", "### Arbitrary public destinations")
	redirect := exampleBlock(t, "skills/go-security/references/INJECTION.md", "### Local redirects")
	lines := strings.Split(strings.TrimSpace(exampleBlock(t, "skills/go-security/SKILL.md", "### A taint finding on a defended line")), "\n")
	if len(lines) != 4 {
		t.Fatalf("the suppression fragment has %d lines, want the request, the Do, the redirect, and the multipart parse", len(lines))
	}
	runExampleTest(t, `package example
import ("fmt"; "net"; "net/http"; "net/http/httptest"; "net/netip"; "slices"; "strings"; "syscall"; "testing"; "time")
`+client+redirect+`
func preview(w http.ResponseWriter, r *http.Request) {
	ctx, target, client := r.Context(), r.URL.Query().Get("url"), publicClient()
	`+lines[0]+`
	if err != nil {
		http.Error(w, "bad url", http.StatusBadRequest)
		return
	}
	`+lines[1]+`
	if err != nil {
		http.Error(w, "fetch failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.WriteHeader(http.StatusNoContent)
}
func login(w http.ResponseWriter, r *http.Request) {
	next := r.URL.Query().Get("next")
	if !localRedirect(next) {
		next = "/"
	}
	`+lines[2]+`
}
func upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	`+lines[3]+`
	if err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func TestHandlers(t *testing.T) {
	// Through a mux, as a service registers them: gosec reports a handler's
	// request only when nothing in the package calls the handler directly.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /login", login)
	mux.HandleFunc("GET /preview", preview)
	mux.HandleFunc("POST /upload", upload)
	serve := func(method, target, body string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(method, target, strings.NewReader(body)))
		return rec
	}
	if got := serve(http.MethodGet, "/login?next=//evil.example", "").Header().Get("Location"); got != "/" {
		t.Errorf("login redirect = %q, want /", got)
	}
	if got := serve(http.MethodGet, "/preview?url=http://127.0.0.1:1/", "").Code; got != http.StatusBadGateway {
		t.Errorf("preview of a loopback URL = %d, want %d", got, http.StatusBadGateway)
	}
	if got := serve(http.MethodPost, "/upload", "not multipart").Code; got != http.StatusBadRequest {
		t.Errorf("upload of a non-multipart body = %d, want %d", got, http.StatusBadRequest)
	}
}
`)
}

// The benchmark pair measures the form perfsprint reports, so it carries the
// suppression; compiled here, it stays under the lint pass.
func TestPerformanceExampleBenchmarkPair(t *testing.T) {
	code := exampleBlock(t, "skills/go-performance/references/BENCHMARKS.md", "## Writing Benchmarks")
	runExampleTest(t, "package example\nimport (\"fmt\"; \"strconv\"; \"testing\")\n"+code)
}

// startExampleLint lints the module in dir with the bundled config while the
// caller runs its tests, and returns the function that waits for the result
// and compares the findings on example lines with exampleLintFindings.
func startExampleLint(t *testing.T, dir string) func() {
	t.Helper()
	golangciLint, err := exec.LookPath("golangci-lint")
	if err != nil {
		t.Fatal("golangci-lint not installed; the examples are checked against the bundled config")
	}
	name, _, _ := strings.Cut(t.Name(), "/")
	exampleBlocksMu.Lock()
	blocks := slices.Clone(exampleBlocks[name])
	exampleBlocksMu.Unlock()
	if len(blocks) == 0 {
		return func() {} // no skill text in this module: nothing to attribute
	}
	report := filepath.Join(t.TempDir(), "lint.json")
	config := filepath.Join(repoRoot(t), "skills", "go-linting", "assets", "golangci.yml")
	cmd := exec.CommandContext(t.Context(), golangciLint, "run", "--config", config,
		"--allow-parallel-runners", "--output.json.path", report, "./...")
	cmd.Dir = dir
	var output strings.Builder
	cmd.Stdout, cmd.Stderr = &output, &output
	if err := cmd.Start(); err != nil {
		t.Fatalf("start golangci-lint: %v", err)
	}
	return func() {
		t.Helper()
		err := cmd.Wait()
		if t.Failed() {
			return // the example does not build or pass: lint has nothing to add
		}
		if exitErr, ok := errors.AsType[*exec.ExitError](err); err != nil && (!ok || exitErr.ExitCode() != 1) {
			t.Errorf("golangci-lint on the example: %v, want exit 0 or 1\n%s", err, output.String())
			return
		}
		data, err := os.ReadFile(report)
		if err != nil {
			t.Errorf("read golangci-lint JSON: %v", err)
			return
		}
		var result struct {
			Issues []struct {
				FromLinter string
				Text       string
				Pos        struct {
					Filename string
					Line     int
				}
			}
		}
		if err := json.Unmarshal(data, &result); err != nil {
			t.Errorf("decode golangci-lint JSON: %v\n%s", err, data)
			return
		}
		files := map[string][]string{}
		seen := map[string]bool{}
		for _, issue := range result.Issues {
			if slices.Contains(wrapperLinters, issue.FromLinter) {
				continue
			}
			lines, ok := files[issue.Pos.Filename]
			if !ok {
				src, err := os.ReadFile(filepath.Join(dir, issue.Pos.Filename))
				if err != nil {
					t.Errorf("read %s: %v", issue.Pos.Filename, err)
					continue
				}
				lines = strings.Split(string(src), "\n")
				files[issue.Pos.Filename] = lines
			}
			if issue.Pos.Line < 1 || issue.Pos.Line > len(lines) {
				continue
			}
			line := strings.TrimSpace(lines[issue.Pos.Line-1])
			for _, block := range blocks {
				if !block.lines[line] {
					continue // the test's own wrapper
				}
				key := block.source + " | " + issue.FromLinter + " | " + line
				seen[key] = true
				if !exampleLintFindings[key] {
					t.Errorf("the bundled golangci.yml reports an example line:\n  %s\n  %s: %s\nfix the example, or pin the key in exampleLintFindings when the block shows the wrong form", key, issue.FromLinter, issue.Text)
				}
			}
		}
		for key := range exampleLintFindings {
			source, _, _ := strings.Cut(key, " | ")
			if !seen[key] && slices.ContainsFunc(blocks, func(b extractedBlock) bool { return b.source == source }) {
				t.Errorf("exampleLintFindings pins a finding the bundled golangci.yml no longer reports: %s", key)
			}
		}
	}
}
