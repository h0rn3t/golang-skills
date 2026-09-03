# Over-Engineering Audit

> Sources: github.com/DietrichGebert/ponytail; Go CodeReviewComments; `$GOROOT/api/go1.2*.txt`
> Authority: project policy for the ladder and the write rules; version claims checked against the api files
> Minimum Go: 1.27 baseline (`COMPATIBILITY.md`); each table row names its own minimum
> Last verified: 2026-09-03 against go1.27.1

**The restraint ladder**, **Reach For What Go Ships**, and **Ship, Then
Question** are normative whenever Go code is written, on any task — the cheapest
over-engineering to remove is the kind never added. They do not restrict what
the task may fix.

**The audit lane** (from Tags on — the cut tags, the output format, the hunt
list) applies when the task is "what can we delete" rather than "make this
read better": a repo or package handed over as bloated, over-abstracted, or
dependency-heavy, and the wanted output is a ranked cut list, not a diff.
In that lane the scope is complexity only — correctness bugs, races, and
security holes found along the way are collected and handed back, never fixed
in the same pass; route them to
[go-code-review](../../go-code-review/SKILL.md). The audit reports; the
refactor workflow in `SKILL.md` applies what the user picks.

## Contents

- [The Restraint Ladder](#the-restraint-ladder)
- [Reach For What Go Ships](#reach-for-what-go-ships)
- [Ship, Then Question](#ship-then-question)
- [Tags](#tags)
- [Output](#output)
- [Go hunt list](#go-hunt-list)
- [Prove it before you cut](#prove-it-before-you-cut)

## The Restraint Ladder

> **Normative**: The ladder runs *after* the problem is understood, not instead
> of it. Read the code the change touches and trace the real flow end to end
> first, then climb. It shortens the solution, never the reading — the smallest
> diff in the wrong place is a second bug, not laziness.

Before writing a new line — type, layer, interface, option, helper, import —
stop at the first rung that holds:

1. **Does this need to exist at all?** Speculative need → skip it and say so in
   one line (YAGNI). On a refactor this rung usually holds: deletion beats
   rewrite.
2. **Already in this codebase?** A helper, type, or pattern two files over →
   reuse it. Re-implementing what the repo already has is the most common slop;
   look before you write.
3. **Does the standard library ship it?** `go doc <pkg>` before assuming it is
   missing; the table in Reach For What Go Ships is the checklist for 3 and 4.
4. **Does a language or toolchain feature cover it?** `go:embed` over an asset
   loader, struct tags over a hand-written marshaler, a DB constraint over an
   app-side check, `t.Cleanup` over hand-rolled teardown, `testing/synctest`
   over sleep-based waits, a build tag over a runtime switch.
5. **A module already in `go.mod`?** Use it. Never add a new one for what a few
   lines do — [go-packages](../../go-packages/SKILL.md) owns the module rungs
   (stdlib → `golang.org/x/...` → existing module → new module).
6. **Can it be one line?** One line.
7. **Only then**: the minimum that works, in the fewest files — a new file only
   when the existing one is unwieldy.

The ladder runs per entity, not once per task: each helper, type, and method
written to implement a needed feature gets its own climb — "the feature is
needed" does not carry over. Judge the final diff, not the opening plan.

Two rungs both work → take the higher one and move on. But if the higher rung
makes the **call site** read worse, take the lower one: the ladder minimizes
concepts, not clarity. [go-style-core](../../go-style-core/SKILL.md) owns the
order — Clarity > Simplicity > Concision — and it outranks every rung here.

Two options on the same rung, same size → take the one correct on edge cases.
Lazy means less code, not a flimsier algorithm.

**A bug fix is a root-cause fix.** A report names a symptom. Before the edit,
find every caller of the function you are about to touch ([GOPLS.md](GOPLS.md))
and land the fix where they all route through: one guard in the shared
function is a smaller diff than a guard in every caller, and patching only the
path the ticket named leaves the sibling callers broken.
[go-troubleshooting](../../go-troubleshooting/SKILL.md) owns the method while
the cause is still unknown.

Check the result at the call site, not the declaration count. A helper that
exists only to be undone where it is called has not removed complexity; it
moved it one line down and added a name to learn.

**Never on the chopping block**: input validation at trust boundaries, error
handling that prevents data loss, security controls, accessibility basics in
anything user-facing, clarity at the call site, the tests that fail when the
logic breaks, and anything the user asked for explicitly. Lazy, not negligent
— a "simplification" that drops a bounds check is a bug.

A deliberate shortcut with a known ceiling is not a cut. Mark it with
`Kept:` / `Ceiling:` / `Fix:` (see `SKILL.md`) so `check-debt.sh` harvests it
instead of it rotting into "later means never".

## Reach For What Go Ships

Rungs 3 and 4, made concrete: before writing a loop, helper, wrapper, or type,
find its row and write the right-hand column. Versions are the minimum `go`
directive — `go vet`'s `stdversion` analyzer flags a newer symbol. On existing
code a swap may be observable; `MODERNIZATION.md` says what each can change.

**Language**

| Instead of writing | Reach for |
|---|---|
| A three-clause loop over an index only | `for i := range n` (Go 1.22) → [go-control-flow](../../go-control-flow/SKILL.md) |
| A walker taking a callback, a channel-fed generator, a slice built only to be ranged | A function returning `iter.Seq[T]` / `iter.Seq2[K, V]`, consumed with `for v := range seq` (Go 1.23) → [go-control-flow](../../go-control-flow/SKILL.md) |
| `x := x` before a closure or goroutine in a loop | Nothing — per-iteration loop variables (Go 1.22) |
| Hand-written `min`/`max`; a loop deleting every key | `min`, `max`, `clear` builtins |
| A temporary declared only to take its address | `new(expr)` (Go 1.26) → [go-declarations](../../go-declarations/SKILL.md) |
| An `if`/`else` chain of fallbacks | `cmp.Or(a, b, c)` (Go 1.22) — it evaluates every argument |
| Copies of one function that differ only in a type | One generic function, or a generic method (Go 1.27) — [go-generics](../../go-generics/SKILL.md) owns the threshold; two rhyming copies can be cheaper than a type parameter |
| A struct with one method plus a constructor, to satisfy a one-method interface | A function type carrying the method, `http.HandlerFunc` style → [go-interfaces](../../go-interfaces/references/EMBEDDING.md) |
| Forwarding methods that all call the same field | Embed the type — [go-interfaces](../../go-interfaces/SKILL.md) owns the exported-struct caveat |
| A hand-written `String()` switch for an enum | `//go:generate go tool stringer -type=T` — `golang.org/x/tools`, tracked by a `tool` directive in `go.mod` (Go 1.24) instead of a `tools.go` of blank imports → [go-declarations](../../go-declarations/SKILL.md) |
| A hand-written marshaler or field mapper | Struct tags; `omitzero` in new code (Go 1.24) → [go-defensive](../../go-defensive/SKILL.md) |
| An asset loader, a file read at start-up, a static-file handler | `//go:embed` with `embed.FS`; `http.FileServerFS` (Go 1.22) → [go-packages](../../go-packages/SKILL.md) |
| A runtime `GOOS`/`GOARCH` switch | A build tag |
| `(T, error)` boxed in a struct with a `get`/`unwrap` method | Return the two values; box only where one value is required, such as a map value |

**Collections and strings**

| Instead of writing | Reach for |
|---|---|
| A linear search loop | `slices.Contains`, `slices.ContainsFunc`, `slices.IndexFunc` |
| A `sort.Slice` comparator; `sort.Strings` after a collect loop | `slices.SortFunc` with `cmp.Compare`; `slices.Sorted(maps.Keys(m))` (Go 1.23) |
| Collect, dedupe, reverse, concat, min/max loops | `slices.Compact`, `Reverse`, `Max`/`Min`, `Concat` (Go 1.22), `Collect` (Go 1.23); `maps.Keys`, `Values`, `Collect` (Go 1.23) → [go-data-structures](../../go-data-structures/SKILL.md) |
| Deep-copy helpers; `map[T]bool` used only for membership | `slices.Clone`, `maps.Clone`; `map[T]struct{}` |
| `Index` plus manual slicing | `strings.Cut`, `CutPrefix`, `CutSuffix`, `CutLast` (Go 1.27) |
| `strings.Split` then `range` over the slice | `strings.SplitSeq`, `FieldsSeq`, `Lines` (Go 1.24) — `Lines` keeps the newline |
| `+=` string building in a loop | `strings.Builder` → [go-performance](../../go-performance/SKILL.md) |
| Per-type nullable wrappers | `sql.Null[T]` (Go 1.22) → [go-database](../../go-database/SKILL.md) |

**Errors, concurrency, context**

| Instead of writing | Reach for |
|---|---|
| A multi-error accumulator | `errors.Join`; `fmt.Errorf` with several `%w` |
| A wrapping error type that only carries context | `fmt.Errorf("...: %w", err)` and `errors.Is` → [go-error-handling](../../go-error-handling/SKILL.md) |
| `errors.As` with a declared target variable | `errors.AsType[T]` (Go 1.26) |
| `sync.Once` plus a captured result field | `sync.OnceFunc`, `sync.OnceValue`, `sync.OnceValues` |
| `wg.Add(1)` / `go func() { defer wg.Done() }()` | `wg.Go(f)` (Go 1.25) → [go-concurrency](../../go-concurrency/SKILL.md) |
| A `WaitGroup`, an error channel, and first-error logic; a semaphore channel | `errgroup.Group` and `SetLimit` (`golang.org/x/sync`, the `x/` rung of the dependency ladder) |
| A mutex around a single counter or flag with no compound invariant | `atomic.Int64`, `atomic.Bool` → [go-concurrency](../../go-concurrency/SKILL.md) |
| A goroutine parked on `ctx.Done()` to run cleanup; a detached copy of a context | `context.AfterFunc`, `context.WithoutCancel` → [go-context](../../go-context/SKILL.md) |

**I/O, HTTP, tests, logging**

| Instead of writing | Reach for |
|---|---|
| Open, read, close by hand; a recursive directory walker | `os.ReadFile`, `os.WriteFile`, `io.ReadAll`; `filepath.WalkDir` |
| Path-traversal guards around `filepath.Join` | `os.Root` (Go 1.24) → [go-defensive](../../go-defensive/SKILL.md) |
| A router module or manual path parsing | `http.ServeMux` patterns like `"GET /users/{id}"` with `r.PathValue` (Go 1.22) → [go-http](../../go-http/SKILL.md) |
| Hand-rolled body limits and CSRF origin checks | `http.MaxBytesReader`; `http.NewCrossOriginProtection` (Go 1.25) → [go-security](../../go-security/SKILL.md) |
| Hand-parsed `os.Args` | `flag`, before any CLI module → [go-packages](../../go-packages/SKILL.md) |
| A random-string generator; a UUID module for `New`/`Parse` | `crypto/rand.Text()` (Go 1.24); stdlib `uuid` (Go 1.27) |
| A fan-out or no-op log handler | `slog.NewMultiHandler` (Go 1.26), `slog.DiscardHandler` (Go 1.24) → [go-logging](../../go-logging/SKILL.md) |
| N near-identical test functions; hand-rolled teardown, temp dirs, env and cwd restore | One table with `t.Run`; `t.Cleanup`, `t.TempDir`, `t.Setenv`, `t.Chdir` (Go 1.24), `t.Context()` (Go 1.24) → [go-testing](../../go-testing/SKILL.md) |
| `httptest.NewServer` + `defer srv.Close()`; HTTP mocks on a real client path | `httptest.NewTestServer(t, h)` (Go 1.27) |
| `time.Sleep` waits in concurrency tests; a `b.N` loop | `testing/synctest` (Go 1.25); `b.Loop()` (Go 1.24) |

A row that makes the call site read worse is skipped — the ladder minimizes
concepts, not clarity ([go-style-core](../../go-style-core/SKILL.md) owns the
order). `go fix -diff` applies the rows it has modernizers for; the rest are yours.

## Ship, Then Question

A request bigger than its need does not stop the turn. Ship the rung that
holds and question the rest in the same reply — "Did X; Y covers it. Need
full X? Say so." Never stall on an answer you can default. Stop only where no
default is safe: the four Stop-and-Ask cases of the refactor workflow
(go-code-refactor's `SKILL.md`), or two readings that lead to materially
different work. When the user hears the lazy version and insists on the full
one, build it — no re-arguing.

For new code, lead with the code and the gate result, then at most three short
lines: `skipped: <X>, add when <Y>`. If the explanation outgrows the code, cut
the explanation — a paragraph defending a simplification is complexity
smuggled back in as prose. A report, walkthrough, or per-step note the user
asked for is not prose debt; give it in full. A refactor reports through
`assets/refactor-report.md`, whose table is already the short form.

Lazy code without its check is unfinished. Non-trivial logic — a branch, a
loop, a parser, a money or security path — leaves a runnable check behind: one
check per behavior, several cases in one table, no fixtures, nothing that only
mirrors the function list. YAGNI applies to tests too: a trivial one-liner
needs none. Tests that fail when the logic breaks are never a cut;
[go-testing](../../go-testing/SKILL.md) owns the form.

## Tags

One tag per finding. The tag names why the code should stop existing.

| Tag | Means | Replacement |
|---|---|---|
| `delete:` | Dead code, unreachable branch, unused export, speculative feature | Nothing |
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
go.mod:L12: dep: github.com/google/uuid for New/Parse only. Stdlib uuid (Go 1.27).
internal/api/wrap.go:L30-52: delete: retry wrapper around an in-process call that cannot fail transiently. Nothing.
internal/config/load.go:L61: shrink: 12-line loop building a map. maps.Collect over the seq, 1 line.
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
  [go-functional-options](../../go-functional-options/SKILL.md).
- Generic function instantiated at exactly one type →
  [go-generics](../../go-generics/SKILL.md).
- `util`, `common`, `helpers`, `base` packages: move each function next to its
  single caller and the package disappears →
  [go-packages](../../go-packages/SKILL.md).

**Hand-rolled standard library** — every row of
[Reach For What Go Ships](#reach-for-what-go-ships), read as a cut: the left
column is the finding, the right column the replacement.

Dependencies to check against the stdlib before anything else: `pkg/errors`
(→ `fmt.Errorf` with `%w`, `errors.Is`/`As`), `golang.org/x/exp/slices` and
`maps` (→ `slices`, `maps`), `github.com/google/uuid` for the common calls
(→ stdlib `uuid`, Go 1.27), assertion libraries used only for equality
(→ `cmp.Diff`). `COMPATIBILITY.md` lists what landed when.

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

A deletion is a behavior change if the code was reachable. For each finding
that reaches the diff:

```bash
go build ./... && go test ./...          # the code was actually unused
go vet ./...                             # unreachable and ineffectual paths
golangci-lint run --enable unused ./...  # unused unexported symbols
go fix -diff ./...                       # stdlib replacements the tool can make
```

Exported symbols need one more step: grep the module for callers, and confirm
nothing outside the module imports the package. If it is a published API,
the finding is a deprecation proposal, not a cut.

`SKILL.md` owns the rest of the contract — what "identical behavior" means,
the never-simplify-away list, and the verification gate the applied cuts run
through.
