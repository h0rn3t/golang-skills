---
name: go-documentation
description: Use when writing or reviewing Go doc comments, or creating exported types, functions, or packages even without a documentation request. Internal code comments belong to go-style-core.
allowed-tools: Bash(bash:*)
---

# Go Documentation

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). `go doc -ex` and
> `go doc <pkg>@<version>` require Go 1.27+; doc links and `#` headings Go 1.19+.

## Resource Routing

- `scripts/check-docs.sh` - Run when checking exported functions, types, methods, constants, and packages for missing doc comments.
- `scripts/check-docs-ast.go` - Implementation helper invoked by `check-docs.sh`; patch this when changing documentation analysis behavior.
- `assets/doc-template.go` - Use when starting a documented package or exported API.
- `references/CONVENTIONS.md` - Read when documenting parameters, context behavior, concurrency safety, cleanup, errors, named results, or a deprecation.
- `references/EXAMPLES.md` - Read when adding runnable examples or package examples.
- `references/FORMATTING.md` - Read when formatting Godoc lists, paragraphs, links, and code blocks.

---

## Doc Comments

> **Normative**: Every top-level exported name has a doc comment, in every
> package except `package main` and those under `internal/` or `cmd/` — the
> API other modules import, which is the set revive `exported` checks in the
> gate.

### Basic Rules

1. Begin with the name of the object being described
2. An article ("a", "an", "the") may precede the name
3. Use full sentences (capitalized, punctuated)

```go
// A Request represents a request to run a command.
type Request struct { ...

// Encode writes the JSON encoding of req to w.
func Encode(w io.Writer, req *Request) { ...
```

Unexported names, and exported names under `internal/` or `cmd/`, get a doc
comment when their behavior is not obvious from the signature.

> **Validation**: `scripts/check-docs.sh` lists exported names without a doc comment outside `package main` — a comment of only directives or only a `Deprecated:` paragraph counts as none — skipping generated files, methods of unexported types, and the methods revive skips (`Error`, `Read`, `ServeHTTP`, `String`, `Write`, `Unwrap`). Unlike the gate it also reports `internal/` and `cmd/`, where a finding is advisory. Run it once at the end of the task, beside the [go-linting](../go-linting/SKILL.md) gate, which does not run it.

---

## Struct Documentation

Group fields with section comments. Mark optional fields with defaults:

```go
type Options struct {
    // General setup:
    Name  string
    Group *FooGroup

    // Customization:
    LargeGroupThreshold int // optional; default: 10
}
```

---

## Package Comments

> **Normative**: Every package must have exactly one package comment.

```go
// Package math provides basic constants and mathematical functions.
package math
```

- For `main` packages, use the binary name: `// The seed_generator command ...`
- For long package comments, use a `doc.go` file

---

## What to Document

> **Advisory**: Document non-obvious behavior, not obvious behavior.

| Topic | Document when... | Skip when... |
|-------|-----------------|--------------|
| Parameters | Non-obvious behavior, edge cases | Restates the type signature |
| Contexts | Behavior differs from standard cancellation | Standard `ctx.Err()` return |
| Concurrency | Ambiguous thread safety (e.g., read that mutates) | Read-only is safe, mutation is unsafe |
| Cleanup | Always document resource release | — |
| Errors | Sentinel values, error types (use `*PathError`) | — |
| Named results | Multiple params of same type, action-oriented names | Type alone is clear enough |

---

## Runnable Examples

> **Advisory**: Provide runnable `Example` functions in test files (`*_test.go`); [EXAMPLES.md](references/EXAMPLES.md) shows the form.

Examples appear in Godoc attached to the documented element. `go doc -ex
<symbol>` (Go 1.27+) lists them from the terminal, and `go doc <pkg>@<version>`
(Go 1.27+) reads the docs of a version you have not imported.

---

## Related Skills

- [go-naming](../go-naming/SKILL.md): the identifiers the comments describe.
- [go-testing](../go-testing/SKILL.md): the tests beside the examples; `Example` functions are [EXAMPLES.md](references/EXAMPLES.md)'s.
- [go-linting](../go-linting/SKILL.md): linters that enforce doc comment presence.
- [go-style-core](../go-style-core/SKILL.md): verbosity against clarity and concision.
