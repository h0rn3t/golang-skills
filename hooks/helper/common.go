package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf16"
)

// event is a hook payload as Claude Code sends it on stdin.
type event map[string]any

// readEvent decodes the payload; ok is false when stdin is not a JSON object.
func readEvent() (event, bool) {
	dec := json.NewDecoder(os.Stdin)
	dec.UseNumber()
	var d map[string]any
	if err := dec.Decode(&d); err != nil || d == nil {
		return nil, false
	}
	return d, true
}

// str is the string field key, or "" when it is absent or not a string.
func (e event) str(key string) string {
	s, _ := e[key].(string)
	return s
}

// toolInput is the tool_input object, or nil.
func (e event) toolInput() event {
	m, _ := e["tool_input"].(map[string]any)
	return m
}

// oneLine keeps a value on the single output line the scripts read.
func oneLine(s string) string {
	return strings.ReplaceAll(s, "\n", " ")
}

// sessionID is the session_id, or "default" when the host sent none.
func (e event) sessionID() string {
	if s := e.str("session_id"); s != "" {
		return s
	}
	return "default"
}

// agentDir is the agent_id made safe for a path component: at most 64
// characters of [A-Za-z0-9_-], every other character replaced by "_".
func agentDir(agent string) string {
	var b strings.Builder
	for _, r := range agent {
		if b.Len() == 64 {
			break
		}
		if r < unicode.MaxASCII && (r == '_' || r == '-' || unicode.IsLetter(r) || unicode.IsDigit(r)) {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	return b.String()
}

// skippedAgent reports a subagent that writes no Go: the plugin's go-verify
// and the host's Explore, claude-code-guide, and statusline-setup. A plugin
// install names the agent "golang-skills:go-verify".
func skippedAgent(agentType string) bool {
	name := strings.ToLower(agentType[strings.LastIndex(agentType, ":")+1:])
	switch name {
	case "go-verify", "explore", "claude-code-guide", "statusline-setup":
		return true
	}
	return false
}

// hasGoFiles reports go.mod or a .go file within two directory levels of
// root, giving up after 200 directories. Hidden directories, vendor,
// node_modules, and testdata are not searched.
func hasGoFiles(root string) bool {
	if fi, err := os.Stat(root); root == "" || err != nil || !fi.IsDir() {
		return false
	}
	seen := 0
	var walk func(dir string, depth int) (found, stop bool)
	walk = func(dir string, depth int) (bool, bool) {
		seen++
		if seen > 200 {
			return false, true
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return false, false
		}
		var dirs []string
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() {
				if !strings.HasPrefix(name, ".") && name != "vendor" && name != "node_modules" && name != "testdata" {
					dirs = append(dirs, name)
				}
				continue
			}
			if name == "go.mod" || strings.HasSuffix(name, ".go") {
				return true, false
			}
		}
		if depth >= 2 {
			return false, false
		}
		for _, name := range dirs {
			if found, stop := walk(filepath.Join(dir, name), depth+1); found || stop {
				return found, stop
			}
		}
		return false, false
	}
	found, _ := walk(root, 0)
	return found
}

// pluginName is the name field of a plugin.json, or "" when it is unreadable
// or not a plain name.
func pluginName(manifest string) string {
	data, err := os.ReadFile(manifest)
	if err != nil {
		return ""
	}
	var m struct{ Name string }
	if json.Unmarshal(data, &m) != nil || !regexp.MustCompile(`^[A-Za-z0-9_.-]+$`).MatchString(m.Name) {
		return ""
	}
	return m.Name
}

// field is one key of an ordered JSON object.
type field struct {
	key   string
	value any
}

// object is a JSON object whose keys keep the order they are listed in.
type object []field

func (o object) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('{')
	for i, f := range o {
		if i > 0 {
			b.WriteByte(',')
		}
		k, err := marshal(f.key)
		if err != nil {
			return nil, err
		}
		v, err := marshal(f.value)
		if err != nil {
			return nil, err
		}
		b.Write(k)
		b.WriteByte(':')
		b.Write(v)
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

func marshal(v any) ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(b.Bytes(), []byte("\n")), nil
}

// encode writes v as one line of ASCII JSON with a space after every ":" and
// ",": the hooks' messages and the tests that read them use that spelling.
func encode(v any) string {
	raw, err := marshal(v)
	if err != nil {
		panic(fmt.Sprintf("encode %T: %v", v, err))
	}
	var b strings.Builder
	inString, escaped := false, false
	for _, r := range string(raw) {
		switch {
		case escaped:
			escaped = false
		case inString && r == '\\':
			escaped = true
		case r == '"':
			inString = !inString
		case !inString && (r == ',' || r == ':'):
			b.WriteRune(r)
			b.WriteByte(' ')
			continue
		}
		if r < 0x80 {
			b.WriteRune(r)
			continue
		}
		if r > 0xffff {
			hi, lo := utf16.EncodeRune(r)
			fmt.Fprintf(&b, `\u%04x\u%04x`, hi, lo)
			continue
		}
		fmt.Fprintf(&b, `\u%04x`, r)
	}
	return b.String()
}

// contextJSON is the answer by which SubagentStart and PostToolUse hand text
// to the model: a subagent never sees plain stdout.
func contextJSON(eventName, text string) string {
	return encode(object{{"hookSpecificOutput", object{
		{"hookEventName", eventName},
		{"additionalContext", text},
	}}})
}

// resolve is the absolute path with every existing symlink resolved; the
// part of the path that does not exist is kept as written.
func resolve(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	if real, err := filepath.EvalSymlinks(abs); err == nil {
		return real
	}
	parent := filepath.Dir(abs)
	if parent == abs {
		return abs
	}
	return filepath.Join(resolve(parent), filepath.Base(abs))
}

// dedupe drops repeated entries and keeps the first of each.
func dedupe(list []string) []string {
	seen := make(map[string]bool, len(list))
	out := list[:0:0]
	for _, s := range list {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
