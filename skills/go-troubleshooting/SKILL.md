---
name: go-troubleshooting
description: Use when a Go program misbehaves and the cause is unknown — a panic with an unclear trace, a hang or deadlock, a goroutine or memory leak, CPU or RSS that climbs over hours, a test that fails only under -race or only in CI, or a service that "just stops responding". Also use when the user pastes a stack trace, a pprof profile, or a GODEBUG line, even if they never say "debug". Does not cover speeding up code already known to be slow (see go-performance) or the fix once a race is identified (see go-concurrency).
---

# Go Troubleshooting

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`).
> `runtime/trace.FlightRecorder` Go 1.25+; `debug.SetCrashOutput` Go 1.23+;
> `testing/synctest` Go 1.25+; `GOTRACEBACK`, `GODEBUG`, `net/http/pprof`,
> and `go tool pprof` are long-standing.

Follow evidence, not intuition. A bug report names a **symptom**; the job is to
find the **mechanism** and fix it once, where every path goes through. The
tools below exist so that no step of that is a guess.

## Resource Routing

- `references/DIAGNOSTIC-TOOLS.md` - Read when capturing a profile, a goroutine dump, or a trace, or when setting `GODEBUG`/`GOTRACEBACK`; full command reference for `pprof`, `trace`, `dlv`, and the race detector.
- `references/SYMPTOM-CATALOG.md` - Read when a symptom is in hand and the cause is not: symptom → likely mechanisms → the command that confirms each → the skill that owns the fix.

## Method

```
1. Reproduce   — smallest input, fixed seed, one process. No repro → instrument first.
2. Capture     — the artifact for this symptom (table below), before changing anything.
3. Read it     — the running goroutine, the blocked goroutines, the top allocators.
4. Hypothesize — one mechanism that explains every observation, not just the loudest.
5. Confirm     — a command whose output differs if the hypothesis is wrong.
6. Fix once    — at the shared function, not at the call site the ticket named.
7. Pin it      — a test that fails on the old code. Then run the full gate.
```

Change one variable per cycle. A fix applied before step 5 that "seems to help"
is the most expensive outcome: the symptom moves, the mechanism stays.

> **Normative**: Do not fix bugs you *notice* along the way in the same diff.
> Record them; a debugging diff must be attributable to one mechanism.

---

## Symptom → First Capture

| Symptom | Capture first | Read for |
|---|---|---|
| Panic | Full trace with `GOTRACEBACK=all` | The `[running]` goroutine; the frame **above** `runtime.` |
| Hang / no response | Goroutine dump: `kill -QUIT <pid>` or `/debug/pprof/goroutine?debug=2` | Goroutines in `chan receive`, `select`, `sync.Mutex.Lock`, `semacquire` |
| Goroutine leak | `/debug/pprof/goroutine?debug=1` twice, minutes apart | A stack whose count grows; the `created by` frame names the leak |
| Memory grows | `go tool pprof -sample_index=inuse_space heap` twice; `GODEBUG=gctrace=1` | Top of `-top`; live heap after each GC in `gctrace` |
| CPU pegged | `go tool pprof -seconds=30 http://.../debug/pprof/profile` | `-top -cum`; a spin loop shows as one frame near 100% |
| Latency spikes | `go tool trace` or `FlightRecorder` snapshot at the spike | GC pauses, goroutine scheduling gaps, blocking syscalls |
| Flaky test | `go test -race -count=50 -shuffle=on -run 'Name$' ./pkg` | The first `WARNING: DATA RACE`, or a timing dependency |
| Wrong result | A table test pinning the input; `dlv test` on that case | The first variable that differs from expectation |
| Works locally, fails in CI | `go env`, `GOOS`/`GOARCH`, `GOFLAGS`, `-race` on, CPU count | `runtime.GOMAXPROCS`, tmp paths, timeouts sized for a faster machine |

`references/SYMPTOM-CATALOG.md` expands each row into mechanisms.

---

## Reading a Panic

```text
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x4a2f1e]

goroutine 42 [running]:
main.(*Server).handle(0xc000010000, {0x0, 0x0})   ← first non-runtime frame: the bug
        /app/server.go:88 +0x1e
main.main.func1()                                  ← how it got there
        /app/main.go:30 +0x2a
created by main.main in goroutine 1                ← who started this goroutine
```

- Only the `[running]` goroutine panicked; the others are context.
- `{0x0, 0x0}` in the argument list is a nil interface or slice — the receiver
  or argument that was nil is often visible right there.
- `concurrent map writes` and `send on closed channel` are **runtime**
  detections of a race: the fix is the synchronization, not a nil check.
  [go-concurrency](../go-concurrency/SKILL.md) owns it.
- A panic inside a `defer` chain hides the original: `GOTRACEBACK=all`, or
  `recover()` and print `%+v` before re-panicking, to see the first one.
- A stack ending in `runtime.goexit` with no user frames is a goroutine that
  panicked after its parent returned — look at `created by`.

`debug.SetCrashOutput(f, debug.CrashOptions{})` (Go 1.23+) writes the trace to
a file in production, where stderr is often lost.

