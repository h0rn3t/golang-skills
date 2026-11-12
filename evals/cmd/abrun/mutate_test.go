package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestMutants(t *testing.T) {
	tests := []struct {
		name    string
		body    string   // statements of func f(a, b int) int, starting on line 4
		want    []string // "kind line: the mutated line"
		wantErr bool
	}{
		{name: "equal", body: "\t_ = a == b", want: []string{"flip 4: _ = a != b"}},
		{name: "not equal", body: "\t_ = a != b", want: []string{"flip 4: _ = a == b"}},
		{name: "less", body: "\t_ = a < b", want: []string{"flip 4: _ = a >= b", "boundary 4: _ = a <= b"}},
		{name: "less or equal", body: "\t_ = a <= b", want: []string{"flip 4: _ = a > b", "boundary 4: _ = a < b"}},
		{name: "greater", body: "\t_ = a > b", want: []string{"flip 4: _ = a <= b", "boundary 4: _ = a >= b"}},
		{name: "greater or equal", body: "\t_ = a >= b", want: []string{"flip 4: _ = a < b", "boundary 4: _ = a > b"}},
		{
			name: "if is dropped and its condition mutated",
			body: "\tif a > 0 {\n\t\treturn a\n\t}\n\treturn b",
			want: []string{"drop-if 4: if (a > 0) && false {", "flip 4: if a <= 0 {", "boundary 4: if a >= 0 {"},
		},
		{
			name: "if keeps its init statement",
			body: "\tif n := a + b; n != 0 {\n\t\treturn n\n\t}\n\treturn 0",
			want: []string{"drop-if 4: if n := a + b; (n != 0) && false {", "flip 4: if n := a + b; n == 0 {"},
		},
		{
			name: "each site reports its own line",
			body: "\t_ = b\n\n\tif a == 0 {\n\t\treturn 0\n\t}\n\treturn b",
			want: []string{"drop-if 6: if (a == 0) && false {", "flip 6: if a != 0 {"},
		},
		{
			name: "arithmetic and logic are left alone",
			body: "\tok := a*b+a&b > 0 || false\n\t_ = ok\n\treturn a - b",
			want: []string{"flip 4: ok := a*b+a&b <= 0 || false", "boundary 4: ok := a*b+a&b >= 0 || false"},
		},
		{name: "nothing to mutate", body: "\treturn a + b"},
		{name: "source that does not parse", body: "\treturn a +", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			src := "package p\n\nfunc f(a, b int) int {\n" + tt.body + "\n}\n"

			got, err := mutants("p.go", []byte(src))

			if (err != nil) != tt.wantErr {
				t.Fatalf("mutants(%q) error = %v, wantErr %v", tt.body, err, tt.wantErr)
			}
			var lines []string
			for _, m := range got {
				if m.file != "p.go" {
					t.Errorf("mutants(%q) file = %q, want p.go", tt.body, m.file)
				}
				if _, err := parser.ParseFile(token.NewFileSet(), m.file, m.src, parser.SkipObjectResolution); err != nil {
					t.Errorf("mutants(%q) %s at line %d does not parse: %v", tt.body, m.kind, m.line, err)
				}
				mutated := strings.Split(string(m.src), "\n")[m.line-1]
				lines = append(lines, fmt.Sprintf("%s %d: %s", m.kind, m.line, strings.TrimSpace(mutated)))
			}
			if !slices.Equal(lines, tt.want) {
				t.Errorf("mutants(%q) =\n%s\nwant\n%s", tt.body, strings.Join(lines, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
}

func TestPickMutants(t *testing.T) {
	ten := make([]mutant, 10)
	for i := range ten {
		ten[i].line = i
	}
	tests := []struct {
		name  string
		limit int
		want  []int // the lines picked
	}{
		{name: "under the limit keeps all", limit: 12, want: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
		{name: "at the limit keeps all", limit: 10, want: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
		{name: "over the limit samples evenly", limit: 3, want: []int{0, 3, 6}},
		{name: "one", limit: 1, want: []int{0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []int
			for _, m := range pickMutants(ten, tt.limit) {
				got = append(got, m.line)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("pickMutants(10 mutants, %d) lines = %v, want %v", tt.limit, got, tt.want)
			}
		})
	}
}

// clampSource has six mutation sites: a drop, a flip and a boundary move on
// each of its two guards. clampTests reach the first guard and the middle of
// the range, so they miss the second guard and both boundary moves.
const (
	clampSource = `package task

// Clamp returns v limited to [0, 10].
func Clamp(v int) int {
	if v < 0 {
		return 0
	}
	if v > 10 {
		return 10
	}
	return v
}
`
	clampTests = `package task

import "testing"

func TestClamp(t *testing.T) {
	if got := Clamp(5); got != 5 {
		t.Errorf("Clamp(5) = %d, want 5", got)
	}
	if got := Clamp(-3); got != 0 {
		t.Errorf("Clamp(-3) = %d, want 0", got)
	}
}
`
)

func TestScoreMutants(t *testing.T) {
	tests := []struct {
		name    string
		tests   string // task_test.go, or "" for a run that left none
		limit   int
		want    mutationScore
		wantErr bool
	}{
		{
			name: "tests catch what they reach", tests: clampTests, limit: 20,
			want: mutationScore{Total: 6, Caught: 3, Tests: true, Survived: []string{"task.go:5 boundary", "task.go:8 drop-if", "task.go:8 boundary"}},
		},
		{
			name: "the limit samples the sites", tests: clampTests, limit: 2,
			want: mutationScore{Total: 2, Caught: 1, Tests: true, Survived: []string{"task.go:8 drop-if"}},
		},
		{name: "no tests catch nothing", limit: 20, want: mutationScore{Total: 6}},
		{
			name:  "red tests are not scored",
			tests: strings.Replace(clampTests, "got != 5", "got != 6", 1), limit: 20, wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			work := t.TempDir()
			writeTestFile(t, filepath.Join(work, "go.mod"), "module abeval\n\ngo 1.27\n")
			writeTestFile(t, filepath.Join(work, "task", "task.go"), clampSource)
			if tt.tests != "" {
				writeTestFile(t, filepath.Join(work, "task", "task_test.go"), tt.tests)
			}

			got, err := scoreMutants(time.Minute, work, "task", tt.limit, tt.tests != "")

			if (err != nil) != tt.wantErr {
				t.Fatalf("scoreMutants(limit %d) error = %v, wantErr %v", tt.limit, err, tt.wantErr)
			}
			if err == nil && (got.Total != tt.want.Total || got.Caught != tt.want.Caught || got.Tests != tt.want.Tests || !slices.Equal(got.Survived, tt.want.Survived)) {
				t.Errorf("scoreMutants(limit %d) = %+v, want %+v", tt.limit, *got, tt.want)
			}
			// The run's tree is what the golden test is overlaid on next.
			if left, err := os.ReadFile(filepath.Join(work, "task", "task.go")); err != nil || string(left) != clampSource {
				t.Errorf("scoreMutants(limit %d) left task.go = %q, err = %v; want the source untouched", tt.limit, left, err)
			}
		})
	}
}

// TestRunOneScoresMutants pins where the score is taken: after the session,
// while its tests are still in the tree, and without disturbing the golden run.
func TestRunOneScoresMutants(t *testing.T) {
	const goldenTest = `package task

import "testing"

func TestGoldenClamp(t *testing.T) {
	if got := Clamp(12); got != 10 {
		t.Fatalf("Clamp(12) = %d, want 10", got)
	}
}
`
	for _, tt := range []struct {
		name    string
		mutants int
		want    *mutationScore
	}{
		{name: "flag set", mutants: 20, want: &mutationScore{Total: 6, Caught: 3, Tests: true}},
		{name: "flag off"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			bin, corpus, model := t.TempDir(), t.TempDir(), t.TempDir()
			cli := filepath.Join(bin, "claude")
			writeTestFile(t, cli, "#!/bin/sh\nset -eu\ncp \"$ABRUN_TEST_SOURCE\"/*.go task/\nprintf '%s\\n' '{\"type\":\"result\",\"result\":\"done\"}'\n")
			if err := os.Chmod(cli, 0o700); err != nil { //nolint:gosec // G302: the fake CLI has to be executable
				t.Fatal(err)
			}
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("ABRUN_TEST_SOURCE", model)
			writeTestFile(t, filepath.Join(corpus, "task/task.go"), "package task\n\n// Clamp returns v limited to [0, 10].\nfunc Clamp(v int) int { panic(\"not implemented\") }\n")
			writeTestFile(t, filepath.Join(corpus, "_golden/task/golden_test.go"), goldenTest)
			writeTestFile(t, filepath.Join(model, "task.go"), clampSource)
			writeTestFile(t, filepath.Join(model, "task_test.go"), clampTests)

			got := runOne(options{runner: runnerClaude, prompt: "Implement %s", timeout: time.Minute, mutants: tt.mutants}, corpus, arm{Name: "baseline"}, "task", 0)

			if status := resultStatus(got); status != "ok " {
				t.Fatalf("runOne(mutants %d) status = %q, want ok: %s %s", tt.mutants, status, got.Err, got.GoFail)
			}
			if (got.Mutation == nil) != (tt.want == nil) || got.MutationUnmeasured != "" {
				t.Fatalf("runOne(mutants %d) mutation = %+v, unmeasured = %q; want %+v", tt.mutants, got.Mutation, got.MutationUnmeasured, tt.want)
			}
			if tt.want != nil && (got.Mutation.Total != tt.want.Total || got.Mutation.Caught != tt.want.Caught || got.Mutation.Tests != tt.want.Tests) {
				t.Errorf("runOne(mutants %d) mutation = %+v, want %+v", tt.mutants, *got.Mutation, *tt.want)
			}
		})
	}
}

func TestMutationSum(t *testing.T) {
	var sum mutationSum
	if got := sum.line(); got != "" {
		t.Errorf("mutationSum{}.line() = %q, want it empty when nothing was measured", got)
	}

	sum.add(nil)
	sum.add(&mutationScore{Total: 6, Caught: 3, Tests: true})
	sum.add(&mutationScore{Total: 4})

	const want = "mutants caught: 30% (3/10) over 2 run(s); where the run left tests: 50% (3/6) over 1"
	if got := sum.line(); got != want {
		t.Errorf("mutationSum.line() = %q, want %q", got, want)
	}
}
