package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRepairFeedback pins the three things the loop's text has to carry: the
// harness's own counting convention, because stage 2 measured a session that
// satisfied its own gate under a different one; both counts and the rule that
// one cannot pay for the other, because stage 3 measured two sessions that met
// the physical gate by deleting a doc comment; and assertion text without the
// position that would name the hidden test.
func TestRepairFeedback(t *testing.T) {
	grown := repairFeedback("gw0", metrics{Lines: 132, Code: 96}, metrics{Lines: 148, Code: 104}, "")
	for _, want := range []string{
		"Package under review: ./gw0",
		"Starting production LOC: 132 physical, 96 code",
		"Current production LOC: 148 physical, 104 code",
		"Gate: FAIL, physical +16, code +8",
		"Physical LOC counts every line",
		"deleting a comment does not pay for a line of code",
		"Done when physical LOC <= 132, code LOC <= 96",
	} {
		if !strings.Contains(grown, want) {
			t.Errorf("repairFeedback(grown) missing %q:\n%s", want, grown)
		}
	}
	if strings.Contains(grown, "Independent contract failure") {
		t.Errorf("repairFeedback(no failure) invented one:\n%s", grown)
	}

	// The stage 3 hole: fewer physical lines, more code. The verdict has to fail.
	traded := repairFeedback("gw2", metrics{Lines: 138, Code: 101}, metrics{Lines: 137, Code: 107}, "")
	if !strings.Contains(traded, "Gate: FAIL, physical -1, code +6") {
		t.Errorf("repairFeedback(doc traded for code) did not fail the gate:\n%s", traded)
	}

	broke := repairFeedback("gw0", metrics{Lines: 135, Code: 99}, metrics{Lines: 130, Code: 95},
		"go test ./gw0/...: exit status 1: --- FAIL: TestEmptyListIsJSONArray/nil (0.00s)\n"+
			"        golden_test.go:200: GET /accounts = 200 \"null\", want 200 []\nFAIL\nFAIL\tabeval/gw0\t0.2s")
	if !strings.Contains(broke, "Gate: PASS, physical -5, code -4") {
		t.Errorf("repairFeedback(shrunk) reported growth:\n%s", broke)
	}
	if !strings.Contains(broke, `GET /accounts = 200 "null", want 200 []`) {
		t.Errorf("repairFeedback(broke) lost the assertion:\n%s", broke)
	}
	if strings.Contains(broke, "golden_test.go") {
		t.Errorf("repairFeedback(broke) named the hidden test file:\n%s", broke)
	}
}

// TestProbeGoldenLeavesTreeUntouched is the safety property the loop rests on.
// The probe has to answer whether the golden passes without putting the golden,
// or the renamed model tests, into the tree the next turn can read.
func TestProbeGoldenLeavesTreeUntouched(t *testing.T) {
	work, golden := t.TempDir(), t.TempDir()
	writeTestFile(t, filepath.Join(work, "go.mod"), "module abeval\n\ngo 1.27\n")
	writeTestFile(t, filepath.Join(work, "task/task.go"), "package task\n\nfunc Value() int { return 2 }\n")
	writeTestFile(t, filepath.Join(work, "task/task_test.go"),
		"package task\n\nimport \"testing\"\n\nfunc TestModel(t *testing.T) {}\n")
	writeTestFile(t, filepath.Join(golden, "golden_test.go"),
		"package task\n\nimport \"testing\"\n\nfunc TestGolden(t *testing.T) {\n\tif Value() != 1 {\n\t\tt.Fatalf(\"Value() = %d, want 1\", Value())\n\t}\n}\n")

	pass, failure, err := probeGolden(time.Minute, work, "task", golden)
	if err != nil {
		t.Fatalf("probeGolden error = %v, want nil", err)
	}
	if pass {
		t.Error("probeGolden = true, want the golden assertion to fail")
	}
	if !strings.Contains(failure, "Value() = 2, want 1") {
		t.Errorf("probeGolden failure = %q, want the assertion", failure)
	}

	entries, err := os.ReadDir(filepath.Join(work, "task"))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if len(names) != 2 || names[0] != "task.go" || names[1] != "task_test.go" {
		t.Errorf("probe left %v in the model's tree, want only task.go and task_test.go", names)
	}

	// And it answers true when the package satisfies the golden.
	writeTestFile(t, filepath.Join(work, "task/task.go"), "package task\n\nfunc Value() int { return 1 }\n")
	if pass, _, err := probeGolden(time.Minute, work, "task", golden); err != nil || !pass {
		t.Errorf("probeGolden(passing) = %v, err = %v, want true and nil", pass, err)
	}
}

