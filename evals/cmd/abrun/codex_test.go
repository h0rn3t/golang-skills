package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestValidateOptionsEffortIsRunnerSpecific(t *testing.T) {
	tests := []struct {
		name    string
		runner  string
		effort  string
		wantErr bool
	}{
		{name: "codex with effort", runner: runnerCodex, effort: "xhigh"},
		{name: "copilot with effort", runner: runnerCopilot, effort: "high"},
		{name: "codex without effort", runner: runnerCodex},
		// Recording a run as xhigh that the CLI could not ask for would put a
		// condition in the report that never reached the model.
		{name: "claude with effort", runner: runnerClaude, effort: "medium"},
		{name: "opencode with effort", runner: runnerOpencode, effort: "xhigh", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := options{corpus: corpusRefactor, runner: tt.runner, model: "m", effort: tt.effort, reps: 1, parallel: 1, timeout: time.Second}
			err := validateOptions(o)
			if tt.wantErr && err == nil {
				t.Errorf("validateOptions(%+v) error = nil, want an error", o)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateOptions(%+v) error = %v, want nil", o, err)
			}
		})
	}
}

func TestParseCodexStream(t *testing.T) {
	// Codex has no skill tool: a skill fires when the model cats its SKILL.md,
	// and the file body comes back in the same transcript. The go-http mention
	// below sits in that output and in a fixture path, where a skill the model
	// never opened must not be scored as one it read.
	transcript := `Reading additional input from stdin...
{"type":"thread.started","thread_id":"t1"}
{"type":"item.completed","item":{"id":"i0","type":"error","message":"Skill descriptions were shortened"}}
{"type":"item.completed","item":{"id":"i1","type":"command_execution","command":"/bin/zsh -lc 'cat /tmp/h/.codex/skills/go-code-refactor/SKILL.md'","exit_code":"0","aggregated_output":"---\nname: go-code-refactor\ndescription: Refactor Go code.\n---\nsee go-http/SKILL.md for servers"}}
{"type":"item.completed","item":{"id":"i2","type":"command_execution","command":"/bin/zsh -lc 'cat /tmp/h/.codex/skills/r0/go-style-core/SKILL.md'","exit_code":"0","aggregated_output":"---\nname: go-style-core\ndescription: Go style.\n---\n# Go Style"}}
{"type":"item.completed","item":{"id":"i3","type":"command_execution","command":"/bin/zsh -lc 'ls /tmp/w/golang-skills/952e678d/report'","exit_code":"0"}}
{"type":"item.completed","item":{"id":"i4","type":"agent_message","text":"first pass"}}
{"type":"item.completed","item":{"id":"i5","type":"agent_message","text":"final answer"}}
{"type":"item.completed","item":{"id":"i6","type":"agent_
`

	skills, final, cost := parseCodexStream([]byte(transcript))
	if want := []string{"go-code-refactor", "go-style-core"}; !reflect.DeepEqual(skills, want) {
		t.Errorf("parseCodexStream skills = %v, want %v", skills, want)
	}
	if final != "final answer" {
		t.Errorf("parseCodexStream final = %q, want %q", final, "final answer")
	}
	// Codex reports token counts, never dollars, so there is no honest $/run.
	if cost != 0 {
		t.Errorf("parseCodexStream cost = %v, want 0", cost)
	}
}

