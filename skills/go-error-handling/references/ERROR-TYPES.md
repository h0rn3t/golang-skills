# Error Types Reference

> Sources: source/uber-go-style/style.md (Error Types); https://pkg.go.dev/errors
> Authority: advisory
> Last verified: 2026-09-10

## Checking Errors

Match a sentinel with `errors.Is`, even when the callee does not wrap today:
`switch err { case ErrDuplicate: }` compares with `==`, stops matching the
first time a wrap is added, and `errorlint` in the gate reports it.

```go
// Good: Works with wrapped errors
switch err := process(an); {
case errors.Is(err, ErrDuplicate):
    return fmt.Errorf("feed %q: %w", an, err)
case errors.Is(err, ErrMarsupial):
    // Try to recover...
case err != nil:
    return fmt.Errorf("process feed %q: %w", an, err)
}
```

## Error Values At An API

- **Message form.** Error strings are not capitalized and do not end with
  punctuation, because they are usually printed after other context;
  exported names, proper nouns, and acronyms keep their case. Displayed
  messages (logs, test failures, API responses) may be capitalized.
  `staticcheck` ST1005 reports the rest.
- **Other results are unspecified on error.** When a function returns a
  non-nil error, callers treat every other return value as meaningless
  unless the documentation says otherwise. A function that takes a
  `context.Context` usually returns an `error`, so the caller can tell a
  cancellation from a result.
- **No in-band errors.** `-1`, `nil`, or `""` as the failure signal forces
  every caller to remember the convention; return `(T, error)` or
  `(T, bool)`, which also makes `Parse(Lookup(key))` a compile error instead
  of a silent misuse.

  ```go
  // Bad: In-band error value
  func Lookup(key string) int  // returns -1 for missing

  // Good: Explicit error or ok value
  func Lookup(key string) (string, bool)
  ```
