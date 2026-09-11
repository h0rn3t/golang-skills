# New-Code Sections Control — implementation corpus, reference (1.7.0) against baseline, Sonnet 5 medium (n=5)

The full implementation corpus at five repetitions per cell, with the plugin
tree the only difference between arms: `reference` is release 1.7.0 as it
shipped, `baseline` is the working tree with the new-code edits after the
[smoke](2026-09-10-go-implement-newcode-smoke-sonnet-5-medium.md) and the
[n=3 direction](2026-09-10-go-implement-newcode-n3-sonnet-5-medium.md). This is
a tree-against-tree comparison, not a control against no skills; the same-day
[no-skill control](2026-09-10-go-implement-control-sonnet-5-medium.md) has the
reference digest as its skilled arm, and a three-arm follow-up carrying its
own `no-skill` arm is the run that can set a README cell.

The edits in the `baseline` arm:

- `go-code` Writing New Code: a **Contract Table** written as a table-driven
  `<pkg>_contract_test.go` before the first production edit and run or read
  before closing; a **Declaration Budget** charging every declaration added
  beyond the specification with one of three reasons, reported as a count.
- `go-http` Routing example registers a `HEAD` pattern beside its `GET`
  pattern and says in code that the GET pattern otherwise serves HEAD.
- `go-data-structures` Copy row split into `Clone` (nil may stay nil) and
  `make`+`copy` / `append([]T{}, s...)` (must encode as `[]`); the
  `go-defensive` boundary-copy example shows both.
- The routing hook no longer names `go-data-structures` for `make([]`,
  `make(map` and `append(`.

## Run

- Finished: 2026-09-10 21:19 UTC
- Runner: `claude` 2.1.267
- Model: `claude-sonnet-5`, reasoning effort `medium`
- Seed: `1`
- Corpus: `implement`
- Fixtures: `catalog`, `feed`, `gateway`, `ledger`
- Arms: `reference`, `baseline`
- Repetitions: 5 per fixture and arm, 40 sessions total
- `reference`: `git archive 7c89b4f` (skills byte-identical to release 1.7.0),
  plugin SHA-256
  `5028c91b61340fe99a22284cf0fe4ba32f48d8e7b1cf83baa0e412c88708a2ea`
- `baseline`: working tree at `7c89b4f` plus the uncommitted edits above,
  plugin SHA-256
  `7e355813e433303a38e12699e7dd3d66809da2fba126b4a09cef32bd2f2d7f70`
- Toolchain: Go 1.27.1 linux/amd64
- Raw report: [`2026-09-10-go-implement-newcode-control-sonnet-5-medium.json`](2026-09-10-go-implement-newcode-control-sonnet-5-medium.json)
  (SHA-256 `965ab2462837f579c9d0d5735cca8986f7e9fbb02f18b835d95be90b694582af`)
- Session transcripts:
  [`2026-09-10-go-implement-newcode-control-sonnet-5-medium.traces.tar.gz`](2026-09-10-go-implement-newcode-control-sonnet-5-medium.traces.tar.gz)
  (SHA-256 `1febf814f370c5a1f8c862b7371ef32914638ca75857bb5bcdd8a9acb335ec11`),
  one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $4.0870 reference, $4.2364 baseline — $0.2043 against $0.2118 per
  session, **+3.7%**

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <git archive 7c89b4f> -arms reference,baseline -n 5 -j 4 -seed 1 -keep \
  -out ../docs/evidence/2026-09-10-go-implement-newcode-control-sonnet-5-medium.json
