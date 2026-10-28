package main

import (
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang-skills/evals/internal/skillreads"
)

// TestMergeSkillReads sums a claude run's skill-file reads over its turns: a
// repair turn's transcript is parsed on its own and added to the first one's.
func TestMergeSkillReads(t *testing.T) {
	first := strings.Join([]string{
		`{"type":"assistant","message":{"id":"m1","content":[{"type":"tool_use","id":"toolu_1","name":"Skill","input":{"skill":"golang-skills:go-code"}}]}}`,
		`{"type":"assistant","message":{"id":"m2","content":[{"type":"tool_use","id":"toolu_2","name":"Read","input":{"file_path":"/w/.eval-plugin-1/skills/go-code/references/NEW-CODE-EXAMPLES.md"}}]}}`,
		`{"type":"assistant","message":{"id":"m2","content":[{"type":"tool_use","id":"toolu_3","name":"Read","input":{"file_path":"/w/.eval-plugin-1/skills/go-testing/SKILL.md"}}]}}`,
	}, "\n")
	repair := `{"type":"assistant","message":{"id":"m9","content":[{"type":"tool_use","id":"toolu_9","name":"Read","input":{"file_path":"/w/.eval-plugin-1/skills/go-code/references/NEW-CODE-EXAMPLES.md"}}]}}`

	var r result
	r.merge(sessionTurn{skillReads: skillreads.Count([]byte(first))})
	r.merge(sessionTurn{skillReads: skillreads.Count([]byte(repair))})
	r.merge(sessionTurn{final: "no reads in this turn"})
	want := map[string]int{"go-code/references/NEW-CODE-EXAMPLES.md": 2, "go-testing/SKILL.md": 1}
	if !maps.Equal(r.SkillReads, want) {
		t.Errorf("merged SkillReads = %v, want %v", r.SkillReads, want)
	}

	var none result
	none.merge(sessionTurn{final: "done"})
	if none.SkillReads != nil {
		t.Errorf("SkillReads after a turn with no reads = %v, want nil", none.SkillReads)
	}
}

// TestArmSkillReads totals reads per arm over every run, errored ones
// included, and lists each file of the arm's tree, so an unread reference is
// a zero rather than an absence. The control arm has no tree and no entry.
func TestArmSkillReads(t *testing.T) {
	baseline := t.TempDir()
	for _, rel := range []string{
		"skills/go-code/SKILL.md",
		"skills/go-code/references/INTENSITY.md",
		"skills/go-code/references/NEW-CODE-EXAMPLES.md",
		"skills/go-testing/SKILL.md",
	} {
		p := filepath.Join(baseline, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	arms := []arm{{Name: controlArm}, {Name: "baseline", dir: baseline}}
	rep := report{
		Corpus: corpusImplement,
		Runner: runnerClaude,
		Arms:   arms,
		Results: []result{
			{Arm: "baseline", Build: true, Golden: true, Edited: true, SkillReads: map[string]int{"go-code/references/NEW-CODE-EXAMPLES.md": 2}},
			{Arm: "baseline", Err: "timed out", SkillReads: map[string]int{"go-code/references/NEW-CODE-EXAMPLES.md": 1, "go-testing/SKILL.md": 1}},
			{Arm: controlArm, Build: true, Golden: true, Edited: true},
		},
	}
	rep.SkillReads = armSkillReads(arms, rep.Results)

	want := map[string]map[string]int{"baseline": {
		"go-code/SKILL.md":                        0,
		"go-code/references/INTENSITY.md":         0,
		"go-code/references/NEW-CODE-EXAMPLES.md": 3,
		"go-testing/SKILL.md":                     1,
	}}
	if !maps.EqualFunc(rep.SkillReads, want, maps.Equal) {
		t.Errorf("armSkillReads(arms, results) = %v, want %v", rep.SkillReads, want)
	}

	out := captureRoutingSummary(t, rep)
	if line := "skill file reads: 4 call(s); SKILL.md read 1/2, references read 1/2"; strings.Count(out, line) != 1 {
		t.Errorf("summary carries %q %d times, want once (baseline only):\n%s", line, strings.Count(out, line), out)
	}
}
