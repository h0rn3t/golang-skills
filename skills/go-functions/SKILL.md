---
name: go-functions
description: Use when designing or reviewing Go function APIs — parameters, return values, signature readability, function ordering, Printf-style helpers, or constructors with optional configuration. Covers choosing ordinary parameters, config structs, or functional options. A routine function-body edit alone does not require this skill; route its actual topic to the relevant Go skill.
---

# Go Function Design

## Resource Routing

- `references/SIGNATURES.md` - Read when designing parameters, return values, named results, or signature readability.
- `references/PRINTF-STRINGER.md` - Read when using fmt verbs, Stringer, GoStringer, Formatter, or Printf-style function naming.
- `references/OPTIONS-VS-STRUCTS.md` - Read when choosing or implementing constructor configuration: config structs, functional options, defaults, validation, and caller ergonomics.

For error strategy, see [go-error-handling](../go-error-handling/SKILL.md).
For identifier names, see [go-naming](../go-naming/SKILL.md). Follow existing
API conventions and preserve signatures unless changing them is in scope.

---

## Function Grouping and Ordering

Organize functions in a file by these rules:

1. Functions sorted in **rough call order**
2. Functions **grouped by receiver**
3. **Exported** functions appear first, after `struct`/`const`/`var` definitions
4. `NewXxx`/`newXxx` constructors appear right after the type definition
5. Plain utility functions appear toward the end of the file

```go
type something struct{ ... }

func newSomething() *something { return &something{} }

func (s *something) Cost() int { return calcCost(s.weights) }

func (s *something) Stop() { ... }

func calcCost(n []int) int { ... }
```

---

## Function Signatures

Keep the signature on a single line when possible. When it must wrap, put **all
arguments on their own lines** with a trailing comma:

```go
func (r *SomeType) SomeLongFunctionName(
    foo1, foo2, foo3 string,
    foo4, foo5, foo6 int,
) {
    foo7 := bar(foo1)
}
```

Add `/* name */` comments for ambiguous arguments, or better yet, replace naked
`bool` parameters with custom types.

---

## Pointers to Interfaces

You almost never need a pointer to an interface. Pass interfaces as values — the
underlying data can still be a pointer.

```go
// Bad: pointer to interface
func process(r *io.Reader) { ... }

// Good: pass the interface value
func process(r io.Reader) { ... }
```

---

## Printf and Stringer

### Printf-style Function Names

Functions that accept a format string should end in `f` for `go vet` support.
Declare format strings as `const` when used outside `Printf` calls.

Prefer `%q` over `%s` with manual quoting when formatting strings for logging
or error messages — it safely escapes special characters and wraps in quotes:

```go
return fmt.Errorf("unknown key %q", key) // produces: unknown key "foo\nbar"
```

## Constructors and Optional Configuration

Choose the API by how callers configure it; an option count alone does not
justify functional options.

| Caller needs | Starting point |
|---|---|
| A few required inputs | Ordinary parameters; meaningful types for ambiguous values |
| Settings usually supplied together, especially internal APIs | Config struct with documented zero values/defaults |
| Public API with independently optional settings and useful defaults | Consider functional options; weigh caller clarity against added types and helpers |

Keep required inputs separate from optional settings. Apply defaults first,
then caller settings, then validate the final configuration, including
interdependent fields. Distinguish unset from explicit zero when they differ.
Do not change an existing config API merely to introduce `With*` helpers.
Check omitted, explicit zero/nil, and repeated settings against the documented
constructor contract. Resolve known mismatches in the code before presenting
it; a caveat does not make a contradictory invariant hold.

When functional options fit and the repository has no established pattern,
this pack prefers an exported `Option` interface with an unexported `apply`
method. Preserve an existing closure-based convention. Implementation and
tradeoffs live in the constructor reference above.

---

## Quick Reference

| Topic | Rule |
|-------|------|
| File ordering | Type -> constructor -> exported -> unexported -> utils |
| Signature wrapping | All args on own lines with trailing comma |
| Naked parameters | Add `/* name */` comments or use custom types |
| Pointers to interfaces | Almost never needed; pass interfaces by value |
| Printf function names | End with `f` for `go vet` support |
| Constructor configuration | Choose by caller needs; defaults, overrides, then validation |

---

## Related Skills

- **Error returns**: See [go-error-handling](../go-error-handling/SKILL.md) when designing error return patterns or wrapping errors in multi-return functions
- **Naming conventions**: See [go-naming](../go-naming/SKILL.md) when naming functions, methods, or choosing getter/setter patterns
- **Option interfaces**: See [go-interfaces](../go-interfaces/SKILL.md) when the abstraction itself needs design or review
- **API documentation**: See [go-documentation](../go-documentation/SKILL.md) when documenting exported constructors, defaults, or `With*` functions
- **Formatting principles**: See [go-style-core](../go-style-core/SKILL.md) when deciding line length, naked returns, or signature formatting
