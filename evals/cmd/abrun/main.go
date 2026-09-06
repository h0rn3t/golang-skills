// Command abrun measures what a wording change to a skill does to the code the
// model actually writes. It runs the same refactoring prompt against the same
// fixtures with the current skills (the baseline arm), an optional complete
// plugin tree from -reference-root, and Markdown variants spliced into the
// current go-code-refactor/SKILL.md. It then reports mechanical deltas: lines,
// declared types, interfaces, pattern-flavored names, and whether the golden
// characterization test still passes.
//
// Nothing here grades prose; the model edits real files in a scratch directory
// and the files are what gets measured. A before/after claim requires the
// reference and baseline arms; the no-skill arm tests fixture sensitivity only.
//
// It is opt-in: it needs the claude CLI and the go toolchain on PATH, plus
// either an authenticated session or ANTHROPIC_API_KEY.
//
//	go run ./cmd/abrun -n 3 -j 4 -out ab.json
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
)

// defaultPrompt is shared by every arm; only the skill text differs between
// them, so any difference in the resulting code is attributable to the wording.
const defaultPrompt = "Refactor the Go package in ./%s so it reads better. " +
	"Keep observable behavior identical: the exported API, error texts, and rendered output must not change. " +
	"Apply the changes to the files."

// anchor is the SKILL.md line a variant block is spliced in front of.
const anchor = "## Workflow"

// patternNames are identifier fragments that mark a design-pattern scaffold
// rather than a domain concept. Idiomatic Go names (Handler, Server, Client)
// stay off the list.
var patternNames = []string{
	"Abstract", "Adapter", "Decorator", "Delegate", "Dispatcher", "Executor",
	"Factory", "Impl", "Manager", "Mediator", "Middleware", "Processor",
	"Provider", "Registry", "Strategy", "Visitor", "Wrapper",
}

// pluginSubdirs are the parts of the repository a plugin arm needs.
var pluginSubdirs = []string{".claude-plugin", "skills", "agents", "hooks"}

type options struct {
	tasks         string
	arms          string
	variants      string
	prompt        string
	model         string
	out           string
	referenceRoot string
	reps          int
	parallel      int
	seed          int64
	timeout       time.Duration
	verbose       bool
	keep          bool
}

