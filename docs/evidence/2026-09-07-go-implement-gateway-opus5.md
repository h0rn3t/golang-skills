# Implementation Corpus — `gateway` on Opus 5

The implementation corpus's first published result. Writing the same working
package from the same specification, the skilled arm produced a third less code
and never once landed in the control's range.

## Run

- Finished: 2026-09-07 12:44 EEST
- Corpus: `implement` (`evals/ab/_implement`)
- Runner: `claude`, Claude Code 2.1.261
- Model: `claude-opus-5`
- Seed: `20260907`
- Fixture: `gateway`
- Arms: `no-skill`, `baseline`
- Repetitions: 5 per arm, 10 sessions total
- Baseline plugin SHA-256: `6443a658f74cd465b98f0065cdb7c117bcde52c64af1acc1a5ef7f3a67369d3c`
- Raw report: [`2026-09-07-go-implement-gateway-opus5.json`](2026-09-07-go-implement-gateway-opus5.json)
- Raw report SHA-256: `90c45bbc3364c6ee02266c0aae3d1e3d4bfa3ea96339471a904765d68413f13d`
- Total cost: $5.50 — $2.06 control, $3.44 skilled

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-opus-5 \
  -tasks gateway -arms no-skill,baseline -n 5 -j 4 -seed 20260907 -keep \
  -out ../docs/evidence/2026-09-07-go-implement-gateway-opus5.json
```

All 10 sessions completed, built, and passed the hidden specification. Every
session changed the fixture and none referenced the repository. A `go-*` skill
loaded in all 5 baseline sessions: `go-code` in every one, with `go-http` and
`go-security` also reached in one of them.

## Results

Correctness is the gate and it is tied at 5/5 golden in both arms, so nothing
here says the skill makes the package more likely to work. The score is what a
working package cost.

| Metric | No skill | Skill | Effect |
|---|---:|---:|---:|
| Lines | 152.6 | 99.8 | **−52.8 (−34.6%)** |
| Functions | 6.80 | 3.80 | −44.1% |
| Branches | 12.60 | 7.80 | −38.1% |
| Types | 0.60 | 0.60 | no difference |
| Interfaces | 0 | 0 | no difference |
| Exported declarations | 0 | 0 | no difference |

Per-run line counts, sorted:

| Arm | Runs | Spread |
|---|---|---:|
| No skill | 109, 124, 169, 170, 191 | 82 |
| Skill | 94, 97, 100, 101, 107 | 13 |

The arms do not overlap: the control's leanest run is 109 lines and the
skilled arm's heaviest is 107. A Welch interval for the mean difference is
approximately −96 to −10 lines at 95% (df ≈ 4.2).

Two things are worth separating in that table. The first is the size: a third
less code for the same passing behavior. The second is the spread — population
standard deviation 30.9 without the skill against 4.4 with it, a sevenfold
reduction. The skilled arm did not merely write less, it wrote the same amount
every time, which is the more useful property of the two when the question is
what a change will cost to review.

Functions and branches fall roughly in step with the line count while types and
interfaces do not move at all. That is the shape of an implementation that is
simply smaller, not one that traded branches for dispatch: no arm grew an
interface, and `Δpattern` was zero throughout.

## What the fixture asks for

`gateway` pins one declaration, `NewServer(addr string, accounts []Account)
*http.Server`, and documents the edge exposure plus five routes with their
status codes, ordering and filtering. How the routing, filtering and encoding
behind that entry point are decomposed is entirely the implementation's choice,
which is where the 3.0 functions and 4.8 branches per run of difference came
from. The trap inside it is the edge timeouts — a zero timeout is no timeout —
and both arms set them, so the trap contributes nothing to this result.

## Limits

One fixture, one model, five runs per arm. The skill's effect on this task is
concentrated in how much scaffolding gets built around a small routing job; it
does not generalize to implementation work of other shapes without measuring
them, and the corpus's three other fixtures show how easily a from-scratch
task turns out to have only one sensible shape. See the
[companion run](2026-09-07-go-implement-feed-catalog-opus5.md) for two that do.

Correctness is tied, so this is not evidence about defect rates. The claude
runner restricts the session to `Skill,Read,Glob,Grep,Edit,Write`, so none of
these implementations could compile or test itself before being measured.
