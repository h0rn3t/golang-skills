package main

import (
	"path/filepath"
	"testing"
)

// TestAnalyzeCountsNamedClosures pins what the column sees and what it leaves
// alone: a helper bound to a name inside a function counts, whether by :=, =
// or var, a literal handed straight to HandleFunc does not, and a package-level
// var func is a declaration Funcs and Exported already account for.
func TestAnalyzeCountsNamedClosures(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "task.go"), `package task

import "net/http"

var render = func(v any) []byte { return nil }

func NewServer() http.Handler {
	mux := http.NewServeMux()
	writeJSON := func(w http.ResponseWriter, v any) { w.Write(render(v)) }
	var methodNotAllowed func(w http.ResponseWriter, r *http.Request)
	methodNotAllowed = func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(405) }
	mux.HandleFunc("GET /x", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 1) })
	mux.HandleFunc("/x", methodNotAllowed)
	return mux
}
`)
	writeTestFile(t, filepath.Join(dir, "task_test.go"), `package task

func helper() { f := func() {}; f() }
`)

	got, err := analyze(dir)
	if err != nil {
		t.Fatalf("analyze() error = %v, want nil", err)
	}
	if got.Closures != 2 {
		t.Errorf("analyze().Closures = %d, want 2 (writeJSON and methodNotAllowed; not the HandleFunc literal, the package var, or the test helper)", got.Closures)
	}
	if got.Funcs != 1 {
		t.Errorf("analyze().Funcs = %d, want 1", got.Funcs)
	}
}

// TestAnalyzeCountsBodyComments pins the boundary of the column: narration
// inside a body counts, whether above a statement, trailing it, or inside a
// function literal; the doc comment on the declaration does not.
func TestAnalyzeCountsBodyComments(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "task.go"), `// Package task serves.
package task

import "net/http"

// NewServer returns the server. This doc comment is the specification and is
// not counted.
func NewServer() http.Handler {
	mux := http.NewServeMux()
	// The patterns carry no method: a known path answers every method but GET
	// with 405.
	mux.HandleFunc("/x", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200) // ok
	})
	return mux
}
`)

	got, err := analyze(dir)
	if err != nil {
		t.Fatalf("analyze() error = %v, want nil", err)
	}
	if got.BodyComments != 3 {
		t.Errorf("analyze().BodyComments = %d, want 3 (two lines above HandleFunc, one trailing inside the literal)", got.BodyComments)
	}
}
