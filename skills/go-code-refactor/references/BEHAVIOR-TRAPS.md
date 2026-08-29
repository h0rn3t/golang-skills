# Behavior Traps

> Sources: Go spec; `encoding/json`, `sync`, `slices` package docs; Go release notes 1.22–1.27
> Authority: normative
> Minimum Go: 1.27 baseline; per-item versions noted inline
> Last verified: 2026-08-29

Rewrites that look like pure cleanup and quietly change what the program does.
These are why a Go refactor needs verification rather than confidence.

## Contents

- [nil vs empty slices and maps](#nil-vs-empty)
- [defer](#defer)
- [Error handling](#error-handling)
- [Interfaces and typed nil](#interfaces-and-typed-nil)
- [Concurrency](#concurrency)
- [Slices sharing backing arrays](#slices-sharing-backing-arrays)
- [Loops and closures](#loops-and-closures)
- [Numbers and conversions](#numbers-and-conversions)
- [Receivers and method sets](#receivers-and-method-sets)
- [Struct layout and tags](#struct-layout-and-tags)
- [Evaluation order](#evaluation-order)
- [Maps and randomness](#maps-and-randomness)
- [Modernization traps](#modernization-traps)
- [Pre-commit checklist](#pre-commit-checklist)

---

## nil vs empty slices and maps

`var s []string` and `s := []string{}` behave identically for `len`, `append`,
and `range` — but not on the wire:

```go
json.Marshal(struct{ Items []string }{nil})        // {"Items":null}
json.Marshal(struct{ Items []string }{[]string{}}) // {"Items":[]}
```

Clients parse those differently, and `s == nil` is a check some code makes
deliberately. Never normalize one into the other in either direction, and
preserve whether a function returns nil or empty on the error path.

Same for maps: reading a nil map is fine, writing panics. Adding a defensive
`make(map...)` where the original had nil turns a panic into silent success — a
behavior change wearing a bug fix's clothes.

`slices.Clone(nil)` and `maps.Clone(nil)` return nil, so swapping a
`make`+`copy` for `Clone` preserves nil-ness. See
[go-defensive](../../go-defensive/SKILL.md).

## defer

**Arguments evaluate immediately; the call is deferred.**

```go
defer log.Printf("took %v", time.Since(start))              // runs NOW
defer func() { log.Printf("took %v", time.Since(start)) }() // runs at return
```

Converting between these forms changes the logged value.

**Order is LIFO.** Reordering defers reorders the operations — mutex release
relative to a channel close, for instance.

**Deferred closures see named results.** Moving returns from named to unnamed
while a defer mutates the result changes what the caller receives:

```go
func f() (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic: %v", r)
        }
    }()
```

**`defer` in a loop accumulates.** Hoisting a loop body into its own function
makes deferred calls fire per iteration instead of at function exit. That is
usually the fix someone wants, but it *is* a behavior change: handles now close
earlier. Flag it rather than doing it silently.

## Error handling

- Error message text is API. Callers grep logs, tests assert on it, alerts
  match it. Do not rephrase, capitalize, or punctuate.
- `%v` and `%w` are not interchangeable: `%w` keeps the chain for `errors.Is`
  and `errors.AsType`. Converting `%v` to `%w` makes previously-failing
  `errors.Is` checks start succeeding.
- Consolidating several distinct error returns into one shared error changes
  what `errors.Is` can distinguish.
- `errors.New` at package level is a comparable sentinel; moving it inside a
  function creates a new value per call and breaks `==`.
- A dropped `err` (`_ = f()`) may be load-bearing at the call site. Adding a
  check where none existed changes control flow — record it as a finding.

## Interfaces and typed nil

```go
var p *MyError = nil
var err error = p
err != nil // true — the interface holds a type
```

Changing a return type from `*MyError` to `error` — a common "cleanup" — flips
`err != nil` at every call site. Narrowing a concrete return to an interface
removes methods callers may use; widening an interface parameter to a concrete
type breaks callers passing other implementations.

## Concurrency

Almost nothing here is safe to tidy:

- Channel buffer size changes blocking behavior and can turn a deadlock into a
  leak, or the reverse.
- Moving `mu.Lock()` changes the critical section; reordering two acquisitions
  can introduce deadlock.
- `wg.Add` must happen before the goroutine starts. Extracting the goroutine
  into a helper often moves `Add` inside — a race. `wg.Go` (Go 1.25+) makes
  this structurally impossible, which is why the swap is worth doing.
- Replacing goroutine + channel with `errgroup` changes cancellation semantics
  and which error surfaces first.
- `sync.Mutex` → `sync.RWMutex` is not a no-op if any "read" path mutates.
- Timeouts, tickers, and `context.WithTimeout` durations are observable.

If concurrency is touched, verify with `go test -race` and say so explicitly in
the report. See [go-concurrency](../../go-concurrency/SKILL.md).

## Slices sharing backing arrays

`a := b[:n]` shares memory with `b`. Replacing it with a copy — or the reverse —
changes whether writes propagate:

```go
s = append(s[:i], s[i+1:]...) // mutates the original backing array
```

`append` may or may not reallocate depending on capacity. Code relying on
aliasing after append is fragile, but changing it is still a behavior change.
Preserve `make([]T, 0, n)` capacity hints: they move the reallocation point and
therefore the aliasing.

## Loops and closures

Loop variables are per-iteration since Go 1.22. Code written for the old
semantics contains `x := x` shadowing; deleting it is safe when the `go`
directive is ≥1.22 and breaks below that — check `go.mod` first.

Converting `for i := 0; i < len(s); i++` to `for i, v := range s` copies each
element into `v`. If the body writes through `s[i]`, or the elements are large
structs, this is not equivalent.

`range` over a slice evaluates `len` once; the three-clause loop re-evaluates
it. If the body appends, the two loops iterate different numbers of times.

## Numbers and conversions

- `int`, `int64`, and `int32` differ in overflow points. Overflow is defined in
  Go, so code may rely on wrapping.
- `float64` arithmetic is not associative — reordering `a + b + c` can change
  the last bits. Since Go 1.25 this is sharper: at `GOAMD64=v3` and above the
  compiler fuses `a*b + c` into one FMA, so results can move across a toolchain
  bump with no diff from you. `float64(a*b) + c` prevents fusing.
- Integer division truncates toward zero; `a/b*c` and `a*c/b` differ.
- `strconv.FormatFloat` precision and `%v` vs `%g` produce different strings.

## Receivers and method sets

Changing a receiver from value to pointer means the method mutates the original
and the type no longer satisfies interfaces when used as a value. Pointer to
value silently drops mutations. Both often compile. **Never change receiver
kind during a readability pass.**

## Struct layout and tags

- Struct tags (`json`, `db`, `yaml`, `validate`) are behavior. Reformatting is
  fine; renaming a key, adding `omitempty`, or dropping `,string` is not.
- Field order affects `encoding/json` output order, `encoding/binary`, `unsafe`
  offsets, and unkeyed struct literals.
- Embedding vs a named field changes the method set and JSON flattening.
- Adding or removing a field changes `==` comparability and `reflect.DeepEqual`.

## Evaluation order

```go
if ok && expensive() { } // expensive() may not run
if a() && b() { }        // reordering changes which side effects fire
```

Extracting subexpressions into variables above the `if` forces evaluation of
both — which matters when one panics on nil or has side effects. This is the
most common accidental breakage in "flatten the condition" refactors.

## Maps and randomness

Map iteration order is randomized. Code that appears to depend on it is already
broken — but "fixing" it by adding a sort changes output downstream systems may
have adapted to. Record it as a finding.

## Modernization traps

Adopting a newer API is a rewrite wearing a cleanup's costume. The ones that
bite most often:

| Swap | Verdict |
|---|---|
| `errors.As` → `errors.AsType[T]` | Equivalent |
| `%v` → `%w` | **Not** equivalent — changes `errors.Is` |
| `omitempty` → `omitzero` | Changes which fields appear in JSON |
| `fmt.Sprintf("%s:%d", host, port)` → `net.JoinHostPort` | Differs for IPv6 — a bug fix, not a swap |
| `os.Open` → `os.Root` | Starts rejecting paths that escape the directory |
| `strings.Split` → `strings.SplitSeq` | Equivalent |
| `strings.Split` → `strings.Lines` | **Not** — `Lines` keeps the trailing newline |
| `sort.Slice` → `slices.SortFunc` | Keeps instability |
| `sort.Slice` → `slices.SortStableFunc` | Changes order of equal keys |
| `runtime.SetFinalizer` → `runtime.AddCleanup` | Changes when cleanup runs and how cycles behave |
| Bumping the `go` directive | New vet diagnostics **and** new semantics — its own change |

`references/MODERNIZATION.md` has the full catalog, including toolchain
releases that break green tests with no diff from you.

---

## Pre-commit checklist

Before calling a step behavior-preserving, confirm none of these moved:

exported signatures · error text and `%w` chains · nil-vs-empty returns ·
defer order and arguments · channel buffers and lock scopes · struct tags and
field order · numeric types · receiver kinds · goroutine counts · timeout
durations · JSON field presence.
