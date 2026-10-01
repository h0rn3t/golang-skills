package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// JSON-проекція перевіряє wire-контракт, зокрема відсутні legacy-поля.
type confirmedRoutingContract struct {
	Version                              int      `json:"version"`
	ToolEvidence                         string   `json:"tool_evidence"`
	HookEvidence                         string   `json:"hook_evidence"`
	BeforeFirstGoEditEvidence            string   `json:"before_first_go_edit_evidence"`
	FirstLoads                           []string `json:"first_loads"`
	LoadsBeforeFirstGoEdit               []string `json:"loads_before_first_go_edit"`
	RouterBeforeFirstGoEdit              *bool    `json:"router_before_first_go_edit"`
	StyleCoreBeforeFirstGoEdit           *bool    `json:"style_core_before_first_go_edit"`
	GoEditAttempts                       int      `json:"go_edit_attempts"`
	SuccessfulGoEdits                    int      `json:"successful_go_edits"`
	SuccessfulGoEditsAttemptedBeforeLoad int      `json:"successful_go_edits_attempted_before_load"`
	AttributedGateBlocks                 int      `json:"attributed_gate_blocks"`
}

func routingContract(t *testing.T, st routingStats) (*confirmedRoutingContract, []confirmedRoutingContract) {
	t.Helper()
	b, err := json.Marshal(st)
	if err != nil {
		t.Fatal(err)
	}
	var wire struct {
		Confirmed *confirmedRoutingContract  `json:"confirmed"`
		Turns     []confirmedRoutingContract `json:"turns"`
	}
	if err := json.Unmarshal(b, &wire); err != nil {
		t.Fatal(err)
	}
	return wire.Confirmed, wire.Turns
}

func skillRequest(id, msg, skill string) string {
	return `{"type":"assistant","message":{"id":"` + msg + `","content":[{"type":"tool_use","id":"` + id + `","name":"Skill","input":{"skill":"` + skill + `"}}]}}`
}

func toolResult(id, body string) string {
	return `{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"` + id + `",` + body + `}]}}`
}

func editRequest(id, tool, path string) string {
	return `{"type":"assistant","message":{"id":"edit-` + id + `","content":[{"type":"tool_use","id":"` + id + `","name":"` + tool + `","input":{"file_path":"` + path + `"}}]}}`
}

func TestClaudeConfirmedSkillResultContract(t *testing.T) {
	tests := []struct {
		name, skill, result, status string
		want                        []string
	}{
		{name: "request only", skill: "go-code", status: "partial"},
		{name: "explicit success", skill: "golang-skills:go-code", result: toolResult("s", `"is_error":false,"content":"loaded"`), status: "observed", want: []string{"go-code"}},
		{name: "optional is error", skill: "go-code", result: toolResult("s", `"content":[{"type":"text","text":"loaded"}]`), status: "observed", want: []string{"go-code"}},
		{name: "failure", skill: "go-code", result: toolResult("s", `"is_error":true,"content":"failed"`), status: "observed"},
		{name: "unknown id", skill: "go-code", result: toolResult("other", `"content":"loaded"`), status: "partial"},
		{name: "empty content", skill: "go-code", result: toolResult("s", `"content":""`), status: "partial"},
		{name: "whitespace content", skill: "go-code", result: toolResult("s", `"content":"  "`), status: "partial"},
		{name: "missing content", skill: "go-code", result: toolResult("s", `"is_error":false`), status: "partial"},
		{name: "null content", skill: "go-code", result: toolResult("s", `"content":null`), status: "partial"},
		{name: "empty blocks", skill: "go-code", result: toolResult("s", `"content":[]`), status: "partial"},
		{name: "unsupported blocks", skill: "go-code", result: toolResult("s", `"content":[{"type":"image","source":{}}]`), status: "partial"},
		{name: "invalid is error", skill: "go-code", result: toolResult("s", `"is_error":"false","content":"loaded"`), status: "partial"},
		{name: "null is error", skill: "go-code", result: toolResult("s", `"is_error":null,"content":"loaded"`), status: "partial"},
		{name: "unrelated skill", skill: "documents", result: toolResult("s", `"content":"loaded"`), status: "observed"},
		{name: "mixed unsupported blocks", skill: "go-code", result: toolResult("s", `"content":[{"type":"text","text":"loaded"},{"type":"image"}]`), status: "partial"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stream := skillRequest("s", "m", tt.skill) + "\n" + tt.result
			got, _ := routingContract(t, claudeRouting([]byte(stream)))
			if got == nil {
				t.Fatalf("claudeRouting(%s).confirmed = nil, want versioned tool-result evidence", tt.name)
			}
			if got.Version != 1 || got.ToolEvidence != tt.status || !reflect.DeepEqual(got.FirstLoads, tt.want) {
				t.Errorf("claudeRouting(%s).confirmed = %+v, want version 1, status %q, loads %v", tt.name, got, tt.status, tt.want)
			}
		})
	}
}

