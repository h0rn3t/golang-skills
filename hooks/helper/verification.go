package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	sessionName = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	assignment  = regexp.MustCompile(`^[A-Za-z_]\w*=`)

	speedClaim = regexp.MustCompile(`(?i)\b(?:faster|speed-?ups?|sped up)\b|[−-]\s?\d+(?:\.\d+)?\s?%\s*(?:sec/op|ns/op|time|latency|cpu)\b|\b(?:sec/op|ns/op|latency|run ?time|cpu time)\b[^.\n]{0,30}?[−-]\s?\d+(?:\.\d+)?\s?%`)
	// A line that says the comparison was skipped is the report
	// go-performance asks for, not a claim.
	speedSkipped = regexp.MustCompile(`(?i)\b(?:skipped|unmeasured|not measured|no verdict|without a verdict|not faster|no faster|no (?:\w+ )?(?:speed-?up|gain))\b`)
	negation     = regexp.MustCompile(`(?i)\b(?:not|never|no)\b|n['’]t\b`)
	// A negation disclaims a claim only about the claim itself: at most three
	// words between, none of them joining another clause. "does not make
	// Parse faster" is a disclaimer; "doesn't change the API and makes Parse
	// 3x faster" is a claim.
	negatedClaim = regexp.MustCompile(`(?i)^((?:\s+[^\s.;,]+){0,3}?)\s+(?:faster|speed-?ups?|sped up)\b`)
	clauseJoin   = regexp.MustCompile(`(?i)\b(?:and|but|while|so|then|yet)\b`)

	passClaim = regexp.MustCompile(`(?i)\b(gofmt|vet|fix|lint|test)\s*(?:[:=]\s*)?pass\s*\([^)]*(?:\bhook\b|\breceipt\b)`)
)

// hedged reports a line whose speed words are a disclaimer, not a claim.
func hedged(line string) bool {
	if speedSkipped.MatchString(line) {
		return true
	}
	for _, loc := range negation.FindAllStringIndex(line, -1) {
		if m := negatedClaim.FindStringSubmatch(line[loc[1]:]); m != nil && !clauseJoin.MatchString(m[1]) {
			return true
		}
	}
	return false
}

