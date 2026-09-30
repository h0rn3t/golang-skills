# Style Principles Reference

> Sources: source/google-go-styleguide/guide.md (Style principles)
> Authority: advisory
> Last verified: 2026-09-10

## 1. Clarity

The code's purpose and rationale must be clear to the reader.

## 2. Simplicity

Code should accomplish goals in the simplest way possible.

### Least Mechanism

Where there are several ways to express the same idea, prefer the most standard
tool:

1. Core language constructs (channel, slice, map, loop, struct)
2. Standard library (HTTP client, template engine)
3. Third-party library — only when (1) and (2) don't suffice

## 3. Concision

Code should have high signal-to-noise ratio.

```go
// Good: Common idiom, high signal
if err := doSomething(); err != nil {
    return err
}

// Good: Signal boost for unusual case
if err := doSomething(); err == nil { // if NO error
    // ...
}
```

## 4. Maintainability

Code is edited many more times than written.

## 5. Consistency

Code should look and behave like similar code in the codebase.

- Package-level consistency is most important
- When ties occur, break in favor of consistency
- Never override documented style principles for consistency
- Consistency keeps conventions, not Go versions: an idiom the module's `go`
  directive has superseded is written in its current form beside the older
  neighbor ([Write Current Go](../SKILL.md#write-current-go); project policy,
  2026-09-13)
