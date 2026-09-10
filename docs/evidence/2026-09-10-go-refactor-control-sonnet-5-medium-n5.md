# Go Refactor Skill Control — Sonnet 5 at medium effort on the Claude CLI (n=5)

The full refactor corpus at five repetitions per cell, the run the
[n=1 medium control](2026-09-10-go-refactor-control-sonnet-5-medium.md) asked
for. It reproduces that reading almost exactly — corpus difference −9.7 lines
against −9.6 — and it is the first Claude-model run in this repository on which
**three of four fixtures separate the arms**, two of them past a Bonferroni
correction.

## Run

- Finished: 2026-09-10 19:09 UTC
- Runner: `claude` 2.1.267
- Model: `claude-sonnet-5`, reasoning effort `medium`
- Seed: `1`
- Corpus: `refactor`
- Fixtures: `dispatch`, `pricing`, `report`, `store`
- Arms: `no-skill`, `baseline`
- Repetitions: 5 per fixture and arm, 40 sessions total
- Baseline plugin SHA-256:
  `5028c91b61340fe99a22284cf0fe4ba32f48d8e7b1cf83baa0e412c88708a2ea`
- Plugin source: checkout `d0a83f360fac7d8cd419a40007a2e14b2ef29660`
  (release 1.7.0), clean apart from this report's own untracked files under
  `docs/evidence`; the arm copies only `.claude-plugin`, `skills`, `agents` and
  `hooks`
- Toolchain: Go 1.27.1 linux/amd64
- Raw report: [`2026-09-10-go-refactor-control-sonnet-5-medium-n5.json`](2026-09-10-go-refactor-control-sonnet-5-medium-n5.json)
- Raw report SHA-256:
  `31ba1c9f117fb17b7d857e25ba74e3b7e6f865365e858d4434f3c3f7d0e9b897`
- Session transcripts:
  [`2026-09-10-go-refactor-control-sonnet-5-medium-n5.traces.tar.gz`](2026-09-10-go-refactor-control-sonnet-5-medium-n5.traces.tar.gz)
  (SHA-256 `c022066af3508be3275cfe25274f7e7be758efc946ac154225eeeec8d0615ad8`),
  one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $0.8937 control, $3.9147 baseline, $4.8084 total (**4.38x**)

```bash
go run ./cmd/abrun -arms no-skill,baseline -n 5 -j 4 -seed 1 \
  -model claude-sonnet-5 -effort medium -keep \
  -out ../docs/evidence/2026-09-10-go-refactor-control-sonnet-5-medium-n5.json
```

All 40 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. Sessions ran with the tool set limited to
`Skill,Read,Glob,Grep,Edit,Write`, so neither arm had a shell and no session
could run `go build` or `go test` on its own work.

`permutation p` is the two-sided exact permutation test over every split of the
two arms' valid runs, unadjusted unless stated. Welch intervals are omitted: at
n=5 they are wider than the effects measured here.

## Structural results

Production-line delta against the shipped fixture; ± is the population SD.
Means cover valid runs only — build, hidden golden test, and the session's own
tests when it wrote any. The effect column is baseline minus no-skill, so
negative favors the skill.

| Fixture | No skill | Skill | Effect | perm p | valid | golden |
|---|---:|---:|---:|---:|---|---|
| `dispatch` | +5.8 ± 2.5 | −10.0 ± 2.6 | **−15.8** | **0.00794** | 5v5 | 5/5 · 5/5 |
| `pricing` | −36.6 ± 2.2 | −45.6 ± 4.8 | **−9.0** | 0.02381 | 5v5 | 5/5 · 5/5 |
| `report` | +14.8 ± 3.4 | +5.5 ± 6.1 | **−9.3** | **0.00794** | 5v4 | 5/5 · 5/5 |
| `store` | −9.2 ± 4.0 | −9.6 ± 4.8 | −0.4 | 1.00000 | 5v5 | 5/5 · 5/5 |
| Corpus | −6.30 | −16.00 | **−9.70** | — | 20v19 | 20/20 · 20/20 |

| Arm | runs | valid | Δlines | Δtypes | Δiface | Δfuncs | Δpattern | Δbranch | build | golden | skill | $/run |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| `no-skill` | 20 | 20 | −6.30 | 0.30 | 0.00 | 1.45 | 0.00 | −5.60 | 100% | 100% | 0% | 0.0447 |
| `baseline` | 20 | 19 | −16.00 | 0.32 | 0.00 | 0.58 | 0.00 | −7.11 | 100% | 100% | 100% | 0.1957 |

New functions fall 60% per session, 1.45 → 0.58. The interface bait was taken
nowhere: `Δiface` and `Δpattern` are 0.00 in both arms and new types are level
(0.30 against 0.32), so the separation here is size and helper count. The
concision gate passes 16/20 with the skill against 10/20 without it.

### `dispatch` — the arms do not overlap

