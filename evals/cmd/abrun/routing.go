package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

// Перші чотири поля зберігають legacy-визначення запитів, включно з
// unattributed hook blocks; Confirmed відокремлює докази tool results.
type routingStats struct {
	FirstTool      string            `json:"first_tool,omitempty"`
	FirstLoads     []string          `json:"first_loads,omitempty"`
	EditBeforeLoad bool              `json:"edit_before_load,omitempty"`
	GateBlocks     int               `json:"gate_blocks,omitempty"`
	Confirmed      *routingEvidence  `json:"confirmed,omitempty"`
	Turns          []routingEvidence `json:"turns,omitempty"`
}

// Успіх tool_result не доводить повного model-visible вмісту чи retention.
// Before-first поля належать першому CLI turn, лічильники складаються;
// Turns зберігає незалежний контекст кожного нового CLI invocation.
type routingEvidence struct {
	Version                              int      `json:"version"`
	ToolEvidence                         string   `json:"tool_evidence"`
	HookEvidence                         string   `json:"hook_evidence"`
	BeforeFirstGoEditEvidence            string   `json:"before_first_go_edit_evidence"`
	FirstLoads                           []string `json:"first_loads,omitempty"`
	LoadsBeforeFirstGoEdit               []string `json:"loads_before_first_go_edit,omitempty"`
	RouterBeforeFirstGoEdit              *bool    `json:"router_before_first_go_edit"`
	StyleCoreBeforeFirstGoEdit           *bool    `json:"style_core_before_first_go_edit"`
	GoEditAttempts                       int      `json:"go_edit_attempts"`
	SuccessfulGoEdits                    int      `json:"successful_go_edits"`
	SuccessfulGoEditsAttemptedBeforeLoad int      `json:"successful_go_edits_attempted_before_load"`
	AttributedGateBlocks                 int      `json:"attributed_gate_blocks"`
}

