# Variable Scope and Declaration Patterns

> Sources: source/uber-go-style/style.md (Reduce Scope of Variables, Local Variable Declarations); source/google-go-styleguide/decisions.md (Declarations)
> Authority: advisory
> Last verified: 2026-09-10

## Top-Level Declarations

Group related `var`, `const`, and `type` declarations in blocks; keep unrelated
top-level declarations separate. Adjacent local declarations may share a block
when this improves readability. Naming globals, including the `_` prefix, is
[go-naming](../../go-naming/references/VARIABLES.md#unexported-globals)'s rule.

At the top level, always use `var`. Do not specify the type unless it differs
from the expression's type.

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

Move declarations as close to usage as possible. Use if-init to limit scope.

### When NOT to reduce scope

Don't reduce scope if it forces deeper nesting or if you need the result after
the `if`.

### Scope constants to functions

Move constants into functions when only used there.

## Decision Tree: var vs :=

```
Is it top-level?
├── Yes → use var
└── No (local)
    ├── Assigning a value? → use :=
    ├── Intentional zero value? → use var
    └── Type differs from RHS? → use var with type
```
