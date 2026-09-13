# Go Compatibility Policy

Baseline for all skills in this repository: **Go 1.27**.

Go supports the two most recent major releases. As of Go 1.27 (August 2026)
that is **1.26 and 1.27**. Skills target 1.27 and name a fallback only where a
maintained toolchain (1.26) lacks the API.

When a skill recommends a standard-library API or language feature introduced
after 1.21, it must name the minimum version inline, e.g. `errors.AsType`
(Go 1.26+).

## Verifying a version claim

Do not guess. Every claim in this table is checkable against the installed
toolchain:

```bash
go version
grep -rn "AsType" "$(go env GOROOT)/api/go1.26.txt"   # when an API landed
go doc encoding/json/v2                               # whether it is reachable
go tool fix help                                      # available modernizers
go tool vet help                                      # available vet analyzers
```

`go vet` runs the `stdversion` analyzer, which reports uses of standard-library
symbols newer than the `go` directive in `go.mod`. That is the enforcement
mechanism for **library** claims — set `go 1.27` in `go.mod` and let vet catch
the rest. The compiler gates most **language** features by the directive as
well: `for i := range n` in a `go 1.21` module, `new(expr)` below 1.26, and a
promoted field in a keyed literal or a generic method below 1.27 all fail with
`requires go1.NN or later (-lang was set to go1.MM; check go.mod)` (verified
with go1.27.1). Two are not gated: function-type inference in a conversion
(1.27) and self-referential constraints (1.26) compile at an older directive
on a 1.27 toolchain and fail only on a real older toolchain, so verify code
that depends on them on the CI toolchain.

## Which version governs

