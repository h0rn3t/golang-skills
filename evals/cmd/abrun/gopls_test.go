package main

import (
	"slices"
	"strings"
	"testing"
)

func TestClaudeToolArgs(t *testing.T) {
	tools := "Skill,Read,Glob,Grep,Edit,Write"
	tests := []struct {
		route string
		want  []string
	}{
		{"", []string{"--tools", tools, "--allowed-tools", tools}},
		{goplsCLI, []string{"--tools", tools + ",Bash", "--allowed-tools", tools + ",Bash(gopls:*)"}},
		{goplsMCP, []string{
			"--mcp-config", `{"mcpServers":{"gopls":{"args":["mcp"],"command":"/opt/bin/gopls"}}}`, "--strict-mcp-config",
			"--tools", tools, "--allowed-tools", tools + ",mcp__gopls",
		}},
	}
	for _, tt := range tests {
		got := claudeToolArgs(options{gopls: tt.route, goplsPath: "/opt/bin/gopls"}, tools)
		if !slices.Equal(got, tt.want) {
			t.Errorf("claudeToolArgs(-gopls %q) = %q, want %q", tt.route, got, tt.want)
		}
	}
}

func TestGoplsCalls(t *testing.T) {
	// Два виклики MCP, два CLI (один після cd, один за абсолютним шляхом);
	// rg із gopls у шаблоні, Read і текст відповіді не рахуються.
	stream := strings.Join([]string{
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"mcp__gopls__go_workspace","input":{}},{"type":"tool_use","name":"mcp__gopls__go_search","input":{"query":"Render"}}]}}`,
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"cd feed && gopls references feed.go:12:6"}}]}}`,
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"/home/u/go/bin/gopls check feed.go"}}]}}`,
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"rg goplsConfig ."}}]}}`,
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read","input":{"file_path":"/x/GOPLS.md"}},{"type":"text","text":"gopls references"}]}}`,
		`not json`,
	}, "\n")
	if got, want := goplsCalls([]byte(stream)), 4; got != want {
		t.Errorf("goplsCalls(stream) = %d, want %d", got, want)
	}
}

func TestValidateOptionsGopls(t *testing.T) {
	base := options{corpus: corpusImplement, runner: runnerClaude, reps: 1, parallel: 1, timeout: 1}
	tests := []struct {
		route, runner string
		wantErr       bool
	}{
		{goplsMCP, runnerClaude, false},
		{goplsCLI, runnerClaude, false},
		{"lsp", runnerClaude, true},
		{goplsMCP, runnerCodex, true},
	}
	for _, tt := range tests {
		o := base
		o.gopls, o.runner = tt.route, tt.runner
		if err := validateOptions(o); (err != nil) != tt.wantErr {
			t.Errorf("validateOptions(-gopls %q -runner %s) error = %v, want error %v", tt.route, tt.runner, err, tt.wantErr)
		}
	}
}
