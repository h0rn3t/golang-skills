# Go Troubleshooting Review

Reviewed on 2026-09-05 against baseline `e94ee92`. Scope: improve the existing
skill for ticket investigation and correct misleading runtime interpretations,
using the shared authoring rules for GPT-6 and Claude. No separate debugging
skill or tracker integration is required.

## Changes

| Finding | Correction |
|---|---|
| Discovery and the method centered on runtime crashes/profiles, with little guidance for wrong business results. | Expand triggers to tickets, missing/wrong results, tenant-specific failures, regressions, and environment differences. |
| A local reproduction could be treated as evidence about deployed code/config. | Match request to deployment identity; compare effective settings and relevant data; preserve uncertainty when matching source is unavailable. |
| The proposed cause in a ticket could anchor the investigation. | Separate contract, observed behavior, and hypothesis; compare a working case and choose a discriminating check. |
| The main method went straight from diagnosis to editing and a full gate. | Separate investigation-only from authorized fixing; report confirmed/probable/unresolved with evidence and remaining checks. |
| Every correction was required at a shared function. | Fix the responsible contract boundary; preserve other callers and explicit input values outside the defect. |
| Runtime guidance equated blocked stacks with leaks or lock-order inversion and confused panic sites with causes. | Require lifetime/ownership/wait-cycle evidence, correct panic/channel interpretations, and distinguish allocation sites from retention. |
| SIGQUIT appeared alongside a passive dump without its terminating effect; profiling cost was declared uniformly safe. | Prefer an existing admin dump, state SIGQUIT's default exit behavior, and bound workload-dependent captures. |

The [core skill](../skills/go-troubleshooting/SKILL.md) stays at 199 lines.
Two new references provide [ticket investigation](../skills/go-troubleshooting/references/TICKET-INVESTIGATION.md)
and [data-flow tracing](../skills/go-troubleshooting/references/DATA-FLOW-TRACING.md).
The router, ownership map, bilingual README, manifest counts, and description
fixture are synchronized. The pack has 60 references, 84 trigger evals, and
33 quality evals; six trigger cases and quality cases 28–33 are new.

Runtime corrections were checked against the official [signal documentation](https://pkg.go.dev/os/signal#hdr-Default_behavior_of_signals_in_Go_programs),
[diagnostics guide](https://go.dev/doc/diagnostics), and [GC memory-limit guide](https://go.dev/doc/gc-guide#Memory_limit).
Local Go documentation also confirmed `http.Server.Shutdown` does not wait for
hijacked connections and `maps.Values` has unspecified iteration order.
These are targeted corrections, not a fresh audit of every diagnostic command.

## Behavioral Probes

[Evidence with prompts, answers, tool calls, hashes, and assessments](evidence/2026-09-05-go-troubleshooting-probes.json)
records three snapshots: baseline, updated, and final. Fresh GPT-6 subagent
contexts applied the skill to the six cases, then four affected cases after
refinement. Assertions and prior answers were withheld from those agents.

Actual Claude Code **2.1.261** ran **`claude-opus-5`**, confirmed in the CLI
initialization records, at `medium` effort. Each case used a fresh directory,
explicitly read the isolated skill snapshot and routed references, and had only
`Skill`, `Read`, `Glob`, and `Grep` available. Hooks were disabled, MCP was empty
and strict, and session persistence was disabled. This tests application of the
instructions to supplied artifacts; it does not execute a Go fix or investigate
a live system. The parent manually assessed the complete answers.

### Baseline and Refinement

GPT-6 already made the correct core decisions on all six baseline prompts,
including refusing a premature cache diagnosis and preserving the shared parser.
The instruction corrections are useful on their own; this baseline does not
establish an increase in GPT-6 quality.

Opus baseline case 33 rejected SIGQUIT and a lone-stack leak diagnosis, but
repeated two false deductions from the old skill: two `Lock` frames prove
lock-order inversion, and grouped channel waits prove the peer has disappeared.
The updated response no longer made those specific deductions.

The first updated Opus pass exposed extra mistakes beyond the original rubric:
assuming repeated unordered `LIMIT 1` queries select the same row, treating an
unreported crash as evidence against a race, attributing overwrite survival to
map iteration order, and defaulting all nonpositive inputs when only a missing
parameter needed correction. Added focused guidance and strengthened assertions
for these issues; kept the incident prompts unchanged and reran cases 28, 30,
31, and 32 in fresh contexts on both models. These are development iterations,
not a held-out benchmark.

| Case | Last observed behavior |
|---|---|
| 28: alleged cache bug, cross-tenant row | Both models locate the dropped tenant predicate, preserve analysis-only scope, and allow different rows from unordered selection. |
| 29: stage differs from local | Both keep deployed-source uncertainty and propose controlled version/config comparison instead of patching correct local SQL. |
| 30: incomplete ticket | Both report unresolved, request a concrete correlated example, and avoid inferring crash absence from missing logs. |
| 31: rows lost after Scan | Both identify the map-key collision, distinguish assignment from iteration, preserve both authorized tenants, and avoid edits. |
| 32: caller-specific default | Both use parameter presence to apply the GET-only default, retain export semantics, and mark regression/gate execution unavailable. |
| 33: live hang | Both choose the existing nonterminating pprof endpoint and reject a lone blocked frame as proof. Opus retains additional overstatements described below. |

Cases 29 and 33 use the updated snapshot results; only the four affected cases
were rerun after the final clarification. Source hashes identify every tested
snapshot rather than implying all responses used identical files.

### Automatic Selection

Six separate native Claude `Skill` probes omitted forced reads. All four new
investigation prompts selected `golang-skills:go-troubleshooting`; both
translation/rewording controls selected no skill. Selection passed **6/6** for
this small sample. GPT-6 discovery and the full pack's trigger suite were not
run; GPT-6 application probes explicitly loaded the skill.

## Limits and Local Checks

One run per case/model/round cannot establish reliability. The cases use
illustrative inline artifacts, and several are close to the reference examples.
There was no live incident investigation or executed before/after Go regression.

The updated Opus hang answer still incorrectly calls a growing population
necessary for leakage and says flat counts with only parked workers put the
hang elsewhere. A single orphaned goroutine can leak, and stable counts alone
do not locate a hang. The skill requires expected-lifetime and wait-dependency
evidence, but this observed model deviation remains; the runtime case is not a
claim of flawless diagnosis. Opus answers were also substantially longer than
GPT-6 answers. Full answers are retained instead of reporting a blanket quality
pass rate.

Local verification: `go test -count=1 ./...` from `evals/` (including schema,
structure, manifest counts, script checks and lint-config validation),
`agentskills-validate@1.0.1` for edited skill entrypoints, the CI relative-link
check, and `git diff --check`. Results apply to repository instructions and
fixtures, not to the illustrative services used in the model prompts.