func claudeRouting(out []byte) routingStats {
	st := legacyClaudeRouting(out)
	evidence := routingEvidence{Version: 1, ToolEvidence: "unmeasured", HookEvidence: "unmeasured"}
	type request struct {
		skill, group       string
		goEdit, beforeLoad bool
		resolved           bool
	}
	requests := map[string]request{}
	seenRequests := map[string]bool{}
	seenResults := map[string]bool{}
	seenHooks := map[string]bool{}
	loaded := map[string]bool{}
	firstGroup := ""
	firstLoads := map[string]bool{}
	for line := range strings.SplitSeq(string(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var e struct {
			Type      string          `json:"type"`
			Subtype   string          `json:"subtype"`
			HookName  string          `json:"hook_name"`
			HookEvent string          `json:"hook_event"`
			HookID    string          `json:"hook_id"`
			SessionID string          `json:"session_id"`
			UUID      string          `json:"uuid"`
			Output    string          `json:"output"`
			Stdout    string          `json:"stdout"`
			Stderr    string          `json:"stderr"`
			ExitCode  json.RawMessage `json:"exit_code"`
			Outcome   json.RawMessage `json:"outcome"`
			Message   struct {
				ID      string `json:"id"`
				Content []struct {
					Type      string          `json:"type"`
					ID        string          `json:"id"`
					Name      string          `json:"name"`
					ToolUseID string          `json:"tool_use_id"`
					IsError   json.RawMessage `json:"is_error"`
					Content   json.RawMessage `json:"content"`
					Input     struct {
						Skill    string `json:"skill"`
						FilePath string `json:"file_path"`
					} `json:"input"`
				} `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal([]byte(line), &e) != nil {
			evidence.ToolEvidence = "partial"
			evidence.HookEvidence = "partial"
			continue
		}
		if e.Type == "system" && e.Subtype == "hook_response" {
			canonical, _ := json.Marshal(e) // JSON-сумісні поля; порядок ключів не змінює emission.
			key := "event:" + string(canonical)
			if e.UUID != "" {
				key = "uuid:" + e.UUID
			}
			if seenHooks[key] {
				continue
			}
			seenHooks[key] = true
			preTool := e.HookEvent == "PreToolUse" || (e.HookEvent == "" && (e.HookName == "PreToolUse" || strings.HasPrefix(e.HookName, "PreToolUse:")))
			marker := strings.Contains(e.Output, "golang-skills routing gate:") || strings.Contains(e.Stdout, "golang-skills routing gate:") || strings.Contains(e.Stderr, "golang-skills routing gate:")
			if !preTool || !marker {
				if (preTool && e.Output == "" && e.Stdout == "" && e.Stderr == "") || (marker && e.HookEvent == "" && e.HookName == "") {
					evidence.HookEvidence = "partial"
				}
				continue
			}
			var exit int
			var outcome string
			if len(e.ExitCode) == 0 || string(e.ExitCode) == "null" || json.Unmarshal(e.ExitCode, &exit) != nil || (exit != 0 && exit != 2) || (len(e.Outcome) > 0 && (json.Unmarshal(e.Outcome, &outcome) != nil || (outcome != "success" && outcome != "error"))) {
				evidence.HookEvidence = "partial"
				continue
			}
			if evidence.HookEvidence == "unmeasured" {
				evidence.HookEvidence = "observed"
			}
			if exit == 2 {
				evidence.AttributedGateBlocks++
			}
			continue
		}
		if e.Type == "assistant" && e.Message.Content != nil {
			if evidence.ToolEvidence == "unmeasured" {
				evidence.ToolEvidence = "observed"
			}
			for _, c := range e.Message.Content {
				if c.Type != "tool_use" {
					continue
				}
				key := "id:" + c.ID
				if c.ID == "" {
					evidence.ToolEvidence = "partial"
					block, _ := json.Marshal(c) // Тип складається лише з JSON-сумісних полів.
					key = "missing:" + e.Message.ID + ":" + string(block)
				}
				if seenRequests[key] {
					continue
				}
				seenRequests[key] = true
				r := request{skill: normalizeSkill(c.Input.Skill), group: "message:" + e.Message.ID}
				if c.Name != "Skill" {
					r.skill = ""
				}
				// Read fallback потребує окремого byte/content decoder у G6.3.
				if c.Name == "Read" && filepath.Base(c.Input.FilePath) == "SKILL.md" && normalizeSkill(filepath.Base(filepath.Dir(c.Input.FilePath))) != "" {
					evidence.ToolEvidence = "partial"
				}
				if r.skill != "" && e.Message.ID == "" {
					r.group = "request:" + c.ID
					evidence.ToolEvidence = "partial"
				}
				r.goEdit = (c.Name == "Edit" || c.Name == "Write" || c.Name == "MultiEdit") && strings.HasSuffix(c.Input.FilePath, ".go")
				if r.goEdit {
					r.beforeLoad = len(loaded) == 0
					if evidence.GoEditAttempts == 0 {
						for name := range loaded {
							evidence.LoadsBeforeFirstGoEdit = append(evidence.LoadsBeforeFirstGoEdit, name)
						}
						slices.Sort(evidence.LoadsBeforeFirstGoEdit)
						evidence.BeforeFirstGoEditEvidence = "observed"
						if evidence.ToolEvidence == "partial" {
							evidence.BeforeFirstGoEditEvidence = "partial"
						}
						for _, pending := range requests {
							if pending.skill != "" && !pending.resolved {
								evidence.BeforeFirstGoEditEvidence = "partial"
							}
						}
						if evidence.BeforeFirstGoEditEvidence == "observed" {
							evidence.RouterBeforeFirstGoEdit = new(loaded["go-code"] || loaded["go-code-refactor"] || loaded["go-code-review"])
							evidence.StyleCoreBeforeFirstGoEdit = new(loaded["go-style-core"])
						}
					}
					evidence.GoEditAttempts++
				}
				if c.ID != "" {
					requests[c.ID] = r
				}
			}
			continue
		}
		if e.Type != "user" {
			continue
		}
		for _, c := range e.Message.Content {
			if c.Type != "tool_result" {
				continue
			}
			// Відомий завершений ID не дозволяє повторному emission змінити успіх.
			if seenResults[c.ToolUseID] {
				continue
			}
			seenResults[c.ToolUseID] = true
			r, ok := requests[c.ToolUseID]
			if !ok || c.ToolUseID == "" {
				evidence.ToolEvidence = "partial"
				continue
			}
			success, valid := successfulClaudeToolResult(c.IsError, c.Content)
			if !valid {
				evidence.ToolEvidence = "partial"
				continue
			}
			r.resolved = true
			requests[c.ToolUseID] = r
			if !success {
				continue
			}
			if r.skill != "" {
				loaded[r.skill] = true
				if firstGroup == "" {
					firstGroup = r.group
				}
				if r.group == firstGroup {
					firstLoads[r.skill] = true
				}
			}
			if r.goEdit {
				evidence.SuccessfulGoEdits++
				if r.beforeLoad {
					evidence.SuccessfulGoEditsAttemptedBeforeLoad++
				}
			}
		}
	}
	for _, r := range requests {
		if !r.resolved {
			evidence.ToolEvidence = "partial"
		}
	}
	for name := range firstLoads {
		evidence.FirstLoads = append(evidence.FirstLoads, name)
	}
	slices.Sort(evidence.FirstLoads)
	if evidence.GoEditAttempts == 0 {
		evidence.BeforeFirstGoEditEvidence = "unmeasured"
		if evidence.ToolEvidence == "observed" {
			evidence.BeforeFirstGoEditEvidence = "not_applicable"
		}
	}
	st.Confirmed = &evidence
	return st
}

// Декодування профілю result окреме від attempt-based метрик і лічильників.
func successfulClaudeToolResult(isError, content json.RawMessage) (success, valid bool) {
	var failed bool
	if len(isError) > 0 && (string(isError) == "null" || json.Unmarshal(isError, &failed) != nil) {
		return false, false
	}
	var text string
	if json.Unmarshal(content, &text) == nil && strings.TrimSpace(text) != "" {
		return !failed, true
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(content, &blocks) != nil || len(blocks) == 0 {
		return false, false
	}
	for _, b := range blocks {
		if b.Type != "text" || strings.TrimSpace(b.Text) == "" {
			return false, false
		}
	}
	return !failed, true
}

func legacyClaudeRouting(out []byte) routingStats {
	var st routingStats
	firstMsg := ""
	loaded := false
	for line := range strings.SplitSeq(string(out), "\n") {
		var e struct {
			Type     string `json:"type"`
			Subtype  string `json:"subtype"`
			HookName string `json:"hook_name"`
			ExitCode int    `json:"exit_code"`
			Message  struct {
				ID      string `json:"id"`
				Content []struct {
					Type  string `json:"type"`
					Name  string `json:"name"`
					Input struct {
						Skill string `json:"skill"`
					} `json:"input"`
				} `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal([]byte(line), &e) != nil {
			continue
		}
		if e.Type == "system" && e.Subtype == "hook_response" && strings.HasPrefix(e.HookName, "PreToolUse") && e.ExitCode == 2 {
			st.GateBlocks++
			continue
		}
		if e.Type != "assistant" {
			continue
		}
		for _, c := range e.Message.Content {
			if c.Type != "tool_use" {
				continue
			}
			if st.FirstTool == "" {
				st.FirstTool = c.Name
			}
			switch c.Name {
			case "Skill":
				loaded = true
				if firstMsg == "" {
					firstMsg = e.Message.ID
				}
				if e.Message.ID == firstMsg {
					if name := normalizeSkill(c.Input.Skill); name != "" {
						st.FirstLoads = append(st.FirstLoads, name)
					}
				}
			case "Edit", "Write", "MultiEdit":
				if !loaded {
					st.EditBeforeLoad = true
				}
			}
		}
	}
	return st
}

// Opening лишається за першим CLI invocation, навіть якщо він не мав tools;
// новий CLI turn не успадковує loads. Legacy per-turn визначення не змінені.
func (st *routingStats) add(t routingStats) {
	if st.Confirmed == nil && len(st.Turns) == 0 && st.FirstTool == "" && len(st.FirstLoads) == 0 {
		blocks := st.GateBlocks
		*st = t
		st.GateBlocks += blocks
		if t.Confirmed != nil {
			opening := *t.Confirmed
			st.Confirmed = &opening
			st.Turns = []routingEvidence{opening}
		}
		return
	}
	st.GateBlocks += t.GateBlocks
	if t.Confirmed == nil {
		return
	}
	st.Turns = append(st.Turns, *t.Confirmed)
	if st.Confirmed == nil {
		return
	}
	st.Confirmed.GoEditAttempts += t.Confirmed.GoEditAttempts
	st.Confirmed.SuccessfulGoEdits += t.Confirmed.SuccessfulGoEdits
	st.Confirmed.SuccessfulGoEditsAttemptedBeforeLoad += t.Confirmed.SuccessfulGoEditsAttemptedBeforeLoad
	st.Confirmed.AttributedGateBlocks += t.Confirmed.AttributedGateBlocks
	if st.Confirmed.ToolEvidence != t.Confirmed.ToolEvidence {
		st.Confirmed.ToolEvidence = "partial"
	}
	if st.Confirmed.HookEvidence != t.Confirmed.HookEvidence {
		st.Confirmed.HookEvidence = "partial"
	}
}

// Спільні знаменники для implement/refactor і review; partial — нижня межа.
func printRoutingSummary(rep report, armName string) {
	toolStatus := map[string]int{}
	hookStatus := map[string]int{}
	var legacyRuns, firstSkill, legacyStyle, legacyEarly, legacyBlocks int
	var openings, firstStyle, editedTurns, router, style, noEdit, beforeFirstPartial int
	var attempts, successful, early, blocks, hookTurns int
	var partialAttempts, partialSuccessful, partialEarly, partialBlocks int
	var completed, commands, refused int
	for _, r := range rep.Results {
		if r.Arm != armName {
			continue
		}
		if r.Refused {
			refused++
		}
		if r.Err == "" {
			completed++
			commands += r.Commands
		}
		if r.Routing == nil {
			toolStatus["unmeasured"]++
			hookStatus["unmeasured"]++
			continue
		}
		st := r.Routing
		legacyRuns++
		if st.FirstTool == "Skill" {
			firstSkill++
		}
		if slices.Contains(st.FirstLoads, "go-style-core") {
			legacyStyle++
		}
		if st.EditBeforeLoad {
			legacyEarly++
		}
		legacyBlocks += st.GateBlocks
		turns := st.Turns
		if len(turns) == 0 && st.Confirmed != nil {
			turns = []routingEvidence{*st.Confirmed}
		}
		if len(turns) == 0 {
			toolStatus["unmeasured"]++
			hookStatus["unmeasured"]++
			continue
		}
		if opening := turns[0]; opening.Version == 1 && opening.ToolEvidence == "observed" {
			openings++
			if slices.Contains(opening.FirstLoads, "go-style-core") {
				firstStyle++
			}
		}
		for _, turn := range turns {
			if turn.Version != 1 {
				toolStatus["unmeasured"]++
				hookStatus["unmeasured"]++
				continue
			}
			if turn.GoEditAttempts > 0 && (turn.ToolEvidence == "partial" || turn.BeforeFirstGoEditEvidence == "partial") {
				beforeFirstPartial++
			}
			switch turn.ToolEvidence {
			case "observed":
				toolStatus["observed"]++
				attempts += turn.GoEditAttempts
				successful += turn.SuccessfulGoEdits
				early += turn.SuccessfulGoEditsAttemptedBeforeLoad
				if turn.GoEditAttempts == 0 {
					noEdit++
				} else if turn.BeforeFirstGoEditEvidence == "observed" && turn.RouterBeforeFirstGoEdit != nil && turn.StyleCoreBeforeFirstGoEdit != nil {
					editedTurns++
					if *turn.RouterBeforeFirstGoEdit {
						router++
					}
					if *turn.StyleCoreBeforeFirstGoEdit {
						style++
					}
				}
			case "partial":
				toolStatus["partial"]++
				partialAttempts += turn.GoEditAttempts
				partialSuccessful += turn.SuccessfulGoEdits
				partialEarly += turn.SuccessfulGoEditsAttemptedBeforeLoad
			default:
				toolStatus["unmeasured"]++
			}
			switch turn.HookEvidence {
			case "observed":
				hookStatus["observed"]++
				hookTurns++
				blocks += turn.AttributedGateBlocks
			case "partial":
				hookStatus["partial"]++
				partialBlocks += turn.AttributedGateBlocks
			default:
				hookStatus["unmeasured"]++
			}
		}
	}
	fmt.Printf("%-24s   routing (Skill tool-result evidence): tool turns observed %d, partial %d, unmeasured %d; hook turns observed %d, partial %d, unmeasured %d\n", "", toolStatus["observed"], toolStatus["partial"], toolStatus["unmeasured"], hookStatus["observed"], hookStatus["partial"], hookStatus["unmeasured"])
	if openings > 0 {
		fmt.Printf("%-24s   go-style-core in first confirmed batch %d/%d opening turns\n", "", firstStyle, openings)
	}
	if editedTurns == 0 {
		fmt.Printf("%-24s   before-first Go edit: n/a; observed no-edit turns %d; partial edited turns excluded %d\n", "", noEdit, beforeFirstPartial)
	} else {
		fmt.Printf("%-24s   before-first Go edit: router %d/%d, go-style-core %d/%d edited turns; observed no-edit turns %d; partial edited turns excluded %d\n", "", router, editedTurns, style, editedTurns, noEdit, beforeFirstPartial)
	}
	if toolStatus["observed"] > 0 {
		fmt.Printf("%-24s   Go tool edits: attempts %d, successful results %d, successful results attempted before confirmed load %d\n", "", attempts, successful, early)
	}
	if hookTurns > 0 {
		fmt.Printf("%-24s   attributable routing blocks %d/%d observed hook turns\n", "", blocks, hookTurns)
	}
	if toolStatus["partial"] > 0 || hookStatus["partial"] > 0 {
		fmt.Printf("%-24s   partial lower bounds: Go attempts %d, successful results %d, successful results attempted before confirmed load %d; attributable blocks %d (excluded from confirmed rates)\n", "", partialAttempts, partialSuccessful, partialEarly, partialBlocks)
	}
	if legacyRuns > 0 {
		fmt.Printf("%-24s   legacy attempts: first tool Skill %d/%d, go-style-core in first requested batch %d/%d, edit before any Skill request %d/%d, unattributed PreToolUse exit2 %d\n", "", firstSkill, legacyRuns, legacyStyle, legacyRuns, legacyEarly, legacyRuns, legacyBlocks)
	}
	runner := rep.Runner
	if runner == "" {
		runner = runnerClaude
	}
	if (runner == runnerClaude || runner == runnerCodex) && completed > 0 {
		fmt.Printf("%-24s   command call attempts %.2f/completed run (runner %s; success not inferred)\n", "", float64(commands)/float64(completed), runner)
	} else {
		fmt.Printf("%-24s   command calls: unmeasured (runner %q)\n", "", runner)
	}
	if refused > 0 {
		fmt.Printf("%-24s   %d run(s) ended in a safety refusal (stop_reason refusal); read their traces before scoring them against the skills\n", "", refused)
	}
}
