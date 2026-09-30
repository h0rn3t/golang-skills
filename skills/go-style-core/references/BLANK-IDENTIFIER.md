# Blank Identifier Patterns

> Sources: source/effective-go/effective_go.html (The blank identifier); source/uber-go-style/style.md (Verify Interface Compliance)
> Authority: advisory for usage patterns; language semantics follow the Go specification
> Last verified: 2026-09-10

## Multiple Assignment

Use `_` to discard unwanted values from multi-value expressions.

### Never Discard Errors Carelessly

If you truly don't need the error, document why:

```go
_ = logger.Sync() // best-effort flush; error is non-actionable
```

## Import for Side Effect

Import a package solely for its `init()` side effects using the blank
identifier. This is commonly used to register drivers, codecs, or debug handlers that
wire themselves into a registry during `init()`. Where a side-effect import
may appear — `main` packages and tests, plus `import _ "embed"` in a library
file that embeds into a `string` or `[]byte` — is
[go-packages](../../go-packages/references/IMPORTS.md#blank-imports-import-_)'s rule.

## Interface Satisfaction Checks

Blank identifiers are also used in compile-time interface assertions, but that
rule is owned by [go-interfaces](../../go-interfaces/SKILL.md).
