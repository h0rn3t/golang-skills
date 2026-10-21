package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"strings"
	"unicode"
	"unicode/utf8"
)

// testSizeSum totals the test-size deltas over the valid runs that carry them;
// a report written before the fields existed prints a dash.
type testSizeSum struct {
	Runs, Lines, Funcs, Cases int
}

// add counts d when every test-size field was measured.
func (s *testSizeSum) add(d metrics) {
	if d.TestLines == nil || d.TestFuncs == nil || d.TestCases == nil {
		return
	}
	s.Runs++
	s.Lines += *d.TestLines
	s.Funcs += *d.TestFuncs
	s.Cases += *d.TestCases
}

// line renders the per-run means, or a dash when no run was measured.
func (s testSizeSum) line(valid int) string {
	if s.Runs == 0 {
		return "tests Δ/run: — (not measured in this report)"
	}
	mean := func(sum int) float64 { return float64(sum) / float64(s.Runs) }
	return fmt.Sprintf("tests Δ/run: lines %.1f, Test funcs %.2f, cases %.2f (%d/%d runs)",
		mean(s.Lines), mean(s.Funcs), mean(s.Cases), s.Runs, valid)
}

// addTestFile adds one _test.go file's lines to the test-size fields and
// returns the parsed file for addTestPackage. A file that does not parse
// still counts its lines: its package's tests fail, the run is excluded from
// the means, and the result's JSON still shows how much was written.
func (m *metrics) addTestFile(fset *token.FileSet, path string, data []byte) (*ast.File, bool) {
	*m.TestLines += lineCount(data)
	file, err := parser.ParseFile(fset, path, data, parser.SkipObjectResolution)
	return file, err == nil
}

// addTestPackage counts the tests and cases of one test package. Its files
// share the package-level tables and the named slice and map types a table
// can be declared with.
func (m *metrics) addTestPackage(files []*ast.File) {
	types := map[string]bool{}
	for _, file := range files {
		ast.Inspect(file, func(n ast.Node) bool {
			if spec, ok := n.(*ast.TypeSpec); ok && isTableType(spec.Type) {
				types[spec.Name.Name] = true
			}
			return true
		})
	}
	pkgRows := map[string]int{}
	for _, file := range files {
		for _, decl := range file.Decls {
			if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.VAR {
				for _, spec := range gen.Specs {
					bindTables(pkgRows, types, spec)
				}
			}
		}
	}
	for _, file := range files {
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && isTestFunc(fn) {
				*m.TestFuncs++
				*m.TestCases += testCases(fn, maps.Clone(pkgRows), types)
			}
		}
	}
}

// isTestFunc reports whether go test runs fn as a test: TestXxx, where Xxx
// does not start with a lower-case letter, taking one *testing.T. TestMain and
// helpers named Test-something are not tests.
func isTestFunc(fn *ast.FuncDecl) bool {
	rest, ok := strings.CutPrefix(fn.Name.Name, "Test")
	if !ok || fn.Recv != nil || fn.Body == nil || fn.Type.Params.NumFields() != 1 {
		return false
	}
	if r, _ := utf8.DecodeRuneInString(rest); unicode.IsLower(r) {
		return false
	}
	star, ok := fn.Type.Params.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	sel, ok := star.X.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "T"
}

// bindTables records the bindings n makes in rows: a name bound to a slice or
// map literal maps to the literal's element count, and a name bound to
// anything else is no table.
func bindTables(rows map[string]int, types map[string]bool, n ast.Node) {
	var names, values []ast.Expr
	switch node := n.(type) {
	case *ast.AssignStmt:
		names, values = node.Lhs, node.Rhs
	case *ast.ValueSpec:
		for _, name := range node.Names {
			names = append(names, name)
		}
		values = node.Values
	}
	for i, name := range names {
		id, ok := name.(*ast.Ident)
		if !ok {
			continue
		}
		var value ast.Expr
		if len(values) == len(names) {
			value = values[i]
		}
		if count, isTable := literalRows(value, types); isTable {
			rows[id.Name] = count
		} else {
			delete(rows, id.Name)
		}
	}
}

// isTableType reports whether typ is a slice or map type.
func isTableType(typ ast.Expr) bool {
	switch typ := typ.(type) {
	case *ast.ArrayType:
		return typ.Len == nil
	case *ast.MapType:
		return true
	}
	return false
}

// literalRows reports the element count of a slice or map composite literal,
// including one of a named type in types.
func literalRows(expr ast.Expr, types map[string]bool) (int, bool) {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return 0, false
	}
	if id, ok := lit.Type.(*ast.Ident); ok {
		return len(lit.Elts), types[id.Name]
	}
	return len(lit.Elts), isTableType(lit.Type)
}

// testCases counts the cases one test function runs: the rows of every table
// it ranges over, plus each subtest outside a loop that adds no case of its
// own (a group counts through its members, a subtest over a table through its
// rows). A test with neither is one case. A range reads the binding of its
// name that comes last before it in the source, over the package-level rows,
// so two groups that each declare `tests` count their own. Tables are matched
// by name, without go/types, so a fixture slice the test ranges over to build
// its input counts as a table too.
func testCases(fn *ast.FuncDecl, rows map[string]int, types map[string]bool) int {
	cases := 0
	var walk func(n ast.Node, inLoop bool)
	walk = func(n ast.Node, inLoop bool) {
		ast.Inspect(n, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.AssignStmt, *ast.ValueSpec:
				bindTables(rows, types, node)
			case *ast.RangeStmt:
				if count, ok := literalRows(node.X, types); ok {
					cases += count
				} else if id, ok := node.X.(*ast.Ident); ok {
					cases += rows[id.Name]
				}
				walk(node.Body, true)
				return false
			case *ast.ForStmt:
				walk(node.Body, true)
				return false
			case *ast.CallExpr:
				if !inLoop && isSubtest(node) {
					before := cases
					walk(node.Args[1], false)
					if cases == before {
						cases++
					}
					return false
				}
			}
			return true
		})
	}
	walk(fn.Body, false)
	return max(cases, 1)
}

// isSubtest reports whether call looks like t.Run(name, f): a Run with two
// arguments, which leaves out exec.Cmd.Run and other zero-argument Runs.
func isSubtest(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "Run" && len(call.Args) == 2
}
