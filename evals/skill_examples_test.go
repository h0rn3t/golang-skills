package evals_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildExampleModule writes files into a fresh module and runs go vet and
// go test -race over it. It complements runExampleTest, which only knows a
// single test file: several skills show a complete package main or a
// package with its own test.
func buildExampleModule(t *testing.T, files map[string]string) {
	t.Helper()
	dir := t.TempDir()
	files["go.mod"] = "module example\n\ngo 1.27\n"
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"vet", "./..."}, {"test", "-count=1", "-race", "./..."}} {
		cmd := exec.CommandContext(t.Context(), "go", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("go %s: %v\n%s", args[0], err, out)
		}
	}
}

// The SSRF defenses are the security reference most likely to be copied
// verbatim: the allowlist must match whole labels, and the client must refuse
// the literal addresses an attacker reaches for before it connects, including
// the CGNAT and NAT64 ranges netip's Is* methods do not cover.
func TestSecurityExampleSSRFTarget(t *testing.T) {
	allow := exampleBlock(t, "skills/go-security/references/INJECTION.md", "## Outbound URLs (SSRF)")
	client := exampleBlock(t, "skills/go-security/references/INJECTION.md", "### Arbitrary public destinations")
	runExampleTest(t, `package example
import ("fmt"; "net"; "net/http"; "net/http/httptest"; "net/netip"; "slices"; "strings"; "syscall"; "testing"; "time")
`+allow+client+`
func TestAllowedHost(t *testing.T) {
	for host, want := range map[string]bool{"partner.example": true, "api.partner.example": true, "evilpartner.example": false, "partner.example.evil": false} {
		if got := allowedHost(host, "partner.example"); got != want {
			t.Errorf("allowedHost(%q, %q) = %t, want %t", host, "partner.example", got, want)
		}
	}
}
func TestPublicClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer srv.Close()
	c := publicClient()
	for _, raw := range []string{
		srv.URL, "http://169.254.169.254/latest", "http://[::1]:1/", "http://10.0.0.8:1/", "http://0.0.0.0:1/",
		"http://100.100.100.200/latest/meta-data/", // CGNAT: Alibaba Cloud metadata
		"http://[64:ff9b::a9fe:a9fe]/latest",       // NAT64 form of 169.254.169.254
		"http://[64:ff9b::a9fe:a9fe%25lo]/latest",  // the same with a zone
	} {
		resp, err := c.Get(raw)
		if err == nil {
			resp.Body.Close()
			t.Errorf("Get(%q) = nil error, want a blocked address range", raw)
			continue
		}
		if !strings.Contains(err.Error(), "blocked address range") {
			t.Errorf("Get(%q) error = %v, want a blocked address range", raw, err)
		}
	}
}
`)
}

// The TLS snippet must stay a valid tls.Config with only the field the
// reference says to set; the http.Server around it is go-http's.
func TestSecurityExampleTLSConfig(t *testing.T) {
	code := exampleBlock(t, "skills/go-security/references/SECRETS-AND-CRYPTO.md", "## TLS")
	runExampleTest(t, `package example
import ("crypto/tls"; "testing")
func config() *tls.Config {
`+code+`
	return cfg
}
func TestMinVersion(t *testing.T) {
	if got := config().MinVersion; got != tls.VersionTLS13 {
		t.Fatalf("MinVersion = %d, want %d", got, tls.VersionTLS13)
	}
}
`)
}

