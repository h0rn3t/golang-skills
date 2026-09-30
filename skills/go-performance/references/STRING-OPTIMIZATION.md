# String Optimization Patterns

> Sources: source/uber-go-style/style.md (Prefer strconv over fmt, Avoid repeated string-to-byte conversions, Prefer Specifying Container Capacity)
> Authority: advisory
> Last verified: 2026-09-10

## Retained Substrings

A long-lived substring can keep a large input allocation alive. When profiling
shows this retention, copy just the part that must survive the operation:

```go
id := strings.Clone(record[:cut])
```

Keep a plain substring for short-lived use; cloning every parsed string adds
allocations and can make memory use worse. Copy at the retention boundary,
not at every intermediate step. `strings.Clone` returns equal text in its own
allocation (except the empty string), so the small retained value no longer
keeps the large input alive. See [strings.Clone](https://pkg.go.dev/strings#Clone).

## String Concatenation

When writing to an `io.Writer`, use `fmt.Fprintf` directly instead of building a
temporary string with `fmt.Sprintf`. When appending to a `[]byte` already in
hand, `buf = fmt.Appendf(buf, ...)` (Go 1.19+) skips the string as well.
