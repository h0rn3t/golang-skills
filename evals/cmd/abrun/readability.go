package main

import (
	"go/ast"
	"go/token"
	"slices"
	"strings"
	"unicode"
)

// readability fills the shape metrics the declaration counts cannot see. On a
// model whose golden tests saturate, a wording change that helps a reader
// shows up here or nowhere: a shorter longest function, shallower nesting, or
// fewer of the three slop proxies. Calls are matched by name within a
// package, without go/types, so each count is an approximation checked by
// hand on a sample before it is trusted.
func (m *metrics) readability(fset *token.FileSet, packages map[string][]*ast.File) {
	var lengths []int
	nest, echoes, helpers, logged := 0, 0, 0, 0
	for _, files := range packages {
		for _, file := range files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					continue
				}
				lengths = append(lengths, fset.Position(fn.End()).Line-fset.Position(fn.Pos()).Line+1)
				nest = max(nest, nesting(fn.Body))
				logged += logAndReturn(fn.Body)
			}
			echoes += echoDocs(file)
		}
		helpers += oneCallHelpers(files)
	}
	slices.Sort(lengths)
	longest, p90 := 0, 0
	if n := len(lengths); n > 0 {
		longest = lengths[n-1]
		p90 = lengths[(9*n+9)/10-1] // nearest rank: ceil(0.9n)
	}
	m.MaxFuncLines, m.P90FuncLines, m.MaxNesting = &longest, &p90, &nest
	m.EchoDocs, m.OneCallHelpers, m.LogAndReturn = &echoes, &helpers, &logged
}

// nesting returns the deepest control structure under body: if, for, range,
// switch, select, and function literals each add a level, and an else-if
// chain stays on one level, as it reads.
func nesting(body *ast.BlockStmt) int {
	deepest := 0
	var walk func(n ast.Node, depth int)
	enter := func(c ast.Node, depth int) bool {
		var inner ast.Node
		switch c := c.(type) {
		case *ast.IfStmt:
			switch e := c.Else.(type) {
			case *ast.IfStmt:
				walk(&ast.BlockStmt{List: []ast.Stmt{e}}, depth)
			case *ast.BlockStmt:
				walk(e, depth+1)
			}
			inner = c.Body
		case *ast.ForStmt:
			inner = c.Body
		case *ast.RangeStmt:
			inner = c.Body
		case *ast.SwitchStmt:
			inner = c.Body
		case *ast.TypeSwitchStmt:
			inner = c.Body
		case *ast.SelectStmt:
			inner = c.Body
		case *ast.FuncLit:
			inner = c.Body
		default:
			return false
		}
		deepest = max(deepest, depth+1)
		walk(inner, depth+1)
		return true
	}
	walk = func(n ast.Node, depth int) {
		ast.Inspect(n, func(c ast.Node) bool {
			if c == n {
				return true
			}
			return c != nil && !enter(c, depth)
		})
	}
	walk(body, 0)
	return deepest
}

// fillerWords carry no information in a first doc sentence: with them and the
// declared name removed, `// NewItem creates a new Item.` has nothing left.
var fillerWords = []string{
	"a", "an", "the", "new", "create", "return", "get", "set", "make", "is", "are",
	"this", "that", "it", "its", "of", "for", "to", "and", "or", "with", "by",
	"given", "instance", "value", "object", "type", "struct", "function", "method",
	"represent", "hold", "contain", "define", "implement", "provide", "specified",
}

// echoDocs counts doc comments on functions and types whose first sentence
// only restates the declared name.
func echoDocs(file *ast.File) int {
	count := 0
	check := func(doc *ast.CommentGroup, names ...string) {
		if doc != nil && echoesName(doc.Text(), names) {
			count++
		}
	}
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			names := []string{d.Name.Name}
			if d.Recv != nil && len(d.Recv.List) > 0 {
				names = append(names, receiverTypeName(d.Recv.List[0].Type))
			}
			check(d.Doc, names...)
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				doc := ts.Doc
				if doc == nil && len(d.Specs) == 1 {
					doc = d.Doc
				}
				check(doc, ts.Name.Name)
			}
		}
	}
	return count
}

