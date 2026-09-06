# Bulk Mechanical Rewrites

> Sources: `gofmt` docs; golang.org/x/tools/cmd/eg; github.com/uber-go/gopatch; golang.org/x/tools/go/analysis; golang.org/x/tools/cmd/deadcode; github.com/dave/dst
> Authority: advisory — tool selection; the behavior promise stays in SKILL.md
> Minimum Go: 1.27 baseline
> Last verified: 2026-09-06

When the same edit recurs across many call sites, hand-editing each one is the
wrong instrument: a generated rewrite is reviewable, re-runnable, and testable
against golden files, and thirty scattered manual edits are none of those. The
tools below are ordered by increasing power. Start at the top and move down only
when the current one cannot express the rewrite.

Single-symbol work — rename, extract, inline, references — is
[GOPLS.md](GOPLS.md)'s, and version-driven idiom updates are
[MODERNIZATION.md](MODERNIZATION.md)'s. This file covers the rest.

> **Normative**: Never `sed` or `perl` a structural Go change. None of the text
> tools have grammar awareness, so a pattern that happens to match inside a
> string literal or a comment is rewritten right alongside real code.

## `gofmt -r` — syntactic, one expression

Purely syntactic and type-unaware: it matches expression shape, not the types
involved. Wildcards are single lowercase identifiers matching any subexpression.

```bash
gofmt -r 'bytes.Compare(a, b) == 0 -> bytes.Equal(a, b)' -w .
gofmt -d file.go     # diff without writing — read it before -w
```

Because it cannot see types, it cannot distinguish a call on the type you meant
from a look-alike of the same name in another package. That distinction needs
`eg`. The `./...` pattern belongs to `go list`, not `gofmt`; for a recursive
rewrite, enumerate the exact non-generated files in a checked script rather
than passing `./...` as a path or sweeping generated/vendor trees.

## `eg` — type-aware, by example

`golang.org/x/tools/cmd/eg` rewrites from a template declaring `before` and
`after` functions of identical type, each with a single-expression body.

```go
// template.go
package template

func before(s string) error { return fmt.Errorf("%s", s) }
func after(s string) error  { return errors.New(s) }
```

```bash
eg -t template.go -w ./...
```

Matching is semantic, so `func(x int)` in the template also matches `func(y int)`
at the call site. Limits worth knowing before you plan a migration around it:
expressions only, no statements or function-literal patterns; the rewrite cannot
change the expression's type; imports are added but never removed; and
duplicating a wildcard in the `after` template duplicates whatever side effect
the matched expression had.

## `gopatch` — statements, import-aware

`github.com/uber-go/gopatch` works at statement level and tracks imports as part
of the patch, so it can operate on a tree that does not fully compile — the
usual state in the middle of a large migration. Metavariables are declared
between `@@` markers, then a diff-like body:

```
@@
var x expression
@@
-errors.New(fmt.Sprintf(x))
+fmt.Errorf(x)
```

```bash
gopatch -d -p rewrite.patch ./...   # dry run first
gopatch -p rewrite.patch ./...
```

Still beta, and the project describes it as covering roughly 80% of a migration
rather than all of it. A pattern cannot match an import statement in isolation —
something has to follow it.

## A `go/analysis` fixer — bespoke and testable

For a rewrite too specific for the three above, write an `analysis.Analyzer`:
`Run(pass)` walks the type-checked AST and reports
`analysis.Diagnostic{SuggestedFixes: ...}` at each match. Test it against
`.golden` files with `analysistest.RunWithSuggestedFixes`, then ship it as a
`singlechecker` binary or run it through `go vet -vettool=<path>`.

This is the rung where the rewrite itself gets a test — which is why it is worth
reaching even for a one-off, when the alternative is an unreviewable diff across
dozens of files.

`//go:fix inline` marks a function or constant so `go fix` folds every call site
into its replacement — the machine-executable end of a deprecation migration
once the replacement exists.

## `dave/dst` — when comments must survive

`go/ast` stores comments in a side table keyed by byte offset, so moving,
reordering, or deleting nodes desyncs them from the code they described. That is
the root cause of the comment loss that makes gopls's extract action a
medium-risk transform.

`github.com/dave/dst` attaches comments and blank-line spacing as node-local
decorations, so a hand-rolled rewrite round-trips them via `decorator.Parse` and
`decorator.Print`; `dstutil.Apply` mirrors `astutil.Apply`, so existing rewrite
logic ports over. Reach for it only when a bespoke fixer has to preserve
comments — it is a dependency, and the ladder in
[OVER-ENGINEERING.md](OVER-ENGINEERING.md) applies to tooling too.

## After every bulk rewrite

```bash
goimports -w .     # even after a tool that claims to manage imports itself
deadcode ./...     # find what the rewrite orphaned
```

`deadcode` builds reachability with rapid type analysis from `main` and `init`,
so it is unsound with respect to assembly, `go:linkname`, and reflection-driven
dispatch. A "dead" verdict on code using any of those is a strong hint, not a
proof — the deletion bar in [SKILL.md](../SKILL.md) is unchanged. Then run the
repository gate: a rewrite that touched thirty files earns the full check, not
the focused one.
