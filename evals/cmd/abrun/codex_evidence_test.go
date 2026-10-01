package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestParseCodexStream(t *testing.T) {
	t.Parallel()
	const path = "/tmp/h/.codex/skills/go-code-refactor/SKILL.md"
	const body = "---\nname: go-code-refactor\n---\nRefactor this package. See go-http/SKILL.md for servers.\n"
	event := func(kind, id, command, status string, exit any, output string) string {
		data, err := json.Marshal(map[string]any{"type": kind, "item": map[string]any{
			"id": id, "type": "command_execution", "command": command, "status": status, "exit_code": exit, "aggregated_output": output,
		}})
		if err != nil {
			t.Fatal(err)
		}
		return string(data) + "\n"
	}
	full := event("item.completed", "i1", "cat "+path, "completed", 0, body)
	transcript := "Reading input...\n{malformed\n" + event("item.started", "i1", "cat "+path, "in_progress", nil, "") + full + full +
		`{"type":"item.completed","item":{"type":"agent_message","text":"first pass"}}` + "\n" +
		`{"type":"item.started","item":{"type":"agent_message","text":"unfinished"}}` + "\n" +
		`{"type":"item.completed","item":{"type":"agent_message","text":"final answer"}}` + "\n" +
		`{"type":"item.completed","item":{"type":"agent_`
	skills, final, cost, evidence := parseCodexStream([]byte(transcript), map[string]string{path: body})
	if !slices.Equal(skills, []string{"go-code-refactor"}) {
		t.Errorf("parseCodexStream(transcript) skills = %v, want go-code-refactor only", skills)
	}
	if final != "final answer" || cost != 0 {
		t.Errorf("parseCodexStream(transcript) final/cost = %q/%v, want final answer/0", final, cost)
	}
	if len(evidence) != 1 {
		t.Fatalf("parseCodexStream(started + duplicate completion) evidence = %+v, want one record", evidence)
	}
	if got := evidence[0]; got.Status != "full_returned_output" || got.MatchedBytes != len(body) || got.ExpectedBytes != len(body) || got.ReturnedBytes != len(body) {
		t.Errorf("parseCodexStream(full cat) evidence = %+v, want complete byte coverage", got)
	}
}

