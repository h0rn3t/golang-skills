package main

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// usageEvent is one line of <data>/usage.jsonl, the record of how the plugin
// is used on this machine: which skills a session loaded and by what route,
// which skills the routing gate named, and what the edit hook found. Beside
// the time, the plugin version, and the session id, which names the
// transcript a line came from, it holds skill names, check statuses, and
// linter and rule names only (no path, prompt, or code), and nothing sends it
// anywhere. The routing and check state beside it is pruned after two days;
// this file is what stays.
type usageEvent struct {
	Time    string            `json:"ts"`
	Version string            `json:"v,omitempty"`
	Session string            `json:"session"`
	Agent   bool              `json:"agent,omitempty"`
	Event   string            `json:"event"`
	Skill   string            `json:"skill,omitempty"`
	Via     string            `json:"via,omitempty"`
	Attempt int               `json:"attempt,omitempty"`
	Missing []string          `json:"missing,omitempty"`
	Checks  map[string]string `json:"checks,omitempty"`
	Linters map[string]int    `json:"linters,omitempty"`
}

// usageCap is the size past which the log moves to usage.jsonl.1, replacing
// the generation before it, so the two files bound the disk use.
const usageCap = 4 << 20

// dataDir is where the hooks keep state: the host's plugin data directory,
// else the fallback the scripts use.
func dataDir() string {
	return cmp.Or(os.Getenv("CLAUDE_PLUGIN_DATA"), filepath.Join(os.TempDir(), "golang-skills-hooks"))
}

// usageMain appends one event: load <session-key> <skill> <via>,
// block <session-key> <attempt> <missing>..., or check <session-key> <run-dir>.
// The session key is the session id, plus "/<agent>" in a subagent.
// GOLANG_SKILLS_USAGE_LOG=off writes nothing.
func usageMain(args []string) error {
	if len(args) < 2 {
		return errors.New("want load|block|check, a session key, and the event's arguments")
	}
	if os.Getenv("GOLANG_SKILLS_USAGE_LOG") == "off" {
		return nil
	}
	e := usageEvent{Event: args[0]}
	e.Session, _, e.Agent = strings.Cut(args[1], "/")
	rest := args[2:]
	switch {
	case e.Event == "load" && len(rest) == 2:
		e.Skill, e.Via = rest[0], rest[1]
	case e.Event == "block" && len(rest) >= 2:
		attempt, err := strconv.Atoi(rest[0])
		if err != nil {
			return fmt.Errorf("block attempt: %w", err)
		}
		e.Attempt, e.Missing = attempt, rest[1:]
	case e.Event == "check" && len(rest) == 1:
		e.Checks = map[string]string{}
		for _, check := range checkNames {
			r, err := readReceipt(filepath.Join(rest[0], check+".json"))
			if err != nil {
				continue // a check the hook never reached left no receipt
			}
			e.Checks[check] = r.Status
			if check == "lint" && r.Status == "fail" {
				e.Linters = linterCounts(r.Diagnostic)
			}
		}
		if len(e.Checks) == 0 {
			return fmt.Errorf("no receipt under %s", rest[0])
		}
	default:
		return fmt.Errorf("%q: unknown event or wrong argument count", args)
	}
	return appendUsage(e)
}

