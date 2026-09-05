---
name: go-resilience
description: Use when designing, implementing, or reviewing Go behavior under dependency failures or overload — retries, backoff and jitter, retry budgets, idempotency or duplicate delivery, circuit breakers, bulkheads, rate limiting, backpressure, or graceful degradation. Also use when retries amplify an outage or timed-out operations risk duplicate side effects. Does not cover an unknown failure's root cause (see go-troubleshooting), basic context propagation (see go-context), or server shutdown alone (see go-http).
---

# Go Resilience

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). The policies are
> version-independent; `testing/synctest` requires Go 1.25+.

Preserve operation semantics while bounding time, work, and failure impact.
A timeout leaves the remote outcome uncertain; another attempt consumes real
capacity and may repeat a side effect.

## Resource Routing

- `references/RETRIES-AND-IDEMPOTENCY.md` - Read for HTTP/RPC retries, Retry-After, request replay, ambiguous outcomes, duplicate events, or durable deduplication.
- `references/LOAD-AND-DEGRADATION.md` - Read for concurrency/rate admission, queue bounds, circuit breakers, recovery, fallback, or multi-tenant/fleet limits.

## Start With the Contract

Identify the operation's side effects, replay guarantee, caller deadline,
dependency capacity/quota, and required behavior when it cannot complete.
Inspect existing SDK, transport, proxy, and queue policies before adding a
wrapper. Defaults in different layers can multiply attempts.

Implement or review the protection the task needs; a routine client does not
need every pattern below. Preserve explicit delivery, consistency, freshness,
and security requirements. If a missing contract changes whether replay or
fallback is safe, ask for that fact while continuing independently valid work.
A review stays read-only; a local implementation request does not authorize
production fault injection, quota changes, or deployment. Missing sibling skills
or a host-specific `Skill` tool do not block use of the available guidance.

## Choose the Protection

| Problem | First mechanism | Boundary |
|---|---|---|
| A dependency call runs too long | End-to-end budget with bounded attempts | Cancellation does not undo remote effects |
| A transient failure on a replay-safe operation | Bounded retry with backoff/jitter | Permanent errors and ambiguous unsafe writes stop |
| Extra attempts sustain an outage | Retry budget and one coordinated owner | Per-call attempt count alone cannot cap fleet load |
| Too much simultaneous work | Bulkhead / bounded concurrency | Bound waiting work too; a semaphore is not a rate limit |
| Requests exceed a quota | Rate admission and bounded backpressure | State whether quota is per process, tenant, or fleet |
| A failing dependency needs recovery time | Circuit breaker if rejection adds value | It neither limits all concurrency nor makes replay safe |
| An optional dependency is unavailable | Contract-approved degraded result | No invented success, weaker authorization, or unmarked stale data |

Prefer existing clients/policies and a small local implementation where it is
sufficient. Follow [go-packages](../go-packages/SKILL.md) before adding a library.
`golang.org/x/time/rate` is an external Go module for token buckets; the standard
library has no general circuit breaker. Use a maintained implementation for a
needed state machine rather than casually building a reusable resilience stack.
Honor an explicitly selected library and verify its actual composition semantics.

## Retry Invariants

- Retry only when both the failure classification and operation contract permit
  replay. A 5xx or network timeout does not prove nothing happened; documented
  429 handling is not forbidden merely because it is a 4xx.
- Count **total attempts including the first**. Bound attempts, elapsed time,
  and aggregate extra traffic; inventory unavoidable lower-layer attempts.
- Queueing, rate admission, backoff, attempts, and response processing share the
  caller's total budget. Each attempt derives from it and can have a shorter
  timeout. Do not reset the total budget or detach retries after cancellation.
- Use capped exponential backoff with jitter and a cancellation-aware wait.
  A valid server retry delay is a lower bound under this policy: do not shrink
  it to a local cap. If it cannot fit, stop or use an already-authorized durable
  rescheduling path; parsing overflow is not permission to retry earlier.
- Use a fresh replayable request body for each attempt. Finish/close a discarded
  response and release attempt resources before waiting. Preserve the attempt
  context until the accepted response body is consumed or explicitly handed
  off with cleanup ownership; returning from `Do` does not finish the body.
