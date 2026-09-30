# Advanced Concurrency Patterns

> Sources: https://pkg.go.dev/golang.org/x/sync/errgroup; source/effective-go/effective_go.html (Concurrency)
> Authority: advisory
> Last verified: 2026-09-29

Situational patterns for bounded work, request/response multiplexing, and
CPU-bound parallelization.

## Bounded Work with errgroup

Use `golang.org/x/sync/errgroup` for related operations when one failure should
cancel the remaining work. For finite input, `SetLimit` bounds active tasks
without a separate worker pool. Keep a worker pool when queue ownership or
long-lived workers are part of the contract.

```go
func processAll(ctx context.Context, items []Item, limit int) error {
    g, groupCtx := errgroup.WithContext(ctx)
    g.SetLimit(limit)
    for _, item := range items {
        g.Go(func() error { return process(groupCtx, item) })
    }
    return g.Wait()
}
```

`process` must honor its context, including blocking I/O and channel operations.
After the first failure the loop still submits the remaining items; each
starts, sees the cancelled context, and returns at once, so a `process` that
ignores its context turns one failure into a full run.

- `Wait` waits for **all started tasks** and returns their first non-nil error;
  it does not return immediately on the first failure or aggregate failures.
- `WithContext` cancels the derived context on the first task error **or when
  `Wait` returns**, even on success. Use the parent context for subsequent work.
  Cancellation is cooperative; it cannot stop a task that ignores it.
- `SetLimit(0)` prevents new tasks; a negative limit means unbounded concurrency.
  The caller passes a positive limit.
  Set the limit before starting tasks and do not change it while tasks run.
- `Go` blocks while the limit is full; that wait is not context-selectable.
  Avoid nested submissions to the same saturated group. If admission itself
  must cancel promptly, use a context-aware semaphore or worker queue.

For independent failures that must all be collected, see
[go-error-handling](../../go-error-handling/SKILL.md#error-flow).
Source: [errgroup API](https://pkg.go.dev/golang.org/x/sync/errgroup).

---

## Channels of Channels

A powerful pattern is embedding a **reply channel** inside a request
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
request := &Request{[]int{3, 4, 5}, sum, make(chan int, 1)} // one slot: the reply never blocks the server
clientRequests <- request
fmt.Printf("answer: %d\n", <-request.resultChan)
```

Because each reply channel has one slot, the server's send completes even when
the client has stopped waiting; an unbuffered reply channel would park the
handler forever on the first abandoned request.

---

## CPU-Bound Parallelization

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
When the main module's `go` directive is 1.25 or later the default `GOMAXPROCS`
is cgroup-aware (an older directive builds with `containermaxprocs=0` in
`DefaultGODEBUG`, and the default stays the host's CPU count), and
`runtime.SetDefaultGOMAXPROCS()` restores that default after a manual override.
