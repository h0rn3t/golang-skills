package evals_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestHTTPClientExampleEnforcesWholeResponseLimit(t *testing.T) {
	code := exampleBlock(t, "skills/go-http/SKILL.md", "### Bounded Response Bodies")
	runExampleTest(t, `package example
import ("encoding/json"; "fmt"; "io"; "net/http"; "strings"; "testing")
func decodeResponse(resp *http.Response, dst any) error {
`+code+`
}
func TestResponseLimit(t *testing.T) {
 const object = "{\"state\":\"ready\"}"
 for _, tt := range []struct { name, body string; wantErr bool }{
  {"small", object, false},
  {"exact limit", object + strings.Repeat(" ", (64<<10)-len(object)), false},
  {"one byte over", object + strings.Repeat(" ", (64<<10)-len(object)+1), true},
  {"second object", object + " {}", true},
  {"trailing junk", object + " junk", true},
 } {
  t.Run(tt.name, func(t *testing.T) {
   resp := &http.Response{Body: io.NopCloser(strings.NewReader(tt.body))}
   defer resp.Body.Close()
   var got struct { State string }
   err := decodeResponse(resp, &got)
   if (err != nil) != tt.wantErr { t.Errorf("decodeResponse(%d bytes) error = %v, want error presence %t", len(tt.body), err, tt.wantErr) }
   if !tt.wantErr && got.State != "ready" { t.Errorf("state = %q, want ready", got.State) }
  })
 }
}
`)
}

func TestHTTPHandlerExampleRejectsTrailingJSON(t *testing.T) {
	code := exampleBlock(t, "skills/go-http/SKILL.md", "## Handler Shape")
	runExampleTest(t, `package example
import (json "encoding/json/v2"; "errors"; "log/slog"; "net/http"; "net/http/httptest"; "strings"; "testing"; "context")
var ErrConflict = errors.New("conflict")
type store struct { calls int; err error }
func (s *store) Create(_ context.Context, name string) (string, error) { s.calls++; return name, s.err }
type Server struct { store *store }
`+code+`
func TestStoreErrors(t *testing.T) {
 for _, tt := range []struct { err error; status int }{
  {ErrConflict, 409},
  {errors.New("disk full"), 500},
 } {
  s := &Server{store: &store{err: tt.err}}
  w := httptest.NewRecorder()
  s.handleCreateUser(w, httptest.NewRequest("POST", "/users", strings.NewReader("{\"name\":\"ok\"}")))
  if w.Code != tt.status || strings.Contains(w.Body.String(), tt.err.Error()) {
   t.Errorf("handler with store error %v: status=%d body=%q, want %d without the error text", tt.err, w.Code, w.Body.String(), tt.status)
  }
 }
}
func TestClientGone(t *testing.T) {
 ctx, cancel := context.WithCancel(t.Context())
 cancel()
 s := &Server{store: &store{err: context.Canceled}}
 w := httptest.NewRecorder()
 s.handleCreateUser(w, httptest.NewRequestWithContext(ctx, "POST", "/users", strings.NewReader("{\"name\":\"ok\"}")))
 if w.Body.Len() != 0 || w.Header().Get("Content-Type") != "" {
  t.Errorf("handler after client disconnect wrote status=%d body=%q, want nothing", w.Code, w.Body.String())
 }
 dctx, dcancel := context.WithTimeout(t.Context(), 0)
 defer dcancel()
 <-dctx.Done()
 s = &Server{store: &store{err: context.DeadlineExceeded}}
 w = httptest.NewRecorder()
 s.handleCreateUser(w, httptest.NewRequestWithContext(dctx, "POST", "/users", strings.NewReader("{\"name\":\"ok\"}")))
 if w.Code != 500 || w.Body.Len() == 0 {
  t.Errorf("handler under an expired request deadline: status=%d body=%q, want a written 500", w.Code, w.Body.String())
 }
}
func TestBody(t *testing.T) {
 for _, tt := range []struct { body string; status, calls int }{
  {"{\"name\":\"ok\"}", 201, 1},
  {"{\"name\":\"ok\"} \n\t", 201, 1},
  {"{\"name\":\"ok\"} garbage", 400, 0},
  {"{\"name\":\"ok\"}]", 400, 0},
  {"{\"name\":\"ok\"}}", 400, 0},
  {"{\"name\":\"ok\"} {}", 400, 0},
  {"{\"name\":\"ok\",\"x\":1}", 400, 0},
  {"{\"Name\":\"ok\"}", 400, 0},
  {"{\"name\":\"ok\",\"name\":\"dup\"}", 400, 0},
  {"null", 400, 0},
  {"{}", 400, 0},
  {"{\"name\":\"ok\"}" + strings.Repeat(" ", 1<<20), 400, 0},
 } {
  t.Run(tt.body[:min(len(tt.body), 30)], func(t *testing.T) {
   s := &Server{store: &store{}}
   w := httptest.NewRecorder()
   s.handleCreateUser(w, httptest.NewRequest("POST", "/users", strings.NewReader(tt.body)))
   if w.Code != tt.status || s.store.calls != tt.calls {
    t.Errorf("handler: status=%d calls=%d, want %d and %d", w.Code, s.store.calls, tt.status, tt.calls)
   }
  })
 }
}
`)
}