func main() {
	var o options
	flag.StringVar(&o.tasks, "tasks", "", "comma-separated fixture names to run (default: every directory under evals/ab)")
	flag.StringVar(&o.arms, "arms", "", "comma-separated arms to run, e.g. \"no-skill,baseline\" (default: all)")
	flag.StringVar(&o.variants, "variants", "", "directory of variant Markdown blocks (default: evals/ab/variants)")
	flag.StringVar(&o.prompt, "prompt", defaultPrompt, "prompt template; %s is the fixture directory")
	flag.StringVar(&o.model, "model", "", "model for the evaluated run (default: claude's default)")
	flag.StringVar(&o.out, "out", "", "write the JSON report to this file")
	flag.StringVar(&o.referenceRoot, "reference-root", "", "alternate plugin root for a reference arm")
	flag.IntVar(&o.reps, "n", 2, "repetitions per fixture per arm")
	flag.IntVar(&o.parallel, "j", 2, "runs to execute concurrently")
	flag.Int64Var(&o.seed, "seed", 1, "deterministic job-order seed")
	flag.DurationVar(&o.timeout, "timeout", 10*time.Minute, "per-run timeout")
	flag.BoolVar(&o.verbose, "verbose", false, "print the model's final message for every run")
	flag.BoolVar(&o.keep, "keep", false, "keep the scratch tree of any run that failed to build or broke the golden test")
	flag.Parse()

	if err := run(o); err != nil {
		if ee, ok := errors.AsType[exitError](err); ok {
			fmt.Fprintln(os.Stderr, ee.msg)
			os.Exit(ee.code)
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

// arm is one variant of the skill tree: a name and the plugin directory the
// claude CLI loads for every run in that arm. The control arm loads no plugin
// at all and carries an empty dir.
type arm struct {
	Name   string `json:"name"`
	Text   string `json:"text,omitempty"`
	Source string `json:"source,omitempty"`
	Digest string `json:"sha256,omitempty"`
	dir    string
}

// controlArm is the run with no skills loaded. It answers the question a
// fixture has to pass before any wording comparison on it means anything: does
// the unaided model already do the right thing here? A fixture where the
// control ties with baseline has no trap in it, and measures nothing.
const controlArm = "no-skill"

const referenceArm = "reference"

// metrics are the structural counts taken from a fixture package. Test files
// are counted but excluded from the structure numbers, because adding a
// characterization test is a legitimate part of a refactor.
type metrics struct {
	Lines      int `json:"lines"`
	Files      int `json:"files"`
	TestFiles  int `json:"test_files"`
	Types      int `json:"types"`
	Interfaces int `json:"interfaces"`
	Funcs      int `json:"funcs"`
	Pattern    int `json:"pattern_names"`
}

func (m metrics) sub(o metrics) metrics {
	return metrics{
		Lines:      m.Lines - o.Lines,
		Files:      m.Files - o.Files,
		TestFiles:  m.TestFiles - o.TestFiles,
		Types:      m.Types - o.Types,
		Interfaces: m.Interfaces - o.Interfaces,
		Funcs:      m.Funcs - o.Funcs,
		Pattern:    m.Pattern - o.Pattern,
	}
}

// result is one (arm, fixture, repetition) run.
type result struct {
	Arm    string   `json:"arm"`
	Task   string   `json:"task"`
	Rep    int      `json:"rep"`
	Skills []string `json:"skills"`
	Before metrics  `json:"before"`
	After  metrics  `json:"after"`
	Delta  metrics  `json:"delta"`
	Build  bool     `json:"build"`
	Golden bool     `json:"golden"`
	// GoFail carries the go build or go test output when one of them failed,
	// so a behavior break is diagnosable from the report alone.
	GoFail  string `json:"go_failure,omitempty"`
	WorkDir string `json:"workdir,omitempty"`
	Output  string `json:"output,omitempty"`
	Err     string `json:"error,omitempty"`
}

type report struct {
	Prompt   string    `json:"prompt"`
	Model    string    `json:"model,omitempty"`
	Reps     int       `json:"reps"`
	Seed     int64     `json:"seed"`
	Arms     []arm     `json:"arms"`
	Results  []result  `json:"results"`
	Finished time.Time `json:"finished"`
}

type job struct {
	arm  arm
	task string
	rep  int
}

func run(o options) error {
	if err := validateOptions(o); err != nil {
		return err
	}
	root, err := repoRoot()
	if err != nil {
		return err
	}
	abDir := filepath.Join(root, "evals", "ab")
	if o.variants == "" {
		o.variants = filepath.Join(abDir, "variants")
	}
	tasks, err := findTasks(abDir, o.tasks)
	if err != nil {
		return err
	}
	if err := validateFixtures(abDir, tasks); err != nil {
		return err
	}
	if _, err := exec.LookPath("claude"); err != nil {
		return exitError{2, "claude CLI not found on PATH; install with: npm install -g @anthropic-ai/claude-code"}
	}
	if _, err := exec.LookPath("go"); err != nil {
		return exitError{2, "go toolchain not found on PATH"}
	}
	arms, cleanup, err := buildArms(root, o.referenceRoot, o.variants, o.arms)
	defer cleanup()
	if err != nil {
		return err
	}

	fmt.Printf("%d fixtures x %d arms x %d reps = %d runs\n\n", len(tasks), len(arms), o.reps, len(tasks)*len(arms)*o.reps)

	jobs := buildJobs(arms, tasks, o.reps, o.seed)

	rep := report{Prompt: o.prompt, Model: o.model, Reps: o.reps, Seed: o.seed, Arms: arms, Results: make([]result, len(jobs))}
	var mu sync.Mutex
	forEach(o.parallel, len(jobs), func(i int) {
		res := runOne(o, abDir, jobs[i].arm, jobs[i].task, jobs[i].rep)
		mu.Lock()
		defer mu.Unlock()
		rep.Results[i] = res
		printResult(res, o.verbose)
	})
	rep.Finished = time.Now()

	fmt.Println()
	printSummary(rep)

	if o.out != "" {
		data, err := json.MarshalIndent(rep, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(o.out, data, 0o644); err != nil {
			return fmt.Errorf("write report: %w", err)
		}
		fmt.Printf("\nreport written to %s\n", o.out)
	}
	return nil
}

func validateOptions(o options) error {
	if o.reps <= 0 {
		return exitError{2, "-n must be greater than zero"}
	}
	if o.parallel <= 0 {
		return exitError{2, "-j must be greater than zero"}
	}
	if o.timeout <= 0 {
		return exitError{2, "-timeout must be greater than zero"}
	}
	return nil
}

func buildJobs(arms []arm, tasks []string, reps int, seed int64) []job {
	jobs := make([]job, 0, len(arms)*len(tasks)*reps)
	for _, a := range arms {
		for _, task := range tasks {
			for rep := range reps {
				jobs = append(jobs, job{arm: a, task: task, rep: rep})
			}
		}
	}
	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(jobs), func(i, j int) {
		jobs[i], jobs[j] = jobs[j], jobs[i]
	})
	return jobs
}

func validateFixtures(abDir string, tasks []string) error {
	for _, task := range tasks {
		fixture := filepath.Join(abDir, task)
		if ok, err := containsFile(fixture, func(name string) bool {
			return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
		}); err != nil {
			return fmt.Errorf("validate fixture %s: %w", task, err)
		} else if !ok {
			return exitError{2, fmt.Sprintf("fixture %s contains no production Go files", task)}
		}

		golden := filepath.Join(abDir, "_golden", task)
		if ok, err := containsFile(golden, func(name string) bool {
			return strings.HasSuffix(name, "_test.go")
		}); err != nil {
			return exitError{2, fmt.Sprintf("fixture %s has no readable golden directory: %v", task, err)}
		} else if !ok {
			return exitError{2, fmt.Sprintf("fixture %s has no golden test", task)}
		}
	}
	return nil
}

func containsFile(dir string, match func(string) bool) (bool, error) {
	found := false
	err := filepath.WalkDir(dir, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && match(d.Name()) {
			found = true
		}
		return nil
	})
	return found, err
}

// findTasks returns the fixture package names under abDir, skipping the golden
// tests and the variant blocks.
func findTasks(abDir, only string) ([]string, error) {
	entries, err := os.ReadDir(abDir)
	if err != nil {
		return nil, fmt.Errorf("read fixtures: %w", err)
	}
	want := map[string]bool{}
	for name := range strings.SplitSeq(only, ",") {
		if name = strings.TrimSpace(name); name != "" {
			want[name] = true
		}
	}
	filtered := len(want) > 0
	var tasks []string
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() || strings.HasPrefix(name, "_") || strings.HasPrefix(name, ".") || name == "variants" {
			continue
		}
		if filtered {
			if !want[name] {
				continue
			}
			delete(want, name)
		}
		tasks = append(tasks, name)
	}
	if len(want) > 0 {
		return nil, exitError{2, "unknown -tasks value: " + strings.Join(sortedKeys(want), ",")}
	}
	if len(tasks) == 0 {
		return nil, exitError{2, "no fixtures selected under evals/ab"}
	}
	return tasks, nil
}

// buildArms materializes one plugin directory per arm: a copy of the skill tree
// with the variant block spliced into go-code-refactor/SKILL.md. The returned
// cleanup removes every copy and is safe to call even on error.
func buildArms(root, referenceRoot, variantsDir, only string) ([]arm, func(), error) {
	var dirs []string
	cleanup := func() {
		for _, d := range dirs {
			_ = os.RemoveAll(d) //nolint:errcheck // best-effort cleanup of temporary plugin copies
		}
	}
	arms := []arm{{Name: controlArm}}
	if referenceRoot != "" {
		if _, err := os.Stat(filepath.Join(referenceRoot, "skills", "go-code-refactor", "SKILL.md")); err != nil {
			return nil, cleanup, fmt.Errorf("reference root: %w", err)
		}
		arms = append(arms, arm{Name: referenceArm, Source: referenceRoot})
	}
	arms = append(arms, arm{Name: "baseline", Source: root})
	entries, err := os.ReadDir(variantsDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, cleanup, fmt.Errorf("read variants: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		text, err := os.ReadFile(filepath.Join(variantsDir, e.Name()))
		if err != nil {
			return nil, cleanup, fmt.Errorf("read variant %s: %w", e.Name(), err)
		}
		arms = append(arms, arm{Name: strings.TrimSuffix(e.Name(), ".md"), Text: strings.TrimSpace(string(text)), Source: root})
	}
	arms, err = selectArms(arms, only)
	if err != nil {
		return nil, cleanup, err
	}

	for i, a := range arms {
		if a.Name == controlArm {
			continue
		}
		dir, err := os.MkdirTemp("", "abrun-arm-")
		if err != nil {
			return nil, cleanup, err
		}
		dirs = append(dirs, dir)
		for _, sub := range pluginSubdirs {
			src := filepath.Join(a.Source, sub)
			if _, err := os.Stat(src); err != nil {
				continue
			}
			if err := os.CopyFS(filepath.Join(dir, sub), os.DirFS(src)); err != nil {
				return nil, cleanup, fmt.Errorf("copy %s: %w", sub, err)
			}
		}
		if a.Text != "" {
			if err := splice(filepath.Join(dir, "skills", "go-code-refactor", "SKILL.md"), a.Text); err != nil {
				return nil, cleanup, err
			}
		}
		digest, err := pluginDigest(dir)
		if err != nil {
			return nil, cleanup, fmt.Errorf("digest %s arm: %w", a.Name, err)
		}
		arms[i].Digest = digest
		arms[i].dir = dir
	}
	return arms, cleanup, nil
}

func pluginDigest(root string) (string, error) {
	h := sha256.New()
	for _, sub := range pluginSubdirs {
		base := filepath.Join(root, sub)
		if _, err := os.Stat(base); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", err
		}
		if err := filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			h.Write([]byte(filepath.ToSlash(rel)))
			h.Write([]byte{0})
			h.Write(data)
			h.Write([]byte{0})
			return nil
		}); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// selectArms keeps only the named arms, or all of them when only is empty.
func selectArms(arms []arm, only string) ([]arm, error) {
	want := map[string]bool{}
	for name := range strings.SplitSeq(only, ",") {
		if name = strings.TrimSpace(name); name != "" {
			want[name] = true
		}
	}
	if len(want) == 0 {
		return arms, nil
	}
	var kept []arm
	for _, a := range arms {
		if want[a.Name] {
			kept = append(kept, a)
			delete(want, a.Name)
		}
	}
	if len(want) > 0 {
		return nil, exitError{2, "unknown -arms value: " + strings.Join(sortedKeys(want), ",")}
	}
	return kept, nil
}

// splice inserts text in front of the anchor heading in a SKILL.md.
func splice(path, text string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read skill: %w", err)
	}
	content := string(data)
	i := strings.Index(content, "\n"+anchor+"\n")
	if i < 0 {
		return fmt.Errorf("anchor %q not found in %s", anchor, path)
	}
	patched := content[:i+1] + text + "\n\n" + content[i+1:]
	if err := os.WriteFile(path, []byte(patched), 0o644); err != nil {
		return fmt.Errorf("write patched skill: %w", err)
	}
	return nil
}

