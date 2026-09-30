# Error Wrapping Reference

> Sources: source/uber-go-style/style.md (Error Wrapping); https://go.dev/blog/go1.13-errors; https://pkg.go.dev/errors#AsType
> Authority: advisory
> Minimum Go: `errors.AsType` 1.26
> Last verified: 2026-09-29

## Wrapping Errors: %v vs %w

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
        return nil, fmt.Errorf("find fortune database: %v", err)
    }
}
```

### Use %w for Error Chain Preservation

Use `%w` when you want callers to programmatically inspect the underlying error:

```go
// Good: %w preserves error chain for errors.Is/errors.As
func (s *Server) internalFunction(ctx context.Context) error {
    if err != nil {
        return fmt.Errorf("fetch remote file: %w", err)
    }
}

// Caller can now check:
if errors.Is(err, fs.ErrNotExist) {
    // Handle not found case
}
```

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

**Pattern**: Use the form `context message: %w`

---

## Adding Information to Errors

### Add Context, Not Redundancy

Add information that you have but the caller/callee might not. Avoid duplicating
information the underlying error already provides:

```go
// Good: Adds meaningful context
settings, err := os.ReadFile("settings.txt")
if err != nil {
    return fmt.Errorf("launch codes unavailable: %w", err)
}
// Output: launch codes unavailable: open settings.txt: no such file or directory
```

```go
// Bad: Duplicates the filename
settings, err := os.ReadFile("settings.txt")
if err != nil {
    return fmt.Errorf("could not open settings.txt: %v", err)
}
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
