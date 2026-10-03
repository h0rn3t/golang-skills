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
