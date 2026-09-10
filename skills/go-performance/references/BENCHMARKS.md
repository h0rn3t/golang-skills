# Benchmark Methodology

> Sources: https://pkg.go.dev/testing#hdr-Benchmarks; https://pkg.go.dev/golang.org/x/perf/cmd/benchstat
> Authority: advisory; `testing.B` semantics normative
> Minimum Go: `b.Loop` 1.24
> Last verified: 2026-09-10

## Contents

- [Writing Benchmarks](#writing-benchmarks)
- [Running Benchmarks](#running-benchmarks)
- [Interpreting Results](#interpreting-results)
- [Using benchstat for Comparison](#using-benchstat-for-comparison)
- [Benchmark Examples from Performance Patterns](#benchmark-examples-from-performance-patterns)
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
        s := strconv.Itoa(n)
        _ = s
    }
}

func BenchmarkFmtSprint(b *testing.B) {
    n := 123456789
    for b.Loop() {
        s := fmt.Sprint(n)
        _ = s
    }
}
```

Key rules:
- Use `b.Loop()`. You will still meet `for i := 0; i < b.N; i++` in existing
  benchmarks — read it, but do not write it in new code
- Inside `for b.Loop()` results stay live without a sink; keep `_ = x` only
  in `b.N` loops and `RunParallel` (see Compiler Elision below)
- `b.Loop()` resets the timer on its first call, so setup before the loop is
  excluded automatically; `b.ResetTimer()` is for `b.N` loops
- Use `b.ReportAllocs()` or the `-benchmem` flag for allocation tracking

### Sub-benchmarks

```go
func BenchmarkConvert(b *testing.B) {
    for _, size := range []int{10, 100, 1000} {
        b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
            data := make([]byte, size)
            for b.Loop() {
                _ = string(data)
            }
        })
    }
}
```

---

## Running Benchmarks

```bash
# Run all benchmarks in a package
go test -bench=. ./...

# Run specific benchmark with memory stats
go test -bench=BenchmarkStrconv -benchmem ./...

# Run with count for statistical significance
go test -bench=. -benchmem -count=10 ./...
```

The `-benchmem` flag reports allocations per operation. The `-count` flag runs
each benchmark N times for statistical significance.

---

## Interpreting Results

These numbers illustrate the output format; they are not Go 1.27 measurements.

```
BenchmarkStrconv-8     18705042    64.2 ns/op    16 B/op    1 allocs/op
BenchmarkFmtSprint-8    8249536   143.0 ns/op    16 B/op    2 allocs/op
```

| Field | Meaning |
|-------|---------|
| `-8` | GOMAXPROCS |
| `18705042` | Number of iterations |
| `64.2 ns/op` | Time per operation |
| `16 B/op` | Bytes allocated per operation |
| `1 allocs/op` | Heap allocations per operation |

---

## Using benchstat for Comparison

`benchstat` compares benchmark results statistically. Install it and save
benchmark output to files:

```bash
# Install benchstat
go install golang.org/x/perf/cmd/benchstat@latest

# Run benchmarks and save results
go test -bench=. -benchmem -count=10 ./... > old.txt

# Make changes, then run again
go test -bench=. -benchmem -count=10 ./... > new.txt

# Compare results
benchstat old.txt new.txt
```

### Interpreting benchstat Output

```
name          old time/op    new time/op    delta
Strconv-8     64.2ns ± 2%    61.8ns ± 1%   -3.74%  (p=0.001 n=10+10)
```

- **delta**: Percentage change (negative = faster)
- **p-value**: Statistical significance (p < 0.05 is significant)
- **n**: Number of valid samples used

Tips:
- Always use `-count=10` or higher for reliable results
- A small p-value is evidence under the sampling assumptions; it does not
  exclude drift, biased inputs, or interference from other workloads
- If benchstat shows `~` (tilde), the difference is not statistically
  significant

---

## Benchmark Examples from Performance Patterns

### strconv vs fmt

Run the two functions above against representative identical inputs. Retain
the raw samples, toolchain, CPU, and benchmark source with any published numbers.

### Repeated Byte Conversions

```go
func BenchmarkRepeatedConversion(b *testing.B) {
    var buf bytes.Buffer
    buf.Grow(len("Hello world"))
    for b.Loop() {
        buf.Reset()
        buf.Write([]byte("Hello world"))
    }
}

func BenchmarkSingleConversion(b *testing.B) {
    var buf bytes.Buffer
    data := []byte("Hello world")
    buf.Grow(len(data))
    for b.Loop() {
        buf.Reset()
        buf.Write(data)
    }
}
```

Both variants reset the buffer on every iteration and preallocate the same
capacity, so retained data does not grow with the iteration count. A concrete
`bytes.Buffer` may let the compiler avoid the conversion allocation; equal
results are valid. Measure the actual writer type before generalizing.

### Slice Capacity

```go
func BenchmarkNoCapacity(b *testing.B) {
    for b.Loop() {
        data := make([]int, 0)
        for k := 0; k < 1000; k++ {
            data = append(data, k)
        }
    }
}

func BenchmarkWithCapacity(b *testing.B) {
    for b.Loop() {
        data := make([]int, 0, 1000)
        for k := 0; k < 1000; k++ {
            data = append(data, k)
        }
    }
}
```

Measure both allocation count and elapsed time; capacity is a workload choice,
not a universal speedup factor.

---

## Profiles from Benchmarks

A benchmark says how much; a profile says where. Capture both from the same
run, then read the profile with `go tool pprof`:

```bash
go test -bench=BenchmarkHotPath -cpuprofile=cpu.prof -memprofile=mem.prof ./...
go tool pprof -top cpu.prof            # or: -alloc_space mem.prof
```

Profiling a running service, `pprof` endpoints, execution traces, and reading
the output belong to
[go-troubleshooting](../../go-troubleshooting/references/DIAGNOSTIC-TOOLS.md).
Re-benchmark after each change and compare with `benchstat`; then re-profile
for the next bottleneck.

---

## Common Mistakes

### Ignoring the benchmark loop

The testing framework adjusts iteration counts to get stable timing. Using a
fixed iteration count produces meaningless results:

```go
// Bad: Ignores the benchmark loop, so the framework can't calibrate
func BenchmarkFixed(b *testing.B) {
    for i := 0; i < 1000; i++ {
        doWork()
    }
}

// Good: Use b.Loop on Go 1.24+
func BenchmarkCorrect(b *testing.B) {
    for b.Loop() {
        doWork()
    }
}
```

### Compiler Elision with `b.Loop`

In the exact `for b.Loop() { ... }` form (Go 1.24+), the compiler keeps
arguments and results of calls inside the loop alive. The call below does not
need a package-level sink:

```go
func BenchmarkWork(b *testing.B) {
    for b.Loop() {
        expensiveFunc()
    }
}
```

That guarantee does not cover manual `b.N` loops or `RunParallel`; ensure
results stay observable there. See [B.Loop](https://pkg.go.dev/testing#B.Loop).
