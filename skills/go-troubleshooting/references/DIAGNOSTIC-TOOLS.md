# Diagnostic Tools Reference

> Sources: `go doc runtime`, `go doc runtime/pprof`, `go doc runtime/trace`, `go doc testing`; `$GOROOT/doc/godebug.md`; go.dev/doc/diagnostics; go.dev/doc/articles/race_detector; github.com/go-delve/delve docs
> Authority: normative for flags and env vars; advisory for the workflows
> Minimum Go: 1.27 baseline; per-item versions inline
> Last verified: 2026-10-01

Commands to run, grouped by what they capture. Every example assumes
an existing, access-controlled `net/http/pprof` listener on `127.0.0.1:6060`
for the HTTP forms. Enabling a listener is a configuration/code change, not a
read-only diagnostic step. Capture one profile at a time with bounded duration;
[profiling overhead](https://go.dev/doc/diagnostics) depends on the workload.

## Contents

- [Environment knobs](#environment-knobs)
- [Stack dumps](#stack-dumps)
- [pprof](#pprof)
- [Execution trace](#execution-trace)
- [Race detector](#race-detector)
- [Delve](#delve)
- [Compiler and runtime introspection](#compiler-and-runtime-introspection)
- [Test flags for debugging](#test-flags-for-debugging)

---

## Environment knobs

| Variable | Effect | Use when |
|---|---|---|
| `GOTRACEBACK=single` (default) | Panic prints the crashing goroutine only | — |
| `GOTRACEBACK=all` | Every user goroutine | A panic whose cause is in another goroutine |
| `GOTRACEBACK=system` | Also runtime goroutines and frames | Suspected runtime or cgo involvement |
| `GOTRACEBACK=crash` | `system` + core dump (`ulimit -c unlimited`) | Post-mortem with `dlv core` |
| `GODEBUG=gctrace=1` | One line per GC: heap before→after, live heap, pause | Leak vs. churn; GC frequency |
| `GODEBUG=schedtrace=1000` | Scheduler state every second | Goroutines runnable but not running |
| `GODEBUG=scheddetail=1,schedtrace=1000` | Per-P and per-M detail | Starvation, `GOMAXPROCS` mismatch |
| `GODEBUG=asyncpreemptoff=1` | Disable async preemption | Ruling preemption in or out for a tight-loop hang |
| `GODEBUG=inittrace=1` | Time and allocation per package `init` | Slow startup |
| `GODEBUG=http2debug=2` | HTTP/2 frame log | Client/server stall on HTTP/2 |
| `GOMAXPROCS=1` | One P | Making a race or ordering bug reproduce |
| `GOMEMLIMIT=500MiB` | Soft limit on Go runtime-managed memory (Go 1.19+) | Tuning memory/GC tradeoffs; excludes native allocations and is not an RSS cap or leak diagnosis |
| `GOFLAGS=-race` | Race detector on for every build in the shell | CI parity locally |

`GODEBUG` accepts a comma-separated list. The Go-version-compat settings
(`GODEBUG=panicnil=1`, `httpmuxgo121=1`, ...) and the release that changed
each default are in `$(go env GOROOT)/doc/godebug.md`; the defaults a build
carries show in `go version -m <binary>` (`build DefaultGODEBUG=…`) or
`go list -f '{{.DefaultGODEBUG}}' ./cmd/app`. A stale one in a Dockerfile is a
finding for [go-security](../../go-security/SKILL.md).

---

## Stack dumps

```bash
curl -fsS --max-time 10 'http://127.0.0.1:6060/debug/pprof/goroutine?debug=1' # grouped counts
```

The `goroutineleak` profile follows the toolchain that built the binary, not
the `go` directive: Go 1.27 toolchain (1.26 needs
`GOEXPERIMENT=goroutineleakprofile`; otherwise the endpoint answers 404
`Unknown profile` and `pprof.Lookup` returns nil). `debug=1` lists leaked
stacks only; `debug=2` falls back to dumping every goroutine, the leaked ones
marked `(leaked)`; the plain endpoint returns a pprof profile for
`go tool pprof`. Every read runs a leak-detecting GC cycle first, so capture it
deliberately rather than scraping it on an interval, and `Profile.Count()`
reports the last cycle's number and stays at zero until the profile has been
read once — take the total from the output.

Programmatic (requires existing instrumentation or an authorized code change):

```go
buf := make([]byte, 1<<20)
n := runtime.Stack(buf, true)            // true = all goroutines
os.Stderr.Write(buf[:n])

pprof.Lookup("goroutine").WriteTo(w, 2)  // same as ?debug=2, from inside the process
if p := pprof.Lookup("goroutineleak"); p != nil { // nil on a 1.26 toolchain without the GOEXPERIMENT
    p.WriteTo(w, 1)                      // leaked goroutines only
}
```

Grouping a `debug=2` dump by top frame:

```bash
awk '/^goroutine /{getline; print}' goroutines.txt | sort | uniq -c | sort -rn | head -20
```

---

## pprof

### Capture

```bash
go tool pprof http://localhost:6060/debug/pprof/block                 # needs runtime.SetBlockProfileRate
go tool pprof http://localhost:6060/debug/pprof/mutex                 # needs runtime.SetMutexProfileFraction
curl -s -o heap.pb.gz http://localhost:6060/debug/pprof/heap          # save for later / diff
go test -memprofilerate=1 -memprofile mem.out ./pkg   # every allocation, slow but exact
```

Enable block/mutex profiling at startup, not permanently in production:

```go
runtime.SetBlockProfileRate(1)          // every blocking event; 10_000 (ns) for sampling
runtime.SetMutexProfileFraction(5)      // 1 in 5 contention events
```

### Read

```bash
go tool pprof -sample_index=alloc_objects -top mem.out   # object count: small-object churn
go tool pprof -base heap1.pb.gz heap2.pb.gz         # what grew between two snapshots
```

Heap profiles are as of the last GC; call `runtime.GC()` before a programmatic
`WriteHeapProfile` for an exact live set.

---

## Execution trace

The trace shows *time*: when each goroutine ran, blocked, was scheduled, and
what the GC did — the tool for latency spikes and "the CPU is idle but the
request is slow".

```bash
curl -s -o trace.out 'http://localhost:6060/debug/pprof/trace?seconds=5'
go test -trace trace.out ./pkg
go tool trace trace.out                 # opens the browser UI
go tool trace -pprof=sync trace.out > sync.pprof   # extract a blocking profile from a trace
```

Programmatic, and the flight recorder for after-the-fact capture:

```go
trace.Start(f); defer trace.Stop()      // runtime/trace, whole-program

fr := trace.NewFlightRecorder(trace.FlightRecorderConfig{MinAge: 5 * time.Second}) // Go 1.25+
fr.Start()
// ... on a latency alarm:
fr.WriteTo(f)                            // the last ~5 s of trace, from the ring buffer
```

In the UI: "View trace" for the timeline; "Goroutine analysis" for
per-function scheduling latency; "Network/Sync/Syscall blocking profile" for
where goroutines waited. Regions and tasks (`trace.WithRegion`,
`trace.NewTask`) label your own spans so a request is findable.

---

## Race detector

```bash
go test -race ./...
go build -race -o app . && ./app         # production-shaped binary; 2–20× slower, 5–10× memory
GORACE="halt_on_error=1 log_path=/tmp/race" ./app   # stop at first race; write reports to files
```

Report anatomy: two stacks (`Write at` / `Previous read at`), each with the
goroutine that did it and, below, where that goroutine was **created**. The
creation stacks only identify the goroutines; the fix is a happens-before edge
(mutex, channel, atomic) between the two racing accesses —
[go-concurrency](../../go-concurrency/SKILL.md) owns it. The detector only
sees executed paths: a race in an untested branch stays hidden until a test
executes that branch; `-count=N` repeats the same paths and only varies their
interleaving.

---

## Delve

```bash
dlv debug ./cmd/app -- --flag value      # build and run under the debugger
dlv test ./pkg -- -test.run '^TestName$' # a single test
dlv attach <pid>                         # a running process (pauses it)
dlv core ./app core.1234                 # post-mortem from GOTRACEBACK=crash
dlv exec ./app                           # a prebuilt binary (build with -gcflags=all=-N\ -l)
```

Delve is the tool for a *deterministic* wrong result; for races and hangs, the
dumps and profiles above are faster and do not perturb timing.

---

## Compiler and runtime introspection

```bash
go build -gcflags='-d=checkptr' ./...                                   # unsafe.Pointer misuse (on by default under -race)
go version -m ./app                                                     # module versions baked into a binary
```

```go
var m runtime.MemStats
runtime.ReadMemStats(&m)                 // stop-the-world; fine for a debug endpoint, not a hot path
metrics.Read(samples)                    // runtime/metrics: cheap, per-metric, the modern form
runtime.NumGoroutine()                   // plot it
debug.SetTraceback("all")                // programmatic GOTRACEBACK
debug.SetCrashOutput(f, debug.CrashOptions{})   // Go 1.23+: crash trace to a file
debug.ReadBuildInfo()                    // module path, VCS revision, -race, settings
```

---

## Test flags for debugging

```bash
go test -count=100 -run '^TestName$' ./pkg | grep -c '^--- FAIL'   # reproduction rate; -failfast would stop at 1
go build ./...                                    # vendor drift: fails "inconsistent vendoring" under the default -mod=vendor
```

`-count=N` with `-shuffle=on` runs one shuffled order N times: the order is
drawn once per invocation, so vary it across invocations. The vendor
check needs `vendor/` and no `-mod` in `go env GOFLAGS`. A module-mode
comparison (`GOFLAGS=-mod=mod`) rewrites `go.mod` and `go.sum`; run it only in
an isolated worktree (`git worktree add ../cmp HEAD`), never in the tree under
investigation.

`t.Context()` (Go 1.24+) is canceled when the test ends — a goroutine still
running after that is what `-race` and goroutine-leak checks catch.
[go-testing](../../go-testing/SKILL.md) owns the test shape.
