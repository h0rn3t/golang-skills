# Prompt Routing — `catalog` and `feed`, three arms, Opus 5 medium (n=3)

The run that measures the two routing edits on the model and fixtures where
the gap was recorded. The [Opus 5 workflow run](2026-09-11-go-implement-newcode-workflow-opus-5-medium.md)
left one open item: 5 of 18 skilled sessions loaded no skill at all, every
one on `feed` or `catalog`, whose prompt says only "Implement the Go package
in ./<dir>". The [Sonnet 5 smoke](2026-09-11-go-implement-prompt-routing-smoke-sonnet-5-medium.md)
established that the new hook fires under the harness; on that model the
router was already reaching every session, so it could not measure the gap.
This run can: `reference` is the committed tree at `f4f373f`, whose `go-code`
description and hook set are those of the workflow run.

The edits in the `baseline` arm:

- **`go-code` description** also names implementing a Go package, function,
  or handler whose declarations and documentation already exist — write the
  bodies, fill in a stub, replace `panic("not implemented")` — even when the
  request names only the package.
- **`hooks/go-prompt-routing.sh`**, a `UserPromptSubmit` hook. When a prompt
  asks for Go work it adds one note to the model's context naming the router
  to load before the first edit: `go-code-refactor` for refactor, clean-up,
  or simplify wording, `go-code` for anything else. Once per skill per
  session; never blocks.

The two are confounded: a difference between the skilled arms is the pair.
The traces say which one acted, see below.

## Run

- Finished: 2026-09-11 11:11 UTC
- Runner: `claude` 2.1.267, with `--include-hook-events` (new in `abrun`
  this day, so hook lifecycle events appear in the trace)
- Model: `claude-opus-5`, reasoning effort `medium`
- Seed: `1`; `-j 4`
- Corpus: `implement`, `-tasks catalog,feed`
- Arms: `no-skill`, `reference`, `baseline`
- Repetitions: 3 per fixture and arm, 18 sessions total
- `reference`: a `git worktree` of `f4f373f`, plugin SHA-256
  `ec904c52ac70fb646eaafe3f271548cf129279a8628fd53ab58ff36dc6542bf1`
- `baseline`: the working tree at `f4f373f` plus the uncommitted edits above,
  plugin SHA-256
  `8c7bc9f028080f74648d33156baf5c22512bb85e23150392407bd48c8bb0d5ad`
- Toolchain: Go 1.27.1 linux/amd64
- Raw report: [`2026-09-11-go-implement-prompt-routing-opus-5-medium.json`](2026-09-11-go-implement-prompt-routing-opus-5-medium.json)
  (SHA-256 `e7d245f8d92ab606b8c5f5184a22eae1df0f2e24dcefbba1e8cf322eb8a0a662`)
- Session transcripts:
  [`2026-09-11-go-implement-prompt-routing-opus-5-medium.traces.tar.gz`](2026-09-11-go-implement-prompt-routing-opus-5-medium.traces.tar.gz)
  (SHA-256 `7a35a37e829a679866fcaa95fd085be6070d6dfbb21efae89bc1d767713e8ed5`),
  one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $0.7103 control, $3.0065 reference, $3.8620 baseline — **4.23x** and
  **5.44x** the control

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-opus-5 -effort medium \
  -reference-root <worktree of f4f373f> -arms no-skill,reference,baseline \
  -tasks catalog,feed -n 3 -j 4 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-11-go-implement-prompt-routing-opus-5-medium.json