// runOne copies one fixture into a scratch module, lets the model refactor it,
// then measures the result and replays the golden characterization test.
func runOne(o options, abDir string, a arm, taskName string, rep int) (res result) {
	res = result{Arm: a.Name, Task: taskName, Rep: rep}
	work, err := os.MkdirTemp("", "abrun-work-")
	if err != nil {
		res.Err = err.Error()
		return res
	}
	// A run whose behavior moved is the one worth reading, so its scratch tree
	// survives when -keep is set; everything else is removed.
	defer func() {
		if o.keep && (res.GoFail != "" || res.Err != "") {
			res.WorkDir = work
			return
		}
		_ = os.RemoveAll(work) //nolint:errcheck // best-effort cleanup after the result is recorded
	}()

	pkgDir := filepath.Join(work, taskName)
	if err := os.CopyFS(pkgDir, os.DirFS(filepath.Join(abDir, taskName))); err != nil {
		res.Err = fmt.Sprintf("copy fixture: %v", err)
		return res
	}
	if err := os.WriteFile(filepath.Join(work, "go.mod"), []byte("module abeval\n\ngo 1.27\n"), 0o644); err != nil {
		res.Err = err.Error()
		return res
	}
	before, err := analyze(pkgDir)
	if err != nil {
		res.Err = fmt.Sprintf("analyze fixture: %v", err)
		return res
	}
	res.Before = before

	tools := "Skill,Read,Glob,Grep,Edit,Write"
	args := []string{
		"-p", fmt.Sprintf(o.prompt, taskName),
		"--output-format", "stream-json", "--verbose",
		"--max-turns", "40",
		"--permission-mode", "acceptEdits",
		// The arms differ only in skill text, so the run must not pick up the
		// operator's own settings or hooks on top of the plugin under test.
		"--restricted",
	}
	if a.dir == "" {
		tools = strings.TrimPrefix(tools, "Skill,")
	} else {
		args = append(args, "--plugin-dir", a.dir)
	}
	args = append(args, "--tools", tools, "--allowed-tools", tools)
	if o.model != "" {
		args = append(args, "--model", o.model)
	}
	out, err := claude(o.timeout, work, args...)
	if err != nil {
		res.Err = err.Error()
	}
	res.Skills, res.Output = parseStream(out)

	after, err := analyze(pkgDir)
	if err != nil {
		if res.Err == "" {
			res.Err = fmt.Sprintf("analyze result: %v", err)
		}
		return res
	}
	res.After = after
	res.Delta = after.sub(before)
	if err := goCmd(o.timeout, work, "build", "./..."); err != nil {
		res.GoFail = err.Error()
	} else {
		res.Build = true
	}

	// The model is expected to leave characterization tests behind, and its
	// helper types collide with the golden file's by name. Its tests have
	// already been counted; set them aside so the golden test compiles against
	// the refactored code alone.
	if err := hideTestFiles(pkgDir); err != nil {
		res.Err = fmt.Sprintf("hide model tests: %v", err)
		return res
	}

	golden := filepath.Join(abDir, "_golden", taskName)
	if err := os.CopyFS(pkgDir, os.DirFS(golden)); err != nil {
		res.Err = fmt.Sprintf("copy golden: %v", err)
		return res
	}
	if err := goCmd(o.timeout, work, "test", "./"+taskName+"/..."); err != nil {
		res.GoFail = err.Error()
	} else {
		res.Golden = true
	}
	return res
}

