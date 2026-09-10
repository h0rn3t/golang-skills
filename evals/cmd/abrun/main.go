// Command abrun measures what a skill does to the code the model actually
// writes. It runs one prompt against a corpus of fixtures with the current
// skills (the baseline arm), an optional complete plugin tree from
// -reference-root, and Markdown variants spliced into the current
// go-code-refactor/SKILL.md. It then reports mechanical deltas: lines, declared
// types, interfaces, pattern-flavored names, and whether the hidden golden test
// passes.
//
// There are two corpora. -corpus refactor hands the model a working package and
// asks it to improve the reading; the golden test characterizes the behavior
// that already existed, so it can only be broken, and the score is how much
// structure the refactor removed. -corpus implement hands the model documented
// but unimplemented declarations; the golden test is the specification, so it
// can be failed outright, and the score is whether the package works at all.
//
// Nothing here grades prose; the model edits real files in a scratch directory
// and the files are what gets measured. A before/after claim requires the
// reference and baseline arms; the no-skill arm tests fixture sensitivity only.
//
// It is opt-in: it needs the go toolchain plus the agent CLI named by -runner
// on PATH — claude with either an authenticated session or ANTHROPIC_API_KEY,
// opencode with a logged-in provider, copilot signed in to a GitHub account
// whose plan serves -model, or codex signed in to an account that serves it.
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
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang-skills/evals/internal/evalplugin"
)

// The corpus prompts are shared by every arm; only the skill text differs
// between them, so any difference in the resulting code is attributable to the
// wording.
const refactorPrompt = "Refactor the Go package in ./%s so it reads better. " +
	"Keep observable behavior identical: the exported API, error texts, and rendered output must not change. " +
	"Apply the changes to the files."

const implementPrompt = "Implement the Go package in ./%s. " +
	"Every exported declaration is already there with its documentation; write the bodies so the package does what the documentation says. " +
	"Do not change the exported signatures. Apply the changes to the files."

// corpora are the two experiments the fixtures support, and they are kept apart
// on purpose. A refactor is scored on how much structure it removes while
// behavior holds, an implementation on whether the hidden test passes at all
// and how much code it took; a mean over both is a number about neither.
const (
	corpusRefactor  = "refactor"
	corpusImplement = "implement"
)

// implementDir holds the implementation corpus. The leading underscore keeps
// findTasks from offering it as a refactor fixture.
const implementDir = "_implement"

// corpusPrompt returns the prompt a corpus is run with unless -prompt overrides it.
func corpusPrompt(corpus string) string {
	if corpus == corpusImplement {
		return implementPrompt
	}
	return refactorPrompt
}

// anchor is the SKILL.md line a variant block is spliced in front of.
const anchor = "## Workflow"

// runners are the agent CLIs the harness can drive. The skills ship for every
// skills-aware agent, so a claim about a wording holds better when the same
// fixtures and the same hidden golden test can be replayed under more than one.
const (
	runnerClaude   = "claude"
	runnerOpencode = "opencode"
	runnerCopilot  = "copilot"
	runnerCodex    = "codex"
)

// effortRunners are the runners whose CLI can set a reasoning effort level. The
// flag is rejected elsewhere rather than ignored, because a run recorded as
// xhigh that was served at the model's default is a report that lies.
var effortRunners = []string{runnerClaude, runnerCodex, runnerCopilot}

// traceFile is the raw session transcript, written at the root of the scratch
// tree so it sits outside the fixture package that gets measured and digested.
const traceFile = "trace.jsonl"

// maxSteps bounds one session where the runner can express a ceiling:
// --max-turns for claude, the build agent's step ceiling for opencode. The
// copilot CLI has no equivalent, so a copilot session is bounded by -timeout
// alone.
const maxSteps = 40

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
	effort        string
	runner        string
	corpus        string
	out           string
	referenceRoot string
	reps          int
	parallel      int
	seed          int64
	timeout       time.Duration
	verbose       bool
	keep          bool
	repair        bool
}

