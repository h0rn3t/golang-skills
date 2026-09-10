---
name: go-performance
description: Use when optimizing slow or performance-critical Go code, allocations, string concatenation in loops, or benchmarks. Concurrent code patterns belong to go-concurrency.
allowed-tools: Bash(bash:*)
---

# Go Performance Patterns

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). `b.Loop()`
> requires Go 1.24+; `encoding/json/v2` Go 1.27+.

## Resource Routing

- `scripts/bench-compare.sh` - Run when comparing benchmark results, saving baselines, or producing JSON benchmark metadata.
- `references/BENCHMARKS.md` - Read when writing benchmarks, using benchstat, or profiling with pprof.
- `references/STRING-OPTIMIZATION.md` - Read when optimizing string conversion, concatenation, byte/string boundaries, or memory retained by substrings.

Apply performance-specific guidance to measured bottlenecks, including retained
memory. Don't add allocations or complexity without evidence that it helps.

---

## Prefer strconv over fmt

For direct primitive conversions, prefer `strconv`; measure its benefit on
the actual inputs and toolchain:

```go
s := strconv.Itoa(n)
```

The [benchmark reference](references/BENCHMARKS.md) shows how to compare
the calls. There is no fixed speedup or allocation count across workloads.

---

## Avoid Repeated String-to-Byte Conversions

For a measured repeated-conversion cost, reuse the bytes outside the loop.
The compiler may already eliminate the conversion allocation for a concrete
writer, so compare the actual call site:

```go
data := []byte("Hello world")
for b.Loop() { // Go 1.24+
    w.Write(data)
}
```

---

## Prefer Specifying Container Capacity

Preallocate when the final size is known at the call site (`len(files)`, a fixed loop bound); otherwise measure before guessing a capacity — an oversized estimate retains memory the workload never uses.

### Map Capacity Hints

Provide capacity hints when initializing maps with `make()`:

```go
m := make(map[string]os.DirEntry, len(files))
```

**Note**: Unlike slices, map capacity hints do not guarantee complete preemptive allocation—they approximate the number of hashmap buckets required.

### Slice Capacity

Provide capacity hints when initializing slices with `make()`, particularly when appending:

```go
data := make([]int, 0, size)
```

Unlike maps, slice capacity is **not a hint**—the compiler allocates exactly that much memory. Subsequent `append()` operations incur zero allocations until capacity is reached.

Preallocation avoids backing-array growth within the chosen capacity. Measure
the resulting time and memory tradeoff; oversized estimates can retain more
memory than the workload needs.

---

## Pass Values

Don't pass pointers as function arguments just to save a few bytes. If a function refers to its argument `x` only as `*x` throughout, then the argument shouldn't be a pointer.

```go
func process(s string) { // not *string — strings are small fixed-size headers
    fmt.Println(s)
}
```

**Common pass-by-value types**: `string`, `io.Reader`, small structs.

**Exceptions**:
- Large structs where copying is expensive
- Small structs that might grow in the future

---

## String Concatenation

Choose the right strategy based on complexity:

| Method | Best For |
|--------|----------|
| `+` | Few strings, simple concat |
| `fmt.Sprintf` | Formatted output with mixed types |
| `strings.Builder` | Loop/piecemeal construction |
| `strings.Join` | Joining a slice |
| Backtick literal | Constant multi-line text |

---

## Benchmarking and Profiling

Always measure before and after optimizing. Use Go's built-in benchmark framework and profiling tools.

```bash
go test -bench=. -benchmem -count=10 ./...
```

> **Validation**: Run `bash scripts/bench-compare.sh` to measure the actual
> impact, and **revert any optimization without a measurable win** — an
> unmeasured optimization is a readability cost with no benefit. Report the
> before/after numbers; do not describe a change as "faster" without them.

### Benchmark discipline

- Keep benchmarks in a file of their own beside the source, ordered to mirror
  the functions they measure. Go only requires some `_test.go` file; the
  `*_bench_test.go` suffix is this pack's convention, and an existing project
  layout outranks it.
- Implement competing variants in isolation, then measure serially on the same
  machine and toolchain; concurrent runs share CPUs and contaminate `ns/op`.
- Compare with `benchstat` (see
  [BENCHMARKS.md](references/BENCHMARKS.md)) and claim only deltas it calls
  significant — never a single run, never a `~` row.
- A comparison that straddles a toolchain bump measures the toolchain, not the
  code — re-run the baseline on the new toolchain first.
- Perf-only changes use a `perf(scope):` subject and paste the benchstat table
  plus hardware context (`goos`/`goarch`/`cpu`) in the body.

### Before reaching for a faster library

Check the standard library first — `encoding/json/v2` (Go 1.27+) and the
iterator variants (`strings.SplitSeq` Go 1.24+, `maps.Keys` Go 1.23+) remove
allocations without a new dependency. See
[go-packages](../go-packages/SKILL.md) for the dependency ladder.

---

## Related Skills

- **Data structures**: See [go-data-structures](../go-data-structures/SKILL.md) when choosing between slices, maps, and arrays, or understanding allocation semantics
- **Declaration patterns**: See [go-style-core](../go-style-core/SKILL.md) when using `make` with capacity hints or initializing maps and slices
- **Concurrency**: See [go-concurrency](../go-concurrency/SKILL.md) when parallelizing work across goroutines or using sync.Pool for buffer reuse
- **Style principles**: See [go-style-core](../go-style-core/SKILL.md) when deciding whether an optimization is worth the readability cost
- **Unknown cause**: See [go-troubleshooting](../go-troubleshooting/SKILL.md) when the program is slow, leaking, or hanging and nobody yet knows why — profile capture and the symptom catalog live there; this skill starts once the hot path is named
