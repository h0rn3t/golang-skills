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
- `references/ERROR-TYPES.md` - Read when choosing sentinel errors, typed errors, or opaque errors, or for the API rules on error values (interface results, message form, in-band values).
- `references/WRAPPING.md` - Read when choosing `%w` versus `%v` or crossing package boundaries.

In Go, [errors are values](https://go.dev/blog/errors-are-values) — they are
created by code and consumed by code. [Error Types](#error-types) chooses
between propagating a cause and defining a new condition;
[Error Wrapping](#error-wrapping) chooses `%w` versus `%v`. The reader knows
the language: this skill carries the decisions that go wrong in review, and
`references/` the rest.

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

An error discarded on purpose says why on the same line
(`n, _ := b.Write(p) // never returns a non-nil error`); a bare `_` is a
finding. Independent failures that must all be reported — validating several
fields, closing several resources — aggregate with `errors.Join`, which
`errors.Is` and `errors.AsType` see through:

```go
return errors.Join(closeErr, flushErr)
```

Related concurrent work returns through
[`errgroup.WithContext`](https://pkg.go.dev/golang.org/x/sync/errgroup): the
first failure cancels the rest and is what `Wait` returns.

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
the existing cause and a contextual message. Whatever the type, the result is
declared as `error`: a concrete `*PathError` result turns a nil pointer into a
non-nil interface
([go-defensive](../go-defensive/SKILL.md#common-pitfalls) owns the mechanism).

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
annotation adds nothing, return `err` directly. One error often serves two
audiences — the operator reading a log line and the caller matching with
`errors.Is` — and `fmt.Errorf("resolve %q: %w", sku, err)` serves both where
`%v` serves only the first.

> **Validation**: Run `bash scripts/check-errors.sh` to detect common
> anti-patterns. The [go-linting](../go-linting/SKILL.md) gate covers the
> rest — `go vet` catches `errorsas` and `lostcancel`, and `go fix -diff`
> flags `errors.As` calls that should be `errors.AsType` — and runs once, at
> the end of the task, not again here.

---

## Related Skills

- [go-naming](../go-naming/SKILL.md#error-names): `ErrX` sentinels and `XError` types.
- [go-testing](../go-testing/SKILL.md): `errors.Is`/`errors.AsType` under test, error-checking helpers.
- [go-defensive](../go-defensive/SKILL.md): panic versus error, recover guards.
- [go-style-core](../go-style-core/SKILL.md): nesting depth, early returns, `if`-init.
- [go-logging](../go-logging/SKILL.md): log levels and what a log line carries.
