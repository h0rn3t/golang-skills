// The opencode runner exists because the plugin is published for every
// skills-aware agent, not only Claude Code, and a wording that helps one model
// is not yet evidence about the skill. Everything downstream of the session —
// the fixtures, the structural metrics, the hidden golden test — is shared with
// the claude runner, so the two differ only in which CLI writes the code.
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

// opencodeConfig is written into every arm home. It pins down what the arms are
// not allowed to differ in: no session sharing, no autoupdate mid-run, and the
// same step ceiling the claude runner passes as --max-turns.
const opencodeConfig = `{
  "$schema": "https://opencode.ai/config.json",
  "share": "disabled",
  "autoupdate": false,
  "agent": { "build": { "steps": %d } }
}
`

// opencodeHomes gives every arm its own HOME and records it on the arm.
//
// opencode discovers skills from HOME-derived directories (~/.config/opencode,
// ~/.opencode, ~/.claude, ~/.agents), so the control arm needs an isolated home
// just as much as the skilled arms do: run under the operator's own home it
// would load whatever go-* skills they have installed globally and stop being a
// control. The returned cleanup removes every home and is safe to call on error.
func opencodeHomes(arms []arm) (func(), error) {
	var homes []string
	cleanup := func() {
		for _, home := range homes {
			_ = os.RemoveAll(home) //nolint:errcheck // best-effort cleanup of temporary arm homes
		}
	}
	auth, err := opencodeAuth()
	if err != nil {
		return cleanup, err
	}
	for i, a := range arms {
		home, err := os.MkdirTemp("", "abrun-home-")
		if err != nil {
			return cleanup, err
		}
		homes = append(homes, home)
		if err := writeOpencodeHome(home, a.dir, auth); err != nil {
			return cleanup, fmt.Errorf("prepare %s arm home: %w", a.Name, err)
		}
		if err := checkOpencodeSkills(home, a.Name); err != nil {
			return cleanup, err
		}
		arms[i].home = home
	}
	return cleanup, nil
}

// opencodeAuth reads the credentials opencode wrote at login. An arm home is a
// fresh HOME, so without this copy every run would fail unauthenticated. A
// missing file is not an error: the provider key can come from the environment.
func opencodeAuth() ([]byte, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("locate opencode credentials: %w", err)
	}
	data, err := os.ReadFile(filepath.Join(home, ".local", "share", "opencode", "auth.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read opencode credentials: %w", err)
	}
	return data, nil
}

// writeOpencodeHome lays out one arm home: the shared config, the credentials,
// and the arm's skills. armDir is the materialized plugin tree, or empty for the
// control arm, which gets a home with no skills directory at all.
func writeOpencodeHome(home, armDir string, auth []byte) error {
	config := filepath.Join(home, ".config", "opencode")
	if err := os.MkdirAll(config, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(config, "opencode.json"), []byte(fmt.Sprintf(opencodeConfig, maxSteps)), 0o644); err != nil {
		return err
	}
	if auth != nil {
		data := filepath.Join(home, ".local", "share", "opencode")
		if err := os.MkdirAll(data, 0o700); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(data, "auth.json"), auth, 0o600); err != nil {
			return err
		}
	}
	if armDir == "" {
		return nil
	}
	return os.CopyFS(filepath.Join(config, "skills"), os.DirFS(filepath.Join(armDir, "skills")))
}

// checkOpencodeSkills asserts that an arm home loads the skills that arm is
// supposed to have and nothing else, which is the one precondition the whole
// comparison rests on. It doubles as a warm-up: opencode installs its plugin
// dependencies into a fresh home on first use, and doing that once here keeps
// concurrent runs from racing on the same install.
func checkOpencodeSkills(home, armName string) error {
	out, err := opencodeCmd(opencodeSetupTimeout, home, home, "debug", "skill")
	if err != nil {
		return fmt.Errorf("list skills for %s arm: %w", armName, err)
	}
	var skills []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(out, &skills); err != nil {
		return fmt.Errorf("decode skills for %s arm: %w", armName, err)
	}
	loaded := 0
	for _, s := range skills {
		if strings.HasPrefix(s.Name, "go-") {
			loaded++
		}
	}
	if armName == controlArm && loaded != 0 {
		return fmt.Errorf("%s arm home loads %d go-* skills; skill discovery is not isolated", controlArm, loaded)
	}
	if armName != controlArm && loaded == 0 {
		return fmt.Errorf("%s arm home loads no go-* skills", armName)
	}
	return nil
}

