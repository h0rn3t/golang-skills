# Modernization Catalog

> Sources: `$GOROOT/api/go1.2*.txt`; `go tool fix help`; Go spec; package docs
> Authority: normative for tier placement; project policy for what may ride in a refactor
> Minimum Go: gated by the `go` directive in `go.mod`, not the installed toolchain
> Last verified: 2026-08-29 against go1.27.0

Check version claims against the installed toolchain using `COMPATIBILITY.md`.
For existing code, work Tier 1, then conditional Tier 2; report Tier 3 changes
separately. For new code, use the [stdlib table](OVER-ENGINEERING.md#reach-for-what-go-ships)
and the required contract rather than refactor equivalence.

## Contents

- [Start with `go fix`](#start-with-go-fix)
- [Tier 1 — safe swaps](#tier-1--safe-swaps)
- [Tier 2 — safe with a condition](#tier-2--safe-with-a-condition)
- [Tier 3 — report, don't apply](#tier-3--report-dont-apply)
- [Deprecated APIs and replacements](#deprecated-apis-and-replacements)
- [Toolchain shifts that break tests on their own](#toolchain-shifts-that-break-tests-on-their-own)
- [Tools worth running once](#tools-worth-running-once)

## Start with `go fix`

Since Go 1.26, `go fix` hosts modernizers that rewrite code to current idioms.
Preview their output and verify the relevant behavior contracts.

Use the package scope and apply conditions in
[Scope mechanical modernization](../SKILL.md#3-scope-mechanical-modernization).
Preview before applying; a scoped refactor does not authorize unrelated
modernization. Keep mechanical changes distinguishable from hand edits.

`go tool fix help` is authoritative. Go 1.27 added `atomictypes`, `embedlit`,
`slicesbackward`, and `unsafefuncs`, renamed `waitgroup` to `waitgroupgo`, and
dropped `fmtappendf`. Report incorrect fixes; do not silently discard them.
[go-linting](../../go-linting/SKILL.md) catalogues the current analyzers.

## Tier 1 — safe swaps

### `new(expr)` — Go 1.26

```go
// before
age := yearsSince(born)
p := Person{Name: name, Age: &age}

// after
p := Person{Name: name, Age: new(yearsSince(born))}
```

Useful for JSON/protobuf optional fields such as `*int`/`*bool`.
Preview with `go fix -newexpr -diff`.

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

Preserve goroutine count and `Add`-before-start. If the old code calls `Add`
inside the goroutine, correcting that race is a Tier 3 fix for that call site.
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

Inference now applies to all assignments and conversions to matching function
types. Typed-variable assignment already worked in Go 1.21; conversion is new:

```go
type Fold func(int, int) int
fold := Fold(combine[int]) // before
fold := Fold(combine)      // Go 1.27
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

// after — nil result permitted
keys := slices.Sorted(maps.Keys(m))

// after — original non-nil empty result is observable
keys := slices.AppendSeq(make([]string, 0, len(m)), maps.Keys(m))
slices.Sort(keys)
```

`slices.Sorted(maps.Keys(m))` is the swap worth making — one line for three —
but it returns nil for an empty map where the loop returned a non-nil empty
slice, turning JSON `[]` into `null`. When that is observable, append into the
allocation with `slices.AppendSeq` (Go 1.23+). Check nilness and capacity
before replacing any collection loop or copy.

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

Replace hand-written integer helpers when their comparisons agree with the
builtins; float helpers need a NaN and signed-zero check first, because
`min`/`max` propagate NaN and distinguish `-0` from `+0`. Replace
`for k := range m { delete(m, k) }` with `clear(m)` only when no key can hold
a NaN, through an array, struct, or interface included: the loop cannot delete
a NaN key and `clear` can, making that a Tier 3 fix. `clear` on a *slice*
zeroes elements rather than truncating — it is not `s = s[:0]`. `go fix -minmax`.

### `cmp.Or` for fallback chains — Go 1.22

`name := cmp.Or(input, defaultName)` replaces the if-chain, but evaluates every
argument — not for expensive or side-effecting fallbacks.

### Test-only conveniences — Go 1.24–1.27

`t.Context()`, `t.Chdir()`, `slog.DiscardHandler`, `b.Loop()`, and — Go 1.27 —
`synctest.Sleep` and `httptest.NewTestServer`. Preserve what the test observes
— cancellation and cleanup timing, clock behavior, transport coverage:
`httptest.NewTestServer` serves its own client over an in-memory network by
default, with no listener for anything else to dial. See
[go-testing](../../go-testing/SKILL.md).

## Tier 2 — safe with a condition

- **`errors.Join` over a text aggregate** (Go 1.20). Verify error text, nil
  behavior, and the exposed error tree. Even with identical text, `Join` can
  make `errors.Is`/`errors.AsType` match causes previously hidden by the string.
  Apply only when all observable error behavior is preserved; otherwise Tier 3.
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

## Deprecated APIs and replacements

**Deprecated in** marks the old API's deprecation, not the replacement's release.
**Tier**: 1 apply, 2 apply if the stated condition holds, 3 report only.
The target module must support the replacement; report unproven Tier 2 swaps.

| Deprecated | Deprecated in | Use instead | Tier |
|---|---|---|---|
| `math/rand.Seed`, `math/rand.Read` — the package itself is not deprecated | 1.20 | `math/rand/v2` (1.22): different API, and a different value sequence for the same seed | 3 |
| `crypto/elliptic` key-agreement helpers (`GenerateKey`, `Marshal`, `Unmarshal`, `Curve` arithmetic) | 1.21 | `crypto/ecdh` (1.20); custom-curve arithmetic has no replacement | 3 |
| `reflect.SliceHeader`, `reflect.StringHeader` | 1.21 | `unsafe.Slice`, `unsafe.String` — recheck what keeps the backing memory alive | 2 |
| `reflect.PtrTo` | 1.22 | `reflect.PointerTo` | 1 |
| `runtime.GOROOT()` | 1.24 | `go env GOROOT` — a subprocess, not an in-process constant | 2 |
| `crypto/cipher` `NewOFB`, `NewCFBEncrypter`, `NewCFBDecrypter` | 1.24 | AEAD modes, or `NewCTR` when the ciphertext size must not change | 3 |
| `golang.org/x/crypto/sha3` — not deprecated | — | `crypto/sha3` (1.24); verify constructor signatures and exposed types remain compatible; `NewLegacyKeccak256`/`512` stay in `x/crypto` | 2 |
| `golang.org/x/crypto/hkdf` — not deprecated | — | `crypto/hkdf` (1.24); replace streaming reads only when the total output length is known and read/error behavior is preserved; adapt the new error returns and `info` argument | 2 |
| `golang.org/x/crypto/pbkdf2` — not deprecated | — | `crypto/pbkdf2` (1.24); adapt argument order, password type, and the new error return; prove reachable inputs satisfy the new validation and preserve failure behavior | 2 |
| `ReverseProxy.Director` | 1.26 | `ReverseProxy.Rewrite` — different `X-Forwarded-*` defaults | 3 |
| `crypto/tls.Config.Rand` | 1.27 | nil in production; `testing/cryptotest.SetGlobalRandom` (1.26) in tests | 2 |

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
- **Seven `GODEBUG` keys were removed in Go 1.27** (`asynctimerchan`,
  `gotypesalias`, `tlsunsafeekm`, `tlsrsakex`, `tls3des`, `tls10server`,
  `x509keypairleaf`). A `godebug` line or `//go:debug` comment pinning any of
  them to its old value now fails the build — delete the pin rather than
  carrying it forward.
- **Float results can move across a toolchain bump.** At `GOAMD64=v3` and above
  the compiler may fuse `a*b + c` into a single FMA. `float64(a*b) + c`
  prevents fusing. Relevant wherever money or aggregates are compared exactly.

Attribute suspected toolchain failures by rerunning the unchanged source
on that toolchain in an isolated checkout, preserving the user's working tree.

## Tools worth running once

To check leaks, write `pprof.Lookup("goroutineleak").WriteTo(w, 1)` in the
tested process (Go 1.27+) or use the existing `goleak` harness: writing the
profile is what triggers detection, so `Count` alone and a passing `go test`
prove nothing. Read the stacks; an empty profile still misses leaks that are
blocked on a reachable channel or mutex.

| Tool | What it finds |
|---|---|
| `go vet ./...` | `waitgroup` (misplaced `wg.Add`), `hostport` (the IPv6 address bug), `stdversion` (stdlib symbols newer than the `go` directive) |
| `bash scripts/verify-refactor.sh leaks ./...` | Runs tests but reports leak verification as incomplete (exit 3 if tests pass); it does not collect a profile |
| `GODEBUG=checkfinalizers=1` | Finalizer and cleanup misuse (Go 1.25+) |
| `golangci-lint run` | Expect the finding count to drop after the refactor; report before and after |
