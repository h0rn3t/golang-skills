# Go Refactor Skill Control — Sonnet 5 at high effort on the Claude CLI (n=1)

The third of three eight-cell runs made the same day against the same plugin
tree and the same seed. It differs from the
[Sonnet 5 medium control](2026-09-10-go-refactor-control-sonnet-5-medium.md) in
one flag: `-effort high`. Taken together, the pair is about reasoning effort;
taken with the [Haiku control](2026-09-10-go-refactor-control-haiku-4-5.md), the
three are about where on the capability range the skill has room to act.

**What one repetition can support.** One session per cell. No variance, no
interval, no p-value. The line deltas below are directions and references for
the next run, not effects. The categorical results — what loaded, what built,
what kept its behavior — are what a single repetition carries.

## Run

- Finished: 2026-09-10 17:26 UTC
- Runner: `claude` 2.1.267
- Model: `claude-sonnet-5`, reasoning effort `high`
- Seed: `1`
- Corpus: `refactor`
- Fixtures: `dispatch`, `pricing`, `report`, `store`
- Arms: `no-skill`, `baseline`
- Repetitions: 1 per fixture and arm, 8 sessions total
- Baseline plugin SHA-256:
  `a36a44f80c8a3b5cbac880c82f89f22293a7f1a7025d5a3a377064595b498d30`
- Plugin source: the **uncommitted working tree** at `3fb9bc2`, 93 modified
  files under the plugin subdirectories — the same tree and digest as the Haiku
  and Sonnet-medium runs, so those three differ only in model and effort.
- Raw report: [`2026-09-10-go-refactor-control-sonnet-5-high.json`](2026-09-10-go-refactor-control-sonnet-5-high.json)
- Raw report SHA-256:
  `938ec387bcf2c70cba9baa5826884c7d3247317779678afa18eb518919f15d4a`
- Session transcripts:
  [`2026-09-10-go-refactor-control-sonnet-5-high.traces.tar.gz`](2026-09-10-go-refactor-control-sonnet-5-high.traces.tar.gz),
  one `traces/<arm>-<fixture>.jsonl` per session
- Cost: $0.2407 control, $1.3484 baseline, $1.5891 total (5.60x)

```bash
go run ./cmd/abrun -arms no-skill,baseline -n 1 -j 4 -seed 1 \
  -model claude-sonnet-5 -effort high -keep \
  -out ../docs/evidence/2026-09-10-go-refactor-control-sonnet-5-high.json
```

8/8 valid: every session completed, built, and passed the hidden golden test,
with zero behavior failures in either arm. No session wrote a test — the claude
arms get no shell.

## The finding: effort lifts the control, and the gap closes

| Fixture | No skill | Skill | Effect |
|---|---:|---:|---:|
| `dispatch` | −11 | −10 | +1 |
| `pricing` | −36 | −41 | −5 |
| `report` | +12 | +12 | 0 |
| `store` | −3 | −6 | −3 |
| All runs | −9.5 (n=4) | −11.2 (n=4) | −1.8 |

Against the same corpus at medium effort:

| | Control | Skill | Effect |
|---|---:|---:|---:|
| Sonnet 5, `-effort medium` | −6.2 | −15.8 | −9.6 |
| Sonnet 5, `-effort high` | −9.5 | −11.2 | −1.8 |

Both columns moved, and they moved toward each other. The unaided model gained
3.3 lines of concision from the extra effort and the skilled arm lost 4.6, so
the separation at n=1 fell from −9.6 to −1.8. A control that solves more of the
task on its own leaves the skill less to do — the same shape the codex
implementation run recorded, where every trap saturated in both arms and no
fixture separated.

`report`, the over-engineering trap, is where the reversal is sharpest. At
medium the skilled session rewrote it to its original 50 lines while the control
grew 14; at high **both arms grew by exactly 12 lines and three functions**. The
extra effort did not make the skilled arm hold the line, it made both arms
elaborate. `store` moves the other way (−6 against −3), and `dispatch` ties
within a line.

New functions across the corpus tell the same story: +1.00 per control run
against +1.25 per skilled run here, where medium had +1.25 against +0.25. The
concision gate is 3/4 in both arms, where medium was 4/4 against 2/4.

None of this is an effect at one repetition per cell. It is a direction that
contradicts the medium run's direction, which is itself the reason to distrust
both numbers until n=5 and to treat effort as a condition that has to be fixed
before two runs are compared.

## Routing: 4/4, and deeper than at medium

`go-code-refactor` fired in every baseline session, as at medium, but the
sessions reached further into the tree:

| Fixture | Skills loaded (baseline) |
|---|---|
| `dispatch` | `go-code-refactor`, `go-style-core`, `go-testing` |
| `pricing` | `go-code-refactor`, `go-style-core` |
| `report` | `go-code`, `go-code-refactor`, `go-naming`, `go-style-core`, `go-testing` |
| `store` | `go-code-refactor` |

Eleven skill loads here against five at medium. The `report` session loaded five
skills including the `go-code` router and still produced the same +12 lines as
the control that loaded none. Firing rate and measured outcome are separate
results, and this run is the cleanest single example of it in the repository:
routing improved, the code did not.

## Correctness

| | Build | Golden | Behavior failures | Model tests |
|---|---|---|---|---|
| `no-skill` | 4/4 | 4/4 | 0 | none written |
| `baseline` | 4/4 | 4/4 | 0 | none written |

Tied and saturated, as at medium.

## Pending modernizations (`fix_hunks`)

| Fixture | Before | No skill | Skill |
|---|---:|---:|---:|
| `dispatch` | 1 | 0 | 0 |
| `pricing` | 0 | 0 | 0 |
| `report` | 0 | 0 | 0 |
| `store` | 0 | 0 | 0 |
| Mean per valid run | 0.25 | 0.00 (4/4 clean) | 0.00 (4/4 clean) |

The third identical reading. Twenty-four sessions across three model/effort
conditions have now cleared `dispatch`'s one pending hunk and introduced
nothing, which is a stable negative result and a firm statement that this
corpus cannot make the metric discriminate. A fixture shipping several pending
modernizations across different analyzers is the prerequisite for using it as a
comparison column rather than as a floor check.

## Cost

$0.3371 per skilled session against $0.0602 per control session, 5.60x — the
highest ratio recorded on this corpus (medium was 3.94x, the 2026-09-08 Sonnet
control 3.28x). The skilled arm both loads more skill text and reasons longer
about it. A full n=5 corpus at these conditions is roughly $8.

## What this changes

- Effort is a first-class condition of a control, not a detail: the same model
  and tree on the same day produce a −9.6 and a −1.8 corpus difference at two
  effort levels. Any before/after comparison has to hold it fixed, and every
  published run has to name it.
- The three runs sketch a range worth testing properly: Haiku barely reaches the
  skill, Sonnet medium reaches it and separates, Sonnet high reaches it more
  deeply and separates less. If that holds at n=5, the skill's value is
  concentrated in the middle of the capability range, and the `report` trap is
  the fixture to measure it on.
- Nothing here should be quoted as an effect. Three runs of one repetition
  disagree with each other by more than any of them claims to measure.
