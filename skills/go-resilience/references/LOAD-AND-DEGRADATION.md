# Load Admission and Degradation

> Sources: [Go rate limiter](https://pkg.go.dev/golang.org/x/time/rate); [Google SRE overload](https://sre.google/sre-book/handling-overload/); [Circuit breaker pattern](https://learn.microsoft.com/en-us/azure/architecture/patterns/circuit-breaker); local `../../go-concurrency/SKILL.md`
> Authority: normative for Go API behavior; advisory for pattern selection; project policy for capacity and fallback contracts
> Last verified: 2026-09-05

## Bound What Actually Accumulates

A semaphore limits work holding a permit, not goroutines waiting to acquire it.
Admit work before spawning or use a bounded worker/queue design. Specify both
active and waiting capacities and the behavior when full: immediate rejection,
bounded wait, or durable upstream backpressure/redelivery. Bound queued bytes as
well as item count when payload sizes vary materially.

Do not acknowledge a required event before its ownership is durably accepted.
If the source must not lose work, use the existing broker's pause/redelivery
contract or durable intake; an in-memory overflow drop changes the requirement.
Backpressure must reach a source that can actually slow down. An unbounded local
queue simply relocates the failure.

Admission waits obey the caller or durable worker lifetime. Avoid starting a
goroutine solely to make an uncancelable wait look cancelable; the waiter can
leak. Coordinate shutdown with [go-concurrency](../../go-concurrency/SKILL.md)
and [go-http](../../go-http/SKILL.md), preserving acknowledged work's contract.

Separate a logical-work bound from an attempt-concurrency bound. Retry backoff
must not occupy a dependency connection or per-attempt semaphore slot; release
it before sleeping. Keep sleeping/requeued logical work bounded elsewhere.
Acquire fresh attempt admission when retrying, and release on every exit path.

## Rate, Concurrency, and Scope

| Bound | Unit | What it protects |
|---|---|---|
| Rate and burst | Attempts per time window, immediate burst | Quota / arrival rate |
| Concurrency | Simultaneously active attempts | Connections, memory, work in progress |
| Queue | Waiting items/bytes and maximum age | Memory and latency before execution |
| Retry budget | Extra attempts over original traffic or a fixed allowance | Failure amplification |

`golang.org/x/time/rate.Limiter` is a local token bucket. Use `Allow` for an
immediate admission decision or context-aware `Wait`/`WaitN` when waiting is
part of the contract. A failed/canceled wait does not authorize a dependency
call. A rate token is consumed admission, not a semaphore permit to return after
every completed request; canceled future reservations have their own API.

Ten replicas with 500/s local limiters can permit 5000/s, plus their configured
bursts. For a global 500/s quota, allocate shares with a bound on replica count
and burst, or use coordinated quota admission. Define behavior during scaling
and coordinator failure; do not let a fallback silently grant unlimited traffic.
A fixed partition is simple but may leave capacity unused. Choose based on the
actual quota and availability contract, not an assumed need for Redis.

Isolate tenants, dependency endpoints, or traffic classes at their failure
boundary. Choose a useful granularity without an unbounded map of per-request
limiters/breakers. Noisy tenants should consume their allocated work budget, not
all shared slots. Account for replica multiplication in half-open probes too.

## Breaker State and Composition

A breaker is justified when temporarily rejecting likely-failing calls helps
protect a dependency or caller. Existing admission limits and a short deadline
may already be sufficient; it is not obligatory on every remote call.

For an implementation in scope, define:

- **Closed:** the observation window, minimum sample size, failure/slow-call
  classification, and threshold appropriate to traffic.
- **Open:** fast rejection with no remote call; a bounded recovery delay and
  clear caller-visible error. Do not hide another retry loop behind rejection.
- **Half-open:** a small admitted probe set with defined success/failure rules.
  Do not release all queued callers when the open interval ends.
- **Generation:** in-flight results belong to the state generation that admitted
  them. A late success/failure from an old generation must not reset newer state.

Local queue rejection, caller cancellation, or a validation error does not by
itself show dependency ill health. A per-attempt timeout with an otherwise-live
caller budget may be a health signal. Separate local exhaustion from dependency
latency; classify based on the configured operation and observation point.
A tenant quota 429 need not open a breaker for unrelated tenants.

There is no universal decorator order across libraries. State whether the breaker
observes wire attempts or logical outcomes and how permits are acquired/released.
Each real attempt must pass applicable rate/concurrency/breaker admission; a
retry cannot bypass open state. Avoid consuming all limited half-open probes
while queued behind another limiter. Release/cancel unused probe reservations
on paths that do not dispatch. Keep health statistics distinct from local
rejections and avoid double-counting one outcome.

Use an existing suitable library and its tested concurrency semantics where
possible. If implementing a small breaker is explicitly required, test state
transitions, stale completions, concurrent probe caps, and recovery under load
with a controlled clock. A mutex alone does not validate the state machine.

## Fallback Preserves Meaning

Choose fallback only when its semantics are allowed. Record:

- Which data/operation can degrade and which must fail explicitly.
- Identity scope, authorization, maximum staleness, and consistency requirements.
- Whether partial/stale/default results are visible to the caller as such.
- The fallback's own time/capacity budget within the remaining caller budget.

A cached authorization allow from another tenant or past its permitted lifetime
is not resilience. An empty result can falsely mean absence; a swallowed failed
write can falsely mean success. Neither is a neutral fallback. Route security
policy to [go-security](../../go-security/SKILL.md), and retain the user's
explicit contract when deciding which optional work can be omitted.

Fallback can overload a second backend or stampede a cache refresh. Bound that
work and share applicable caller/attempt limits. Hedged parallel requests are
extra attempts with replay/capacity requirements; do not add them as a routine
latency fix or assume canceling the loser undoes its effects.

## Observe and Exercise the Policy

Separate logical success/failure and latency from attempt count, retry delay,
retry suppression, admission rejection/wait, breaker transitions/probes, and
fallback/degraded results. Use bounded dimensions such as dependency/operation
and outcome class; do not put arbitrary tenant IDs, request IDs, or keys into
metric labels. Follow [go-logging](../../go-logging/SKILL.md) for correlated logs.

Test controlled slow calls, permanent errors, overload, cancellation and recovery.
Assert bounds and semantic results, not only eventual success. Include noisy-
tenant isolation, all-replica rate/burst accounting, stale breaker completions,
and the crash/delivery cases relevant to this task. Use an isolated environment;
production fault injection requires that operational scope explicitly.
