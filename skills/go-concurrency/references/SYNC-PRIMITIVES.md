# Sync Primitives Patterns

> Sources: source/uber-go-style/style.md (Zero-value Mutexes are Valid, Do not embed mutexes, Atomic); https://pkg.go.dev/sync/atomic; https://pkg.go.dev/time#NewTimer
> Authority: advisory
> Minimum Go: typed atomics (`atomic.Int64`) 1.19; `sync.OnceFunc`/`OnceValue`/`OnceValues` 1.21; synchronous timer channels 1.23
> Last verified: 2026-10-19

## Don't Embed Mutexes

If you use a struct by pointer, the mutex should be a non-pointer field. Do not
embed the mutex on the struct, even if the struct is not exported.

---

## Atomic Operations

Prefer typed atomics in new code. In an existing struct, swapping a raw `int64` or
`unsafe.Pointer` field for `atomic.Int64` or `atomic.Pointer[T]` makes the
struct no-copy: every by-value copy of it then fails `go vet` (`passes lock by
value: … contains sync/atomic.noCopy`), and `go fix -atomictypes ./...` makes
the swap without rewriting those copies. `atomic.Value` differs from
`atomic.Pointer[T]` (`Value` panics on a nil store and on a change of dynamic
type): migrate a type as one change, not call by call.

---

## Run Once

`sync.OnceFunc`, `sync.OnceValue`, and `sync.OnceValues` (Go 1.21+) replace a
`sync.Once` paired with a result field and a getter: the returned function
runs the wrapped one at most once and hands every caller the same result.

```go
// Good: one declaration; every call returns the same *Config
var config = sync.OnceValue(load)
```

If the wrapped function panics, every later call panics with the same value,
so one-time work that can fail returns its error — `sync.OnceValues` carries
`(T, error)`. The error is cached like the value: every later call gets the
first failure. Work whose failure can be transient (a dial, a token fetch)
keeps a mutex-guarded field that stores only a success, so the next call
retries.

---

## Timers

Go 1.23 timer semantics hold for every module on a Go 1.27 toolchain (on
1.26, for a `go` directive of 1.23 or later): `t.C` is unbuffered, `Stop` and
`Reset` never leave a stale value behind, and an unreferenced timer is
collected without `Stop`.

- The old drain `if !t.Stop() { <-t.C }` now blocks forever: `Stop` returns
  false only when the value was already received or the timer was already
  stopped. Delete it; `t.Reset(d)` alone is correct.
- A leftover `GODEBUG=asynctimerchan=1` in the deployment's environment stops
  a binary built with Go 1.27 at startup (`removed GODEBUG "asynctimerchan"`);
  delete it with the drain.
- `time.After` in a `select` loop allocates a timer per iteration and restarts
  the timeout every time another case wins. A per-message idle timeout is one
  `time.NewTimer` before the loop and `t.Reset(d)` after each message; a
  deadline for the whole loop is a `context.WithTimeout` created once.

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