func main() {
	var o options
	flag.StringVar(&o.tasks, "tasks", "", "comma-separated fixture names to run (default: every directory under evals/ab)")
	flag.StringVar(&o.arms, "arms", "", "comma-separated arms to run, e.g. \"no-skill,baseline\" (default: all)")
	flag.StringVar(&o.variants, "variants", "", "directory of variant Markdown blocks (default: evals/ab/variants)")
	flag.StringVar(&o.prompt, "prompt", "", "prompt template; %s is the fixture directory (default: the corpus prompt)")
	flag.StringVar(&o.corpus, "corpus", corpusRefactor, "fixture corpus to run: refactor or implement")
	flag.StringVar(&o.model, "model", "", "model for the evaluated run (default: the runner's own default; required for opencode)")
	flag.StringVar(&o.effort, "effort", "", "reasoning effort for the evaluated run, e.g. medium (claude, codex and copilot only)")
	flag.StringVar(&o.runner, "runner", runnerClaude, "agent CLI to drive: claude, opencode, copilot or codex")
	flag.StringVar(&o.out, "out", "", "write the JSON report to this file")
	flag.StringVar(&o.referenceRoot, "reference-root", "", "alternate plugin root for a reference arm")
	flag.IntVar(&o.reps, "n", 2, "repetitions per fixture per arm")
	flag.IntVar(&o.parallel, "j", 2, "runs to execute concurrently")
	flag.Int64Var(&o.seed, "seed", 1, "deterministic job-order seed")
	flag.DurationVar(&o.timeout, "timeout", 10*time.Minute, "per-run timeout")
	flag.BoolVar(&o.verbose, "verbose", false, "print the model's final message for every run")
	flag.BoolVar(&o.keep, "keep", false, "keep every run's scratch tree, including successful source and model tests (.model)")
	flag.BoolVar(&o.repair, "repair", false, "after the session, return the measured line count and any independent failure and grant one repair turn")
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
// agent CLI loads for every run in that arm. The control arm loads no plugin at
// all and carries an empty dir. home is set only by the opencode runner, which
// loads skills from HOME rather than from a plugin flag.
type arm struct {
	Name   string `json:"name"`
	Text   string `json:"text,omitempty"`
	Source string `json:"source,omitempty"`
	Digest string `json:"sha256,omitempty"`
	dir    string
	home   string
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
	// Lines is the production line count, and it is the number every line
	// claim in an evidence report has to mean: physical lines in the package's
	// non-test *.go files, blank lines and comments included, a trailing line
	// without a newline counted once. No file under a _test.go name
	// contributes, so a session cannot shrink this number by moving code into
	// a test, and cannot grow it by writing one.
	Lines      int `json:"lines"`
	Files      int `json:"files"`
	TestFiles  int `json:"test_files"`
	Types      int `json:"types"`
	Interfaces int `json:"interfaces"`
	Funcs      int `json:"funcs"`
	Pattern    int `json:"pattern_names"`
	// Exported counts the package's public surface. On an implementation task
	// every exported declaration the specification needs is already there, so
	// growth here is scope the task never asked for.
	Exported int `json:"exported"`
	// Branches counts decision points: if, for, range, and each non-default
	// case. It stands in for hand-rolled control flow — the difference between
	// leaning on what the standard library already decides and re-deciding it
	// by hand — which the declaration counts cannot see when the API is fixed.
	Branches int `json:"branches"`
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
		Exported:   m.Exported - o.Exported,
		Branches:   m.Branches - o.Branches,
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
	// ModelTests is pass, fail, or skipped (no test files). Older reports omit it.
	ModelTests    string `json:"model_tests,omitempty"`
	ModelTestFail string `json:"model_test_failure,omitempty"`
	// Edited reports whether the fixture files actually changed. A session that
	// touched nothing and reported success is not a behavior-preserving refactor
	// with a zero delta; it is a run that never happened where it was measured,
	// and averaging it in would hide that as a tie.
	Edited bool `json:"edited"`
	// Leaked reports that the transcript mentions the fixture corpus in the
	// repository rather than the scratch copy. The hidden golden test sits there
	// next to the fixtures, so a session that found its way back to the checkout
	// is not evidence about anything.
	Leaked bool `json:"leaked,omitempty"`
	// BehaviorFailure is a golden overlay that compiled and then failed an
	// assertion: the observable contract moved. HarnessFailure is a golden run
	// that never reached an assertion, because a name the overlay declares is
	// also declared by the model's production code. The second one is the
	// harness colliding with the model, not evidence about behavior, and
	// counting it as a behavior break is what made one 5×5 arm look worse than
	// it was. Neither is set when the package failed to build on its own: that
	// failure is already Build, and the golden result carries no information.
	BehaviorFailure bool `json:"behavior_failure,omitempty"`
	HarnessFailure  bool `json:"harness_failure,omitempty"`
	// LineGatePass reports whether production lines failed to grow, which is the
	// concision gate's own criterion measured the way analyze measures it. It is
	// not a validity verdict — read it next to Build, Golden and Edited.
	LineGatePass bool `json:"line_gate_pass"`
	// FixHunksBefore and FixHunks are the modernizations `go fix -diff` still
	// proposes for the package's production files, before the session and
	// after the last turn. Zero after means the toolchain has no modernizer
	// left for the code in the tree; it says nothing about what no analyzer
	// covers (cmp.Or, errors.Join, iter.Seq), so read it as a floor and not as
	// a modernity score. Both are absent when the reading could not be taken
	// and FixUnmeasured says which one failed and why: a package that does not
	// type-check produces an empty diff for the wrong reason, and recording
	// that as zero would read as fully modern.
	FixHunksBefore *int   `json:"fix_hunks_before,omitempty"`
	FixHunks       *int   `json:"fix_hunks,omitempty"`
	FixUnmeasured  string `json:"fix_hunks_unmeasured,omitempty"`
	// EmptyDiff is a session that ran to completion and deliberately left every
	// file as it found it. Edited==false with an error is a different thing: a
	// session that never got as far as writing, which is why this is its own
	// field rather than the negation of Edited.
	EmptyDiff bool `json:"empty_diff"`
	// ReportedCounts reports whether the final message stated a line count or
	// delta. It reads the model's prose, so it measures what the session
	// claimed, never what it did — Commands and the retained trace are what
	// show whether it measured anything.
	ReportedCounts bool `json:"reported_counts"`
	// RepairFired reports that -repair granted a second turn, which happens only
	// when the first turn missed the line gate or failed the independent check.
	// PreRepair and PreRepairGolden are the readings that triggered it, so the
	// report shows what the loop was handed and what it did with it.
	RepairFired     bool     `json:"repair_fired,omitempty"`
	PreRepair       *metrics `json:"pre_repair,omitempty"`
	PreRepairGolden bool     `json:"pre_repair_golden,omitempty"`
	// RepairFeedback is the exact text returned to the model. It is recorded
	// because the feedback is generated per run: without it the report cannot
	// say what the session was actually told.
	RepairFeedback string `json:"repair_feedback,omitempty"`
	// Commands counts the shell calls across every turn. Only the codex runner
	// reports it; the claude arms are granted no shell at all.
	Commands int `json:"commands,omitempty"`
	// Trace is the retained JSONL transcript, written for every runner and kept
	// with -keep. A final message claiming a measurement is checkable against
	// the commands the session actually ran.
	Trace string `json:"trace_path,omitempty"`
	// GoFail carries the go build or go test output when one of them failed,
	// so a behavior break is diagnosable from the report alone.
	GoFail  string  `json:"go_failure,omitempty"`
	Cost    float64 `json:"cost_usd,omitempty"`
	WorkDir string  `json:"workdir,omitempty"`
	Output  string  `json:"output,omitempty"`
	Err     string  `json:"error,omitempty"`
}

type report struct {
	Prompt   string    `json:"prompt"`
	Corpus   string    `json:"corpus"`
	Runner   string    `json:"runner"`
	Model    string    `json:"model,omitempty"`
	Effort   string    `json:"effort,omitempty"`
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
	if o.corpus == corpusImplement {
		abDir = filepath.Join(abDir, implementDir)
	}
	if o.prompt == "" {
		o.prompt = corpusPrompt(o.corpus)
	}
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
	if _, err := exec.LookPath(o.runner); err != nil {
		return exitError{2, missingRunner(o.runner)}
	}
	if _, err := exec.LookPath("go"); err != nil {
		return exitError{2, "go toolchain not found on PATH"}
	}
	arms, cleanup, err := buildArms(root, o.referenceRoot, o.variants, o.arms)
	defer cleanup()
	if err != nil {
		return err
	}
	switch o.runner {
	case runnerOpencode:
		homes, err := opencodeHomes(arms)
		defer homes()
		if err != nil {
			return err
		}
	case runnerCopilot:
		homes, err := copilotHomes(arms)
		defer homes()
		if err != nil {
			return err
		}
	case runnerCodex:
		homes, err := codexHomes(arms)
		defer homes()
		if err != nil {
			return err
		}
	}

	fmt.Printf("%d fixtures x %d arms x %d reps = %d runs\n\n", len(tasks), len(arms), o.reps, len(tasks)*len(arms)*o.reps)

	jobs := buildJobs(arms, tasks, o.reps, o.seed)

	rep := report{Prompt: o.prompt, Corpus: o.corpus, Runner: o.runner, Model: o.model, Effort: o.effort, Reps: o.reps, Seed: o.seed, Arms: arms, Results: make([]result, len(jobs))}
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
	if o.corpus != corpusRefactor && o.corpus != corpusImplement {
		return exitError{2, fmt.Sprintf("-corpus must be %s or %s", corpusRefactor, corpusImplement)}
	}
	switch o.runner {
	case runnerClaude, runnerCopilot, runnerCodex:
	case runnerOpencode:
		// An arm home carries no model preference of its own, so opencode has
		// nothing to fall back on and the model has to be named explicitly.
		if o.model == "" {
			return exitError{2, "-model is required for the opencode runner, e.g. -model opencode-go/minimax-m3"}
		}
	default:
		return exitError{2, fmt.Sprintf("-runner must be one of %s, %s, %s, %s", runnerClaude, runnerOpencode, runnerCopilot, runnerCodex)}
	}
	if o.effort != "" && !slices.Contains(effortRunners, o.runner) {
		return exitError{2, fmt.Sprintf("-effort is only supported by the %s runners", strings.Join(effortRunners, " and "))}
	}
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

func missingRunner(runner string) string {
	hint := "install with: npm install -g @anthropic-ai/claude-code"
	switch runner {
	case runnerOpencode:
		hint = "install with: npm install -g opencode-ai"
	case runnerCopilot:
		hint = "install with: npm install -g @github/copilot"
	case runnerCodex:
		hint = "install with: npm install -g @openai/codex"
	}
	return runner + " CLI not found on PATH; " + hint
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
		if err := checkArmFrontmatter(dir); err != nil {
			return nil, cleanup, fmt.Errorf("%s arm: %w", a.Name, err)
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
	// Successful implementations also need source evidence for structural review.
	defer func() {
		if o.keep {
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
	fixBefore, fixBeforeErr := fixHunks(o.timeout, work, taskName)
	if fixBeforeErr == nil {
		res.FixHunksBefore = &fixBefore
	}
	beforeDigest, err := fixtureDigest(pkgDir)
	if err != nil {
		res.Err = fmt.Sprintf("digest fixture: %v", err)
		return res
	}

	// measure re-reads everything a turn can change. It runs after the first
	// session and again after a repair turn, so the recorded numbers always
	// describe the tree as the run left it. It reports false when the run
	// cannot continue.
	measure := func() bool {
		if afterDigest, err := fixtureDigest(pkgDir); err == nil {
			res.Edited = afterDigest != beforeDigest
		} else if res.Err == "" {
			res.Err = fmt.Sprintf("digest result: %v", err)
			return false
		}
		after, err := analyze(pkgDir)
		if err != nil {
			if res.Err == "" {
				res.Err = fmt.Sprintf("analyze result: %v", err)
			}
			return false
		}
		res.After = after
		res.Delta = after.sub(before)
		res.LineGatePass = res.Delta.Lines <= 0
		res.EmptyDiff = !res.Edited && res.Err == ""
		res.GoFail, res.Build = "", false
		if err := goCmd(o.timeout, work, "build", "./..."); err != nil {
			res.GoFail = err.Error()
		} else {
			res.Build = true
		}
		res.FixHunks = nil
		hunks, fixErr := fixHunks(o.timeout, work, taskName)
		if fixErr == nil {
			res.FixHunks = &hunks
		}
		res.FixUnmeasured = fixNote(fixBeforeErr, fixErr)
		res.ModelTests, res.ModelTestFail = "skipped", ""
		if after.TestFiles > 0 {
			res.ModelTests = "pass"
			if err := goCmd(o.timeout, work, "test", "-count=1", "./..."); err != nil {
				res.ModelTests = "fail"
				res.ModelTestFail = err.Error()
			}
		}
		return true
	}

	turn := runSession(o, a, work, fmt.Sprintf(o.prompt, taskName))
	res.merge(turn)
	// The transcript is written before anything is measured, so it survives even
	// a run that fails from here on. A final message that claims a measurement
	// is only checkable against the commands the session actually ran.
	if err := appendTrace(work, turn.out); err == nil && o.keep {
		res.Trace = filepath.Join(work, traceFile)
	}
	res.Leaked = bytes.Contains(turn.out, []byte(abDir))
	if !measure() {
		return res
	}

	golden := filepath.Join(abDir, "_golden", taskName)
	// The repair loop closes the gap stage 2 measured: a session that is handed
	// its own numbers back, in the harness's own counting convention, together
	// with the independent failure, repairs what a generic re-review does not.
	// The probe runs on a throwaway copy of the whole module, so the golden
	// never enters the tree the model can see.
	if o.repair && res.Err == "" {
		probePass, probeFail, probeErr := probeGolden(o.timeout, work, taskName, golden)
		switch {
		case probeErr != nil:
			res.Err = fmt.Sprintf("probe golden: %v", probeErr)
			return res
		case res.LineGatePass && probePass:
			// Nothing to repair; the loop costs no extra turn.
		default:
			pre := res.After
			res.RepairFired, res.PreRepair, res.PreRepairGolden = true, &pre, probePass
			res.RepairFeedback = repairFeedback(taskName, before.Lines, pre.Lines, probeFail)
			repair := runSession(o, a, work, res.RepairFeedback)
			res.merge(repair)
			if err := appendTrace(work, repair.out); err == nil && o.keep {
				res.Trace = filepath.Join(work, traceFile)
			}
			if bytes.Contains(repair.out, []byte(abDir)) {
				res.Leaked = true
			}
			if !measure() {
				return res
			}
		}
	}

	// Run the model's tests first, then isolate them to avoid name collisions
	// or a model test being the reason the independent golden check passes.
	if err := hideTestFiles(pkgDir); err != nil {
		res.Err = fmt.Sprintf("hide model tests: %v", err)
		return res
	}

	if err := os.CopyFS(pkgDir, os.DirFS(golden)); err != nil {
		res.Err = fmt.Sprintf("copy golden: %v", err)
		return res
	}
	if err := goCmd(o.timeout, work, "test", "./"+taskName+"/..."); err != nil {
		res.GoFail = err.Error()
		res.BehaviorFailure, res.HarnessFailure = classifyGolden(res.Build, err.Error())
	} else {
		res.Golden = true
	}
	return res
}

// sessionTurn is what one model invocation leaves behind. It exists because a
// repair run has two turns and both have to be folded into one result.
type sessionTurn struct {
	out      []byte
	skills   []string
	final    string
	cost     float64
	commands int
	err      error
}

// runSession dispatches one turn to the configured runner.
func runSession(o options, a arm, work, prompt string) sessionTurn {
	var t sessionTurn
	switch o.runner {
	case runnerOpencode:
		t.out, t.err = opencodeSession(o, a.home, work, prompt)
		t.skills, t.final, t.cost = parseOpencodeStream(t.out)
	case runnerCopilot:
		t.out, t.err = copilotSession(o, a, work, prompt)
		t.skills, t.final, t.cost = parseCopilotStream(t.out)
	case runnerCodex:
		t.out, t.err = codexSession(o, a.home, work, prompt)
		t.skills, t.final, t.cost = parseCodexStream(t.out)
		t.commands = codexCommands(t.out)
	default:
		t.out, t.err = claudeSession(o, a.dir, work, prompt)
		t.skills, t.final, t.cost = parseClaudeStream(t.out)
	}
	return t
}

// merge folds one turn into the result. Skills and cost accumulate across
// turns, because a repair turn can reach a skill the first one did not and
// spends money of its own; the final message is the latest turn's, since that
// is the session's own last word on what it did.
func (r *result) merge(t sessionTurn) {
	if t.err != nil && r.Err == "" {
		r.Err = t.err.Error()
	}
	for _, s := range t.skills {
		if !slices.Contains(r.Skills, s) {
			r.Skills = append(r.Skills, s)
		}
	}
	sort.Strings(r.Skills)
	r.Cost += t.cost
	r.Commands += t.commands
	if t.final != "" {
		r.Output = t.final
	}
	r.ReportedCounts = reportedCounts(r.Output)
}

// appendTrace adds one turn's transcript to the run's trace file, so a repair
// run keeps both turns in the order they happened.
func appendTrace(work string, out []byte) error {
	f, err := os.OpenFile(filepath.Join(work, traceFile), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(out); err != nil {
		_ = f.Close() //nolint:errcheck // the write error is the one worth reporting
		return err
	}
	return f.Close()
}

// probeGolden replays the golden test against a throwaway copy of the module
// and returns whether it passed plus the failure text.
//
// The copy is the whole point. The golden must never appear in the tree the
// model can read, or the next turn would be repairing against a test it can
// see, which is a different experiment. The probe leaves the model's tree
// untouched, including the tests the model wrote.
func probeGolden(timeout time.Duration, work, taskName, goldenDir string) (bool, string, error) {
	probe, err := os.MkdirTemp("", "abrun-probe-")
	if err != nil {
		return false, "", err
	}
	defer func() {
		_ = os.RemoveAll(probe) //nolint:errcheck // best-effort cleanup of the probe copy
	}()
	// MkdirTemp already created the directory, and CopyFS refuses to overwrite,
	// so the copy goes into a fresh child.
	root := filepath.Join(probe, "module")
	if err := os.CopyFS(root, os.DirFS(work)); err != nil {
		return false, "", err
	}
	pkgDir := filepath.Join(root, taskName)
	if err := hideTestFiles(pkgDir); err != nil {
		return false, "", err
	}
	if err := os.CopyFS(pkgDir, os.DirFS(goldenDir)); err != nil {
		return false, "", err
	}
	if err := goCmd(timeout, root, "test", "./"+taskName+"/..."); err != nil {
		return false, err.Error(), nil
	}
	return true, "", nil
}

// repairFeedback is the text the loop returns to the model.
//
// It names the counting convention on purpose. Stage 2 measured a session that
// read the same file as 101 lines where the harness read 138 — non-blank and
// non-comment against physical — declared the gate met and changed nothing.
// The disagreement was about the definition, not the code, so the definition
// travels with the number.
//
// The independent failure arrives as assertion text with file positions
// stripped: the model is told what broke, not which file holds the test.
func repairFeedback(taskName string, start, current int, failure string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Package under review: ./%s\n\n", taskName)
	fmt.Fprintf(&b, "Starting production LOC: %d\n", start)
	fmt.Fprintf(&b, "Current production LOC: %d\n", current)
	if current > start {
		fmt.Fprintf(&b, "Gate: FAIL, growth +%d\n", current-start)
	} else {
		b.WriteString("Gate: PASS on line count.\n")
	}
	b.WriteString("LOC counts physical lines in the package's non-test *.go files, blank lines and comments included.\n")
	if failure != "" {
		fmt.Fprintf(&b, "\nIndependent contract failure:\n%s\n", assertionsOnly(failure))
	}
	b.WriteString("\nRepair the implementation or restore the starting version.\n")
	fmt.Fprintf(&b, "Done when LOC <= %d and the contract test passes.\n", start)
	return b.String()
}

// assertionsOnly reduces a go test failure to the lines a model can act on: the
// assertion messages, without the positions that would name the hidden test.
func assertionsOnly(failure string) string {
	var kept []string
	for line := range strings.SplitSeq(failure, "\n") {
		line = strings.TrimSpace(filePosition.ReplaceAllString(line, ""))
		line = strings.TrimSpace(strings.TrimPrefix(line, ":"))
		if line == "" || strings.HasPrefix(line, "FAIL") || strings.HasPrefix(line, "ok ") {
			continue
		}
		if !slices.Contains(kept, line) {
			kept = append(kept, line)
		}
		if len(kept) == 8 {
			break
		}
	}
	return strings.Join(kept, "\n")
}

// claudeSession runs one headless refactoring session in work with the arm's
// plugin loaded. armDir is empty for the control arm, which also loses the Skill
// tool so it cannot reach a skill the operator installed outside the plugin.
func claudeSession(o options, armDir, work, prompt string) ([]byte, error) {
	tools := "Skill,Read,Glob,Grep,Edit,Write"
	args := []string{
		"-p", prompt,
		"--output-format", "stream-json", "--verbose",
		"--max-turns", fmt.Sprint(maxSteps),
		"--permission-mode", "acceptEdits",
		// The arms differ only in skill text, so the run must not pick up the
		// operator's own settings or hooks on top of the plugin under test.
		"--restricted",
	}
	if armDir == "" {
		tools = strings.TrimPrefix(tools, "Skill,")
	} else {
		pluginDir, err := evalplugin.Copy(armDir, work)
		if err != nil {
			return nil, err
		}
		args = append(args, "--plugin-dir", pluginDir)
	}
	args = append(args, "--tools", tools, "--allowed-tools", tools)
	if o.model != "" {
		args = append(args, "--model", o.model)
	}
	if o.effort != "" {
		args = append(args, "--effort", o.effort)
	}
	return claude(o.timeout, work, args...)
}

// goldenCollision matches the compiler reporting that the golden overlay and
// the model's production code declare the same name. hideTestFiles already
// takes the model's own test files out of the build, so what is left to collide
// is a production declaration — a helper named serve against the overlay's own
// serve, which is what happened once in the 5×5 concision run.
//
// Only this exact shape is excused. A golden that fails to build because the
// model renamed or deleted an exported symbol says "undefined", and that is a
// broken contract the model owns, not a harness fault.
var goldenCollision = regexp.MustCompile(`redeclared in this block|other declaration of`)

// classifyGolden splits a failed golden run into the two causes that must never
// be averaged together. built is whether the model's package compiled before
// the overlay was added.
func classifyGolden(built bool, output string) (behavior, harness bool) {
	if !built {
		return false, false
	}
	if goldenCollision.MatchString(output) {
		return false, true
	}
	return true, false
}

// lineCountClaim matches a line or LOC count in the model's final message: a
// number next to the word, in either order, within one clause. It is a
// heuristic about prose and is reported as such — a session can state a count
// it never measured, and the trace is what settles that.
var lineCountClaim = regexp.MustCompile(`(?i)(\d+\s*(?:production\s+)?(?:lines?|loc)\b|\b(?:lines?|loc)\b[^.;\n]{0,24}?\d+)`)

// reportedCounts reports whether final states a line count at all. A file:line
// reference is not a count, so a bare path is stripped before the search.
func reportedCounts(final string) bool {
	if final == "" {
		return false
	}
	return lineCountClaim.MatchString(filePosition.ReplaceAllString(final, ""))
}

// filePosition matches a compiler-style path:line reference, which names a
// location rather than counting anything.
var filePosition = regexp.MustCompile(`[\w./-]+\.go:\d+(?::\d+)?`)

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

// fixtureDigest hashes every file under dir. The structural metrics cannot tell
// a refactor that happened to keep every count from a session that never wrote
// anything, and those two have to be told apart.
func fixtureDigest(dir string) (string, error) {
	h := sha256.New()
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
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
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		h.Write([]byte(filepath.ToSlash(rel)))
		h.Write([]byte{0})
		h.Write(data)
		h.Write([]byte{0})
		return nil
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
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

// fixHunks counts the modernizations `go fix -diff` still proposes for the
// package. It is the cheapest objective reading of "reaches for what the
// toolchain already ships": zero means no analyzer has anything left to say
// about the code in the tree, and the count is deterministic and costs no
// model call.
//
// go fix -diff exits non-zero exactly when the diff is not empty, so the exit
// status carries no error information. A real failure — most often a package
// that does not type-check — shows up as diagnostics on stderr with an empty
// diff on stdout, and that case has to stay distinguishable from a package
// with nothing left to fix rather than being recorded as a clean zero.
func fixHunks(timeout time.Duration, work, pkg string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "fix", "-diff", "./"+pkg)
	cmd.Dir = work
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if ctx.Err() != nil {
		return 0, fmt.Errorf("go fix -diff timed out after %s", timeout)
	}
	if diagnostics := strings.TrimSpace(stderr.String()); diagnostics != "" {
		return 0, fmt.Errorf("go fix -diff: %s", diagnostics)
	}
	if err != nil && len(out) == 0 {
		return 0, fmt.Errorf("go fix -diff: %w", err)
	}
	return countFixHunks(out), nil
}

// fixNote names the reading that could not be taken, so an absent count is
// never read as a zero one. It is empty when both readings came back.
func fixNote(before, after error) string {
	var notes []string
	if before != nil {
		notes = append(notes, "before: "+firstLine(before.Error()))
	}
	if after != nil {
		notes = append(notes, "after: "+firstLine(after.Error()))
	}
	return strings.Join(notes, "; ")
}

// countFixHunks counts the unified-diff hunks in a go fix diff, skipping the
// ones in a _test.go file so the number follows the same production-only rule
// as the line count: a session can neither improve nor worsen it by writing a
// test. A file header is `--- <path> (old)`, which is why both ends of the
// line are checked — a removed line of Go source can begin with `--- ` too.
func countFixHunks(diff []byte) int {
	hunks, production := 0, false
	for line := range strings.Lines(string(diff)) {
		line = strings.TrimRight(line, "\n")
		switch {
		case strings.HasPrefix(line, "--- ") && strings.HasSuffix(line, " (old)"):
			production = !strings.HasSuffix(strings.TrimSuffix(line, " (old)"), "_test.go")
		case production && strings.HasPrefix(line, "@@"):
			hunks++
		}
	}
	return hunks
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
			// A method on an unexported type adds no reachable surface, so only
			// plain functions count toward the public API here.
			if d.Recv == nil && d.Name.IsExported() {
				m.Exported++
			}
		case *ast.GenDecl:
			switch d.Tok {
			case token.TYPE:
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					m.Types++
					m.Pattern += patternHits(ts.Name.Name)
					if ts.Name.IsExported() {
						m.Exported++
					}
					if _, ok := ts.Type.(*ast.InterfaceType); ok {
						m.Interfaces++
					}
				}
			case token.VAR, token.CONST:
				for _, spec := range d.Specs {
					vs, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for _, name := range vs.Names {
						if name.IsExported() {
							m.Exported++
						}
					}
				}
			}
		}
	}
	m.Branches += branchCount(file)
}

// branchCount counts the decision points in a file: if, for, range, and each
// case that names a value. A switch or select statement is not counted itself,
// because its cases already are, and a default clause is not a decision.
func branchCount(file *ast.File) int {
	count := 0
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt:
			count++
		case *ast.CaseClause:
			count += len(node.List)
		case *ast.CommClause:
			if node.Comm != nil {
				count++
			}
		}
		return true
	})
	return count
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

// parseClaudeStream pulls the go-* skills the model invoked, its final message,
// and the session cost out of a stream-json transcript.
func parseClaudeStream(out []byte) (skills []string, final string, cost float64) {
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
			if usd, ok := obj["total_cost_usd"].(float64); ok {
				cost = usd
			}
		}
	}
	return sortedKeys(fired), final, cost
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

