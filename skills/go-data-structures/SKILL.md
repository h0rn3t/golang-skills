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
Writing the loop instead is a reviewable defect, not a style choice —
`go fix ./...` rewrites many of them automatically.

| Loop you were about to write | Use |
|---|---|
| Compare | `slices.Equal`, `maps.Equal` |
| Concatenate | `slices.Concat` (Go 1.22+) |
| Insert/delete in the middle | `slices.Insert`, `slices.Delete` |
| Iterate in reverse | `slices.Backward` |
| Split off the last segment | `strings.CutLast` / `bytes.CutLast` (Go 1.27+) |

`slices.Clone` and `maps.Clone` preserve nilness: nil input yields nil, while
an initialized empty container stays non-nil. `slices.Collect` and
`slices.Sorted` return nil for an empty iterator. An unconditional `make` plus
copy produces a non-nil container even for nil input. Preserve that allocation
when JSON, a later map write, or observable capacity requires it — see
[Declaring Empty Slices](#declaring-empty-slices). Check the contract before
replacing a loop.

These clones are shallow: references inside elements or map values remain
shared. Copy depth and type-specific `Clone` contracts belong to
[go-defensive](../go-defensive/references/BOUNDARY-COPYING.md#copy-depth-is-part-of-the-contract).

```go
dir, file, ok := strings.CutLast("a/b/c.txt", "/") // "a/b", "c.txt", true
```

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

**JSON depends on the encoder**: `encoding/json` v1 encodes a nil slice as
`null` and `[]string{}` as `[]`. JSON v2 defaults encode nil non-byte slices
as `[]`; `FormatNilSliceAsNull(true)` restores null. Require non-nil only when
the selected encoder/options or another caller contract needs it. See
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

Be careful when copying a struct from another package. If the type has methods
on its pointer type (`*T`), copying the value can cause aliasing bugs.

**General rule:** Do not copy a value of type `T` if its methods are associated
with the pointer type `*T`. This applies to `bytes.Buffer`, `sync.Mutex`,
`sync.WaitGroup`, and types containing them.

```go
type SafeCounter struct {
    mu    sync.Mutex
    count int
}

// Bad: copying a mutex
var mu sync.Mutex
mu2 := mu // almost always a bug

// Good: pass by pointer
func increment(sc *SafeCounter) {
    sc.mu.Lock()
    sc.count++
    sc.mu.Unlock()
}
```

---

## Related Skills

- [go-defensive](../go-defensive/SKILL.md): copying slices and maps at API boundaries.
- [go-performance](../go-performance/SKILL.md): capacity hints for known workloads.
- [go-style-core](../go-style-core/SKILL.md): range forms and `new`/`make`/`var`/literal choices.
