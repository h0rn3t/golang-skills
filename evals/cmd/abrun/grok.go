// The grok runner drives `grok -p`. Grok has no --plugin-dir on that path and
// no Skill tool: a skill counts when a read_file result returns every line of
// that skill's SKILL.md. Hooks from this plugin are not installed. The session
// already has a shell, so -shell is not accepted; a review run has the edit
// and shell tools removed so it cannot change the fixture.
//
// Skills are discovered from GROK_HOME and from the real HOME (~/.claude,
// ~/.cursor, ~/.grok). Both are replaced per arm. The operator's bundled
// skills are copied into every arm, including the control, so they are not a
// difference between arms. Claude and Cursor skill compatibility is off, or
// the control arm would still see the operator's go-* skills.
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

// grokConfig is written into every arm's GROK_HOME. It pins what the arms are
// not allowed to differ in: no leader shared with the operator, no memory, no
// workflows, and no scan of Claude or Cursor skill directories.
const grokConfig = `[cli]
use_leader = false

[compat.claude]
skills = false
agents = false
hooks = false
mcps = false
rules = false

[compat.cursor]
skills = false

[memory_v2]
enabled = false

[workflows]
enabled = false
`

// grokReviewDenied removes the tools a review session could use to change the
// fixture. Grok accepts both the tool id and the name the transcript shows.
const grokReviewDenied = "search_replace,write,run_terminal_command,run_terminal_cmd"

// grokSetupTimeout bounds the per-home `grok inspect` that checks skill discovery.
const grokSetupTimeout = 2 * time.Minute

// grokHomes gives every arm its own HOME and GROK_HOME and records HOME on the
// arm. The returned cleanup removes every home and is safe to call on error.
func grokHomes(arms []arm) (func(), error) {
	var homes []string
	cleanup := func() {
		for _, home := range homes {
			_ = os.RemoveAll(home) //nolint:errcheck // best-effort cleanup of temporary arm homes
		}
	}
	auth, err := grokAuth()
	if err != nil {
		return cleanup, err
	}
	for i, a := range arms {
		home, err := os.MkdirTemp("", "abrun-grok-")
		if err != nil {
			return cleanup, err
		}
		homes = append(homes, home)
		if err := writeGrokHome(home, a.dir, auth); err != nil {
			return cleanup, fmt.Errorf("prepare %s arm home: %w", a.Name, err)
		}
		if err := checkGrokSkills(home, a.Name, a.dir); err != nil {
			return cleanup, err
		}
		arms[i].home = home
	}
	return cleanup, nil
}

// grokAuth reads the credentials grok wrote at login. An arm home is a fresh
// GROK_HOME, so without this copy every run would fail unauthenticated. A
// missing file is not an error: the key can come from the environment.
func grokAuth() ([]byte, error) {
	root, err := grokOperatorRoot()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(root, "auth.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read grok credentials: %w", err)
	}
	return data, nil
}

// grokOperatorRoot is the GROK_HOME the operator is logged in under.
func grokOperatorRoot() (string, error) {
	if dir := os.Getenv("GROK_HOME"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate grok credentials: %w", err)
	}
	return filepath.Join(home, ".grok"), nil
}

// grokHomeDir is the GROK_HOME inside one arm home.
func grokHomeDir(home string) string { return filepath.Join(home, ".grok") }

// writeGrokHome lays out one arm home: the config, the credentials, the
// operator's bundled skills, and the arm's skills. armDir is empty for the
// control arm, which gets no go-* skills directory.
func writeGrokHome(home, armDir string, auth []byte) error {
	dir := grokHomeDir(home)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(grokConfig), 0o644); err != nil {
		return err
	}
	if auth != nil {
		if err := os.WriteFile(filepath.Join(dir, "auth.json"), auth, 0o600); err != nil {
			return err
		}
	}
	if err := copyGrokBundled(dir); err != nil {
		return err
	}
	if armDir == "" {
		return nil
	}
	return os.CopyFS(filepath.Join(dir, "skills"), os.DirFS(filepath.Join(armDir, "skills")))
}

// copyGrokBundled copies the operator's bundled skill cache into one arm.
// Grok materializes that cache into GROK_HOME on first use; doing it here
// keeps concurrent runs from racing on the same extract, and puts the same
// non-go skills in front of every arm. A missing cache is left for the CLI.
func copyGrokBundled(dst string) error {
	root, err := grokOperatorRoot()
	if err != nil {
		return err
	}
	src := filepath.Join(root, "bundled")
	info, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}
	bundled := filepath.Join(dst, "bundled")
	if err := os.MkdirAll(bundled, 0o755); err != nil {
		return err
	}
	if err := os.CopyFS(bundled, os.DirFS(src)); err != nil {
		return fmt.Errorf("copy grok bundled skills: %w", err)
	}
	return nil
}