```

All 40 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. No shell in either arm. `go-code` fired in
20/20 reference sessions and 18/20 baseline sessions: two `gateway` sessions
loaded `go-http` alone and never reached the router.

`perm p` is the exact two-sided permutation test over the valid runs'
line deltas; Fisher is the exact two-sided test on pass counts.

## Results

| Fixture | Arm | Golden | Valid | Δlines (valid) | Δfuncs | Δtypes | $ / run |
|---|---|---|---|---|---|---|---:|
| `catalog` | reference | 4/5 | 4 | 20, 20, 38, 21 → 24.8 | 0, 0, 2, 0 | 0, 0, 1, 0 | 0.182 |
| `catalog` | baseline | 4/5 | 4 | 39, 20, 36, 35 → 32.5 | 2, 0, 2, 2 | 1, 0, 1, 1 | 0.188 |
| | | | | +7.8, perm p = 0.40 | | | |
| `feed` | reference | 5/5 | 5 | 44, 44, 39, 40, 45 → 42.4 | 0 ×5 | 0, 2, 0, 0, 1 | 0.196 |
| `feed` | baseline | 4/5 | 4 | 43, 45, 45, 45 → 44.5 | 0 ×4 | 2, 2, 1, 2 | 0.220 |
| | | | | +2.1, perm p = 0.27 | | | |
| `gateway` | reference | **0/5** | 0 | (72–77, all failing) | 1 ×5 | 0 ×5 | 0.276 |
| `gateway` | baseline | **4/5** | 4 | 82, 91, 76, 76 → 81.3 | 2, 5, 2, 2 | 0, 1, 0, 0 | 0.217 |
| | | | | Fisher p = 0.048 | | | |
| `ledger` | reference | 5/5 | 5 | 37, 35, 35, 36, 32 → 35.0 | 0 ×5 | 0 ×5 | 0.163 |
| `ledger` | baseline | 5/5 | 4 | 32, 37, 47, 35 → 37.8 | 0 ×4 | 0, 0, 1, 0 | 0.222 |
| | | | | +2.8, perm p = 0.48 | | | |

| Arm | golden | valid | contract tests written · pass · fail | skills / session | hook blocks / session | turns | output tokens / session | $ / run |
|---|---|---|---|---:|---:|---:|---:|---:|
| `reference` | 14/20 | 14 | 0 · — · — | 4.65 | 1.10 | 22.4 | 3342 | 0.2043 |
| `baseline` | 17/20 | 16 | 10 · 8 · 2 | 3.95 | 0.50 | 20.9 | 4624 | 0.2118 |

Overall golden 17/20 against 14/20 is Fisher p = 0.45: the corpus-level
correctness difference is `gateway`'s, and `gateway` is the fixture whose
traps are live on this model.

Skill loads over twenty sessions — reference: `go-code` 20, `go-style-core`
20, `go-data-structures` 20, `go-error-handling` 19, `go-http` 5, `go-linting`
4, `go-defensive` 3, `go-documentation` 1, `go-security` 1. Baseline:
`go-code` 18, `go-style-core` 18, `go-error-handling` 14, `go-testing` 10,
`go-linting` 8, `go-http` 5, `go-data-structures` 4, `go-defensive` 1,
`go-documentation` 1.

## `gateway`: 0/5 against 4/5

Every reference session fails the HEAD assertions and four of five also
render the empty list as `null` — the pattern of the two n=5 controls and the
n=3 run. Across the 2026-09-09 control, the 2026-09-10 control, the smoke, the
n=3 run and this one, the 1.7.0 tree is **0 of 19** skilled sessions on the
HEAD assertions at this effort.

In the baseline arm every session that loaded `go-code` passes: three of
three here, plus three of three in the n=3 run and the smoke's one, **7 of 7**.
Each registers `HEAD /healthz`, `HEAD /accounts` and `HEAD /accounts/{id}`
beside the `GET` patterns with a 405 handler carrying `Allow: GET`, the shape
of the `go-http` Routing example, and none returns `null`. The two sessions
that loaded `go-http` alone split: one registered the HEAD patterns from the
example and passed with a single skill loaded at $0.089; the other wrote
`GET` patterns only and failed as the reference does. The example is enough on
its own in one of two sessions; with the router it was enough in seven of
seven.

The cost of the fix is lines and helpers: passing implementations run 76–91
lines against the reference's 72–77, and the extra is the three HEAD
registrations, the shared 405 handler and a `writeJSON` — 2, 5, 2, 2 functions
against 1. The 5 is a session that also declared a server struct with three
methods and charged all of it to the Declaration Budget as reuse across three
routes. Per session the baseline arm is cheaper here (0.217 against 0.276)
only because two sessions skipped the router.

## `feed`: 5/5 against 4/5, and the shape of the one failure

The baseline miss is `kinds` rendered as `null` for an account with no
activity. The session initialized `Kinds: []string{}` and then overwrote it
with `slices.Sorted(maps.Keys(doc.Counts))`, which returns nil for an empty
map. `go-data-structures` was loaded — by the model's own routing, the hook
no longer forces it — and its reach-for table offers `slices.Collect(maps.Keys(m))`
as the row while the sentence below the table says `Collect` and `Sorted`
return nil for an empty iterator. The row was applied; the sentence was not.
It is the `Clone` finding again with a different function, and it motivates
the follow-up split of that row.

The session's own contract test caught it exactly:

```text
feed_contract_test.go:70: Render("acct-1", []) = {"account":"acct-1","events":[],"kinds":null,"counts":{}},
                          want {"account":"acct-1","events":[],"kinds":[],"counts":{}}