// checkArmFrontmatter rejects a materialized arm that carries a SKILL.md whose
// frontmatter no YAML parser will accept, before any CLI is asked to load it.
//
// This runs for every runner, including claude, which has no listing command to
// check an arm home against. It tests one failure mode rather than validating
// YAML: an unquoted single-line scalar containing a colon followed by a space,
// which YAML reads as a nested mapping key and rejects. That is what broke
// `go-code` on 2026-09-08 and cost 80 sessions, and a description is the field
// most likely to want a colon in it.
func checkArmFrontmatter(armDir string) error {
	root := filepath.Join(armDir, "skills")
	entries, err := os.ReadDir(root)
	if err != nil {
		return fmt.Errorf("read arm skills: %w", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(root, e.Name(), "SKILL.md")
		text, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if err := checkFrontmatter(string(text)); err != nil {
			return fmt.Errorf("%s/SKILL.md: %w", e.Name(), err)
		}
	}
	return nil
}

// checkFrontmatter reports the first frontmatter line of a SKILL.md that a YAML
// parser would reject. It returns nil when the file has no frontmatter at all,
// which is a different problem and one every loader reports for itself.
func checkFrontmatter(text string) error {
	rest, ok := strings.CutPrefix(text, "---\n")
	if !ok {
		return nil
	}
	block, _, ok := strings.Cut(rest, "\n---")
	if !ok {
		return errors.New("frontmatter is not closed")
	}
	for line := range strings.SplitSeq(block, "\n") {
		key, value, ok := strings.Cut(line, ": ")
		if !ok || strings.HasPrefix(line, " ") || strings.HasPrefix(line, "#") {
			continue
		}
		// A quoted or block scalar may contain anything; a plain one may not
		// contain ": ", which starts a mapping the parser has nowhere to put.
		if strings.ContainsAny(value[:1], `"'>|[{&*`) {
			continue
		}
		if strings.Contains(value, ": ") {
			return fmt.Errorf("%s value contains an unquoted colon-space and will not parse as YAML; quote it or use an em-dash", key)
		}
	}
	return nil
}

// armSkillNames lists the go-* skills present in one materialized arm tree,
// which is the set that arm's home is supposed to put in front of the model. It
// returns nil for the control arm, whose tree is empty by construction.
//
// A directory counts only if it holds a SKILL.md, because that file is what
// every runner discovers; a stray directory is not a skill.
func armSkillNames(armDir string) ([]string, error) {
	if armDir == "" {
		return nil, nil
	}
	root := filepath.Join(armDir, "skills")
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read arm skills: %w", err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "go-") {
			continue
		}
		if _, err := os.Stat(filepath.Join(root, e.Name(), "SKILL.md")); err != nil {
			continue
		}
		names = append(names, e.Name())
	}
	return names, nil
}