```

All 18 sessions completed without a CLI error, changed the fixture, and
stayed out of the repository checkout. No shell in any arm.

## Results

| Fixture | Arm | Golden | Δlines | Δtypes | skill loaded | `go-code` loaded | $ / run |
|---|---|---|---|---|---|---|---:|
| `catalog` | no-skill | 3/3 | 24, 27, 24 → 25.0 | 0 | — | — | 0.084 |
| `catalog` | reference | 3/3 | 22, 19, 19 → 20.0 | 0 | **1/3** | 1/3 | 0.312 |
| `catalog` | baseline | 3/3 | 19, 19, 19 → **19.0** | 0 | **3/3** | 3/3 | 0.681 |
| `feed` | no-skill | 3/3 | 50, 47, 52 → 49.7 | 2 ×3 | — | — | 0.152 |
| `feed` | reference | 3/3 | 36, 33, 36 → 35.0 | 0 | 3/3 | 3/3 | 0.690 |
| `feed` | baseline | 3/3 | 34, 34, 36 → 34.7 | 0 | 3/3 | 3/3 | 0.607 |

| Arm | golden | prompt hook fired | first tool call | turn of the `go-code` load | `checks:` line | budget line | contract test written | skills / session | turns | median report chars | $ / run |
|---|---|---|---|---|---|---|---|---:|---:|---:|---:|
| `no-skill` | 6/6 | — | `Glob` 6/6 | — | 0/6 | 0/6 | 2/6 | 0 | 12.8 | 856 | 0.1184 |
| `reference` | 6/6 | — | `Glob` 6/6 | 6th–7th in 4; never in 2 | 4/6 | 4/6 | 4/6 | 3.83 | 18.8 | 925 | 0.5011 |
| `baseline` | 6/6 | **6/6** | **`Skill go-code` 6/6** | **2nd in 6/6** | **6/6** | **6/6** | **6/6** | 5.50 | 22.0 | 1020 | 0.6437 |

Skill loads over six sessions — reference: `go-code` 4, `go-style-core` 4,
`go-error-handling` 4, `go-testing` 4, `go-data-structures` 4,
`go-defensive` 2. Baseline: `go-code` 6, `go-style-core` 6,
`go-error-handling` 6, `go-testing` 6, `go-data-structures` 6,
`go-defensive` 1. `go-linting` was not loaded in any session of either arm.

## The gap, reproduced and closed

**Reference: two of three `catalog` sessions loaded nothing.** Both ran
`Glob`, read the stub and `go.mod`, edited the file twice, and reported in
prose with no `checks:` line, no budget line, no test file — the shape the
workflow run recorded 5 times in 18. Their bodies are 22 and 19 lines and
both pass the golden, so on this fixture the unaided Opus 5 body is already
close to the skilled one; what the router adds here is the contract test and
the report, not correctness. The third `catalog` session and all three `feed`
sessions loaded `go-code` at the sixth or seventh turn, after reading the
files. Across the two Opus 5 runs the reference description therefore misses
7 of 24 skilled sessions, 6 of them on `catalog`.

**Baseline: `go-code` in 6/6, as the first tool call, at turn 2.** Every
transcript opens with the hook's `hook_response` event carrying the note,
then the model's first turn is `Skill go-code`, with no `Glob` before it.
Fisher on skill loaded, 4/6 against 6/6, is p = 0.45 at this n; the
observation that carries is the sequence, which is the same in six of six
sessions and in none of the six reference sessions.

**Which edit acted.** The hook's note is in the context in every baseline
session and the router is loaded in the turn after it, before the model has
read a file. The description can only act through the host's matcher once
the model has decided to look for a skill; in the reference arm that decision
came at turn 6 or 7 or not at all. This run cannot separate the two edits by
outcome — both arms with the hook route 6/6 — but the timing says the hook is
what moved the first call, and the two reference `catalog` misses are the
cases the description alone has to be trusted to catch. A run with the hook
removed from `baseline` would isolate the description; it was not made.

## What followed from the load

- **Every baseline session followed the workflow**: `checks:` line, budget
  line, no-shell statement and a contract test in 6/6, against 4/6 in the
  reference arm — exactly the sessions that loaded the router. The workflow
  edits already measured on this model hold; the hook delivers the sessions
  that were not reaching them.
- **Correctness is saturated**: 6/6 in every arm. `catalog` and `feed` on
  Opus 5 are passed unaided; this run says nothing about correctness.
- **Size.** `catalog` 19.0 in every baseline session against 20.0 and 25.0;
  `feed` 34.7 against 35.0 and 49.7, the unaided arm declaring two
  package-level types in every session where neither skilled arm declared
  any. The skilled arms are level with each other; the control gap is the
  Plain Code form, measured before.
- **Cost: 5.44x against 4.23x.** The whole difference is the two `catalog`
  sessions that loaded nothing and cost $0.16 and $0.21 against $0.54–$0.86
  for a session that runs the workflow. Routing every session costs what the
  workflow costs; the reference arm was cheaper by skipping it.
- **Turns 22.0 against 18.8**; report length 1020 against 925 characters,
  the shape being the four-line report plus the test file's name.

## What this changes

- **The hook closes the Opus 5 routing gap on the fixtures that showed it**:
  6/6 against 4/6 here and 13/18 in the workflow run, with the router loaded
  as the first action in every session. Kept.
- **The description edit is kept but not isolated**: both edits ride in the
  same arm. Its measurement would be a `baseline` without the hook; the
  hook's presence makes the question cheaper to leave open.
- **README**: the Opus 5 new-code cell stays on the workflow run. This run is
  two fixtures, n=3, and correctness is 6/6 in every arm.
- **Next**: the refactor corpus on Haiku 4.5, where the router reached 1 of 4
  sessions, run the same day; `gateway` on Opus 5 with the hook, since it is
  the one fixture where routing changes the golden on this model.
