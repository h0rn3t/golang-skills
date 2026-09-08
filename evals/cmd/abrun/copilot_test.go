package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestValidateOptionsAcceptsCopilotWithoutModel(t *testing.T) {
	// Unlike an opencode arm home, a copilot home falls back to the account's
	// own default model, so -model stays optional.
	o := options{corpus: corpusRefactor, runner: runnerCopilot, reps: 1, parallel: 1, timeout: time.Second}

	if err := validateOptions(o); err != nil {
		t.Fatalf("validateOptions(%+v) error = %v, want nil", o, err)
	}
}

func TestParseCopilotStream(t *testing.T) {
	// The view line carries an argument shape that is not the skill tool's, and
	// the truncated line is what a killed session leaves behind; neither may cost
	// the parser the events around it.
	transcript := `{"type":"session.info","data":{"model":"copilot-test-model"}}
{"type":"tool.execution_start","data":{"toolCallId":"c1","toolName":"view","arguments":{"path":"/tmp/report.go"}}}
{"type":"tool.execution_start","data":{"toolCallId":"c2","toolName":"skill","arguments":{"skill":"go-code-refactor"}}}
{"type":"tool.execution_start","data":{"toolCallId":"c3","toolName":"skill","arguments":{"skill":"go-code-refactor"}}}
{"type":"tool.execution_start","data":{"toolCallId":"c4","toolName":"skill","arguments":{"skill":"github-pr-media"}}}
{"type":"assistant.message","data":{"content":"first pass","phase":"tool_call"}}
{"type":"assistant.message","data":{"content":"final answer","phase":"final_answer"}}
{"type":"assistant.message","data":{"content":"
`

	skills, final, cost := parseCopilotStream([]byte(transcript))
	if want := []string{"go-code-refactor"}; !reflect.DeepEqual(skills, want) {
		t.Errorf("parseCopilotStream skills = %v, want %v", skills, want)
	}
	if final != "final answer" {
		t.Errorf("parseCopilotStream final = %q, want %q", final, "final answer")
	}
	// Copilot reports premium requests and AI credits, never dollars, so there is
	// no honest $/run to report and the summary must leave the column empty.
	if cost != 0 {
		t.Errorf("parseCopilotStream cost = %v, want 0", cost)
	}
}

func TestWriteCopilotHome(t *testing.T) {
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
			if err := writeCopilotHome(home, tt.armDir); err != nil {
				t.Fatalf("writeCopilotHome(%q) error = %v, want nil", tt.armDir, err)
			}
			skill := filepath.Join(home, "skills", "go-code-refactor", "SKILL.md")
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

func TestCopilotEnvRedirectsConfigAndWorkingDirectory(t *testing.T) {
	// HOME is deliberately left alone: the credential store copilot authenticates
	// against is keyed to it, and redirecting it fails every run unauthenticated.
	t.Setenv("HOME", "/operator/home")
	t.Setenv("COPILOT_HOME", "/operator/home/.copilot")
	t.Setenv("PWD", "/operator/repo")
	t.Setenv("OLDPWD", "/operator/repo")
	home := t.TempDir()
	work := t.TempDir()

	env := copilotEnv(home, work)

	want := map[string]string{
		"HOME":         "/operator/home",
		"COPILOT_HOME": home,
		"PWD":          work,
		"OLDPWD":       work,
	}
	// Later entries win in exec.Cmd.Env, so the effective value is the last one.
	for key, value := range want {
		if got := lastEnv(env, key); got != value {
			t.Errorf("copilotEnv(%q, %q)[%s] = %q, want %q", home, work, key, got, value)
		}
	}
}

func TestCopilotSessionDropsSkillToolForControlArm(t *testing.T) {
	// The control arm home carries no skills, but leaving the tool in place would
	// still let the session reach a builtin one, which the claude control cannot.
	if copilotTools[0] != "skill" {
		t.Fatalf("copilotTools[0] = %q, want the skill tool first so the control arm can drop it", copilotTools[0])
	}
	for _, tool := range copilotTools[1:] {
		if tool == "skill" {
			t.Errorf("copilotTools = %v, want the skill tool listed once", copilotTools)
		}
	}
}
