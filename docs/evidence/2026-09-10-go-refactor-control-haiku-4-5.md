# Go Refactor Skill Control — Haiku 4.5 on the Claude CLI (n=1)

The first A/B run of this corpus on Haiku. Anthropic's skill-authoring guidance
asks for a check on Haiku, Sonnet and Opus; the repository had Sonnet 5 and
Opus 5 controls and nothing on Haiku. This closes that gap with the cheapest
run that can close it, and it is the first run to carry the `fix_hunks` metric.

**What one repetition can support.** Eight sessions, one per cell, give no
variance estimate, no interval and no p-value, so nothing below is evidence
that the skill moved a structural number on this model. What a single
repetition does establish is categorical: whether the corpus runs on Haiku at
all, whether the plugin is reached, whether behavior survives, and what the new
metric reads. Those are the findings here; the line deltas are recorded so the
next run has a reference, not because four pairs support a claim.

## Run

- Finished: 2026-09-10 16:43 UTC
- Runner: `claude` 2.1.267
- Model: `claude-haiku-4-5-20251001`, no reasoning effort set
- Seed: `1`
- Corpus: `refactor`
- Fixtures: `dispatch`, `pricing`, `report`, `store`
- Arms: `no-skill`, `baseline`
- Repetitions: 1 per fixture and arm, 8 sessions total
- Baseline plugin SHA-256:
  `a36a44f80c8a3b5cbac880c82f89f22293a7f1a7025d5a3a377064595b498d30`
- Plugin source: the **uncommitted working tree** at `3fb9bc2`, which carries 93
  modified files under the plugin subdirectories. This arm is therefore not any
  committed revision and cannot be recovered from a commit hash; the digest
  above is what pins it.
- Raw report: [`2026-09-10-go-refactor-control-haiku-4-5.json`](2026-09-10-go-refactor-control-haiku-4-5.json)
- Raw report SHA-256:
  `1760ceb297617460cda40e4dbc46ea0d50e7b64d0850f31b61fc990474bc9333`
- Session transcripts:
  [`2026-09-10-go-refactor-control-haiku-4-5.traces.tar.gz`](2026-09-10-go-refactor-control-haiku-4-5.traces.tar.gz),
  one `traces/<arm>-<fixture>.jsonl` per session
- Cost: $0.1287 control, $0.2382 baseline, $0.3669 total (1.85x)

```bash
go run ./cmd/abrun -arms no-skill,baseline -n 1 -j 4 -seed 1 \
  -model claude-haiku-4-5-20251001 -keep \
  -out ../docs/evidence/2026-09-10-go-refactor-control-haiku-4-5.json
```

All 8 sessions completed without a CLI error and all 8 built. No session wrote
a test: the claude arms are granted `Skill,Read,Glob,Grep,Edit,Write` and no
shell, so `model_tests` is `skipped` throughout.

## The finding: the plugin is mostly not reached on Haiku

`go-code-refactor` fired in **1 of 4** baseline sessions (`dispatch`). The other
three never called the `Skill` tool at all.

The arm loaded correctly, so this is routing and not a broken arm. Each baseline
session's `init` event lists the `Skill` tool and all 25 `golang-skills:*`
skills, and each control session lists neither — the control is clean.

| Model / runner | Baseline sessions that reached `go-code-refactor` |
|---|---|
| `gpt-5.6-luna` medium, codex | 20/20 |
| Opus 5 medium, claude | 19/20 |
| Haiku 4.5, claude | 1/4 |

On this model a `no-skill` versus `baseline` comparison is therefore mostly a
comparison of the same unaided model against itself, and the structural columns
below inherit that. The routing question has to be answered before a wording
question is worth asking on Haiku.

The one session that did reach the skill produced the run's worst structural
result — `dispatch` at +13 lines and +3 functions, where the unaided control
returned −1. That is one session. It is an anecdote, and it is recorded here so
the next Haiku run knows where to look.

## Structural results

Mean production-line delta. Negative is less production code. Every cell is a
single session.

| Fixture | No skill | Skill |
|---|---:|---:|
| `dispatch` | −1 | +13 |
| `pricing` | −38 | −27 |
| `report` | +20 (behavior broken) | +19 |
| `store` | −6 | −6 |
| Valid runs | −15.0 (n=3) | −0.2 (n=4) |

The two corpus means are not comparable as printed: the control's `report` run
broke behavior and is excluded, and `report` is the fixture where both arms grow
the most. Excluding `report` from both arms gives −15.0 against −6.7, still on
one session per cell.

`store` is identical in both arms, and `pricing` moves 20 branches out of the
code in both. Neither arm created an interface or a pattern-flavored name
anywhere in the run: `Δiface` and `Δpattern` are 0.00 in both columns, which on
this corpus is the over-engineering trap not being taken by either arm.

## Correctness

| | Build | Golden | Behavior failures |
|---|---|---|---|
| `no-skill` | 4/4 | 3/4 | 1 |
| `baseline` | 4/4 | 4/4 | 0 |

The control's `report` session changed the rendered column widths — the golden
overlay reads `123.45` where the fixture pads to `   123.45` — which is exactly
the observable contract the prompt told it to keep. This is the only cell in
the run where the arms separate on something other than routing, and it is one
session.

## Pending modernizations — first reading of `fix_hunks`

| Fixture | Before | No skill | Skill |
|---|---:|---:|---:|
| `dispatch` | 1 | 0 | 0 |
| `pricing` | 0 | 0 | 0 |
| `report` | 0 | 0 | 0 |
| `store` | 0 | 0 | 0 |
| Mean per valid run | 0.33 control, 0.25 skill | 0.00 (3/3 clean) | 0.00 (4/4 clean) |

`dispatch` ships with one modernization pending, `for i := 0; i < len(events); i++`
in `DispatchAll`, and both arms removed it. Nothing else in the corpus has one,
and no session introduced a construct the toolchain would undo.

So the metric separates nothing here — it is at its floor in both arms — and its
value in this run is the negative result: eight sessions of rewriting produced
no code `go fix` wants to change. Making it discriminate would need a fixture
that ships with several pending modernizations across different analyzers, in
the way `pricing` ships with duplicated plan data. The metric also says nothing
about the constructs no analyzer covers (`cmp.Or`, `errors.Join`, `iter.Seq`),
so it stays a floor rather than a modernity score.

## What this changes

- Haiku now has a control, and it says the corpus runs and behavior mostly
  survives. It does not say anything about a wording.
- The next Haiku question is routing, not concision: does `go-code` reach the
  model at all on this tier, and does the description or the invocation path
  need to change for it. Measuring a skill's wording on a model that loads the
  skill in a quarter of sessions measures the router.
- The 1.6.0 routing gate cannot cover this gap. `go-code-routing.sh` blocks the
  first `.go` edit only in a session that loaded `go-code`, and none of these
  sessions loaded anything, so the gate stayed silent by design. A hook that
  fires on the edit rather than on the router would be a different mechanism,
  and this run is one reason to consider it — not four.
- A Haiku claim about structure needs n≥5 per cell. At $0.046 per session that
  is about $1.85 for the full corpus at n=5, so cost is not the constraint.
