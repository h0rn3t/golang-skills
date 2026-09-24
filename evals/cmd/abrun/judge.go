package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// judgeModelDefault differs from the generator the pack is measured on, so the
// judge does not prefer its own model's style.
const judgeModelDefault = "claude-fable-5-1"

// judgement is one blind pairwise comparison of two arms' production code on
// the same task and repetition. The judge sees the two diffs as A and B, in
// both orders; a preference counts only when both orders name the same arm.
type judgement struct {
	Task string    `json:"task"`
	Rep  int       `json:"rep"`
	Pair [2]string `json:"pair"`
	// Verdicts holds the arm each order chose, or "tie": [0] showed Pair[0]
	// as A, [1] showed Pair[1] as A.
	Verdicts  [2]string `json:"verdicts,omitzero"`
	Reasons   [2]string `json:"reasons,omitzero"`
	Preferred string    `json:"preferred,omitempty"`
	// Disagree marks a verdict that moved with the order: position, not code,
	// decided it, and the pair is scored as a tie.
	Disagree bool   `json:"positional_disagreement,omitempty"`
	Skipped  string `json:"skipped,omitempty"`
}

type verdict struct {
	Winner string `json:"winner"`
	Reason string `json:"reason"`
}

// judgeFunc asks the judge one question. Tests replace the model with a fake.
type judgeFunc func(prompt string) (verdict, error)

const verdictSchema = `{"type":"object","additionalProperties":false,"required":["winner","reason"],` +
	`"properties":{"winner":{"type":"string","enum":["A","B","tie"]},"reason":{"type":"string"}}}`

const judgeRubric = `You compare two changes to the same Go package for readability. Both pass the
same tests; judge only how the resulting production code reads.

Prefer the change where a reader finds the main decision, the error exits, and
the state the code tracks without chasing wrappers. Count against a change:
helpers that only rename two or three lines, comments that narrate the next
line, doc comments that repeat the name, interfaces or options nothing needs,
logging an error and also returning it, long functions and deep nesting that a
plainer shape would avoid. Do not reward length or brevity for its own sake.
Answer "tie" when neither reads better.

CHANGE A:
<<<
%s
>>>

CHANGE B:
<<<
%s
>>>`

// claudeJudge asks the judge model through claude -p with no tools and the
// verdict enforced by --json-schema.
func claudeJudge(o options) judgeFunc {
	return func(prompt string) (verdict, error) {
		dir, err := os.MkdirTemp("", "abrun-judge-")
		if err != nil {
			return verdict{}, err
		}
		defer func() { _ = os.RemoveAll(dir) }() // an empty scratch dir; a leftover costs nothing
		out, err := claude(o.timeout, dir, "-p", prompt, "--output-format", "json", "--json-schema", verdictSchema,
			"--max-turns", "2", "--tools", "", "--restricted", "--model", o.judgeModel)
		if err != nil {
			return verdict{}, err
		}
		var res struct {
			StructuredOutput *verdict `json:"structured_output"`
		}
		if err := json.Unmarshal(out, &res); err != nil {
			return verdict{}, fmt.Errorf("unreadable judge output: %w", err)
		}
		if res.StructuredOutput == nil {
			return verdict{}, errors.New("no structured output")
		}
		if !slices.Contains([]string{"A", "B", "tie"}, res.StructuredOutput.Winner) {
			return verdict{}, fmt.Errorf("winner %q is not A, B, or tie", res.StructuredOutput.Winner)
		}
		return *res.StructuredOutput, nil
	}
}

