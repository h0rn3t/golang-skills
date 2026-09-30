# Function Signatures

> Sources: source/uber-go-style/style.md (Avoid Naked Parameters, Function Grouping and Ordering); source/google-go-styleguide/decisions.md (Function formatting)
> Authority: advisory
> Last verified: 2026-09-10

---

## Shortening Call Sites

Factor out local variables instead of splitting function calls across lines:

```go
// Bad: long inline call
ranked := search.Rank(
    scoring.Weighted(votes, weights),
    text.Normalize(query),
    defaultOptions,
)

// Good: locals named for what they hold
score := scoring.Weighted(votes, weights)
terms := text.Normalize(query)
ranked := search.Rank(score, terms, defaultOptions)
```

Preserve the original left-to-right call order when introducing locals.

---

## Avoid Naked Parameters

Naked parameters in function calls hurt readability. Add C-style comments for
ambiguous arguments:

```go
// Bad: what do these booleans mean?
printInfo("foo", true, true)

// Good: inline comments clarify intent
printInfo("foo", true /* isLocal */, true /* done */)
```

Better yet, replace naked `bool` parameters with custom types:

```go
type Region int

const (
    UnknownRegion Region = iota
    Local
)

type Status int

const (
    Pending Status = iota
    Done
)

func printInfo(name string, region Region, status Status)
```

### When to Use Each Approach

| Approach | When |
|----------|------|
| C-style comments | Quick fix; few call sites; third-party API you can't change |
| Custom types | Multiple call sites; public API; more than one bool/int parameter |
| Config structs or functional options | Optional constructor settings; choose by caller needs using [OPTIONS-VS-STRUCTS.md](OPTIONS-VS-STRUCTS.md) |
