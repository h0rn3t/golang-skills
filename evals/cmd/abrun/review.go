package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

// The review corpus hands the model a working package with seeded defects and
// asks for a review. Nothing is edited: the final message is the review, and
// it is scored against a hidden key that names each defect by the source line
// it sits on. Recall is the score; a citation on a bait line — code that is
// correct and looks suspicious — and a citation that matches no key entry are
// counted next to it, so a review that flags every line cannot win.
const reviewPrompt = "Review the Go package in ./%s as a pull request reviewer would. " +
	"Report every defect you find as a finding with the file and line it is at (file.go:NN), " +
	"its severity (Must Fix, Should Fix, or Nit), what is wrong, and the fix. " +
	"Do not modify any file; your final message is the review."

const corpusReview = "review"

// reviewDir holds the review corpus. The leading underscore keeps findTasks
// from offering it as a refactor fixture, the same way implementDir does.
const reviewDir = "_review"

// reviewKeyFile is the answer key under _golden/<task>/.
const reviewKeyFile = "key.json"

// reviewBaseDir, under _golden/<task>/, makes a fixture a pull request: it
// holds the files the change modified as they were before it. The session's
// tree gets a git history of two commits, base and change, and the prompt —
// the fixture's prompt.txt, which states what the change is for — gets the
// diff between them, since a review session has no shell to run git.
const reviewBaseDir = "base"

// reviewTolerance is how many lines a citation may sit from the defect's own
// lines and still count. A reviewer cites the statement, the enclosing
// declaration, or the line after the opening brace; two lines covers all
// three without reaching the next defect in these fixtures.
const reviewTolerance = 2

// reviewDefect is one seeded defect. Match is a substring of one source line
// in File; Nth picks the occurrence when the substring repeats (1-based,
// default 1); Span is how many lines from that one the defect covers (default
// 1). Tool names the linter or analyzer in the bundled gate that reports the
// line, so the report can separate what a tool would have said from what only
// reading finds. PreExisting marks a defect on lines the fixture's change did
// not touch; only a fixture with a base (reviewBaseDir) can have one.
type reviewDefect struct {
	ID          string `json:"id"`
	Owner       string `json:"owner"`
	Severity    string `json:"severity"`
	File        string `json:"file"`
	Match       string `json:"match"`
	Nth         int    `json:"nth,omitempty"`
	Span        int    `json:"span,omitempty"`
	Tool        string `json:"tool,omitempty"`
	PreExisting bool   `json:"pre_existing,omitempty"`
	Note        string `json:"note"`
	line        int
}

// reviewBait is a correct line a weak review flags anyway.
type reviewBait struct {
	ID    string `json:"id"`
	File  string `json:"file"`
	Match string `json:"match"`
	Nth   int    `json:"nth,omitempty"`
	Note  string `json:"note"`
	line  int
}

type reviewKey struct {
	Defects []reviewDefect `json:"defects"`
	Baits   []reviewBait   `json:"baits"`
}

// reviewScore is what one review earned against the key.
type reviewScore struct {
	Total    int      `json:"total"`
	Found    int      `json:"found"`
	FoundIDs []string `json:"found_ids"`
	Missed   []string `json:"missed"`
	// MustTotal and MustFound count the must-severity defects; MustAsMust is
	// how many of the found ones the review filed under Must Fix.
	MustTotal  int `json:"must_total"`
	MustFound  int `json:"must_found"`
	MustAsMust int `json:"must_as_must"`
	// ReadOnlyTotal and ReadOnlyFound count the defects no tool in the bundled
	// gate reports: what the review found by reading.
	ReadOnlyTotal int `json:"read_only_total"`
	ReadOnlyFound int `json:"read_only_found"`
	BaitTotal     int `json:"bait_total"`
	BaitsHit      int `json:"baits_hit"`
	// Citations is every distinct file:line a finding is anchored at (the
	// first citation on its line); Unkeyed is how many of them sit on neither
	// a defect nor a bait.
	Citations int `json:"citations"`
	Unkeyed   int `json:"unkeyed"`
	// Verified and Plausible count the evidence markers the review template
	// asks for; they read prose, so they say what the review claimed.
	Verified  int `json:"verified"`
	Plausible int `json:"plausible"`
	// PreTotal counts the pre-existing defects; PreTagged is how many of the
	// found ones the review marked pre-existing, and NewTagged how many found
	// defects of the change itself it marked so, wrongly.
	PreTotal  int `json:"pre_total,omitempty"`
	PreFound  int `json:"pre_found,omitempty"`
	PreTagged int `json:"pre_tagged,omitempty"`
	NewTagged int `json:"new_tagged,omitempty"`
}

