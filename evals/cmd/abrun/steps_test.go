package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// stepsJSON renders st as the report stores it, so a case pins the wire names
// a replay script reads and tells a zero count from an absent one.
func stepsJSON(t *testing.T, st *stepStats) string {
	t.Helper()
	b, err := json.Marshal(st)
	if err != nil {
		t.Fatalf("json.Marshal(%+v) error = %v", st, err)
	}
	return string(b)
}

func TestClaudeSteps(t *testing.T) {
	tests := []struct {
		name   string
		stream []string
		want   string
	}{
		{
			name: "loads, reads, then Go edits",
			stream: []string{
				`{"type":"system","subtype":"init"}`,
				`{"type":"assistant","message":{"id":"m1","content":[{"type":"tool_use","id":"t1","name":"Skill","input":{"skill":"golang-skills:go-code"}}]}}`,
				`{"type":"assistant","message":{"id":"m1","content":[{"type":"tool_use","id":"t2","name":"Skill","input":{"skill":"golang-skills:go-style-core"}}]}}`,
				`{"type":"assistant","message":{"id":"m2","content":[{"type":"tool_use","id":"t3","name":"Read","input":{"file_path":"/w/a/a.go"}},{"type":"tool_use","id":"t4","name":"Glob","input":{"pattern":"go.mod"}}]}}`,
				`{"type":"assistant","message":{"id":"m3","content":[{"type":"text","text":"Edit a.go next"},{"type":"tool_use","id":"t5","name":"Write","input":{"file_path":"/w/a/NOTES.md"}}]}}`,
				`{"type":"assistant","message":{"id":"m4","content":[{"type":"tool_use","id":"t6","name":"Edit","input":{"file_path":"/w/a/a_test.go"}}]}}`,
				`{"type":"assistant","message":{"id":"m4","content":[{"type":"tool_use","id":"t7","name":"Edit","input":{"file_path":"/w/a/a.go"}}]}}`,
				`not json`,
				`{"type":"assistant","message":{"id":"m5","content":[{"type":"tool_use","id":"t8","name":"Bash","input":{"command":"go test ./..."}}]}}`,
				// A session whose work ran in a fork reports zeros in usage and
				// the real totals per model.
				`{"type":"result","subtype":"success","usage":{"output_tokens":0},"modelUsage":{"claude-sonnet-5-5":{"outputTokens":880,"thinkingTokens":300},"claude-haiku-5-5":{"outputTokens":20,"thinkingTokens":0}}}`,
			},
			want: `{"turns":5,"tools":{"Bash":1,"Edit":2,"Glob":1,"Read":1,"Skill":2,"Write":1},"before_first_go_edit":5,"output_tokens":900,"thinking_tokens":300}`,
		},
		{
			name: "review cut short: no Go edit, no result event",
			stream: []string{
				`{"type":"assistant","message":{"id":"m1","content":[{"type":"tool_use","id":"t1","name":"Read","input":{"file_path":"/w/a/a.go"}}]}}`,
				`{"type":"assistant","message":{"id":"m2","content":[{"type":"text","text":"a.go:3: Must Fix"}]}}`,
			},
			want: `{"turns":2,"tools":{"Read":1}}`,
		},
		{
			name: "Go edit is the first call",
			stream: []string{
				`{"type":"assistant","message":{"id":"m1","content":[{"type":"tool_use","id":"t1","name":"Write","input":{"file_path":"/w/a/a.go"}}]}}`,
			},
			want: `{"turns":1,"tools":{"Write":1},"before_first_go_edit":0}`,
		},
		{
			name: "no assistant message",
			stream: []string{
				`{"type":"system","subtype":"init"}`,
				`{"type":"result","subtype":"error_during_execution","modelUsage":{}}`,
			},
			want: `null`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stream := strings.Join(tt.stream, "\n")
			if got := stepsJSON(t, claudeSteps([]byte(stream))); got != tt.want {
				t.Errorf("claudeSteps(%s) =\n%s, want\n%s", tt.name, got, tt.want)
			}
		})
	}
}

func TestMergeSteps(t *testing.T) {
	tests := []struct {
		name  string
		turns []*stepStats
		want  string
	}{
		{
			name: "repair turn makes the first Go edit",
			turns: []*stepStats{
				{Turns: 2, Tools: map[string]int{"Read": 2, "Skill": 1}, OutputTokens: 100, ThinkingTokens: 10},
				nil,
				{Turns: 1, Tools: map[string]int{"Edit": 1, "Read": 1}, BeforeFirstGoEdit: new(1), OutputTokens: 50, ThinkingTokens: 5},
			},
			want: `{"turns":3,"tools":{"Edit":1,"Read":3,"Skill":1},"before_first_go_edit":4,"output_tokens":150,"thinking_tokens":15}`,
		},
		{
			name: "first turn already edited Go",
			turns: []*stepStats{
				{Turns: 1, Tools: map[string]int{"Edit": 1, "Read": 2}, BeforeFirstGoEdit: new(2)},
				{Turns: 1, Tools: map[string]int{"Edit": 1}, BeforeFirstGoEdit: new(0)},
			},
			want: `{"turns":2,"tools":{"Edit":2,"Read":2},"before_first_go_edit":2}`,
		},
		{
			name:  "runner that reports no steps",
			turns: []*stepStats{nil, nil},
			want:  `null`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var r result
			for _, st := range tt.turns {
				r.merge(sessionTurn{steps: st})
			}
			if got := stepsJSON(t, r.Steps); got != tt.want {
				t.Errorf("Steps after merging %s =\n%s, want\n%s", tt.name, got, tt.want)
			}
		})
	}
}

func TestStepsLine(t *testing.T) {
	tests := []struct {
		name    string
		results []result
		want    string
	}{
		{
			name: "means over the arm's completed runs",
			results: []result{
				{Arm: "baseline", Steps: &stepStats{Turns: 4, Tools: map[string]int{"Skill": 4, "Read": 2, "Edit": 2}, BeforeFirstGoEdit: new(6), OutputTokens: 1000, ThinkingTokens: 200}},
				{Arm: "baseline", Steps: &stepStats{Turns: 2, Tools: map[string]int{"Read": 2, "Glob": 2}}},
				{Arm: "baseline", Err: "timeout", Steps: &stepStats{Turns: 9, Tools: map[string]int{"Bash": 9}}},
				{Arm: "baseline"},
				{Arm: "no-skill", Steps: &stepStats{Turns: 7, Tools: map[string]int{"Read": 7}}},
			},
			want: "steps/run: 6.0 tool calls (Read 2.0, Skill 2.0, Edit 1.0, Glob 1.0), 3.0 turns over 2 run(s)" +
				"; 6.0 calls before the first Go edit (1 run(s))" +
				"; output 1000 tokens, 200 of them thinking (1 run(s))",
		},
		{
			name:    "runs with no Go edit and no token reading",
			results: []result{{Arm: "baseline", Steps: &stepStats{Turns: 3, Tools: map[string]int{"Read": 3, "Skill": 1}}}},
			want:    "steps/run: 4.0 tool calls (Read 3.0, Skill 1.0), 3.0 turns over 1 run(s)",
		},
		{
			name:    "report written before steps were recorded",
			results: []result{{Arm: "baseline"}},
			want:    "steps/run: — (not measured in this report)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := stepsLine(report{Results: tt.results}, "baseline"); got != tt.want {
				t.Errorf("stepsLine(%s, baseline) =\n%q, want\n%q", tt.name, got, tt.want)
			}
		})
	}
}
