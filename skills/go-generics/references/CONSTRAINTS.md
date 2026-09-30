# Type Constraints in Go Generics

> Sources: https://go.dev/ref/spec#Type_constraints; https://pkg.go.dev/cmp; source/google-go-styleguide/guide.md
> Authority: normative for constraint semantics
> Minimum Go: `cmp.Ordered` 1.21
> Last verified: 2026-09-10

Constraints define what operations a type parameter supports. Choose the
narrowest constraint that satisfies your function's needs — no more.

---

## Composing and Writing Constraints

> **Advisory**: Define a custom constraint only when no standard one fits.

Constraints can require methods alongside type elements:

```go
type Stringer interface {
    comparable
    String() string
}
```

A type satisfying `Stringer` must be comparable **and** have a `String()` method.

## Type Inference

> **Advisory**: Let the compiler infer type arguments when unambiguous.

The compiler infers type parameters from function arguments:

```go
result := slices.Contains[[]string](names, "alice") // explicit — unnecessary
result := slices.Contains(names, "alice")           // inferred — preferred
```

Supply type arguments explicitly only when there are no function arguments to
infer from, the inferred type is wrong (e.g., untyped constant promotes to the
wrong type), or readability benefits from making the type visible.