func TestClaudeConfirmedEditOrderingContract(t *testing.T) {
	tests := []struct {
		name, tool, path       string
		success                bool
		attempts, edits, early int
	}{
		{name: "Edit success", tool: "Edit", path: "/x/a.go", success: true, attempts: 1, edits: 1, early: 1},
		{name: "Write success", tool: "Write", path: "/x/a.go", success: true, attempts: 1, edits: 1, early: 1},
		{name: "MultiEdit success", tool: "MultiEdit", path: "/x/a.go", success: true, attempts: 1, edits: 1, early: 1},
		{name: "blocked edit", tool: "Edit", path: "/x/a.go", attempts: 1},
		{name: "non Go edit", tool: "Edit", path: "/x/a.go.md", success: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stream := []string{skillRequest("s", "m", "go-code"), editRequest("e", tt.tool, tt.path), toolResult("s", `"content":"loaded"`)}
			if tt.success {
				stream = append(stream, toolResult("e", `"content":"applied"`))
			} else {
				stream = append(stream, toolResult("e", `"is_error":true,"content":"blocked"`))
			}
			got, _ := routingContract(t, claudeRouting([]byte(strings.Join(stream, "\n"))))
			if got == nil {
				t.Fatal("claudeRouting(edit).confirmed = nil, want evidence")
			}
			if got.GoEditAttempts != tt.attempts || got.SuccessfulGoEdits != tt.edits || got.SuccessfulGoEditsAttemptedBeforeLoad != tt.early {
				t.Errorf("claudeRouting(%s).confirmed = %+v, want attempts %d, successful %d, early %d", tt.name, got, tt.attempts, tt.edits, tt.early)
			}
			if tt.attempts == 0 {
				if got.RouterBeforeFirstGoEdit != nil || got.StyleCoreBeforeFirstGoEdit != nil {
					t.Errorf("claudeRouting(%s) before-first flags = %+v, want nil", tt.name, got)
				}
				return
			}
			if got.BeforeFirstGoEditEvidence != "partial" || got.RouterBeforeFirstGoEdit != nil || got.StyleCoreBeforeFirstGoEdit != nil {
				t.Errorf("claudeRouting(%s) before-first flags = %+v, want partial while Skill result was pending", tt.name, got)
			}
		})
	}
}

