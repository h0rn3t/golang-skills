package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseGrokStream(t *testing.T) {
	full := "---\nname: go-naming\ndescription: d\n---\n# n\n"
	partial := "---\nname: go-http\n"
	transcript := strings.Join([]string{
		grokEvent(t, map[string]any{
			"type": "assistant",
			"message": map[string]any{"content": []any{
				map[string]any{"type": "tool_use", "id": "full", "name": "read_file", "input": map[string]any{
					"target_file": "/tmp/h/.grok/skills/go-naming/SKILL.md",
				}},
				map[string]any{"type": "tool_use", "id": "part", "name": "read_file", "input": map[string]any{
					"target_file": "/tmp/h/.grok/skills/go-http/SKILL.md",
				}},
				map[string]any{"type": "tool_use", "id": "miss", "name": "read_file", "input": map[string]any{
					"target_file": "/tmp/vendor-go-http/SKILL.md",
				}},
				map[string]any{"type": "tool_use", "id": "bad", "name": "Read", "input": map[string]any{
					"path": "/tmp/h/.grok/skills/go-code/SKILL.md",
				}},
				map[string]any{"type": "tool_use", "id": "swap", "name": "read_file", "input": map[string]any{
					"file_path": "/tmp/h/.grok/skills/go-linting/SKILL.md",
				}},
				map[string]any{"type": "tool_use", "id": "sh", "name": "run_terminal_command", "input": map[string]any{"command": "go test"}},
				map[string]any{"type": "tool_use", "name": "Skill", "input": map[string]any{"skill": "golang-skills:go-style-core"}},
			}},
		}),
		grokEvent(t, map[string]any{
			"type": "user",
			"message": map[string]any{"content": []any{
				grokReadResult(t, "full", false, full, 5, "/tmp/h/.grok/skills/go-naming/SKILL.md"),
				grokReadResult(t, "part", false, partial, 4, "/tmp/h/.grok/skills/go-http/SKILL.md"),
				grokReadResult(t, "bad", true, full, 5, "/tmp/h/.grok/skills/go-code/SKILL.md"),
				grokReadResult(t, "swap", false, full, 5, "/tmp/h/.grok/skills/go-naming/SKILL.md"),
			}},
		}),
		grokEvent(t, map[string]any{
			"type": "result", "result": "final answer", "total_cost_usd": 0.02,
		}),
		`{"type":"assistant","message":`,
	}, "\n")

	skills, final, cost, commands := parseGrokStream([]byte(transcript))
	if want := []string{"go-naming", "go-style-core"}; !reflect.DeepEqual(skills, want) {
		t.Errorf("parseGrokStream skills = %v, want %v", skills, want)
	}
	if final != "final answer" {
		t.Errorf("parseGrokStream final = %q, want %q", final, "final answer")
	}
	if cost != 0.02 {
		t.Errorf("parseGrokStream cost = %v, want 0.02", cost)
	}
	if commands != 1 {
		t.Errorf("parseGrokStream commands = %d, want 1", commands)
	}
}

func TestGrokReadCoversDecodedObject(t *testing.T) {
	body := "---\n# go-naming\n"
	content := map[string]any{
		"FileContent": map[string]any{
			"raw_output":    body,
			"total_lines":   2,
			"absolute_path": "/tmp/h/.grok/skills/go-naming/SKILL.md",
		},
	}
	if !grokReadCovers(content, "go-naming") {
		t.Errorf("grokReadCovers(full object, go-naming) = false, want true")
	}
	if grokReadCovers(content, "go-http") {
		t.Errorf("grokReadCovers(full object, go-http) = true, want false")
	}
}

func TestWriteGrokHome(t *testing.T) {
	t.Setenv("GROK_HOME", t.TempDir())
	armDir := t.TempDir()
	writeTestFile(t, filepath.Join(armDir, "skills", "go-code-refactor", "SKILL.md"), "skill body\n")

	tests := []struct {
		name      string
		armDir    string
		wantSkill bool
	}{
		{name: "skilled arm", armDir: armDir, wantSkill: true},
		{name: "control arm", wantSkill: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			if err := writeGrokHome(home, tt.armDir, []byte(`{"token":"k"}`)); err != nil {
				t.Fatalf("writeGrokHome(%q) error = %v, want nil", tt.armDir, err)
			}
			config, err := os.ReadFile(filepath.Join(grokHomeDir(home), "config.toml"))
			if err != nil {
				t.Fatalf("read config: %v", err)
			}
			for _, want := range []string{"use_leader = false", "skills = false"} {
				if !strings.Contains(string(config), want) {
					t.Errorf("config = %q, want substring %q", config, want)
				}
			}
			if _, err := os.Stat(filepath.Join(grokHomeDir(home), "auth.json")); err != nil {
				t.Errorf("auth.json error = %v, want the copied credentials", err)
			}
			skill := filepath.Join(grokHomeDir(home), "skills", "go-code-refactor", "SKILL.md")
			_, err = os.Stat(skill)
			if tt.wantSkill && err != nil {
				t.Errorf("os.Stat(%q) error = %v, want the arm skill", skill, err)
			}
			if !tt.wantSkill && err == nil {
				t.Errorf("os.Stat(%q) found a skill in the control arm", skill)
			}
		})
	}
}

func grokEvent(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	return string(b)
}

func grokReadResult(t *testing.T, id string, isError bool, raw string, lines int, path string) map[string]any {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"FileContent": map[string]any{
			"raw_output":    raw,
			"total_lines":   lines,
			"absolute_path": path,
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal(read result) error = %v", err)
	}
	return map[string]any{
		"type": "tool_result", "tool_use_id": id, "is_error": isError, "content": string(body),
	}
}
