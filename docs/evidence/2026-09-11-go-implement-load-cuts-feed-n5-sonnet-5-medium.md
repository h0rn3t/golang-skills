# Load Cuts — `feed`, reference against baseline, Sonnet 5 medium (n=5)

The follow-up to the [Sonnet 5 smoke](2026-09-11-go-implement-load-cuts-smoke-sonnet-5-medium.md),
whose one failure was a baseline `feed` session rendering `kinds` as `null`:
the fixture repeated at five repetitions per arm, reference against baseline,
no unaided arm. Two questions: does the miss recur more often in the tree
without the go-style-core sentence on nil and empty values, and what do the
loading cuts measure on this fixture.

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

- Finished: 2026-09-11 21:20 UTC
- Runner: `claude` 2.1.267, `--include-hook-events`
- Model: `claude-sonnet-5`, reasoning effort `medium`
- Seed: `1`; `-j 4`
- Corpus: `implement`
- Fixtures: `feed`
- Arms: `reference`, `baseline`
- Repetitions: 5 per fixture and arm, 10 sessions total
- `reference`: a `git worktree` of `49b4e25`, plugin SHA-256
  `6b2a182086e1397b964432e2fccd22263e0ea5e35c67660dc1bc086fc31af39b`
- `baseline`: the working tree at `49b4e25` plus the uncommitted edits above,
  plugin SHA-256 `34b4f1418237eb5039f39f25d0f10be8d3cd35c8252a5652f27b3256d04a8113`
- Toolchain: Go 1.27.1 linux/amd64
- Raw report: [`2026-09-11-go-implement-load-cuts-feed-n5-sonnet-5-medium.json`](2026-09-11-go-implement-load-cuts-feed-n5-sonnet-5-medium.json)
  (SHA-256 `5406f7de9394fc31a24143c5eb07f44a5140a9e96a512fd17f756a03a80fdff8`)
- Session transcripts: [`2026-09-11-go-implement-load-cuts-feed-n5-sonnet-5-medium.traces.tar.gz`](2026-09-11-go-implement-load-cuts-feed-n5-sonnet-5-medium.traces.tar.gz)
  (SHA-256 `d6c0bd4f8a9fc895be60ffcd481d9c6a33008d42b3b4151bd5cc79698ec6daf2`), one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $1.4971 reference, $1.3749 baseline — **-8%**

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <worktree of 49b4e25> -arms reference,baseline \
  -tasks feed -n 5 -j 4 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-11-go-implement-load-cuts-feed-n5-sonnet-5-medium.json
```

All 10 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. The tool set had no shell.

`go-code` fired in 5/5 sessions of each arm, as the first tool call.

## Results

| Arm | golden | model tests | Δlines by rep | `Skill` turns by rep | API calls by rep | $ / run by rep |
|---|---|---|---|---|---|---|
| `reference` | 5/5 | 5/5 | 30, 39, 32, 33, 30 → 32.8 | 4, 4, 5, 4, 5 | 13, 15, 18, 16, 14 | 0.276, 0.311, 0.308, 0.283, 0.319 → 0.299 |
| `baseline` | 5/5 | 5/5 | 31, 34, 38, 30, 38 → 34.2 | 5, 4, 4, 3, 3 | 14, 14, 16, 13, 14 | 0.298, 0.250, 0.349, 0.242, 0.236 → 0.275 |

Δtypes and Δfuncs are 0 in all ten sessions; every session wrote a contract
test and reported the `checks:` and budget lines.

| Arm | golden | `Skill` turns | API calls | cache writes | cache reads | output tokens | thinking tokens | median report chars | $ / run |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|
| `reference` | 5/5 | 4.4 | 15.2 | 40.9K | 431K | 4846 | 505 | 800 | 0.299 |
| `baseline` | 5/5 | 3.8 | 14.2 | 35.9K | 398K | 5067 | 880 | 918 | 0.275 |

Payloads written to the cache on a lone `Skill` load, tokens (reference →
baseline): `go-style-core` 2867 → 2329–2340; `go-code` 6601–6635 → 6543–6588;
`go-error-handling` 3089 → 2988–3000; `go-linting` 5157–5161 → 5087–5119;
`go-testing` 3647 in one reference session, batched in every baseline
session. Summed over the skills a `feed` session loads, about 750 tokens less
per session; the changelog estimate from character counts was 1.1K.

`go-linting` loaded with no shell in 5/5 reference and 4/5 baseline sessions.

## Reading

- **The smoke failure did not recur: golden 5/5 in both arms, `null` in
  0 of 10 sessions.** The miss sits at its recorded rate — one session in
  five in the 2026-09-10 control and the 2026-09-11 Plain Code run, one in
  three in the smoke, none in ten here — and does not separate the trees.
- **Size: 34.2 against 32.8 lines**, both arms spread over 30–39, no
  helpers or types added in either. Not a signal at n=5.
- **Cost −8%, cache writes −12%** ($0.275 against $0.299, 35.9K against
  40.9K tokens a session) with `Skill` turns 3.8 against 4.4. This is the
  loading cut in numbers on a fixture that loads four to six skills.
- **Thinking and reports moved the other way** — 880 against 505 thinking
  tokens, median report 918 against 800 characters — inside the spread of
  single sessions (thinking 199–1336 across the run).

## What this changes

- The 1.12.0 edits ship: the one open question from the smokes is closed on
  the fixture and model that raised it.
- The README table is unchanged: this run has no unaided arm, so it sets no
  cell; the README's Loading cost paragraph cites it.
