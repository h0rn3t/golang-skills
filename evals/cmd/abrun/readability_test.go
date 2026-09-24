package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// function returns a declaration that spans exactly lines lines.
func function(name string, lines int) string {
	return "func " + name + "() {\n" + strings.Repeat("\t_ = 0\n", lines-2) + "}\n"
}

func analyzeSource(t *testing.T, files map[string]string) metrics {
	t.Helper()
	root := t.TempDir()
	for name, src := range files {
		writeTestFile(t, filepath.Join(root, name), src)
	}
	got, err := analyze(root)
	if err != nil {
		t.Fatalf("analyze(%v) error = %v, want nil", files, err)
	}
	return got
}

func TestReadabilityFunctionShape(t *testing.T) {
	nested := "func deep(xs []int) {\n\tfor range xs {\n\t\tif true {\n\t\t\tswitch {\n\t\t\tcase true:\n\t\t\t\tfunc() {}()\n\t\t\t}\n\t\t}\n\t}\n" +
		strings.Repeat("\t_ = 0\n", 60-10) + "}\n"
	src := "package p\n\n" + nested + function("a", 10) + function("b", 10)
	got := analyzeSource(t, map[string]string{
		"p.go":      src,
		"p_test.go": "package p\n\n" + function("longTest", 200),
	})
	if *got.MaxFuncLines != 60 || *got.P90FuncLines != 60 || *got.MaxNesting != 4 {
		t.Errorf("analyze(60-line func nested 4, two 10-line funcs) = max %d, p90 %d, nesting %d; want 60, 60, 4",
			*got.MaxFuncLines, *got.P90FuncLines, *got.MaxNesting)
	}
}

func TestReadabilityNestingElseIfChain(t *testing.T) {
	src := "package p\n\nfunc f(n int) int {\n\tif n == 1 {\n\t\treturn 1\n\t} else if n == 2 {\n\t\treturn 2\n\t} else if n == 3 {\n\t\treturn 3\n\t} else {\n\t\treturn 0\n\t}\n}\n"
	if got := analyzeSource(t, map[string]string{"p.go": src}); *got.MaxNesting != 1 {
		t.Errorf("analyze(else-if chain).MaxNesting = %d, want 1", *got.MaxNesting)
	}
}

func TestReadabilitySlopProxies(t *testing.T) {
	tests := []struct {
		name  string
		src   string
		field func(metrics) int
		want  int
	}{
		{"doc repeats the name", "// NewItem creates a new Item.\nfunc NewItem() {}\n",
			func(m metrics) int { return *m.EchoDocs }, 1},
		{"method doc repeats name and receiver", "type Store struct{}\n\n// Close closes the store.\nfunc (s *Store) Close() {}\n",
			func(m metrics) int { return *m.EchoDocs }, 1},
		{"doc carries information", "// NewItem validates id and returns ErrEmpty for \"\".\nfunc NewItem(id string) {}\n",
			func(m metrics) int { return *m.EchoDocs }, 0},
		{"one-statement helper called once", "func total(xs []int) int { return len(xs) }\n\nfunc Use(xs []int) int { return total(xs) }\n",
			func(m metrics) int { return *m.OneCallHelpers }, 1},
		{"helper called twice", "func total(xs []int) int { return len(xs) }\n\nfunc Use(xs []int) int { return total(xs) + total(xs) }\n",
			func(m metrics) int { return *m.OneCallHelpers }, 0},
		{"exported function called once", "func Total(xs []int) int { return len(xs) }\n\nfunc Use(xs []int) int { return Total(xs) }\n",
			func(m metrics) int { return *m.OneCallHelpers }, 0},
		{"four-statement helper", "func steps() { _ = 1; _ = 2; _ = 3; _ = 4 }\n\nfunc Use() { steps() }\n",
			func(m metrics) int { return *m.OneCallHelpers }, 0},
		{"log then return err", "import \"log/slog\"\n\nfunc F(err error) error {\n\tif err != nil {\n\t\tslog.Error(\"load\", \"err\", err)\n\t\treturn err\n\t}\n\treturn nil\n}\n",
			func(m metrics) int { return *m.LogAndReturn }, 1},
		{"log then return wrap", "import (\"fmt\"; \"log\")\n\nfunc F(err error) error {\n\tlog.Printf(\"load: %v\", err)\n\treturn fmt.Errorf(\"load: %w\", err)\n}\n",
			func(m metrics) int { return *m.LogAndReturn }, 1},
		{"log then degrade", "import \"log/slog\"\n\nfunc F(err error) error {\n\tslog.Warn(\"metrics\", \"err\", err)\n\treturn nil\n}\n",
			func(m metrics) int { return *m.LogAndReturn }, 0},
		{"wrap without logging", "import \"fmt\"\n\nfunc F(err error) error {\n\treturn fmt.Errorf(\"load: %w\", err)\n}\n",
			func(m metrics) int { return *m.LogAndReturn }, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzeSource(t, map[string]string{"p.go": "package p\n\n" + tt.src})
			if n := tt.field(got); n != tt.want {
				t.Errorf("analyze(%q) = %d, want %d", tt.src, n, tt.want)
			}
		})
	}
}

