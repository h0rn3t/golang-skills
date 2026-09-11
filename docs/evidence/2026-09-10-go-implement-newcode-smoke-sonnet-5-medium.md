# New-Code Sections Smoke — Sonnet 5 at medium effort, implementation corpus (n=1)

A first look at two edits made on the same day, against the tree they change
and against no skills at all. Both arms with skills fired the router in every
session, so this run reads *which instructions the model followed*, not what
their effect is: one repetition per fixture and arm supports no claim about
size, correctness rate, or cost, and none is made below.

The edits under test:

- `go-code` Writing New Code gains a **Contract Table** (one row per observable
  clause of the documentation, written before the first edit and walked before
  closing) and a **Declaration Budget** (every declaration added beyond the
  specification carries one of three reasons and is reported as a count).
- Caveats moved into the examples that primed the defects the
  [2026-09-10 control](2026-09-10-go-implement-control-sonnet-5-medium.md)
  recorded: the `go-http` Routing example registers a `HEAD` pattern beside its
  `GET` pattern; the `go-data-structures` Copy row and the `go-defensive`
  boundary-copy example say in the cell and in the code comment that `Clone`
  keeps nil, which `encoding/json` v1 writes as `null`.

## Run

- Finished: 2026-09-10 20:46 UTC
- Runner: `claude` 2.1.267
- Model: `claude-sonnet-5`, reasoning effort `medium`
- Seed: `1`
- Corpus: `implement`
- Fixtures: `catalog`, `feed`, `gateway`, `ledger`
- Arms: `no-skill`, `reference`, `baseline`
- Repetitions: 1 per fixture and arm, 12 sessions total
- `reference`: `git archive 7c89b4f` (skills byte-identical to release 1.7.0),
  plugin SHA-256
  `5028c91b61340fe99a22284cf0fe4ba32f48d8e7b1cf83baa0e412c88708a2ea` — the
  same digest as the `baseline` arm of the 2026-09-10 n=5 control, so that
  control's twenty skilled sessions are this arm's history
- `baseline`: working tree at `7c89b4f` plus the uncommitted edits to
  `skills/go-code`, `skills/go-http`, `skills/go-data-structures` and
  `skills/go-defensive`, plugin SHA-256
  `f624ec953df18cc52adbfa2f1b6514ed647980d32f78bba14576f144768124f4`
- Toolchain: Go 1.27.1 linux/amd64
- Raw report: [`2026-09-10-go-implement-newcode-smoke-sonnet-5-medium.json`](2026-09-10-go-implement-newcode-smoke-sonnet-5-medium.json)
  (SHA-256 `e1c57ade5aa6d2ae16119437902458e816f1ab17e3b83efb212c500b6edf2b58`)
- Session transcripts:
  [`2026-09-10-go-implement-newcode-smoke-sonnet-5-medium.traces.tar.gz`](2026-09-10-go-implement-newcode-smoke-sonnet-5-medium.traces.tar.gz)
  (SHA-256 `d8027a531176fb4df9a5448284330d957845d1d6c5aa92e403e5faf3cfb32d0d`),
  one `traces/<arm>-<fixture>-r0.jsonl` per session
- Cost: $0.1679 control, $0.8263 reference, $0.9712 baseline — **4.92x** and
  **5.79x** the control

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <git archive 7c89b4f> -arms no-skill,reference,baseline \
  -n 1 -j 4 -seed 1 -keep \
  -out ../docs/evidence/2026-09-10-go-implement-newcode-smoke-sonnet-5-medium.json