// verification is go-verification-routing.sh: PreToolUse on Bash asks for
// the receipt verifier before a repeated check of the last Go edit, and Stop
// holds back a report whose hook-pass or speed claims lack evidence. It
// returns the exit code.
func verification() int {
	d, ok := readEvent()
	if !ok {
		return 0
	}
	stopping := d.str("hook_event_name") == "Stop"
	if (!stopping && (d.str("hook_event_name") != "PreToolUse" || d.str("tool_name") != "Bash")) || os.Getenv("GOLANG_SKILLS_VERIFICATION_GATE") == "off" {
		return 0
	}
	session := d.sessionID()
	agent, agentOK := d["agent_id"].(string)
	command, commandOK := d.toolInput()["command"].(string)
	if (d["agent_id"] != nil && !agentOK) || (d.toolInput()["command"] != nil && !commandOK) || !sessionName.MatchString(session) {
		return 0
	}
	agent = agentDir(agent)
	root := os.Getenv("GOLANG_SKILLS_VERIFICATION_ROOT")
	base := stateBase()
	// Each agent of a session keeps its own loads and edits; a fork starts
	// from its parent's loads.
	routing := filepath.Join(base, "routing", session)
	loadedDir := routing
	if agent != "" && d.str("agent_type") != "fork" {
		loadedDir = filepath.Join(routing, "agent-"+agent)
	}
	loaded := loadedSkills(filepath.Join(loadedDir, "loaded"))
	if !loaded["go-code"] && !loaded["go-code-refactor"] {
		return 0
	}
	if !isFile(filepath.Join(root, "skills/go-linting/SKILL.md")) {
		return 0
	}
	key := session
	if agent != "" {
		key += "/" + agent
	}
	sum := sha256.Sum256([]byte(key))
	state := filepath.Join(base, "verification", hex.EncodeToString(sum[:]))
	if !isFile(filepath.Join(state, "latest")) {
		return 0
	}
	if stopping {
		return stopCheck(d, state)
	}

	kinds := commandForms(command, root, 0)
	var checks []form
	for _, f := range kinds {
		if f.kind == "verdict" {
			_ = touch(filepath.Join(state, "verdict")) // a lost mark only re-asks for a verdict
			continue
		}
		checks = append(checks, f)
	}
	if len(checks) == 0 {
		return 0
	}
	var latest, attempted generation
	if readJSON(filepath.Join(state, "latest"), &latest) != nil || readOptionalJSON(filepath.Join(state, "attempted"), &attempted) != nil {
		fmt.Fprintln(os.Stderr, "golang-skills verification gate: receipt state unavailable; run required checks directly and do not credit hooks.")
		return 0
	}
	skill := "go-linting"
	if ns := pluginName(filepath.Join(root, ".claude-plugin/plugin.json")); ns != "" {
		skill = ns + ":" + skill
	}
	recipe := shellJoin([]string{"bash", filepath.Join(root, "hooks/go-check-receipt.sh"), "--gate", latest.RunDir, latest.Package})
	if !loaded["go-linting"] {
		fmt.Fprintf(os.Stderr, "golang-skills verification gate: before verifying Go work, load the `%s` skill (Skill tool, name `%s`; without a Skill tool, Read %s whole), then run:\n%s\n",
			skill, skill, filepath.Join(root, "skills/go-linting/SKILL.md"), recipe)
		return 2
	}
	anyCheck := slices.ContainsFunc(checks, func(f form) bool { return f.kind == "check" })
	if (verifierFirst(checks, latest) && !anyCheck) || attempted.RunDir == latest.RunDir {
		return 0
	}
	fmt.Fprintf(os.Stderr, "golang-skills verification gate: before repeating checks for the last Go edit, run this verifier once (Bash tool):\n%s\n"+
		"A redirect or a pipe that only displays output (2>&1, | tail, | head) is still that command. Another verification command in the same call is not.\n"+
		"Only hook_credit=true checks may be reported as pass (hook). A failed verification allows direct checks; the next Go edit requires another attempt.\n", recipe)
	return 2
}

// loadedSkills is the set of skills a loaded file records, one per line.
func loadedSkills(path string) map[string]bool {
	loaded := map[string]bool{}
	if data, err := os.ReadFile(path); err == nil {
		for l := range strings.SplitSeq(string(data), "\n") {
			loaded[l] = true
		}
	}
	return loaded
}

// verifierFirst reports a call of the current generation's verifier before
// any other check in the command.
func verifierFirst(forms []form, latest generation) bool {
	for _, f := range forms {
		if f.kind == "check" {
			return false
		}
		switch {
		case len(f.args) == 3 && f.args[0] == "--gate":
			if resolve(f.args[1]) == latest.RunDir && resolve(f.args[2]) == latest.Package {
				return true
			}
		case len(f.args) >= 3:
			if isFile(f.args[0]) && filepath.Dir(resolve(f.args[0])) == latest.RunDir && resolve(f.args[1]) == latest.Package {
				return true
			}
		}
	}
	return false
}

