# Sync Primitives Patterns

> Sources: source/uber-go-style/style.md (Zero-value Mutexes are Valid, Do not embed mutexes, Atomic); https://pkg.go.dev/sync/atomic
> Authority: advisory
> Minimum Go: typed atomics (`atomic.Int64`) 1.19
> Last verified: 2026-09-10

Detailed patterns for mutexes and atomic operations — covering mutex embedding
pitfalls and type-safe atomic access.

---

## Don't Embed Mutexes

If you use a struct by pointer, the mutex should be a non-pointer field. Do not
embed the mutex on the struct, even if the struct is not exported.

```go
// Bad: Embedded mutex exposes Lock/Unlock as part of API
type SMap struct {
    sync.Mutex // Lock() and Unlock() become methods of SMap
    data map[string]string
}

func (m *SMap) Get(k string) string {
    m.Lock()
    defer m.Unlock()
    return m.data[k]
}
```

```go
// Good: Named field keeps mutex as implementation detail
type SMap struct {
    mu   sync.Mutex
    data map[string]string
}

func (m *SMap) Get(k string) string {
    m.mu.Lock()
    defer m.mu.Unlock()
    return m.data[k]
}
```

With the bad example, `Lock` and `Unlock` methods are unintentionally part of
the exported API. With the good example, the mutex is an implementation detail
hidden from callers.

---

## Atomic Operations: Full Example

The raw-value functions in `sync/atomic` operate on `int32`, `int64`, etc.,
making it easy to forget atomic access on some paths. Prefer its typed atomics:

```go
// Bad: Easy to forget atomic operation
type foo struct {
    running int32 // atomic
}

func (f *foo) start() {
    if atomic.SwapInt32(&f.running, 1) == 1 {
        return // already running
    }
    // start the Foo
}

func (f *foo) isRunning() bool {
    return f.running == 1 // race! forgot atomic.LoadInt32
}
```

```go
// Good: Type-safe atomic operations
type foo struct {
    running atomic.Bool
}

func (f *foo) start() {
    if f.running.Swap(true) {
        return // already running
    }
    // start the Foo
}

func (f *foo) isRunning() bool {
    return f.running.Load() // can't accidentally read non-atomically
}
```

The standard-library `atomic.Bool`, `atomic.Int64`, etc. types add type safety
by hiding the underlying type. They need no external dependency on Go 1.27.
Preserve an existing `go.uber.org/atomic` convention or use it for a required
operation the standard library does not provide.

---

## Channel Direction Examples

Specifying direction prevents accidental misuse:

```go
// Good: Direction specified - clear ownership
func sum(values <-chan int) int {
    total := 0
    for v := range values {
        total += v
    }
    return total
}
```

```go
// Bad: No direction - allows accidental misuse
func sum(values chan int) (out int) {
    for v := range values {
        out += v
    }
    close(values) // Bug! The consumer closes a channel it does not own; a
    return out    // send-only or receive-only type would have rejected this.
}
```
