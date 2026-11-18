package evals_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestLintInstallPinMatchesCI keeps every golangci-lint install command in
// skills/ and docs/ on the version CI installs. The scripts print one when the
// linter is missing; a drifted pin installs a linter the bundled config was
// not verified against.
func TestLintInstallPinMatchesCI(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	workflow, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "validate-skills.yml"))
	if err != nil {
		t.Fatal(err)
	}
	// The same pattern go-release-watch.yml reads the pin with.
	m := regexp.MustCompile(`(?m)^\s*GOLANGCI_LINT_VERSION:\s*(v[0-9.]+)\s*$`).FindSubmatch(workflow)
	if m == nil {
		t.Fatal("GOLANGCI_LINT_VERSION not found in validate-skills.yml")
	}
	pinned := string(m[1])
	install := regexp.MustCompile(`golangci-lint/v2/cmd/golangci-lint@(v\d+\.\d+\.\d+)`)
	repo := os.DirFS(root)
	for _, dir := range []string{"skills", "docs"} {
		err := fs.WalkDir(repo, dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			data, err := fs.ReadFile(repo, path)
			if err != nil {
				return err
			}
			for _, got := range install.FindAllSubmatch(data, -1) {
				if string(got[1]) != pinned {
					t.Errorf("%s installs golangci-lint %s, CI pins %s", path, got[1], pinned)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, script := range []string{
		"go-code-refactor/scripts/verify-refactor.sh",
		"go-code-review/scripts/pre-review.sh",
		"go-linting/scripts/setup-lint.sh",
	} {
		data, err := fs.ReadFile(repo, "skills/"+script)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(data, []byte("golangci-lint/v2/cmd/golangci-lint@"+pinned)) {
			t.Errorf("%s does not print the pinned install command for a missing golangci-lint", script)
		}
	}
}

// TestRepoLintConfigTracksBundled keeps the repository's own .golangci.yml on
// the bundled linter set: golangci-lint has no config inheritance, so the
// root file copies the bundled linters block verbatim and adds only its
// exclusions. A linter or setting changed in one copy alone fails here.
func TestRepoLintConfigTracksBundled(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	bundled, err := os.ReadFile(filepath.Join(root, "skills", "go-linting", "assets", "golangci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	repo, err := os.ReadFile(filepath.Join(root, ".golangci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	_, linters, ok := bytes.Cut(bundled, []byte("\nlinters:\n"))
	if !ok {
		t.Fatal("bundled golangci.yml has no top-level linters block")
	}
	linters, _, ok = bytes.Cut(linters, []byte("\n  exclusions:\n"))
	if !ok {
		t.Fatal("bundled golangci.yml has no linters.exclusions block")
	}
	if !bytes.Contains(repo, append([]byte("\nlinters:\n"), linters...)) {
		t.Error(".golangci.yml linters block differs from skills/go-linting/assets/golangci.yml; copy it from \"linters:\" to its exclusions")
	}
}

// TestBundledLintConfig runs the bundled golangci.yml over fixtures/lint, one
// scenario per file. The edit hook lints with this config when a repository
// has none, and go-style-core says a finding it prints is fixed, so a finding
// on an idiom the skills teach becomes code the skills call slop.
func TestBundledLintConfig(t *testing.T) {
	golangciLint, err := exec.LookPath("golangci-lint")
	if err != nil {
		t.Fatal("golangci-lint not installed; the bundled config is checked against it")
	}
	root := repoRoot(t)
	config := filepath.Join(root, "skills", "go-linting", "assets", "golangci.yml")
	jsonPath := filepath.Join(t.TempDir(), "report.json")
	cmd := exec.CommandContext(t.Context(), golangciLint, "run", "--config", config, "--output.json.path", jsonPath, "./...")
	cmd.Dir = filepath.Join(root, "evals", "fixtures", "lint")
	if out, err := cmd.CombinedOutput(); err != nil {
		if exitErr, ok := errors.AsType[*exec.ExitError](err); !ok || exitErr.ExitCode() != 1 {
			t.Fatalf("golangci-lint run: %v, want exit 0 or 1\n%s", err, out)
		}
	}
	out, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read golangci-lint JSON: %v", err)
	}
	var report struct {
		Issues []struct {
			FromLinter string
			Text       string
			Pos        struct{ Filename string }
		}
	}
	if err := json.Unmarshal(out, &report); err != nil {
		t.Fatalf("decode golangci-lint JSON: %v\n%s", err, out)
	}

	tests := []struct {
		file, linter, text string
		want               bool
	}{
		{"internal/probe/readclose.go", "errcheck", "rows.Close", false},
		{"internal/probe/readclose.go", "errcheck", "tx.Rollback", false},
		{"internal/probe/readclose.go", "errcheck", "resp.Body.Close", false},
		{"internal/probe/write.go", "errcheck", "f.Close", true},
		{"internal/probe/decode.go", "errcheck", "json.Unmarshal", true},
		{"internal/probe/collect.go", "prealloc", "", false},
		{"internal/probe/collect.go", "perfsprint", "", false},
		{"internal/probe/flat.go", "gocyclo", "", false},
		{"internal/probe/readfile.go", "gosec", "G304", false},
		{"internal/probe/write.go", "gosec", "G304", false},
		{"internal/probe/eof.go", "errorlint", "", true},
		{"internal/probe/naming.go", "revive", "userId", true},
		{"internal/probe/slogmsg.go", "sloglint", "string literal or a constant", true},
		{"internal/probe/slogattrs.go", "sloglint", "", false},
		{"internal/probe/testctx_test.go", "usetesting", "context.Background()", true},
		{"internal/probe/testctx_test.go", "usetesting", "context.TODO()", true},
		{"internal/probe/cleanupctx_test.go", "usetesting", "", false},
		{"internal/probe/receivers.go", "recvcheck", `"Counter"`, true},
		{"internal/probe/receivers.go", "recvcheck", `"Celsius"`, false},
		{"internal/probe/decoders.go", "recvcheck", `"Money"`, false},
		{"internal/probe/decoders.go", "recvcheck", `"Flag"`, false},
		{"internal/probe/ctxfield.go", "containedctx", "", true},
		{"internal/probe/builtin.go", "predeclared", "max", true},
		{"internal/store/store.go", "revive", "should have comment", false},
		{"cmd/tool/main.go", "revive", "should have comment", false},
		{"api.go", "revive", "should have comment", true},
	}
	for _, tt := range tests {
		got := false
		for _, issue := range report.Issues {
			if issue.FromLinter == tt.linter && pathHasSuffix(issue.Pos.Filename, tt.file) && strings.Contains(issue.Text, tt.text) {
				got = true
			}
		}
		if got != tt.want {
			t.Errorf("golangci-lint %s: %s finding %q = %t, want %t", tt.file, tt.linter, tt.text, got, tt.want)
		}
	}
}
