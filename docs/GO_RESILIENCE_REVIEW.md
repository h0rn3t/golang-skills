# Go Resilience Review and Validation

Added on 2026-09-05 against baseline `558f8f9`.

## Scope and Ownership

[go-resilience](../skills/go-resilience/SKILL.md) owns policy for replay safety,
retry budgets, durable idempotency, overload admission, circuit recovery, and
fallback. Two references cover [retries/idempotency](../skills/go-resilience/references/RETRIES-AND-IDEMPOTENCY.md)
and [load/degradation](../skills/go-resilience/references/LOAD-AND-DEGRADATION.md).
It does not require a new framework, tracker integration, or a Claude-only host.

HTTP transport/server mechanics stay in `go-http`, context propagation in
`go-context`, synchronization/lifetimes in `go-concurrency`, and SQL mechanics
in `go-database`. The router and those skills now point to the policy owner;
`go-troubleshooting` routes known failure containment there after diagnosis.

Removed the blanket "Never retry a 4xx" instruction from `go-http`. Retry
classification must consider documented 429 behavior alongside replay safety,
server delay, and remaining budget. Also clarified complete SQL transaction
replay and its separation from remote side effects.

The pack now has **27 skills, 62 references, 90 trigger evals, and 39 quality
evals**. The new entrypoint has 146 lines. Added quality cases 34–39 and six
trigger cases; descriptions, manifest counts, ownership and bilingual README
are synchronized. Existing version metadata is retained.

## Source Checks

- [RFC 9110](https://www.rfc-editor.org/rfc/rfc9110.html#section-9.2.2) and its [Retry-After section](https://www.rfc-editor.org/rfc/rfc9110.html#section-10.2.3), plus [RFC 6585](https://www.rfc-editor.org/rfc/rfc6585.html#section-4), inform replay and delay classification.
- [Go Transport documentation](https://pkg.go.dev/net/http#Transport) and local `Request.isReplayable` / `persistConn.shouldRetryRequest` source confirm that idempotency headers can affect automatic network-error replay; `Request.Close` is not a retry-disable switch.
- [AWS idempotent APIs](https://aws.amazon.com/builders-library/making-retries-safe-with-idempotent-APIs/) informs operation identity and atomicity boundaries; the actual receiver contract remains decisive.
- [Go rate limiter](https://pkg.go.dev/golang.org/x/time/rate), [Google SRE overload](https://sre.google/sre-book/handling-overload/), and [Azure circuit breaker guidance](https://learn.microsoft.com/en-us/azure/architecture/patterns/circuit-breaker) inform admission and recovery decisions. Queue/delivery/fallback safeguards are repository policy, not universal provider guarantees.
- [PostgreSQL serialization failure handling](https://www.postgresql.org/docs/current/mvcc-serialization-failure-handling.html) confirms that a retry includes the complete transaction and the logic deciding its queries.

## Behavioral Evidence

[Recorded prompts, complete answers, tool calls, hashes and assessments](evidence/2026-09-05-go-resilience-probes.json)
retain baseline, updated, and final snapshots. GPT-6 used independent subagent
contexts per round. Actual Claude Code **2.1.261** ran **`claude-opus-5`** at
`medium` effort, confirmed by CLI initialization records, in fresh directories
per case. Application probes explicitly read the skill and routed references;
assertions and prior answers were withheld. Only read tools were enabled,
hooks were disabled, MCP was empty/strict, and persistence was disabled.

Baseline GPT-6 already made most of the correct decisions using existing
skills. Baseline Opus recognized the 429 conflict but added contradictory delay
clamping advice and unsupported claims around duplicate delivery. A baseline
instruction gap is not automatically evidence of a model failure or of a
subsequent quality increase.

The first updated answers exposed two areas needing clarification:

- A persisted logical operation ID does not prove it was the idempotency key
  on the first wire attempt. Both models initially assumed a persisted key;
  Opus also recommended unsafe transport replay controls. Clarified preservation
  of the actual sent key, unresolved earlier different-key outcomes, and Go's
  automatic `GetBody` / idempotency-header behavior.
- Opus calculated 24 attempts but gave inconsistent amplification ratios and
  assumed a short deadline precluded exhausting them. Clarified metric
  denominators, fast failures, and conditional SDK ownership.

Strengthened assertions for 35 and 37 without changing their prompts, then
reran those cases in fresh contexts. These are development iterations rather
than a held-out benchmark.

| Case | Observed core decisions |
|---|---|
| 34: 429 and Retry-After | Both allow the documented safe retry when time permits, reject a 60-second wait under a 250 ms budget, and stop unchanged 401 replay. |
| 35: timed-out POST and key changes | Final GPT distinguishes the actual first key from a logical ID and preserves unknown outcomes. Final Opus recognizes the historical-key problem but retains conflicting advice; see below. |
| 36: duplicate delivery across replicas | Both reject local SQL/mutex/outbox as an exactly-once remote-effect guarantee and use durable claims, payload binding, and unknown states. |
| 37: layered retries | Final answers give 24 total/23 extra attempts and coordinate retry ownership, caller deadline, and aggregate admission. |
| 38: overload and fleet quota | Both bound active/waiting work, preserve required delivery, separate rate/concurrency, and account for replica/tenant scope. |
| 39: breaker and authorization fallback | Both bound probes, reject stale-generation state updates, and reject expired cross-tenant allow decisions. |

Final source refinements affected cases 35 and 37; other observations remain
from the updated snapshot. Hashes identify the exact files each round used.

Six separate native Claude `Skill` probes tested automatic selection without
forced reads. Four implementation/design requests selected
`golang-skills:go-resilience`; the translation and grammar controls selected no
skill: **6/6 selection results**. GPT-6 automatic discovery and the whole pack's
model eval suite were not run.

## Limits

These were simulated design/review tasks, not executed Go implementations or
live outage tests. One sample per case/model/round cannot establish reliability.
Process success and correct core decisions do not imply flawless answers.

Opus still overstates some policies: case 34 treats the full attempt-timeout
ceiling as a minimum; case 36 assumes a 202 response causes redelivery; case 39
assumes queued callers have expired and prescribes universal breaker thresholds.
In final case 35 it flags historical keys as unresolved but first proposes a new
stable key, calls 429/503 rejections safe without establishing absence of side
effects, and includes invalid illustrative Go pseudocode. Its large-payload and
retention claims also need qualification. Those responses are not ready-made
implementation guidance; the skill's contract/replay constraints still apply.
There is no blanket quality pass rate. Full responses preserve these limits.

## Repository Verification

- `go test -count=1 ./...` from `evals/`: passed, including structure, description
  fixtures, schema, manifest counts, script behavior and lint configuration.
- `skill-creator/scripts/quick_validate.py`: new skill passed.
- `agentskills-validate@1.0.1`: new skill and all six edited sibling entrypoints passed.
- CI relative Markdown link check and `git diff --check`: passed.

These checks validate this repository's instructions and fixtures. They do not
execute the proposed implementations in the model answers.
