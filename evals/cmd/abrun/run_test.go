package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

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
