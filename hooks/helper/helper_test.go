package main

import (
	"slices"
	"testing"
)

// The tokenizer and the redirect filter replaced Python's shlex and a
// regular expression; the wants are what those produced.
func TestShellTokens(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want []string
	}{
		{"go vet ./... 2>&1 | tail -5", []string{"go", "vet", "./...", "2>", "&", "1", "|", "tail", "-5"}},
		{`echo "a b" 'c d' e\ f`, []string{"echo", "a b", "c d", "e f"}},
		{`bash -c "go vet . && gofmt -l ."`, []string{"bash", "-c", "go vet . && gofmt -l ."}},
		{"a;b&&c||d", []string{"a", ";", "b", "&&", "c", "||", "d"}},
		{`echo "x\"y" "a\b"`, []string{"echo", `x"y`, `a\b`}},
		{"cmd # comment\nnext", []string{"cmd", "next"}},
		{"ab#c d", []string{"ab"}},
		{"x>out y", []string{"x>out", "y"}},
		{"''", []string{""}},
	}
	for _, tt := range tests {
		got, err := shellTokens(tt.in)
		if err != nil || !slices.Equal(got, tt.want) {
			t.Errorf("shellTokens(%q) = %q, %v; want %q", tt.in, got, err, tt.want)
		}
	}
	if got, err := shellTokens(`unclosed "q`); err == nil {
		t.Errorf("shellTokens(unclosed quote) = %q, want an error", got)
	}
}

func TestWithoutOutputRedirects(t *testing.T) {
	t.Parallel()
	tests := []struct{ in, want string }{
		{"go vet . 2>&1", "go vet .  "},
		{"go test >log.txt .", "go test   ."},
		{"a2>file", "a2 "},
		{`echo ">x" > y`, `echo ">x"  `},
		{"cmd &>> all.log tail", "cmd   tail"},
		{"x >&- y", "x   y"},
		{`x >   "a b" z`, "x   z"},
	}
	for _, tt := range tests {
		if got := withoutOutputRedirects(tt.in); got != tt.want {
			t.Errorf("withoutOutputRedirects(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestCommandForms(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want []string
	}{
		{"go test -race ./... 2>&1 | tail -20", []string{"check"}},
		{"timeout -k 10 60 go vet .", []string{"check"}},
		{"env -C /tmp GOFLAGS= go -C x build ./...", []string{"check"}},
		{`bash -c "cd x && golangci-lint run ./..."`, []string{"check"}},
		{"bash /p/skills/go-performance/scripts/bench-compare.sh --verdict /tmp/s ./pkg", []string{"verdict"}},
		{"bash /p/skills/go-performance/scripts/bench-compare.sh --verdict=/tmp/s ./pkg", nil},
		{"bash /p/hooks/go-check-receipt.sh --gate /r /pkg", []string{"verifier"}},
		{"bash /elsewhere/go-check-receipt.sh --gate /r /pkg", nil},
		{"go mod tidy; ls", nil},
	}
	for _, tt := range tests {
		var got []string
		for _, f := range commandForms(tt.in, "/p", 0) {
			got = append(got, f.kind)
		}
		if !slices.Equal(got, tt.want) {
			t.Errorf("commandForms(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestHedged(t *testing.T) {
	t.Parallel()
	tests := []struct {
		line string
		want bool
	}{
		{"This change does not make Parse faster; it only removes one copy.", true},
		{"The pool isn't measurably faster here.", true},
		{"speedup: skipped (no verdict)", true},
		{"This doesn't change the API and makes Parse 3x faster.", false},
		{"It no longer allocates and runs 40% faster.", false},
		{"Run is now 3x faster.", false},
	}
	for _, tt := range tests {
		if got := hedged(tt.line); got != tt.want {
			t.Errorf("hedged(%q) = %t, want %t", tt.line, got, tt.want)
		}
	}
}

func TestNamesGo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		prompt string
		want   bool
	}{
		{"Rewrite this Python service in Go to cut latency", true},
		{"Port the CLI to Go on Linux", true},
		{"Go ahead and add a changelog entry", false},
		{"Go through the README and fix the typos", false},
		{"Fix the bug. Go to the handler first.", false},
		{"Напиши це на Go", true},
		{"edit internal/x/y.go", true},
		{"a Gopher drawing", false},
	}
	for _, tt := range tests {
		if got := namesGo(tt.prompt); got != tt.want {
			t.Errorf("namesGo(%q) = %t, want %t", tt.prompt, got, tt.want)
		}
	}
}

func TestMentionedRouter(t *testing.T) {
	t.Parallel()
	tests := []struct{ prompt, want string }{
		{"use go-code for this", "go-code"},
		{"/opsx:apply add-auth /go-code", "go-code"},
		{"$go-code-refactor the store", "go-code-refactor"},
		{"read skills/go-code/SKILL.md", ""},
		{"go-code-reviewer is a name", ""},
		{"see go-code.md", ""},
		{"(golang-skills:go-code-review)", "go-code-review"},
	}
	for _, tt := range tests {
		if got := mentionedRouter(tt.prompt); got != tt.want {
			t.Errorf("mentionedRouter(%q) = %q, want %q", tt.prompt, got, tt.want)
		}
	}
}

func TestStripHostTags(t *testing.T) {
	t.Parallel()
	in := "<ide_opened_file>The user opened golang-skills/MODERNIZATION.md</ide_opened_file>study this <agent-message>go test</agent-message>"
	if got, want := stripHostTags(in, true), " study this  "; got != want {
		t.Errorf("stripHostTags(agent) = %q, want %q", got, want)
	}
	if got, want := stripHostTags("<ide_selection>x", false), "<ide_selection>x"; got != want {
		t.Errorf("stripHostTags(unclosed) = %q, want %q", got, want)
	}
}

func TestPromptHints(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, src string
		want      []string
	}{
		{"aliased import", "package p\n\nimport h \"net/http\"\n\nfunc F(w h.ResponseWriter) {}\n", []string{"go-http"}},
		{"import alone", "package p\n\nimport _ \"database/sql\"\n", nil},
		{"grpc server", "package p\n\nimport \"google.golang.org/grpc\"\n\nvar srv = grpc.NewServer()\n", []string{"go-http"}},
		{"wrapped error", "package p\n\nimport \"fmt\"\n\nfunc F(err error) error { return fmt.Errorf(\"f: %w\", err) }\n", []string{"go-error-handling"}},
		{"retry stub", "package p\n\n// Get retries with backoff and honors Retry-After.\nfunc Get() { panic(\"not implemented\") }\n", []string{"go-resilience"}},
		{"Retry-After alone", "package p\n\n// Get reads the Retry-After header.\nfunc Get() { panic(\"not implemented\") }\n", nil},
		{"cache stub", "package p\n\n// Cache keeps each answer for ttl.\nfunc Get() { panic(\"not implemented\") }\n", []string{"go-performance"}},
		{"content hash", "package p\n\nimport \"crypto/sha256\"\n\nvar _ = sha256.Sum256\n", nil},
	}
	for _, tt := range tests {
		if got := promptHints("/x/p.go", tt.src); !slices.Equal(got, tt.want) {
			t.Errorf("%s: promptHints = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestEncode(t *testing.T) {
	t.Parallel()
	got := encode(object{{"decision", "block"}, {"reason", "a <b> ї"}, {"n", []string{"x", "y"}}})
	want := "{\"decision\": \"block\", \"reason\": \"a <b> \\u0457\", \"n\": [\"x\", \"y\"]}"
	if got != want {
		t.Errorf("encode = %s, want %s", got, want)
	}
}