// hideTestFiles renames every _test.go in dir out of the build, so a test the
// model wrote cannot collide with the golden file or, worse, be the reason the
// golden run passes.
func hideTestFiles(dir string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		return os.Rename(path, path+".model")
	})
}

// analyze counts the structure of every Go file in dir.
func analyze(dir string) (metrics, error) {
	var m metrics
	fset := token.NewFileSet()
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() || !strings.HasSuffix(name, ".go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.HasSuffix(name, "_test.go") {
			m.TestFiles++
			return nil
		}
		m.Files++
		m.Lines += lineCount(data)
		file, err := parser.ParseFile(fset, path, data, parser.SkipObjectResolution)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		countDecls(file, &m)
		return nil
	})
	return m, err
}

func lineCount(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	lines := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		lines++
	}
	return lines
}

// countDecls adds one file's declarations to m.
func countDecls(file *ast.File, m *metrics) {
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			m.Funcs++
			m.Pattern += patternHits(d.Name.Name)
		case *ast.GenDecl:
			if d.Tok != token.TYPE {
				continue
			}
			for _, spec := range d.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				m.Types++
				m.Pattern += patternHits(ts.Name.Name)
				if _, ok := ts.Type.(*ast.InterfaceType); ok {
					m.Interfaces++
				}
			}
		}
	}
}

