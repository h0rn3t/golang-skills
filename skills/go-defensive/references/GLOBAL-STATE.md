# Global State Patterns

> Sources: source/google-go-styleguide/best-practices.md (Global state); source/effective-go/effective_go.html (Initialization)
> Authority: advisory
> Last verified: 2026-09-29

## When Global State Is Acceptable

Not all package-level variables are harmful. Global state is appropriate when
it is **truly process-wide** and **not worth injecting**:

- **Default instances** — `slog.Default()`, `log.Default()`, `flag.CommandLine`
- **Compiled-once values** — `regexp.MustCompile(...)` at package level
- **Registries** — `database/sql.Register`, `image.RegisterFormat`
- **Singleton infrastructure** — a process-wide metric collector or trace exporter

## Injecting Time

A common case: replacing `time.Now` for deterministic tests.

**Bad**
```go
func IsExpired(expiry time.Time) bool {
    return time.Now().After(expiry) // untestable
}
```

**Good**
```go
func IsExpired(now, expiry time.Time) bool {
    return now.After(expiry)
}
```

The caller passes `time.Now()`; a test passes a fixed instant, and no type
exists to hold a clock:

```go
expiry := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
if !IsExpired(expiry.Add(time.Second), expiry) {
    t.Error("IsExpired(expiry+1s, expiry) = false, want true")
}
```

Code that sleeps or waits on a timer keeps calling `time.Now`, and its test
runs inside `synctest.Test`, whose fake clock starts at midnight UTC
2000-01-01 and advances only when every goroutine in the bubble is blocked
([go-testing](../../go-testing/SKILL.md#use-the-toolchains-test-apis)).
