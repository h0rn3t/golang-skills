# Loops and Iterators

> Sources: Go specification, For statements; Effective Go, For; COMPATIBILITY.md
> Authority: normative
> Minimum Go: integer range and per-iteration declared variables Go 1.22; iterator functions Go 1.23
> Last verified: 2026-09-10

## Choose the Loop by Its Semantics

On Go 1.22+, `for i := range n` expresses a simple count; do not mechanically
replace a loop whose bound changes during iteration. A scoped `go fix -diff`
previews supported modernization.

## Range Details That Affect Behavior

- Slice range values are copies. Use indices when changing the original elements.
- Map iteration order is unspecified. Sort keys explicitly when order is part
  of the result contract.
- String range produces a byte offset and a rune, not a byte value.
- Variables declared by a loop are per-iteration with Go 1.22+ language
  semantics; assignments to preexisting variables still reuse those variables.
  Check `go.mod` and file constraints before removing a capture workaround.

## Iterator Functions

For a collection naturally consumed once, consider `iter.Seq[T]` or
`iter.Seq2[K, V]` (Go 1.23+) instead of materializing a temporary slice.
Use slice results when callers need indexing, reuse, or an existing API contract.

```go
func (l *List[T]) All() iter.Seq[T] {
    return func(yield func(T) bool) {
        for n := l.head; n != nil; n = n.next {
            if !yield(n.val) {
                return
            }
        }
    }
}
```

Honor `yield` on every path: after it returns false, stop and never call it
again. Prefer `All` for whole-collection iteration; `Keys` and `Values` name
more specific views where useful. `iter.Pull` serves consumers needing explicit
next/stop control, such as lockstep traversal. Ensure its stop function is
called when ending consumption early; it is unnecessary for ordinary range.

## Conditional Scope and Exit Targets

In a switch nested inside a loop, an unlabeled `break` exits the switch.
Read [SWITCH-PATTERNS.md](SWITCH-PATTERNS.md) for labeled exits and fallthrough.
