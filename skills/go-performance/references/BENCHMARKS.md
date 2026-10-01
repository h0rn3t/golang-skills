# Benchmark Methodology

> Sources: https://pkg.go.dev/testing#hdr-Benchmarks; https://pkg.go.dev/golang.org/x/perf/cmd/benchstat
> Authority: advisory; `testing.B` semantics normative
> Minimum Go: `b.Loop` 1.24
> Last verified: 2026-10-01

## Contents

- [Writing Benchmarks](#writing-benchmarks)
- [Using benchstat for Comparison](#using-benchstat-for-comparison)
- [Profiles from Benchmarks](#profiles-from-benchmarks)
- [Common Mistakes](#common-mistakes)

## Writing Benchmarks

Go benchmarks use the `testing.B` type and live in `_test.go` files. Function
names must start with `Benchmark`. Use `b.Loop()` (Go 1.24+): unlike `b.N`
loops, it keeps the loop body from being optimized away.

```go
func BenchmarkStrconv(b *testing.B) {
    n := 123456789
    for b.Loop() {
        strconv.Itoa(n)
    }
}

func BenchmarkFmtSprint(b *testing.B) {
    n := 123456789
    for b.Loop() {
        _ = fmt.Sprint(n) // vet's unusedresult reports a bare fmt.Sprint
    }
}
```

Key rules:
- Use `b.Loop()`. You will still meet `for i := 0; i < b.N; i++` in existing
  benchmarks — read it, but do not write it in new code
- Inside `for b.Loop()` results stay live without a sink. A `b.N` loop or
  `RunParallel` needs a package-level sink (`var sink T`), not `_ = f(x)`,
  which lets the compiler delete an inlined call (see Compiler Elision below)
- Live is not heap-allocated: a result that never leaves the loop can stay on
  the stack, so store it where the real caller would (`sink = data`), and when
  B/op is 0 check `go test -gcflags=-m` for `does not escape`
- `b.Loop()` resets the timer on its first call, so setup before the loop is
  excluded automatically; `b.ResetTimer()` is for `b.N` loops
- `b.N` is the iteration count only after `b.Loop()` has returned false, so a
  benchmark that sizes a fixture by `b.N` is restructured — a constant size, or
  a fixture built per iteration — not translated line by line
- Use `b.ReportAllocs()` or the `-benchmem` flag for allocation tracking

### Sub-benchmarks

```go
var sinkStr string

func BenchmarkConvert(b *testing.B) {
    for _, size := range []int{10, 100, 1000} {
        b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
            data := make([]byte, size)
            for b.Loop() {
                sinkStr = string(data) // _ = string(data) is 0 B/op at size=10
            }
        })
    }
}
```

---

## Using benchstat for Comparison

```bash
go get -tool golang.org/x/perf/cmd/benchstat@latest   # pinned in go.mod, per go-packages
go tool benchstat old.txt new.txt
```

`bench-compare.sh` runs `go tool benchstat` when `go.mod` declares the tool,
otherwise a `benchstat` on `PATH`; with neither it prints raw lines and says
the comparison was skipped.

### Interpreting benchstat Output

One table per unit (`sec/op`, then `B/op` and `allocs/op` with `-benchmem`);
the `sec/op` table from a go1.27.1 run where `Format` switched from
`fmt.Sprint` to `strconv.Itoa` and `Parse` did not change:

```
         │   old.txt    │               new.txt               │
         │    sec/op    │   sec/op     vs base                │
Format-8   106.85n ± 3%   44.18n ± 5%  -58.66% (p=0.000 n=10)
Parse-8     12.79n ± 6%   12.71n ± 4%        ~ (p=0.383 n=10)
geomean     36.97n        23.70n       -35.90%
```

- **Value ± %**: the median and its 95% confidence interval
- **vs base**: change of the median from the first file (negative = faster),
  with the p-value and the samples per file; `~` means no significant
  difference at p < 0.05, and the row gives no percentage
- **geomean**: geometric mean of the column across benchmarks

Tips:
- Always use `-count=10` or higher for reliable results
- A small p-value is evidence under the sampling assumptions; it does not
  exclude drift, biased inputs, or interference from other workloads

---

## Profiles from Benchmarks

A benchmark says how much; a profile says where. Capture both from the same
run, then read the profile with `go tool pprof`:

```bash
go test -run '^$' -bench=BenchmarkHotPath -cpuprofile=cpu.prof -memprofile=mem.prof ./path/to/pkg
go tool pprof -top cpu.prof            # or: -alloc_space mem.prof
```

Profiling a running service, `pprof` endpoints, execution traces, and reading
the output belong to
[go-troubleshooting](../../go-troubleshooting/references/DIAGNOSTIC-TOOLS.md).
Re-benchmark after each change and compare with `benchstat`; then re-profile
for the next bottleneck.

---

## Common Mistakes

### Compiler Elision with `b.Loop`

In the exact `for b.Loop() { ... }` form (Go 1.24+), the compiler keeps
arguments and results of calls inside the loop alive.

That guarantee does not cover manual `b.N` loops or `RunParallel`. In a `b.N`
loop assign the result to a package-level sink (`sink = expensiveFunc()`); in
`RunParallel` keep the sink per goroutine and store it once after the loop,
because goroutines writing one plain variable race:

```go
var sinkParallel atomic.Int64

func BenchmarkWorkParallel(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        var last int64
        for pb.Next() {
            last = expensiveFunc()
        }
        sinkParallel.Store(last)
    })
}
```

See [B.Loop](https://pkg.go.dev/testing#B.Loop).
