package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"
)

// ownerPatterns is the one owner table: a decision-bearing form of the edited
// text names the owner skill behind it. go-code/SKILL.md owns the routing
// decision; each pattern stays narrow enough that ordinary syntax
// (fmt.Errorf with %v, r.Context(), http.StatusOK, "<-" inside a string, an
// empty interface{}, a sha256 content hash) names no owner. A crypto import
// other than a content hash is matched by securityCrypto.
var ownerPatterns = []struct {
	owner string
	re    *regexp.Regexp
}{
	{"go-http", regexp.MustCompile(`(?m)\bnet/http\b|\bhttp\.(Handle|HandleFunc|Server\b|Client\b|NewServeMux|NewRequest|ResponseWriter|Error|ListenAndServe|Redirect)`)},
	{"go-error-handling", regexp.MustCompile(`(?m)fmt\.Errorf\([^)]*%w|\berrors\.(Is|As|AsType|Join|New)\(`)},
	{"go-concurrency", regexp.MustCompile(`(?m)\bgo\s+func\b|\bgo\s+[A-Za-z_]\w*\(|\bmake\(chan\b|\bchan\s|\bsync\.(Mutex|RWMutex|WaitGroup|Once|Map)\b`)},
	{"go-context", regexp.MustCompile(`(?m)\bcontext\.(Background|TODO|With[A-Za-z]+|AfterFunc)\b|^\s+\w+\s+context\.Context\s*(//.*)?$`)},
	{"go-database", regexp.MustCompile(`(?m)database/sql|\bsql\.(Open|DB|Tx|Rows|Null)\b|\bpgx(pool)?\.`)},
	{"go-logging", regexp.MustCompile(`(?m)\bslog\.`)},
	{"go-security", regexp.MustCompile(`(?m)os/exec|html/template|text/template|\bexec\.Command|\bhttp\.(SetCookie|Cookie)\b|\bfilepath\.Join\(`)},
	{"go-defensive", regexp.MustCompile(`(?m)\bdefer\s+func\b|\brecover\(\)`)},
	{"go-generics", regexp.MustCompile(`(?m)\[[A-Z][A-Za-z0-9]*\s+(any|comparable|~|[A-Za-z]+\.[A-Za-z]+)\b`)},
	{"go-interfaces", regexp.MustCompile(`(?m)\binterface\s*\{\s*[^\s}]`)},
	{"go-packages", regexp.MustCompile(`(?m)^package main\b|\bfunc main\(`)},
	{"go-resilience", regexp.MustCompile(`(?m)x/time/rate|\bbackoff\b|\bRetry-After\b`)},
}

var (
	cryptoPath = regexp.MustCompile(`\bcrypto/`)
	contentSum = regexp.MustCompile(`^(?:sha256|sha3|sha512)\b`)
)

// securityCrypto reports a crypto/ path other than a content hash.
func securityCrypto(text string) bool {
	for _, loc := range cryptoPath.FindAllStringIndex(text, -1) {
		if !contentSum.MatchString(text[loc[1]:]) {
			return true
		}
	}
	return false
}

// hints names the owner skills the edited text points at. A test file has
// one owner: its body is test plumbing — a defer, an http.Request, an
// errors.Is on a want — not the decisions the content patterns recognize.
func hints(path, text string) []string {
	if strings.HasSuffix(path, "_test.go") {
		return []string{"go-testing"}
	}
	var out []string
	for _, p := range ownerPatterns {
		if p.re.MatchString(text) || (p.owner == "go-security" && securityCrypto(text)) {
			out = append(out, p.owner)
		}
	}
	return out
}