// go-resilience's only code: the retry loop stops on success, on a
// non-retryable error, when the budget is spent, and when ctx ends; a
// server-requested delay is a floor, a wait past the deadline stops at once,
// and zero attempts is an error, not a success.
func TestResilienceExampleRetry(t *testing.T) {
	code := exampleBlock(t, "skills/go-resilience/SKILL.md", "## Retry Invariants")
	runExampleTest(t, `package example
import ("context"; "errors"; "math/rand/v2"; "testing"; "testing/synctest"; "time")
`+code+`
func TestRetry(t *testing.T) {
	ctx := t.Context()
	transient := errors.New("transient")
	calls := 0
	err := retry(ctx, 3, time.Millisecond, 2*time.Millisecond, func(context.Context) (time.Duration, bool, error) {
		calls++
		if calls < 3 { return 0, true, transient }
		return 0, false, nil
	})
	if err != nil || calls != 3 { t.Fatalf("recovers within budget: err=%v calls=%d, want nil after 3", err, calls) }

	calls = 0
	err = retry(ctx, 3, time.Millisecond, 2*time.Millisecond, func(context.Context) (time.Duration, bool, error) {
		calls++; return 0, false, errors.New("permanent")
	})
	if err == nil || calls != 1 { t.Fatalf("non-retryable: err=%v calls=%d, want error after 1 call", err, calls) }

	calls = 0
	err = retry(ctx, 3, time.Millisecond, 2*time.Millisecond, func(context.Context) (time.Duration, bool, error) {
		calls++; return 0, true, transient
	})
	if !errors.Is(err, transient) || calls != 3 { t.Fatalf("budget: err=%v calls=%d, want transient after 3 calls", err, calls) }

	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	calls = 0
	err = retry(cancelled, 3, time.Second, time.Second, func(context.Context) (time.Duration, bool, error) {
		calls++; return 0, true, transient
	})
	if !errors.Is(err, context.Canceled) || !errors.Is(err, transient) || calls != 1 {
		t.Fatalf("cancel during backoff: err=%v calls=%d, want Canceled joined with the attempt error after 1 call", err, calls)
	}

	start := time.Now()
	_ = retry(ctx, 2, time.Millisecond, time.Millisecond, func(context.Context) (time.Duration, bool, error) {
		return 40 * time.Millisecond, true, transient
	})
	if d := time.Since(start); d < 40*time.Millisecond { t.Fatalf("Retry-After floor: waited %v, want >= 40ms", d) }

	synctest.Test(t, func(t *testing.T) {
		short, cancel := context.WithTimeout(t.Context(), 250*time.Millisecond)
		defer cancel()
		start := time.Now()
		calls := 0
		err := retry(short, 3, time.Millisecond, time.Second, func(context.Context) (time.Duration, bool, error) {
			calls++; return 60 * time.Second, true, transient
		})
		if d := time.Since(start); d != 0 || calls != 1 || !errors.Is(err, transient) || errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("Retry-After past the deadline: waited %v, calls=%d, err=%v; want no wait, 1 call, the attempt error", d, calls, err)
		}
	})

	calls = 0
	err = retry(ctx, 0, time.Millisecond, time.Millisecond, func(context.Context) (time.Duration, bool, error) {
		calls++; return 0, false, nil
	})
	if err == nil || calls != 0 { t.Fatalf("zero attempts: err=%v calls=%d, want an error and no call", err, calls) }
	_ = rand.N[int]
}
`)
}

// The slogtest example is a complete test file; it has to pass slogtest's
// own cases against the stdlib JSON handler it uses as a stand-in.
func TestLoggingExampleSlogtest(t *testing.T) {
	code := exampleBlock(t, "skills/go-logging/references/LOGGING-PATTERNS.md", "## Testing with slogtest")
	buildExampleModule(t, map[string]string{"handler_test.go": code})
}

// The structured-logging pair (good and bad) must at least compile; sloglint,
// not vet, is what flags the bad half.
func TestLoggingExampleStructured(t *testing.T) {
	code := exampleBlock(t, "skills/go-logging/SKILL.md", "## Structured Logging")
	buildExampleModule(t, map[string]string{"log.go": `package example
import ("fmt"; "log/slog")
func placed(orderID int, total float64) {
` + code + `
}
`})
}

