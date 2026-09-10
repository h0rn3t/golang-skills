# Error Wrapping Reference

> Sources: source/uber-go-style/style.md (Error Wrapping); https://go.dev/blog/go1.13-errors; https://pkg.go.dev/errors#AsType
> Authority: advisory
> Minimum Go: `errors.AsType` 1.26
> Last verified: 2026-09-10

This reference covers error wrapping with `%v` vs `%w`, placement conventions,
adding context to errors, and logging best practices.

---

## Wrapping Errors: %v vs %w

The choice between `%v` and `%w` significantly impacts how errors are propagated
and inspected.

### Use %v for Simple Annotation

Use `%v` when you want to:

- Add context without preserving the error chain for programmatic inspection
- Deliberately keep the cause opaque under the API's error contract
- Log or display errors to humans

`%v` preserves the error text, not its identity or type. It does not redact
sensitive details; external responses may need a separate safe message.

```go
// Good: keep the cause opaque when that is the API's contract
func (s *Server) SuggestFortune(ctx context.Context, req *pb.Request) (*pb.Response, error) {
    if err != nil {
        return nil, fmt.Errorf("couldn't find fortune database: %v", err)
    }
}
```

### Use %w for Error Chain Preservation

Use `%w` when you want callers to programmatically inspect the underlying error:

```go
// Good: %w preserves error chain for errors.Is/errors.As
func (s *Server) internalFunction(ctx context.Context) error {
    if err != nil {
        return fmt.Errorf("couldn't find remote file: %w", err)
    }
}

// Caller can now check:
if errors.Is(err, fs.ErrNotExist) {
    // Handle not found case
}
```

### When to Use Each

**Use %w when**:
- Adding context while preserving the original error for programmatic inspection
- You explicitly document and test the underlying errors you expose

**Use %v when**:
- The API deliberately omits the underlying cause from programmatic inspection
- Logging or displaying to humans
- Creating independent errors that hide implementation details

---

## Placement of %w

Place `%w` at the **end** of the error string so error text mirrors error chain
structure:

```go
// Good: %w at end - prints newest to oldest
err1 := fmt.Errorf("err1")
err2 := fmt.Errorf("err2: %w", err1)
err3 := fmt.Errorf("err3: %w", err2)
fmt.Println(err3) // err3: err2: err1
```

```go
// Bad: %w at start - prints oldest to newest (confusing)
err1 := fmt.Errorf("err1")
err2 := fmt.Errorf("%w: err2", err1)
err3 := fmt.Errorf("%w: err3", err2)
fmt.Println(err3) // err1: err2: err3
```

```go
// Bad: %w in middle - incoherent order
err1 := fmt.Errorf("err1")
err2 := fmt.Errorf("err2-1 %w err2-2", err1)
err3 := fmt.Errorf("err3-1 %w err3-2", err2)
fmt.Println(err3) // err3-1 err2-1 err1 err2-2 err3-2
```

**Pattern**: Use the form `context message: %w`

---

## Adding Information to Errors

### Add Context, Not Redundancy

Add information that you have but the caller/callee might not. Avoid duplicating
information the underlying error already provides:

```go
// Good: Adds meaningful context
f, err := os.Open("settings.txt")
if err != nil {
    return fmt.Errorf("launch codes unavailable: %v", err)
}
defer f.Close()
// Output: launch codes unavailable: open settings.txt: no such file or directory
```

```go
// Bad: Duplicates the filename
f, err := os.Open("settings.txt")
if err != nil {
    return fmt.Errorf("could not open settings.txt: %v", err)
}
defer f.Close()
// Output: could not open settings.txt: open settings.txt: no such file or directory
```

### Don't Annotate Without Purpose

If the annotation only indicates failure without adding information, just return
the error:

```go
// Bad: Annotation adds nothing
return fmt.Errorf("failed: %v", err)

// Good: Just return the error
return err
```

---

## Logging Errors

When the error is handled by logging, [go-logging](../../go-logging/SKILL.md)
owns the level, the attribute shape, and the `Enabled` guard for expensive
attributes; log it once, at the point that handles it.

### Protect Sensitive Information

Be careful with PII (Personally Identifiable Information) in log messages. Many
log sinks are not appropriate for sensitive user data.

---

## Quick Reference

| Pattern | Guidance |
|---------|----------|
| `%v` | Display text or deliberately omit the chain; does not redact the message |
| `%w` | Use to preserve error chain for programmatic inspection |
| `%w` placement | Always at the end: `"context: %w"` |
| Adding context | Add new info, don't duplicate existing info |
| Empty annotation | Just return `err` instead of `fmt.Errorf("failed: %v", err)` |
| Logging | Don't log and return; use appropriate log levels |