// judgePairs compares pair[0] with pair[1] on every task and repetition
// pair[0] ran. A pair whose runs are not both valid, or whose judge call
// fails, is recorded as skipped with its reason; the run never fails on the
// judge.
func judgePairs(rep report, abDir string, pair [2]string, judge judgeFunc, parallel int) []judgement {
	byKey := map[string]result{}
	for _, r := range rep.Results {
		byKey[fmt.Sprintf("%s/%s/%d", r.Arm, r.Task, r.Rep)] = r
	}
	var out []judgement
	for _, r := range rep.Results {
		if r.Arm == pair[0] {
			out = append(out, judgement{Task: r.Task, Rep: r.Rep, Pair: pair})
		}
	}
	forEach(parallel, len(out), func(i int) {
		j := &out[i]
		runs := [2]result{byKey[fmt.Sprintf("%s/%s/%d", pair[0], j.Task, j.Rep)], byKey[fmt.Sprintf("%s/%s/%d", pair[1], j.Task, j.Rep)]}
		for k, r := range runs {
			if r.Arm == "" || resultStatus(r) != "ok " || r.Source == nil {
				j.Skipped = fmt.Sprintf("skipped (%s run is not valid)", pair[k])
				return
			}
		}
		original, err := productionSources(filepath.Join(abDir, j.Task))
		if err != nil {
			j.Skipped = fmt.Sprintf("skipped (%v)", err)
			return
		}
		diffs := [2]string{sourceDiff(original, runs[0].Source), sourceDiff(original, runs[1].Source)}
		for k := range 2 {
			a, b := k, 1-k // order 0 shows pair[0] as A; order 1 shows pair[1] as A
			v, err := judge(fmt.Sprintf(judgeRubric, diffs[a], diffs[b]))
			if err != nil {
				j.Skipped = fmt.Sprintf("skipped (%v)", err)
				return
			}
			j.Reasons[k] = v.Reason
			switch v.Winner {
			case "A":
				j.Verdicts[k] = pair[a]
			case "B":
				j.Verdicts[k] = pair[b]
			default:
				j.Verdicts[k] = "tie"
			}
		}
		j.Disagree = j.Verdicts[0] != j.Verdicts[1]
		if !j.Disagree && j.Verdicts[0] != "tie" {
			j.Preferred = j.Verdicts[0]
		}
	})
	return out
}

// productionSources reads every non-test Go file under dir, keyed by its
// slash-separated path relative to dir.
func productionSources(dir string) (map[string]string, error) {
	fsys := os.DirFS(dir)
	sources := map[string]string{}
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, err := fs.ReadFile(fsys, path)
		sources[path] = string(data)
		return err
	})
	return sources, err
}

// sourceDiff renders every changed file as a whole-file line diff: " " kept,
// "-" removed, "+" added. Fixture files are small, so there are no hunks.
func sourceDiff(before, after map[string]string) string {
	names := slices.Sorted(maps.Keys(before))
	for name := range after {
		if _, ok := before[name]; !ok {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	var b strings.Builder
	for _, name := range names {
		if before[name] == after[name] {
			continue
		}
		fmt.Fprintf(&b, "--- %s\n", name)
		for _, line := range lineDiff(splitLines(before[name]), splitLines(after[name])) {
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

// lineDiff is a longest-common-subsequence diff of a and b.
func lineDiff(a, b []string) []string {
	lcs := make([][]int, len(a)+1)
	for i := range lcs {
		lcs[i] = make([]int, len(b)+1)
	}
	for i, x := range slices.Backward(a) {
		for j, y := range slices.Backward(b) {
			if x == y {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	var out []string
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] == b[j]:
			out = append(out, " "+a[i])
			i, j = i+1, j+1
		case lcs[i+1][j] >= lcs[i][j+1]:
			out = append(out, "-"+a[i])
			i++
		default:
			out = append(out, "+"+b[j])
			j++
		}
	}
	for ; i < len(a); i++ {
		out = append(out, "-"+a[i])
	}
	for ; j < len(b); j++ {
		out = append(out, "+"+b[j])
	}
	return out
}

// printJudgeSummary prints the preferences both orders agreed on, the ties,
// and how often the verdict followed the position instead of the code.
func printJudgeSummary(js []judgement, pair [2]string) {
	wins := map[string]int{}
	ties, disagree, skipped := 0, 0, 0
	for _, j := range js {
		switch {
		case j.Skipped != "":
			skipped++
		case j.Preferred != "":
			wins[j.Preferred]++
		default:
			ties++
			if j.Disagree {
				disagree++
			}
		}
	}
	fmt.Printf("\njudge %s vs %s: %s preferred %d, %s preferred %d, ties %d; positional disagreement %d/%d; skipped %d\n",
		pair[0], pair[1], pair[0], wins[pair[0]], pair[1], wins[pair[1]], ties, disagree, len(js)-skipped, skipped)
}

// judgePair parses -judge-pair and checks both arms will run. The review
// corpus has no production diff to compare.
func judgePair(o options, arms []arm) ([2]string, error) {
	if o.corpus == corpusReview {
		return [2]string{}, exitError{2, "-judge compares production code; the review corpus leaves none"}
	}
	x, y, ok := strings.Cut(o.judgePair, ",")
	if !ok || x == "" || y == "" || x == y {
		return [2]string{}, exitError{2, fmt.Sprintf("-judge-pair %q: want two different arms, e.g. reference,baseline", o.judgePair)}
	}
	for _, name := range []string{x, y} {
		if !slices.ContainsFunc(arms, func(a arm) bool { return a.Name == name }) {
			return [2]string{}, exitError{2, fmt.Sprintf("-judge-pair names arm %q, which this run does not have", name)}
		}
	}
	return [2]string{x, y}, nil
}
