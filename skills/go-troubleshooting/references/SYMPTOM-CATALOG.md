# Symptom Catalog

> Sources: `go doc runtime`; `$GOROOT/doc/godebug.md`; go.dev/doc/diagnostics; go.dev/doc/articles/race_detector; Go Wiki CodeReviewComments; runtime panic messages as printed by Go 1.27
> Authority: advisory — candidate mechanisms; ordering is not a measured likelihood
> Minimum Go: 1.27 baseline
> Last verified: 2026-09-02; test-order bisection note added 2026-09-18; panic, leak, build, and routing rows rechecked 2026-09-29

Confirm before fixing; two mechanisms often share a symptom. A stack/profile
pattern alone rarely proves ownership, causality, or a leak.

## Contents

- [Panics by message](#panics-by-message)
- [Hangs](#hangs)
- [Leaks and growth](#leaks-and-growth)
- [Wrong results](#wrong-results)
- [Tests](#tests)
- [Startup and build](#startup-and-build)
- [Edits do not change behavior](#edits-do-not-change-behavior)
- [Environment differences](#environment-differences)

---

## Panics by message

| Message | Mechanism | Confirm | Owner |
|---|---|---|---|
| `nil pointer dereference` | Method on a nil receiver; field of a nil struct pointer; a constructor's error ignored so the value is nil | Inspect the faulting operation and inputs in matching source/debugger; printed argument words are only clues | [go-defensive](../../go-defensive/SKILL.md) |
| `nil pointer dereference` with a non-nil-looking value | Interface holding a typed nil pointer (`var p *T; var i I = p; i != nil`) | `fmt.Printf("%T %v", i, i)` prints the type with `<nil>` | [go-defensive](../../go-defensive/SKILL.md#common-pitfalls) |
| `concurrent map writes` / `concurrent map read and map write` | Unsupported concurrent access to a shared map; runtime checks are not a complete race detector | `go test -race` shows the two stacks | [go-concurrency](../../go-concurrency/SKILL.md) |
| `send on closed channel` | Producer still running after `close`; multiple closers | Trace send/close ownership and ordering; this panic can occur without a data race | [go-concurrency](../../go-concurrency/SKILL.md) |
| `sync: negative WaitGroup counter` | More `Done` calls than `Add`: `Done` without `Add`, `Done` twice on one path, or a negative `Add`. `Add` inside the goroutine instead lets `Wait` return early or panics `sync: WaitGroup misuse: Add called concurrently with Wait` | `wg.Go` replaces the pair (Go 1.25+); `go vet` `waitgroup` analyzer | [go-concurrency](../../go-concurrency/SKILL.md) |
| `all goroutines are asleep - deadlock!` | Every goroutine blocked: unbuffered send with no receiver; `wg.Wait` with a missing `Done`; `Lock` twice on a non-reentrant mutex | The dump printed with the panic lists each wait state | [go-concurrency](../../go-concurrency/SKILL.md) |
| Panic inside `net/http` handler, connection closed | Handler panicked; `http.Server` recovers per connection and logs `http: panic serving` | Server error log; wrap with a recovering middleware that logs the panic value, `debug.Stack()`, and the request ID — `%+v` on a panic value prints no stack | [go-http](../../go-http/SKILL.md) |
| Truncated or uninformative trace | Capture/log truncation, hidden runtime frames, or missing matching source | Preserve complete panic output and matching binary/source; use `GOTRACEBACK=all` on an authorized subsequent run | this skill |

---

## Hangs

| Symptom | Mechanism | Confirm | Owner |
|---|---|---|---|
| Whole process stops responding | Lock-order inversion between two mutexes | Establish an ownership/wait cycle from stacks and code; two `Lock` frames alone do not prove inversion | [go-concurrency](../../go-concurrency/SKILL.md) |
| | Unbuffered channel, receiver exited on error | Dump: N goroutines in `chan send` at one line | [go-concurrency](../../go-concurrency/SKILL.md) |
| | `wg.Wait` forever | One worker returned early without `Done`, or panicked and recovered elsewhere | [go-concurrency](../../go-concurrency/SKILL.md) |
| | Connection pool exhausted | `db.Stats().WaitCount` climbing; goroutines in `database/sql.(*DB).conn` | [go-database](../../go-database/SKILL.md) |
| One request never returns | `select` without `ctx.Done()`; `time.After` in a loop; HTTP client with no timeout | Dump shows the goroutine in `select` or `net.(*netFD).Read` for minutes | [go-context](../../go-context/SKILL.md) / [go-http](../../go-http/SKILL.md) |
| Shutdown never completes | `srv.Shutdown` waits on an active streaming handler; a worker ignores the shutdown context | Dump during shutdown; `Shutdown` with its own deadline | [go-http](../../go-http/SKILL.md) |
| CPU 100%, no progress | Spin loop on a condition another goroutine sets without synchronization; `for { select { default: } }` | CPU profile: one frame ~100% | [go-concurrency](../../go-concurrency/SKILL.md) |
| Slow, but CPU idle | Blocking syscall or network wait; GC assist; lock contention | `go tool trace` — long "Syscall" or "Blocked" bands; block/mutex profile | this skill, then owner |
| Starts fast, hangs after N requests | Semaphore or buffered channel never released on the error path | Count acquires vs releases; a `defer` missing on the early return | [go-concurrency](../../go-concurrency/SKILL.md) |

---

## Leaks and growth

| Symptom | Mechanism | Confirm | Owner |
|---|---|---|---|
| Goroutine count rises | Goroutine waiting after its owner finished; worker has no effective termination path | `/debug/pprof/goroutineleak?debug=1` (Go 1.27+) lists goroutines that can never unblock; else comparable dumps show growth beyond the expected lifetime — inspect the growing stack and its owner | [go-concurrency](../../go-concurrency/SKILL.md) |
| | HTTP client body not closed → connection goroutines held | `bodyclose` linter; dump shows `net/http.(*persistConn)` stacks | [go-http](../../go-http/SKILL.md) |
| Heap `inuse` rises, GC runs | Map used as a cache without eviction; slice of pointers retaining everything; global `append` | `pprof -sample_index=inuse_space -top`; `gctrace` live heap climbs | [go-data-structures](../../go-data-structures/SKILL.md) |
| | Subslice `s[:n]` of a large buffer keeps the whole array alive | Look for `bytes` held from a read buffer; `slices.Clone` the part you keep | [go-data-structures](../../go-data-structures/SKILL.md) |
| | `sync.Pool` of huge buffers; `bytes.Buffer` grown once and pooled | `inuse` at `bytes.(*Buffer).grow` | [go-concurrency](../../go-concurrency/SKILL.md) |
| | Timer/ticker per request without `Stop` (pre-Go 1.23 kept them alive until fire) | Heap at `time.NewTimer`, and the binary was built with a toolchain < 1.27 — 1.27 removed `asynctimerchan`, so the old behavior is unreachable whatever the `go` directive says | [go-context](../../go-context/SKILL.md) |
| RSS high, heap `inuse` low | Goroutine stacks (thousands of goroutines × stack size); CGO/`mmap`; runtime not yet returning memory | `MemStats.StackInuse`, `Sys - HeapSys`; `madvdontneed=0` in the process's `GODEBUG` (the Linux default is 1) keeps freed pages in RSS until memory pressure — RSS shape only | [go-concurrency](../../go-concurrency/SKILL.md) if stacks |
| GC constantly running | Allocation rate high, not a leak | `alloc_space` top; `gctrace` shows frequent, small heaps | [go-performance](../../go-performance/SKILL.md) |
| OOM-killed with modest heap | `GOMEMLIMIT` unset in a memory-capped container; GC targets 2× live heap | Set `GOMEMLIMIT` ~ 90% of the cgroup limit; confirm with `gctrace` | [go-performance](../../go-performance/SKILL.md) |
| File descriptors rise | `os.Open` without `Close` on the error path; `Rows` not closed; listeners per request | `lsof`; `sqlclosecheck`; a missing `defer` right after the open | [go-defensive](../../go-defensive/SKILL.md) |

---

## Wrong results

For a concrete ticket, first use [ticket investigation](TICKET-INVESTIGATION.md)
and [data-flow tracing](DATA-FLOW-TRACING.md). Follow identity, row count, values,
and serialization before choosing a runtime capture.

| Symptom | Mechanism | Confirm | Owner |
|---|---|---|---|
| Value sometimes stale or garbled | Data race | `-race` | [go-concurrency](../../go-concurrency/SKILL.md) |
| Slice changes appear elsewhere | Two slices sharing a backing array after `s[:n]` or `append` within capacity | Print `cap` and pointers; `slices.Clone` at the boundary | [go-defensive](../../go-defensive/SKILL.md#common-pitfalls) |
| `errors.Is` never matches | Wrapped with `%v` not `%w`; a new error value each call instead of a sentinel; compared across a JSON/gRPC boundary | `check-errors.sh`; print `%T` of the chain | [go-error-handling](../../go-error-handling/SKILL.md) |
| Rows missing after a loop | `rows.Err()` unchecked; `rows.Next` stopped on a scan error silently | `rowserrcheck` | [go-database](../../go-database/SKILL.md) |
| Behavior differs after upgrade | `GODEBUG` default changed with the `go` directive; loop variable semantics (1.22); `for range` over a function | `go version -m <binary>` (`build DefaultGODEBUG=…`) or `go list -f '{{.DefaultGODEBUG}}' ./cmd/app` for both builds; `MODERNIZATION.md` release notes | [go-code-refactor](../../go-code-refactor/SKILL.md) |

---

## Tests

| Symptom | Mechanism | Confirm | Owner |
|---|---|---|---|
| Passes alone, fails with the package | Shared package-level state; `t.Parallel` tests touching a global; `os.Chdir`; shared temp file name | `-shuffle=on` changes the failure; replay the seed with `-v`, then bisect the tests that ran before it with `-run` until the pair remains | [go-testing](../../go-testing/SKILL.md) |
| Fails 1 in N | Race; `time.Sleep`-based ordering; map iteration order assumed | `-race -count=100`; rewrite under `synctest` | [go-testing](../../go-testing/SKILL.md) |
| Hangs | Goroutine waiting on a channel the test never feeds; `wg.Wait` with a missing `Done` | `-timeout 30s` dump | [go-concurrency](../../go-concurrency/SKILL.md) |
| `-race` fails only in CI | Faster local machine never interleaves the two accesses | It is a real race; `GOMAXPROCS=1` or `-cpu 1,4` locally | [go-concurrency](../../go-concurrency/SKILL.md) |
| Test is green but the bug persists | Test asserts on a mock, not behavior; test edited to match the code | Revert the code; the test must fail | [go-testing](../../go-testing/SKILL.md) |

---

## Startup and build

| Symptom | Mechanism | Confirm | Owner |
|---|---|---|---|
| Slow start | Heavy `init`; global regexp compiles; TLS handshakes at boot | `GODEBUG=inittrace=1` | [go-packages](../../go-packages/SKILL.md) |
| `undefined: X` after upgrade | A toolchain older than the release that added `X` (an API newer only than the `go` directive still builds; `go vet` `stdversion` reports it), or `X` removed from an upgraded dependency | `go version` where it fails; `grep X $(go env GOROOT)/api/go1.*.txt`; `go list -m <dep>` and `go doc <pkg> X` | [go-code-refactor](../../go-code-refactor/SKILL.md) |

---

## Edits do not change behavior

First establish which source and artifact the failing command actually uses.
Run these from the failing command's working directory, replacing `.` with its
package target and preserving its build flags (`-tags`, `-mod`, `-modfile`,
`-overlay`) and environment:

```bash
go env GOMOD GOWORK GOFLAGS GOOS GOARCH CGO_ENABLED GOTOOLCHAIN
go list -compiled -json .
go list -m -json all
```

- Check `Dir`, `ImportPath`, `GoFiles`, `CgoFiles`, `CompiledGoFiles`, and
  `IgnoredGoFiles`. Match `//go:build`, filename OS/architecture suffixes, and
  cgo settings to the failing invocation. For tests, use `go list -test -json`
  with the same target/flags and inspect `TestGoFiles` and `XTestGoFiles` too.
  If the edit is in a dependency, list that package's import path explicitly
  or add `-deps`; listing only the caller does not show dependency files.
- Inspect the active `go.mod` and, when enabled, `go.work`: workspace `use`
  entries and `replace` directives can select a different checkout. Module
  JSON exposes `Replace` and `Dir`; an active workspace replacement overrides
  a module replacement. Account for vendor mode rather than switching it off
  to make a diagnostic command succeed.
- Distinguish `go run main.go` (explicit files) from `go run .` (the selected
  package). For a running binary/container, establish its executable/image
  identity and inspect `go version -m /path/to/binary`; a local `go list` cannot
  prove which artifact is deployed. Confirm the edited code path executes.

Build caching and test-result caching are different. For a suspicious cached
test result, rerun the focused test with `-count=1`; this does not disable the
build cache. Go normally tracks Go source and compiler-input changes, so avoid
routine cache deletion. Toolchain/cache defects remain possible: if evidence
points there, preserve the reproduction and inspect `GODEBUG=gocachetest=1`
(test reuse) or `gocacheverify=1` (rebuild and verify build-cache entries).
Changes to external C libraries used by cgo are a documented exception to
build-cache tracking; a targeted rebuild with `-a` may then be appropriate.

---

## Environment differences

When "it works on my machine", diff these before reading code:

```bash
go env GOOS GOARCH GOFLAGS CGO_ENABLED GOTOOLCHAIN
go version -m ./app | head            # exact module versions and build settings in the binary
nproc; ulimit -n; cat /sys/fs/cgroup/memory.max 2>/dev/null
echo "$TZ" "$LANG"; date +%Z
```
