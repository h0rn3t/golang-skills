# Independent current-agent skill application probes

2026-09-09. Six identical task scenarios were evaluated in independent agent
contexts before and after instruction changes. These are qualitative application
probes using the current session model, not Opus 5 / GPT-5.6 runs or a controlled
model benchmark. The main agent reviewed the decisions; no automatic judge or
runtime execution of these hypothetical snippets was used. Both passes selected
contract-preserving approaches, so this establishes no measured quality gain.
The candidate snapshot precedes final summary-row and BEHAVIOR-TRAPS corrections.

Baseline snapshot SHA-256: `dc68d7a600e943e0edcac93e37a8561bf658c3fd9d77b10a6d12816fb7543d2b`.
Candidate snapshot SHA-256: `5b93552292d32ef03d8aa845a86023381b7bc2250be451b055afd6b43859e775`.

## Baseline response

# Current-agent baseline application probes

Snapshot: `/tmp/golang-skill-probes.rTJ5x6/baseline`. All references below are relative to that immutable snapshot. This is a read-only application exercise, not a runtime test or model comparison. No repository edits, audit report reads, shell model invocations, or paid model tests occurred. `graft map` / `graft ask` supplied orientation; Markdown skills are not indexed. Graft search surfaced an audit filename, but no audit report was opened.

## A — Review one changed handler

Decision: review only the changed function, reading its dependencies/neighbors where necessary to establish its behavior. Keep the workspace read-only. Run the prescribed `go test ./internal/api`; record its actual output and exit code. Do not automatically run the module-wide pre-review script or HTTP checklist commands, and do not repair unrelated docs/style. An identified concern can justify a focused additional check; report why and its exact scope. No pass is claimed here because the scenario supplied no target handler/module.

Report shape: `Reviewed the changed handler. <Concrete attributable finding with file:line, severity, verified/plausible evidence, if any>. Verification: go test ./internal/api — <observed result>. Unrelated documentation/style issues were outside this review.` If no finding is established, state that limited scope rather than claim universal correctness. Do not invent findings from the scenario alone.

Instructions: `go-code-review/SKILL.md:20-27` diff scope and repository precedence; `go-style-core/SKILL.md:34-45` repository/user precedence and review remains read-only; `go-linting/SKILL.md:20-22,36-48,54-58` repository gate replaces defaults, scope and observed results; `go-code-review/SKILL.md:39-42` provenance for findings/checks.

Ambiguity encountered: `go-code-review/SKILL.md:28-30` prescribes module-wide pre-review plus go fix preview; `go-http/SKILL.md:167-171` prescribes broader checks; `go-code-review/SKILL.md:34-35` says report every finding. I resolve these through explicit repository precedence and the requested single-function review, rather than expanding into unrelated style findings.

## B — Snapshot of map[string][]int

Decision: preserve independent mutable inner slices. Never replace a deep copy with `return maps.Clone(state)`. For an existing snapshot that already preserves nil maps and nil values, the local refactor is:

```go
s.mu.RLock()
defer s.mu.RUnlock()
out := maps.Clone(s.values)
for key, values := range out {
    out[key] = slices.Clone(values)
}
return out
```

Keep the existing locking primitive/lifetime; `RLock` here merely illustrates an existing RWMutex. If the existing implementation uses `make(map[string][]int, len(s.values))` and makes every inner slice, retain those allocations when they deliberately return non-nil empties for nil inputs. I would inspect that exact existing loop before choosing the final snippet; a nil-normalizing implementation cannot be replaced by the above unchanged. Retaining the original inner `make`+`copy` is valid when needed for equivalence. Add no exported Clone API for a single snapshot operation.

Verification targets: mutate an existing returned element and append to a returned slice, confirm internal state is unchanged; replace/delete a returned map value, confirm original map unchanged; compare nil/empty map and slice cases against existing behavior. Existing adequate tests can be reused.

