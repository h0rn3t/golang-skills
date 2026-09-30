# Formatting Reference

> Sources: source/effective-go/effective_go.html (Formatting); source/google-go-styleguide/guide.md (Formatting, Line length); source/uber-go-style/style.md (Line Length)
> Authority: advisory; gofmt conformance is normative
> Last verified: 2026-09-10

## gofmt is Required

All Go source files **must** conform to `gofmt` output. No exceptions.

## Line Length

There is **no rigid line length limit** in Go, but avoid uncomfortably long
lines. Uber suggests a soft limit of 99 characters.

Guidelines:
- If a line feels too long, **refactor** rather than just wrap
- Don't split before indentation changes (function declarations, conditionals)
- Don't split long strings (URLs) into multiple lines
- Wrapping a signature or its arguments is
  [go-functions](../../go-functions/SKILL.md#function-signatures)'s rule
- If it's already as short as practical, let it remain long

## Semicolons and Local Consistency

Go inserts semicolons at specified line endings. Keep an opening control-block
brace with its statement; let `gofmt` handle layout. Explicit semicolons are
normally needed only to separate if/switch initializers or for clauses.
