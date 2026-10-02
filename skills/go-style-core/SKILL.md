---
name: go-style-core
description: Use when resolving Go style or language mechanics such as formatting, nesting, declarations, initialization, scope, shadowing, loops, switches, or enum zero values. API design, naming, errors, and tests have specialized skills.
---

# Go Style and Language Mechanics

Apply the baseline below; read detailed syntax guidance only for the decision
the task requires. An ordinary function edit does not require every reference.

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). Respect the target
> module's language version; references name version-sensitive features inline.

## Resource Routing

- `references/PRINCIPLES.md` - Read when resolving a tradeoff between clarity, simplicity, concision, maintainability, and consistency.
- `references/FORMATTING.md` - Read for line breaks, whitespace, comments, and semicolon mechanics.
- `references/SCOPE.md` - Read for `var` vs `:=`, grouping declarations, if-init, and reassignment across scopes.
- `references/SHADOWING.md` - Read when an inner declaration hides an outer variable or a predeclared identifier.
- `references/IOTA.md` - Read when designing enum defaults, bitmasks, or grouped constants.
- `references/INITIALIZATION.md` - Read for struct/map initialization, keyed literals (including embedded fields in Go 1.27), zero values, and pointers to optional values.
- `references/CONTROL-FLOW.md` - Read when choosing loop/range forms, writing iterators, or preserving iteration behavior.
- `references/SWITCH-PATTERNS.md` - Read for expression switches, fallthrough, and labeled breaks; route interface semantics to go-interfaces.
- `references/BLANK-IDENTIFIER.md` - Read for intentional discards and side-effect imports; route interface assertions to go-interfaces.

## Style Principles

Resolve readability tradeoffs in this order: clarity, simplicity, concision,
maintainability, consistency. Use the least mechanism that delivers the user's
requirements. These defaults operate within the precedence below.

## House Style Wins

Follow the host's instruction hierarchy. Within it, explicit user requirements
and repository instructions take precedence over these skill defaults. Read
`AGENTS.md`, `CLAUDE.md`, `CONTRIBUTING.md`, `.golangci.yml`, the CI
configuration, and neighboring code before editing. Skills do not authorize
extra work or require renewed approval for work the user already authorized.

