// The codex runner drives the OpenAI Codex CLI. It differs from the other three
// in the one place that matters most for reading its results: Codex has no skill
// tool. Skills reach the model as a listing in the system prompt, and a skill
// fires when the model decides to read its SKILL.md through the shell. So this
// runner measures a slightly different question than claude, copilot and
// opencode do — not "did the model call the skill tool" but "did the model go
// and read the skill" — and a session that never opens a SKILL.md is a control
// that happened to see a table of contents.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

// codexSkillPath matches the SKILL.md path of a go-* skill. It has to cover two
// shapes, because Codex names a skill one way in the prompt listing it renders
// ("(file: r0/go-http/SKILL.md)", a root alias with no skills/ segment) and
// another way in the shell command that reads it (the absolute path under
// $CODEX_HOME/skills). Anchoring on the skill's own directory covers both. The
// leading boundary keeps a directory that merely ends in the name, such as
// vendor-go-http, from matching.
var codexSkillPath = regexp.MustCompile(`(?:^|[^A-Za-z0-9_-])(go-[a-z0-9-]+)/SKILL\.md`)

var codexFrontmatter = regexp.MustCompile(`(?ms)^---\r?\n(.*?)^---[ \t]*\r?$`)
var codexSkillName = regexp.MustCompile(`(?m)^name:[ \t]*["']?(go-[a-z0-9-]+)["']?[ \t]*\r?$`)

// codexHomes gives every arm its own HOME and records it on the arm.
//
// CODEX_HOME alone is not enough. Codex also discovers skills from host
// locations outside its own home — a control arm run with only CODEX_HOME
// redirected still sees whatever the operator installed under ~/.claude,
// ~/.agents and the rest, and stops being a control. HOME is what closes that,
// and it costs nothing here because Codex keeps its credentials in a file this
// harness can copy, unlike copilot's credential store. The returned cleanup
// removes every home and is safe to call on error.
func codexHomes(arms []arm) (func(), error) {
	var homes []string
	cleanup := func() {
		for _, home := range homes {
			_ = os.RemoveAll(home) //nolint:errcheck // best-effort cleanup of temporary arm homes
		}
	}
	auth, err := codexAuth()
	if err != nil {
		return cleanup, err
	}
	for i, a := range arms {
		home, err := os.MkdirTemp("", "abrun-codex-")
		if err != nil {
			return cleanup, err
		}
		homes = append(homes, home)
		if err := writeCodexHome(home, a.dir, auth); err != nil {
			return cleanup, fmt.Errorf("prepare %s arm home: %w", a.Name, err)
		}
		if err := checkCodexSkills(home, a.Name, a.dir); err != nil {
			return cleanup, err
		}
		arms[i].home = home
	}
	return cleanup, nil
}

// codexAuth reads the credentials codex wrote at login. An arm home is a fresh
// HOME, so without this copy every run would fail unauthenticated. A missing
// file is not an error: the key can come from the environment.
func codexAuth() ([]byte, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("locate codex credentials: %w", err)
	}
	data, err := os.ReadFile(filepath.Join(home, ".codex", "auth.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read codex credentials: %w", err)
	}
	return data, nil
}

// codexHomeDir is the CODEX_HOME inside one arm home.
func codexHomeDir(home string) string { return filepath.Join(home, ".codex") }

// writeCodexHome lays out one arm home: the credentials and the arm's skills.
// armDir is the materialized plugin tree, or empty for the control arm, which
// gets a home with no skills directory at all.
func writeCodexHome(home, armDir string, auth []byte) error {
	dir := codexHomeDir(home)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	if auth != nil {
		if err := os.WriteFile(filepath.Join(dir, "auth.json"), auth, 0o600); err != nil {
			return err
		}
	}
	if armDir == "" {
		return nil
	}
	return os.CopyFS(filepath.Join(dir, "skills"), os.DirFS(filepath.Join(armDir, "skills")))
}

// checkCodexSkills asserts that an arm home puts exactly the skills that arm is
// supposed to have in front of the model, which is the one precondition the whole
// comparison rests on; checkArmSkills owns the comparison.
//
// `codex debug prompt-input` renders the prompt the model would actually see
// without spending a request, so the check is both free and the real thing
// rather than a directory listing that stands in for it.
func checkCodexSkills(home, armName, armDir string) error {
	out, err := codexCmd(codexSetupTimeout, home, home, "debug", "prompt-input", "list skills")
	if err != nil {
		return fmt.Errorf("render prompt for %s arm: %w", armName, err)
	}
	return checkArmSkills(armName, armDir, skillsInPaths(string(out)), "")
}

// codexSetupTimeout bounds the per-home prompt render that runs before the first
// session.
const codexSetupTimeout = 2 * time.Minute

// codexSession runs one headless session in work under the arm's home.
//
// The sandbox is workspace-write, so the session can edit the fixture and run
// the toolchain but cannot write outside the scratch tree; approval_policy=never
// is what keeps a headless run from blocking on a prompt. --ephemeral keeps
// concurrent runs from accumulating session files in one shared arm home.
func codexSession(o options, home, work, prompt string) ([]byte, error) {
	args := []string{
		"exec", "--json",
		"--cd", work,
		"--skip-git-repo-check",
		"--ephemeral",
		"--color", "never",
		"-s", "workspace-write",
		"-c", `approval_policy="never"`,
	}
	if o.model != "" {
		args = append(args, "-m", o.model)
	}
	if o.effort != "" {
		args = append(args, "-c", fmt.Sprintf("model_reasoning_effort=%q", o.effort))
	}
	args = append(args, prompt)
	return codexCmd(o.timeout, work, home, args...)
}

