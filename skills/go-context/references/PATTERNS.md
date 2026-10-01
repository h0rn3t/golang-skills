# Context Patterns

> Sources: https://pkg.go.dev/context; source/golang-wiki/CodeReviewComments.md (Contexts)
> Authority: advisory; `context` API semantics follow the package documentation
> Last verified: 2026-10-01

Common patterns for deriving, checking, and propagating `context.Context`.

## Contents

- [Context Immutability](#context-immutability)
- [When to Use context.Background()](#when-to-use-contextbackground)
- [Deriving Contexts](#deriving-contexts)
- [Checking Cancellation](#checking-cancellation)
- [Respecting Cancellation in HTTP Handlers](#respecting-cancellation-in-http-handlers)
- [Context Value Best Practices](#context-value-best-practices)
- [Quick Reference](#quick-reference)

## Context Immutability

Contexts are immutable. It's safe to pass the same `ctx` to multiple calls that
share the same deadline, cancellation signal, credentials, and parent trace:

```go
// Safe: same context to sequential calls
func ProcessBatch(ctx context.Context, items []Item) error {
    for _, item := range items {
        if err := process(ctx, item); err != nil {
            return err
        }
    }
    return nil
}

// Safe: same context to concurrent calls
func ProcessConcurrently(ctx context.Context, a, b *Data) error {
    g, ctx := errgroup.WithContext(ctx)
    g.Go(func() error { return processA(ctx, a) })
    g.Go(func() error { return processB(ctx, b) })
    return g.Wait()
}
```

---

## When to Use context.Background()

Use `context.Background()` only for functions that are **never request-specific**:

```go
func main() {
    if err := run(); err != nil {
        log.Fatal(err) // exits: a defer in main would not run
    }
}

func run() error {
    // Bind the process lifetime to signals here, not with a hand-rolled
    // signal channel and goroutine.
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()
    return serve(ctx)
}
```

Two exceptions where a fresh root is correct rather than lazy:

- **In tests**: use `t.Context()` (Go 1.24+), which is cancelled at test end.
  `context.Background()` in a test leaks work past the test.
- **Cleanup that must outlive cancellation**: shutdown, flush, and audit writes
  need their own timeout, since the incoming context is already cancelled. Use
  `context.WithoutCancel(ctx)` (Go 1.21+) to keep the values while dropping the
  cancellation, or a fresh `context.WithTimeout(context.Background(), ...)`.

**Default to passing a Context** even if you think you don't need to. Only use
`context.Background()` directly if you have a good reason why passing a context
would be a mistake:

```go
func LoadConfig(ctx context.Context) (*Config, error) {
    // Even if not using ctx now, accepting it allows future
    // additions without API changes
}
```

---

## Deriving Contexts

```go
// Add timeout — cancel fires after duration elapses
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

// Add cancellation — caller controls when to cancel
ctx, cancel := context.WithCancel(ctx)
defer cancel()

// Add deadline — cancel fires at a specific wall-clock time
ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Hour))
defer cancel()

// Add value (use sparingly — only for request-scoped data)
ctx = context.WithValue(ctx, requestIDKey, reqID)
```

`defer cancel()` immediately after creating a derived context. This ensures
resources are released even if the function returns early. A context derived
inside a loop is the exception: call `cancel()` at the end of each iteration
(the [loop form](../SKILL.md#deriving-contexts)) or move the iteration into its
own function with the `defer`, since a deferred call waits for the function to
return.

### Nested Derivation

Derived contexts form a tree. Cancelling a parent cancels all its children:

```go
func handleRequest(ctx context.Context) error {
    // Parent timeout for the whole request
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // Tighter timeout for the database call
    dbCtx, dbCancel := context.WithTimeout(ctx, 5*time.Second)
    defer dbCancel()

    data, err := queryDB(dbCtx)
    if err != nil {
        return err
    }

    // Remaining time from parent context applies here
    return sendResponse(ctx, data)
}
```

---

## Checking Cancellation

### In Long-Running Loops

```go
func LongRunningOperation(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Do work
        }
    }
}
```

### Before Expensive Operations

Check cancellation before starting work that can't be interrupted:

```go
func ProcessItems(ctx context.Context, items []Item) error {
    for _, item := range items {
        if ctx.Err() != nil {
            return ctx.Err()
        }
        if err := expensiveProcess(item); err != nil {
            return err
        }
    }
    return nil
}
```

### Distinguishing Cancellation Causes

```go
if err := ctx.Err(); err != nil {
    switch {
    case errors.Is(err, context.Canceled):
        // Caller explicitly cancelled (e.g., client disconnected)
    case errors.Is(err, context.DeadlineExceeded):
        // Timeout or deadline passed
    }
}
```

---

## Respecting Cancellation in HTTP Handlers

```go
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    result, err := slowOperation(ctx)
    if err != nil {
        if errors.Is(ctx.Err(), context.Canceled) {
            // The incoming request itself was cancelled; no response is needed.
            return
        }
        slog.ErrorContext(ctx, "slow operation", "err", err)
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }
    body, err := json.Marshal(result) // encoding/json/v2; an encode error must still get a 500
    if err != nil {
        slog.ErrorContext(ctx, "encode result", "err", err)
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    _, _ = w.Write(body) // headers are sent; a failed write is the client's disconnect
}
```

An operation's `context.Canceled` may come from a child context while the
incoming request is still alive. Check `r.Context().Err()` before omitting a
response; otherwise an unwritten response becomes an implicit 200. Keep
internal error details out of the response and use the
[HTTP error mapping](../../go-http/SKILL.md#mapping-errors-to-status-codes)
for status selection and sanitized server-side diagnostics.

The `r.Context()` is cancelled when:
- The client closes the connection
- The request is cancelled by the client or HTTP/2 transport
- The `ServeHTTP` method returns

---

## Context Value Best Practices

### Use Unexported Key Types

```go
type userIDKey struct{} // one type per key: every value of one empty struct type is the same key

func WithUserID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, userIDKey{}, id)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
    id, ok := ctx.Value(userIDKey{}).(string)
    return id, ok
}
```

An unexported key type prevents collisions with keys from other packages —
even if they use the same string or int value. Inside the package, a second
key gets its own type (`type tenantIDKey struct{}`): a second
`var tenantIDKey contextKey` of a shared empty type equals the first, and the
tenant ID overwrites the user ID.

### Provide Accessor Functions

Always wrap `context.WithValue` and `ctx.Value` in typed helper functions (as
shown above) rather than exposing keys. This gives you type safety and a single
place to change the implementation.

---

## Quick Reference

| Pattern | Guidance |
|---------|----------|
| Parameter position | Always first: `func F(ctx context.Context, ...)` |
| Struct storage | Don't store in structs; pass to methods |
| Custom types | Don't create; use `context.Context` interface |
| Application data | Prefer parameters > receiver > globals > context values |
| Request-scoped data | Appropriate for context values |
| Sharing context | Safe — contexts are immutable |
| `context.Background()` | Only for non-request-specific code |
| Default | Pass context even if you think you don't need it |
| `defer cancel()` | Defer immediately after `WithTimeout`/`WithCancel`/`WithDeadline` outside a loop |
| Derived context in a loop | Call its cancel at the end of each iteration, or the iteration in its own function with `defer cancel()` |
| Value keys | One unexported struct type per key, provide accessor functions |
| Cancellation check | `ctx.Err()` before expensive ops; `select` on `ctx.Done()` in loops |
