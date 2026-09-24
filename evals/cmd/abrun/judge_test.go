package main

import (
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// judgeReport is one task where reference and baseline both left valid code:
// reference writes "return 1", baseline "return 2".
func judgeReport(t *testing.T) (report, string) {
	t.Helper()
	abDir := t.TempDir()
	writeTestFile(t, filepath.Join(abDir, "feed", "feed.go"), "package feed\n\nfunc F() int { panic(0) }\n")
	run := func(armName, body string) result {
		return result{Arm: armName, Task: "feed", Build: true, Golden: true, Edited: true,
			Source: map[string]string{"feed.go": "package feed\n\nfunc F() int { " + body + " }\n"}}
	}
	return report{Results: []result{run("reference", "return 1"), run("baseline", "return 2")}}, abDir
}

// fakeJudge prefers the change containing want; with position set it always
// answers A, whatever the code.
func fakeJudge(want string, position bool, prompts *[]string) judgeFunc {
	var mu sync.Mutex
	return func(prompt string) (verdict, error) {
		mu.Lock()
		*prompts = append(*prompts, prompt)
		mu.Unlock()
		if position {
			return verdict{Winner: "A"}, nil
		}
		a, _, _ := strings.Cut(prompt, "CHANGE B:")
		if strings.Contains(a, want) {
			return verdict{Winner: "A", Reason: "plainer"}, nil
		}
		return verdict{Winner: "B", Reason: "plainer"}, nil
	}
}

func TestJudgeBothOrdersAgree(t *testing.T) {
	rep, abDir := judgeReport(t)
	var prompts []string
	got := judgePairs(rep, abDir, [2]string{"reference", "baseline"}, fakeJudge("+func F() int { return 2 }", false, &prompts), 1)
	if len(got) != 1 || got[0].Preferred != "baseline" || got[0].Disagree || got[0].Skipped != "" {
		t.Fatalf("judgePairs(judge prefers baseline) = %+v, want one baseline preference", got)
	}
	if len(prompts) != 2 {
		t.Fatalf("judge calls = %d, want 2 (both orders)", len(prompts))
	}
	first, _, _ := strings.Cut(prompts[0], "CHANGE B:")
	second, _, _ := strings.Cut(prompts[1], "CHANGE B:")
	if !strings.Contains(first, "return 1") || !strings.Contains(second, "return 2") {
		t.Errorf("orders did not swap: A in first prompt %q, in second %q", first, second)
	}
	for _, p := range prompts {
		if strings.Contains(p, "reference") || strings.Contains(p, "baseline") {
			t.Errorf("judge prompt names an arm: %q", p)
		}
	}
}

func TestJudgePositionalDisagreementIsATie(t *testing.T) {
	rep, abDir := judgeReport(t)
	var prompts []string
	got := judgePairs(rep, abDir, [2]string{"reference", "baseline"}, fakeJudge("", true, &prompts), 1)
	if len(got) != 1 || got[0].Preferred != "" || !got[0].Disagree || got[0].Verdicts != [2]string{"reference", "baseline"} {
		t.Fatalf("judgePairs(judge always answers A) = %+v, want a tie with positional disagreement", got)
	}
}

func TestJudgeFailureIsSkipped(t *testing.T) {
	rep, abDir := judgeReport(t)
	failing := func(string) (verdict, error) { return verdict{}, errors.New("no structured output") }
	got := judgePairs(rep, abDir, [2]string{"reference", "baseline"}, failing, 1)
	if len(got) != 1 || got[0].Skipped != "skipped (no structured output)" {
		t.Fatalf("judgePairs(failing judge) = %+v, want skipped (no structured output)", got)
	}

	rep.Results[1].Golden = false
	var prompts []string
	got = judgePairs(rep, abDir, [2]string{"reference", "baseline"}, fakeJudge("", false, &prompts), 1)
	if got[0].Skipped != "skipped (baseline run is not valid)" || len(prompts) != 0 {
		t.Fatalf("judgePairs(invalid baseline) = %+v after %d calls, want skipped with no call", got, len(prompts))
	}
}

func TestSourceDiff(t *testing.T) {
	before := map[string]string{"a.go": "one\ntwo\nthree\n", "same.go": "x\n"}
	after := map[string]string{"a.go": "one\n2\nthree\n", "same.go": "x\n", "new.go": "n\n"}
	want := "--- a.go\n one\n-two\n+2\n three\n--- new.go\n+n\n"
	if got := sourceDiff(before, after); got != want {
		t.Errorf("sourceDiff() =\n%s\nwant\n%s", got, want)
	}
}
