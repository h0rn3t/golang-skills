# String Optimization Patterns

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

## strconv vs fmt

For primitive conversions, `strconv` avoids the general formatting machinery
of `fmt`. Benchmark the actual values and toolchain before claiming a speedup.

Benchmark snippets use `b.Loop()` (Go 1.24+).

**Bad:**

```go
for b.Loop() {
    _ = fmt.Sprint(n)
}
```

**Good:**

```go
for b.Loop() {
    _ = strconv.Itoa(n)
}
```

Use the same `n` for both variants; benchmark setup is outside the loop.
See [benchmark methodology](BENCHMARKS.md) for executable examples.

Common conversions:

| Task | `fmt` | `strconv` |
|------|-------|-----------|
| Int → string | `fmt.Sprint(n)` | `strconv.Itoa(n)` |
| Int64 → string | `fmt.Sprint(n)` | `strconv.FormatInt(n, 10)` |
| Float → string | `fmt.Sprint(f)` | `strconv.FormatFloat(f, 'f', -1, 64)` |
| String → int | — | `strconv.Atoi(s)` |
| Bool → string | `fmt.Sprint(b)` | `strconv.FormatBool(b)` |

---

## Repeated String-to-Byte Conversions

Reuse converted bytes when measurement shows a benefit. The compiler can
already avoid conversion allocations for some concrete writers.

**Bad:**

```go
for b.Loop() {
    w.Write([]byte("Hello world"))
}
```

**Good:**

```go
data := []byte("Hello world")
for b.Loop() {
    w.Write(data)
}
```

Compare both forms with a bounded writer workload; see
[benchmark methodology](BENCHMARKS.md#benchmark-examples-from-performance-patterns).
Do not assume each conversion allocates or that hoisting it gives a fixed gain.

---

## String Concatenation

Choose the right string building strategy based on complexity.

### Use `+` for Simple Cases

```go
key := "projectid: " + p
```

The `+` operator is efficient for a small, fixed number of strings. The compiler
can often optimize adjacent string literals.

### Use `fmt.Sprintf` for Formatting

```go
// Good: clear formatting
str := fmt.Sprintf("%s [%s:%d]-> %s", src, qos, mtu, dst)

// Bad: + with manual conversions
str := src.String() + " [" + qos.String() + ":" + strconv.Itoa(mtu) + "]-> " + dst.String()
```

When writing to an `io.Writer`, use `fmt.Fprintf` directly instead of building a
temporary string with `fmt.Sprintf`.

### Use `strings.Builder` for Piecemeal Construction

`strings.Builder` takes amortized linear time, whereas repeated `+` or
`fmt.Sprintf` take quadratic time when building a large string:

```go
b := new(strings.Builder)
for i, d := range digitsOfPi {
    fmt.Fprintf(b, "the %d digit of pi is: %d\n", i, d)
}
str := b.String()
```

### Use Backticks for Constant Multi-line Strings

```go
// Good: raw string literal
usage := `Usage:

custom_tool [args]`

// Bad: concatenation with escape sequences
usage := "" +
    "Usage:\n" +
    "\n" +
    "custom_tool [args]"
```

### Strategy Summary

| Method | Best For | Performance |
|--------|----------|-------------|
| `+` | Few strings, simple concat | O(n) for small n |
| `fmt.Sprintf` | Formatted output | Slower, but clearer |
| `strings.Builder` | Loop/piecemeal construction | Amortized O(n) |
| `strings.Join` | Joining a slice | O(n) |
| `strings.Clone` | Detach a retained substring when profiling justifies the copy | Copies the retained bytes |
| Backtick literal | Constant multi-line text | Zero cost |
