# Retries and Idempotency

> Sources: [HTTP semantics](https://www.rfc-editor.org/rfc/rfc9110.html#section-9.2.2); [Retry-After](https://www.rfc-editor.org/rfc/rfc9110.html#section-10.2.3); [429](https://www.rfc-editor.org/rfc/rfc6585.html#section-4); [Go HTTP transport](https://pkg.go.dev/net/http#Transport); [Idempotent APIs](https://aws.amazon.com/builders-library/making-retries-safe-with-idempotent-APIs/)
> Authority: normative for protocol/Go API behavior; project policy for retry admission and delivery safeguards
> Last verified: 2026-09-05

## Classify Before Replaying

| Observation | Decision |
|---|---|
| Caller canceled or total budget expired | Stop this request's work; do not restart with `Background` |
| Attempt timed out while caller budget remains | Remote outcome may be unknown; retry only if replay-safe and useful |
| Validation/authentication error | Stop unchanged replay; a documented credential-refresh flow is a separate bounded recovery action |
| Documented transient 429/503 or RPC equivalent | Retry only with replay safety, delay, and remaining budgets |
| Other 5xx / transport failure | Inspect the dependency contract and error; neither category is uniformly transient or safe |
| Lost response after a write | Reconcile or use receiver-enforced deduplication; do not equate lack of response with rollback |

For HTTP, idempotence concerns intended effects, not identical response bytes.
A repeated DELETE may return a different status without repeating its intended
effect. A POST can be safely retried under an explicit idempotency contract;
a client-supplied header alone creates no such guarantee.

Go's `Transport` already retries some network failures on previously used
connections for requests it considers idempotent, when the body is absent or
replayable via `GetBody`. Its recognition of an `Idempotency-Key` header is a
transport heuristic, not verification that the server implements deduplication.
Do not mark an unsafe POST as idempotent for an unsupported receiver: adding
this header can enable automatic replay after an ambiguous network failure.
`Request.Close` is not a switch disabling retries. `NewRequestWithContext`
automatically supplies `GetBody` for some readers, including `bytes.Reader`;
inspect the constructed request and transport policy instead of assuming an
omitted assignment disables replay. Proven pre-write retries differ from
ambiguous post-write retries; classify the actual attempt, not just an error name.

Inspect SDK/proxy retry settings as well as explicit loops. Count application
attempts separately from observed wire attempts when lower-layer behavior matters.

## One Time and Work Budget

Decide a maximum total-attempt count, overall duration, attempt timeout, and
extra-traffic budget at the dependency's relevant scope. Three layers allowing
3, 4, and 2 total attempts can produce 24 downstream attempts for one logical
request. Report total attempts per original operation and extra attempts
(total minus one); name the denominator of any amplification ratio. A deadline
can reduce counts for slow failures, but fast failures can exhaust all attempts.
Designate a retry owner or account for unavoidable nested attempts. The owner
can be an operation-aware SDK; it need not always be application code.

A request's remaining time includes admission, sleep, connection setup, request
write, response headers, body processing, and cleanup. Derive each attempt's
context from the caller. An attempt timeout is a ceiling, not a requirement to
always spend that entire duration; start only when a useful attempt fits the
chosen policy. Recheck cancellation and budget after waits and before dispatch.

For locally computed delay, use a saturating/capped exponential calculation and
sample jitter from its bounded window (for example, full jitter from zero to
the cap for that attempt). Avoid shifts/multiplication that overflow durations.
Bound the retry count even when random delay happens to be zero. Production
clients should not share a fixed seed that synchronizes retries; tests can use
controlled randomness.

Per-request limits still permit a retry storm with many callers. An extra-attempt
token budget or equivalent admission policy should suppress retries when the
dependency is saturated. Do not spend unlimited retries merely because another
caller's total deadline is long.

## Retry-After Without Deadline Violations

Parse nonnegative decimal delay-seconds or an HTTP-date (`http.ParseTime`).
Calculate a date's remaining wait with the available clock; handle past dates,
clock skew, malformed values, and overflow explicitly. A date already in the
past adds no positive server delay; it does not remove local backoff/budgets.

For a valid server delay, use at least that delay and the selected local delay;
any added jitter must not move the attempt earlier. The local backoff cap caps
local policy, not the server's valid instruction. If the resulting delay leaves
no useful attempt before the caller deadline, return the current failure with
relevant retry metadata or use an already-authorized durable rescheduling path.
Do not sleep out the request just to return a timeout, or shorten a large valid
value to a supposedly safe maximum. A numeric overflow must not become a short
retry interval; decline the retry when its requested wait cannot be represented.

Example: a safe GET with 3 seconds left, a 500 ms attempt limit and
`Retry-After: 2` can retry after that wait if other budgets allow. With 250 ms
left and `Retry-After: 60`, it cannot; a 1-second local cap changes neither fact.

## HTTP Attempt Ownership

Build a new request/body per attempt from immutable replay data or `GetBody`;
a consumed stream is not reusable. Persist the operation key and payload when
retries must survive process restarts. Keep redirects and transport-level
replays within the operation's intended destinations and side-effect contract.

Inspect status/error before decoding. Close each discarded response inside its
attempt scope, not with a loop-wide defer that accumulates bodies. For connection
reuse, drain only within bounded size/time; do not consume an unbounded error
stream just to preserve a connection. Release the dependency permit before
backoff, and reacquire before a later attempt.

A successful `Client.Do` return still leaves a live response body. If a helper
returns `*http.Response`, a deferred attempt `cancel()` must not invalidate the
body before its caller reads it. Either consume/decode/close within the attempt,
or transfer both body and cancellation/permit cleanup ownership to its consumer.
Test cancellation while reading as well as before headers.

## Durable Duplicate Handling

Bind a stable logical operation identifier to the receiver's identity scope and
payload semantics. Reusing a key for a different intent is a conflict; distinct
intent may legitimately have an identical payload. When fixing an existing caller, recover the exact key used on the first wire
attempt. A persisted logical ID is not evidence that the request used it as its
key. If that key is lost, or previous attempts used different keys, treat those
outcomes as unresolved; introducing a stable key only protects future work under
that key. Retention starts according to the receiver's contract, not when the
current retry loop starts.

Retention must cover the supported replay/redelivery window; late arrivals after expiry need an explicit
contract, not an unconditional deduplication promise.

On the receiver, use an atomic durable claim/unique constraint keyed by the
required tenant/account and operation identity. A `SELECT` then side effect then
`INSERT`, even with a process mutex, races across replicas. Define completed
response replay, changed-payload rejection, and in-progress duplicate handling.
Use a consistent canonical payload representation or an explicitly byte-exact
contract; do not silently change the equality rule between requests.

For local database effects, commit the receipt and mutation atomically using
[go-database](../../go-database/SKILL.md). For remote effects, record durable
intent and state, but acknowledge the gap between remote completion and local
acknowledgement. An outbox makes local intent durable; its delivery can repeat.
Without receiver deduplication or reliable reconciliation, it cannot promise
both no duplicate effects and completion after an ambiguous outcome.

Crash recovery needs ownership/version checks, not just a lease timeout. The old
worker or remote call may still run after its lease expires; a new worker must
not blindly repeat an unsafe effect. Use receiver-enforced operation identity,
fencing where supported, or preserve an unknown state for reconciliation and a
business decision about at-most-once versus at-least-once behavior.

Work that must survive the original request belongs to an explicitly durable
handoff and a worker-owned bounded lifetime. Detaching a goroutine or using
`WithoutCancel` by itself does not make delivery durable. Keep retryable SQL
transaction work separate from nontransactional remote side effects.
