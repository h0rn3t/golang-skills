# Symptom Catalog

> Sources: `go doc runtime`; go.dev/doc/diagnostics; Go Wiki CodeReviewComments; runtime panic messages as printed by Go 1.27
> Authority: advisory — mechanisms are ranked by how often they turn out to be the cause
> Minimum Go: 1.27 baseline
> Last verified: 2026-09-02

Each entry: the symptom as reported → the mechanisms that produce it, most
common first → the command whose output confirms or rules out each → the skill
that owns the fix. Confirm before fixing; two mechanisms often share a symptom.

## Contents

- [Panics by message](#panics-by-message)
- [Hangs](#hangs)
- [Leaks and growth](#leaks-and-growth)
- [Wrong results](#wrong-results)
- [Tests](#tests)
- [Startup and build](#startup-and-build)
- [Environment differences](#environment-differences)

---

## Panics by message

| Message | Mechanism | Confirm | Owner |
|---|---|---|---|
| `nil pointer dereference` | Method on a nil receiver; field of a nil struct pointer; a constructor's error ignored so the value is nil | Frame above `runtime.` in the trace; `{0x0, ...}` in the args | [go-defensive](../../go-defensive/SKILL.md) |
| `nil pointer dereference` with a non-nil-looking value | Interface holding a typed nil pointer (`var p *T; var i I = p; i != nil`) | `fmt.Printf("%T %v", i, i)` prints the type with `<nil>` | [go-interfaces](../../go-interfaces/SKILL.md) |
| `index out of range [N] with length M` | Loop bound from a different slice; off-by-one on `len-1`; slice reused after `append` reallocated | `-list` the frame; check which length was used for the bound | [go-data-structures](../../go-data-structures/SKILL.md) |
| `slice bounds out of range` | `s[a:b]` with `a > b` or `b > cap` from parsed input | Validate the indices from input | [go-security](../../go-security/SKILL.md) (if input-driven) |
| `assignment to entry in nil map` | `var m map[K]V` written without `make`; struct field map never initialized | Grep the declaration; constructor missing | [go-data-structures](../../go-data-structures/SKILL.md) |
| `concurrent map writes` / `concurrent map read and map write` | Unsynchronized map shared across goroutines — the runtime's own race check | `go test -race` shows the two stacks | [go-concurrency](../../go-concurrency/SKILL.md) |
| `send on closed channel` | Producer still running after `close`; multiple closers | Race report or goroutine dump at the `close` and the `send` | [go-concurrency](../../go-concurrency/SKILL.md) |
| `close of closed channel` / `close of nil channel` | Two owners closing; channel field never made | Who owns the channel — exactly one sender should close | [go-concurrency](../../go-concurrency/SKILL.md) |
| `sync: negative WaitGroup counter` | `Done` without `Add`, or `Add` inside the goroutine after `Wait` started | `wg.Go` replaces the pair (Go 1.25+); `go vet` `waitgroup` analyzer | [go-concurrency](../../go-concurrency/SKILL.md) |
| `sync: unlock of unlocked mutex` | Double `Unlock`; `Unlock` on a copied struct (`go vet` copylocks) | `go vet ./...` | [go-concurrency](../../go-concurrency/SKILL.md) |
| `interface conversion: X is Y, not Z` | Unchecked type assertion `v.(Z)` | Use `v, ok := x.(Z)` or a type switch | [go-interfaces](../../go-interfaces/SKILL.md) |
| `integer divide by zero` | Parsed or computed denominator | Table test with zero | [go-defensive](../../go-defensive/SKILL.md) |
| `all goroutines are asleep - deadlock!` | Every goroutine blocked: unbuffered send with no receiver; `wg.Wait` with a missing `Done`; `Lock` twice on a non-reentrant mutex | The dump printed with the panic lists each wait state | [go-concurrency](../../go-concurrency/SKILL.md) |
| `too many open files` (as an error, then panics downstream) | `resp.Body`, `*os.File`, `sql.Rows` not closed; goroutine-per-connection with no limit | `lsof -p <pid> \| wc -l` over time; `bodyclose`, `sqlclosecheck` linters | [go-defensive](../../go-defensive/SKILL.md) |
| `fatal error: concurrent map iteration and map write` | `range` over a map another goroutine writes | Race detector | [go-concurrency](../../go-concurrency/SKILL.md) |
| `fatal error: stack overflow` / `goroutine stack exceeds 1000000000-byte limit` | Unbounded recursion; `String()` calling `Sprintf("%v", x)` on itself; `MarshalJSON` marshaling its own type | Trace shows the same frame thousands of times | [go-functions](../../go-functions/SKILL.md) |
| Panic inside `net/http` handler, connection closed | Handler panicked; `http.Server` recovers per connection and logs `http: panic serving` | Server error log; wrap with a recovering middleware that logs `%+v` and the request ID | [go-http](../../go-http/SKILL.md) |
| Panic with no user frames | Goroutine panicked after its parent returned; `created by` names the origin | `GOTRACEBACK=all` | [go-concurrency](../../go-concurrency/SKILL.md) |

---

## Hangs

| Symptom | Mechanism | Confirm | Owner |
|---|---|---|---|
| Whole process stops responding | Lock-order inversion between two mutexes | Dump: two goroutines each in `Lock` at different lines of the same package | [go-concurrency](../../go-concurrency/SKILL.md) |
| | Unbuffered channel, receiver exited on error | Dump: N goroutines in `chan send` at one line | [go-concurrency](../../go-concurrency/SKILL.md) |
| | `wg.Wait` forever | One worker returned early without `Done`, or panicked and recovered elsewhere | [go-concurrency](../../go-concurrency/SKILL.md) |
| | Connection pool exhausted | `db.Stats().WaitCount` climbing; goroutines in `database/sql.(*DB).conn` | [go-database](../../go-database/SKILL.md) |
| One request never returns | `select` without `ctx.Done()`; `time.After` in a loop; HTTP client with no timeout | Dump shows the goroutine in `select` or `net.(*netFD).Read` for minutes | [go-context](../../go-context/SKILL.md) / [go-http](../../go-http/SKILL.md) |
| | `context.WithTimeout` created but `cancel` never called (leak, not hang) | `go vet` `lostcancel` | [go-context](../../go-context/SKILL.md) |
| Shutdown never completes | `srv.Shutdown` waits on a hijacked/streaming connection; a worker ignores the shutdown context | Dump during shutdown; `Shutdown` with its own deadline | [go-http](../../go-http/SKILL.md) |
| CPU 100%, no progress | Spin loop on a condition another goroutine sets without synchronization; `for { select { default: } }` | CPU profile: one frame ~100% | [go-concurrency](../../go-concurrency/SKILL.md) |
| Slow, but CPU idle | Blocking syscall or network wait; GC assist; lock contention | `go tool trace` — long "Syscall" or "Blocked" bands; block/mutex profile | this skill, then owner |
| Starts fast, hangs after N requests | Semaphore or buffered channel never released on the error path | Count acquires vs releases; a `defer` missing on the early return | [go-concurrency](../../go-concurrency/SKILL.md) |

---

## Leaks and growth

| Symptom | Mechanism | Confirm | Owner |
|---|---|---|---|
| Goroutine count rises | Goroutine blocked on a channel nobody reads; `time.Ticker` never stopped; `context` never canceled | Two `debug=1` dumps; the stack whose count grows | [go-concurrency](../../go-concurrency/SKILL.md) |
| | HTTP client body not closed → connection goroutines held | `bodyclose` linter; dump shows `net/http.(*persistConn)` stacks | [go-http](../../go-http/SKILL.md) |
| Heap `inuse` rises, GC runs | Map used as a cache without eviction; slice of pointers retaining everything; global `append` | `pprof -sample_index=inuse_space -top`; `gctrace` live heap climbs | [go-data-structures](../../go-data-structures/SKILL.md) |
| | Subslice `s[:n]` of a large buffer keeps the whole array alive | Look for `bytes` held from a read buffer; `slices.Clone` the part you keep | [go-data-structures](../../go-data-structures/SKILL.md) |
| | `sync.Pool` of huge buffers; `bytes.Buffer` grown once and pooled | `inuse` at `bytes.(*Buffer).grow` | [go-concurrency](../../go-concurrency/SKILL.md) |
| | Timer/ticker per request without `Stop` (pre-Go 1.23 kept them alive until fire) | Heap at `time.NewTimer`; `go.mod` `go` directive < 1.23 | [go-context](../../go-context/SKILL.md) |
| RSS high, heap `inuse` low | Goroutine stacks (thousands of goroutines × stack size); CGO/`mmap`; runtime not yet returning memory | `MemStats.StackInuse`, `Sys - HeapSys`; `GODEBUG=madvdontneed=1` changes RSS shape only | [go-concurrency](../../go-concurrency/SKILL.md) if stacks |
| GC constantly running | Allocation rate high, not a leak | `alloc_space` top; `gctrace` shows frequent, small heaps | [go-performance](../../go-performance/SKILL.md) |
| OOM-killed with modest heap | `GOMEMLIMIT` unset in a memory-capped container; GC targets 2× live heap | Set `GOMEMLIMIT` ~ 90% of the cgroup limit; confirm with `gctrace` | [go-performance](../../go-performance/SKILL.md) |
| File descriptors rise | `os.Open` without `Close` on the error path; `Rows` not closed; listeners per request | `lsof`; `sqlclosecheck`; a missing `defer` right after the open | [go-defensive](../../go-defensive/SKILL.md) |

---

## Wrong results

| Symptom | Mechanism | Confirm | Owner |
|---|---|---|---|
| Value sometimes stale or garbled | Data race | `-race` | [go-concurrency](../../go-concurrency/SKILL.md) |
| Modified copy has no effect | Method with a value receiver mutating; `range` value variable; struct in a map (`m[k].f = v` does not compile, but a copy through a local does) | `go vet` does not catch it; read the receiver | [go-interfaces](../../go-interfaces/SKILL.md) |
| Slice changes appear elsewhere | Two slices sharing a backing array after `s[:n]` or `append` within capacity | Print `cap` and pointers; `slices.Clone` at the boundary | [go-data-structures](../../go-data-structures/SKILL.md) |
| Deferred value is wrong | `defer f(x)` evaluated `x` at defer time; closure captured a variable that changed | Trace the value at the `defer` line | [go-defensive](../../go-defensive/SKILL.md) |
| `errors.Is` never matches | Wrapped with `%v` not `%w`; a new error value each call instead of a sentinel; compared across a JSON/gRPC boundary | `check-errors.sh`; print `%T` of the chain | [go-error-handling](../../go-error-handling/SKILL.md) |
| JSON field missing or empty | Unexported field; wrong tag; `omitempty` on a zero that is meaningful; `nil` vs empty slice | `json.Marshal` round-trip test | [go-defensive](../../go-defensive/SKILL.md) |
| Time off by hours / DST | `time.Now()` vs `time.Now().UTC()`; `time.Date` in `Local`; comparing with `==` instead of `Equal` | Print `t.Location()` | [go-defensive](../../go-defensive/SKILL.md) |
| Rows missing after a loop | `rows.Err()` unchecked; `rows.Next` stopped on a scan error silently | `rowserrcheck` | [go-database](../../go-database/SKILL.md) |
| Behavior differs after upgrade | `GODEBUG` default changed with the `go` directive; loop variable semantics (1.22); `for range` over a function | `go doc runtime` godebug table; `MODERNIZATION.md` release notes | [go-code-refactor](../../go-code-refactor/SKILL.md) |
| Float comparison fails | `==` on computed floats | Compare with a tolerance; `math.Nextafter` | [go-defensive](../../go-defensive/SKILL.md) |

---

## Tests

| Symptom | Mechanism | Confirm | Owner |
|---|---|---|---|
| Passes alone, fails with the package | Shared package-level state; `t.Parallel` tests touching a global; `os.Chdir`; shared temp file name | `-shuffle=on` changes the failure; `-run` the pair | [go-testing](../../go-testing/SKILL.md) |
| Fails 1 in N | Race; `time.Sleep`-based ordering; map iteration order assumed | `-race -count=100`; rewrite under `synctest` | [go-testing](../../go-testing/SKILL.md) |
| Hangs | Goroutine waiting on a channel the test never feeds; `wg.Wait` with a missing `Done` | `-timeout 30s` dump | [go-concurrency](../../go-concurrency/SKILL.md) |
| `-race` fails only in CI | Faster local machine never interleaves the two accesses | It is a real race; `GOMAXPROCS=1` or `-cpu 1,4` locally | [go-concurrency](../../go-concurrency/SKILL.md) |
| Passes on macOS, fails on Linux | Case-insensitive filesystem; `/tmp` layout; file permissions; `\r\n`; signal semantics | Run in a container | this skill |
| Test is green but the bug persists | Test asserts on a mock, not behavior; test edited to match the code | Revert the code; the test must fail | [go-testing](../../go-testing/SKILL.md) |
| Coverage says the line ran but the branch is wrong | Table test with duplicate cases; `t.Run` names colliding | `-v` shows each subtest once | [go-testing](../../go-testing/SKILL.md) |

---

## Startup and build

| Symptom | Mechanism | Confirm | Owner |
|---|---|---|---|
| Slow start | Heavy `init`; global regexp compiles; TLS handshakes at boot | `GODEBUG=inittrace=1` | [go-packages](../../go-packages/SKILL.md) |
| `import cycle not allowed` | Two packages depending on each other's types | `go list -deps` on each; move the shared type to a third package | [go-packages](../../go-packages/SKILL.md) |
| `undefined: X` after upgrade | API newer than the `go` directive (`stdversion` vet) or removed | `go vet`; `grep X $(go env GOROOT)/api/go1.*.txt` | [go-code-refactor](../../go-code-refactor/SKILL.md) |
| Binary works with `go run`, fails as a container | `CGO_ENABLED`, missing CA certificates, `scratch` image without tzdata or `/tmp` | `go version -m ./app`; `ldd ./app` | [go-packages](../../go-packages/SKILL.md) |
| Binary size doubled | A dependency pulled in `net/http` + `reflect` + a template engine; debug info | `go tool nm -size -sort size`; `-ldflags='-s -w'` | [go-packages](../../go-packages/SKILL.md) |
| `go.sum` mismatch / checksum error | Proxy or replaced module differs from the recorded hash | `GOFLAGS=-mod=mod go mod verify`; `GONOSUMDB` scope | [go-packages](../../go-packages/SKILL.md) |

---

## Environment differences

When "it works on my machine", diff these before reading code:

```bash
go env GOOS GOARCH GOFLAGS GOMAXPROCS CGO_ENABLED GOTOOLCHAIN
go version -m ./app | head            # exact module versions and build settings in the binary
nproc; ulimit -n; cat /sys/fs/cgroup/memory.max 2>/dev/null
echo "$TZ" "$LANG"; date +%Z
```

A test timeout that fits an 8-core laptop fails on a 2-core runner; a
`-race` build is 5–10× slower; a container with `memory.max` set and no
`GOMEMLIMIT` is OOM-killed while the Go heap looks healthy.
