# Modernization Catalog

> Sources: `$GOROOT/api/go1.2*.txt`; `go tool fix help`; Go spec; package docs
> Authority: normative for tier placement; project policy for what may ride in a refactor
> Minimum Go: gated by the `go` directive in `go.mod`, not the installed toolchain
> Last verified: 2026-08-29 against go1.27.0

Every entry below was checked against an installed toolchain. When you extend
this file, check yours the same way — `COMPATIBILITY.md` documents how. An
unverified modernization claim is worse than none, because it rides into a diff
that promised not to change behavior.

## Contents

- [Start with `go fix`](#start-with-go-fix)
- [Tier 1 — safe swaps](#tier-1--safe-swaps)
- [Tier 2 — safe with a condition](#tier-2--safe-with-a-condition)
- [Tier 3 — report, don't apply](#tier-3--report-dont-apply)
- [Toolchain shifts that break tests on their own](#toolchain-shifts-that-break-tests-on-their-own)
- [Tools worth running once](#tools-worth-running-once)

---

## Start with `go fix`

Since Go 1.26, `go fix` hosts the *modernizers*: analyzers that rewrite code to
current idioms, built on the same framework as `go vet` and designed not to
change behavior.

```bash
go fix -diff ./...   # inspect first
go fix ./...         # then apply
```

Run it before hand-editing. Whatever `go fix` handles is mechanical,
reviewable, and attributable to a tool; your hand-written changes should be the
ones that needed judgment. Mixing both into one unexplained blob is what makes
reviewers distrust a refactor.

`go tool fix help` prints the set your toolchain actually has — trust that over
any list, including this one. If a fixer produces something wrong, say so in
the report instead of quietly reverting it.
[go-linting](../../go-linting/SKILL.md) catalogues the current analyzers.

---

## Tier 1 — safe swaps

Each removes ceremony that exists only because the language lacked the feature.

### `new(expr)` — Go 1.26

```go
// before
age := yearsSince(born)
p := Person{Name: name, Age: &age}

// after
p := Person{Name: name, Age: new(yearsSince(born))}
```

Biggest payoff in code building JSON or protobuf structs with `*int`/`*bool`
optional fields, where the temporaries outnumber the logic. `go fix -newexpr`.

### `errors.AsType[T]` — Go 1.26

Generic `errors.As` with the same matching semantics and no out-parameter:

```go
// before
var perr *fs.PathError
if errors.As(err, &perr) { log.Println(perr.Path) }

// after
if perr, ok := errors.AsType[*fs.PathError](err); ok { log.Println(perr.Path) }
```

`go fix -errorsastype`. Note this replaces `As`, never `Is`.

### `sync.WaitGroup.Go` — Go 1.25

```go
// before
wg.Add(1)
go func() { defer wg.Done(); work(item) }()

// after
wg.Go(func() { work(item) })
```

Same goroutine count, same `Add`-before-start. Worth extra attention: the old
form is where the classic "`Add` inside the goroutine" race lives. If you find
that bug, the swap fixes it as a side effect — which makes it Tier 3 for that
call site. Say so rather than letting a race fix ride along unannounced.
`go fix -waitgroupgo`.

### `strings.CutLast` / `bytes.CutLast` — Go 1.27

```go
// before
i := strings.LastIndex(name, ".")
if i < 0 { return name, "" }
return name[:i], name[i+1:]

// after
base, ext, _ := strings.CutLast(name, ".")
```

Same contract as `Cut`: not found → `(s, "", false)`. Check what the old `-1`
branch returned before folding it into `ok` — if it returned anything other
than `(s, "")`, the swap needs an explicit `if !ok`.

### Drop now-redundant type arguments — Go 1.27

Inference now applies where a generic function is assigned or converted to a
matching function type, so instantiations added to satisfy the old compiler go:

```go
var fold func(int, int) int = combine[int] // before
var fold func(int, int) int = combine      // after
```

### `url.URL.Clone` / `url.Values.Clone` — Go 1.27

Replaces hand-rolled copying. Same caveat as `wg.Go`: if the old copy was
shallow (`u2 := *u`, or a `maps.Clone`d `Values` whose slices stay shared), the
hand-rolled version was a latent bug and the swap is a fix — Tier 3 for that
call site. See [go-defensive](../../go-defensive/SKILL.md).

### Iterator forms of splitting — Go 1.24

```go
for line := range strings.SplitSeq(text, "\n") { ... }
```

Also `FieldsSeq`, `FieldsFuncSeq`, and the `bytes` equivalents: same elements in
the same order, no intermediate slice. Only when the slice is not indexed,
re-ranged, or kept — otherwise it is a rewrite, not a swap.

`strings.Lines` is **not** a drop-in: it yields lines *with* their trailing
newline, unlike `strings.Split(s, "\n")`.

### `slices` and `maps` over hand-rolled loops — Go 1.21–1.23

```go
// before
keys := make([]string, 0, len(m))
for k := range m { keys = append(keys, k) }
sort.Strings(keys)

// after
keys := slices.Sorted(maps.Keys(m))
```

Likewise `slices.Contains`, `Index`, `Reverse`, `Collect`, `Max`/`Min`,
`Clone`. **Watch the sort**: `sort.Slice` is unstable, `slices.SortFunc` is
unstable, `slices.SortStableFunc` is stable — match what the original used,
because for equal keys the output order is observable.

### `for i := range n` — Go 1.22

Only when the body does not mutate `i` or `n`. `range n` evaluates `n` once;
the three-clause loop re-evaluates it each iteration. `go fix -rangeint`.

### Delete `x := x` loop-variable shadowing — Go 1.22

Per-iteration loop variables make the copy redundant. Safe once the `go`
directive says 1.22+; a mechanical delete under an older directive is a real
bug. `go fix -forvar`.

### `min`, `max`, `clear` — Go 1.21

Replace hand-written helpers and `for k := range m { delete(m, k) }`. Note
`clear` on a *slice* zeroes elements rather than truncating — it is not
`s = s[:0]`. `go fix -minmax`.

### `cmp.Or` for fallback chains — Go 1.22

```go
name := cmp.Or(input, defaultName)
```

`cmp.Or` evaluates all arguments — not for expensive or side-effecting
fallbacks.

### `errors.Join` — Go 1.21

Replaces accumulating errors into a string. It **changes the error's text**, so
this is Tier 1 only when the aggregate error is newly constructed or its text
is provably unobserved.

### Test-only conveniences — Go 1.24–1.27

`t.Context()`, `t.Chdir()`, `slog.DiscardHandler`, `b.Loop()`, and — Go 1.27 —
`synctest.Sleep` and `httptest.NewTestServer`. Test code has no production
observers, so the bar is lower; a changed test is still worth mentioning. See
[go-testing](../../go-testing/SKILL.md).

---

## Tier 2 — safe with a condition

- **`net.JoinHostPort` over `fmt.Sprintf("%s:%d", host, port)`.** Identical for
  IPv4 and hostnames; for IPv6 the old form produced an unusable address, so if
  IPv6 can reach this code the swap is a **bug fix** — Tier 3. `go vet`'s
  `hostport` analyzer flags these.
- **`os.Root` over `os.Open` with path joining.** `os.Root` refuses paths that
  escape the directory, including via symlinks. That is the point, and it is a
  behavior change wherever an escaping path currently succeeds.
- **`runtime.AddCleanup` over `runtime.SetFinalizer`.** Better for new code, but
  cleanup timing and cycle behavior differ. Fine for a finalizer that only frees
  a resource; not fine when ordering is relied on.
- **`testing/synctest`** for concurrency tests. Additive, and it makes flaky
  time-based tests deterministic — but it virtualizes the clock inside the
  bubble, so a test that measured real durations behaves differently.
- **Generic methods (Go 1.27).** A package-level helper that conceptually
  belongs to one type can now live on it: `stream.Map(s, f)` becomes
  `s.Map(f)`. Condition: the helper is unexported or the package is internal —
  moving or renaming an exported function is Tier 3. Generic methods still
  cannot satisfy interfaces ([go-interfaces](../../go-interfaces/SKILL.md)).
- **`slog.GroupAttrs`** replacing a manually built group. Same output; verify
  the attribute order your handler emits.

---

## Tier 3 — report, don't apply

Genuinely better and genuinely observable. List them in the findings with a
one-line rationale so the user can schedule them.

- **`omitzero` over `omitempty`** (Go 1.24). Clearer intent, finally omits a
  zero `time.Time` — and changes the JSON on the wire.
- **Migrating call sites to `encoding/json/v2`** (GA in Go 1.27). Faster and a
  saner API, but v2 defaults are stricter, so it changes what parses.
- **Stdlib `uuid`** (Go 1.27) replacing `github.com/google/uuid`. Deletes a
  dependency, but the APIs differ so every call site moves. Usually a small
  mechanical PR once approved. See [go-packages](../../go-packages/SKILL.md).
- **`slog.NewMultiHandler`** (Go 1.26) replacing a hand-rolled fan-out — log
  routing is observable.
- **Adding `context.Context` plumbing.** The single most common "improvement"
  that changes cancellation behavior end to end. Its own piece of work.
- **Bumping the `go` directive.** Enables new language semantics *and* new vet
  diagnostics at once. Its own change, never a rider.

---

## Toolchain shifts that break tests on their own

If the refactor coincides with a toolchain bump, some failures are not yours.
Attribute before rewriting. Verified against go1.27.0:

- **`encoding/json` v1 is backed by v2 by default** (the `jsonv2` GOEXPERIMENT
  is on). Marshal/unmarshal semantics are preserved, but the exact text of
  error messages can differ — a test asserting a JSON error string breaks with
  no diff from you. `GOEXPERIMENT=nojsonv2` restores the old backend while you
  confirm the cause.
- **`GOMAXPROCS` is cgroup-aware since Go 1.25.** Under a container CPU limit,
  effective parallelism changes, which changes timing-dependent behavior.
  `runtime.SetDefaultGOMAXPROCS()` restores the default after an override.
- **Float results can move across a toolchain bump.** At `GOAMD64=v3` and above
  the compiler may fuse `a*b + c` into a single FMA. `float64(a*b) + c`
  prevents fusing. Relevant wherever money or aggregates are compared exactly.

Anything else you suspect is a toolchain change rather than yours: confirm it
before saying so. `git stash` the diff and re-run the failing test on the new
toolchain — if it still fails, it was never yours. Reporting a guess here is
how a refactor loses the reviewer's trust.

---

## Tools worth running once

| Tool | What it finds |
|---|---|
| `go vet ./...` | `waitgroupgo` (misplaced `wg.Add`), `hostport` (the IPv6 address bug), `stdversion` (stdlib symbols newer than the `go` directive) |
| `bash scripts/verify-refactor.sh leaks ./...` | Goroutine leaks — the `goroutineleak` pprof profile is GA in Go 1.27, so this turns "looks like it leaks" into a concrete list |
| `GODEBUG=checkfinalizers=1` | Finalizer and cleanup misuse (Go 1.25+) |
| `golangci-lint run` | Expect the finding count to drop after the refactor; report before and after |
