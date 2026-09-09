package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRunOneSeparatesHarnessCollisionFromRegression is the end-to-end form of
// the fault the 5×5 concision run hit: a production helper sharing a name with
// the hidden overlay's own helper failed the golden run, and the failure was
// first recorded as a behavior regression. The two causes now have to come out
// of runOne as different results, and the transcript has to survive so a claim
// in the final message is checkable.
func TestRunOneSeparatesHarnessCollisionFromRegression(t *testing.T) {
	const goldenTest = `package task

import "testing"

func goldenValue() int { return 1 }

func TestValue(t *testing.T) {
	if got := Value(); got != goldenValue() {
		t.Fatalf("Value() = %d, want %d", got, goldenValue())
	}
}
`
	const fixture = `package task

func Value() int { return 1 }
`
	// collision keeps the fixture's behavior and adds a helper the overlay also
	// declares. It compiles on its own, which is what makes the golden failure
	// the harness's fault and not the model's.
	const collision = `package task

func goldenValue() int { return 1 }

func Value() int { return goldenValue() }
`
	const regression = `package task

func Value() int { return 2 }
`

	for _, tt := range []struct {
		name              string
		source            string
		harness, behavior bool
		edited, gate      bool
		status            string
	}{
		{name: "name collision", source: collision, harness: true, edited: true, status: "HRN"},
		{name: "real regression", source: regression, behavior: true, edited: true, gate: true, status: "ERR"},
		{name: "empty diff", gate: true, status: "ERR"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			bin, corpus, model := t.TempDir(), t.TempDir(), t.TempDir()
			cli := filepath.Join(bin, "claude")
			writeTestFile(t, cli, `#!/bin/sh
set -eu
pwd > "$ABRUN_TEST_WORK"
if [ -f "$ABRUN_TEST_SOURCE/task.go" ]; then
  cp "$ABRUN_TEST_SOURCE/task.go" task/task.go
fi
printf '%s\n' '{"type":"result","result":"Rewrote the helper; 3 lines now."}'
`)
			if err := os.Chmod(cli, 0o755); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			writeTestFile(t, filepath.Join(corpus, "task/task.go"), fixture)
			writeTestFile(t, filepath.Join(corpus, "_golden/task/golden_test.go"), goldenTest)
			if tt.source != "" {
				writeTestFile(t, filepath.Join(model, "task.go"), tt.source)
			}
			workFile := filepath.Join(model, "work")
			t.Setenv("ABRUN_TEST_SOURCE", model)
			t.Setenv("ABRUN_TEST_WORK", workFile)

			got := runOne(options{runner: runnerClaude, prompt: "Refactor %s", keep: true, timeout: time.Minute}, corpus, arm{Name: "baseline"}, "task", 0)
			t.Cleanup(func() { _ = os.RemoveAll(got.WorkDir) })

			if !got.Build {
				t.Fatalf("runOne(%s) build = false, want the package to compile: %s", tt.name, got.GoFail)
			}
			if got.HarnessFailure != tt.harness || got.BehaviorFailure != tt.behavior {
				t.Errorf("runOne(%s) harness=%v behavior=%v, want %v and %v: %s",
					tt.name, got.HarnessFailure, got.BehaviorFailure, tt.harness, tt.behavior, firstLine(got.GoFail))
			}
			if got.Edited != tt.edited || got.EmptyDiff != !tt.edited {
				t.Errorf("runOne(%s) edited=%v empty_diff=%v, want %v and %v", tt.name, got.Edited, got.EmptyDiff, tt.edited, !tt.edited)
			}
			if got.LineGatePass != tt.gate {
				t.Errorf("runOne(%s) line_gate_pass=%v (Δlines %+d), want %v", tt.name, got.LineGatePass, got.Delta.Lines, tt.gate)
			}
			if status := resultStatus(got); status != tt.status {
				t.Errorf("resultStatus(%s) = %q, want %q", tt.name, status, tt.status)
			}
			if !got.ReportedCounts {
				t.Errorf("runOne(%s) reported_counts = false, want the claim in the final message counted", tt.name)
			}

			// The trace is what separates a measured count from a stated one.
			if got.Trace == "" {
				t.Fatalf("runOne(%s) kept no trace", tt.name)
			}
			trace, err := os.ReadFile(got.Trace)
			if err != nil || !strings.Contains(string(trace), `"result"`) {
				t.Errorf("trace = %q, err = %v, want the session transcript", trace, err)
			}

			summary := summarizeArm(report{Results: []result{got}}, "baseline")
			if summary.HarnessFailures != boolCount(tt.harness) || summary.BehaviorFailures != boolCount(tt.behavior) {
				t.Errorf("summary harness=%d behavior=%d, want %d and %d",
					summary.HarnessFailures, summary.BehaviorFailures, boolCount(tt.harness), boolCount(tt.behavior))
			}
			if summary.Valid != 0 {
				t.Errorf("summary valid = %d, want 0: no run here yields a usable behavioral verdict", summary.Valid)
			}
		})
	}
}

