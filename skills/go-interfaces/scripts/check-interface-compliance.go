package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const version = "1.4.0"

type ifaceInfo struct {
	Name string `json:"name"`
	File string `json:"file"`
	Line int    `json:"line"`
	// ReturnedBy names the exported function that returns the interface from
	// its own package; empty when the finding is an unconverted interface.
	ReturnedBy string `json:"returned_by,omitempty"`

	key string
	obj *types.TypeName
}

type result struct {
	Interfaces     []ifaceInfo `json:"interfaces"`
	Missing        []ifaceInfo `json:"missing"`
	CountInterface int         `json:"count_interfaces"`
	CountMissing   int         `json:"count_missing"`
	Truncated      bool        `json:"truncated"`
	Status         string      `json:"status,omitempty"`
}

type sourceFile struct {
	path          string
	pkgName       string
	includeInScan bool
}

type packageGroup struct {
	key   string
	dir   string
	name  string
	files []sourceFile
}

type options struct {
	jsonOutput  bool
	includeTest bool
	limit       int
	target      string
	help        bool
	version     bool
}

func usage() {
	fmt.Fprintf(os.Stdout, `check-interface-compliance.sh v%s - List exported interfaces implemented beside their declaration that nothing converts to or that an exported function returns

USAGE
    bash check-interface-compliance.sh [options] [path]

OPTIONS
    -h, --help       Show this help message
    -v, --version    Show version
    --json           Output results as JSON
    --include-test   Also scan _test.go files for interface definitions and implementations
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
		fmt.Printf("check-interface-compliance.sh v%s\n", version)
		return
	}

	files, err := findGoFiles(opts.target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	// Test files are scanned only with --include-test, generated files never;
	// both are still type-checked with their package.
	allFiles := make([]sourceFile, 0, len(files))
	scanned := 0
	for _, path := range files {
		pkgName, generated, err := packageClause(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: parse package %s: %v\n", path, err)
			os.Exit(2)
		}
		include := (!strings.HasSuffix(path, "_test.go") || opts.includeTest) && !generated
		if include {
			scanned++
		}
		allFiles = append(allFiles, sourceFile{path: path, pkgName: pkgName, includeInScan: include})
	}

	if scanned == 0 {
		out := result{
			Interfaces: []ifaceInfo{},
			Missing:    []ifaceInfo{},
			Status:     "no_go_files",
		}
		emit(out, opts.jsonOutput)
		return
	}

	groups := groupByPackage(allFiles)
	exports := listExports(opts.target)
	assertions := map[string]map[string]bool{}
	interfaces := []ifaceInfo{}
	localImpl := map[string]map[string]bool{}
	returned := map[string]map[string]string{}

	for _, group := range groups {
		groupAssertions, groupInterfaces, groupImpls, groupReturned, err := analyzePackage(group, exports)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(2)
		}
		assertions[group.key] = groupAssertions
		interfaces = append(interfaces, groupInterfaces...)
		localImpl[group.key] = groupImpls
		returned[group.key] = groupReturned
	}

	sort.Slice(interfaces, func(i, j int) bool {
		if interfaces[i].Name == interfaces[j].Name {
			if interfaces[i].File == interfaces[j].File {
				return interfaces[i].Line < interfaces[j].Line
			}
			return interfaces[i].File < interfaces[j].File
		}
		return interfaces[i].Name < interfaces[j].Name
	})

	missing := []ifaceInfo{}
	for _, iface := range interfaces {
		if !localImpl[iface.key][iface.Name] {
			continue
		}
		iface.ReturnedBy = returned[iface.key][iface.Name]
		if iface.ReturnedBy != "" || !assertions[iface.key][iface.Name] {
			missing = append(missing, iface)
		}
	}
	sort.Slice(missing, func(i, j int) bool {
		if missing[i].File == missing[j].File {
			return missing[i].Line < missing[j].Line
		}
		return missing[i].File < missing[j].File
	})

	totalMissing := len(missing)
	truncated := false
	if opts.limit > 0 && totalMissing > opts.limit {
		missing = missing[:opts.limit]
		truncated = true
	}

	out := result{
		Interfaces:     interfaces,
		Missing:        missing,
		CountInterface: len(interfaces),
		CountMissing:   totalMissing,
		Truncated:      truncated,
	}
	if len(interfaces) == 0 {
		out.Status = "no_exported_interfaces"
	}

	emit(out, opts.jsonOutput)
	if totalMissing > 0 {
		os.Exit(1)
	}
}

func emit(out result, jsonOutput bool) {
	if jsonOutput {
		data, err := json.Marshal(out)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: marshal JSON: %v\n", err)
			os.Exit(2)
		}
		fmt.Println(string(data))
		return
	}

	if out.Status == "no_go_files" {
		fmt.Println("No Go files found.")
		return
	}
	if out.Status == "no_exported_interfaces" {
		fmt.Println("No exported interfaces found.")
		return
	}

	fmt.Printf("Exported interfaces found: %d\n\n", out.CountInterface)
	if out.CountMissing == 0 {
		fmt.Println("No interface implemented beside its declaration lacks a static conversion or is returned by an exported function.")
		return
	}

	fmt.Println("Interfaces implemented beside their declaration that nothing converts to or that an exported function returns:")
	fmt.Println()
	for _, item := range out.Missing {
		if item.ReturnedBy != "" {
			fmt.Printf("  %s:%d  does a consumer need interface '%s', or should %s return the concrete type?\n", item.File, item.Line, item.Name, item.ReturnedBy)
			continue
		}
		fmt.Printf("  %s:%d  does a consumer need interface '%s', or is the concrete type enough?\n", item.File, item.Line, item.Name)
	}
	if out.Truncated {
		fmt.Printf("  ... and %d more (use --limit to adjust)\n", out.CountMissing-len(out.Missing))
	}
	fmt.Println()
	fmt.Println("An interface belongs to the package that consumes it (go-interfaces).")
	fmt.Println()
	fmt.Printf("Total: %d interface(s) to question\n", out.CountMissing)
}

func analyzePackage(group packageGroup, exports exportIndex) (map[string]bool, []ifaceInfo, map[string]bool, map[string]string, error) {
	fset := token.NewFileSet()
	parsed := make([]*ast.File, 0, len(group.files))
	fileByAST := map[*ast.File]sourceFile{}
	for _, sf := range group.files {
		file, err := parser.ParseFile(fset, sf.path, nil, 0)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("parse %s: %w", sf.path, err)
		}
		parsed = append(parsed, file)
		fileByAST[file] = sf
	}

	info := &types.Info{
		Defs:      map[*ast.Ident]types.Object{},
		Uses:      map[*ast.Ident]types.Object{},
		Types:     map[ast.Expr]types.TypeAndValue{},
		Instances: map[*ast.Ident]types.Instance{},
	}
	conf := types.Config{
		Importer: exports.importerFor(fset, group.dir),
		Error:    func(error) {},
	}
	_, _ = conf.Check(group.key, fset, parsed, info)

	assertions := map[string]bool{}
	interfaces := []ifaceInfo{}
	typeNames := []*types.TypeName{}

	for _, file := range parsed {
		sf := fileByAST[file]
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			switch gen.Tok {
			case token.VAR:
				for _, spec := range gen.Specs {
					vs, ok := spec.(*ast.ValueSpec)
					if !ok || vs.Type == nil {
						continue
					}
					for _, name := range vs.Names {
						if name.Name != "_" {
							continue
						}
						if asserted := assertedInterface(vs.Type); asserted != "" {
							assertions[asserted] = true
						}
					}
				}
			case token.TYPE:
				for _, spec := range gen.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					obj, ok := info.Defs[ts.Name].(*types.TypeName)
					if !ok {
						continue
					}
					typeNames = append(typeNames, obj)
					if !sf.includeInScan || !ast.IsExported(ts.Name.Name) {
						continue
					}
					if iface, ok := obj.Type().Underlying().(*types.Interface); ok {
						iface.Complete()
						interfaces = append(interfaces, ifaceInfo{
							Name: ts.Name.Name,
							File: sf.path,
							Line: fset.Position(ts.Pos()).Line,
							key:  group.key,
							obj:  obj,
						})
					}
				}
			}
		}
	}

	impls := map[string]bool{}
	for _, iface := range interfaces {
		for _, ifaceType := range interfaceForms(iface.obj, info) {
			ifaceType.Complete()
			if ifaceType.NumMethods() == 0 || impls[iface.Name] {
				continue
			}
			for _, typeName := range typeNames {
				if typeName == iface.obj {
					continue
				}
				if !typeBelongsToScannedFile(typeName, group.files, fset) {
					continue
				}
				if _, ok := typeName.Type().Underlying().(*types.Interface); ok {
					continue
				}
				if implements(typeName.Type(), ifaceType) {
					impls[iface.Name] = true
					break
				}
			}
		}
	}

	for name := range staticConversions(parsed, info, interfaces) {
		assertions[name] = true
	}
	scanned := map[*ast.File]bool{}
	for file, sf := range fileByAST {
		scanned[file] = sf.includeInScan
	}
	return assertions, interfaces, impls, returnedInterfaces(parsed, scanned, info, interfaces), nil
}

// staticConversions returns the interfaces that a concrete value is assigned,
// returned, passed, sent, converted to, or placed in a composite literal as
// somewhere in the package. The compiler already checks those pairs, so they
// need no assertion.
func staticConversions(files []*ast.File, info *types.Info, interfaces []ifaceInfo) map[string]bool {
	converted := map[string]bool{}
	note := func(target types.Type, value ast.Expr) {
		tv, ok := info.Types[value]
		if target == nil || !ok || tv.Type == nil || types.IsInterface(tv.Type) {
			return
		}
		for _, iface := range interfaces {
			if sameInterface(target, iface.obj) {
				converted[iface.Name] = true
			}
		}
	}
	typeOf := func(e ast.Expr) types.Type { return info.Types[e].Type }

	var walk func(root ast.Node, results *types.Tuple)
	walk = func(root ast.Node, results *types.Tuple) {
		ast.Inspect(root, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.FuncLit:
				if sig, ok := typeOf(n).(*types.Signature); ok {
					walk(n.Body, sig.Results())
				}
				return false
			case *ast.ValueSpec:
				if n.Type != nil {
					for _, v := range n.Values {
						note(typeOf(n.Type), v)
					}
				}
			case *ast.AssignStmt:
				if n.Tok == token.ASSIGN && len(n.Lhs) == len(n.Rhs) {
					for i, v := range n.Rhs {
						note(typeOf(n.Lhs[i]), v)
					}
				}
			case *ast.ReturnStmt:
				if results != nil && results.Len() == len(n.Results) {
					for i, v := range n.Results {
						note(results.At(i).Type(), v)
					}
				}
			case *ast.SendStmt:
				if ch, ok := underlying(typeOf(n.Chan)).(*types.Chan); ok {
					note(ch.Elem(), n.Value)
				}
			case *ast.CompositeLit:
				noteElements(n, info, note)
			case *ast.CallExpr:
				fun := info.Types[n.Fun]
				if fun.IsType() && len(n.Args) == 1 {
					note(fun.Type, n.Args[0])
					break
				}
				sig, ok := fun.Type.(*types.Signature)
				if !ok || sig.Params().Len() == 0 {
					break
				}
				last := sig.Params().Len() - 1
				for i, arg := range n.Args {
					param := sig.Params().At(min(i, last)).Type()
					if sig.Variadic() && i >= last && !n.Ellipsis.IsValid() {
						if slice, ok := param.(*types.Slice); ok {
							param = slice.Elem()
						}
					}
					note(param, arg)
				}
			}
			return true
		})
	}
	for _, file := range files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				walk(decl, nil)
				continue
			}
			var results *types.Tuple
			if obj, ok := info.Defs[fn.Name].(*types.Func); ok {
				results = obj.Type().(*types.Signature).Results()
			}
			if fn.Body != nil {
				walk(fn.Body, results)
			}
		}
	}
	return converted
}

// noteElements passes each element of a composite literal to note with the
// type it is assigned to: a struct field, a slice or array element, or a map
// key and value.
func noteElements(lit *ast.CompositeLit, info *types.Info, note func(types.Type, ast.Expr)) {
	litType := underlying(info.Types[lit].Type)
	if ptr, ok := litType.(*types.Pointer); ok {
		litType = underlying(ptr.Elem()) // elided &T in an outer literal
	}
	for i, elt := range lit.Elts {
		var key ast.Expr
		value := elt
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			key, value = kv.Key, kv.Value
		}
		switch t := litType.(type) {
		case *types.Struct:
			if ident, ok := key.(*ast.Ident); ok {
				if field, ok := info.Uses[ident].(*types.Var); ok {
					note(field.Type(), value)
				}
			} else if key == nil && i < t.NumFields() {
				note(t.Field(i).Type(), value)
			}
		case *types.Slice:
			note(t.Elem(), value)
		case *types.Array:
			note(t.Elem(), value)
		case *types.Map:
			if key != nil {
				note(t.Key(), key)
			}
			note(t.Elem(), value)
		}
	}
}

func underlying(t types.Type) types.Type {
	if t == nil {
		return nil
	}
	return t.Underlying()
}

// returnedInterfaces maps each exported interface that an exported function
// returns to the first such function, the producer-owned interface that
// go-interfaces calls the Bad case. An interface the package also takes as a
// parameter is consumed there, as http.Handler is by its middleware, and is
// left out.
func returnedInterfaces(files []*ast.File, scanned map[*ast.File]bool, info *types.Info, interfaces []ifaceInfo) map[string]string {
	match := func(t types.Type) string {
		for _, iface := range interfaces {
			if sameInterface(t, iface.obj) {
				return iface.Name
			}
		}
		return ""
	}
	type exportedFunc struct {
		name string
		sig  *types.Signature
	}
	consumed := map[string]bool{}
	var exported []exportedFunc
	for _, file := range files {
		if !scanned[file] {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			obj, ok := info.Defs[fn.Name].(*types.Func)
			if !ok {
				continue
			}
			sig := obj.Type().(*types.Signature)
			for i := range sig.Params().Len() {
				param := sig.Params().At(i).Type()
				if slice, ok := param.(*types.Slice); ok && sig.Variadic() && i == sig.Params().Len()-1 {
					param = slice.Elem()
				}
				if name := match(param); name != "" {
					consumed[name] = true
				}
			}
			if fn.Recv == nil && fn.Name.IsExported() {
				exported = append(exported, exportedFunc{name: fn.Name.Name, sig: sig})
			}
		}
	}
	returned := map[string]string{}
	for _, fn := range exported {
		for result := range fn.sig.Results().Variables() {
			name := match(result.Type())
			if name == "" || consumed[name] || returned[name] != "" {
				continue
			}
			returned[name] = fn.name
		}
	}
	return returned
}

func typeBelongsToScannedFile(typeName *types.TypeName, files []sourceFile, fset *token.FileSet) bool {
	pos := fset.Position(typeName.Pos())
	for _, sf := range files {
		if sf.includeInScan && filepath.Clean(sf.path) == filepath.Clean(pos.Filename) {
			return true
		}
	}
	return false
}

// interfaceForms returns the interface types to test implementations against:
// the declared interface, or for a generic one each instantiation the package
// names, since a method set written in terms of T matches no concrete type.
func interfaceForms(obj *types.TypeName, info *types.Info) []*types.Interface {
	decl, ok := types.Unalias(obj.Type()).(*types.Named)
	if !ok || decl.TypeParams().Len() == 0 {
		if iface, ok := obj.Type().Underlying().(*types.Interface); ok {
			return []*types.Interface{iface}
		}
		return nil
	}
	var forms []*types.Interface
	for _, inst := range info.Instances {
		named, ok := inst.Type.(*types.Named)
		if !ok || named.Origin() != decl.Origin() {
			continue
		}
		if iface, ok := named.Underlying().(*types.Interface); ok {
			forms = append(forms, iface)
		}
	}
	return forms
}

// sameInterface reports whether t is the interface obj declares or, for a
// generic interface, one of its instantiations.
func sameInterface(t types.Type, obj *types.TypeName) bool {
	if types.Identical(t, obj.Type()) {
		return true
	}
	named, ok := types.Unalias(t).(*types.Named)
	if !ok || named.TypeArgs().Len() == 0 {
		return false
	}
	decl, ok := types.Unalias(obj.Type()).(*types.Named)
	return ok && decl.TypeParams().Len() > 0 && named.Origin() == decl.Origin()
}

func implements(t types.Type, iface *types.Interface) bool {
	if types.Implements(t, iface) {
		return true
	}
	if _, ok := t.(*types.Pointer); ok {
		return false
	}
	return types.Implements(types.NewPointer(t), iface)
}

func assertedInterface(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		if ast.IsExported(e.Name) {
			return e.Name
		}
	case *ast.IndexExpr:
		return assertedInterface(e.X)
	case *ast.IndexListExpr:
		return assertedInterface(e.X)
	}
	return ""
}

func groupByPackage(files []sourceFile) []packageGroup {
	byKey := map[string]*packageGroup{}
	for _, sf := range files {
		key := filepath.Dir(sf.path) + "|" + sf.pkgName
		group, ok := byKey[key]
		if !ok {
			group = &packageGroup{key: key, dir: filepath.Dir(sf.path), name: sf.pkgName}
			byKey[key] = group
		}
		group.files = append(group.files, sf)
	}

	groups := make([]packageGroup, 0, len(byKey))
	for _, group := range byKey {
		sort.Slice(group.files, func(i, j int) bool { return group.files[i].path < group.files[j].path })
		groups = append(groups, *group)
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].key < groups[j].key })
	return groups
}

// packageClause parses the package clause and the comments before it, where
// the "Code generated ... DO NOT EDIT." marker lives.
func packageClause(path string) (name string, generated bool, err error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.PackageClauseOnly|parser.ParseComments)
	if err != nil {
		return "", false, err
	}
	return file.Name.Name, ast.IsGenerated(file), nil
}

// exportIndex maps import paths to the export data `go list -export` wrote
// for the scanned module, so a module-local type in a method signature is a
// real type instead of an invalid one that any other invalid type matches.
type exportIndex struct {
	files map[string]string // import path -> export data file
	dirs  map[string]bool   // absolute directories of the packages go list saw
}

// listExports runs go list once for the target. Outside a module, or where go
// list fails, the index stays empty and importerFor falls back to
// importer.Default, which resolves the standard library only.
func listExports(target string) exportIndex {
	idx := exportIndex{files: map[string]string{}, dirs: map[string]bool{}}
	dir, pattern := target, "./..."
	if base, ok := strings.CutSuffix(target, "/..."); ok {
		dir = base
		if dir == "" {
			dir = "."
		}
	} else if info, err := os.Stat(target); err == nil && !info.IsDir() {
		dir, pattern = filepath.Dir(target), "."
	}
	cmd := exec.Command("go", "list", "-e", "-export", "-deps", "-test",
		"-f", "{{.ImportPath}}\t{{.Dir}}\t{{.Export}}", pattern)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOPROXY=off") // never download to resolve an import
	out, _ := cmd.Output()                        // with -e, a partial listing is still usable
	for line := range strings.Lines(string(out)) {
		fields := strings.Split(strings.TrimSuffix(line, "\n"), "\t")
		// Test variants ("p [p.test]") carry test-only declarations.
		if len(fields) != 3 || strings.Contains(fields[0], " ") {
			continue
		}
		if fields[1] != "" {
			idx.dirs[filepath.Clean(fields[1])] = true
		}
		if fields[2] != "" {
			idx.files[fields[0]] = fields[2]
		}
	}
	return idx
}

func (idx exportIndex) importerFor(fset *token.FileSet, dir string) types.Importer {
	abs, err := filepath.Abs(dir)
	if err != nil || !idx.dirs[abs] {
		return importer.Default()
	}
	return importer.ForCompiler(fset, "gc", func(path string) (io.ReadCloser, error) {
		if file, ok := idx.files[path]; ok {
			return os.Open(file)
		}
		return nil, fmt.Errorf("no export data for %s", path)
	})
}

func findGoFiles(target string) ([]string, error) {
	info, err := os.Stat(target)
	if err == nil {
		if !info.IsDir() {
			if strings.HasSuffix(target, ".go") {
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
		if !d.IsDir() && strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

// ignored reports a directory or file name that go ./... leaves out
// (go help packages): vendor, testdata, and names that begin with "." or "_".
func ignored(name string) bool {
	return name == "vendor" || name == "testdata" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}

func parseArgs(args []string) (options, error) {
	opts := options{target: "."}
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
		case arg == "--include-test":
			opts.includeTest = true
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
