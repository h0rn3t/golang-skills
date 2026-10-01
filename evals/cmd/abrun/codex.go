// Codex не має skill tool: runner підтверджує повний captured output читання
// SKILL.md за snapshot staged-копії arm. Це не доводить retention моделлю.
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
	"strconv"
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

// codexSkillEvidence відрізняє порожній inventory від недоступного та старих звітів.
type codexSkillEvidence struct {
	InventoryStatus string           `json:"inventory_status"`
	InventoryReason string           `json:"inventory_reason,omitempty"`
	Reads           []codexSkillRead `json:"reads,omitempty"`
}

// codexSkillRead описує captured output, а не завантаження або retention моделлю.
type codexSkillRead struct {
	ItemID        string `json:"item_id,omitempty"`
	Skill         string `json:"skill"`
	Path          string `json:"path,omitempty"`
	Command       string `json:"command"`
	Status        string `json:"status"`
	Reason        string `json:"reason"`
	ReturnedBytes int    `json:"command_output_bytes"`
	ExpectedBytes int    `json:"expected_file_bytes"`
	// MatchedBytes — точний збіг bytes конкретного файла; частковий head/sed
	// може містити внутрішній фрагмент. Загальний output команди рахується окремо.
	MatchedBytes int `json:"matched_file_bytes"`
}

// codexSkillInventory читає довірену installed-копію до запуску CLI.
// Aliases походять тільки з filesystem staged home, ніколи з transcript.
func codexSkillInventory(a arm) (map[string]string, error) {
	inventory := map[string]string{}
	if a.dir == "" {
		return inventory, nil
	}
	if a.home == "" {
		return nil, fmt.Errorf("codex arm home is missing")
	}
	root := filepath.Join(codexHomeDir(a.home), "skills")
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read staged codex skills: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "go-") {
			continue
		}
		path := filepath.Join(root, entry.Name(), "SKILL.md")
		body, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("snapshot staged skill: %w", err)
		}
		if len(body) == 0 {
			return nil, fmt.Errorf("staged skill %q is empty", path)
		}
		physical, err := filepath.EvalSymlinks(path)
		if err != nil {
			return nil, fmt.Errorf("resolve staged skill: %w", err)
		}
		inventory[path] = string(body)
		if filepath.Base(physical) == "SKILL.md" && filepath.Base(filepath.Dir(physical)) == entry.Name() {
			inventory[physical] = string(body)
		}
	}
	if len(inventory) == 0 {
		return nil, fmt.Errorf("staged codex arm has no Go skills")
	}
	return inventory, nil
}

// codexEvent містить лише потрібні поля JSONL lifecycle та command output.
type codexEvent struct {
	Type string `json:"type"`
	Item struct {
		ID       string          `json:"id"`
		Type     string          `json:"type"`
		Command  string          `json:"command"`
		Text     string          `json:"text"`
		Status   string          `json:"status"`
		ExitCode json.RawMessage `json:"exit_code"`
		Output   string          `json:"aggregated_output"`
	} `json:"item"`
}

// parseCodexStream зараховує тільки успішне completed читання повного output.
// Token usage не конвертується в dollars; cost залишається нульовим.
func parseCodexStream(out []byte, inventory map[string]string) (skills []string, final string, cost float64, evidence []codexSkillRead) {
	fired := map[string]bool{}
	completed := map[string]bool{}
	pending := map[string]codexEvent{}
	var order []string
	for line := range strings.SplitSeq(string(out), "\n") {
		var ev codexEvent
		if json.Unmarshal([]byte(line), &ev) != nil {
			continue
		}
		if ev.Item.Type == "agent_message" {
			if ev.Type == "item.completed" && ev.Item.Text != "" {
				final = ev.Item.Text
			}
			continue
		}
		if ev.Item.Type != "command_execution" {
			continue
		}
		key := ev.Item.ID
		if key == "" {
			key = line
		}
		if ev.Type == "item.started" {
			if _, ok := pending[key]; !ok && !completed[key] {
				pending[key] = ev
				order = append(order, key)
			}
			continue
		}
		if ev.Type != "item.completed" || completed[key] {
			continue
		}
		completed[key] = true
		delete(pending, key)
		reads := codexReadEvidence(ev, inventory)
		evidence = append(evidence, reads...)
		for _, read := range reads {
			if read.Status == "full_returned_output" {
				fired[read.Skill] = true
			}
		}
	}
	for _, key := range order {
		if ev, ok := pending[key]; ok {
			evidence = append(evidence, codexReadEvidence(ev, inventory)...)
		}
	}
	return sortedKeys(fired), final, 0, evidence
}

