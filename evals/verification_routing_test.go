package evals_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPromptLoadsEverySelectedSkillInPrimaryAction(t *testing.T) {
	t.Parallel()
	cwd := filepath.Join(repoRoot(t), "evals", "ab", "_implement")
	_, out := promptEvent(t, t.TempDir(), "primary", cwd, "Implement the Go package in ./fetch")
	var primary string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "Before the first edit,") {
			primary = line
			break
		}
	}
	for _, owner := range []string{"go-code", "go-style-core", "go-http", "go-error-handling", "go-context", "go-resilience", "go-testing", "go-linting"} {
		if !strings.Contains(primary, "Skill tool, name `golang-skills:"+owner+"`") {
			t.Errorf("primary action = %q, want explicit Skill action for %s", primary, owner)
		}
	}
}

func TestVerificationRouting(t *testing.T) {
	t.Parallel()
	root, state, dir := repoRoot(t), t.TempDir(), t.TempDir()
	file := filepath.Join(dir, "main.go")
	for name, data := range map[string]string{"go.mod": "module verifyroute\n\ngo 1.27\n", "main.go": "package verifyroute\n\nfunc Value() int { return 1 }\n"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	routing := filepath.Join(root, "hooks", "go-code-routing.sh")
	guard := filepath.Join(root, "hooks", "go-verification-routing.sh")
	load := func(skill string) {
		t.Helper()
		hookEvent(t, routing, state, routingPayload("PostToolUse", "vroute", "Skill", map[string]any{"skill": skill}))
	}
	check := func(command string) (int, string) {
		t.Helper()
		return hookEvent(t, guard, state, routingPayload("PreToolUse", "vroute", "Bash", map[string]any{"command": command}))
	}
	load("go-code")
	if code, msg := check("go vet ./..."); code != 0 {
		t.Fatalf("no edit verification = (%d,%q), want no guard", code, msg)
	}
	code, _, msg := hookOutput(t, filepath.Join(root, "hooks", "go-vet-on-edit.sh"), state, routingPayload("PostToolUse", "vroute", "Write", map[string]any{"file_path": file}), "GOWORK=off", "GOLANG_SKILLS_EDIT_LINT=off")
	if code != 0 {
		t.Fatalf("prepare receipts = (%d,%q), want 0", code, msg)
	}
	if code, msg := check("gofmt -l .; go vet ./..."); code != 2 || !strings.Contains(msg, "go-linting") {
		t.Fatalf("missing gate owner = (%d,%q), want load block", code, msg)
	}
	load("go-linting")
	if code, msg := check("go test -race ./..."); code != 2 || !strings.Contains(msg, "go-check-receipt.sh") {
		t.Fatalf("unverified receipts = (%d,%q), want concrete verifier block", code, msg)
	}
	for _, command := range []string{"go -C /tmp vet .", "env -C /tmp go vet .", "timeout -k 10 60 go vet .", "go vet .; bash " + filepath.Join(root, "hooks", "go-check-receipt.sh") + " --gate /missing ."} {
		if code, msg := check(command); code != 2 {
			t.Errorf("verification bypass %q = (%d,%q), want blocked", command, code, msg)
		}
	}
	for _, command := range []string{"go version", "echo 'go vet ./...'", "cat main.go"} {
		if code, msg := check(command); code != 0 {
			t.Errorf("inventory command %q = (%d,%q), want allowed", command, code, msg)
		}
	}
	paths, err := filepath.Glob(filepath.Join(state, "checks", "run.*"))
	if err != nil || len(paths) != 1 {
		t.Fatalf("receipt dirs = %v,%v, want one", paths, err)
	}
	recipe := "bash " + filepath.Join(root, "hooks", "go-check-receipt.sh") + " --gate " + paths[0] + " " + dir
	for _, command := range []string{"false && " + recipe + "; go vet .", "true || " + recipe + "; go vet .", recipe + " & go vet .", recipe + "; go vet ."} {
		if code, msg := check(command); code != 2 {
			t.Errorf("mixed prerequisite %q = (%d,%q), want standalone attempt first", command, code, msg)
		}
	}
	for _, command := range []string{
		recipe + " | tail",
		recipe + " | tail -n 40",
		recipe + " 2>&1 | tail",
		recipe + " >/tmp/golang-skills-out",
		recipe + " > /tmp/golang-skills-out",
		recipe + ` >"/tmp/my log"`,
		recipe + ` >'/tmp/my log'`,
		recipe + ` >/tmp/my\ log`,
		recipe + " 2>&1|tail -n 40",
		recipe + " &>/tmp/log",
		recipe + " >>/tmp/log",
		recipe + " >/tmp/log|tail -n 40",
	} {
		if code, msg := check(command); code != 0 {
			t.Errorf("display suffix %q = (%d,%q), want the verifier allowed", command, code, msg)
		}
	}
	for _, command := range []string{
		recipe + " 2>&1 | tail; go vet .",
		"go vet . 2>&1 | tail",
		"go test -race ./... | tail",
		`go vet . >"/tmp/my log"`,
		`bash -c "go vet . >/tmp/log"`,
		`echo ' > '; go vet .`,
		recipe + ` >"/tmp/my log"; go vet .`,
		recipe + " >/tmp/log;go vet .",
		recipe + " >/tmp/log&&go vet .",
	} {
		if code, msg := check(command); code != 2 {
			t.Errorf("piped check %q = (%d,%q), want blocked until a standalone verifier", command, code, msg)
		}
	}
	cmd := exec.Command("bash", filepath.Join(root, "hooks", "go-check-receipt.sh"), "--gate", paths[0], dir)
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gate verifier = %v: %s", err, out)
	}
	var result struct {
		Checks []struct {
			Check  string `json:"check"`
			Credit bool   `json:"hook_credit"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("gate JSON = %s: %v", out, err)
	}
	credits := map[string]bool{}
	for _, row := range result.Checks {
		credits[row.Check] = row.Credit
	}
	if !credits["vet"] || !credits["fix"] || credits["gofmt"] || credits["test"] || credits["lint"] {
		t.Errorf("credits = %v, want package vet/fix only", credits)
	}
	if code, msg := check("go vet . && go test -race ."); code != 0 {
		t.Fatalf("verified generation = (%d,%q), want direct checks allowed", code, msg)
	}
	stop := func(message string) string {
		t.Helper()
		payload := routingPayload("Stop", "vroute", "", nil)
		payload["last_assistant_message"] = message
		_, out, _ := hookOutput(t, guard, state, payload, "GOWORK=off")
		return out
	}
	if out := stop("checks: vet pass (hook) · lint pass (hook)"); !strings.Contains(out, `"decision": "block"`) || !strings.Contains(out, "lint") {
		t.Errorf("unsupported final credit = %q, want corrective block", out)
	}
	if out := stop("checks: vet pass (hook) · lint observed (hook, reuse unverified)"); strings.Contains(out, `"decision": "block"`) {
		t.Errorf("qualified final evidence = %q, want allowed", out)
	}
	if code, out := receiptCommand(t, "verify", filepath.Join(paths[0], "vet.json"), dir, "go", "vet", "."); code != 0 {
		t.Fatalf("individual vet = (%d,%q), want success", code, out)
	}
	if out := stop("checks: vet pass (hook) · fix pass (hook)"); strings.Contains(out, `"decision": "block"`) {
		t.Errorf("same-generation credits = %q, want retained", out)
	}
	if err := os.WriteFile(file, []byte("package verifyroute\n\nfunc Value() int { return 2 }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out := stop("checks: vet pass (hook)"); !strings.Contains(out, `"decision": "block"`) {
		t.Errorf("untracked code mutation = %q, want stale credit rejected", out)
	}
	hookOutput(t, filepath.Join(root, "hooks", "go-vet-on-edit.sh"), state, routingPayload("PostToolUse", "vroute", "Edit", map[string]any{"file_path": file}), "GOWORK=off", "GOLANG_SKILLS_EDIT_LINT=off")
	if code, msg := check("go vet ."); code != 2 {
		t.Fatalf("later edit = (%d,%q), want verifier again", code, msg)
	}
	// An old run can never release the guard for the later edit.
	cmd = exec.Command("bash", filepath.Join(root, "hooks", "go-check-receipt.sh"), "--gate", paths[0], dir)
	cmd.Env = append(os.Environ(), "GOWORK=off")
	_ = cmd.Run()
	if code, msg := check("go vet ."); code != 2 {
		t.Fatalf("old verifier = (%d,%q), want newer generation still blocked", code, msg)
	}
}

// TestVerificationStatePerAgent: a Go edit by one agent asks that agent for
// the verifier attempt, not the main thread or a sibling, which share its
// session_id.
func TestVerificationStatePerAgent(t *testing.T) {
	t.Parallel()
	root, state, dir := repoRoot(t), t.TempDir(), t.TempDir()
	file := filepath.Join(dir, "main.go")
	for name, data := range map[string]string{"go.mod": "module peragent\n\ngo 1.27\n", "main.go": "package peragent\n\nfunc Value() int { return 1 }\n"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	guard := filepath.Join(root, "hooks", "go-verification-routing.sh")
	as := func(p map[string]any, agent string) map[string]any {
		if agent != "" {
			p["agent_id"], p["agent_type"] = agent, "general-purpose"
		}
		return p
	}
	for _, agent := range []string{"", "a1", "a2"} {
		for _, skill := range []string{"go-code", "go-linting"} {
			hookEvent(t, filepath.Join(root, "hooks", "go-code-routing.sh"), state, as(routingPayload("PostToolUse", "peragent", "Skill", map[string]any{"skill": skill}), agent))
		}
	}
	if code, _, msg := hookOutput(t, filepath.Join(root, "hooks", "go-vet-on-edit.sh"), state, as(routingPayload("PostToolUse", "peragent", "Write", map[string]any{"file_path": file}), "a1"), "GOWORK=off", "GOLANG_SKILLS_EDIT_LINT=off"); code != 0 {
		t.Fatalf("subagent edit hook = (%d,%q), want 0", code, msg)
	}
	check := func(agent string) int {
		t.Helper()
		code, _ := hookEvent(t, guard, state, as(routingPayload("PreToolUse", "peragent", "Bash", map[string]any{"command": "go vet ./..."}), agent))
		return code
	}
	if code := check("a1"); code != 2 {
		t.Errorf("check by the editing subagent = %d, want 2 until its verifier attempt", code)
	}
	for _, agent := range []string{"", "a2"} {
		if code := check(agent); code != 0 {
			t.Errorf("check by %q after a1's edit = %d, want 0: the edit is not theirs", agent, code)
		}
	}
}

func TestSpeedClaimGate(t *testing.T) {
	t.Parallel()
	root, state, dir := repoRoot(t), t.TempDir(), t.TempDir()
	file, bench := filepath.Join(dir, "main.go"), filepath.Join(dir, "main_bench_test.go")
	for name, data := range map[string]string{
		"go.mod":             "module speedgate\n\ngo 1.27\n",
		"main.go":            "package speedgate\n\nfunc Value() int { return 1 }\n",
		"main_bench_test.go": "package speedgate\n\nimport \"testing\"\n\nfunc BenchmarkValue(b *testing.B) {\n\tfor b.Loop() {\n\t\tValue()\n\t}\n}\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	guard := filepath.Join(root, "hooks", "go-verification-routing.sh")
	stop := func(message string, env ...string) string {
		t.Helper()
		payload := routingPayload("Stop", "speed", "", nil)
		payload["last_assistant_message"] = message
		_, out, _ := hookOutput(t, guard, state, payload, append(env, "GOWORK=off")...)
		return out
	}
	hookEvent(t, filepath.Join(root, "hooks", "go-code-routing.sh"), state, routingPayload("PostToolUse", "speed", "Skill", map[string]any{"skill": "go-code"}))
	if out := stop("strings.Builder is faster than += in a loop."); out != "" {
		t.Errorf("answer before any Go edit = %q, want allowed", out)
	}
	edit := func(path string) {
		t.Helper()
		if code, _, msg := hookOutput(t, filepath.Join(root, "hooks", "go-vet-on-edit.sh"), state, routingPayload("PostToolUse", "speed", "Write", map[string]any{"file_path": path}), "GOWORK=off", "GOLANG_SKILLS_EDIT_LINT=off"); code != 0 {
			t.Fatalf("edit hook on %s = (%d,%q), want 0", path, code, msg)
		}
	}
	edit(bench)
	if out := stop("SHA hardware support would hash several times faster."); out != "" {
		t.Errorf("advice after a benchmark-only edit = %q, want allowed", out)
	}
	edit(file)
	for _, claim := range []string{
		"Run is now about 3x faster.",
		"BenchmarkRun: sec/op −42%, allocs/op −99%",
		"latency drops by -30% on the hot path",
		"The arena gives a 2.1x speedup.",
	} {
		if out := stop(claim); !strings.Contains(out, `"decision": "block"`) || !strings.Contains(out, "--verdict") {
			t.Errorf("unmeasured claim %q = %q, want speed block", claim, out)
		}
	}
	for _, report := range []string{
		"speedup: skipped (no verdict)",
		"Not measured: the change should be faster, but the comparison was skipped.",
		"hex.EncodeToString would give the same names but no measurable speedup.",
		"checks: gofmt pass · vet pass",
	} {
		if out := stop(report); out != "" {
			t.Errorf("report %q = %q, want allowed", report, out)
		}
	}
	if out := stop("Run is now 3x faster.", "GOLANG_SKILLS_SPEED_GATE=off"); out != "" {
		t.Errorf("gate off = %q, want allowed", out)
	}
	// bench-compare.sh rejects --verdict=DIR as an unknown option, so it runs no verdict.
	equals := "bash " + filepath.Join(root, "skills", "go-performance", "scripts", "bench-compare.sh") + " --verdict=/tmp/perf-snap ./pkg"
	hookEvent(t, guard, state, routingPayload("PreToolUse", "speed", "Bash", map[string]any{"command": equals}))
	if out := stop("Run is now 3x faster."); !strings.Contains(out, `"decision": "block"`) {
		t.Errorf("claim after --verdict=DIR = %q, want speed block", out)
	}
	verdict := "bash " + filepath.Join(root, "skills", "go-performance", "scripts", "bench-compare.sh") + " --verdict /tmp/perf-snap ./pkg 2>&1 | tail -5"
	if code, msg := hookEvent(t, guard, state, routingPayload("PreToolUse", "speed", "Bash", map[string]any{"command": verdict})); code != 0 {
		t.Fatalf("verdict command = (%d,%q), want allowed", code, msg)
	}
	if out := stop("BenchmarkRun: sec/op −42% (keep)"); out != "" {
		t.Errorf("claim after a verdict = %q, want allowed", out)
	}
}
