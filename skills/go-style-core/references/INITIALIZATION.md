# Initialization and Composite Literals

> Sources: source/google-go-styleguide/decisions.md; source/uber-go-style/style.md; COMPATIBILITY.md
> Authority: project policy for preferred forms; language semantics follow Go
> Minimum Go: `any` Go 1.18; `new(expr)` Go 1.26

## Preserve the Value Contract

Choose zero, nil, and empty values by the API's meaning before choosing syntax.
For example, with `encoding/json`, a nil slice encodes as `null` and an empty
non-nil slice as `[]`. Replacing one with the other can change behavior.
Keep explicit zero fields when they clarify a test case or configuration.

## Structs

Prefer keyed fields for readable, resilient construction. This is a style
default for local types; `go vet`'s composite-literal check catches unkeyed
external types subject to its exceptions. Small obvious test rows and types
with a documented positional convention can follow the repository's practice.

```go
cfg := Config{
    Host:    "localhost",
    Port:    8080,
    Timeout: 30 * time.Second,
}
```

Use `var user User` for an intentional zero-value struct. Omit redundant zero
fields when that preserves clarity. To allocate and initialize a struct,
prefer `&T{Field: value}`; both `&T{}` and `new(T)` produce a pointer to a
zero-value `T`.

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
temporary on older targets. `go fix -diff -newexpr ./...` previews the rewrite
through the [shared gate](../../go-linting/SKILL.md).

## Maps

| Need | Form |
|---|---|
| Intentional nil map | `var m map[string]int` |
| Empty map that will receive entries | `m := make(map[string]int)` |
| Known initial entries | `m := map[string]int{"a": 1}` |

A nil map can be read but assigning an entry panics. This reference owns
declaration form; [go-data-structures](../../go-data-structures/SKILL.md) owns
collection choice and aliasing, and [go-performance](../../go-performance/SKILL.md)
owns capacity decisions.

## Literal Layout

Keep a short literal on one line when readable. For a multiline literal, keep
the opening brace with the expression, use trailing commas, and align the
closing brace with the opening line. Let `gofmt` settle indentation.

```go
items := []*Item{
    {Name: "widget", Price: 9.99},
    {Name: "gadget", Price: 19.99},
}
headers := map[string]string{
    "Content-Type": "application/json",
    "X-Request-ID": reqID,
}
```

Omit repeated element types (and redundant `&T` in pointer-element literals);
`gofmt -s` applies these simplifications. Adjacent nested braces are fine when
the structure remains clear; follow local layout rather than a field-count rule.
Extract a long function literal when a named operation clarifies its role,
while preserving captures and side-effect order.

## Raw Strings and `any`

Raw strings make regex, SQL, JSON, and multiline text easier to read when they
avoid escaped quotes. Preserve the actual bytes, including newlines, when
changing the literal form. Use `any` instead of `interface{}` in new code on
Go 1.18+; broader modernization remains subject to the task's scope.
