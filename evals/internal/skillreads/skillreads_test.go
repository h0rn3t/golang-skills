package skillreads

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestKey(t *testing.T) {
	tests := []struct {
		path, want string
	}{
		{"/tmp/w/.eval-plugin-1/skills/go-http/SKILL.md", "go-http/SKILL.md"},
		{"/tmp/w/.eval-plugin-1/skills/go-http/references/JSON-V2.md", "go-http/references/JSON-V2.md"},
		{"/home/u/.claude/skills/go-code/references/INTENSITY.md", "go-code/references/INTENSITY.md"},
		{"/tmp/w/skills/go-http/assets/template.md", ""},
		{"/tmp/w/skills/go-http/scripts/check.sh", ""},
		{"/tmp/w/skills/other/SKILL.md", ""},
		{"/tmp/w/skills/other/references/X.md", ""},
		{"/tmp/w/feed/feed.go", ""},
		{"SKILL.md", ""},
		{"", ""},
	}
	for _, tt := range tests {
		if got := Key(tt.path); got != tt.want {
			t.Errorf("Key(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestCount(t *testing.T) {
	// One assistant line per content block, as stream-json emits them: the
	// two Reads of JSON-V2.md are separate calls, the repeated toolu_2 line is
	// the same call, the go-code SKILL.md read is the Skill-tool fallback, and
	// a Glob, a Read of a fixture, a Read in a user message and a line that is
	// not JSON count for nothing.
	stream := strings.Join([]string{
		`{"type":"system","subtype":"init"}`,
		`{"type":"assistant","message":{"id":"m1","content":[{"type":"tool_use","id":"toolu_1","name":"Read","input":{"file_path":"/w/.eval-plugin-1/skills/go-http/references/JSON-V2.md"}}]}}`,
		`{"type":"assistant","message":{"id":"m1","content":[{"type":"tool_use","id":"toolu_2","name":"Read","input":{"file_path":"/w/.eval-plugin-1/skills/go-code/SKILL.md"}}]}}`,
		`{"type":"assistant","message":{"id":"m1","content":[{"type":"tool_use","id":"toolu_2","name":"Read","input":{"file_path":"/w/.eval-plugin-1/skills/go-code/SKILL.md"}}]}}`,
		`{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"toolu_1","content":"..."}]}}`,
		`{"type":"assistant","message":{"id":"m2","content":[{"type":"text","text":"Read JSON-V2.md again"},{"type":"tool_use","id":"toolu_3","name":"Read","input":{"file_path":"/w/.eval-plugin-1/skills/go-http/references/JSON-V2.md","offset":40}}]}}`,
		`{"type":"assistant","message":{"id":"m2","content":[{"type":"tool_use","id":"toolu_4","name":"Glob","input":{"pattern":"skills/go-http/references/*.md"}}]}}`,
		`{"type":"assistant","message":{"id":"m3","content":[{"type":"tool_use","id":"toolu_5","name":"Read","input":{"file_path":"/w/feed/feed.go"}}]}}`,
		`{"type":"user","message":{"content":[{"type":"tool_use","id":"toolu_6","name":"Read","input":{"file_path":"/w/skills/go-http/SKILL.md"}}]}}`,
		`not json`,
	}, "\n")
	want := map[string]int{"go-http/references/JSON-V2.md": 2, "go-code/SKILL.md": 1}
	if got := Count([]byte(stream)); !maps.Equal(got, want) {
		t.Errorf("Count(stream) = %v, want %v", got, want)
	}

	// --output-format json --verbose prints the same messages as one array;
	// a user message whose content is a string must not hide the rest.
	array := `[{"type":"user","message":{"content":"Implement the stub"}},` +
		`{"type":"assistant","message":{"id":"m1","content":[{"type":"tool_use","id":"toolu_1","name":"Read","input":{"file_path":"/w/skills/go-testing/SKILL.md"}}]}},` +
		`{"type":"result","result":"done"}]`
	want = map[string]int{"go-testing/SKILL.md": 1}
	if got := Count([]byte(array)); !maps.Equal(got, want) {
		t.Errorf("Count(array) = %v, want %v", got, want)
	}

	// A block without an id is still one call per message.
	noID := strings.Join([]string{
		`{"type":"assistant","message":{"id":"m1","content":[{"type":"tool_use","name":"Read","input":{"file_path":"/w/skills/go-http/SKILL.md"}}]}}`,
		`{"type":"assistant","message":{"id":"m1","content":[{"type":"tool_use","name":"Read","input":{"file_path":"/w/skills/go-http/SKILL.md"}}]}}`,
		`{"type":"assistant","message":{"id":"m2","content":[{"type":"tool_use","name":"Read","input":{"file_path":"/w/skills/go-http/SKILL.md"}}]}}`,
	}, "\n")
	want = map[string]int{"go-http/SKILL.md": 2}
	if got := Count([]byte(noID)); !maps.Equal(got, want) {
		t.Errorf("Count(noID) = %v, want %v", got, want)
	}

	for _, out := range []string{"", `{"type":"result","result":"done"}`, "not json"} {
		if got := Count([]byte(out)); got != nil {
			t.Errorf("Count(%q) = %v, want nil", out, got)
		}
	}
}

func TestSummary(t *testing.T) {
	total := map[string]int{
		"go-code/SKILL.md":                 1,
		"go-http/SKILL.md":                 0,
		"go-http/references/JSON-V2.md":    3,
		"go-http/references/WEB-SERVER.md": 0,
		"go-code/references/INTENSITY.md":  0,
	}
	want := "skill file reads: 4 call(s); SKILL.md read 1/2, references read 1/3"
	if got := Summary(total); got != want {
		t.Errorf("Summary(%v) = %q, want %q", total, got, want)
	}
}

func TestFiles(t *testing.T) {
	root := t.TempDir()
	for _, rel := range []string{
		"skills/go-http/SKILL.md",
		"skills/go-http/references/JSON-V2.md",
		"skills/go-http/references/WEB-SERVER.md",
		"skills/go-http/assets/template.md",
		"skills/go-code/SKILL.md",
		"skills/notes/SKILL.md",
	} {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want := []string{"go-code/SKILL.md", "go-http/SKILL.md", "go-http/references/JSON-V2.md", "go-http/references/WEB-SERVER.md"}
	if got := Files(root); !slices.Equal(got, want) {
		t.Errorf("Files(root) = %q, want %q", got, want)
	}
	if got := Files(filepath.Join(root, "missing")); got != nil {
		t.Errorf("Files(missing) = %q, want nil", got)
	}
}
