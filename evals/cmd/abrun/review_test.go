package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

func reviewRoot() string { return filepath.Join("..", "..", "ab", reviewDir) }

// TestReviewKeysResolve is the regression test for editing a fixture: every
// key entry must still land on exactly one line, and no bait may sit inside a
// defect's window, or a citation of the defect would count as a false alarm.
func TestReviewKeysResolve(t *testing.T) {
	root := reviewRoot()
	tasks, err := findTasks(root, "")
	if err != nil {
		t.Fatalf("findTasks(review corpus) error = %v", err)
	}
	if len(tasks) == 0 {
		t.Fatal("review corpus has no fixtures")
	}
	if err := validateReviewFixtures(root, tasks); err != nil {
		t.Fatalf("validateReviewFixtures() error = %v", err)
	}
	for _, task := range tasks {
		key, err := loadReviewKey(filepath.Join(root, "_golden", task, reviewKeyFile), filepath.Join(root, task))
		if err != nil {
			t.Fatalf("%s: %v", task, err)
		}
		for _, b := range key.Baits {
			for _, d := range key.Defects {
				c := cited{file: b.File, from: b.line, to: b.line}
				if c.near(d.File, d.line, d.Span) {
					t.Errorf("%s: bait %s (line %d) is within the window of defect %s (line %d, span %d)", task, b.ID, b.line, d.ID, d.line, d.Span)
				}
			}
		}
		for i, d := range key.Defects {
			for _, e := range key.Defects[i+1:] {
				c := cited{file: d.File, from: d.line, to: d.line + d.Span - 1}
				if c.near(e.File, e.line, e.Span) {
					t.Errorf("%s: defects %s (line %d) and %s (line %d) overlap within the tolerance; one citation would count for both", task, d.ID, d.line, e.ID, e.line)
				}
			}
		}
	}
}

// TestReviewFixturesAreToolClean pins the property the corpus README states:
// gofmt, go vet, and go fix have nothing to say about a fixture, so the tools
// a reviewer runs first do not hand over the defects that are meant to be
// found by reading, and the fixture's own tests pass.
func TestReviewFixturesAreToolClean(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not on PATH")
	}
	root := reviewRoot()
	tasks, err := findTasks(root, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, task := range tasks {
		t.Run(task, func(t *testing.T) {
			work := t.TempDir()
			if err := os.CopyFS(filepath.Join(work, task), os.DirFS(filepath.Join(root, task))); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(work, "go.mod"), []byte("module abeval\n\ngo 1.27\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			out, err := exec.Command("gofmt", "-l", filepath.Join(work, task)).CombinedOutput()
			if err != nil || len(strings.TrimSpace(string(out))) > 0 {
				t.Errorf("gofmt -l: %v %s", err, out)
			}
			for _, args := range [][]string{{"vet", "./..."}, {"test", "-count=1", "./..."}} {
				cmd := exec.Command("go", args...)
				cmd.Dir = work
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Errorf("go %s: %v\n%s", strings.Join(args, " "), err, out)
				}
			}
			if hunks, err := fixHunks(2*time.Minute, work, task); err != nil {
				t.Errorf("go fix -diff: %v", err)
			} else if hunks != 0 {
				t.Errorf("go fix -diff proposes %d hunk(s); the fixture must be modern before it is reviewed", hunks)
			}
		})
	}
}

func TestScoreReview(t *testing.T) {
	key := reviewKey{
		Defects: []reviewDefect{
			{ID: "a", Severity: "must", File: "server.go", line: 41, Span: 1},
			{ID: "b", Severity: "must", File: "store.go", line: 50, Span: 8, Tool: "rowserrcheck"},
			{ID: "c", Severity: "should", File: "server.go", line: 74, Span: 1},
			{ID: "d", Severity: "nit", File: "store.go", line: 10, Span: 1},
		},
		Baits: []reviewBait{{ID: "bait", File: "server.go", line: 90}},
	}
	review := `# Review

### Must Fix
- [ ] server.go:41 compares a wrapped error with ==
      Evidence: plausible
- [ ] ./orders/store.go:55:4 rows never closed (verified by reading the loop)

### Should Fix
- [ ] server.go:72-73 the goroutine outlives the request

### Nits
- [ ] server.go:90 consider checking the write error
- [ ] server.go:120 naming
`
	got := scoreReview(key, review)
	if got.Found != 3 || !slices.Equal(got.FoundIDs, []string{"a", "b", "c"}) || !slices.Equal(got.Missed, []string{"d"}) {
		t.Errorf("found = %d %v, missed = %v; want a, b, c found and d missed", got.Found, got.FoundIDs, got.Missed)
	}
	if got.MustTotal != 2 || got.MustFound != 2 || got.MustAsMust != 2 {
		t.Errorf("must = %d/%d, as must %d; want 2/2 and 2", got.MustFound, got.MustTotal, got.MustAsMust)
	}
	if got.ReadOnlyTotal != 3 || got.ReadOnlyFound != 2 {
		t.Errorf("read-only = %d/%d; want 2/3", got.ReadOnlyFound, got.ReadOnlyTotal)
	}
	if got.BaitsHit != 1 || got.BaitTotal != 1 {
		t.Errorf("baits = %d/%d; want 1/1", got.BaitsHit, got.BaitTotal)
	}
	if got.Citations != 5 || got.Unkeyed != 1 {
		t.Errorf("citations = %d, unkeyed = %d; want 5 and 1 (server.go:120)", got.Citations, got.Unkeyed)
	}
	if got.Verified != 1 || got.Plausible != 1 {
		t.Errorf("markers verified=%d plausible=%d; want 1 and 1", got.Verified, got.Plausible)
	}

	// The same defect filed under Should Fix is found but not as must.
	demoted := strings.Replace(review, "### Must Fix", "### Should Fix", 1)
	if got := scoreReview(key, demoted); got.MustFound != 2 || got.MustAsMust != 0 {
		t.Errorf("demoted: must found %d, as must %d; want 2 and 0", got.MustFound, got.MustAsMust)
	}

	// An inline label on the citation's own line overrides the heading.
	inline := "### Nits\n- server.go:41 **Must fix**: wrapped error compared with ==\n"
	if got := scoreReview(key, inline); got.MustAsMust != 1 {
		t.Errorf("inline label: as must %d; want 1", got.MustAsMust)
	}

	// A supporting reference later on the line counts toward the defect it
	// lands on and is never an unkeyed finding.
	supporting := "- server.go:41 compares with == although Get wraps the sentinel at store.go:10\n"
	if got := scoreReview(key, supporting); got.Found != 2 || got.Citations != 1 || got.Unkeyed != 0 {
		t.Errorf("supporting reference: found %d, citations %d, unkeyed %d; want 2, 1, 0", got.Found, got.Citations, got.Unkeyed)
	}

	// Outside the tolerance is a miss.
	if got := scoreReview(key, "server.go:44 something"); got.Found != 0 || got.Unkeyed != 1 {
		t.Errorf("three lines off: found %d, unkeyed %d; want 0 and 1", got.Found, got.Unkeyed)
	}
	if got := scoreReview(key, ""); got.Found != 0 || got.Citations != 0 || len(got.Missed) != 4 {
		t.Errorf("empty review: %+v", got)
	}
}

