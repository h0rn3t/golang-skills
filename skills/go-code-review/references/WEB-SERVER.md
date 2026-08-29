# Web Server: Skills Applied Together

This example shows how Go skills integrate in a real HTTP server. Each section
references the relevant skill for detailed guidance.

## Structure

```go
package main

import (
    "context"
    "encoding/json"
    "errors"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)

// --- Interfaces (go-interfaces) ---

// Store defines the data access boundary. Defined in the consumer
// package, not the implementation package.
type Store interface {
    GetUser(ctx context.Context, id string) (*User, error)
}

// --- Types and constructors (go-naming, go-declarations) ---

// Server handles HTTP requests for the user API.
type Server struct {
    store  Store
    router *http.ServeMux
}

// NewServer creates a Server with the given dependencies.
// The caller must call Shutdown to release resources.
func NewServer(store Store) *Server {
    s := &Server{store: store}
    s.router = http.NewServeMux()
    s.router.HandleFunc("GET /users/{id}", s.handleGetUser)
    return s
}

// --- Error handling (go-error-handling) ---

// Domain errors as sentinels — checked with errors.Is.
var ErrNotFound = errors.New("not found")

// --- HTTP handler (go-control-flow, go-context, go-error-handling) ---

func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()  // go-context: derive from request
    id := r.PathValue("id")

    user, err := s.store.GetUser(ctx, id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {  // go-error-handling: errors.Is
            http.Error(w, "user not found", http.StatusNotFound)
            return  // go-control-flow: early return
        }
        // HTTP handlers are an exception to "log OR return": log detail server-side, return sanitized error to client.
        slog.Error("GetUser failed", "id", id, "err", err)
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    // go-error-handling: never discard this — a half-written body is a real failure
    if err := json.NewEncoder(w).Encode(user); err != nil {
        slog.ErrorContext(ctx, "encode response failed", "id", id, "err", err)
    }
}

// --- Graceful shutdown (go-concurrency, go-defensive, go-packages) ---

func main() {
    if err := run(); err != nil {
        slog.Error("server failed", "err", err)
        os.Exit(1)  // go-packages: exit only from main
    }
}

func run() error {
    // go-context: signal handling owns the lifetime; no hand-rolled signal goroutine
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    store := NewDBStore(os.Getenv("DATABASE_URL"))
    srv := NewServer(store)

    httpSrv := &http.Server{
        Addr: ":8080",
        // go-defensive: origin check for state-changing requests (Go 1.25+)
        Handler:           http.NewCrossOriginProtection().Handler(srv.router),
        ReadHeaderTimeout: 5 * time.Second,  // go-defensive: use time.Duration
        ReadTimeout:       10 * time.Second,
        WriteTimeout:      10 * time.Second,
        IdleTimeout:       120 * time.Second,
        // MaxHeaderValueCount defaults to 500 (Go 1.27+); lower it for stricter APIs.
    }

    // go-concurrency: buffered so this goroutine can never block on exit
    errCh := make(chan error, 1)
    go func() {
        slog.Info("starting server", "addr", httpSrv.Addr)
        errCh <- httpSrv.ListenAndServe()
    }()

    select {
    case err := <-errCh:
        if errors.Is(err, http.ErrServerClosed) {
            return nil
        }
        return err
    case <-ctx.Done():
    }

    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()  // go-defensive: defer cleanup
    return httpSrv.Shutdown(shutdownCtx)
}
```

## Skills Applied

| Area | Skill | What's demonstrated |
|------|-------|---------------------|
| Interface at consumer | [go-interfaces](../../go-interfaces/SKILL.md) | `Store` defined where it's used |
| Naming | [go-naming](../../go-naming/SKILL.md) | MixedCaps, receiver abbreviation, clear func names |
| Error handling | [go-error-handling](../../go-error-handling/SKILL.md) | Sentinels, `errors.Is`, log-or-return |
| Context | [go-context](../../go-context/SKILL.md) | Derived from request, passed through |
| Control flow | [go-control-flow](../../go-control-flow/SKILL.md) | Early returns for error cases |
| Concurrency | [go-concurrency](../../go-concurrency/SKILL.md) | Clear goroutine lifetime, channel sizing |
| Defensive | [go-defensive](../../go-defensive/SKILL.md) | `defer cancel()`, `time.Duration`, graceful shutdown |
| Packages | [go-packages](../../go-packages/SKILL.md) | Exit only in `main()` |
| Logging | [go-logging](../../go-logging/SKILL.md) | Structured slog, handle error once |
