package evals_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// installLinkAllow — посилання, яких немає в копії npx skills add (лише skills/),
// і які лишаються навмисно: у тексті сказано, що їх немає в окремій установці.
var installLinkAllow = []string{
	"../../hooks/",
}

func TestSkillsOnlyInstallLinks(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	dst := t.TempDir()
	if err := os.CopyFS(dst, os.DirFS(filepath.Join(root, "skills"))); err != nil {
		t.Fatalf("copy skills: %v", err)
	}
	mdLink := regexp.MustCompile(`\]\(([^)\s]+)\)`)
	tick := regexp.MustCompile("`([^`\\s]+)`")
	var checked int
	err := filepath.WalkDir(dst, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		rel, err := filepath.Rel(dst, path)
		if err != nil {
			t.Fatalf("rel %s: %v", path, err)
		}
		text := string(raw)
		skillRoot := filepath.Join(dst, relSkill(rel))
		var refs []string
		for _, m := range mdLink.FindAllStringSubmatch(text, -1) {
			if ref, ok := installRef(m[1]); ok {
				refs = append(refs, ref)
			}
		}
		for _, m := range tick.FindAllStringSubmatch(text, -1) {
			if ref, ok := installRef(m[1]); ok {
				refs = append(refs, ref)
			}
		}
		for _, ref := range refs {
			checked++
			if allowedInstallLink(ref) {
				continue
			}
			base := filepath.Dir(path)
			if ref == "COMPATIBILITY.md" || strings.HasPrefix(ref, "references/") || strings.HasPrefix(ref, "scripts/") || strings.HasPrefix(ref, "assets/") {
				base = skillRoot
			}
			target := filepath.Join(base, filepath.FromSlash(ref))
			inside, err := filepath.Rel(dst, target)
			if err != nil || strings.HasPrefix(inside, "..") {
				t.Errorf("%s: %s leaves the skills-only install", filepath.ToSlash(rel), ref)
				continue
			}
			if _, err := os.Stat(target); err != nil {
				t.Errorf("%s: %s does not exist in a skills-only install", filepath.ToSlash(rel), ref)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk skills copy: %v", err)
	}
	if checked == 0 {
		t.Fatal("skills copy has no links to check")
	}
}

func relSkill(rel string) string {
	skill, _, _ := strings.Cut(rel, string(filepath.Separator))
	return skill
}

func installRef(token string) (string, bool) {
	if i := strings.IndexByte(token, '#'); i >= 0 {
		token = token[:i]
	}
	if token == "COMPATIBILITY.md" {
		return token, true
	}
	if token == "" || strings.Contains(token, "://") || strings.Contains(token, "...") || strings.ContainsAny(token, "<>*$") {
		return "", false
	}
	switch ext := filepath.Ext(token); ext {
	case ".md", ".sh", ".py", ".yml", ".yaml", ".go", ".json":
	default:
		return "", false
	}
	if strings.HasPrefix(token, "references/") || strings.HasPrefix(token, "scripts/") || strings.HasPrefix(token, "assets/") ||
		strings.HasPrefix(token, "../") || strings.HasPrefix(token, "./") {
		return token, true
	}
	return "", false
}

func allowedInstallLink(ref string) bool {
	for _, prefix := range installLinkAllow {
		if strings.HasPrefix(ref, prefix) {
			return true
		}
	}
	return false
}

func TestRouterSectionsLead(t *testing.T) {
	t.Parallel()
	const maxLine = 200
	const maxBytes = 12 * 1024
	cases := []struct {
		file, heading string
	}{
		{"skills/go-code-refactor/SKILL.md", "## Workflow"},
		{"skills/go-code/SKILL.md", "## Route Before The First Edit"},
	}
	root := repoRoot(t)
	for _, tc := range cases {
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(tc.file)))
		if err != nil {
			t.Fatalf("read %s: %v", tc.file, err)
		}
		off, line := 0, 1
		found := false
		for _, ln := range strings.Split(string(raw), "\n") {
			if ln == tc.heading {
				found = true
				break
			}
			off += len(ln) + 1
			line++
		}
		if !found {
			t.Errorf("%s: no heading %s", tc.file, tc.heading)
			continue
		}
		if line > maxLine && off > maxBytes {
			t.Errorf("%s: %s at line %d, %d bytes from the start; want within %d lines or %d bytes",
				tc.file, tc.heading, line, off, maxLine, maxBytes)
		}
	}
}

func TestHookPrintedLineIsConditional(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	err := filepath.WalkDir(filepath.Join(root, "skills"), func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		text := string(raw)
		for i, line := range strings.Split(text, "\n") {
			if strings.Contains(line, "Skill tool") && !strings.Contains(line, "or read") {
				t.Errorf("%s:%d: Skill tool without an or-read alternative: %s", rel, i+1, line)
			}
		}
		for _, para := range hookParagraphs(text) {
			if !strings.Contains(para, "printed by the hook") && !strings.Contains(para, "printed by the last edit hook") {
				continue
			}
			if !hostHookCondition.MatchString(para) {
				t.Errorf("%s: paragraph names a command printed by the hook without %q:\n%s",
					rel, "Where the plugin hook runs", para)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk skills: %v", err)
	}
}

// hostHookCondition — умова хоста біля команди, яку друкує edit-хук.
// «Where the plugin edit hook runs» і «Where the Claude Code plugin hook runs»
// уже є такою умовою; абзац без жодної з них — ні.
var hostHookCondition = regexp.MustCompile(`Where the (Claude Code )?plugin( edit)? hook runs`)

func hookParagraphs(text string) []string {
	var paras []string
	for _, chunk := range strings.Split(text, "\n\n") {
		paras = append(paras, strings.Split("\n"+chunk, "\n- ")[1:]...)
		if !strings.Contains(chunk, "\n- ") && !strings.HasPrefix(chunk, "- ") {
			paras = append(paras, chunk)
		}
	}
	return paras
}