Control deltas are +1, +6, +7, +7, +8; skilled deltas −12, −12, −11, −10, −5.
p = 0.00794 is 2/252, the floor for a 5v5 permutation test, and 0.032 after
Bonferroni across the four fixtures — zero stays excluded.

The mechanism is visible in the final messages. Four of five control sessions
converge on the same move: replace the if/else chain with a `switch` and extract
`put`/`del`/`eventKey` helpers — +2.40 functions per session, package 65 → 70.8
lines. The skilled sessions fold the three near-identical branches into one
selection plus one shared computation and add **+0.20** functions, 65 → 55.0
lines. Both arms preserve the error strings; the difference is whether the
duplication is removed or renamed, which is the distinction
`go-code-refactor`'s «once, not once under a name» clause exists to make.

Prior readings on this fixture: Sonnet 5 at the CLI default −4.4 (p = 0.381),
Opus 5 medium +3.4 against the skill, `gpt-5.6-luna` medium on codex −5.8,
Sonnet 5 medium n=1 −12. This run's −15.8 is the largest and the only one on a
Claude model whose arms do not overlap at all.

### `report` — the over-engineering trap

Control sessions grow the package every time: +12, +12, +12, +19, +19, adding
3.40 functions per session. Skilled valid sessions are −5, +8, +9, +10 with
1.50 functions, and one of them returns a *smaller* package with the golden test
green. p = 0.00794 (1/126 at 5v4), 0.032 after Bonferroni.

### `pricing` and `store`

`pricing` is the largest absolute reduction in the corpus in both arms — the
fixture ships 131 lines of duplicated plan literals — and the skill takes 9.0
more lines out of it (p = 0.024, 0.095 after Bonferroni). `store` is a tie
(−0.4, p = 1.00): both arms flatten the nesting and delete the duplicate miss
branch, and neither builds the `Repository` interface the fixture baits with.

## Correctness

| | Build | Golden | Behavior failures | Model tests |
|---|---|---|---|---|
| `no-skill` | 20/20 | 20/20 | 0 | none written |
| `baseline` | 20/20 | 20/20 | 0 | 2 written, 1 failed |

Correctness is tied and saturated: every session in both arms built and passed
the hidden golden test, and no session moved the observable contract.

One run is excluded from the structural means. `baseline`/`report` #0 kept
behavior — the golden test is green — and then failed a test it wrote itself:
its `TestRenderText` expected the `UNITS`/`AMOUNT` columns one space wider than
its own code produces. The refactor is sound and the assertion is wrong; the
harness excludes it because a failed model test is not evidence of a
behavior-preserving refactor, and the same session on the same fixture failed
the same way in the [2026-09-09 run](2026-09-09-go-controls-sonnet-5-medium.ru.md).
Test files were written by 2/20 skilled sessions and 0/20 control sessions, so
this exclusion class exists only in the skilled arm.

## Pending modernizations (`fix_hunks`)

| | Before | After | Runs with nothing left to propose |
|---|---:|---:|---|
| `no-skill` | 0.25 | 0.00 | 20/20 |
| `baseline` | 0.25 | 0.00 | 19/19 |

Third reading of this metric and the third at the floor. `dispatch`'s
`for i := 0; i < len(events); i++` is modernized by every session in both arms,
and no session introduced a construct `go fix` would undo. On this corpus the
metric confirms that nothing regressed and cannot discriminate between arms.

## Reported counts

16 of 20 skilled sessions state a line count in their final message; 0 of 20
control sessions do. The claims that could be checked against the harness agree
with it (`dispatch` #0 reports 66 → 56 physical lines against the harness's
65 → 55, the one-line difference being the trailing newline convention). Several
skilled sessions also state that no shell was available and that they verified
by hand instead of claiming a green build — the reporting behavior
`go-code-refactor` asks for, and the reason the trace is archived beside the
number.

## Skill routing

`go-code-refactor` fired in 20/20 baseline sessions, `go-style-core` in 2 and
`go-code-review` in 1. No control session loaded anything. Routing is therefore
not the limiting factor on this tier, unlike
[Haiku 4.5](2026-09-10-go-refactor-control-haiku-4-5.md) at 1/4.

## What this changes

- The medium-effort direction from the n=1 control reproduces at n=5: −9.6
  becomes −9.7 on the same fixtures, same model, same day, one commit later.
- `dispatch` and `report` are established results at this effort level, not
  directions: both survive Bonferroni, and `dispatch`'s arms do not overlap.
  `pricing` is exploratory after correction; `store` is a tie.
- The cost multiple is 4.38x, against 3.94x for the n=1 medium reading, 3.28x
  for the CLI-default control and 5.60x for the same-day high-effort run — at
  $0.12 per session across both arms.
- Effort remains part of a control's identity. This run says nothing about
  Sonnet 5 at `high`, where the [same-day control](2026-09-10-go-refactor-control-sonnet-5-high.md)
  measured −1.8 at n=1 because the unaided arm improves faster than the skilled
  one. Re-measure before comparing across effort levels.
