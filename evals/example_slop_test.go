package evals_test

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestPositiveExamplesCarryNoSlop scans every Go block a skill shows as the
// right form: a model copies the code, not the caveat beside it. A block whose
// last line of prose before the fence says Bad, Before, or fragment is skipped,
// and inside a block so are the lines from a `// Bad` or `// Before` comment to
// the next `// Good` or `// After`.
func TestPositiveExamplesCarryNoSlop(t *testing.T) {
	root := repoRoot(t)
	var files []string
	for _, pattern := range []string{"skills/*/SKILL.md", "skills/*/references/*.md"} {
		matches, err := filepath.Glob(filepath.Join(root, pattern))
		if err != nil {
			t.Fatalf("glob %s: %v", pattern, err)
		}
		files = append(files, matches...)
	}

	slop := regexp.MustCompile(`// ---|// go-[a-z-]+:|(?i:failed to|could not|couldn't)`)
	negativeProse := regexp.MustCompile(`(?i)\b(bad|before|fragment)\b`)
	negativeStart := regexp.MustCompile(`(?i)^//\s*(bad|before|avoid|wrong)\b`)
	negativeEnd := regexp.MustCompile(`(?i)^//\s*(good|after|better|fix)\b`)

	for _, path := range files {
		rel, _ := filepath.Rel(root, path)
		lines := strings.Split(readFile(t, path), "\n")
		prose := ""
		for i := 0; i < len(lines); i++ {
			trimmed := strings.TrimSpace(lines[i])
			if !strings.HasPrefix(trimmed, "```go") {
				if trimmed != "" && !strings.HasPrefix(trimmed, "```") {
					prose = trimmed
				}
				continue
			}
			negative := negativeProse.MatchString(prose)
			inNegative := false
			for i++; i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "```"); i++ {
				code := strings.TrimSpace(lines[i])
				switch {
				case negativeStart.MatchString(code):
					inNegative = true
				case negativeEnd.MatchString(code):
					inNegative = false
				}
				if !negative && !inNegative && slop.MatchString(code) {
					t.Errorf("%s:%d: positive example carries %q: %s", rel, i+1, slop.FindString(code), code)
				}
			}
		}
	}
}
