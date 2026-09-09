package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestClassifyGolden pins the split that a raw golden rate cannot express. The
// collision case is the one the 5×5 concision run hit: the model added a
// production helper whose name the overlay also declares, and the run was first
// recorded as a behavior regression it was not.
func TestClassifyGolden(t *testing.T) {
	tests := []struct {
		name              string
		built             bool
		output            string
		behavior, harness bool
	}{
		{
			name:    "name collision with the overlay",
			built:   true,
			output:  "go test ./gateway/...: exit status 1: # abeval/gateway\n./gateway.go:41:6: goldenServe redeclared in this block\n\t./golden_test.go:22:6: other declaration of goldenServe",
			harness: true,
		},
		{
			name:     "assertion failed",
			built:    true,
			output:   "--- FAIL: TestStatusCodes/post_to_list (0.00s)\n    golden_test.go:179: POST /accounts = 200, want 405",
			behavior: true,
		},
		{
			name:     "exported symbol the overlay needs is gone",
			built:    true,
			output:   "# abeval/gateway [abeval/gateway.test]\n./golden_test.go:24:9: undefined: NewServer",
			behavior: true,
		},
		{
			name:   "package never built on its own",
			output: "./gateway.go:12:1: syntax error: unexpected }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			behavior, harness := classifyGolden(tt.built, tt.output)
			if behavior != tt.behavior || harness != tt.harness {
				t.Errorf("classifyGolden(%v, ...) = behavior %v, harness %v, want %v and %v",
					tt.built, behavior, harness, tt.behavior, tt.harness)
			}
		})
	}
}

// TestReportedCounts covers the prose heuristic. It answers what a session
// claimed, so a stated count is a hit whether or not the session measured one,
// and a path:line reference is not a count at all.
func TestReportedCounts(t *testing.T) {
	tests := []struct {
		name  string
		final string
		want  bool
	}{
		{name: "empty message", final: ""},
		{name: "count first", final: "Removed the wrapper: 148 lines down to 132.", want: true},
		{name: "word first", final: "Production lines went from 148 to 132.", want: true},
		{name: "loc", final: "LOC is now 132.", want: true},
		{name: "singular", final: "Net effect is 1 line shorter.", want: true},
		{name: "claim without a number", final: "The package is shorter and behavior is unchanged."},
		{name: "file position only", final: "Simplified the handler in gateway.go:41 and kept the contract."},
		{name: "position plus a real count", final: "gateway.go:41 was the wrapper; 12 lines are gone.", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reportedCounts(tt.final); got != tt.want {
				t.Errorf("reportedCounts(%q) = %v, want %v", tt.final, got, tt.want)
			}
		})
	}
}

// TestCodexCommands counts attempts, not transcript events: codex emits a
// started and a completed item for one shell call.
func TestCodexCommands(t *testing.T) {
	transcript := strings.Join([]string{
		`{"type":"item.started","item":{"type":"command_execution","command":"wc -l gateway/gateway.go"}}`,
		`{"type":"item.completed","item":{"type":"command_execution","command":"wc -l gateway/gateway.go"}}`,
		`{"type":"item.completed","item":{"type":"command_execution","command":"go build ./..."}}`,
		`{"type":"item.completed","item":{"type":"agent_message","text":"Down to 132 lines."}}`,
		`not json`,
		``,
	}, "\n")

	if got := codexCommands([]byte(transcript)); got != 2 {
		t.Errorf("codexCommands(transcript) = %d, want 2", got)
	}
	if got := codexCommands(nil); got != 0 {
		t.Errorf("codexCommands(nil) = %d, want 0", got)
	}
}

// TestGoldenHelpersAreCollisionResistant is the guard for the harness fault
// itself. Every name a hidden golden declares at package level shares a
// namespace with the production code the model writes, so a plain name like
// serve or errTransport is a collision waiting to be scored as a regression.
// Test entry points are exempt: the model has no reason to declare one, and
// hideTestFiles has already removed the tests it did write.
func TestGoldenHelpersAreCollisionResistant(t *testing.T) {
	refactorRoot, implementRoot := corpusRoots()

	for _, root := range []string{refactorRoot, implementRoot} {
		goldenDir := filepath.Join(root, "_golden")
		entries, err := os.ReadDir(goldenDir)
		if err != nil {
			t.Fatalf("os.ReadDir(%q) error = %v, want nil", goldenDir, err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			path := filepath.Join(goldenDir, entry.Name(), "golden_test.go")
			for _, name := range topLevelNames(t, path) {
				if strings.HasPrefix(name, "Test") || strings.HasPrefix(name, "Benchmark") || strings.HasPrefix(name, "Example") {
					continue
				}
				if !strings.Contains(strings.ToLower(name), "golden") {
					t.Errorf("%s declares %q; name it golden* so the model's production code cannot collide with it", path, name)
				}
			}
		}
	}
}

// topLevelNames returns every package-level identifier a Go file declares.
func topLevelNames(t *testing.T, path string) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parser.ParseFile(%q) error = %v, want nil", path, err)
	}
	var names []string
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			// A method is namespaced by its receiver, which is itself checked.
			if d.Recv == nil {
				names = append(names, d.Name.Name)
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					names = append(names, s.Name.Name)
				case *ast.ValueSpec:
					for _, id := range s.Names {
						names = append(names, id.Name)
					}
				}
			}
		}
	}
	return names
}
