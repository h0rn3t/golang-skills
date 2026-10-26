---
name: go-performance
description: Use when optimizing slow or performance-critical Go code, allocations, string concatenation in loops, or benchmarks. Also use when writing or reviewing a cache, including implementing one from a stub — its key, its expiry, and how a miss loads. Concurrent code patterns belong to go-concurrency.
allowed-tools: Bash(bash:*)
---

# Go Performance Patterns

> Compatibility: Baseline Go 1.27 (see the plugin repository's COMPATIBILITY.md). `b.Loop()`
> requires Go 1.24+; `encoding/json/v2` Go 1.27+; cgroup-aware `GOMAXPROCS`
> Go 1.25+.

## Resource Routing

- `scripts/bench-compare.sh` - Run when comparing benchmark results, saving baselines, or producing JSON benchmark metadata.
- `scripts/bench-verdict.go` - The statistics helper `bench-compare.sh --verdict` builds; `scripts/bench-verdict_test.go` holds its unit tests.
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
  A loader that dedups within one request (a GraphQL DataLoader) is built per
  request; shared by the process, it is a cache with no tenant in its key.
- **One expiry rule, stated.** The TTL is the staleness the caller accepts. A
  balance, a permission, or stock at checkout is not cached. A failed load is
  returned and not kept. A write deletes the key after its transaction
  commits and never stores the value it was handed: two writers that commit
  in one order and set the cache in the other keep the older value until the
  TTL. Only a load fills.
- **Concurrent misses share one load**, or a cold hot key sends every waiting
  request to the origin at once. Use `golang.org/x/sync/singleflight` where the
  module already requires `golang.org/x/sync`, otherwise the in-flight entry
  below. The lock guards the map and is never held across the load, so one
  slow key does not stall the rest. The load takes the first caller's `ctx`
  values but not its cancellation (`context.WithoutCancel`), so a caller that
  leaves fails no other waiter, and `load` sets its own deadline; the entry is
  finished in a `defer`, so a panicking `load` does not leave its key loading
  for good.
- **Bounded**: a cap on the entry count, or a sweep of expired entries; a key
  space the code does not bound (user IDs, query strings) needs one, since
  replacing an expired entry on its next read never frees a key that is not
  read again. A map that only grows is a leak, and a swept one keeps its
  peak table: [go-data-structures](../go-data-structures/SKILL.md#a-map-never-shrinks).
  Interning repeated values is `unique.Make` (Go 1.23); a map of
  `weak.Pointer` (Go 1.24) is not a cache, since its entries die at the next
  GC and its keys stay unless `runtime.AddCleanup` deletes them.

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
	mu      sync.Mutex         // guards entries; never held across load
	entries map[rateKey]*entry // bounded: tenants times currency pairs
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
	e = &entry{done: make(chan struct{}), err: errors.New("rates: load panicked")}
	r.entries[key] = e
	r.mu.Unlock()
	defer func() { // also after a panic in load: waiters wake, the key is retried
		r.mu.Lock()
		e.at = time.Now()
		if e.err != nil {
			delete(r.entries, key) // a failed load is not kept
		}
		r.mu.Unlock()
		close(e.done)
	}()
	e.rate, e.err = r.load(context.WithoutCancel(ctx), key) // its values, not its cancel
	return e.rate, e.err
}
```

---

## Benchmarking and Profiling

A request to make Go code faster runs these steps in order, from the project.
The defaults above (`strconv` for a lone value, a capacity known at the call
site, `strings.Builder` in a loop) need none of them.

1. **A benchmark of the path the caller runs**, before any production edit:
   an existing `Benchmark*` that covers it, or a new one in
   `*_bench_test.go`. Its inputs have production's size and variety — a cache
   benchmark with one key measures only hits — and it uses `b.RunParallel`
   when requests run concurrently. The benchmark is a sample of the workload,
   not the specification: nothing the change adds is keyed to its inputs or
   kept between its iterations.
2. **Freeze the baseline** after step 1 and before the first production edit,
   one package at a time; the snapshot keeps the test binary and the test
   files as they are now. `--filter` names the step 1 benchmark:

   ```bash
   bash "<installed-skill-dir>/scripts/bench-compare.sh" --snapshot "${TMPDIR:-/tmp}/perf-snap" --filter '^BenchmarkX$' ./path/to/pkg
   ```

3. **Profile the benchmark and name where the time is**, as
   [Read the Profile](#read-the-profile) says. The function a request names
   ("X looks slow", "speed up X") is a guess this step checks, not the
   target: under 5% of the profile it stays as it is however wasteful it
   reads — making it free saves less than 5% — and the report gives its share
   and where the time went instead.
4. **One change, from the cheapest class** in [Fix Order](#fix-order).
5. **The verdict decides**, run serially, never beside another benchmark:

   ```bash
   bash "<installed-skill-dir>/scripts/bench-compare.sh" --verdict "${TMPDIR:-/tmp}/perf-snap" --filter '^BenchmarkX$' ./path/to/pkg
   ```

   With the same `--filter`, the verdict judges the path the request is
   about: every other benchmark of the package costs about 25 s at the
   defaults and raises the bar each row must clear. `keep` (exit 0) keeps the
   change; `revert` reverts it; `noisy` raises
   `--count` or `--benchtime` and asks once more, and a change is never kept
   on `noisy`. A second `noisy` is reported as no measurable change: rerunning
   until a `keep` appears picks the noise, not the change. A `keep` on `allocs/op` or `B/op` alone is a speedup only where the
   profile spent its time in allocation or GC; otherwise it is reported as
   fewer allocations. After a kept change, profile again and name the next
   hotspot; chase it only within the request, and snapshot the kept code into
   a new directory before that next change: against the first snapshot, a
   second change that is slower still reads `keep` on the first one's gain.
   A last verdict against the first snapshot gives the total.
6. **Guard an allocation win** with an [allocation test](#allocation-guard)
   when the change exists to remove allocations.
7. **Report** each kept change with its verdict line and the production lines
   it added, and each reverted attempt with its reason, so the next change
   does not try a dead idea again. "Faster" needs a `keep` on `sec/op`;
   without a verdict the comparison is reported as skipped.

A behavior-preserving refactor asks a different question — did anything get
slower — and keeps `--save`/`--baseline` with `benchstat`
([go-code-refactor](../go-code-refactor/SKILL.md)).
[BENCHMARKS.md](references/BENCHMARKS.md) reads benchstat output and profiles.

### Read the Profile

`go test -run '^$' -bench <name> -o "${TMPDIR:-/tmp}/pkg.test" -cpuprofile "${TMPDIR:-/tmp}/cpu.prof" ./path/to/pkg`
(`-o` and absolute paths keep the test binary and the profile out of the
tree), then `go tool pprof -top -cum "${TMPDIR:-/tmp}/cpu.prof"`; the row
decides the class of fix:

| The profile shows | The time goes to | First move |
|---|---|---|
| a package function and what it calls: `slices.Index`, `strings.Fields`, `regexp.Compile`, a parse per call | repeated or superlinear work | do it once where its input changes; a set in place of a scan |
| `runtime.mallocgc` above about 15% cumulative, or `gcBgMarkWorker` and `gcAssistAlloc` above about 5% | allocation and GC | `-memprofile`, then `pprof -sample_index=alloc_space`; fewer values on the heap per call |
| `runtime.scanobject` or `gcDrain` high while allocs/op is low | marking a large, pointer-dense live heap | values instead of `*T` in long-lived slices and maps, indexes instead of pointer links |
| `sync.(*Mutex).Unlock` at the top of a `-mutexprofile`, or `sync.(*Mutex).Lock` in a `-blockprofile`, of a `RunParallel` benchmark | contention | a shorter critical section; per-key state behind a read-mostly map ([go-concurrency](../go-concurrency/SKILL.md)) |
| total samples far below the profile's duration | waiting on I/O, a round trip per item, or a lock | an existing batch API; [go-database](../go-database/SKILL.md) for a query per row |
| a syscall (`syscall.Write`, `Syscall6`) at the top | one system call per small read or write | count calls per op: a `bufio.Writer`/`bufio.Reader` or one batched call is class 1; decline only when each unit of work already makes one call |
| `crypto/*` or `compress/*` doing the payload's own work, set up once | work the platform already does at speed | nothing in this code: report the share and decline |

### Fix Order

Take the first class that reaches the speedup asked for, and stop:

1. Do less: work repeated per call moved to where its input changes, a
   quadratic scan replaced by a set, one round trip per batch.
2. Allocate less with what ships: a struct in place of `map[string]any`, one
   reused `Encoder` or buffer, `strconv.Append*`, a capacity known at the
   call site.
3. Contend less: a shorter critical section, atomics for independent counters.
4. More CPUs per request, only where CPUs sit idle under the real load — a
   batch job, not a server where every CPU already serves a request: there a
   fan-out is faster in a one-call benchmark and no faster in `RunParallel`.
5. Hand-written replacements for the standard library (an encoder, an
   escaper, a hash table), shards, cache-line padding, `unsafe`: each takes its
   own verdict against the best change from the classes above, and the report
   names the lines it added.

### Allocation Guard

A kept allocation win stays won through a test of the count, not of time.
`AllocsPerRun` counts every goroutine's allocations, panics under parallel
tests, and the race detector drops pooled objects, so the test runs alone and
outside `-race`:

```go
//go:build !race

package ledger

import (
	"io"
	"testing"
)

func TestWriteReportAllocs(t *testing.T) {
	rows := sampleRows(t, 100)
	got := testing.AllocsPerRun(20, func() {
		if err := WriteReport(io.Discard, rows); err != nil {
			t.Fatal(err)
		}
	})
	if got > 1 {
		t.Errorf("WriteReport(100 rows) = %v allocs/op, want <= 1: one buffer, not one per row", got)
	}
}
```

### Benchmark discipline

- Keep benchmarks in a file of their own beside the source, ordered to mirror
  the functions they measure. Go only requires some `_test.go` file; the
  `*_bench_test.go` suffix is this pack's convention, and an existing project
  layout outranks it.
- Implement competing variants in isolation, so each run measures one of them.
- A comparison that straddles a toolchain bump measures the toolchain, not the
  code — re-run the baseline on the new toolchain first.
- Perf-only changes use a `perf(scope):` subject and paste the verdict or
  benchstat table plus hardware context (`goos`/`goarch`/`cpu`) in the body.

### Before reaching for a faster library

Check the standard library first — the iterator variants (`strings.SplitSeq`
Go 1.24+, `maps.Keys` Go 1.23+) skip the intermediate slice without a new
dependency. JSON is not such a lever on Go 1.27: `encoding/json` already runs
on the v2 implementation (unmarshal significantly faster, marshal at parity),
so moving an import to `encoding/json/v2` measures nothing, and a comparison
across the 1.26 → 1.27 bump measures the new engine. The JSON wins are in
the call site: one reused `Encoder` or buffer, a struct in place of
`map[string]any`. See [go-packages](../go-packages/SKILL.md) for the
dependency ladder.

---

## Runtime Settings

`GOMAXPROCS` follows the cgroup CPU limit from a `go 1.25` directive, so
`go.uber.org/automaxprocs` is a dependency to remove, as its own change
([MODERNIZATION.md](../go-code-refactor/references/MODERNIZATION.md)). The
removal changes the value under a small or fractional limit: the runtime
rounds the limit up and stays at 2 or more on a machine with two CPUs or
more, where `automaxprocs` rounded down to at least 1 — a 1.5-CPU container
goes from 1 to 2. The runtime reads the CPU limit, not the request, and a
`GOMAXPROCS` variable or call turns its automatic updates off
(`go doc runtime.GOMAXPROCS`).

The garbage collector's settings belong to the process, not to a package:
library code never calls `debug.SetGCPercent`, `debug.SetMemoryLimit`,
`runtime.GOMAXPROCS`, or `runtime.GC`. A `GOGC` value is a deployment
recommendation, measured against the unchanged snapshot with
`--verdict DIR --new-env GOGC=200`. `GOMEMLIMIT` goes where the process owns
its container's memory, 5–10% under the limit, and never to avoid an
out-of-memory kill in a process already close to that limit: it trades the
kill for a severe slowdown, where a higher limit or a lower `GOGC` does not
([gc-guide](https://go.dev/doc/gc-guide)). Tuning chosen before Go 1.26 was
measured against the previous collector; re-measure it.

Profile-guided optimization is a build input, not a code change: a merged
CPU profile from production saved as `default.pgo` in the main package's
directory is used by `go build` (`-pgo=auto`, the default since Go 1.21),
typically for 2–14%. A microbenchmark's profile is a poor source: it
exercises one small part of the program
([go.dev/doc/pgo](https://go.dev/doc/pgo)).

---

## Related Skills

- [go-data-structures](../go-data-structures/SKILL.md): slices, maps, arrays and their allocation semantics.
- [go-concurrency](../go-concurrency/SKILL.md): parallel work, `sync.Pool` for buffers.
- [go-database](../go-database/SKILL.md): a query per row, pools, and batching when the time is a round trip.
- [go-http](../go-http/SKILL.md): client reuse and timeouts when the time is waiting on a service.
- [go-style-core](../go-style-core/SKILL.md): whether an optimization is worth its readability cost.
- [go-troubleshooting](../go-troubleshooting/SKILL.md): a slow, leaking, or hanging program with no named hot path yet; profile capture and the symptom catalog live there.