// appendUsage stamps e with the time and the plugin version and writes it as
// one line.
func appendUsage(e usageEvent) error {
	e.Time = now()
	// An empty root would read the project's own manifest, not the plugin's.
	if root := os.Getenv("GOLANG_SKILLS_HOOK_ROOT"); root != "" {
		var manifest struct{ Version string }
		if readJSON(filepath.Join(root, ".claude-plugin", "plugin.json"), &manifest) == nil {
			e.Version = manifest.Version
		}
	}
	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	file := filepath.Join(dataDir(), "usage.jsonl")
	if err := os.MkdirAll(filepath.Dir(file), 0o750); err != nil {
		return err
	}
	// Two hooks crossing the cap in the same instant both rotate: the second
	// finds the log already moved, which is no failure, or moves the first
	// one's new file over the older generation. The log is statistics, so no
	// lock guards it.
	if fi, err := os.Stat(file); err == nil && fi.Size() > usageCap {
		if err := os.Rename(file, file+".1"); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	// One Write of a short line under O_APPEND: hooks running at once never
	// interleave their events.
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	_, err = f.Write(append(line, '\n'))
	return errors.Join(err, f.Close())
}

// lintFinding is a golangci-lint text line as the edit hook takes one
// (lint_issue in go-vet-on-edit.sh, positionLine in receipt.go): an absolute
// path, a line, a column unless the linter reports whole lines, the message,
// and the linter in parentheses at the end. The first group is the message's
// first word when a colon follows it.
var lintFinding = regexp.MustCompile(`(?m)^/.*?:\d+(?::\d+)?: (?:([A-Za-z][\w-]*): )?.*\(([\w-]+)\)$`)

// ruleLinters are the linters whose message leads with the rule that fired
// ("G104: ", "exported: "), as golangci-lint 2.15.0 prints them. Any other
// linter's first word is text from the code, such as the identifier in
// predeclared's "max: shadows predeclared identifier", and stays out of the
// log.
var ruleLinters = map[string]bool{
	"gocritic": true, "gosec": true, "govet": true, "modernize": true, "revive": true, "staticcheck": true,
}

// linterCounts counts the findings of a lint diagnostic by linter, or by
// "linter/rule" for the linters that name a rule.
func linterCounts(diagnostic string) map[string]int {
	found := map[string]int{}
	for _, m := range lintFinding.FindAllStringSubmatch(diagnostic, -1) {
		rule, name := m[1], m[2]
		if rule != "" && ruleLinters[name] {
			name += "/" + rule
		}
		found[name]++
	}
	return found
}

// readUsage returns the events of a log and of the generation rotated out
// before it, oldest first; either file may be missing, not both. A line that
// does not parse, such as a write cut short, is skipped.
func readUsage(file string) ([]usageEvent, error) {
	var events []usageEvent
	found := false
	for _, name := range []string{file + ".1", file} {
		data, err := os.ReadFile(name)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		found = true
		for line := range bytes.Lines(data) {
			var e usageEvent
			if json.Unmarshal(line, &e) == nil {
				events = append(events, e)
			}
		}
	}
	if !found {
		return nil, fmt.Errorf("%s: %w", file, os.ErrNotExist)
	}
	return events, nil
}

// usageSummary is what stats reports for one log. Loaded, Versions, and
// Linters count sessions, as BlockedSessions and StoppedSessions do; Blocks
// counts gate answers, Named the blocks that named a skill, and Runs and
// Failed edit-hook runs.
type usageSummary struct {
	File            string         `json:"file"`
	First           string         `json:"first,omitempty"`
	Last            string         `json:"last,omitempty"`
	Sessions        int            `json:"sessions"`
	Versions        map[string]int `json:"versions"`
	Loaded          map[string]int `json:"loaded"`
	NeverLoaded     []string       `json:"never_loaded,omitempty"`
	Blocks          int            `json:"blocks"`
	BlockedSessions int            `json:"blocked_sessions"`
	StoppedSessions int            `json:"stopped_sessions"`
	Named           map[string]int `json:"named"`
	Runs            int            `json:"runs"`
	Failed          map[string]int `json:"failed"`
	Linters         map[string]int `json:"linters"`
}

// summarize totals the events at or after since. skills is the plugin's skill
// list, from which NeverLoaded is taken; a subagent's events count toward its
// session.
func summarize(events []usageEvent, skills []string, since time.Time) usageSummary {
	s := usageSummary{Versions: map[string]int{}, Loaded: map[string]int{}, Named: map[string]int{},
		Failed: map[string]int{}, Linters: map[string]int{}}
	seen := map[[3]string]bool{}
	firstIn := func(kind, name, session string) bool {
		key := [3]string{kind, name, session}
		if seen[key] {
			return false
		}
		seen[key] = true
		return true
	}
	var oldest, newest time.Time
	for _, e := range events {
		ts, err := time.Parse(time.RFC3339, e.Time)
		if err != nil || ts.Before(since) {
			continue
		}
		if oldest.IsZero() || ts.Before(oldest) {
			oldest = ts
		}
		if ts.After(newest) {
			newest = ts
		}
		if firstIn("session", "", e.Session) {
			s.Sessions++
		}
		if version := cmp.Or(e.Version, "unknown"); firstIn("version", version, e.Session) {
			s.Versions[version]++
		}
		switch e.Event {
		case "load":
			if firstIn("load", e.Skill, e.Session) {
				s.Loaded[e.Skill]++
			}
		case "block":
			s.Blocks++
			if firstIn("block", "", e.Session) {
				s.BlockedSessions++
			}
			// The gate logs every stalled retry, and it stops each one from
			// the third on: the session was stopped, however often.
			if e.Attempt >= 3 && firstIn("stop", "", e.Session) {
				s.StoppedSessions++
			}
			for _, skill := range e.Missing {
				s.Named[skill]++
			}
		case "check":
			s.Runs++
			for check, status := range e.Checks {
				if status == "fail" {
					s.Failed[check]++
				}
			}
			// The hook lints again after every edit and reports a finding
			// until it is fixed, so a sum would rank the linters by how long
			// the sessions ran.
			for linter := range e.Linters {
				if firstIn("linter", linter, e.Session) {
					s.Linters[linter]++
				}
			}
		}
	}
	if s.Sessions > 0 {
		s.First, s.Last = oldest.UTC().Format(time.DateOnly), newest.UTC().Format(time.DateOnly)
	}
	for _, skill := range skills {
		if s.Loaded[skill] == 0 {
			s.NeverLoaded = append(s.NeverLoaded, skill)
		}
	}
	return s
}

// ranked is the keys of m, largest count first, ties by name.
func ranked(m map[string]int) []string {
	return slices.SortedFunc(maps.Keys(m), func(a, b string) int {
		return cmp.Or(cmp.Compare(m[b], m[a]), cmp.Compare(a, b))
	})
}

// counts renders m as "name n, name n" in ranked order.
func counts(m map[string]int) string {
	var parts []string
	for _, name := range ranked(m) {
		parts = append(parts, name+" "+strconv.Itoa(m[name]))
	}
	return strings.Join(parts, ", ")
}

func (s usageSummary) text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "usage log: %s\n", s.File)
	if s.Sessions == 0 {
		b.WriteString("no events\n")
		return b.String()
	}
	fmt.Fprintf(&b, "%s to %s: sessions %d (by version: %s)\n", s.First, s.Last, s.Sessions, counts(s.Versions))
	b.WriteString("\nskills loaded, by sessions:\n")
	for _, skill := range ranked(s.Loaded) {
		fmt.Fprintf(&b, "  %-20s %4d %4d%%\n", skill, s.Loaded[skill], 100*s.Loaded[skill]/s.Sessions)
	}
	if len(s.NeverLoaded) > 0 {
		fmt.Fprintf(&b, "never loaded: %s\n", strings.Join(s.NeverLoaded, ", "))
	}
	fmt.Fprintf(&b, "\nrouting gate: blocks %d; sessions blocked %d of %d; sessions stopped %d\n",
		s.Blocks, s.BlockedSessions, s.Sessions, s.StoppedSessions)
	if len(s.Named) > 0 {
		fmt.Fprintf(&b, "  named: %s\n", counts(s.Named))
	}
	fmt.Fprintf(&b, "\nedit hook: runs %d", s.Runs)
	if len(s.Failed) > 0 {
		fmt.Fprintf(&b, "; failed: %s", counts(s.Failed))
	}
	b.WriteString("\n")
	if len(s.Linters) > 0 {
		b.WriteString("lint findings, by sessions:\n")
		for _, linter := range ranked(s.Linters) {
			fmt.Fprintf(&b, "  %-28s %4d\n", linter, s.Linters[linter])
		}
	}
	return b.String()
}

