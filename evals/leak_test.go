package evals_test

import (
	"maps"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"unicode"
)

// A validation eval measures whether the skills generalize only while the
// skill text does not carry the case itself. TestLeakValidationEvals flags a
// validation case whose prompt reappears in skills/go-*/SKILL.md or
// skills/go-*/references/*.md: shared word 5-grams (tokenized as the review's
// leak.py does) or identifiers that belong to the case's scenario rather than
// to Go. Only the prompt is compared: an expected answer that restates the
// owner skill's rule is the skill doing its job, while the prompt's names and
// wording in a skill example are the case answered in advance.
//
// Both signals count only what at most leakMaxFiles skill files carry. A
// phrase or name spread wider is house vocabulary (`if err != nil return`,
// ErrNotFound), not one case's scenario. A paraphrase that keeps the case's
// numbers or decisions but not its wording passes both signals; the known ones
// carry their confirmed passage as evidence in knownLeaks.
const (
	leakShingle  = 5 // words per shingle
	leakMaxFiles = 2 // skill files a shingle or identifier may appear in and still count
	leakShingles = 3 // shared prompt shingles that flag a case
	leakIdents   = 2 // shared characteristic identifiers that flag a case
)

// knownLeaks are the validation cases whose answer the skill text is known to
// carry, each with the passage that carries it. An entry stays only while the
// case still leaks, by either signal or by its evidence; fixing the skill text
// (or moving the case to train) means deleting the entry, so the list only
// shrinks. The review of 2026-10-28 (X7) listed 19, 34–39, 57 and 59; this
// guard found 31, 63 and 70.
var knownLeaks = map[int]struct{ reason, evidence string }{
	19: {"go-database SQL-PATTERNS.md implements the prompt's exact Transfer signature (X7)",
		"Transfer(ctx context.Context, from, to int64, cents int64) error"},
	31: {"go-troubleshooting DATA-FLOW-TRACING.md walks the prompt's byID[row.LocalID] overwrite and (tenant, local_id) contract (guard)",
		"byID[row.LocalID] = item"},
	34: {"go-resilience RETRIES-AND-IDEMPOTENCY.md works the prompt's 3 s, 500 ms, Retry-After 2 and 60, 250 ms, 1 s cap example (X7, D-8)",
		"a safe GET with 3 seconds left, a 500 ms attempt limit and `Retry-After: 2`"},
	35: {"go-resilience RETRIES-AND-IDEMPOTENCY.md states the answer: a persisted logical ID is not the key the first attempt sent (X7, D-8)",
		"A persisted logical ID is not evidence that the request used it as its key"},
	36: {"go-resilience RETRIES-AND-IDEMPOTENCY.md names the prompt's SELECT, side effect, INSERT under a process mutex (X7, D-8)",
		"A `SELECT` then side effect then `INSERT`, even with a process mutex, races across replicas"},
	37: {"go-resilience RETRIES-AND-IDEMPOTENCY.md computes the prompt's 3 x 4 x 2 = 24 attempts (X7, D-8)",
		"3, 4, and 2 total attempts can produce 24 downstream attempts"},
	38: {"go-resilience LOAD-AND-DEGRADATION.md: the prompt's ten replicas at 500/s permit 5000/s against a 500/s quota (X7, D-8)",
		"Ten replicas with 500/s local limiters can permit 5000/s"},
	39: {"go-resilience LOAD-AND-DEGRADATION.md answers the prompt's half-open release of every queued caller (X7, D-8)",
		"Do not release all queued callers when the open interval ends"},
	57: {"go-style-core INITIALIZATION.md declares the prompt's Audit and Document types and writes the answer's literal (X7)",
		"doc := Document{CreatedBy: owner, Name: name}"},
	59: {"go-data-structures SKILL.md writes the answer with the prompt's dst, updates and minimum (X7, C-1)",
		"maps.DeleteFunc(dst, func(_ string, score int) bool { return score < minimum })"},
	63: {"go-security SECRETS-AND-CRYPTO.md carries the prompt's ErrNoUser login and 19 MiB argon2.IDKey cap, added with the case (guard)",
		"case errors.Is(err, ErrNoUser):"},
	70: {"go-logging SKILL.md states the answer for the prompt's otelslog and TraceIDRatioBased setup (guard)",
		"`otelslog.NewHandler` writes to the global OTel `LoggerProvider`, a no-op until `global.SetLoggerProvider` installs one"},
}

