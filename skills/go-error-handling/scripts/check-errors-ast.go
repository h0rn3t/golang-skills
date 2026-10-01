package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

const version = "1.5.0"

type finding struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

type parseError struct {
	File    string `json:"file"`
	Message string `json:"message"`
}

type options struct {
	jsonOutput      bool
	checkBareReturn bool
	limit           int
	target          string
	help            bool
	version         bool
}

func usage() {
	fmt.Fprintf(os.Stdout, `check-errors.sh v%s - Check Go code for common error handling anti-patterns

USAGE
    bash check-errors.sh [options] [path]

    Reports text matching on err.Error() (==, !=, switch, and strings.Contains,
    HasPrefix, HasSuffix, EqualFold, Index) and an error that is both logged
    and returned. Skips _test.go and generated files and, as go ./... does,
    vendor and testdata directories and names that begin with "." or "_".

    Exits 0 if nothing is found, 1 on findings, 2 on a usage error or when a
    file does not parse. The other files are still checked; --json then adds
    "status":"parse_error" and a "parse_errors" list.

OPTIONS
    -h, --help       Show this help message
    -v, --version    Show version
    --json           Output results as JSON
    --bare-return    Also flag bare 'return err' for review (off by default: the
                     skill allows a bare return when annotation adds nothing)
    --no-bare-return Accepted for compatibility; the check is already off
    --limit N        Show at most N results (default: all)
`, version)
}

func main() {
	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
	if opts.help {
		usage()
		return
	}
	if opts.version {
		fmt.Printf("check-errors.sh v%s\n", version)
		return
	}

	files, err := findGoFiles(opts.target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}
	if len(files) == 0 {
		if opts.jsonOutput {
			fmt.Println(`{"findings":[],"total":0,"truncated":false,"status":"no_go_files"}`)
		} else {
			fmt.Printf("No Go files found in: %s\n", opts.target)
		}
		return
	}

	findings := []finding{}
	var parseErrors []parseError
	for _, file := range files {
		fileFindings, err := analyzeFile(file, opts.checkBareReturn)
		if err != nil {
			parseErrors = append(parseErrors, parseError{File: file, Message: err.Error()})
			continue
		}
		findings = append(findings, fileFindings...)
	}

	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].File == findings[j].File {
			if findings[i].Line == findings[j].Line {
				return findings[i].Rule < findings[j].Rule
			}
			return findings[i].Line < findings[j].Line
		}
		return findings[i].File < findings[j].File
	})

	total := len(findings)
	truncated := false
	if opts.limit > 0 && total > opts.limit {
		findings = findings[:opts.limit]
		truncated = true
	}

	if opts.jsonOutput {
		out := struct {
			Findings    []finding    `json:"findings"`
			Total       int          `json:"total"`
			Truncated   bool         `json:"truncated"`
			Status      string       `json:"status,omitempty"`
			ParseErrors []parseError `json:"parse_errors,omitempty"`
		}{Findings: findings, Total: total, Truncated: truncated, ParseErrors: parseErrors}
		if len(parseErrors) > 0 {
			out.Status = "parse_error"
		}
		data, err := json.Marshal(out)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: marshal JSON: %v\n", err)
			os.Exit(2)
		}
		fmt.Println(string(data))
	} else {
		printText(findings, parseErrors, total, truncated, opts.limit)
	}

	if len(parseErrors) > 0 {
		os.Exit(2)
	}
	if total > 0 {
		os.Exit(1)
	}
}

func printText(shown []finding, parseErrors []parseError, total int, truncated bool, limit int) {
	if len(parseErrors) > 0 {
		fmt.Println("Malformed Go files:")
		fmt.Println()
		for _, pe := range parseErrors {
			fmt.Printf("  %s  %s\n", pe.File, pe.Message)
		}
		fmt.Println()
	}
	switch {
	case total == 0 && len(parseErrors) == 0:
		fmt.Println("No error handling anti-patterns found.")
	case total > 0:
		fmt.Println("Error handling anti-patterns found:")
		fmt.Println()
		for _, item := range shown {
			fmt.Printf("  %s:%d  [%s] %s\n", item.File, item.Line, item.Rule, item.Message)
		}
		if truncated {
			fmt.Printf("  ... and %d more (use --limit to adjust)\n", total-limit)
		}
		fmt.Println()
		fmt.Printf("Total: %d finding(s)\n", total)
	}
}

