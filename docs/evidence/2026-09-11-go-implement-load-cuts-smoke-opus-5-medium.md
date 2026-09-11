# Load Cuts Smoke — `catalog`, `feed`, `gateway`, three arms, Opus 5 medium (n=1)

A first look at the 1.12.0 loading edits on the model whose traces motivated
them. One repetition per fixture and arm supports no claim about size,
correctness rate, or cost, and none is made below. What this run reads is
whether the owners now load in one message and whether correctness held.

The edits in the `baseline` arm, released as 1.12.0:

- **`go-code` step 3** issues the `Skill` calls for every selected owner in one
  message; the routing gate's message for a blocked edit says the same.
- **`go-testing`, `go-naming`, `go-documentation`**: the `> **Validation**`
  callouts that told the model to run the new tests now, or a script and then
  `go build`, are one sentence each routing to the `go-linting` gate.
- **`go-style-core`** drops Declarations and Scope, Loops and Switches, and
  Naked Returns; **`go-code`** compresses its Resource Routing preamble and
  states the contract-test reason once; **Related Skills** in 22 skills are one
  line per pointer.

The `reference` arm is the committed tree at `49b4e25` (release 1.11.0).

## Run

- Finished: 2026-09-11 19:03 UTC
- Runner: `claude` 2.1.267, `--include-hook-events`
- Model: `claude-opus-5`, reasoning effort `medium`
- Seed: `1`; `-j 3`
- Corpus: `implement`
- Fixtures: `catalog`, `feed`, `gateway`
- Arms: `no-skill`, `reference`, `baseline`
- Repetitions: 1 per fixture and arm, 9 sessions total
- `reference`: a `git worktree` of `49b4e25`, plugin SHA-256
  `6b2a182086e1397b964432e2fccd22263e0ea5e35c67660dc1bc086fc31af39b`
- `baseline`: the working tree at `49b4e25` plus the uncommitted edits above,
  plugin SHA-256 `34b4f1418237eb5039f39f25d0f10be8d3cd35c8252a5652f27b3256d04a8113`
- Toolchain: Go 1.27.1 linux/amd64
- Raw report: [`2026-09-11-go-implement-load-cuts-smoke-opus-5-medium.json`](2026-09-11-go-implement-load-cuts-smoke-opus-5-medium.json)
  (SHA-256 `0a58ffb06c49972a43dc832448db9c1527f29e34153d59ae7bb4104ca8906847`)
- Session transcripts: [`2026-09-11-go-implement-load-cuts-smoke-opus-5-medium.traces.tar.gz`](2026-09-11-go-implement-load-cuts-smoke-opus-5-medium.traces.tar.gz)
  (SHA-256 `0024b016e5aa1226dc4c980e8db5b8513ec8e3c5dbce14d6470f7b444cacd252`), one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $0.5252 control, $2.3652 reference, $2.0805 baseline — **4.50x** and **3.96x** the control

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-opus-5 -effort medium \
  -reference-root <worktree of 49b4e25> -arms no-skill,reference,baseline \
  -tasks catalog,feed,gateway -n 1 -j 3 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-11-go-implement-load-cuts-smoke-opus-5-medium.json
```

All 9 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. The tool set had no shell.

`go-code` fired in 3/3 reference and 3/3 baseline sessions, as the first tool
call in every one of them, after the prompt hook's note.

## Results

| Fixture | Arm | Golden | Δlines | Δtypes | Δfuncs | `Skill` turns | API calls | $ / run |
|---|---|---|---:|---:|---:|---:|---:|---:|
| `catalog` | no-skill | 1/1 | 30 | 0 | 0 | — | 7 | 0.108 |
| `catalog` | reference | 1/1 | 21 | 0 | 0 | 2 | 7 | 0.649 |
| `catalog` | baseline | 1/1 | 19 | 0 | 0 | 2 | 7 | 0.513 |
| `feed` | no-skill | 1/1 | 50 | 2 | 0 | — | 7 | 0.126 |
| `feed` | reference | 1/1 | 34 | 0 | 0 | 2 | 7 | 0.644 |
| `feed` | baseline | 1/1 | 34 | 0 | 0 | 2 | 7 | 0.610 |
| `gateway` | no-skill | **0/1** | 134 | 0 | 6 | — | 8 | 0.291 |
| `gateway` | reference | 1/1 | 80 | 0 | 0 | 6 | 16 | 1.072 |
| `gateway` | baseline | 1/1 | 76 | 0 | 0 | 2 | 10 | 0.957 |

| Arm | golden | `Skill` turns | API calls | cache writes | cache reads | output tokens | thinking tokens | median report chars | $ / run |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|
| `no-skill` | 2/3 | 0.0 | 7.3 | 5.5K | 49K | 3771 | 1143 | 1311 | 0.175 |
| `reference` | 3/3 | 3.3 | 10.0 | 42.2K | 306K | 8486 | 3592 | 1219 | 0.788 |
| `baseline` | 3/3 | 2.0 | 8.0 | 37.9K | 215K | 8260 | 3506 | 1316 | 0.693 |

Skill loads — reference: `catalog`: `go-code`, `go-style-core`, `go-error-handling`, `go-testing`, `go-data-structures` in 2 messages; `feed`: `go-code`, `go-style-core`, `go-testing`, `go-data-structures`, `go-error-handling`, `go-defensive` in 2 messages; `gateway`: `go-code`, `go-style-core`, `go-http`, `go-testing`, `go-defensive`, `go-error-handling` in 6 messages.
Baseline: `catalog`: `go-code`, `go-style-core`, `go-error-handling`, `go-testing`, `go-data-structures` in 2 messages; `feed`: `go-code`, `go-style-core`, `go-testing`, `go-data-structures`, `go-error-handling`, `go-defensive` in 2 messages; `gateway`: `go-code`, `go-style-core`, `go-http`, `go-error-handling`, `go-testing`, `go-data-structures` in 2 messages.

## Reading

- **Correctness: 3/3 in both skilled arms against 2/3 unaided.** The control
  `gateway` session answered `HEAD` with 200 on every route, at 134 lines with
  six helpers — the trap this fixture exists to catch, unchanged from the
  2026-09-11 workflow run.
- **The batching edit acted.** Every baseline session loaded `go-code` first
  and every owner in one message after it. The reference `gateway` session
  loaded its six owners one per turn: 16 API calls against 10, cache reads
  562K tokens against 294K, $1.07 against $0.96. On `catalog` and `feed` the
  reference tree already batched, so the arms are level there.
- **Size is level:** `catalog` 21 → 19, `feed` 34 → 34, `gateway` 80 → 76.
  One repetition; no claim.
- **Reports** carry the `checks:` line and the budget line in 3/3 sessions of
  both skilled arms. Median report length 1316 against 1219 characters.
- **Per-skill payloads are not measured here**: no baseline session loaded a
  skill on a turn of its own, which is the point of the edit. The
  [`feed` n=5 run](2026-09-11-go-implement-load-cuts-feed-n5-sonnet-5-medium.md)
  measures them.

## What this changes

- The 1.12.0 edits ship. Nothing moved against them on this model, and the
  one mechanism they target is visible in the `gateway` pair.
- The README table is unchanged: n=1 does not displace the n=3 Opus 5 cell.
