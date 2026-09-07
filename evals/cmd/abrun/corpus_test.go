package main

import (
	"path/filepath"
	"slices"
	"testing"
	"time"
)

// corpusRoots locates the two fixture roots from the test's own package, so the
// check does not depend on a git checkout being present.
func corpusRoots() (refactor, implement string) {
	refactor = filepath.Join("..", "..", "ab")
	return refactor, filepath.Join(refactor, implementDir)
}

// TestCorporaDoNotOverlap pins the property the underscore prefix buys: the
// implementation fixtures are invisible to a refactor run, so the two line
// deltas can never end up in one mean.
func TestCorporaDoNotOverlap(t *testing.T) {
	refactorRoot, implementRoot := corpusRoots()

	refactorTasks, err := findTasks(refactorRoot, "")
	if err != nil {
		t.Fatalf("findTasks(refactor corpus) error = %v, want nil", err)
	}
	if slices.Contains(refactorTasks, implementDir) {
		t.Errorf("findTasks(refactor corpus) = %v, want %s excluded", refactorTasks, implementDir)
	}

	implementTasks, err := findTasks(implementRoot, "")
	if err != nil {
		t.Fatalf("findTasks(implement corpus) error = %v, want nil", err)
	}
	if len(implementTasks) == 0 {
		t.Fatal("findTasks(implement corpus) returned no fixtures")
	}
	for _, task := range implementTasks {
		if slices.Contains(refactorTasks, task) {
			t.Errorf("fixture %q appears in both corpora; golden directories are keyed by name", task)
		}
	}
}

// TestEveryFixtureIsRunnable is the regression test for adding a fixture: a new
// directory without a golden test would otherwise only fail once a paid run had
// already started.
func TestEveryFixtureIsRunnable(t *testing.T) {
	refactorRoot, implementRoot := corpusRoots()

	for _, root := range []string{refactorRoot, implementRoot} {
		tasks, err := findTasks(root, "")
		if err != nil {
			t.Fatalf("findTasks(%q) error = %v, want nil", root, err)
		}
		if err := validateFixtures(root, tasks); err != nil {
			t.Errorf("validateFixtures(%q, %v) error = %v, want nil", root, tasks, err)
		}
	}
}

func TestCorpusPrompt(t *testing.T) {
	if got := corpusPrompt(corpusRefactor); got != refactorPrompt {
		t.Errorf("corpusPrompt(%q) = %q, want the refactor prompt", corpusRefactor, got)
	}
	if got := corpusPrompt(corpusImplement); got != implementPrompt {
		t.Errorf("corpusPrompt(%q) = %q, want the implement prompt", corpusImplement, got)
	}
}

func TestValidateOptionsRejectsUnknownCorpus(t *testing.T) {
	o := options{corpus: "rewrite", runner: runnerClaude, reps: 1, parallel: 1, timeout: time.Second}

	if err := validateOptions(o); err == nil {
		t.Error("validateOptions(unknown corpus) error = nil, want non-nil")
	}
}

func TestSkillFired(t *testing.T) {
	tests := []struct {
		name   string
		corpus string
		skills []string
		want   bool
	}{
		{name: "refactor owner fired", corpus: corpusRefactor, skills: []string{"go-code-refactor"}, want: true},
		{name: "refactor with another skill only", corpus: corpusRefactor, skills: []string{"go-naming"}, want: false},
		{name: "implement routes to its own owner", corpus: corpusImplement, skills: []string{"go-http"}, want: true},
		{name: "implement reached nothing", corpus: corpusImplement, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := skillFired(tt.corpus, tt.skills); got != tt.want {
				t.Errorf("skillFired(%q, %v) = %v, want %v", tt.corpus, tt.skills, got, tt.want)
			}
		})
	}
}