```

With a shell that failure is one edit away from green. Without one the model
reported the case as read against the code and passed it, and the harness
recorded a wrong implementation beside a right test.

Size is level (+2.1, p = 0.27). Types are not: four of four valid baseline
sessions declared package-level wire types (`document`, `eventJSON`) against
two of five reference sessions, and the contract tests do not use them, so the
test file did not pull them up. The budget reports name the reason: "reason 2,
a distinct format boundary between internal and wire representation". Reason 2
as written — algorithm, resource lifetime, or trust boundary — was read to
cover a representation. The follow-up run's `go-code` narrows it.

## `catalog`: the budget legitimizes

Both arms lose one session to the same bug — the repeated-SKU deduplication
keyed on the results map, so a SKU that fails is asked for twice — at the same
rate as every previous run. The structural difference is in what passed: three
of four valid baseline sessions declared a `resolveError` type with `Error` and
`Unwrap`, 35–39 lines, where one of five reference sessions did and the rest
wrote `fmt.Errorf("...%q: %w", sku, err)` in 20 lines. Both satisfy the
contract's two audiences. The budget made the type visible and reasoned; it did
not stop it, and the +7.8 lines (p = 0.40) are that type. A gate that counts
and asks for reasons is not the concision gate, which counts and refuses
growth; the follow-up names the declaration-free forms to take first.

## `ledger`: 5/5 against 5/5, one own-test exclusion

Baseline `ledger` #3 passed the hidden golden and failed its own contract test
on the width of the `TOTAL` line in the empty case — the test expected 27
characters, the code and the documentation produce 30. The harness excludes a
run whose model tests fail, so it is out of the means; it is the same
exclusion class the refactor corpus records for a wrong self-written
assertion, and the second reading here of what a test costs without a shell to
run it. #4 grew to 47 lines with a `balance` struct charged to the budget.

## The instructions, followed and not

- **Contract test: 10 of 20 sessions** (`catalog` 2, `feed` 3, `gateway` 1,
  `ledger` 4), written before the first production edit in 7. Eight pass, two
  fail: one caught a real defect the session could not run to see, one carried
  a wrong expectation. The file form is followed half the time where the prose
  form was followed never; its payoff needs a shell or a repair turn.
- **Declaration Budget: reported in every session that loaded `go-code`.**
  It changed what is explained, not what is written: `catalog` and `feed` show
  the reasons stretching to cover a representation.
- **`go-http` example: copied wherever `go-http` loaded**, 8 of 9 sessions
  across the three runs on this tree.
- **Hook**: blocks per session 1.10 → 0.50, skills per session 4.65 → 3.95,
  `go-data-structures` 20 → 4 sessions (all self-routed); the test-file hint
  adds `go-testing` in 10. The output-token cost of the test files (+38%) is
  offset by the loads that no longer happen, so the session cost is flat.

## What this changes

- `gateway`'s HEAD and empty-list clauses are live traps on Sonnet 5 medium
  that separate skill trees: 0/19 for the 1.7.0 tree, 7/7 for the new tree
  when the router fired. The `go-http` Routing example is the edit with the
  mechanism visible in every trace.
- Three refinements go into the follow-up three-arm run: reason 2 of the
  Declaration Budget excludes representations and the budget names the
  declaration-free forms to take first; the key-collection row splits the way
  the Copy row did; a `_test.go` names `go-testing` alone in the hook.
- No README cell moves on this run: it has no `no-skill` arm. The follow-up
  carries one.
