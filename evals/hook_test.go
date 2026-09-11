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

	// The hints are heuristics for decision-bearing forms. Routine syntax that
	// go-code's router says does not trigger a load — fmt.Errorf with %v, an
	// http constant in a comment — must leave the gate silent.
	t.Run("routine syntax names no owner", func(t *testing.T) {
		t.Parallel()
		state := t.TempDir()
		for _, skill := range []string{"go-code", "go-style-core"} {
			hookEvent(t, script, state, routingPayload("PostToolUse", "s6", "Skill", map[string]any{"skill": skill}))
		}
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
		collections := "package store\n\nfunc (s *Store) IDs() []string {\n\tids := make([]string, 0, len(s.m))\n\tseen := make(map[string]struct{})\n\tfor id := range s.m {\n\t\tids = append(ids, id)\n\t}\n\treturn ids\n}\n"
		code, msg := hookEvent(t, script, state, routingPayload("PreToolUse", "s10", "Write",
			map[string]any{"file_path": "/repo/store/ids.go", "content": collections}))
		if code != 0 || msg != "" {
			t.Fatalf("make/append/make(map): exit %d, stderr %q; want silent 0", code, msg)
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
			{"defer", "s9", "/repo/x/read.go",
				"package x\n\nfunc read(f *os.File) {\n\tdefer f.Close()\n}\n", "go-defensive"},
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
}
