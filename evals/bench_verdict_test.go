package evals_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The verdict helper ships as Go source next to bench-compare.sh; its unit
// tests run as a command-line-arguments package, no module needed.
func TestBenchVerdictUnitTests(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(repoRoot(t), "skills", "go-performance", "scripts")
	for _, args := range [][]string{
		{"vet", "./bench-verdict.go", "./bench-verdict_test.go"},
		{"test", "-count=1", "./bench-verdict.go", "./bench-verdict_test.go"},
	} {
		cmd := exec.CommandContext(t.Context(), "go", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("go %s: %v\n%s", args[0], err, out)
		}
	}
}

const (
	joinSlow  = "package join\n\n// Line returns n copies of \"x\".\nfunc Line(n int) string {\n\ts := \"\"\n\tfor range n {\n\t\ts += \"x\"\n\t}\n\treturn s\n}\n"
	joinFast  = "package join\n\nimport \"strings\"\n\n// Line returns n copies of \"x\".\nfunc Line(n int) string { return strings.Repeat(\"x\", n) }\n"
	joinBench = "package join\n\nimport \"testing\"\n\nvar sink string\n\nfunc BenchmarkLine(b *testing.B) {\n\tfor b.Loop() {\n\t\tsink = Line(64)\n\t}\n}\n"
)

type benchVerdict struct {
	Mode       string   `json:"mode"`
	Status     string   `json:"status"`
	Verdict    string   `json:"verdict"`
	ExitCode   int      `json:"exit_code"`
	Benchmarks []string `json:"benchmarks"`
	Comparison *struct {
		Verdict string `json:"verdict"`
		Rows    []struct {
			Unit   string `json:"unit"`
			Change string `json:"change"`
		} `json:"rows"`
	} `json:"comparison"`
}

// TestBenchCompareVerdict drives --snapshot and --verdict on a package whose
// two versions differ by 62 allocations per op, so the verdict does not
// depend on timing noise.
func TestBenchCompareVerdict(t *testing.T) {
	t.Parallel()
	script := filepath.Join(repoRoot(t), "skills", "go-performance", "scripts", "bench-compare.sh")
	mod := t.TempDir()
	snap := filepath.Join(t.TempDir(), "snap")
	write := func(name, text string) {
		t.Helper()
		path := filepath.Join(mod, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	bench := func(wantExit int, args ...string) benchVerdict {
		t.Helper()
		cmd := exec.CommandContext(t.Context(), "bash", append([]string{script, "--json"}, args...)...)
		cmd.Dir = mod
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		err := cmd.Run()
		gotExit := 0
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			gotExit = exitErr.ExitCode()
		} else if err != nil {
			t.Fatalf("bench-compare.sh %s: %v", strings.Join(args, " "), err)
		}
		var got benchVerdict
		if jsonErr := json.Unmarshal(stdout.Bytes(), &got); jsonErr != nil || gotExit != wantExit || got.ExitCode != wantExit {
			t.Fatalf("bench-compare.sh %s exit = %d, want %d (json: %v)\nstdout:\n%s\nstderr:\n%s", strings.Join(args, " "), gotExit, wantExit, jsonErr, stdout.String(), stderr.String())
		}
		return got
	}
	write("go.mod", "module verdictfix\n\ngo 1.27\n")
	write("join/join.go", joinSlow)
	write("join/join_test.go", joinBench)
	run := []string{"-n", "6", "--benchtime", "500x", "./join"}

	if got := bench(0, "--snapshot", snap, "./join"); got.Mode != "snapshot" || len(got.Benchmarks) != 1 {
		t.Fatalf("--snapshot = %+v, want one benchmark", got)
	}
	// -test.list sees no sub-benchmarks; the snapshot lists by the first element.
	if got := bench(0, "--snapshot", filepath.Join(t.TempDir(), "sub"), "--filter", "BenchmarkLine/n=64", "./join"); len(got.Benchmarks) != 1 {
		t.Fatalf("--snapshot --filter BenchmarkLine/n=64 = %+v, want BenchmarkLine", got)
	}
	write("join/join.go", joinFast)
	if got := bench(0, append([]string{"--verdict", snap}, run...)...); got.Verdict != "keep" || got.Comparison == nil || got.Comparison.Verdict != "keep" {
		t.Fatalf("--verdict after removing 62 allocations = %+v, want keep", got)
	}

	// A benchmark edited after the snapshot is not what gets compared: the
	// slow code behind a benchmark that no longer calls it still loses.
	bench(0, "--snapshot", snap, "--force", "./join")
	write("join/join.go", joinSlow)
	write("join/join_test.go", strings.Replace(joinBench, "sink = Line(64)", `sink = "x"`, 1))
	write("join/extra_test.go", "package join\n\nimport \"testing\"\n\nfunc BenchmarkExtra(b *testing.B) {\n\tfor b.Loop() {\n\t}\n}\n")
	got := bench(1, append([]string{"--verdict", snap}, run...)...)
	if got.Verdict != "revert" || got.Comparison == nil {
		t.Fatalf("--verdict with the benchmark gutted after the snapshot = %+v, want revert", got)
	}
	for _, r := range got.Comparison.Rows {
		if r.Unit == "allocs/op" && r.Change != "worse" {
			t.Errorf("allocs/op row change = %q, want worse: the frozen benchmark still calls Line", r.Change)
		}
	}

	runCommandInDir(t, mod, 2, "bash", script, "--snapshot", snap, "./join")
	runCommandInDir(t, mod, 2, "bash", script, "--verdict", snap, "-n", "3", "./join")
	runCommandInDir(t, mod, 2, "bash", script, "--verdict", filepath.Join(mod, "nosnap"), "./join")
	runCommandInDir(t, mod, 2, "bash", script, "--verdict", snap, "--threshold", "0", "./join")
	runCommandInDir(t, mod, 2, "bash", script, "--verdict", snap, "--threshold", "05", "./join") // not a JSON number
	runCommandInDir(t, mod, 2, "bash", script, "--verdict", snap, "--new-env", "not-a-var", "./join")
	runCommandInDir(t, mod, 2, "bash", script, "--new-env", "GOGC=400", "./join")
	write("other/other.go", "package other\n")
	runCommandInDir(t, mod, 2, "bash", script, "--snapshot", snap, "--force", "./...")
}
