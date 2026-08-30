# Over-Engineering Audit

Read this when the task is "what can we delete" rather than "make this read
better": a repo or package handed over as bloated, over-abstracted, or
dependency-heavy, and the wanted output is a ranked cut list, not a diff.

Scope is complexity only. Correctness bugs, races, and security holes found
along the way are collected and handed back, never fixed in the same pass —
route them to [go-code-review](../../go-code-review/SKILL.md). The audit
reports; the refactor workflow in `SKILL.md` applies what the user picks.

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