// loadReviewKey reads the key and resolves every Match to a line of the
// pristine fixture. An entry whose substring is missing, or repeats without
// an Nth, is an error: a key that cannot place its own defects scores nothing.
func loadReviewKey(path, fixtureDir string) (reviewKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return reviewKey{}, err
	}
	var key reviewKey
	if err := json.Unmarshal(data, &key); err != nil {
		return reviewKey{}, fmt.Errorf("%s: %w", path, err)
	}
	if len(key.Defects) == 0 {
		return reviewKey{}, fmt.Errorf("%s: no defects", path)
	}
	seen := map[string]bool{}
	for i := range key.Defects {
		d := &key.Defects[i]
		if d.ID == "" || seen[d.ID] {
			return reviewKey{}, fmt.Errorf("%s: defect %d: missing or repeated id %q", path, i, d.ID)
		}
		seen[d.ID] = true
		switch d.Severity {
		case "must", "should", "nit":
		default:
			return reviewKey{}, fmt.Errorf("%s: defect %s: severity %q is not must, should, or nit", path, d.ID, d.Severity)
		}
		if d.Span <= 0 {
			d.Span = 1
		}
		if d.line, err = matchLine(fixtureDir, d.File, d.Match, d.Nth); err != nil {
			return reviewKey{}, fmt.Errorf("%s: defect %s: %w", path, d.ID, err)
		}
	}
	for i := range key.Baits {
		b := &key.Baits[i]
		if b.ID == "" || seen[b.ID] {
			return reviewKey{}, fmt.Errorf("%s: bait %d: missing or repeated id %q", path, i, b.ID)
		}
		seen[b.ID] = true
		if b.line, err = matchLine(fixtureDir, b.File, b.Match, b.Nth); err != nil {
			return reviewKey{}, fmt.Errorf("%s: bait %s: %w", path, b.ID, err)
		}
	}
	return key, nil
}

// matchLine returns the 1-based line of the nth occurrence of match in the
// file, or of the only one when nth is zero.
func matchLine(dir, file, match string, nth int) (int, error) {
	data, err := os.ReadFile(filepath.Join(dir, file))
	if err != nil {
		return 0, err
	}
	var hits []int
	for i, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, match) {
			hits = append(hits, i+1)
		}
	}
	switch {
	case len(hits) == 0:
		return 0, fmt.Errorf("%s: %q not found", file, match)
	case nth == 0 && len(hits) > 1:
		return 0, fmt.Errorf("%s: %q occurs %d times; set nth", file, match, len(hits))
	case nth == 0:
		return hits[0], nil
	case nth > len(hits):
		return 0, fmt.Errorf("%s: %q occurs %d times, nth is %d", file, match, len(hits), nth)
	}
	return hits[nth-1], nil
}

// citation matches a file:line reference in prose, with an optional range:
// server.go:41, ./orders/server.go:41:6, pool.go:38-40.
var citation = regexp.MustCompile(`([\w./-]*?([\w-]+\.go)):(\d+)(?::\d+)?(?:\s*[-–]\s*(\d+))?`)

