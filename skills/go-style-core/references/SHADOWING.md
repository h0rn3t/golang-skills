# Variable Shadowing

> **Normative**: Be aware that `:=` in inner scopes creates a new variable that shadows the outer one.

### The Trap

```go
// Bug: err in the inner scope shadows the outer err
var err error
if condition {
    val, err := someFunc()  // new err — outer err stays nil
    if err == nil {
        use(val)
    }
}
return err  // always nil!
```

### Fix: Assign to the Outer Variable

```go
var err error
if condition {
    var val int
    val, err = someFunc()  // assigns to outer err
    if err == nil {
        use(val)
    }
}
return err  // correct
```

### Detection

Avoid hiding predeclared identifiers such as `error`, `string`, `len`, `new`,
`make`, or `any`; choose a name for the value's role instead. Plain `go vet`
does not generally diagnose shadowing. For outer-variable checks, use the
optional shadow analyzer or the repository's configured lint checks.

With the `shadow` analyzer installed, invoke it via `go vet`:

```bash
go vet -vettool=$(which shadow) ./...
```

Or add `govet` with shadow check enabled in `.golangci.yml`.
