---
name: go-performance
description: Use when optimizing slow or performance-critical Go code, allocations, string concatenation in loops, caches, or benchmarks. Concurrent code patterns belong to go-concurrency.
allowed-tools: Bash(bash:*)
---

# Go Performance Patterns

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). `b.Loop()`
> requires Go 1.24+; `encoding/json/v2` Go 1.27+.

## Resource Routing

- `scripts/bench-compare.sh` - Run when comparing benchmark results, saving baselines, or producing JSON benchmark metadata.
- `references/BENCHMARKS.md` - Read when writing benchmarks, using benchstat, or profiling with pprof.
- `references/STRING-OPTIMIZATION.md` - Read when optimizing concatenation or memory retained by substrings.

Apply performance-specific guidance to measured bottlenecks, including retained
memory. Don't add allocations or complexity without evidence that it helps.

---

## Prefer strconv over fmt

For direct primitive conversions, prefer `strconv` by default; choosing it
needs no benchmark:

```go
s := strconv.Itoa(n)
```

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

---

## Pass Values

A pointer parameter is not a speed fix: value versus pointer parameters belong to [go-functions](../go-functions/SKILL.md#pointers-to-interfaces).

---

## String Concatenation

Build in a loop with `strings.Builder`, calling `Grow(n)` first when the final size is known.

---

## Caching

> **Normative**: Cache what a profile shows is expensive and read far more
> often than it changes. A cached fast call buys nothing and adds a staleness
> bug.

- **The key is every input the answer depends on**: the tenant, caller,
  locale, or permissions it was computed for, including those that arrive in
  `ctx`. A key of the user ID alone serves one tenant's record to another.
- **One expiry rule, stated.** The TTL is the staleness the caller accepts. A
  balance, a permission, or stock at checkout is not cached. A failed load is
  returned and not kept.
- **Concurrent misses share one load**, or a cold hot key sends every waiting
  request to the origin at once. Use `golang.org/x/sync/singleflight` where the
  module already requires `golang.org/x/sync`, otherwise the in-flight entry
  below. The lock guards the map and is never held across the load, so one
  slow key does not stall the rest. The load runs under the first caller's
  `ctx`.
- **Bounded**: cap the entries or replace expired ones as they are read; a map
  that only grows is a leak.

```go
type rateKey struct{ tenant, pair string } // every input the rate depends on
type entry struct {
	done chan struct{} // closed once rate and err are set
	rate float64
	err  error
	at   time.Time // zero while loading
}
type Rates struct {
	load    func(ctx context.Context, key rateKey) (float64, error)
	ttl     time.Duration
	mu      sync.Mutex // guards entries; never held across load
	entries map[rateKey]*entry
}

func (r *Rates) Get(ctx context.Context, key rateKey) (float64, error) {
	r.mu.Lock()
	e, ok := r.entries[key]
	if ok && (e.at.IsZero() || time.Since(e.at) < r.ttl) {
		r.mu.Unlock()
		select {
		case <-e.done:
			return e.rate, e.err
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}
	e = &entry{done: make(chan struct{})}
	r.entries[key] = e
	r.mu.Unlock()
	e.rate, e.err = r.load(ctx, key)
	r.mu.Lock()
	if e.err != nil {
		delete(r.entries, key) // a failed load is not kept
	} else {
		e.at = time.Now()
	}
	r.mu.Unlock()
	close(e.done)
	return e.rate, e.err
}
```

---

## Benchmarking and Profiling

A change made for speed beyond the defaults above (`strconv` for primitive
conversions, capacity when the final size is known, `strings.Builder` in a
loop) needs a baseline saved before the edit and a comparison after it, run
serially on the same machine and toolchain — concurrent runs share CPUs and
contaminate `ns/op`. Run from the project:

```bash
bash "<installed-skill-dir>/scripts/bench-compare.sh" --save "${TMPDIR:-/tmp}/before.txt" ./path/to/pkg      # before the edit
bash "<installed-skill-dir>/scripts/bench-compare.sh" --baseline "${TMPDIR:-/tmp}/before.txt" ./path/to/pkg  # after it
```

> **Validation**: keep the change only for a delta `benchstat` calls
> significant — never a single run, never a `~` row — and revert it
> otherwise. Report the before/after table, or report the comparison as
> skipped; do not describe a change as "faster" without it.
> [BENCHMARKS.md](references/BENCHMARKS.md) reads benchstat output and profiles.

### Benchmark discipline

- Keep benchmarks in a file of their own beside the source, ordered to mirror
  the functions they measure. Go only requires some `_test.go` file; the
  `*_bench_test.go` suffix is this pack's convention, and an existing project
  layout outranks it.
- Implement competing variants in isolation, so each run measures one of them.
- A comparison that straddles a toolchain bump measures the toolchain, not the
  code — re-run the baseline on the new toolchain first.
- Perf-only changes use a `perf(scope):` subject and paste the benchstat table
  plus hardware context (`goos`/`goarch`/`cpu`) in the body.
- The report or PR description lists the attempts that were reverted, each
  with its benchstat row, so the next change does not try a dead idea again.

### Before reaching for a faster library

Check the standard library first — `encoding/json/v2` (Go 1.27+) and the
iterator variants (`strings.SplitSeq` Go 1.24+, `maps.Keys` Go 1.23+) remove
allocations without a new dependency. See
[go-packages](../go-packages/SKILL.md) for the dependency ladder.

---

## Related Skills

- [go-data-structures](../go-data-structures/SKILL.md): slices, maps, arrays and their allocation semantics.
- [go-concurrency](../go-concurrency/SKILL.md): parallel work, `sync.Pool` for buffers.
- [go-style-core](../go-style-core/SKILL.md): whether an optimization is worth its readability cost.
- [go-troubleshooting](../go-troubleshooting/SKILL.md): a slow, leaking, or hanging program with no named hot path yet; profile capture and the symptom catalog live there.