Instructions: `go-defensive/SKILL.md:77-91` copy boundary and shallow Clone caveat; `go-defensive/references/BOUNDARY-COPYING.md:46-55,59-71` snapshot under lock, nil behavior, clone each map slice value; `go-code-refactor/references/BEHAVIOR-TRAPS.md:30-46` preserve nil distinctions; `go-code/SKILL.md:44-51,65-70,72-79` correctness/ownership over brevity and no unnecessary helper.

Ambiguity encountered: `go-defensive/SKILL.md:80-81` says never hand-written make+copy; `BOUNDARY-COPYING.md:13-15` says clones replace every such pair, despite the shallow caveat. `BEHAVIOR-TRAPS.md:48-49` claims make+copy to Clone preserves nilness; that inference does not hold when the original allocation returns non-nil empty output. The explicit behavior preservation requirement decides this case.

## C — Endless queue, at most 8 calls, cancellation

Decision: exactly eight long-lived workers, each receives with cancellation and invokes the synchronous operation with the same cancellation context. Function waits for every worker before returning. Do not spawn one goroutine for every received queue item. Do not close the caller-owned input channel. With a processing operation that has no error result, stdlib suffices:

```go
func consume(ctx context.Context, queue <-chan Item) error {
    var wg sync.WaitGroup
    for range 8 {
        wg.Go(func() {
            for {
                select {
                case <-ctx.Done():
                    return
                case item, ok := <-queue:
                    if !ok || ctx.Err() != nil {
                        return
                    }
                    process(ctx, item)
                }
            }
        })
    }
    wg.Wait()
    return ctx.Err()
}
```

Assumptions: `process` cooperates with cancellation and does not panic, as required for the shown lifecycle. If it returns an error and failure should stop consumption, use eight identical worker loops in `errgroup.WithContext`, returning the processing error and `g.Wait()`; fixed worker count itself bounds concurrency, so SetLimit is unnecessary. I would preserve an existing continue-on-error policy if present. I would not promise forcibly terminating an operation that ignores its ctx.

Verification targets: queue never closes yet cancellation returns; no work queued and cancellation returns; 9 queued items with blocked processors yield max active 8; cancellation reaches in-flight calls and consume does not return before they exit. Run the applicable scope with `-race`.

Instructions: `go-concurrency/SKILL.md:21-39` lifetime, stop, wait, bounded unbounded input; `go-context/SKILL.md:106-115` cancellation checks; `go-context/references/PATTERNS.md:129-161` loop and pre-operation checks; `go-concurrency/references/ADVANCED-PATTERNS.md:111-120` WaitGroup.Go and bounded work; `go-code/SKILL.md:44-51` stdlib sufficient before dependencies; `go-linting/SKILL.md:49-51` race check.

Ambiguity encountered: the “Good” loop at `go-concurrency/SKILL.md:41-48` ranges queue and launches unbounded goroutines; I do not apply it to this explicit endless-input limit. The selector at lines 74-83 points bounded fan-out toward SetLimit, but a fixed worker count is the simpler stdlib fit here. Error policy and cooperative cancellation depend on the existing process operation, absent from the scenario.

## D — Exactly one JSON request object, cap 1 KiB before Create

Decision: HTTP owner controls body mechanics. Decode into a pointer to the existing request struct so null is distinguishable from an object; enforce EOF before Create. Keep validation/domain conversion local unless existing shared helpers are useful.

```go
r.Body = http.MaxBytesReader(w, r.Body, 1<<10)
dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields()
var req *createRequest
if err := dec.Decode(&req); err != nil || req == nil {
    http.Error(w, "invalid JSON object", http.StatusBadRequest)
    return
}
if err := dec.Decode(new(any)); err != io.EOF {
    http.Error(w, "body must contain one JSON object", http.StatusBadRequest)
    return
}
// Validate required fields here, before the existing Create call.
created, err := s.store.Create(r.Context(), *req)
// Apply the existing error/status mapping and write the response.
```

