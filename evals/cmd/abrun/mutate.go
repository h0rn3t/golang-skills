package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"time"
)

// The -mutants score asks what the tests a session wrote are worth: small bugs
// are injected into the production code it left, one at a time, and its own
// tests run against each. The three operators keep every type and the
// statement structure as they were, so a mutant always compiles and a failing
// run means a test noticed the bug.
var (
	// flipped negates a comparison.
	flipped = map[token.Token]token.Token{
		token.EQL: token.NEQ, token.NEQ: token.EQL,
		token.LSS: token.GEQ, token.GEQ: token.LSS,
		token.GTR: token.LEQ, token.LEQ: token.GTR,
	}
	// boundary moves a relational comparison's edge by one.
	boundary = map[token.Token]token.Token{
		token.LSS: token.LEQ, token.LEQ: token.LSS,
		token.GTR: token.GEQ, token.GEQ: token.GTR,
	}
)

// mutant is one injected bug.
type mutant struct {
	file string // relative to the package directory
	line int
	kind string // flip, boundary or drop-if
	src  []byte // the whole file with the bug in place
}

// mutationScore is how many injected bugs one run's own tests caught. A run
// that left no tests catches none, and lists no survivors.
type mutationScore struct {
	Total    int      `json:"total"`
	Caught   int      `json:"caught"`
	Tests    bool     `json:"tests"`
	Survived []string `json:"survived,omitempty"` // file:line kind
}

// mutants returns every bug the operators can inject into one production
// file, in the order ast.Inspect reaches the sites.
func mutants(name string, src []byte) ([]mutant, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, name, src, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}
	tf := fset.File(file.Pos())
	var out []mutant
	add := func(kind string, pos, end token.Pos, text string) {
		mutated := slices.Concat(src[:tf.Offset(pos)], []byte(text), src[tf.Offset(end):])
		out = append(out, mutant{file: name, line: tf.Line(pos), kind: kind, src: mutated})
	}
	ast.Inspect(file, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.BinaryExpr:
			end := n.OpPos + token.Pos(len(n.Op.String()))
			if op, ok := flipped[n.Op]; ok {
				add("flip", n.OpPos, end, op.String())
			}
			if op, ok := boundary[n.Op]; ok {
				add("boundary", n.OpPos, end, op.String())
			}
		case *ast.IfStmt:
			// The condition is still evaluated, so what it declares and calls
			// stays used; only the body stops running.
			cond := src[tf.Offset(n.Cond.Pos()):tf.Offset(n.Cond.End())]
			add("drop-if", n.Cond.Pos(), n.Cond.End(), "("+string(cond)+") && false")
		}
		return true
	})
	return out, nil
}

// pickMutants returns at most limit of ms, evenly spaced, so a cap samples the
// whole package and not its first file.
func pickMutants(ms []mutant, limit int) []mutant {
	if len(ms) <= limit {
		return ms
	}
	picked := make([]mutant, limit)
	for i := range picked {
		picked[i] = ms[i*len(ms)/limit]
	}
	return picked
}

// scoreMutants injects up to limit bugs into the package the run left in
// work/taskName and runs the run's own tests against each. The tests run in a
// throwaway copy of the module, so work stays as the session left it. hasTests
// is whether the run left any test: without one every mutant is missed and
// nothing is executed. Tests that fail before any bug is injected are an
// error, since a red suite catches every mutant for no reason.
func scoreMutants(timeout time.Duration, work, taskName string, limit int, hasTests bool) (*mutationScore, error) {
	sources, err := productionSources(filepath.Join(work, taskName))
	if err != nil {
		return nil, err
	}
	var all []mutant
	for _, name := range slices.Sorted(maps.Keys(sources)) {
		ms, err := mutants(name, []byte(sources[name]))
		if err != nil {
			return nil, err
		}
		all = append(all, ms...)
	}
	all = pickMutants(all, limit)
	score := &mutationScore{Total: len(all), Tests: hasTests}
	if !hasTests || len(all) == 0 {
		return score, nil
	}

	probe, err := os.MkdirTemp("", "abrun-mutate-")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = os.RemoveAll(probe) // best-effort cleanup of the mutation copy
	}()
	// MkdirTemp already created the directory, and CopyFS refuses to overwrite,
	// so the copy goes into a fresh child.
	root := filepath.Join(probe, "module")
	if err := os.CopyFS(root, os.DirFS(work)); err != nil {
		return nil, err
	}
	// The mutated condition `(c) && false` is what vet's bool check is for.
	args := []string{"test", "-count=1", "-vet=off", "./" + taskName + "/..."}
	start := time.Now()
	if err := goCmd(timeout, root, args...); err != nil {
		return nil, fmt.Errorf("tests fail before any mutant: %w", err)
	}
	// A dropped break condition loops forever. go test's own timeout stops the
	// test binary, which the context that kills the go command would orphan.
	args = append(args, "-timeout="+(3*time.Since(start)+10*time.Second).Round(time.Second).String())
	for _, m := range all {
		path := filepath.Join(root, taskName, m.file)
		if err := os.WriteFile(path, m.src, 0o600); err != nil {
			return nil, err
		}
		missed := goCmd(timeout, root, args...) == nil
		if err := os.WriteFile(path, []byte(sources[m.file]), 0o600); err != nil {
			return nil, err
		}
		if missed {
			score.Survived = append(score.Survived, fmt.Sprintf("%s:%d %s", m.file, m.line, m.kind))
			continue
		}
		score.Caught++
	}
	return score, nil
}

// mutationSum totals the mutation scores of an arm's valid runs. Over every
// run the rate falls when a session writes no test; over the runs that left
// tests it says how sharp those tests are.
type mutationSum struct {
	Runs, Total, Caught     int
	TestedRuns, TestedTotal int
}

func (s *mutationSum) add(m *mutationScore) {
	if m == nil {
		return
	}
	s.Runs++
	s.Total += m.Total
	s.Caught += m.Caught
	if m.Tests {
		s.TestedRuns++
		s.TestedTotal += m.Total
	}
}

// line renders both rates, or "" when no run was scored.
func (s *mutationSum) line() string {
	if s.Runs == 0 {
		return ""
	}
	return fmt.Sprintf("mutants caught: %s over %d run(s); where the run left tests: %s over %d",
		caughtRate(s.Caught, s.Total), s.Runs, caughtRate(s.Caught, s.TestedTotal), s.TestedRuns)
}

// caughtRate renders caught of total as "50% (3/6)".
func caughtRate(caught, total int) string {
	if total == 0 {
		return "— (0/0)"
	}
	return fmt.Sprintf("%d%% (%d/%d)", 100*caught/total, caught, total)
}