func TestClaudeConfirmedGroupingAndDedupContract(t *testing.T) {
	stream := []string{
		skillRequest("late", "m-late", "go-testing"), skillRequest("router", "m-first", "go-code-review"),
		skillRequest("style", "m-first", "go-style-core"), skillRequest("router", "m-first", "go-code-review"),
		toolResult("style", `"content":"loaded"`), toolResult("router", `"content":"loaded"`), toolResult("style", `"content":"loaded"`),
		toolResult("late", `"content":"loaded"`), editRequest("e", "Edit", "a.go"), editRequest("e", "Edit", "a.go"),
		toolResult("e", `"content":"applied"`), toolResult("e", `"content":"applied"`),
	}
	got, _ := routingContract(t, claudeRouting([]byte(strings.Join(stream, "\n"))))
	if got == nil {
		t.Fatal("claudeRouting(group).confirmed = nil, want evidence")
	}
	if !reflect.DeepEqual(got.FirstLoads, []string{"go-code-review", "go-style-core"}) || !reflect.DeepEqual(got.LoadsBeforeFirstGoEdit, []string{"go-code-review", "go-style-core", "go-testing"}) || got.GoEditAttempts != 1 || got.SuccessfulGoEdits != 1 || got.SuccessfulGoEditsAttemptedBeforeLoad != 0 || got.RouterBeforeFirstGoEdit == nil || !*got.RouterBeforeFirstGoEdit || got.StyleCoreBeforeFirstGoEdit == nil || !*got.StyleCoreBeforeFirstGoEdit {
		t.Errorf("claudeRouting(interleaved duplicates).confirmed = %+v, want deduped successful first group, all loads before first edit and both routing flags", got)
	}
}

func TestClaudeMissingIDsAndAvailabilityContract(t *testing.T) {
	tests := []struct {
		name, stream, status string
		loads                []string
	}{
		{name: "empty", status: "unmeasured"},
		{name: "final only", stream: `{"type":"result","result":"done"}`, status: "unmeasured"},
		{name: "empty final", stream: `{"type":"result","result":""}`, status: "unmeasured"},
		{name: "missing final", stream: `{"type":"result"}`, status: "unmeasured"},
		{name: "text assistant", stream: `{"type":"assistant","message":{"content":[{"type":"text","text":"review"}]}}`, status: "observed"},
		{name: "missing request ID", stream: skillRequest("", "", "go-code") + "\n" + toolResult("", `"content":"loaded"`), status: "partial"},
		{name: "result before request", stream: toolResult("s", `"content":"loaded"`) + "\n" + skillRequest("s", "m", "go-code"), status: "partial"},
		{name: "missing message IDs", stream: skillRequest("s", "", "go-code") + "\n" + skillRequest("s2", "", "go-testing") + "\n" + toolResult("s", `"content":"loaded"`) + "\n" + toolResult("s2", `"content":"loaded"`), status: "partial", loads: []string{"go-code"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, _ := routingContract(t, claudeRouting([]byte(tt.stream)))
			if got == nil {
				t.Fatal("claudeRouting(availability).confirmed = nil, want explicit availability")
			}
			if got.ToolEvidence != tt.status || !reflect.DeepEqual(got.FirstLoads, tt.loads) || got.RouterBeforeFirstGoEdit != nil || got.StyleCoreBeforeFirstGoEdit != nil {
				t.Errorf("claudeRouting(%s).confirmed = %+v, want status %q loads %v no-edit N/A", tt.name, got, tt.status, tt.loads)
			}
		})
	}
}

func TestClaudeDuplicateHookFormattingContract(t *testing.T) {
	event := `{"type":"system","subtype":"hook_response","hook_event":"PreToolUse","stderr":"golang-skills routing gate: missing","exit_code":2}`
	var raw map[string]any
	if err := json.Unmarshal([]byte(event), &raw); err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := routingContract(t, claudeRouting([]byte(event+"\n"+string(b))))
	if got == nil || got.AttributedGateBlocks != 1 {
		t.Errorf("claudeRouting(equivalent hook emissions) = %+v, want 1 attributable block", got)
	}
}

