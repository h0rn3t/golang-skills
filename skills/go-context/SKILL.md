---
name: go-context
description: Use when handling Go context.Context, cancellation, deadlines, timeouts, or request-scoped values and parameter placement. Goroutine lifecycle and synchronization belong to go-concurrency.
---

# Go Context Usage

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`).
> `context.WithoutCancel`, `context.AfterFunc`, `context.WithTimeoutCause` and
> `context.WithDeadlineCause` require Go 1.21+;
> `t.Context()` in tests, Go 1.24+.

## Resource Routing

- `references/PATTERNS.md` - Read when deriving contexts, checking cancellation, handling HTTP request contexts, or using typed context-value keys.

## Context as First Parameter

Functions that use a Context accept it as their **first parameter**. Do not
add a Context member to a struct type; pass `ctx` to each method that needs
it. **Exception**: methods whose signature must match an interface in the
standard library or a third-party library. Do not create custom Context types
or use interfaces other than `context.Context` in function signatures.

---

## Where to Put Application Data

Consider these options in order of preference:

1. **Function parameters** — most explicit and type-safe
2. **Receiver** — for data that belongs to the type
3. **Globals** — for truly global configuration (use sparingly)
4. **Context value** — only for request-scoped data

Context values are appropriate for:
- Request IDs and trace IDs
- Authentication/authorization info that flows with requests

Context values are **not** appropriate for:
- Deadlines and cancellation signals — derive them with `WithTimeout`,
  `WithDeadline`, or `WithCancel`, never `WithValue`
- Optional function parameters
- Data that could be passed explicitly
- Configuration that doesn't vary per-request

---

## Common Patterns

### Deriving Contexts

`defer cancel()` immediately after creating a derived context. Inside a loop a
`defer` waits for the function to return, so every iteration's context stays
live until then: cancel at the end of each iteration, or move the iteration
into its own function that defers `cancel()`.

```go
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

for _, item := range items {
    itemCtx, cancelItem := context.WithTimeout(ctx, time.Second)
    err := send(itemCtx, item)
    cancelItem() // per iteration, not deferred
    if err != nil {
        return err
    }
}
```

### Cancellation With a Reason

`context.WithCancelCause` records *why* a context was cancelled: `ctx.Err()`
still returns `Canceled`, and `context.Cause(ctx)` returns the reason. Use it
when several paths can cancel and the caller must tell them apart:

```go
ctx, cancel := context.WithCancelCause(ctx)
defer cancel(nil)
// ...
cancel(fmt.Errorf("upstream closed: %w", err))
// later, in the caller
if err := context.Cause(ctx); err != nil {
    return err
}
```

`context.WithTimeoutCause` and `context.WithDeadlineCause` (Go 1.21+) do the
same for a deadline: `ctx.Err()` stays `DeadlineExceeded`, `context.Cause(ctx)`
is the error you supplied, so a caller can tell this timeout from an upstream
one.

`context.AfterFunc(ctx, f)` runs `f` once `ctx` is done; call `stop` when the
work ends first:

```go
stop := context.AfterFunc(ctx, func() { conn.Close() })
defer stop()
```

`context.WithoutCancel(ctx)` keeps the values but drops cancellation; use it
for work that must outlive the request (audit log, cleanup) and give that work
its own timeout.

---

## Related Skills

- [go-resilience](../go-resilience/SKILL.md): total and attempt time budgets, cancellation-aware backoff, durable retry ownership.
- [go-concurrency](../go-concurrency/SKILL.md): goroutine cancellation, select timeouts, errgroup.
- [go-error-handling](../go-error-handling/SKILL.md): wrapping or returning `ctx.Err()`.
- [go-interfaces](../go-interfaces/SKILL.md): APIs that take context alongside interfaces.
- [go-logging](../go-logging/SKILL.md): loggers and request IDs carried in context.