- Scope retry ownership where operation semantics are known. Application, SDK,
  mesh, and broker retries must compose to a bounded policy rather than retry
  independently. Every real attempt passes dependency admission and breaker gates.

## Idempotency and Delivery

A key is meaningful only when the receiver enforces it. Keep one key per logical
operation, stable payload, and documented account/tenant scope and retention.
A new attempt is not a new operation. For an already-sent request, recover the
key actually sent; a newly chosen stable key cannot deduplicate an earlier one. An in-memory map is insufficient for
cross-replica or restart-safe deduplication.

Use durable atomic claims and explicit in-progress/completed/unknown states for
duplicate delivery. A lease expiring does not prove the previous worker stopped
or the remote operation failed. SQL can atomically bind local effects to a
receipt; a local transaction/outbox does not atomically commit an unrelated
remote effect. If its outcome cannot be reconciled and replay is unsafe, expose
that limit instead of promising exactly-once completion.

## Overload and Recovery

Bound admission **before** spawning unbounded goroutines. Choose reject, wait,
or durable backpressure from the caller's contract; do not silently drop work
that must be retained. Both admission and waiting need cancellation/lifetime
bounds. Isolate expensive tenants or dependencies without unbounded per-key state.

Release a dependency's per-attempt concurrency permit during backoff; retain a
separate bound on pending logical work. Rate tokens govern attempts over time,
not in-flight calls. Multiple local limiters do not enforce one global quota.

A breaker samples relevant dependency outcomes, limits half-open probes, and
ignores stale completions from earlier state generations. Classify caller
cancellation, local rejection, validation, and dependency timeouts deliberately.
Coordinate recovery admission so retries cannot bypass an open breaker or flood
the recovering dependency. Choose windows/thresholds from traffic and failure
behavior, not universal magic numbers.

Fallback is an alternate result with a contract: identity scope, freshness,
consistency, and caller-visible degradation. Preserve security decisions and
required side effects. Returning an empty list or cached allow on failure is
not a neutral substitute. Route authorization rules to
[go-security](../go-security/SKILL.md).

## Verify Failure Behavior

Test the selected policy, not just eventual success. Use controlled outcomes,
recorded attempt times/counts, injected randomness/clock or `testing/synctest`
(Go 1.25+) where suitable; real network I/O still needs appropriate test seams.

- Cancellation before admission, during backoff, and during the active call;
  exhausted deadlines start no further intended attempts.
- Nonretryable errors stop; safe transient errors recover within budget;
  server delay is honored; request bodies replay and response cleanup occurs.
- Concurrent duplicates, payload mismatch, and crash windows do not produce
  unsupported guarantees; test durable constraints with the actual database.
- Saturation bounds in-flight and queued work, preserves required delivery, and
  contains a noisy tenant. Recovery probes remain bounded under concurrency.
- Fallback preserves identity/freshness/error semantics; old breaker outcomes
  cannot change a newer state. Observe logical requests separately from attempts,
  retries exhausted/suppressed, queue wait/rejection, and degraded responses.

Use [go-testing](../go-testing/SKILL.md) for test mechanics and
[go-linting](../go-linting/SKILL.md) for the actual repository gate. Race tests
can expose in-process synchronization bugs; they do not establish distributed
idempotency. Report executed results, unrun checks, and remaining contract
limits honestly; follow [go-style-core](../go-style-core/SKILL.md#how-much-to-say)
for report length. A proposed policy is not evidence of outage recovery.

## Related Skills

- **HTTP mechanics**: [go-http](../go-http/SKILL.md) owns clients, transport configuration, response handling, and server shutdown.
- **Context and concurrency mechanics**: [go-context](../go-context/SKILL.md) and [go-concurrency](../go-concurrency/SKILL.md) own cancellation propagation, goroutine lifetime, and synchronization.
- **Durable local effects**: [go-database](../go-database/SKILL.md) owns transactions, constraints, and pool/query mechanics.
- **Failure classification and reporting**: [go-error-handling](../go-error-handling/SKILL.md) and [go-logging](../go-logging/SKILL.md) own error matching and structured logs; avoid high-cardinality or secret metric labels.
- **Unknown cause**: [go-troubleshooting](../go-troubleshooting/SKILL.md) owns incident evidence and diagnosis before speculative retry or breaker changes.