func TestClaudeHookAttributionContract(t *testing.T) {
	tests := []struct {
		name, event, status string
		blocks              int
	}{
		{name: "attributable stderr", event: `{"type":"system","subtype":"hook_response","hook_event":"PreToolUse","stderr":"golang-skills routing gate: load go-code","exit_code":2,"uuid":"h1"}`, status: "observed", blocks: 1},
		{name: "attributable output", event: `{"type":"system","subtype":"hook_response","hook_name":"PreToolUse:Edit","output":"golang-skills routing gate: missing","exit_code":2}`, status: "observed", blocks: 1},
		{name: "attributable stdout", event: `{"type":"system","subtype":"hook_response","hook_event":"PreToolUse","stdout":"golang-skills routing gate: ready","exit_code":0}`, status: "observed"},
		{name: "legacy unattributed", event: `{"type":"system","subtype":"hook_response","hook_name":"PreToolUse:Edit","exit_code":2}`, status: "partial"},
		{name: "unrelated", event: `{"type":"system","subtype":"hook_response","hook_event":"PreToolUse","stderr":"other hook","exit_code":2}`, status: "unmeasured"},
		{name: "PostToolUse", event: `{"type":"system","subtype":"hook_response","hook_event":"PostToolUse","stderr":"golang-skills routing gate: missing","exit_code":2}`, status: "unmeasured"},
		{name: "missing exit", event: `{"type":"system","subtype":"hook_response","hook_event":"PreToolUse","stderr":"golang-skills routing gate: missing","outcome":"error"}`, status: "partial"},
		{name: "missing identity", event: `{"type":"system","subtype":"hook_response","stderr":"golang-skills routing gate: missing","exit_code":2}`, status: "partial"},
		{name: "not system", event: `{"type":"assistant","subtype":"hook_response","hook_event":"PreToolUse","stderr":"golang-skills routing gate: missing","exit_code":2}`, status: "unmeasured"},
		{name: "cancelled", event: `{"type":"system","subtype":"hook_response","hook_event":"PreToolUse","stderr":"golang-skills routing gate: missing","exit_code":2,"outcome":"cancelled"}`, status: "partial"},
		{name: "invalid outcome", event: `{"type":"system","subtype":"hook_response","hook_event":"PreToolUse","stderr":"golang-skills routing gate: missing","exit_code":2,"outcome":"unknown"}`, status: "partial"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, _ := routingContract(t, claudeRouting([]byte(tt.event+"\n"+tt.event)))
			if got == nil {
				t.Fatal("claudeRouting(hook).confirmed = nil, want explicit hook availability")
			}
			if got.HookEvidence != tt.status || got.AttributedGateBlocks != tt.blocks {
				t.Errorf("claudeRouting(%s twice).confirmed = %+v, want status %q blocks %d", tt.name, got, tt.status, tt.blocks)
			}
		})
	}
}

func TestClaudeReadFallbackCoverageContract(t *testing.T) {
	tests := []struct{ path, status string }{
		{path: "/skills/go-code/SKILL.md", status: "partial"},
		{path: "/skills/go-style-core/references/CURRENT-GO.md", status: "observed"},
		{path: "/source/main.go", status: "observed"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()
			stream := editRequest("read", "Read", tt.path) + "\n" + toolResult("read", `"content":"read content"`) + "\n" + editRequest("e", "Edit", "a.go") + "\n" + toolResult("e", `"content":"applied"`)
			got, _ := routingContract(t, claudeRouting([]byte(stream)))
			if got == nil || got.ToolEvidence != tt.status {
				t.Errorf("claudeRouting(Read %q).confirmed = %+v, want coverage %q", tt.path, got, tt.status)
			}
			if tt.status == "partial" && (got.BeforeFirstGoEditEvidence != "partial" || got.RouterBeforeFirstGoEdit != nil || got.StyleCoreBeforeFirstGoEdit != nil) {
				t.Errorf("claudeRouting(Read fallback %q).before-first = %+v, want partial with nil flags", tt.path, got)
			}
		})
	}
}

func TestClaudePendingSkillBeforeFirstEditContract(t *testing.T) {
	for _, body := range []string{`"content":"loaded"`, `"content":"failed","is_error":true`} {
		t.Run(body, func(t *testing.T) {
			t.Parallel()
			stream := skillRequest("s", "m", "go-style-core") + "\n" + editRequest("e", "Write", "a.go") + "\n" + toolResult("s", body) + "\n" + toolResult("e", `"content":"applied"`)
			got, _ := routingContract(t, claudeRouting([]byte(stream)))
			if got == nil || got.ToolEvidence != "observed" || got.BeforeFirstGoEditEvidence != "partial" || got.RouterBeforeFirstGoEdit != nil || got.StyleCoreBeforeFirstGoEdit != nil || got.SuccessfulGoEditsAttemptedBeforeLoad != 1 {
				t.Errorf("claudeRouting(pending Skill then %s) = %+v, want observed final tools, partial first-edit and early successful edit", body, got)
			}
		})
	}
}

