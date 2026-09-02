---
name: go-http
description: Use when writing or reviewing Go HTTP code — handlers, routing with net/http ServeMux, middleware, request decoding and response encoding, server timeouts and graceful shutdown, or HTTP clients. Also use when building a REST or JSON API endpoint or calling an external HTTP service, even if the user names a framework instead of net/http. Does not cover the test server helpers (see go-testing) or request-scoped logging (see go-logging).
---

# Go HTTP Servers and Clients

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). Method and
> wildcard patterns in `http.ServeMux` require Go 1.22+;
> `http.NewCrossOriginProtection` Go 1.25+; `http.Server.MaxHeaderValueCount`
> Go 1.27+.

## Resource Routing

- `references/WEB-SERVER.md` - Read when assembling a complete server: routing, handler, graceful shutdown, and where the other go-* skills meet in one `main`.

## Stdlib First

> **Normative**: `net/http` routes by method and path wildcard since Go 1.22.
> A router module is rung four of the dependency ladder in
> [go-packages](../go-packages/SKILL.md), not rung one.

Reach for a framework only when the repository already uses one — then match it
([go-style-core](../go-style-core/SKILL.md) owns house style) and keep every
rule below; they are about HTTP, not about `net/http`.

---

## Routing (Go 1.22+)

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /users/{id}", s.handleGetUser)
mux.HandleFunc("POST /users", s.handleCreateUser)
mux.HandleFunc("GET /{$}", s.handleIndex) // exact "/", not a subtree