var (
	goTokens     = regexp.MustCompile("(?s)//[^\\n]*|/\\*.*?\\*/|\"(?:\\\\.|[^\"\\\\])*\"|`[^`]*`|'(?:\\\\.|[^'\\\\])*'")
	importWord   = regexp.MustCompile(`\bimport\b`)
	groupedOpen  = regexp.MustCompile(`^\s*\(`)
	importSplit  = regexp.MustCompile(`[;\n(]`)
	importAlias  = regexp.MustCompile(`([A-Za-z_]\w*|\.)\s*$`)
	selectorName = regexp.MustCompile(`\b([A-Za-z_]\w*)\s*\.`)
	panicCall    = regexp.MustCompile(`\bpanic\s*\(`)
	contextType  = regexp.MustCompile(`\bcontext\.Context\b`)

	extraOwners = []struct {
		owner string
		re    *regexp.Regexp
	}{
		{"go-http", regexp.MustCompile(`\bhttp\.(?:Request|Response|Transport|RoundTripper|Handler)\b`)},
		{"go-database", regexp.MustCompile(`\bsql\.(?:Conn|Stmt|Result|NamedArg|Null\w*)\b`)},
		{"go-security", regexp.MustCompile(`\b(?:template\.\w+|exec\.Cmd)\b`)},
	}
	dotOwners = []struct {
		pkg, owner string
		re         *regexp.Regexp
	}{
		{"sync", "go-concurrency", regexp.MustCompile(`(?:Mutex|RWMutex|WaitGroup|Once|Map)\b`)},
		{"errors", "go-error-handling", regexp.MustCompile(`(?:Is|As|AsType|Join|New)\s*\(`)},
		{"slog", "go-logging", regexp.MustCompile(`(?:Logger|Handler|Attr|LogAttrs|Info|Error)\b`)},
		{"http", "go-http", regexp.MustCompile(`(?:ResponseWriter|Request|Server|Client|NewServeMux)\b`)},
		{"sql", "go-database", regexp.MustCompile(`(?:DB|Tx|Rows|Open)\b`)},
		{"exec", "go-security", regexp.MustCompile(`(?:Command|CommandContext|Cmd)\b`)},
	}
	dotContext = regexp.MustCompile(`Context\b`)

	boundedWorkers = regexp.MustCompile(`(?i)\bbounded\b.{0,50}\bworkers?\b|\bat most\b.{0,70}\bin flight\b|\bowns\b.{0,40}\bgoroutines\b`)
	// "retry" as a verb, not the start of "Retry-After".
	retryContract = regexp.MustCompile(`(?is)\b(?:honor(?:s|ing)?|respect(?:s|ing)?|retr(?:ies|ied|ying))\b.{0,100}\b(?:Retry-After|backoff)\b|\bretry[^\w-].{0,99}\b(?:Retry-After|backoff)\b`)
	// Cache-Control is not a cache.
	cacheContract = regexp.MustCompile(`(?is)\bcach(?:e[sd]?|ing)[^\w-].{0,159}\bttl\b|\bttl\b.{0,160}\bcach(?:e[sd]?|ing)(?:[^\w-]|$)`)
)

// canonical maps an import path to the name the owner patterns use for it.
var canonical = map[string]string{
	"context": "context", "sync": "sync", "errors": "errors",
	"log/slog": "slog", "net/http": "http", "database/sql": "sql",
	"os/exec": "exec", "fmt": "fmt", "html/template": "template",
	"text/template": "template",
}

// matchesUnqualified reports a match of re that does not follow an
// identifier or a selector dot: the name a dot import brings in.
func matchesUnqualified(re *regexp.Regexp, code string) bool {
	for _, loc := range re.FindAllStringIndex(code, -1) {
		if loc[0] == 0 {
			return true
		}
		c := code[loc[0]-1]
		ident := c == '_' || '0' <= c && c <= '9' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
		if c != '.' && !ident {
			return true
		}
	}
	return false
}

// literalAfter reports a string literal that follows end with only blanks
// between, and for which accept holds.
func literalAfter(lexical string, end int, literals map[int]string, accept func(string) bool) bool {
	for pos, value := range literals {
		if pos >= end && strings.TrimSpace(lexical[end:pos]) == "" && accept(value) {
			return true
		}
	}
	return false
}

