package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// The prompt patterns read Ukrainian and Russian as well as English. RE2's
// \b and \w are ASCII-only, so word boundaries are spelled out: wb before a
// word, we after it, and w for one word character.
const (
	w  = `[\pL\pN_]`
	wb = `(?:^|[^\pL\pN_])`
	we = `(?:[^\pL\pN_]|$)`
)

// words compiles a case-insensitive alternation that must stand as whole
// words; `\w` in alts is a Unicode word character.
func words(alts string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)` + wb + `(?:` + strings.ReplaceAll(alts, `\w`, w) + `)` + we)
}

// wordStart is words without the closing boundary: the alternation may run on
// into a longer word.
func wordStart(alts string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)` + wb + `(?:` + strings.ReplaceAll(alts, `\w`, w) + `)`)
}

func isWordRune(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// stripHostTags removes the context an IDE host or a subagent report adds to
// the prompt: <ide_opened_file>…</ide_opened_file>, <ide_selection>, and, with
// agentMessages, <agent-message>. Only the words the user typed are classified.
func stripHostTags(prompt string, agentMessages bool) string {
	open := regexp.MustCompile(`<(ide_[a-z_]+)\b[^>]*>`)
	if agentMessages {
		open = regexp.MustCompile(`<(ide_[a-z_]+|agent-message)\b[^>]*>`)
	}
	var b strings.Builder
	last, from := 0, 0
	for from < len(prompt) {
		loc := open.FindStringSubmatchIndex(prompt[from:])
		if loc == nil {
			break
		}
		start, tagEnd := from+loc[0], from+loc[1]
		closing := "</" + prompt[from+loc[2]:from+loc[3]] + ">"
		end := strings.Index(prompt[tagEnd:], closing)
		if end < 0 {
			from = start + 1
			continue
		}
		b.WriteString(prompt[last:start])
		b.WriteByte(' ')
		last = tagEnd + end + len(closing)
		from = last
	}
	b.WriteString(prompt[last:])
	return b.String()
}

var (
	ladderSlash = regexp.MustCompile(`^\s*/(?:[\pL\pN_.-]+:)?go-code\s+(lite|full|ultra)` + we)
	ladderWord  = regexp.MustCompile(`(?i)` + wb + `(?:(lite|full|ultra)[ -]mode|режим\s+(lite|full|ultra))` + we)
)

// ladderParse prints the event, the session, and a restraint level the
// prompt sets (or an empty line) for go-restraint-ladder.sh; nothing when the
// hook has nothing to say.
func ladderParse() {
	d, ok := readEvent()
	if !ok {
		return
	}
	eventName := d.str("hook_event_name")
	word := ""
	switch eventName {
	case "UserPromptSubmit":
		prompt, ok := d["prompt"].(string)
		if !ok {
			return
		}
		// IDE context is not the user typing a level word, and a prefix of it
		// would hide a leading /go-code. `/go-code-refactor ultra` sets
		// nothing: a refactor runs at full.
		prompt = stripHostTags(prompt, false)
		m := ladderSlash.FindStringSubmatch(prompt)
		if m == nil {
			m = ladderWord.FindStringSubmatch(prompt)
		}
		if m == nil {
			return
		}
		for _, g := range m[1:] {
			if g != "" {
				word = strings.ToLower(g)
				break
			}
		}
	case "SessionStart", "SubagentStart":
		if skippedAgent(d.str("agent_type")) || !hasGoFiles(d.str("cwd")) {
			return
		}
	default:
		return
	}
	fmt.Println(eventName)
	fmt.Println(oneLine(d.sessionID()))
	fmt.Println(word)
}

// subagentParse prints three lines for go-subagent-routing.sh: the event when
// the hook should print its note (a session or a Go-writing subagent starting
// in a directory with Go), else an empty line; the session id; and the
// SessionStart source (startup, resume, clear, compact). The session and the
// source are printed without the note too: a compaction is recorded whatever
// the directory holds.
func subagentParse() {
	d, ok := readEvent()
	if !ok {
		return
	}
	eventName := pick(d.str("hook_event_name") != "", d.str("hook_event_name"), "SubagentStart")
	if eventName != "SubagentStart" && eventName != "SessionStart" {
		return
	}
	note := eventName
	// SessionStart has no agent; the filter is only for a subagent.
	if (eventName == "SubagentStart" && skippedAgent(d.str("agent_type"))) || !hasGoFiles(d.str("cwd")) {
		note = ""
	}
	source := ""
	if eventName == "SessionStart" {
		source = d.str("source")
	}
	fmt.Println(note)
	fmt.Println(oneLine(d.sessionID()))
	fmt.Println(oneLine(source))
}

var (
	slashRouter = regexp.MustCompile(`^\s*/(?:[A-Za-z0-9_.-]+:)?(go-code-refactor|go-code-review|go-code)` + we)
	namespace   = regexp.MustCompile(`^[A-Za-z0-9_.-]+:`)

	// `implement` counts as a verb only: "one implementation" in a design
	// question is not new code.
	newCode = words(`implement(?:s|ed|ing)?|write|add|create|build|stub\w*|not implemented|fill in|` +
		`реаліз\w*|реализ\w*|напиш\w*|напис\w*|дода\w*|добав\w*|створ\w*|созда\w*|допиш\w*|заполн\w*`)
	// Bare "slow" asks why, not for a change; "optimistic" and
	// "high-performance" are not requests.
	speedRequest = words(`faster|speed\s+(?:\w+\s+){0,3}?up|speed-?up|optimi[sz](?:e|es|ed|ing|ation)|` +
		`(?:reduce|cut|lower)\s+(?:the\s+|its\s+)?(?:latency|allocations?|allocs|cpu\s+time|memory\s+use)|` +
		`allocate\s+less|fewer\s+allocations|прискор\w*|оптиміз\w*|ускор\w*|оптимиз\w*`)
	humanLanguage = wordStart(`english|ukrainian|russian|german|french|spanish|polish|` +
		`англ\w*|українськ\w*|російськ\w*|німецьк\w*|русск\w*|украинск\w*|немецк\w*`)
	proofread = words(`grammar|proofread\w*`)
	translate = wordStart(`translat\w*|переклад\w*|перекла\w*|перевед\w*|перевод\w*`)
	codeNoun  = words(`function\w?|func|method\w?|package\w?|handler\w?|endpoint\w?|struct\w?|interface\w?|type\w?|` +
		`test\w*|stub\w?|bug\w?|panic|бод[иі]\w*|функці\w*|метод\w*|пакет\w*|хендлер\w*|обробник\w*|` +
		`структур\w*|інтерфейс\w*|тест\w*|заглушк\w*|баг\w*|функци\w*|обработчик\w*|эндпоинт\w*`)
	refactorWords = words(`refactor\w*|clean(?:\s|-)?up|simplif\w*|restructur\w*|moderni[sz]\w*|tidy(?:\s|-)?up|reads? better|` +
		`messy|mess|bloated|over-?engineer\w*|over-?abstract\w*|dead code|too long|hard to follow|monolith\w*|modulari[sz]\w*|` +
		`(?:more\s+|make\s+(?:it|this|the\s+code)\s+)readable|readability|unreadable|` +
		`split(?:ting)?\s+(?:this|the|that|it)\s+(?:[\pL\pN_-]+\s+)?(?:function|method|file|package|handler|type|struct)\w*|` +
		`into\s+smaller\s+(?:ones|functions|methods|pieces|parts|files|packages)|` +
		`extract\w*\s+(?:an?\s+|the\s+)?(?:helper|function|method|interface|package|type)|` +
		`рефактор\w*|спрост\w*|упрост\w*|почист\w*|переструктур\w*|модерніз\w*|модерниз\w*|монол[иі]т\w*|модуляриз\w*`)
	// An audit for what can stop existing is the over-engineering audit of
	// go-code-refactor (docs/RULE_OWNERSHIP.md), not a review.
	subtract = words(`over-?engineer\w*|over-?abstract\w*|dead code|what (?:can|could|should) (?:we|i|be) (?:delete|remove|cut)\w*`)
	work     = words(`implement(?:s|ed|ing)?|write|add|create|build|make|fix\w*|bug\w*|stub\w*|not implemented|fill in|` +
		`реаліз\w*|реализ\w*|напиш\w*|напис\w*|дода\w*|добав\w*|виправ\w*|исправ\w*|створ\w*|созда\w*|зроби|сделай|баг\w*|` +
		`почин\w*|поправ\w*|допиш\w*|заполн\w*`)
	navigate = words(`explain|find|locate|where|who calls|how does|show|trace|understand|` +
		`поясн\w*|объясн\w*|знайд\w*|найд\w*|покаж\w*|де|где|хто виклика\w*|кто вызыва\w*|` +
		`як працю\w*|как работа\w*`)
	questionStart = regexp.MustCompile(`(?i)^\s*(?:explain|find|locate|where|who|how|show|trace|understand|` +
		strings.ReplaceAll(`поясн\w*|объясн\w*|знайд\w*|найд\w*|покаж\w*|`, `\w`, w) + `(?:де|где|хто|кто|як|как)` + we + `)`)
	review     = words(`review\w*|audit\w*|рев[ьи]ю\w*|ревью\w*`)
	panicLine  = regexp.MustCompile(wb + `(?:panic|fatal error): `)
	goroutineN = regexp.MustCompile(wb + `goroutine \d+ \[`)

	// The language named by more than the bare word "Go".
	goNamed = regexp.MustCompile(wb + `(?:golang|[\pL\pN_][\pL\pN_./-]*\.go|go\.mod|goroutines?|gofmt|go (?:build|test|vet|run|mod|fix))` + we + "|```go" + we)
	// "Go ahead" and "Go back" are the verb anywhere; "Go to", "Go through",
	// "Go over", and "Go on" only where a sentence starts.
	goVerbAnywhere = regexp.MustCompile(`^\s+(?:ahead|back)` + we)
	goVerbOpening  = regexp.MustCompile(`^\s+(?:to|through|over|on)` + we)

	pathToken = regexp.MustCompile(`^(?:\.{1,2}/)?[\pL\pN_.-]+(?:/[\pL\pN_.-]+)*`)
)

// namesGo reports a prompt that names the Go language.
func namesGo(text string) bool {
	if goNamed.MatchString(text) {
		return true
	}
	for i := 0; ; {
		j := strings.Index(text[i:], "Go")
		if j < 0 {
			return false
		}
		start, end := i+j, i+j+2
		i = end
		before, _ := utf8.DecodeLastRuneInString(text[:start])
		after, _ := utf8.DecodeRuneInString(text[end:])
		if (start > 0 && isWordRune(before)) || (end < len(text) && isWordRune(after)) {
			continue
		}
		rest := text[end:]
		if goVerbAnywhere.MatchString(rest) || (goVerbOpening.MatchString(rest) && sentenceStart(text[:start])) {
			continue
		}
		return true
	}
}

// sentenceStart reports whether a word after prefix opens a sentence or a line.
func sentenceStart(prefix string) bool {
	trimmed := strings.TrimRightFunc(prefix, unicode.IsSpace)
	if trimmed == "" || strings.Contains(prefix[len(trimmed):], "\n") {
		return true
	}
	last, _ := utf8.DecodeLastRuneInString(trimmed)
	return strings.ContainsRune(".!?;:*-•>", last)
}

// mentionedRouter is a router named in the text: a word at a boundary (space,
// $, quote, bracket), possibly with / or a namespace, and not part of a path
// or of a longer name.
func mentionedRouter(prompt string) string {
	for i := 0; i < len(prompt); i++ {
		if i > 0 {
			prev, _ := utf8.DecodeLastRuneInString(prompt[:i])
			if !unicode.IsSpace(prev) && !strings.ContainsRune("`'\"([$", prev) {
				continue
			}
		}
		rest := strings.TrimPrefix(prompt[i:], "/")
		if ns := namespace.FindString(rest); ns != "" {
			rest = rest[len(ns):]
		}
		for _, name := range []string{"go-code-refactor", "go-code-review", "go-code"} {
			if !strings.HasPrefix(rest, name) {
				continue
			}
			after := rest[len(name):]
			next, size := utf8.DecodeRuneInString(after)
			if after != "" && (isWordRune(next) || next == '/' || next == '-') {
				continue
			}
			if next == '.' {
				if r, _ := utf8.DecodeRuneInString(after[size:]); isWordRune(r) {
					continue
				}
			}
			return name
		}
	}
	return ""
}

// targetFiles are the non-test Go files the prompt points at: a directory
// under cwd ("./dispatch", or a bare word naming one) or a .go path. With
// pathlikeOnly a bare word does not count.
func targetFiles(prompt, cwd string, pathlikeOnly bool) []string {
	if fi, err := os.Stat(cwd); cwd == "" || err != nil || !fi.IsDir() {
		return nil
	}
	inToken := func(r rune) bool { return isWordRune(r) || r == '.' || r == '/' || r == '-' }
	var files []string
	seen := map[string]bool{}
	for i := 0; i < len(prompt); {
		r, size := utf8.DecodeRuneInString(prompt[i:])
		prev, _ := utf8.DecodeLastRuneInString(prompt[:i])
		if !inToken(r) || (i > 0 && inToken(prev)) {
			i += size
			continue
		}
		tok := pathToken.FindString(prompt[i:])
		if tok == "" {
			i += size
			continue
		}
		i += len(tok)
		tok = strings.TrimRight(tok, ".,;:!?")
		if tok == "" || tok == "." || tok == ".." || seen[tok] {
			continue
		}
		seen[tok] = true
		pathlike := strings.Contains(tok, "/") || strings.HasSuffix(tok, ".go")
		if pathlikeOnly && !pathlike {
			continue
		}
		p := filepath.Join(cwd, tok)
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			files = append(files, goFiles(p)...)
		} else if pathlike && err == nil && fi.Mode().IsRegular() && strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go") {
			files = append(files, p)
		}
	}
	return files
}

