package evals_test

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func skillScript(t *testing.T, skill, name string) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "skills", skill, "scripts", name)
}

func writeTree(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		writeFile(t, filepath.Join(dir, name), content)
	}
}

// A file target is read with its whole package: the package comment in doc.go
// counts, and an implementation in another file makes the interface the same
// finding the directory target reports.
func TestScriptFileTargetReadsWholePackage(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"go.mod":        "module filetarget\n\ngo 1.27\n",
		"pay/doc.go":    "// Package pay charges cards.\npackage pay\n",
		"pay/charge.go": "package pay\n\n// Store keeps charges.\ntype Store interface{ Get() int }\n\nfunc Refund() {}\n",
		"pay/mem.go":    "package pay\n\ntype memStore struct{}\n\nfunc (memStore) Get() int { return 1 }\n",
	})
	names := func(t *testing.T, out []byte) []string {
		t.Helper()
		var result struct {
			Missing []jsonFinding `json:"missing"`
		}
		if err := json.Unmarshal(out, &result); err != nil {
			t.Fatalf("parse JSON: %v\n%s", err, out)
		}
		var got []string
		for _, m := range result.Missing {
			got = append(got, m.Name)
		}
		return got
	}

	out := runCommandInDir(t, dir, 1, "bash", skillScript(t, "go-documentation", "check-docs.sh"), "--json", "pay/charge.go")
	if got := names(t, out); !slices.Equal(got, []string{"Refund"}) {
		t.Errorf("check-docs.sh pay/charge.go missing = %q, want only Refund: the package comment is in pay/doc.go", got)
	}
	for _, target := range []string{"./pay", "pay/charge.go"} {
		out := runCommandInDir(t, dir, 1, "bash", skillScript(t, "go-interfaces", "check-interface-compliance.sh"), "--json", target)
		if got := names(t, out); !slices.Equal(got, []string{"Store"}) {
			t.Errorf("check-interface-compliance.sh %s missing = %q, want Store: pay/mem.go implements it", target, got)
		}
	}
}

// check-docs marks the findings the lint gate excludes, and skips Len, Less,
// and Swap only on a type that declares all three, as revive does.
func TestCheckDocsAdvisoryAndSortMethods(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"go.mod":                  "module docsadvisory\n\ngo 1.27\n",
		"internal/store/store.go": "// Package store keeps rows.\npackage store\n\nfunc Open() {}\n",
		"api/api.go": `// Package api is the public surface.
package api

func Serve() {}

// ByLen sorts by length.
type ByLen []string

func (s ByLen) Len() int           { return len(s) }
func (s ByLen) Less(i, j int) bool { return len(s[i]) < len(s[j]) }
func (s ByLen) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// Half declares one of the three.
type Half []int

func (h Half) Len() int { return len(h) }
`,
	})
	out := runCommandInDir(t, dir, 1, "bash", skillScript(t, "go-documentation", "check-docs.sh"), "--json", "./...")
	var result struct {
		Missing []struct {
			File     string `json:"file"`
			Line     int    `json:"line"`
			Name     string `json:"name"`
			Advisory bool   `json:"advisory"`
		} `json:"missing"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("parse JSON: %v\n%s", err, out)
	}
	type finding struct {
		file     string
		line     int
		name     string
		advisory bool
	}
	var got []finding
	for _, m := range result.Missing {
		got = append(got, finding{m.File, m.Line, m.Name, m.Advisory})
	}
	want := []finding{
		{"api/api.go", 4, "Serve", false},
		{"api/api.go", 16, "Len", false},
		{"internal/store/store.go", 4, "Open", true},
	}
	if !slices.Equal(got, want) {
		t.Errorf("check-docs.sh ./... missing = %+v, want %+v", got, want)
	}
}

// A call through a value of a named func type converts its argument to the
// interface, as a call through a plain func does.
func TestInterfaceComplianceCallThroughNamedFuncType(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"go.mod": "module namedfunc\n\ngo 1.27\n",
		"sink.go": `package namedfunc

// Sink takes bytes.
type Sink interface{ Put([]byte) }

type buf struct{}

func (buf) Put([]byte) {}

type handler func(Sink)

func run(h handler) { h(buf{}) }
`,
	})
	out := runCommandInDir(t, dir, 0, "bash", skillScript(t, "go-interfaces", "check-interface-compliance.sh"), "--json", ".")
	if !strings.Contains(string(out), `"count_missing":0`) {
		t.Errorf("check-interface-compliance.sh: h(buf{}) through type handler func(Sink) is a conversion to Sink, got\n%s", out)
	}
}

// A branch behind its base is compared with the branch point: the test the
// base gained after it is not the branch's deletion.
func TestPreReviewComparesWithMergeBase(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		runCommandInDir(t, dir, 0, "git", append([]string{"-c", "user.email=t@example.com", "-c", "user.name=t", "-c", "commit.gpgsign=false"}, args...)...)
	}
	const calcTest = "package calc\n\nimport \"testing\"\n\nfunc TestAdd(t *testing.T) {\n\tif Add(1, 2) != 3 {\n\t\tt.Error(\"Add(1, 2) != 3\")\n\t}\n}\n"
	writeTree(t, dir, map[string]string{
		"go.mod":            "module mergebase\n\ngo 1.27\n",
		".golangci.yml":     "version: \"2\"\nlinters:\n  default: none\n",
		"calc/calc.go":      "package calc\n\n// Add returns a + b.\nfunc Add(a, b int) int { return a + b }\n",
		"calc/calc_test.go": calcTest,
	})
	git("init", "-q")
	git("add", ".")
	git("commit", "-q", "-m", "base")
	git("branch", "-M", "base")
	git("checkout", "-q", "-b", "feature")
	writeFile(t, filepath.Join(dir, "calc", "sub.go"), "package calc\n\n// Sub returns a - b.\nfunc Sub(a, b int) int { return a - b }\n")
	git("add", ".")
	git("commit", "-q", "-m", "sub")
	git("checkout", "-q", "base")
	writeFile(t, filepath.Join(dir, "calc", "calc_test.go"), calcTest+"\nfunc TestAddZero(t *testing.T) {\n\tif Add(0, 0) != 0 {\n\t\tt.Error(\"Add(0, 0) != 0\")\n\t}\n}\n")
	git("commit", "-q", "-am", "zero")
	git("checkout", "-q", "feature")

	// An import path target resolves to its directory for gofmt and the diff.
	for _, target := range []string{"./...", "mergebase/..."} {
		out := runCommandInDir(t, dir, 0, "bash", skillScript(t, "go-code-review", "pre-review.sh"), "--json", "--new-from-rev", "base", target)
		if !strings.Contains(string(out), `"bar":{"status":"pass","findings":[]}`) {
			t.Errorf("pre-review.sh --new-from-rev base %s on a branch behind base: want bar pass with no findings, got\n%s", target, out)
		}
	}
}

// check-debt scans what ./... scans, counts a marker only in a comment, and
// takes one path.
func TestCheckDebtScanScope(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	const stray = "package x\n\n// Kept: never read by the go command.\n"
	writeTree(t, dir, map[string]string{
		"pkg/a.go":      "package a\n\nconst usage = \"// Kept: text in a string\"\n\n// Kept: linear scan. Fix: index it when the list grows.\nvar n = 1\n",
		"testdata/x.go": stray,
		"_gen/x.go":     stray,
		".cache/x.go":   stray,
		"pkg/_skip.go":  stray,
	})
	script := skillScript(t, "go-code-refactor", "check-debt.sh")
	out := runCommandInDir(t, dir, 0, "bash", script, "--json", ".")
	var result struct {
		Markers []struct {
			File string `json:"file"`
			Line int    `json:"line"`
		} `json:"markers"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("parse JSON: %v\n%s", err, out)
	}
	if len(result.Markers) != 1 || !pathHasSuffix(result.Markers[0].File, "pkg/a.go") || result.Markers[0].Line != 5 {
		t.Errorf("check-debt.sh . markers = %+v, want the one comment marker at pkg/a.go:5", result.Markers)
	}
	runCommandInDir(t, dir, 2, "bash", script, ".", "pkg")
}

