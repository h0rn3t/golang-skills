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
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	cmd := exec.Command("bash", script)
	cmd.Stdin = strings.NewReader(string(body))
	cmd.Env = append(os.Environ(), "CLAUDE_PLUGIN_DATA="+state)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err == nil {
		return 0, stderr.String()
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("run %s: %v", script, err)
	}
	return exitErr.ExitCode(), stderr.String()
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

// TestRoutingGate drives the go-code routing hook through one session: no
// gate without go-code, a block that names exactly the missing owners, a pass
// once they are loaded, and a pass on retry even when they are not.
func TestRoutingGate(t *testing.T) {
	t.Parallel()
	script := filepath.Join(repoRoot(t), "hooks", "go-code-routing.sh")
	handler := "package api\n\nfunc (s *Server) handle(w http.ResponseWriter, r *http.Request) {\n\tif err := s.store.Save(r.Context(), u); err != nil {\n\t\thttp.Error(w, fmt.Errorf(\"save: %w\", err).Error(), 500)\n\t}\n}\n"

	t.Run("silent without go-code", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		code, msg := hookEvent(t, script, state, routingPayload("PreToolUse", "s1", "Write",
			map[string]any{"file_path": "/repo/api/handler.go", "content": handler}))
		if code != 0 || msg != "" {
			t.Fatalf("edit without go-code loaded: exit %d, stderr %q; want silent 0", code, msg)
		}
	})

	t.Run("blocks once then passes", func(t *testing.T) {
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
		for _, want := range []string{"go-style-core", "go-http", "go-error-handling"} {
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

		// The retry passes without loading anything: the gate reminds once.
		if code, msg := hookEvent(t, script, state, edit); code != 0 {
			t.Fatalf("retry after one reminder: exit %d, stderr %q; want 0 (no deadlock)", code, msg)
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

		code, msg := hookEvent(t, script, state, routingPayload("PreToolUse", "s3", "Write",
			map[string]any{"file_path": "/repo/api/handler.go", "content": handler}))
		if code != 0 || msg != "" {
			t.Fatalf("edit with every owner loaded: exit %d, stderr %q; want silent 0", code, msg)
		}
	})

	t.Run("test file requires go-testing", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		hookEvent(t, script, state, routingPayload("PostToolUse", "s4", "Skill", map[string]any{"skill": "go-code"}))
		hookEvent(t, script, state, routingPayload("PostToolUse", "s4", "Skill", map[string]any{"skill": "go-style-core"}))
		code, msg := hookEvent(t, script, state, routingPayload("PreToolUse", "s4", "Edit",
			map[string]any{"file_path": "/repo/api/handler_test.go", "old_string": "a", "new_string": "func TestX(t *testing.T) {}"}))
		if code != 2 || !strings.Contains(msg, "go-testing") {
			t.Fatalf("_test.go edit: exit %d, stderr %q; want 2 naming go-testing", code, msg)
		}
		if strings.Contains(msg, "go-style-core") {
			t.Fatalf("go-style-core is loaded and must not be named:\n%s", msg)
		}
	})

	t.Run("ignores non-Go files and other sessions", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		hookEvent(t, script, state, routingPayload("PostToolUse", "s5", "Skill", map[string]any{"skill": "go-code"}))
		if code, msg := hookEvent(t, script, state, routingPayload("PreToolUse", "s5", "Write",
			map[string]any{"file_path": "/repo/README.md", "content": "http."})); code != 0 || msg != "" {
			t.Fatalf("Markdown edit: exit %d, stderr %q; want silent 0", code, msg)
		}
		if code, msg := hookEvent(t, script, state, routingPayload("PreToolUse", "other", "Write",
			map[string]any{"file_path": "/repo/main.go", "content": "package main"})); code != 0 || msg != "" {
			t.Fatalf("another session's Go edit: exit %d, stderr %q; want silent 0", code, msg)
		}
	})
}
