# Godoc Formatting Reference

> Sources: https://go.dev/doc/comment (Go 1.19 doc comment syntax)
> Authority: normative
> Minimum Go: doc links, `#` headings, and list syntax 1.19; `go doc -http` 1.25
> Last verified: 2026-10-01

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

**Lists** — lines indented and starting with `-`, `*`, or `1.`; text after the
list starts after a blank `//` line, which gofmt inserts along with the
normalized marker and indentation:

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

An old-style heading — a lone capitalized line with no final punctuation,
between blank `//` lines — still renders as a heading, and gofmt rewrites it to
`// # Heading`. An old-style list written as an indented block without `-`,
`*`, or `1.` markers renders as code, not as a list.

---

## Signal Boosting

A comment inside a function that flags an easily-missed form, such as
`err == nil` where `!= nil` is usual, is an internal comment:
[go-style-core PRINCIPLES.md](../../go-style-core/references/PRINCIPLES.md#3-concision)
owns it.

---

## Documentation Preview

> **Advisory**: Preview documentation before and during code review.

```bash
go doc -all .    # terminal rendering, Go 1.19+ syntax
go doc -http     # the pkg.go.dev view in a browser, pkgsite pinned by the toolchain (Go 1.25+)
```

Both render the syntax above.