// goFiles are the non-test .go files of dir, sorted.
func goFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if n := e.Name(); strings.HasSuffix(n, ".go") && !strings.HasSuffix(n, "_test.go") {
			out = append(out, filepath.Join(dir, n))
		}
	}
	sort.Strings(out)
	return out
}

// promptParse prints the verdict go-prompt-routing.sh acts on: session,
// skill, kind, mode (slash, mention, or note), "new" for new code, "speed"
// for a speed request, then up to 40 target files; nothing for a prompt that
// gets no note.
func promptParse() {
	d, ok := readEvent()
	if !ok {
		return
	}
	if e := d.str("hook_event_name"); e != "" && e != "UserPromptSubmit" {
		return
	}
	prompt, ok := d["prompt"].(string)
	if !ok && d["prompt"] != nil {
		return
	}
	prompt = stripHostTags(prompt, true)
	if strings.TrimSpace(prompt) == "" {
		return
	}
	cwd := d.str("cwd")
	emit := func(skill, kind, mode string) {
		fmt.Println(oneLine(d.sessionID()))
		fmt.Println(skill)
		fmt.Println(kind)
		fmt.Println(mode)
		fmt.Println(pick(newCode.MatchString(prompt), "new", ""))
		fmt.Println(pick(speedRequest.MatchString(prompt), "speed", ""))
		files := targetFiles(prompt, cwd, false)
		for _, p := range files[:min(len(files), 40)] {
			fmt.Println(p)
		}
	}
	// A slash command raises this event and no other: the host expands it,
	// inserts the router SKILL.md, and calls no tool.
	if m := slashRouter.FindStringSubmatch(prompt); m != nil {
		emit(m[1], "", "slash")
		return
	}
	if name := mentionedRouter(prompt); name != "" {
		emit(name, "", "mention")
		return
	}
	// Translating or copy-editing prose that mentions Go is not Go work.
	// Translating code into Go names no human language and still routes.
	if proofread.MatchString(prompt) || (translate.MatchString(prompt) && humanLanguage.MatchString(prompt)) {
		return
	}
	// A prompt that does not name Go counts as Go work only when the working
	// directory holds Go and the prompt names a code element, by a noun or by
	// a path to Go source; "add a line to the README" must not fire.
	goWork := namesGo(prompt) || (hasGoFiles(cwd) && (codeNoun.MatchString(prompt) || len(targetFiles(prompt, cwd, true)) > 0))
	if !goWork {
		return
	}
	question := navigate.MatchString(prompt) && (strings.HasSuffix(strings.TrimSpace(prompt), "?") || questionStart.MatchString(prompt))
	// A pasted runtime trace with no work verb asks what went wrong; review
	// is checked before the work verbs, since a review prompt asks for "the
	// fix" too.
	switch {
	case panicLine.MatchString(prompt) && goroutineN.MatchString(prompt) && !work.MatchString(prompt):
		emit("go-troubleshooting", "a pasted Go panic or goroutine trace", "note")
	case question:
	case subtract.MatchString(prompt):
		emit("go-code-refactor", "a behavior-preserving refactor", "note")
	case review.MatchString(prompt):
		emit("go-code-review", "a Go code review", "note")
	case refactorWords.MatchString(prompt):
		emit("go-code-refactor", "a behavior-preserving refactor", "note")
	case work.MatchString(prompt):
		emit("go-code", "Go code to write, implement, or fix", "note")
	case speedRequest.MatchString(prompt):
		emit("go-code", "a request to make Go code faster", "note")
	}
}

// gateJSON is a PreToolUse answer for exit 0. stop denies the edit and stops
// the session (continue: false), so the model does not repeat an edit the
// gate will not pass; notice only shows the text to the user.
func gateJSON(kind, text string) string {
	if kind != "stop" {
		return encode(object{{"systemMessage", text}})
	}
	return encode(object{
		{"continue", false},
		{"stopReason", text},
		{"hookSpecificOutput", object{
			{"hookEventName", "PreToolUse"},
			{"permissionDecision", "deny"},
			{"permissionDecisionReason", text},
		}},
	})
}

// editFile prints tool_input.file_path, parsed as JSON, so an escaped quote
// or a "file_path" string inside the edited content cannot redirect the hook.
func editFile() {
	if d, ok := readEvent(); ok {
		fmt.Print(strings.ReplaceAll(d.toolInput().str("file_path"), "\n", ""))
	}
}

// sessionKey prints the key of the verification state: the session, plus
// "/<agent_id>" in a subagent, as go-verification-routing.sh keys it.
func sessionKey() {
	d, ok := readEvent()
	if !ok {
		return
	}
	key := d.sessionID()
	if agent := agentDir(d.str("agent_id")); agent != "" {
		key += "/" + agent
	}
	fmt.Println(key)
}
