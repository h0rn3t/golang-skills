---
name: go-data-structures
description: Use when creating or manipulating Go slices, maps, arrays, or sets, including new/make, append, copies, and nil versus empty JSON collections. Thread safety belongs to go-concurrency.
---

# Go Data Structures

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). `slices`/`maps`
> and `slices.Clone`/`maps.Clone` require Go 1.21+; `slices.Concat` Go 1.22+;
> `strings.CutLast`/`bytes.CutLast` Go 1.27+.

## Resource Routing

- `references/SLICES.md` - Read when managing slice capacity and aliasing, or when a subslice keeps a large backing array alive.

## Slices

### Reach for `slices` and `maps` First

The `slices` and `maps` packages (Go 1.21+) cover most hand-written loops.
Writing the loop instead is a reviewable defect, not a style choice.
`go fix -diff` on the edited package previews the loops a fixer exists for
(`slicescontains`, `slicesbackward`, `slicessort`, `mapsloop`, `minmax`);
`Equal`, `Concat`, `Insert`, and `Delete` have none and are written by hand.

| Loop you were about to write | Use |
|---|---|
| Compare | `slices.Equal`, `maps.Equal` |
| Concatenate | `slices.Concat` (Go 1.22+) |
| Insert/delete in the middle | `slices.Insert`, `slices.Delete` |
| Iterate in reverse | `slices.Backward` (Go 1.23+) |
| Split off the last segment | `strings.CutLast` / `bytes.CutLast` (Go 1.27+) |

```go
ns, name, ok := strings.CutLast(key, ".") // "app.http.requests" → "app.http", "requests", true
if !ok {
    ns, name = "", key // no ".": CutLast returned key, "", false
}
```

`slices.Clone` and `maps.Clone` preserve nilness: nil input yields nil, while
an initialized empty container stays non-nil. `slices.Collect`,
`slices.Sorted`, and `slices.Concat` return nil for an empty result —
`Concat` even when every input is a non-nil empty slice. An unconditional
`make` plus copy produces a non-nil container even for nil input. Preserve
that allocation when JSON, a later map write, or observable capacity requires
it — see [Declaring Empty Slices](#declaring-empty-slices). Check the contract
before replacing a loop:

```go
all := slices.Concat(a, b)
if all == nil {
    all = []Item{} // encoding/json v1 writes nil as null; this list's contract is []
}
```

These clones are shallow: references inside elements or map values remain
shared. Copy depth and type-specific `Clone` contracts belong to
[go-defensive](../go-defensive/references/BOUNDARY-COPYING.md#copy-depth-is-part-of-the-contract).

### Update and Filter an Existing Map

When updates must overwrite matching keys before filtering the resulting map:

```go
maps.Copy(dst, updates)
maps.DeleteFunc(dst, func(_ string, score int) bool { return score < minimum })
```

Both operations mutate `dst`, so existing aliases observe the changes.
The destination must be initialized if the source contains entries.

### Declaring Empty Slices

Prefer nil slices over empty literals:

```go
// Good: nil slice
var t []string

// Avoid: non-nil but zero-length
t := []string{}
```

Both have `len` and `cap` of zero, but the nil slice is the preferred style.

**JSON depends on the encoder**: `encoding/json` v1 encodes a nil slice or
nil map as `null`, and `[]string{}` as `[]`, `map[string]int{}` as `{}`. JSON
v2 defaults encode nil non-byte slices as `[]` and nil maps as `{}`;
`FormatNilSliceAsNull(true)` and `FormatNilMapAsNull(true)` restore null.
Require non-nil only when the selected encoder/options or another caller
contract needs it. See
[JSON v2 defaults](../go-http/references/JSON-V2.md#defaults-that-can-change-the-contract).

When designing interfaces, avoid distinguishing between nil and non-nil
zero-length slices.

---

## Maps

### Implementing a Set

Use `map[T]struct{}` when the map is only a set. The empty struct takes no
storage and makes membership intent explicit.

Use boolean map values only when the value carries a separate meaning beyond
presence.

---

## Copying

A struct whose methods have pointer receivers or that holds a lock
(`bytes.Buffer`, `sync.Mutex`) is passed by pointer, never copied;
[go-functions](../go-functions/SKILL.md#pointers-to-interfaces) owns the rule.

---

## Related Skills

- [go-defensive](../go-defensive/SKILL.md): copying slices and maps at API boundaries.
- [go-performance](../go-performance/SKILL.md): capacity hints for known workloads.
- [go-style-core](../go-style-core/SKILL.md): range forms and `new`/`make`/`var`/literal choices.
