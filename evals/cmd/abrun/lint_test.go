package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCountLintIssues(t *testing.T) {
	const report = `{"Issues":[
		{"FromLinter":"errcheck","Text":"Error return value of ` + "`w.Write`" + ` is not checked","Pos":{"Filename":"task/task.go","Line":8}},
		{"FromLinter":"errcheck","Text":"Error return value of ` + "`os.Remove`" + ` is not checked","Pos":{"Filename":"task/task_test.go","Line":5}},
		{"FromLinter":"revive","Text":"exported: exported function Serve should have comment","Pos":{"Filename":"task/task.go","Line":7}}
	],"Report":{}}`
	const broken = `{"Issues":[{"FromLinter":"typecheck","Text":"undefined: missingHelper","Pos":{"Filename":"task/task.go","Line":3}}],"Report":{}}`

	tests := []struct {
		name       string
		in         string
		want       int
		unmeasured bool
	}{
		{name: "no issues", in: `{"Issues":null,"Report":{}}`, want: 0},
		// Both production findings count, the one in the test file does not.
		{name: "production only", in: report, want: 2},
		// A package that does not compile is a compiler report, not a clean one.
		{name: "typecheck", in: broken, unmeasured: true},
		{name: "not a report", in: "level=error msg=...", unmeasured: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := countLintIssues([]byte(tt.in))
			if tt.unmeasured {
				if err == nil {
					t.Fatalf("countLintIssues(%s) error = nil, want a reason the reading was not taken", tt.name)
				}
				return
			}
			if err != nil {
				t.Fatalf("countLintIssues(%s) error = %v, want nil", tt.name, err)
			}
			if got != tt.want {
				t.Errorf("countLintIssues(%s) = %d, want %d", tt.name, got, tt.want)
			}
		})
	}
}

func TestReadingColumn(t *testing.T) {
	one, two := 1, 2
	tests := []struct {
		name          string
		before, after *int
		note, want    string
	}{
		{name: "never attempted", want: ""},
		{name: "both", before: &one, after: &two, want: "1->2"},
		{name: "after unmeasured", before: &one, note: "after: broken", want: "1->n/a"},
		{name: "only a note", note: "before: no config", want: "n/a->n/a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := readingColumn(tt.before, tt.after, tt.note); got != tt.want {
				t.Errorf("readingColumn(%s) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

// TestRunOneCountsLintFindings pins the reading the column exists for: the
// unchecked w.Write that every skilled Opus 5 gateway session left behind on
// 2026-09-11 is a finding under the bundled configuration, a session that
// checks it is clean, and a tree that does not compile is unmeasured rather
// than clean. It drives golangci-lint for real, so it is skipped where the
// binary is absent and under -short.
func TestRunOneCountsLintFindings(t *testing.T) {
	if testing.Short() {
		t.Skip("runs golangci-lint")
	}
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		t.Skip("golangci-lint not on PATH")
	}
	config, err := filepath.Abs(filepath.Join("..", "..", "..", "skills", "go-linting", "assets", "golangci.yml"))
	if err != nil {
		t.Fatal(err)
	}

	const stub = `// Package task writes bodies.
package task

import "net/http"

// Serve writes body to w.
func Serve(w http.ResponseWriter, body []byte) {
	panic("not implemented")
}
`
	const unchecked = `// Package task writes bodies.
package task

import "net/http"

// Serve writes body to w.
func Serve(w http.ResponseWriter, body []byte) {
	w.Write(body)
}
`
	const checked = `// Package task writes bodies.
package task

import "net/http"

// Serve writes body to w.
func Serve(w http.ResponseWriter, body []byte) {
	if _, err := w.Write(body); err != nil {
		return
	}
}
`
	const broken = `// Package task writes bodies.
package task

import "net/http"

// Serve writes body to w.
func Serve(w http.ResponseWriter, body []byte) { missing(w, body) }
`

	for _, tt := range []struct {
		name, source string
		want         int
		unmeasured   bool
	}{
		{name: "unchecked write", source: unchecked, want: 1},
		{name: "checked write", source: checked, want: 0},
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
			writeTestFile(t, filepath.Join(corpus, "task/task.go"), stub)
			writeTestFile(t, filepath.Join(corpus, "_golden/task/golden_test.go"),
				`package task; import ("net/http/httptest"; "testing"); func TestServe(t *testing.T) { rec := httptest.NewRecorder(); Serve(rec, []byte("ok")); if rec.Body.String() != "ok" { t.Fatal("golden mismatch") } }`)
			writeTestFile(t, filepath.Join(model, "task.go"), tt.source)
			t.Setenv("ABRUN_TEST_SOURCE", model)

			got := runOne(options{runner: runnerClaude, prompt: "Implement %s", timeout: 2 * time.Minute, lintConfig: config}, corpus, arm{Name: "baseline"}, "task", 0)

			if got.LintBefore == nil || *got.LintBefore != 0 {
				t.Fatalf("runOne(%s) lint_before = %v, want 0 for the stub (%s)", tt.name, got.LintBefore, got.LintUnmeasured)
			}
			if tt.unmeasured {
				if got.Lint != nil || !strings.Contains(got.LintUnmeasured, "after:") {
					t.Fatalf("runOne(%s) lint = %v, unmeasured = %q, want no reading and a reason", tt.name, got.Lint, got.LintUnmeasured)
				}
				return
			}
			if got.Lint == nil || *got.Lint != tt.want {
				t.Fatalf("runOne(%s) lint = %v, want %d (%s)", tt.name, got.Lint, tt.want, got.LintUnmeasured)
			}
			if column := lintColumn(got); column != "0->"+strconv.Itoa(tt.want) {
				t.Errorf("lintColumn(%s) = %q, want %q", tt.name, column, "0->"+strconv.Itoa(tt.want))
			}
			summary := summarizeArm(report{Results: []result{got}}, "baseline")
			if summary.Valid != 1 || summary.LintMeasured != 1 || summary.LintAfter != tt.want {
				t.Errorf("summary valid=%d measured=%d after=%d, want 1, 1 and %d", summary.Valid, summary.LintMeasured, summary.LintAfter, tt.want)
			}
			if clean := summary.LintClean == 1; clean != (tt.want == 0) {
				t.Errorf("summary clean=%v, want %v", clean, tt.want == 0)
			}
		})
	}
}
