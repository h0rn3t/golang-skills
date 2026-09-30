# Levels and Context

> Sources: https://pkg.go.dev/log/slog; https://google.github.io/styleguide/go/best-practices#logging
> Authority: advisory
> Minimum Go: `log/slog` 1.21; `slog.NewMultiHandler` 1.26
> Last verified: 2026-09-10

## Contents

- [Level Semantics](#level-semantics)
- [Custom Verbosity Levels](#custom-verbosity-levels)
- [Context-Based Logging](#context-based-logging)
- [What NOT to Log](#what-not-to-log)

## Level Semantics

### Choosing Between Warn and Error

```
Is the error returned to the caller?
├─ Yes → do not log it; the top of the chain logs it once
└─ No, it is handled here
    ├─ The operation succeeded after retry/fallback → Warn
    ├─ It failed and needs an operator now          → Error
    └─ It failed and can wait for the next review   → Warn
```

---

## Custom Verbosity Levels

```go
const (
    LevelTrace = slog.Level(-8)  // below Debug
    LevelNotice = slog.Level(2)  // between Info and Warn
)

slog.Log(ctx, LevelTrace, "detailed trace", "span_id", spanID)
```

---

## Context-Based Logging

### Pattern 1: Logger in Context

Use this when HTTP middleware needs to add request-scoped fields as the request
moves through a handler chain. The canonical context-key and middleware
implementation lives in `LOGGING-PATTERNS.md`.

### Pattern 2: Explicit Logger Parameter

Pass `*slog.Logger` as a function parameter alongside context.

### When to Use Each

| Situation | Recommendation |
|-----------|---------------|
| HTTP handlers / middleware chains | Logger in context |
| Library code with no HTTP dependency | Explicit parameter |
| Background workers / batch jobs | Explicit parameter |
| Deep call chains (5+ levels) | Logger in context |

---

## What NOT to Log

Secrets and credentials belong to [go-security](../../go-security/SKILL.md#secrets).

### Personally Identifiable Information (PII)

Avoid logging unless required for debugging and your retention policy
allows it:
- Email addresses, phone numbers
- Full names, physical addresses
- IP addresses (in some jurisdictions)
- Credit card numbers, SSNs

If you must log a user identifier, use an opaque ID rather than PII.

### High-Cardinality Unbounded Data

Don't log entire request bodies, full stack traces at Info level, or
unbounded collections:

```go
// Bad: unbounded data
slog.Info("received", "body", string(requestBody))
slog.Info("users loaded", "users", users) // could be 100k entries

// Good: bounded summary
slog.Info("received", "content_length", len(requestBody), "content_type", ct)
slog.Info("users loaded", "count", len(users))
```

### Decision Table

| Data type | Log it? | Alternative |
|-----------|---------|-------------|
| Request ID / trace ID | Yes | — |
| User ID (opaque) | Yes | — |
| HTTP method, path, status | Yes | — |
| Error messages | Yes | — |
| Passwords / tokens | **Never** | A secret type whose `LogValue` returns `[REDACTED]` ([go-logging](../SKILL.md#what-not-to-log)) |
| Full request body | **No** | Log content length and type |
| PII (email, name) | **Avoid** | Log opaque user ID |
| Large collections | **No** | Log count or summary |
| Stack traces | Debug only | Use `slog.Debug` |