// opencodeSetupTimeout bounds the per-home warm-up, which downloads and installs
// opencode's plugin dependencies into a fresh HOME.
const opencodeSetupTimeout = 5 * time.Minute

// opencodeSession runs one headless refactoring session in work under the arm's
// home. --auto approves the edits the prompt asks for, the same role
// --permission-mode acceptEdits plays for the claude runner.
func opencodeSession(o options, home, work, prompt string) ([]byte, error) {
	args := []string{"run", "--format", "json", "--auto"}
	if o.model != "" {
		args = append(args, "--model", o.model)
	}
	args = append(args, prompt)
	return opencodeCmd(o.timeout, work, home, args...)
}

// opencodeCmd runs one opencode invocation and returns everything it wrote to
// stdout.
//
// The transcript is captured through a temporary file rather than a pipe on
// purpose: opencode exits without draining stdout, so a pipe truncates it at the
// first full buffer — 64 KiB, well under one session — and the tail that goes
// missing is exactly the part this harness reads, the final message and the last
// cost lines. The truncation is silent and still parses as NDJSON, so nothing
// downstream would notice.
func opencodeCmd(timeout time.Duration, dir, home string, args ...string) ([]byte, error) {
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
	cmd := exec.CommandContext(ctx, "opencode", args...)
	cmd.Dir = dir
	cmd.Env = opencodeEnv(home, dir)
	cmd.Stdout = stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	out, err := os.ReadFile(stdout.Name())
	if err != nil {
		return nil, fmt.Errorf("read opencode transcript: %w", err)
	}
	if ctx.Err() != nil {
		return out, fmt.Errorf("timed out after %s", timeout)
	}
	if runErr != nil {
		if message := strings.TrimSpace(stderr.String()); message != "" {
			return out, fmt.Errorf("opencode: %w: %s", runErr, message)
		}
		return out, fmt.Errorf("opencode: %w", runErr)
	}
	return out, nil
}

// opencodeEnv points every path opencode derives from the environment at the
// arm's own home and the run's own scratch tree.
//
// The XDG variables are overridden alongside HOME so that an operator who sets
// them cannot leak their own config or skills into a run. PWD has to be set too,
// and it is not a formality: opencode takes its working directory from PWD and
// ignores the working directory of its own process, so with the inherited value
// left in place every session edits the repository the harness was launched from
// instead of the scratch copy — including, if it goes looking, the golden test
// it is being measured against.
func opencodeEnv(home, dir string) []string {
	return append(os.Environ(),
		"HOME="+home,
		"PWD="+dir,
		// OLDPWD would otherwise hand the session a path back to the repository.
		"OLDPWD="+dir,
		"XDG_CONFIG_HOME="+filepath.Join(home, ".config"),
		"XDG_DATA_HOME="+filepath.Join(home, ".local", "share"),
		"XDG_STATE_HOME="+filepath.Join(home, ".local", "state"),
		"XDG_CACHE_HOME="+filepath.Join(home, ".cache"),
	)
}

// opencodeEvent is the part of one line of `opencode run --format json` the
// benchmark reads. Input stays raw because it is a different shape per tool and
// a decode failure on one tool must not drop the rest of the transcript.
type opencodeEvent struct {
	Type string `json:"type"`
	Part struct {
		Tool  string  `json:"tool"`
		Text  string  `json:"text"`
		Cost  float64 `json:"cost"`
		State struct {
			Input json.RawMessage `json:"input"`
		} `json:"state"`
	} `json:"part"`
}

// parseOpencodeStream pulls the go-* skills the model loaded, its last message,
// and the session cost out of an NDJSON transcript. opencode reports a skill as
// a call to one `skill` tool carrying the name, where claude emits a Skill
// tool_use block per call.
func parseOpencodeStream(out []byte) (skills []string, final string, cost float64) {
	fired := map[string]bool{}
	for line := range strings.SplitSeq(string(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var ev opencodeEvent
		if json.Unmarshal([]byte(line), &ev) != nil {
			continue
		}
		switch ev.Type {
		case "tool_use":
			if ev.Part.Tool != "skill" {
				continue
			}
			var in struct {
				Name string `json:"name"`
			}
			if json.Unmarshal(ev.Part.State.Input, &in) != nil {
				continue
			}
			if name := normalizeSkill(in.Name); name != "" {
				fired[name] = true
			}
		case "text":
			if ev.Part.Text != "" {
				final = ev.Part.Text
			}
		case "step_finish":
			cost += ev.Part.Cost
		}
	}
	return sortedKeys(fired), final, cost
}