// usageLogs is the logs stats reads when none is named: the data directory's
// inside a hook, else one per install under Claude Code's plugin data
// (golang-skills-<marketplace>, or golang-skills-inline for --plugin-dir, in
// CLAUDE_CONFIG_DIR or ~/.claude) plus the fallback directory's. A directory
// left with only the rotated generation counts.
func usageLogs() []string {
	dirs := []string{dataDir()}
	if os.Getenv("CLAUDE_PLUGIN_DATA") == "" {
		config := os.Getenv("CLAUDE_CONFIG_DIR")
		if home, err := os.UserHomeDir(); config == "" && err == nil {
			config = filepath.Join(home, ".claude")
		}
		if config != "" {
			installs, _ := filepath.Glob(filepath.Join(config, "plugins", "data", "golang-skills*")) // a malformed pattern is the only error
			dirs = append(installs, dirs...)
		}
	}
	var logs []string
	for _, dir := range dirs {
		if file := filepath.Join(dir, "usage.jsonl"); isFile(file) || isFile(file+".1") {
			logs = append(logs, file)
		}
	}
	return logs
}

// stats prints one summary per usage log, as text or with --json as one JSON
// object per line; --days N keeps the last N days, and flags and logs come in
// any order. It returns 0 when a log was read or --help asked, 1 when no log
// was read, and 2 for a usage error.
func stats(w io.Writer, args []string) int {
	// The name is the line --help prints after "Usage of".
	flags := flag.NewFlagSet("go-hook-helper.sh stats [--json] [--days N] [usage.jsonl ...]", flag.ContinueOnError)
	asJSON := flags.Bool("json", false, "print one JSON object per log")
	days := flags.Int("days", 0, "count only the events of the last `N` days")
	flags.SetOutput(io.Discard) // Parse would send --help to stderr too
	var logs []string
	for {
		err := flags.Parse(args)
		if errors.Is(err, flag.ErrHelp) {
			flags.SetOutput(w)
			flags.Usage()
			return 0
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "helper: stats:", err)
			return 2
		}
		if flags.NArg() == 0 {
			break
		}
		// Parse stops at the first log: go on behind it, so that a flag
		// there is not read as a second log.
		logs = append(logs, flags.Arg(0))
		args = flags.Args()[1:]
	}
	if *days < 0 {
		fmt.Fprintln(os.Stderr, "helper: stats: --days wants a number of days, not", *days)
		return 2
	}
	if len(logs) == 0 {
		logs = usageLogs()
	}
	var since time.Time
	if *days > 0 {
		since = time.Now().AddDate(0, 0, -*days)
	}
	var skills []string
	if root := os.Getenv("GOLANG_SKILLS_HOOK_ROOT"); root != "" {
		files, _ := filepath.Glob(filepath.Join(root, "skills", "go-*", "SKILL.md")) // a malformed pattern is the only error
		for _, f := range files {
			skills = append(skills, filepath.Base(filepath.Dir(f)))
		}
	}
	var out strings.Builder
	code := 1
	for _, file := range logs {
		events, err := readUsage(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, "helper: stats:", err)
			continue
		}
		s := summarize(events, skills, since)
		s.File = file
		switch {
		case *asJSON:
			fmt.Fprintln(&out, encode(s))
		case code == 0:
			fmt.Fprint(&out, "\n", s.text())
		default:
			out.WriteString(s.text())
		}
		code = 0
	}
	if len(logs) == 0 {
		fmt.Fprintln(os.Stderr, "helper: stats: no usage log under", dataDir(),
			"or Claude Code's plugins/data/golang-skills*; name the file, as another host keeps its plugin data elsewhere")
	}
	if _, err := io.WriteString(w, out.String()); err != nil {
		fmt.Fprintln(os.Stderr, "helper: stats:", err)
		return 1
	}
	return code
}
