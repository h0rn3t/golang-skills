---
name: go-data-structures
description: Use when creating or manipulating Go slices, maps, arrays, or sets, including new/make, append, copies, and nil versus empty JSON collections. Thread safety belongs to go-concurrency.
---

# Go Data Structures

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). `slices`/`maps`
> and `slices.Clone`/`maps.Clone` require Go 1.21+; `strings.SplitSeq`
> Go 1.24+; `strings.CutLast`/`bytes.CutLast` Go 1.27+.

## Resource Routing

- `references/SLICES.md` - Read when deciding nil versus empty slices, copying slices, or managing slice capacity and aliasing.

## Choosing a Data Structure

```
What do you need?
├─ Ordered collection of items
│  ├─ Fixed size known at compile time → Array [N]T
│  └─ Dynamic size → Slice []T
│     ├─ Know approximate size? → make([]T, 0, capacity)
│     └─ Unknown size or nil-safe for JSON? → var s []T (nil)
├─ Key-value lookup
│  └─ Map map[K]V
│     ├─ Know approximate size? → make(map[K]V, capacity)
│     └─ Need a set? → map[T]struct{} (zero-size values)
└─ Need to pass to a function?
   └─ Copy at the boundary if the caller might mutate it
```

> **When this skill does NOT apply**: For concurrent access to data structures (mutexes, atomic operations), see [go-concurrency](../go-concurrency/SKILL.md). For defensive copying at API boundaries, see [go-defensive](../go-defensive/SKILL.md). For pre-sizing capacity for performance, see [go-performance](../go-performance/SKILL.md).

---

## Slices

### Reach for `slices` and `maps` First

The `slices` and `maps` packages (Go 1.21+) cover most hand-written loops.
Writing the loop instead is a reviewable defect, not a style choice —
`go fix ./...` rewrites many of them automatically.

| Loop you were about to write | Use |
|---|---|
| Search for a value | `slices.Contains`, `slices.IndexFunc` |
| Sort | `slices.Sort`, `slices.SortFunc` (not `sort.Slice`) |
| Copy | `slices.Clone`, `maps.Clone` |
| Merge entries into an existing map | `maps.Copy` |
| Delete map entries by predicate | `maps.DeleteFunc` |
| Compare | `slices.Equal`, `maps.Equal` |
| Collect keys/values | `slices.Collect(maps.Keys(m))` |
| Insert/delete in the middle | `slices.Insert`, `slices.Delete` |
| Iterate in reverse | `slices.Backward` |
| Split a string once, iterate | `strings.SplitSeq` (no slice allocated) |
| Split off the last segment | `strings.CutLast` / `bytes.CutLast` (Go 1.27+) |

`slices.Clone`, `maps.Clone`, `slices.Collect`, and `slices.Sorted` return nil
for empty input, where a `make`+`copy` pair returns a non-nil empty container.
When the result reaches JSON or a later map write, that difference is
observable — see [Declaring Empty Slices](#declaring-empty-slices) below and
keep the allocation. Replacing a loop is not free of contract until you check it.

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
`Copy` leaves unrelated keys intact; it is a merge, not replacement or deep
cloning. The destination must be initialized if the source contains entries.
With distinct source and destination maps, the source is unchanged; if they
are the same map, filtering naturally changes both. Preserve copy/filter order
and use a predicate without map mutations or iteration-order assumptions.
See [maps.Copy](https://pkg.go.dev/maps#Copy) and
[maps.DeleteFunc](https://pkg.go.dev/maps#DeleteFunc).

### The append Function

**Always assign the result** — the underlying array may change:

```go
x := []int{1, 2, 3}
x = append(x, 4, 5, 6)

// Append a slice to a slice
x = append(x, y...)  // Note the ...
```

### Two-Dimensional Slices

**Independent inner slices** (can grow/shrink independently):

```go
picture := make([][]uint8, YSize)
for i := range picture {
    picture[i] = make([]uint8, XSize)
}
```

**Single allocation** (more efficient for fixed sizes):

```go
picture := make([][]uint8, YSize)
pixels := make([]uint8, XSize*YSize)
for i := range picture {
    picture[i], pixels = pixels[:XSize], pixels[XSize:]
}
```

### Declaring Empty Slices

Prefer nil slices over empty literals:

```go
// Good: nil slice
var t []string

// Avoid: non-nil but zero-length
t := []string{}
```

Both have `len` and `cap` of zero, but the nil slice is the preferred style.

**Exception for JSON**: A nil slice encodes to `null`, while `[]string{}`
encodes to `[]`. Use non-nil when you need a JSON array.

When designing interfaces, avoid distinguishing between nil and non-nil
zero-length slices.

---

## Maps

### Implementing a Set

Use `map[T]struct{}` when the map is only a set. The empty struct takes no
storage and makes membership intent explicit:

```go
attended := map[string]struct{}{"Ann": {}, "Joe": {}}
if _, ok := attended[person]; ok {
    fmt.Println(person, "was at the meeting")
}
```

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
// Bad: copying a mutex
var mu sync.Mutex
mu2 := mu  // almost always a bug

// Good: pass by pointer
func increment(sc *SafeCounter) {
    sc.mu.Lock()
    sc.count++
    sc.mu.Unlock()
}
```

---

## Quick Reference

| Topic | Key Point |
|-------|-----------|
| Slices | Always assign `append` result; `nil` slice preferred over `[]T{}` |
| Sets | `map[T]struct{}` for membership-only sets |
| Copying | `slices.Clone` / `maps.Clone` (nil in, nil out); don't copy `T` if methods are on `*T` |
| Loops | Check `slices`/`maps` before writing one; `go fix -diff ./...` to confirm |

## Related Skills

- **Defensive copying**: See [go-defensive](../go-defensive/SKILL.md) when copying slices or maps at API boundaries to prevent mutation
- **Capacity hints**: See [go-performance](../go-performance/SKILL.md) when pre-sizing slices or maps for known workloads
- **Iteration patterns**: See [go-style-core](../go-style-core/SKILL.md) when using range loops over slices, maps, or channels
- **Declaration style**: See [go-style-core](../go-style-core/SKILL.md) when choosing between `new`, `make`, `var`, and composite literals
