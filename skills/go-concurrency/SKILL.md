---
name: go-concurrency
description: Use when writing concurrent Go code with goroutines, channels, or mutexes; parallelizing work; fixing data races; or protecting shared state. Context cancellation belongs to go-context.
---

# Go Concurrency

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). `sync.WaitGroup.Go`
> and `testing/synctest` require Go 1.25+; typed atomics (`atomic.Bool`) are
> stdlib since Go 1.19.

## Resource Routing

- `references/GOROUTINE-PATTERNS.md` - Read when starting, stopping, or waiting for goroutines.
- `references/SYNC-PRIMITIVES.md` - Read when choosing between mutexes, atomics, channels, and once-like primitives.
- `references/BUFFER-POOLING.md` - Read when considering channel-backed or sync.Pool-style reuse.
- `references/ADVANCED-PATTERNS.md` - Read for bounded errgroup work, cooperative cancellation, request/reply channels, and CPU-bound parallelization.

## Goroutine Lifetimes

> **Normative**: When you spawn goroutines, make it clear when or whether they
> exit.

Goroutines can leak by blocking on channel sends/receives. The GC **will not
terminate** a blocked goroutine even if no other goroutine holds a reference to
the channel. Even non-leaking in-flight goroutines cause panics (send on closed
channel), data races, memory issues, and resource leaks.

### Core Rules

1. **Every goroutine needs a stop mechanism** — a predictable end time, a
   cancellation signal, or both
2. **Code must be able to wait** for the goroutine to finish
3. **No goroutines in `init()`** — expose lifecycle methods (`Close`, `Stop`,
   `Shutdown`) instead
4. **Keep synchronization scoped** — constrain to function scope, factor logic
   into synchronous functions
5. **Bound fan-out** — `errgroup.Group.SetLimit(n)` or a semaphore; never one
   goroutine per element of an unbounded input

```go
// Good: a fixed pair of tasks, both joined before returning
var wg sync.WaitGroup
wg.Go(func() { process(ctx, first) })
wg.Go(func() { process(ctx, second) })
wg.Wait()
```

`process` must honor `ctx` and must not panic; `WaitGroup.Go` does not propagate
errors or cancel tasks. For variable-size input or sibling cancellation on
error, use the bounded `errgroup` pattern in `references/ADVANCED-PATTERNS.md`.

```go
// Bad: no way to stop or wait
go func() { for { flush(); time.Sleep(delay) } }()
```

No `item := item` capture line — loop variables are per-iteration since Go
1.22, and `go fix -forvar ./...` deletes leftovers. `go fix -waitgroupgo ./...`
rewrites `Add(1)`/`go`/`defer Done()` into `wg.Go`.