// patternHits reports how many pattern-scaffold fragments name contains.
func patternHits(name string) int {
	hits := 0
	for _, p := range patternNames {
		if strings.Contains(name, p) {
			hits++
		}
	}
	return hits
}

// parseStream pulls the go-* skills the model invoked and its final message out
// of a stream-json transcript.
func parseStream(out []byte) (skills []string, final string) {
	fired := map[string]bool{}
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
		if obj, ok := v.(map[string]any); ok {
			if text, ok := obj["result"].(string); ok && text != "" {
				final = text
			}
		}
	}
	return sortedKeys(fired), final
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

// normalizeSkill strips a plugin prefix ("golang-skills:go-http") and a leading
// slash, and drops anything that is not a go-* skill.
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

func claude(timeout time.Duration, dir string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Dir = dir
	out, err := commandOutput(cmd)
	if ctx.Err() != nil {
		return out, fmt.Errorf("timed out after %s", timeout)
	}
	if err != nil {
		return out, fmt.Errorf("claude: %w", err)
	}
	return out, nil
}

func commandOutput(cmd *exec.Cmd) ([]byte, error) {
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err == nil {
		return out, nil
	}
	if message := strings.TrimSpace(stderr.String()); message != "" {
		return out, fmt.Errorf("%w: %s", err, message)
	}
	return out, err
}

