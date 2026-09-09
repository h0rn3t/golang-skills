---
name: go-http
description: Use when writing or reviewing Go HTTP code — handlers, routing with net/http ServeMux, middleware, request decoding and response encoding, server timeouts and graceful shutdown, or HTTP clients. Also use when building a REST or JSON API endpoint or calling an external HTTP service, even if the user names a framework instead of net/http. Does not cover the test server helpers (see go-testing) or request-scoped logging (see go-logging).
---

# Go HTTP Servers and Clients

> Compatibility: Baseline Go 1.27 (`COMPATIBILITY.md`). `ServeMux` method and
> wildcard patterns require Go 1.22+; `http.NewCrossOriginProtection` Go 1.25+;
> `http.Server.MaxHeaderValueCount` Go 1.27+.

## Resource Routing

- `references/WEB-SERVER.md` - Read when assembling a complete server: routing, handler, graceful shutdown, and where the other go-* skills meet in one `main`.

## Stdlib First

Use `net/http` method/path routing before adding a router module
([go-packages](../go-packages/SKILL.md) owns the dependency ladder). Match an
existing framework and house style; the HTTP rules still apply.
Use a framework only when the repository already uses one.

## Routing (Go 1.22+)

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /users/{id}", s.handleGetUser)
mux.HandleFunc("POST /users", s.handleCreateUser)
mux.HandleFunc("GET /{$}", s.handleIndex) // exact "/", not a subtree

id := r.PathValue("id")
```

- A pattern without a method matches every method; `GET` also matches `HEAD`.
  If the endpoint contract requires 405 for HEAD, reject it explicitly on
  that endpoint or register its matching HEAD pattern with a 405 handler.
- Conflicting patterns panic at registration; overlapping patterns are valid
  when one is more specific.
- Trailing `/` is a subtree; `{$}` pins the exact path.

## Handler Shape

Use plain functions, closures, or methods. A struct holds dependencies or state
shared by handlers; a small handler can capture them directly. Preserve the
existing structure. No package-level state.

Bound and decode → validate → call the domain with `r.Context()` → map the
error → write once.

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
    if err := dec.Decode(new(any)); err != io.EOF {
        http.Error(w, "body must contain one JSON value", http.StatusBadRequest)
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

- Bound every decoded body with `http.MaxBytesReader`. For a single-document
  endpoint, require EOF after the first value before calling the domain;
  `DisallowUnknownFields` alone accepts trailing data. The second decode
  allows trailing whitespace while still enforcing the cap.
- Pass `r.Context()` downstream; it is cancelled on client disconnect.
- For a JSON array contract, build the response with `make([]T, 0, n)` on
  both unfiltered and filtered paths, including nil input and no matches.
  Preserve the wire type independently of how the server snapshots its input.
- Set headers before `WriteHeader`, and call it once. Buffer encoding when an
  encode error must change the status; otherwise log the encode error because
  headers have already been sent.

### Mapping errors to status codes

Centralize shared error-to-status rules in `writeError`, using
`errors.Is`/`errors.AsType`; keep a single-handler mapping local. Apply each
row where that failure can occur:

| Error | Status | Body |
|---|---|---|
| Validation, malformed input | 400 | The validation message |
| `ErrNotFound` sentinel | 404 | Generic |
| `ErrConflict`, version mismatch | 409 | Generic |
| `context.Canceled` (client left) | 499-style: log at Debug, write nothing | — |
| `context.DeadlineExceeded` from downstream | 504 | Generic |
| Anything else | 500 | Generic — **never** `err.Error()` |

At the 500 boundary, log the full error server-side with the request ID and
return a generic message. This is the handle-once exception owned by
[go-error-handling](../go-error-handling/SKILL.md).

## Middleware

`func(next http.Handler) http.Handler`. Outside in: recover → request ID +
logging → auth → mux. Wrap `http.ResponseWriter` only to capture status and
keep it transparent; use `http.NewResponseController(w)` for deadlines and
flushing instead of asserting optional interfaces.

## Server Construction

Construct an `http.Server`; bare `http.ListenAndServe` sets no timeouts.

| Field | Why |
|---|---|
| `ReadHeaderTimeout` | Slowloris defense; must never be zero |
| `ReadTimeout`, `WriteTimeout` | Bound slow clients; `WriteTimeout` exceeds the slowest handler |
| `IdleTimeout` | Reclaim keep-alive connections |
| `MaxHeaderBytes`, `MaxHeaderValueCount` (Go 1.27+) | Cap header abuse |
| `Handler: http.NewCrossOriginProtection().Handler(mux)` | CSRF for state-changing requests (Go 1.25+) |
| `BaseContext` | Expose process shutdown to handlers |

For graceful shutdown, `signal.NotifyContext` owns the lifetime;
`ListenAndServe` runs in one goroutine feeding a buffered error channel;
`srv.Shutdown(ctx)` gets its own timeout.

## Clients

- Never `http.Get`, `http.Post`, or `http.DefaultClient` — no timeout. One
  `*http.Client{Timeout: d}` per dependency, built once and reused; it owns the
  connection pool.
- `http.NewRequestWithContext(ctx, ...)` — the ctx-less form is unbounded;
  `noctx` in the lint gate flags it.
- `defer resp.Body.Close()` on every response, error or not (`bodyclose`
  flags it). Bound untrusted bodies and close them on every path. Go 1.27
  drains an unread HTTP/1 body on `Close` — up to 256 KiB and 50 ms — to keep
  the connection reusable, so closing is the whole obligation; a manual read to
  EOF only buys reuse for responses past those bounds.
- Check `resp.StatusCode` before decoding; a 5xx body is not your struct.
- For retry eligibility, `Retry-After`, replay safety, and coordinated budgets,
  use [go-resilience](../go-resilience/SKILL.md). HTTP status alone does not
  establish whether repeating an operation is safe.


### Bounded Response Bodies

For a strict response-size limit, read at most the limit plus one byte,
reject an oversized result, then decode the complete buffer into `dst`.
A decoder can finish its first object before `io.LimitReader` reaches its
limit; that alone does not enforce the size of the whole response.

```go
const maxBody = 64 << 10
body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
if err != nil {
    return fmt.Errorf("read response: %w", err)
}
if len(body) > maxBody {
    return fmt.Errorf("response exceeds %d bytes", maxBody)
}
return json.Unmarshal(body, dst)
```

## Validation

Run `go vet ./...` (`httpresponse`), `golangci-lint run` with `bodyclose` and
`noctx` from [go-linting](../go-linting/SKILL.md), and `go test -race ./...`:
handlers run concurrently. Test handlers with `httptest.NewTestServer(t, h)`.

## Related Skills

- [go-context](../go-context/SKILL.md): derived deadlines and request lifetime.
- [go-error-handling](../go-error-handling/SKILL.md): sentinels and wrapping.
- [go-logging](../go-logging/SKILL.md): request IDs, log fields, redaction.
- [go-database](../go-database/SKILL.md): handler queries and transactions.
- [go-testing](../go-testing/SKILL.md): `httptest` and `synctest`.
- [go-concurrency](../go-concurrency/SKILL.md): server goroutine and shared state.
- [go-packages](../go-packages/SKILL.md): router, JSON, and client dependencies.
- [go-security](../go-security/SKILL.md): input naming files, URLs or commands;
  cookies, SSRF, and error disclosure.