```

All 12 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. The tool set was `Skill,Read,Glob,Grep,Edit,Write`,
so no session had a shell; the routing hook was active in both skilled arms.

## Results

Production-line delta against the shipped fixture, declared functions and
types added, and the hidden golden verdict. One session per cell.

| Fixture | No skill | Reference (1.7.0) | Baseline (edits) |
|---|---|---|---|
| `catalog` | pass, +20, 0 funcs | pass, +20, 0 funcs | pass, +21, 0 funcs; budget reported: 0 |
| `feed` | pass, +48 | pass, +43, 0 types | pass, +44, **+2 types**; budget reported: 2 |
| `gateway` | pass, +78, 0 funcs (`r.Method` checks) | **fail** `HEAD … = 200, want 405`, +90, +4 funcs | **fail** `GET /accounts = 200 "null", want []`, +79, +2 funcs; HEAD assertions pass |
| `ledger` | pass, +35 | pass, +45, +1 type | pass, +39, 0 types; budget reported: 0 |

| Arm | golden | skills / session | hook blocks / session | turns | $ / run |
|---|---|---:|---:|---:|---:|
| `no-skill` | 4/4 | 0 | 0 | 8.5 | 0.0420 |
| `reference` | 3/4 | 4.50 | 1.25 | 21.5 | 0.2066 |
| `baseline` | 3/4 | 4.75 | 1.00 | 23.2 | 0.2428 |

Skills reached by the baseline arm: `go-code` 4/4, `go-style-core` 4/4,
`go-data-structures` 4/4, `go-error-handling` 3/4, `go-linting` 3/4, `go-http`
1/4 (on `gateway`). No skilled session in either arm read a reference file.

## What the model followed

**Declaration Budget: reported in 4 of 4 baseline sessions**, as a count with a
reason per entry. On `gateway` the baseline session added 2 functions against
the reference session's 4, and named them: a `methodNotAllowed(allow)` factory
registered on three routes and a `writeJSON` shared by two. On `feed` the
budget did not cut, it rationalized: the session declared two package-level
types, `outEvent` and `document`, and charged them to "reason 2, a distinct
format boundary", where the reference session shaped the same wire document
with a function-local type and an anonymous struct, which the harness does not
count and the reader does not have to learn. Whether the count changes what
gets written, rather than how it is explained, is the n=5 question.

**`go-http` HEAD example: followed.** The baseline `gateway` session registered
`HEAD /healthz`, `HEAD /accounts` and `HEAD /accounts/{id}` beside the `GET`
patterns, with `Allow: GET` on the 405, and its report carries the row
`HEAD /healthz → 405 with Allow: GET`. The reference session wrote the same
sentence every skilled session of the two n=5 controls wrote — "automatic 405
via Go's method-specific `ServeMux` patterns" — and failed the same three
assertions. This is the first skilled session at this effort to pass the HEAD
assertions, after 0 of 10 on 2026-09-09 and 2026-09-10; it is one session,
and the control arm passed too by checking `r.Method` by hand.

## What the model did not follow

**Contract Table before the first edit: 0 of 4.** Every baseline session wrote
its contract check in the final message, after the code, as a post-hoc
checklist — a 7-row table on `feed`, bullets elsewhere. The skill's step 2 says
to write it in the transcript before routing; the model implemented first and
listed second. This is the shape the 1.6.0 changelog already recorded for
owner loads ("prose alone did not make it happen") and the reason the routing
hook exists; the hook cannot see text, so nothing enforced this step. The
post-hoc list on `gateway` had no row for an empty account list — the fixture's
doc comment never names that case, the table rule does ("the empty, nil, and
invalid inputs"), and a row written before the code is what the rule was for.

**`go-data-structures` Copy caveat: not followed, and the 2026-09-10 chain
reproduced step by step.** In the baseline `gateway` trace the first body edit
used `make([]Account, len(accounts))` plus `copy` (non-nil), the routing hook
blocked it for `go-data-structures`, the skill loaded, the next message reads
"I'll use `slices.Clone`/`slices.SortFunc` instead of manual copy/`sort.Slice`",
and the body that landed carries `sorted := slices.Clone(accounts)` with
`result := sorted` on the unfiltered path — `null` for a nil list. The Copy row
now says in the same cell that `Clone` keeps nil and that a non-nil copy is
`make` plus `copy`. The positive half of the row was applied; the caveat in the
same cell was not. The reference session kept `make`+`copy` and would have
passed that assertion. Read against the two example fixes together, the
asymmetry is the finding of this run: the fix that showed the code to write
(a `HEAD` pattern) was copied; the fix that said when not to use the code
shown (`Clone`) was not.

**Cost.** 5.79x against 4.92x, 1.7 more turns per session: the table and the
budget are output tokens, and the hook still forces `go-data-structures` on
`make(`/`append(` in every session (4/4 loads, 1.00 blocks per session).

## What this changes

- Nothing in the READMEs: n=1 per cell is not a control.
- The `go-http` Routing example is worth measuring: `-tasks gateway`,
  `reference` against `baseline`, n≥5, HEAD assertions and `Δfuncs` as the
  readings.
- The `go-data-structures` cell edit is not enough on its own. Either the Copy
  row shows the non-nil copy as code in its own row, so the positive form is
  the one that gets copied, or the routing hook stops naming
  `go-data-structures` for `make(`/`append(` — the load itself is what starts
  the rewrite, in this run as in the control.
- The Contract Table needs a mechanism, not a sentence. A hook cannot see
  transcript text; a table the model must *write as a file* can be seen — the
  rows as a `_test.go` the harness already runs as `model_tests` is the
  candidate, and it is the shape a shell-equipped session would use anyway.