// checkGrokSkills asserts that an arm home puts exactly the go-* skills that
// arm is supposed to have in front of the model. `grok inspect` is the listing
// the session uses, so the check is the real discovery path.
func checkGrokSkills(home, armName, armDir string) error {
	out, err := grokCmd(grokSetupTimeout, home, home, "inspect", "--json")
	if err != nil {
		return fmt.Errorf("list skills for %s arm: %w", armName, err)
	}
	var decoded struct {
		Skills []struct {
			Name string `json:"name"`
		} `json:"skills"`
	}
	if err := json.Unmarshal(out, &decoded); err != nil {
		return fmt.Errorf("decode skills for %s arm: %w", armName, err)
	}
	var loaded []string
	for _, s := range decoded.Skills {
		if name := normalizeSkill(s.Name); name != "" {
			loaded = append(loaded, name)
		}
	}
	return checkArmSkills(armName, armDir, loaded, "")
}

// grokSession runs one headless session in work under the arm's home.
// --always-approve and bypassPermissions keep a headless run from blocking on
// a prompt. A private leader socket keeps the run off the operator's leader.
func grokSession(o options, home, work, prompt string) ([]byte, error) {
	sock, err := os.CreateTemp("", "abrun-grok-*.sock")
	if err != nil {
		return nil, err
	}
	sockName := sock.Name()
	if err := sock.Close(); err != nil {
		return nil, err
	}
	if err := os.Remove(sockName); err != nil {
		return nil, err
	}
	defer func() {
		_ = os.Remove(sockName)             //nolint:errcheck // best-effort cleanup of the session socket
		_ = os.Remove(sockName + ".daemon") //nolint:errcheck // best-effort cleanup of the session socket
	}()
	args := []string{
		"-p", prompt,
		"--output-format", "streaming-messages-json",
		"--max-turns", fmt.Sprint(maxSteps),
		"--always-approve",
		"--permission-mode", "bypassPermissions",
		"--no-subagents",
		"--disable-web-search",
		"--cwd", work,
		"--leader-socket", sockName,
	}
	if o.corpus == corpusReview {
		args = append(args, "--disallowed-tools", grokReviewDenied)
	}
	if o.model != "" {
		args = append(args, "--model", o.model)
	}
	if o.effort != "" {
		args = append(args, "--reasoning-effort", o.effort)
	}
	return grokCmd(o.timeout, work, home, args...)
}

// grokCmd runs one grok invocation and returns everything it wrote to stdout.
// The transcript is captured through a temporary file rather than a pipe
// because a session's tool results run to megabytes.
func grokCmd(timeout time.Duration, dir, home string, args ...string) ([]byte, error) {
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
	cmd := exec.CommandContext(ctx, "grok", args...)
	cmd.Dir = dir
	cmd.Env = grokEnv(home, dir)
	cmd.Stdout = stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	out, err := os.ReadFile(stdout.Name())
	if err != nil {
		return nil, fmt.Errorf("read grok transcript: %w", err)
	}
	if ctx.Err() != nil {
		return out, fmt.Errorf("timed out after %s", timeout)
	}
	if runErr != nil {
		if message := strings.TrimSpace(stderr.String()); message != "" {
			return out, fmt.Errorf("grok: %w: %s", runErr, message)
		}
		return out, fmt.Errorf("grok: %w", runErr)
	}
	return out, nil
}

// grokEnv points grok's configuration and skill discovery at the arm's own
// home and the run's own scratch tree. PWD is overridden alongside the process
// working directory so an inherited value cannot hand the session a path back
// to the repository, where the hidden golden test lives next to the fixtures.
func grokEnv(home, dir string) []string {
	return append(os.Environ(),
		"HOME="+home,
		"GROK_HOME="+grokHomeDir(home),
		"GROK_MEMORY=0",
		"GROK_WORKFLOWS=0",
		"GROK_CLAUDE_SKILLS_ENABLED=false",
		"GROK_CURSOR_SKILLS_ENABLED=false",
		"PWD="+dir,
		"OLDPWD="+dir,
	)
}

// grokReadFile is the JSON a read_file tool result carries in its content.
type grokReadFile struct {
	FileContent struct {
		RawOutput    string `json:"raw_output"`
		TotalLines   int    `json:"total_lines"`
		AbsolutePath string `json:"absolute_path"`
	} `json:"FileContent"`
}