var (
	mustHeading   = regexp.MustCompile(`(?i)\bmust[ -]?fix\b|\bcritical\b|\bblocker\b`)
	shouldHeading = regexp.MustCompile(`(?i)\bshould[ -]?fix\b`)
	nitHeading    = regexp.MustCompile(`(?i)\bnits?\b|\bminor\b`)
	verifiedMark  = regexp.MustCompile(`(?i)\bverified\b`)
	plausibleMark = regexp.MustCompile(`(?i)\bplausible\b`)
	preMark       = regexp.MustCompile(`(?i)\bpre-?existing\b`)
)

// cited is one distinct location the review named, with the severity it was
// filed under: the label on its own line when there is one, else the last
// heading above it. anchor marks the first citation on its line — the
// finding's own location; a later one on the same line is a supporting
// reference ("Get wraps the sentinel at store.go:34"), which counts toward a
// defect it lands on but is never an unkeyed finding. pre is set when the
// citation's line or the Markdown heading above it says pre-existing.
type cited struct {
	file     string
	from, to int
	severity string
	anchor   bool
	pre      bool
}

// parseCitations walks the review line by line. A heading sets the severity
// for what follows; a severity word on the citation's own line overrides it.
// A Markdown heading also sets or clears the pre-existing mark.
func parseCitations(review string) []cited {
	var out []cited
	current := ""
	preSection := false
	for line := range strings.SplitSeq(review, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			preSection = preMark.MatchString(line)
		}
		inline := severityOf(line)
		if inline != "" && !citation.MatchString(line) {
			current = inline
			continue
		}
		pre := preSection || preMark.MatchString(line)
		for i, m := range citation.FindAllStringSubmatch(line, -1) {
			from, _ := strconv.Atoi(m[3])
			to := from
			if m[4] != "" {
				to, _ = strconv.Atoi(m[4])
			}
			if to < from {
				from, to = to, from
			}
			sev := current
			if inline != "" {
				sev = inline
			}
			out = append(out, cited{file: m[2], from: from, to: to, severity: sev, anchor: i == 0, pre: pre})
		}
	}
	return out
}

func severityOf(line string) string {
	switch {
	case mustHeading.MatchString(line):
		return "must"
	case shouldHeading.MatchString(line):
		return "should"
	case nitHeading.MatchString(line):
		return "nit"
	}
	return ""
}

// near reports whether a citation lands on the window [line, line+span-1]
// widened by the tolerance.
func (c cited) near(file string, line, span int) bool {
	return c.file == file && c.to >= line-reviewTolerance && c.from <= line+span-1+reviewTolerance
}

// scoreReview matches the review's citations against the key.
func scoreReview(key reviewKey, review string) reviewScore {
	cites := parseCitations(review)
	score := reviewScore{Total: len(key.Defects), BaitTotal: len(key.Baits)}
	score.Verified = len(verifiedMark.FindAllString(review, -1))
	score.Plausible = len(plausibleMark.FindAllString(review, -1))
	for _, d := range key.Defects {
		if d.Severity == "must" {
			score.MustTotal++
		}
		if d.Tool == "" {
			score.ReadOnlyTotal++
		}
		if d.PreExisting {
			score.PreTotal++
		}
		found, asMust, pre := false, false, false
		for _, c := range cites {
			if c.near(d.File, d.line, d.Span) {
				found = true
				asMust = asMust || c.severity == "must"
				pre = pre || c.pre
			}
		}
		if !found {
			score.Missed = append(score.Missed, d.ID)
			continue
		}
		score.Found++
		switch {
		case d.PreExisting:
			score.PreFound++
			if pre {
				score.PreTagged++
			}
		case pre:
			score.NewTagged++
		}
		score.FoundIDs = append(score.FoundIDs, d.ID)
		if d.Severity == "must" {
			score.MustFound++
			if asMust {
				score.MustAsMust++
			}
		}
		if d.Tool == "" {
			score.ReadOnlyFound++
		}
	}
	for _, b := range key.Baits {
		for _, c := range cites {
			if c.near(b.File, b.line, 1) {
				score.BaitsHit++
				break
			}
		}
	}
	distinct := map[string]cited{}
	for _, c := range cites {
		if c.anchor {
			distinct[c.file+":"+strconv.Itoa(c.from)+"-"+strconv.Itoa(c.to)] = c
		}
	}
	score.Citations = len(distinct)
	for _, c := range distinct {
		keyed := false
		for _, d := range key.Defects {
			if c.near(d.File, d.line, d.Span) {
				keyed = true
			}
		}
		for _, b := range key.Baits {
			if c.near(b.File, b.line, 1) {
				keyed = true
			}
		}
		if !keyed {
			score.Unkeyed++
		}
	}
	sort.Strings(score.FoundIDs)
	sort.Strings(score.Missed)
	return score
}