// The composed server is presented as a complete package main.
func TestHTTPExampleWebServer(t *testing.T) {
	code := exampleBlock(t, "skills/go-http/references/WEB-SERVER.md", "## Structure")
	buildExampleModule(t, map[string]string{"main.go": code})
}

// The synchronous-function example must compile as a whole declaration.
func TestConcurrencyExampleSynchronousFunctions(t *testing.T) {
	code := exampleBlock(t, "skills/go-concurrency/references/GOROUTINE-PATTERNS.md", "## Prefer Synchronous Functions")
	buildExampleModule(t, map[string]string{"work.go": `package example
type Item struct{}
type Result struct{}
func processItem(Item) (Result, error) { return Result{}, nil }
` + code + `
`})
}

func TestConcurrencyExampleChannelDirection(t *testing.T) {
	code := exampleBlock(t, "skills/go-concurrency/references/SYNC-PRIMITIVES.md", "## Channel Direction Examples")
	runExampleTest(t, `package example
import "testing"
`+code+`
func TestSum(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 1; ch <- 2; ch <- 3
	close(ch)
	if got := sum(ch); got != 6 { t.Fatalf("sum = %d, want 6", got) }
}
`)
}

func TestConcurrencyExampleWaitGroupGo(t *testing.T) {
	code := exampleBlock(t, "skills/go-concurrency/SKILL.md", "## Goroutine Lifetimes")
	runExampleTest(t, `package example
import ("context"; "sync"; "sync/atomic"; "testing")
var ran atomic.Int32
func process(context.Context, int) { ran.Add(1) }
func both(ctx context.Context, first, second int) {
`+code+`
}
func TestBoth(t *testing.T) {
	both(t.Context(), 1, 2)
	if got := ran.Load(); got != 2 { t.Fatalf("process ran %d times, want 2", got) }
}
`)
}

// Struct tags are the serialization contract the skill talks about; the
// example must produce exactly the bytes its tags promise under json v1.
func TestDefensiveExampleStructTags(t *testing.T) {
	code := exampleBlock(t, "skills/go-defensive/SKILL.md", "## Struct Field Tags")
	runExampleTest(t, `package example
import ("encoding/json"; "testing")
`+code+`
func TestTags(t *testing.T) {
	b, err := json.Marshal(User{Name: "a", Email: "b"})
	if err != nil || string(b) != `+"`"+`{"name":"a","email":"b"}`+"`"+` {
		t.Fatalf("Marshal = %s, %v", b, err)
	}
}
`)
}

// The root is opened once and held on the server; a name inside it reads,
// and a parent path, an absolute path, and a symlink out of it all fail.
func TestDefensiveExampleOpenRoot(t *testing.T) {
	code := exampleBlock(t, "skills/go-defensive/SKILL.md", "## Confine Filesystem Access")
	runExampleTest(t, `package example
import ("os"; "path/filepath"; "testing")
`+code+`
func TestOpenOutsideRoot(t *testing.T) {
	dir := t.TempDir()
	uploads := filepath.Join(dir, "uploads")
	secret := filepath.Join(dir, "secret")
	if err := os.Mkdir(uploads, 0o750); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{secret: "secret", filepath.Join(uploads, "ok.txt"): "ok"} {
		if err := os.WriteFile(name, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(secret, filepath.Join(uploads, "link")); err != nil {
		t.Fatal(err)
	}
	s, err := newServer(uploads)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := s.upload("ok.txt"); err != nil || string(got) != "ok" {
		t.Fatalf("upload(ok.txt) = %q, %v; want ok", got, err)
	}
	for _, name := range []string{"../secret", secret, "link"} {
		if got, err := s.upload(name); err == nil {
			t.Errorf("upload(%q) = %q, want an error", name, got)
		}
	}
}
`)
}

