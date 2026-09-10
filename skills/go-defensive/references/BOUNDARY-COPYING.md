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

`slices.Clone` and `maps.Clone` preserve nilness, including non-nil empty
inputs. Replacing an unconditional allocation with Clone changes JSON v1
output for nil input from `[]` to `null`. JSON v2 defaults already encode nil
non-byte slices as `[]`; check the selected encoder and compatibility options
in [go-data-structures](../../go-data-structures/SKILL.md#declaring-empty-slices).

For a JSON array contract with v1 or v2's `FormatNilSliceAsNull(true)`, keep the
copy non-nil even when the input is nil:

```go
// Good: still encodes as [], not null, for a nil input under JSON v1
func (q *Queue) Items() []Item { return append(make([]Item, 0, len(q.items)), q.items...) }
```

The same holds for a map a caller is expected to write into: `maps.Clone` of a
nil map returns nil, and the first write panics.

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
- Profiling shows the copy is a real hot-path cost (measure, do not assume)

When in doubt, copy. The cost is almost always negligible next to the bugs
shared references cause.
