package evals_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

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
	cmd := exec.Command(golangciLint, "run", "--config", config, "--output.json.path", jsonPath, "./...")
	cmd.Dir = filepath.Join(root, "evals", "fixtures", "lint")
	if out, err := cmd.CombinedOutput(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 1 {
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