The `go` directive of the module's `go.mod`. A `go.mod` with no `go` line is
language version 1.16 ([go.dev/ref/mod](https://go.dev/ref/mod#go-mod-file-go)),
and a `go.work` with none is 1.18; a workspace's directive selects the
toolchain for the workspace and does not raise a member module's language
version. A `toolchain` line, `GOTOOLCHAIN`, and the Go installed on the
machine choose the compiler, not the language version: none of them makes a
newer form legal, and bumping the directive is its own change
(`skills/go-code-refactor/references/MODERNIZATION.md`, Tier 3). The
write-time list of forms per version is
`skills/go-style-core/references/CURRENT-GO.md`.

## Language features by version

| Feature | Since | Notes |
|---|---|---|
| Generic methods (`func (s *S) Get[T any](...)`) | 1.27 | Methods may declare their own type parameters |
| Promoted fields in keyed struct literals | 1.27 | Direct initialization through embedded value fields; `go fix` analyzer `embedlit`; pointer embedding and overlapping enclosing/promoted fields are excluded |
| Function type inference in all assignments and conversions | 1.27 | Typed-variable assignment already supported inference in 1.21 |
| Trailing comma in type parameter lists | 1.18 | `[T any,]`; part of the original generics syntax |
| `new(expr)` — allocate and initialize in one expression | 1.26 | `p := new(compute())`; `go fix` analyzer `newexpr` |
| Generic type aliases | 1.24 | `type Set[T comparable] = map[T]struct{}` |
| `tool` directive in `go.mod`, run with `go tool <name>` | 1.24 | Tracked tool dependencies; replaces a `tools.go` of blank imports |
| Per-iteration loop variables | 1.22 | The `x := x` capture line is dead code; `go fix` analyzer `forvar` removes it |
| `range` over integer and over function iterators | 1.22 / 1.23 | `for i := range n`, `for v := range seq` |
| `min`, `max`, `clear` builtins | 1.21 | `go fix` analyzer `minmax` |
| `any` as the spelling of `interface{}` | 1.18 | Predeclared alias; `go fix` analyzer `any` |

## Standard-library APIs by version

Only entries a skill actually recommends are listed. Anything else is out of
scope for this repository.

### Go 1.27

| API | Replaces |
|---|---|
| `uuid` (`New`, `NewV4`, `NewV7`, `Parse`, `MustParse`, `Nil`, `Max`, `Compare`) | `github.com/google/uuid` for the common cases |
| `encoding/json/v2` + `encoding/json/jsontext` | `encoding/json` for new code that needs its semantics or streaming |
| `httptest.NewTestServer(tb, handler)` (in-memory; reached only via `srv.Client()`) | `httptest.NewServer` + `defer srv.Close()` |
| `synctest.Sleep` | Real sleeps inside `synctest.Test` bubbles |
| `strings.CutLast`, `bytes.CutLast` | `LastIndex` + manual slicing |
| `url.URL.Clone`, `url.Values.Clone` | Hand-written deep copies at boundaries |
| `http.Server.MaxHeaderValueCount` (default `DefaultMaxHeaderValueCount` = 500) | Custom header-count guards |
| `http.Server.DisableClientPriority` | A custom HTTP/2 write scheduler installed only to ignore client priority |
| `hash/maphash.ComparableHasher[T]`, `maphash.Hasher[T]` | Hand-rolled hash/equality pairs for generic containers |
| `math/rand/v2.(*Rand).N` | `rand.N` package function when you need an explicit source |
| `crypto/mldsa`, `crypto.MLDSAMu` | Post-quantum signatures |
| `database/sql/driver.RowsColumnScanner` | Per-column driver scanning |
| `database/sql.ConvertAssign` | Reimplementing `Rows.Scan`'s conversions in a custom `sql.Scanner` |

### Go 1.26

| API | Replaces |
|---|---|
| `errors.AsType[T](err) (T, bool)` | `errors.As` with a declared target variable; `go fix` analyzer `errorsastype` |
| `slog.NewMultiHandler(handlers...)` | Hand-written fan-out handlers |
| `testing.TB.ArtifactDir()` | Ad-hoc temp dirs for test output that must survive the run |
| `bytes.Buffer.Peek` | Read-then-unread dances |

### Go 1.25

| API | Replaces |
|---|---|
| `sync.WaitGroup.Go(func())` | `wg.Add(1)` / `go func(){ defer wg.Done() }()`; `go fix` analyzer `waitgroupgo` |
| `testing/synctest.Test`, `synctest.Wait` | Sleep-based waits in concurrency tests |
| `slog.GroupAttrs(key, attrs...)` | `slog.Group` with `any` varargs |
| `http.NewCrossOriginProtection()` | Hand-rolled CSRF origin checks |
| `runtime.SetDefaultGOMAXPROCS()` | Manual `GOMAXPROCS` math; 1.25 makes the default cgroup-aware |
| `testing.TB.Output()`, `TB.Attr()` | `fmt.Println` in tests, untyped metadata in failure text |

### Go 1.24

| API | Replaces |
|---|---|
| `b.Loop()` | `for i := 0; i < b.N; i++` |
| `t.Context()` | `context.Background()` in tests; `go fix` analyzer `testingcontext` |
| `crypto/rand.Text()` | Manual random-string generation |
| `cipher.NewGCMWithRandomNonce` | Manual random nonce generation/prefixing for AES-GCM in new formats |
| `os.Root`, `os.OpenRoot` | Path-traversal-prone `filepath.Join` + `os.Open` |
| `strings.SplitSeq`, `FieldsSeq`, `Lines` (and `bytes`) | `Split` followed by a `range` over the slice |
| `omitzero` struct tag option | `omitempty` plus a custom `IsZero` |
| `slog.DiscardHandler` | Hand-written no-op handlers |
| `t.Chdir()` | `os.Chdir` plus a restore in cleanup |

### Go 1.18–1.23

`any`, `strings.Cut`, `bytes.Cut`, `strings.Clone` (1.18); `fmt.Appendf` and
the typed atomics `atomic.Bool`, `atomic.Int64`, `atomic.Pointer[T]` (1.19);
`strings.CutPrefix`/`CutSuffix`, `bytes.Clone`, `errors.Join`,
`context.WithCancelCause`/`Cause` (1.20); `log/slog`, `cmp.Ordered`,
`cmp.Compare`, the `slices` and `maps` packages including `slices.Index`,
`Max`/`Min`, `Reverse`, `Compact`, `Clip` and `maps.DeleteFunc`,
`sync.OnceFunc`/`OnceValue`/`OnceValues`, `context.AfterFunc`,
`WithoutCancel`, `WithTimeoutCause`/`WithDeadlineCause` (1.21); `cmp.Or`,
`slices.Concat`, `sql.Null[T]`, `reflect.TypeFor`, `http.ServeMux` method and
wildcard patterns with `r.PathValue`, `http.ServeFileFS` and
`http.FileServerFS`, `testing/slogtest` (1.22); `iter.Seq` with the iterator
forms in `slices` (`All`, `Values`, `Collect`, `Sorted`, `AppendSeq`) and
`maps` (`All`, `Keys`, `Values`, `Collect`), and a `time.Tick` ticker the
garbage collector reclaims once unreferenced (1.23) — the `strings` `*Seq`
forms are 1.24.

### Deprecated APIs and replacements

The [modernization catalog](skills/go-code-refactor/references/MODERNIZATION.md#deprecated-apis-and-replacements)
owns the replacement table, deprecation versions, and risk conditions. It ships
inside the skill so these rules remain available in skill-only installations.

## Fallback policy

State a fallback only when it is real:

- The API is newer than 1.26 **and** the guidance matters on 1.26.
- The alternative is not merely more verbose but semantically different.

Do not add "on older Go, do X" notes for anything available in both supported
releases — it is noise the reader must skip.

## When a new Go release lands

1. `grep` the new `api/go1.NN.txt` in `GOROOT` for packages this repository
   recommends.
2. `go tool fix help` — new modernizers usually mean an obsolete pattern in a
   skill.
3. Update this file, then the skills that name the affected APIs.
4. Bump `go-version` in `.github/workflows/validate-skills.yml` and the `go`
   directive in `evals/go.mod`.
5. Add a regression assertion in `evals/eval_test.go` so the new guidance
   cannot silently rot.
