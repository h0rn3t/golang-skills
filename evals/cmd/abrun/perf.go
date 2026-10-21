package main

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// The perf corpus hands the model a package with one slow path and a prompt
// naming the symptom, or, in a decline fixture, a request to speed up a
// function the package's own profile shows is not worth it. The score is the
// harness's measurement, never the session's: once every session has ended,
// the golden benchmark runs serially through bench-compare.sh --verdict, the
// fixture as given against the tree the session left, so parallel sessions
// share no CPU with it. Beside the verdict the report records which functions
// the session changed (a bait is a cold function written to look slow) and
// whether a benchmark or a profile ran before the first edit.
const corpusPerf = "perf"

// perfDir holds the perf corpus. The leading underscore keeps findTasks from
// offering it as a refactor fixture, the same way implementDir does.
const perfDir = "_perf"

// perfKeyFile is the per-fixture key under _golden/<task>/.
const perfKeyFile = "key.json"

const (
	expectSpeedup = "speedup"
	expectDecline = "decline"
)

// perfKeep is the bench-compare.sh verdict for a significant improvement.
const perfKeep = "keep"

// perfRounds is how many runs of each side the harness's verdict takes.
const perfRounds = 10

// perfKey describes one fixture. Prompt is the task, %s the fixture directory.
// Bench and Benchtime select and size the golden benchmark. Targets are the
// functions the request is about and Baits cold ones that only look slow,
// both as funcKey spells them (Name, or Type.Name for a method). Class names
// what the profile shows (gc, alloc, cpu, contention, io); Note says why, for
// the reader of a report.
type perfKey struct {
	Prompt    string   `json:"prompt"`
	Expect    string   `json:"expect"`
	Class     string   `json:"class"`
	Bench     string   `json:"bench"`
	Benchtime string   `json:"benchtime"`
	Targets   []string `json:"targets"`
	Baits     []string `json:"baits"`
	Note      string   `json:"note"`
}

func loadPerfKey(path string) (perfKey, error) {
	var k perfKey
	data, err := os.ReadFile(path)
	if err != nil {
		return k, err
	}
	if err := json.Unmarshal(data, &k); err != nil {
		return k, fmt.Errorf("parse %s: %w", path, err)
	}
	switch {
	case strings.Count(k.Prompt, "%s") != 1:
		return k, fmt.Errorf("%s: prompt needs exactly one %%s for the fixture directory", path)
	case k.Expect != expectSpeedup && k.Expect != expectDecline:
		return k, fmt.Errorf("%s: expect must be %s or %s, got %q", path, expectSpeedup, expectDecline, k.Expect)
	case k.Bench == "":
		return k, fmt.Errorf("%s: bench names no golden benchmark", path)
	case k.Expect == expectDecline && len(k.Targets) == 0:
		return k, fmt.Errorf("%s: a decline fixture names the targets it must leave alone", path)
	}
	return k, nil
}

func validatePerfFixtures(abDir string, tasks []string) error {
	if err := validateFixtures(abDir, tasks); err != nil {
		return err
	}
	for _, task := range tasks {
		if _, err := loadPerfKey(filepath.Join(abDir, "_golden", task, perfKeyFile)); err != nil {
			return exitError{2, fmt.Sprintf("fixture %s: %v", task, err)}
		}
	}
	return nil
}

// perfScore is what one perf run earned. Verdict and the deltas are the
// harness's verdict on the golden benchmark; Unmeasured says why there is
// none. BenchRuns, Profiles and Verdicts count the session's own commands,
// read from a claude trace; MeasuredFirst is a benchmark or profile command
// before the first edit of a production file, or at all when there was none.
type perfScore struct {
	Expect         string   `json:"expect"`
	Class          string   `json:"class"`
	Prompt         string   `json:"prompt"`
	Verdict        string   `json:"verdict,omitempty"`
	SecDeltaPct    *float64 `json:"sec_delta_pct,omitempty"`
	AllocsDeltaPct *float64 `json:"allocs_delta_pct,omitempty"`
	Unmeasured     string   `json:"unmeasured,omitempty"`
	Changed        []string `json:"changed"`
	TargetsChanged []string `json:"targets_changed,omitempty"`
	BaitsChanged   []string `json:"baits_changed,omitempty"`
	BenchRuns      int      `json:"bench_runs"`
	Profiles       int      `json:"profiles"`
	Verdicts       int      `json:"verdicts"`
	MeasuredFirst  bool     `json:"measured_before_edit"`
	Pass           bool     `json:"pass"`
}

