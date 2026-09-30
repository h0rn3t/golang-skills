# Modernization Catalog

> Sources: `$GOROOT/api/go1.2*.txt`; `go tool fix help`; Go spec; package docs; JetBrains go-modern-guidelines `guidelines.json` at `155dc7c` (all 54 items cross-checked 2026-09-13; the write-time card is `go-style-core/references/CURRENT-GO.md`)
> Authority: normative for tier placement; project policy for what may ride in a refactor
> Minimum Go: gated by the `go` directive in `go.mod`, not the installed toolchain
> Last verified: 2026-09-13

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
A default run includes `hostport`, whose hunks are Tier 2: keep them only where
IPv6 cannot reach the code, or leave them out with `-hostport=false`. Scope
follows [Scope mechanical modernization](../SKILL.md#3-scope-mechanical-modernization).
`go tool fix help` is authoritative;
[go-linting](../../go-linting/SKILL.md#modernization-go-fix) lists the
analyzers the idiom card does not show and the Go 1.27 renames. Report incorrect fixes; do not silently discard them.

## Tier 1 — safe swaps

[CURRENT-GO.md](../../go-style-core/references/CURRENT-GO.md) gives each form
and its trap; an entry here names only what keeps the swap behavior-identical
on existing code, and the `go fix` analyzer where one exists.

- **`new(expr)`** (Go 1.26): a hand edit, only where the temp has no other use;
  `go fix -newexpr` rewrites only `func ptr(x T) *T { return &x }` helpers and
  their calls.
- **`errors.AsType[T]`** (Go 1.26): same matching semantics as `errors.As`.
  `go fix -errorsastype`.
- **`wg.Go`** (Go 1.25): preserve goroutine count and `Add`-before-start. If the
  old code calls `Add` inside the goroutine, correcting that race is a Tier 3
  fix for that call site. `go fix -waitgroupgo`.
- **`strings.CutLast`, `bytes.CutLast`** (Go 1.27): check what the old `-1`
  branch returned before folding it into `ok` — if it returned anything other
  than `(s, "")`, the swap needs an explicit `if !ok`.
- **Redundant type arguments** (Go 1.27): inference now applies to conversions
  to matching function types, so `Fold(combine[int])` becomes `Fold(combine)`.
- **`strings.SplitSeq`, `FieldsSeq`, `FieldsFuncSeq`** (Go 1.24): only when the
  slice is not indexed, re-ranged, or kept — otherwise it is a rewrite, not a
  swap. `strings.Lines` is **not** a drop-in: it keeps each trailing newline.
- **`for i := range n`** (Go 1.22): only when the body does not mutate `i` or
  `n`; `range n` evaluates `n` once. `go fix -rangeint`.
- **Delete `x := x`** (Go 1.22): a mechanical delete under a `go` directive
  below 1.22 is a real bug. `go fix -forvar`.
- **`min`, `max`, `clear`** (Go 1.21): float helpers need a NaN and signed-zero
  check first. Replace a delete loop with `clear(m)` only when no key can hold
  a NaN, through an array, struct, or interface included: the loop cannot
  delete a NaN key and `clear` can, making that a Tier 3 fix. `go fix -minmax`.
- **`cmp.Or`** (Go 1.22): not for expensive or side-effecting fallbacks.
- **Test-only conveniences**: `t.Context()`, `t.Chdir()`, `slog.DiscardHandler`,
  `b.Loop()` (Go 1.24); `synctest.Sleep`, `httptest.NewTestServer` (Go 1.27).
  Preserve what the test observes — cancellation and cleanup timing, clock
  behavior, transport coverage: `httptest.NewTestServer` has no listener for
  anything else to dial. See [go-testing](../../go-testing/SKILL.md).

### `url.URL.Clone` / `url.Values.Clone` — Go 1.27

Use only when documented depth matches the contract; `url.Values.Clone` copies
its value slices, unlike `maps.Clone`. Treat a shallow-to-deep swap as a Tier 3
semantic fix with a contract test, not a behavior-preserving cleanup. `Clone`
returns nil for nil, where a copy loop over `make` returned a writable map:
`out := v.Clone(); if out == nil { out = url.Values{} }` keeps that. See the [go-defensive copy-depth rule](../../go-defensive/references/BOUNDARY-COPYING.md#copy-depth-is-part-of-the-contract).

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

Check nilness and capacity before replacing any collection loop or copy.
**Watch the sort**: `sort.Slice` is unstable, `slices.SortFunc` is
unstable, `slices.SortStableFunc` is stable — match what the original used,
because for equal keys the output order is observable.

## Tier 2 — safe with a condition

- **`errors.Join` over a text aggregate** (Go 1.20). Verify error text, nil
  behavior, and the exposed error tree. Even with identical text, `Join` can
  make `errors.Is`/`errors.AsType` match causes previously hidden by the string.
  Apply only when all observable error behavior is preserved; otherwise Tier 3.
- **`net.JoinHostPort` over `fmt.Sprintf("%s:%d", host, port)`.** Identical for
  IPv4 and hostnames; for IPv6 the old form produced an unusable address, so if
  IPv6 can reach this code the swap is a **bug fix** — Tier 3. `go vet` flags
  these, and a default `go fix` rewrites them (`hostport`).
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
| `go fix -diff -buildtag -plusbuild <pkgs>` | `buildtag`: malformed `//go:build` or `// +build` directives; `plusbuild`: obsolete `// +build` lines the `//go:build` form made redundant (Go 1.17+) — hygiene, not modernization |
| `bash scripts/verify-refactor.sh leaks ./...` | Runs tests but reports leak verification as incomplete (exit 3 if tests pass); it does not collect a profile |
| `GODEBUG=checkfinalizers=1` | Finalizer and cleanup misuse (Go 1.25+) |
| `golangci-lint run` | Expect the finding count to drop after the refactor; report before and after |
