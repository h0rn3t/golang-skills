// Command evalrun runs the trigger and quality evals in evals.json against
// Claude Code in headless mode. Trigger evals check which go-* skills the model
// invokes for a prompt; quality evals run the prompt to completion and have a
// second model grade the answer against the eval's assertions.
//
// It is opt-in: it needs the claude CLI on PATH and either an authenticated
// session or ANTHROPIC_API_KEY. Skills are loaded from the repository with
// --plugin-dir, so the checked-out tree is what gets measured.
//
//	go run ./cmd/evalrun -set validation -kind all -j 2 -out results.json
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type triggerEval struct {
	Query         string   `json:"query"`
	ShouldTrigger []string `json:"should_trigger"`
	Set           string   `json:"set"`
}

type qualityEval struct {
	ID         int      `json:"id"`
	Prompt     string   `json:"prompt"`
	Files      []string `json:"files"`
	Assertions []string `json:"assertions"`
	Set        string   `json:"set"`
}

type evalsFile struct {
	TriggerEvals []triggerEval `json:"trigger_evals"`
	QualityEvals []qualityEval `json:"quality_evals"`
}

type triggerResult struct {
	Query string   `json:"query"`
	Want  []string `json:"want"`
	Got   []string `json:"got"`
	Pass  bool     `json:"pass"`
	Err   string   `json:"error,omitempty"`
}

type assertionResult struct {
	Assertion string `json:"assertion"`
	Pass      bool   `json:"pass"`
	Evidence  string `json:"evidence"`
}

type qualityResult struct {
	ID         int               `json:"id"`
	Pass       bool              `json:"pass"`
	Assertions []assertionResult `json:"assertions"`
	Output     string            `json:"output,omitempty"`
	Err        string            `json:"error,omitempty"`
}

type report struct {
	Set      string          `json:"set"`
	Trigger  []triggerResult `json:"trigger"`
	Quality  []qualityResult `json:"quality"`
	Summary  string          `json:"summary"`
	Finished time.Time       `json:"finished"`
}

type options struct {
	set, kind, model, judgeModel, out string
	limit, parallel                   int
	timeout                           time.Duration
	verbose                           bool
	restricted                        bool // claude supports --restricted: no built-in tools beyond --tools, no user/project settings or hooks
}

func main() {
	var o options
	flag.StringVar(&o.set, "set", "validation", "eval set: train, validation, or all")
	flag.StringVar(&o.kind, "kind", "all", "which evals: trigger, quality, or all")
	flag.IntVar(&o.limit, "limit", 0, "run at most N evals of each kind (0 = all)")
	flag.IntVar(&o.parallel, "j", 2, "evals to run concurrently")
	flag.StringVar(&o.model, "model", "", "model for the evaluated run (default: claude's default)")
	flag.StringVar(&o.judgeModel, "judge-model", "", "model for grading quality evals (default: claude's default)")
	flag.DurationVar(&o.timeout, "timeout", 10*time.Minute, "per-eval timeout")
	flag.StringVar(&o.out, "out", "", "write the JSON report to this file")
	flag.BoolVar(&o.verbose, "verbose", false, "print the model's full output for failed quality evals")
	flag.Parse()

	if err := run(o); err != nil {
		var ef exitError
		if errors.As(err, &ef) {
			fmt.Fprintln(os.Stderr, ef.msg)
			os.Exit(ef.code)
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
}

type exitError struct {
	code int
	msg  string
}

func (e exitError) Error() string { return e.msg }

func run(o options) error {
	if _, err := exec.LookPath("claude"); err != nil {
		return exitError{2, "claude CLI not found on PATH; install with: npm install -g @anthropic-ai/claude-code"}
	}
	if help, err := exec.Command("claude", "--help").CombinedOutput(); err == nil {
		o.restricted = strings.Contains(string(help), "--restricted")
	}
	root, err := repoRoot()
	if err != nil {
		return err
	}
	var evals evalsFile
	raw, err := os.ReadFile(filepath.Join(root, "evals", "evals.json"))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, &evals); err != nil {
		return fmt.Errorf("parse evals.json: %w", err)
	}

	rep := report{Set: o.set}
	if o.kind == "trigger" || o.kind == "all" {
		var selected []triggerEval
		for _, ev := range evals.TriggerEvals {
			if o.set == "all" || ev.Set == o.set {
				selected = append(selected, ev)
			}
		}
		selected = capN(selected, o.limit)
		rep.Trigger = make([]triggerResult, len(selected))
		forEach(o.parallel, len(selected), func(i int) {
			rep.Trigger[i] = runTrigger(o, root, selected[i])
			printTrigger(rep.Trigger[i])
		})
	}
	if o.kind == "quality" || o.kind == "all" {
		var selected []qualityEval
		for _, ev := range evals.QualityEvals {
			if o.set == "all" || ev.Set == o.set {
				selected = append(selected, ev)
			}
		}
		selected = capN(selected, o.limit)
		rep.Quality = make([]qualityResult, len(selected))
		forEach(o.parallel, len(selected), func(i int) {
			rep.Quality[i] = runQuality(o, root, selected[i])
			printQuality(rep.Quality[i], o.verbose)
		})
	}

	tp, tn := count(rep.Trigger, func(r triggerResult) bool { return r.Pass })
	qp, qn := count(rep.Quality, func(r qualityResult) bool { return r.Pass })
	rep.Summary = fmt.Sprintf("trigger %d/%d, quality %d/%d", tp, tn, qp, qn)
	rep.Finished = time.Now()
	fmt.Println()
	fmt.Println(rep.Summary)

	if o.out != "" {
		data, err := json.MarshalIndent(rep, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(o.out, data, 0o644); err != nil {
			return err
		}
	}
	if tp != tn || qp != qn {
		return exitError{1, "evals failed"}
	}
	return nil
}

func capN[T any](s []T, n int) []T {
	if n > 0 && len(s) > n {
		return s[:n]
	}
	return s
}

func count[T any](s []T, ok func(T) bool) (pass, total int) {
	for _, v := range s {
		if ok(v) {
			pass++
		}
	}
	return pass, len(s)
}

// forEach runs f(i) for i in [0, n) with at most parallel goroutines at once.
func forEach(parallel, n int, f func(int)) {
	sem := make(chan struct{}, max(parallel, 1))
	var wg sync.WaitGroup
	for i := range n {
		sem <- struct{}{}
		wg.Go(func() {
			defer func() { <-sem }()
			f(i)
		})
	}
	wg.Wait()
}

// workDir creates a scratch directory for one eval and copies the eval's
// fixture files into it at their repository-relative paths, so prompts that
// name evals/files/... resolve without exposing the rest of the repository.
func workDir(root string, files []string) (string, error) {
	dir, err := os.MkdirTemp("", "evalrun-")
	if err != nil {
		return "", err
	}
	for _, rel := range files {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			os.RemoveAll(dir)
			return "", fmt.Errorf("fixture %s: %w", rel, err)
		}
		dst := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			os.RemoveAll(dir)
			return "", err
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			os.RemoveAll(dir)
			return "", err
		}
	}
	return dir, nil
}

func repoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --show-toplevel: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// claude runs one headless invocation in dir. The repository is passed as a
// plugin so the skills under test load; dir is a scratch directory so the model
// cannot read the skill sources instead of invoking them.
func claude(o options, dir string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), o.timeout)
	defer cancel()
	if o.restricted {
		args = append(args, "--restricted")
	}
	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Dir = dir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if ctx.Err() != nil {
		return out, fmt.Errorf("timed out after %s", o.timeout)
	}
	if err != nil && len(out) == 0 {
		return out, fmt.Errorf("claude: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

// runTrigger asks the model with only the Skill tool available and records
// which go-* skills it invoked in its first turns.
func runTrigger(o options, root string, ev triggerEval) triggerResult {
	res := triggerResult{Query: ev.Query, Want: ev.ShouldTrigger}
	dir, err := workDir(root, nil)
	if err != nil {
		res.Err = err.Error()
		return res
	}
	defer os.RemoveAll(dir)
	args := []string{"-p", ev.Query, "--plugin-dir", root, "--output-format", "stream-json",
		"--verbose", "--max-turns", "3", "--tools", "Skill", "--allowed-tools", "Skill"}
	if o.model != "" {
		args = append(args, "--model", o.model)
	}
	out, err := claude(o, dir, args...)
	if err != nil {
		res.Err = err.Error()
	}
	got := map[string]bool{}
	for line := range strings.SplitSeq(string(out), "\n") {
		if !strings.Contains(line, `"Skill"`) {
			continue
		}
		var v any
		if json.Unmarshal([]byte(line), &v) != nil {
			continue
		}
		for _, name := range skillCalls(v) {
			got[name] = true
		}
	}
	res.Got = sortedKeys(got)
	res.Pass = res.Err == "" && triggered(ev.ShouldTrigger, got)
	return res
}

// skillCalls walks a decoded stream-json message and returns the go-* skill
// names passed to the Skill tool, wherever the tool_use block sits.
func skillCalls(v any) []string {
	var names []string
	switch t := v.(type) {
	case map[string]any:
		if t["type"] == "tool_use" && t["name"] == "Skill" {
			if in, ok := t["input"].(map[string]any); ok {
				if s, ok := in["skill"].(string); ok {
					if name := normalizeSkill(s); name != "" {
						names = append(names, name)
					}
				}
			}
		}
		for _, child := range t {
			names = append(names, skillCalls(child)...)
		}
	case []any:
		for _, child := range t {
			names = append(names, skillCalls(child)...)
		}
	}
	return names
}

// normalizeSkill strips a plugin prefix ("golang-skills:go-http") and a
// leading slash, and drops anything that is not a go-* skill.
func normalizeSkill(s string) string {
	s = strings.TrimPrefix(strings.TrimSpace(s), "/")
	if i := strings.LastIndex(s, ":"); i >= 0 {
		s = s[i+1:]
	}
	if !strings.HasPrefix(s, "go-") {
		return ""
	}
	return s
}

// triggered passes when every expected skill fired; for a negative control
// (no expected skills) it passes only when no go-* skill fired at all.
func triggered(want []string, got map[string]bool) bool {
	if len(want) == 0 {
		return len(got) == 0
	}
	for _, w := range want {
		if !got[w] {
			return false
		}
	}
	return true
}

// runQuality lets the model answer with read-only tools, then has a second,
// plugin-free model grade the answer against the assertions.
func runQuality(o options, root string, ev qualityEval) qualityResult {
	res := qualityResult{ID: ev.ID}
	dir, err := workDir(root, ev.Files)
	if err != nil {
		res.Err = err.Error()
		return res
	}
	defer os.RemoveAll(dir)
	const tools = "Skill,Read,Glob,Grep"
	args := []string{"-p", ev.Prompt, "--plugin-dir", root, "--output-format", "json",
		"--max-turns", "30", "--tools", tools, "--allowed-tools", tools}
	if o.model != "" {
		args = append(args, "--model", o.model)
	}
	out, err := claude(o, dir, args...)
	if err != nil {
		res.Err = err.Error()
		return res
	}
	answer, err := resultText(out)
	if err != nil {
		res.Err = err.Error()
		return res
	}
	res.Output = answer

	var checklist strings.Builder
	for i, a := range ev.Assertions {
		fmt.Fprintf(&checklist, "%d. %s\n", i+1, a)
	}
	judgePrompt := fmt.Sprintf(`You are grading a Go coding assistant's answer against a checklist.
Judge only what is present in the answer; do not assume unstated code.

ANSWER:
<<<
%s
>>>

CHECKLIST:
%s
Respond with JSON only, no prose, in this exact shape and order:
{"results":[{"assertion":"<checklist item>","pass":true,"evidence":"<short quote or reason>"}]}`, answer, checklist.String())

	jargs := []string{"-p", judgePrompt, "--output-format", "json", "--max-turns", "1", "--tools", ""}
	if o.judgeModel != "" {
		jargs = append(jargs, "--model", o.judgeModel)
	}
	jout, err := claude(o, dir, jargs...)
	if err != nil {
		res.Err = "judge: " + err.Error()
		return res
	}
	verdict, err := resultText(jout)
	if err != nil {
		res.Err = "judge: " + err.Error()
		return res
	}
	var graded struct {
		Results []assertionResult `json:"results"`
	}
	if err := json.Unmarshal([]byte(extractJSON(verdict)), &graded); err != nil {
		res.Err = fmt.Sprintf("judge returned non-JSON: %v", err)
		return res
	}
	res.Assertions = graded.Results
	res.Pass = len(graded.Results) == len(ev.Assertions)
	for _, a := range graded.Results {
		res.Pass = res.Pass && a.Pass
	}
	return res
}

// resultText returns the final assistant text from `--output-format json`,
// accepting both the single-object and the message-array shapes.
func resultText(out []byte) (string, error) {
	var obj struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(out, &obj); err == nil && obj.Result != "" {
		return obj.Result, nil
	}
	var arr []struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(out, &arr); err == nil {
		for i := len(arr) - 1; i >= 0; i-- {
			if arr[i].Result != "" {
				return arr[i].Result, nil
			}
		}
	}
	return "", fmt.Errorf("no result field in claude output: %.200s", out)
}

func extractJSON(s string) string {
	if i := strings.Index(s, "{"); i >= 0 {
		if j := strings.LastIndex(s, "}"); j > i {
			return s[i : j+1]
		}
	}
	return s
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func printTrigger(r triggerResult) {
	status := "PASS"
	if !r.Pass {
		status = "FAIL"
	}
	q := r.Query
	if len(q) > 70 {
		q = q[:67] + "..."
	}
	fmt.Printf("[%s] trigger  want=%v got=%v  %q", status, r.Want, r.Got, q)
	if r.Err != "" {
		fmt.Printf("  error: %s", r.Err)
	}
	fmt.Println()
}

func printQuality(r qualityResult, verbose bool) {
	status := "PASS"
	if !r.Pass {
		status = "FAIL"
	}
	fmt.Printf("[%s] quality #%d", status, r.ID)
	if r.Err != "" {
		fmt.Printf("  error: %s", r.Err)
	}
	fmt.Println()
	for _, a := range r.Assertions {
		mark := "ok  "
		if !a.Pass {
			mark = "MISS"
		}
		fmt.Printf("       %s %s — %s\n", mark, a.Assertion, a.Evidence)
	}
	if verbose && !r.Pass && r.Output != "" {
		fmt.Println("       --- output ---")
		fmt.Println(r.Output)
	}
}
