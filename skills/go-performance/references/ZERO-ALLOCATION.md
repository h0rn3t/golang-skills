# Zero-Allocation Hot Paths

> Sources: https://medium.com/@ksandeeptech07/zero-allocation-in-go-ce29e6a9ffdc (strategies, corrected below); https://go.dev/doc/gc-guide; https://go.dev/wiki/CompilerOptimizations; https://pkg.go.dev/testing#AllocsPerRun
> Authority: advisory; every allocation count below measured with `testing.AllocsPerRun` on Go 1.27.4
> Last verified: 2026-11-26

## Contents

- [When It Applies](#when-it-applies)
- [The Loop](#the-loop)
- [Reading the Escape Report](#reading-the-escape-report)
- [Why a Value Reaches the Heap](#why-a-value-reaches-the-heap)
- [Caller-Owned Buffers](#caller-owned-buffers)
- [Pooling](#pooling)
- [Traps That Add Allocations](#traps-that-add-allocations)

## When It Applies

Zero allocations is a property of one named path, not a style for a codebase.
Start only when [Read the Profile](../SKILL.md#read-the-profile) gives the
allocation-and-GC row (`runtime.mallocgc` above about 15% cumulative, or GC
assist and mark workers above about 5%), or when a latency target is missed
and the GC trace (`GODEBUG=gctrace=1`) shows collections in the tail. Stop at
the first `keep` that reaches the target: going from 2 allocs/op to 0 is
worth the code only where the verdict shows it.

## The Loop

It is the speed loop in [Benchmarking and Profiling](../SKILL.md#benchmarking-and-profiling)
with the allocation steps spelled out:

1. **Count.** The benchmark from step 1 runs with `b.ReportAllocs()`; the
   number to move is `allocs/op` on the caller's path, not `B/op` of a
   helper.
2. **Find the sites.** `-memprofile .perf-snap/mem.prof`, then
   `go tool pprof -sample_index=alloc_objects -list '<func>' .perf-snap/mem.prof`:
   objects, not bytes, since the per-object cost is what the GC pays for.
3. **Ask the compiler why.** `go build -gcflags='-m' ./path/to/pkg 2>&1 | grep '<file>.go'`
   names each value moved to the heap; `-gcflags='-m=2'` prints the chain
   of assignments that forced it. Read only the lines of the function the
   profile named.
4. **One change**, from [Why a Value Reaches the Heap](#why-a-value-reaches-the-heap),
   then the verdict. A count that moves without `sec/op` moving is reported
   as fewer allocations, not as faster.
5. **Guard it** with the [allocation test](../SKILL.md#allocation-guard), so
   the next edit that adds one fails.

## Reading the Escape Report

| `-m` says | Meaning |
|---|---|
| `moved to heap: x` | the variable `x` itself lives on the heap |
| `<expr> escapes to heap` | the value built by that expression is heap-allocated, or boxed into an interface |
| `does not escape` | the compiler proved it can stay on the stack |
| `leaking param: p` | `p` is stored somewhere that outlives the call; the caller's argument escapes |
| `leaking param: p to result ~r0` | `p` flows into the result; whether it escapes is decided at the caller |
| `inlining call to f` | `f`'s body was copied into the caller, which re-runs escape analysis there |

The report is evidence of why, not a count. Two readings it gives wrong on
1.27 when taken alone:

- **Escape is decided per call site after inlining.** A function returning
  `&d` reports `moved to heap: d`, yet a caller that inlines it and only reads
  a field allocates nothing; a caller that stores the pointer in a global
  allocates once. Read the report at the caller's line, not the callee's.
- **`does not escape` is not zero.** `make([]byte, n)` with a non-constant
  `n` reports `does not escape` and still allocates once `n` exceeds 32 bytes:
  only a small backing array fits the stack slot the compiler reserves.

`testing.AllocsPerRun` is the count; `-m` says which line to change.

## Why a Value Reaches the Heap

| Cause | Typical site | Fix that keeps it on the stack |
|---|---|---|
| Stored in an interface | `fmt` arguments, `any` fields, `slog` attrs, an `int` above 255 converted to `any` | a typed API (`strconv.AppendInt`, `slog.Int`), or formatting outside the hot path |
| Outlives the frame | a pointer returned from a function that is not inlined, stored in a struct field, map, global, or channel | return the value (`T`, not `*T`) for small structs; let the caller own the storage |
| Captured by a closure that outlives the call | `go func()`, a callback kept by the callee | pass the value as an argument |
| Size unknown at compile time | `make([]T, n)` above the stack slot; `append` past `cap` | a capacity known at the call site, or a caller-owned buffer |
| Converted across string and `[]byte` and kept | `[]byte(s)` passed to a writer the compiler cannot see; `string(b)` used as a key that is stored | `io.WriteString`; `m[string(b)]` for a lookup (no allocation), a copy only when stored |
| Called through an interface the compiler cannot devirtualize | `w.Write(p)` on an `io.Writer` parameter leaks `p` | accept; a concrete type helps only where the profile names this call |

## Caller-Owned Buffers

The allocation-free API shape in the standard library is append-style: the
function takes `dst []byte`, appends, and returns the extended slice
(`strconv.AppendInt`, `time.Time.AppendFormat`, `fmt.Appendf`). The caller
reuses one buffer across calls; the function allocates only when the
caller's capacity runs out. The function body, as it appears in an
append-style encoder:

```go
func AppendLine(dst []byte, id int, name string) []byte {
	dst = strconv.AppendInt(dst, int64(id), 10)
	dst = append(dst, ' ')
	dst = append(dst, name...)
	return append(dst, '\n')
}
```

The caller resets the length and keeps the array:
`buf = AppendLine(buf[:0], id, name)` followed by `w.Write(buf)` is 0
allocs/op once `buf` has grown to the line length.

The buffer replaces allocations, not writes. On an `io.Writer` the caller
passed in, the number and order of `Write` calls is observable: a writer that
fails on its second call, a network stream, or a log that must show every
row written before an error. Collecting the header, rows, and trailer into
one buffer and writing it once turns "error at row two" into "no error", and
no test of the function's output notices. Keep one `Write` per unit the code
wrote before and reuse the buffer between them; coalescing writes is a
separate change, made only where the writer is the code's own file or socket
and the profile names the system call.

## Pooling

A pool is for objects that are large or costly to build and short-lived per
request; it is the last allocation fix, after the buffer moved to the caller
or out of the loop. The pattern, the size cap on return, and the copy-out
before `Put` are owned by
[BUFFER-POOLING.md](../../go-concurrency/references/BUFFER-POOLING.md#production-alternative).
Two allocation-specific rules:

- **Pool pointers, never slices.** `pool.Put(buf)` with a `[]byte` boxes the
  slice header into `any`: 1 alloc/op on every `Put`, the cost the pool
  existed to remove (staticcheck SA6002). Store `*bytes.Buffer` or
  `*[]byte` and write back `*bp = buf[:0]` before `Put`.
- **Nothing obtained from the pool is used after `Put`.** A slice returned to
  the caller, a string made with `unsafe.String` over the buffer, or a
  request body still read by the transport hands the next `Get` live data.

## Traps That Add Allocations

Each was measured on 1.27 against the version it was meant to improve:

- **`strings.Builder` followed by `[]byte(b.String())`.** Building
  `"Hello, " + name` with a Builder and writing `[]byte(b.String())` costs
  2 allocs/op; the plain `+` costs 1; two `io.WriteString` calls cost 0. A
  Builder pays off for many pieces in a loop, not for two, and never when
  the result is converted back to bytes.
- **`sync.Pool` of `[]byte`**: see [Pooling](#pooling).
- **A fixed buffer that silently truncates.** `copy(buf, input)` into a
  1024-byte buffer drops everything past byte 1024; a reused buffer grows
  with `append(buf[:0], input...)` and keeps its capacity.
- **`fmt` in a loop declared allocation-free.** `fmt.Printf("%d", n)` boxes
  `n` into `any` and allocates for values above 255.
- **`r.URL.Query()` in a handler.** It parses the whole query into a fresh
  `url.Values` map on every call; the concatenation beside it is not the cost.
  Call it once per request, and only where the profile names it.
- **Preallocating where the size is a guess.** An oversized capacity retains
  memory the workload never uses; the rule is in
  [Prefer Specifying Container Capacity](../SKILL.md#prefer-specifying-container-capacity).