// analyzeFile reports the findings in one file; a generated file has none,
// since its author is the generator.
func analyzeFile(path string, checkBareReturn bool) ([]finding, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}
	if ast.IsGenerated(file) {
		return nil, nil
	}

	var findings []finding

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.BinaryExpr:
			if node.Op != token.EQL && node.Op != token.NEQ {
				return true
			}
			line := fset.Position(node.Pos()).Line
			switch {
			case isErrorCall(node.X) && isTextOperand(node.Y):
				findings = append(findings, finding{
					File:    path,
					Line:    line,
					Rule:    "string-error-compare",
					Message: "comparing err.Error() to string; use errors.Is or errors.AsType instead",
				})
			case isTextOperand(node.X) && isErrorCall(node.Y):
				findings = append(findings, finding{
					File:    path,
					Line:    line,
					Rule:    "string-error-compare",
					Message: "comparing string to err.Error(); use errors.Is or errors.AsType instead",
				})
			}
		case *ast.SwitchStmt:
			if node.Tag != nil && isErrorCall(node.Tag) {
				findings = append(findings, finding{
					File:    path,
					Line:    fset.Position(node.Pos()).Line,
					Rule:    "string-error-compare",
					Message: "switching on err.Error(); use errors.Is or errors.AsType instead",
				})
			}
		case *ast.CallExpr:
			line := fset.Position(node.Pos()).Line
			if name, ok := stringsMatchOnErrorText(node); ok {
				findings = append(findings, finding{
					File:    path,
					Line:    line,
					Rule:    "string-error-compare",
					Message: fmt.Sprintf("using strings.%s on err.Error(); use errors.Is or errors.AsType instead", name),
				})
			}
		case *ast.BlockStmt:
			findings = append(findings, logAndReturn(fset, path, node.List)...)
		case *ast.CaseClause:
			findings = append(findings, logAndReturn(fset, path, node.Body)...)
		case *ast.CommClause:
			findings = append(findings, logAndReturn(fset, path, node.Body)...)
		case *ast.ReturnStmt:
			line := fset.Position(node.Return).Line
			if checkBareReturn && returnsErr(node) {
				findings = append(findings, finding{
					File:    path,
					Line:    line,
					Rule:    "bare-return-err",
					Message: "bare return err: confirm the caller does not need context this frame could add (fmt.Errorf(\"...: %w\", err))",
				})
			}
		}
		return true
	})

	return findings, nil
}

// logAndReturn reports a log call carrying err that is followed, in the same
// statement list, by a return of that err, bare or wrapped: typically the body
// of an `if err != nil`. A log in one block and a return in another never pair,
// so a "log and degrade" branch followed by an unrelated return, or a log that
// ends one function next to a return that starts the next, is not a finding.
// An assignment to err between the two makes the returned error a new one.
func logAndReturn(fset *token.FileSet, path string, stmts []ast.Stmt) []finding {
	var findings []finding
	logLine := 0
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.ExprStmt:
			if call, ok := s.X.(*ast.CallExpr); ok && isLogCallWithErr(call) {
				logLine = fset.Position(call.Pos()).Line
			}
		case *ast.AssignStmt:
			if assignsIdent(s, "err") {
				logLine = 0
			}
		case *ast.ReturnStmt:
			if logLine != 0 && (returnsErr(s) || returnsWrappedErr(s)) {
				retLine := fset.Position(s.Return).Line
				findings = append(findings, finding{
					File:    path,
					Line:    logLine,
					Rule:    "log-and-return",
					Message: fmt.Sprintf("error is both logged (line %d) and returned (line %d); handle errors once", logLine, retLine),
				})
			}
			logLine = 0
		}
	}
	return findings
}

func assignsIdent(stmt *ast.AssignStmt, name string) bool {
	return slices.ContainsFunc(stmt.Lhs, func(lhs ast.Expr) bool {
		ident, ok := lhs.(*ast.Ident)
		return ok && ident.Name == name
	})
}

func isErrorCall(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok || len(call.Args) != 0 {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "Error"
}

// isTextOperand reports the other side of a comparison with err.Error() that
// makes it text matching: a string literal, a named string constant or
// variable (`notFoundMsg`, `pkg.Msg`), or another error's text. nil, true, and
// false are left out: an Error method that returns a value, not a string, is
// compared with them.
func isTextOperand(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return e.Kind == token.STRING
	case *ast.Ident:
		return e.Name != "nil" && e.Name != "true" && e.Name != "false"
	case *ast.SelectorExpr:
		return true
	}
	return isErrorCall(expr)
}

// stringsMatchOnErrorText reports a strings call that matches err.Error()
// text, such as strings.HasPrefix(err.Error(), "dial"), and the function name.
func stringsMatchOnErrorText(call *ast.CallExpr) (string, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	switch sel.Sel.Name {
	case "Contains", "HasPrefix", "HasSuffix", "EqualFold", "Index":
	default:
		return "", false
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok || pkg.Name != "strings" {
		return "", false
	}
	return sel.Sel.Name, slices.ContainsFunc(call.Args, containsErrorCall)
}

func containsErrorCall(expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if found {
			return false
		}
		if e, ok := n.(ast.Expr); ok && isErrorCall(e) {
			found = true
			return false
		}
		return true
	})
	return found
}