Assume ordinary encoding/json decoding into the existing struct (no custom UnmarshalJSON accepting other JSON types). Scalars and arrays cannot decode into that struct; null is rejected by pointer check. Both decodes occur under MaxBytesReader, so whitespace past the byte cap also prevents Create. Status 400 is chosen from supplied HTTP example; if the API contract requires 413 for size failures, inspect errors from both decodes for MaxBytesError and map consistently. No error internals or raw request values are echoed. Duplicate keys are not specified in this request; if the contract requires rejecting them, an additional decoder policy is needed, not a claim that DisallowUnknownFields does it.

Verification targets: valid object at 1024 bytes accepted; 1025 bytes rejected; object plus whitespace exceeding cap rejected; two objects, trailing malformed data, empty body, null, array/scalar, unknown field rejected; all rejected cases prove Create was not called.

Instructions: `go-http/SKILL.md:47-80` cap/first decode/EOF/validate/domain ordering; `go-security/SKILL.md:113-115` delegates these mechanics to HTTP owner; `go-security/SKILL.md:125-126` generic errors; `go-code/SKILL.md:86-92` contract-derived checks and overlap cases.

Ambiguity encountered: HTTP example says one JSON value and uses non-pointer struct (`go-http/SKILL.md:55-62`), so direct copy alone does not establish the explicit object-only contract. No size-specific status is specified by this scenario.

## E — Existing enum zero is stdout, add third option

Decision: append the third member and add its branch in the existing implementation, preserving numeric values and zero default. No unknown sentinel, shifted iota, pointer config, or With-style API.

```go
const (
    LogToStdout LogOutput = iota
    LogToFile
    LogToRemote
)
```

Names mirror the skill example; retain real repository names. Verify omitted configuration still selects stdout and prior numeric values remain unchanged; exercise the new branch through the existing API.

Instructions: `go-style-core/SKILL.md:84-85` deliberate useful zero and preserved zero distinctions; `go-style-core/references/IOTA.md:25-40` precisely this stdout-zero exception; `go-functions/SKILL.md:103-109` defaults/config contract and no gratuitous With helpers; `go-code/SKILL.md:59-63,86-92` existing API and defaults.

Ambiguity encountered: `go-style-core/references/IOTA.md:9-12` starts with default start-at-one guidance; the explicit useful-zero exception resolves this scenario. `go-defensive/SKILL.md:130-132` likewise leads with invalid zero but delegates the exception.

## F — Pure local string formatting helper

Decision: no context parameter. It neither consumes cancellation/deadlines/request values nor passes them downstream. Keep a simple synchronous signature:

```go
func formatLabel(name string, count int) string {
    return fmt.Sprintf("%s (%d)", name, count)
}
```

Keep the actual existing format unchanged. No ctx.TODO/Background inside the helper. Preserve existing required callback/interface signature if one exists; none is given here.

Instructions: `go-context/SKILL.md:16-25` first-parameter rule is conditional on functions using a context; `go-code/SKILL.md:44-51,72-79` minimum required operation, no speculative plumbing; `go-functions/SKILL.md:14-16` preserve signatures; `go-style-core/SKILL.md:28-30` least mechanism.

Ambiguity encountered: `go-context/references/PATTERNS.md:68-76,247` defaults to context even unused, with future-proofing rationale, and `go-code-review/SKILL.md:97` repeats it. For this explicitly pure local helper, speculative API growth has no current purpose, so I follow the narrower context-use rule and restraint guidance.

## Evaluation boundary

These are proposed decisions, not compiled implementations. Six scenarios yielded six contract-compatible decisions through the current instruction set, but several required resolving contradictory or over-broad examples. This does not establish that another agent, a different prompt, or repeated runs will make the same decisions. No baseline-vs-candidate or causal improvement claim is supported yet.


## Candidate response

# Current-snapshot behavioral probes

