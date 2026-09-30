# Copying Slices and Maps at API Boundaries

> Sources: source/uber-go-style/style.md; https://pkg.go.dev/slices, https://pkg.go.dev/maps
> Authority: advisory
> Minimum Go: `slices.Clone` / `maps.Clone` 1.21; `url.Values.Clone` 1.27
> Last verified: 2026-08-29

Slices and maps contain references to their underlying data. Copy them at API
boundaries so callers cannot mutate internal state, or vice versa.

## Use the stdlib clone functions

Prefer `slices.Clone` and `maps.Clone` (Go 1.21+) when their nil and capacity
behavior fits the contract. An unconditional `make` creates a non-nil result
even for nil input; keep it when callers need a writable map, an observable
slice capacity, or a JSON array with an encoder that maps nil to null. A copy
is not redundant merely because a clone function exists.

```go
// Good: snapshot under the lock, clone before releasing it
func (s *Stats) Snapshot() map[string]int {
  s.mu.Lock()
  defer s.mu.Unlock()
  return maps.Clone(s.counters)
}
```

For a JSON array contract with v1 or v2's `FormatNilSliceAsNull(true)`, keep the
copy non-nil even when the input is nil:

```go
// Good: still encodes as [], not null, for a nil input under JSON v1
func (q *Queue) Items() []Item { return append(make([]Item, 0, len(q.items)), q.items...) }
```

The same holds for a map a caller is expected to write into: `maps.Clone` of a
nil map returns nil, and the first write panics.

## Copy depth is part of the contract

A shallow copy duplicates the outer value or container and keeps references to
nested data. A deep copy duplicates nested mutable data to the depth required
for independent ownership. Go has no general-purpose deep-copy operation.

`slices.Clone` and `maps.Clone` are shallow: pointer elements and reference
values such as slices or maps remain shared. Struct assignment is also shallow
for fields that contain references. The name `Clone` alone does not specify
copy depth; follow that method's documented contract. For example,
`url.Values.Clone` (Go 1.27+) copies the map and its `[]string` values.

| Source | Operation | Copy depth |
|---|---|---|
| `[]int`, `map[string]int` | `slices.Clone`, `maps.Clone` | Independent outer data; elements/values are scalars |
| `[]*Trip` | `slices.Clone` | New slice; the `*Trip` values are shared |
| `map[string][]string` | `maps.Clone` | New map; the `[]string` values are shared |
| `url.Values` | `v.Clone()` (Go 1.27+) | New map and independent value slices |

```go
// Bad: the maps.Clone copy shares every []string with the caller
params := maps.Clone(userParams)

// Good
params := userParams.Clone() // url.Values.Clone, Go 1.27+
```

For a struct with reference fields, an explicit `Clone` method must still meet
the required copy depth; the method name does not make it deep. Copy each
nested mutable field only when the ownership contract requires it.

Do not replace a shallow copy with a deep copy, or a deep copy with a shallow
copy, as a style cleanup. They have different aliasing and cost contracts. If
the required ownership changes, make that semantic change explicit in the API
contract and its tests.

## When copies are not needed

Defensive copies have a cost. Skip them when:

- The data is **immutable by convention** and the doc comment says so
- The slice/map is **created fresh** for the caller and never stored internally
- Profiling shows the copy is a real hot-path cost (measure, do not assume)

When in doubt, copy. The cost is almost always negligible next to the bugs
shared references cause.