- Assertion style, error-wrapping style, logger, and test layout follow the
  nearest existing code. The `_` global prefix is
  [go-naming](../go-naming/SKILL.md#where-review-sends-names-back)'s rule.
- Introduce a *convention* the guide prefers only in new code with no neighbor
  to match, or as a whole-package migration the user asked for. An older Go
  idiom in the neighbor is not a convention:
  [Write Current Go](#write-current-go) outranks it.
- A bug is not house style. Fix it within the authorized scope; report unrelated
  findings separately. A review-only request remains read-only.

## Formatting

Use `gofmt` for Go source. This guide imposes no rigid line-length limit;
break by meaning and readability, while respecting repository requirements.

A comment states what the code cannot show: a constraint, a default
deliberately overridden, the business or historical reason behind a choice,
the clause a branch serves. Code that reads as its
documentation reads carries none; a comment that narrates the next line or
argues that a change is correct is removed before the diff closes. Match the
neighboring code's comment density. Doc comments on exported API belong to
[go-documentation](../go-documentation/SKILL.md).

## Write Current Go

> **Normative**: The module's `go` directive, not the neighboring code, sets
> the idiom. Every line you write or edit uses the current language form and
> standard-library API available at that version, even when the rest of the
> file keeps the older form. Consistency with an older neighbor is not a
> reason to write the older form.

```go
// At go 1.27, the line you write:
return slices.ContainsFunc(items, func(it Item) bool { return it.ID == id })

// The neighbor two functions up stays as it is — not rewritten, not copied:
for i := 0; i < len(items); i++ {
    if items[i].ID == id {
        return true
    }
}
```

The [idiom card](#current-go-idiom-card) below lists the forms. An API not
written before is looked up with `go doc` first, never written from memory:
`maps.Keys` returns an iterator (`iter.Seq`), not the slice the retired
`golang.org/x/exp/maps` returned. Without a shell, the owner skill's example is
the source, and a symbol no skill shows is reported, not invented.

Write the older form only when one of four things is true, and name which in
the report: the current form is not available at the `go` directive
(`COMPATIBILITY.md` lists the language features); it changes observable
behavior (the tiers in
[MODERNIZATION.md](../go-code-refactor/references/MODERNIZATION.md) say which
swaps do); it does not fit the code at hand; or an explicit user or repository
instruction requires the older form ([House Style Wins](#house-style-wins)).
The directive is the `go` line of the module's `go.mod` — language 1.16 when the line is missing; a
`go.work` directive does not raise a member module's version — and neither a
`toolchain` line, `GOTOOLCHAIN`, nor an installed newer Go raises it or
authorizes a version bump; respect build constraints and the CI toolchains.
A language form newer than the directive fails to compile with `requires
go1.NN or later (-lang was set to go1.MM; check go.mod)`; a library symbol
newer than it builds, and only `go vet` (`stdversion`) reports it:
`strings.CutLast requires go1.27 or later (module is go1.26)`.

The rule covers idioms — language features and standard-library APIs — not
dependencies: the logger, assertion library, router, or ORM the package already
uses stays ([House Style Wins](#house-style-wins);
[go-packages](../go-packages/SKILL.md) owns replacing one). Scope stays too:
untouched neighbors are not rewritten for consistency, and the report names
the older forms left in place as an opportunity. A scoped `go fix -diff`
previews what the mechanical modernizers would change, and the
[edit hook](#the-edit-hook-record) prints it after each edit where installed.
[go-linting](../go-linting/SKILL.md) owns the analyzers and verification gate;
[OVER-ENGINEERING.md](../go-code-refactor/references/OVER-ENGINEERING.md#reach-for-what-go-ships)
lists the standard-library replacements beyond the automated modernizers.

## Current Go Idiom Card

> Sources: Go release notes 1.13–1.27; `$GOROOT/api/go1.*.txt`; `go tool fix help`; `go doc encoding/json/v2`; JetBrains go-modern-guidelines `guidelines.json` at `155dc7c` (54 items, cross-checked one by one 2026-09-13). Last verified: 2026-10-01.

One line per idiom: the form an older habit produces, the form the directive
allows, and the trap. A row newer than the module's `go` directive does not
apply; the older rows apply at every directive.
[MODERNIZATION.md](../go-code-refactor/references/MODERNIZATION.md) says
which of these swaps change behavior when *existing* code is rewritten;
[go-data-structures](../go-data-structures/SKILL.md) owns nil against empty.

### In nearly every file

| Habit | Write at the directive |
|---|---|
| `for i := 0; i < n; i++` | `for i := range n` (Go 1.22) — only when the body changes neither `i` nor `n` |
| a loop that searches | `slices.Contains`, `slices.ContainsFunc`, `slices.IndexFunc` (Go 1.21) |
| `err == target`, `err == io.EOF` | `errors.Is(err, target)` (Go 1.13) — `==` sees only the outermost error and misses every `%w` wrap |
| `interface{}` | `any` (Go 1.18) |

### Loops, builtins, defaults

| Habit | Write at the directive |
|---|---|
| `v := v` before a closure or `&v` | nothing (Go 1.22) — loop variables are per-iteration; a pointer to the element itself is `&s[i]` |
| `if a < b { m = a } else { m = b }` | `min(a, b)`, `max(a, b)` (Go 1.21) — for floats, NaN and `-0` behave differently from the hand-written branch |
| `for k := range m { delete(m, k) }` | `clear(m)` (Go 1.21) — on a slice, `clear` zeroes the elements and keeps the length |
| an `if` chain picking the first non-empty value | `cmp.Or(a, b, c)` (Go 1.22) — every argument is evaluated; wrong when `0` or `false` is a real value rather than a missing one |
| `tmp := f(); p := &tmp` | `new(f())` (Go 1.26) — `new(30)` is `*int`; a `*time.Duration` field takes `new(30 * time.Second)` |
| `Outer{Inner: Inner{Field: v}}` | `Outer{Field: v}` (Go 1.27) — promoted fields in a keyed literal; not through a pointer embed, not mixed with the embedded field itself |

### Collections

| Habit | Write at the directive |
|---|---|
| collect the keys, then `sort.Strings` | `slices.Sorted(maps.Keys(m))` (Go 1.23) — nil for an empty map, which `encoding/json/v2` writes as `[]` and v1 as `null`; under v1, when `[]` must encode, `slices.AppendSeq(make([]K, 0, len(m)), maps.Keys(m))` then `slices.Sort` |
| `sort.Slice(s, func(i, j int) bool {...})` | `slices.SortFunc(s, func(a, b T) int { return cmp.Compare(a.Key, b.Key) })` (Go 1.21) — the comparator returns `int`, not `bool`; both are unstable, and `sort.SliceStable` becomes `slices.SortStableFunc` |
| `sort.Ints`, `sort.Strings` | `slices.Sort` (Go 1.21) |
| a shallow copy loop, `append([]T(nil), s...)` | `slices.Clone`, `maps.Clone` (Go 1.21) — preserve nilness, capacity when observable, and nested aliasing; under `encoding/json` v1, `make` plus `copy` when `[]` must encode |
| index, max, min, reverse, dedupe, filter loops | `slices.Index`, `slices.Max`, `slices.Min`, `slices.Reverse`, `slices.Compact`, `slices.DeleteFunc` (Go 1.21) — `Compact` drops adjacent duplicates only, so sort first |
| a max or min loop that returns zero for no input | `if len(s) == 0 { return 0 }` before `slices.Max(s)` or `slices.Min(s)` — both panic on an empty slice |
| a reverse, dedupe, or filter loop that builds a new slice | `slices.Reverse`, `slices.Compact`, and `slices.DeleteFunc` edit in place, so they run on `slices.Clone(s)` when `s` is kept |
| merging or filtering a map by hand | `maps.Copy`, `maps.DeleteFunc` (Go 1.21) |
| `s = s[:len(s):len(s)]` | `slices.Clip` (Go 1.21) |
| a slice built only to be ranged over | a function returning `iter.Seq[T]` (Go 1.23); `slices.Collect` when a slice is needed after all |

### Strings and bytes

| Habit | Write at the directive |
|---|---|
| `strings.Index` plus slicing | `strings.Cut` (Go 1.18); `strings.CutPrefix`, `strings.CutSuffix` (Go 1.20); `strings.CutLast` (Go 1.27) — `bytes` has the same four |
| `strings.HasPrefix` followed by `strings.TrimPrefix` | `strings.CutPrefix` (Go 1.20) — one check, both results |
| `strings.Split` followed by `range` | `strings.SplitSeq`, `strings.FieldsSeq` (Go 1.24) — only when the slice is not indexed or kept; `strings.Lines` keeps each `\n` |
| a substring kept from a large input; `append([]byte(nil), b...)` | `strings.Clone` (Go 1.18), `bytes.Clone` (Go 1.20) |
| `buf = append(buf, fmt.Sprintf(...)...)` | `buf = fmt.Appendf(buf, ...)` (Go 1.19) — no intermediate string |
| `+=` on a string in a loop | `strings.Builder` |

### Errors

| Habit | Write at the directive |
|---|---|
| a slice of errors joined as text | `errors.Join(errs...)` (Go 1.20) — `errors.Is` and `errors.As` still match through it |
| `var target *T` then `errors.As(err, &target)` | `errors.AsType[*T](err)` (Go 1.26) — returns the value; replaces `As`, never `Is` |

### Context, sync, time

| Habit | Write at the directive |
|---|---|
| `wg.Add(1)` and `go func() { defer wg.Done() }()` | `wg.Go(f)` (Go 1.25) |
| a goroutine that interrupts a blocking call while this function runs | `stop := context.AfterFunc(ctx, f)` and `defer stop()` (Go 1.21) — without `stop`, each registration on a long-lived context stays live until the context ends |
| a goroutine parked on `ctx.Done()` to run cleanup whenever the context ends, after this function returns | `context.AfterFunc(ctx, f)` (Go 1.21), its `stop` kept by the owner and called from its `Close` — a `defer stop()` here deregisters the cleanup at return, and it never runs |
| `cancel()` that loses the reason | `context.WithCancelCause`, `context.Cause` (Go 1.20); `context.WithTimeoutCause`, `context.WithDeadlineCause` (Go 1.21) |
| `sync.Once` plus a result field and a getter | `sync.OnceFunc`, `sync.OnceValue`, `sync.OnceValues` (Go 1.21) |
| an `int32` flag with `atomic.LoadInt32`; `unsafe.Pointer` with `atomic.StorePointer` | `atomic.Bool`, `atomic.Int64`, `atomic.Pointer[T]` (Go 1.19) — in an existing struct, the new field makes every by-value copy of that struct a `go vet` copylocks finding; `atomic.Value` differs from `atomic.Pointer[T]` in nil handling |
| a mutex around one counter or flag | `atomic.Int64`, `atomic.Bool` (Go 1.19) — only with no compound invariant |
| `time.Now().Sub(start)`, `deadline.Sub(time.Now())` | `time.Since(start)`; `time.Until(deadline)` (Go 1.8) |
| `time.NewTicker` with `defer t.Stop()` in the process's own main loop, run by `main` itself | `for range time.Tick(d)` (Go 1.23) — an unreferenced ticker is collected since 1.23, so the old warning against `time.Tick` is over there; `Tick` returns nil for `d <= 0`, so the loop blocks forever where `NewTicker` panicked: a configured `d` is checked `> 0` first; a spawned goroutine keeps `time.NewTicker` plus `select` on its stop channel or `ctx.Done()`, as does a loop that calls `Stop` or `Reset` |

### Types, HTTP, JSON

| Habit | Write at the directive |
|---|---|
| `reflect.TypeOf((*T)(nil)).Elem()` | `reflect.TypeFor[T]()` (Go 1.22) |
| a package-level generic helper that belongs to one type | a generic method on the type (Go 1.27) — it cannot satisfy an interface |
| a new router dependency; `strings.TrimPrefix(r.URL.Path, "/users/")` | `mux.HandleFunc("GET /users/{id}", h)` and `r.PathValue("id")` (Go 1.22) — a package already on a router module keeps it |
| `omitempty` on a struct, bool, number or `time.Time` field | `omitzero` (Go 1.24) — `omitempty` stays for strings, slices and maps; on an existing field the wire changes |
| a new UUID dependency for creating and parsing | the standard `uuid` package (Go 1.27) — a package already on a UUID module keeps it until a migration is asked for |
| `math/rand` with `rand.Seed(1)` for a fixed sequence; `rand.Intn(n)` | `import "math/rand/v2"`: `rand.IntN(n)`, `rand.N(d)` (Go 1.22); a fixed sequence is its own generator, `rand.New(rand.NewPCG(1, 2))` — top-level `rand.Seed` is a no-op at a `go 1.24` or later directive, so a seeded run comes out random |
| `*u` copied by hand; `url.Values` copied in a loop | `u.Clone()`, `v.Clone()` (Go 1.27) — only when the old copy's depth matches; shallow-to-deep is a semantic fix, not a refactor ([MODERNIZATION.md](../go-code-refactor/references/MODERNIZATION.md#urlurlclone--urlvaluesclone--go-127)) |
| `encoding/json` in a package with no JSON yet; `json.NewEncoder(w).Encode(v)` | `import json "encoding/json/v2"` (Go 1.27): `json.Marshal`, `json.MarshalWrite(w, v)`, `json.UnmarshalRead(r, &v)` — nil slices encode as `[]` and nil maps as `{}`, so no `make` for the wire; duplicate names are rejected; `MarshalWrite` adds no newline; `Marshal` fails where v1 did not — invalid UTF-8 in any string, a `time.Duration` field (no default representation) — and returns partial bytes with the error, so the error is checked before anything is written; a package already on `encoding/json` stays on it until a migration is asked for |

### Tests and benchmarks

| Habit | Write at the directive |
|---|---|
| `context.Background()` in a test | `t.Context()` (Go 1.24) — cancelled when the test returns, before `t.Cleanup` runs; cleanup that needs a live context uses `context.WithoutCancel(t.Context())` (`usetesting` reports `context.Background()` there) |
| `for i := 0; i < b.N; i++` | `for b.Loop()` (Go 1.24) — `b.N` is unknown until the loop ends, so a fixture sized by `b.N` is restructured, not translated |
| `os.Chdir` with a restore in cleanup; a hand-made temp dir | `t.Chdir` (Go 1.24); `t.TempDir` (Go 1.15) |
| `time.Sleep` to let goroutines settle | `synctest.Test` and `synctest.Wait` (Go 1.25) |
| `httptest.NewServer` with `defer srv.Close()` | `httptest.NewTestServer(t, h)` (Go 1.27) — in-memory, closed by `t.Cleanup`, and reached only through `srv.Client()`: `srv.URL` is empty until that call and `http://example.com` after it, so any other client reaches the real host; code that dials `srv.URL` with its own client calls `srv.Start()` first (loopback; the cleanup stays) |

## Reduce Nesting

Handle errors and special conditions first. Return early or continue the loop
so the success path stays unindented. Preserve the order of validation, side
effects, and returned errors when flattening existing code.
Negate the original predicate exactly: for floating-point values, `!(x > 0)`
also rejects NaN, while `x <= 0` does not. Preserve short-circuit evaluation.

### Unnecessary Else

Omit `else` after a branch that exits. For two branches assigning one value,
use default plus override when the default is safe to evaluate unconditionally;
keep the branches when evaluation has side effects or is expensive.

## How Much To Say

Follow the host's communication requirements. Give a short initial update,
then progress updates on what changed the work: a finding, a decision, a
blocker, the next check; a read that changed nothing is not an update. Close
with the outcome, observed verification results, and material limitations;
never imply a skipped check ran. Size reports, reviews, and design notes to
the task; use applicable `assets/` templates without filler sections or
repeated summaries.

Keep routine edits and checks inline. When the user or host authorizes parallel
work, delegate only bounded, independent tasks with clear ownership and useful
work remaining locally; do not spawn a second agent merely to repeat a
completed check.

### The Edit Hook Record

Where the Claude Code plugin hook runs, it checks each edit of a `.go` file
and records `pass`, `fail`, `skipped` or `unavailable` per check; it never blocks
an edit that already happened. It runs `gofmt -l` on
the file and `go vet`, `go fix -diff`, `go test -short` (no `-race`), and
`golangci-lint` on the file's package: vet and test output covers the whole
package, fix hunks and lint findings in other files of the package arrive as a
count, and in a git checkout lint counts only issues new since HEAD. Immutable
JSON receipts include command, scope, config, tool version, exit code,
timestamps, diagnostics and input digests before/after the check. Success
uses PostToolUse context; findings include the receipt directory on stderr.
Missing tools, disabled checks, no tests and no module have explicit statuses.
A finding in code the diff touched is fixed before the next step, not reported
around; one in code
the diff did not touch is pre-existing: report it, do not fix it. Without a
shell tool, its observed receipts are the available check evidence.
[go-linting](../go-linting/SKILL.md#verification-gate) owns reuse and scope:
with a shell, load that skill and execute the matching Bash verification
command printed by the hook after the final edit. Its batch `--gate` result
lists credited checks and required direct checks; the Bash workflow guard
requires an attempt for the current edit generation. Only `hook_credit=true`
permits `pass (hook)` for the verified scope. Unverified hook results are
observed evidence; silence is never pass. A race check needs separate evidence.

## Related Skills

- [go-functions](../go-functions/SKILL.md): signatures, constructors, config structs, functional options.
- [go-naming](../go-naming/SKILL.md): identifiers and receiver names.
- [go-error-handling](../go-error-handling/SKILL.md): error strategy, wrapping, log-vs-return.
- [go-interfaces](../go-interfaces/SKILL.md): type assertions, type switches, compile-time checks.
- [go-data-structures](../go-data-structures/SKILL.md): slices and maps; [go-performance](../go-performance/SKILL.md): capacity hints.
- [go-documentation](../go-documentation/SKILL.md): exported API comments and examples.
- [go-linting](../go-linting/SKILL.md): the shared gate and CI configuration.
- [go-code-review](../go-code-review/SKILL.md): systematic review; [go-code-refactor](../go-code-refactor/SKILL.md): behavior-preserving restructuring.
