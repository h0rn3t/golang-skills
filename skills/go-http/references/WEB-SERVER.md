# Web Server: Skills Applied Together

> Sources: https://pkg.go.dev/net/http; the owner skills the table names
> Authority: project policy (a composition example, not a rule source)
> Minimum Go: `ServeMux` patterns 1.22; `encoding/json/v2` 1.27
> Last verified: 2026-09-24

A whole `package main` that serves one read-only route, in the form the owner
skills ask for. The store is a concrete type: an interface arrives with a
second implementation or a test fake
([go-interfaces](../../go-interfaces/SKILL.md)). The handler is registered
once, so it is written at its registration and captures the store. A server
with state-changing routes also wraps the mux in
`http.NewCrossOriginProtection().Handler(mux)`
([Server Construction](../SKILL.md#server-construction)); on a GET-only
server it checks nothing.

## Structure

```go
package main

import (
    "context"
    json "encoding/json/v2"
    "errors"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)

var errNotFound = errors.New("not found")

type user struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

// store reads users from the database; the query belongs to go-database.
type store struct{ dsn string }

func (s *store) user(ctx context.Context, id string) (user, error) {
    return user{}, errNotFound
}

func main() {
    if err := run(); err != nil {
        slog.Error("serve", "err", err)
        os.Exit(1)
    }
}

func run() error {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    users := &store{dsn: os.Getenv("DATABASE_URL")}
    mux := http.NewServeMux()
    mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        u, err := users.user(r.Context(), id)
        if errors.Is(err, errNotFound) {
            http.Error(w, "user not found", http.StatusNotFound)
            return
        }
        if err != nil {
            // The detail stays in the server log; the client gets only a status.
            slog.ErrorContext(r.Context(), "get user", "id", id, "err", err)
            http.Error(w, "internal error", http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        _ = json.MarshalWrite(w, u) // headers are sent; a failed write is the client's disconnect
    })

    srv := &http.Server{
        Addr:              ":8080",
        Handler:           mux,
        ReadHeaderTimeout: 5 * time.Second,
        ReadTimeout:       10 * time.Second,
        WriteTimeout:      10 * time.Second,
        IdleTimeout:       120 * time.Second,
    }
    errCh := make(chan error, 1) // buffered: the goroutine exits even after run returns
    go func() { errCh <- srv.ListenAndServe() }()
    slog.Info("serving", "addr", srv.Addr)

    select {
    case err := <-errCh:
        return err
    case <-ctx.Done():
    }
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    return srv.Shutdown(shutdownCtx)
}
```

`ListenAndServe` returns `http.ErrServerClosed` only after `Shutdown`, so an
error that arrives first is a real one and is returned as it is.

## Skills Applied

| Area | Skill | What's demonstrated |
|------|-------|---------------------|
| Concrete dependency | [go-interfaces](../../go-interfaces/SKILL.md) | `store` is a type, not an interface with one implementation |
| Naming | [go-naming](../../go-naming/SKILL.md) | MixedCaps, one-letter receiver, unexported names in `package main` |
| Error handling | [go-error-handling](../../go-error-handling/SKILL.md) | Sentinel matched with `errors.Is`; the handler logs once and answers with a status |
| Context | [go-context](../../go-context/SKILL.md) | Signal context owns the lifetime; the request context reaches the store |
| Control flow | [go-style-core](../../go-style-core/SKILL.md) | Early returns for error cases |
| Concurrency | [go-concurrency](../../go-concurrency/SKILL.md) | One goroutine with a buffered result channel |
| Defensive | [go-defensive](../../go-defensive/SKILL.md) | `defer cancel()`, `time.Duration` timeouts, graceful shutdown |
| Packages | [go-packages](../../go-packages/SKILL.md) | `main` calls `run` and is the only place that exits |
| Logging | [go-logging](../../go-logging/SKILL.md) | Structured slog, a static message, the error logged once |
