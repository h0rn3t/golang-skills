package evals_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodexEditHook(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name       string
		files      []string
		patch      string
		clean      bool
		cleanFirst bool
	}{
		{name: "relative update", files: []string{"one/main.go"}, patch: "*** Update File: one/main.go\n@@\n-old\n+new"},
		{name: "multiple packages continue after findings", files: []string{"one/main.go", "two/main.go"}, patch: "*** Update File: one/main.go\n@@\n-old\n+new\n*** Add File: two/main.go\n+package p"},
		{name: "rename checks destination", files: []string{"renamed/main.go"}, patch: "*** Update File: removed.go\n*** Move to: renamed/main.go\n@@\n-old\n+new"},
		{name: "path passed without shell evaluation", files: []string{"one/odd ' $(touch injected).go"}, patch: "*** Add File: one/odd ' $(touch injected).go\n+package p"},
		{name: "clean output is one JSON object", files: []string{"one/main.go", "two/main.go"}, patch: "*** Add File: one/main.go\n+package p\n*** Add File: two/main.go\n+package p", clean: true},
		{name: "passing receipts survive a later finding", files: []string{"one/main.go", "two/main.go"}, patch: "*** Add File: one/main.go\n+package p\n*** Add File: two/main.go\n+package p", cleanFirst: true},
		{name: "deletion is ignored", patch: "*** Delete File: removed.go"},
		{name: "non-Go is ignored", patch: "*** Add File: README.md\n+hello"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cwd, state := t.TempDir(), t.TempDir()
			for i, file := range tt.files {
				path := filepath.Join(cwd, file)
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(filepath.Dir(path), "go.mod"), []byte("module scratch\n\ngo 1.26\n"), 0o644); err != nil {
					t.Fatal(err)
				}
				src := "package p\n\nfunc F() {\n  println(1)\n}\n"
				if tt.clean || (i == 0 && tt.cleanFirst) {
					src = "package p\n\nfunc F() {}\n"
				}
				if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			payload := routingPayload("PostToolUse", "codex-edit", "apply_patch", map[string]any{"command": "*** Begin Patch\n" + tt.patch + "\n*** End Patch"})
			payload["cwd"] = cwd
			body, err := json.Marshal(payload)
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("bash", filepath.Join(repoRoot(t), "hooks/go-hook-helper.sh"), "codex-edit")
			cmd.Stdin = strings.NewReader(string(body))
			cmd.Env = append(os.Environ(), "PLUGIN_ROOT="+repoRoot(t), "CLAUDE_PLUGIN_ROOT="+repoRoot(t), "CLAUDE_PLUGIN_DATA="+state, "GOLANG_SKILLS_EDIT_LINT=off", "GOLANG_SKILLS_EDIT_TESTS=off")
			var stdout, stderr strings.Builder
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			err = cmd.Run()
			if len(tt.files) == 0 {
				if err != nil || stdout.Len() != 0 || stderr.Len() != 0 {
					t.Fatalf("non-edit: %v, stdout %q, stderr %q; want silent success", err, stdout.String(), stderr.String())
				}
				return
			}
			if tt.clean {
				var result struct {
					HookSpecificOutput struct {
						HookEventName, AdditionalContext string
					}
				}
				if err != nil || json.Unmarshal([]byte(stdout.String()), &result) != nil || result.HookSpecificOutput.HookEventName != "PostToolUse" {
					t.Fatalf("clean edit: %v, stdout %q, stderr %q; want one PostToolUse JSON object", err, stdout.String(), stderr.String())
				}
				for _, file := range tt.files {
					if !strings.Contains(result.HookSpecificOutput.AdditionalContext, filepath.Dir(filepath.Join(cwd, file))) {
						t.Errorf("missing receipt context for %s: %s", file, stdout.String())
					}
				}
				return
			}
			exitErr, ok := err.(*exec.ExitError)
			if !ok || exitErr.ExitCode() != 2 {
				t.Fatalf("unformatted edit: %v, stdout %q, stderr %q; want exit 2", err, stdout.String(), stderr.String())
			}
			for i, file := range tt.files {
				if i == 0 && tt.cleanFirst {
					if !strings.Contains(stderr.String(), filepath.Dir(filepath.Join(cwd, file))) {
						t.Errorf("missing passing receipt context for %s: %s", file, stderr.String())
					}
				} else if !strings.Contains(stderr.String(), "gofmt: "+filepath.Join(cwd, file)) {
					t.Errorf("missing finding for %s: %s", file, stderr.String())
				}
			}
			if _, err := os.Stat(filepath.Join(cwd, "injected")); !os.IsNotExist(err) {
				t.Fatalf("path was evaluated by a shell: %v", err)
			}
		})
	}
}
