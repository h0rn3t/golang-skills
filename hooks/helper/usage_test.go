package main

import (
	"bytes"
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

// usageEnv points the usage log at a fresh data directory and the plugin root
// at a manifest of the given version, and returns the data directory.
func usageEnv(t *testing.T, version string) string {
	t.Helper()
	data, root := t.TempDir(), t.TempDir()
	t.Setenv("CLAUDE_PLUGIN_DATA", data)
	t.Setenv("GOLANG_SKILLS_HOOK_ROOT", root)
	t.Setenv("GOLANG_SKILLS_USAGE_LOG", "")
	if version != "" {
		manifest := filepath.Join(root, ".claude-plugin", "plugin.json")
		if err := os.MkdirAll(filepath.Dir(manifest), 0o750); err != nil {
			t.Fatalf("mkdir manifest dir: %v", err)
		}
		if err := os.WriteFile(manifest, []byte(`{"name": "x", "version": "`+version+`"}`), 0o600); err != nil {
			t.Fatalf("write manifest: %v", err)
		}
	}
	return data
}

// usageLines decodes the log under data with the timestamps checked and
// cleared, so a test compares the rest of each event.
func usageLines(t *testing.T, data string) []usageEvent {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(data, "usage.jsonl"))
	if err != nil {
		t.Fatalf("read usage log: %v", err)
	}
	var events []usageEvent
	for line := range strings.Lines(string(raw)) {
		var e usageEvent
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatalf("usage log line %q: %v", line, err)
		}
		if _, err := time.Parse(time.RFC3339, e.Time); err != nil {
			t.Errorf("usage log line %q: ts does not parse as RFC 3339: %v", line, err)
		}
		e.Time = ""
		events = append(events, e)
	}
	return events
}

func mustUsage(t *testing.T, args ...string) {
	t.Helper()
	if err := usageMain(args); err != nil {
		t.Fatalf("usageMain(%q) error = %v, want nil", args, err)
	}
}

func TestUsageLog(t *testing.T) {
	data := usageEnv(t, "9.8.7")
	mustUsage(t, "load", "s1/agent7", "go-testing", "Skill")
	mustUsage(t, "block", "s1", "2", "go-http", "go-testing")
	t.Setenv("GOLANG_SKILLS_USAGE_LOG", "off")
	mustUsage(t, "load", "s1", "go-http", "Read")

	want := []usageEvent{
		{Version: "9.8.7", Session: "s1", Agent: true, Event: "load", Skill: "go-testing", Via: "Skill"},
		{Version: "9.8.7", Session: "s1", Event: "block", Attempt: 2, Missing: []string{"go-http", "go-testing"}},
	}
	if got := usageLines(t, data); !reflect.DeepEqual(got, want) {
		t.Errorf("usage log after a load, a block, and a load with the log off:\n got %+v\nwant %+v", got, want)
	}

	t.Setenv("GOLANG_SKILLS_USAGE_LOG", "")
	for _, args := range [][]string{nil, {"bogus", "s1"}, {"load", "s1", "go-http"}, {"block", "s1", "twice", "go-http"}, {"check", "s1"}} {
		if err := usageMain(args); err == nil {
			t.Errorf("usageMain(%q) error = nil, want a usage error", args)
		}
	}
}

func TestUsageCheck(t *testing.T) {
	data := usageEnv(t, "")
	run := t.TempDir()
	statuses := map[string]string{"gofmt": "pass", "vet": "fail", "fix": "pass", "test": "skipped", "lint": "fail"}
	for check, status := range statuses {
		r := receipt{Version: 1, Check: check, Status: status}
		if check == "lint" {
			r.Diagnostic = "/m/a.go:9:1: G104: Errors unhandled (gosec)\n"
		}
		if err := atomicJSON(filepath.Join(run, check+".json"), r); err != nil {
			t.Fatalf("write %s receipt: %v", check, err)
		}
	}
	mustUsage(t, "check", "s2", run)

	want := []usageEvent{{Session: "s2", Event: "check", Checks: statuses, Linters: map[string]int{"gosec/G104": 1}}}
	if got := usageLines(t, data); !reflect.DeepEqual(got, want) {
		t.Errorf("usage log after one edit-hook run:\n got %+v\nwant %+v", got, want)
	}
}

