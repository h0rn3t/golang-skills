package evals_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestLocalRedirectExample(t *testing.T) {
	code := exampleBlock(t, "skills/go-security/references/INJECTION.md", "### Local redirects")
	runExampleTest(t, `package example
import ("net/http"; "net/http/httptest"; "net/url"; "strings"; "testing")
`+code+`
func TestRedirect(t *testing.T) {
 for _, tt := range []struct { next string; allowed bool }{
  {"/", true}, {"/account?tab=profile#name", true}, {"/профіль", true},
  {"/search?q=https://example.test", true}, {"/a%20b", true},
  {"", false}, {"account", false}, {"https://attacker.example", false},
  {"//attacker.example", false}, {"/\\attacker.example", false},
  {"/\t/attacker.example", false}, {"/\r/attacker.example", false},
  {"/\n/attacker.example", false}, {"/\x00/attacker.example", false},
  {"/\x7f/attacker.example", false}, {"/ /attacker.example", false},
 } {
  t.Run(tt.next, func(t *testing.T) {
   if got := localRedirect(tt.next); got != tt.allowed {
    t.Fatalf("localRedirect(%q)=%t, want %t", tt.next, got, tt.allowed)
   }
   if !tt.allowed { return }
   r := httptest.NewRequest("GET", "https://trusted.example/login", nil)
   w := httptest.NewRecorder()
   http.Redirect(w, r, tt.next, http.StatusSeeOther)
   u, err := url.Parse(w.Header().Get("Location"))
   if err != nil { t.Fatal(err) }
   if got := r.URL.ResolveReference(u).Host; got != "trusted.example" {
    t.Errorf("redirect(%q) host=%q, want trusted.example", tt.next, got)
   }
  })
 }
}
`)
}

func TestContextHandlerFailureBoundaries(t *testing.T) {
	code := exampleBlock(t, "skills/go-context/references/PATTERNS.md", "## Respecting Cancellation in HTTP Handlers")
	runExampleTest(t, `package example
import ("context"; "encoding/json"; "errors"; "fmt"; "log/slog"; "net/http"; "net/http/httptest"; "strings"; "testing")
var slowOperation func(context.Context) (string, error)
`+code+`
func TestHandler(t *testing.T) {
 for _, tt := range []struct { name string; err error; cancelRequest bool; status int }{
  {"success", nil, false, 200},
  {"internal details", errors.New("password=example-secret host=internal"), false, 500},
  {"dependency canceled", fmt.Errorf("child: %w", context.Canceled), false, 500},
  {"request canceled", context.Canceled, true, 0},
 } {
  t.Run(tt.name, func(t *testing.T) {
   var logs strings.Builder
   slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
   slowOperation = func(context.Context) (string, error) { return "ok", tt.err }
   ctx, cancel := context.WithCancel(t.Context())
   defer cancel()
   if tt.cancelRequest { cancel() }
   w := httptest.NewRecorder()
   w.Code = 0
   r := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
   handler(w, r)
   status := w.Code
   if !tt.cancelRequest { status = w.Result().StatusCode }
   if status != tt.status { t.Errorf("handler(%s) status=%d, want %d", tt.name, status, tt.status) }
   if strings.Contains(w.Body.String(), "example-secret") || strings.Contains(w.Body.String(), "host=internal") {
    t.Errorf("handler(%s) розкрив внутрішні дані: %q", tt.name, w.Body.String())
   }
   if tt.cancelRequest && w.Body.Len() != 0 { t.Errorf("скасований запит отримав body=%q", w.Body.String()) }
   wantLogged := tt.err != nil && !tt.cancelRequest
   if wantLogged && !strings.Contains(logs.String(), tt.err.Error()) {
    t.Errorf("handler(%s) не залогував деталь server-side: %q", tt.name, logs.String())
   }
   if !wantLogged && logs.Len() != 0 {
    t.Errorf("handler(%s) залогував зайве: %q", tt.name, logs.String())
   }
   if tt.err == nil {
    var got string
    if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || got != "ok" {
     t.Errorf("handler(success) body=%q, error=%v, want JSON string ok", w.Body.String(), err)
    }
   }
  })
 }
}
`)
}

func TestRefactorLintStatus(t *testing.T) {
	for _, tt := range []struct {
		name, script, status, findings string
		exitCode                       int
	}{
		{"clean", "exit 0", "pass", "0", 0},
		{"findings", "echo 'sample.go:3:1: example diagnostic (test)'; exit 1", "fail", "1", 1},
		{"config error", "echo 'Error: unsupported config version' >&2; exit 3", "unavailable", "n/a", 3},
		{"failure without findings", "echo 'Error: failed to load packages' >&2; exit 1", "unavailable", "n/a", 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := refactorToolFixture(t, tt.script)
			script := filepath.Join(repoRoot(t), "skills/go-code-refactor/scripts/verify-refactor.sh")
			out := runCommandInDir(t, dir, 0, "bash", script, "--json", "baseline", "./...")
			var got struct {
				Status   string `json:"lint_status"`
				ExitCode *int   `json:"lint_exit_code"`
				Findings string `json:"lint_findings"`
				LogPath  string `json:"lint_log_path"`
				Passed   bool   `json:"passed"`
			}
			if err := json.Unmarshal(out, &got); err != nil {
				t.Fatalf("baseline JSON: %v\n%s", err, out)
			}
			if got.Status != tt.status || got.ExitCode == nil || *got.ExitCode != tt.exitCode || got.Findings != tt.findings {
				t.Errorf("baseline(%s)=%s, want lint %s, exit %d, findings %s", tt.name, out, tt.status, tt.exitCode, tt.findings)
			}
			if !got.Passed {
				t.Error("інформаційний lint змінив результат основних перевірок")
			}
			if got.LogPath == "" {
				t.Fatal("не повернуто шлях до журналу lint")
			}
			if _, err := os.Stat(filepath.Join(dir, got.LogPath)); err != nil {
				t.Fatalf("lint log %q: %v", got.LogPath, err)
			}
		})
	}
}

