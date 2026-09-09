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

Start with concrete types. Generalize only when a second type appears.

```
Do multiple types share identical logic?
├─ No  → concrete types
└─ Yes → Do they share a useful interface?
         ├─ Yes → interface
         └─ No  → generics
```

**Prefer generics when**: multiple types share identical logic (sort, filter,
map/reduce); the alternative is `any` plus type switching; you are building a
reusable container.

**Avoid generics when**: only one type is ever instantiated; an interface
already models the shared behavior; the generic version is harder to read.

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
- Same restraint as generic functions: add the parameter when a second type
  actually appears, not in anticipation.

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

`hash/maphash.Hasher[K]` (Go 1.27+) is the seam between a hash-based container
and its keys: one hash function plus the matching equivalence relation. It buys
what the built-in map cannot do — keys that are not `comparable`, or an equality
other than `==` — so constrain the container to `any`, not `comparable`.
`maphash.ComparableHasher[K]{}` is the ready-made implementation for ordinary
keys; a custom `Hasher` is a two-method type.

```go
type table[K, V any] struct {
    hasher maphash.Hasher[K] // maphash.ComparableHasher[string]{} at an ordinary call site
    seed   maphash.Seed      // one seed per table instance
}

func (t *table[K, V]) hash(k K) uint64 {
    var h maphash.Hash
    h.SetSeed(t.seed)
    t.hasher.Hash(&h, k)
    return h.Sum64()
}
```

A `Hasher` must be logically stateless, and `Equal(x, y)` must imply an equal
hash — the container relies on both. `ComparableHasher` is itself constrained to
`comparable`, so it can only be named where the key type is concrete: inside a
`K any` container the field stays the `Hasher[K]` interface.

---

## Type Aliases vs Type Definitions

Type aliases (`type Old = new.Name`) are rare — use for package migration or
gradual API refactoring. Generic type aliases (Go 1.24+) are legal
(`type Set[T comparable] = map[T]struct{}`) and carry the same caution: an alias
adds a name, not a type, so it buys nothing but a migration path.

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
| When to use | Multiple types, identical logic, no adequate interface |
| Starting point | Concrete first; generalize on the second type |
| Naming | `T`, `K`, `V`, `E` |
| Generic methods | Go 1.27+; cannot satisfy an interface |
| Ordering constraint | `cmp.Ordered`, never a hand-written one |
| Generic hashing | `maphash.ComparableHasher[T]` (Go 1.27+) |
| Type aliases | Migration only |
| Pitfall | Single-use generics, `T` used only as an interface |

---

## Related Skills

- **Interfaces vs generics**: See [go-interfaces](../go-interfaces/SKILL.md) when deciding whether an interface already models the shared behavior without generics
- **Type declarations**: See [go-style-core](../go-style-core/SKILL.md) when defining new types, type aliases, or choosing between type definitions and aliases
- **Documenting generic APIs**: See [go-documentation](../go-documentation/SKILL.md) when writing doc comments and runnable examples for generic functions
- **Naming type parameters**: See [go-naming](../go-naming/SKILL.md) when choosing names for type parameters or constraint interfaces
