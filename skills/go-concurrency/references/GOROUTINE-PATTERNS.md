# Goroutine Lifecycle Patterns

> Sources: source/uber-go-style/style.md (Goroutine Lifetimes, No goroutine leaks); source/golang-wiki/CodeReviewComments.md (Goroutine Lifetimes)
> Authority: advisory
> Last verified: 2026-09-29

## Stop/Done Channel Pattern

Every goroutine must have a predictable stop mechanism. Use a stop channel to
signal shutdown and a done channel to confirm exit:

```go
var (
    stop = make(chan struct{}) // tells the goroutine to stop
    done = make(chan struct{}) // tells us that the goroutine exited
)
go func() {
    defer close(done)
    ticker := time.NewTicker(delay)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            flush()
        case <-stop:
            return
        }
    }
}()

// To shut down:
close(stop)  // signal the goroutine to stop
<-done       // and wait for it to exit
```

A spawned goroutine keeps this form — `time.NewTicker` plus `select` on
its stop channel or `ctx.Done()` — even when it is meant to run until exit: a
`for range time.Tick(d)` goroutine has no stop and nothing to wait on
([Core Rules](../SKILL.md#core-rules) 1-2).

---

## No Goroutines in init()

```go
// Bad: Spawns uncontrollable background goroutine
func init() {
    go doWork()
}
```

```go
// Good: Explicit lifecycle management
type Worker struct {
    stop chan struct{}
    done chan struct{}
}

func NewWorker() *Worker {
    w := &Worker{
        stop: make(chan struct{}),
        done: make(chan struct{}),
    }
    go w.doWork()
    return w
}

func (w *Worker) Shutdown() {
    close(w.stop)
    <-w.done
}
```

---

## Prefer Synchronous Functions

```go
// Good: Synchronous function - caller controls concurrency
func ProcessItems(items []Item) ([]Result, error) {
    var results []Result
    for _, item := range items {
        result, err := processItem(item)
        if err != nil {
            return nil, err
        }
        results = append(results, result)
    }
    return results, nil
}
```