func TestRefactorTruncation(t *testing.T) {
	dir := refactorToolFixture(t, "exit 0")
	outDir := filepath.Join(dir, ".refactor-verify")
	if err := os.Mkdir(outDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{
		"baseline.summary": "build: pass\nvet: pass\ntest: pass\n",
		"after.summary":    "build: fail\nvet: fail\ntest: fail\n",
	} {
		if err := os.WriteFile(filepath.Join(outDir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	script := filepath.Join(repoRoot(t), "skills/go-code-refactor/scripts/verify-refactor.sh")
	run := func(t *testing.T, mode, limit string) (truncated bool, lines int) {
		t.Helper()
		exitCode := 1
		if mode == "leaks" {
			exitCode = 3
		}
		out := runCommandInDir(t, dir, exitCode, "bash", script, "--json", "--limit", limit, mode)
		var got struct {
			Truncated bool   `json:"truncated"`
			Diff      string `json:"diff"`
			Output    string `json:"output"`
		}
		if err := json.Unmarshal(out, &got); err != nil {
			t.Fatalf("%s JSON: %v\n%s", mode, err, out)
		}
		text := got.Diff + got.Output
		if text == "" {
			t.Fatalf("%s --limit %s: порожній вивід", mode, limit)
		}
		return got.Truncated, len(strings.Split(text, "\n"))
	}
	for _, mode := range []string{"diff", "leaks"} {
		unlimited, full := run(t, mode, "0")
		if unlimited {
			t.Errorf("%s --limit 0: truncated=true, хоча ліміт вимкнено", mode)
		}
		if full < 3 {
			t.Fatalf("%s: фікстура дає %d рядків, замало щоб перевірити межу", mode, full)
		}
		// Рівно на межі скорочення немає: інакше повний блоб повідомляється
		// як урізаний, і виклик шукає вивід, якого не бракує.
		for _, limit := range []int{1, full - 1, full, full + 1} {
			t.Run(fmt.Sprintf("%s/%d", mode, limit), func(t *testing.T) {
				gotTruncated, gotLines := run(t, mode, strconv.Itoa(limit))
				wantTruncated, wantLines := limit < full, min(limit, full)
				if gotTruncated != wantTruncated || gotLines != wantLines {
					t.Errorf("%s --limit %d: truncated=%t з %d рядками; want %t з %d (повний блоб — %d)",
						mode, limit, gotTruncated, gotLines, wantTruncated, wantLines, full)
				}
			})
		}
	}
}

// TestRefactorApplyLimitCounting перевіряє лічильник рядків самої apply_limit,
// витягнутої з відвантажуваного скрипта. Через режими цей дефект недосяжний:
// усі троє передають їй результат `$(...)`, який зрізає кінцевий "\n".
func TestRefactorApplyLimitCounting(t *testing.T) {
	script := filepath.Join(repoRoot(t), "skills/go-code-refactor/scripts/verify-refactor.sh")
	const harness = `set -uo pipefail
eval "$(awk '/^apply_limit\(\) \{/,/^\}/' "$1")"
LIMIT="$2"
text=$(printf '%b' "$3"; printf x)
apply_limit "${text%x}"
printf '%s' "$TRUNCATED"`
	for _, tt := range []struct {
		text, limit string
		truncated   bool
	}{
		{`a\nb`, "2", false},
		{`a\nb`, "1", true},
		{`a\nb\n`, "2", false}, // кінцевий "\n" не є зайвим рядком
		{`a\nb\n`, "1", true},
		{`a\n`, "1", false},
		{`a\nb\nc\n`, "2", true},
	} {
		t.Run(fmt.Sprintf("%s/limit=%s", tt.text, tt.limit), func(t *testing.T) {
			out := runCommandInDir(t, t.TempDir(), 0, "bash", "-c", harness, "harness", script, tt.limit, tt.text)
			if got := string(out) == "true"; got != tt.truncated {
				t.Errorf("apply_limit(%q, LIMIT=%s) truncated=%s, want %t", tt.text, tt.limit, out, tt.truncated)
			}
		})
	}
}

// refactorToolFixture ізолює статуси зовнішніх інструментів для тестів shell-контракту.
func refactorToolFixture(t *testing.T, lintScript string) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "bin")
	if err := os.Mkdir(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{
		"bin/go":            "#!/bin/sh\ncase \"$1\" in\nversion) echo 'go version go1.27.1 test/test';;\ntest) printf '=== RUN   TestExample\\n--- PASS: TestExample (0.00s)\\nPASS\\n';;\nesac\n",
		"bin/golangci-lint": "#!/bin/sh\n" + lintScript + "\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example\n\ngo 1.27\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return dir
}