// promptHints extends hints for the prompt note with lexical signals: the
// imports and type names in use, and a stub's documented contract. They never
// widen the edit gate. Comments and literals are not executable code.
func promptHints(path, text string) []string {
	if strings.HasSuffix(path, "_test.go") {
		return []string{"go-testing"}
	}
	code := []byte(text)
	literals := map[int]string{}
	var comments []string
	for _, loc := range goTokens.FindAllStringIndex(text, -1) {
		value := text[loc[0]:loc[1]]
		if strings.HasPrefix(value, "//") || strings.HasPrefix(value, "/*") {
			comments = append(comments, value)
		} else {
			literals[loc[0]] = value
		}
		for i := loc[0]; i < loc[1]; i++ {
			if code[i] != '\n' {
				code[i] = ' '
			}
		}
	}
	lexical := string(code)

	imports, dot, extras := parseImports(lexical, literals)
	// Selectors read through the import table, so an aliased import
	// matches the canonical patterns.
	resolved := selectorName.ReplaceAllStringFunc(lexical, func(m string) string {
		ident := selectorName.FindStringSubmatch(m)[1]
		if name, ok := imports[ident]; ok {
			return name + "."
		}
		return ident + "."
	})

	out := hints(path, resolved)
	for _, p := range extraOwners {
		if p.re.MatchString(resolved) {
			out = append(out, p.owner)
		}
	}
	for _, x := range extras {
		if regexp.MustCompile(`\b` + regexp.QuoteMeta(x.name) + `\s*\.\s*\w+`).MatchString(resolved) {
			out = append(out, x.owner)
		}
	}
	if contextType.MatchString(resolved) || (slices.Contains(dot, "context") && matchesUnqualified(dotContext, resolved)) {
		out = append(out, "go-context")
	}
	for _, p := range dotOwners {
		if slices.Contains(dot, p.pkg) && matchesUnqualified(p.re, resolved) {
			out = append(out, p.owner)
		}
	}

	// A %w literal matters only as an argument of an actual Errorf call.
	var fmtNames []string
	for alias, name := range imports {
		if name == "fmt" {
			fmtNames = append(fmtNames, alias)
		}
	}
	if _, ok := imports["fmt"]; !ok {
		fmtNames = append(fmtNames, "fmt")
	}
	for _, alias := range fmtNames {
		call := regexp.MustCompile(`\b` + regexp.QuoteMeta(alias) + `\.Errorf\s*\(`)
		for _, loc := range call.FindAllStringIndex(lexical, -1) {
			if literalAfter(lexical, loc[1], literals, func(v string) bool { return strings.Contains(v, "%w") }) {
				out = append(out, "go-error-handling")
			}
		}
	}
	out = append(out, stubOwners(lexical, literals, comments)...)
	return dedupe(out)
}

// importOwner is an import outside the canonical table whose use names an
// owner: a crypto package other than a content hash, or x/time/rate.
type importOwner struct{ name, owner string }

// parseImports reads the import declarations of lexical, the source with
// comments and literals blanked, whose literals sit at their offsets. It
// returns the name each import is used by, keyed by its local name; the
// canonical names of dot imports; and the imports that name an owner.
func parseImports(lexical string, literals map[int]string) (imports map[string]string, dot []string, extras []importOwner) {
	imports = map[string]string{}
	positions := make([]int, 0, len(literals))
	for pos := range literals {
		positions = append(positions, pos)
	}
	sort.Ints(positions)
	// Locate import spans in literal- and comment-free code, then recover
	// their string tokens: grouped, aliased, and raw imports alike.
	for _, kw := range importWord.FindAllStringIndex(lexical, -1) {
		start, end := kw[1], -1
		if groupedOpen.MatchString(lexical[start:]) {
			if i := strings.Index(lexical[start:], ")"); i >= 0 {
				end = start + i + 1
			}
		} else if i := strings.Index(lexical[start:], "\n"); i >= 0 {
			end = start + i
		}
		if end <= start {
			end = len(lexical)
		}
		for _, pos := range positions {
			value := literals[pos]
			if pos < start || pos >= end || (value[0] != '"' && value[0] != '`') {
				continue
			}
			pkg := value[1 : len(value)-1]
			parts := importSplit.Split(lexical[start:pos], -1)
			alias := pkg[strings.LastIndex(pkg, "/")+1:]
			if m := importAlias.FindStringSubmatch(parts[len(parts)-1]); m != nil {
				alias = m[1]
			}
			switch alias {
			case ".":
				if name, ok := canonical[pkg]; ok {
					dot = append(dot, name)
				}
			case "_":
			default:
				owner := ""
				if strings.HasPrefix(pkg, "crypto/") && pkg != "crypto/sha256" && pkg != "crypto/sha3" && pkg != "crypto/sha512" {
					owner = "go-security"
				}
				if pkg == "x/time/rate" || pkg == "golang.org/x/time/rate" {
					owner = "go-resilience"
				}
				name, ok := canonical[pkg]
				if !ok {
					name = pick(owner != "", alias, "foreign")
				}
				imports[alias] = name
				if owner != "" {
					extras = append(extras, importOwner{name, owner})
				}
			}
		}
	}
	return imports, dot, extras
}