// decide sets Pass. A speedup fixture passes on a kept verdict with behavior
// held and no bait touched; a decline fixture passes when behavior held and
// no target changed. A decline run that changes something else is recorded in
// Changed as scope, not failed for it.
func (s *perfScore) decide(golden bool) {
	if s.Expect == expectDecline {
		s.Pass = golden && len(s.TargetsChanged) == 0
		return
	}
	s.Pass = golden && s.Verdict == perfKeep && len(s.BaitsChanged) == 0
}

// perfJob is a measurement still owed. base is a module holding the fixture as
// given plus the golden files, work the session's tree with the same golden
// files; both lay the package out as abeval/<pkg>.
type perfJob struct {
	base, work, pkg string
	key             perfKey
}

// scorePerf fills res.Perf once the golden test has run, and leaves a perfJob
// on res when there is a change to measure.
func scorePerf(abDir, task, work string, key perfKey, res *result) {
	sc := res.Perf
	before, err := productionSources(filepath.Join(abDir, task))
	if err == nil {
		sc.Changed, err = changedFuncs(before, res.Source)
	}
	if err != nil {
		sc.Unmeasured = fmt.Sprintf("compare functions: %v", err)
		sc.decide(false)
		return
	}
	sc.TargetsChanged = intersect(sc.Changed, key.Targets)
	sc.BaitsChanged = intersect(sc.Changed, key.Baits)
	if trace, err := os.ReadFile(filepath.Join(work, traceFile)); err == nil {
		sc.BenchRuns, sc.Profiles, sc.Verdicts, sc.MeasuredFirst = perfSignals(trace)
	}
	switch {
	case !res.Build:
		sc.Unmeasured = "the package does not build"
	case len(sc.Changed) == 0:
		sc.Unmeasured = "no production function changed"
	default:
		base, err := perfBase(abDir, task)
		if err != nil {
			sc.Unmeasured = fmt.Sprintf("prepare the base tree: %v", err)
			break
		}
		res.perfJob = &perfJob{base: base, work: work, pkg: task, key: key}
		return
	}
	sc.decide(res.Golden)
}

// perfBase builds the module the verdict's snapshot is taken in: the fixture
// as given, its own test files hidden as the session's are, and the golden
// files copied in.
func perfBase(abDir, task string) (string, error) {
	base, err := os.MkdirTemp("", "abrun-perf-base-")
	if err != nil {
		return "", err
	}
	pkg := filepath.Join(base, task)
	err = os.CopyFS(pkg, os.DirFS(filepath.Join(abDir, task)))
	if err == nil {
		err = hideTestFiles(pkg)
	}
	if err == nil {
		err = os.CopyFS(pkg, os.DirFS(filepath.Join(abDir, "_golden", task)))
	}
	if err == nil {
		err = os.WriteFile(filepath.Join(base, "go.mod"), []byte("module abeval\n\ngo 1.27\n"), 0o644)
	}
	if err != nil {
		_ = os.RemoveAll(base) // best-effort cleanup of a tree that failed to build
		return "", err
	}
	return base, nil
}

// measurePerfResults runs the owed measurements one at a time, after every
// session has ended, and removes the trees no longer needed.
func measurePerfResults(o options, root string, results []result) {
	script := filepath.Join(root, "skills", "go-performance", "scripts", "bench-compare.sh")
	if abs, err := filepath.Abs(script); err == nil {
		script = abs // each call runs in the tree it measures
	}
	for i := range results {
		r := &results[i]
		job := r.perfJob
		if job == nil {
			continue
		}
		r.Perf.Verdict, r.Perf.SecDeltaPct, r.Perf.AllocsDeltaPct, r.Perf.Unmeasured = measurePerf(o, script, job)
		r.Perf.decide(r.Golden)
		_ = os.RemoveAll(job.base) // best-effort cleanup after the verdict is recorded
		if !o.keep {
			_ = os.RemoveAll(job.work) // best-effort cleanup after the verdict is recorded
		}
		r.perfJob = nil
		printPerfResult(*r)
	}
}

// benchCompareJSON is the part of bench-compare.sh --json the harness reads.
type benchCompareJSON struct {
	Status     string `json:"status"`
	Verdict    string `json:"verdict"`
	Output     string `json:"output"`
	Comparison *struct {
		Rows []struct {
			Unit     string   `json:"unit"`
			DeltaPct *float64 `json:"delta_pct"`
		} `json:"rows"`
	} `json:"comparison"`
}

