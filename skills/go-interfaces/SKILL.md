---
name: go-interfaces
description: Use when designing or implementing Go interfaces, abstractions, mockable boundaries, embedding, assertions, type switches, or interface versus concrete API types. Generics belong to go-generics.
allowed-tools: Bash(bash:*)
---

# Go Interfaces and Composition

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). Generic methods
> require Go 1.27+.

## Resource Routing

- `scripts/check-interface-compliance.sh` - Run as a heuristic to find exported interfaces that may need compile-time assertions.
- `scripts/check-interface-compliance.go` - Implementation helper invoked by `check-interface-compliance.sh`; patch this when changing method-set analysis.
- `references/EMBEDDING.md` - Read when embedding interfaces or structs in public APIs.
- `references/RECEIVER-TYPE.md` - Read when pointer/value receivers affect interface satisfaction.

---

## Accept Interfaces, Return Concrete Types

Define a new interface at the consumer that needs substitution. Prefer concrete
constructor returns when callers use that type's API; return an existing
interface when it is the full public contract and implementation hiding is
intentional, as in [Generality](#generality-hide-implementation-expose-interface).

```go
// Good: consumer defines the interface it needs
package consumer

type Thinger interface { Thing() bool }

func Foo(t Thinger) string { ... }
```

```go
// Good: producer returns concrete type
package producer

type Thinger struct{ ... }
func (t Thinger) Thing() bool { ... }
func NewThinger() Thinger { return Thinger{ ... } }
```

```go
// Unnecessary when callers need the concrete API, without an implementation-hiding contract
package producer

type Thinger interface { Thing() bool }
type defaultThinger struct{ ... }
func NewThinger() Thinger { return defaultThinger{ ... } }
```

**Do not define interfaces before they are used.** Identify the consumer and
the substitution it needs, including a test double. One production
implementation neither requires nor rules out an interface.

---

## Generality: Hide Implementation, Expose Interface

If a type exists only to implement an interface with no exported methods beyond
that interface, return the interface from constructors to hide the implementation:

```go
func NewHash() hash.Hash32 {
    return &myHash{}  // unexported type
}
```

Benefits: implementation can change without affecting callers, substituting
algorithms requires only changing the constructor call.

---

## Type Assertions: Comma-Ok Idiom

Without checking, a failed assertion causes a runtime panic. Always use the
comma-ok idiom to test safely:

```go
str, ok := value.(string)
if ok {
    fmt.Printf("string value is: %q\n", str)
}
```

To check if a value implements an interface:

```go
if _, ok := val.(json.Marshaler); ok {
    fmt.Printf("value %v implements json.Marshaler\n", val)
}
```

---

## Type Switch

It's idiomatic to reuse the variable name (`t := t.(type)`) — the variable has
the correct type in each case branch. When a case lists multiple types
(`case int, int64:`), the variable has the interface type.

---

## Generic Methods Cannot Satisfy Interfaces

Go 1.27 lets a method declare its own type parameters, but interface methods
still cannot. A type whose only implementation of `Do` is `Do[T any]()` does
**not** satisfy `interface{ Do() }` — the compile-time assertion
`var _ I = (*T)(nil)` is how you find out. When callers dispatch through an
interface, keep a non-generic method (or a free generic function) as the entry
point. See [go-generics](../go-generics/SKILL.md).

---

## Embedding

Avoid embedding types in public structs — the inner type's full method set
becomes part of your public API. Use unexported fields instead.

---

## Interface Satisfaction Checks

Use a blank identifier assignment to verify a type implements an interface at
compile time:

```go
var _ json.Marshaler = (*RawMessage)(nil)
```

This causes a compile error if `*RawMessage` doesn't implement `json.Marshaler`.

Use this pattern when:
- There are no static conversions that would verify the interface automatically
- The type must satisfy an interface for correct behavior (e.g., custom JSON
  marshaling)
- Interface changes should break compilation, not silently degrade

**Don't** add these checks for every interface — only when no other static
conversion would catch the error.

> **Validation**: Use `scripts/check-interface-compliance.sh` when a heuristic
> scan would help find missing assertions. Review its candidates against the
> conditions above; a finding is not a requirement to add an assertion.

---

## Receiver Type

Choose receivers by mutation, copying safety, and the required method set.
Consistent pointer receivers are a useful default for new mutable types; small
value types can use value receivers. Preserve existing value-method interface
satisfaction rather than converting all methods for consistency alone.

---

## Quick Reference

| Concept | Pattern | Notes |
|---------|---------|-------|
| Consumer owns interface | Define interfaces where used | Not in the implementing package |
| Safe type assertion | `v, ok := x.(Type)` | Returns zero value + false |
| Type switch | `switch v := x.(type)` | Variable has correct type per case |
| Interface embedding | `type RW interface { Reader; Writer }` | Union of methods |
| Struct embedding | `type S struct { *T }` | Promotes T's methods |
| Interface check | `var _ I = (*T)(nil)` | Compile-time verification |
| Generality | Return interface from constructor | Hide implementation |

---

## Related Skills

- **Interface naming**: See [go-naming](../go-naming/SKILL.md) when naming interfaces (the `-er` suffix convention) or choosing receiver names
- **Error types**: See [go-error-handling](../go-error-handling/SKILL.md) when implementing the `error` interface, custom error types, or `errors.As` matching
- **Generics vs interfaces**: See [go-generics](../go-generics/SKILL.md) when deciding whether generics are needed or an interface already suffices
- **Functional options**: See [go-functions](../go-functions/SKILL.md) when using an interface-based Option pattern for flexible constructors
- **Defensive boundaries**: See [go-defensive](../go-defensive/SKILL.md) when interface assertions are one part of a broader API-boundary hardening pass
