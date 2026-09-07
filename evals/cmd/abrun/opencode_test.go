package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestValidateOptionsAcceptsOpencodeWithModel(t *testing.T) {
	o := options{corpus: corpusRefactor, runner: runnerOpencode, model: "opencode-go/minimax-m3", reps: 1, parallel: 1, timeout: time.Second}

	if err := validateOptions(o); err != nil {
		t.Fatalf("validateOptions(%+v) error = %v, want nil", o, err)
	}
}

func TestParseOpencodeStream(t *testing.T) {
	// The bash line carries an input shape that is not the skill tool's, and the
	// truncated line is what a killed session leaves behind; neither may cost the
	// parser the events around it.
	transcript := `{"type":"step_start","part":{"type":"step-start"}}
{"type":"tool_use","part":{"type":"tool","tool":"bash","state":{"status":"completed","input":{"command":"ls"}}}}
{"type":"tool_use","part":{"type":"tool","tool":"skill","state":{"status":"completed","input":{"name":"go-code-refactor"}}}}
{"type":"tool_use","part":{"type":"tool","tool":"skill","state":{"status":"completed","input":{"name":"go-code-refactor"}}}}
{"type":"tool_use","part":{"type":"tool","tool":"skill","state":{"status":"completed","input":{"name":"customize-opencode"}}}}
{"type":"text","part":{"type":"text","text":"first pass"}}
{"type":"step_finish","part":{"type":"step-finish","cost":0.01}}
{"type":"text","part":{"type":"text","text":"final answer"}}
{"type":"step_finish","part":{"type":"step-finish","cost":0.02}}
{"type":"step_finish","part":{"type":"step
`

	skills, final, cost := parseOpencodeStream([]byte(transcript))
	if want := []string{"go-code-refactor"}; !reflect.DeepEqual(skills, want) {
		t.Errorf("parseOpencodeStream skills = %v, want %v", skills, want)
	}
	if final != "final answer" {
		t.Errorf("parseOpencodeStream final = %q, want %q", final, "final answer")
	}
	if cost != 0.03 {
		t.Errorf("parseOpencodeStream cost = %v, want 0.03", cost)
	}
}

func TestWriteOpencodeHome(t *testing.T) {
	armDir := t.TempDir()
	writeTestFile(t, filepath.Join(armDir, "skills", "go-code-refactor", "SKILL.md"), "skill body\n")

	tests := []struct {
		name      string
		armDir    string
		wantSkill bool
	}{
		{name: "skilled arm", armDir: armDir, wantSkill: true},
		{name: "control arm", armDir: "", wantSkill: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			if err := writeOpencodeHome(home, tt.armDir, []byte(`{"opencode-go":{"type":"api","key":"k"}}`)); err != nil {
				t.Fatalf("writeOpencodeHome(%q) error = %v, want nil", tt.armDir, err)
			}
			for _, path := range []string{
				filepath.Join(home, ".config", "opencode", "opencode.json"),
				filepath.Join(home, ".local", "share", "opencode", "auth.json"),
			} {
				if _, err := os.Stat(path); err != nil {
					t.Errorf("os.Stat(%q) error = %v, want a written file", path, err)
				}
			}
			skill := filepath.Join(home, ".config", "opencode", "skills", "go-code-refactor", "SKILL.md")
			_, err := os.Stat(skill)
			if tt.wantSkill && err != nil {
				t.Errorf("os.Stat(%q) error = %v, want the arm's skill copied in", skill, err)
			}
			// A control arm home with a skills directory would silently load
			// whatever the operator installed globally and stop being a control.
			if !tt.wantSkill && err == nil {
				t.Errorf("os.Stat(%q) found a skill in the %s arm home", skill, controlArm)
			}
		})
	}
}

func TestOpencodeEnvRedirectsEveryInheritedPath(t *testing.T) {
	// opencode takes its working directory from PWD, so an inherited PWD is not
	// cosmetic: it decides which checkout the session edits.
	t.Setenv("PWD", "/operator/repo")
	t.Setenv("OLDPWD", "/operator/repo")
	t.Setenv("XDG_CONFIG_HOME", "/operator/config")
	t.Setenv("XDG_DATA_HOME", "/operator/data")
	home := t.TempDir()
	work := t.TempDir()

	env := opencodeEnv(home, work)

	want := map[string]string{
		"HOME":            home,
		"PWD":             work,
		"OLDPWD":          work,
		"XDG_CONFIG_HOME": filepath.Join(home, ".config"),
		"XDG_DATA_HOME":   filepath.Join(home, ".local", "share"),
		"XDG_STATE_HOME":  filepath.Join(home, ".local", "state"),
		"XDG_CACHE_HOME":  filepath.Join(home, ".cache"),
	}
	// Later entries win in exec.Cmd.Env, so the effective value is the last one.
	for key, value := range want {
		if got := lastEnv(env, key); got != value {
			t.Errorf("opencodeEnv(%q, %q)[%s] = %q, want %q", home, work, key, got, value)
		}
	}
}

func lastEnv(env []string, key string) string {
	value := ""
	for _, entry := range env {
		if name, v, ok := strings.Cut(entry, "="); ok && name == key {
			value = v
		}
	}
	return value
}