func measurePerf(o options, script string, job *perfJob) (verdict string, sec, allocs *float64, unmeasured string) {
	snap := filepath.Join(job.base, ".snapshot")
	if out, err := benchCompare(o, job.base, script, "--snapshot", snap, "./"+job.pkg); err != nil {
		return "", nil, nil, fmt.Sprintf("snapshot: %v", err)
	} else if out.Status != "ok" {
		return "", nil, nil, fmt.Sprintf("snapshot %s: %s", out.Status, firstLine(out.Output))
	}
	args := []string{"--verdict", snap, "-n", strconv.Itoa(perfRounds), "--filter", job.key.Bench}
	if job.key.Benchtime != "" {
		args = append(args, "--benchtime", job.key.Benchtime)
	}
	out, err := benchCompare(o, job.work, script, append(args, "./"+job.pkg)...)
	switch {
	case err != nil:
		return "", nil, nil, fmt.Sprintf("verdict: %v", err)
	case out.Status != "ok" || out.Comparison == nil:
		return "", nil, nil, fmt.Sprintf("verdict %s: %s", out.Status, firstLine(out.Output))
	}
	for _, row := range out.Comparison.Rows {
		switch {
		case row.Unit == "sec/op" && sec == nil:
			sec = row.DeltaPct
		case row.Unit == "allocs/op" && allocs == nil:
			allocs = row.DeltaPct
		}
	}
	return out.Verdict, sec, allocs, ""
}

// benchCompare runs bench-compare.sh --json in dir. Exit 1 is a verdict of
// revert or noisy, or a status the JSON names, so only a missing or
// unparsable object is an error.
func benchCompare(o options, dir, script string, args ...string) (benchCompareJSON, error) {
	ctx, cancel := context.WithTimeout(context.Background(), o.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", append([]string{script, "--json"}, args...)...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	runErr := cmd.Run()
	var out benchCompareJSON
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		return out, fmt.Errorf("%w: %s", errors.Join(runErr, err), firstLine(stderr.String()))
	}
	return out, nil
}

// changedFuncs lists the functions and methods whose declaration differs
// between two sets of production sources, added and removed ones included.
// Comments do not count, and neither does a function moved between files.
func changedFuncs(before, after map[string]string) ([]string, error) {
	old, err := funcDecls(before)
	if err != nil {
		return nil, err
	}
	cur, err := funcDecls(after)
	if err != nil {
		return nil, err
	}
	var changed []string
	for name, text := range cur {
		if prev, ok := old[name]; !ok || prev != text {
			changed = append(changed, name)
		}
	}
	for name := range old {
		if _, ok := cur[name]; !ok {
			changed = append(changed, name)
		}
	}
	slices.Sort(changed)
	return changed, nil
}

// funcDecls maps funcKey to the declaration's tokens, from the func keyword
// on, so neither the doc comment nor the layout of the body counts.
func funcDecls(sources map[string]string) (map[string]string, error) {
	decls := map[string]string{}
	fset := token.NewFileSet()
	for path, src := range sources {
		file, err := parser.ParseFile(fset, path, src, 0)
		if err != nil {
			return nil, err
		}
		tf := fset.File(file.Pos())
		for _, d := range file.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok {
				decls[funcKey(fn)] = tokens(src[tf.Offset(fn.Pos()):tf.Offset(fn.End())])
			}
		}
	}
	return decls, nil
}

// tokens renders Go source as its token sequence without comments and
// semicolons, the two things a line break or a reflow changes.
func tokens(src string) string {
	var s scanner.Scanner
	s.Init(token.NewFileSet().AddFile("", -1, len(src)), []byte(src), nil, 0)
	var b strings.Builder
	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			return b.String()
		}
		if tok != token.SEMICOLON {
			b.WriteString(tok.String())
			b.WriteString(lit)
			b.WriteByte(' ')
		}
	}
}

// funcKey is Name for a function and Type.Name for a method.
func funcKey(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return fn.Name.Name
	}
	t := fn.Recv.List[0].Type
	if star, ok := t.(*ast.StarExpr); ok {
		t = star.X
	}
	switch x := t.(type) {
	case *ast.IndexExpr:
		t = x.X
	case *ast.IndexListExpr:
		t = x.X
	}
	if id, ok := t.(*ast.Ident); ok {
		return id.Name + "." + fn.Name.Name
	}
	return fn.Name.Name
}

func intersect(names, want []string) []string {
	var out []string
	for _, n := range names {
		if slices.Contains(want, n) {
			out = append(out, n)
		}
	}
	return out
}