func goCmd(timeout time.Duration, dir string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

func repoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --show-toplevel: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
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

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// firstLine returns s up to its first newline, so a multi-line go test failure
// stays one line in the live log while the full text lives in the report.
func firstLine(s string) string {
	if before, _, ok := strings.Cut(s, "\n"); ok {
		return before
	}
	return s
}

func printResult(r result, verbose bool) {
	status := resultStatus(r)
	fmt.Printf("[%s] %-24s %-10s #%d  lines %+d  types %+d  iface %+d  funcs %+d  pattern %+d  build=%v golden=%v skills=%v",
		status, r.Arm, r.Task, r.Rep, r.Delta.Lines, r.Delta.Types, r.Delta.Interfaces, r.Delta.Funcs, r.Delta.Pattern, r.Build, r.Golden, r.Skills)
	if r.Err != "" {
		fmt.Printf("  error: %s", r.Err)
	}
	fmt.Println()
	if r.GoFail != "" {
		fmt.Printf("       %s\n", firstLine(r.GoFail))
	}
	if verbose && r.Output != "" {
		fmt.Println("       --- output ---")
		fmt.Println(r.Output)
	}
}

func resultStatus(r result) string {
	if r.Err != "" || !r.Build || !r.Golden {
		return "ERR"
	}
	return "ok "
}

type armSummary struct {
	Runs       int
	Errors     int
	Valid      int
	Build      int
	Golden     int
	Refactor   int
	Lines      int
	Types      int
	Interfaces int
	Funcs      int
	Pattern    int
}

func summarizeArm(rep report, name string) armSummary {
	var summary armSummary
	for _, r := range rep.Results {
		if r.Arm != name {
			continue
		}
		summary.Runs++
		if r.Err != "" {
			summary.Errors++
			continue
		}
		if r.Build {
			summary.Build++
		}
		if r.Golden {
			summary.Golden++
		}
		if slices.Contains(r.Skills, "go-code-refactor") {
			summary.Refactor++
		}
		if !r.Build || !r.Golden {
			continue
		}
		summary.Valid++
		summary.Lines += r.Delta.Lines
		summary.Types += r.Delta.Types
		summary.Interfaces += r.Delta.Interfaces
		summary.Funcs += r.Delta.Funcs
		summary.Pattern += r.Delta.Pattern
	}
	return summary
}

// printSummary averages structural deltas only over runs that both build and
// pass their hidden golden test. Failed sessions remain visible in the counts.
func printSummary(rep report) {
	fmt.Printf("%-24s %5s %6s %5s %8s %7s %7s %7s %9s %7s %7s %9s\n",
		"arm", "runs", "errors", "valid", "Δlines", "Δtypes", "Δiface", "Δfuncs", "Δpattern", "build", "golden", "refactor")
	for _, a := range rep.Arms {
		summary := summarizeArm(rep, a.Name)
		completed := summary.Runs - summary.Errors
		if summary.Valid == 0 {
			fmt.Printf("%-24s %5d %6d %5d %8s\n", a.Name, summary.Runs, summary.Errors, 0, "no data")
			continue
		}
		mean := func(sum int) float64 { return float64(sum) / float64(summary.Valid) }
		percent := func(count int) int {
			if completed == 0 {
				return 0
			}
			return 100 * count / completed
		}
		fmt.Printf("%-24s %5d %6d %5d %8.1f %7.2f %7.2f %7.2f %9.2f %6d%% %6d%% %8d%%\n",
			a.Name, summary.Runs, summary.Errors, summary.Valid,
			mean(summary.Lines), mean(summary.Types), mean(summary.Interfaces), mean(summary.Funcs), mean(summary.Pattern),
			percent(summary.Build), percent(summary.Golden), percent(summary.Refactor))
	}
}