func TestReadabilityDeltaMissingSide(t *testing.T) {
	measured := metrics{MaxFuncLines: new(40)}
	if got := measured.sub(metrics{}).MaxFuncLines; got != nil {
		t.Errorf("sub(measured, unmeasured).MaxFuncLines = %d, want nil", *got)
	}
	if got := measured.sub(metrics{MaxFuncLines: new(10)}).MaxFuncLines; got == nil || *got != 30 {
		t.Errorf("sub(40, 10).MaxFuncLines = %v, want 30", fmt.Sprint(got))
	}
}

// A report saved before the readability fields existed still loads, and its
// summary says the metrics were not measured instead of averaging zeros.
func TestReadabilitySummaryOfOldReport(t *testing.T) {
	const saved = `{"corpus":"implement","runner":"claude","reps":1,"arms":[{"name":"baseline"}],
"results":[{"arm":"baseline","task":"feed","rep":0,"skills":["go-code"],
"before":{"lines":20,"files":1,"test_files":0,"types":1,"interfaces":0,"funcs":2,"closures":0,"body_comments":0,"pattern_names":0,"exported":2,"branches":0},
"after":{"lines":45,"files":1,"test_files":1,"types":1,"interfaces":0,"funcs":2,"closures":0,"body_comments":1,"pattern_names":0,"exported":2,"branches":4},
"delta":{"lines":25,"files":0,"test_files":1,"types":0,"interfaces":0,"funcs":0,"closures":0,"body_comments":1,"pattern_names":0,"exported":0,"branches":4},
"build":true,"golden":true,"edited":true}]}`
	var rep report
	if err := json.Unmarshal([]byte(saved), &rep); err != nil {
		t.Fatalf("json.Unmarshal(saved report) error = %v, want nil", err)
	}
	if d := rep.Results[0].Delta; d.MaxFuncLines != nil || d.Lines != 25 {
		t.Fatalf("saved report delta = lines %d, max func %v; want 25 and unmeasured", d.Lines, d.MaxFuncLines)
	}
	summary := summarizeArm(rep, "baseline")
	if summary.Valid != 1 || summary.Readability.Runs != 0 {
		t.Fatalf("summarizeArm(saved report) = valid %d, readability runs %d; want 1 and 0", summary.Valid, summary.Readability.Runs)
	}
	if got := summary.Readability.line(summary.Valid); !strings.Contains(got, "—") {
		t.Errorf("readability line for an old report = %q, want a dash", got)
	}

	rep.Results[0].Delta.MaxFuncLines = new(12)
	rep.Results[0].Delta.P90FuncLines = new(8)
	rep.Results[0].Delta.MaxNesting = new(2)
	rep.Results[0].Delta.EchoDocs = new(0)
	rep.Results[0].Delta.OneCallHelpers = new(1)
	rep.Results[0].Delta.LogAndReturn = new(0)
	got := summarizeArm(rep, "baseline").Readability.line(1)
	if !strings.Contains(got, "max func 12.0") || !strings.Contains(got, "(1/1 runs)") {
		t.Errorf("readability line for a measured run = %q, want max func 12.0 over 1/1 runs", got)
	}
}
