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

Interfaces belong in the package that **consumes** values, not the package that
**implements** them. Return concrete (usually pointer or struct) types from
constructors so new methods can be added without refactoring.

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
// Bad: producer defines and returns its own interface
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

Return an interface from a constructor only when it is a **pre-existing** one
owned by the consumer or the standard library, and the type has no exported
methods beyond it:

```go
func NewHash() hash.Hash32 {
    return &myHash{}  // unexported type; hash.Hash32 is the stdlib's interface
}
```

An interface declared beside the constructor in order to be returned is the
Bad case above, whatever the type's method set.

---

## Type Assertions: Comma-Ok Idiom

Without checking, a failed assertion panics. Use the comma-ok idiom whenever
the dynamic type is data-dependent; a bare `x.(T)` is acceptable only where a
mismatch is a programming error that should panic, such as a recover guard
re-panicking a foreign value ([PANIC-RECOVER.md](../go-defensive/references/PANIC-RECOVER.md)).
Reflection code uses `reflect.TypeAssert[T]` (Go 1.25+) instead of
`v.Interface().(T)`.

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
becomes part of your public API, so adding, removing, or replacing the embedded
type is a breaking change. Use an unexported field and forward the methods you
mean to export; [EMBEDDING.md](references/EMBEDDING.md#dont-embed-in-public-structs)
has the before/after. This skill owns the rule; go-defensive routes here.

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

If in doubt, use a pointer receiver. Don't mix receiver types on a single
type — if any method needs a pointer, use pointers for all methods. Value
receivers fit maps, funcs, channels, slices that are not resliced or
reallocated, and small immutable structs or basic types;
[RECEIVER-TYPE.md](references/RECEIVER-TYPE.md) has the full decision list.

---

## Related Skills

- [go-naming](../go-naming/SKILL.md): the `-er` suffix, receiver names.
- [go-error-handling](../go-error-handling/SKILL.md): implementing `error`, custom error types, `errors.As` matching.
- [go-generics](../go-generics/SKILL.md): whether generics are needed or an interface suffices.
- [go-functions](../go-functions/SKILL.md): interface-based Option patterns for constructors.
- [go-defensive](../go-defensive/SKILL.md): assertions as part of an API-boundary hardening pass.
