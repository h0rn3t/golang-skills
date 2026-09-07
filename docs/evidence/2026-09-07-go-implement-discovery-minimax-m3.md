# Implementation Corpus Discovery — MiniMax M3 on opencode

A negative result. The four implementation fixtures do not yet earn their
place, and this file records why so the next attempt does not repeat it.

## Run

- Finished: 2026-09-07 11:48 EEST
- Corpus: `implement` (`evals/ab/_implement`)
- Runner: `opencode` 1.18.29
- Model: `opencode-go/minimax-m3`
- Seed: `1`
- Fixtures: `catalog`, `feed`, `gateway`, `ledger`
- Arms: `no-skill`, `baseline`
- Repetitions: 3 per fixture and arm, 24 sessions total
- Baseline plugin SHA-256: `6443a658f74cd465b98f0065cdb7c117bcde52c64af1acc1a5ef7f3a67369d3c`
- Raw report: [`2026-09-07-go-implement-discovery-minimax-m3.json`](2026-09-07-go-implement-discovery-minimax-m3.json)
- Raw report SHA-256: `dfe2c5bd0063614aa8e3b8008258c95a866c523c5ed46a8e050eb08e292d6416`
- Total cost: $0.33

```bash
go run ./cmd/abrun -corpus implement \
  -runner opencode -model opencode-go/minimax-m3 \
  -arms no-skill,baseline -n 3 -j 4 -seed 1 -keep \
  -out ../docs/evidence/2026-09-07-go-implement-discovery-minimax-m3.json
```

The report predates the `Δexp` and `Δbranch` metrics, so its results carry
neither field.

## Result

| Fixture | Control golden | Skill golden | Skill loaded | Control Δlines | Skill Δlines |
|---|---:|---:|---:|---:|---:|
| `catalog` | 3/3 | 3/3 | 3/3 | +23.3 | +18.0 |
| `feed` | 3/3 | 3/3 | **0/3** | +24.7 | +24.7 |
| `gateway` | 3/3 | 3/3 | 2/3 | +14.0 | +28.0 |
| `ledger` | 3/3 | **2/3** | 1/3 | +15.7 | +17.0 |

## Why the fixtures are not admitted

**The traps are saturated.** The control passed the hidden specification in 12
of 12 runs. Unaided, the model returned a non-nil slice so the JSON rendered
`[]`, set all four server timeouts, cloned the caller's slice, and wrapped with
`%w`. The traps themselves are live: every golden test fails against the stub's
panic and passes against an idiomatic reference implementation, checked before
the fixtures were committed. The model just does not fall into them. Because
the control was perfect on a small and cheap model, a stronger model cannot do
worse, so this is not a matter of choosing a different one.

**Triggering ruled out the comparison.** A `go-*` skill loaded in only 6 of 12
baseline runs and in none of the three `feed` runs, so half the skilled arm was
a control that happened to cost more. No claim about conciseness survives that.

**The structural metrics had no room.** `Δtypes`, `Δiface`, `Δfuncs` and
`Δpattern` were zero in all 24 runs. Pinning every exported declaration in the
stub is what makes the golden test compile, and it is also what leaves the model
nothing to over-engineer. `Δlines` was the only structural metric that moved;
`Δbranch` and `Δexp` were added afterwards for exactly this reason.

## What this run does not show

It does not show that the skill makes an implementation longer. The single
worst line count in the table — the 43-line `gateway` run behind that arm's
+28.0 mean — is a baseline run in which **no skill loaded at all**. The two
`gateway` runs that did load a skill came in at 17 and 24 lines against a
control of 3, 17 and 22. At `n=3`, with a control spread from 3 to 22 lines on
one fixture, nothing separates the arms.

## Triggering, measured separately

The 50% firing rate is specific to this runner and model. Seven trigger evals
written from the exact phrasings that missed here were added to the `train` set
and run against the `claude` CLI, where all seven passed: `go-code` fires for
"the bodies all panic, fill in the implementations" and for a greenfield
package, `go-data-structures` for the JSON feed document, `go-defensive` for the
immutable snapshot over a caller's slice, `go-http` for the edge server, and
`go-error-handling` for the lookup that must name the record and keep the
reason. Raw report:
[`2026-09-07-trigger-train-claude.json`](2026-09-07-trigger-train-claude.json),
SHA-256 `12748a1fae87139d8879ce5dee180cd5e4b8a08eb11d7e203a6dcc83fa611e34`,
Claude Code 2.1.261, 50 of 64 train-set trigger evals passing overall.

So the skill descriptions route implementation work correctly, and the gap
observed here is the model's, not the description's. Tuning descriptions against
it would be tuning against one small model.

## Next attempt

Two things have to change before this corpus produces a result.

The measurement has to run where the skill reliably loads, which on today's
evidence means the `claude` runner. Note that runner restricts the session to
`Skill,Read,Glob,Grep,Edit,Write`, so an implementation session there cannot
compile or test its own work; whether that restriction is right for a
from-scratch task is an open question the refactor corpus never had to answer.

The fixtures have to leave room to over-engineer. A stub that names every
exported declaration measures only the bodies. A fixture that pins one entry
point and leaves the structure behind it free would let `Δtypes`, `Δiface`,
`Δfuncs` and `Δbranch` move, which is where a conciseness claim would come from.
