---
name: go-logging
description: Use when choosing a logging approach, configuring slog, writing structured log statements, or deciding log levels in Go. Also use when making a Go service observable — request-scoped context, trace correlation, metric shape and label cardinality, dashboards and alerts as done-criteria — or when setting up production logging or migrating from log to slog, even if the user doesn't explicitly mention logging. Does not cover error handling strategy (see go-error-handling).
---

# Go Logging

> Compatibility: Baseline Go 1.27 (see the plugin repository's COMPATIBILITY.md).
> `slog.NewMultiHandler` requires Go 1.26+; `slog.GroupAttrs` Go 1.25+;
> `slog.DiscardHandler` Go 1.24+; `log/slog` itself Go 1.21+.

## Resource Routing

- `references/LEVELS-AND-CONTEXT.md` - Read when choosing log levels, deciding logger-in-context versus explicit parameters, or excluding sensitive fields.
- `references/LOGGING-PATTERNS.md` - Read when configuring slog handlers, logging HTTP requests, testing handlers, or migrating from `log.Printf`.
- `references/OBSERVABILITY.md` - Read when adding metrics, histograms, or a route label, correlating logs with OpenTelemetry traces, wiring propagation or sampling, or exposing profiles in a production service.

## Choosing a Logger

> **Normative**: Use `log/slog` for new Go code in a package with no logger
> yet. A package already on zap, zerolog, or logrus keeps it
> ([House Style Wins](../go-style-core/SKILL.md#house-style-wins)); replacing
> it is a migration [go-packages](../go-packages/SKILL.md) owns.

Do not introduce a third-party logging library unless profiling shows `slog`
is a bottleneck in your hot path.

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
A test or caller that ignores logs passes `slog.New(slog.DiscardHandler)`
(Go 1.24+), never a hand-written no-op handler.

---

## Log Levels

**Rules of thumb**:
- If nobody should act on it, it's not Error — use Warn or Info
- If it's only useful with a debugger attached, it's Debug
- `slog.Error` carries the error that caused it as an `"err"` attribute

On a hot path, guard attribute construction that allocates with
`logger.Enabled(ctx, slog.LevelDebug)` so a disabled level costs one check.

---

## Request-Scoped Logging

> **Advisory**: Derive loggers from context to carry request-scoped fields.

Use middleware to enrich a logger with request ID, user ID, or trace ID, then
pass the enriched logger downstream — in the context for handler and
middleware chains, as an explicit parameter for libraries and background
workers ([the decision table](references/LEVELS-AND-CONTEXT.md#when-to-use-each)).

Log through the `*Context` variants (`slog.InfoContext`, `slog.ErrorContext`)
so the context reaches `Handler.Handle`. That call alone adds nothing:
`TextHandler` and `JSONHandler` ignore the context. A trace or request ID
reaches the record only from a logger already enriched with it, or from a
handler that reads it out of the context; that handler re-wraps `WithAttrs`
and `WithGroup`, or every logger built with `With` drops the field. Both forms
are in [LOGGING-PATTERNS.md](references/LOGGING-PATTERNS.md#context-attributes).

---

## Log or Return, Not Both

The handle-once rule and its one exception — a handler at the top of the
chain logs the detail and answers with a status — belong to
[go-error-handling](../go-error-handling/SKILL.md#error-flow). The logging
side of that exception is one `*Context` call carrying the fields the record
needs:

```go
slog.ErrorContext(r.Context(), "checkout failed", "err", err, "user_id", uid)
http.Error(w, "internal error", http.StatusInternalServerError)
```

---

## Production Observability

A feature in a service is not done until an operator can see it fail: a
metric for its operations and errors, logs and traces joined by the trace ID,
and each metric wired into a dashboard or alert rule. Label values stay
bounded (method, route pattern, status); never user IDs, full URLs, or
request bodies. Metric shapes, the route label, OpenTelemetry log and trace
wiring, and profiles: [OBSERVABILITY.md](references/OBSERVABILITY.md).

---

## What NOT to Log

> **Normative**: Never log secrets, credentials, PII, or high-cardinality unbounded data.

A secret gets its own type, and every way a value is printed returns
`[REDACTED]`: `slog` calls `LogValue` only on an attribute's own value, while a
struct, slice, or map that holds the secret goes through `fmt` (`TextHandler`,
`%v`, `%#v`, an error built with `%s`) or `encoding/json` (`JSONHandler`,
`json.Marshal`):

```go
// Token is a credential; no log record, format verb, or JSON encoding shows it.
type Token string

func (Token) LogValue() slog.Value         { return slog.StringValue("[REDACTED]") }
func (Token) String() string               { return "[REDACTED]" }
func (Token) GoString() string             { return "[REDACTED]" }
func (Token) MarshalText() ([]byte, error) { return []byte("[REDACTED]"), nil }

// client holds its token in an unexported field, where fmt cannot call
// Token's methods, so client prints itself without it.
type client struct {
	baseURL string
	token   Token
}

func (c client) String() string       { return "client{" + c.baseURL + "}" }
func (c client) GoString() string     { return c.String() }
func (c client) LogValue() slog.Value { return slog.StringValue(c.String()) }
```

A struct with an exported secret field gets its redaction from the field's
type, so a `Config` holding a `Token` logs whole. An unexported field does
not: `fmt` cannot call `String` or `GoString` on it, so `TextHandler`, `%v`,
`%+v`, and `%#v` print the raw secret. A struct that keeps a secret in an
unexported field prints itself without it, as `client` does. The one site
that must send the secret converts with `string(tok)`.

A request value goes into an `slog` attribute, which both handlers quote or
escape; `log.Printf` or `fmt` with `%s` writes its CR/LF verbatim and lets the
client forge log lines. An error tracker or analytics SDK (Sentry, Segment) is
a log sink too: it gets the user ID, never the email, phone, or request
headers, scrubbed in the SDK's before-send hook.

---

## Related Skills

- [go-error-handling](../go-error-handling/SKILL.md): log or return, the handle-once pattern.
- [go-context](../go-context/SKILL.md): which request-scoped values a context may carry.
- [go-performance](../go-performance/SKILL.md): hot-path logging and allocations in log calls.
- [go-code-review](../go-code-review/SKILL.md): reviewing logging in a PR.
