package evals_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode"
)

// TestRelativeLinks checks every relative Markdown link in the skills, docs/,
// agents/, and the tracked root files: the target exists and a #anchor names
// one of its headings. A dead path sends the model looking for a file that is not
// there; a dead anchor sends it to the top of the wrong file. Code spans and
// fenced blocks are not links.
func TestRelativeLinks(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	var files []string
	err := filepath.WalkDir(filepath.Join(root, "skills"), func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(path, ".md") {
			files = append(files, path)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, pattern := range []string{"docs/*.md", "agents/*.md"} {
		matches, err := filepath.Glob(filepath.Join(root, pattern))
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, matches...)
	}
	// The root holds working notes nobody commits; only its tracked files ship.
	tracked, err := exec.CommandContext(t.Context(), "git", "-C", root, "ls-files", "--", "*.md").Output()
	if err != nil {
		t.Fatal(err)
	}
	for name := range strings.SplitSeq(strings.TrimSpace(string(tracked)), "\n") {
		if name != "" && !strings.Contains(name, "/") {
			files = append(files, filepath.Join(root, name))
		}
	}
	fence := regexp.MustCompile("(?s)```.*?```")
	span := regexp.MustCompile("`[^`]*`")
	link := regexp.MustCompile(`\]\(([^)\s]+)\)`)
	anchors := map[string]map[string]bool{}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		text := span.ReplaceAllString(fence.ReplaceAllString(string(data), ""), "")
		for _, m := range link.FindAllStringSubmatch(text, -1) {
			target := m[1]
			if strings.HasPrefix(target, "http:") || strings.HasPrefix(target, "https:") || strings.HasPrefix(target, "mailto:") {
				continue
			}
			rel, anchor, _ := strings.Cut(target, "#")
			path := f
			if rel != "" {
				path = filepath.Join(filepath.Dir(f), rel)
			}
			name, _ := filepath.Rel(root, f)
			if _, err := os.Stat(path); err != nil {
				t.Errorf("%s -> %s", name, target)
				continue
			}
			if anchor == "" || !strings.HasSuffix(path, ".md") {
				continue
			}
			if anchors[path] == nil {
				anchors[path] = headingSlugs(t, path)
			}
			if !anchors[path][anchor] {
				t.Errorf("%s -> %s (no such heading)", name, target)
			}
		}
	}
}

// headingSlugs are the anchors of a Markdown file's headings, spelled the way
// GitHub spells them: lower case, letters, digits, _ and - kept, spaces to -.
func headingSlugs(t *testing.T, path string) map[string]bool {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	heading := regexp.MustCompile(`^#{1,6}\s`)
	slugs := map[string]bool{}
	for line := range strings.SplitSeq(string(data), "\n") {
		if !heading.MatchString(line) {
			continue
		}
		text := strings.ToLower(strings.TrimSpace(strings.TrimLeft(line, "#")))
		slug := strings.Map(func(r rune) rune {
			switch {
			case r == ' ':
				return '-'
			case r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r):
				return r
			}
			return -1
		}, text)
		slugs[slug] = true
	}
	return slugs
}
