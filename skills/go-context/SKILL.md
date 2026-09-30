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

Parameters first, then the receiver, then a global; a context value carries
only request-scoped data that crosses APIs — a request or trace ID, the
authenticated principal — never an optional parameter, per-process
configuration, or a deadline, which is derived with `WithTimeout`,
`WithDeadline`, or `WithCancel`.

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

With `context.WithCancelCause`, `WithTimeoutCause`, or `WithDeadlineCause`,
`ctx.Err()` still returns `Canceled` or `DeadlineExceeded`; only
`context.Cause(ctx)` returns the reason supplied, so a caller that must tell
this cancellation or timeout from an upstream one reads `Cause`, and
`defer cancel(nil)` releases a cause context on the success path.

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
