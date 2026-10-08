package main

import (
	"slices"
	"strings"
	"testing"
)

func TestClaudeToolArgsShell(t *testing.T) {
	tools := "Skill,Read,Glob,Grep,Edit,Write"
	shell := ",Bash(go:*),Bash(gofmt:*),Bash(golangci-lint:*),Bash(govulncheck:*),Bash(cd:*),Bash(bash:*)"
	tests := []struct {
		name string
		o    options
		want []string
	}{
		{"shell go", options{shell: shellGo}, []string{"--tools", tools + ",Bash", "--allowed-tools", tools + shell}},
		{"shell go with gopls cli", options{shell: shellGo, gopls: goplsCLI}, []string{"--tools", tools + ",Bash", "--allowed-tools", tools + ",Bash(gopls:*)" + shell}},
	}
	for _, tt := range tests {
		if got := claudeToolArgs(tt.o, tools); !slices.Equal(got, tt.want) {
			t.Errorf("claudeToolArgs(%s) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestValidateOptionsShell(t *testing.T) {
	base := options{corpus: corpusImplement, runner: runnerClaude, reps: 1, parallel: 1, timeout: 1}
	tests := []struct {
		shell, runner string
		wantErr       bool
	}{
		{shellGo, runnerClaude, false},
		{"bash", runnerClaude, true},
		{shellGo, runnerCodex, true},
	}
	for _, tt := range tests {
		o := base
		o.shell, o.runner = tt.shell, tt.runner
		if err := validateOptions(o); (err != nil) != tt.wantErr {
			t.Errorf("validateOptions(-shell %q -runner %s) error = %v, want error %v", tt.shell, tt.runner, err, tt.wantErr)
		}
	}
}

func TestClaudeCommands(t *testing.T) {
	// Two Bash calls in one message and one in another; a Read and answer text
	// naming Bash do not count.
	stream := strings.Join([]string{
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"go test ./..."}},{"type":"tool_use","name":"Bash","input":{"command":"gofmt -l ."}}]}}`,
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read","input":{"file_path":"/x/a.go"}},{"type":"text","text":"Bash"}]}}`,
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"golangci-lint run ./..."}}]}}`,
		`not json`,
	}, "\n")
	if got, want := claudeCommands([]byte(stream)), 3; got != want {
		t.Errorf("claudeCommands(stream) = %d, want %d", got, want)
	}
}

func TestClaudeRefused(t *testing.T) {
	tests := []struct {
		name   string
		stream string
		want   bool
	}{
		{"result line", `{"type":"result","subtype":"success","stop_reason":"refusal"}`, true},
		{"assistant message", `{"type":"assistant","message":{"stop_reason":"refusal","content":[]}}`, true},
		{"end turn", `{"type":"assistant","message":{"stop_reason":null,"content":[{"type":"text","text":"refusal"}]}}` + "\n" + `{"type":"result","stop_reason":"end_turn"}`, false},
		{"not json", `not json "stop_reason":"refusal"`, false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := claudeRefused([]byte(tt.stream)); got != tt.want {
				t.Errorf("claudeRefused(%q) = %v, want %v", tt.stream, got, tt.want)
			}
		})
	}
}