// parseGrokStream pulls the go-* skills the model read in full, its final
// message, the reported cost, and the shell-command count from a
// streaming-messages-json transcript.
//
// A Skill tool call counts too, the same way the claude parser counts one. A
// read counts only when the tool result is not an error, the returned text has
// total_lines lines, and the returned path is that skill's SKILL.md. A head
// or an offset read stays out: a partial read is not evidence the model saw
// the skill.
func parseGrokStream(out []byte) (skills []string, final string, cost float64, commands int) {
	fired := map[string]bool{}
	pending := map[string]string{}
	for line := range strings.SplitSeq(string(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var v any
		if json.Unmarshal([]byte(line), &v) != nil {
			continue
		}
		for _, name := range skillCalls(v) {
			fired[name] = true
		}
		obj, ok := v.(map[string]any)
		if !ok {
			continue
		}
		if text, ok := obj["result"].(string); ok && text != "" {
			final = text
		}
		if usd, ok := obj["total_cost_usd"].(float64); ok {
			cost = usd
		}
		switch obj["type"] {
		case "assistant":
			grokNoteReads(obj, pending)
			commands += grokCommandCalls(obj)
		case "user":
			grokNoteReadResults(obj, pending, fired)
		}
	}
	return sortedKeys(fired), final, cost, commands
}

// grokNoteReads records read_file calls whose path is a go-* SKILL.md, keyed
// by tool-use id, so the matching result can accept or reject them.
func grokNoteReads(obj map[string]any, pending map[string]string) {
	for _, block := range grokContent(obj) {
		if block["type"] != "tool_use" {
			continue
		}
		name, _ := block["name"].(string)
		if name != "read_file" && name != "Read" {
			continue
		}
		in, _ := block["input"].(map[string]any)
		skill, ok := grokSkillPath(in)
		if !ok {
			continue
		}
		id, _ := block["id"].(string)
		if id == "" {
			continue
		}
		pending[id] = skill
	}
}

// grokNoteReadResults accepts a pending read whose result returned the whole
// file at that skill's path.
func grokNoteReadResults(obj map[string]any, pending map[string]string, fired map[string]bool) {
	for _, block := range grokContent(obj) {
		if block["type"] != "tool_result" {
			continue
		}
		id, _ := block["tool_use_id"].(string)
		skill, ok := pending[id]
		if !ok {
			continue
		}
		delete(pending, id)
		if errFlag, _ := block["is_error"].(bool); errFlag {
			continue
		}
		if grokReadCovers(block["content"], skill) {
			fired[skill] = true
		}
	}
}

// grokCommandCalls counts shell tool calls. Grok names the tool
// run_terminal_command in the transcript and run_terminal_cmd as its id.
func grokCommandCalls(obj map[string]any) int {
	n := 0
	for _, block := range grokContent(obj) {
		if block["type"] != "tool_use" {
			continue
		}
		name, _ := block["name"].(string)
		if name == "run_terminal_command" || name == "run_terminal_cmd" {
			n++
		}
	}
	return n
}

// grokContent returns the message content blocks of one stream event.
func grokContent(obj map[string]any) []map[string]any {
	msg, _ := obj["message"].(map[string]any)
	raw, _ := msg["content"].([]any)
	var blocks []map[string]any
	for _, item := range raw {
		block, ok := item.(map[string]any)
		if ok {
			blocks = append(blocks, block)
		}
	}
	return blocks
}

// grokSkillPath returns the go-* skill named by a read_file input. The keys
// are the ones grok and the messages-shaped stream have used for the path.
func grokSkillPath(in map[string]any) (string, bool) {
	if in == nil {
		return "", false
	}
	var path string
	for _, key := range []string{"target_file", "path", "file_path"} {
		s, ok := in[key].(string)
		if ok && s != "" {
			path = s
			break
		}
	}
	m := codexSkillPath.FindStringSubmatch(path)
	if m == nil {
		return "", false
	}
	return m[1], true
}

// grokReadCovers reports whether a tool result returned every line of skill's
// SKILL.md. content is either the JSON text grok puts in the result or that
// object already decoded.
func grokReadCovers(content any, skill string) bool {
	var raw []byte
	switch c := content.(type) {
	case string:
		raw = []byte(c)
	case map[string]any:
		encoded, err := json.Marshal(c)
		if err != nil {
			return false
		}
		raw = encoded
	default:
		return false
	}
	var doc grokReadFile
	if json.Unmarshal(raw, &doc) != nil {
		return false
	}
	body := doc.FileContent
	if body.TotalLines <= 0 || textLines(body.RawOutput) != body.TotalLines {
		return false
	}
	if body.AbsolutePath == "" {
		return false
	}
	got, ok := grokSkillPath(map[string]any{"path": body.AbsolutePath})
	return ok && got == skill
}

// textLines counts the lines in s the way a read_file total_lines does: a
// trailing newline does not add a line, and an empty string has none.
func textLines(s string) int {
	if s == "" {
		return 0
	}
	n := strings.Count(s, "\n")
	if !strings.HasSuffix(s, "\n") {
		n++
	}
	return n
}
