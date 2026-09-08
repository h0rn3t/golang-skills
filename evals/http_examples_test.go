package evals_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHTTPHandlerExampleRejectsTrailingJSON(t *testing.T) {
	code := exampleBlock(t, "skills/go-http/SKILL.md", "## Handler Shape")
	runExampleTest(t, `package example
import ("encoding/json"; "io"; "net/http"; "net/http/httptest"; "strings"; "testing"; "context")
var _ = io.EOF
type createUserRequest struct { Name string }
func (r createUserRequest) validate() error { return nil }
func (r createUserRequest) toUser() string { return r.Name }
type store struct { calls int }
func (s *store) Create(_ context.Context, name string) (string, error) { s.calls++; return name, nil }
type Server struct { store *store }
func (s *Server) writeError(w http.ResponseWriter, _ *http.Request, _ error) { w.WriteHeader(500) }
func writeJSON(w http.ResponseWriter, code int, _ any) { w.WriteHeader(code) }
`+code+`
func TestBody(t *testing.T) {
 for _, tt := range []struct { body string; status, calls int }{
  {"{\"Name\":\"ok\"}", 201, 1},
  {"{\"Name\":\"ok\"} \n\t", 201, 1},
  {"{\"Name\":\"ok\"} garbage", 400, 0},
  {"{\"Name\":\"ok\"} {}", 400, 0},
  {"{\"Name\":\"ok\"}" + strings.Repeat(" ", 1<<20), 400, 0},
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
