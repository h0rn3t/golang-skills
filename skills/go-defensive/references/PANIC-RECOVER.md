# Panic and Recover Patterns

> Sources: source/effective-go/effective_go.html (Panic, Recover); source/uber-go-style/style.md (Do not Panic)
> Authority: advisory
> Last verified: 2026-09-29

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

---

## Recover Patterns

`recover` regains control of a panicking goroutine. It only works inside
deferred functions.

### Package-Internal Panic/Recover

Use panic internally but convert to errors at API boundaries:

```go
type Error string

func (e Error) Error() string { return string(e) }

func (regexp *Regexp) error(err string) {
    panic(Error(err))
}

func Compile(str string) (regexp *Regexp, err error) {
    regexp = new(Regexp)
    defer func() {
        if e := recover(); e != nil {
            regexp = nil
            err = e.(Error)  // Re-panics if not our Error type
        }
    }()
    return regexp.doParse(str), nil
}
```

**Key points:**

- Deferred functions can modify named return values
- Type assertion `e.(Error)` re-panics on unexpected error types
- A panic raised for the package's own control flow never reaches a caller—convert it at the API boundary

---

## Quick Reference

| Pattern | Description |
|---------|-------------|
| Package-internal | Panic internally, recover and return error at API boundary |
| Type-safe recovery | Use type assertion to re-panic on unexpected errors |

## When to Use

- **Panic**: A programming error — API misuse, an unreachable case, a `MustX` failure at init
- **Recover**: Goroutines you start, because `net/http` recovers only its handler goroutine; package-internal error simplification
- **Never**: Let an input or environment error cross a package boundary as a panic—return it
