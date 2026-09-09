package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRepairFeedback pins the two things the loop's text has to carry: the
// harness's own counting convention, because stage 2 measured a session that
// satisfied its own gate under a different one, and assertion text without the
// position that would name the hidden test.
func TestRepairFeedback(t *testing.T) {
	grown := repairFeedback("gw0", 132, 148, "")
	for _, want := range []string{
		"Package under review: ./gw0",
		"Starting production LOC: 132",
		"Current production LOC: 148",
		"Gate: FAIL, growth +16",
		"physical lines",
		"Done when LOC <= 132",
	} {
		if !strings.Contains(grown, want) {
			t.Errorf("repairFeedback(grown) missing %q:\n%s", want, grown)
		}
	}
	if strings.Contains(grown, "Independent contract failure") {
		t.Errorf("repairFeedback(no failure) invented one:\n%s", grown)
	}

	broke := repairFeedback("gw0", 135, 130, "go test ./gw0/...: exit status 1: --- FAIL: TestEmptyListIsJSONArray/nil (0.00s)\n"+
		"        golden_test.go:200: GET /accounts = 200 \"null\", want 200 []\nFAIL\nFAIL\tabeval/gw0\t0.2s")
	if !strings.Contains(broke, "Gate: PASS on line count.") {
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
			bin, corpus, model := t.TempDir(), t.TempDir(), t.TempDir()
			cli := filepath.Join(bin, "claude")
			// The turn counter lives in the model directory: turn one writes the
			// grown package, turn two the repaired one.
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
			writeTestFile(t, filepath.Join(model, "turn1.go"), grown)
			writeTestFile(t, filepath.Join(model, "turn2.go"), repaired)
			workFile := filepath.Join(model, "work")
			t.Setenv("ABRUN_TEST_SOURCE", model)
			t.Setenv("ABRUN_TEST_WORK", workFile)

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
			if !strings.Contains(got.RepairFeedback, "Gate: FAIL, growth +3") {
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
