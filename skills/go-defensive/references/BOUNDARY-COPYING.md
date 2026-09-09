# Copying Slices and Maps at API Boundaries

> Sources: source/uber-go-style/style.md; https://pkg.go.dev/slices, https://pkg.go.dev/maps
> Authority: advisory
> Minimum Go: `slices.Clone` / `maps.Clone` 1.21; `url.Values.Clone` 1.27
> Last verified: 2026-08-29

Slices and maps contain references to their underlying data. Copy when the
ownership contract requires independent mutation; preserve intentional shared
views and caller-owned fresh results.

## Use the stdlib clone functions

Use `slices.Clone` and `maps.Clone` (Go 1.21+) for shallow copies when their
nil, capacity, and element-sharing behavior matches the contract. Retain
explicit copying when deeper ownership or another representation is required.

### Receiving

```go
// Bad: caller keeps a live reference to d.trips
func (d *Driver) SetTrips(trips []Trip) { d.trips = trips }

// Good
func (d *Driver) SetTrips(trips []Trip) { d.trips = slices.Clone(trips) }
```

```go
// Bad
func (s *Server) SetConfig(cfg map[string]string) { s.config = cfg }

// Good
func (s *Server) SetConfig(cfg map[string]string) { s.config = maps.Clone(cfg) }
```

### Returning

```go
// Bad: exposes internal state
func (q *Queue) Items() []Item { return q.items }

// Good
func (q *Queue) Items() []Item { return slices.Clone(q.items) }
```

```go
// Good: snapshot under the lock, clone before releasing it
func (s *Stats) Snapshot() map[string]int {
  s.mu.Lock()
  defer s.mu.Unlock()
  return maps.Clone(s.counters)
}
```

Note the nil behavior: `slices.Clone(nil)` and `maps.Clone(nil)` return `nil`.
Replacing an always-allocated result with a clone can therefore change
`encoding/json` v1 output from `[]` to `null` for nil input; cloning a non-nil
empty slice preserves that distinction. Default v2 collection encoding differs — see
[go-data-structures](../../go-data-structures/SKILL.md).

## Clone is shallow

`Clone` copies one level. If the element type contains a reference, the copy
still aliases it:

| Type | `Clone` gives you | What you need |
|---|---|---|
| `[]int`, `map[string]int` | A real, independent copy | Nothing more |
| `[]*Trip` | New slice, **same pointers** | Clone each element too |
| `map[string][]string` | New map, **same slices** | Clone each value |
| `url.Values` | — | `v.Clone()` (Go 1.27+), which deep-copies the value slices |
| `*url.URL` | — | `u.Clone()` (Go 1.27+) |
| `http.Header` | — | `h.Clone()` |

```go
// Bad: the maps.Clone copy shares every []string with the caller
params := maps.Clone(userParams)

// Good
params := userParams.Clone() // url.Values.Clone, Go 1.27+
```

For a struct with reference fields, write an explicit `Clone` method rather
than relying on assignment — struct assignment is shallow for the same reason.

## When copies are not needed

Defensive copies have a cost. Skip them when:

- The data is **immutable by convention** and the doc comment says so
- The slice/map is **created fresh** for the caller and never stored internally
- The caller explicitly transfers ownership and no longer accesses the data

If profiling shows a required copy is expensive, change representation or
ownership only within scope; performance alone does not make aliasing safe.