// The logged and returned error is matched by name, whatever the name: cerr
// logged and cerr returned is a finding, cerr logged and err returned is not.
func TestCheckErrorsNamedErrorVariables(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"go.mod": "module namederr\n\ngo 1.27\n",
		"close.go": `package namederr

import (
	"fmt"
	"io"
	"log/slog"
)

func closeBoth(a, b io.Closer) error {
	if cerr := a.Close(); cerr != nil {
		slog.Error("close a", "err", cerr)
		return cerr
	}
	if closeErr := b.Close(); closeErr != nil {
		slog.Error("close b", "err", closeErr)
		return fmt.Errorf("close b: %w", closeErr)
	}
	return nil
}

func degrade(a io.Closer) error {
	err := a.Close()
	if cerr := a.Close(); cerr != nil {
		slog.Warn("second close", "err", cerr)
		return err
	}
	return nil
}
`,
	})
	out := runCommandInDir(t, dir, 1, "bash", skillScript(t, "go-error-handling", "check-errors.sh"), "--json", "./...")
	var result struct {
		Findings []jsonFinding `json:"findings"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("parse JSON: %v\n%s", err, out)
	}
	var lines []int
	for _, f := range result.Findings {
		lines = append(lines, f.Line)
	}
	if want := []int{11, 15}; !slices.Equal(lines, want) {
		t.Errorf("check-errors.sh log-and-return lines = %v, want %v (cerr and closeErr; not degrade)", lines, want)
	}
}

// A go command that cannot answer go env is the environment (exit 2 with a
// message), not exit 1, which the contract reads as findings.
func TestWrappersExitTwoWhenGoEnvFails(t *testing.T) {
	t.Parallel()
	bin := t.TempDir()
	fakeGo := filepath.Join(bin, "go")
	writeFile(t, fakeGo, "#!/bin/sh\nexit 1\n")
	if err := os.Chmod(fakeGo, 0o755); err != nil {
		t.Fatal(err)
	}
	path := strings.Join([]string{bin, "/usr/bin", "/bin"}, string(os.PathListSeparator))
	for _, s := range []struct{ skill, name string }{
		{"go-naming", "check-naming.sh"},
		{"go-documentation", "check-docs.sh"},
		{"go-error-handling", "check-errors.sh"},
		{"go-interfaces", "check-interface-compliance.sh"},
	} {
		cmd := exec.CommandContext(t.Context(), "bash", skillScript(t, s.skill, s.name), ".")
		cmd.Dir = t.TempDir()
		cmd.Env = append(os.Environ(), "PATH="+path, "XDG_CACHE_HOME="+t.TempDir())
		out, err := cmd.CombinedOutput()
		exitErr, ok := errors.AsType[*exec.ExitError](err)
		if !ok || exitErr.ExitCode() != 2 || !strings.Contains(string(out), "go env") {
			t.Errorf("%s with a failing go env: err = %v, output %q; want exit 2 naming go env", s.name, err, out)
		}
	}
}
