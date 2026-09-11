# Load Cuts Smoke — `catalog`, `feed`, `gateway`, three arms, Sonnet 5 medium (n=1)

The same tree pair as the [Opus 5 smoke](2026-09-11-go-implement-load-cuts-smoke-opus-5-medium.md)
of the same hour, on Sonnet 5 at medium effort. One repetition per fixture and
arm supports no claim about size, correctness rate, or cost; the run records
whether the owners load in fewer messages and what the one failure was.

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

- Finished: 2026-09-11 19:01 UTC
- Runner: `claude` 2.1.267, `--include-hook-events`
- Model: `claude-sonnet-5`, reasoning effort `medium`
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
- Raw report: [`2026-09-11-go-implement-load-cuts-smoke-sonnet-5-medium.json`](2026-09-11-go-implement-load-cuts-smoke-sonnet-5-medium.json)
  (SHA-256 `802b93d7ad9d835acfabb4bee066e01ce6f5ac9af5fbfbb05c3c18dd29744f85`)
- Session transcripts: [`2026-09-11-go-implement-load-cuts-smoke-sonnet-5-medium.traces.tar.gz`](2026-09-11-go-implement-load-cuts-smoke-sonnet-5-medium.traces.tar.gz)
  (SHA-256 `604286c7d80729fac608e3779ef4497fe912e3db9663c9fda544b67c1c2823a9`), one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $0.1373 control, $0.9720 reference, $0.8911 baseline — **7.08x** and **6.49x** the control

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <worktree of 49b4e25> -arms no-skill,reference,baseline \
  -tasks catalog,feed,gateway -n 1 -j 3 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-11-go-implement-load-cuts-smoke-sonnet-5-medium.json
```

All 9 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. The tool set had no shell.

`go-code` fired in 3/3 reference and 3/3 baseline sessions, as the first tool
call in every one of them.

## Results

| Fixture | Arm | Golden | Δlines | Δtypes | Δfuncs | `Skill` turns | API calls | $ / run |
|---|---|---|---:|---:|---:|---:|---:|---:|
| `catalog` | no-skill | 1/1 | 22 | 0 | 0 | — | 5 | 0.033 |
| `catalog` | reference | 1/1 | 19 | 0 | 0 | 3 | 14 | 0.296 |
| `catalog` | baseline | 1/1 | 19 | 0 | 0 | 3 | 13 | 0.257 |
| `feed` | no-skill | 1/1 | 48 | 0 | 0 | — | 6 | 0.042 |
| `feed` | reference | 1/1 | 35 | 0 | 0 | 4 | 13 | 0.298 |
| `feed` | baseline | **0/1** | 39 | 0 | 0 | 2 | 8 | 0.210 |
| `gateway` | no-skill | 1/1 | 84 | 0 | 0 | — | 7 | 0.063 |
| `gateway` | reference | 1/1 | 80 | 0 | 1 | 3 | 13 | 0.377 |
| `gateway` | baseline | 1/1 | 73 | 0 | 0 | 2 | 15 | 0.424 |

| Arm | golden | `Skill` turns | API calls | cache writes | cache reads | output tokens | thinking tokens | median report chars | $ / run |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|
| `no-skill` | 3/3 | 0.0 | 6.0 | 4.4K | 57K | 1587 | 385 | 270 | 0.046 |
| `reference` | 3/3 | 3.3 | 13.3 | 41.2K | 405K | 7729 | 2471 | 836 | 0.324 |
| `baseline` | 2/3 | 2.3 | 12.0 | 37.7K | 388K | 6740 | 2413 | 894 | 0.297 |

Skill loads — reference: `catalog`: `go-code`, `go-style-core`, `go-error-handling`, `go-testing`, `go-linting` in 3 messages; `feed`: `go-code`, `go-style-core`, `go-testing`, `go-error-handling`, `go-linting` in 4 messages; `gateway`: `go-code`, `go-style-core`, `go-http`, `go-testing`, `go-error-handling` in 3 messages.
Baseline: `catalog`: `go-code`, `go-style-core`, `go-error-handling`, `go-testing`, `go-linting` in 3 messages; `feed`: `go-code`, `go-style-core`, `go-data-structures`, `go-error-handling`, `go-testing` in 2 messages; `gateway`: `go-code`, `go-style-core`, `go-http`, `go-security`, `go-testing`, `go-error-handling` in 2 messages.

Payloads written to the cache on a lone `Skill` load, tokens: `go-code` 6633
reference against 6554 baseline; `go-linting` 5176 against 5115;
`go-error-handling` 3088 in the reference arm, batched in the baseline arm.

## The `feed` failure

The baseline `feed` session rendered `kinds` as `null` for an account with no
events, and the golden `TestRenderEmptyKeepsEveryMemberTyped` failed on it.
Its own contract test, written before the body, carried the case — "no
activity is an ordinary response, not a special case", want
`"kinds":[]` — and the body pre-sized `events` and `counts` but left `kinds`
nil and appended to it. The session's sequence: `Skill go-code` with a `Glob`
in the first message, `Read`, four owners in one `Skill` message, `Write` of
the contract test, two `Edit`s of the body, a `Read` of the result, and the
text "complete and correct" before the report, with no shell to run the test
it had written. The same miss appears in the
[2026-09-10 control](2026-09-10-go-implement-newcode-control-sonnet-5-medium.md)
and the [2026-09-11 Plain Code run](2026-09-11-go-implement-newcode-plaincode-sonnet-5-medium.md),
baseline `feed` 4/5 in each, with the go-style-core sentence on nil and empty
values still in the tree. The
[`feed` n=5 run](2026-09-11-go-implement-load-cuts-feed-n5-sonnet-5-medium.md)
below repeats the fixture at five repetitions per arm.

## Reading

- **`Skill` turns 2.3 against 3.3** per session, the first one always
  `go-code`; Sonnet 5 still spreads owners over two or three messages where
  Opus 5 uses one.
- **Cost $0.297 against $0.324 a session.** Without the failed `feed`
  session, `catalog` is −13% and `gateway` +12%, the latter loading
  `go-security` as a sixth owner and making 15 API calls. Noise at n=1.
- **`go-linting` loaded with no shell** in 2/3 reference and 1/3 baseline
  sessions. Opus 5 stopped doing this in 1.11.0; Sonnet 5 has not, and the
  edits here do not address it.
- **Size:** `catalog` 19 → 19, `feed` 35 → 39 (the failed session),
  `gateway` 80 → 73.

## What this changes

- One failure at n=1 on a fixture with a recorded 1-in-5 miss rate is not a
  regression; it is the reason the `feed` n=5 run was made before release.
- The README table is unchanged: n=1 does not displace the n=5 Sonnet 5 cell.
