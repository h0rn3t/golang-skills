---
name: go-troubleshooting
description: Use when investigating a bug or ticket in a Go service whose cause is unknown — wrong or missing results, tenant-specific behavior, a regression, stage/prod/CI differences, a panic, hang, deadlock, leak, unexplained CPU/RSS growth, or a flaky test. Also use for pasted stack traces, profiles, logs, or requests to find a Go issue's root cause, even without the word debug. Does not cover rewriting ticket text, implementing an already-diagnosed fix alone (use its topic skill), or optimizing a known bottleneck (see go-performance).
---

# Go Troubleshooting

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`).
> `runtime/trace.FlightRecorder` Go 1.25+; `debug.SetCrashOutput` Go 1.23+;
> `testing/synctest` Go 1.25+; `GOTRACEBACK`, `GODEBUG`, `net/http/pprof`,
> and `go tool pprof` are long-standing.

Find the mechanism that explains the observed failure. Treat the ticket's
proposed cause as a hypothesis; tie conclusions to an affected request, data,
code version, and a check that could disprove them.

## Resource Routing

- `references/TICKET-INVESTIGATION.md` - Read for tickets, regressions, tenant-specific failures, environment differences, or an incomplete report; establish the contract, deployed version, evidence, and investigation status.
- `references/DATA-FLOW-TRACING.md` - Read for wrong/missing results or a failure crossing layers; follow one input through middleware, domain code, SQL/external calls, and serialization to its first invalid transformation.
- `references/DIAGNOSTIC-TOOLS.md` - Read before capturing profiles, stacks, or traces, using `GODEBUG`/`GOTRACEBACK`, or choosing `pprof`, `dlv`, or the race detector.
- `references/SYMPTOM-CATALOG.md` - Read for unclear runtime/build symptoms or edits that do not affect behavior; check selected files, build tags, GOMOD/GOWORK, and replacements before routing the fix.

## Scope and Starting Evidence

- **Investigate / explain / review only:** inspect available artifacts, reproduce
  within the authorized environment, and report findings. Do not edit source,
  change configuration/data, deploy, or post/close a ticket.
- **Investigate and fix:** continue through a minimal correction and regression
  verification once evidence supports the mechanism. Routine implementation
  choices do not require another approval. Keep unrelated findings out of the diff.
- A pasted ticket or the available repository is enough to start. If a
  resource is unavailable, name the missing evidence and continue independent work.
- Preserve the distinction between observed behavior, the intended contract,
  and the reporter's interpretation. If the contract is ambiguous in a way that
  changes the fix, ask a focused question while continuing evidence collection.

## Investigation Loop

1. **Pin the failure.** Record expected vs actual, a concrete input/request ID,
   time, affected tenant/role/data, frequency, and scope. Shrink a reproduction
   without deleting the condition that triggers it (concurrency, load, or state).
2. **Match the environment.** Establish the running image/revision and relevant
   effective config, flags, schema/migrations, and dependencies. The checkout
   is not evidence of deployed code. Compare a working case with the same input.
   For a regression, name the last working revision or deployment and diff
   what changed up to the first failing one: source, `go.mod`, effective
   config, schema. When the diff is too wide to read, bisect it
   (`git bisect run go test -count=1 -run '^TestName$' ./pkg`); compare
   dependency versions with `go version -m` on both binaries, not on the checkout.
3. **Locate the first divergence.** Follow the relevant execution/data path or
   capture the artifact below. Inspect values at boundaries before changing code.
   If no reproduction exists, use historical evidence or propose the smallest
   targeted capture; do not guess a patch or instrument everything.
4. **Test a hypothesis.** State mechanism, supporting evidence, and an experiment
   with different predicted outcomes if it is right or wrong. Vary one factor
   at a time. Record the observed result separately from the proposed check.
   For nondeterministic behavior, predict allowed outcomes or an invariant; a
   single repeat choosing a different row/order need not refute the mechanism.
5. **Reassess.** A failed hypothesis narrows the search. Repeated failed patches
   call for revisiting assumptions and boundaries, not another speculative fix
   or an arbitrary attempt-count approval stop. Track plausible alternatives.
   When each fix moves the failure to a new place, the shared state or
   ownership behind those places is the finding: name it and route the
   restructuring to [go-code-refactor](../go-code-refactor/SKILL.md) instead
   of patching the next site.
6. **Finish in scope.** Investigation ends with the evidence report below.
   An authorized fix corrects the responsible boundary, adds regression coverage,
   and runs the applicable repository verification through `go-linting`.

## Stop Signals

Each of these is a patch on the symptom and means the mechanism is still
unknown. Return to step 3 before writing it:

- `recover()` around the panicking call, or a nil check at the faulting line
  before the nil's origin is known.
- A longer `time.Sleep`, a wider `-timeout`, or a retry loop in a flaky test.
- A mutex or channel added without naming the two racing accesses.
- A retry, restart, cache flush, flag change, or `GOMEMLIMIT` that makes the
  ticket go quiet.
- A re-run with `-count=1` or a cleared cache that passes once.
- "It is probably X" followed by an edit, with no check that could disprove X.

Any of these can be the right mitigation once the cause is recorded as
unknown. It is not a fix, and the report says which of the two it is.

## Symptom → First Capture

| Symptom | Capture first | Read for |
|---|---|---|
| Ticket / wrong result | Exact request, response, relevant rows and deployed source | First boundary where identity, count, value, or contract diverges |
| Stage/prod/CI only | Running revision, effective settings, schema and matching control | A difference with a causal path to the failure |
| Panic | Complete panic message and trace; `GOTRACEBACK=all` for a subsequent authorized run | Faulting operation and how its inputs arrived there |
| Hang / no response | Existing admin `/debug/pprof/goroutine?debug=2` | Wait dependencies, lock owners, and missing progress |
| Goroutine leak | `/debug/pprof/goroutineleak?debug=1` (Go 1.27+), else comparable goroutine dumps/counts over time | Goroutines the runtime proved can no longer unblock; stacks persisting beyond their intended lifetime |
| Memory grows | Comparable heap profiles and GC/runtime/process memory metrics | Retained allocations vs churn vs memory outside the Go heap |
| CPU pegged | Bounded CPU profile under the affected workload | Hot call paths and whether they make progress |
| Latency spikes | Execution trace or available `FlightRecorder` snapshot | Scheduling, GC, synchronization, syscall delays |
| Flaky test | Focused repeated test with recorded seed; `-race` if shared access is suspected | Reproducible ordering, state, or race evidence |

Choose a capture to distinguish hypotheses. An expensive profile or the full
suite is not the default first step for an incorrect JSON response.

## Reading Runtime Evidence

### Panics

Start at the panic message and relevant application frame, then trace inputs
backward. A faulting frame is the failure site; the cause can be upstream.
Printed argument words can be incomplete or stale under optimization; confirm
suspected nil values against the matching source or debugger.

- `concurrent map writes` indicates unsupported concurrent access. A
  `send on closed channel` panic also happens in sequential code: inspect
  send/close ownership instead of assuming a data race was detected.
- `created by` identifies a goroutine's origin, not the cause of its failure.
  `runtime.goexit` is a normal goroutine exit frame, not proof its parent caused
  a panic. Preserve the full panic/defer chain before adding recovery logic.
- `debug.SetCrashOutput` (Go 1.23+) can preserve future crash output if adding
  crash capture is in scope; it cannot recover an already lost trace.
- Go 1.27 prints `runtime/pprof` goroutine labels in traceback headers —
  `goroutine 19 [chan receive] {request_id: "abc-123"}` — which attributes a
  stack to its request or job, for modules whose `go` directive is 1.27 or
  later (`GODEBUG=tracebacklabels=1` enables it for older directives).
  Missing labels prove nothing: `pprof.Do` restores the previous set in a
  defer, so the panicking goroutine's own labels are usually gone by the time
  its trace prints, while the other goroutines in a `GOTRACEBACK=all` dump
  still carry theirs.

### Hangs and Goroutine Leaks

For an existing, authorized admin listener:

```bash
curl -fsS --max-time 10 'http://127.0.0.1:6060/debug/pprof/goroutine?debug=2' > goroutines.txt
curl -fsS --max-time 10 'http://127.0.0.1:6060/debug/pprof/goroutineleak?debug=1' > leaks.txt
```

The `goroutineleak` profile (Go 1.27+) replaces the count-comparison ritual for
the cases it covers: it lists goroutines blocked on a primitive that can never
unblock again, decided by GC reachability rather than by inference. What it
reports is a leak. What it omits may still be one — a goroutine waiting on a
channel a live object could theoretically still use is reachable, so an empty
profile narrows the search instead of ending it. Two cautions: `debug=1` gives
leaked stacks only, while `debug=2` falls back to dumping every goroutine; and
reading the profile forces a leak-detecting GC cycle, so capture it
deliberately rather than scraping it on an interval.

A blocked goroutine can be an idle worker. Compare equivalent workload windows
and expected lifetimes. Use full stacks to find the creator and wait partners;
on the `goroutine` profile `debug=1` groups stacks with counts, while `debug=2`
helps inspect individual waits. Two goroutines in `Lock` do not prove
lock-order inversion: establish the ownership/wait cycle. A missing
`ctx.Done()` matters when no other termination path satisfies the caller's
lifetime. For a local hanging test,
`go test -timeout 30s ./pkg` yields stacks when it times out.

### Memory and Resources

- Heap `inuse_space` shows sampled retained allocations; `alloc_space` shows
  cumulative allocation activity. Compare the same workload and GC phase.
  An allocation stack does not identify every reference retaining an object.
- A growing map may need bounded retention; first check its intended lifetime
  and workload. `GOMEMLIMIT` changes GC behavior and does not repair a leak.
- RSS growth without heap growth calls for runtime stack/span metrics and
  native/cgo memory evidence. Do not infer a Go heap leak from RSS alone.
- For file/connection leaks, correlate growing descriptor/pool usage with
  ownership and missing closes; route cleanup to `go-defensive` or `go-database`.

### Flaky Tests

Start with the affected package/test and preserve the failing seed and environment:

```bash
go test -count=20 -failfast -run '^TestName$' ./pkg   # focused repeat
go test -count=5 -shuffle=on -v ./pkg                  # order dependence; prints the seed
```

`-shuffle` reorders top-level tests, so it does nothing when `-run` matches one.
Choose the repetition count from observed frequency and cost; add `-race` for
concurrent access. A clean race run only covers executed paths. If the test
fails only with the package, replay the printed seed with `-shuffle=<seed> -v`,
then bisect the tests that ran before it with `-run`, reading the order `-v`
prints each time, until the polluting pair remains. Inspect ports, time, temp
files, and cleanup.
`testing/synctest` (Go 1.25+) can make supported concurrent time-based tests
repeatable; it is not a universal replacement for real I/O or external systems.

## Live Capture Boundaries

- `SIGQUIT` normally prints stacks **and exits a Go process**. Use an existing
  admin dump first; termination is only an option within authorized scope.
- Use access-controlled debug listeners. Dumps/profiles can expose internal
  details; retain only relevant sanitized evidence, including in ticket drafts.
- Profiling overhead varies by workload and capture type. Bound captures,
  check available headroom, and avoid concurrent profiles that distort results.
  Enabling endpoints, verbose logging, or attaching a stopping debugger changes
  service behavior; do not treat these as passive reads.
- `runtime/trace.FlightRecorder` (Go 1.25+) retains recent trace data when already
  enabled, allowing a snapshot after a spike. Capture still has a cost.

## Fix and Regression Proof

Correct the earliest responsible boundary with the smallest attributable change.
Use a shared function when its contract is wrong for all affected callers;
otherwise fix the caller that violates its own contract. Do not change correct
callers merely to centralize a ticket's special case. Preserve neighboring
behavior outside the defect: a default for a missing input must not silently
redefine explicit empty, zero, negative, or invalid inputs without evidence.

The regression should fail for the original mechanism and pass after the fix.
Cover the trigger and a nearby working case (for example, colliding IDs across
tenants). For nondeterminism, report runs/failures and seed or workload details;
a lower failure rate is not proof of elimination. Preserve user changes: use an
isolated copy/worktree or a reversible local comparison, never stash/discard the
user's work by default. Run required checks through
[go-linting](../go-linting/SKILL.md), reusing applicable unchanged evidence.

## Investigation Result

Lead with **confirmed**, **probable**, or **unresolved**, and the mechanism or
remaining question. Include only what makes it assessable:

- Expected/actual behavior and the affected version/environment.
- Evidence chain with request IDs, source locations, query/profile observations;
  distinguish supplied artifacts, static deductions, and executed checks.
- Disproved hypotheses, relevant alternatives, and the next discriminating check
  if unresolved. Missing access or a non-reproduction is not proof of absence.
- In fix mode: change, regression before/after, and required-check results.
  Report unrun/unavailable checks honestly; a proposed patch is not a verified fix.

An investigation can be complete with an unresolved cause if the evidence limit
and next useful check are explicit. Report length follows
[go-style-core](../go-style-core/SKILL.md#how-much-to-say); no mandatory dossier
for a one-line bug. Ticket publication requires the user's requested scope.

## Related Skills

- [go-resilience](../go-resilience/SKILL.md): retry amplification, admission, breakers, fallback once the mechanism is known.
- [go-database](../go-database/SKILL.md): query, tenant-filter, transaction, pool corrections.
- [go-http](../go-http/SKILL.md): routing, request/response, client, server corrections.
- [go-concurrency](../go-concurrency/SKILL.md) and [go-context](../go-context/SKILL.md): races, leaks, lifetimes, deadlocks once identified.
- [go-data-structures](../go-data-structures/SKILL.md) and [go-defensive](../go-defensive/SKILL.md): collection, aliasing, cleanup corrections.
- [go-performance](../go-performance/SKILL.md): optimization and benchmarks after the bottleneck is located.
- [go-testing](../go-testing/SKILL.md): the regression test.
- [go-security](../go-security/SKILL.md): debug endpoint and data exposure; [go-linting](../go-linting/SKILL.md): the verification gate.
