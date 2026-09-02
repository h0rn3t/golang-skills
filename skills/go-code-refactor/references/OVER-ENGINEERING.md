# Over-Engineering Audit

Two lanes, and the scope note below belongs to only one of them.

**The restraint ladder** (next section) is normative whenever Go code is
written, on any task — the cheapest over-engineering to remove is the kind
never added. It does not restrict what the task may fix.

**The audit lane** (everything after the ladder — tags, output format, hunt
list) applies when the task is "what can we delete" rather than "make this
read better": a repo or package handed over as bloated, over-abstracted, or
dependency-heavy, and the wanted output is a ranked cut list, not a diff.
In that lane the scope is complexity only — correctness bugs, races, and
security holes found along the way are collected and handed back, never fixed
in the same pass; route them to
[go-code-review](../../go-code-review/SKILL.md). The audit reports; the
refactor workflow in `SKILL.md` applies what the user picks. On a write-mode
task the audit scope does not apply: fix what the task asked for.

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
   missing.
4. **Does a language or toolchain feature cover it?** `go:embed` over an asset
   loader, struct tags over a hand-written marshaler, a DB constraint over an
   app-side check, `t.Cleanup` over hand-rolled teardown, `testing/synctest`
   over sleep-based waits, a build tag over a runtime switch.
5. **A module already in `go.mod`?** Use it. Never add a new one for what a few
   lines do — [go-packages](../../go-packages/SKILL.md) owns the module rungs
   (stdlib → `golang.org/x/...` → existing module → new module).
6. **Can it be one line?** One line.
7. **Only then**: the minimum that works.

Two rungs both work → take the higher one and move on. But if the higher rung
makes the **call site** read worse, take the lower one: the ladder minimizes
concepts, not clarity. [go-style-core](../../go-style-core/SKILL.md) owns the
order — Clarity > Simplicity > Concision — and it outranks every rung here.

Check the result at the call site, not the declaration count. A helper that
exists only to be undone where it is called has not removed complexity; it
moved it one line down and added a name to learn.

**Never on the chopping block**: input validation at trust boundaries, error
handling that prevents data loss, security controls, accessibility basics in
anything user-facing, clarity at the call site, and anything the user asked
for explicitly. Lazy, not negligent — a "simplification" that drops a bounds
check is a bug.

A deliberate shortcut with a known ceiling is not a cut. Mark it with
`Kept:` / `Ceiling:` / `Fix:` (see `SKILL.md`) so `check-debt.sh` harvests it
instead of it rotting into "later means never".

## Tags

One tag per finding. The tag names why the code should stop existing.

| Tag | Means | Replacement |
|---|---|---|
| `delete:` | Dead code, unreachable branch, unused export, speculative feature | Nothing |
| `stdlib:` | Hand-rolled thing the standard library ships | Name the function |
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

**Hand-rolled standard library**

| Hand-rolled | Replacement |
|---|---|
| Linear search loop over a slice | `slices.Contains`, `slices.IndexFunc` |
| Manual sort comparator boilerplate | `slices.SortFunc`, `cmp.Compare` |
| `+=` string building in a loop | `strings.Builder` → [go-performance](../../go-performance/SKILL.md) |
| Custom multi-error accumulator | `errors.Join` |
| `sync.Once` + captured result field | `sync.OnceFunc`, `sync.OnceValue` |
| `map[T]bool` used only for membership | `map[T]struct{}` |
| `wg.Add(1)` / `go func(){ defer wg.Done() }()` | `wg.Go(f)` (Go 1.25+) |
| `errors.As` with a declared target var | `errors.AsType[T]` (Go 1.26+) |
| Hand-rolled min/max/clear helpers | `min`, `max`, `clear` builtins |
| `for i := 0; i < n; i++` over an index only | `for i := range n` |
| Deep-copy helpers for slices and maps | `slices.Clone`, `maps.Clone` |
| Hand-written fan-out `slog.Handler` | `slog.NewMultiHandler` (Go 1.26+) |
| Path-traversal guards around `filepath.Join` | `os.Root` |
| Test HTTP mocks for a real client path | `httptest.NewTestServer(t, h)` |
| Struct boxing `(T, error)` plus an `unwrap`/`get` method | Return the two values — Go has multiple returns; box only where one value is required, such as a map value, and unbox at that boundary |
| `time.Sleep` waits in concurrency tests | `testing/synctest` → [go-testing](../../go-testing/SKILL.md) |

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
- Table tests with one row, or a subtest per input where a table is shorter →
  [go-testing](../../go-testing/SKILL.md).

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