// twoTurnCorpus builds a one-fixture corpus and puts a fake agent CLI on PATH
// that writes turn on its first invocation and repaired on its second, so a
// run with -repair can be driven end to end. It returns the corpus directory.
func twoTurnCorpus(t *testing.T, fixture, turn, repaired string) string {
	t.Helper()
	bin, corpus, model := t.TempDir(), t.TempDir(), t.TempDir()
	cli := filepath.Join(bin, "claude")
	// The turn counter lives in the model directory: turn one writes the first
	// package, turn two the repaired one.
	writeTestFile(t, cli, `#!/bin/sh
set -eu
pwd > "$ABRUN_TEST_WORK"
turn=1
if [ -f "$ABRUN_TEST_SOURCE/turn" ]; then turn=2; fi
: > "$ABRUN_TEST_SOURCE/turn"
cp "$ABRUN_TEST_SOURCE/turn$turn.go" task/task.go
printf '%s\n' "{\"type\":\"result\",\"result\":\"turn $turn done\"}"
`)
	if err := os.Chmod(cli, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	writeTestFile(t, filepath.Join(corpus, "task/task.go"), fixture)
	writeTestFile(t, filepath.Join(corpus, "_golden/task/golden_test.go"),
		"package task\n\nimport \"testing\"\n\nfunc TestGolden(t *testing.T) {\n\tif Value() != 1 {\n\t\tt.Fatal(\"golden mismatch\")\n\t}\n}\n")
	writeTestFile(t, filepath.Join(model, "turn1.go"), turn)
	writeTestFile(t, filepath.Join(model, "turn2.go"), repaired)
	t.Setenv("ABRUN_TEST_SOURCE", model)
	t.Setenv("ABRUN_TEST_WORK", filepath.Join(model, "work"))
	return corpus
}

// TestRunOneRepairLoop drives the whole loop through runOne with a fake CLI that
// grows the package on its first turn and shrinks it on the second, so the
// recorded result has to describe the tree after the repair, not before it.
func TestRunOneRepairLoop(t *testing.T) {
	const fixture = "package task\n\nfunc Value() int { return 1 }\n"
	const grown = "package task\n\n// grown past the gate\nfunc helper() int { return 1 }\n\nfunc Value() int { return helper() }\n"
	const repaired = "package task\n\nfunc Value() int { return 1 } // repaired\n"

	for _, tt := range []struct {
		name        string
		repair      bool
		wantFired   bool
		wantGate    bool
		wantLines   int
		wantPreGold bool
	}{
		{name: "loop off leaves the growth", wantLines: 3},
		{name: "loop on repairs it", repair: true, wantFired: true, wantGate: true, wantLines: 0, wantPreGold: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			corpus := twoTurnCorpus(t, fixture, grown, repaired)

			got := runOne(options{runner: runnerClaude, prompt: "Refactor %s", keep: true,
				repair: tt.repair, timeout: time.Minute}, corpus, arm{Name: "baseline"}, "task", 0)
			t.Cleanup(func() { _ = os.RemoveAll(got.WorkDir) })

			if got.Err != "" || !got.Build || !got.Golden {
				t.Fatalf("runOne(%s) = %+v, want a clean build and golden pass", tt.name, got)
			}
			if got.RepairFired != tt.wantFired {
				t.Errorf("repair_fired = %v, want %v", got.RepairFired, tt.wantFired)
			}
			if got.Delta.Lines != tt.wantLines {
				t.Errorf("Δlines = %+d, want %+d", got.Delta.Lines, tt.wantLines)
			}
			if got.LineGatePass != tt.wantGate {
				t.Errorf("line_gate_pass = %v, want %v", got.LineGatePass, tt.wantGate)
			}
			if !tt.wantFired {
				if got.PreRepair != nil || got.RepairFeedback != "" {
					t.Errorf("loop off still recorded pre-repair state: %+v %q", got.PreRepair, got.RepairFeedback)
				}
				return
			}
			if got.PreRepair == nil || got.PreRepair.Lines-got.Before.Lines != 3 {
				t.Errorf("pre_repair = %+v, want the +3 reading that triggered the loop", got.PreRepair)
			}
			if got.PreRepairGolden != tt.wantPreGold {
				t.Errorf("pre_repair_golden = %v, want %v", got.PreRepairGolden, tt.wantPreGold)
			}
			if !strings.Contains(got.RepairFeedback, "physical +3") {
				t.Errorf("repair feedback did not carry the measured growth:\n%s", got.RepairFeedback)
			}
			// Both turns have to survive in one trace.
			trace, err := os.ReadFile(got.Trace)
			if err != nil || !strings.Contains(string(trace), "turn 1 done") || !strings.Contains(string(trace), "turn 2 done") {
				t.Errorf("trace = %q, err = %v, want both turns", trace, err)
			}
			if got.Output != "turn 2 done" {
				t.Errorf("final message = %q, want the repair turn's", got.Output)
			}
			summary := summarizeArm(report{Results: []result{got}}, "baseline")
			if summary.RepairsFired != 1 || summary.RepairsRescued != 1 {
				t.Errorf("summary fired=%d rescued=%d, want 1 and 1", summary.RepairsFired, summary.RepairsRescued)
			}
		})
	}
}