func TestParseCodexSkillReadContract(t *testing.T) {
	t.Parallel()
	const path = "/tmp/arm home/.codex/skills/go-http/SKILL.md"
	const other = "/tmp/arm home/.codex/skills/go-style-core/SKILL.md"
	const body = "---\nname: go-http\n---\nHTTP handlers must decode requests. See go-style-core/SKILL.md.\n"
	const otherBody = "---\nname: go-style-core\n---\nFormat code and use guard clauses.\n"
	const partial = "---\nname: go-http\n---\n"
	tests := []struct {
		name, command, output, status, eventType string
		exit                                     any
		inventory                                map[string]string
		wantSkills                               []string
		wantStatus                               string
		wantMatched                              int
	}{
		{name: "full cat", command: "cat '" + path + "'", output: body, wantSkills: []string{"go-http"}, wantStatus: "full_returned_output", wantMatched: len(body)},
		{name: "zsh wrapper", command: "/bin/zsh -lc \"cat '" + path + "'\"", output: body, wantSkills: []string{"go-http"}, wantStatus: "full_returned_output", wantMatched: len(body)},
		{name: "escaped quote wrapper", command: "/bin/zsh -lc 'cat '\\''" + path + "'\\'''", output: body, wantSkills: []string{"go-http"}, wantStatus: "full_returned_output", wantMatched: len(body)},
		{name: "string zero", command: "cat '" + path + "'", output: body, exit: "0", wantSkills: []string{"go-http"}, wantStatus: "full_returned_output", wantMatched: len(body)},
		{name: "list", command: "ls '" + path + "'", output: path, wantStatus: "unverified"},
		{name: "echo", command: "echo '" + path + "'", output: body, wantStatus: "unverified"},
		{name: "path search", command: "rg SKILL.md '" + path + "'", output: body, wantStatus: "unverified"},
		{name: "failed cat", command: "cat '" + path + "'", output: body, exit: 1, status: "failed", wantStatus: "failed", wantMatched: len(body)},
		{name: "missing exit", command: "cat '" + path + "'", output: body, exit: "missing", wantStatus: "unverified", wantMatched: len(body)},
		{name: "null exit", command: "cat '" + path + "'", output: body, exit: "null", wantStatus: "unverified", wantMatched: len(body)},
		{name: "boolean exit", command: "cat '" + path + "'", output: body, exit: true, wantStatus: "unverified", wantMatched: len(body)},
		{name: "fractional exit", command: "cat '" + path + "'", output: body, exit: 0.5, wantStatus: "unverified", wantMatched: len(body)},
		{name: "missing id", command: "cat '" + path + "'", output: body, wantStatus: "unverified", wantMatched: len(body)},
		{name: "invalid exit", command: "cat '" + path + "'", output: body, exit: "bad", wantStatus: "unverified", wantMatched: len(body)},
		{name: "nonzero exit", command: "cat '" + path + "'", output: body, exit: 7, wantStatus: "failed", wantMatched: len(body)},
		{name: "missing status", command: "cat '" + path + "'", output: body, status: "missing", wantStatus: "unverified", wantMatched: len(body)},
		{name: "incomplete", command: "cat '" + path + "'", output: body, eventType: "item.started", status: "in_progress", wantStatus: "unverified", wantMatched: len(body)},
		{name: "head frontmatter", command: "head -n 3 '" + path + "'", output: partial, wantStatus: "partial_returned_output", wantMatched: len(partial)},
		{name: "sed triple address", command: "sed -n '1,2,3p' '" + path + "'", output: body, wantStatus: "unverified"},
		{name: "sed positive sign", command: "sed -n '+1p' '" + path + "'", output: body, wantStatus: "unverified"},
		{name: "sed signed range", command: "sed -n '1,+3p' '" + path + "'", output: body, wantStatus: "unverified"},
		{name: "sed negative address", command: "sed -n '-1p' '" + path + "'", output: body, wantStatus: "unverified"},
		{name: "sed regex address", command: "sed -n '/name/p' '" + path + "'", output: body, wantStatus: "unverified"},
		{name: "sed stepped address", command: "sed -n '1~2p' '" + path + "'", output: body, wantStatus: "unverified"},
		{name: "sed single address", command: "sed -n '1p' '" + path + "'", output: "---\n", wantStatus: "partial_returned_output", wantMatched: 4},
		{name: "sed full range", command: "sed -n '1,$p' '" + path + "'", output: body, wantSkills: []string{"go-http"}, wantStatus: "full_returned_output", wantMatched: len(body)},
		{name: "sed frontmatter", command: "sed -n '1,3p' '" + path + "'", output: partial, wantStatus: "partial_returned_output", wantMatched: len(partial)},
		{name: "sed interior", command: "sed -n '4,4p' '" + path + "'", output: strings.TrimPrefix(body, partial), wantStatus: "partial_returned_output", wantMatched: len(body) - len(partial)},
		{name: "head full", command: "head -n 999 '" + path + "'", output: body, wantSkills: []string{"go-http"}, wantStatus: "full_returned_output", wantMatched: len(body)},
		{name: "cropped cat", command: "cat '" + path + "'", output: partial, wantStatus: "partial_returned_output", wantMatched: len(partial)},
		{name: "truncation marker", command: "cat '" + path + "'", output: partial + "[output truncated]", wantStatus: "partial_returned_output", wantMatched: len(partial)},
		{name: "wrong body", command: "cat '" + path + "'", output: otherBody, wantStatus: "unverified"},
		{name: "wrong name", command: "cat '" + path + "'", output: strings.Replace(body, "name: go-http", "name: go-naming", 1), wantStatus: "unverified"},
		{name: "crosslink only", command: "cat '" + other + "'", output: "See go-http/SKILL.md.", wantStatus: "unverified"},
		{name: "backup file", command: "cat '" + path + ".bak'", output: body, wantStatus: "unverified"},
		{name: "outside arm", command: "cat /operator/skills/go-http/SKILL.md", output: body, wantStatus: "unverified"},
		{name: "empty inventory", command: "cat '" + path + "'", output: body, inventory: map[string]string{}, wantStatus: "unverified"},
		{name: "empty expected", command: "cat '" + path + "'", output: body, inventory: map[string]string{path: ""}, wantStatus: "unverified"},
		{name: "empty output", command: "cat '" + path + "'", wantStatus: "unverified"},
		{name: "complex pipeline", command: "cat '" + path + "' | head", output: body, wantStatus: "unverified"},
		{name: "complex compound", command: "echo '" + path + "'; cat '" + path + "'", output: body, wantStatus: "unverified"},
		{name: "substitution", command: "cat $(echo '" + path + "')", output: body, wantStatus: "unverified"},
		{name: "custom head executable", command: "/tmp/head -n 999 '" + path + "'", output: body, wantStatus: "unverified"},
		{name: "custom sed executable", command: "/tmp/sed -n '1,$p' '" + path + "'", output: body, wantStatus: "unverified"},
		{name: "custom executable", command: "/tmp/cat '" + path + "'", output: body, wantStatus: "unverified"},
		{name: "prefix cropped", command: "cat '" + path + "'", output: body[:8], wantStatus: "partial_returned_output", wantMatched: 8},
		{name: "escaped spaces", command: "cat " + strings.ReplaceAll(path, " ", "\\ "), output: body, wantSkills: []string{"go-http"}, wantStatus: "full_returned_output", wantMatched: len(body)},
		{name: "cat option", command: "cat -n '" + path + "'", output: body, wantStatus: "unverified"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.status == "" {
				tt.status = "completed"
			}
			if tt.eventType == "" {
				tt.eventType = "item.completed"
			}
			if tt.exit == nil {
				tt.exit = 0
			}
			item := map[string]any{"id": "i1", "type": "command_execution", "command": tt.command, "status": tt.status, "exit_code": tt.exit, "aggregated_output": tt.output}
			if tt.name == "missing id" {
				delete(item, "id")
			}
			if tt.exit == "missing" {
				delete(item, "exit_code")
			}
			if tt.exit == "null" {
				item["exit_code"] = nil
			}
			if tt.status == "missing" {
				delete(item, "status")
			}
			event, err := json.Marshal(map[string]any{"type": tt.eventType, "item": item})
			if err != nil {
				t.Fatal(err)
			}
			inventory := tt.inventory
			if inventory == nil {
				inventory = map[string]string{path: body, other: otherBody}
			}
			skills, _, _, evidence := parseCodexStream(event, inventory)
			if !slices.Equal(skills, tt.wantSkills) {
				t.Errorf("parseCodexStream(%s) skills = %v, want %v", event, skills, tt.wantSkills)
			}
			if len(evidence) != 1 {
				t.Fatalf("parseCodexStream(%s) evidence = %+v, want one record", event, evidence)
			}
			if got := evidence[0]; got.Status != tt.wantStatus || got.MatchedBytes != tt.wantMatched || got.Reason == "" {
				t.Errorf("parseCodexStream(%s) evidence = %+v, want status %q matched %d with reason", event, got, tt.wantStatus, tt.wantMatched)
			}
		})
	}
}

