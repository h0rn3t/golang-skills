# Logging Patterns

Detailed patterns for slog setup, handler configuration, testing, HTTP
middleware, and migration from the legacy `log` package.

## Contents

- [Setting Up slog](#setting-up-slog)
- [Handler Configuration](#handler-configuration)
- [Testing with slogtest](#testing-with-slogtest)
- [HTTP Request Logging Middleware](#http-request-logging-middleware)
- [Migration from log.Printf to slog](#migration-from-logprintf-to-slog)

## Setting Up slog

### Basic Configuration

In `main`, configure a JSON handler for production:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
slog.SetDefault(logger)
slog.Info("server started", "addr", ":8080")
// Output: {"time":"...","level":"INFO","msg":"server started","addr":":8080"}
```

### Text Handler for Development

```go
// Human-readable output for local development
logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))
slog.SetDefault(logger)
// Output: time=... level=DEBUG msg="cache lookup" key=user:42 hit=true
```

### Dynamic Level Control

Use `slog.LevelVar` to change the minimum level at runtime (e.g., via an
admin endpoint or signal handler):

```go
var programLevel = new(slog.LevelVar) // default Info

func init() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: programLevel,
    }))
    slog.SetDefault(logger)
}

// Call from an admin endpoint or signal handler
func enableDebug() {
    programLevel.Set(slog.LevelDebug)
}
```

---

## Handler Configuration

### Adding Source Location

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    AddSource: true,
    Level:     slog.LevelInfo,
}))
// Output includes: "source":{"function":"main.handleRequest","file":"server.go","line":42}
```

### Default Attributes

Use `logger.With` for fixed fields; a custom handler is only needed when fields
must be extracted dynamically from each call's context:

```go
logger = logger.With("service", "orders")
```

### Multi-Handler (Fan-Out)

Use `slog.NewMultiHandler` (Go 1.26+) to write to multiple destinations:

```go
logger := slog.New(slog.NewMultiHandler(
    slog.NewJSONHandler(os.Stdout, nil),
    slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}),
))
```

It handles fan-out and each destination's enabled level; no custom wrapper is
needed. See [go-logging](../SKILL.md#fanning-out-to-several-sinks) for the shared rule.

---

## Testing with slogtest

Go 1.22+ provides `testing/slogtest` to verify handler implementations:

```go
package myhandler_test

import (
    "testing"
    "testing/slogtest"
)

func TestHandler(t *testing.T) {
    // newHandler returns your custom slog.Handler and a func that
    // parses the output into []map[string]any for verification.
    results := func(t *testing.T) map[string]any {
        // parse your handler's output here
    }

    h := NewMyHandler(buf, nil)
    slogtest.Run(t, func(t *testing.T) slog.Handler { return h }, results)
}
```

### Capturing Logs in Tests

For unit tests that assert on log output, write to a buffer:

```go
func TestOrderProcessing(t *testing.T) {
    var buf bytes.Buffer
    logger := slog.New(slog.NewJSONHandler(&buf, nil))

    processOrder(logger, order)

    if !strings.Contains(buf.String(), `"order_id"`) {
        t.Error("expected order_id in log output")
    }
}
```

---

## HTTP Request Logging Middleware

A complete middleware that logs each request with timing, status, and
request-scoped fields:

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        reqID := r.Header.Get("X-Request-ID")
        if reqID == "" {
            reqID = uuid.NewString()
        }

        logger := slog.With(
            "request_id", reqID,
            "method", r.Method,
            "path", r.URL.Path,
        )

        // Wrap the response writer to capture the status code
        rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

        // Store logger in context for downstream handlers
        ctx := context.WithValue(r.Context(), loggerKey, logger)
        next.ServeHTTP(rw, r.WithContext(ctx))

        logger.Info("request completed",
            "status", rw.status,
            "elapsed_ms", time.Since(start).Milliseconds(),
        )
    })
}

type responseWriter struct {
    http.ResponseWriter
    status int
    wroteHeader bool
}

func (rw *responseWriter) WriteHeader(code int) {
    if rw.wroteHeader {
        return
    }
    rw.ResponseWriter.WriteHeader(code)
    if code >= 200 || code == http.StatusSwitchingProtocols {
        rw.status = code
        rw.wroteHeader = true
    }
}

func (rw *responseWriter) Write(p []byte) (int, error) {
    if !rw.wroteHeader {
        rw.WriteHeader(http.StatusOK)
    }
    return rw.ResponseWriter.Write(p)
}

func (rw *responseWriter) Unwrap() http.ResponseWriter {
    return rw.ResponseWriter
}

func (rw *responseWriter) FlushError() error {
    if !rw.wroteHeader {
        rw.WriteHeader(http.StatusOK)
    }
    return http.NewResponseController(rw.ResponseWriter).Flush()
}
```

Use `http.NewResponseController(w)` downstream for flush, hijack, and deadline
operations: `Unwrap` lets it reach the underlying writer. `FlushError` also
records the implicit 200 when flushing commits the response. Intermediate 1xx
responses leave the final status open, except 101, which switches protocols.
Libraries that assert `http.Flusher` or `http.Hijacker` directly need an adapter
that preserves those interfaces; `Unwrap` alone does not satisfy them.

### Retrieving the Logger from Context

```go
type ctxKey struct{}

var loggerKey = ctxKey{}

func loggerFromCtx(ctx context.Context) *slog.Logger {
    if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
        return l
    }
    return slog.Default()
}
```

---

## Migration from log.Printf to slog

### Step 1: Replace Direct Calls

```go
// Before
log.Printf("user %s logged in from %s", userID, ip)

// After
slog.Info("user logged in", "user_id", userID, "ip", ip)
```

### Step 2: Replace log.Fatalf in main()

```go
// Before
log.Fatalf("failed to connect: %v", err)

// After — slog has no Fatal; use slog + os.Exit in main
slog.Error("failed to connect", "err", err)
os.Exit(1)
```

### Step 3: Bridge Legacy Code

If migrating incrementally, redirect the standard `log` package output
through slog:

```go
// In main(), after setting up slog:
slog.SetDefault(logger)

// The standard log package now writes through slog's default handler.
// This works because slog.SetDefault also updates log.Default().
```

### Step 4: Replace Logger Parameters

```go
// Before: passing *log.Logger around
func NewServer(addr string, logger *log.Logger) *Server

// After: pass *slog.Logger explicitly
func NewServer(addr string, logger *slog.Logger) *Server

// Or derive from context in handlers
func (s *Server) handleRequest(ctx context.Context) {
    logger := loggerFromCtx(ctx)
    logger.Info("handling request")
}
```

### Migration Checklist

| Step | What to change | Verify |
|------|---------------|--------|
| 1 | `log.Printf` → `slog.Info/Warn/Error` | `rg 'log\.Printf'` returns 0 hits |
| 2 | `log.Fatalf` → `slog.Error` + `os.Exit(1)` in main | Only in `main()` |
| 3 | Set `slog.SetDefault` early in main | Legacy `log` calls route through slog |
| 4 | `*log.Logger` params → `*slog.Logger` | All constructors updated |
| 5 | Remove `"log"` imports where replaced | `goimports` handles this |
