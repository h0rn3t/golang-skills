# Current Go Idiom Card

> Sources: Go release notes 1.13–1.27; `$GOROOT/api/go1.*.txt`; `go tool fix help`; JetBrains go-modern-guidelines `guidelines.json` at `155dc7c` (54 items with frequency ranks, cross-checked one by one 2026-09-13)
> Authority: normative — the forms [Write Current Go](../SKILL.md#write-current-go) requires at the module's `go` directive; project policy for the grouping and order
> Minimum Go: each row names its own; a row newer than the module's `go` directive does not apply
> Last verified: 2026-09-13

One line per idiom: the form an older habit produces, the form the directive
allows, and the trap on the same line. Read the whole card before the first
edit of a task that writes Go — the older rows apply at every directive, so a
`head` or a `grep` over it misses what a Go 1.19 module still gets. The card
says what to write; [MODERNIZATION.md](../../go-code-refactor/references/MODERNIZATION.md)
says which of these swaps change behavior when *existing* code is rewritten,
and [go-data-structures](../../go-data-structures/SKILL.md) owns nil against
empty for every collection row.

## In nearly every file

| Habit | Write at the directive |
|---|---|
| `for i := 0; i < n; i++` | `for i := range n` (Go 1.22) — only when the body changes neither `i` nor `n` |
| a loop that searches | `slices.Contains`, `slices.ContainsFunc`, `slices.IndexFunc` (Go 1.21) |
| `err == target`, `err == io.EOF` | `errors.Is(err, target)` (Go 1.13) — `==` sees only the outermost error and misses every `%w` wrap |
| `interface{}` | `any` (Go 1.18) |

## Loops, builtins, defaults

| Habit | Write at the directive |
|---|---|
| `v := v` before a closure or `&v` | nothing (Go 1.22) — loop variables are per-iteration; a pointer to the element itself is `&s[i]` |
| `if a < b { m = a } else { m = b }` | `min(a, b)`, `max(a, b)` (Go 1.21) — for floats, NaN and `-0` behave differently from the hand-written branch |
| `for k := range m { delete(m, k) }` | `clear(m)` (Go 1.21) — on a slice, `clear` zeroes the elements and keeps the length |
| an `if` chain picking the first non-empty value | `cmp.Or(a, b, c)` (Go 1.22) — every argument is evaluated; wrong when `0` or `false` is a real value rather than a missing one |
| `tmp := f(); p := &tmp` | `new(f())` (Go 1.26) — `new(30)` is `*int`; a `*time.Duration` field takes `new(30 * time.Second)` |
| `Outer{Inner: Inner{Field: v}}` | `Outer{Field: v}` (Go 1.27) — promoted fields in a keyed literal; not through a pointer embed, not mixed with the embedded field itself |

## Collections

| Habit | Write at the directive |
|---|---|
| collect the keys, then `sort.Strings` | `slices.Sorted(maps.Keys(m))` (Go 1.23) — nil for an empty map; when `[]` must encode, `slices.AppendSeq(make([]K, 0, len(m)), maps.Keys(m))` then `slices.Sort` |
| `sort.Slice(s, func(i, j int) bool {...})` | `slices.SortFunc(s, func(a, b T) int { return cmp.Compare(a.Key, b.Key) })` (Go 1.21) — the comparator returns `int`, not `bool`; both are unstable, and `sort.SliceStable` becomes `slices.SortStableFunc` |
| `sort.Ints`, `sort.Strings` | `slices.Sort` (Go 1.21) |
| a copy loop, `append([]T(nil), s...)` | `slices.Clone`, `maps.Clone` (Go 1.21) — nil stays nil; `make` plus `copy` when `[]` must encode |
| index, max, min, reverse, dedupe loops | `slices.Index`, `slices.Max`, `slices.Min`, `slices.Reverse`, `slices.Compact` (Go 1.21) — `Compact` drops adjacent duplicates only, so sort first |
| merging or filtering a map by hand | `maps.Copy`, `maps.DeleteFunc` (Go 1.21) |
| `s = s[:len(s):len(s)]` | `slices.Clip` (Go 1.21) |
| a slice built only to be ranged over | a function returning `iter.Seq[T]` (Go 1.23); `slices.Collect` when a slice is needed after all |

## Strings and bytes

| Habit | Write at the directive |
|---|---|
| `strings.Index` plus slicing | `strings.Cut` (Go 1.18); `strings.CutPrefix`, `strings.CutSuffix` (Go 1.20); `strings.CutLast` (Go 1.27) — `bytes` has the same four |
| `strings.HasPrefix` followed by `strings.TrimPrefix` | `strings.CutPrefix` (Go 1.20) — one check, both results |
| `strings.Split` followed by `range` | `strings.SplitSeq`, `strings.FieldsSeq` (Go 1.24) — only when the slice is not indexed or kept; `strings.Lines` keeps each `\n` |
| a substring kept from a large input; `append([]byte(nil), b...)` | `strings.Clone` (Go 1.18), `bytes.Clone` (Go 1.20) |
| `buf = append(buf, fmt.Sprintf(...)...)` | `buf = fmt.Appendf(buf, ...)` (Go 1.19) — no intermediate string |
| `+=` on a string in a loop | `strings.Builder` |

## Errors

| Habit | Write at the directive |
|---|---|
| a slice of errors joined as text | `errors.Join(errs...)` (Go 1.20) — `errors.Is` and `errors.As` still match through it |
| `var target *T` then `errors.As(err, &target)` | `errors.AsType[*T](err)` (Go 1.26) — returns the value; replaces `As`, never `Is` |

## Context, sync, time

| Habit | Write at the directive |
|---|---|
| `wg.Add(1)` and `go func() { defer wg.Done() }()` | `wg.Go(f)` (Go 1.25) |
| a goroutine parked on `ctx.Done()` to run cleanup | `context.AfterFunc(ctx, f)` (Go 1.21) — returns a stop function |
| `cancel()` that loses the reason | `context.WithCancelCause`, `context.Cause` (Go 1.20); `context.WithTimeoutCause`, `context.WithDeadlineCause` (Go 1.21) |
| `sync.Once` plus a result field and a getter | `sync.OnceFunc`, `sync.OnceValue`, `sync.OnceValues` (Go 1.21) |
| an `int32` flag with `atomic.LoadInt32`; `unsafe.Pointer` with `atomic.StorePointer` | `atomic.Bool`, `atomic.Int64`, `atomic.Pointer[T]` (Go 1.19) — in an existing struct, swapping the field type changes its size and layout; `atomic.Value` differs from `atomic.Pointer[T]` in nil handling |
| a mutex around one counter or flag | `atomic.Int64`, `atomic.Bool` (Go 1.19) — only with no compound invariant |
| `time.Now().Sub(start)`, `deadline.Sub(time.Now())` | `time.Since(start)`, `time.Until(deadline)` (Go 1.8) |
| `time.NewTicker` with `defer t.Stop()` in a loop that runs for the life of the process | `for range time.Tick(d)` (Go 1.23) — an unreferenced ticker is collected since 1.23, so the old warning against `time.Tick` is over; `time.NewTicker` stays where `Stop` or `Reset` is called |

## Types, HTTP, JSON

| Habit | Write at the directive |
|---|---|
| `reflect.TypeOf((*T)(nil)).Elem()` | `reflect.TypeFor[T]()` (Go 1.22) |
| a package-level generic helper that belongs to one type | a generic method on the type (Go 1.27) — it cannot satisfy an interface |
| a router module; `strings.TrimPrefix(r.URL.Path, "/users/")` | `mux.HandleFunc("GET /users/{id}", h)` and `r.PathValue("id")` (Go 1.22) |
| `omitempty` on a struct, bool, number or `time.Time` field | `omitzero` (Go 1.24) — `omitempty` stays for strings, slices and maps; on an existing field the wire changes |
| a UUID module for creating and parsing | the standard `uuid` package (Go 1.27) |
| `*u` copied by hand; `url.Values` copied in a loop | `u.Clone()`, `v.Clone()` (Go 1.27) |
| `encoding/json` in new JSON code | `encoding/json/v2` (Go 1.27) — nil slices encode as `[]`, duplicate keys are rejected; existing `encoding/json` code stays until a migration is asked for |

## Tests and benchmarks

| Habit | Write at the directive |
|---|---|
| `context.Background()` in a test | `t.Context()` (Go 1.24) — cancelled when the test returns, before `t.Cleanup` runs; cleanup that needs a live context makes its own |
| `for i := 0; i < b.N; i++` | `for b.Loop()` (Go 1.24) — `b.N` is unknown until the loop ends, so a fixture sized by `b.N` is restructured, not translated |
| `os.Chdir` with a restore in cleanup; a hand-made temp dir | `t.Chdir`, `t.TempDir` (Go 1.24) |
| `time.Sleep` to let goroutines settle | `synctest.Test` and `synctest.Wait` (Go 1.25) |
| `httptest.NewServer` with `defer srv.Close()` | `httptest.NewTestServer(t, h)` (Go 1.27) — an in-memory server reached only through `srv.Client()` |

## An API you have not written before

`go doc` is the signature: `go doc strings.CutLast`, `go doc sync.OnceValue`,
`go doc errors.AsType` before the line is written, never from memory —
`maps.Keys` returns an iterator (`iter.Seq`), not the slice the retired
`golang.org/x/exp/maps` returned; a `slices.SortFunc` comparator returns
`int`; `strings.Lines` keeps the newline. Without a shell, the owner skill's
example is the source, and a symbol no skill shows is reported, not invented.
`go vet` runs `stdversion`, which reports a standard-library symbol newer than
the directive; `go fix -diff` shows the rewrites the analyzers know
([go-linting](../../go-linting/SKILL.md#modernization-go-fix)).

## Which version governs

The `go` directive of the module's `go.mod`, read before the first edit. A
`go.mod` with no `go` line is language version 1.16; a `go.work` has its own
directive for the workspace, and it does not raise a member module's language
version. A `toolchain` line, `GOTOOLCHAIN`, and the Go installed on the machine
choose the compiler, not the language version, so none of them makes a newer
form legal: the compiler refuses one with `requires go1.NN or later (-lang was
set to go1.MM; check go.mod)`, and `go vet`'s `stdversion` reports a newer
library symbol. Bumping the directive is its own change
([MODERNIZATION.md](../../go-code-refactor/references/MODERNIZATION.md#tier-3--report-dont-apply)).
