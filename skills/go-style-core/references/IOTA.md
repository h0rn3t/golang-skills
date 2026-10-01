# Constants and iota Patterns

> Sources: source/uber-go-style/style.md (Start Enums at One); source/google-go-styleguide/decisions.md (Constant naming)
> Authority: advisory
> Last verified: 2026-10-01

## Start Enums at One

Start enums at one so the zero value represents an invalid/unset state. This
catches uninitialized variables:

```go
type Operation int

const (
    Add Operation = iota + 1
    Subtract
    Multiply
)
```

### When Zero Makes Sense

Use zero when the default behavior is desirable.
The key question: **is the zero value a valid, useful default?** If yes, start
at zero. If no, start at one.

## Bitmask Patterns

Use bit-shifting with `iota` for flag/bitmask enums:

```go
type Permission int

const (
    Read Permission = 1 << iota
    Write
    Execute
)

perms := Read | Write
```

## String Representation

An enum whose values reach logs or `%v` output gets a `String()`: generated
by `stringer` where `go.mod` already tracks it
(`tool golang.org/x/tools/cmd/stringer`) or the task asks for it; otherwise a
hand-written switch, since `go get -tool` adds `golang.org/x/tools` to `go.mod`
(a new module on the [dependency ladder](../../go-packages/SKILL.md#dependency-ladder)).
The tracked tool (Go 1.24) replaces a `tools.go` of blank imports, and
`go generate ./...` writes `operation_string.go`:

```go
//go:generate go tool stringer -type=Operation
type Operation int
```

Write the switch by hand too when the text differs from the constant names.

## Grouping Rules

- Each enum type gets its own `const` block — `iota` resets to 0 in each block
- Unrelated constants go in separate blocks
- Document the enum type, not each individual constant (unless behavior is
  non-obvious)
