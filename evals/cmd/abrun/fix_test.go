package main

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestCountFixHunks(t *testing.T) {
	const diff = `--- /tmp/abrun/task/task.go (old)
+++ /tmp/abrun/task/task.go (new)
@@ -3,7 +3,7 @@
 func Sum(values []int) int {
 	total := 0
-	for i := 0; i < len(values); i++ {
+	for i := range values {
 		total += values[i]
 	}
@@ -20,3 +20,3 @@
-	banner := "--- old (old)"
+	banner := "--- new (old)"
--- /tmp/abrun/task/task_test.go (old)
+++ /tmp/abrun/task/task_test.go (new)
@@ -8,3 +8,3 @@
-	ctx := context.Background()
+	ctx := t.Context()
`

	tests := []struct {
		name string
		in   string
		want int
	}{
		{name: "empty diff", in: "", want: 0},
		// Both production hunks count, the one in the test file does not, and
		// the rewritten string literal does not open a new file section.
		{name: "production only", in: diff, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := countFixHunks([]byte(tt.in)); got != tt.want {
				t.Errorf("countFixHunks(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

// TestRunOneCountsPendingModernizations pins the property the metric exists
// for and the one that would quietly invert it: a tree the toolchain still has
// a modernizer for reads above zero, and a tree that cannot be read at all
// reads as unmeasured rather than as clean.
func TestRunOneCountsPendingModernizations(t *testing.T) {
	const stale = `package task

func Sum(values []int) int {
	total := 0
	for i := 0; i < len(values); i++ {
		total += values[i]
	}
	return total
}
`
	const modern = `package task

func Sum(values []int) int {
	total := 0
	for _, v := range values {
		total += v
	}
	return total
}
`
	// A comment is enough of an edit to keep the run valid while leaving the
	// loop go fix wants to rewrite exactly where it was.
	const commented = "// Package task sums values.\n" + stale
	const broken = `package task

func Sum(values []int) int { return missingHelper(values) }
`

	for _, tt := range []struct {
		name, source string
		want         int
		unmeasured   bool
	}{
		{name: "modernized", source: modern, want: 0},
		{name: "left alone", source: commented, want: 1},
		{name: "does not compile", source: broken, unmeasured: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			bin, corpus, model := t.TempDir(), t.TempDir(), t.TempDir()
			cli := filepath.Join(bin, "claude")
			writeTestFile(t, cli, `#!/bin/sh
set -eu
cp "$ABRUN_TEST_SOURCE/task.go" task/task.go
`)
			if err := os.Chmod(cli, 0o755); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			writeTestFile(t, filepath.Join(corpus, "task/task.go"), stale)
			writeTestFile(t, filepath.Join(corpus, "_golden/task/golden_test.go"),
				`package task; import "testing"; func TestSum(t *testing.T) { if Sum([]int{1,2,3})!=6 { t.Fatal("golden mismatch") } }`)
			writeTestFile(t, filepath.Join(model, "task.go"), tt.source)
			t.Setenv("ABRUN_TEST_SOURCE", model)

			got := runOne(options{runner: runnerClaude, prompt: "Refactor %s", timeout: time.Minute}, corpus, arm{Name: "baseline"}, "task", 0)

			if got.FixHunksBefore == nil || *got.FixHunksBefore != 1 {
				t.Fatalf("runOne(%s) fix_hunks_before = %v, want the fixture's one pending rewrite (%s)",
					tt.name, got.FixHunksBefore, got.FixUnmeasured)
			}
			if tt.unmeasured {
				if got.FixHunks != nil || got.FixUnmeasured == "" {
					t.Fatalf("runOne(%s) fix_hunks = %v, unmeasured = %q, want no reading and a reason",
						tt.name, got.FixHunks, got.FixUnmeasured)
				}
				return
			}
			if got.FixHunks == nil || *got.FixHunks != tt.want {
				t.Fatalf("runOne(%s) fix_hunks = %v, want %d (%s)", tt.name, got.FixHunks, tt.want, got.FixUnmeasured)
			}
			if column := fixColumn(got); column != "1->"+strconv.Itoa(tt.want) {
				t.Errorf("fixColumn(%s) = %q, want %q", tt.name, column, "1->"+strconv.Itoa(tt.want))
			}
			summary := summarizeArm(report{Results: []result{got}}, "baseline")
			if summary.Valid != 1 || summary.FixMeasured != 1 || summary.FixAfter != tt.want {
				t.Errorf("summary valid=%d measured=%d after=%d, want 1, 1 and %d",
					summary.Valid, summary.FixMeasured, summary.FixAfter, tt.want)
			}
			if clean := summary.FixClean == 1; clean != (tt.want == 0) {
				t.Errorf("summary clean=%v, want %v", clean, tt.want == 0)
			}
		})
	}
}