func boolCount(b bool) int {
	if b {
		return 1
	}
	return 0
}

func TestRunOneValidatesModelTestsAndKeepsSource(t *testing.T) {
	bin := t.TempDir()
	cli := filepath.Join(bin, "claude")
	writeTestFile(t, cli, `#!/bin/sh
set -eu
pwd > "$ABRUN_TEST_WORK"
cp "$ABRUN_TEST_SOURCE/task.go" task/task.go
if [ -f "$ABRUN_TEST_SOURCE/task_test.go" ]; then
  cp "$ABRUN_TEST_SOURCE/task_test.go" task/task_test.go
fi
`)
	if err := os.Chmod(cli, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	const implementation = "package task\nfunc Value() int { return 1 }\n"
	for _, tt := range []struct {
		name, test  string
		keep, valid bool
	}{
		{name: "passing", test: `package task; import "testing"; func TestValue(t *testing.T) { if Value()!=1 { t.Fatal("wrong value") } }`, keep: true, valid: true},
		{name: "failing", test: `package task; import "testing"; func TestValue(t *testing.T) { t.Fatal("model assertion failed") }`, keep: true},
		{name: "uncompilable", test: `package task; import "testing"; func TestValue(t *testing.T) { missing() }`, keep: true},
		{name: "absent", keep: true, valid: true},
		{name: "cleanup", valid: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			corpus, model := t.TempDir(), t.TempDir()
			writeTestFile(t, filepath.Join(corpus, "task/task.go"), "package task\nfunc Value() int { return 0 }\n")
			writeTestFile(t, filepath.Join(corpus, "_golden/task/golden_test.go"), `package task; import "testing"; func TestValue(t *testing.T) { if Value()!=1 { t.Fatal("golden mismatch") } }`)
			writeTestFile(t, filepath.Join(model, "task.go"), implementation)
			if tt.test != "" {
				writeTestFile(t, filepath.Join(model, "task_test.go"), tt.test)
			}
			workFile := filepath.Join(model, "work")
			t.Setenv("ABRUN_TEST_SOURCE", model)
			t.Setenv("ABRUN_TEST_WORK", workFile)
			got := runOne(options{runner: runnerClaude, prompt: "Implement %s", keep: tt.keep, timeout: time.Minute}, corpus, arm{Name: "baseline"}, "task", 0)
			path, err := os.ReadFile(workFile)
			if err != nil {
				t.Fatal(err)
			}
			work := strings.TrimSpace(string(path))
			t.Cleanup(func() { _ = os.RemoveAll(work) })
			if got.Err != "" || !got.Build || !got.Golden || !got.Edited {
				t.Fatalf("runOne(%s) = %+v, want edited, built and golden pass", tt.name, got)
			}
			if valid := resultStatus(got) == "ok "; valid != tt.valid {
				t.Errorf("runOne(%s) valid=%v, want %v", tt.name, valid, tt.valid)
			}
			wantTests := "pass"
			if tt.test == "" {
				wantTests = "skipped"
			} else if !tt.valid {
				wantTests = "fail"
			}
			if got.ModelTests != wantTests || (got.ModelTestFail != "") != !tt.valid {
				t.Errorf("model tests=%q failure=%q, want %q with failure=%v", got.ModelTests, got.ModelTestFail, wantTests, !tt.valid)
			}
			summary := summarizeArm(report{Results: []result{got}}, "baseline")
			if valid := summary.Valid == 1; valid != tt.valid {
				t.Errorf("summary valid=%v, want %v", valid, tt.valid)
			}
			if failed := summary.ModelTestFailures == 1; failed != !tt.valid {
				t.Errorf("summary model-test failure=%v, want %v", failed, !tt.valid)
			}
			if !tt.keep {
				if _, err := os.Stat(work); !os.IsNotExist(err) {
					t.Errorf("scratch directory still exists: %v", err)
				}
				return
			}
			if got.WorkDir == "" {
				t.Fatal("-keep lost scratch directory for a completed run")
			}
			data, err := os.ReadFile(filepath.Join(got.WorkDir, "task/task.go"))
			if err != nil || string(data) != implementation {
				t.Errorf("retained source=%q, err=%v, want original model source", data, err)
			}
			if tt.test != "" {
				data, err := os.ReadFile(filepath.Join(got.WorkDir, "task/task_test.go.model"))
				if err != nil || string(data) != tt.test {
					t.Errorf("retained test=%q, err=%v, want original model test", data, err)
				}
			}
		})
	}
}
