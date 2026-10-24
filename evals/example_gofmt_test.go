package evals_test

import (
	"bytes"
	"go/format"
	"path/filepath"
	"strings"
	"testing"
)

// TestFileExamplesAreGofmtClean holds every Go block that shows a whole file
// (it has a package clause) to gofmt's output: a model copies the indentation,
// the import order, and the doc comment layout along with the code. A block
// that does not parse as a file (an elided body, two files in one block) is
// left to the compile tests.
func TestFileExamplesAreGofmtClean(t *testing.T) {
	root := repoRoot(t)
	var files []string
	for _, pattern := range []string{"skills/*/SKILL.md", "skills/*/references/*.md"} {
		matches, err := filepath.Glob(filepath.Join(root, pattern))
		if err != nil {
			t.Fatalf("glob %s: %v", pattern, err)
		}
		files = append(files, matches...)
	}

	for _, path := range files {
		rel, _ := filepath.Rel(root, path)
		lines := strings.Split(readFile(t, path), "\n")
		for i := 0; i < len(lines); i++ {
			indent, ok := strings.CutSuffix(lines[i], "```go")
			if !ok || strings.TrimSpace(indent) != "" {
				continue
			}
			start := i + 1
			var block strings.Builder
			for i++; i < len(lines) && strings.TrimSpace(lines[i]) != "```"; i++ {
				block.WriteString(strings.TrimPrefix(lines[i], indent))
				block.WriteByte('\n')
			}
			src := []byte(block.String())
			if !bytes.HasPrefix(src, []byte("package ")) && !bytes.Contains(src, []byte("\npackage ")) {
				continue
			}
			got, err := format.Source(src)
			if err == nil && !bytes.Equal(got, src) {
				t.Errorf("%s:%d: file example is not gofmt-clean; gofmt writes:\n%s", rel, start, got)
			}
		}
	}
}
