package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
)

// stepStats is what a claude session did, counted from its stream-json trace.
// Event counts stay comparable between separate runs, where dollars drift.
// Other runners leave it nil.
type stepStats struct {
	// Turns counts the model's responses: distinct assistant message ids.
	Turns int `json:"turns"`
	// Tools counts tool calls by tool name. A call a hook blocked is counted:
	// it cost the step.
	Tools map[string]int `json:"tools,omitempty"`
	// BeforeFirstGoEdit is how many tool calls came before the first Edit,
	// Write, or MultiEdit of a .go file was attempted, skill loads included;
	// nil when the run attempted none.
	BeforeFirstGoEdit *int `json:"before_first_go_edit,omitempty"`
	// OutputTokens includes ThinkingTokens. Both are totals over every model
	// the session used, and absent when it ended without a result event.
	OutputTokens   int `json:"output_tokens,omitzero"`
	ThinkingTokens int `json:"thinking_tokens,omitzero"`
}

// claudeSteps counts the steps in a claude stream-json transcript, or returns
// nil when it holds no assistant message.
func claudeSteps(out []byte) *stepStats {
	st := stepStats{Tools: map[string]int{}}
	messages := map[string]bool{}
	calls := 0
	for line := range strings.SplitSeq(string(out), "\n") {
		var e struct {
			Type    string `json:"type"`
			Message struct {
				ID      string `json:"id"`
				Content []struct {
					Type  string `json:"type"`
					Name  string `json:"name"`
					Input struct {
						FilePath string `json:"file_path"`
					} `json:"input"`
				} `json:"content"`
			} `json:"message"`
			// The result event's own usage reads zero when the work ran in a
			// fork (a slash command); modelUsage has the totals either way.
			ModelUsage map[string]struct {
				OutputTokens   int `json:"outputTokens"`
				ThinkingTokens int `json:"thinkingTokens"`
			} `json:"modelUsage"`
		}
		if json.Unmarshal([]byte(line), &e) != nil {
			continue
		}
		switch e.Type {
		case "result":
			for _, u := range e.ModelUsage {
				st.OutputTokens += u.OutputTokens
				st.ThinkingTokens += u.ThinkingTokens
			}
		case "assistant":
			messages[e.Message.ID] = true
			for _, c := range e.Message.Content {
				if c.Type != "tool_use" {
					continue
				}
				goEdit := (c.Name == "Edit" || c.Name == "Write" || c.Name == "MultiEdit") && strings.HasSuffix(c.Input.FilePath, ".go")
				if goEdit && st.BeforeFirstGoEdit == nil {
					st.BeforeFirstGoEdit = new(calls)
				}
				st.Tools[c.Name]++
				calls++
			}
		}
	}
	if len(messages) == 0 {
		return nil
	}
	st.Turns = len(messages)
	return &st
}

// add folds a later turn of the same run into s. A first Go edit made in that
// turn comes after every call of the turns before it.
func (s *stepStats) add(t stepStats) {
	if s.BeforeFirstGoEdit == nil && t.BeforeFirstGoEdit != nil {
		earlier := 0
		for _, n := range s.Tools {
			earlier += n
		}
		s.BeforeFirstGoEdit = new(earlier + *t.BeforeFirstGoEdit)
	}
	s.Turns += t.Turns
	s.OutputTokens += t.OutputTokens
	s.ThinkingTokens += t.ThinkingTokens
	if s.Tools == nil {
		s.Tools = map[string]int{}
	}
	for name, n := range t.Tools {
		s.Tools[name] += n
	}
}

// stepsLine renders an arm's per-run step means over its completed runs that
// carry steps. The edit and token clauses count only the runs with that
// reading, and are left out when no run has it.
func stepsLine(rep report, armName string) string {
	var runs, turns, total, edited, beforeEdit, metered, output, thinking int
	tools := map[string]int{}
	for _, r := range rep.Results {
		if r.Arm != armName || r.Err != "" || r.Steps == nil {
			continue
		}
		runs++
		turns += r.Steps.Turns
		for name, n := range r.Steps.Tools {
			tools[name] += n
			total += n
		}
		if r.Steps.BeforeFirstGoEdit != nil {
			edited++
			beforeEdit += *r.Steps.BeforeFirstGoEdit
		}
		if r.Steps.OutputTokens > 0 {
			metered++
			output += r.Steps.OutputTokens
			thinking += r.Steps.ThinkingTokens
		}
	}
	if runs == 0 {
		return "steps/run: — (not measured in this report)"
	}
	mean := func(sum, n int) float64 { return float64(sum) / float64(n) }
	names := slices.SortedFunc(maps.Keys(tools), func(a, b string) int {
		return cmp.Or(cmp.Compare(tools[b], tools[a]), cmp.Compare(a, b))
	})
	perTool := make([]string, len(names))
	for i, name := range names {
		perTool[i] = fmt.Sprintf("%s %.1f", name, mean(tools[name], runs))
	}
	line := fmt.Sprintf("steps/run: %.1f tool calls (%s), %.1f turns over %d run(s)",
		mean(total, runs), strings.Join(perTool, ", "), mean(turns, runs), runs)
	if edited > 0 {
		line += fmt.Sprintf("; %.1f calls before the first Go edit (%d run(s))", mean(beforeEdit, edited), edited)
	}
	if metered > 0 {
		line += fmt.Sprintf("; output %.0f tokens, %.0f of them thinking (%d run(s))", mean(output, metered), mean(thinking, metered), metered)
	}
	return line
}
