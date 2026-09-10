---
name: go-error-handling
description: Use when returning, wrapping, or handling Go errors; choosing sentinels or custom errors; using errors.Is/As or %w; or deciding whether to log or return. Panic/recover belongs to go-defensive.
allowed-tools: Bash(bash:*)
---

# Go Error Handling

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). `errors.AsType[T]`
> requires Go 1.26+; `errors.Is`/`errors.As` and `%w` work on every supported
> release.

## Resource Routing

- `scripts/check-errors.sh` - Run when checking string-based error matching and log-and-return patterns; `--bare-return` adds the opt-in bare `return err` review.
- `scripts/check-errors-ast.go` - Implementation helper invoked by `check-errors.sh`; patch this when changing error-flow analysis behavior.
- `references/ERROR-FLOW.md` - Read when deciding where to handle, wrap, log, or return errors.
- `references/ERROR-TYPES.md` - Read when choosing sentinel errors, typed errors, or opaque errors.
- `references/WRAPPING.md` - Read when choosing `%w` versus `%v` or crossing package boundaries.

In Go, [errors are values](https://go.dev/blog/errors-are-values) — they are
created by code and consumed by code. [Error Types](#error-types) chooses
between propagating a cause and defining a new condition;
[Error Wrapping](#error-wrapping) chooses `%w` versus `%v`.

---

## Core Rules

### Never Return Concrete Error Types

**Never return concrete error types from exported functions** — a concrete `nil`
pointer stored in the `error` interface is non-nil ([go-defensive](../go-defensive/SKILL.md#common-pitfalls)
owns the typed-nil mechanism):

```go
// Bad: Concrete type can cause subtle bugs
func Bad() *os.PathError { /*...*/ }

// Good: Always return the error interface
func Good() error { /*...*/ }
```

### Error Strings

Error strings should **not** be capitalized and should **not** end with
punctuation. Exception: exported names, proper nouns, or acronyms.

```go
// Bad
err := fmt.Errorf("Something bad happened.")

// Good
err := fmt.Errorf("something bad happened")
```

For displayed messages (logs, test failures, API responses), capitalization is
appropriate.

### Return Values on Error

When a function returns an error, callers must treat all non-error return values
as unspecified unless explicitly documented.

**Tip**: Functions taking a `context.Context` should usually return an `error`
so callers can determine if the context was cancelled.

---

## Handling Errors

When encountering an error, make a **deliberate choice** — do not discard
with `_`:

1. **Handle immediately** — address the error and continue
2. **Return to caller** — optionally wrapped with context
3. **In exceptional cases** — `log.Fatal` or `panic`

To intentionally ignore: add a comment explaining why.

```go
n, _ := b.Write(p) // never returns a non-nil error
```

For related concurrent operations, use
[`errgroup`](https://pkg.go.dev/golang.org/x/sync/errgroup):

```go
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { return task1(ctx) })
g.Go(func() error { return task2(ctx) })
if err := g.Wait(); err != nil { return err }
```

For independent failures that must all be reported — validating several
fields, closing several resources — aggregate with `errors.Join`.
`errors.Is` and `errors.AsType` see through the joined error:

```go
return errors.Join(closeErr, flushErr)
```

### Avoid In-Band Errors

Don't return `-1`, `nil`, or empty string to signal errors. Use multiple
returns:

```go
// Bad: In-band error value
func Lookup(key string) int  // returns -1 for missing

// Good: Explicit error or ok value
func Lookup(key string) (string, bool)
```

This prevents callers from writing `Parse(Lookup(key))` — it causes a
compile-time error since `Lookup(key)` has 2 outputs.

---

## Error Flow

Handle errors first and return, so the success path stays unindented;
[go-style-core](../go-style-core/SKILL.md#reduce-nesting) owns nesting,
`if`-init scope, and `else`.

**Handle errors once** — either log or return, never both:

```
Error encountered?
├─ Caller can act on it? → Return (with context via %w)
├─ Top of call chain (main, a request handler)? → Log the detail and handle it
│   A handler writes a status, never the error text; that is the one place
│   the same error is both logged and answered.
└─ Neither? → Log at appropriate level, continue
```

---

## Error Types

| Caller contract | Use |
|-----------------|-----|
| Preserve an existing cause while adding context | `fmt.Errorf("...: %w", err)` |
| Match a stable condition without additional data | Reuse a suitable sentinel; define one for a new condition |
| Inspect additional structured fields or type-specific behavior | Custom `error` type |
| Message only, no stable matching contract | `errors.New` for static text; `fmt.Errorf` for dynamic text |

Dynamic wording alone does not justify a new type. `%w` already preserves
existing sentinels and typed causes for `errors.Is` and `errors.AsType`.
Add a custom type when callers need new programmatic data or behavior beyond
the existing cause and a contextual message.

### Matching a typed error

Use `errors.AsType[T]` (Go 1.26+) — it returns the value instead of writing
through a pointer, so the target variable and the `if` collapse into one line:

```go
// Good (Go 1.26+)
if pathErr, ok := errors.AsType[*fs.PathError](err); ok {
    return pathErr.Path
}

// Older toolchains, or when T must be computed at runtime
var pathErr *fs.PathError
if errors.As(err, &pathErr) { /* ... */ }
```

`go fix -errorsastype ./...` rewrites the old form. Keep `errors.Is` for
sentinel comparison — `AsType` replaces `As`, not `Is`.

### Matching multiple typed errors

Keep the original `err` visible to every branch; name each extracted cause:

```go
if pathErr, ok := errors.AsType[*fs.PathError](err); ok {
    return pathErr.Path
} else if linkErr, ok := errors.AsType[*os.LinkError](err); ok {
    return linkErr.New
}
```

With `if err, ok := errors.AsType[*fs.PathError](err); ok`, a following
`else if` sees that result's typed nil on a failed match, not the original
error. Do not reuse `err` for the extracted cause in such a chain.
[gopls `errorsastypeshadow`](https://pkg.go.dev/golang.org/x/tools/gopls/internal/analysis/errorsastypeshadow)
detects this mistake; test a cause matching the second branch, including wrapping.

---

## Error Wrapping

- **Use `%v`**: For display or annotation that deliberately omits the error chain
- **Use `%w`**: When the underlying cause is part of the caller-facing contract

**Key rules**: Place `%w` at the end. Add context callers don't have. If
annotation adds nothing, return `err` directly.

> **Validation**: Run `bash scripts/check-errors.sh` to detect common
> anti-patterns. The [go-linting](../go-linting/SKILL.md) gate covers the
> rest — `go vet` catches `errorsas` and `lostcancel`, and `go fix -diff`
> flags `errors.As` calls that should be `errors.AsType` — and runs once, at
> the end of the task, not again here.

---

## Related Skills

- **Error naming**: [go-naming](../go-naming/SKILL.md#error-names) owns `ErrX` sentinels and `XError` types
- **Testing errors**: See [go-testing](../go-testing/SKILL.md) when testing error semantics with `errors.Is`/`errors.AsType` or writing error-checking helpers
- **Panic handling**: See [go-defensive](../go-defensive/SKILL.md) when deciding between panic and error returns, or writing recover guards
- **Guard clauses**: See [go-style-core](../go-style-core/SKILL.md) — it owns nesting depth, early returns, `if`-init, and statement mechanics
- **Logging decisions**: See [go-logging](../go-logging/SKILL.md) when choosing log levels, configuring structured logging, or deciding what context to include in log messages
