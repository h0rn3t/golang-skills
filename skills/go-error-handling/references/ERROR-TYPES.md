# Error Types Reference

> Sources: source/uber-go-style/style.md (Error Types); https://pkg.go.dev/errors
> Authority: advisory
> Last verified: 2026-09-10

This reference covers structured error types, sentinel errors, and how to choose
the right error type for your use case.

---

## Error Structure

> The error-type decision table is in [the parent skill](../SKILL.md#error-types).
> This reference covers: expanded code examples, sentinel errors, error checking
> with `errors.Is`/`errors.AsType`, and structured error types.

**Key considerations**:

- Does the caller need to inspect an existing cause or a new condition?
- Does the caller need additional structured fields, or just context in the message?
- Exported error variables/types become part of your public API

```go
// No matching needed, static message
func Open() error {
    return errors.New("could not open")
}
```

```go
// A new stable condition callers need to match - export a sentinel
var ErrCouldNotOpen = errors.New("could not open")

func Open() error {
    return ErrCouldNotOpen
}
```

```go
// Existing cause plus dynamic context - retain the cause without a new type
if err != nil {
    return fmt.Errorf("open %q: %w", file, err)
}
```

---

## Sentinel Errors

The simplest structured errors are unparameterized global values:

```go
// Good: Sentinel errors for programmatic checking
var (
    // ErrDuplicate occurs if this animal has already been seen.
    ErrDuplicate = errors.New("duplicate")

    // ErrMarsupial occurs because we're allergic to marsupials.
    ErrMarsupial = errors.New("marsupials are not supported")
)

func process(animal Animal) error {
    switch {
    case seen[animal]:
        return ErrDuplicate
    case marsupial(animal):
        return ErrMarsupial
    }
    seen[animal] = true
    return nil
}
```

---

## Checking Errors

For direct comparison (when errors are not wrapped):

```go
// Good: Direct comparison with sentinel
switch err := process(an); err {
case ErrDuplicate:
    return fmt.Errorf("feed %q: %v", an, err)
case ErrMarsupial:
    alternate := an.BackupAnimal()
    return handlePet(alternate)
}
```

When errors may be wrapped, use `errors.Is`:

```go
// Good: Works with wrapped errors
switch err := process(an); {
case errors.Is(err, ErrDuplicate):
    return fmt.Errorf("feed %q: %v", an, err)
case errors.Is(err, ErrMarsupial):
    // Try to recover...
}
```

**Never** match errors based on string content:

```go
// Bad: Fragile string matching
if regexp.MatchString(`duplicate`, err.Error()) {...}
if regexp.MatchString(`marsupial`, err.Error()) {...}
```

---

## Structured Error Types

Use a struct when callers need additional fields as programmatic information.
If the operation and path are only message context, `%w` is sufficient to
preserve an existing cause. Here the caller also needs to inspect `Op` and
`Path` as fields:

```go
// Good: Structured error with accessible fields
type PathError struct {
    Op   string
    Path string
    Err  error
}

func (e *PathError) Error() string {
    return e.Op + " " + e.Path + ": " + e.Err.Error()
}

func (e *PathError) Unwrap() error { return e.Err }
```

Callers can use `errors.AsType` (Go 1.26+) to extract the structured error:

```go
if pathErr, ok := errors.AsType[*PathError](err); ok {
    fmt.Println("Failed path:", pathErr.Path)
}
```

---

## Error Values At An API

The rules a reviewer applies to the error side of a signature. The parent
skill carries only the decisions; these are the conventions.

- **Return `error`, never a concrete type.** A `*os.PathError` result that is
  nil becomes a non-nil `error` in the caller
  ([go-defensive](../../go-defensive/SKILL.md#common-pitfalls) owns the
  typed-nil mechanism).

  ```go
  // Bad: Concrete type can cause subtle bugs
  func Bad() *os.PathError { /*...*/ }

  // Good: Always return the error interface
  func Good() error { /*...*/ }
  ```

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

- **Discard deliberately.** An ignored error carries a comment saying why on
  the same line (`n, _ := b.Write(p) // never returns a non-nil error`);
  `errcheck` reports the ones that do not.

---

## Quick Reference

Choose the representation with [Error Types](../SKILL.md#error-types).

| Inspection | Use |
|------------|-----|
| Checking sentinel errors | `errors.Is(err, ErrFoo)` |
| Extracting structured errors | `errors.AsType[*PathError](err)` (Go 1.26+) |
