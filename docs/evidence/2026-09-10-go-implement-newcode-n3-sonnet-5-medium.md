# New-Code Sections — `gateway` and `feed`, reference against baseline, Sonnet 5 medium (n=3)

The second reading of the new-code edits, on the two implementation fixtures
whose traps are live on this model at this effort — `gateway` (HEAD is a 405;
the empty list is `[]`) and `feed` (every member keeps its JSON type). Three
repetitions per fixture and arm, the plugin the only difference between arms.
Three per cell is a direction, not a control: it decides whether the full
corpus at n=5 is worth running, and it was.

What changed between the [smoke](2026-09-10-go-implement-newcode-smoke-sonnet-5-medium.md)
and this run, all in the `baseline` arm:

- The routing hook no longer names `go-data-structures` for `make([]`,
  `make(map` and `append(` — the forced load that started the
  `make`+`copy` → `slices.Clone` → `null` chain in the control and the smoke.
- The `go-data-structures` Copy row is two rows: `Clone` where nil may stay
  nil, and `make`+`copy` or `append([]T{}, s...)` where the copy must encode
  as `[]`; the `go-defensive` example shows both forms as code.
- The `go-code` Contract Table is a file, not prose: a table-driven
  `<pkg>_contract_test.go` written before the first production edit, run with
  the gate where a shell exists and read against the code where it does not.

## Run

- Finished: 2026-09-10 21:09 UTC
- Runner: `claude` 2.1.267
- Model: `claude-sonnet-5`, reasoning effort `medium`
- Seed: `1`
- Corpus: `implement`, `-tasks gateway,feed`
- Arms: `reference`, `baseline`
- Repetitions: 3 per fixture and arm, 12 sessions total
- `reference`: `git archive 7c89b4f` (skills byte-identical to release 1.7.0),
  plugin SHA-256
  `5028c91b61340fe99a22284cf0fe4ba32f48d8e7b1cf83baa0e412c88708a2ea` — the
  digest of the skilled arm in the
  [2026-09-10 n=5 control](2026-09-10-go-implement-control-sonnet-5-medium.md)
- `baseline`: working tree at `7c89b4f` plus the uncommitted edits to
  `hooks/go-code-routing.sh`, `skills/go-code`, `skills/go-http`,
  `skills/go-data-structures` and `skills/go-defensive`, plugin SHA-256
  `7e355813e433303a38e12699e7dd3d66809da2fba126b4a09cef32bd2f2d7f70`
- Toolchain: Go 1.27.1 linux/amd64
- Raw report: [`2026-09-10-go-implement-newcode-n3-sonnet-5-medium.json`](2026-09-10-go-implement-newcode-n3-sonnet-5-medium.json)
  (SHA-256 `fe5ff30a4914948a73d6a64bb38c63c17cecb6bbe9fa6ece53f5dfea1a3020cb`)
- Session transcripts:
  [`2026-09-10-go-implement-newcode-n3-sonnet-5-medium.traces.tar.gz`](2026-09-10-go-implement-newcode-n3-sonnet-5-medium.traces.tar.gz)
  (SHA-256 `de11e7975171e758558b1b7e67938d5197965a90f4b9abb3e64d93e3d1dfe80e`),
  one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $1.4162 reference, $2.0239 baseline — $0.2360 against $0.3373 per
  session, **+43%**

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <git archive 7c89b4f> -arms reference,baseline -tasks gateway,feed \
  -n 3 -j 4 -seed 1 -keep \
  -out ../docs/evidence/2026-09-10-go-implement-newcode-n3-sonnet-5-medium.json