Snapshot: `/tmp/go-skill-probes.TAQsV7/skills` (copied from repository `skills/` before skill evaluation).
All instruction citations below are relative to that snapshot. This is a current-agent, read-only reasoning probe: no repository handler fixtures were supplied, no target test commands were executed, no shell model calls or paid A/B runs were used, and no memory/history/audit report was opened. Required `graft map` reported an automatic one-file graph refresh; no explicit repository edit was made. `graft ask` did not locate the Markdown skills, so skill Markdown was read in the snapshot. One audit filename appeared in ranked graph output; its contents were not read.

## A. Review one changed handler

**Decision and actions.** Keep the review read-only and scoped to the changed handler plus the callers/types needed to establish its contracts. Read repository conventions, then trace request inputs, validation, side effects, errors, and response behavior. Run the specified repository gate, `go test ./internal/api`, from the target module, or reuse an observed passing result for the identical relevant state. Do not automatically run full-repository formatting, documentation checks, modernization, lint, or a default full Go gate. Do not fix even attributable bugs without fix authorization. Unrelated documentation/style issues are outside this review; if they appear in diagnostics, identify them as pre-existing rather than turning them into findings against this handler.

**Report I would return.** Findings grouped by severity, each containing the changed `file:line`, concrete failing input/path, impact, and executed or static evidence. Material hypotheses would explicitly name missing evidence. Verification would say `go test ./internal/api — pass/fail/unavailable (actual result)`; in this probe it is **not run**, because no target handler/module was supplied. A clean review would say “No supported findings in the changed handler,” not assert whole-repository cleanliness. No fabricated finding or passing result.

**Instructions used.** `go-code-review/SKILL.md:20-46` (scope, local conventions, evidence, report); `:51-53` (review remains read-only); `:217-221` (selected checks, authorized fix mode); `go-linting/SKILL.md:20-22,36-42,54-62` (repository gate precedence, reuse, accurate outcomes); `go-style-core/SKILL.md:34-45` (house style and scope); `go-http/SKILL.md:41-81` (handler ordering and contracts); `go-code-review/assets/review-template.md:11-39` (evidence/report fields).

**Ambiguities.** The specific handler, changed lines, and existing test results are absent. The short review cannot honestly contain actual findings or a test status. The broad review checklist has documentation and style rows, but the explicit procedure limits their application to this diff.

## B. Snapshot map[string][]int with independently mutable returned slices

**Decision.** Retain the outer loop that builds an independent map. Copy every inner slice. A bare `maps.Clone(state)` is incorrect because `snapshot["x"][0] = 99` would mutate the source's inner array. If the existing implementation is nested element-copy loops with always-allocated inner slices, simplify only the element loop to built-in `copy`:

```go
out := make(map[string][]int, len(state))
for key, values := range state {
    out[key] = make([]int, len(values))
    copy(out[key], values)
}
return out
```

Keep the existing lock and nil-map branch, if any, exactly where they already are. This snippet assumes the current outer result and inner results are always allocated. If the existing contract instead preserves nil slices, the inner operation can be `out[key] = slices.Clone(values)`. I would not introduce nil-preserving cloning over a `make`-based implementation without checking that representation. If the existing code already uses `make` + `copy`, it is already a small, correct implementation: keeping it is an acceptable simplification outcome, rather than adding a generic deep-clone helper.

**Checks I would use.** Existing snapshot tests, plus targeted ownership cases if missing: mutate an existing returned inner element; replace/delete a returned map entry; change original data after snapshot; nil and empty inner slices remain as before. Keep any required synchronization throughout the full deep copy.

**Instructions used.** `go-data-structures/SKILL.md:41-44,191` (matching semantics before stdlib replacement; clone is shallow); `go-defensive/SKILL.md:79-91` (independent mutation and nested references); `go-defensive/references/BOUNDARY-COPYING.md:14-16,47-58,63-70` (depth, lock, nil/capacity); `go-code-refactor/SKILL.md:100-102` (observer-visible behavior unchanged); `go-code/SKILL.md:44-51` (minimum correct operation); `go-data-structures/references/SLICES.md:80-91` (built-in copy).

