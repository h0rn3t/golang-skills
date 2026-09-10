package evals_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The skills carry two kinds of claim that a reader cannot check by reading:
// "(Go 1.NN)" next to a standard-library symbol, and an analyzer name next to
// the tool that runs it. Both were wrong at least once, and neither shows up as
// a broken link or a failed build. These tests resolve both against the
// installed toolchain instead of a hand-maintained list.

func goroot(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		t.Skipf("go env GOROOT: %v", err)
	}
	root := strings.TrimSpace(string(out))
	if _, err := os.Stat(filepath.Join(root, "api")); err != nil {
		t.Skipf("no api directory under %s", root)
	}
	return root
}

// apiIndex maps "package.Symbol" to the earliest Go minor that introduced it,
// where "package" is the last element of the import path. Fields of exported
// structs are indexed under their package too, so `http.Server.MaxHeaderValueCount`
// resolves via its last component.
func apiIndex(t *testing.T) map[string]string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(goroot(t), "api", "go1.*.txt"))
	if err != nil {
		t.Fatalf("glob api files: %v", err)
	}

	// pkg <path>, <kind> <rest>
	line := regexp.MustCompile(`^pkg ([^,]+), (?:func|type|const|var|method \([^)]*\)) (\w+)`)
	// A struct field arrives as "type Server struct, MaxHeaderValueCount int".
	field := regexp.MustCompile(`^pkg ([^,]+), type \w+ struct, (\w+) `)
	version := regexp.MustCompile(`go1\.(\d+)\.txt$`)

	index := map[string]string{}
	record := func(pkgPath, symbol, minor string) {
		key := path(pkgPath) + "." + symbol
		if prev, ok := index[key]; !ok || less(minor, prev) {
			index[key] = minor
		}
	}
	for _, f := range files {
		m := version.FindStringSubmatch(filepath.Base(f))
		if m == nil {
			continue // go1.txt, the 1.0 baseline
		}
		content, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		for _, l := range strings.Split(string(content), "\n") {
			if g := field.FindStringSubmatch(l); g != nil {
				record(g[1], g[2], m[1])
				continue
			}
			if g := line.FindStringSubmatch(l); g != nil {
				record(g[1], g[2], m[1])
			}
		}
	}
	if len(index) == 0 {
		t.Fatal("api index is empty")
	}

	// "?Symbol" resolves a bare name that the standard library defines exactly
	// once. Ambiguous names (`Clone`, `Reverse`) stay unresolvable and are left
	// to the carried package.
	owners := map[string][]string{}
	for key := range index {
		symbol := key[strings.IndexByte(key, '.')+1:]
		owners[symbol] = append(owners[symbol], key)
	}
	for symbol, keys := range owners {
		if len(keys) == 1 {
			index["?"+symbol] = index[keys[0]]
		}
	}
	return index
}

func path(importPath string) string {
	if i := strings.LastIndex(importPath, "/"); i >= 0 {
		return importPath[i+1:]
	}
	return importPath
}

func less(a, b string) bool {
	if len(a) != len(b) {
		return len(a) < len(b)
	}
	return a < b
}

// symbolKey turns a backticked token as written in prose — `slices.Compact`,
// `maphash.ComparableHasher[T]`, `http.Server.MaxHeaderValueCount`,
// `errors.AsType[*fs.PathError](err)` — into "package.Symbol". A bare `Reverse`
// takes the carried package, which is how these tables list a family after
// naming it once. The second result is the package this token establishes for
// the ones after it, empty when the token names none.
func symbolKey(token, carried string) (key, qualifies string) {
	token = strings.TrimSpace(token)
	if i := strings.IndexAny(token, "[("); i >= 0 {
		token = token[:i]
	}
	token = strings.Trim(token, ".")
	parts := strings.Split(token, ".")
	for _, p := range parts {
		if !isIdentifier(p) {
			return "", ""
		}
	}
	symbol := parts[len(parts)-1]
	// A lowercase symbol is a method on a value (`wg.Go`, `srv.Client`) or a
	// package name on its own; neither dates an API in this index.
	if symbol == "" || symbol[0] < 'A' || symbol[0] > 'Z' {
		return "", ""
	}
	if len(parts) == 1 {
		// A table row often lists struct fields bare — "`MaxHeaderBytes`,
		// `MaxHeaderValueCount` (Go 1.27+)". Take the carried package when the
		// line established one, otherwise fall back to a name the standard
		// library defines in exactly one place.
		if carried != "" {
			return carried + "." + symbol, ""
		}
		return "?" + symbol, ""
	}
	pkg := parts[0]
	// `t.Context`, `b.Loop`: a single-letter receiver is a variable, not a
	// package, and inventing a package from it would resolve to nothing.
	if len(pkg) < 2 {
		return "", ""
	}
	return pkg + "." + symbol, pkg
}

func isIdentifier(s string) bool {
	for i, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r == '_':
		case r >= '0' && r <= '9' && i > 0:
		default:
			return false
		}
	}
	return s != ""
}