```

All 12 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. No shell in either arm; `go-code` fired in
12/12.

## Results

| Fixture | Arm | Golden | Model tests | Δlines | Δfuncs | Δtypes | $ / run |
|---|---|---|---|---|---|---|---:|
| `gateway` | reference | **0/3** — `null`; HEAD+`null`; HEAD+`null` | none written | 103, 73, 75 | 4, 1, 1 | 1, 0, 0 | 0.289 |
| `gateway` | baseline | **3/3** | written 3/3, pass 3/3 | 83, 81, 93 | 2, 1, 5 | 0, 0, 1 | 0.443 |
| `feed` | reference | 3/3 | none written | 44, 48, 41 | 0, 0, 0 | 0, 2, 0 | 0.183 |
| `feed` | baseline | 3/3 | written 3/3, pass 3/3 | 45, 47, 45 | 0, 0, 0 | 0, 2, 0 | 0.232 |

| Arm | golden | skills / session | hook blocks / session | turns | test file before first production edit |
|---|---|---:|---:|---:|---|
| `reference` | 3/6 | 5.00 | 1.33 | 24.8 | — |
| `baseline` | 6/6 | 5.67 | 1.50 | 30.3 | 3/6 (`feed` 3/3, `gateway` 0/3) |

Skill loads over six sessions — reference: `go-code` 6, `go-style-core` 6,
`go-error-handling` 6, `go-data-structures` 6, `go-http` 3, `go-linting` 2,
`go-defensive` 1. Baseline: `go-code` 6, `go-style-core` 6,
`go-error-handling` 6, `go-testing` 6, `go-http` 3, `go-linting` 3,
`go-defensive` 3, `go-documentation` 1, `go-data-structures` **0**.

## `gateway`: 0/3 against 3/3

The reference arm fails the way the 1.7.0 tree has failed every skilled
`gateway` session at this effort: HEAD served as 200 by the `GET` patterns in
two sessions, the empty list rendered as `null` in all three (the
`make`+`copy` → `slices.Clone` rewrite after the forced `go-data-structures`
load, as in the control). Across the 2026-09-09 and 2026-09-10 controls, the
smoke's reference session and this run, that tree is now **0 of 14** on the
HEAD assertions.

The baseline arm passes 3/3. Every session registers `HEAD /healthz`,
`HEAD /accounts` and `HEAD /accounts/{id}` beside the `GET` patterns with a
405 handler carrying `Allow: GET` — the shape of the `go-http` Routing example
— and none returns `null`: two sessions keep the list non-nil from
construction, and the third wrote `slices.Clone` first and replaced it with
`make`+`copy` six edits later, before writing its contract test.
`go-data-structures` was never loaded in this arm. With the smoke's baseline
session, the new tree is 4 of 4 on the HEAD assertions and 3 of 4 on the
empty list — the miss being the smoke session, whose tree still carried the
collections hint. A Fisher exact test on 0/3 against 3/3 gives p = 0.10
two-sided; the history behind the 0 is what makes the direction worth n=5.

Cost of the fix on this fixture: +53% per session (0.289 → 0.443), 32–44
turns against 26–29. The contract tests are 112–193 lines each; two of the
three test HEAD explicitly and all three test the empty list. Functions added:
2, 1, 5 against 4, 1, 1 — the 5 is a session that declared a server struct
with three methods plus `writeJSON` and `methodNotAllowed`, and reported them
under the Declaration Budget as "single-purpose helpers reused across 3 routes
each"; the budget made the growth visible and reasoned, it did not stop it.

## `feed`: 3/3 against 3/3

Correctness was never in question here — the doc comment names the trap — and
size is level: 45.7 against 44.3 lines, two package-level types in one session
of each arm. What the arm adds is the file: three contract tests of four cases
each, named after the clauses (`no activity is an ordinary response, not a
special case`, `an event with an empty ID never happened`), written **before**
the first production edit in 3 of 3 sessions, passing in 3 of 3. Cost +27%.

## What the model followed and what it did not

- **Contract test written: 6/6; passing: 6/6.** The prose form of the same
  instruction produced a post-hoc list in 4/4 smoke sessions and no test
  file; the file form produced a test in every session. On `feed` it came
  first every time; on `gateway` the model wrote the server body first and
  the test after it in all three sessions, so on the larger task the
  "before the first production edit" clause was not followed, and the test
  still caught nothing it would have needed to catch — the implementation
  was already right. Whether a test written second can still move
  correctness is the n=5 question.
- **Declaration Budget reported: 6/6**, with reasons. Two of three `gateway`
  sessions justify exactly two helpers by reuse across three routes; the
  third justifies six declarations the same way.
- **Hook.** Removing the collections hint took `go-data-structures` from 6/6
  to 0/6 sessions with no visible loss — the Copy rows were not read and were
  not needed. The contract test write is itself blocked once for `go-testing`,
  and on `gateway` also for `go-defensive`, because the test body carries a
  `defer`. That second load is noise from a test file and is the next cost to
  remove.
- **Cost +43%** per session: the test file is 80–190 lines of output, and the
  test write pulls one or two more skill loads.

## What this changes

- Nothing in the READMEs yet: the full corpus at n=5, reference against
  baseline, is the run that can set a cell, and it follows this one.
- The `go-http` Routing example and the hook change are the two edits with a
  mechanism visible in the traces; the Copy rows were not exercised in this
  arm and remain unmeasured on their own.
