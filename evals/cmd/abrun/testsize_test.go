package main

import (
	"path/filepath"
	"testing"
)

// TestAnalyzeCountsModelTests pins what the test-size columns see: a table
// counts its rows whether declared in the test, at package level, or in a
// sibling test file, a t.Run outside a loop is one case, a test with neither
// is one case, a slice built as input and never ranged over is not a table,
// and a helper is not a test.
func TestAnalyzeCountsModelTests(t *testing.T) {
	tests := []struct {
		name                 string
		src                  string
		sibling              string
		wantFuncs, wantCases int
	}{
		{
			name: "local table",
			src: `func TestA(t *testing.T) {
	tests := []struct{ in, want string }{{"a", "A"}, {"b", "B"}, {"", ""}}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {})
	}
}`,
			wantFuncs: 1, wantCases: 3,
		},
		{
			name: "package map table",
			src: `var cases = map[string]int{"one": 1, "two": 2}

func TestB(t *testing.T) {
	for name := range cases {
		_ = name
	}
}`,
			wantFuncs: 1, wantCases: 2,
		},
		{
			name: "loose subtests",
			src: `func TestC(t *testing.T) {
	t.Run("empty", func(t *testing.T) {})
	t.Run("nil", func(t *testing.T) {})
}`,
			wantFuncs: 1, wantCases: 2,
		},
		{
			name: "plain tests, input slice and helper",
			src: `func TestD(t *testing.T) {
	items := []int{1, 2, 3}
	_ = items
}

func TestE(t *testing.T) {}

func helper(t *testing.T) {}`,
			wantFuncs: 2, wantCases: 2,
		},
		{
			name: "tables share a name",
			src: `func TestF(t *testing.T) {
	tests := []string{"a", "b"}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {})
	}
}

func TestG(t *testing.T) {
	tests := []string{"a", "b", "c"}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {})
	}
}`,
			wantFuncs: 2, wantCases: 5,
		},
		{
			name: "grouped subtests",
			src: `func TestH(t *testing.T) {
	t.Run("group", func(t *testing.T) {
		t.Run("a", func(t *testing.T) {})
		t.Run("b", func(t *testing.T) {})
	})
}`,
			wantFuncs: 1, wantCases: 2,
		},
		{
			name: "TestMain, Testable helper and a Run that is not a subtest",
			src: `func TestMain(m *testing.M) { m.Run() }

func Testable(t *testing.T) {}

func TestI(t *testing.T) {
	cmd := exec.Command("true")
	_ = cmd.Run()
	t.Run("one", func(t *testing.T) {})
}`,
			wantFuncs: 1, wantCases: 1,
		},
		{
			name: "table in a sibling file",
			src: `func TestJ(t *testing.T) {
	for _, tt := range shared {
		t.Run(tt, func(t *testing.T) {})
	}
}`,
			sibling:   `var shared = []string{"a", "b", "c"}`,
			wantFuncs: 1, wantCases: 3,
		},
		{
			name: "named slice type",
			src: `type cases []string

func TestK(t *testing.T) {
	tests := cases{"a", "b"}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {})
	}
}`,
			wantFuncs: 1, wantCases: 2,
		},
		{
			name: "one name, a table per group",
			src: `func TestL(t *testing.T) {
	t.Run("long", func(t *testing.T) {
		tests := []string{"a", "b", "c", "d"}
		for _, tt := range tests {
			t.Run(tt, func(t *testing.T) {})
		}
	})
	t.Run("short", func(t *testing.T) {
		tests := []string{"a"}
		for _, tt := range tests {
			t.Run(tt, func(t *testing.T) {})
		}
	})
}`,
			wantFuncs: 1, wantCases: 5,
		},
		{
			name: "subtest over a table",
			src: `func TestM(t *testing.T) {
	t.Run("all", func(t *testing.T) {
		for _, in := range []int{1, 2, 3} {
			_ = in
		}
	})
}`,
			wantFuncs: 1, wantCases: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeTestFile(t, filepath.Join(dir, "task_test.go"), "package task\n\nimport \"testing\"\n\n"+tt.src+"\n")
			if tt.sibling != "" {
				writeTestFile(t, filepath.Join(dir, "shared_test.go"), "package task\n\n"+tt.sibling+"\n")
			}

			got, err := analyze(dir)
			if err != nil {
				t.Fatalf("analyze() error = %v, want nil", err)
			}
			if *got.TestFuncs != tt.wantFuncs || *got.TestCases != tt.wantCases {
				t.Errorf("analyze() test funcs, cases = %d, %d, want %d, %d", *got.TestFuncs, *got.TestCases, tt.wantFuncs, tt.wantCases)
			}
		})
	}
}

// TestTestSizeSumLine averages only the runs that carry all three fields and
// prints a dash for a report written before they existed.
func TestTestSizeSumLine(t *testing.T) {
	var old testSizeSum
	old.add(metrics{Lines: -3})
	if got, want := old.line(1), "tests Δ/run: — (not measured in this report)"; got != want {
		t.Errorf("line() for a report without test fields = %q, want %q", got, want)
	}

	var s testSizeSum
	s.add(metrics{TestLines: new(10), TestFuncs: new(1), TestCases: new(4)})
	s.add(metrics{TestLines: new(20), TestFuncs: new(2), TestCases: new(2)})
	s.add(metrics{TestLines: new(99)})
	want := "tests Δ/run: lines 15.0, Test funcs 1.50, cases 3.00 (2/3 runs)"
	if got := s.line(3); got != want {
		t.Errorf("line(3) = %q, want %q", got, want)
	}
}

// TestAnalyzeCountsUnparsableTestLines keeps the line count of a test file the
// session left broken, so the per-run row still shows what was written.
func TestAnalyzeCountsUnparsableTestLines(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "task_test.go"), "package task\n\nfunc TestX(t *testing.T) {\n")

	got, err := analyze(dir)
	if err != nil {
		t.Fatalf("analyze() error = %v, want nil", err)
	}
	if *got.TestLines != 3 || *got.TestFuncs != 0 {
		t.Errorf("analyze() test lines, funcs = %d, %d, want 3, 0", *got.TestLines, *got.TestFuncs)
	}
}
