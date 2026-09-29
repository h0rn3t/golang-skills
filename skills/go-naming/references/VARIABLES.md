# Variable Names

> Sources: source/google-go-styleguide/decisions.md (Variable names); source/golang-wiki/CodeReviewComments.md (Variable Names); source/uber-go-style/style.md (Prefix Unexported Globals with _)
> Authority: advisory
> Last verified: 2026-09-10

This reference provides detailed guidance on naming variables in Go, covering scope-based
naming, single-letter conventions, and avoiding type redundancy.

## Length Proportional to Scope

> **Advisory**: Short names for small scopes, longer names for large scopes.

| Scope        | Lines  | Name Length |
|--------------|--------|-------------|
| Small        | 1-7    | 1-2 chars   |
| Medium       | 8-15   | short word  |
| Large        | 15-25  | descriptive |
| Very large   | 25+    | full words  |

```go
// Good - short scope, short name
for i := range items {
    process(items[i])
}

// Good - larger scope, clearer name
func processOrders(orders []*Order) error {
    pendingOrders := filterPending(orders)
    // ... 20+ lines of processing ...
    return nil
}
```

## Single-Letter Variables

> **Advisory**: Use single letters only when meaning is obvious.

Appropriate uses:
- Loop indices: `i`, `j`, `k`
- Coordinates: `x`, `y`, `z`
- Receivers: one or two letters
- Common types: `r` for `io.Reader`, `w` for `io.Writer`
- Short loops: `for _, n := range nodes`

```go
// Good - familiar conventions
func Copy(w io.Writer, r io.Reader) (int64, error)

for i, v := range values {
    process(v)
}

// Bad - unclear single letters
func Process(a, b, c string) error  // what are a, b, c?
```

## Avoid Type in Variable Name

> **Advisory**: Don't include the type in the variable name.

| Repetitive (Bad)             | Better               |
|------------------------------|----------------------|
| `var numUsers int`           | `var users int`      |
| `var nameString string`      | `var name string`    |
| `var primaryProject *Project`| `var primary *Project`|
| `var userSlice []User`       | `var users []User`   |

When disambiguating multiple forms, use meaningful qualifiers:

```go
// Good - meaningful distinction
limitRaw := r.FormValue("limit")
limit, err := strconv.Atoi(limitRaw)

// Also good
limitStr := r.FormValue("limit")
limit, err := strconv.Atoi(limitStr)
```

## Unexported Globals

> **Advisory**: Name an unexported top-level `var` or `const` for its role,
> with no prefix — new code with no neighbor to match included. The `_` prefix
> is the Uber style guide's convention, which the Google style guide does not
> use; follow it only in a package that already uses it.

```go
// Good - new code
const (
    defaultPort = 8080
    defaultUser = "user"
)

// Only where the package already prefixes its globals (Uber convention)
const _defaultTimeout = 30 * time.Second
```

Unexported error values keep the `err` prefix without the underscore, in a
package that uses `_` too:

```go
var errUserNotFound = errors.New("user not found")
var errInvalidInput = errors.New("invalid input")
```