// codexCmd runs one codex invocation and returns everything it wrote to stdout.
// The transcript is captured through a temporary file rather than a pipe because
// a session's aggregated command output runs to megabytes.
func codexCmd(timeout time.Duration, dir, home string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	stdout, err := os.CreateTemp("", "abrun-transcript-")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = stdout.Close()           //nolint:errcheck // best-effort cleanup of the transcript buffer
		_ = os.Remove(stdout.Name()) //nolint:errcheck // best-effort cleanup of the transcript buffer
	}()
	var stderr strings.Builder
	cmd := exec.CommandContext(ctx, "codex", args...)
	cmd.Dir = dir
	cmd.Env = codexEnv(home, dir)
	cmd.Stdout = stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	out, err := os.ReadFile(stdout.Name())
	if err != nil {
		return nil, fmt.Errorf("read codex transcript: %w", err)
	}
	if ctx.Err() != nil {
		return out, fmt.Errorf("timed out after %s", timeout)
	}
	if runErr != nil {
		if message := strings.TrimSpace(stderr.String()); message != "" {
			return out, fmt.Errorf("codex: %w: %s", runErr, message)
		}
		return out, fmt.Errorf("codex: %w", runErr)
	}
	return out, nil
}

// codexEnv points every path codex derives from the environment at the arm's own
// home and the run's own scratch tree.
//
// HOME carries the skill isolation and one thing more: codex runs its shell as
// `zsh -lc`, so an inherited HOME would also load the operator's shell profile
// into every session. PWD has to be redirected alongside the process working
// directory so that no inherited value hands the session a path back to the
// repository, where the hidden golden test lives next to the fixtures.
func codexEnv(home, dir string) []string {
	return append(os.Environ(),
		"HOME="+home,
		"CODEX_HOME="+codexHomeDir(home),
		"PWD="+dir,
		"OLDPWD="+dir,
		"XDG_CONFIG_HOME="+filepath.Join(home, ".config"),
		"XDG_DATA_HOME="+filepath.Join(home, ".local", "share"),
		"XDG_STATE_HOME="+filepath.Join(home, ".local", "state"),
		"XDG_CACHE_HOME="+filepath.Join(home, ".cache"),
	)
}

// codexEvent is the part of one line of `codex exec --json` the benchmark reads.
type codexEvent struct {
	Type string `json:"type"`
	Item struct {
		Type     string          `json:"type"`
		Command  string          `json:"command"`
		Text     string          `json:"text"`
		ExitCode json.RawMessage `json:"exit_code"`
		Output   string          `json:"aggregated_output"`
	} `json:"item"`
}

// parseCodexStream pulls the go-* skills the model read and its last message out
// of a JSONL transcript.
//
// Evidence requires a completed successful command naming the SKILL.md path
// and matching frontmatter returned in its output. Path mentions, failed reads
// and cross-references alone do not count. Partial reads without frontmatter
// remain unverified. This is transcript evidence, not filesystem provenance:
// shell commands can synthesize output, and truncation can hide a real read.
//
// The cost return is always zero: codex reports token counts, not dollars, so
// there is no number here that means the same thing as the claude and opencode
// runners' $/run and a converted one would only look comparable.
func parseCodexStream(out []byte) (skills []string, final string, cost float64) {
	fired := map[string]bool{}
	for line := range strings.SplitSeq(string(out), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "{") {
			continue
		}
		var ev codexEvent
		if json.Unmarshal([]byte(line), &ev) != nil {
			continue
		}
		switch ev.Item.Type {
		case "command_execution":
			if ev.Type != "item.completed" || (string(ev.Item.ExitCode) != "0" && string(ev.Item.ExitCode) != `"0"`) {
				continue
			}
			paths := skillsInPaths(ev.Item.Command)
			for _, block := range codexFrontmatter.FindAllStringSubmatch(ev.Item.Output, -1) {
				for _, match := range codexSkillName.FindAllStringSubmatch(block[1], -1) {
					if slices.Contains(paths, match[1]) {
						fired[match[1]] = true
					}
				}
			}
		case "agent_message":
			if ev.Type == "item.completed" && ev.Item.Text != "" {
				final = ev.Item.Text
			}
		}
	}
	return sortedKeys(fired), final, 0
}

// codexCommands counts the shell calls in a transcript. A session that claims a
// line count in its final message and ran no command never measured one, and
// that difference is the whole reason the transcript is kept.
func codexCommands(out []byte) int {
	count := 0
	for line := range strings.SplitSeq(string(out), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "{") {
			continue
		}
		var ev codexEvent
		if json.Unmarshal([]byte(line), &ev) != nil {
			continue
		}
		// Codex reports a command twice, when it starts and when it finishes.
		// Counting the completion counts attempts, not events.
		if ev.Item.Type == "command_execution" && ev.Type == "item.completed" {
			count++
		}
	}
	return count
}

// skillsInPaths returns the go-* skills named by a SKILL.md path in text.
func skillsInPaths(text string) []string {
	found := map[string]bool{}
	for _, m := range codexSkillPath.FindAllStringSubmatch(text, -1) {
		if name := normalizeSkill(m[1]); name != "" {
			found[name] = true
		}
	}
	return sortedKeys(found)
}
