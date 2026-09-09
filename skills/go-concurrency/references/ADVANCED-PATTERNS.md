# Advanced Concurrency Patterns

Detailed reference for advanced concurrency patterns from Effective Go. These
patterns are situational — use when you need request/response multiplexing or
CPU-bound parallelization.

---

## Waiting, Errors, and Cancellation

`WaitGroup` joins tasks without collecting errors or cancelling siblings.
`errgroup.Group` collects the first error and waits for all started tasks;
only `errgroup.WithContext` also cancels the derived context on an error.
Tasks must observe that context for cancellation to stop their work. `Wait`
also cancels the derived context on return, so do not reuse it for later work.

```go
g, workCtx := errgroup.WithContext(ctx)
g.SetLimit(maxWorkers) // positive limit, set before starting tasks
for _, item := range items {
    g.Go(func() error { return process(workCtx, item) })
}
return g.Wait()
```

This finite-batch form bounds active tasks. A blocked `g.Go` admission is not
itself cancellation-aware; an endless queue needing prompt cancellation uses
the fixed-worker loop in the parent skill. Bound queued logical work separately.
`golang.org/x/sync/errgroup` is an external module; keep `WaitGroup` when no
error collection is needed. Choose from actual lifetime and failure contracts.

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

This demonstrates request/reply ownership only. Both sends can block; a real
service still needs cancellation, admission limits, and a receiver lifetime.

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
overwhelm downstream resources. Use a semaphore to cap concurrency:

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