// TestVersionClaimsMatchToolchain checks every inline "(Go 1.NN)" marker in the
// skills against $GOROOT/api. A marker governs the backticked symbols written
// since the previous marker on the same line, and states the minimum toolchain
// they need: no symbol it governs may have arrived *later* than the claim.
//
// That is the direction that breaks a build — guidance offering an API a
// release before it exists. The opposite slip, a marker dated later than the
// oldest symbol in a group it shares with a newer one, is imprecision these
// tables can carry deliberately ("`strings.Cut`, ..., `CutLast` (Go 1.27)"),
// so it is not an error here.
//
// A bare symbol inherits the package of the last qualified one on the line, so
// "`slices.Compact`, `Reverse`, `Concat`" resolves all three.
func TestVersionClaimsMatchToolchain(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	index := apiIndex(t)

	marker := regexp.MustCompile(`\(Go 1\.(\d+)\+?\)`)
	backticked := regexp.MustCompile("`([^`]+)`")

	var files []string
	for _, pattern := range []string{
		"COMPATIBILITY.md",
		"skills/*/SKILL.md",
		"skills/*/references/*.md",
	} {
		matches, err := filepath.Glob(filepath.Join(root, filepath.FromSlash(pattern)))
		if err != nil {
			t.Fatalf("glob %s: %v", pattern, err)
		}
		files = append(files, matches...)
	}

	for _, file := range files {
		rel, err := filepath.Rel(root, file)
		if err != nil {
			t.Fatalf("relative path for %s: %v", file, err)
		}
		rel = filepath.ToSlash(rel)

		for n, line := range strings.Split(readFile(t, file), "\n") {
			markers := marker.FindAllStringSubmatchIndex(line, -1)
			if markers == nil {
				continue
			}
			prev, pkg := 0, ""
			for _, m := range markers {
				claimed := line[m[2]:m[3]]
				segment := line[prev:m[0]]
				prev = m[1]

				for _, tok := range backticked.FindAllStringSubmatch(segment, -1) {
					key, qualified := symbolKey(tok[1], pkg)
					if qualified != "" {
						pkg = qualified
					}
					if key == "" {
						continue
					}
					actual, ok := index[key]
					if !ok || !less(claimed, actual) {
						continue
					}
					t.Errorf("%s:%d claims (Go 1.%s), but %s arrived in Go 1.%s",
						rel, n+1, claimed, strings.TrimPrefix(key, "?"), actual)
				}
			}
		}
	}
}

// Some analyzers are named after an ordinary Go or English word the skills also
// use in prose — `any` the predeclared identifier, `atomic` the package,
// `tests` the noun. A backticked one of these says nothing about which tool
// runs it, so attribution is not checked for them.
var ambiguousAnalyzerNames = map[string]bool{
	"any": true, "appends": true, "assign": true, "atomic": true, "bools": true,
	"composites": true, "defers": true, "directive": true, "inline": true,
	"printf": true, "shift": true, "slog": true, "tests": true,
	"unmarshal": true, "unreachable": true,
}

// analyzerSets returns the analyzers registered with `go tool vet` and
// `go tool fix`. The two tools share a framework and a naming style but not a
// namespace: `waitgroup` is a vet analyzer, `waitgroupgo` a fix one.
func analyzerSets(t *testing.T) (vet, fix map[string]bool) {
	t.Helper()
	load := func(tool string) map[string]bool {
		out, err := exec.Command("go", "tool", tool, "help").CombinedOutput()
		if err != nil {
			t.Skipf("go tool %s help: %v", tool, err)
		}
		names := map[string]bool{}
		inList := false
		for _, line := range strings.Split(string(out), "\n") {
			switch {
			case strings.HasPrefix(line, "Registered analyzers:"):
				inList = true
			case strings.HasPrefix(line, "By default"):
				inList = false
			case inList:
				if fields := strings.Fields(line); len(fields) >= 2 {
					names[fields[0]] = true
				}
			}
		}
		if len(names) == 0 {
			t.Fatalf("go tool %s help listed no analyzers", tool)
		}
		return names
	}
	return load("vet"), load("fix")
}

// TestAnalyzerToolAttribution catches an analyzer credited to the wrong tool: a
// line that tells the reader to run `go vet` must not name a fix-only analyzer,
// and the reverse. Both commands reject an unknown -NAME flag, so a misfiled
// name is a broken instruction, not a cosmetic slip.
func TestAnalyzerToolAttribution(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	vet, fix := analyzerSets(t)

	backticked := regexp.MustCompile("`([^`]+)`")
	err := filepath.WalkDir(filepath.Join(root, "skills"), func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(p, ".md") {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		content, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		for n, line := range strings.Split(string(content), "\n") {
			namesVet := strings.Contains(line, "go vet")
			namesFix := strings.Contains(line, "go fix")
			if namesVet == namesFix {
				continue // neither tool, or both — the attribution is ambiguous
			}
			for _, tok := range backticked.FindAllStringSubmatch(line, -1) {
				fields := strings.Fields(tok[1])
				if len(fields) == 0 {
					continue
				}
				name := strings.TrimPrefix(fields[0], "-")
				if ambiguousAnalyzerNames[name] {
					continue
				}
				switch {
				case namesVet && fix[name] && !vet[name]:
					t.Errorf("%s:%d runs `go vet` but names %q, which is a `go fix` analyzer", rel, n+1, name)
				case namesFix && vet[name] && !fix[name]:
					t.Errorf("%s:%d runs `go fix` but names %q, which is a `go vet` analyzer", rel, n+1, name)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk skills: %v", err)
	}
}