// stopCheck blocks a final message that claims a speedup no verdict
// measured, or a hook pass the current receipts do not credit.
func stopCheck(d event, state string) int {
	text := d.str("last_assistant_message")
	var reasons []string
	// A speedup claimed after an edit of production Go needs a
	// bench-compare.sh --verdict call somewhere in the session.
	if os.Getenv("GOLANG_SKILLS_SPEED_GATE") != "off" && isFile(filepath.Join(state, "production")) && !isFile(filepath.Join(state, "verdict")) {
		for line := range strings.SplitSeq(text, "\n") {
			if hedged(line) {
				continue
			}
			if found := speedClaim.FindString(line); found != "" {
				reasons = append(reasons, "golang-skills speed gate: the report claims a speedup (\""+found+"\") but no bench-compare.sh --verdict ran in this session. Run the verdict as go-performance says (The verdict decides) and report its line, or report the comparison as skipped and drop the claim.")
				break
			}
		}
	}
	claims := map[string]bool{}
	for _, m := range passClaim.FindAllStringSubmatch(text, -1) {
		claims[strings.ToLower(m[1])] = true
	}
	if len(claims) == 0 && len(reasons) == 0 {
		return 0
	}
	credits := map[string]bool{}
	var latest, attempted generation
	if readJSON(filepath.Join(state, "latest"), &latest) == nil && readOptionalJSON(filepath.Join(state, "attempted"), &attempted) == nil && attempted.RunDir == latest.RunDir {
		var claimed []string
		for _, c := range attempted.Credits {
			if claims[c] {
				claimed = append(claimed, c)
			}
		}
		for _, c := range currentCredits(latest.RunDir, claimed) {
			credits[c] = true
		}
	}
	var unsupported, current []string
	for c := range claims {
		if !credits[c] {
			unsupported = append(unsupported, c)
		}
	}
	for c := range credits {
		current = append(current, c)
	}
	sort.Strings(unsupported)
	sort.Strings(current)
	if len(unsupported) > 0 {
		reasons = append(reasons, "golang-skills verification gate: unsupported hook-pass report for "+strings.Join(unsupported, ", ")+
			". Current verified credits: "+pick(len(current) > 0, strings.Join(current, ", "), "none")+
			". Correct the report to observed/unavailable or verify the current receipt. A direct rerun is pass (direct), not pass (hook).")
	}
	if len(reasons) > 0 {
		reason := strings.Join(reasons, "\n")
		if stopActive, _ := d["stop_hook_active"].(bool); stopActive {
			fmt.Println(encode(object{{"continue", false}, {"stopReason", reason}}))
		} else {
			fmt.Println(encode(object{{"decision", "block"}, {"reason", reason}}))
		}
	}
	return 0
}

// form is one simple command the gate recognizes: a receipt "verifier" call
// with its arguments, a bench-compare "verdict", or a verification "check".
type form struct {
	kind string
	args []string
}

// commandForms splits a Bash command into simple commands and classifies
// them, looking through env, command, timeout, and bash -c wrappers.
func commandForms(text, root string, depth int) []form {
	if depth > 3 {
		return nil
	}
	tokens, err := shellTokens(withoutOutputRedirects(text))
	if err != nil {
		return nil
	}
	var result []form
	var segment []string
	for _, token := range append(tokens, ";") {
		if !isPunctuation(token) {
			segment = append(segment, token)
			continue
		}
		args := unwrap(segment)
		segment = nil
		if len(args) == 0 {
			continue
		}
		name := filepath.Base(args[0])
		script := ""
		if name == "bash" && len(args) > 1 {
			script = filepath.Base(args[1])
		}
		switch {
		case name == "bash" && len(args) > 2 && args[1] == "-c":
			result = append(result, commandForms(args[2], root, depth+1)...)
		case script == "go-check-receipt.sh":
			if resolve(args[1]) == resolve(filepath.Join(root, "hooks/go-check-receipt.sh")) {
				result = append(result, form{"verifier", args[2:]})
			}
		case script == "bench-compare.sh" || (name == "bench-compare.sh"):
			if slices.Contains(args, "--verdict") { // bench-compare.sh rejects --verdict=DIR
				result = append(result, form{kind: "verdict"})
			}
		case name == "go":
			sub := args[1:]
			if len(sub) > 0 && sub[0] == "-C" {
				sub = sub[min(2, len(sub)):]
			} else if len(sub) > 0 && strings.HasPrefix(sub[0], "-C=") {
				sub = sub[1:]
			}
			if len(sub) > 0 && slices.Contains([]string{"build", "vet", "test", "fix"}, sub[0]) {
				result = append(result, form{kind: "check"})
			}
		case name == "gofmt" || name == "govulncheck" || (name == "golangci-lint" && len(args) > 1 && args[1] == "run"):
			result = append(result, form{kind: "check"})
		case script == "verify-refactor.sh" || script == "pre-review.sh":
			result = append(result, form{kind: "check"})
		}
	}
	return result
}

