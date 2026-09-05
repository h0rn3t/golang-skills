# Variable Scope and Declaration Patterns

> **Source**: Uber Style Guide, Google Style Guide

Detailed patterns for choosing between `var` and `:=` and reducing variable
scope in Go.

---

## Top-Level Declarations

Group related `var`, `const`, and `type` declarations in blocks; keep unrelated
top-level declarations separate. Adjacent local declarations may share a block
when this improves readability. Follow the repository's naming convention for
globals; the `_` prefix in examples is not a universal rule.

At the top level, always use `var`. Do not specify the type unless it differs
from the expression's type:

```go
// Bad: redundant type
var _s string = F()

// Good: type inferred
var _s = F()
```

Specify the type when the desired type differs from the expression:

```go
type myError struct{}

func (myError) Error() string { return "error" }
func F() myError              { return myError{} }

// F returns myError but we want the error interface
var _e error = F()
```

---

## Local Variable Patterns

### Use `:=` with explicit values

```go
// Bad
var s = "foo"

// Good
s := "foo"
```

### Use `var` for intentional zero values

`var` signals "this starts empty on purpose":

```go
// Bad: empty literal hides intent
filtered := []int{}

// Good: var signals intentional nil slice
var filtered []int
```

This is especially important for slices: `[]int{}` marshals to `[]` in JSON
while `nil` marshals to `null`. Choose based on your API contract.

### Type annotation when RHS is unclear

Use `var` with an explicit type when the type isn't obvious from the right-hand
side:

```go
// Type not obvious from function name alone
var ratio float64 = computeRatio()
```

---

## Reducing Scope

### Redeclaration versus reassignment

In the same block, `:=` may reuse an existing variable if at least one other
non-blank variable is new and the reused variable keeps its type:

```go
f, err := os.Open(name)
if err != nil {
    return err
}
defer f.Close()
d, err := f.Stat() // new d, same err
```

An inner block instead declares a new variable with that name. See
[SHADOWING.md](SHADOWING.md) before changing assignment across a block boundary.

### If-init pattern

Move declarations as close to usage as possible. Use if-init to limit scope:

```go
// Bad: err lives beyond where it's needed
err := os.WriteFile(name, data, 0644)
if err != nil {
    return err
}

// Good: err scoped to the if block
if err := os.WriteFile(name, data, 0644); err != nil {
    return err
}
```

### When NOT to reduce scope

Don't reduce scope if it forces deeper nesting or if you need the result after
the `if`:

```go
// Good: data used after the error check
data, err := os.ReadFile(name)
if err != nil {
    return err
}

if err := cfg.Decode(data); err != nil {
    return err
}

fmt.Println(cfg)
```

### Scope constants to functions

Move constants into functions when only used there:

```go
func Bar() {
    const (
        defaultPort = 8080
        defaultUser = "user"
    )
    fmt.Println("Default port", defaultPort)
}
```

---

## Decision Tree: var vs :=

```
Is it top-level?
├── Yes → use var
└── No (local)
    ├── Assigning a value? → use :=
    ├── Intentional zero value? → use var
    └── Type differs from RHS? → use var with type
```