func TestClaudeLegacyAttemptsRemainIndependentContract(t *testing.T) {
	stream := skillRequest("s", "m", "go-code") + "\n" + toolResult("s", `"is_error":true,"content":"failed"`) + "\n" + editRequest("e", "Edit", "a.go") + "\n" + toolResult("e", `"content":"applied"`) + "\n" + `{"type":"system","subtype":"hook_response","hook_name":"PreToolUse:Edit","exit_code":2}`
	st := claudeRouting([]byte(stream))
	got, _ := routingContract(t, st)
	if st.FirstTool != "Skill" || st.EditBeforeLoad || st.GateBlocks != 1 || got == nil || got.AttributedGateBlocks != 0 || got.SuccessfulGoEditsAttemptedBeforeLoad != 1 || got.BeforeFirstGoEditEvidence != "observed" || got.RouterBeforeFirstGoEdit == nil || *got.RouterBeforeFirstGoEdit {
		t.Errorf("claudeRouting(failed Skill plus unrelated hook) = %+v confirmed %+v, want legacy request separate from success and attribution", st, got)
	}
	bash := `{"type":"assistant","message":{"content":[{"type":"tool_use","id":"b","name":"Bash","input":{"command":"go test ./..."}}]}}`
	if claudeCommands([]byte(bash+"\n"+toolResult("b", `"is_error":"invalid","content":null`))) != 1 {
		t.Error("claudeCommands(malformed result) = non-1, want independent attempt count 1")
	}
}

func TestClaudeRoutingConsoleCoverageContract(t *testing.T) {
	var legacy routingStats
	if err := json.Unmarshal([]byte(`{"first_tool":"Skill","first_loads":["go-style-core"],"gate_blocks":2}`), &legacy); err != nil {
		t.Fatal(err)
	}
	confirmed := claudeRouting([]byte(skillRequest("s", "m", "go-code-review") + "\n" + toolResult("s", `"content":"loaded"`)))
	partial := claudeRouting([]byte(editRequest("e", "Edit", "a.go")))
	for _, corpus := range []string{corpusImplement, corpusRefactor, corpusReview} {
		t.Run(corpus, func(t *testing.T) {
			rep := report{Corpus: corpus, Arms: []arm{{Name: "baseline"}}, Results: []result{
				{Arm: "baseline", Routing: &legacy},
				{Arm: "baseline", Routing: &confirmed},
				{Arm: "baseline", Routing: &partial},
				{Arm: "baseline"},
			}}
			b := captureRoutingSummary(t, rep)
			for _, want := range []string{"routing (Skill tool-result evidence)", "before-first Go edit: n/a", "tool turns observed 1, partial 1, unmeasured 2", "hook turns observed 0, partial 0, unmeasured 4", "legacy attempts", "partial lower bounds"} {
				if !strings.Contains(string(b), want) {
					t.Errorf("printSummary(%s) = %q, want substring %q", corpus, b, want)
				}
			}
		})
	}
}

