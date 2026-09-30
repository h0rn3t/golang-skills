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

The GC **will not terminate** a blocked goroutine even if no other goroutine
holds a reference to the channel.

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

**Test for leaks** with `synctest.Test` (Go 1.25+), which fails when a goroutine
started in its bubble is still blocked once the test function returns, or write
`pprof.Lookup("goroutineleak").WriteTo(w, 1)` (Go 1.27+) in the tested process
and read the stacks. Keep [go.uber.org/goleak](https://pkg.go.dev/go.uber.org/goleak)
where the project already uses it; do not add the module for this.

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

## Synchronous Functions

> **Normative**: Prefer synchronous functions over asynchronous ones.

## Mutexes

**Don't embed mutexes** — use a named `mu` field to keep `Lock`/`Unlock` as
implementation details, not exported API.

## Channel Direction

> **Normative**: Specify channel direction where possible.

### Channel Size: One or None

Channels should have size **zero** (unbuffered) or **one**. Any other size
requires justification for:

- How the size was determined
- What prevents the channel from filling under load
- What happens when writers block

## Atomic Operations

Use `atomic.Bool`, `atomic.Int64`, etc. from the standard `sync/atomic` package
for type-safe atomic operations. Keep `go.uber.org/atomic` when already
used by the project or when a required operation is absent from the standard
library.

## Related Skills

- [go-resilience](../go-resilience/SKILL.md): bulkhead admission, backpressure, rate scope, recovery budgets.
- [go-context](../go-context/SKILL.md): cancellation, deadlines, request-scoped values through goroutines.
- [go-error-handling](../go-error-handling/SKILL.md): errors from goroutines, errgroup.
- [go-defensive](../go-defensive/SKILL.md): shared state at API boundaries, defer for cleanup.
- [go-interfaces](../go-interfaces/SKILL.md): receiver types for types holding sync primitives.
- [go-troubleshooting](../go-troubleshooting/SKILL.md): goroutine dumps, race reports, the symptom catalog when the cause is unknown.
