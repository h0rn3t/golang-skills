package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestCommandOutputReturnsPartialOutputAndError(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess", "--", "partial-error")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")

	got, err := commandOutput(cmd)
	if err == nil {
		t.Fatal("commandOutput(partial-error) error = nil, want non-nil")
	}
	if string(got) != "partial output\n" {
		t.Fatalf("commandOutput(partial-error) output = %q, want %q", got, "partial output\n")
	}
}

func TestResultStatus(t *testing.T) {
	tests := []struct {
		name string
		in   result
		want string
	}{
		{name: "success", in: result{Build: true, Golden: true, Edited: true}, want: "ok "},
		{name: "runner error", in: result{Build: true, Golden: true, Edited: true, Err: "failed"}, want: "ERR"},
		{name: "build failure", in: result{Golden: true, Edited: true}, want: "ERR"},
		{name: "golden failure", in: result{Build: true, Edited: true}, want: "ERR"},
		{name: "no edit", in: result{Build: true, Golden: true}, want: "ERR"},
		{name: "repository reference", in: result{Build: true, Golden: true, Edited: true, Leaked: true}, want: "ERR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resultStatus(tt.in); got != tt.want {
				t.Errorf("resultStatus(%+v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSummarizeArmExcludesInvalidDeltas(t *testing.T) {
	rep := report{Results: []result{
		{Arm: "baseline", Build: true, Golden: true, Edited: true, Delta: metrics{Lines: -10}},
		{Arm: "baseline", Build: true, Golden: false, Edited: true, Delta: metrics{Lines: -100}},
		{Arm: "baseline", Err: "session failed", Delta: metrics{Lines: -200}},
		// A session that edited nothing scores a zero delta on every metric,
		// which would read as a behavior-preserving tie rather than a miss.
		{Arm: "baseline", Build: true, Golden: true},
		{Arm: "baseline", Build: true, Golden: true, Edited: true, Leaked: true, Delta: metrics{Lines: -400}},
	}}

	got := summarizeArm(rep, "baseline")
	if got.Runs != 5 || got.Errors != 1 || got.Valid != 1 {
		t.Fatalf("summarizeArm counts = runs:%d errors:%d valid:%d, want 5, 1, 1", got.Runs, got.Errors, got.Valid)
	}
	if got.NoEdit != 1 || got.Leaked != 1 {
		t.Errorf("summarizeArm counts = noedit:%d leaked:%d, want 1, 1", got.NoEdit, got.Leaked)
	}
	if got.Lines != -10 {
		t.Errorf("summarizeArm lines = %d, want -10 from the valid run only", got.Lines)
	}
}

func TestAnalyzeIncludesNestedGoFiles(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "fixture.go"), "package fixture\n\ntype Local struct{}\nfunc Top() {}\n")
	writeTestFile(t, filepath.Join(root, "nested", "nested.go"), "package nested\ntype Service interface{ Run() }\nfunc Use() {}")
	writeTestFile(t, filepath.Join(root, "nested", "nested_test.go"), "package nested\n")

	got, err := analyze(root)
	if err != nil {
		t.Fatalf("analyze(%q) error = %v, want nil", root, err)
	}
	want := metrics{Lines: 7, Code: 6, Tokens: 29, Files: 2, TestFiles: 1, Types: 2, Interfaces: 1, Funcs: 2, Exported: 4}
	if got != want {
		t.Errorf("analyze(%q) = %+v, want %+v", root, got, want)
	}
}

// TestCodeSize pins the count a deleted comment cannot move. It is the half of
// the gate stage 3 was missing: under the physical count alone, two sessions
// bought new helpers by deleting the paragraph that justified the server's
// timeouts, and both trades passed.
func TestCodeSize(t *testing.T) {
	for _, tt := range []struct {
		name       string
		src        string
		code, toks int
	}{
		{name: "blank and comment-only lines hold no code", src: "package p\n\n// a\n//\n// b\nvar X = 1\n", code: 2, toks: 6},
		{name: "code with a trailing comment counts once", src: "package p\n\nvar X = 1 // why\n", code: 2, toks: 6},
		{name: "a block comment counts only where code shares the line",
			src: "package p\n\n/*\nprose\n*/\nvar X = 1 /* here */\n", code: 2, toks: 6},
		{name: "a multi-line string is code on every line it spans",
			src: "package p\n\nvar X = `one\ntwo\nthree`\n", code: 4, toks: 6},
		{name: "an empty file holds none", src: "", code: 0, toks: 0},
		{name: "a last line without a newline still counts", src: "package p\n\nvar X = 1", code: 2, toks: 6},
	} {
		t.Run(tt.name, func(t *testing.T) {
			code, toks := codeSize("p.go", []byte(tt.src))
			if code != tt.code || toks != tt.toks {
				t.Errorf("codeSize(%q) = %d code, %d tokens, want %d and %d", tt.src, code, toks, tt.code, tt.toks)
			}
		})
	}
}

// TestCodeSizeTokensIgnoreLineShape is the finding the gw2 smoke produced. A
// session met both line gates by packing an eight-field struct literal onto one
// 201-character line and spending the seven lines on helpers, and gofmt left
// the result alone. Both line counts move under that reflow and the token count
// must not, or the harness cannot tell a shorter package from a rearranged one.
func TestCodeSizeTokensIgnoreLineShape(t *testing.T) {
	const spread = `package p

var X = T{
	A: 1,
	B: 2,
	C: 3,
}
`
	const packed = "package p\n\nvar X = T{A: 1, B: 2, C: 3}\n"

	spreadCode, spreadToks := codeSize("p.go", []byte(spread))
	packedCode, packedToks := codeSize("p.go", []byte(packed))
	if spreadToks != packedToks {
		t.Errorf("tokens = %d spread, %d packed; reflow must not move the count", spreadToks, packedToks)
	}
	if spreadCode <= packedCode {
		t.Errorf("code = %d spread, %d packed; the line count is what reflow moves", spreadCode, packedCode)
	}
}

func TestAnalyzeCountsPublicSurfaceAndBranches(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "surface.go"), `package surface

const Public = 1

const private = 2

var Exported, unexported = 3, 4

type Kind int

type hidden struct{}

// Method is exported but hangs off an unexported type, so it adds no
// reachable surface.
func (hidden) Method() {}

func Walk(items []int) int {
	total := 0
	for _, n := range items {
		if n < 0 {
			continue
		}
		switch n {
		case 1, 2:
			total += n
		default:
			total++
		}
	}
	return total
}
`)

	got, err := analyze(root)
	if err != nil {
		t.Fatalf("analyze(%q) error = %v, want nil", root, err)
	}

	// Public, Exported, Kind, Walk. private, unexported, hidden and the method
	// on hidden are all unreachable from outside the package.
	if got.Exported != 4 {
		t.Errorf("analyze exported = %d, want 4", got.Exported)
	}
	// range, if, and the two values of the one non-default case.
	if got.Branches != 4 {
		t.Errorf("analyze branches = %d, want 4", got.Branches)
	}
}