id := r.PathValue("id")
```

- A pattern without a method matches every method; `GET` also matches `HEAD`.
- Two patterns that can match the same request panic at registration — a
  startup failure, which is the right time.
- Trailing `/` is a subtree; `{$}` pins the exact path.

---

## Handler Shape

Handlers are methods on a struct that holds the dependencies, or closures that
return `http.HandlerFunc`. Never package-level state.

Order inside a handler: bound and decode → validate → call the domain with
`r.Context()` → map the error → write once.

```go
func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
    r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    var req createUserRequest
    if err := dec.Decode(&req); err != nil {
        http.Error(w, "invalid JSON body", http.StatusBadRequest)
        return
    }
    if err := req.validate(); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    user, err := s.store.Create(r.Context(), req.toUser())
    if err != nil {
        s.writeError(w, r, err)
        return
    }
    writeJSON(w, http.StatusCreated, user)
}
```

- `http.MaxBytesReader` on every body you decode; an unbounded body is a
  memory DoS.
- `r.Context()` into every downstream call — it is cancelled when the client
  disconnects. [go-context](../go-context/SKILL.md) owns the rest.
- Set headers before `WriteHeader`; call `WriteHeader` once. Encode into a
  buffer first when the encode error must change the status; otherwise handle
  the encode error by logging it — the client already has the header.

### Mapping errors to status codes

One `writeError` per service, driven by `errors.Is`/`errors.AsType`:

| Error | Status | Body |
|---|---|---|
| Validation, malformed input | 400 | The validation message |
| `ErrNotFound` sentinel | 404 | Generic |
| `ErrConflict`, version mismatch | 409 | Generic |
| `context.Canceled` (client left) | 499-style: log at Debug, write nothing | — |
| `context.DeadlineExceeded` from downstream | 504 | Generic |
| Anything else | 500 | Generic — **never** `err.Error()` |

The 500 branch is the one place that both logs and returns: log the full error
server-side with the request ID, return a generic message.
[go-error-handling](../go-error-handling/SKILL.md) owns the handle-once rule
and this exception.

---

## Middleware

`func(next http.Handler) http.Handler`. Order from the outside in: recover →
request ID + logging → auth → the mux. Wrap `http.ResponseWriter` only to
capture the status code, and keep the wrapper transparent: use
`http.NewResponseController(w)` for deadlines and flushing rather than
asserting on optional interfaces.

---

## Server Construction

> **Normative**: Never `http.ListenAndServe(addr, h)` in production — it sets
> no timeouts. Construct an `http.Server`.

| Field | Why |
|---|---|
| `ReadHeaderTimeout` | Slowloris defense; the one field that must never be zero |
| `ReadTimeout`, `WriteTimeout` | Bound a slow client; `WriteTimeout` must exceed the slowest handler |
| `IdleTimeout` | Reclaim keep-alive connections |
| `MaxHeaderBytes`, `MaxHeaderValueCount` (Go 1.27+) | Cap header abuse |
| `Handler: http.NewCrossOriginProtection().Handler(mux)` | CSRF for state-changing requests (Go 1.25+) |
| `BaseContext` | Inject the process-lifetime context so handlers can see shutdown |

Graceful shutdown: `signal.NotifyContext` owns the lifetime, `ListenAndServe`
runs in one goroutine feeding a buffered `chan error`, and `srv.Shutdown(ctx)`
gets its own timeout. The full `run()` is in `references/WEB-SERVER.md`.

---

## Clients

- Never `http.Get`, `http.Post`, or `http.DefaultClient` — no timeout. One
  `*http.Client{Timeout: d}` per dependency, built once and reused; it owns the
  connection pool.
- `http.NewRequestWithContext(ctx, ...)` — the ctx-less form is unbounded;
  `noctx` in the lint gate flags it.
- `defer resp.Body.Close()` on every response, error or not (`bodyclose`
  flags it). Read the body to EOF, `io.LimitReader` if the size is untrusted.
- Check `resp.StatusCode` before decoding; a 5xx body is not your struct.
- Retry only idempotent requests, only on 5xx and transport errors, with
  backoff and jitter, and stop when `ctx` is done. Never retry a 4xx.

> **Validation**: `go vet ./...` (the `httpresponse` analyzer catches `Body`
> use before the error check), `golangci-lint run` with `bodyclose` and
> `noctx` from the [go-linting](../go-linting/SKILL.md) baseline, and
> `go test -race ./...` — handlers run concurrently by definition. Test
> handlers with `httptest.NewTestServer(t, h)`; [go-testing](../go-testing/SKILL.md)
> owns that.

---

## Quick Reference

| Do | Don't |
|----|-------|
| `mux.HandleFunc("GET /users/{id}", h)` | Router module for method matching |
| `http.MaxBytesReader` + `DisallowUnknownFields` | `json.Unmarshal(io.ReadAll(r.Body))` |
| `r.Context()` into every call | `context.Background()` inside a handler |
| `writeError` maps sentinels to status | `http.Error(w, err.Error(), 500)` |
| `&http.Server{ReadHeaderTimeout: ...}` | `http.ListenAndServe` |
| One `*http.Client` with `Timeout` per dependency | `http.Get` / `DefaultClient` |
| `defer resp.Body.Close()` always | Close only on the happy path |

---

## Related Skills

- **Context**: See [go-context](../go-context/SKILL.md) when deriving timeouts from `r.Context()` or deciding what may outlive the request
- **Errors**: See [go-error-handling](../go-error-handling/SKILL.md) for sentinels, wrapping, and the log-and-return exception at the handler boundary
- **Logging**: See [go-logging](../go-logging/SKILL.md) for request-scoped loggers, request IDs, and what never goes in a log line
- **Database**: See [go-database](../go-database/SKILL.md) when the handler's work is a query or a transaction
- **Testing**: See [go-testing](../go-testing/SKILL.md) for `httptest.NewTestServer`, `httptest.NewRecorder`, and `synctest` for timeout paths
- **Concurrency**: See [go-concurrency](../go-concurrency/SKILL.md) for the server goroutine, channel sizing, and shared state behind handlers
- **Dependencies**: See [go-packages](../go-packages/SKILL.md) before adding a router, JSON, or client module the standard library already covers
- **Security**: See [go-security](../go-security/SKILL.md) when a request value names a file, URL, or command, for cookie flags, SSRF checks, and what an error response may reveal