// validateReviewFixtures is validateFixtures for the review corpus: a fixture
// needs production Go files and a key that resolves against them. A fixture
// with a base also needs its prompt.txt, and every defect must sit where the
// key says: a pre-existing one on lines the change left alone, any other on a
// line it changed.
func validateReviewFixtures(abDir string, tasks []string) error {
	for _, task := range tasks {
		fixture := filepath.Join(abDir, task)
		if ok, err := containsFile(fixture, func(name string) bool {
			return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
		}); err != nil {
			return fmt.Errorf("validate fixture %s: %w", task, err)
		} else if !ok {
			return exitError{2, fmt.Sprintf("fixture %s contains no production Go files", task)}
		}
		golden := filepath.Join(abDir, "_golden", task)
		key, err := loadReviewKey(filepath.Join(golden, reviewKeyFile), fixture)
		if err != nil {
			return exitError{2, fmt.Sprintf("fixture %s: %v", task, err)}
		}
		changed, err := baseChanges(filepath.Join(golden, reviewBaseDir), fixture)
		if err != nil {
			return exitError{2, fmt.Sprintf("fixture %s: %v", task, err)}
		}
		if len(changed) == 0 {
			for _, d := range key.Defects {
				if d.PreExisting {
					return exitError{2, fmt.Sprintf("fixture %s: defect %s is pre_existing, but the fixture has no %s/", task, d.ID, reviewBaseDir)}
				}
			}
			continue
		}
		if prompt, err := fixturePrompt(golden); err != nil || prompt == "" {
			return exitError{2, fmt.Sprintf("fixture %s: a fixture with a %s/ needs a %s stating the change (%v)", task, reviewBaseDir, fixturePromptFile, err)}
		}
		for _, d := range key.Defects {
			touched := slices.ContainsFunc(changed[d.File], func(line int) bool {
				return line >= d.line && line < d.line+d.Span
			})
			if touched == d.PreExisting {
				return exitError{2, fmt.Sprintf("fixture %s: defect %s (pre_existing=%t) at %s:%d; changed lines there: %v", task, d.ID, d.PreExisting, d.File, d.line, changed[d.File])}
			}
		}
	}
	return nil
}

// baseChanges returns, per file of the fixture's base, the lines of the
// fixture the change added or rewrote; empty when there is no base. A base
// file identical to the fixture's is an error: it describes no change.
func baseChanges(baseDir, fixture string) (map[string][]int, error) {
	changed := map[string][]int{}
	entries, err := os.ReadDir(baseDir)
	if errors.Is(err, fs.ErrNotExist) {
		return changed, nil
	}
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		base, err := os.ReadFile(filepath.Join(baseDir, e.Name()))
		if err != nil {
			return nil, err
		}
		head, err := os.ReadFile(filepath.Join(fixture, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("%s/%s has no counterpart in the fixture: %w", reviewBaseDir, e.Name(), err)
		}
		lines := changedLines(strings.Split(string(base), "\n"), strings.Split(string(head), "\n"))
		if len(lines) == 0 {
			return nil, fmt.Errorf("%s/%s is identical to the fixture's", reviewBaseDir, e.Name())
		}
		changed[e.Name()] = lines
	}
	return changed, nil
}