**Ambiguity found during application.** `go-code-refactor/references/BEHAVIOR-TRAPS.md:48-49` says replacing `make`+`copy` with Clone preserves nil-ness. That broad statement conflicts with the concrete warning in `BOUNDARY-COPYING.md:55-58`: `make([]int, 0)` is non-nil while `slices.Clone(nil)` is nil. I follow the actual existing representation and retain `make` + `copy` when needed. Source code was not supplied to settle which existing representation applies.

## C. Endless queue, at most eight calls, cancellation

**Decision.** Use eight fixed workers and join them. The producer owns and closes the queue; workers receive only. The run operation is synchronous: it returns after all workers stop. Cancellation interrupts waiting for the next item and is passed into processing. For this minimal example, `process` is synchronous and handles its per-item outcome; it must observe cancellation and must not panic.

```go
func run(ctx context.Context, queue <-chan Item) error {
    var wg sync.WaitGroup
    for range 8 {
        wg.Go(func() {
            for {
                if ctx.Err() != nil {
                    return
                }
                select {
                case <-ctx.Done():
                    return
                case item, ok := <-queue:
                    if !ok {
                        return
                    }
                    if ctx.Err() != nil {
                        return
                    }
                    process(ctx, item)
                }
            }
        })
    }
    wg.Wait()
    return ctx.Err()
}
```

Eight worker goroutines give at most eight simultaneous synchronous calls. Checks around receiving avoid processing already-observed cancellation when the queue is continuously ready; they do not promise an impossible atomic cutoff between checking cancellation and entering `process`. If processing ignores context, prompt shutdown cannot be guaranteed by this wrapper.

If `process` returns errors and the existing contract requires fail-fast sibling cancellation, use the same eight-worker loop inside `errgroup.WithContext`; return each worker's processing error and join with `g.Wait()`. Do not feed an endless queue directly into blocking `g.Go` admission with `SetLimit(8)`. Do not invent retries, acknowledgements, or a new buffering layer. Test peak concurrency, cancellation with idle and busy queues, and worker completion; run the affected package with `-race` in an implementation task.

**Instructions used.** `go-concurrency/SKILL.md:31-57,103-116,135-160` (fixed workers, wait, synchronous API, direction, buffer policy); `go-concurrency/references/ADVANCED-PATTERNS.md:11-30` (error contract; blocked Go admission; endless queues); `go-context/SKILL.md:18-26,108-117` (placement and cancellation); `go-context/references/PATTERNS.md:69-71` (context use); `go-error-handling/SKILL.md:70-82` (cancellation result and deliberate error handling); `go-linting/SKILL.md:49-51` (race evidence).

**Ambiguities.** The error/acknowledgement contract and producer ownership are not specified. I chose no newly invented failure policy. The queue may contain unbounded logical work elsewhere; eight active workers bounds execution, not all upstream storage.

## D. Exactly one JSON request object, 1 KiB, then Create

**Decision.** Limit the body to 1024 bytes before decoding; reject unknown fields for the strict request schema; decode into a pointer to the existing request struct so JSON `null` is rejected as a non-object; require EOF on a second decode before validation and Create. Keep safe generic parser/error responses. This uses the existing repository's JSON decoder and request type, illustrated with encoding/json v1:

```go
r.Body = http.MaxBytesReader(w, r.Body, 1<<10)
dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields()

var req *createRequest
if err := dec.Decode(&req); err != nil || req == nil {
    http.Error(w, "invalid JSON object", http.StatusBadRequest)
    return
}
if err := dec.Decode(new(any)); err != io.EOF {
    http.Error(w, "body must contain one JSON object", http.StatusBadRequest)
    return
}
if err := req.validate(); err != nil {
    http.Error(w, "invalid request", http.StatusBadRequest)
    return
}
result, err := store.Create(r.Context(), *req)
// Continue with the existing error mapping and success writer.
```

