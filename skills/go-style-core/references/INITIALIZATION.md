# Initialization and Composite Literals

> Sources: source/google-go-styleguide/decisions.md; source/uber-go-style/style.md; COMPATIBILITY.md
> Authority: project policy for preferred forms; language semantics follow Go
> Target Go: 1.27
> Last verified: 2026-09-10

## Structs

Prefer keyed fields for readable, resilient construction. This is a style
default for local types; `go vet`'s composite-literal check catches unkeyed
external types subject to its exceptions. Small obvious test rows and types
with a documented positional convention can follow the repository's practice.

Use `var user User` for an intentional zero-value struct. Omit redundant zero
fields when that preserves clarity. To allocate and initialize a struct,
prefer `&T{Field: value}`; both `&T{}` and `new(T)` produce a pointer to a
zero-value `T`.

### Embedded Fields in Go 1.27

Set an unambiguous promoted field directly when the nested literal only adds
ceremony. Keep the embedded type; no flattening of the data model is needed:

```go
type Audit struct { CreatedBy string }
type Document struct {
    Audit
    Name string
}

doc := Document{CreatedBy: owner, Name: name}
```

This initializes `doc.Audit.CreatedBy`. Do not combine `Audit: ...` with
`CreatedBy: ...` in the same literal, or use this shorthand through an embedded
pointer. Keep explicit nesting when names are ambiguous or it makes the value
clearer. For existing code, preserve expression evaluation order and preview
the scoped `embedlit` modernizer through [go-linting](../../go-linting/SKILL.md).
See the [Go 1.27 language changes](https://go.dev/doc/go1.27#language).

## Pointers to Optional Values

`new(expr)` (Go 1.26+) allocates and initializes a value in one expression:

```go
// Before
n := computeLimit()
cfg.Limit = &n

// Go 1.26+
cfg.Limit = new(computeLimit())
```

This is useful when nil means unset and an explicit zero has a different
meaning. Check the module's language version before using it; retain the
temporary on older targets. The temporary-then-address form above is a hand
edit: `go fix -diff -newexpr ./...` rewrites only a pointer helper, previewed
through the [shared gate](../../go-linting/SKILL.md).

## Maps

| Need | Form |
|---|---|
| Intentional nil map | `var m map[string]int` |
| Empty map that will receive entries | `m := make(map[string]int)` |
| Known initial entries | `m := map[string]int{"a": 1}` |

This reference owns
declaration form; [go-data-structures](../../go-data-structures/SKILL.md) owns
collection choice and aliasing, and [go-performance](../../go-performance/SKILL.md)
owns capacity decisions.
