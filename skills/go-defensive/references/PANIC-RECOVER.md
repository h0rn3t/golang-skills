# Panic and Recover Patterns

> Sources: source/effective-go/effective_go.html (Panic, Recover); source/uber-go-style/style.md (Do not Panic)
> Authority: advisory
> Last verified: 2026-10-01

## Panic Guidelines

`panic` creates a run-time error that stops the program. An error caused by
input or the environment is returned, never panicked; a panic marks a
programming error.

### When to Panic

Real library functions should **avoid panic**. If the problem can be masked or
worked around, let things continue rather than taking down the whole program.
A case no input can reach is the exception. This fragment switches over an
unexported enum whose every value is a constant of the same package, so the
`default` is a bug here, never caller input:

```go
func eval(c opcode, a, b int) int {
    switch c {
    case opAdd:
        return a + b
    case opSub:
        return a - b
    default:
        panic(fmt.Sprintf("eval: unknown opcode %d", c))
    }
}
```

### Panic in Initialization

`init()` may panic when a package cannot set itself up from its own constants
and embedded files — a bug the first test run reports. A value from the
environment is not that case: read it in `main` and return the error.

```go
// Bad: a missing variable is the environment, not a bug in this package
func init() {
    if os.Getenv("USER") == "" {
        panic("no value for $USER")
    }
}

// Good: run, called from main, returns it and main exits non-zero
user := os.Getenv("USER")
if user == "" {
    return errors.New("USER is not set")
}
```

### When Panic IS Acceptable

Beyond initialization, panic is acceptable in these narrow cases:

1. **API misuse** — analogous to how core language panics on out-of-bounds
   access. The `reflect` package uses this approach.
2. **Internal implementation detail with matching `recover`** at the package
   boundary. Panic simplifies deeply-nested control flow while the public API
   still returns errors (the Parse/parseInt pattern below).
3. **`panic("unreachable")`** after `log.Fatal` when the compiler can't detect
   unreachable code.

#### Parse/parseInt Pattern

Use panic internally to unwind complex recursion, but always convert to an error
at the package boundary:

```go
func parseInt(in string) int {
    n, err := strconv.Atoi(in)
    if err != nil {
        panic(&syntaxError{"not a valid integer"})
    }
    return n
}

func Parse(in string) (_ *Node, err error) {
    defer func() {
        if p := recover(); p != nil {
            sErr, ok := p.(*syntaxError)
            if !ok {
                panic(p)  // not ours — re-panic
            }
            err = fmt.Errorf("syntax error: %v", sErr.msg)
        }
    }()
    // ... calls parseInt internally
}
```

**Key**: The type check `p.(*syntaxError)` ensures only *our* panics are caught.
Unexpected panics (nil pointer, etc.) propagate normally.

## Recover in Goroutines and Middleware

`net/http` recovers a handler's panic, logs it, and closes the connection; a
goroutine you start is outside that, and its panic ends the process unless its
own deferred function recovers. `recover` returns nil unless the deferred
function calls it directly: `defer handlePanic()` works, a closure that calls
`handlePanic()` does not.

```go
go func() {
    defer func() {
        if p := recover(); p != nil {
            logger.Error("worker panic", "panic", p, "stack", string(debug.Stack()))
        }
    }()
    work(ctx)
}()
```

A recovery middleware re-panics `http.ErrAbortHandler`.
`httputil.ReverseProxy` panics with it when the backend body fails mid-copy;
a middleware that swallows it and writes a 500 hands the client a truncated
body that reads as a complete response:

```go
func recoverer(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            p := recover()
            if p == nil {
                return
            }
            if err, ok := p.(error); ok && errors.Is(err, http.ErrAbortHandler) {
                panic(p) // keep the abort: the client must see a broken response
            }
            slog.ErrorContext(r.Context(), "handler panic", "panic", p, "stack", string(debug.Stack()))
            http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        }()
        next.ServeHTTP(w, r)
    })
}
```