func captureRoutingSummary(t *testing.T, rep report) string {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "summary")
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stdout
	os.Stdout = file
	func() { defer func() { os.Stdout = previous }(); printSummary(rep) }()
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestRoutingConsoleCommandScopeContract(t *testing.T) {
	for _, corpus := range []string{corpusImplement, corpusReview} {
		for _, runner := range []string{runnerClaude, runnerCodex, runnerOpencode, runnerCopilot, ""} {
			t.Run(corpus+" "+runner, func(t *testing.T) {
				r := result{Arm: "baseline", Build: true, Golden: true, Edited: true, Commands: 3}
				if corpus == corpusReview {
					r.Edited = false
					r.Output = "review complete"
					r.Review = &reviewScore{Total: 1}
				}
				rep := report{Corpus: corpus, Runner: runner, Arms: []arm{{Name: "baseline"}}, Results: []result{r}}
				got := captureRoutingSummary(t, rep)
				want := "command calls: unmeasured"
				if runner == "" || runner == runnerClaude || runner == runnerCodex {
					want = "command call attempts 3.00/completed run"
				}
				if !strings.Contains(got, want) {
					t.Errorf("printSummary(%s %s) = %q, want %q", corpus, runner, got, want)
				}
				if runner == "" && !strings.Contains(got, "(runner claude; success not inferred)") {
					t.Errorf("printSummary(%s default runner) = %q, want displayed runner claude", corpus, got)
				}
			})
		}
	}
}

func TestClaudeRoutingConfirmedDenominatorsContract(t *testing.T) {
	full := claudeRouting([]byte(skillRequest("s", "m", "go-code-refactor") + "\n" + skillRequest("style", "m", "go-style-core") + "\n" + toolResult("s", `"content":"loaded"`) + "\n" + toolResult("style", `"content":"loaded"`) + "\n" + editRequest("e", "Edit", "a.go") + "\n" + toolResult("e", `"content":"applied"`) + "\n" + `{"type":"system","subtype":"hook_response","hook_event":"PreToolUse","stderr":"golang-skills routing gate: block","exit_code":2}`))
	partial := claudeRouting([]byte(skillRequest("s", "m", "go-code") + "\n" + editRequest("e", "Edit", "a.go") + "\n" + toolResult("s", `"content":"loaded"`) + "\n" + toolResult("e", `"content":"applied"`) + "\n" + `{"type":"system","subtype":"hook_response","hook_event":"PreToolUse","stderr":"golang-skills routing gate: block","exit_code":2,"uuid":"h"}` + "\n" + `{"type":"system","subtype":"hook_response","hook_event":"PreToolUse","stderr":"golang-skills routing gate: unknown"}`))
	rep := report{Arms: []arm{{Name: "baseline"}}, Results: []result{{Arm: "baseline", Routing: &full}, {Arm: "baseline", Routing: &partial}}}
	got := captureRoutingSummary(t, rep)
	for _, want := range []string{"before-first Go edit: router 1/1, go-style-core 1/1", "partial edited turns excluded 1", "attributable routing blocks 1/1 observed hook turns", "attributable blocks 1 (excluded from confirmed rates)"} {
		if !strings.Contains(got, want) {
			t.Errorf("printSummary(complete and partial) = %q, want %q", got, want)
		}
	}
}

func TestClaudeRoutingPartialEditedTurnLabelsContract(t *testing.T) {
	loaded := skillRequest("router", "loads", "go-code") + "\n" + skillRequest("style", "loads", "go-style-core") + "\n" + toolResult("router", `"content":"loaded"`) + "\n" + toolResult("style", `"content":"loaded"`) + "\n"
	edit := editRequest("e", "Edit", "a.go") + "\n" + toolResult("e", `"content":"applied"`) + "\n"
	pendingBash := `{"type":"assistant","message":{"id":"late","content":[{"type":"tool_use","id":"bash","name":"Bash","input":{"command":"go test ./..."}}]}}`
	latePartial := claudeRouting([]byte(loaded + edit + pendingBash))
	bothPartial := claudeRouting([]byte(loaded + skillRequest("pending", "owners", "go-testing") + "\n" + edit + pendingBash))
	got, _ := routingContract(t, latePartial)
	if got == nil || got.ToolEvidence != "partial" || got.BeforeFirstGoEditEvidence != "observed" || got.RouterBeforeFirstGoEdit == nil || !*got.RouterBeforeFirstGoEdit || got.StyleCoreBeforeFirstGoEdit == nil || !*got.StyleCoreBeforeFirstGoEdit || got.SuccessfulGoEdits != 1 {
		t.Fatalf("claudeRouting(confirmed loads and edit then pending Bash) = %+v, want partial turn with observed successful first-edit evidence", got)
	}
	rep := report{Arms: []arm{{Name: "baseline"}}, Results: []result{{Arm: "baseline", Routing: &latePartial}, {Arm: "baseline", Routing: &bothPartial}}}
	console := captureRoutingSummary(t, rep)
	for _, want := range []string{"tool turns observed 0, partial 2", "before-first Go edit: n/a", "partial edited turns excluded 2"} {
		if !strings.Contains(console, want) {
			t.Errorf("printSummary(late partial and both partial) = %q, want %q with each excluded turn counted once", console, want)
		}
	}
}