// The HEAD-as-405 recipe runs on the routes the Routing example registers:
// every method a path does not serve is a 405 whose Allow lists what it does
// serve, HEAD included, and the served methods still reach their handlers.
func TestHTTPExampleHeadNotAllowed(t *testing.T) {
	routing := exampleBlock(t, "skills/go-http/SKILL.md", "## Routing (Go 1.22+)")
	recipe := exampleBlock(t, "skills/go-http/SKILL.md", "Only when the contract makes `HEAD` a 405")
	handler := regexp.MustCompile(`s\.handle\w+`)
	var routes []string
	for _, line := range strings.Split(routing, "\n") {
		if strings.HasPrefix(line, "mux.HandleFunc(") {
			routes = append(routes, handler.ReplaceAllString(line, "served"))
		}
	}
	decl, registrations, ok := strings.Cut(recipe, "\nmux.HandleFunc(")
	if !ok || len(routes) == 0 {
		t.Fatalf("routing example registers %d routes; HEAD recipe registrations found: %t", len(routes), ok)
	}
	runExampleTest(t, `package example
import ("net/http"; "net/http/httptest"; "testing")
`+decl+`
func served(http.ResponseWriter, *http.Request) {}
func newMux() *http.ServeMux {
 mux := http.NewServeMux()
`+strings.Join(routes, "\n")+`
mux.HandleFunc(`+registrations+`
 return mux
}
func TestMethods(t *testing.T) {
 mux := newMux()
 for _, tt := range []struct { method, path, allow string; code int }{
  {"GET", "/users/1", "", 200}, {"HEAD", "/users/1", "GET", 405}, {"POST", "/users/1", "GET", 405}, {"DELETE", "/users/1", "GET", 405},
  {"GET", "/users", "", 200}, {"POST", "/users", "", 200}, {"HEAD", "/users", "GET, POST", 405}, {"DELETE", "/users", "GET, POST", 405},
  {"GET", "/", "", 200}, {"HEAD", "/", "GET", 405}, {"POST", "/", "GET", 405},
 } {
  rec := httptest.NewRecorder()
  mux.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))
  if rec.Code != tt.code || rec.Header().Get("Allow") != tt.allow {
   t.Errorf("%s %s = %d Allow=%q, want %d Allow=%q", tt.method, tt.path, rec.Code, rec.Header().Get("Allow"), tt.code, tt.allow)
  }
 }
}
`)
}

func TestLoggingResponseWriterExample(t *testing.T) {
	code := exampleBlock(t, "skills/go-logging/references/LOGGING-PATTERNS.md", "## HTTP Request Logging Middleware")
	_, wrapper, ok := strings.Cut(code, "type responseWriter struct")
	if !ok {
		t.Fatal("logging example has no responseWriter declaration")
	}
	runExampleTest(t, `package example
import ("net/http"; "net/http/httptest"; "testing")
type responseWriter struct`+wrapper+`
func TestStatus(t *testing.T) {
 for _, tt := range []struct { name string; write func(*responseWriter); want int }{
  {"explicit", func(w *responseWriter) { w.WriteHeader(201); w.WriteHeader(500) }, 201},
  {"implicit", func(w *responseWriter) { _, _ = w.Write([]byte("ok")); w.WriteHeader(500) }, 200},
  {"flush", func(w *responseWriter) {
   if err := http.NewResponseController(w).Flush(); err != nil { t.Errorf("Flush: %v", err) }
   w.WriteHeader(500)
  }, 200},
 } {
  t.Run(tt.name, func(t *testing.T) {
   rec := httptest.NewRecorder()
   w := &responseWriter{ResponseWriter: rec, status: http.StatusOK}
   tt.write(w)
   if rec.Code != tt.want || w.status != tt.want { t.Errorf("wire=%d log=%d, want %d", rec.Code, w.status, tt.want) }
  })
 }
}
// A real server is needed for informational responses; ResponseRecorder
// treats its first WriteHeader as final even for 103.
func TestInformational(t *testing.T) {
 for _, final := range []int{200, 201} {
  t.Run(http.StatusText(final), func(t *testing.T) {
   srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
    rw.WriteHeader(103)
    if final == 200 { _, _ = rw.Write([]byte("ok")) } else { rw.WriteHeader(final) }
    if rw.status != final { t.Errorf("log status=%d, want %d", rw.status, final) }
   }))
   defer srv.Close()
   resp, err := srv.Client().Get(srv.URL)
   if err != nil { t.Fatal(err) }
   defer resp.Body.Close()
   if resp.StatusCode != final { t.Errorf("wire status=%d, want %d", resp.StatusCode, final) }
  })
 }
}
`)
}