// codexReadEvidence зв'язує body з operand path та порядком multi-file cat.
func codexReadEvidence(ev codexEvent, inventory map[string]string) []codexSkillRead {
	words, ok := codexLiteralWords(ev.Item.Command)
	if ok && len(words) == 3 && slices.Contains([]string{"/bin/zsh", "zsh", "/bin/bash", "bash", "/bin/sh", "sh"}, words[0]) && words[1] == "-lc" {
		words, ok = codexLiteralWords(words[2])
	}
	var paths []string
	command := ""
	if ok && len(words) > 1 {
		command = filepath.Base(words[0])
		ok = words[0] == command || words[0] == "/bin/"+command || words[0] == "/usr/bin/"+command
	}
	if ok && len(words) > 1 {
		args := words[1:]
		switch command {
		case "cat":
			if args[0] == "--" {
				args = args[1:]
			}
			for _, arg := range args {
				if strings.HasPrefix(arg, "-") {
					ok = false
					break
				}
			}
			if ok {
				paths = args
			}
		case "head":
			if len(args) == 3 && (args[0] == "-n" || args[0] == "-c") {
				n, err := strconv.Atoi(args[1])
				ok = err == nil && n >= 0
				args = args[2:]
			} else if len(args) == 2 && strings.HasPrefix(args[0], "-") {
				n, err := strconv.Atoi(strings.TrimPrefix(args[0], "-"))
				ok = err == nil && n >= 0
				args = args[1:]
			}
			if ok && len(args) == 1 && !strings.HasPrefix(args[0], "-") {
				paths = args
			}
		case "sed":
			if len(args) == 3 && args[0] == "-n" {
				expr, ends := strings.CutSuffix(args[1], "p")
				ok = ends && strings.Count(expr, ",") <= 1
				for part := range strings.SplitSeq(expr, ",") {
					if part == "$" {
						continue
					}
					for _, digit := range part {
						if digit < '0' || digit > '9' {
							ok = false
						}
					}
					n, err := strconv.Atoi(part)
					if err != nil || n <= 0 {
						ok = false
					}
				}
				if ok {
					paths = args[2:]
				}
			}
		}
	}
	supported := ok && len(paths) > 0 && (command == "cat" || command == "head" || command == "sed")
	for _, path := range paths {
		if filepath.Base(path) != "SKILL.md" {
			supported = false
		}
	}
	if !supported {
		var reads []codexSkillRead
		for _, name := range skillsInPaths(ev.Item.Command) {
			reads = append(reads, codexSkillRead{ItemID: ev.Item.ID, Skill: name, Command: ev.Item.Command, Status: "unverified", Reason: "unsupported literal read command", ReturnedBytes: len(ev.Item.Output)})
		}
		return reads
	}
	var combined strings.Builder
	allTrusted := true
	for _, path := range paths {
		body := inventory[path]
		if body == "" {
			allTrusted = false
		}
		combined.WriteString(body)
	}
	// Multi-file bytes можна розділити лише за довіреними operands у їхньому порядку.
	multiPrefix := allTrusted && strings.HasPrefix(combined.String(), ev.Item.Output)
	offset := 0
	var reads []codexSkillRead
	for _, path := range paths {
		names := skillsInPaths(path)
		if len(names) != 1 {
			continue
		}
		expected := inventory[path]
		read := codexSkillRead{ItemID: ev.Item.ID, Skill: names[0], Path: path, Command: ev.Item.Command, Status: "unverified", Reason: "path absent from staged inventory or expected file empty", ReturnedBytes: len(ev.Item.Output), ExpectedBytes: len(expected)}
		switch {
		case expected == "" || !allTrusted:
		case ev.Item.Output == "":
			read.Reason = "command returned no captured output"
		default:
			output := ev.Item.Output
			if len(paths) > 1 {
				if !multiPrefix {
					read.Reason = "multi-file output does not match staged operands in order"
					break
				}
				end := min(offset+len(expected), len(output))
				output = output[min(offset, len(output)):end]
			}
			if output == expected {
				read.Status = "full_returned_output"
				read.Reason = "captured output equals complete staged file"
				read.MatchedBytes = len(expected)
			} else if output != "" && (strings.HasPrefix(expected, output) || len(paths) == 1 && command == "sed" && strings.Contains(expected, output)) {
				read.Status = "partial_returned_output"
				read.Reason = "captured output matches a staged file fragment"
				read.MatchedBytes = len(output)
			} else if strings.HasPrefix(output, "---\nname: "+names[0]+"\n") {
				for i := 0; i < min(len(expected), len(output)) && expected[i] == output[i]; i++ {
					read.MatchedBytes++
				}
				if read.MatchedBytes > 0 && read.MatchedBytes < len(expected) {
					read.Status = "partial_returned_output"
					read.Reason = "captured output contains only a matching prefix"
				} else {
					read.MatchedBytes = 0
					read.Reason = "captured output differs from staged file"
				}
			} else {
				read.Reason = "captured output differs from staged file"
			}
		}
		// Coverage обчислюється незалежно; невдалий або непідтверджений lifecycle
		// не стає skill load навіть за повного captured body.
		switch {
		case ev.Type != "item.completed" || ev.Item.Status != "completed":
			read.Status = "unverified"
			read.Reason = "command not completed successfully"
			if ev.Item.Status == "failed" {
				read.Status = "failed"
			}
		case ev.Item.ID == "":
			read.Status = "unverified"
			read.Reason = "command item ID missing"
		case string(ev.Item.ExitCode) != "0" && string(ev.Item.ExitCode) != `"0"`:
			read.Status = "unverified"
			read.Reason = "exit code missing, invalid, or nonzero"
			var code int
			if json.Unmarshal(ev.Item.ExitCode, &code) == nil && code != 0 {
				read.Status = "failed"
			}
			var text string
			if json.Unmarshal(ev.Item.ExitCode, &text) == nil {
				if n, err := strconv.Atoi(text); err == nil && n != 0 {
					read.Status = "failed"
				}
			}
		}

		reads = append(reads, read)
		offset += len(expected)
	}
	return reads
}