// isLogCallWithErr reports a logger call that carries err in its arguments or
// in a call it is chained from. The receiver is the logger itself
// (`slog.Error`, `logger.Warn`), a field holding it (`s.logger.Error`,
// `h.log.Printf`), or a chain that starts there (`logger.With("op", op).Error`,
// `slog.Default().Error`, zerolog's `log.Error().Err(err).Msg`).
func isLogCallWithErr(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	carriesErr := func(args []ast.Expr) bool {
		return slices.ContainsFunc(args, func(arg ast.Expr) bool { return containsIdent(arg, "err") })
	}
	withErr := carriesErr(call.Args)
	methods := []string{sel.Sel.Name}
	receiver := sel.X
	for {
		inner, ok := receiver.(*ast.CallExpr)
		if !ok {
			break
		}
		innerSel, ok := inner.Fun.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		withErr = withErr || carriesErr(inner.Args)
		methods = append(methods, innerSel.Sel.Name)
		receiver = innerSel.X
	}
	if !withErr {
		return false
	}
	var name string
	switch x := receiver.(type) {
	case *ast.Ident:
		name = x.Name
	case *ast.SelectorExpr:
		name = x.Sel.Name
	}
	return isLogger(strings.ToLower(name), methods)
}

// isLogger reports a receiver named as a logger. A short or suffixed name
// (`l`, `lg`, `auditLog`, `reqLogger`) counts only with a logging method in
// the chain, since `l` also names lists and listeners.
func isLogger(name string, methods []string) bool {
	switch name {
	case "log", "logger", "slog":
		return true
	}
	if name != "l" && name != "lg" && !strings.HasSuffix(name, "log") && !strings.HasSuffix(name, "logger") {
		return false
	}
	return slices.ContainsFunc(methods, func(m string) bool {
		for _, prefix := range []string{"Debug", "Info", "Warn", "Error", "Print", "Log", "Msg"} {
			if strings.HasPrefix(m, prefix) {
				return true
			}
		}
		return false
	})
}

func containsIdent(expr ast.Expr, name string) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if found {
			return false
		}
		ident, ok := n.(*ast.Ident)
		if ok && ident.Name == name {
			found = true
			return false
		}
		return true
	})
	return found
}

func returnsErr(stmt *ast.ReturnStmt) bool {
	if len(stmt.Results) == 0 {
		return false
	}
	ident, ok := stmt.Results[len(stmt.Results)-1].(*ast.Ident)
	return ok && ident.Name == "err"
}

// returnsWrappedErr reports a last result that is a call carrying err, such as
// fmt.Errorf("save %d: %w", id, err): the same error, returned with context.
func returnsWrappedErr(stmt *ast.ReturnStmt) bool {
	if len(stmt.Results) == 0 {
		return false
	}
	call, ok := stmt.Results[len(stmt.Results)-1].(*ast.CallExpr)
	return ok && slices.ContainsFunc(call.Args, func(arg ast.Expr) bool { return containsIdent(arg, "err") })
}

func findGoFiles(target string) ([]string, error) {
	info, err := os.Stat(target)
	if err == nil {
		if !info.IsDir() {
			if strings.HasSuffix(target, ".go") && !strings.HasSuffix(target, "_test.go") {
				return []string{target}, nil
			}
			return nil, nil
		}
		return walkGoFiles(target)
	}

	if strings.HasSuffix(target, "/...") {
		dir := strings.TrimSuffix(target, "/...")
		if dir == "" {
			dir = "."
		}
		if info, statErr := os.Stat(dir); statErr == nil && info.IsDir() {
			return walkGoFiles(dir)
		}
	}

	return nil, fmt.Errorf("path not found: %s", target)
}

// walkGoFiles returns the non-test Go files beneath root, skipping what
// go ./... skips: vendor, testdata, and names that begin with "." or "_".
func walkGoFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path != root && ignored(d.Name()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

// ignored reports a directory or file name that go ./... leaves out
// (go help packages).
func ignored(name string) bool {
	return name == "vendor" || name == "testdata" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}

func parseArgs(args []string) (options, error) {
	opts := options{target: ".", checkBareReturn: false}
	var positionals []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			opts.help = true
		case arg == "-v" || arg == "--version":
			opts.version = true
		case arg == "--json":
			opts.jsonOutput = true
		case arg == "--bare-return":
			opts.checkBareReturn = true
		case arg == "--no-bare-return":
			opts.checkBareReturn = false
		case arg == "--limit":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--limit requires a number")
			}
			i++
			limit, err := parseNonNegativeInt(args[i])
			if err != nil {
				return opts, fmt.Errorf("--limit must be a non-negative integer, got: %s", args[i])
			}
			opts.limit = limit
		case strings.HasPrefix(arg, "--limit="):
			value := strings.TrimPrefix(arg, "--limit=")
			limit, err := parseNonNegativeInt(value)
			if err != nil {
				return opts, fmt.Errorf("--limit must be a non-negative integer, got: %s", value)
			}
			opts.limit = limit
		case strings.HasPrefix(arg, "-"):
			return opts, fmt.Errorf("unknown option: %s", arg)
		default:
			positionals = append(positionals, arg)
		}
	}
	if len(positionals) > 1 {
		return opts, fmt.Errorf("expected at most one path")
	}
	if len(positionals) == 1 {
		opts.target = positionals[0]
	}
	return opts, nil
}

func parseNonNegativeInt(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty")
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("invalid")
		}
		n = n*10 + int(r-'0')
	}
	return n, nil
}