// checkArmSkills compares the go-* skills a runner's own listing reports against
// the ones the arm tree contains, and is the precondition the whole comparison
// rests on.
//
// The exact set matters, not the count. A skill the CLI cannot parse — one
// unquoted `: ` in a description is enough to make its YAML frontmatter
// invalid — is dropped from the listing silently, with the reason on stderr and
// a zero exit status. An earlier version of this check only rejected an empty
// set, so a broken skill in a 24-skill tree passed it, and 80 sessions of
// 2026-09-08 measured a baseline arm with no router in it. reason carries
// whatever the CLI wrote to stderr while listing, which is where it explains
// itself; it is quoted back only when something is actually missing.
func checkArmSkills(armName, armDir string, loaded []string, reason string) error {
	want, err := armSkillNames(armDir)
	if err != nil {
		return fmt.Errorf("%s arm: %w", armName, err)
	}
	if armName == controlArm {
		if len(loaded) != 0 {
			return fmt.Errorf("%s arm home loads %d go-* skills (%s); skill discovery is not isolated",
				controlArm, len(loaded), strings.Join(loaded, ", "))
		}
		return nil
	}
	if len(want) == 0 {
		return fmt.Errorf("%s arm tree contains no go-* skills", armName)
	}
	have := map[string]bool{}
	for _, name := range loaded {
		have[name] = true
	}
	var missing []string
	for _, name := range want {
		if !have[name] {
			missing = append(missing, name)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	err = fmt.Errorf("%s arm home loads %d of %d go-* skills; missing %s",
		armName, len(loaded), len(want), strings.Join(missing, ", "))
	if reason = strings.TrimSpace(reason); reason != "" {
		err = fmt.Errorf("%w: %s", err, reason)
	}
	return err
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
	fmt.Printf("[%s] %-24s %-10s #%d  lines %+d  types %+d  funcs %+d  exp %+d  branch %+d  build=%v model_tests=%s golden=%v gate=%v skills=%v",
		status, r.Arm, r.Task, r.Rep, r.Delta.Lines, r.Delta.Types, r.Delta.Funcs, r.Delta.Exported, r.Delta.Branches, r.Build, r.ModelTests, r.Golden, r.LineGatePass, r.Skills)
	if r.Err != "" {
		fmt.Printf("  error: %s", r.Err)
	}
	fmt.Println()
	if r.GoFail != "" {
		fmt.Printf("       %s\n", firstLine(r.GoFail))
	}
	if r.ModelTestFail != "" {
		fmt.Printf("       model tests: %s\n", firstLine(r.ModelTestFail))
	}
	if r.HarnessFailure {
		fmt.Printf("       golden failed on a name collision with the model's code, not on behavior\n")
	}
	if r.RepairFired && r.PreRepair != nil {
		fmt.Printf("       repair turn: lines %+d -> %+d, golden %v -> %v\n",
			r.PreRepair.Lines-r.Before.Lines, r.Delta.Lines, r.PreRepairGolden, r.Golden)
	}
	if pending := fixColumn(r); pending != "" {
		fmt.Printf("       go fix pending: %s hunk(s) before->after", pending)
		if r.FixUnmeasured != "" {
			fmt.Printf(" (%s)", r.FixUnmeasured)
		}
		fmt.Println()
	}
	if r.WorkDir != "" {
		fmt.Printf("       source: %s\n", r.WorkDir)
	}
	if r.Trace != "" {
		fmt.Printf("       trace: %s (%d command(s), counts reported=%v)\n", r.Trace, r.Commands, r.ReportedCounts)
	}
	if verbose && r.Output != "" {
		fmt.Println("       --- output ---")
		fmt.Println(r.Output)
	}
}

// resultStatus labels one run for the operator. A harness collision gets its
// own label rather than being folded into ERR: it is still excluded from the
// means, because the run produced no behavioral verdict, but the reason it was
// excluded is the harness and not the model, and only the label carries that.
func resultStatus(r result) string {
	if r.HarnessFailure {
		return "HRN"
	}
	if r.Err != "" || !r.Build || !r.Golden || r.ModelTests == "fail" || !r.Edited || r.Leaked {
		return "ERR"
	}
	return "ok "
}

// fixColumn renders one run's pending modernizations as before->after. A
// reading that could not be taken prints as n/a rather than as a zero, which
// would read as a package with nothing left to modernize.
func fixColumn(r result) string {
	if r.FixHunksBefore == nil && r.FixHunks == nil && r.FixUnmeasured == "" {
		return ""
	}
	count := func(n *int) string {
		if n == nil {
			return "n/a"
		}
		return strconv.Itoa(*n)
	}
	return count(r.FixHunksBefore) + "->" + count(r.FixHunks)
}

type armSummary struct {
	Runs              int
	Errors            int
	Valid             int
	Build             int
	Golden            int
	SkillFired        int
	Lines             int
	Types             int
	Interfaces        int
	Funcs             int
	Pattern           int
	Exported          int
	Branches          int
	NoEdit            int
	Leaked            int
	ModelTestFailures int
	// BehaviorFailures and HarnessFailures split the golden failures. Only the
	// first is a claim about the model; reporting a raw golden rate that mixes
	// them in overstates one arm's regressions.
	BehaviorFailures int
	HarnessFailures  int
	// LineGatePasses and EmptyDiffs count over completed runs, not valid ones: a
	// gate the model met while breaking behavior still has to be visible.
	LineGatePasses int
	EmptyDiffs     int
	ReportedCounts int
	// RepairsFired and RepairsRescued count the loop: how often a second turn
	// was granted, and how often the run cleared both the gate and the golden
	// afterwards. The second number is the only one that says the loop worked.
	RepairsFired   int
	RepairsRescued int
	// FixBefore and FixAfter total the pending go fix hunks over the valid
	// runs that could be read at both ends, FixMeasured counts those runs, and
	// FixClean the ones the toolchain had nothing left to propose for. Runs
	// whose diff could not be taken are left out of all four rather than
	// averaged in as clean.
	FixBefore   int
	FixAfter    int
	FixMeasured int
	FixClean    int
	// Cost covers every session that reported one, including the invalid runs:
	// a failed session still spends money, so excluding it would understate
	// what the corpus costs to replay.
	Cost   float64
	Costed int
}

// skillFired reports whether a run reached the skill its corpus is about. The
// refactor corpus has one owner for every fixture, go-code-refactor. An
// implementation task instead routes to whichever go-* skill owns its topic —
// go-http for a server, go-defensive for a boundary copy — so there the only
// question the summary can ask of every fixture at once is whether the plugin
// was reached at all. Which skill it was stays per-run in the report.
func skillFired(corpus string, skills []string) bool {
	if corpus == corpusImplement {
		return len(skills) > 0
	}
	return slices.Contains(skills, "go-code-refactor")
}

func summarizeArm(rep report, name string) armSummary {
	var summary armSummary
	for _, r := range rep.Results {
		if r.Arm != name {
			continue
		}
		summary.Runs++
		if r.Cost > 0 {
			summary.Cost += r.Cost
			summary.Costed++
		}
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
		if r.ModelTests == "fail" {
			summary.ModelTestFailures++
		}
		if r.BehaviorFailure {
			summary.BehaviorFailures++
		}
		if r.HarnessFailure {
			summary.HarnessFailures++
		}
		if r.LineGatePass {
			summary.LineGatePasses++
		}
		if r.EmptyDiff {
			summary.EmptyDiffs++
		}
		if r.ReportedCounts {
			summary.ReportedCounts++
		}
		if r.RepairFired {
			summary.RepairsFired++
			if r.LineGatePass && r.Golden {
				summary.RepairsRescued++
			}
		}
		if skillFired(rep.Corpus, r.Skills) {
			summary.SkillFired++
		}
		if !r.Edited {
			summary.NoEdit++
		}
		if r.Leaked {
			summary.Leaked++
		}
		if resultStatus(r) != "ok " {
			continue
		}
		summary.Valid++
		summary.Lines += r.Delta.Lines
		summary.Types += r.Delta.Types
		summary.Interfaces += r.Delta.Interfaces
		summary.Funcs += r.Delta.Funcs
		summary.Pattern += r.Delta.Pattern
		summary.Exported += r.Delta.Exported
		summary.Branches += r.Delta.Branches
		if r.FixHunksBefore != nil && r.FixHunks != nil {
			summary.FixMeasured++
			summary.FixBefore += *r.FixHunksBefore
			summary.FixAfter += *r.FixHunks
			if *r.FixHunks == 0 {
				summary.FixClean++
			}
		}
	}
	return summary
}

// printSummary averages structural deltas only over runs that both build and
// pass their model tests (if present) and hidden golden test.
// Failed sessions remain visible in the counts.
func printSummary(rep report) {
	fmt.Printf("%-24s %5s %6s %5s %8s %7s %7s %7s %9s %6s %8s %7s %7s %6s %8s\n",
		"arm", "runs", "errors", "valid", "Δlines", "Δtypes", "Δiface", "Δfuncs", "Δpattern", "Δexp", "Δbranch", "build", "golden", "skill", "$/run")
	for _, a := range rep.Arms {
		summary := summarizeArm(rep, a.Name)
		completed := summary.Runs - summary.Errors
		if summary.ModelTestFailures > 0 {
			fmt.Printf("%-24s   %d run(s) failed model tests and were excluded\n", a.Name, summary.ModelTestFailures)
		}
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
		cost := 0.0
		if summary.Costed > 0 {
			cost = summary.Cost / float64(summary.Costed)
		}
		fmt.Printf("%-24s %5d %6d %5d %8.1f %7.2f %7.2f %7.2f %9.2f %6.2f %8.2f %6d%% %6d%% %5d%% %8.4f\n",
			a.Name, summary.Runs, summary.Errors, summary.Valid,
			mean(summary.Lines), mean(summary.Types), mean(summary.Interfaces), mean(summary.Funcs), mean(summary.Pattern),
			mean(summary.Exported), mean(summary.Branches),
			percent(summary.Build), percent(summary.Golden), percent(summary.SkillFired), cost)
		// Neither of these belongs in a column: they are not a worse score, they
		// are a reason to distrust the row above them and go read the report.
		if summary.NoEdit > 0 {
			fmt.Printf("%-24s   %d run(s) changed no file, %d of them deliberately\n", "", summary.NoEdit, summary.EmptyDiffs)
		}
		if summary.Leaked > 0 {
			fmt.Printf("%-24s   %d run(s) referenced the repository and were excluded\n", "", summary.Leaked)
		}
		if summary.HarnessFailures > 0 {
			fmt.Printf("%-24s   %d golden failure(s) were harness collisions, not behavior changes\n", "", summary.HarnessFailures)
		}
		fmt.Printf("%-24s   line gate %d/%d, behavior failures %d, counts reported %d/%d\n",
			"", summary.LineGatePasses, completed, summary.BehaviorFailures, summary.ReportedCounts, completed)
		if summary.FixMeasured > 0 {
			fixMean := func(sum int) float64 { return float64(sum) / float64(summary.FixMeasured) }
			fmt.Printf("%-24s   go fix pending %.2f -> %.2f hunk(s)/run, nothing left to propose in %d/%d\n",
				"", fixMean(summary.FixBefore), fixMean(summary.FixAfter), summary.FixClean, summary.FixMeasured)
		}
		if summary.RepairsFired > 0 {
			fmt.Printf("%-24s   repair turn fired %d time(s), cleared gate and golden in %d\n",
				"", summary.RepairsFired, summary.RepairsRescued)
		}
	}
}
