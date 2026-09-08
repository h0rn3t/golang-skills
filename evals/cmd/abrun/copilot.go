// The copilot runner drives the GitHub Copilot CLI, the third agent the plugin
// is published for. Everything downstream of the session — the fixtures, the
// structural metrics, the hidden golden test — is shared with the claude and
// opencode runners, so the three differ only in which CLI writes the code.
//
// Copilot is closer to claude than to opencode in one respect that matters for
// comparing them: --available-tools restricts the session to the same read and
// edit surface the claude arm gets, so neither runner can shell out and run
// `go test` on its own work. It is closer to opencode in the other: skills are
// discovered from a home directory rather than passed on a flag, so every arm
// needs its own.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// copilotTools is the tool set every copilot session is restricted to. It is the
// claude arm's Skill,Read,Glob,Grep,Edit,Write under copilot's names, so a
// difference between the two runners is not a difference in what the model was
// allowed to do. Leaving bash out is the load-bearing part: with a shell the
// session could run the fixture's tests, which no other arm can.
//
// Copilot names the read and write tools per model family, so the list carries
// every spelling of the same surface: create/edit/grep on the models that use
// them, apply_patch/rg on the ones that use those. A name the session does not
// have is reported as an unknown allowlist entry and ignored, while a missing
// name costs the session its editor — with only create/edit allowlisted,
// gpt-5.6-luna finishes every run explaining it has no way to write a file.
// Nothing here grants a shell, so the surface stays what it claims to be.
var copilotTools = []string{"skill", "view", "create", "edit", "grep", "glob", "apply_patch", "rg"}

// copilotHomes gives every arm its own COPILOT_HOME and records it on the arm.
//
// Copilot discovers personal skills from the home it is pointed at, so the
// control arm needs an isolated one just as much as the skilled arms do: run
// under the operator's own home it would load whatever go-* skills they have
// installed globally and stop being a control. HOME itself is left alone,
// because the credential store the CLI authenticates against is keyed to it and
// a redirected HOME fails the run with "No authentication information found".
// The returned cleanup removes every home and is safe to call on error.
func copilotHomes(arms []arm) (func(), error) {
	var homes []string
	cleanup := func() {
		for _, home := range homes {
			_ = os.RemoveAll(home) //nolint:errcheck // best-effort cleanup of temporary arm homes
		}
	}
	for i, a := range arms {
		home, err := os.MkdirTemp("", "abrun-copilot-")
		if err != nil {
			return cleanup, err
		}
		homes = append(homes, home)
		if err := writeCopilotHome(home, a.dir); err != nil {
			return cleanup, fmt.Errorf("prepare %s arm home: %w", a.Name, err)
		}
		if err := checkCopilotSkills(home, a.Name, a.dir); err != nil {
			return cleanup, err
		}
		arms[i].home = home
	}
	return cleanup, nil
}

// writeCopilotHome lays out one arm home by copying the arm's skills where
// copilot looks for personal ones. armDir is the materialized plugin tree, or
// empty for the control arm, which gets a home with no skills directory at all.
//
// Only skills are copied. The plugin's subagent and its gofmt/vet hook have no
// copilot equivalent here, which is a difference to state in any claim that
// spans runners rather than one to paper over.
func writeCopilotHome(home, armDir string) error {
	if armDir == "" {
		return nil
	}
	return os.CopyFS(filepath.Join(home, "skills"), os.DirFS(filepath.Join(armDir, "skills")))
}

// checkCopilotSkills asserts that an arm home loads exactly the skills that arm
// is supposed to have, which is the one precondition the whole comparison rests
// on. checkArmSkills owns the comparison; this function's job is to report what
// copilot actually loaded, and why it dropped anything.
//
// `copilot skill list` exits zero when a skill fails to load and names it on
// stderr, so the reason is captured and handed to the comparison rather than
// discarded.
func checkCopilotSkills(home, armName, armDir string) error {
	out, stderr, err := copilotOutput(copilotSetupTimeout, home, home, "skill", "list", "--json")
	if err != nil {
		return fmt.Errorf("list skills for %s arm: %w", armName, err)
	}
	var skills []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(out, &skills); err != nil {
		return fmt.Errorf("decode skills for %s arm: %w", armName, err)
	}
	var loaded []string
	for _, s := range skills {
		if strings.HasPrefix(s.Name, "go-") {
			loaded = append(loaded, s.Name)
		}
	}
	return checkArmSkills(armName, armDir, loaded, stderr)
}

// copilotSetupTimeout bounds the per-home skill listing that runs before the
// first session.
const copilotSetupTimeout = 2 * time.Minute