// codexLiteralWords підтримує тільки literal words, quotes та backslash.
// Operators, substitutions і expansions відхиляються; command не виконується.
func codexLiteralWords(command string) ([]string, bool) {
	var words []string
	var word strings.Builder
	quote := byte(0)
	active := false
	for i := 0; i < len(command); i++ {
		ch := command[i]
		if quote == '\'' {
			if ch == quote {
				quote = 0
			} else {
				word.WriteByte(ch)
			}
			continue
		}
		if ch == '\\' {
			if i+1 == len(command) {
				return nil, false
			}
			i++
			next := command[i]
			if next == '\n' {
				return nil, false
			}
			if quote == '"' && !strings.ContainsRune("\"$\\`", rune(next)) {
				word.WriteByte('\\')
			}
			word.WriteByte(next)
			active = true
			continue
		}
		if ch == '$' || ch == '`' {
			return nil, false
		}
		if quote == '"' {
			if ch == quote {
				quote = 0
			} else {
				word.WriteByte(ch)
			}
			continue
		}
		if ch == '\'' || ch == '"' {
			quote = ch
			active = true
			continue
		}
		if ch == ' ' || ch == '\t' {
			if active {
				words = append(words, word.String())
				word.Reset()
				active = false
			}
			continue
		}
		if strings.ContainsRune(";|&<>(){}[]*?~\n\r", rune(ch)) {
			return nil, false
		}
		word.WriteByte(ch)
		active = true
	}
	if quote != 0 {
		return nil, false
	}
	if active {
		words = append(words, word.String())
	}
	return words, true
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
		// Read metadata не впливають на підрахунок lifecycle attempts.
		var ev struct {
			Type string `json:"type"`
			Item struct {
				Type string `json:"type"`
			} `json:"item"`
		}
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
