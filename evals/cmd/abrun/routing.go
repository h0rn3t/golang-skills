package main

import (
	"encoding/json"
	"strings"
)

// routingStats is what a claude stream-json transcript says about how a
// session reached its skills: the first tool it called, the go-* skills in its
// first Skill-bearing message, whether it edited before loading any skill, and
// how many edits the routing gate blocked. The 2026-09-30 prompt-note runs
// were read from exactly these counts — a note wording changed which tool a
// session opened with and what its first load carried long before it moved
// golden or cost.
type routingStats struct {
	FirstTool      string   `json:"first_tool,omitempty"`
	FirstLoads     []string `json:"first_loads,omitempty"`
	EditBeforeLoad bool     `json:"edit_before_load,omitempty"`
	GateBlocks     int      `json:"gate_blocks,omitempty"`
}

// claudeRouting reads routingStats from one turn's stream-json. The stream has
// one assistant line per content block, so the first load is every Skill call
// sharing the message id of the first one; a gate block is a PreToolUse hook
// response with exit code 2.
func claudeRouting(out []byte) routingStats {
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

// add folds a later turn into the run's stats: the opening belongs to the
// first turn, blocks accumulate.
func (st *routingStats) add(t routingStats) {
	if st.FirstTool == "" && len(st.FirstLoads) == 0 {
		blocks := st.GateBlocks
		*st = t
		st.GateBlocks += blocks
		return
	}
	st.GateBlocks += t.GateBlocks
}
