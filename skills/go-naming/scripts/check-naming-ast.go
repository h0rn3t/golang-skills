// check-naming-ast is the implementation behind check-naming.sh. It parses Go
// files with go/ast instead of matching lines with regular expressions, so a
// SCREAMING_SNAKE word inside a string or a comment is no longer a finding.
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
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const version = "2.0.0"

type violation struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

type options struct {
	help, version, jsonOutput bool
	limit                     int
	target                    string
}

var (
	screamingName   = regexp.MustCompile(`^[A-Z][A-Z0-9]*_[A-Z0-9_]+$`)
	getterName      = regexp.MustCompile(`^Get([A-Z][A-Za-z0-9]*)$`)
	getterException = regexp.MustCompile(`^(By|From|Or|With|All)`)
	genericPackages = map[string]bool{
		"util": true, "utils": true, "helper": true, "helpers": true,
		"common": true, "misc": true, "shared": true, "base": true, "lib": true,
	}
)

func usage() {
	fmt.Printf(`check-naming.sh v%s — Check Go code for common naming anti-patterns

USAGE
    bash check-naming.sh [options] [path]

DESCRIPTION
    Parses Go source files and reports naming violations from the Go style
    guides:
      - SCREAMING_SNAKE_CASE constants (should be MixedCaps)
      - Get-prefixed getter methods (should omit Get)
      - Packages named util/helper/common/misc
      - Receivers named "this" or "self"

    Exits 0 if no violations found, 1 if violations found, 2 on error.

OPTIONS
    -h, --help       Show this help message
    -v, --version    Show version
    --json           Output results as JSON
    --limit N        Show at most N results (default: all)

ARGUMENTS
    path             Directory, ./... pattern, or Go file (default: current directory)
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
		fmt.Printf("check-naming.sh v%s\n", version)
		return
	}

	files, err := findGoFiles(opts.target)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
	if len(files) == 0 {
		if opts.jsonOutput {
			fmt.Println(`{"violations":[],"total":0,"truncated":false,"status":"no_go_files"}`)
		} else {
			fmt.Printf("No Go files found in: %s\n", opts.target)
		}
		return
	}

	var all []violation
	fset := token.NewFileSet()
	for _, path := range files {
		f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", path, err)
			continue
		}
		all = append(all, check(fset, path, f)...)
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].File != all[j].File {
			return all[i].File < all[j].File
		}
		if all[i].Line != all[j].Line {
			return all[i].Line < all[j].Line
		}
		return all[i].Rule < all[j].Rule
	})

	total := len(all)
	shown := all
	truncated := false
	if opts.limit > 0 && total > opts.limit {
		shown = all[:opts.limit]
		truncated = true
	}

	if opts.jsonOutput {
		out := struct {
			Violations []violation `json:"violations"`
			Total      int         `json:"total"`
			Truncated  bool        `json:"truncated"`
		}{Violations: shown, Total: total, Truncated: truncated}
		if shown == nil {
			out.Violations = []violation{}
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(out); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
	} else if total == 0 {
		fmt.Println("No naming violations found.")
	} else {
		fmt.Println("Naming violations found:")
		fmt.Println()
		for _, v := range shown {
			fmt.Printf("  %s:%d  [%s] %s\n", v.File, v.Line, v.Rule, v.Message)
		}
		if truncated {
			fmt.Printf("  ... and %d more (use --limit to adjust)\n", total-opts.limit)
		}
		fmt.Printf("\nTotal: %d violation(s)\n", total)
	}
	if total > 0 {
		os.Exit(1)
	}
}

func parseArgs(args []string) (options, error) {
	opts := options{target: "."}
	for i := 0; i < len(args); i++ {
		switch a := args[i]; a {
		case "-h", "--help":
			opts.help = true
		case "-v", "--version":
			opts.version = true
		case "--json":
			opts.jsonOutput = true
		case "--limit":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--limit requires a number")
			}
			n, err := strconv.Atoi(args[i+1])
			if err != nil || n < 0 {
				return opts, fmt.Errorf("--limit must be a non-negative integer, got: %s", args[i+1])
			}
			opts.limit = n
			i++
		default:
			if strings.HasPrefix(a, "-") {
				return opts, fmt.Errorf("unknown option: %s", a)
			}
			opts.target = a
		}
	}
	return opts, nil
}

// findGoFiles accepts a file, a directory, or a ./... pattern and returns the
// non-test Go files beneath it, skipping vendor and .git.
func findGoFiles(target string) ([]string, error) {
	dir := strings.TrimSuffix(target, "/...")
	if dir == "" {
		dir = "."
	}
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("path not found: %s", target)
	}
	if !info.IsDir() {
		if strings.HasSuffix(dir, ".go") {
			return []string{dir}, nil
		}
		return nil, nil
	}
	var files []string
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if name := d.Name(); path != dir && (name == "vendor" || name == ".git") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func check(fset *token.FileSet, path string, f *ast.File) []violation {
	var out []violation
	add := func(pos token.Pos, rule, msg string) {
		out = append(out, violation{File: path, Line: fset.Position(pos).Line, Rule: rule, Message: msg})
	}

	if pkg := f.Name.Name; genericPackages[pkg] {
		add(f.Name.Pos(), "bad-package-name",
			fmt.Sprintf("package '%s' is too generic; use a specific, descriptive name", pkg))
	}

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok != token.CONST {
				continue
			}
			for _, spec := range d.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range vs.Names {
					if screamingName.MatchString(name.Name) {
						add(name.Pos(), "screaming-const",
							fmt.Sprintf("constant '%s' uses SCREAMING_SNAKE_CASE; use MixedCaps instead", name.Name))
					}
				}
			}
		case *ast.FuncDecl:
			if d.Recv == nil || len(d.Recv.List) == 0 {
				continue
			}
			for _, recvName := range d.Recv.List[0].Names {
				if n := recvName.Name; n == "this" || n == "self" {
					add(d.Pos(), "bad-receiver",
						fmt.Sprintf("receiver named '%s'; use a short 1-2 letter abbreviation of the type instead", n))
				}
			}
			if m := getterName.FindStringSubmatch(d.Name.Name); m != nil && !getterException.MatchString(m[1]) {
				add(d.Pos(), "get-prefix",
					fmt.Sprintf("method '%s' has Get prefix; Go getters should omit Get (use '%s')", d.Name.Name, m[1]))
			}
		}
	}
	return out
}
