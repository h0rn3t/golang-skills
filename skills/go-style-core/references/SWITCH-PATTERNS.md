# Switch Patterns

> Sources: https://go.dev/ref/spec#Switch_statements; source/effective-go/effective_go.html (Switch)
> Authority: normative
> Last verified: 2026-09-10

## No Automatic Fallthrough

Go `switch` cases do **not** fall through by default (unlike C/Java). Each case
body implicitly breaks. Use `fallthrough` only when explicitly needed — it is
rare in idiomatic Go.

## Expression-less Switch

A `switch` with no expression switches on `true`. Use it for clean if-else-if
chains when comparing a single variable against multiple conditions.

## Break with Labels

`break` inside a `switch` terminates only the switch, **not** an enclosing
`for` loop. Use a label to break out of the loop:

```go
Loop:
    for _, v := range items {
        switch v.Type {
        case "done":
            break Loop
        }
    }
```

A plain `break` in that case would leave only the `switch`.

**Rule of thumb**: Whenever you have a `switch` inside a `for` and need to
exit the loop from a case, always use a labeled break.

## Type Switches

For type switches (`switch v := x.(type)`), see
[go-interfaces](../../go-interfaces/SKILL.md): Type Switch.