func TestHideTestFilesRecurses(t *testing.T) {
	root := t.TempDir()
	paths := []string{
		filepath.Join(root, "root_test.go"),
		filepath.Join(root, "nested", "nested_test.go"),
	}
	for _, path := range paths {
		writeTestFile(t, path, "package fixture\n")
	}

	if err := hideTestFiles(root); err != nil {
		t.Fatalf("hideTestFiles(%q) error = %v, want nil", root, err)
	}
	for _, path := range paths {
		if _, err := os.Stat(path + ".model"); err != nil {
			t.Errorf("os.Stat(%q) error = %v, want renamed test file", path+".model", err)
		}
	}
}

func TestValidateFixturesRequiresGoldenGoFile(t *testing.T) {
	abDir := t.TempDir()
	writeTestFile(t, filepath.Join(abDir, "task", "task.go"), "package task\n")

	if err := validateFixtures(abDir, []string{"task"}); err == nil {
		t.Fatal("validateFixtures(task without golden) error = nil, want non-nil")
	}
	writeTestFile(t, filepath.Join(abDir, "_golden", "task", "README.md"), "not a test\n")
	if err := validateFixtures(abDir, []string{"task"}); err == nil {
		t.Fatal("validateFixtures(task without golden Go file) error = nil, want non-nil")
	}
	writeTestFile(t, filepath.Join(abDir, "_golden", "task", "helper.go"), "package task\n")
	if err := validateFixtures(abDir, []string{"task"}); err == nil {
		t.Fatal("validateFixtures(task without golden test) error = nil, want non-nil")
	}
	writeTestFile(t, filepath.Join(abDir, "_golden", "task", "golden_test.go"), "package task\n")
	if err := validateFixtures(abDir, []string{"task"}); err != nil {
		t.Fatalf("validateFixtures(task with golden) error = %v, want nil", err)
	}
}

