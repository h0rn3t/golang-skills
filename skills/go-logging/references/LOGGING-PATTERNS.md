# Logging Patterns

> Sources: https://pkg.go.dev/log/slog; https://go.dev/blog/slog
> Authority: advisory
> Minimum Go: `log/slog` 1.21; `slog.DiscardHandler` 1.24; `slog.NewMultiHandler` 1.26
> Last verified: 2026-09-10

## Contents

- [Setting Up slog](#setting-up-slog)
- [Handler Configuration](#handler-configuration)
- [Testing with slogtest](#testing-with-slogtest)
- [HTTP Request Logging Middleware](#http-request-logging-middleware)
- [Migration from log.Printf to slog](#migration-from-logprintf-to-slog)

## Setting Up slog

### Dynamic Level Control

Use `slog.LevelVar` to change the minimum level at runtime (e.g., via an
admin endpoint or signal handler):

```go
level := new(slog.LevelVar) // Info until changed
slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})))
level.Set(slog.LevelDebug)
```

---

## Handler Configuration

### Default Attributes

Use `logger.With` for fixed fields; a custom handler is only needed when fields
must be extracted dynamically from each call's context.

### Multi-Handler (Fan-Out)

Use `slog.NewMultiHandler` (Go 1.26+) to write to multiple destinations:

```go
logger := slog.New(slog.NewMultiHandler(
    slog.NewJSONHandler(os.Stdout, nil),
    slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}),
))
```

Each destination keeps its own level; [go-logging](../SKILL.md#fanning-out-to-several-sinks) has the rule.

---

## Testing with slogtest

`testing/slogtest` verifies handler implementations — the package is Go 1.21+,
the `Run` form below Go 1.22+:

```go
package myhandler_test

import (
    "bytes"
    "encoding/json"
    "log/slog"
    "testing"
    "testing/slogtest"
)

func TestHandler(t *testing.T) {
    var buf bytes.Buffer
    newHandler := func(*testing.T) slog.Handler {
        buf.Reset()
        return slog.NewJSONHandler(&buf, nil) // replace with your handler
    }
    // result parses the one record each slogtest case wrote.
    result := func(t *testing.T) map[string]any {
        var m map[string]any
        if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
            t.Fatal(err)
        }
        return m
    }
    slogtest.Run(t, newHandler, result)
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
            reqID = uuid.New().String() // standard-library uuid, Go 1.27+
        }

        logger := slog.With(
            "request_id", reqID,
            "method", r.Method,
            "path", r.URL.Path,
        )

        rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
        ctx := context.WithValue(r.Context(), loggerKey{}, logger)
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
type loggerKey struct{}

func loggerFromCtx(ctx context.Context) *slog.Logger {
    if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
        return l
    }
    return slog.Default()
}
```

---

## Migration from log.Printf to slog

### Replace log.Fatalf in main()

```go
// Before
log.Fatalf("failed to connect: %v", err)

// After — slog has no Fatal; use slog + os.Exit in main
slog.Error("connect", "err", err)
os.Exit(1)
```

### Bridge Legacy Code

If migrating incrementally, redirect the standard `log` package output
through slog:

```go
// In main(), after setting up slog:
slog.SetDefault(logger)

// The standard log package now writes through slog's default handler.
// This works because slog.SetDefault also updates log.Default().
```
