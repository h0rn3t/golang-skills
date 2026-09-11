# Workflow Retune — `catalog`, `feed`, `gateway`, three arms, Opus 5 medium (n=3)

The same trees as the [Sonnet 5 medium run](2026-09-11-go-implement-newcode-workflow-sonnet-5-medium.md)
of the same hour, on Opus 5 at medium effort: the first measurement of the
Writing New Code section, the Plain Code retune, or the workflow edits on
this model. The last Opus 5 runs on this corpus are the
[2026-09-07 `gateway`](2026-09-07-go-implement-gateway-opus5.md) and
[`feed`/`catalog`](2026-09-07-go-implement-feed-catalog-opus5.md) runs on a
tree four releases older, with no effort flag. Three repetitions support no
size claim; the run records correctness and which instructions this model
follows.

## Run

- Finished: 2026-09-11 08:11 UTC
- Runner: `claude` 2.1.267
- Model: `claude-opus-5`, reasoning effort `medium`
- Seed: `1`; `-j 4`
- Corpus: `implement`, `-tasks catalog,feed,gateway`
- Arms: `no-skill`, `reference`, `baseline`
- Repetitions: 3 per fixture and arm, 27 sessions total
- `reference`: a `git worktree` of `ead8ce6`, plugin SHA-256
  `52832538f566efff400a1f2a5f1a1cf5e77877a8f2d60f08f0f86053c7d79a0b`
- `baseline`: the working tree with the workflow edits, plugin SHA-256
  `f816124fac7202590aa6849a796454920781ad3d2c63afca2007ecc4cc61a68f`
- Toolchain: Go 1.27.1 linux/amd64
- Raw report: [`2026-09-11-go-implement-newcode-workflow-opus-5-medium.json`](2026-09-11-go-implement-newcode-workflow-opus-5-medium.json)
  (SHA-256 `b25923bfbc708e730ed9b42c378ebbfc1cd7fecc61f48cfadb3e3082e1d68044`)
- Session transcripts:
  [`2026-09-11-go-implement-newcode-workflow-opus-5-medium.traces.tar.gz`](2026-09-11-go-implement-newcode-workflow-opus-5-medium.traces.tar.gz)
  (SHA-256 `ac8d7988a8c6ad8c8545022d93923817440d75ba354d6c58a9595b724227b73f`),
  one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $1.8457 control, $5.9592 reference, $6.5058 baseline — **3.23x** and
  **3.52x**

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-opus-5 -effort medium \
  -reference-root <worktree of ead8ce6> -arms no-skill,reference,baseline \
  -tasks catalog,feed,gateway -n 3 -j 4 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-11-go-implement-newcode-workflow-opus-5-medium.json
