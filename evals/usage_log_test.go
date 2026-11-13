package evals_test

import (
	"encoding/json"
	"errors"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// usageEvent is a line of the usage log the hooks append to
// <CLAUDE_PLUGIN_DATA>/usage.jsonl (hooks/helper/usage.go), without its
// timestamp.
type usageEvent struct {
	Version string            `json:"v"`
	Session string            `json:"session"`
	Agent   bool              `json:"agent"`
	Event   string            `json:"event"`
	Skill   string            `json:"skill"`
	Via     string            `json:"via"`
	Attempt int               `json:"attempt"`
	Missing []string          `json:"missing"`
	Checks  map[string]string `json:"checks"`
	Linters map[string]int    `json:"linters"`
}

// readUsageLog returns the events under state, or nil when the hooks wrote
// no log.
func readUsageLog(t *testing.T, state string) []usageEvent {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(state, "usage.jsonl"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatalf("read usage log: %v", err)
	}
	var events []usageEvent
	for line := range strings.Lines(string(raw)) {
		var e usageEvent
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatalf("usage log line %q: %v", line, err)
		}
		events = append(events, e)
	}
	return events
}

// scratchMain writes a module holding one clean main.go and returns the
// file's path.
func scratchMain(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	for name, src := range map[string]string{
		filepath.Join(dir, "go.mod"): "module scratch\n\ngo 1.27\n",
		path:                         "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hi\")\n}\n",
	} {
		if err := os.WriteFile(name, []byte(src), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return path
}

// TestUsageLog drives the hooks as a session does and reads the log they
// leave: the routing and check state is pruned after two days, so this file
// is the only record of which skills sessions load, which the gate has to
// name, and what the edit hook finds.
func TestUsageLog(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	gate := filepath.Join(root, "hooks", "go-code-routing.sh")
	editHook := filepath.Join(root, "hooks", "go-vet-on-edit.sh")
	var manifest struct{ Version string }
	raw, err := os.ReadFile(filepath.Join(root, ".claude-plugin", "plugin.json"))
	if err != nil || json.Unmarshal(raw, &manifest) != nil || manifest.Version == "" {
		t.Fatalf("plugin.json version: %q, read error %v", manifest.Version, err)
	}
	version := manifest.Version
	load := func(session, skill string) map[string]any {
		return routingPayload("PostToolUse", session, "Skill", map[string]any{"skill": skill})
	}

	t.Run("a gate block and each load route", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		edit := routingPayload("PreToolUse", "u1", "Write", map[string]any{"file_path": "/repo/api/plain.go", "content": "package api\n"})
		if code, msg := hookEvent(t, gate, state, edit); code != 2 {
			t.Fatalf("edit without a router: exit %d, stderr %q; want 2", code, msg)
		}
		hookEvent(t, gate, state, load("u1", "golang-skills:go-code"))
		hookEvent(t, gate, state, routingPayload("PostToolUse", "u1", "Read",
			map[string]any{"file_path": filepath.Join(root, "skills", "go-style-core", "SKILL.md")}))
		inSubagent := load("u1", "golang-skills:go-testing")
		inSubagent["agent_id"], inSubagent["agent_type"] = "a1", "general-purpose"
		hookEvent(t, gate, state, inSubagent)
		hookEvent(t, gate, state, load("u1", "other-plugin:deploy"))
		hookEvent(t, gate, state, load("u1", "other-plugin:go-deploy"))

		want := []usageEvent{
			{Version: version, Session: "u1", Event: "block", Attempt: 1, Missing: []string{"go-code", "go-style-core"}},
			{Version: version, Session: "u1", Event: "load", Skill: "go-code", Via: "Skill"},
			{Version: version, Session: "u1", Event: "load", Skill: "go-style-core", Via: "Read"},
			{Version: version, Session: "u1", Agent: true, Event: "load", Skill: "go-testing", Via: "Skill"},
		}
		if got := readUsageLog(t, state); !reflect.DeepEqual(got, want) {
			t.Errorf("usage log after a block, three loads of this plugin's skills, and two of another plugin's:\n got %+v\nwant %+v", got, want)
		}
	})

	t.Run("a slash command", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		if code, _ := promptEvent(t, state, "u2", t.TempDir(), "/golang-skills:go-code-refactor tidy ./dispatch"); code != 0 {
			t.Fatalf("slash invocation: exit %d, want 0", code)
		}
		want := []usageEvent{{Version: version, Session: "u2", Event: "load", Skill: "go-code-refactor", Via: "slash"}}
		if got := readUsageLog(t, state); !reflect.DeepEqual(got, want) {
			t.Errorf("usage log after /go-code-refactor:\n got %+v\nwant %+v", got, want)
		}
	})

	// The switch stops the log, not the record the gate reads.
	t.Run("switched off", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		hookOutput(t, gate, state, load("u3", "golang-skills:go-code"), "GOLANG_SKILLS_USAGE_LOG=off")
		if got := readUsageLog(t, state); got != nil {
			t.Errorf("usage log with GOLANG_SKILLS_USAGE_LOG=off = %+v, want none", got)
		}
		loaded, err := os.ReadFile(filepath.Join(state, "routing", "u3", "loaded"))
		if err != nil || string(loaded) != "go-code\n" {
			t.Errorf("gate record with the log off = %q, %v; want the go-code load", loaded, err)
		}
	})

	t.Run("an edit-hook run", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		edited := routingPayload("PostToolUse", "u4", "Write", map[string]any{"file_path": scratchMain(t)})
		if code, msg := hookEvent(t, editHook, state, edited); code != 0 {
			t.Fatalf("edit hook on a clean file: exit %d, stderr %q; want 0", code, msg)
		}
		got := readUsageLog(t, state)
		if len(got) != 1 || got[0].Event != "check" || got[0].Session != "u4" || got[0].Version != version {
			t.Fatalf("usage log after one edit-hook run = %+v, want one check event of session u4 at %s", got, version)
		}
		checks := got[0].Checks
		if len(checks) != 5 || checks["gofmt"] != "pass" || checks["vet"] != "pass" {
			t.Errorf("check event statuses = %v, want all five checks with gofmt and vet passing", checks)
		}
	})

	// revive's line-length-limit reports a whole line, so its finding has no
	// column: the hook, the lint receipt, and the log still take it as one.
	t.Run("a lint finding without a column", func(t *testing.T) {
		t.Parallel()
		state, path, bin := t.TempDir(), scratchMain(t), t.TempDir()
		finding := path + ":6: line-length-limit: line is 129 characters, out of limit 120 (revive)"
		stub := "#!/usr/bin/env bash\nif [[ \"$1\" == config ]]; then exit 1; fi\nprintf '%s\\n' \"$LINT_OUTPUT\"\nexit 1\n"
		if err := os.WriteFile(filepath.Join(bin, "golangci-lint"), []byte(stub), 0o755); err != nil { //nolint:gosec // G306: the hook runs the stub from PATH, so it is executable
			t.Fatalf("write golangci-lint stub: %v", err)
		}
		edited := routingPayload("PostToolUse", "u5", "Write", map[string]any{"file_path": path})
		code, msg := hookEventEnv(t, editHook, state, edited, "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"), "LINT_OUTPUT="+finding)
		if code != 2 || !strings.Contains(msg, finding) || strings.Contains(msg, "failed (exit 1)") {
			t.Errorf("edit hook on a finding without a column: exit %d, stderr %q; want 2 with the finding and no lint failure", code, msg)
		}
		got := readUsageLog(t, state)
		if len(got) != 1 || got[0].Checks["lint"] != "fail" || !maps.Equal(got[0].Linters, map[string]int{"revive/line-length-limit": 1}) {
			t.Errorf("usage log after that run = %+v, want one check event with lint failed on revive/line-length-limit", got)
		}
	})
}
