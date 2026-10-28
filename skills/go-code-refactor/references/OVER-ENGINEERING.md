# Over-Engineering Audit

> Sources: github.com/DietrichGebert/ponytail; Go CodeReviewComments; `$GOROOT/api/go1.2*.txt`
> Authority: project policy for the audit lane; version claims checked against the api files
> Minimum Go: 1.27 baseline (the plugin repository's COMPATIBILITY.md); each table row names its own minimum
> Last verified: 2026-10-01 against go1.27.1

**Reach For What Go Ships** is the replacement catalog for rungs 3 and 4 of
the [restraint ladder](LADDER.md): read it when one of those rungs needs a
concrete standard-library or toolchain replacement.

**The audit lane** (from Tags on) applies when the wanted output is a ranked
cut list, not a diff. Its scope is complexity only — correctness bugs, races,
and security holes found along the way are handed back to
[go-code-review](../../go-code-review/SKILL.md), never fixed in the same pass.

## Contents

- [Reach For What Go Ships](#reach-for-what-go-ships)
- [Tags](#tags)
- [Output](#output)
- [Go hunt list](#go-hunt-list)
- [Prove it before you cut](#prove-it-before-you-cut)

## Reach For What Go Ships

[Ladder](LADDER.md) rungs 3 and 4, made concrete: before writing a loop, helper, wrapper, or type,
find its row and write the right-hand column. Versions are the minimum `go`
directive — `go vet`'s `stdversion` analyzer flags a newer symbol. On existing
code a swap may be observable; `MODERNIZATION.md` says what each can change.
The [idiom card](../../go-style-core/SKILL.md#current-go-idiom-card) already
carries `for i := range n`, `iter.Seq`, `cmp.Or`, `slices.Sorted(maps.Keys(m))`,
`slices.SortFunc`, the `slices` and `maps` loop helpers, `errors.Join`,
`errors.AsType[T]`, `wg.Go`, `context.AfterFunc`, `omitzero`, the stdlib
`uuid`, and `t.Context()`, each with its trap; read
those rows there. This table lists only what the card does not.

**Language**

| Instead of writing | Reach for |
|---|---|
| Copies of one function that differ only in a type | One generic function, or a generic method (Go 1.27) — [go-generics](../../go-generics/SKILL.md) owns the threshold; two rhyming copies can be cheaper than a type parameter |
| A struct with one method plus a constructor, to satisfy a one-method interface | A function type carrying the method, `http.HandlerFunc` style → [go-interfaces](../../go-interfaces/references/EMBEDDING.md) |
| Forwarding methods that all call the same field | Embed the type — [go-interfaces](../../go-interfaces/SKILL.md) owns the exported-struct caveat |
| A hand-written `String()` switch for an enum | `//go:generate go tool stringer -type=T` where `go.mod` already tracks it (`tool golang.org/x/tools/cmd/stringer`, Go 1.24) or the task asks for it; otherwise a hand-written switch → [go-style-core](../../go-style-core/SKILL.md) |
| A hand-written marshaler or field mapper | Struct tags → [go-defensive](../../go-defensive/SKILL.md) |
| An asset loader, a file read at start-up, a static-file handler | `//go:embed` with `embed.FS`; `http.FileServerFS` (Go 1.22) → [go-packages](../../go-packages/SKILL.md) |
| A runtime `GOOS`/`GOARCH` switch | A build tag |
| `(T, error)` boxed in a struct with a `get`/`unwrap` method | Return the two values; box only where one value is required, such as a map value |

**Collections and strings**

| Instead of writing | Reach for |
|---|---|
| A loop concatenating slices | `slices.Concat` (Go 1.22) → [go-data-structures](../../go-data-structures/SKILL.md) |
| `map[T]bool` used only for membership | `map[T]struct{}` |
| Per-type nullable wrappers | `sql.Null[T]` (Go 1.22) → [go-database](../../go-database/SKILL.md) |

**Errors, concurrency, context**

| Instead of writing | Reach for |
|---|---|
| A multi-error accumulator over a fixed set of errors | `fmt.Errorf` with several `%w` (Go 1.20) |
| A wrapping error type that only carries context | `fmt.Errorf("...: %w", err)` and `errors.Is` → [go-error-handling](../../go-error-handling/SKILL.md) |
| A `WaitGroup`, an error channel, and first-error logic; a semaphore channel | `errgroup.Group` and `SetLimit` (`golang.org/x/sync`: a new module on the dependency ladder unless `go.mod` already has it) → [go-concurrency](../../go-concurrency/SKILL.md) |
| A detached copy of a context | `context.WithoutCancel` → [go-context](../../go-context/SKILL.md) |

**I/O, HTTP, tests, logging**

| Instead of writing | Reach for |
|---|---|
| Open, read, close by hand; a recursive directory walker | `os.ReadFile`, `os.WriteFile`, `io.ReadAll`; `filepath.WalkDir` |
| Path-traversal guards around `filepath.Join` | `os.Root` (Go 1.24) → [go-defensive](../../go-defensive/SKILL.md) |
| Hand-rolled body limits and CSRF origin checks | `http.MaxBytesReader`; `http.NewCrossOriginProtection` (Go 1.25) → [go-security](../../go-security/SKILL.md) |
| Hand-parsed `os.Args` | `flag`, before any CLI module → [go-packages](../../go-packages/SKILL.md) |
| A random-string generator | `crypto/rand.Text()` (Go 1.24) |
| A fan-out or no-op log handler | `slog.NewMultiHandler` (Go 1.26), `slog.DiscardHandler` (Go 1.24) → [go-logging](../../go-logging/SKILL.md) |
| N near-identical test functions; hand-rolled teardown and env restore | One table with `t.Run`; `t.Cleanup`, `t.Setenv` → [go-testing](../../go-testing/SKILL.md) |

`go fix -diff` applies the rows it has modernizers for; the rest are yours.

## Tags

One tag per finding. The tag names why the code should stop existing.

| Tag | Means | Replacement |
|---|---|---|
| `delete:` | Dead code, unreachable branch, unused export, speculative feature | Nothing |
| `reuse:` | Helper, type, or pattern the repository already has, written again | Name the existing symbol and its file |
| `stdlib:` | Hand-rolled thing the standard library, the language, or the toolchain ships (`slices`, `go:embed`, struct tags, `synctest`) | Name the function or feature |
| `dep:` | Module doing what the stdlib or the toolchain already does | Name the stdlib replacement |
| `yagni:` | Abstraction with one implementation, config nobody sets, layer with one caller | Inline it |
| `shrink:` | Same behavior, fewer lines | Show the shorter form |

## Output

Ranked biggest cut first, one line per finding:

```
<file>:L<line>: <tag> <what to cut>. <replacement>.
```

End with `net: -<N> lines, -<M> deps`. Nothing to cut: `Lean already.`

Examples:

```
internal/store/repo.go:L14: yagni: UserRepository interface, one implementation, defined next to it. Return *PostgresStore; the consumer declares the interface it needs.
internal/util/strings.go:L8: stdlib: 22-line contains() over a slice. slices.Contains.
internal/api/wrap.go:L30-52: delete: retry wrapper around an in-process call that cannot fail transiently. Nothing.
net: -180 lines, -2 deps
```

## Go hunt list

Ordered by how much usually comes out.

**Layers that only forward**

- Interface declared in the same package as its single implementation, and
  returned by the constructor → [go-interfaces](../../go-interfaces/SKILL.md)
  owns the rule; the cut is to return the concrete type and let the consumer
  declare what it needs.
- Wrapper struct whose every method calls the same field with the same
  arguments. Call the inner type.
- `Manager`, `Service`, `Handler` layer with one caller and no state.
- Functional options constructor with one option, or options nobody passes →
  [go-functions](../../go-functions/SKILL.md).
- Generic function instantiated at exactly one type →
  [go-generics](../../go-generics/SKILL.md).
- `util`, `common`, `helpers`, `base` packages: move each function next to its
  single caller and the package disappears →
  [go-naming](../../go-naming/references/IDENTIFIERS.md#avoid-uninformative-names).

**Hand-rolled standard library** — every row of
[Reach For What Go Ships](#reach-for-what-go-ships) and of the
[idiom card](../../go-style-core/SKILL.md#current-go-idiom-card), read as a
cut: the left column is the finding, the right column the replacement.

**Flexibility nobody asked for**

- Config field, flag, or environment variable with one value across the repo
  and no test that varies it.
- `any` parameter that only ever receives one concrete type.
- Callback or hook slice with zero registrations outside tests.
- Channel or mutex where a return value would do; a goroutine whose caller
  immediately waits for it → [go-concurrency](../../go-concurrency/SKILL.md).
- Exported symbol in an `internal/` package with one in-tree caller — export
  is free flexibility that costs a rename later.
- `init()` doing work a caller could do explicitly.

**Volume without content**

- Files that export one small function and import nothing special.
- Constructors that only assign fields a composite literal could set.
- Comments restating the code; doc comments on unexported one-liners →
  [go-documentation](../../go-documentation/SKILL.md).
- Table tests with one row, or a subtest per input where a table is shorter —
  the checks stay, the scaffolding goes → [go-testing](../../go-testing/SKILL.md).

## Prove it before you cut

A deletion is a behavior change if the code was reachable. Exported symbols
need one more step: grep the module for callers, and confirm nothing outside
the module imports the package. If it is a published API, the finding is a
deprecation proposal, not a cut.