```

All 27 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. No shell in any arm. `go-code` fired in 6/9
reference and 7/9 baseline sessions; every miss loaded no skill at all — two
`feed` and one `catalog` session in the reference arm, one `feed` and one
`catalog` session in the baseline arm — so on this model the router's
description, not its text, is the ceiling on `feed` and `catalog`.

## Results

| Fixture | Arm | Golden | Valid | Δlines (valid) | Δfuncs | Δtypes | $ / run |
|---|---|---|---|---|---|---|---:|
| `catalog` | no-skill | 3/3 | 3 | 24, 26, 30 → 26.7 | 0 | 0 | 0.119 |
| `catalog` | reference | 3/3 | 3 | 22, 22, 19 → 21.0 | 0 | 0 | 0.538 |
| `catalog` | baseline | 3/3 | 3 | 19, 19, 22 → 20.0 | 0 | 0 | 0.490 |
| `feed` | no-skill | 3/3 | 3 | 51, 48, 49 → 49.3 | 0 | 2 ×3 | 0.179 |
| `feed` | reference | 3/3 | 3 | 47, 49, 33 → 43.0 | 0 | 2 ×2, 0 ×1 | 0.388 |
| `feed` | baseline | 3/3 | 3 | 34, 36, 47 → 39.0 | 0 | 2 ×1, 0 ×2 | 0.531 |
| `gateway` | no-skill | **1/3** | 1 | 147 | 6 ×1 | 1 ×1 | 0.318 |
| `gateway` | reference | 3/3 | 3 | 81, 85, 80 → 82.0 | 0 | 0 | 1.060 |
| `gateway` | baseline | 3/3 | 3 | 74, 79, 82 → 78.3 | 0 | 0 | 1.147 |

| Comparison | `catalog` | `feed` | `gateway` |
|---|---|---|---|
| baseline − no-skill, lines (perm p) | −6.7 (0.10) | −10.3 (0.10) | −68.7 (0.25) |
| baseline − reference, lines (perm p) | −1.0 (1.00) | −4.0 (0.80) | −3.7 (0.40) |
| golden, no-skill vs baseline (Fisher) | 3/3 vs 3/3 | 3/3 vs 3/3 | 1/3 vs 3/3 (0.40) |

| Arm | golden | contract tests written · before the body | budget line | `checks:` line | says no shell | `go-linting` loads | hook blocks / session | skills / session | turns | median report chars | $ / run |
|---|---|---|---|---|---|---|---:|---:|---:|---:|---:|
| `no-skill` | 7/9 | 5 · 0 | 0/9 | 0/9 | 5/9 | 0/9 | 0 | 0 | 15.2 | 1578 | 0.2051 |
| `reference` | 9/9 | 6 · 6 | 6/9 | 0/9 | 8/9 | 1/9 | 0.22 | 3.67 | 21.8 | 1802 | 0.6621 |
| `baseline` | 9/9 | 7 · 6 | 7/9 | 7/9 | 7/9 | **0/9** | 0.11 | 4.56 | 23.0 | **1151** | 0.7229 |

Skill loads over nine sessions — reference: `go-code` 6, `go-style-core` 6,
`go-error-handling` 6, `go-testing` 6, `go-data-structures` 3, `go-http` 3,
`go-defensive` 2, `go-linting` 1. Baseline: `go-code` 7, `go-style-core` 7,
`go-error-handling` 7, `go-testing` 7, `go-defensive` 5,
`go-data-structures` 5, `go-http` 3.

## Reading

- **Correctness: 9/9 in both skilled arms against 7/9 unaided.** Both
  control failures are `gateway` sessions answering `HEAD` with 200 on all
  three routes — the clause the skilled arms carry through `go-http`. On this
  model the unaided `gateway` is also the largest code in the run: 147, 128
  and 110 lines with three to six helpers, against 74–85 with none in either
  skilled arm. The 2026-09-07 Opus run had the same shape (152.6 → 99.8) at
  5/5 in both arms; the control's `HEAD` failures are new to this run and,
  at n=3, a dated observation.
- **Every session that loaded `go-code` followed the workflow**: 7/7 budget
  lines, 7/7 `checks:` lines, 7/7 no-shell statements, and 6 of 7 Contract
  Tables before the body (the seventh is a `feed` session that wrote the test
  after). The reference arm, on the same model, reported the budget in 6/6
  `go-code` sessions and wrote its test first in 6/6 as well: Opus 5 follows
  the prose form of the instruction that Sonnet 5 followed only with the
  template, so the edits change this model's behavior less.
- **`go-linting` loads: 0/9** against 1/9. The rule was already nearly
  followed here; the named tool closed the last case.
- **Reports are shorter**: median 1151 characters against 1802 for the
  reference tree and 1578 unaided. This is the one place the template moved
  Opus 5 — the shape replaced a longer narrative that the
  [2026-09-05 probes](../CROSS_MODEL_REVIEW.md#opus-verification-through-claude-cli)
  had already flagged.
- **`HEAD` in the model's own test: 3/3 in both skilled arms**, with the
  reference arm reaching it through `go-http` alone. `HEAD` patterns 2/3
  against 1/3, `r.Method` 1/3 against 2/3; all six pass.
- **Cost 3.52x against 3.23x**, +9% a session: `go-testing` 7 against 6,
  `go-defensive` and `go-data-structures` 5 against 2 and 3 — the `feed` and
  `catalog` sessions routed to more owners — and one more turn. The
  2026-09-07 Opus `gateway` run cost 1.67x without an effort flag.

## What this changes

- **README**: the Opus 5 new-code cell moves to this run — three fixtures,
  three arms, medium effort, 9/9 against 7/9, 3.5x — replacing the two-task
  2026-09-07 reading. It is n=3 and says nothing about size.
- **The workflow edits are safe on Opus 5**: no measure moved against them,
  correctness is level with the reference tree at 9/9, reports are shorter,
  the last shell-less `go-linting` load is gone.
- **The router description is the open item on this model**: 5 of 18 skilled
  sessions loaded nothing, all on `feed` and `catalog`, whose prompt names
  only the package to implement.
