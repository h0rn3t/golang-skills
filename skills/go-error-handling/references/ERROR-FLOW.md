# Error Flow Patterns

> Sources: source/uber-go-style/style.md (Handle Errors Once); source/golang-wiki/CodeReviewComments.md (Indent Error Flow)
> Authority: advisory
> Last verified: 2026-09-10

## Handle Errors Once

When a caller receives an error, it should handle each error **only once**.
Choose ONE response:

1. **Return the error** (wrapped or verbatim) for the caller to handle
2. **Log and degrade gracefully** (don't return the error)
3. **Match and handle** specific error cases, return others

**If you return an error, don't log it yourself** — let the caller handle it.
Logging and returning the same error is the most common "handle errors once"
violation, causing duplicate noise as callers up the stack also handle the error.

```go
// Bad: Logs AND returns - causes noise in logs
u, err := getUser(id)
if err != nil {
    log.Printf("Could not get user %q: %v", id, err)
    return err  // Callers will also log this!
}

// Good: Wrap and return - let caller decide how to handle
u, err := getUser(id)
if err != nil {
    return fmt.Errorf("get user %q: %w", id, err)
}

// Good: Log and degrade gracefully (don't return error)
if err := emitMetrics(); err != nil {
    // Failure to write metrics should not break the application
    log.Printf("emit metrics: %v", err)
}
// Continue execution...

// Good: Match specific errors, return others
tz, err := getUserTimeZone(id)
switch {
case errors.Is(err, ErrUserNotFound):
    tz = time.UTC
case err != nil:
    return fmt.Errorf("get user %q: %w", id, err)
}
```

---

## Logging vs Returning Errors

### Levels and Structure

Which level, structured attributes, and what must never reach a log line are
[go-logging](../../go-logging/SKILL.md)'s rules; this reference only decides
whether the error is logged or returned.