func TestMatchLine(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "f.go"), []byte("a\nneedle\nb\nneedle\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := matchLine(dir, "f.go", "needle", 0); err == nil {
		t.Error("ambiguous match without nth: error = nil, want non-nil")
	}
	if got, err := matchLine(dir, "f.go", "needle", 2); err != nil || got != 4 {
		t.Errorf("nth 2 = %d, %v; want 4", got, err)
	}
	if got, err := matchLine(dir, "f.go", "b", 0); err != nil || got != 3 {
		t.Errorf("unique = %d, %v; want 3", got, err)
	}
	if _, err := matchLine(dir, "f.go", "absent", 0); err == nil {
		t.Error("absent match: error = nil, want non-nil")
	}
}

func TestSessionTools(t *testing.T) {
	if got := sessionTools(corpusReview, true); strings.Contains(got, "Edit") || strings.Contains(got, "Write") || !strings.HasPrefix(got, "Skill,") {
		t.Errorf("review tools = %q; want Skill and read-only tools", got)
	}
	if got := sessionTools(corpusReview, false); strings.HasPrefix(got, "Skill,") {
		t.Errorf("review tools without a plugin = %q; want no Skill", got)
	}
	if got := sessionTools(corpusImplement, true); got != "Skill,Read,Glob,Grep,Edit,Write" {
		t.Errorf("implement tools = %q", got)
	}
}

func TestReviewResultStatus(t *testing.T) {
	ok := result{Build: true, Output: "server.go:41 ...", Review: &reviewScore{}}
	if got := resultStatus(ok); got != "ok " {
		t.Errorf("valid review = %q, want ok", got)
	}
	edited := ok
	edited.Edited = true
	if got := resultStatus(edited); got != "ERR" {
		t.Errorf("review that edited the fixture = %q, want ERR", got)
	}
	silent := ok
	silent.Output = ""
	if got := resultStatus(silent); got != "ERR" {
		t.Errorf("review with no final message = %q, want ERR", got)
	}
}

// TestRescoreReport pins that a saved report is scored again against the
// key as it is now: a citation on a defect the key names counts, the
// re-scored report is written when -out is given, and Rescored is set.
func TestRescoreReport(t *testing.T) {
	root := reviewRoot()
	key, err := loadReviewKey(filepath.Join(root, "_golden", "books", reviewKeyFile), filepath.Join(root, "books"))
	if err != nil {
		t.Fatal(err)
	}
	d := key.Defects[0]
	stale := reviewScore{Total: 1}
	rep := report{Corpus: corpusReview, Arms: []arm{{Name: "no-skill"}}, Results: []result{{
		Arm: "no-skill", Task: "books", Build: true, Review: &stale,
		Output: "### Must Fix\n- " + d.File + ":" + strconv.Itoa(d.line) + " the field is missing\n",
	}}}
	data, err := json.Marshal(rep)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	in, out := filepath.Join(dir, "in.json"), filepath.Join(dir, "out.json")
	if err := os.WriteFile(in, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := rescoreReport(options{rescore: in, out: out}); err != nil {
		t.Fatalf("rescoreReport() error = %v", err)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var scored report
	if err := json.Unmarshal(got, &scored); err != nil {
		t.Fatal(err)
	}
	r := scored.Results[0].Review
	if r == nil || r.Total != len(key.Defects) || r.Found != 1 || !slices.Equal(r.FoundIDs, []string{d.ID}) {
		t.Errorf("re-scored review = %+v, want total %d, found 1 (%s)", r, len(key.Defects), d.ID)
	}
	if scored.Rescored.IsZero() {
		t.Error("Rescored is zero, want the time of the re-score")
	}
}