func TestParseCodexMultiFileRead(t *testing.T) {
	t.Parallel()
	const first = "/tmp/h/.codex/skills/go-http/SKILL.md"
	const second = "/tmp/h/.codex/skills/go-style-core/SKILL.md"
	const body = "---\nname: go-http\n---\nHTTP body with go-style-core/SKILL.md crosslink.\n"
	const next = "---\nname: go-style-core\n---\nStyle body.\n"
	tests := []struct {
		name, output string
		want         []string
	}{
		{name: "full", output: body + next, want: []string{"go-http", "go-style-core"}},
		{name: "second cropped", output: body + next[:12], want: []string{"go-http"}},
		{name: "reversed bodies", output: next + body},
		{name: "one body", output: body, want: []string{"go-http"}},
		{name: "prose mentioning both", output: "Read " + first + " and " + second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			event, err := json.Marshal(map[string]any{"type": "item.completed", "item": map[string]any{"id": "i1", "type": "command_execution", "command": "cat -- '" + first + "' '" + second + "'", "status": "completed", "exit_code": 0, "aggregated_output": tt.output}})
			if err != nil {
				t.Fatal(err)
			}
			skills, _, _, evidence := parseCodexStream(event, map[string]string{first: body, second: next})
			if !slices.Equal(skills, tt.want) || len(evidence) != 2 {
				t.Errorf("parseCodexStream(%s) = %v/%+v, want skills %v and two records", event, skills, evidence, tt.want)
			}
		})
	}
}

