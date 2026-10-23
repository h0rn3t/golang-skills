package main

import (
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestClaudeRouting(t *testing.T) {
	// A session that explores through Bash, edits, is blocked, then loads two
	// skills in one message (two lines, one id) and one more later.
	stream := strings.Join([]string{
		`{"type":"assistant","message":{"id":"m1","content":[{"type":"tool_use","name":"Bash","input":{"command":"ls"}}]}}`,
		`{"type":"assistant","message":{"id":"m2","content":[{"type":"tool_use","name":"Edit","input":{"file_path":"/x/a.go"}}]}}`,
		`{"type":"system","subtype":"hook_response","hook_name":"PreToolUse:Edit","exit_code":2}`,
		`{"type":"system","subtype":"hook_response","hook_name":"PostToolUse:Edit","exit_code":2}`,
		`{"type":"assistant","message":{"id":"m3","content":[{"type":"tool_use","name":"Skill","input":{"skill":"golang-skills:go-code"}}]}}`,
		`{"type":"assistant","message":{"id":"m3","content":[{"type":"tool_use","name":"Skill","input":{"skill":"golang-skills:go-style-core"}}]}}`,
		`{"type":"assistant","message":{"id":"m4","content":[{"type":"tool_use","name":"Skill","input":{"skill":"golang-skills:go-testing"}}]}}`,
		`{"type":"system","subtype":"hook_response","hook_name":"PreToolUse:Write","exit_code":0}`,
		`not json`,
	}, "\n")
	got := claudeRouting([]byte(stream))
	if got.FirstTool != "Bash" || !got.EditBeforeLoad || got.GateBlocks != 1 || !slices.Equal(got.FirstLoads, []string{"go-code", "go-style-core"}) {
		t.Errorf("claudeRouting(stream) = %+v, want first Bash, edit before load, 1 block, first loads go-code and go-style-core", got)
	}

	clean := `{"type":"assistant","message":{"id":"m1","content":[{"type":"tool_use","name":"Skill","input":{"skill":"go-code"}}]}}` + "\n" +
		`{"type":"assistant","message":{"id":"m2","content":[{"type":"tool_use","name":"Write","input":{"file_path":"/x/a.go"}}]}}`
	if got := claudeRouting([]byte(clean)); got.FirstTool != "Skill" || got.EditBeforeLoad || got.GateBlocks != 0 {
		t.Errorf("claudeRouting(clean) = %+v, want first Skill, no edit before load, no blocks", got)
	}
}

func TestRoutingStatsAdd(t *testing.T) {
	st := routingStats{FirstTool: "Skill", FirstLoads: []string{"go-code"}, GateBlocks: 1}
	st.add(routingStats{FirstTool: "Edit", GateBlocks: 2})
	if st.FirstTool != "Skill" || st.GateBlocks != 3 {
		t.Errorf("add(repair turn) = %+v, want the first turn's opening and 3 blocks", st)
	}
}

func TestContractFirst(t *testing.T) {
	edit := func(name, path string) string {
		return `{"type":"assistant","message":{"id":"m","content":[{"type":"tool_use","name":"` + name + `","input":{"file_path":"` + path + `"}}]}}`
	}
	tests := []struct {
		name   string
		stream []string
		want   *bool
	}{
		{"test first", []string{edit("Write", "/w/signup/signup_test.go"), edit("Edit", "/w/signup/signup.go")}, new(true)},
		{"body first", []string{edit("Edit", "/w/signup/signup.go"), edit("Write", "/w/signup/signup_test.go")}, new(false)},
		{"markdown edit does not decide", []string{edit("Write", "/w/NOTES.md"), edit("MultiEdit", "/w/signup/signup_test.go")}, new(true)},
		{"no go edit", []string{edit("Write", "/w/NOTES.md")}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := claudeRouting([]byte(strings.Join(tt.stream, "\n"))).ContractFirst
			if (got == nil) != (tt.want == nil) || got != nil && *got != *tt.want {
				t.Errorf("claudeRouting(%s).ContractFirst = %v, want %v", tt.name, ptrString(got), ptrString(tt.want))
			}
		})
	}

	// A repair turn cannot rewrite what the opening turn already decided.
	st := routingStats{ContractFirst: new(false), FirstTool: "Edit"}
	st.add(routingStats{ContractFirst: new(true)})
	if st.ContractFirst == nil || *st.ContractFirst {
		t.Errorf("add(repair turn).ContractFirst = %v, want false from the opening turn", ptrString(st.ContractFirst))
	}

	rep := report{Results: []result{
		{Arm: "a", Routing: &routingStats{ContractFirst: new(true)}},
		{Arm: "a", Routing: &routingStats{ContractFirst: new(false)}},
		{Arm: "a", Routing: &routingStats{}},
		{Arm: "a", Err: "timeout", Routing: &routingStats{ContractFirst: new(true)}},
		{Arm: "b", Routing: &routingStats{ContractFirst: new(true)}},
	}}
	if first, measured := contractFirst(rep, "a"); first != 1 || measured != 2 {
		t.Errorf("contractFirst(rep, a) = %d/%d, want 1/2", first, measured)
	}
}

func ptrString(b *bool) string {
	if b == nil {
		return "nil"
	}
	return strconv.FormatBool(*b)
}