// unwrap drops what runs a command without being it: variable assignments,
// env with its options, command, and timeout with its options and duration.
func unwrap(args []string) []string {
	for len(args) > 0 && assignment.MatchString(args[0]) {
		args = args[1:]
	}
	if len(args) > 0 && filepath.Base(args[0]) == "env" {
		args = args[1:]
		for len(args) > 0 && (strings.HasPrefix(args[0], "-") || assignment.MatchString(args[0])) {
			args = args[min(optionWidth(args[0], "-C", "--chdir", "-u", "--unset"), len(args)):]
		}
	}
	if len(args) > 0 && args[0] == "command" {
		args = args[1:]
	}
	if len(args) > 0 && (filepath.Base(args[0]) == "timeout" || filepath.Base(args[0]) == "gtimeout") {
		args = args[1:]
		for len(args) > 0 && strings.HasPrefix(args[0], "-") {
			args = args[min(optionWidth(args[0], "-k", "--kill-after", "-s", "--signal"), len(args)):]
		}
		args = args[min(1, len(args)):]
	}
	return args
}

// optionWidth is 2 for an option that takes a separate value, else 1.
func optionWidth(opt string, withValue ...string) int {
	if slices.Contains(withValue, opt) {
		return 2
	}
	return 1
}

func isPunctuation(token string) bool {
	return token != "" && strings.Trim(token, ";&|()\n") == ""
}

// withoutOutputRedirects removes unquoted output redirections (>file, 2>&1,
// >> log) and keeps quoted words, bash -c bodies included, for the parser.
func withoutOutputRedirects(text string) string {
	var b strings.Builder
	for i := 0; i < len(text); {
		if n := quotedLen(text[i:]); n > 0 {
			b.WriteString(text[i : i+n])
			i += n
			continue
		}
		if n := redirectLen(text, i); n > 0 {
			b.WriteByte(' ')
			i += n
			continue
		}
		b.WriteByte(text[i])
		i++
	}
	return b.String()
}

// quotedLen is the length of a "…" or '…' string or a backslash escape at the
// start of s, or 0.
func quotedLen(s string) int {
	if s == "" {
		return 0
	}
	switch s[0] {
	case '"':
		for i := 1; i < len(s); i++ {
			switch s[i] {
			case '\\':
				if i+1 >= len(s) || s[i+1] == '\n' {
					return 0
				}
				i++
			case '"':
				return i + 1
			}
		}
	case '\'':
		if j := strings.IndexByte(s[1:], '\''); j >= 0 {
			return j + 2
		}
	case '\\':
		if len(s) > 1 && s[1] != '\n' {
			_, size := utf8.DecodeRuneInString(s[1:])
			return 1 + size
		}
	}
	return 0
}

// redirectLen is the length of an output redirection starting at text[i]:
// an optional fd number or & (not inside a word), > or >>, then &N or a word.
func redirectLen(text string, i int) int {
	j := i
	prev, _ := utf8.DecodeLastRuneInString(text[:i])
	if i == 0 || !isWordRune(prev) {
		switch {
		case text[j] == '&':
			j++
		case text[j] >= '0' && text[j] <= '9':
			for j < len(text) && text[j] >= '0' && text[j] <= '9' {
				j++
			}
		}
		if j >= len(text) || text[j] != '>' {
			j = i
		}
	}
	if j >= len(text) || text[j] != '>' {
		return 0
	}
	j++
	if j < len(text) && text[j] == '>' {
		j++
	}
	if j < len(text) && text[j] == '&' {
		k := j + 1
		for k < len(text) && (text[k] >= '0' && text[k] <= '9' || text[k] == '-') {
			k++
		}
		if k > j+1 {
			return k - i
		}
	}
	for j < len(text) && (text[j] == ' ' || text[j] == '\t') {
		j++
	}
	start := j
	for j < len(text) {
		if n := quotedLen(text[j:]); n > 0 {
			j += n
			continue
		}
		r, size := utf8.DecodeRuneInString(text[j:])
		if unicode.IsSpace(r) || strings.ContainsRune(";&|()<>\"'\\", r) {
			break
		}
		j += size
	}
	if j == start {
		return 0
	}
	return j - i
}

