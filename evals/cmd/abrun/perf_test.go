package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// TestPerfCorpusIsRunnable keeps the perf keys honest: every fixture has a
// golden test and a valid key, and every target and bait the key names is a
// function the fixture declares, so a renamed function cannot silently turn
// a bait check into a check of nothing.
func TestPerfCorpusIsRunnable(t *testing.T) {
	refactorRoot, _ := corpusRoots()
	root := filepath.Join(refactorRoot, perfDir)
	tasks, err := findTasks(root, "")
	if err != nil || len(tasks) == 0 {
		t.Fatalf("findTasks(perf corpus) = %v, %v; want fixtures", tasks, err)
	}
	if err := validatePerfFixtures(root, tasks); err != nil {
		t.Fatalf("validatePerfFixtures(%v) error = %v", tasks, err)
	}
	refactorTasks, err := findTasks(refactorRoot, "")
	if err != nil {
		t.Fatal(err)
	}
	declines := 0
	for _, task := range tasks {
		if slices.Contains(refactorTasks, task) {
			t.Errorf("fixture %q appears in the perf corpus and the refactor corpus", task)
		}
		key, err := loadPerfKey(filepath.Join(root, "_golden", task, perfKeyFile))
		if err != nil {
			t.Fatal(err)
		}
		if key.Expect == expectDecline {
			declines++
		}
		sources, err := productionSources(filepath.Join(root, task))
		if err != nil {
			t.Fatal(err)
		}
		decls, err := funcDecls(sources)
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range append(slices.Clone(key.Targets), key.Baits...) {
			if _, ok := decls[name]; !ok {
				t.Errorf("fixture %s: key names %q, which the fixture does not declare", task, name)
			}
		}
	}
	if declines == 0 {
		t.Error("the perf corpus has no decline fixture; the corpus would only reward changing code")
	}
}

func TestChangedFuncs(t *testing.T) {
	before := map[string]string{
		"a.go": "package p\n\n// F doc.\nfunc F() int { return 1 }\n\nfunc (c *C) M() {}\n\nfunc gone() {}\n",
		"b.go": "package p\n\ntype C struct{}\n\nfunc Moved() {}\n",
	}
	after := map[string]string{
		"a.go": "package p\n\n// F has a new doc comment.\nfunc F() int {\n\treturn 1 // same body, new comment\n}\n\nfunc (c *C) M() { _ = c }\n\nfunc Moved() {}\n",
		"b.go": "package p\n\ntype C struct{}\n\nfunc added() {}\n",
	}
	got, err := changedFuncs(before, after)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"C.M", "added", "gone"}; !slices.Equal(got, want) {
		t.Errorf("changedFuncs() = %v, want %v: comments and a moved function do not count", got, want)
	}
}

func TestPerfSignals(t *testing.T) {
	bash := func(cmd string) string {
		return `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"` + cmd + `"}}]}}`
	}
	edit := func(path string) string {
		return `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Edit","input":{"file_path":"` + path + `"}}]}}`
	}
	for _, tt := range []struct {
		name                      string
		lines                     []string
		bench, profiles, verdicts int
		measuredFirst             bool
	}{
		{"measured then edited", []string{bash("go test -bench . -cpuprofile cpu.out ./p"), edit("p/p.go"), bash("bash /x/bench-compare.sh --verdict /tmp/s ./p")}, 2, 1, 1, true},
		{"edited by eye", []string{edit("p/p.go"), bash("go test -bench . ./p")}, 1, 0, 0, false},
		{"a test file first is not an edit", []string{edit("p/p_bench_test.go"), bash("go tool pprof -top cpu.out")}, 0, 1, 0, true},
		{"nothing run", []string{bash("go build ./...")}, 0, 0, 0, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			bench, profiles, verdicts, first := perfSignals([]byte(strings.Join(tt.lines, "\n")))
			if bench != tt.bench || profiles != tt.profiles || verdicts != tt.verdicts || first != tt.measuredFirst {
				t.Errorf("perfSignals() = %d, %d, %d, %v; want %d, %d, %d, %v", bench, profiles, verdicts, first, tt.bench, tt.profiles, tt.verdicts, tt.measuredFirst)
			}
		})
	}
}

func TestPerfDecide(t *testing.T) {
	for _, tt := range []struct {
		name   string
		score  perfScore
		golden bool
		want   bool
	}{
		{"kept", perfScore{Expect: expectSpeedup, Verdict: perfKeep}, true, true},
		{"kept but behavior broke", perfScore{Expect: expectSpeedup, Verdict: perfKeep}, false, false},
		{"kept with a bait touched", perfScore{Expect: expectSpeedup, Verdict: perfKeep, BaitsChanged: []string{"formatLine"}}, true, false},
		{"not significant", perfScore{Expect: expectSpeedup, Verdict: "noisy"}, true, false},
		{"kept below the fixture's floor", perfScore{Expect: expectSpeedup, Verdict: perfKeep, MinSpeedupPct: 50, SecDeltaPct: new(-12.8)}, true, false},
		{"kept above the fixture's floor", perfScore{Expect: expectSpeedup, Verdict: perfKeep, MinSpeedupPct: 50, SecDeltaPct: new(-85.0)}, true, true},
		{"kept on allocations alone under a floor", perfScore{Expect: expectSpeedup, Verdict: perfKeep, MinSpeedupPct: 50}, true, false},
		{"declined", perfScore{Expect: expectDecline}, true, true},
		{"declined, fixed the real hot path", perfScore{Expect: expectDecline, Changed: []string{"discount"}}, true, true},
		{"optimized the cold function", perfScore{Expect: expectDecline, TargetsChanged: []string{"NormalizeName"}}, true, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.score.decide(tt.golden)
			if tt.score.Pass != tt.want {
				t.Errorf("decide(golden=%v) on %+v: Pass = %v, want %v", tt.golden, tt.score, tt.score.Pass, tt.want)
			}
		})
	}
}