// changedLines returns the 1-based lines of head outside a longest common
// subsequence with base: what a diff shows as added. Fixture files are a few
// hundred lines, so the quadratic table is fine.
func changedLines(base, head []string) []int {
	lcs := make([][]int, len(base)+1)
	for i := range lcs {
		lcs[i] = make([]int, len(head)+1)
	}
	for i, b := range slices.Backward(base) {
		for j := len(head) - 1; j >= 0; j-- {
			if b == head[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	var out []int
	i, j := 0, 0
	for j < len(head) {
		switch {
		case i < len(base) && base[i] == head[j]:
			i++
			j++
		case i < len(base) && lcs[i+1][j] >= lcs[i][j+1]:
			i++
		default:
			out = append(out, j+1)
			j++
		}
	}
	return out
}

// stageReviewBase turns the session's copy of the fixture into a two-commit
// history — the base files, then the fixture as it is — and returns the diff
// between them. The tree ends as it started; only .git is added.
func stageReviewBase(timeout time.Duration, work, task, baseDir string) (string, error) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return "", err
	}
	pkg := filepath.Join(work, task)
	head := map[string][]byte{}
	for _, e := range entries {
		if head[e.Name()], err = os.ReadFile(filepath.Join(pkg, e.Name())); err != nil {
			return "", err
		}
		base, err := os.ReadFile(filepath.Join(baseDir, e.Name()))
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(filepath.Join(pkg, e.Name()), base, 0o600); err != nil { //nolint:gosec // G703: names come from the repository's own base/ listing

			return "", err
		}
	}
	git := func(args ...string) (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		// The user's git configuration — hooks, signing, templates — stays out.
		args = append([]string{"-c", "user.name=abrun", "-c", "user.email=abrun@example.invalid", "-c", "commit.gpgsign=false"}, args...)
		cmd := exec.CommandContext(ctx, "git", args...) //nolint:gosec // G204: fixed program, argv built here, no input
		cmd.Dir = work
		cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
		}
		return string(out), nil
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "-A"}, {"commit", "-q", "-m", "base"}} {
		if _, err := git(args...); err != nil {
			return "", err
		}
	}
	for name, data := range head {
		if err := os.WriteFile(filepath.Join(pkg, name), data, 0o600); err != nil {
			return "", err
		}
	}
	if _, err := git("commit", "-q", "-am", "change"); err != nil {
		return "", err
	}
	return git("diff", "HEAD~1", "HEAD")
}

// reviewDiffPrompt appends the change's diff to a fixture prompt.
func reviewDiffPrompt(prompt, diff string) string {
	return prompt + "\n\nThe change is the last commit in the tree (git diff HEAD~1 HEAD):\n\n```diff\n" + diff + "```"
}