// echoesName reports whether the first sentence of text has no word left once
// the words of names and the filler words are removed. A word also matches a
// stem plus s, es, d, or ed, so "closes" matches Close.
func echoesName(text string, names []string) bool {
	sentence, _, _ := strings.Cut(strings.Join(strings.Fields(text), " "), ". ")
	var stems []string
	for _, name := range names {
		stems = append(stems, identWords(name)...)
		stems = append(stems, strings.ToLower(name))
	}
	stems = append(stems, fillerWords...)
	words := strings.FieldsFunc(strings.ToLower(sentence), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	if len(words) == 0 {
		return false
	}
	for _, w := range words {
		if !slices.ContainsFunc(stems, func(s string) bool {
			return w == s || w == s+"s" || w == s+"es" || w == s+"d" || w == s+"ed"
		}) {
			return false
		}
	}
	return true
}

// identWords splits a MixedCaps identifier into lower-case words, keeping an
// initialism together: NewHTTPServer is new, http, server.
func identWords(name string) []string {
	runes := []rune(name)
	var words []string
	start := 0
	for i := 1; i < len(runes); i++ {
		lowerToUpper := unicode.IsLower(runes[i-1]) && unicode.IsUpper(runes[i])
		initialismEnd := unicode.IsUpper(runes[i-1]) && unicode.IsUpper(runes[i]) && i+1 < len(runes) && unicode.IsLower(runes[i+1])
		if lowerToUpper || initialismEnd || runes[i] == '_' {
			words = append(words, strings.ToLower(strings.Trim(string(runes[start:i]), "_")))
			start = i
		}
	}
	words = append(words, strings.ToLower(strings.Trim(string(runes[start:]), "_")))
	return slices.DeleteFunc(words, func(w string) bool { return w == "" })
}

func receiverTypeName(expr ast.Expr) string {
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	switch t := expr.(type) {
	case *ast.IndexExpr:
		expr = t.X
	case *ast.IndexListExpr:
		expr = t.X
	}
	if id, ok := expr.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}

// oneCallHelpers counts unexported package functions, methods excluded, whose
// body is at most three statements and whose name appears exactly once in the
// package besides its declaration. A name passed as a value counts as a use,
// so a handler registered by name is still counted when that is its one use.
func oneCallHelpers(files []*ast.File) int {
	uses := map[string]int{}
	var candidates []*ast.FuncDecl
	for _, file := range files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if ok && fn.Recv == nil && fn.Body != nil && !fn.Name.IsExported() &&
				fn.Name.Name != "main" && fn.Name.Name != "init" && len(fn.Body.List) <= 3 {
				candidates = append(candidates, fn)
			}
		}
		ast.Inspect(file, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok {
				uses[id.Name]++
			}
			return true
		})
	}
	count := 0
	for _, fn := range candidates {
		if uses[fn.Name.Name]-1 == 1 {
			count++
		}
	}
	return count
}

// logMethods are the calls that write a log record.
var logMethods = []string{
	"Print", "Printf", "Println", "Error", "Errorf", "Warn", "Warnf", "Info", "Infof",
	"Debug", "Debugf", "ErrorContext", "WarnContext", "InfoContext", "DebugContext", "Log", "LogAttrs",
}

// logAndReturn counts blocks that log an error and then return it, bare or
// wrapped: the handle-once rule gives a caller one or the other.
func logAndReturn(body *ast.BlockStmt) int {
	count := 0
	ast.Inspect(body, func(n ast.Node) bool {
		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}
		logged := map[string]bool{}
		for _, stmt := range block.List {
			switch s := stmt.(type) {
			case *ast.ExprStmt:
				if call, ok := s.X.(*ast.CallExpr); ok && isLogCall(call) {
					for _, arg := range call.Args {
						for _, name := range identNames(arg) {
							logged[name] = true
						}
					}
				}
			case *ast.ReturnStmt:
				returnsLogged := slices.ContainsFunc(s.Results, func(r ast.Expr) bool {
					return slices.ContainsFunc(identNames(r), func(name string) bool {
						return logged[name] && strings.HasSuffix(strings.ToLower(name), "err")
					})
				})
				if returnsLogged {
					count++
				}
			}
		}
		return true
	})
	return count
}

// isLogCall reports a method call named like a log write, leaving out fmt,
// http, and errors, whose Println, Error, and Errorf write no log record.
func isLogCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || !slices.Contains(logMethods, sel.Sel.Name) {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return !ok || (pkg.Name != "fmt" && pkg.Name != "http" && pkg.Name != "errors")
}

func identNames(expr ast.Expr) []string {
	var names []string
	ast.Inspect(expr, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			names = append(names, id.Name)
		}
		return true
	})
	return names
}