func TestCodexSkillInventory(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	root := filepath.Join(codexHomeDir(home), "skills")
	path := filepath.Join(root, "go-http", "SKILL.md")
	const body = "---\nname: go-http\n---\nExpected staged arm content.\n"
	writeTestFile(t, path, body)
	inventory, err := codexSkillInventory(arm{dir: "staged", home: home})
	if err != nil {
		t.Fatalf("codexSkillInventory(%q) error = %v, want nil", home, err)
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	if inventory[path] != body || inventory[resolved] != body {
		t.Errorf("codexSkillInventory(%q) = %v, want both physical and staged paths", home, inventory)
	}
	alias := filepath.Join(t.TempDir(), "arm alias")
	if err := os.Symlink(home, alias); err != nil {
		t.Fatal(err)
	}
	inventory, err = codexSkillInventory(arm{dir: "staged", home: alias})
	if err != nil || inventory[filepath.Join(codexHomeDir(alias), "skills", "go-http", "SKILL.md")] != body || inventory[resolved] != body {
		t.Errorf("codexSkillInventory(alias %q) = %v/%v, want trusted alias and physical paths", alias, inventory, err)
	}
	empty, err := codexSkillInventory(arm{home: home})
	if err != nil || len(empty) != 0 {
		t.Errorf("codexSkillInventory(control) = %v/%v, want empty inventory", empty, err)
	}
	writeTestFile(t, path, "")
	if _, err := codexSkillInventory(arm{dir: "staged", home: home}); err == nil {
		t.Error("codexSkillInventory(empty SKILL.md) error = nil, want error")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err := codexSkillInventory(arm{dir: "staged", home: home}); err == nil {
		t.Error("codexSkillInventory(missing SKILL.md) error = nil, want error")
	}
	if _, err := codexSkillInventory(arm{dir: "staged", home: t.TempDir()}); err == nil {
		t.Error("codexSkillInventory(missing skills directory) error = nil, want error")
	}
}

func TestCodexSkillInventoryRejectsBackupAlias(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	path := filepath.Join(codexHomeDir(home), "skills", "go-http", "SKILL.md")
	backup := path + ".bak"
	const body = "---\nname: go-http\n---\nTrusted staged content.\n"
	writeTestFile(t, backup, body)
	if err := os.Symlink("SKILL.md.bak", path); err != nil {
		t.Fatal(err)
	}
	physical, err := filepath.EvalSymlinks(backup)
	if err != nil {
		t.Fatal(err)
	}
	inventory, err := codexSkillInventory(arm{dir: "staged", home: home})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := inventory[physical]; ok {
		t.Errorf("codexSkillInventory(SKILL.md symlink to %q) trusts backup alias, want logical SKILL.md only", physical)
	}
	for _, tt := range []struct {
		name, path string
		want       []string
	}{
		{name: "logical skill symlink", path: path, want: []string{"go-http"}},
		{name: "physical backup alias", path: physical},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			event, err := json.Marshal(map[string]any{"type": "item.completed", "item": map[string]any{"id": "i1", "type": "command_execution", "command": "cat '" + tt.path + "'", "status": "completed", "exit_code": 0, "aggregated_output": body}})
			if err != nil {
				t.Fatal(err)
			}
			skills, _, _, evidence := parseCodexStream(event, inventory)
			if !slices.Equal(skills, tt.want) {
				t.Errorf("parseCodexStream(%s, filesystem inventory) skills = %v, want %v; evidence = %+v", event, skills, tt.want, evidence)
			}
		})
	}
}

func TestCodexSkillInventoryRejectsForeignSkillAlias(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	logical := filepath.Join(codexHomeDir(home), "skills", "go-http", "SKILL.md")
	foreign := filepath.Join(t.TempDir(), "go-style-core", "SKILL.md")
	const body = "---\nname: go-http\n---\nThis arm contains only go-http.\n"
	writeTestFile(t, foreign, body)
	if err := os.MkdirAll(filepath.Dir(logical), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(foreign, logical); err != nil {
		t.Fatal(err)
	}
	physical, err := filepath.EvalSymlinks(foreign)
	if err != nil {
		t.Fatal(err)
	}
	inventory, err := codexSkillInventory(arm{dir: "staged", home: home})
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name, path string
		want       []string
	}{
		{name: "logical owner", path: logical, want: []string{"go-http"}},
		{name: "foreign physical name", path: physical},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			event, err := json.Marshal(map[string]any{"type": "item.completed", "item": map[string]any{"id": "i1", "type": "command_execution", "command": "cat '" + tt.path + "'", "status": "completed", "exit_code": 0, "aggregated_output": body}})
			if err != nil {
				t.Fatal(err)
			}
			skills, _, _, evidence := parseCodexStream(event, inventory)
			if !slices.Equal(skills, tt.want) {
				t.Errorf("parseCodexStream(%s, go-http-only filesystem inventory) skills = %v, want %v; evidence = %+v", event, skills, tt.want, evidence)
			}
		})
	}
	if _, ok := inventory[physical]; ok {
		t.Errorf("codexSkillInventory(%q -> %q) trusts foreign skill name, want logical owner only", logical, physical)
	}
}

