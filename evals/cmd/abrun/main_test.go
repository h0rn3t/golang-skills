package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
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
		{name: "success", in: result{Build: true, Golden: true}, want: "ok "},
		{name: "runner error", in: result{Build: true, Golden: true, Err: "failed"}, want: "ERR"},
		{name: "build failure", in: result{Golden: true}, want: "ERR"},
		{name: "golden failure", in: result{Build: true}, want: "ERR"},
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
		{Arm: "baseline", Build: true, Golden: true, Delta: metrics{Lines: -10}},
		{Arm: "baseline", Build: true, Golden: false, Delta: metrics{Lines: -100}},
		{Arm: "baseline", Err: "session failed", Delta: metrics{Lines: -200}},
	}}

	got := summarizeArm(rep, "baseline")
	if got.Runs != 3 || got.Errors != 1 || got.Valid != 1 {
		t.Fatalf("summarizeArm counts = runs:%d errors:%d valid:%d, want 3, 1, 1", got.Runs, got.Errors, got.Valid)
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
	want := metrics{Lines: 7, Files: 2, TestFiles: 1, Types: 2, Interfaces: 1, Funcs: 2}
	if got != want {
		t.Errorf("analyze(%q) = %+v, want %+v", root, got, want)
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
	valid := options{reps: 1, parallel: 1, timeout: time.Second}
	tests := []struct {
		name string
		in   options
	}{
		{name: "zero repetitions", in: options{parallel: 1, timeout: time.Second}},
		{name: "zero parallelism", in: options{reps: 1, timeout: time.Second}},
		{name: "zero timeout", in: options{reps: 1, parallel: 1}},
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