// copilotSession runs one headless session in work under the arm's home.
// --allow-all-tools is what non-interactive mode requires to proceed without a
// confirmation prompt, the role --permission-mode acceptEdits plays for the
// claude runner; it grants nothing beyond the six tools named above.
//
// The control arm loses the skill tool as well as the skills, so it cannot
// reach a builtin skill the way the claude control loses the Skill tool.
func copilotSession(o options, a arm, work, prompt string) ([]byte, error) {
	tools := copilotTools
	if a.dir == "" {
		tools = tools[1:]
	}
	args := []string{
		"-p", prompt,
		"--output-format", "json",
		"--allow-all-tools",
		"--available-tools", strings.Join(tools, ","),
		// The arms differ only in skill text, so the run must not pick up an
		// AGENTS.md, a built-in MCP server, or a CLI upgrade mid-corpus, and it
		// must not export the session off the machine.
		"--no-custom-instructions",
		"--disable-builtin-mcps",
		"--no-auto-update",
		"--no-remote",
		"--no-remote-export",
		"--no-color",
	}
	if o.model != "" {
		args = append(args, "--model", o.model)
	}
	if o.effort != "" {
		args = append(args, "--effort", o.effort)
	}
	return copilotCmd(o.timeout, work, a.home, args...)
}

// copilotCmd runs one copilot invocation and returns everything it wrote to
// stdout, discarding stderr on success the way a session does not need it.
func copilotCmd(timeout time.Duration, dir, home string, args ...string) ([]byte, error) {
	out, _, err := copilotOutput(timeout, dir, home, args...)
	return out, err
}

// copilotOutput runs one copilot invocation and returns stdout and stderr
// separately. The setup check needs both: `copilot skill list` reports a skill
// it could not parse on stderr and still exits zero, so a caller that only reads
// stdout cannot tell a complete listing from a truncated one.
//
// The transcript is captured through a temporary file rather than a pipe because
// a streaming session emits a few megabytes of tool-call deltas, and the part
// this harness reads — the skill calls, the final message — is spread across all
// of it.
func copilotOutput(timeout time.Duration, dir, home string, args ...string) ([]byte, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	stdout, err := os.CreateTemp("", "abrun-transcript-")
	if err != nil {
		return nil, "", err
	}
	defer func() {
		_ = stdout.Close()           //nolint:errcheck // best-effort cleanup of the transcript buffer
		_ = os.Remove(stdout.Name()) //nolint:errcheck // best-effort cleanup of the transcript buffer
	}()
	var stderr strings.Builder
	cmd := exec.CommandContext(ctx, "copilot", args...)
	cmd.Dir = dir
	cmd.Env = copilotEnv(home, dir)
	cmd.Stdout = stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	out, err := os.ReadFile(stdout.Name())
	if err != nil {
		return nil, stderr.String(), fmt.Errorf("read copilot transcript: %w", err)
	}
	if ctx.Err() != nil {
		return out, stderr.String(), fmt.Errorf("timed out after %s", timeout)
	}
	if runErr != nil {
		if message := strings.TrimSpace(stderr.String()); message != "" {
			return out, stderr.String(), fmt.Errorf("copilot: %w: %s", runErr, message)
		}
		return out, stderr.String(), fmt.Errorf("copilot: %w", runErr)
	}
	return out, stderr.String(), nil
}

// copilotEnv points copilot's configuration and state at the arm's own home and
// the run's own scratch tree. HOME stays as it is so the credential store keeps
// working; COPILOT_HOME is what decides which skills the session sees. PWD is
// overridden alongside the process working directory so that an inherited value
// cannot hand the session a path back to the repository, where the hidden golden
// test lives next to the fixtures.
func copilotEnv(home, dir string) []string {
	return append(os.Environ(),
		"COPILOT_HOME="+home,
		"PWD="+dir,
		"OLDPWD="+dir,
	)
}

// copilotEvent is the part of one line of `copilot --output-format json` the
// benchmark reads. Arguments stays raw because it is a different shape per tool
// and a decode failure on one tool must not drop the rest of the transcript.
type copilotEvent struct {
	Type string `json:"type"`
	Data struct {
		ToolName  string          `json:"toolName"`
		Arguments json.RawMessage `json:"arguments"`
		Content   string          `json:"content"`
	} `json:"data"`
}

// parseCopilotStream pulls the go-* skills the model loaded and its last message
// out of an NDJSON transcript.
//
// The cost return is always zero: copilot reports a session in premium requests
// and AI credits, not dollars, so there is no number here that means the same
// thing as the claude and opencode runners' $/run and a converted one would only
// look comparable. The summary prints the column for the arms that report a
// cost and leaves it empty for the ones that do not.
func parseCopilotStream(out []byte) (skills []string, final string, cost float64) {
	fired := map[string]bool{}
	for line := range strings.SplitSeq(string(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var ev copilotEvent
		if json.Unmarshal([]byte(line), &ev) != nil {
			continue
		}
		switch ev.Type {
		case "tool.execution_start":
			if ev.Data.ToolName != "skill" {
				continue
			}
			var in struct {
				Skill string `json:"skill"`
			}
			if json.Unmarshal(ev.Data.Arguments, &in) != nil {
				continue
			}
			if name := normalizeSkill(in.Skill); name != "" {
				fired[name] = true
			}
		case "assistant.message":
			if ev.Data.Content != "" {
				final = ev.Data.Content
			}
		}
	}
	return sortedKeys(fired), final, 0
}
