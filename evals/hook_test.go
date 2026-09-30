package evals_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// hookEvent runs one plugin hook with a Claude Code hook payload on stdin and
// returns its exit code and stderr. state is the CLAUDE_PLUGIN_DATA directory,
// which the routing hook uses to remember what a session has loaded.
func hookEvent(t *testing.T, script, state string, payload map[string]any) (int, string) {
	t.Helper()
	code, _, stderr := hookOutput(t, script, state, payload)
	return code, stderr
}

// hookOutput runs a hook the way the host runs a plugin hook:
// CLAUDE_PLUGIN_ROOT points at the checkout, CLAUDE_PLUGIN_DATA at state, and
// env is appended last and overrides both. It returns the exit code, stdout,
// and stderr: the gate reports a session stop as JSON on stdout.
func hookOutput(t *testing.T, script, state string, payload map[string]any, env ...string) (int, string, string) {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	cmd := exec.Command("bash", script)
	cmd.Stdin = strings.NewReader(string(body))
	cmd.Env = append(append(os.Environ(), "CLAUDE_PLUGIN_ROOT="+repoRoot(t), "CLAUDE_PLUGIN_DATA="+state), env...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err == nil {
		return 0, stdout.String(), stderr.String()
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("run %s: %v", script, err)
	}
	return exitErr.ExitCode(), stdout.String(), stderr.String()
}

func routingPayload(event, session, tool string, input map[string]any) map[string]any {
	return map[string]any{
		"hook_event_name": event,
		"session_id":      session,
		"tool_name":       tool,
		"tool_input":      input,
	}
}

func TestHookScriptsSyntax(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "hooks", "hooks.json"))
	if err != nil {
		t.Fatalf("read hooks.json: %v", err)
	}
	var manifest struct {
		Hooks map[string][]struct {
			Matcher string `json:"matcher"`
			Hooks   []struct {
				Command string `json:"command"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatalf("hooks.json is not valid JSON: %v", err)
	}
	if len(manifest.Hooks["PreToolUse"]) == 0 || len(manifest.Hooks["PostToolUse"]) == 0 {
		t.Fatalf("hooks.json must register PreToolUse and PostToolUse hooks, got %v", manifest.Hooks)
	}
	seen := map[string]bool{}
	for event, entries := range manifest.Hooks {
		for _, entry := range entries {
			for _, h := range entry.Hooks {
				// Commands look like: bash "${CLAUDE_PLUGIN_ROOT}/hooks/<name>.sh"
				start := strings.Index(h.Command, "/hooks/")
				end := strings.Index(h.Command, ".sh")
				if start < 0 || end < 0 {
					t.Fatalf("%s hook command does not name a hooks/*.sh script: %q", event, h.Command)
				}
				script := filepath.Join(root, h.Command[start+1:end+3])
				seen[script] = true
				if _, err := os.Stat(script); err != nil {
					t.Errorf("%s hook points at a missing script: %s", event, script)
				}
			}
		}
	}
	for script := range seen {
		if out, err := exec.Command("bash", "-n", script).CombinedOutput(); err != nil {
			t.Errorf("bash -n %s: %v\n%s", script, err, out)
		}
	}
}

// TestRoutingGate drives the routing hook through one session: a block without
// a router that names the entry skill, a block that names exactly the missing
// owners after any of the three routers, a pass once they are loaded, a block
// again on a retry that loaded nothing, and a JSON stop instead of a third
// block.
func TestRoutingGate(t *testing.T) {
	t.Parallel()
	script := filepath.Join(repoRoot(t), "hooks", "go-code-routing.sh")
	handler := "package api\n\nfunc (s *Server) handle(w http.ResponseWriter, r *http.Request) {\n\tif err := s.store.Save(r.Context(), u); err != nil {\n\t\thttp.Error(w, fmt.Errorf(\"save: %w\", err).Error(), 500)\n\t}\n}\n"

	// A session without a router used to pass silently, and the loads depended
	// on the model's choice alone. Now the first .go edit is blocked and names
	// the entry skill, go-style-core, the owners, and the card by their exact
	// plugin names.
	t.Run("blocks without a router and names the entry skill", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		code, msg := hookEvent(t, script, state, routingPayload("PreToolUse", "s1", "Write",
			map[string]any{"file_path": "/repo/api/handler.go", "content": handler}))
		if code != 2 {
			t.Fatalf("edit without a router loaded: exit %d, stderr %q; want 2", code, msg)
		}
		cardPath := filepath.Join(repoRoot(t), "skills", "go-style-core", "references", "CURRENT-GO.md")
		for _, want := range []string{"loaded no router skill", "`golang-skills:go-code`", "`golang-skills:go-code-refactor`",
			"`golang-skills:go-style-core`", "`golang-skills:go-http`", "`golang-skills:go-error-handling`", cardPath,
			filepath.Join(repoRoot(t), "skills") + "/<name>/SKILL.md"} {
			if !strings.Contains(msg, want) {
				t.Errorf("block without a router must name %s:\n%s", want, msg)
			}
		}
	})

	// The entry skill is the one the prompt hook picked: a refactor prompt leads
	// to go-code-refactor, not to go-code.
	t.Run("entry skill is the router the prompt picked", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		if _, out := promptEvent(t, state, "s13", t.TempDir(), "Refactor the Go package in ./dispatch so it reads better."); !strings.Contains(out, "go-code-refactor") {
			t.Fatalf("refactor prompt: stdout %q; want go-code-refactor", out)
		}
		code, msg := hookEvent(t, script, state, routingPayload("PreToolUse", "s13", "Write",
			map[string]any{"file_path": "/repo/dispatch/run.go", "content": "package dispatch\n"}))
		if code != 2 || !strings.Contains(msg, "`golang-skills:go-code-refactor`") {
			t.Fatalf("edit after a refactor prompt: exit %d, stderr %q; want 2 naming go-code-refactor", code, msg)
		}
		if strings.Contains(msg, "`golang-skills:go-code`") {
			t.Errorf("the prompt picked go-code-refactor; the block must not ask for go-code:\n%s", msg)
		}
	})

	// A reminder is not a load: reminded is only a log. A retry without a load
	// is blocked again, the third one stops the session with JSON instead of
	// endless blocks, and after the loads the edit passes.
	t.Run("a retry without loading is blocked, then the session stops", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		edit := routingPayload("PreToolUse", "s14", "Write",
			map[string]any{"file_path": "/repo/api/handler.go", "content": handler})
		if code, msg := hookEvent(t, script, state, edit); code != 2 {
			t.Fatalf("first edit: exit %d, stderr %q; want 2", code, msg)
		}
		reminded, err := os.ReadFile(filepath.Join(state, "routing", "s14", "reminded"))
		if err != nil || !strings.Contains(string(reminded), "go-style-core") {
			t.Fatalf("reminded after the first block = %q, %v; want go-style-core in it", reminded, err)
		}
		if _, err := os.Stat(filepath.Join(state, "routing", "s14", "loaded")); !os.IsNotExist(err) {
			t.Fatalf("a block must not write loaded: stat error = %v, want not exist", err)
		}

		code, msg := hookEvent(t, script, state, edit)
		if code != 2 || !strings.Contains(msg, "A reminder is not a load") || !strings.Contains(msg, "`golang-skills:go-code`") {
			t.Fatalf("retry without loading: exit %d, stderr %q; want 2 naming what is still missing", code, msg)
		}

		code, out, msg := hookOutput(t, script, state, edit)
		if code != 0 || msg != "" {
			t.Fatalf("third stalled retry: exit %d, stderr %q; want 0 with a JSON stop", code, msg)
		}
		var stop struct {
			Continue   *bool  `json:"continue"`
			StopReason string `json:"stopReason"`
			Hook       struct {
				Decision string `json:"permissionDecision"`
			} `json:"hookSpecificOutput"`
		}
		if err := json.Unmarshal([]byte(out), &stop); err != nil {
			t.Fatalf("third stalled retry: stdout %q is not JSON: %v", out, err)
		}
		if stop.Continue == nil || *stop.Continue || stop.Hook.Decision != "deny" ||
			!strings.Contains(stop.StopReason, "GOLANG_SKILLS_ROUTING_GATE=off") || !strings.Contains(stop.StopReason, "golang-skills:go-code") {
			t.Fatalf("third stalled retry = %+v; want continue false, deny, and a reason naming the skills and the off switch", stop)
		}

		for _, skill := range []string{"golang-skills:go-code", "golang-skills:go-style-core", "golang-skills:go-http", "golang-skills:go-error-handling"} {
			hookEvent(t, script, state, routingPayload("PostToolUse", "s14", "Skill", map[string]any{"skill": skill}))
		}
		hookEvent(t, script, state, routingPayload("PostToolUse", "s14", "Read",
			map[string]any{"file_path": filepath.Join(repoRoot(t), "skills", "go-style-core", "references", "CURRENT-GO.md")}))
		if code, out, msg := hookOutput(t, script, state, edit); code != 0 || out != "" || msg != "" {
			t.Fatalf("edit after the loads: exit %d, stdout %q, stderr %q; want silent 0", code, out, msg)
		}
	})

	// Progress (any new load) resets the counter, and parallel edits of
	// different files in one message are not retries.
	t.Run("progress and parallel edits do not count as stalled retries", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		edit := func(path string) map[string]any {
			return routingPayload("PreToolUse", "s15", "Write", map[string]any{"file_path": path, "content": handler})
		}
		for _, path := range []string{"/repo/api/a.go", "/repo/api/b.go", "/repo/api/c.go"} {
			if code, msg := hookEvent(t, script, state, edit(path)); code != 2 {
				t.Fatalf("parallel edit %s: exit %d, stderr %q; want 2", path, code, msg)
			}
		}
		hookEvent(t, script, state, edit("/repo/api/a.go"))
		hookEvent(t, script, state, routingPayload("PostToolUse", "s15", "Skill", map[string]any{"skill": "golang-skills:go-code"}))
		code, out, msg := hookOutput(t, script, state, edit("/repo/api/a.go"))
		if code != 2 || out != "" || strings.Contains(msg, "`golang-skills:go-code`,") {
			t.Fatalf("retry after loading go-code: exit %d, stdout %q, stderr %q; want 2, no stop, go-code no longer missing", code, out, msg)
		}
	})

	// A skill missing from this copy of the plugin is not required: the gate
	// names it as missing instead of raising blocks the model cannot satisfy.
	t.Run("a skill missing from the plugin copy is reported, not required", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		src := repoRoot(t)
		for _, rel := range []string{"hooks/go-code-routing.sh", "skills/go-code/SKILL.md", "skills/go-style-core/SKILL.md",
			"skills/go-style-core/references/CURRENT-GO.md", "skills/go-error-handling/SKILL.md"} {
			data, err := os.ReadFile(filepath.Join(src, rel))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Dir(filepath.Join(root, rel)), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, rel), data, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		copied := filepath.Join(root, "hooks", "go-code-routing.sh")
		state := t.TempDir()
		edit := routingPayload("PreToolUse", "s16", "Write",
			map[string]any{"file_path": "/repo/api/handler.go", "content": handler})
		code, msg := hookEvent(t, copied, state, edit)
		if code != 2 || !strings.Contains(msg, "Not installed in this plugin copy, so not required: `golang-skills:go-http`") {
			t.Fatalf("edit with go-http missing from the copy: exit %d, stderr %q; want 2 reporting go-http as not installed", code, msg)
		}
		if strings.Contains(msg, "but not: `golang-skills:go-http`") || strings.Contains(msg, "Missing: `golang-skills:go-code`, `golang-skills:go-style-core`, `golang-skills:go-http`") {
			t.Errorf("go-http is not installed and must not be required:\n%s", msg)
		}
		for _, skill := range []string{"go-code", "go-style-core", "go-error-handling"} {
			hookEvent(t, copied, state, routingPayload("PostToolUse", "s16", "Skill", map[string]any{"skill": skill}))
		}
		hookEvent(t, copied, state, routingPayload("PostToolUse", "s16", "Read",
			map[string]any{"file_path": filepath.Join(root, "skills", "go-style-core", "references", "CURRENT-GO.md")}))
		if code, out, msg := hookOutput(t, copied, state, edit); code != 0 || out != "" || msg != "" {
			t.Fatalf("edit with every installed skill loaded: exit %d, stdout %q, stderr %q; want silent 0", code, out, msg)
		}
	})

	t.Run("GOLANG_SKILLS_ROUTING_GATE=off turns the gate off", func(t *testing.T) {
		t.Parallel()
		code, msg := hookEventEnv(t, script, t.TempDir(), routingPayload("PreToolUse", "s17", "Write",
			map[string]any{"file_path": "/repo/api/handler.go", "content": handler}), "GOLANG_SKILLS_ROUTING_GATE=off")
		if code != 0 || msg != "" {
			t.Fatalf("gate off: exit %d, stderr %q; want silent 0", code, msg)
		}
	})

	// Without CLAUDE_PLUGIN_ROOT the hook is not wired as a plugin: the skills
	// are registered under bare names, and the gate names them the same way.
	t.Run("bare names outside a plugin", func(t *testing.T) {
		t.Parallel()
		code, msg := hookEventEnv(t, script, t.TempDir(), routingPayload("PreToolUse", "s18", "Write",
			map[string]any{"file_path": "/repo/api/handler.go", "content": handler}), "CLAUDE_PLUGIN_ROOT=")
		if code != 2 || !strings.Contains(msg, "`go-code`") || strings.Contains(msg, "golang-skills:go-") {
			t.Fatalf("gate outside a plugin: exit %d, stderr %q; want 2 naming `go-code` without a namespace", code, msg)
		}
	})

	// A refactor prompt reaches go-code-refactor alone, and a review that turns
	// into edits reaches go-code-review alone; in the 2026-09-10 and 2026-09-13
	// refactor runs such sessions loaded go-style-core in 2/20 and 6/20. The
	// gate keyed on go-code alone stayed silent for all of them.
	t.Run("blocks after go-code-refactor or go-code-review", func(t *testing.T) {
		t.Parallel()
		for _, router := range []string{"go-code-refactor", "go-code-review"} {
			state := t.TempDir()
			edit := routingPayload("PreToolUse", router, "Write",
				map[string]any{"file_path": "/repo/api/handler.go", "content": handler})
			hookEvent(t, script, state, routingPayload("PostToolUse", router, "Skill",
				map[string]any{"skill": "golang-skills:" + router}))
			code, msg := hookEvent(t, script, state, edit)
			if code != 2 {
				t.Fatalf("first Go edit after %s: exit %d, stderr %q; want 2", router, code, msg)
			}
			for _, want := range []string{router, "go-style-core", "go-http", "go-error-handling"} {
				if !strings.Contains(msg, want) {
					t.Errorf("block after %s must name %s:\n%s", router, want, msg)
				}
			}
			if code, msg := hookEvent(t, script, state, edit); code != 2 {
				t.Fatalf("retry after the %s reminder without loading: exit %d, stderr %q; want 2", router, code, msg)
			}
		}
	})

	t.Run("blocks until loaded", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		edit := routingPayload("PreToolUse", "s2", "Write",
			map[string]any{"file_path": "/repo/api/handler.go", "content": handler})

		// A plugin install names the skill "golang-skills:go-code"; the gate must
		// see through the prefix, or it never learns that go-code was loaded.
		if code, _ := hookEvent(t, script, state, routingPayload("PostToolUse", "s2", "Skill",
			map[string]any{"skill": "golang-skills:go-code"})); code != 0 {
			t.Fatalf("recording Skill go-code: exit %d, want 0", code)
		}

		code, msg := hookEvent(t, script, state, edit)
		if code != 2 {
			t.Fatalf("first Go edit after go-code: exit %d, stderr %q; want 2", code, msg)
		}
		// The card is named by the path this checkout carries it at, so the
		// model can Read it in the same message as the loads.
		cardPath := filepath.Join(repoRoot(t), "skills", "go-style-core", "references", "CURRENT-GO.md")
		for _, want := range []string{"`golang-skills:go-style-core`", "`golang-skills:go-http`", "`golang-skills:go-error-handling`", "and has not read the idiom card", cardPath} {
			if !strings.Contains(msg, want) {
				t.Errorf("block message must name %s:\n%s", want, msg)
			}
		}
		// r.Context() is HTTP plumbing, not a context.* decision; the gate must not
		// demand go-context for it.
		for _, unwanted := range []string{"go-context", "go-concurrency", "go-testing", "go-database"} {
			if strings.Contains(msg, unwanted) {
				t.Errorf("block message names %s, which the edit does not touch:\n%s", unwanted, msg)
			}
		}

		// A reminder is not a load: a retry without a load is blocked.
		if code, msg := hookEvent(t, script, state, edit); code != 2 {
			t.Fatalf("retry after one reminder without loading: exit %d, stderr %q; want 2", code, msg)
		}
	})

	// --hints runs the same owner table over whole files for the prompt hook:
	// a handler names go-http and go-error-handling, a test file go-testing,
	// a missing file nothing, and a query go-database and go-error-handling,
	// as the go-code SQL row loads both.
	t.Run("hints mode names owners for files", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		src := filepath.Join(dir, "handler.go")
		if err := os.WriteFile(src, []byte(handler), 0o644); err != nil {
			t.Fatal(err)
		}
		tf := filepath.Join(dir, "handler_test.go")
		if err := os.WriteFile(tf, []byte("package api\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		out, err := exec.Command("bash", script, "--hints", src, tf, filepath.Join(dir, "missing.go")).Output()
		if err != nil {
			t.Fatalf("--hints: %v", err)
		}
		if got, want := strings.TrimSpace(string(out)), "go-http go-error-handling go-testing"; got != want {
			t.Fatalf("--hints = %q, want %q", got, want)
		}
		db := filepath.Join(dir, "store.go")
		if err := os.WriteFile(db, []byte("package store\n\nimport \"database/sql\"\n\nfunc count(db *sql.DB) {}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		out, err = exec.Command("bash", script, "--hints", db).Output()
		if err != nil {
			t.Fatalf("--hints %s: %v", db, err)
		}
		if got, want := strings.TrimSpace(string(out)), "go-database go-error-handling"; got != want {
			t.Fatalf("--hints on a database/sql file = %q, want %q", got, want)
		}
	})

	t.Run("passes when owners are loaded", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		for _, skill := range []string{"go-code", "go-style-core", "go-error-handling"} {
			hookEvent(t, script, state, routingPayload("PostToolUse", "s3", "Skill", map[string]any{"skill": skill}))
		}
		// go-http arrives through a direct read of its SKILL.md, the Codex path.
		hookEvent(t, script, state, routingPayload("PostToolUse", "s3", "Read",
			map[string]any{"file_path": "/home/u/.claude/skills/go-http/SKILL.md"}))
		hookEvent(t, script, state, routingPayload("PostToolUse", "s3", "Read",
			map[string]any{"file_path": "/home/u/.claude/skills/go-style-core/references/CURRENT-GO.md"}))

		code, msg := hookEvent(t, script, state, routingPayload("PreToolUse", "s3", "Write",
			map[string]any{"file_path": "/repo/api/handler.go", "content": handler}))
		if code != 0 || msg != "" {
			t.Fatalf("edit with every owner loaded: exit %d, stderr %q; want silent 0", code, msg)
		}
	})

	// The idiom card is a gate item of its own: no wording of go-code made
	// Sonnet 5 medium read it (0/24 sessions on 2026-09-18). One whole Read per
	// session satisfies it; a head over it does not, since the card's older
	// rows apply at every go directive.
	t.Run("the idiom card is read whole once per session", func(t *testing.T) {
		t.Parallel()
		cardPath := filepath.Join(repoRoot(t), "skills", "go-style-core", "references", "CURRENT-GO.md")
		load := func(state, session string) {
			for _, skill := range []string{"go-code", "go-style-core", "go-http", "go-error-handling"} {
				hookEvent(t, script, state, routingPayload("PostToolUse", session, "Skill", map[string]any{"skill": skill}))
			}
		}
		edit := func(session string) map[string]any {
			return routingPayload("PreToolUse", session, "Write",
				map[string]any{"file_path": "/repo/api/handler.go", "content": handler})
		}

		// Every owner loaded, the card unread: the block names the card alone.
		state := t.TempDir()
		load(state, "c1")
		code, msg := hookEvent(t, script, state, edit("c1"))
		if code != 2 || !strings.Contains(msg, "has not read the idiom card") || !strings.Contains(msg, cardPath) {
			t.Fatalf("edit with the card unread: exit %d, stderr %q; want 2 naming the card at %s", code, msg, cardPath)
		}
		if strings.Contains(msg, "but not:") || strings.Contains(msg, "Skill call") {
			t.Errorf("every skill is loaded; the block must ask for the card only:\n%s", msg)
		}
		if code, msg := hookEvent(t, script, state, edit("c1")); code != 2 || !strings.Contains(msg, "the idiom card") {
			t.Fatalf("retry after the card reminder without a Read: exit %d, stderr %q; want 2 naming the card", code, msg)
		}

		// A Read with an offset is not the whole card.
		state = t.TempDir()
		load(state, "c2")
		hookEvent(t, script, state, routingPayload("PostToolUse", "c2", "Read",
			map[string]any{"file_path": cardPath, "offset": 40}))
		if code, msg := hookEvent(t, script, state, edit("c2")); code != 2 || !strings.Contains(msg, "idiom card") {
			t.Fatalf("edit after a partial card read: exit %d, stderr %q; want 2 naming the card", code, msg)
		}

		// A limit that covers the file is a whole read; the path is the
		// installed one, wherever the plugin lives.
		state = t.TempDir()
		load(state, "c3")
		hookEvent(t, script, state, routingPayload("PostToolUse", "c3", "Read",
			map[string]any{"file_path": cardPath, "limit": 2000}))
		if code, msg := hookEvent(t, script, state, edit("c3")); code != 0 || msg != "" {
			t.Fatalf("edit after a whole card read with a wide limit: exit %d, stderr %q; want silent 0", code, msg)
		}
		state = t.TempDir()
		load(state, "c4")
		hookEvent(t, script, state, routingPayload("PostToolUse", "c4", "Read",
			map[string]any{"file_path": "/home/u/.claude/plugins/x/skills/go-style-core/references/CURRENT-GO.md"}))
		if code, msg := hookEvent(t, script, state, edit("c4")); code != 0 || msg != "" {
			t.Fatalf("edit after a whole card read: exit %d, stderr %q; want silent 0", code, msg)
		}
	})

	t.Run("test file requires go-testing", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		hookEvent(t, script, state, routingPayload("PostToolUse", "s4", "Skill", map[string]any{"skill": "go-code"}))
		hookEvent(t, script, state, routingPayload("PostToolUse", "s4", "Skill", map[string]any{"skill": "go-style-core"}))
		hookEvent(t, script, state, routingPayload("PostToolUse", "s4", "Read",
			map[string]any{"file_path": "/home/u/.claude/skills/go-style-core/references/CURRENT-GO.md"}))
		code, msg := hookEvent(t, script, state, routingPayload("PreToolUse", "s4", "Edit",
			map[string]any{"file_path": "/repo/api/handler_test.go", "old_string": "a", "new_string": "func TestX(t *testing.T) {}"}))
		if code != 2 || !strings.Contains(msg, "go-testing") {
			t.Fatalf("_test.go edit: exit %d, stderr %q; want 2 naming go-testing", code, msg)
		}
		if strings.Contains(msg, "go-style-core") {
			t.Fatalf("go-style-core is loaded and must not be named:\n%s", msg)
		}
	})

	// The body of a test file is plumbing, not a decision: a defer on a
	// recorder, an http.NewRequest, an errors.Is on the wanted sentinel must
	// name no owner beyond go-testing.
	t.Run("test file body names go-testing only", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		for _, skill := range []string{"go-code", "go-style-core"} {
			hookEvent(t, script, state, routingPayload("PostToolUse", "s11", "Skill", map[string]any{"skill": skill}))
		}
		body := "package api\n\nfunc TestGet(t *testing.T) {\n\tsrv := NewServer(\":0\", nil)\n\tdefer srv.Close()\n\treq := http.NewRequest(http.MethodGet, \"/x\", nil)\n\tif !errors.Is(err, ErrNotFound) {\n\t\tt.Fatal(err)\n\t}\n}\n"
		code, msg := hookEvent(t, script, state, routingPayload("PreToolUse", "s11", "Write",
			map[string]any{"file_path": "/repo/api/contract_test.go", "content": body}))
		if code != 2 || !strings.Contains(msg, "go-testing") {
			t.Fatalf("_test.go write: exit %d, stderr %q; want 2 naming go-testing", code, msg)
		}
		for _, unwanted := range []string{"go-defensive", "go-http", "go-error-handling"} {
			if strings.Contains(msg, unwanted) {
				t.Errorf("test body must not name %s:\n%s", unwanted, msg)
			}
		}
	})

	t.Run("ignores non-Go files; sessions are separate", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		hookEvent(t, script, state, routingPayload("PostToolUse", "s5", "Skill", map[string]any{"skill": "go-code"}))
		if code, msg := hookEvent(t, script, state, routingPayload("PreToolUse", "s5", "Write",
			map[string]any{"file_path": "/repo/README.md", "content": "http."})); code != 0 || msg != "" {
			t.Fatalf("Markdown edit: exit %d, stderr %q; want silent 0", code, msg)
		}
		// Another session does not inherit go-code from s5: it has no router loaded.
		if code, msg := hookEvent(t, script, state, routingPayload("PreToolUse", "other", "Write",
			map[string]any{"file_path": "/repo/main.go", "content": "package main"})); code != 2 || !strings.Contains(msg, "loaded no router skill") {
			t.Fatalf("another session's Go edit: exit %d, stderr %q; want 2 without a router", code, msg)
		}
	})

	// The hints are heuristics for decision-bearing forms. Routine syntax that
	// go-code's router says does not trigger a load — fmt.Errorf with %v, an
	// http constant in a comment — must leave the gate silent.
	t.Run("routine syntax names no owner", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		for _, skill := range []string{"go-code", "go-style-core"} {
			hookEvent(t, script, state, routingPayload("PostToolUse", "s6", "Skill", map[string]any{"skill": skill}))
		}
		hookEvent(t, script, state, routingPayload("PostToolUse", "s6", "Read",
			map[string]any{"file_path": "/home/u/.claude/skills/go-style-core/references/CURRENT-GO.md"}))
		routine := "package store\n\n// Save maps a miss to http.StatusOK.\nfunc (s *Store) Save() error {\n\treturn fmt.Errorf(\"x: %v\", err)\n}\n"
		code, msg := hookEvent(t, script, state, routingPayload("PreToolUse", "s6", "Write",
			map[string]any{"file_path": "/repo/store/save.go", "content": routine}))
		if code != 0 || msg != "" {
			t.Fatalf("fmt.Errorf(%%v) and http.StatusOK in a comment: exit %d, stderr %q; want silent 0", code, msg)
		}
	})

	// make and append appear in nearly every Go body; a forced
	// go-data-structures load was what started the make+copy -> slices.Clone
	// rewrite in the 2026-09-10 sessions that returned null for a nil list.
	t.Run("collections are routine syntax", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		for _, skill := range []string{"go-code", "go-style-core"} {
			hookEvent(t, script, state, routingPayload("PostToolUse", "s10", "Skill", map[string]any{"skill": skill}))
		}
		hookEvent(t, script, state, routingPayload("PostToolUse", "s10", "Read",
			map[string]any{"file_path": "/home/u/.claude/skills/go-style-core/references/CURRENT-GO.md"}))
		collections := "package store\n\nfunc (s *Store) IDs() []string {\n\tids := make([]string, 0, len(s.m))\n\tseen := make(map[string]struct{})\n\tfor id := range s.m {\n\t\tids = append(ids, id)\n\t}\n\treturn ids\n}\n"
		code, msg := hookEvent(t, script, state, routingPayload("PreToolUse", "s10", "Write",
			map[string]any{"file_path": "/repo/store/ids.go", "content": collections}))
		if code != 0 || msg != "" {
			t.Fatalf("make/append/make(map): exit %d, stderr %q; want silent 0", code, msg)
		}
	})

	// A plain defer and a ctx parameter passed on appear in nearly every body;
	// the 1.7.0 hints for them named go-defensive in 16 of 73 implement gate
	// blocks on 2026-09-30.
	t.Run("plain defer and a ctx parameter are routine syntax", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		for _, skill := range []string{"go-code", "go-style-core"} {
			hookEvent(t, script, state, routingPayload("PostToolUse", "s12", "Skill", map[string]any{"skill": skill}))
		}
		hookEvent(t, script, state, routingPayload("PostToolUse", "s12", "Read",
			map[string]any{"file_path": "/home/u/.claude/skills/go-style-core/references/CURRENT-GO.md"}))
		routine := "package store\n\nfunc (s *Store) Read(ctx context.Context, f *os.File) error {\n\ts.mu.Lock()\n\tdefer s.mu.Unlock()\n\tdefer f.Close()\n\treturn s.load(ctx, f)\n}\n"
		code, msg := hookEvent(t, script, state, routingPayload("PreToolUse", "s12", "Write",
			map[string]any{"file_path": "/repo/store/read.go", "content": routine}))
		if code != 0 || msg != "" {
			t.Fatalf("defer Unlock/Close and a ctx parameter: exit %d, stderr %q; want silent 0", code, msg)
		}
	})

	t.Run("decision-bearing forms name their owner", func(t *testing.T) {
		t.Parallel()
		cases := []struct {
			name, session, path, content, owner string
		}{
			{"type parameter list", "s7", "/repo/x/map.go",
				"package x\n\nfunc Map[T any](xs []T) []T { return xs }\n", "go-generics"},
			{"pgxpool", "s8", "/repo/x/db.go",
				"package x\n\nfunc open(dsn string) (*pgxpool.Pool, error) {\n\treturn pgxpool.New(ctx, dsn)\n}\n", "go-database"},
			{"deferred closure with recover", "s9", "/repo/x/run.go",
				"package x\n\nfunc run() (err error) {\n\tdefer func() {\n\t\tif r := recover(); r != nil {\n\t\t\terr = fmt.Errorf(\"panic: %v\", r)\n\t\t}\n\t}()\n\treturn work()\n}\n", "go-defensive"},
			{"context stored in a struct", "s11", "/repo/x/worker.go",
				"package x\n\ntype Worker struct {\n\tctx  context.Context\n\tjobs []Job\n}\n", "go-context"},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				state := t.TempDir()
				hookEvent(t, script, state, routingPayload("PostToolUse", tc.session, "Skill", map[string]any{"skill": "go-code"}))
				code, msg := hookEvent(t, script, state, routingPayload("PreToolUse", tc.session, "Write",
					map[string]any{"file_path": tc.path, "content": tc.content}))
				if code != 2 || !strings.Contains(msg, tc.owner) {
					t.Fatalf("%s edit: exit %d, stderr %q; want 2 naming %s", tc.name, code, msg, tc.owner)
				}
			})
		}
	})
}

// TestVetHook drives the PostToolUse vet hook against throwaway modules:
// gofmt, go vet, and go fix -diff findings reach stderr with exit 2; a clean
// file and a non-Go path stay silent with exit 0.
func TestVetHook(t *testing.T) {
	t.Parallel()
	script := filepath.Join(repoRoot(t), "hooks", "go-vet-on-edit.sh")

	// module writes a one-file module in a fresh temp dir and returns the
	// absolute path of main.go, the file the payload names.
	module := func(t *testing.T, src string) string {
		t.Helper()
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module scratch\n\ngo 1.27\n"), 0o644); err != nil {
			t.Fatalf("write go.mod: %v", err)
		}
		path := filepath.Join(dir, "main.go")
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatalf("write main.go: %v", err)
		}
		return path
	}
	edited := func(path string) map[string]any {
		return routingPayload("PostToolUse", "v1", "Write", map[string]any{"file_path": path})
	}
	clean := "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hi\")\n}\n"

	t.Run("unformatted file", func(t *testing.T) {
		t.Parallel()
		path := module(t, "package main\n\nfunc main() {\n  x := 1\n_ = x\n}\n")
		code, msg := hookEvent(t, script, t.TempDir(), edited(path))
		if code != 2 || !strings.Contains(msg, "gofmt:") {
			t.Fatalf("unformatted file: exit %d, stderr %q; want 2 naming gofmt:", code, msg)
		}
	})

	t.Run("clean file", func(t *testing.T) {
		t.Parallel()
		path := module(t, clean)
		if code, msg := hookEvent(t, script, t.TempDir(), edited(path)); code != 0 || msg != "" {
			t.Fatalf("clean file: exit %d, stderr %q; want silent 0", code, msg)
		}
	})

	t.Run("non-Go path", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), "README.md")
		if err := os.WriteFile(path, []byte("  not gofmt material\n"), 0o644); err != nil {
			t.Fatalf("write README.md: %v", err)
		}
		if code, msg := hookEvent(t, script, t.TempDir(), edited(path)); code != 0 || msg != "" {
			t.Fatalf("Markdown edit: exit %d, stderr %q; want silent 0", code, msg)
		}
		missing := filepath.Join(t.TempDir(), "gone.go")
		if code, msg := hookEvent(t, script, t.TempDir(), edited(missing)); code != 0 || msg != "" {
			t.Fatalf("deleted .go file: exit %d, stderr %q; want silent 0", code, msg)
		}
	})

	// go fix -diff on go1.27 rewrites a counted loop to range-over-int; the
	// file is gofmt-clean and vets clean, so the diff is the only finding.
	t.Run("modernizable loop", func(t *testing.T) {
		t.Parallel()
		path := module(t, "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tn := 3\n\tfor i := 0; i < n; i++ {\n\t\tfmt.Println(i)\n\t}\n}\n")
		code, msg := hookEvent(t, script, t.TempDir(), edited(path))
		if code != 2 || !strings.Contains(msg, "go fix -diff") || !strings.Contains(msg, "range n") {
			t.Fatalf("pre-1.22 loop: exit %d, stderr %q; want 2 with a go fix -diff section rewriting to range n", code, msg)
		}
		for _, unwanted := range []string{"gofmt:", "go vet"} {
			if strings.Contains(msg, unwanted) {
				t.Errorf("clean-but-modernizable file must not report %s:\n%s", unwanted, msg)
			}
		}
	})

	t.Run("go fix hunks elsewhere in the package are one count", func(t *testing.T) {
		t.Parallel()
		path := module(t, clean)
		loop := "package main\n\nimport \"fmt\"\n\nfunc count() {\n\tn := 3\n\tfor i := 0; i < n; i++ {\n\t\tfmt.Println(i)\n\t}\n}\n"
		if err := os.WriteFile(filepath.Join(filepath.Dir(path), "other.go"), []byte(loop), 0o644); err != nil {
			t.Fatalf("write other.go: %v", err)
		}
		code, msg := hookEventEnv(t, script, t.TempDir(), edited(path), "GOLANG_SKILLS_EDIT_LINT=off")
		if code != 2 || !strings.Contains(msg, "1 hunk(s) in other files") || strings.Contains(msg, "range n") {
			t.Fatalf("modernizable other.go: exit %d, stderr %q; want 2 with a hunk count and no hunk text", code, msg)
		}
	})

	t.Run("vet failure", func(t *testing.T) {
		t.Parallel()
		path := module(t, "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Printf(\"%d\\n\", \"s\")\n}\n")
		code, msg := hookEvent(t, script, t.TempDir(), edited(path))
		if code != 2 || !strings.Contains(msg, "go vet") {
			t.Fatalf("printf misuse: exit %d, stderr %q; want 2 naming go vet", code, msg)
		}
	})

	// A compile error is reported once, by go vet; go fix would restate it.
	t.Run("compile error reported once", func(t *testing.T) {
		t.Parallel()
		path := module(t, "package main\n\nfunc main() {\n\tx := undefined\n}\n")
		code, msg := hookEvent(t, script, t.TempDir(), edited(path))
		if code != 2 || !strings.Contains(msg, "go vet") {
			t.Fatalf("compile error: exit %d, stderr %q; want 2 naming go vet", code, msg)
		}
		if strings.Contains(msg, "go fix -diff") {
			t.Errorf("go fix must not restate a compile error go vet already reported:\n%s", msg)
		}
	})

	// The payload is JSON: a "file_path" string inside the written content
	// must not redirect the hook away from tool_input.file_path.
	t.Run("parses the payload as JSON", func(t *testing.T) {
		t.Parallel()
		path := module(t, "package main\n\nfunc main() {\n  x := 1\n_ = x\n}\n")
		payload := edited(path)
		payload["tool_input"].(map[string]any)["content"] = "// {\"file_path\": \"/decoy/other.go\"}\n"
		code, msg := hookEvent(t, script, t.TempDir(), payload)
		if code != 2 || !strings.Contains(msg, "gofmt: "+path) {
			t.Fatalf("payload with a decoy path in content: exit %d, stderr %q; want 2 naming %s", code, msg, path)
		}
	})

	// The tests and the linter run only for a package that type-checks, so a
	// session without a shell still sees what the gate would have said.
	withTest := func(t *testing.T, src, test string) string {
		t.Helper()
		path := module(t, src)
		if err := os.WriteFile(filepath.Join(filepath.Dir(path), "main_test.go"), []byte(test), 0o644); err != nil {
			t.Fatalf("write main_test.go: %v", err)
		}
		return path
	}
	failing := "package main\n\nimport \"testing\"\n\nfunc TestMain2(t *testing.T) {\n\tt.Errorf(\"main() = 1, want 2\")\n}\n"

	t.Run("failing package test", func(t *testing.T) {
		t.Parallel()
		path := withTest(t, clean, failing)
		code, msg := hookEvent(t, script, t.TempDir(), edited(path))
		if code != 2 || !strings.Contains(msg, "go test") || !strings.Contains(msg, "main() = 1, want 2") {
			t.Fatalf("failing test: exit %d, stderr %q; want 2 with the go test failure", code, msg)
		}
	})

	t.Run("a test timeout prints the running tests, not the goroutine dump", func(t *testing.T) {
		t.Parallel()
		slow := "package main\n\nimport (\n\t\"testing\"\n\t\"time\"\n)\n\nfunc TestSlow(t *testing.T) {\n\ttime.Sleep(5 * time.Second)\n}\n"
		path := withTest(t, clean, slow)
		code, msg := hookEventEnv(t, script, t.TempDir(), edited(path), "GOLANG_SKILLS_EDIT_LINT=off", "GOLANG_SKILLS_EDIT_TEST_TIMEOUT=1")
		if code != 2 || !strings.Contains(msg, "panic: test timed out after 1s") || !strings.Contains(msg, "TestSlow") || !strings.Contains(msg, "goroutine dump cut") {
			t.Fatalf("slow test: exit %d, stderr %q; want 2 naming the timeout and the running test", code, msg)
		}
		if strings.Contains(msg, "testing.tRunner") || strings.Contains(msg, "goroutine 1 [") {
			t.Errorf("a timeout must not print the goroutine dump:\n%s", msg)
		}
	})

	t.Run("package tests can be switched off", func(t *testing.T) {
		t.Parallel()
		path := withTest(t, clean, failing)
		code, msg := hookEventEnv(t, script, t.TempDir(), edited(path), "GOLANG_SKILLS_EDIT_TESTS=off", "GOLANG_SKILLS_EDIT_LINT=off")
		if code != 0 || msg != "" {
			t.Fatalf("tests off: exit %d, stderr %q; want silent 0", code, msg)
		}
	})

	t.Run("passing package test stays silent", func(t *testing.T) {
		t.Parallel()
		path := withTest(t, clean, "package main\n\nimport \"testing\"\n\nfunc TestMain2(t *testing.T) {}\n")
		if code, msg := hookEventEnv(t, script, t.TempDir(), edited(path), "GOLANG_SKILLS_EDIT_LINT=off"); code != 0 || msg != "" {
			t.Fatalf("passing test: exit %d, stderr %q; want silent 0", code, msg)
		}
	})

	bareWrite := "package main\n\nimport \"net/http\"\n\nfunc handle(w http.ResponseWriter, r *http.Request) {\n\tw.Write([]byte(\"ok\"))\n}\n\nfunc main() { http.HandleFunc(\"/\", handle) }\n"
	needLint := func(t *testing.T) {
		t.Helper()
		if _, err := exec.LookPath("golangci-lint"); err != nil {
			t.Skip("golangci-lint not installed")
		}
	}

	t.Run("lint finding in the edited file", func(t *testing.T) {
		t.Parallel()
		needLint(t)
		path := module(t, bareWrite)
		code, msg := hookEvent(t, script, t.TempDir(), edited(path))
		if code != 2 || !strings.Contains(msg, "golangci-lint") || !strings.Contains(msg, "errcheck") || !strings.Contains(msg, path+":6:") {
			t.Fatalf("bare w.Write: exit %d, stderr %q; want 2 with an errcheck finding at %s:6", code, msg, path)
		}
	})

	t.Run("lint findings elsewhere in the package are one count", func(t *testing.T) {
		t.Parallel()
		needLint(t)
		path := module(t, "package main\n\nfunc main() {}\n")
		other := strings.Replace(bareWrite, "func main() { http.HandleFunc(\"/\", handle) }\n", "func init() { http.HandleFunc(\"/\", handle) }\n", 1)
		if err := os.WriteFile(filepath.Join(filepath.Dir(path), "other.go"), []byte(other), 0o644); err != nil {
			t.Fatalf("write other.go: %v", err)
		}
		code, msg := hookEvent(t, script, t.TempDir(), edited(path))
		if code != 2 || !strings.Contains(msg, "1 finding(s) in other files") || strings.Contains(msg, "errcheck") {
			t.Fatalf("finding in other.go: exit %d, stderr %q; want 2 with a count and no finding text", code, msg)
		}
	})

	t.Run("in a git checkout only lint issues new since HEAD count", func(t *testing.T) {
		t.Parallel()
		needLint(t)
		if _, err := exec.LookPath("git"); err != nil {
			t.Skip("git not installed")
		}
		path := module(t, bareWrite)
		git := func(args ...string) {
			t.Helper()
			cmd := exec.Command("git", append([]string{"-c", "user.email=t@example.com", "-c", "user.name=t", "-c", "commit.gpgsign=false"}, args...)...)
			cmd.Dir = filepath.Dir(path)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("git %v: %v\n%s", args, err, out)
			}
		}
		git("init", "-q")
		git("add", ".")
		git("commit", "-q", "-m", "older debt")
		src := bareWrite + "\nfunc handle2(w http.ResponseWriter, _ *http.Request) {\n\tw.Write([]byte(\"two\"))\n}\n"
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatalf("write main.go: %v", err)
		}
		code, msg := hookEvent(t, script, t.TempDir(), edited(path))
		if code != 2 || !strings.Contains(msg, "errcheck") || !strings.Contains(msg, path+":12:") {
			t.Fatalf("new bare write: exit %d, stderr %q; want 2 with an errcheck finding at %s:12", code, msg, path)
		}
		if strings.Contains(msg, path+":6:") {
			t.Errorf("the committed finding at line 6 predates the session and must not print:\n%s", msg)
		}
	})

	t.Run("repository lint configuration wins", func(t *testing.T) {
		t.Parallel()
		needLint(t)
		path := module(t, bareWrite)
		cfg := "version: \"2\"\nlinters:\n  default: none\n  enable:\n    - govet\n"
		if err := os.WriteFile(filepath.Join(filepath.Dir(path), ".golangci.yml"), []byte(cfg), 0o644); err != nil {
			t.Fatalf("write .golangci.yml: %v", err)
		}
		if code, msg := hookEvent(t, script, t.TempDir(), edited(path)); code != 0 || msg != "" {
			t.Fatalf("repository config without errcheck: exit %d, stderr %q; want silent 0", code, msg)
		}
	})
}

// hookEventEnv is hookEvent with extra environment variables for the hook.
func hookEventEnv(t *testing.T, script, state string, payload map[string]any, env ...string) (int, string) {
	t.Helper()
	code, _, stderr := hookOutput(t, script, state, payload, env...)
	return code, stderr
}

// promptEvent runs the UserPromptSubmit hook and returns its exit code and
// stdout, which is what the host adds to the model's context. Like hookOutput
// it sets CLAUDE_PLUGIN_ROOT to the checkout; env comes last.
func promptEvent(t *testing.T, state, session, cwd, prompt string, env ...string) (int, string) {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"hook_event_name": "UserPromptSubmit",
		"session_id":      session,
		"cwd":             cwd,
		"prompt":          prompt,
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	cmd := exec.Command("bash", filepath.Join(repoRoot(t), "hooks", "go-prompt-routing.sh"))
	cmd.Stdin = strings.NewReader(string(body))
	cmd.Env = append(append(os.Environ(), "CLAUDE_PLUGIN_ROOT="+repoRoot(t), "CLAUDE_PLUGIN_DATA="+state), env...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if stderr.Len() > 0 {
		t.Errorf("prompt hook wrote to stderr, which a UserPromptSubmit hook must not:\n%s", stderr.String())
	}
	if err == nil {
		return 0, stdout.String()
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("run prompt hook: %v", err)
	}
	return exitErr.ExitCode(), stdout.String()
}

// TestPromptRouting drives the UserPromptSubmit hook: the two corpus prompts
// name their router, a prompt without Go and a read-only Go question stay
// silent, a router named in the prompt is selected rather than silenced,
// Ukrainian and Russian wording route like English, the note is printed once
// per skill per session with the plugin's exact skill names, and a session
// that already loaded the skill is left alone.
func TestPromptRouting(t *testing.T) {
	t.Parallel()
	const implement = "Implement the Go package in ./feed. Every exported declaration is already there with its documentation; write the bodies so the package does what the documentation says. Do not change the exported signatures. Apply the changes to the files."
	const refactor = "Refactor the Go package in ./dispatch so it reads better. Keep observable behavior identical: the exported API, error texts, and rendered output must not change. Apply the changes to the files."

	// goRepo is a directory holding a Go module two levels down, for prompts
	// that do not name Go themselves.
	goRepo := func(t *testing.T) string {
		t.Helper()
		dir := t.TempDir()
		pkg := filepath.Join(dir, "internal", "feed")
		if err := os.MkdirAll(pkg, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(pkg, "feed.go"), []byte("package feed\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		return dir
	}

	t.Run("implement prompt names go-code", func(t *testing.T) {
		t.Parallel()
		code, out := promptEvent(t, t.TempDir(), "p1", t.TempDir(), implement)
		if code != 0 || !strings.Contains(out, "`golang-skills:go-code`") {
			t.Fatalf("implement prompt: exit %d, stdout %q; want 0 naming go-code", code, out)
		}
		if strings.Contains(out, "go-code-refactor") {
			t.Fatalf("implement prompt must not name go-code-refactor:\n%s", out)
		}
		// The idiom card by its installed path, for a Read in the same message
		// as the go-style-core load; the gate asks for it otherwise.
		cardPath := filepath.Join(repoRoot(t), "skills", "go-style-core", "references", "CURRENT-GO.md")
		if !strings.Contains(out, "Read the idiom card whole") || !strings.Contains(out, cardPath) {
			t.Fatalf("note must name the idiom card at %s:\n%s", cardPath, out)
		}
	})

	t.Run("new code names go-testing without a condition", func(t *testing.T) {
		t.Parallel()
		_, out := promptEvent(t, t.TempDir(), "p21", t.TempDir(), implement)
		if !strings.Contains(out, "`golang-skills:go-testing`, since new code starts with its contract test") || strings.Contains(out, "if you write or edit a test") {
			t.Fatalf("implement prompt must name go-testing unconditionally:\n%s", out)
		}
		_, out = promptEvent(t, t.TempDir(), "p22", t.TempDir(), "Fix the Go bug in ./feed where the totals come out wrong")
		if !strings.Contains(out, "`golang-skills:go-testing` if you write or edit a test") {
			t.Fatalf("fix prompt keeps the condition on go-testing:\n%s", out)
		}
	})

	t.Run("refactor prompt names go-code-refactor", func(t *testing.T) {
		t.Parallel()
		code, out := promptEvent(t, t.TempDir(), "p2", t.TempDir(), refactor)
		if code != 0 || !strings.Contains(out, "`golang-skills:go-code-refactor`") {
			t.Fatalf("refactor prompt: exit %d, stdout %q; want 0 naming go-code-refactor", code, out)
		}
	})

	// The review corpus prompt says "what is wrong, and the fix": `fix` used to
	// send it to go-code with the condition "before the first edit", which a
	// review never reaches. The review note names go-code-review alone, with no
	// card and no owner list.
	t.Run("review prompt names go-code-review alone", func(t *testing.T) {
		t.Parallel()
		prompt := "Review the Go package in ./orders as a pull request reviewer would. Report every defect you find " +
			"with its file and line, its severity, what is wrong, and the fix. Do not modify any file."
		code, out := promptEvent(t, t.TempDir(), "p20", t.TempDir(), prompt)
		if code != 0 || !strings.Contains(out, "`golang-skills:go-code-review`") || !strings.Contains(out, "Before the first finding") {
			t.Fatalf("review prompt: exit %d, stdout %q; want 0 naming go-code-review before the first finding", code, out)
		}
		for _, unwanted := range []string{"`golang-skills:go-code`", "before the first edit", "CURRENT-GO.md", "go-style-core"} {
			if strings.Contains(out, unwanted) {
				t.Errorf("review note names %q:\n%s", unwanted, out)
			}
		}
	})

	t.Run("clean-up wording is a refactor", func(t *testing.T) {
		t.Parallel()
		_, out := promptEvent(t, t.TempDir(), "p3", t.TempDir(), "This Go file is messy, clean it up")
		if !strings.Contains(out, "`golang-skills:go-code-refactor`") {
			t.Fatalf("messy/clean up: stdout %q; want go-code-refactor", out)
		}
	})

	t.Run("monolith wording is a refactor", func(t *testing.T) {
		t.Parallel()
		_, out := promptEvent(t, t.TempDir(), "p13", t.TempDir(), "Our Go monolith has one models package that every other package imports — propose how to modularize it")
		if !strings.Contains(out, "`golang-skills:go-code-refactor`") {
			t.Fatalf("monolith/modularize: stdout %q; want go-code-refactor", out)
		}
	})

	t.Run("silent without Go", func(t *testing.T) {
		t.Parallel()
		code, out := promptEvent(t, t.TempDir(), "p4", t.TempDir(), "Write a Python script that parses this CSV and prints the totals")
		if code != 0 || out != "" {
			t.Fatalf("non-Go prompt in a non-Go directory: exit %d, stdout %q; want silent 0", code, out)
		}
	})

	t.Run("read-only Go question stays silent", func(t *testing.T) {
		t.Parallel()
		state, cwd := t.TempDir(), goRepo(t)
		for _, prompt := range []string{
			"Explain what this Go function does and why it uses a mutex",
			"Where is CreateUser implemented in Go?",
		} {
			if code, out := promptEvent(t, state, "p5", cwd, prompt); code != 0 || out != "" {
				t.Fatalf("question %q: exit %d, stdout %q; want silent 0", prompt, code, out)
			}
		}
	})

	t.Run("directory with Go and a code noun fires", func(t *testing.T) {
		t.Parallel()
		_, out := promptEvent(t, t.TempDir(), "p6", goRepo(t), "Add a handler that returns the account balance as JSON")
		if !strings.Contains(out, "`golang-skills:go-code`") {
			t.Fatalf("handler in a Go directory: stdout %q; want go-code", out)
		}
	})

	t.Run("directory with Go but no code noun stays silent", func(t *testing.T) {
		t.Parallel()
		code, out := promptEvent(t, t.TempDir(), "p7", goRepo(t), "Add a line to the README about the release schedule")
		if code != 0 || out != "" {
			t.Fatalf("README edit in a Go directory: exit %d, stdout %q; want silent 0", code, out)
		}
	})

	// A router named in the text is selected: such a prompt used to turn the
	// note off, and the loads were left to the model's choice. A modifier in a
	// host command (/opsx:apply add-auth /go-code) loads nothing by itself either.
	t.Run("a router named in the prompt is selected, not silenced", func(t *testing.T) {
		t.Parallel()
		for _, tc := range []struct{ prompt, want string }{
			{"$go-code-refactor " + refactor, "`golang-skills:go-code-refactor`"},
			{"use the go-code skill: " + implement, "`golang-skills:go-code`"},
			{"/opsx:apply add-auth /go-code", "`golang-skills:go-code`"},
			{"Используй golang-skills:go-code-refactor для пакета dispatch", "`golang-skills:go-code-refactor`"},
		} {
			code, out := promptEvent(t, t.TempDir(), "p8", t.TempDir(), tc.prompt)
			if code != 0 || !strings.Contains(out, "names the "+tc.want+" skill") || !strings.Contains(out, "Skill tool, name "+tc.want) {
				t.Errorf("prompt %q: exit %d, stdout %q; want a note selecting %s", tc.prompt, code, out, tc.want)
			}
		}
		// A path to a skill file is not a mention: a prompt about Markdown
		// stays silent.
		if code, out := promptEvent(t, t.TempDir(), "p8", t.TempDir(), "Fix the routing table in skills/go-code/SKILL.md"); code != 0 || out != "" {
			t.Errorf("a skill path in the prompt: exit %d, stdout %q; want silent 0", code, out)
		}

		// A mention alone loads nothing: the gate blocks the first edit and asks
		// for exactly the named router.
		state := t.TempDir()
		promptEvent(t, state, "p17", t.TempDir(), "$go-code-refactor "+refactor)
		gate := filepath.Join(repoRoot(t), "hooks", "go-code-routing.sh")
		code, msg := hookEvent(t, gate, state, routingPayload("PreToolUse", "p17", "Write",
			map[string]any{"file_path": "/repo/dispatch/run.go", "content": "package dispatch\n"}))
		if code != 2 || !strings.Contains(msg, "`golang-skills:go-code-refactor`") {
			t.Fatalf("edit after a prompt that only names the router: exit %d, stderr %q; want 2 naming go-code-refactor", code, msg)
		}
	})

	// A slash command is expanded by the host: it inserts the router SKILL.md
	// and calls no tool, so PostToolUse never fires and go-code-routing.sh
	// records no load — the edit gate, which needs a router in `loaded`, would
	// stay silent for the whole session. UserPromptSubmit is the only event a
	// slash invocation raises, so this hook records the router itself.
	t.Run("slash command records the router and arms the gate", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		cwd := t.TempDir()
		pkg := filepath.Join(cwd, "dispatch")
		if err := os.MkdirAll(pkg, 0o755); err != nil {
			t.Fatal(err)
		}
		src := "package dispatch\n\nfunc Run(ctx context.Context) error {\n\tctx, cancel := context.WithTimeout(ctx, time.Second)\n\tdefer cancel()\n\treturn fmt.Errorf(\"run: %w\", run(ctx))\n}\n"
		if err := os.WriteFile(filepath.Join(pkg, "dispatch.go"), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}

		code, out := promptEvent(t, state, "p16", cwd, "/golang-skills:go-code Реалізуй пакет ./dispatch")
		if code != 0 {
			t.Fatalf("slash invocation: exit %d; want 0", code)
		}
		for _, want := range []string{"`/go-code`", "`golang-skills:go-style-core`", "`golang-skills:go-context`", "`golang-skills:go-error-handling`", "in one message", "CURRENT-GO.md"} {
			if !strings.Contains(out, want) {
				t.Errorf("slash note must name %s:\n%s", want, out)
			}
		}
		if strings.Contains(out, "Skill tool, name") {
			t.Errorf("the skill is already in context; the note must not ask for it again:\n%s", out)
		}

		// The gate now sees a router for this session and blocks the first edit.
		gate := filepath.Join(repoRoot(t), "hooks", "go-code-routing.sh")
		edit := routingPayload("PreToolUse", "p16", "Edit", map[string]any{
			"file_path":  filepath.Join(pkg, "dispatch.go"),
			"new_string": "func run(ctx context.Context) error { return ctx.Err() }",
		})
		blockCode, msg := hookEvent(t, gate, state, edit)
		if blockCode != 2 || !strings.Contains(msg, "go-style-core") {
			t.Fatalf("edit after a slash invocation: exit %d, stderr %q; want 2 naming go-style-core", blockCode, msg)
		}

		// The router is recorded, so a second slash in the same session is silent.
		if _, out := promptEvent(t, state, "p16", cwd, "/go-code ще раз"); out != "" {
			t.Errorf("second slash invocation in the same session: stdout %q; want silent", out)
		}
	})

	t.Run("slash command for the other routers", func(t *testing.T) {
		t.Parallel()
		for _, router := range []string{"go-code-refactor", "go-code-review"} {
			state := t.TempDir()
			_, out := promptEvent(t, state, router, t.TempDir(), "/"+router+" ./dispatch")
			if !strings.Contains(out, "`/"+router+"`") || !strings.Contains(out, "`golang-skills:go-style-core`") {
				t.Fatalf("slash %s: stdout %q; want the note naming the command and go-style-core", router, out)
			}
			// A review writes nothing, so its note names no card; a refactor does.
			if wantCard := router == "go-code-refactor"; strings.Contains(out, "CURRENT-GO.md") != wantCard {
				t.Errorf("slash %s names the card = %v, want %v:\n%s", router, !wantCard, wantCard, out)
			}
			gate := filepath.Join(repoRoot(t), "hooks", "go-code-routing.sh")
			code, msg := hookEvent(t, gate, state, routingPayload("PreToolUse", router, "Write",
				map[string]any{"file_path": "/repo/api/handler.go", "content": "package api\n"}))
			if code != 2 {
				t.Fatalf("edit after slash %s: exit %d, stderr %q; want 2", router, code, msg)
			}
		}
	})

	t.Run("once per skill per session", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		if _, out := promptEvent(t, state, "p9", t.TempDir(), implement); out == "" {
			t.Fatal("first prompt: want the note")
		}
		if _, out := promptEvent(t, state, "p9", t.TempDir(), implement); out != "" {
			t.Fatalf("second prompt in the same session: stdout %q; want silent", out)
		}
		// A refactor later in the same session still gets its own note once.
		if _, out := promptEvent(t, state, "p9", t.TempDir(), refactor); !strings.Contains(out, "`golang-skills:go-code-refactor`") {
			t.Fatalf("refactor after implement: stdout %q; want go-code-refactor", out)
		}
		if _, out := promptEvent(t, state, "other", t.TempDir(), implement); out == "" {
			t.Fatal("another session: want its own note")
		}
	})

	t.Run("silent when the session already loaded the skill", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		hookEvent(t, filepath.Join(repoRoot(t), "hooks", "go-code-routing.sh"), state,
			routingPayload("PostToolUse", "p10", "Skill", map[string]any{"skill": "golang-skills:go-code"}))
		if code, out := promptEvent(t, state, "p10", t.TempDir(), implement); code != 0 || out != "" {
			t.Fatalf("go-code already loaded: exit %d, stdout %q; want silent 0", code, out)
		}
	})

	// The owners come from the gate's own table, read off the files the prompt
	// names, so a session can load everything before its first edit instead of
	// meeting the gate once per owner (five blocks in three sessions on
	// 2026-09-13). go-style-core is always named; test files stay out of the
	// scan; routine code names no owner.
	t.Run("names go-style-core and the owners the target's code points at", func(t *testing.T) {
		t.Parallel()
		cwd := t.TempDir()
		pkg := filepath.Join(cwd, "dispatch")
		if err := os.MkdirAll(pkg, 0o755); err != nil {
			t.Fatal(err)
		}
		src := "package dispatch\n\nfunc Run(ctx context.Context) error {\n\tctx, cancel := context.WithTimeout(ctx, time.Second)\n\tdefer cancel()\n\treturn fmt.Errorf(\"run: %w\", run(ctx))\n}\n"
		if err := os.WriteFile(filepath.Join(pkg, "dispatch.go"), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(pkg, "dispatch_test.go"), []byte("package dispatch\n\nimport \"net/http\"\n\nvar _ = http.StatusOK\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		_, out := promptEvent(t, t.TempDir(), "p13", cwd, refactor)
		for _, want := range []string{"`golang-skills:go-code-refactor`", "`golang-skills:go-style-core`", "`golang-skills:go-error-handling`", "`golang-skills:go-context`", "before the first edit"} {
			if !strings.Contains(out, want) {
				t.Errorf("note must name %s:\n%s", want, out)
			}
		}
		if strings.Contains(out, "`golang-skills:go-http`") {
			t.Errorf("a test file's imports must not name an owner:\n%s", out)
		}
		if strings.Contains(out, "`golang-skills:go-defensive`") {
			t.Errorf("defer cancel() is routine; it must not name go-defensive:\n%s", out)
		}
		// A bare package name in the prompt resolves against cwd too.
		if _, out := promptEvent(t, t.TempDir(), "p14", cwd, "Спрости Go-пакет dispatch, не змінюючи поведінки"); !strings.Contains(out, "`golang-skills:go-context`") {
			t.Errorf("bare package name: stdout %q; want go-context", out)
		}
		plain := filepath.Join(cwd, "plain")
		if err := os.MkdirAll(plain, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(plain, "plain.go"), []byte("package plain\n\nfunc Add(a, b int) int { return a + b }\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		_, out = promptEvent(t, t.TempDir(), "p15", cwd, "Refactor the Go package in ./plain so it reads better.")
		if strings.Contains(out, "owners its code points at") || !strings.Contains(out, "`golang-skills:go-style-core`") {
			t.Errorf("routine code must name go-style-core and no owner:\n%s", out)
		}
	})

	t.Run("ukrainian wording", func(t *testing.T) {
		t.Parallel()
		if _, out := promptEvent(t, t.TempDir(), "p11", t.TempDir(), "Реалізуй Go-пакет у ./catalog за документацією"); !strings.Contains(out, "`golang-skills:go-code`") {
			t.Fatalf("Ukrainian implement: stdout %q; want go-code", out)
		}
		if _, out := promptEvent(t, t.TempDir(), "p12", t.TempDir(), "Спрости цей Go-пакет, не змінюючи поведінки"); !strings.Contains(out, "`golang-skills:go-code-refactor`") {
			t.Fatalf("Ukrainian simplify: stdout %q; want go-code-refactor", out)
		}
	})

	t.Run("russian wording", func(t *testing.T) {
		t.Parallel()
		cases := []struct {
			prompt, cwd, want string
		}{
			{"Реализуй пакет на Go в ./catalog по документации", t.TempDir(), "`golang-skills:go-code`"},
			{"Упрости этот Go-пакет, не меняя поведения", t.TempDir(), "`golang-skills:go-code-refactor`"},
			{"Используй go-code и почини обработчик заказов", t.TempDir(), "`golang-skills:go-code`"},
			// Without the word Go: a directory of Go code and a Russian code noun.
			{"Добавь обработчик, который возвращает баланс счёта в JSON", goRepo(t), "`golang-skills:go-code`"},
			{"Почини баг в функции ParseConfig", goRepo(t), "`golang-skills:go-code`"},
		}
		for _, tc := range cases {
			if _, out := promptEvent(t, t.TempDir(), "p18", tc.cwd, tc.prompt); !strings.Contains(out, tc.want) {
				t.Errorf("Russian prompt %q: stdout %q; want %s", tc.prompt, out, tc.want)
			}
		}
		// A question does not start the edit router.
		if _, out := promptEvent(t, t.TempDir(), "p19", goRepo(t), "Объясни, как работает эта функция на Go?"); out != "" {
			t.Errorf("Russian question: stdout %q; want silent", out)
		}
	})

	// Without CLAUDE_PLUGIN_ROOT the hook is not wired as a plugin, and the note
	// names the skills by bare names, as such a host registers them.
	t.Run("bare names outside a plugin", func(t *testing.T) {
		t.Parallel()
		_, out := promptEvent(t, t.TempDir(), "p20", t.TempDir(), implement, "CLAUDE_PLUGIN_ROOT=")
		if !strings.Contains(out, "Skill tool, name `go-code`") || strings.Contains(out, "golang-skills:go-") {
			t.Fatalf("note outside a plugin: stdout %q; want bare `go-code`", out)
		}
	})
}

// subagentEvent runs the SubagentStart hook and returns its exit code and
// stdout, which the host adds to the subagent's context.
func subagentEvent(t *testing.T, cwd, agentType string) (int, string) {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"hook_event_name": "SubagentStart",
		"session_id":      "sub1",
		"cwd":             cwd,
		"agent_id":        "a1",
		"agent_type":      agentType,
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	cmd := exec.Command("bash", filepath.Join(repoRoot(t), "hooks", "go-subagent-routing.sh"))
	cmd.Stdin = strings.NewReader(string(body))
	cmd.Env = append(os.Environ(), "CLAUDE_PLUGIN_ROOT="+repoRoot(t))
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if stderr.Len() > 0 {
		t.Errorf("subagent hook wrote to stderr, which a SubagentStart hook must not:\n%s", stderr.String())
	}
	if err == nil {
		return 0, stdout.String()
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("run subagent hook: %v", err)
	}
	return exitErr.ExitCode(), stdout.String()
}

// TestSubagentRouting drives the SubagentStart hook: a subagent started in a
// Go project is told which router to load, and so is the next one, since
// each subagent starts with an empty context; a subagent started outside Go,
// or as the plugin's own go-verify agent, hears nothing.
func TestSubagentRouting(t *testing.T) {
	t.Parallel()
	goDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(goDir, "go.mod"), []byte("module scratch\n\ngo 1.27\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	t.Run("Go project names the routers", func(t *testing.T) {
		t.Parallel()
		for i := 0; i < 2; i++ {
			code, out := subagentEvent(t, goDir, "general-purpose")
			if code != 0 {
				t.Fatalf("subagent %d in a Go directory: exit %d, want 0", i, code)
			}
			for _, want := range []string{"name `golang-skills:go-code`", "`golang-skills:go-code-refactor`", "before the first edit"} {
				if !strings.Contains(out, want) {
					t.Errorf("subagent %d note must mention %q:\n%s", i, want, out)
				}
			}
		}
	})

	t.Run("silent outside Go", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("docs\n"), 0o644); err != nil {
			t.Fatalf("write README.md: %v", err)
		}
		if code, out := subagentEvent(t, dir, "general-purpose"); code != 0 || out != "" {
			t.Fatalf("subagent outside Go: exit %d, stdout %q; want silent 0", code, out)
		}
	})

	t.Run("silent for go-verify and agents that write no Go", func(t *testing.T) {
		t.Parallel()
		for _, agent := range []string{"go-verify", "golang-skills:go-verify", "Explore", "claude-code-guide", "statusline-setup"} {
			if code, out := subagentEvent(t, goDir, agent); code != 0 || out != "" {
				t.Fatalf("%s subagent: exit %d, stdout %q; want silent 0", agent, code, out)
			}
		}
	})

	t.Run("Plan still hears the note", func(t *testing.T) {
		t.Parallel()
		if code, out := subagentEvent(t, goDir, "Plan"); code != 0 || !strings.Contains(out, "`golang-skills:go-code`") {
			t.Fatalf("Plan subagent: exit %d, stdout %q; want the router note", code, out)
		}
	})
}

// ladderEvent runs the restraint ladder hook with payload on stdin and returns
// its exit code and stdout, which the host adds to the context.
func ladderEvent(t *testing.T, state string, payload map[string]any, env ...string) (int, string) {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	cmd := exec.Command("bash", filepath.Join(repoRoot(t), "hooks", "go-restraint-ladder.sh"))
	cmd.Stdin = strings.NewReader(string(body))
	cmd.Env = append(append(os.Environ(), "CLAUDE_PLUGIN_DATA="+state, "GOLANG_SKILLS_LADDER="), env...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if stderr.Len() > 0 {
		t.Errorf("ladder hook wrote to stderr:\n%s", stderr.String())
	}
	if err == nil {
		return 0, stdout.String()
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("run ladder hook: %v", err)
	}
	return exitErr.ExitCode(), stdout.String()
}

// TestLadderHook drives the restraint ladder hook: a session or subagent
// started in a Go project gets the ladder from OVER-ENGINEERING.md and the
// level line from go-code's Intensity table; a level word in a prompt sets the
// level for the rest of the session; GOLANG_SKILLS_LADDER sets the starting
// level or turns the hook off.
func TestLadderHook(t *testing.T) {
	t.Parallel()
	goDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(goDir, "go.mod"), []byte("module scratch\n\ngo 1.27\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	start := func(session, cwd string) map[string]any {
		return map[string]any{"hook_event_name": "SessionStart", "source": "startup", "session_id": session, "cwd": cwd}
	}
	prompt := func(session, text string) map[string]any {
		return map[string]any{"hook_event_name": "UserPromptSubmit", "session_id": session, "cwd": goDir, "prompt": text}
	}
	subagent := func(session, agent string) map[string]any {
		return map[string]any{"hook_event_name": "SubagentStart", "session_id": session, "cwd": goDir, "agent_type": agent}
	}

	t.Run("session start in Go prints the ladder at full", func(t *testing.T) {
		t.Parallel()
		code, out := ladderEvent(t, t.TempDir(), start("l1", goDir))
		if code != 0 {
			t.Fatalf("exit %d, want 0", code)
		}
		owner := filepath.Join(repoRoot(t), "skills", "go-code-refactor", "references", "OVER-ENGINEERING.md")
		for _, want := range []string{owner, "## The Restraint Ladder", "Does this need to exist at all?", "Can it be one line?", "**Never on the chopping block**", "Restraint level `full`", "as written. The default."} {
			if !strings.Contains(out, want) {
				t.Errorf("session start must print %q:\n%s", want, out)
			}
		}
		if strings.Contains(out, "## Reach For What Go Ships") {
			t.Errorf("only the ladder section is printed, not the rest of the file:\n%s", out)
		}
	})

	t.Run("silent outside Go", func(t *testing.T) {
		t.Parallel()
		if code, out := ladderEvent(t, t.TempDir(), start("l2", t.TempDir())); code != 0 || out != "" {
			t.Fatalf("session outside Go: exit %d, stdout %q; want silent 0", code, out)
		}
	})

	t.Run("a level word holds for the session", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		_, out := ladderEvent(t, state, prompt("l3", "/golang-skills:go-code ultra ./feed"))
		if !strings.Contains(out, "restraint level `ultra`") || !strings.Contains(out, "`Need <X>? <Y> covers it.`") {
			t.Fatalf("/go-code ultra: stdout %q; want the ultra line", out)
		}
		if _, out := ladderEvent(t, state, start("l3", goDir)); !strings.Contains(out, "Restraint level `ultra`") {
			t.Errorf("compact after ultra: stdout %q; want level ultra", out)
		}
		if _, out := ladderEvent(t, state, subagent("l3", "general-purpose")); !strings.Contains(out, "Restraint level `ultra`") {
			t.Errorf("subagent after ultra: stdout %q; want level ultra", out)
		}
		if _, out := ladderEvent(t, state, prompt("l3", "Lite mode: add a handler that returns the balance")); !strings.Contains(out, "`lazier: <X>`") {
			t.Errorf("lite mode: stdout %q; want the lite line", out)
		}
		if _, out := ladderEvent(t, state, prompt("l4", "режим ultra, додай хендлер")); !strings.Contains(out, "`ultra`") {
			t.Errorf("режим ultra: stdout %q; want the ultra line", out)
		}
	})

	t.Run("prompts without a level word stay silent", func(t *testing.T) {
		t.Parallel()
		// A refactor runs at full, so its command sets no level.
		for _, p := range []string{"Implement the Go package in ./feed", "/go-code-refactor ultra ./dispatch", "/go-code ./feed"} {
			if code, out := ladderEvent(t, t.TempDir(), prompt("l5", p)); code != 0 || out != "" {
				t.Errorf("prompt %q: exit %d, stdout %q; want silent 0", p, code, out)
			}
		}
	})

	t.Run("GOLANG_SKILLS_LADDER sets the start level or turns the hook off", func(t *testing.T) {
		t.Parallel()
		if _, out := ladderEvent(t, t.TempDir(), start("l6", goDir), "GOLANG_SKILLS_LADDER=lite"); !strings.Contains(out, "Restraint level `lite`") {
			t.Errorf("GOLANG_SKILLS_LADDER=lite: stdout %q; want level lite", out)
		}
		for _, payload := range []map[string]any{start("l7", goDir), prompt("l7", "/go-code ultra ./feed"), subagent("l7", "general-purpose")} {
			if code, out := ladderEvent(t, t.TempDir(), payload, "GOLANG_SKILLS_LADDER=off"); code != 0 || out != "" {
				t.Errorf("GOLANG_SKILLS_LADDER=off, %v: exit %d, stdout %q; want silent 0", payload["hook_event_name"], code, out)
			}
		}
	})

	t.Run("silent for go-verify and agents that write no Go", func(t *testing.T) {
		t.Parallel()
		for _, agent := range []string{"go-verify", "golang-skills:go-verify", "Explore", "claude-code-guide", "statusline-setup"} {
			if code, out := ladderEvent(t, t.TempDir(), subagent("l8", agent)); code != 0 || out != "" {
				t.Errorf("%s subagent: exit %d, stdout %q; want silent 0", agent, code, out)
			}
		}
	})

	t.Run("Plan still gets the ladder", func(t *testing.T) {
		t.Parallel()
		if code, out := ladderEvent(t, t.TempDir(), subagent("l9", "Plan")); code != 0 || !strings.Contains(out, "## The Restraint Ladder") {
			t.Errorf("Plan subagent: exit %d, stdout %q; want the ladder", code, out)
		}
	})
}
