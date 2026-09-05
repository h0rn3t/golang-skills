# Diagnostic Tools Reference

> Sources: `go doc runtime`, `go doc runtime/pprof`, `go doc runtime/trace`, `go doc testing`; go.dev/doc/diagnostics; github.com/go-delve/delve docs
> Authority: normative for flags and env vars; advisory for the workflows
> Minimum Go: 1.27 baseline; per-item versions inline
> Last verified: 2026-09-02; signal, profiling-cost, and memory-limit notes rechecked 2026-09-05

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

The [Go GC guide](https://go.dev/doc/gc-guide#Memory_limit) defines which memory
the soft limit covers. Changing the limit does not establish why memory grows.

`GODEBUG` accepts a comma-separated list. The Go-version-compat settings
(`GODEBUG=panicnil=1`, `tlsrsakex=1`, ...) are documented under `go doc
runtime` → "godebug"; a stale one in a Dockerfile is a finding for
[go-security](../../go-security/SKILL.md).

---

## Stack dumps

```bash
curl -fsS --max-time 10 'http://127.0.0.1:6060/debug/pprof/goroutine?debug=2' > goroutines.txt
curl -fsS --max-time 10 'http://127.0.0.1:6060/debug/pprof/goroutine?debug=1' # grouped counts
go test -timeout 30s ./pkg             # on timeout: dump of every goroutine, then fail
```

[SIGQUIT normally dumps stacks and exits](https://pkg.go.dev/os/signal#hdr-Default_behavior_of_signals_in_Go_programs).
Only use `kill -QUIT <pid>` when termination is authorized; prefer the existing
admin endpoint for a process that must stay alive.

Programmatic (requires existing instrumentation or an authorized code change):

```go
buf := make([]byte, 1<<20)
n := runtime.Stack(buf, true)            // true = all goroutines
os.Stderr.Write(buf[:n])

pprof.Lookup("goroutine").WriteTo(w, 2)  // same as ?debug=2, from inside the process
```

Grouping a `debug=2` dump by top frame:

```bash
awk '/^goroutine /{getline; print}' goroutines.txt | sort | uniq -c | sort -rn | head -20
```

Wait states worth recognizing: `chan receive`, `chan send`, `select`,
`select (no cases)` (a `select {}` — intentional or a bug), `sync.Mutex.Lock`,
`sync.WaitGroup.Wait`, `semacquire`, `IO wait`, `sleep`, `syscall`. A goroutine
in `[chan receive, 47 minutes]` may be an idle worker. Confirm leakage by
comparing counts/stacks under equivalent workload and checking whether its
owner has finished and its intended lifetime has expired.

---

## pprof

### Capture

```bash
# Live process (HTTP)
go tool pprof -seconds=30 http://localhost:6060/debug/pprof/profile   # CPU
go tool pprof http://localhost:6060/debug/pprof/heap                  # heap (inuse by default)
go tool pprof http://localhost:6060/debug/pprof/allocs                # heap, alloc_space by default
go tool pprof http://localhost:6060/debug/pprof/block                 # needs runtime.SetBlockProfileRate
go tool pprof http://localhost:6060/debug/pprof/mutex                 # needs runtime.SetMutexProfileFraction
curl -s -o heap.pb.gz http://localhost:6060/debug/pprof/heap          # save for later / diff

# Tests and benchmarks
go test -cpuprofile cpu.out -memprofile mem.out -bench . ./pkg
go test -blockprofile block.out -mutexprofile mutex.out ./pkg
go test -memprofilerate=1 -memprofile mem.out ./pkg   # every allocation, slow but exact
```

Enable block/mutex profiling at startup, not permanently in production:

```go
runtime.SetBlockProfileRate(1)          // every blocking event; 10_000 (ns) for sampling
runtime.SetMutexProfileFraction(5)      // 1 in 5 contention events
```

### Read

```bash
go tool pprof -top -cum cpu.out                     # inclusive time, callers first
go tool pprof -top mem.out                          # inuse_space
go tool pprof -sample_index=alloc_space -top mem.out
go tool pprof -sample_index=alloc_objects -top mem.out   # object count: small-object churn
go tool pprof -list 'pkg.Func' cpu.out              # per-line inside one function
go tool pprof -peek 'runtime.mallocgc' cpu.out      # who calls this
go tool pprof -base heap1.pb.gz heap2.pb.gz         # what grew between two snapshots
go tool pprof -diff_base heap1.pb.gz heap2.pb.gz    # same, normalized
go tool pprof -http=:8081 cpu.out                   # flame graph, source view, browser
go tool pprof -focus 'handleOrder' -top cpu.out     # only stacks through this function
go tool pprof -ignore 'runtime\.' -top cpu.out      # hide a subtree
```

Sample indexes for heap profiles: `inuse_space` (live bytes — leaks),
`inuse_objects`, `alloc_space` (cumulative — churn), `alloc_objects`. Heap
profiles are as of the last GC; call `runtime.GC()` before a programmatic
`WriteHeapProfile` for an exact live set.

Interactive mode (`go tool pprof cpu.out`) accepts the same verbs: `top`,
`list`, `peek`, `web`, `weblist`, `disasm`, `tags`, `focus=`, `ignore=`.

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
go build -race -o app . && ./app         # production-shaped binary; 5–10× slower, ~10× memory
GORACE="halt_on_error=1 log_path=/tmp/race" ./app   # stop at first race; write reports to files
```

Report anatomy: two stacks (`Write at` / `Previous read at`), each with the
goroutine that did it and, below, where that goroutine was **created**. The
fix is the synchronization between those two creation sites —
[go-concurrency](../../go-concurrency/SKILL.md) owns it. The detector only
sees executed paths: a race in an untested branch stays hidden, so
`-count=N -shuffle=on` widens coverage.

---

## Delve

```bash
go install github.com/go-delve/delve/cmd/dlv@latest

dlv debug ./cmd/app -- --flag value      # build and run under the debugger
dlv test ./pkg -- -test.run 'TestName$'  # a single test
dlv attach <pid>                         # a running process (pauses it)
dlv core ./app core.1234                 # post-mortem from GOTRACEBACK=crash
dlv exec ./app                           # a prebuilt binary (build with -gcflags=all=-N\ -l)
```

Inside:

```text
break pkg.Func            b main.(*Server).handle:12
condition 1 id == 42      # break only when
continue / next / step / stepout
print v   / p v           # values; p -v v for full nested
locals / args / goroutines / goroutine 42 / bt / frame 2
watch -w s.count          # hardware watchpoint on a write
on 1 print id             # attach a command to a breakpoint
```

`dlv dap` speaks the Debug Adapter Protocol for editors. Delve is the tool
for a *deterministic* wrong result; for races and hangs, the dumps and
profiles above are faster and do not perturb timing.

---

## Compiler and runtime introspection

```bash
go build -gcflags='-m' ./pkg 2>&1 | grep -E 'escapes|moved to heap'   # escape analysis
go build -gcflags='-m -m' ./pkg                                         # with reasons
go build -gcflags='-d=checkptr' ./...                                   # unsafe.Pointer misuse (on by default under -race)
go vet ./...                                                            # copylocks, lostcancel, printf, waitgroup, unusedresult
go tool nm -size -sort size ./app | head                                # what makes the binary big
go version -m ./app                                                     # module versions baked into a binary
GOSSAFUNC=Func go build ./pkg                                           # ssa.html for one function
```

```go
var m runtime.MemStats
runtime.ReadMemStats(&m)                 // stop-the-world; fine for a debug endpoint, not a hot path
metrics.Read(samples)                    // runtime/metrics: cheap, per-metric, the modern form
runtime.NumGoroutine()                   // plot it; a monotone rise is a leak
debug.SetTraceback("all")                // programmatic GOTRACEBACK
debug.SetCrashOutput(f, debug.CrashOptions{})   // Go 1.23+: crash trace to a file
debug.ReadBuildInfo()                    // module path, VCS revision, -race, settings
```

---

## Test flags for debugging

```bash
go test -run 'TestName$' -v ./pkg                # exact match; -v for t.Log output
go test -count=100 -failfast -run 'TestName$'    # reproduction rate
go test -shuffle=on ./pkg                         # order dependence; prints the seed
go test -shuffle=1712345678 ./pkg                 # replay that order
go test -race -count=20 ./pkg
go test -timeout 30s ./pkg                        # hang → goroutine dump
go test -cpu 1,2,8 ./pkg                          # GOMAXPROCS sweep
go test -run TestName -args -my.flag=1            # flags to the test binary
go test -c -o pkg.test ./pkg && ./pkg.test -test.run TestName   # run the binary directly / under dlv
go test -json ./... | go run gotest.tools/gotestsum@latest --raw-command -- cat   # structured output
GOFLAGS=-mod=mod go test ./...                    # rule out vendor drift
```

`t.Context()` (Go 1.24+) is canceled when the test ends — a goroutine still
running after that is what `-race` and goroutine-leak checks catch.
[go-testing](../../go-testing/SKILL.md) owns the test shape.