var (
	leakWord  = regexp.MustCompile(`[a-z0-9_.%<>=*/:+-]+`)
	leakIdent = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)*`)
	mixedCaps = regexp.MustCompile(`[a-z0-9][A-Z]|[A-Z]{2}[a-z]`)
)

func TestLeakValidationEvals(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	var files []string
	for _, pattern := range []string{"skills/go-*/SKILL.md", "skills/go-*/references/*.md"} {
		matches, err := filepath.Glob(filepath.Join(root, filepath.FromSlash(pattern)))
		if err != nil {
			t.Fatalf("glob %s: %v", pattern, err)
		}
		files = append(files, matches...)
	}
	if len(files) == 0 {
		t.Fatal("no skill text found")
	}
	shingleFiles := map[string]int{}
	identFiles := map[string]int{}
	var normalized strings.Builder
	for _, f := range files {
		text := readFile(t, f)
		for s := range shingles(text) {
			shingleFiles[s]++
		}
		for id := range identifiers(text) {
			identFiles[id]++
		}
		// One line per file, so an evidence phrase never spans two files.
		normalized.WriteString(strings.Join(strings.Fields(text), " ") + "\n")
	}
	corpus := normalized.String()

	stdPkgs, stdNames := map[string]bool{}, map[string]bool{}
	for key := range apiIndex(t) {
		pkg, name, ok := strings.Cut(key, ".")
		if !ok {
			continue
		}
		// "syscall (linux-386)" is the syscall package on one platform.
		pkg, _, _ = strings.Cut(pkg, " ")
		stdPkgs[pkg] = true
		stdNames[name] = true
	}

	var evals struct {
		Quality []struct {
			ID     int    `json:"id"`
			Prompt string `json:"prompt"`
			Set    string `json:"set"`
		} `json:"quality_evals"`
	}
	readJSON(t, filepath.Join(root, "evals", "evals.json"), &evals)

	validation := map[int]bool{}
	for _, ev := range evals.Quality {
		if ev.Set != "validation" {
			continue
		}
		validation[ev.ID] = true
		var shared, ids []string
		for s := range shingles(ev.Prompt) {
			if n := shingleFiles[s]; n > 0 && n <= leakMaxFiles {
				shared = append(shared, s)
			}
		}
		for id := range identifiers(ev.Prompt) {
			if n := identFiles[id]; n > 0 && n <= leakMaxFiles && characteristic(id, stdPkgs, stdNames) {
				ids = append(ids, id)
			}
		}
		slices.Sort(shared)
		slices.Sort(ids)
		known, isKnown := knownLeaks[ev.ID]
		evidence := isKnown && strings.Contains(corpus, known.evidence)
		leaks := len(shared) >= leakShingles || len(ids) >= leakIdents || evidence
		switch {
		case isKnown && leaks:
			t.Logf("validation eval %d leaks (known: %s): shingles %q, identifiers %q, evidence present %t",
				ev.ID, known.reason, shared, ids, evidence)
		case isKnown:
			t.Errorf("validation eval %d no longer leaks into skill text (%s): remove it from knownLeaks", ev.ID, known.reason)
		case leaks:
			t.Errorf("validation eval %d leaks into skill text: %d shared prompt %d-grams %q, %d characteristic identifiers %q; "+
				"reword the skill example, or move the case to train and write a fresh validation case",
				ev.ID, len(shared), leakShingle, shared, len(ids), ids)
		}
	}
	for _, id := range slices.Sorted(maps.Keys(knownLeaks)) {
		if !validation[id] {
			t.Errorf("knownLeaks lists eval %d, which is not a validation quality eval: remove it", id)
		}
	}
}

// shingles returns the word 5-grams of text, lowercased and split as leak.py
// splits them.
func shingles(text string) map[string]bool {
	words := leakWord.FindAllString(strings.ToLower(text), -1)
	set := map[string]bool{}
	for i := 0; i+leakShingle <= len(words); i++ {
		set[strings.Join(words[i:i+leakShingle], " ")] = true
	}
	return set
}

// identifiers returns the identifiers and selector chains in text, as written.
func identifiers(text string) map[string]bool {
	set := map[string]bool{}
	for _, id := range leakIdent.FindAllString(text, -1) {
		set[id] = true
	}
	return set
}

// characteristic reports whether an identifier names part of a case's scenario
// rather than of Go: a mixedCaps name the standard library does not export
// (CreatedBy, byID), or an exported or mixedCaps selector on something other
// than a standard-library package that selects at least one name the library
// does not export (row.LocalID, otelslog.NewHandler). A selector on a value of
// a library type (r.URL.Query, client.Do) is Go, and so are file names
// (report.go) and prose (e.g).
func characteristic(id string, stdPkgs, stdNames map[string]bool) bool {
	parts := strings.Split(id, ".")
	if len(parts) == 1 {
		return mixedCaps.MatchString(id) && !stdNames[id]
	}
	last := parts[len(parts)-1]
	if stdPkgs[parts[0]] || !mixedCaps.MatchString(last) && !unicode.IsUpper(rune(last[0])) {
		return false
	}
	return slices.ContainsFunc(parts[1:], func(p string) bool { return !stdNames[p] })
}
