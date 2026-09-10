# Advanced Concurrency Patterns

> Sources: https://pkg.go.dev/golang.org/x/sync/errgroup; source/effective-go/effective_go.html (Concurrency)
> Authority: advisory
> Last verified: 2026-09-10

Situational patterns for bounded work, request/response multiplexing, and
CPU-bound parallelization.

## Contents

- [Bounded Work with errgroup](#bounded-work-with-errgroup)
- [Channels of Channels](#channels-of-channels)
- [CPU-Bound Parallelization](#cpu-bound-parallelization)
- [Common Mistakes](#common-mistakes)

## Bounded Work with errgroup

Use `golang.org/x/sync/errgroup` for related operations when one failure should
cancel the remaining work. For finite input, `SetLimit` bounds active tasks
without a separate worker pool. Keep a worker pool when queue ownership or
long-lived workers are part of the contract.

```go
func processAll(ctx context.Context, items []Item, limit int) error {
    if limit <= 0 {
        return fmt.Errorf("concurrency limit must be positive: %d", limit)
    }
    g, groupCtx := errgroup.WithContext(ctx)
    g.SetLimit(limit)
    for _, item := range items {
        if groupCtx.Err() != nil {
            break
        }
        g.Go(func() error {
            if err := groupCtx.Err(); err != nil {
                return err
            }
            return process(groupCtx, item)
        })
    }
    if err := g.Wait(); err != nil {
        return err
    }
    return ctx.Err()
}
```

`process` must honor its context, including blocking I/O and channel operations.
The checks stop observed cancellation from reaching new work; a `Go` call
already waiting for a slot can still start its wrapper after cancellation.
The final `ctx.Err()` preserves parent cancellation even when no task ran.

- `Wait` waits for **all started tasks** and returns their first non-nil error;
  it does not return immediately on the first failure or aggregate failures.
- `WithContext` cancels the derived context on the first task error **or when
  `Wait` returns**, even on success. Use the parent context for subsequent work.
  Cancellation is cooperative; it cannot stop a task that ignores it.
- `SetLimit(0)` prevents new tasks; a negative limit means unbounded concurrency.
  This example rejects both because its contract requires a positive bound.
  Set the limit before starting tasks and do not change it while tasks run.
- `Go` blocks while the limit is full; that wait is not context-selectable.
  Avoid nested submissions to the same saturated group. If admission itself
  must cancel promptly, use a context-aware semaphore or worker queue.

For independent failures that must all be collected, see
[go-error-handling](../../go-error-handling/SKILL.md#handling-errors).
Source: [errgroup API](https://pkg.go.dev/golang.org/x/sync/errgroup).

---

## Channels of Channels

> **Source**: Effective Go

A channel is a first-class value that can be allocated and passed around like
any other. A powerful pattern is embedding a **reply channel** inside a request
struct, letting each client provide its own path for the answer:

```go
type Request struct {
    args       []int
    f          func([]int) int
    resultChan chan int
}
```

The client sends a request with a function, its arguments, and a channel on
which to receive the result:

```go
request := &Request{[]int{3, 4, 5}, sum, make(chan int)}
clientRequests <- request
fmt.Printf("answer: %d\n", <-request.resultChan)
```

The server handler reads from the queue and sends results back on each
request's reply channel:

```go
func handle(queue chan *Request) {
    for req := range queue {
        req.resultChan <- req.f(req.args)
    }
}
```

This pattern forms the basis for a rate-limited, parallel, non-blocking RPC
system without a mutex in sight.

---

## CPU-Bound Parallelization

> **Source**: Effective Go (modernized)

When a computation can be broken into independent pieces, parallelize it across
CPU cores using a `sync.WaitGroup` to wait for completion. Use `WaitGroup.Go`
(Go 1.25+) so the add/done bookkeeping stays coupled to the goroutine:

```go
type Vector []float64

func (v Vector) DoSome(i, n int, u Vector) {
    for ; i < n; i++ {
        v[i] += u.Op(v[i])
    }
}

func (v Vector) DoAll(u Vector) {
    numCPU := runtime.GOMAXPROCS(0)
    var wg sync.WaitGroup
    for i := range numCPU {
        wg.Go(func() {
            v.DoSome(i*len(v)/numCPU, (i+1)*len(v)/numCPU, u)
        })
    }
    wg.Wait()
}
```

Size the fan-out with `runtime.GOMAXPROCS(0)`, not `runtime.NumCPU()`:
`NumCPU` reports hardware cores and ignores the cgroup CPU limit, so a
container capped at 2 CPUs on a 64-core host spawns 64 workers that thrash.
Since Go 1.25 the default `GOMAXPROCS` is cgroup-aware, and
`runtime.SetDefaultGOMAXPROCS()` restores that default after a manual override.

No `i := i` capture line — loop variables are per-iteration since Go 1.22, and
`for i := range numCPU` replaces the three-clause loop.

> **Important**: Don't confuse concurrency (structuring a program as
> independently executing components) with parallelism (executing calculations
> simultaneously on multiple CPUs). Go is a concurrent language; not all
> parallelization problems fit its model.

---

## Common Mistakes

### Forgetting to signal completion

If a goroutine never calls `wg.Done()` (or never sends on a done channel), the
waiting goroutine blocks forever:

```go
// Bad: Missing wg.Done — deadlocks
var wg sync.WaitGroup
wg.Add(1)
go func() {
    doWork()
}()
wg.Wait()

// Good: WaitGroup.Go owns completion (Go 1.25+); doWork must not panic
var wg sync.WaitGroup
wg.Go(doWork)
wg.Wait()
```

### Unbounded goroutine spawning

Launching one goroutine per work item with no limit can exhaust memory or
overwhelm downstream resources. Use a semaphore to cap concurrency. Validate
`maxWorkers > 0` before this pattern: zero blocks the first send for nonempty
input, and a negative channel capacity panics.

```go
// Bad: Spawns len(items) goroutines at once
var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func(it Item) {
        defer wg.Done()
        process(it)
    }(item)
}
wg.Wait()

// Good: Semaphore limits concurrency to maxWorkers
var wg sync.WaitGroup
sem := make(chan struct{}, maxWorkers)
for _, item := range items {
    sem <- struct{}{}
    wg.Go(func() {
        defer func() { <-sem }()
        process(item) // must not panic
    })
}
wg.Wait()
```