// TestRunOneCodeGateBlocksDocsPayingForCode is the stage 3 hole in miniature.
// Two of that run's fifteen sessions met the gate by deleting the doc comment
// that justifies the server's timeouts and spending the lines on helpers: the
// file got physically shorter while the code in it grew. The physical gate
// passes that trade and has to keep passing it, because past reports are read
// against it; the code gate is what refuses it, and the loop has to fire.
func TestRunOneCodeGateBlocksDocsPayingForCode(t *testing.T) {
	// Nine physical lines, two of code.
	const fixture = `package task

// Value reports the answer.
//
// The paragraph below is the justification a session can spend: it explains
// why the answer is what it is, which is the kind of comment that costs
// physical lines without costing code. Deleting it buys room for helpers
// under a gate that counts every line the same.
func Value() int { return 1 }
`
	// Eight physical lines, five of code: shorter file, more code.
	const traded = `package task

// Value reports the answer.
func Value() int { return helper() }

func helper() int {
	return 1
}
`
	// Nine physical, two of code: the documentation is back and the code is not.
	const repaired = `package task

// Value reports the answer.
//
// The paragraph below is the justification a session can spend: it explains
// why the answer is what it is, which is the kind of comment that costs
// physical lines without costing code. Deleting it buys room for helpers
// under a gate that counts every line the same.
func Value() int { return 1 } // restored
`

	for _, tt := range []struct {
		name         string
		repair       bool
		wantFired    bool
		wantCodeGate bool
		wantLines    int
		wantCode     int
	}{
		{name: "loop off records the trade", wantLines: -1, wantCode: 3},
		{name: "loop on refuses it", repair: true, wantFired: true, wantCodeGate: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			corpus := twoTurnCorpus(t, fixture, traded, repaired)

			got := runOne(options{runner: runnerClaude, prompt: "Refactor %s", keep: true,
				repair: tt.repair, timeout: time.Minute}, corpus, arm{Name: "baseline"}, "task", 0)
			t.Cleanup(func() { _ = os.RemoveAll(got.WorkDir) })

			if got.Err != "" || !got.Build || !got.Golden {
				t.Fatalf("runOne(%s) = %+v, want a clean build and golden pass", tt.name, got)
			}
			if got.Delta.Lines != tt.wantLines || got.Delta.Code != tt.wantCode {
				t.Errorf("Δlines = %+d, Δcode = %+d, want %+d and %+d",
					got.Delta.Lines, got.Delta.Code, tt.wantLines, tt.wantCode)
			}
			// The physical gate passes either way; only the code gate separates them.
			if !got.LineGatePass {
				t.Errorf("line_gate_pass = false (Δlines %+d), want the physical gate to keep its meaning", got.Delta.Lines)
			}
			if got.CodeGatePass != tt.wantCodeGate {
				t.Errorf("code_gate_pass = %v (Δcode %+d), want %v", got.CodeGatePass, got.Delta.Code, tt.wantCodeGate)
			}
			if got.RepairFired != tt.wantFired {
				t.Errorf("repair_fired = %v, want %v — the trade is what the loop is for", got.RepairFired, tt.wantFired)
			}
			if !tt.wantFired {
				return
			}
			if got.PreRepairGolden != true {
				t.Error("pre_repair_golden = false, want the loop fired on the code gate alone")
			}
			if !strings.Contains(got.RepairFeedback, "Gate: FAIL, physical -1, code +3") {
				t.Errorf("repair feedback did not name the trade:\n%s", got.RepairFeedback)
			}
			summary := summarizeArm(report{Results: []result{got}}, "baseline")
			if summary.CodeGatePasses != 1 || summary.RepairsRescued != 1 {
				t.Errorf("summary code gate=%d rescued=%d, want 1 and 1", summary.CodeGatePasses, summary.RepairsRescued)
			}
		})
	}
}
