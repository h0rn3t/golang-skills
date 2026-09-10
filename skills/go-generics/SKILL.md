---
name: go-generics
description: Use when choosing or writing Go generics, constraints, type aliases versus definitions, or utilities for multiple types. Non-generic interface design belongs to go-interfaces.
---

# Go Generics and Type Parameters

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). Generic **methods**
> and inference in function-type conversions require Go 1.27+; self-referential
> constraints Go 1.26+; generic type aliases Go 1.24+; generics themselves, Go
> 1.18+. Neither the compiler nor `stdversion` gates these language features
> by the `go` directive — code that uses them compiles on a 1.27 toolchain with
> `go 1.26` in go.mod and fails on a real 1.26 toolchain; verify on the CI
> toolchain.

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

- A generic method cannot satisfy an interface; [go-interfaces](../go-interfaces/SKILL.md#generic-methods-cannot-satisfy-interfaces)
  owns that rule and the assertion that catches it.
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

### Self-referential constraints (Go 1.26+)

A type parameter may appear in its own constraint, so a constraint can require
that a type combines with, compares to, or produces its own kind:

```go
type Adder[A Adder[A]] interface{ Add(A) A }

func Sum[A Adder[A]](zero A, xs ...A) A {
    for _, x := range xs {
        zero = zero.Add(x)
    }
    return zero
}
```

Reach for it only when the method genuinely takes or returns the implementing
type; `comparable` or `cmp.Ordered` covers equality and ordering. Go 1.25
rejects the declaration as an invalid recursive type, and the `go` directive
does not gate it (see the Compatibility note).

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

## Related Skills

- **Interfaces vs generics**: See [go-interfaces](../go-interfaces/SKILL.md) when deciding whether an interface already models the shared behavior without generics
- **Type declarations**: See [go-style-core](../go-style-core/SKILL.md) when defining new types, type aliases, or choosing between type definitions and aliases
- **Documenting generic APIs**: See [go-documentation](../go-documentation/SKILL.md) when writing doc comments and runnable examples for generic functions
- **Naming type parameters**: See [go-naming](../go-naming/SKILL.md) when choosing names for type parameters or constraint interfaces