func TestFindTasksRejectsUnknownSelection(t *testing.T) {
	abDir := t.TempDir()
	writeTestFile(t, filepath.Join(abDir, "known", "known.go"), "package known\n")

	if _, err := findTasks(abDir, "known,missing"); err == nil {
		t.Fatal("findTasks(known,missing) error = nil, want unknown-task error")
	}
}

func TestBuildArmsUsesReferenceAndCurrentRoots(t *testing.T) {
	current := t.TempDir()
	reference := t.TempDir()
	variants := t.TempDir()
	const skillPath = "skills/go-code-refactor/SKILL.md"
	writeTestFile(t, filepath.Join(current, skillPath), "current\n\n## Workflow\n")
	writeTestFile(t, filepath.Join(reference, skillPath), "reference\n\n## Workflow\n")

	arms, cleanup, err := buildArms(current, reference, variants, "reference,baseline")
	t.Cleanup(cleanup)
	if err != nil {
		t.Fatalf("buildArms(reference, baseline) error = %v, want nil", err)
	}
	if len(arms) != 2 {
		t.Fatalf("buildArms(reference, baseline) returned %d arms, want 2", len(arms))
	}

	want := map[string]string{"reference": "reference", "baseline": "current"}
	digests := map[string]string{}
	for _, arm := range arms {
		data, err := os.ReadFile(filepath.Join(arm.dir, skillPath))
		if err != nil {
			t.Fatalf("os.ReadFile(%s arm skill) error = %v", arm.Name, err)
		}
		if string(data) != want[arm.Name]+"\n\n## Workflow\n" {
			t.Errorf("%s arm skill = %q, want source marker %q", arm.Name, data, want[arm.Name])
		}
		if arm.Digest == "" {
			t.Errorf("%s arm digest is empty", arm.Name)
		}
		digests[arm.Name] = arm.Digest
	}
	if digests["reference"] == digests["baseline"] {
		t.Errorf("reference digest = baseline digest = %q, want distinct plugin content", digests["reference"])
	}
}

func TestCheckFrontmatter(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool // true when the file is expected to be rejected
	}{
		{"em-dash description", "---\nname: go-code\ndescription: Use when writing Go — it routes.\n---\n", false},
		{"unquoted colon-space", "---\nname: go-code\ndescription: Use when writing Go: it routes.\n---\n", true},
		{"double-quoted colon-space", "---\nname: go-code\ndescription: \"Use when writing Go: it routes.\"\n---\n", false},
		{"single-quoted colon-space", "---\nname: go-code\ndescription: 'Use when writing Go: it routes.'\n---\n", false},
		{"colon without a space", "---\nname: go-code\ndescription: Use /opsx:apply with it.\n---\n", false},
		{"indented continuation", "---\nname: go-code\nallowed-tools:\n  - Read: all\n---\n", false},
		{"no frontmatter", "# Go Code\n\nA skill with no frontmatter at all.\n", false},
		{"unclosed frontmatter", "---\nname: go-code\n", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkFrontmatter(tt.text)
			if (err != nil) != tt.want {
				t.Errorf("checkFrontmatter() error = %v, want rejected = %v", err, tt.want)
			}
		})
	}
}

