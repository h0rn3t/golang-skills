# Go Refactor Skill Control — Sonnet 5 at medium effort on the Claude CLI (n=1)

The same eight-cell run as the
[Haiku 4.5 control](2026-09-10-go-refactor-control-haiku-4-5.md) — same seed,
same fixtures, same plugin tree, same day — on Sonnet 5 at `-effort medium`. It
exists to give the Haiku routing result a same-conditions comparison, and it
carries the second reading of the `fix_hunks` metric.

**What one repetition can support.** Eight sessions, one per cell, give no
variance estimate, no interval and no p-value. Every line-delta figure below is
a single session and is recorded as a direction and a reference for the next
run, not as an effect. The categorical results — what loaded, what built, what
kept its behavior — are what a single repetition can carry.

## Run

- Finished: 2026-09-10 17:16 UTC
- Runner: `claude` 2.1.267
- Model: `claude-sonnet-5`, reasoning effort `medium`
- Seed: `1`
- Corpus: `refactor`
- Fixtures: `dispatch`, `pricing`, `report`, `store`
- Arms: `no-skill`, `baseline`
- Repetitions: 1 per fixture and arm, 8 sessions total
- Baseline plugin SHA-256:
  `a36a44f80c8a3b5cbac880c82f89f22293a7f1a7025d5a3a377064595b498d30`
- Plugin source: the **uncommitted working tree** at `3fb9bc2`, 93 modified
  files under the plugin subdirectories — the same tree the Haiku run used, and
  the same digest, so the two runs differ only in the model and the effort.
  This arm is not a committed revision; the digest is what pins it.
- Raw report: [`2026-09-10-go-refactor-control-sonnet-5-medium.json`](2026-09-10-go-refactor-control-sonnet-5-medium.json)
- Raw report SHA-256:
  `e780aea745fbb69bf097addcc106568bd383a3ebe6b8acda9325f12929f7653d`
- Session transcripts:
  [`2026-09-10-go-refactor-control-sonnet-5-medium.traces.tar.gz`](2026-09-10-go-refactor-control-sonnet-5-medium.traces.tar.gz),
  one `traces/<arm>-<fixture>.jsonl` per session
- Cost: $0.2281 control, $0.8986 baseline, $1.1267 total (3.94x)

```bash
go run ./cmd/abrun -arms no-skill,baseline -n 1 -j 4 -seed 1 \
  -model claude-sonnet-5 -effort medium -keep \
  -out ../docs/evidence/2026-09-10-go-refactor-control-sonnet-5-medium.json
```

All 8 sessions completed, built, and passed the hidden golden test: 8/8 valid,
zero behavior failures, zero harness collisions. No session wrote a test — the
claude arms get `Skill,Read,Glob,Grep,Edit,Write` and no shell.

## Routing: 4/4, against 1/4 on Haiku

`go-code-refactor` fired in **every** baseline session. The `report` session
also loaded `go-style-core`. No control session loaded anything, as the control
arm carries no plugin.

| Model / runner (same corpus) | Baseline sessions that reached `go-code-refactor` |
|---|---|
| Sonnet 5 medium, claude (this run) | 4/4 |
| Opus 5 medium, claude (2026-09-08, n=5) | 19/20 |
| `gpt-5.6-luna` medium, codex (2026-09-07, n=5) | 20/20 |
| Haiku 4.5, claude (2026-09-10, n=1) | 1/4 |

Same plugin tree, same digest, same prompt, same day: the arm loads and the
skill is reachable. Haiku's 1/4 is therefore a property of the model tier, not
of the arm, and this run is the control that establishes it.

## Structural results

Production-line delta, one session per cell. Negative is less production code;
the effect column is baseline minus no-skill, so negative favors the skill.

| Fixture | No skill | Skill | Effect |
|---|---:|---:|---:|
| `dispatch` | +2 | −10 | −12 |
| `pricing` | −35 | −47 | −12 |
| `report` | +14 | 0 | −14 |
| `store` | −6 | −6 | 0 |
| All runs | −6.2 (n=4) | −15.8 (n=4) | −9.6 |

Three fixtures move the same way and none moves against the skill. That is the
shape earlier multi-repetition runs found on other models, reproduced here at
one repetition — a direction worth spending n=5 on, not a measured effect.

`report` is the strongest cell. The control took the over-engineering bait,
adding three functions and 14 lines; the skilled session rewrote the package to
exactly the same size (50 → 50 lines, functions 1 → 1) with the golden test
green, and its final message flagged that it had no shell to run the
verification with rather than claiming it had. New functions across the corpus
are +1.25 per control run against +0.25 per skilled run, and the concision gate
passes 4/4 against 2/4.

Neither arm created an interface or a pattern-flavored name anywhere in the run
(`Δiface` and `Δpattern` are 0.00 in both columns), so the structural separation
here is size and helper count, not the interface trap.

`store` is identical in both arms, as it was on Haiku.

## Correctness

| | Build | Golden | Behavior failures | Model tests |
|---|---|---|---|---|
| `no-skill` | 4/4 | 4/4 | 0 | none written |
| `baseline` | 4/4 | 4/4 | 0 | none written |

Correctness is tied and saturated. The Haiku control's one behavior break
(`report`, changed column widths) does not reproduce here in either arm.

## Pending modernizations (`fix_hunks`)

| Fixture | Before | No skill | Skill |
|---|---:|---:|---:|
| `dispatch` | 1 | 0 | 0 |
| `pricing` | 0 | 0 | 0 |
| `report` | 0 | 0 | 0 |
| `store` | 0 | 0 | 0 |
| Mean per valid run | 0.25 | 0.00 (4/4 clean) | 0.00 (4/4 clean) |

Identical to the Haiku reading: both arms removed `dispatch`'s
`for i := 0; i < len(events); i++`, and no session introduced anything `go fix`
would undo. Two models and sixteen sessions now agree that this corpus keeps the
metric at its floor, which is the negative result the metric can give here and
also the reason it cannot discriminate on these fixtures. A fixture that ships
with several pending modernizations across different analyzers is what would
change that.

## Reading this next to the 2026-09-08 Sonnet 5 control

The earlier [Sonnet 5 control](2026-09-08-go-refactor-control-sonnet-5.md) ran
n=5 without an effort setting against a different plugin tree and reported a
corpus difference of −3.20 lines. This run reports −9.6 at n=1 with medium
effort and a working tree carrying 93 modified plugin files. Three conditions
moved at once, and one repetition cannot attribute the gap to any of them.
Treat the two as separate runs, not as a before/after: a control is a property
of the model version and the tree it ran against, and comparing a skill edit
across runs needs the same conditions on both sides.

## What this changes

- Haiku's low routing rate is confirmed as model-specific rather than a broken
  arm, on the same tree and the same day.
- On Sonnet 5 medium the direction of the effect looks stronger than the
  2026-09-08 measurement, and `report` — the fixture built as the
  over-engineering trap — separates the arms completely in this pair. That is
  the cell to spend n=5 on next.
- Cost is 3.94x for the skilled arm, higher than the 3.28x of the earlier
  Sonnet control, at $0.14 per session across both arms. A full n=5 corpus at
  these conditions is roughly $5.60.