func TestClaudeNoToolOpeningRemainsFirstInvocationContract(t *testing.T) {
	first := claudeRouting([]byte(`{"type":"result","result":""}`))
	repair := claudeRouting([]byte(skillRequest("s", "m", "go-code") + "\n" + toolResult("s", `"content":"loaded"`)))
	var r result
	r.merge(sessionTurn{routing: &first})
	r.merge(sessionTurn{routing: &repair})
	got, turns := routingContract(t, *r.Routing)
	if got == nil || len(turns) != 2 || r.Routing.FirstTool != "" || len(got.FirstLoads) != 0 || turns[0].ToolEvidence != "unmeasured" || !reflect.DeepEqual(turns[1].FirstLoads, []string{"go-code"}) {
		t.Errorf("merge(no-tool opening then repair) = %+v turns %+v, want opening tied to first invocation", r.Routing, turns)
	}
}

func TestClaudeFreshRepairEvidenceContract(t *testing.T) {
	first := claudeRouting([]byte(skillRequest("s", "m", "go-code") + "\n" + toolResult("s", `"content":"loaded"`)))
	repair := claudeRouting([]byte(editRequest("e", "Write", "a.go") + "\n" + toolResult("e", `"content":"applied"`)))
	var r result
	r.merge(sessionTurn{routing: &first})
	r.merge(sessionTurn{routing: &repair})
	got, turns := routingContract(t, *r.Routing)
	if got == nil || len(turns) != 2 {
		t.Fatalf("merge(fresh repair) = %+v turns %v, want confirmed and 2 snapshots", got, turns)
	}
	if !reflect.DeepEqual(got.FirstLoads, []string{"go-code"}) || got.RouterBeforeFirstGoEdit != nil || got.SuccessfulGoEditsAttemptedBeforeLoad != 1 || turns[1].RouterBeforeFirstGoEdit == nil || *turns[1].RouterBeforeFirstGoEdit {
		t.Errorf("merge(fresh repair) = %+v turns %+v, want opening preserved and repair edit before fresh load", got, turns)
	}
}

func TestClaudeRoutingFakeCLIContract(t *testing.T) {
	bin := t.TempDir()
	work := t.TempDir()
	stream := skillRequest("s", "m", "go-style-core") + "\n" + toolResult("s", `"content":"loaded"`) + "\n" + `{"type":"result","result":"done"}` + "\n"
	path := filepath.Join(bin, "transcript.jsonl")
	if err := os.WriteFile(path, []byte(stream), 0o600); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(bin, "claude"), "#!/bin/sh\ncat \"$CLAUDE_ROUTING_TEST_TRANSCRIPT\"\n")
	if err := os.Chmod(filepath.Join(bin, "claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("CLAUDE_ROUTING_TEST_TRANSCRIPT", path)
	got := runSession(options{timeout: time.Second, corpus: corpusImplement}, arm{}, work, "safe stub")
	if got.err != nil || got.routing == nil {
		t.Fatalf("runSession(fake CLI) = %+v, want routing without error", got)
	}
	evidence, _ := routingContract(t, *got.routing)
	if evidence == nil || evidence.ToolEvidence != "observed" || !reflect.DeepEqual(evidence.FirstLoads, []string{"go-style-core"}) {
		t.Errorf("runSession(fake CLI).routing = %+v, want confirmed tool-result evidence", evidence)
	}
}
