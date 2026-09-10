# Godoc Formatting Reference

> Sources: https://go.dev/doc/comment (Go 1.19 doc comment syntax)
> Authority: normative
> Minimum Go: doc links, `#` headings, and list syntax 1.19
> Last verified: 2026-09-10

## Godoc Formatting

Doc comments use the syntax `go doc` and pkg.go.dev have rendered since Go
1.19 ([go.dev/doc/comment](https://go.dev/doc/comment)); `gofmt` reformats
comments into it, so an old-style heading is rewritten on the next save.

**Paragraphs** — separate with a blank `//` line:

```go
// LoadConfig reads a configuration out of the named file.
//
// See [ParseConfig] for the file format.
```

**Headings** — a line that starts with `# `, alone in its paragraph:

```go
// # Using headings
//
// A heading gets an anchor on pkg.go.dev; keep it short, no trailing period.
```

**Lists** — lines indented and starting with `-`, `*`, or `1.`; a blank `//`
line before and after; gofmt normalizes the marker and indentation:

```go
// LoadConfig treats these keys specially:
//   - "import" makes this configuration inherit from the named file.
//   - "env" is populated with the process environment.
```

**Code blocks** — indent by one tab (or two spaces) after a blank line:

```go
// Update runs the function in an atomic transaction:
//
//	if err := db.Update(func(s *State) { s.Foo = bar }); err != nil {
//		return err
//	}
```

**Doc links** — `[Name]`, `[Type.Method]`, or `[pkg.Name]` render as links to
the symbol; `[text]: URL` at the end of the comment defines a named link.
Bare URLs are linked automatically:

```go
// Process returns [ErrNotFound] if the input references a missing item.
// See [io.Writer] and the [design note].
//
// [design note]: https://example.com/design
```

Old-style Godoc (a heading as a plain capitalized line, lists as verbatim
blocks) still renders as paragraphs or code, not as headings or lists.

---

## Signal Boosting

> **Advisory**: Add comments to highlight unusual or easily-missed patterns.

These two are hard to distinguish:

```go
if err := doSomething(); err != nil {  // common
    // ...
}

if err := doSomething(); err == nil {  // unusual!
    // ...
}
```

Add a comment to boost the signal:

```go
// Good:
if err := doSomething(); err == nil { // if NO error
    // ...
}
```

---

## Documentation Preview

> **Advisory**: Preview documentation before and during code review.

```bash
go doc -all .                                   # terminal rendering, Go 1.19+ syntax
go install golang.org/x/pkgsite/cmd/pkgsite@latest && pkgsite   # the pkg.go.dev view
```

Both render the syntax above; an old-style heading shows up as a paragraph.