// The archive loop writes nested entries through the root and refuses the
// entries that escape it: parent and absolute names, links, and an archive
// over the size cap.
func TestDefensiveExampleExtractTar(t *testing.T) {
	code := exampleBlock(t, "skills/go-defensive/SKILL.md", "### Archive Entries")
	runExampleTest(t, `package example
import ("archive/tar"; "bytes"; "errors"; "fmt"; "io"; "os"; "path"; "path/filepath"; "strings"; "testing")
`+code+`
func archive(t *testing.T, hdrs ...*tar.Header) *tar.Reader {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, h := range hdrs {
		if err := tw.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(strings.Repeat("x", int(h.Size)))); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	return tar.NewReader(&buf)
}
func TestExtract(t *testing.T) {
	parent := t.TempDir()
	dest := filepath.Join(parent, "dest")
	if err := os.Mkdir(dest, 0o750); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(dest)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	err = extract(root, archive(t,
		&tar.Header{Name: "top.txt", Typeflag: tar.TypeReg, Size: 1, Mode: 0o600},
		&tar.Header{Name: "dir/", Typeflag: tar.TypeDir, Mode: 0o750},
		&tar.Header{Name: "a/b/c.txt", Typeflag: tar.TypeReg, Size: 2, Mode: 0o600},
	), 10)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if got, err := os.ReadFile(filepath.Join(dest, "a", "b", "c.txt")); err != nil || string(got) != "xx" {
		t.Fatalf("a/b/c.txt = %q, %v; want xx", got, err)
	}
	for _, hdr := range []*tar.Header{
		{Name: "../evil.txt", Typeflag: tar.TypeReg, Size: 1},
		{Name: "a/../../evil.txt", Typeflag: tar.TypeReg, Size: 1},
		{Name: filepath.Join(parent, "evil.txt"), Typeflag: tar.TypeReg, Size: 1},
		{Name: "link", Typeflag: tar.TypeSymlink, Linkname: parent},
		{Name: "hard", Typeflag: tar.TypeLink, Linkname: "top.txt"},
		{Name: "big.txt", Typeflag: tar.TypeReg, Size: 11},
	} {
		if err := extract(root, archive(t, hdr), 10); err == nil {
			t.Errorf("extract(%q) succeeded, want an error", hdr.Name)
		}
	}
	for _, name := range []string{filepath.Join(parent, "evil.txt"), filepath.Join(dest, "link"), filepath.Join(dest, "big.txt")} {
		if _, err := os.Lstat(name); err == nil {
			t.Errorf("%s exists after a refused entry", name)
		}
	}
}
`)
}

func TestGenericsExampleSelfReferentialConstraint(t *testing.T) {
	code := exampleBlock(t, "skills/go-generics/SKILL.md", "### Self-referential constraints (Go 1.26+)")
	runExampleTest(t, `package example
import "testing"
`+code+`
type money int
func (m money) Add(o money) money { return m + o }
func TestSum(t *testing.T) {
	if got := Sum(money(0), 1, 2, 3); got != 6 { t.Fatalf("Sum = %d, want 6", got) }
}
`)
}

func TestInterfacesExampleSatisfactionCheck(t *testing.T) {
	code := exampleBlock(t, "skills/go-interfaces/SKILL.md", "## Interface Satisfaction Checks")
	buildExampleModule(t, map[string]string{"raw.go": `package example
import "encoding/json"
type RawMessage []byte
func (m RawMessage) MarshalJSON() ([]byte, error) { return []byte(m), nil }
` + code + `
`})
}

func TestTestingExampleHelper(t *testing.T) {
	code := exampleBlock(t, "skills/go-testing/SKILL.md", "## Test Helpers")
	runExampleTest(t, `package example
import "testing"
type Store struct{ dir string }
func Open(dir string) (*Store, error) { return &Store{dir: dir}, nil }
func (s *Store) Close() error { return nil }
`+code+`
func TestNewStore(t *testing.T) {
	if s := newStore(t); s == nil || s.dir == "" { t.Fatal("newStore returned no store") }
}
`)
}