func TestCheckArmSkillsRequiresEverySkillInTheTree(t *testing.T) {
	armDir := t.TempDir()
	for _, name := range []string{"go-code", "go-http", "go-testing"} {
		writeTestFile(t, filepath.Join(armDir, "skills", name, "SKILL.md"), "---\nname: "+name+"\n---\n")
	}

	if err := checkArmSkills("baseline", armDir, []string{"go-code", "go-http", "go-testing"}, ""); err != nil {
		t.Errorf("checkArmSkills(complete listing) error = %v, want nil", err)
	}

	// The bug this guards: one unloadable skill in a full tree used to pass,
	// because the check only rejected an empty listing.
	err := checkArmSkills("baseline", armDir, []string{"go-http", "go-testing"}, "failed to parse YAML frontmatter")
	if err == nil {
		t.Fatal("checkArmSkills(missing go-code) error = nil, want a rejection")
	}
	if !strings.Contains(err.Error(), "go-code") {
		t.Errorf("checkArmSkills(missing go-code) error = %q, want the missing skill named", err)
	}
	if !strings.Contains(err.Error(), "failed to parse YAML frontmatter") {
		t.Errorf("checkArmSkills(missing go-code) error = %q, want the CLI's own reason quoted", err)
	}

	if err := checkArmSkills(controlArm, "", nil, ""); err != nil {
		t.Errorf("checkArmSkills(control, empty) error = %v, want nil", err)
	}
	if err := checkArmSkills(controlArm, "", []string{"go-code"}, ""); err == nil {
		t.Error("checkArmSkills(control, leaked skill) error = nil, want a rejection")
	}
}

func TestBuildJobsIsSeededAndNotArmMajor(t *testing.T) {
	arms := []arm{{Name: "a"}, {Name: "b"}}
	tasks := []string{"one", "two"}

	got := buildJobs(arms, tasks, 3, 7)
	again := buildJobs(arms, tasks, 3, 7)
	if !reflect.DeepEqual(got, again) {
		t.Fatalf("buildJobs(seed 7) = %#v, second run = %#v", got, again)
	}

	var armMajor []job
	for _, arm := range arms {
		for _, task := range tasks {
			for rep := range 3 {
				armMajor = append(armMajor, job{arm: arm, task: task, rep: rep})
			}
		}
	}
	if reflect.DeepEqual(got, armMajor) {
		t.Fatalf("buildJobs(seed 7) retained arm-major order: %#v", got)
	}
}

func TestValidateOptions(t *testing.T) {
	valid := options{corpus: corpusRefactor, runner: runnerClaude, reps: 1, parallel: 1, timeout: time.Second}
	tests := []struct {
		name string
		in   options
	}{
		{name: "zero repetitions", in: options{corpus: corpusRefactor, runner: runnerClaude, parallel: 1, timeout: time.Second}},
		{name: "zero parallelism", in: options{corpus: corpusRefactor, runner: runnerClaude, reps: 1, timeout: time.Second}},
		{name: "zero timeout", in: options{corpus: corpusRefactor, runner: runnerClaude, reps: 1, parallel: 1}},
		{name: "unknown runner", in: options{corpus: corpusRefactor, runner: "cursor", reps: 1, parallel: 1, timeout: time.Second}},
		{name: "opencode without a model", in: options{corpus: corpusRefactor, runner: runnerOpencode, reps: 1, parallel: 1, timeout: time.Second}},
	}

	if err := validateOptions(valid); err != nil {
		t.Fatalf("validateOptions(valid) error = %v, want nil", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateOptions(tt.in); err == nil {
				t.Errorf("validateOptions(%+v) error = nil, want non-nil", tt.in)
			}
		})
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args
	i := slices.Index(args, "--")
	if i < 0 || i+1 >= len(args) {
		os.Exit(2)
	}

	switch args[i+1] {
	case "partial-error":
		if _, err := fmt.Fprintln(os.Stdout, "partial output"); err != nil {
			os.Exit(3)
		}
		if _, err := fmt.Fprintln(os.Stderr, "partial failure"); err != nil {
			os.Exit(3)
		}
		os.Exit(7)
	default:
		os.Exit(2)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", path, err)
	}
}