func TestUsageLogRotates(t *testing.T) {
	data := usageEnv(t, "")
	log := filepath.Join(data, "usage.jsonl")
	if err := os.WriteFile(log, nil, 0o600); err != nil {
		t.Fatalf("create usage log: %v", err)
	}
	if err := os.Truncate(log, usageCap+1); err != nil {
		t.Fatalf("grow usage log past the cap: %v", err)
	}
	mustUsage(t, "load", "s1", "go-code", "Skill")

	if fi, err := os.Stat(log + ".1"); err != nil || fi.Size() != usageCap+1 {
		t.Errorf("usage.jsonl.1 after a write past the cap: %v, %v; want the %d-byte log moved there", fi, err, usageCap+1)
	}
	if got := usageLines(t, data); len(got) != 1 {
		t.Errorf("usage.jsonl after rotation holds %d events, want 1", len(got))
	}
}

func TestLinterCounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, in string
		want     map[string]int
	}{
		{"no output", "", map[string]int{}},
		{"linter only", "/m/a.go:3:2: Error return value of `f.Close` is not checked (errcheck)\n", map[string]int{"errcheck": 1}},
		{"rule prefix", "/m/a.go:9:1: SA4006: this value of err is never used (staticcheck)\n", map[string]int{"staticcheck/SA4006": 1}},
		{"same linter twice with source lines between",
			"/m/a.go:3:2: unchecked (errcheck)\n\tf.Close()\n\t^\n/m/b.go:1:1: unchecked (errcheck)\n", map[string]int{"errcheck": 2}},
		{"first word of a linter that prints no rule", "/m/builtin.go:5:2: max: shadows predeclared identifier (predeclared)\n", map[string]int{"predeclared": 1}},
		{"path with a space", "/m/my dir/a.go:3:2: unchecked (errcheck)\n", map[string]int{"errcheck": 1}},
		{"no column", "/m/a.go:37: line-length-limit: line is 129 characters, out of limit 80 (revive)\n", map[string]int{"revive/line-length-limit": 1}},
		{"position line without a linter", "/m/a.go:1:2: declared and not used: x\n", map[string]int{}},
		{"parenthesis that is not last", "/m/a.go:1:2: call (of f) copies a lock\n", map[string]int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := linterCounts(tt.in); !maps.Equal(got, tt.want) {
				t.Errorf("linterCounts(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// usageSample is two sessions inside the window and one before it. s1 loads
// go-testing twice and go-http from a subagent, is blocked three times (the
// third and fourth attempts at one edit, both of which stop the session), and
// has two failing edit-hook runs that report the same unfixed findings; s2
// runs a later plugin version.
func usageSample() []usageEvent {
	return []usageEvent{
		{Time: "2026-11-01T08:00:00.000000+00:00", Version: "1.0.0", Session: "s0", Event: "load", Skill: "go-naming", Via: "Skill"},
		{Time: "2026-11-10T09:00:00.000000+00:00", Version: "1.0.0", Session: "s1", Event: "block", Attempt: 1, Missing: []string{"go-testing", "go-http"}},
		{Time: "2026-11-10T09:00:01.000000+00:00", Version: "1.0.0", Session: "s1", Event: "load", Skill: "go-code", Via: "Skill"},
		{Time: "2026-11-10T09:00:02.000000+00:00", Version: "1.0.0", Session: "s1", Event: "load", Skill: "go-testing", Via: "Skill"},
		{Time: "2026-11-10T09:00:03.000000+00:00", Version: "1.0.0", Session: "s1", Event: "load", Skill: "go-testing", Via: "Read"},
		{Time: "2026-11-10T09:00:04.000000+00:00", Version: "1.0.0", Session: "s1", Agent: true, Event: "load", Skill: "go-http", Via: "Skill"},
		{Time: "2026-11-10T09:00:05.000000+00:00", Version: "1.0.0", Session: "s1", Event: "block", Attempt: 3, Missing: []string{"go-http"}},
		{Time: "2026-11-10T09:00:06.000000+00:00", Version: "1.0.0", Session: "s1", Event: "block", Attempt: 4, Missing: []string{"go-http"}},
		{Time: "2026-11-10T09:00:07.000000+00:00", Version: "1.0.0", Session: "s1", Event: "check",
			Checks:  map[string]string{"gofmt": "pass", "vet": "fail", "fix": "pass", "test": "skipped", "lint": "fail"},
			Linters: map[string]int{"errcheck": 2}},
		{Time: "2026-11-10T09:00:08.000000+00:00", Version: "1.0.0", Session: "s1", Event: "check",
			Checks:  map[string]string{"gofmt": "pass", "vet": "fail", "fix": "pass", "test": "skipped", "lint": "fail"},
			Linters: map[string]int{"errcheck": 2}},
		{Time: "2026-11-11T10:00:00.000000+00:00", Version: "2.0.0", Session: "s2", Event: "load", Skill: "go-code", Via: "slash"},
		{Time: "2026-11-11T10:00:01.000000+00:00", Version: "2.0.0", Session: "s2", Event: "check",
			Checks: map[string]string{"gofmt": "pass", "vet": "pass", "fix": "pass", "test": "pass", "lint": "pass"}},
		{Time: "not a time", Session: "s9", Event: "load", Skill: "go-naming"},
	}
}

func TestSummarize(t *testing.T) {
	t.Parallel()
	since := time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)
	got := summarize(usageSample(), []string{"go-code", "go-http", "go-naming", "go-testing"}, since)
	want := usageSummary{
		First: "2026-11-10", Last: "2026-11-11", Sessions: 2,
		Versions:    map[string]int{"1.0.0": 1, "2.0.0": 1},
		Loaded:      map[string]int{"go-code": 2, "go-testing": 1, "go-http": 1},
		NeverLoaded: []string{"go-naming"},
		Blocks:      3, BlockedSessions: 1, StoppedSessions: 1,
		Named:   map[string]int{"go-testing": 1, "go-http": 3},
		Runs:    3,
		Failed:  map[string]int{"vet": 2, "lint": 2},
		Linters: map[string]int{"errcheck": 1},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summarize(sample, four skills, since 2026-11-05):\n got %+v\nwant %+v", got, want)
	}

	if all := summarize(usageSample(), nil, time.Time{}); all.Sessions != 3 || all.First != "2026-11-01" || all.NeverLoaded != nil {
		t.Errorf("summarize(sample, no skill list, no window) = %d sessions from %s, never loaded %v; want 3 from 2026-11-01 and no list",
			all.Sessions, all.First, all.NeverLoaded)
	}
}

func TestSummaryText(t *testing.T) {
	t.Parallel()
	s := summarize(usageSample(), []string{"go-code", "go-http", "go-naming", "go-testing"}, time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC))
	s.File = "usage.jsonl"
	want := `usage log: usage.jsonl
2026-11-10 to 2026-11-11: sessions 2 (by version: 1.0.0 1, 2.0.0 1)

skills loaded, by sessions:
go-code 2 100%
go-http 1 50%
go-testing 1 50%
never loaded: go-naming

routing gate: blocks 3; sessions blocked 1 of 2; sessions stopped 1
named: go-http 3, go-testing 1

edit hook: runs 3; failed: lint 2, vet 2
lint findings, by sessions:
errcheck 1
`
	// Column widths are layout, not contract: compare the fields of each line.
	var got strings.Builder
	for line := range strings.Lines(s.text()) {
		got.WriteString(strings.Join(strings.Fields(line), " "))
		got.WriteByte('\n')
	}
	if got.String() != want {
		t.Errorf("summary text:\n got:\n%s\nwant:\n%s", got.String(), want)
	}

	if got := (usageSummary{File: "empty.jsonl"}).text(); got != "usage log: empty.jsonl\nno events\n" {
		t.Errorf("text of a summary with no session = %q, want the file and \"no events\"", got)
	}
}

func TestStats(t *testing.T) {
	log := filepath.Join(usageEnv(t, ""), "usage.jsonl")
	recent := time.Now().UTC().Format(time.RFC3339)
	old := `{"ts": "2020-01-01T00:00:00Z", "session": "old", "event": "load", "skill": "go-code"}` + "\n"
	current := "torn line\n" + `{"ts": "` + recent + `", "session": "new", "event": "load", "skill": "go-http"}` + "\n"
	if err := os.WriteFile(log+".1", []byte(old), 0o600); err != nil {
		t.Fatalf("write rotated log: %v", err)
	}
	if err := os.WriteFile(log, []byte(current), 0o600); err != nil {
		t.Fatalf("write log: %v", err)
	}

	// loaded checks that stats exits 0 with one JSON summary of log holding want.
	loaded := func(name string, args []string, want map[string]int) {
		t.Helper()
		var out bytes.Buffer
		if code := stats(&out, args); code != 0 {
			t.Errorf("%s: stats(%q) = exit %d, want 0", name, args, code)
			return
		}
		var got usageSummary
		if err := json.Unmarshal(out.Bytes(), &got); err != nil {
			t.Errorf("%s: stats(%q) output %q is not one JSON object: %v", name, args, out.String(), err)
			return
		}
		if got.File != log || !maps.Equal(got.Loaded, want) {
			t.Errorf("%s: stats(%q) = file %q, loaded %v; want %q, %v", name, args, got.File, got.Loaded, log, want)
		}
	}
	both := map[string]int{"go-code": 1, "go-http": 1}
	loaded("both generations, torn line skipped", []string{"--json", log}, both)
	loaded("window drops the old session", []string{"--json", "--days", "1", log}, map[string]int{"go-http": 1})
	loaded("default log is the data directory's", []string{"--json"}, both)
	loaded("flag after the path", []string{log, "--json"}, both)
	loaded("flags on both sides of the path", []string{"--days", "1", log, "--json"}, map[string]int{"go-http": 1})

	if err := os.Remove(log); err != nil {
		t.Fatalf("remove the current log: %v", err)
	}
	loaded("named log with only its rotated generation", []string{"--json", log}, map[string]int{"go-code": 1})
	loaded("default log with only its rotated generation", []string{"--json"}, map[string]int{"go-code": 1})

	var out bytes.Buffer
	if code := stats(&out, []string{filepath.Join(t.TempDir(), "usage.jsonl")}); code != 1 {
		t.Errorf("stats(missing log) = exit %d, want 1", code)
	}
	for _, args := range [][]string{{"--days", "soon"}, {"--days", "-1"}, {"--bogus"}} {
		if code := stats(&out, args); code != 2 {
			t.Errorf("stats(%q) = exit %d, want 2", args, code)
		}
	}
	out.Reset()
	if code := stats(&out, []string{"--help"}); code != 0 || !strings.Contains(out.String(), "json") || !strings.Contains(out.String(), "days") {
		t.Errorf("stats(--help) = exit %d, output %q; want 0 with both flags named", code, out.String())
	}
}

// TestUsageLogs covers a stats run from a terminal, where no hook names the
// data directory: one log per install under Claude Code's configuration
// directory, counting an install left with only its rotated generation.
func TestUsageLogs(t *testing.T) {
	for _, tt := range []struct {
		name, env, sub string
	}{
		{"CLAUDE_CONFIG_DIR", "CLAUDE_CONFIG_DIR", ""},
		{"home directory", "HOME", ".claude"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			usageEnv(t, "")
			t.Setenv("CLAUDE_PLUGIN_DATA", "")
			t.Setenv("CLAUDE_CONFIG_DIR", "")
			t.Setenv("TMPDIR", t.TempDir()) // no fallback log of this machine's own
			base := t.TempDir()
			t.Setenv(tt.env, base)
			data := filepath.Join(base, tt.sub, "plugins", "data")
			rotated := filepath.Join(data, "golang-skills-golang-skills", "usage.jsonl")
			inline := filepath.Join(data, "golang-skills-inline", "usage.jsonl")
			for _, file := range []string{rotated + ".1", inline, filepath.Join(data, "golang-skills-unused", "routing"), filepath.Join(data, "other-plugin", "usage.jsonl")} {
				if err := os.MkdirAll(filepath.Dir(file), 0o750); err != nil {
					t.Fatalf("mkdir for %s: %v", file, err)
				}
				if err := os.WriteFile(file, nil, 0o600); err != nil {
					t.Fatalf("write %s: %v", file, err)
				}
			}
			if got, want := usageLogs(), []string{rotated, inline}; !slices.Equal(got, want) {
				t.Errorf("usageLogs() with %s=%s = %q, want %q", tt.env, base, got, want)
			}
		})
	}
}