func TestLoggingFanOutExample(t *testing.T) {
	code := exampleBlock(t, "skills/go-logging/references/LOGGING-PATTERNS.md", "### Multi-Handler (Fan-Out)")
	// Substitute only the output destinations so the documented setup itself runs.
	code = strings.NewReplacer("os.Stdout", "&stdout", "os.Stderr", "&stderr").Replace(code)
	runExampleTest(t, `package example
import ("bytes"; "encoding/json"; "log/slog"; "strings"; "testing")
func TestFanOut(t *testing.T) {
 var stdout, stderr bytes.Buffer
`+code+`
 logger = logger.With("service", "orders").WithGroup("request")
 logger.Debug("hidden")
 logger.Info("accepted", "id", 1)
 logger.Error("failed", "id", 2)
 dec := json.NewDecoder(&stdout)
 for _, want := range []string{"INFO", "ERROR"} {
  var record struct { Level, Service string; Request struct { ID int } }
  if err := dec.Decode(&record); err != nil { t.Fatal(err) }
  if record.Level != want || record.Service != "orders" || record.Request.ID == 0 { t.Errorf("record=%+v, want %s with shared fields", record, want) }
 }
 if got := stderr.String(); strings.Count(got, "\n") != 1 || !strings.Contains(got, "level=ERROR") || !strings.Contains(got, "service=orders") || !strings.Contains(got, "request.id=2") {
  t.Errorf("stderr=%q, want only error with shared fields", got)
 }
}
`)
}

// The redaction type in go-logging's What NOT to Log is the form go-security
// routes to: no stdlib handler, attribute form, or group may print the secret.
func TestLoggingExampleRedaction(t *testing.T) {
	code := exampleBlock(t, "skills/go-logging/SKILL.md", "## What NOT to Log")
	runExampleTest(t, `package example
import ("bytes"; "log/slog"; "strings"; "testing")
`+code+`
func TestRedacted(t *testing.T) {
 tok := Token("s3cr3t")
 var buf bytes.Buffer
 for _, h := range []slog.Handler{slog.NewJSONHandler(&buf, nil), slog.NewTextHandler(&buf, nil)} {
  logger := slog.New(h)
  logger.Info("login", "token", tok)
  logger.With("token", tok).Info("with")
  logger.Info("group", slog.Group("auth", "token", tok))
 }
 if got := buf.String(); strings.Contains(got, "s3cr3t") || strings.Count(got, "[REDACTED]") != 6 {
  t.Errorf("log output leaks the token or misses a redaction:\n%s", got)
 }
}
`)
}

func exampleBlock(t *testing.T, path, heading string) string {
	t.Helper()
	text := readFile(t, filepath.Join(repoRoot(t), path))
	_, section, ok := strings.Cut(text, heading)
	if !ok {
		t.Fatalf("%s missing %q", path, heading)
	}
	_, code, ok := strings.Cut(section, "```go\n")
	if !ok {
		t.Fatalf("%s has no Go example after %q", path, heading)
	}
	code, _, ok = strings.Cut(code, "```")
	if !ok {
		t.Fatalf("%s has an unclosed Go example", path)
	}
	return code
}

func runExampleTest(t *testing.T, code string) {
	t.Helper()
	dir := t.TempDir()
	for name, content := range map[string]string{"go.mod": "module example\n\ngo 1.27\n", "example_test.go": code} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.CommandContext(t.Context(), "go", "test", "-count=1", "-race", "./...")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("example go test: %v\n%s", err, out)
	}
}