`createRequest` is an ordinary existing struct, not a custom permissive Unmarshaler. Arrays/scalars fail its decode; null is explicitly rejected. A second object, trailing junk, truncated JSON, and oversized trailing whitespace all fail before Create. Valid trailing whitespace within the limit is allowed. The cap includes whitespace. The example returns 400 for malformed and oversized input; if the endpoint contract mandates 413, classify `*http.MaxBytesError` in both decode paths. Do not rely only on Content-Length or `DisallowUnknownFields`, and do not expose raw validation errors unless confirmed safe public messages. Keep authorization and validation before the side effect; trace actual fields into sensitive sinks if applicable.

**Checks I would use.** Valid object at <=1024 bytes calls Create once; empty input, null, array, scalar, unknown key, second object, junk suffix, malformed JSON, and >1024-byte whitespace suffix call Create zero times. Include exactly 1024 and 1025 bytes. These are proposed cases, not claimed executed tests.

**Instructions used.** `go-http/SKILL.md:47-81` (limit, decode, EOF, validation, then Create with request context); `go-security/SKILL.md:13-17,24-39,125-126` (trust boundary, sink tracing, generic responses); `go-code/SKILL.md:76-79,86-92` (example helpers are not mandatory APIs; boundary and precedence checks); `go-error-handling/SKILL.md:77-85` (handle errors).

**Ambiguities.** Exact error statuses, field rules, existing JSON package, duplicate-member and case-sensitivity policy were not specified. The pointer is my adaptation of the HTTP example to the stricter word “object”; the snapshot's nonpointer example by itself does not exclude JSON null. v1 `DisallowUnknownFields` does not reject duplicate known fields. If “exactly one object” is also intended to require duplicate rejection, that is an additional contract to implement explicitly.

## E. Existing enum zero is stdout; add third option

**Decision.** Append the new member while preserving old numeric values and useful zero behavior. Assuming the two existing members are stdout=0 and file=1:

```go
const (
    LogToStdout LogOutput = iota
    LogToFile
    LogToRemote
)
```

Use the actual new destination name from the task. Extend existing selection logic for that member, preserve omitted configuration and explicit zero as stdout, and keep the existing config/constructor API. Do not insert an invalid member ahead of stdout, add pointers to detect unset, or introduce WithOutput/default-normalization scaffolding. Check existing values, zero-config behavior, and routing of the new member.

**Instructions used.** `go-style-core/SKILL.md:84-85`; `go-style-core/references/IOTA.md:25-40` (this exact useful-zero form); `go-defensive/SKILL.md:129-131` (preserve zero and existing numeric enum); `go-code/SKILL.md:59-63,94-98` (existing API and no unused surface); `go-functions/SKILL.md:96-110` (no configuration API change just to introduce options).

**Ambiguities.** Actual member names and whether values are persisted must be read in the target code. Both support appending with stable values; neither warrants inventing a new API.

## F. Pure local string formatter

**Decision.** No context parameter. Keep the ordinary local helper signature:

```go
func formatLabel(name string, count int) string {
    return fmt.Sprintf("%s (%d)", name, count)
}
```

There is no I/O, cancellation point, request metadata, or interface contract requiring ctx. Do not thread an unused context through callers or create a background context inside it. Preserve an existing required signature if this turns out to be an interface callback. No new context tests are needed.

**Instructions used.** `go-context/references/PATTERNS.md:69-71` (explicit pure-formatting exemption); `go-context/SKILL.md:18-26` (first parameter applies when the function uses context); `go-code-review/SKILL.md:105` (avoid unused context on pure helpers); `go-functions/SKILL.md:14-16` (existing signatures); `go-code/SKILL.md:44-51,72-79` (minimum useful mechanism).

**Ambiguities.** None under the stated local/pure assumptions.

## Verification boundary

All six choices are applications of the captured instructions to hypothetical requests, with concrete proposed code and unresolved contract details stated above. Snippets were not compiled against nonexistent target types and no target repository gate was run. No claim is made about Opus, GPT, matched baselines, or statistical improvement. Graft reported 91,464 + 27,475 = approximately 118,939 tokens saved during repository orientation; this is its estimate, not a benchmark measurement.