// rescoreReport scores the review results of a saved report again, against
// the keys as they are now, and prints the summary. A key is amended after a
// run from what the reviews cited, and the report's numbers are then about a
// key that no longer exists; the review score depends on nothing but the
// final message, so it is recomputed without spending a session. -out writes
// the re-scored report, with Rescored set.
func rescoreReport(o options) error {
	data, err := os.ReadFile(o.rescore)
	if err != nil {
		return err
	}
	var rep report
	if err := json.Unmarshal(data, &rep); err != nil {
		return fmt.Errorf("%s: %w", o.rescore, err)
	}
	if rep.Corpus != corpusReview {
		return exitError{2, fmt.Sprintf("%s is a %s report; -rescore applies to the review corpus", o.rescore, rep.Corpus)}
	}
	root, err := repoRoot()
	if err != nil {
		return err
	}
	abDir := filepath.Join(root, "evals", "ab", reviewDir)
	keys := map[string]reviewKey{}
	for i := range rep.Results {
		r := &rep.Results[i]
		if r.Review == nil {
			continue
		}
		key, ok := keys[r.Task]
		if !ok {
			if key, err = loadReviewKey(filepath.Join(abDir, "_golden", r.Task, reviewKeyFile), filepath.Join(abDir, r.Task)); err != nil {
				return err
			}
			keys[r.Task] = key
		}
		score := scoreReview(key, r.Output)
		r.Review = &score
		printReviewResult(*r, false)
	}
	rep.Rescored = time.Now().UTC()
	if o.out != "" {
		out, err := json.MarshalIndent(rep, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(o.out, out, 0o644); err != nil {
			return err
		}
	}
	printSummary(rep)
	return nil
}

// sessionTools is the claude tool set for a corpus. A review session gets no
// editing tool: the prompt says not to modify anything, and a session that
// cannot is one whose Edited flag needs no explaining. Without a plugin the
// Skill tool goes too.
func sessionTools(corpus string, hasPlugin bool) string {
	tools := "Skill,Read,Glob,Grep,Edit,Write"
	if corpus == corpusReview {
		tools = "Skill,Read,Glob,Grep"
	}
	if !hasPlugin {
		tools = strings.TrimPrefix(tools, "Skill,")
	}
	return tools
}

// reviewSummary is the review corpus's share of an armSummary, summed over
// valid runs.
type reviewSummary struct {
	Runs                         int
	Found, Total                 int
	MustFound, MustTotal         int
	MustAsMust                   int
	ReadOnlyFound, ReadOnlyTotal int
	BaitsHit, BaitTotal          int
	Citations, Unkeyed           int
	Verified, Plausible          int
	PreFound, PreTotal           int
	PreTagged, NewTagged         int
	// PerDefect counts the valid runs that found each defect id.
	PerDefect map[string]int
}

func (s *reviewSummary) add(r reviewScore) {
	s.Runs++
	s.Found += r.Found
	s.Total += r.Total
	s.MustFound += r.MustFound
	s.MustTotal += r.MustTotal
	s.MustAsMust += r.MustAsMust
	s.ReadOnlyFound += r.ReadOnlyFound
	s.ReadOnlyTotal += r.ReadOnlyTotal
	s.BaitsHit += r.BaitsHit
	s.BaitTotal += r.BaitTotal
	s.Citations += r.Citations
	s.Unkeyed += r.Unkeyed
	s.Verified += r.Verified
	s.Plausible += r.Plausible
	s.PreFound += r.PreFound
	s.PreTotal += r.PreTotal
	s.PreTagged += r.PreTagged
	s.NewTagged += r.NewTagged
	if s.PerDefect == nil {
		s.PerDefect = map[string]int{}
	}
	for _, id := range r.FoundIDs {
		s.PerDefect[id]++
	}
	for _, id := range r.Missed {
		s.PerDefect[id] += 0
	}
}

func printReviewResult(r result, verbose bool) {
	s := r.Review
	fmt.Printf("[%s] %-24s %-10s #%d  found %d/%d  must %d/%d (as must %d)  read-only %d/%d  baits %d/%d  unkeyed %d/%d  skills=%v",
		resultStatus(r), r.Arm, r.Task, r.Rep, s.Found, s.Total, s.MustFound, s.MustTotal, s.MustAsMust,
		s.ReadOnlyFound, s.ReadOnlyTotal, s.BaitsHit, s.BaitTotal, s.Unkeyed, s.Citations, r.Skills)
	if s.PreTotal > 0 {
		fmt.Printf("  pre-existing found %d/%d tagged %d, new tagged %d", s.PreFound, s.PreTotal, s.PreTagged, s.NewTagged)
	}
	if r.Err != "" {
		fmt.Printf("  error: %s", r.Err)
	}
	fmt.Println()
	if r.Edited {
		fmt.Printf("       the session modified the fixture, which a review must not\n")
	}
	if len(s.Missed) > 0 {
		fmt.Printf("       missed: %s\n", strings.Join(s.Missed, ", "))
	}
	if r.Trace != "" {
		fmt.Printf("       trace: %s\n", r.Trace)
	}
	if verbose && r.Output != "" {
		fmt.Println("       --- output ---")
		fmt.Println(r.Output)
	}
}

// printReviewSummary is printSummary for the review corpus: recall per arm,
// then the defects one by one, so the report can say which defect the skill
// text made the difference on.
func printReviewSummary(rep report) {
	fmt.Printf("%-24s %5s %6s %5s %7s %7s %8s %9s %7s %8s %6s %8s\n",
		"arm", "runs", "errors", "valid", "recall", "must", "as-must", "read-only", "baits", "unkeyed", "skill", "$/run")
	summaries := map[string]armSummary{}
	for _, a := range rep.Arms {
		summary := summarizeArm(rep, a.Name)
		summaries[a.Name] = summary
		rs := summary.Review
		if summary.Valid == 0 || rs.Runs == 0 {
			fmt.Printf("%-24s %5d %6d %5d %7s\n", a.Name, summary.Runs, summary.Errors, 0, "no data")
			printRoutingSummary(rep, a.Name)
			continue
		}
		ratio := func(num, den int) string {
			if den == 0 {
				return "  n/a"
			}
			return fmt.Sprintf("%5.2f", float64(num)/float64(den))
		}
		completed := summary.Runs - summary.Errors
		skill := 0
		if completed > 0 {
			skill = 100 * summary.SkillFired / completed
		}
		cost := 0.0
		if summary.Costed > 0 {
			cost = summary.Cost / float64(summary.Costed)
		}
		fmt.Printf("%-24s %5d %6d %5d %7s %7s %8s %9s %7s %8s %5d%% %8.4f\n",
			a.Name, summary.Runs, summary.Errors, summary.Valid,
			ratio(rs.Found, rs.Total), ratio(rs.MustFound, rs.MustTotal), ratio(rs.MustAsMust, rs.MustFound),
			ratio(rs.ReadOnlyFound, rs.ReadOnlyTotal), ratio(rs.BaitsHit, rs.BaitTotal), ratio(rs.Unkeyed, rs.Citations),
			skill, cost)
		fmt.Printf("%-24s   %.1f citations/run, evidence markers verified %.1f plausible %.1f /run\n",
			"", float64(rs.Citations)/float64(rs.Runs), float64(rs.Verified)/float64(rs.Runs), float64(rs.Plausible)/float64(rs.Runs))
		if rs.PreTotal > 0 {
			fmt.Printf("%-24s   pre-existing found %d/%d, tagged pre-existing %d; change defects tagged pre-existing %d\n",
				"", rs.PreFound, rs.PreTotal, rs.PreTagged, rs.NewTagged)
		}
		printRoutingSummary(rep, a.Name)
	}

	// Per-defect table: rows are defect ids, columns are arms, cells are
	// found/valid runs on that fixture.
	ids := map[string]string{}
	for _, r := range rep.Results {
		if r.Review == nil {
			continue
		}
		for _, id := range append(append([]string{}, r.Review.FoundIDs...), r.Review.Missed...) {
			ids[r.Task+"/"+id] = r.Task
		}
	}
	if len(ids) == 0 {
		return
	}
	keys := make([]string, 0, len(ids))
	for k := range ids {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Printf("\n%-40s", "defect")
	for _, a := range rep.Arms {
		fmt.Printf(" %12s", a.Name)
	}
	fmt.Println()
	for _, k := range keys {
		task, id, _ := strings.Cut(k, "/")
		fmt.Printf("%-40s", k)
		for _, a := range rep.Arms {
			found, valid := 0, 0
			for _, r := range rep.Results {
				if r.Arm != a.Name || r.Task != task || r.Review == nil || resultStatus(r) != "ok " {
					continue
				}
				valid++
				if slices.Contains(r.Review.FoundIDs, id) {
					found++
				}
			}
			fmt.Printf(" %12s", fmt.Sprintf("%d/%d", found, valid))
		}
		fmt.Println()
	}
}
