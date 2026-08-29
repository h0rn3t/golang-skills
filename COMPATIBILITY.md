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
mechanism — set `go 1.27` in `go.mod` and let vet catch the rest.

## Language features by version

| Feature | Since | Notes |
|---|---|---|
| Generic methods (`func (s *S) Get[T any](...)`) | 1.27 | Methods may declare their own type parameters |
| Relaxed inference for partially instantiated generic functions | 1.27 | Fewer explicit type arguments needed |
| Trailing comma in type parameter lists | 1.27 | `[T any,]` |
| `new(expr)` — allocate and initialize in one expression | 1.26 | `p := new(compute())`; `go fix` analyzer `newexpr` |
| Generic type aliases | 1.24 | `type Set[T comparable] = map[T]struct{}` |
| Per-iteration loop variables | 1.22 | The `x := x` capture line is dead code; `go fix` analyzer `forvar` removes it |
| `range` over integer and over function iterators | 1.22 / 1.23 | `for i := range n`, `for v := range seq` |
| `min`, `max`, `clear` builtins | 1.21 | `go fix` analyzer `minmax` |

## Standard-library APIs by version

Only entries a skill actually recommends are listed. Anything else is out of
scope for this repository.

### Go 1.27

| API | Replaces |
|---|---|
| `uuid` (`uuid.New`, `NewV4`, `NewV7`, `Parse`) | `github.com/google/uuid` for the common cases |
| `encoding/json/v2` + `encoding/json/jsontext` | `encoding/json` for new code that needs its semantics or streaming |
| `httptest.NewTestServer(tb, handler)` | `httptest.NewServer` + `defer srv.Close()` |
| `synctest.Sleep` | Real sleeps inside `synctest.Test` bubbles |
| `strings.CutLast`, `bytes.CutLast` | `LastIndex` + manual slicing |
| `url.URL.Clone`, `url.Values.Clone` | Hand-written deep copies at boundaries |
| `http.Server.MaxHeaderValueCount` (default `DefaultMaxHeaderValueCount` = 500) | Custom header-count guards |
| `http.Server.DisableClientPriority` | — |
| `hash/maphash.ComparableHasher[T]`, `maphash.Hasher[T]` | Hand-rolled hash/equality pairs for generic containers |
| `math/rand/v2.(*Rand).N` | `rand.N` package function when you need an explicit source |
| `crypto/mldsa`, `crypto.MLDSAMu` | Post-quantum signatures |
| `database/sql/driver.RowsColumnScanner` | Per-column driver scanning |

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
| `os.Root`, `os.OpenRoot` | Path-traversal-prone `filepath.Join` + `os.Open` |

### Go 1.21–1.23

`log/slog` (1.21), `cmp.Ordered` (1.21), `slices` and `maps` packages (1.21),
`testing/slogtest` (1.22), `iter.Seq` and the `*Seq` iterator variants across
`slices`, `maps`, and `strings` (1.23).

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