// stubOwners names the owners a stub's documented contract points at: a
// panic("not implemented") body whose comments describe bounded workers,
// retries, or a cache with a ttl.
func stubOwners(lexical string, literals map[int]string, comments []string) []string {
	stub := false
	for _, loc := range panicCall.FindAllStringIndex(lexical, -1) {
		if literalAfter(lexical, loc[1], literals, func(v string) bool { return v == `"not implemented"` || v == "`not implemented`" }) {
			stub = true
		}
	}
	contract := strings.Join(comments, " ")
	var owners []string
	if stub && boundedWorkers.MatchString(contract) {
		owners = append(owners, "go-concurrency")
	}
	if stub && retryContract.MatchString(contract) {
		owners = append(owners, "go-resilience")
	}
	// A cache stub (key, ttl, a shared miss) is go-performance when the task
	// is to implement one. Prompt-only: an edit of finished code does not gain
	// a gate requirement.
	if stub && cacheContract.MatchString(contract) {
		owners = append(owners, "go-performance")
	}
	return owners
}

// pick is a when cond holds, else b.
func pick(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

// fileHints is promptHints over whole files, for the prompt note: the owners
// in first-seen order. An unreadable file is skipped.
func fileHints(paths []string) []string {
	var out []string
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		out = append(out, promptHints(p, strings.ToValidUTF8(string(data), "�"))...)
	}
	return dedupe(out)
}

// gateParse prints, one per line, what go-code-routing.sh needs from a
// payload: event, session, agent, tool, skill, file path, the owners the
// edited text points at, and the edit key.
func gateParse() {
	d, ok := readEvent()
	if !ok {
		return
	}
	agent := d.str("agent_id")
	if d.str("agent_type") == "fork" {
		agent = ""
	}
	ti := d.toolInput()
	skill := pick(ti.str("skill") != "", ti.str("skill"), ti.str("name"))
	path := ti.str("file_path")
	// A Read loads a skill only whole: a later offset, or a limit short of
	// the file, is a peek and records nothing.
	if d.str("tool_name") == "Read" && (truthy(ti["offset"]) || truthy(ti["limit"])) {
		offset, okOffset := intValue(ti["offset"])
		limit, okLimit := intValue(ti["limit"])
		lines := -1
		if data, err := os.ReadFile(path); err == nil {
			lines = strings.Count(string(data), "\n")
			if len(data) > 0 && data[len(data)-1] != '\n' {
				lines++
			}
		}
		if !okOffset || !okLimit || offset > 1 || (limit != 0 && (lines < 0 || limit < lines)) {
			path = ""
		}
	}
	parts := []string{pick(ti.str("new_string") != "", ti.str("new_string"), ti.str("content"))}
	if edits, ok := ti["edits"].([]any); ok {
		for _, e := range edits {
			m, _ := e.(map[string]any)
			parts = append(parts, event(m).str("new_string"))
		}
	}
	text := strings.Join(parts, "\n")
	// A retry of the same edit after a block has the same key, while
	// parallel edits of different files in one message have different keys.
	sum := sha256.Sum256([]byte(path + "\x00" + text))
	for _, v := range []string{d.str("hook_event_name"), d.str("session_id"), agentDir(agent),
		d.str("tool_name"), skill, path, strings.Join(hints(path, text), " "), hex.EncodeToString(sum[:])[:16]} {
		fmt.Println(oneLine(v))
	}
}

// truthy is a JSON value that counts as set: not null, false, 0, or "".
func truthy(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case json.Number:
		f, err := x.Float64()
		return err != nil || f != 0
	case string:
		return x != ""
	}
	return true
}

// intValue is an absent value as 0, or a JSON integer; ok is false for any
// other value.
func intValue(v any) (int, bool) {
	switch x := v.(type) {
	case nil:
		return 0, true
	case bool:
		if x {
			return 1, true
		}
		return 0, true
	case json.Number:
		n, err := x.Int64()
		return int(n), err == nil
	}
	return 0, false
}