func TestRunSessionCodexEvidence(t *testing.T) {
	bin, home, work, staged := t.TempDir(), t.TempDir(), t.TempDir(), t.TempDir()
	path := filepath.Join(codexHomeDir(home), "skills", "go-http", "SKILL.md")
	const body = "---\nname: go-http\n---\nStaged installed body.\n"
	const stale = "---\nname: go-http\n---\nOther arm body.\n"
	writeTestFile(t, path, body)
	writeTestFile(t, filepath.Join(staged, "skills", "go-http", "SKILL.md"), stale)
	transcript, err := json.Marshal(map[string]any{"type": "item.completed", "item": map[string]any{"id": "i1", "type": "command_execution", "command": "cat '" + path + "'", "exit_code": 0, "status": "completed", "aggregated_output": body}})
	if err != nil {
		t.Fatal(err)
	}
	transcriptPath := filepath.Join(work, "fixture.jsonl")
	writeTestFile(t, transcriptPath, string(transcript)+"\n")
	cli := filepath.Join(bin, "codex")
	writeTestFile(t, cli, "#!/bin/sh\ncat fixture.jsonl\nprintf 'mutated after invocation\\n' > \"$CODEX_HOME/skills/go-http/SKILL.md\"\n")
	if err := os.Chmod(cli, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	o := options{runner: runnerCodex, timeout: time.Second}
	a := arm{dir: staged, home: home}
	turn := runSession(o, a, work, "fixture")
	if turn.err != nil || !slices.Equal(turn.skills, []string{"go-http"}) || turn.codexEvidence == nil || turn.codexEvidence.InventoryStatus != "available" {
		t.Fatalf("runSession(staged arm) = %+v, want successful evidence from pre-invocation installed copy", turn)
	}
	var res result
	res.merge(turn)
	// Repair використовує новий snapshot; первинний доказ зберігається.
	writeTestFile(t, path, body)
	writeTestFile(t, transcriptPath, strings.Replace(string(transcript), "Staged installed body.", "Staged", 1)+"\n")
	repair := runSession(o, a, work, "repair")
	res.merge(repair)
	if res.CodexSkillEvidence == nil || len(res.CodexSkillEvidence.Reads) != 2 || !slices.Equal(res.Skills, []string{"go-http"}) {
		t.Errorf("merge(first + repair) = %+v, want both turns' evidence and one confirmed skill", res.CodexSkillEvidence)
	}
	control := runSession(o, arm{home: home}, work, "control")
	if len(control.skills) != 0 || control.codexEvidence == nil || control.codexEvidence.InventoryStatus != "empty" {
		t.Errorf("runSession(control) = %+v, want empty inventory without loads", control)
	}
	marker := filepath.Join(work, "unexpected-codex-invocation")
	writeTestFile(t, cli, "#!/bin/sh\nprintf invoked > unexpected-codex-invocation\nexit 1\n")
	unavailable := runSession(o, arm{dir: staged, home: t.TempDir()}, work, "unavailable")
	if unavailable.err == nil || unavailable.codexEvidence == nil || unavailable.codexEvidence.InventoryStatus != "unavailable" || len(unavailable.out) != 0 {
		t.Errorf("runSession(unavailable inventory) = %+v, want explicit failure before CLI invocation", unavailable)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Errorf("runSession(unavailable inventory) invocation marker error = %v, want absent marker", err)
	}
	var old result
	if err := json.Unmarshal([]byte(`{"skills":["go-http"]}`), &old); err != nil {
		t.Fatal(err)
	}
	if old.CodexSkillEvidence != nil {
		t.Error("old JSON report has evidence, want absent evidence distinguishable")
	}
	data, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"codex_skill_evidence"`) || strings.Contains(string(data), "model_visible") {
		t.Errorf("json.Marshal(result) = %s, want captured-output evidence", data)
	}
}

func TestCodexCommandsIgnoresReadMetadataTypes(t *testing.T) {
	t.Parallel()
	const path = "/tmp/h/.codex/skills/go-http/SKILL.md"
	const body = "---\nname: go-http\n---\nRead the full staged skill.\n"
	tests := []struct {
		name           string
		status, output any
	}{
		{name: "numeric output", status: "completed", output: 42},
		{name: "boolean status", status: true, output: body},
		{name: "object output", status: "completed", output: map[string]any{"text": body}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			event, err := json.Marshal(map[string]any{"type": "item.completed", "item": map[string]any{"id": "i1", "type": "command_execution", "command": "cat " + path, "status": tt.status, "exit_code": 0, "aggregated_output": tt.output}})
			if err != nil {
				t.Fatal(err)
			}
			if got := codexCommands(event); got != 1 {
				t.Errorf("codexCommands(%s) = %d, want one attempt regardless of unusable read metadata", event, got)
			}
			skills, _, _, evidence := parseCodexStream(event, map[string]string{path: body})
			if len(skills) != 0 {
				t.Errorf("parseCodexStream(%s) skills = %v, want none for unusable read metadata; evidence = %+v", event, skills, evidence)
			}
		})
	}
}