---

## Hangs and Deadlocks

`fatal error: all goroutines are asleep - deadlock!` is the easy case — the
runtime saw *every* goroutine blocked. A server hangs with most goroutines
idle, so the runtime says nothing; take the dump yourself and group by stack:

```bash
curl -s 'http://127.0.0.1:6060/debug/pprof/goroutine?debug=2' > goroutines.txt
grep -c '^goroutine ' goroutines.txt              # how many
grep -A1 '^goroutine ' goroutines.txt | grep -v '^--' | sort | uniq -c | sort -rn | head
```

Hundreds of goroutines in the same `chan receive` frame means the sender is
gone or never started; one goroutine in `Lock` with another in the same
package's `Lock` on a different line is lock-order inversion; `select` with
no `ctx.Done()` case never returns when the caller gives up. `go test -timeout
30s` prints the same dump when a test hangs.

---

## Leaks

Goroutine leaks and heap growth are the same investigation: **what is still
referenced, and who created it**.

- Goroutines: diff two `debug=1` dumps. The growing stack's `created by` frame
  is the leak; the fix is a `ctx.Done()` case, a closed channel, or a bounded
  worker set ([go-concurrency](../go-concurrency/SKILL.md)).
- Heap: `-sample_index=inuse_space` shows what is *live*; `alloc_space` shows
  churn (a performance question, [go-performance](../go-performance/SKILL.md)).
  A growing `inuse` at a `map` insert or `append` with no matching delete is
  a cache with no eviction. `GODEBUG=gctrace=1` confirms: the live-heap number
  after each GC (`->N MB`) climbs instead of oscillating.
- RSS without heap growth: `runtime.MemStats.Sys` vs `HeapInuse` — goroutine
  stacks, CGO allocations, or the runtime holding freed spans. `GOMEMLIMIT`
  makes the GC work harder; it does not fix a leak.
- Resource leaks: `lsof -p <pid>` climbing with `http.Response.Body` or
  `sql.Rows` never closed; the `bodyclose` and `sqlclosecheck` linters in the
  gate catch the static cases.

---

## Flaky Tests

A test that fails one run in fifty has a real bug or a real timing
dependency; neither is fixed by `retry`.

```bash
go test -race -count=100 -shuffle=on -failfast -run 'TestName$' ./pkg
```

- `-race` finding → the race is the bug, not the flake.
- Passes alone, fails in the package → shared state: package-level `var`,
  `t.Setenv` without `t.Parallel` guard, a fixed port, a shared temp path.
- Depends on `time.Sleep` → rewrite under `testing/synctest` (Go 1.25+), where
  virtual time makes the ordering deterministic ([go-testing](../go-testing/SKILL.md)).
- Fails only in CI → fewer CPUs, slower disk, `-race` on, a different
  timezone or locale. `runtime.GOMAXPROCS(0)` in the log tells you which.

---

## Production Safety

- Mount `net/http/pprof` on an internal listener (`127.0.0.1:6060` or a
  sidecar-only port), never the public mux — profiles contain heap contents.
  [go-security](../go-security/SKILL.md) owns why.
- A 30-second CPU profile costs ~5% overhead; a heap profile is a snapshot.
  Both are safe under load. `go tool trace` is heavier; keep captures short.
- `runtime/trace.FlightRecorder` (Go 1.25+) keeps the last few seconds in a
  ring buffer so a spike can be captured **after** it happens.

---

## Hand Off the Fix

Once the mechanism is named, the fix belongs to its owner: races and leaks to
[go-concurrency](../go-concurrency/SKILL.md), lost cancellation to
[go-context](../go-context/SKILL.md), nil-interface and boundary bugs to
[go-defensive](../go-defensive/SKILL.md), allocation churn to
[go-performance](../go-performance/SKILL.md). Close with the
[go-linting](../go-linting/SKILL.md) gate — a debugging change is still a change.

> **Validation**: The regression test from step 7 fails against the
> pre-fix commit (`git stash` / `git worktree` to prove it) and passes after.
> Report the reproduction rate before and after (`-count=N`), not "seems fixed".
> Lead the report with the mechanism and the output that confirmed it;
> [go-style-core](../go-style-core/SKILL.md#how-much-to-say) owns the length.

---

## Related Skills

- **Fixing what you found**: See [go-concurrency](../go-concurrency/SKILL.md) for races, leaks, and channel deadlocks once identified
- **Cancellation**: See [go-context](../go-context/SKILL.md) when a goroutine outlives its request or a timeout never fires
- **Making it faster**: See [go-performance](../go-performance/SKILL.md) when the profile shows churn rather than a leak, and for benchmark methodology
- **Deterministic time in tests**: See [go-testing](../go-testing/SKILL.md) for `synctest`, `t.Context`, and table tests that pin a repro
- **Exposure of debug endpoints**: See [go-security](../go-security/SKILL.md) before mounting pprof or expvar anywhere reachable
- **Gate**: See [go-linting](../go-linting/SKILL.md) for `-race`, `bodyclose`, `sqlclosecheck`, and the rest of the verification gate