func TestParseCodexStreamRequiresReadEvidence(t *testing.T) {
	const path = "/tmp/h/.codex/skills/go-http/SKILL.md"
	const body = "---\nname: go-http\ndescription: Go HTTP.\n---\n# Go HTTP\n"
	for _, tt := range []struct {
		name, event, command, output, exit string
		want                               bool
	}{
		{name: "cat numeric exit", event: "item.completed", command: "cat " + path, output: body, exit: "0", want: true},
		{name: "sed string exit", event: "item.completed", command: "sed -n '1,80p' " + path, output: body, exit: `"0"`, want: true},
		{name: "failed read", event: "item.completed", command: "cat " + path, output: "cat: " + path + ": No such file or directory", exit: "1"},
		{name: "failed after output", event: "item.completed", command: "cat " + path, output: body, exit: "1"},
		{name: "absent conditional", event: "item.completed", command: "if [ -f " + path + " ]; then cat " + path + "; else echo ABSENT; fi", output: "ABSENT\n", exit: "0"},
		{name: "echoed path", event: "item.completed", command: "echo " + path, output: path + "\n", exit: "0"},
		{name: "started event", event: "item.started", command: "cat " + path, output: body, exit: "0"},
		{name: "missing exit status", event: "item.completed", command: "cat " + path, output: body, exit: "null"},
		{name: "missing output", event: "item.completed", command: "cat " + path, exit: "0"},
		{name: "partial read without identity", event: "item.completed", command: "sed -n '20,40p' " + path, output: "# Go HTTP\nRead go-http/SKILL.md.\n", exit: "0"},
		{name: "cross reference only", event: "item.completed", command: "cat " + path, output: "---\nname: go-code\ndescription: Go code.\n---\nRead go-http/SKILL.md.\n", exit: "0"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			transcript := fmt.Sprintf(`{"type":%q,"item":{"type":"command_execution","command":%q,"aggregated_output":%q,"exit_code":%s}}`, tt.event, tt.command, tt.output, tt.exit)
			skills, _, _ := parseCodexStream([]byte(transcript))
			if got := len(skills) == 1 && skills[0] == "go-http"; got != tt.want || len(skills) > 1 {
				t.Errorf("parseCodexStream(%s) skills = %v, want go-http read=%v", tt.name, skills, tt.want)
			}
		})
	}
}

func TestSkillsInPaths(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []string
	}{
		{
			// The shape codex renders in the prompt listing: a root alias, and
			// no skills/ segment anywhere in it.
			name: "prompt listing",
			text: "- go-http: Use when writing Go HTTP code (file: r0/go-http/SKILL.md)\n- go-naming: ... (file: r0/go-naming/SKILL.md)",
			want: []string{"go-http", "go-naming"},
		},
		{
			name: "shell command",
			text: "/bin/zsh -lc 'cat /tmp/h/.codex/skills/go-code-refactor/SKILL.md'",
			want: []string{"go-code-refactor"},
		},
		{
			// The scratch tree lives under a path ending in "golang-skills",
			// which a looser reading would treat as a skills directory.
			name: "fixture path",
			text: "cat /private/tmp/-Users-x-golang-skills/952e678d/report/report.go",
		},
		{
			name: "directory ending in a skill name",
			text: "cat /tmp/vendor-go-http/SKILL.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := skillsInPaths(tt.text)
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("skillsInPaths(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}

func TestWriteCodexHome(t *testing.T) {
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
			if err := writeCodexHome(home, tt.armDir, []byte(`{"tokens":{}}`)); err != nil {
				t.Fatalf("writeCodexHome(%q) error = %v, want nil", tt.armDir, err)
			}
			auth := filepath.Join(codexHomeDir(home), "auth.json")
			if _, err := os.Stat(auth); err != nil {
				t.Errorf("os.Stat(%q) error = %v, want the credentials copied in", auth, err)
			}
			skill := filepath.Join(codexHomeDir(home), "skills", "go-code-refactor", "SKILL.md")
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

func TestCodexEnvRedirectsHomeAndWorkingDirectory(t *testing.T) {
	// CODEX_HOME alone leaves host skill discovery intact, so HOME is the part
	// that makes the control arm a control; it also keeps codex's `zsh -lc` from
	// sourcing the operator's shell profile into every session.
	t.Setenv("HOME", "/operator/home")
	t.Setenv("CODEX_HOME", "/operator/home/.codex")
	t.Setenv("PWD", "/operator/repo")
	t.Setenv("OLDPWD", "/operator/repo")
	home := t.TempDir()
	work := t.TempDir()

	env := codexEnv(home, work)

	want := map[string]string{
		"HOME":            home,
		"CODEX_HOME":      filepath.Join(home, ".codex"),
		"PWD":             work,
		"OLDPWD":          work,
		"XDG_CONFIG_HOME": filepath.Join(home, ".config"),
		"XDG_DATA_HOME":   filepath.Join(home, ".local", "share"),
	}
	// Later entries win in exec.Cmd.Env, so the effective value is the last one.
	for key, value := range want {
		if got := lastEnv(env, key); got != value {
			t.Errorf("codexEnv(%q, %q)[%s] = %q, want %q", home, work, key, got, value)
		}
	}
	if strings.Contains(lastEnv(env, "HOME"), "operator") {
		t.Errorf("codexEnv left the operator's HOME in place")
	}
}
