---
name: go-generics
description: Use when choosing or writing Go generics, constraints, type aliases versus definitions, or utilities for multiple types. Non-generic interface design belongs to go-interfaces.
---

# Go Generics and Type Parameters

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). Generic **methods**
> and inference in function-type conversions require Go 1.27+;
> generic type aliases require Go 1.24+; generics themselves, Go 1.18+.

## Resource Routing

- `references/CONSTRAINTS.md` - Read when composing constraints, using type sets, or choosing between generics and interfaces.

## When to Use Generics

Start with concrete types for a concrete task. Generalize when actual callers
need multiple types or the requested public contract is itself generic; do not
wait for a second local instantiation when implementing that explicit API.

```
When the public API is not already prescribed:
Do multiple types share identical logic?
├─ No  → concrete types
└─ Yes → Do they share a useful interface?
         ├─ Yes → interface
         └─ No  → generics
```

**Prefer generics when**: multiple types share identical logic (sort, filter,
map/reduce); the alternative is `any` plus type switching; you are building a
reusable container.

**Avoid adding generics when**: one concrete type satisfies the requested API;
an interface already models the required behavior; or generalization makes the
operation harder to read without serving a caller contract.

> "Write code, don't design types." — Robert Griesemer and Ian Lance Taylor

```go
// Bad: premature — only ever called with int
func Sum[T constraints.Integer | constraints.Float](vals []T) T { /* ... */ }

// Good
func SumInts(vals []int) int { /* ... */ }
```

---

## Generic Methods (Go 1.27+)

Methods may declare their own type parameters, independent of the receiver's.
This removes the old workaround of a package-level generic function taking the
receiver as its first argument.

```go
type Store struct{ data map[string]any }

// Go 1.27: the type parameter belongs to the method
func (s *Store) Get[T any](key string) (T, bool) {
    v, ok := s.data[key].(T)
    return v, ok
}

n, ok := store.Get[int]("count")
```

Constraints:

- A generic method cannot satisfy an interface — interface methods have no type
  parameters. If callers dispatch through an interface, keep the free function.
- The receiver's own type parameters stay on the receiver; do not redeclare them.
- Same restraint as generic functions: serve demonstrated callers or an
  explicitly requested generic API, not speculative future flexibility.

Go 1.27 extends function type inference to all assignments and conversions to
matching function types. Assignment to a typed variable already supported
inference in Go 1.21. Prefer inference when the call site remains clear.

---

## Type Parameter Naming

| Name | Typical Use |
|------|-------------|
| `T` | General type parameter |
| `K` / `V` | Map key / value type |
| `E` | Element/item type |

For complex constraints, a short descriptive name can clarify the parameter's role.

---

## Constraints

Prefer standard-library constraints over hand-written ones:

| Need | Use |
|---|---|
| `<`, `>` ordering | `cmp.Ordered` (Go 1.21+) — not `constraints.Ordered` |
| `==` only | `comparable` |
| Anything | `any` |
| Numeric union | Write a local union; the `constraints` module is still `x/exp` |

```go
type Numeric interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}
```

Use `~` for underlying types so named types (`type Celsius float64`) satisfy the
constraint, and `|` for unions.

### Hashing generic keys

For a generic container that needs its own hash table, use
`hash/maphash.Hasher[T]` and `maphash.ComparableHasher[T]` (Go 1.27+) instead of
hand-rolling a hash/equality pair:

```go
type table[K comparable, V any] struct {
    hasher maphash.Hasher[K] // maphash.ComparableHasher[K]{}
    seed   maphash.Seed
}
```

---

## Type Aliases vs Type Definitions

Use a definition for a distinct type with its own methods or invariants. An alias
preserves type identity and can support migration or a useful public vocabulary.
Generic aliases (Go 1.24+), such as `type Set[T comparable] = map[T]struct{}`,
are valid when that identity is intended. Avoid an extra name that helps no caller.

---

## Common Pitfalls

**Don't wrap standard library types.** A single-use generic is indirection:

```go
// Bad
type Set[T comparable] struct{ m map[T]struct{} }

// Better
seen := map[string]struct{}{}
```

**Don't use generics for interface satisfaction.** If `T` is only used to
satisfy an interface, accept the interface:

```go
// Bad
func Process[T io.Reader](r T) error

// Good
func Process(r io.Reader) error
```

**Don't over-constrain.** `comparable` beats `interface{ ~int | ~string }` when
you only need `==`.

---

## Quick Reference

| Topic | Guidance |
|-------|----------|
| When to use | Explicit generic API, or multiple types with shared logic and no adequate interface |
| Starting point | Concrete first unless callers or the explicit API require generics |
| Naming | `T`, `K`, `V`, `E` |
| Generic methods | Go 1.27+; cannot satisfy an interface |
| Ordering constraint | `cmp.Ordered`, never a hand-written one |
| Generic hashing | `maphash.ComparableHasher[T]` (Go 1.27+) |
| Type aliases | Preserve identity for migration or a useful public contract |
| Pitfall | Single-use generics, `T` used only as an interface |

---

## Related Skills

- **Interfaces vs generics**: See [go-interfaces](../go-interfaces/SKILL.md) when deciding whether an interface already models the shared behavior without generics
- **Type declarations**: See [go-style-core](../go-style-core/SKILL.md) when defining new types, type aliases, or choosing between type definitions and aliases
- **Documenting generic APIs**: See [go-documentation](../go-documentation/SKILL.md) when writing doc comments and runnable examples for generic functions
- **Naming type parameters**: See [go-naming](../go-naming/SKILL.md) when choosing names for type parameters or constraint interfaces
