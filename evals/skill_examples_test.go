package evals_test

import (
	"os"
	"os/exec"
	"path/filepath"
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

// The SSRF check is the security reference most likely to be copied
// verbatim; it must compile with a caller-supplied context and reject the
// literal addresses an attacker reaches for.
func TestSecurityExampleSSRFTarget(t *testing.T) {
	code := exampleBlock(t, "skills/go-security/references/INJECTION.md", "## Outbound URLs (SSRF)")
	runExampleTest(t, `package example
import ("context"; "fmt"; "net"; "net/url"; "testing")
`+code+`
func TestSafeTarget(t *testing.T) {
	for _, raw := range []string{"http://127.0.0.1/admin", "http://[::1]/", "http://169.254.169.254/latest", "http://10.0.0.8/", "ftp://8.8.8.8/", "http://0.0.0.0/"} {
		if _, err := safeTarget(t.Context(), raw); err == nil {
			t.Errorf("safeTarget(%q) = nil error, want a blocked-range or scheme error", raw)
		}
	}
	if _, err := safeTarget(t.Context(), "https://8.8.8.8/"); err != nil {
		t.Errorf("safeTarget(public literal) error = %v, want nil", err)
	}
	var _ = fmt.Sprint
	var _ = net.IPv4len
	var _ = url.Parse
	var _ context.Context
}
`)
}

// The TLS snippet must stay a valid tls.Config with only the field the
// reference says to set.
func TestSecurityExampleTLSConfig(t *testing.T) {
	code := exampleBlock(t, "skills/go-security/references/SECRETS-AND-CRYPTO.md", "## TLS")
	runExampleTest(t, `package example
import ("crypto/tls"; "net/http"; "testing")
func server() *http.Server {
`+code+`
	return srv
}
func TestMinVersion(t *testing.T) {
	if got := server().TLSConfig.MinVersion; got != tls.VersionTLS13 {
		t.Fatalf("MinVersion = %d, want %d", got, tls.VersionTLS13)
	}
}
`)
}

// go-resilience's only code: the retry loop stops on success, on a
// non-retryable error, when the budget is spent, and when ctx ends; a
// server-requested delay is a floor.
func TestResilienceExampleRetry(t *testing.T) {
	code := exampleBlock(t, "skills/go-resilience/SKILL.md", "## Retry Invariants")
	runExampleTest(t, `package example
import ("context"; "errors"; "math/rand/v2"; "testing"; "time")
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

// The "caller adds concurrency" example was two declarations glued to a
// top-level statement; it must be a compilable pair of functions.
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

func TestDefensiveExampleOpenRoot(t *testing.T) {
	code := exampleBlock(t, "skills/go-defensive/SKILL.md", "## Confine Filesystem Access")
	runExampleTest(t, `package example
import ("os"; "testing")
func open(userSuppliedName string) error {
`+code+`
	if err != nil {
		return err
	}
	return f.Close()
}
func TestOpenOutsideRoot(t *testing.T) {
	// /srv/uploads does not exist here; the point is that the code compiles
	// and never opens an escaping name.
	if err := open("../../etc/passwd"); err == nil {
		t.Fatal("open(../../etc/passwd) succeeded, want an error")
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
