# Budget Closure Smoke — `catalog`, `feed`, `gateway`, three arms, Sonnet 5 medium (n=1)

The cross-model check on the 2026-09-12 edits before they are released. One
repetition per fixture and arm supports no claim about size, correctness
rate, or cost, and none is made below. What this run reads is whether the
edits measured on Opus 5 the same morning break anything on the model whose
`gateway` and `feed` readings have been the volatile ones.

The edits in the `baseline` arm are those of the
[Opus 5 smoke](2026-09-12-go-implement-budget-closure-smoke-opus-5-medium.md),
plus the `go-testing` `synctest.Sleep` caveat added after it; the plugin
digest differs from the Opus runs for that one line. The `reference` arm is
the committed tree at `2fab3d6` (release 1.12.0).

## Run

- Finished: 2026-09-12 09:09 UTC
- Runner: `claude` 2.1.267
- Model: `claude-sonnet-5`, reasoning effort `medium`
- Seed: `1`; `-j 3`
- Corpus: `implement`; fixtures: `catalog`, `feed`, `gateway`
- Arms: `no-skill`, `reference`, `baseline`; 1 repetition each, 9 sessions
- `reference`: a `git worktree` of `2fab3d6`, plugin SHA-256
  `778830a584281cc9fcd6949cabaf2c9596ee73f581e0e22b670832ab9acffe8f`
- `baseline`: the working tree at `2fab3d6` plus the uncommitted edits,
  plugin SHA-256 `4a0e9f43902d97dc256038a4cb5b2174ad3d3c58683006420673d7f945441f06`
- Toolchain: Go 1.27.1 linux/amd64; golangci-lint 2.13.2 for the lint line
- Raw report: [`2026-09-12-go-implement-budget-closure-smoke-sonnet-5-medium.json`](2026-09-12-go-implement-budget-closure-smoke-sonnet-5-medium.json)
  (SHA-256 `8882446cf00caebecc5064f5a2bdb0a8a10a69ee38504537e397bb663f75e39f`)
- Session transcripts: [`2026-09-12-go-implement-budget-closure-smoke-sonnet-5-medium.traces.tar.gz`](2026-09-12-go-implement-budget-closure-smoke-sonnet-5-medium.traces.tar.gz)
  (SHA-256 `9abc9819518016dce9bab34242b26636311bd6c0929eb2c82c7787c6570f5a6a`), one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $0.1295 control, $0.8166 reference, $0.9810 baseline — **6.31x** and **7.58x** the control

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <worktree of 2fab3d6> -arms no-skill,reference,baseline \
  -tasks catalog,feed,gateway -n 1 -j 3 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-12-go-implement-budget-closure-smoke-sonnet-5-medium.json
```

All 9 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. The tool set had no shell. `go-code` fired
first in 3/3 reference and 3/3 baseline sessions.

## Results

| Fixture | Arm | Golden | Δlines | Δfuncs | Δclos | lint before→after | `Skill` msgs | API calls | Edits | $ / run |
|---|---|---|---:|---:|---:|---:|---:|---:|---:|---:|
| `catalog` | no-skill | 1/1 | 22 | 0 | 0 | 0→0 | — | 5 | — | 0.035 |
| `catalog` | reference | 1/1 | 19 | 0 | 0 | 0→0 | 3 | 11 | 3 | 0.240 |
| `catalog` | baseline | 1/1 | 21 | 0 | 0 | 0→0 | 3 | 18 | 7 | 0.317 |
| `feed` | no-skill | 1/1 | 47 | 0 | 0 | 0→0 | — | 7 | — | 0.044 |
| `feed` | reference | 1/1 | 32 | 0 | 0 | 0→1 | 5 | 14 | 4 | 0.273 |
| `feed` | baseline | 1/1 | 47 | 0 | 0 | 0→0 | 3 | 21 | 8 | 0.349 |
| `gateway` | no-skill | **0/1** | 63 | 0 | 0 | 0→3 | — | 7 | — | 0.050 |
| `gateway` | reference | 1/1 | 66 | 1 | 0 | 0→3 | 3 | 12 | 4 | 0.304 |
| `gateway` | baseline | 1/1 | 69 | 0 | 1 | 0→1 | 3 | 11 | 3 | 0.315 |

| Arm | golden | `Skill` msgs | API calls | cache writes | cache reads | output tokens | thinking tokens | median report chars | $ / run |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|
| `no-skill` | 2/3 | 0.0 | 6.3 | 3.5K | 59K | 1636 | 441 | 336 | 0.043 |
| `reference` | 3/3 | 3.7 | 12.3 | 37.5K | 358K | 4951 | 1193 | 996 | 0.272 |
| `baseline` | 3/3 | 3.0 | 16.7 | 37.6K | 530K | 6960 | 2210 | 806 | 0.327 |

Skill loads — reference: `catalog` `go-code`, `go-error-handling`,
`go-testing`, `go-style-core` in 3 messages; `feed` those plus
`go-data-structures` and `go-linting` in 5; `gateway` `go-code`,
`go-style-core`, `go-http`, `go-security`, `go-testing`, `go-defensive`,
`go-error-handling` in 3. Baseline: `catalog` reference's four plus
`go-linting` in 3; `feed` `go-code`, `go-data-structures`, `go-testing`,
`go-style-core`, `go-error-handling` in 3; `gateway` `go-code`, `go-http`,
`go-error-handling`, `go-security`, `go-testing`, `go-style-core` in 3.

Budget lines: reference `gateway` `1 — methodNotAllowed: shared by the HEAD
/healthz, HEAD /accounts, HEAD /accounts/{id} routes`, a package-level
function; baseline `gateway` `0 — body written behind the existing NewServer
signature only`, with `methodNotAllowed` as a closure inside `NewServer`.
The other four skilled sessions reported `0`.

## Reading

- **Nothing broke.** Golden 3/3 in both skilled arms against 2/3 unaided;
  the unaided miss is `gateway` answering `HEAD` with 200, the same clause as
  every unaided Opus 5 miss today. `feed` rendered `kinds` as `[]` in all
  three arms; the 2026-09-11 baseline `null` did not recur here.
- **The closure rule did not act on Sonnet 5 in this one session.** The
  arms came out reversed against Opus 5: the reference session wrote
  `methodNotAllowed` as a package function and counted it, the baseline
  session wrote it as a closure and reported `0`. One session; the Opus 5
  `gateway` pair at n=5 is the measurement of the rule, and a Sonnet 5 pair
  at n=5 is what would say whether the model reads the new paragraph.
- **The baseline cost more here** — $0.33 against $0.27 a session, 16.7
  against 12.3 API calls, cache reads 530K against 358K — and the extra
  calls are edit rounds, not reads: the baseline `catalog` and `feed`
  sessions made 7 and 8 `Edit` calls against 3 and 4, re-reading their own
  contract test between them. No session in either arm read a `references/`
  file, so the moved examples and the cut rules induced no extra loading.
  `Skill` messages 3.0 against 3.7. Noise at n=1 on a model that iterates.
- **Lint.** The unaided `gateway` session left three findings, the reference
  one three, the baseline one; the reference `feed` session one. The
  skills' gate never runs without a shell on either model.
- **Size:** `catalog` 19 → 21, `feed` 32 → 47, `gateway` 66 → 69. One
  repetition; the Sonnet `feed` spread in the n=5 and n=10 runs is 30–52
  lines.

## What this changes

- The 2026-09-12 edits ship for both models: no correctness regression on
  either, and the Opus 5 measurement stands.
- The README's Sonnet 5 cells are unchanged: n=1 does not displace the n=5
  and n=10 runs that set them.