// TestRunOnePerfMeasuresAfterTheSession drives runOne and measurePerfResults
// end to end: a fake claude edits the fixture, and the verdict comes from the
// repository's own bench-compare.sh on the golden benchmark. The versions
// differ by 62 allocations per op, so the verdict does not depend on timing.
func TestRunOnePerfMeasuresAfterTheSession(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and runs the golden benchmark through bench-compare.sh")
	}
	const slow = "package task\n\nfunc Line(n int) string {\n\ts := \"\"\n\tfor range n {\n\t\ts += \"x\"\n\t}\n\treturn s\n}\n\nfunc Label() string { return \"line\" }\n"
	const fast = "package task\n\nimport \"strings\"\n\nfunc Line(n int) string { return strings.Repeat(\"x\", n) }\n\nfunc Label() string { return \"line\" }\n"
	const baited = "package task\n\nimport \"strings\"\n\nfunc Line(n int) string { return strings.Repeat(\"x\", n) }\n\nfunc Label() string { return strings.Clone(\"line\") }\n"
	bin := t.TempDir()
	cli := filepath.Join(bin, "claude")
	writeTestFile(t, cli, `#!/bin/sh
set -eu
pwd > "$ABRUN_TEST_WORK"
printf '%s\n' '{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"go test -run ^$ -bench . ./task"}}]}}'
if [ -f "$ABRUN_TEST_SOURCE/task.go" ]; then
  printf '%s\n' '{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Edit","input":{"file_path":"task/task.go"}}]}}'
  cp "$ABRUN_TEST_SOURCE/task.go" task/task.go
fi
printf '%s\n' '{"type":"result","result":"done"}'
`)
	if err := os.Chmod(cli, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	root := filepath.Join("..", "..", "..")

	for _, tt := range []struct {
		name, source, expect string
		verdict              string
		pass, measured       bool
	}{
		{name: "faster", source: fast, expect: expectSpeedup, verdict: perfKeep, pass: true, measured: true},
		{name: "faster with the bait touched", source: baited, expect: expectSpeedup, verdict: perfKeep, measured: true},
		{name: "declined", expect: expectDecline, pass: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			corpus, model := t.TempDir(), t.TempDir()
			writeTestFile(t, filepath.Join(corpus, "task/task.go"), slow)
			writeTestFile(t, filepath.Join(corpus, "_golden/task/golden_test.go"), "package task\n\nimport \"testing\"\n\nfunc TestGoldenLine(t *testing.T) {\n\tif got := Line(3); got != \"xxx\" {\n\t\tt.Fatalf(\"Line(3) = %q\", got)\n\t}\n}\n")
			writeTestFile(t, filepath.Join(corpus, "_golden/task/golden_bench_test.go"), "package task\n\nimport \"testing\"\n\nvar goldenSink string\n\nfunc BenchmarkGoldenLine(b *testing.B) {\n\tfor b.Loop() {\n\t\tgoldenSink = Line(64)\n\t}\n}\n")
			writeTestFile(t, filepath.Join(corpus, "_golden/task/key.json"), `{"prompt":"Speed up Line in ./%s.","expect":"`+tt.expect+`","bench":"^BenchmarkGoldenLine$","benchtime":"300x","targets":["Line"],"baits":["Label"]}`)
			if tt.source != "" {
				writeTestFile(t, filepath.Join(model, "task.go"), tt.source)
			}
			workFile := filepath.Join(model, "work")
			t.Setenv("ABRUN_TEST_SOURCE", model)
			t.Setenv("ABRUN_TEST_WORK", workFile)
			o := options{corpus: corpusPerf, runner: runnerClaude, shell: shellGo, timeout: 2 * time.Minute}
			results := []result{runOne(o, corpus, arm{Name: "baseline"}, "task", 0)}
			got := &results[0]
			if got.Err != "" || got.Perf == nil || !got.Golden {
				t.Fatalf("runOne() = %+v, want a golden pass with a perf score", got)
			}
			if got.Perf.Prompt != "Speed up Line in ./task." {
				t.Errorf("Perf.Prompt = %q, want the key's prompt for the fixture", got.Perf.Prompt)
			}
			if (got.perfJob != nil) != tt.measured {
				t.Fatalf("perfJob = %v, want a measurement owed: %v", got.perfJob, tt.measured)
			}
			measurePerfResults(o, root, results)
			if got.Perf.Verdict != tt.verdict || got.Perf.Pass != tt.pass || !got.Perf.MeasuredFirst {
				t.Errorf("Perf = %+v, want verdict %q, pass %v, measured before the edit", got.Perf, tt.verdict, tt.pass)
			}
			path, err := os.ReadFile(workFile)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(strings.TrimSpace(string(path))); !os.IsNotExist(err) {
				t.Errorf("scratch tree still exists after the measurement: %v", err)
			}
		})
	}
}