// The table-test asset is copied as a starting point; it must compile and
// pass against the white-box package it declares.
func TestTestingAssetTableTemplate(t *testing.T) {
	asset := readFile(t, filepath.Join(repoRoot(t), "skills", "go-testing", "assets", "table-test-template.go"))
	if strings.Contains(asset, "TODO") || strings.Contains(asset, "//") {
		t.Errorf("table-test template carries TODO or commented-out code:\n%s", asset)
	}
	buildExampleModule(t, map[string]string{
		"example.go":      "package example\n\nfunc Example(s string) string { return s }\n",
		"example_test.go": asset,
	})
}

// The documentation asset is a complete file; its doc comments must survive
// vet and its deprecation notice must follow the skill's own rule.
func TestDocumentationAssetTemplate(t *testing.T) {
	asset := readFile(t, filepath.Join(repoRoot(t), "skills", "go-documentation", "assets", "doc-template.go"))
	buildExampleModule(t, map[string]string{"widget.go": asset})
}

// The subcommand example returns a usage error on empty input instead of
// indexing os.Args, and returns a failed Parse instead of exiting. A usage
// error (already printed by Parse) stays distinguishable from a runtime error,
// so main prints each once and exits 2 only for usage.
func TestPackagesExampleSubcommands(t *testing.T) {
	code := exampleBlock(t, "skills/go-packages/references/PACKAGE-SIZE.md", "### Subcommands")
	runExampleTest(t, `package example
import ("errors"; "flag"; "fmt"; "os"; "testing")
var served, dry = 0, false
var errBind = errors.New("bind: permission denied")
func serve(port int) error { if port == 1 { return errBind }; served = port; return nil }
func migrate(d bool) error { dry = d; return nil }
`+code+`
func TestRun(t *testing.T) {
 for _, tt := range []struct { args []string; want error }{
  {nil, errUsage},
  {[]string{"deploy"}, errUsage},
  {[]string{"serve", "-port", "x"}, errUsage},
  {[]string{"serve", "-h"}, flag.ErrHelp},
  {[]string{"serve", "-port", "1"}, errBind},
  {[]string{"serve", "-port", "9090"}, nil},
  {[]string{"migrate", "-dry_run"}, nil},
 } {
  if err := run(tt.args); !errors.Is(err, tt.want) {
   t.Errorf("run(%q) = %v, want %v", tt.args, err, tt.want)
  }
 }
 if err := run([]string{"serve", "-port", "1"}); errors.Is(err, errUsage) {
  t.Errorf("a runtime error from serve is reported as a usage error: %v", err)
 }
 if served != 9090 || !dry {
  t.Errorf("serve got %d, migrate got %t; want 9090 and true", served, dry)
 }
}
`)
}

func TestPackagesExampleRunPattern(t *testing.T) {
	code := exampleBlock(t, "skills/go-packages/SKILL.md", "## Exit in Main")
	buildExampleModule(t, map[string]string{"main.go": `package main
import "log"
func run() error { return nil }
` + code + `
`})
}

func TestDatabaseExampleRowsLoop(t *testing.T) {
	code := exampleBlock(t, "skills/go-database/SKILL.md", "## Rows: Close and Check `Err`")
	buildExampleModule(t, map[string]string{"orders.go": `package example
import ("context"; "database/sql"; "fmt")
type Order struct{ ID int; Total float64 }
func listOrders(ctx context.Context, db *sql.DB, customerID int) ([]Order, error) {
	const q = "SELECT id, total FROM orders WHERE customer_id = $1"
` + code + `
}
`})
}

func TestPerformanceExampleByteReuse(t *testing.T) {
	code := exampleBlock(t, "skills/go-performance/SKILL.md", "## Avoid Repeated String-to-Byte Conversions")
	runExampleTest(t, `package example
import ("bytes"; "testing")
func BenchmarkWrite(b *testing.B) {
	var w bytes.Buffer
`+code+`
}
`)
}