// perfSignals reads a claude stream-json trace in order: Bash calls that run
// a benchmark, take or read a profile, or ask bench-compare.sh for a verdict,
// and whether one of the first two came before the first edit of a
// production .go file. Other runners' traces count nothing.
func perfSignals(trace []byte) (bench, profiles, verdicts int, measuredFirst bool) {
	edited := false
	for line := range bytes.SplitSeq(trace, []byte("\n")) {
		var msg struct {
			Message struct {
				Content []struct {
					Type  string `json:"type"`
					Name  string `json:"name"`
					Input struct {
						Command  string `json:"command"`
						FilePath string `json:"file_path"`
					} `json:"input"`
				} `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal(line, &msg) != nil {
			continue
		}
		for _, c := range msg.Message.Content {
			if c.Type != "tool_use" {
				continue
			}
			switch c.Name {
			case "Bash":
				cmd := c.Input.Command
				isBench := strings.Contains(cmd, "-bench") || strings.Contains(cmd, "bench-compare.sh")
				isProfile := strings.Contains(cmd, "profile") || strings.Contains(cmd, "pprof") || strings.Contains(cmd, "-trace")
				if isBench {
					bench++
				}
				if isProfile {
					profiles++
				}
				if strings.Contains(cmd, "bench-compare.sh") && strings.Contains(cmd, "--verdict") {
					verdicts++
				}
				if (isBench || isProfile) && !edited {
					measuredFirst = true
				}
			case "Edit", "Write", "MultiEdit":
				if p := c.Input.FilePath; strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go") {
					edited = true
				}
			}
		}
	}
	return bench, profiles, verdicts, measuredFirst
}

func printPerfResult(r result) {
	p := r.Perf
	delta := func(v *float64) string {
		if v == nil {
			return "-"
		}
		return fmt.Sprintf("%+.1f%%", *v)
	}
	fmt.Printf("[perf] %-24s %-8s #%d  %-7s verdict=%s sec/op=%s allocs/op=%s golden=%v pass=%v changed=%v baits=%v measured_first=%v bench=%d profiles=%d verdicts=%d\n",
		r.Arm, r.Task, r.Rep, p.Expect, cmp.Or(p.Verdict, "-"), delta(p.SecDeltaPct), delta(p.AllocsDeltaPct), r.Golden, p.Pass, p.Changed, p.BaitsChanged, p.MeasuredFirst, p.BenchRuns, p.Profiles, p.Verdicts)
	if p.Unmeasured != "" {
		fmt.Printf("       unmeasured: %s\n", p.Unmeasured)
	}
}

// printPerfSummary is the perf corpus's table: one row per arm, then the
// passes per fixture, since the speedup fixtures and the decline one are
// scored on different things and a mean over both hides which one moved.
func printPerfSummary(rep report) {
	fmt.Printf("%-24s %5s %6s %7s %5s %5s %6s %8s %7s %7s %7s %8s\n",
		"arm", "runs", "errors", "golden", "pass", "keep", "baits", "measured", "bench/r", "prof/r", "verd/r", "$/run")
	var tasks []string
	for _, a := range rep.Arms {
		var runs, errs, golden, pass, keep, baits, first, bench, profiles, verdicts, costed int
		var cost float64
		for _, r := range rep.Results {
			if r.Arm != a.Name {
				continue
			}
			runs++
			if !slices.Contains(tasks, r.Task) {
				tasks = append(tasks, r.Task)
			}
			if r.Cost > 0 {
				cost += r.Cost
				costed++
			}
			if r.Err != "" || r.Perf == nil {
				errs++
				continue
			}
			p := r.Perf
			for _, hit := range []struct {
				ok bool
				n  *int
			}{{r.Golden, &golden}, {p.Pass, &pass}, {p.Verdict == perfKeep, &keep}, {len(p.BaitsChanged) > 0, &baits}, {p.MeasuredFirst, &first}} {
				if hit.ok {
					*hit.n++
				}
			}
			bench += p.BenchRuns
			profiles += p.Profiles
			verdicts += p.Verdicts
		}
		done := runs - errs
		perRun := func(n int) float64 {
			if done == 0 {
				return 0
			}
			return float64(n) / float64(done)
		}
		if costed > 0 {
			cost /= float64(costed)
		}
		fmt.Printf("%-24s %5d %6d %7s %5s %5d %6d %8s %7.1f %7.1f %7.1f %8.4f\n",
			a.Name, runs, errs, fmt.Sprintf("%d/%d", golden, done), fmt.Sprintf("%d/%d", pass, done), keep, baits,
			fmt.Sprintf("%d/%d", first, done), perRun(bench), perRun(profiles), perRun(verdicts), cost)
		printRoutingSummary(rep, a.Name)
	}
	slices.Sort(tasks)
	for _, task := range tasks {
		fmt.Printf("  %-10s", task)
		for _, a := range rep.Arms {
			var n, pass int
			for _, r := range rep.Results {
				if r.Arm == a.Name && r.Task == task && r.Perf != nil {
					n++
					if r.Perf.Pass {
						pass++
					}
				}
			}
			fmt.Printf("  %s %d/%d", a.Name, pass, n)
		}
		fmt.Println()
	}
}
