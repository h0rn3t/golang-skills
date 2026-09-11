---
name: go-logging
description: Use when choosing a logging approach, configuring slog, writing structured log statements, or deciding log levels in Go. Also use when making a Go service observable — request-scoped context, trace correlation, metric shape and label cardinality, dashboards and alerts as done-criteria — or when setting up production logging or migrating from log to slog, even if the user doesn't explicitly mention logging. Does not cover error handling strategy (see go-error-handling).
---

# Go Logging

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`).
> `slog.NewMultiHandler` requires Go 1.26+; `slog.GroupAttrs` Go 1.25+;
> `log/slog` itself Go 1.21+.

## Resource Routing

- `references/LEVELS-AND-CONTEXT.md` - Read when choosing log levels, deciding logger-in-context versus explicit parameters, or excluding sensitive fields.
- `references/LOGGING-PATTERNS.md` - Read when configuring slog handlers, logging HTTP requests, testing handlers, or migrating from `log.Printf`.

## Core Principle

Logs are for **operators**, not developers. Every log line should help someone
diagnose a production issue. If it doesn't serve that purpose, it's noise.

---

## Choosing a Logger

> **Normative**: Use `log/slog` for new Go code.

`slog` is structured, leveled, and in the standard library (Go 1.21+). It
covers the vast majority of production logging needs.

```
Which logger?
├─ New production code      → log/slog
├─ Trivial CLI / one-off    → log (standard)
└─ Measured perf bottleneck → zerolog or zap (benchmark first)
```

Do not introduce a third-party logging library unless profiling shows `slog`
is a bottleneck in your hot path. When you do, keep the same structured
key-value style.

---

## Structured Logging

> **Normative**: Always use key-value pairs. Never interpolate values into the message string.

The message is a **static description** of what happened. Dynamic data goes in
key-value attributes:

```go
// Good: static message, structured fields
slog.Info("order placed", "order_id", orderID, "total", total)

// Bad: dynamic data baked into the message string
slog.Info(fmt.Sprintf("order %d placed for $%.2f", orderID, total))
```

### Key Naming

> **Advisory**: Use `snake_case` for log attribute keys.

Keys should be lowercase, underscore-separated, and consistent across the
codebase: `user_id`, `request_id`, `elapsed_ms`.

### Typed Attributes

For performance-critical paths, use typed constructors to avoid allocations.
Group related attributes with `slog.GroupAttrs` (Go 1.25+) — it takes `...Attr`
rather than `slog.Group`'s `...any`, so the compiler checks the arguments:

```go
slog.LogAttrs(ctx, slog.LevelInfo, "request handled",
    slog.String("method", r.Method),
    slog.Int("status", code),
    slog.Duration("elapsed", elapsed),
    slog.GroupAttrs("client", slog.String("ip", ip), slog.String("ua", ua)),
)
```

### Fanning out to several sinks

`slog.NewMultiHandler` (Go 1.26+) replaces hand-written fan-out handlers — e.g.
JSON to stdout for the log pipeline plus text to stderr for a human:

```go
logger := slog.New(slog.NewMultiHandler(
    slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
    slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}),
))
```

Each handler keeps its own level and format; `Enabled` is true if any child is.

---

## Log Levels

> **Advisory**: Follow these level semantics consistently.

| Level | When to use | Production default |
|-------|-------------|--------------------|
| Debug | Developer-only diagnostics, tracing internal state | Disabled |
| Info  | Notable lifecycle events: startup, shutdown, config loaded | Enabled |
| Warn  | Unexpected but recoverable: deprecated feature used, retry succeeded | Enabled |
| Error | Operation failed, requires operator attention | Enabled |

**Rules of thumb**:
- If nobody should act on it, it's not Error — use Warn or Info
- If it's only useful with a debugger attached, it's Debug
- `slog.Error` should always include an `"err"` attribute

```go
slog.Error("payment failed", "err", err, "order_id", id)
slog.Warn("retry succeeded", "attempt", n, "endpoint", url)
slog.Info("server started", "addr", addr)
slog.Debug("cache lookup", "key", key, "hit", hit)
```

On a hot path, guard attribute construction that allocates with
`logger.Enabled(ctx, slog.LevelDebug)` so a disabled level costs one check.

---

## Request-Scoped Logging

> **Advisory**: Derive loggers from context to carry request-scoped fields.

Use middleware to enrich a logger with request ID, user ID, or trace ID, then
pass the enriched logger downstream — in the context for handler and
middleware chains, as an explicit parameter for libraries and background
workers ([the decision table](references/LEVELS-AND-CONTEXT.md#when-to-use-each)).
Keep the full context-key and middleware implementation in the logging patterns
reference so request-scoped logging has one owner.

Log through the `*Context` variants (`slog.InfoContext`, `slog.ErrorContext`)
so the context reaches `Handler.Handle`. That call alone adds nothing:
`TextHandler` and `JSONHandler` ignore the context. A trace or request ID
reaches the record only from a logger already enriched with it, or from a
handler that reads it out of the context. For the existing enriched-logger
approach, see "HTTP Request Logging Middleware" and "Retrieving the Logger
from Context" in [LOGGING-PATTERNS.md](references/LOGGING-PATTERNS.md).
Use that retrieved logger's `InfoContext` method downstream.

---

## Log or Return, Not Both

The handle-once rule and its one exception — a handler at the top of the
chain logs the detail and answers with a status — belong to
[go-error-handling](../go-error-handling/SKILL.md#error-flow). The logging
side of that exception is the `*Context` call, so the request's fields reach
the record:

```go
slog.ErrorContext(r.Context(), "checkout failed", "err", err, "user_id", uid)
http.Error(w, "internal error", http.StatusInternalServerError)
```

---

## Production Observability Checklist

> **Advisory**: for a production service, in the stack the project already
> runs. The metric shapes below are Prometheus terms because that is the
> common case — translate them for OpenTelemetry or a vendor agent rather
> than adding a second stack, and skip the checklist for a library or CLI.

A feature in a service is not done until an operator can see it fail:

- **Metrics** — counters for operations and errors, histograms for latency
  (histograms aggregate across instances; summaries do not). Keep the query
  that reads a metric next to its declaration.
- **Cardinality** — label values stay bounded (method, route pattern, status);
  never user IDs, full URLs, or request bodies.
- **Logs** — structured key-value records carrying the request or trace ID,
  which needs the enriched logger or context handler described above.
- **Dashboards and alerts** — a metric nobody queries is not observability:
  land each one in the project's dashboards and alert rules, or say plainly
  that it shipped unwired.
- **Profiles** — mount `net/http/pprof` on a separate internal listener,
  never on the public mux ([go-security](../go-security/SKILL.md#http-surface)
  owns the rule); [go-troubleshooting](../go-troubleshooting/SKILL.md) owns
  reading them.

---

## What NOT to Log

> **Normative**: Never log secrets, credentials, PII, or high-cardinality unbounded data.

- Passwords, API keys, tokens, session IDs
- Full credit card numbers, SSNs
- Request/response bodies that may contain user data
- Entire slices or maps of unbounded size

---

## Related Skills

- [go-error-handling](../go-error-handling/SKILL.md): log or return, the handle-once pattern.
- [go-context](../go-context/SKILL.md): request-scoped values and loggers in context.
- [go-performance](../go-performance/SKILL.md): hot-path logging and allocations in log calls.
- [go-code-review](../go-code-review/SKILL.md): reviewing logging in a PR.