**Test for leaks** with [go.uber.org/goleak](https://pkg.go.dev/go.uber.org/goleak)
in unit tests; in a running service capture `/debug/pprof/goroutineleak?debug=1`
(Go 1.27+), which lists goroutines the runtime proved can never unblock —
[go-troubleshooting](../go-troubleshooting/SKILL.md) owns reading it. Test
timing-dependent behavior with `testing/synctest` (fake clock, no real sleeps) —
see [go-testing](../go-testing/SKILL.md).

> **Principle**: Never start a goroutine without knowing how it will stop.
> **Validation**: `go test -race ./...` and `go vet ./...` (the `waitgroup`,
> `loopclosure`, and `testinggoroutine` analyzers) — see
> [go-linting](../go-linting/SKILL.md).

---

## Share by Communicating

> "Do not communicate by sharing memory; instead, share memory by communicating."

Pick the primitive by the shape of the problem, not by preference:

```
Do you need a goroutine at all?
├─ No — a return value or a synchronous call does it → no goroutine, no channel
└─ Yes
   ├─ Handing a value or ownership to another goroutine → channel
   ├─ Protecting shared state (cache, counter, map)      → sync.Mutex / RWMutex; atomic.* for one word
   ├─ Fan-out that must be bounded                       → errgroup.Group with SetLimit
   └─ Both fit                                           → the one with fewer lines
```

A goroutine whose caller immediately waits for it is a function call with
extra steps — it is on the hunt list in
[go-code-refactor](../go-code-refactor/references/OVER-ENGINEERING.md).

---

## Synchronous Functions

> **Normative**: Prefer synchronous functions over asynchronous ones.

| Benefit | Why |
|---|---|
| Localized goroutines | Lifetimes easier to reason about |
| Avoids leaks and races | Easier to prevent resource leaks and data races |
| Easier to test | Check input/output without polling |
| Caller flexibility | Caller adds concurrency when needed |

> **Advisory**: It is quite difficult (sometimes impossible) to remove
> unnecessary concurrency at the caller side. Let the caller add concurrency
> when needed.

---

## Zero-value Mutexes

The zero-value of `sync.Mutex` and `sync.RWMutex` is valid — almost never need
a pointer to a mutex.

```go
var mu sync.Mutex // Ready to use; no allocation or constructor needed
```

**Don't embed mutexes** — use a named `mu` field to keep `Lock`/`Unlock` as
implementation details, not exported API.

---

## Channel Direction

> **Normative**: Specify channel direction where possible.

Direction prevents errors (compiler catches closing a receive-only channel),
conveys ownership, and is self-documenting.

```go
func produce(out chan<- int) { /* send-only */ }
func consume(in <-chan int)  { /* receive-only */ }
func transform(in <-chan int, out chan<- int) { /* both */ }
```

### Channel Size: One or None

Channels should have size **zero** (unbuffered) or **one**. Any other size
requires justification for:

- How the size was determined
- What prevents the channel from filling under load
- What happens when writers block

```go
c := make(chan int)    // unbuffered — Good
c := make(chan int, 1) // size one — Good
c := make(chan int, 64) // arbitrary — needs justification
```

---

## Atomic Operations

Use `atomic.Bool`, `atomic.Int64`, etc. from the standard `sync/atomic` package
for type-safe atomic operations. Raw `int32`/`int64` fields make it easy to
forget atomic access on some code paths. Keep `go.uber.org/atomic` when already
used by the project or when a required operation is absent from the standard
library; the types shown here need no external dependency.

```go
var running atomic.Bool
running.Store(true)
running.Load() // Read through the atomic API too
```

A plain read such as `running == 1` on a raw `int32` can race with an atomic
write. See `references/SYNC-PRIMITIVES.md` for the complete comparison.

---

## Documenting Concurrency

> **Advisory**: Document thread-safety when it's not obvious from the operation
> type.

Go users assume read-only operations are safe for concurrent use, and mutating
operations are not. Document concurrency when:

1. **Read vs mutating is unclear** — e.g., a `Lookup` that mutates LRU state
2. **API provides synchronization** — e.g., thread-safe clients
3. **Interface has concurrency requirements** — document in type definition

---

## Context Usage

> For context.Context guidance (parameter placement, struct storage, custom
> types, derivation patterns), see the dedicated
> [go-context](../go-context/SKILL.md) skill.

---

## Buffer Pooling with Channels

Use a buffered channel as a free list to reuse allocated buffers. This "leaky
buffer" pattern uses `select` with `default` for non-blocking operations.
Read `references/BUFFER-POOLING.md` for the full pattern and when to prefer
`sync.Pool` instead.

---

## Related Skills

- [go-resilience](../go-resilience/SKILL.md): bulkhead admission, backpressure, rate scope, recovery budgets.
- [go-context](../go-context/SKILL.md): cancellation, deadlines, request-scoped values through goroutines.
- [go-error-handling](../go-error-handling/SKILL.md): errors from goroutines, errgroup.
- [go-defensive](../go-defensive/SKILL.md): shared state at API boundaries, defer for cleanup.
- [go-interfaces](../go-interfaces/SKILL.md): receiver types for types holding sync primitives.
- [go-troubleshooting](../go-troubleshooting/SKILL.md): goroutine dumps, race reports, the symptom catalog when the cause is unknown.
