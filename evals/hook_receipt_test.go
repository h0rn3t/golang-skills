package evals_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestHookReceipts(t *testing.T) {
	t.Parallel()
	script := filepath.Join(repoRoot(t), "hooks", "go-vet-on-edit.sh")
	module := func(t *testing.T) string {
		t.Helper()
		dir := t.TempDir()
		for name, content := range map[string]string{
			"go.mod":  "module receipt\n\ngo 1.27\n",
			"main.go": "package receipt\n\nfunc Value() int { return 1 }\n",
		} {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return filepath.Join(dir, "main.go")
	}
	type receipt struct {
		Check       string            `json:"check"`
		Status      string            `json:"status"`
		ExitCode    *int              `json:"exit_code"`
		Command     []string          `json:"command"`
		Scope       map[string]string `json:"scope"`
		Config      string            `json:"config"`
		ToolVersion string            `json:"tool_version"`
		Started     string            `json:"started"`
		Ended       string            `json:"ended"`
		Before      string            `json:"inputs_before"`
		After       string            `json:"inputs_after"`
		Reusable    bool              `json:"reusable"`
		Diagnostic  string            `json:"diagnostic"`
	}
	read := func(t *testing.T, state string) map[string]receipt {
		t.Helper()
		out := map[string]receipt{}
		err := filepath.WalkDir(state, func(path string, ent os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if ent.IsDir() || !strings.HasSuffix(path, ".json") {
				return nil
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			var r receipt
			if err := json.Unmarshal(raw, &r); err != nil {
				return err
			}
			out[r.Check] = r
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	t.Run("explicit results include skipped checks and actual scope", func(t *testing.T) {
		t.Parallel()
		path, state := module(t), t.TempDir()
		code, out, msg := hookOutput(t, script, state, routingPayload("PostToolUse", "receipts", "Write", map[string]any{"file_path": path}), "GOLANG_SKILLS_EDIT_LINT=off", "GOWORK=off")
		if code != 0 || msg != "" || !strings.Contains(out, "additionalContext") {
			t.Fatalf("clean hook = (%d, %q, %q), want exit 0 with context", code, out, msg)
		}
		r := read(t, state)
		if len(r) != 5 {
			t.Fatalf("receipts = %v, want five explicit checks", r)
		}
		for _, check := range []string{"gofmt", "vet", "fix"} {
			v := r[check]
			if v.Status != "pass" || v.ExitCode == nil || *v.ExitCode != 0 || !v.Reusable || v.Before == "" || v.Before != v.After || len(v.Command) == 0 || v.ToolVersion == "" || v.Started == "" || v.Ended == "" {
				t.Errorf("%s receipt = %+v, want complete reusable pass", check, v)
			}
		}
		if r["gofmt"].Scope["kind"] != "file" || r["vet"].Scope["kind"] != "package" {
			t.Errorf("receipt scopes = %v, want file gofmt and package vet", r)
		}
		for _, check := range []string{"test", "lint"} {
			if r[check].Status != "skipped" || r[check].ExitCode != nil || r[check].Reusable {
				t.Errorf("%s receipt = %+v, want skipped without exit or reuse", check, r[check])
			}
		}
		paths, err := filepath.Glob(filepath.Join(state, "checks", "run.*", "vet.json"))
		if err != nil || len(paths) != 1 {
			t.Fatalf("vet receipt paths = %v, %v; want one", paths, err)
		}
		if code, out := receiptCommand(t, "verify", paths[0]); code != 0 {
			t.Fatalf("verify hook result from shell = (%d, %q), want current pass", code, out)
		}
	})
	t.Run("formatting findings are failure despite exit zero", func(t *testing.T) {
		t.Parallel()
		path, state := module(t), t.TempDir()
		if err := os.WriteFile(path, []byte("package receipt\nfunc Value() int{return 1}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		hookOutput(t, script, state, routingPayload("PostToolUse", "format", "Write", map[string]any{"file_path": path}), "GOLANG_SKILLS_EDIT_LINT=off", "GOWORK=off")
		v := read(t, state)["gofmt"]
		if v.Status != "fail" || v.ExitCode == nil || *v.ExitCode != 0 || v.Reusable {
			t.Fatalf("gofmt receipt = %+v, want fail with actual exit 0", v)
		}
	})
	for _, tc := range []struct{ name, status, output, want string }{
		{"new finding", "1", "/tmp/example.go:2:1: unchecked (errcheck)", "fail"},
		{"config failure", "3", "invalid configuration", "unavailable"},
		{"empty failure", "1", "", "unavailable"},
		{"timeout", "124", "", "unavailable"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path, state, bin := module(t), t.TempDir(), t.TempDir()
			stub := "#!/usr/bin/env bash\ncase $1 in version) echo test-lint-1; exit 0;; config) exit 1;; esac\nprintf '%s' \"$LINT_OUTPUT\"\nexit \"$LINT_STATUS\"\n"
			if err := os.WriteFile(filepath.Join(bin, "golangci-lint"), []byte(stub), 0o755); err != nil {
				t.Fatal(err)
			}
			hookOutput(t, script, state, routingPayload("PostToolUse", tc.name, "Write", map[string]any{"file_path": path}), "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"), "GOWORK=off", "LINT_STATUS="+tc.status, "LINT_OUTPUT="+tc.output)
			v := read(t, state)["lint"]
			if v.Status != tc.want || v.ExitCode == nil || v.Reusable || v.Config == "" || !strings.Contains(v.Diagnostic, tc.output) {
				t.Errorf("lint %s receipt = %+v, want %s with exit/config/diagnostic", tc.name, v, tc.want)
			}
		})
	}
	t.Run("outside a module records skipped package checks", func(t *testing.T) {
		t.Parallel()
		path, state := module(t), t.TempDir()
		if err := os.Remove(filepath.Join(filepath.Dir(path), "go.mod")); err != nil {
			t.Fatal(err)
		}
		hookOutput(t, script, state, routingPayload("PostToolUse", "no-module", "Write", map[string]any{"file_path": path}), "GOWORK=off")
		for check, v := range read(t, state) {
			if check != "gofmt" && (v.Status != "skipped" || v.ExitCode != nil || v.Reusable || len(v.Command) == 0) {
				t.Errorf("outside module %s = %+v, want skipped with planned command", check, v)
			}
		}
	})
	t.Run("typecheck failure explains skipped dependent checks", func(t *testing.T) {
		t.Parallel()
		path, state := module(t), t.TempDir()
		if err := os.WriteFile(path, []byte("package receipt\n\nvar value = undefined\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		hookOutput(t, script, state, routingPayload("PostToolUse", "typecheck", "Write", map[string]any{"file_path": path}), "GOWORK=off")
		v := read(t, state)
		if v["vet"].Status != "fail" {
			t.Errorf("vet = %+v, want fail", v["vet"])
		}
		for _, check := range []string{"fix", "test", "lint"} {
			if v[check].Status != "skipped" || v[check].Reusable {
				t.Errorf("%s = %+v, want skipped", check, v[check])
			}
		}
	})
}

func TestPromptStubHints(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, source string
		want, absent []string
	}{
		{"aliased signatures", "package stub\nimport (ctx \"context\"; lock \"sync\"; log \"log/slog\")\ntype Pool struct { mu lock.Mutex; logger *log.Logger }\nfunc Run(c ctx.Context) error { panic(\"not implemented\") }\n", []string{"go-context", "go-concurrency", "go-logging"}, nil},
		{"errors alias", "package stub\nimport e \"errors\"\nvar ErrInvalid = e.New(\"invalid\")\nfunc Run() error { panic(\"not implemented\") }\n", []string{"go-error-handling"}, nil},
		{"blank and unused imports", "package stub\nimport (_ \"sync\"; _ \"context\"; \"errors\"; \"log/slog\")\nfunc Run() { panic(\"not implemented\") }\n", nil, []string{"go-context", "go-concurrency", "go-error-handling", "go-logging"}},
		{"dot signature", "package stub\nimport . \"context\"\nfunc Run(c Context) { panic(\"not implemented\") }\n", []string{"go-context"}, nil},
		{"comments and strings are not imports", "package stub\n// import \"sync\"\n// context.WithCancel and errors.New are examples, not this contract.\nvar text = `import \"log/slog\"; sync.Mutex; go func`\nfunc Run() { panic(\"not implemented\") }\n", nil, []string{"go-context", "go-concurrency", "go-error-handling", "go-logging"}},
		{"foreign aliases are not standard owners", "package stub\nimport (sync \"example.com/locks\"; context \"example.com/ctx\")\nfunc Run(c context.Context, m sync.Mutex) { panic(\"not implemented\") }\n", nil, []string{"go-context", "go-concurrency"}},
		{"existing HTTP signature", "package stub\nimport \"net/http\"\nfunc Run(r *http.Request) { panic(\"not implemented\") }\n", []string{"go-http", "go-error-handling"}, nil},
		{"existing security and resilience imports", "package stub\nimport (\"html/template\"; \"crypto/rand\"; \"golang.org/x/time/rate\")\nfunc Run(t *template.Template, limiter *rate.Limiter) { rand.Read(nil); panic(\"not implemented\") }\n", []string{"go-security", "go-resilience"}, nil},
		{"raw stub contract", "// Package stub runs jobs on bounded workers.\npackage stub\nfunc Run() { panic(`not implemented`) }\n", []string{"go-concurrency"}, nil},
		{"raw sample is not a stub", "// Sample of bounded workers.\npackage stub\nvar sample = `panic(\"not implemented\")`\nfunc Run() {}\n", nil, []string{"go-concurrency"}},
		{"format wrap example is not code", "package stub\nvar sample = `fmt.Errorf(\"%w\", err)`\nfunc Run() {}\n", nil, []string{"go-error-handling"}},
		{"retry contract", "package stub\n// Run retries unavailable requests, honoring Retry-After.\nfunc Run() { panic(\"not implemented\") }\n", []string{"go-resilience"}, nil},
		{"backoff contract", "package stub\n// Run retries failures with bounded exponential backoff.\nfunc Run() { panic(\"not implemented\") }\n", []string{"go-resilience"}, nil},
		{"unrelated retry prose", "package stub\n// Retry-After and backoff are names in an example, not this contract.\nfunc Run() { panic(\"not implemented\") }\n", nil, []string{"go-resilience"}},
		{"retry prose without a stub", "package stub\n// Run used to retry unavailable requests honoring Retry-After.\nfunc Run() {}\n", nil, []string{"go-resilience"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cwd := t.TempDir()
			if err := os.WriteFile(filepath.Join(cwd, "stub.go"), []byte(tc.source), 0o644); err != nil {
				t.Fatal(err)
			}
			code, out := promptEvent(t, t.TempDir(), "stub", cwd, "Implement the Go package in ./stub.go")
			if code != 0 {
				t.Fatalf("prompt hook exit = %d, want 0", code)
			}
			for _, owner := range tc.want {
				if !strings.Contains(out, "`golang-skills:"+owner+"`") {
					t.Errorf("prompt = %q, want %s", out, owner)
				}
			}
			for _, owner := range tc.absent {
				if strings.Contains(out, "`golang-skills:"+owner+"`") {
					t.Errorf("prompt = %q, unexpected %s", out, owner)
				}
			}
		})
	}
	t.Run("pool contract signals concurrency before bodies exist", func(t *testing.T) {
		t.Parallel()
		cwd := filepath.Join(repoRoot(t), "evals", "ab", "_implement")
		_, out := promptEvent(t, t.TempDir(), "pool-stub", cwd, "Implement the Go package in ./pool")
		for _, owner := range []string{"go-context", "go-concurrency"} {
			if !strings.Contains(out, "`golang-skills:"+owner+"`") {
				t.Errorf("pool prompt = %q, want %s", out, owner)
			}
		}
	})
	t.Run("fetch retains resilience from the Retry-After contract", func(t *testing.T) {
		t.Parallel()
		cwd := filepath.Join(repoRoot(t), "evals", "ab", "_implement")
		_, out := promptEvent(t, t.TempDir(), "fetch-stub", cwd, "Implement the Go package in ./fetch")
		for _, owner := range []string{"go-context", "go-http", "go-error-handling", "go-resilience"} {
			if !strings.Contains(out, "`golang-skills:"+owner+"`") {
				t.Errorf("fetch note = %q, want %s", out, owner)
			}
		}
	})
	// New prompt heuristics must not become unconditional edit requirements.
	t.Run("context signature alone does not widen edit gate", func(t *testing.T) {
		t.Parallel()
		script, state := filepath.Join(repoRoot(t), "hooks", "go-code-routing.sh"), t.TempDir()
		for _, skill := range []string{"go-code", "go-style-core"} {
			hookEvent(t, script, state, routingPayload("PostToolUse", "gate", "Skill", map[string]any{"skill": skill}))
		}
		code, msg := hookEvent(t, script, state, routingPayload("PreToolUse", "gate", "Edit", map[string]any{"file_path": "/repo/stub.go", "new_string": "func Run(ctx context.Context) error { return nil }"}))
		if code != 0 || msg != "" {
			t.Fatalf("signature edit = (%d, %q), want unchanged permissive gate", code, msg)
		}
	})
}

// receiptCommand runs the public receipt verifier, with the same build profile
// as the hook tests. A stale record must be rejected even after an edit via Bash.
func receiptCommand(t *testing.T, args ...string) (int, string) {
	t.Helper()
	return receiptCommandEnv(t, nil, args...)
}

func receiptCommandEnv(t *testing.T, env []string, args ...string) (int, string) {
	t.Helper()
	cmd := exec.Command("python3", append([]string{filepath.Join(repoRoot(t), "hooks", "go-check-receipt.py")}, args...)...)
	cmd.Env = append(append(os.Environ(), "GOWORK=off"), env...)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return 0, string(out)
	}
	if exit, ok := err.(*exec.ExitError); ok {
		return exit.ExitCode(), string(out)
	}
	t.Fatalf("receipt command %v: %v", args, err)
	return 0, ""
}

func TestReceiptShellCredit(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "module with ' quote")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "main.go")
	for name, contents := range map[string]string{"go.mod": "module credit\n\ngo 1.27\n", "main.go": "package credit\n\nfunc Value() int { return 1 }\n"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	records := t.TempDir()
	if code, out := receiptCommand(t, "run", records, "vet", file, "", "all", "run", "", "go", "vet", "."); code != 0 {
		t.Fatalf("vet run = (%d, %q), want success", code, out)
	}
	record := filepath.Join(records, "vet.json")
	verify := func(t *testing.T, cwd string, command ...string) (int, string) {
		t.Helper()
		args := append([]string{filepath.Join(repoRoot(t), "hooks", "go-check-receipt.sh"), record, cwd}, command...)
		cmd := exec.Command("bash", args...)
		cmd.Env = append(os.Environ(), "GOWORK=off")
		out, err := cmd.CombinedOutput()
		if err == nil {
			return 0, string(out)
		}
		if exit, ok := err.(*exec.ExitError); ok {
			return exit.ExitCode(), string(out)
		}
		t.Fatalf("shell verifier: %v", err)
		return 0, ""
	}
	t.Run("matching package and command grants hook credit", func(t *testing.T) {
		code, out := verify(t, dir, "go", "vet", ".")
		var result struct {
			Valid  bool `json:"valid"`
			Credit bool `json:"hook_credit"`
		}
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("verifier JSON = %q, error %v", out, err)
		}
		if code != 0 || !result.Valid || !result.Credit {
			t.Errorf("verifier = (%d, %s), want verified hook credit", code, out)
		}
	})
	t.Run("wider command cannot borrow package receipt", func(t *testing.T) {
		if code, out := verify(t, dir, "go", "vet", "./..."); code == 0 {
			t.Errorf("wide vet = (%d, %q), want rejected", code, out)
		}
	})
	t.Run("same dot command in a different directory is rejected", func(t *testing.T) {
		if code, out := verify(t, filepath.Dir(dir), "go", "vet", "."); code == 0 {
			t.Errorf("different cwd = (%d, %q), want rejected", code, out)
		}
	})
	t.Run("state-only legacy verifier does not grant hook credit", func(t *testing.T) {
		code, out := receiptCommand(t, "verify", record)
		var result map[string]any
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatal(err)
		}
		if code != 0 || result["valid"] != true || result["hook_credit"] != false {
			t.Errorf("state-only verify = (%d, %q), want valid without hook credit", code, out)
		}
	})
	t.Run("host-only hook and shell metadata does not invalidate inputs", func(t *testing.T) {
		code, out := receiptCommandEnv(t, []string{"AI_AGENT=changed", "CLAUDE_CODE_EXECPATH=/another/cli", "CLAUDE_PROJECT_DIR=/another/host-dir", "GIT_EDITOR=changed"}, "verify", record, dir, "go", "vet", ".")
		if code != 0 || !strings.Contains(out, `"hook_credit": true`) {
			t.Errorf("host bookkeeping = (%d,%q), want current credit", code, out)
		}
	})
	t.Run("changed input cancels previously verified credit", func(t *testing.T) {
		if err := os.WriteFile(file, []byte("package credit\n\nfunc Value() int { return 2 }\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if code, out := verify(t, dir, "go", "vet", "."); code == 0 {
			t.Errorf("stale verify = (%d, %q), want rejected", code, out)
		}
	})
}

func TestHookOffersRunnableVerifier(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := filepath.Join(dir, "main.go")
	for name, contents := range map[string]string{"go.mod": "module offer\n\ngo 1.27\n", "main.go": "package offer\n\nfunc Value() int { return 1 }\n"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	code, out, msg := hookOutput(t, filepath.Join(repoRoot(t), "hooks", "go-vet-on-edit.sh"), t.TempDir(), routingPayload("PostToolUse", "offer", "Write", map[string]any{"file_path": file}), "GOWORK=off", "GOLANG_SKILLS_EDIT_LINT=off")
	if code != 0 || msg != "" {
		t.Fatalf("hook = (%d, %q), want clean check offer", code, msg)
	}
	var result struct {
		Output struct {
			Context string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatal(err)
	}
	var commands []string
	for _, line := range strings.Split(result.Output.Context, "\n") {
		if strings.HasPrefix(line, "bash ") && !strings.Contains(line, " --gate ") {
			commands = append(commands, line)
		}
	}
	if len(commands) == 0 {
		t.Fatalf("hook context = %q, want an executable Bash verification command", result.Output.Context)
	}
	// Run the command exactly as copied from model context, without plugin env.
	cmd := exec.Command("bash", "-c", commands[0])
	cmd.Env = append(os.Environ(), "GOWORK=off")
	verified, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("offered command %q = %v: %s", commands[0], err, verified)
	}
	var credit struct {
		Credit bool `json:"hook_credit"`
	}
	if err := json.Unmarshal(verified, &credit); err != nil {
		t.Fatal(err)
	}
	if !credit.Credit {
		t.Errorf("offered command = %s, want hook_credit true", verified)
	}
}

func TestFailingHookOffersVerificationForPassingChecks(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := filepath.Join(dir, "main.go")
	for name, contents := range map[string]string{"go.mod": "module offer\n\ngo 1.27\n", "main.go": "package offer\nfunc Value() int{return 1}\n"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	code, _, msg := hookOutput(t, filepath.Join(repoRoot(t), "hooks", "go-vet-on-edit.sh"), t.TempDir(), routingPayload("PostToolUse", "offer-fail", "Write", map[string]any{"file_path": file}), "GOWORK=off", "GOLANG_SKILLS_EDIT_LINT=off")
	if code != 2 {
		t.Fatalf("unformatted hook = (%d, %q), want failure", code, msg)
	}
	var command string
	for _, line := range strings.Split(msg, "\n") {
		if strings.HasPrefix(line, "bash ") && !strings.Contains(line, " --gate ") {
			command = line
			break
		}
	}
	if command == "" {
		t.Fatalf("hook error = %q, want executable verifier for a passing check", msg)
	}
	cmd := exec.Command("bash", "-c", command)
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("error-path offered verifier = %v: %s", err, out)
	}
	var result struct {
		Check  string `json:"check"`
		Credit bool   `json:"hook_credit"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatal(err)
	}
	if result.Check == "gofmt" || !result.Credit {
		t.Errorf("error-path verification = %s, want passing package check only", out)
	}
}

func TestReceiptInvalidation(t *testing.T) {
	t.Parallel()
	newRecord := func(t *testing.T, check, config string, command ...string) (string, string) {
		t.Helper()
		dir, records := t.TempDir(), t.TempDir()
		file := filepath.Join(dir, "main.go")
		for name, content := range map[string]string{"go.mod": "module receipts\n\ngo 1.27\n", "main.go": "package receipts\n\nfunc Value() int { return 1 }\n"} {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if config != "" {
			config = filepath.Join(dir, config)
			if err := os.WriteFile(config, []byte("version: \"2\"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		args := append([]string{"run", records, check, file, config, "all", "run", ""}, command...)
		if code, out := receiptCommand(t, args...); code != 0 {
			t.Fatalf("receipt run = (%d, %q), want success", code, out)
		}
		return dir, filepath.Join(records, check+".json")
	}
	for _, changed := range []string{"other.go", "main_test.go", "go.mod", "go.sum", ".golangci.yml"} {
		t.Run(changed, func(t *testing.T) {
			t.Parallel()
			dir, record := newRecord(t, "vet", ".golangci.yml", "go", "vet", ".")
			if code, out := receiptCommand(t, "verify", record); code != 0 {
				t.Fatalf("verify unchanged = (%d, %q), want reusable", code, out)
			}
			if err := os.WriteFile(filepath.Join(dir, changed), []byte("changed inputs"), 0o644); err != nil {
				t.Fatal(err)
			}
			if code, out := receiptCommand(t, "verify", record); code == 0 {
				t.Fatalf("verify after %s change = (%d, %q), want stale", changed, code, out)
			}
		})
	}
	t.Run("inputs changed during command never become reusable", func(t *testing.T) {
		t.Parallel()
		_, record := newRecord(t, "vet", "", "bash", "-c", "printf 'package receipts\\n' > other.go")
		if code, out := receiptCommand(t, "verify", record); code == 0 {
			t.Fatalf("verify changed-during-check = (%d, %q), want rejected", code, out)
		}
	})
	t.Run("invalid receipt fails closed", func(t *testing.T) {
		t.Parallel()
		_, record := newRecord(t, "vet", "", "go", "vet", ".")
		if err := os.WriteFile(record, []byte("{\"status\":\"pass\""), 0o644); err != nil {
			t.Fatal(err)
		}
		if code, out := receiptCommand(t, "verify", record); code == 0 {
			t.Fatalf("verify truncated = (%d, %q), want rejected", code, out)
		}
	})
	t.Run("external local replacement changes invalidate", func(t *testing.T) {
		t.Parallel()
		dir, _ := newRecord(t, "vet", "", "go", "vet", ".")
		dep, records := t.TempDir(), t.TempDir()
		if err := os.WriteFile(filepath.Join(dep, "go.mod"), []byte("module example.com/dep\n\ngo 1.27\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		mod := "module receipts\n\ngo 1.27\n\nreplace example.com/dep => " + strconv.Quote(dep) + "\n"
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0o644); err != nil {
			t.Fatal(err)
		}
		file := filepath.Join(dir, "main.go")
		if code, out := receiptCommand(t, "run", records, "vet", file, "", "all", "run", "", "go", "vet", "."); code != 0 {
			t.Fatalf("replacement run = (%d, %q), want 0", code, out)
		}
		record := filepath.Join(records, "vet.json")
		if code, out := receiptCommand(t, "verify", record); code != 0 {
			t.Fatalf("replacement verify = (%d, %q), want current pass", code, out)
		}
		if err := os.WriteFile(filepath.Join(dep, "dep.go"), []byte("package dep\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if code, out := receiptCommand(t, "verify", record); code == 0 {
			t.Fatalf("replacement changed = (%d, %q), want stale", code, out)
		}
	})
	t.Run("workspace sibling inputs invalidate", func(t *testing.T) {
		t.Parallel()
		dir, _ := newRecord(t, "vet", "", "go", "vet", ".")
		dep, records, workDir := t.TempDir(), t.TempDir(), t.TempDir()
		if err := os.WriteFile(filepath.Join(dep, "go.mod"), []byte("module example.com/dep\n\ngo 1.27\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		work := filepath.Join(workDir, "go.work")
		physicalDir, err := filepath.EvalSymlinks(dir)
		if err != nil {
			t.Fatal(err)
		}
		physicalDep, err := filepath.EvalSymlinks(dep)
		if err != nil {
			t.Fatal(err)
		}
		content := "go 1.27\n\nuse (\n" + strconv.Quote(physicalDir) + "\n" + strconv.Quote(physicalDep) + "\n)\n"
		if err := os.WriteFile(work, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		env := []string{"GOWORK=" + work}
		if code, out := receiptCommandEnv(t, env, "run", records, "vet", filepath.Join(dir, "main.go"), "", "all", "run", "", "go", "vet", "."); code != 0 {
			t.Fatalf("workspace run = (%d, %q), want 0", code, out)
		}
		record := filepath.Join(records, "vet.json")
		if code, out := receiptCommandEnv(t, env, "verify", record); code != 0 {
			t.Fatalf("workspace verify = (%d, %q), want pass", code, out)
		}
		if err := os.WriteFile(filepath.Join(dep, "dep.go"), []byte("package dep\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if code, out := receiptCommandEnv(t, env, "verify", record); code == 0 {
			t.Fatalf("workspace changed = (%d, %q), want stale", code, out)
		}
	})
	t.Run("build environment changes invalidate", func(t *testing.T) {
		t.Parallel()
		_, record := newRecord(t, "vet", "", "go", "vet", ".")
		if code, out := receiptCommandEnv(t, []string{"GOFLAGS=-tags=receipt"}, "verify", record); code == 0 {
			t.Fatalf("changed flags = (%d, %q), want stale", code, out)
		}
	})
	t.Run("FIFO inputs do not block snapshot or allow reuse", func(t *testing.T) {
		t.Parallel()
		dir, _ := newRecord(t, "vet", "", "go", "vet", ".")
		if out, err := exec.Command("mkfifo", filepath.Join(dir, "runtime.pipe")).CombinedOutput(); err != nil {
			t.Fatalf("mkfifo = %v: %s", err, out)
		}
		records := t.TempDir()
		if code, out := receiptCommand(t, "run", records, "vet", filepath.Join(dir, "main.go"), "", "all", "run", "", "go", "vet", "."); code != 0 {
			t.Fatalf("FIFO run = (%d, %q), want completed check", code, out)
		}
		if code, out := receiptCommand(t, "verify", filepath.Join(records, "vet.json")); code == 0 {
			t.Fatalf("FIFO verify = (%d, %q), want incomplete snapshot", code, out)
		}
	})
}