// shellTokens splits a command the way a POSIX shell lexer does for the
// parts the gate reads: quotes, backslash escapes, comments, and runs of the
// operators ;&|() and newline as their own tokens.
func shellTokens(s string) ([]string, error) {
	const (
		space = iota
		word
		operator
	)
	var tokens []string
	var tok strings.Builder
	state, quoted := space, false
	emit := func() {
		if tok.Len() > 0 || quoted {
			tokens = append(tokens, tok.String())
		}
		tok.Reset()
		quoted = false
	}
	isOp := func(c byte) bool { return strings.IndexByte(";&|()\n", c) >= 0 }
	for i := 0; i < len(s); i++ {
		c := s[i]
		if state == operator {
			if isOp(c) {
				tok.WriteByte(c)
				continue
			}
			emit()
			state = space
		}
		switch {
		case c == ' ' || c == '\t' || c == '\r':
			if state == word {
				emit()
				state = space
			}
		case c == '#':
			if state == word {
				emit()
			}
			state = space
			if j := strings.IndexByte(s[i:], '\n'); j >= 0 {
				i += j
			} else {
				i = len(s)
			}
		case isOp(c):
			if state == word {
				emit()
			}
			tok.WriteByte(c)
			state = operator
		case c == '\\':
			if i+1 >= len(s) {
				return nil, errors.New("no escaped character")
			}
			i++
			tok.WriteByte(s[i])
			state = word
		case c == '\'':
			j := strings.IndexByte(s[i+1:], '\'')
			if j < 0 {
				return nil, errors.New("no closing quotation")
			}
			tok.WriteString(s[i+1 : i+1+j])
			i += j + 1
			quoted, state = true, word
		case c == '"':
			i++
			for ; i < len(s) && s[i] != '"'; i++ {
				if s[i] == '\\' && i+1 < len(s) {
					if s[i+1] != '"' && s[i+1] != '\\' {
						tok.WriteByte('\\')
					}
					i++
				}
				tok.WriteByte(s[i])
			}
			if i >= len(s) {
				return nil, errors.New("no closing quotation")
			}
			quoted, state = true, word
		default:
			tok.WriteByte(c)
			state = word
		}
	}
	emit()
	return tokens, nil
}

// shellJoin quotes each argument for a POSIX shell when it needs it.
func shellJoin(args []string) string {
	safe := regexp.MustCompile(`^[\w@%+=:,./-]+$`)
	quoted := make([]string, len(args))
	for i, a := range args {
		if a != "" && safe.MatchString(a) {
			quoted[i] = a
		} else {
			quoted[i] = "'" + strings.ReplaceAll(a, "'", `'"'"'`) + "'"
		}
	}
	return strings.Join(quoted, " ")
}

// stateBase is where the hooks keep session state.
func stateBase() string {
	if d := os.Getenv("CLAUDE_PLUGIN_DATA"); d != "" {
		return d
	}
	return pick(os.Getenv("TMPDIR") != "", os.Getenv("TMPDIR"), "/tmp") + "/golang-skills-hooks"
}

func isFile(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.Mode().IsRegular()
}

func touch(p string) error {
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	return f.Close()
}

func readJSON(p string, v any) error {
	data, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// readOptionalJSON is readJSON for a file that may not exist yet.
func readOptionalJSON(p string, v any) error {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return nil
	}
	return readJSON(p, v)
}
