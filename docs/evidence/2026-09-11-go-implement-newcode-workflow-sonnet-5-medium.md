# Workflow Retune — `catalog`, `feed`, `gateway`, three arms, Sonnet 5 medium (n=5)

The run that measures the `go-code` workflow edits made after the
[n=10 Plain Code run](2026-09-11-go-implement-newcode-plaincode-n10-sonnet-5-medium.md),
which asked for three things: a mechanism for the shell-less `go-linting`
exemption that prose did not deliver (17/30 loads with no shell in any
session), a way for the Contract Table to reach the `HEAD` clause it had
missed in every one of eleven tests, and a hook message that stops the model
from treating a blocked edit as applied (4 stale edits in 30 sessions).
`reference` is the committed tree at `ead8ce6`, byte for byte the `baseline`
arm of the n=10 run; `baseline` is the working tree with the edits below;
`no-skill` is the unaided control on the same day, model, and effort.

The edits in the `baseline` arm:

- **`go-code` workflow, six steps.** Step 2 checks the tool list for a shell
  tool by name (`Bash` in Claude Code) and decides the task's shape; step 3
  names `go-testing` among the owners when a Contract Table will be written;
  step 4 is the Contract Table itself, for new code only, before the first
  production edit; step 6 reports in a literal four-line shape — outcome,
  `checks:` line, `added package-level declarations:` line, gaps — with
  `checks: unavailable (no shell)` as the whole gate report when no shell tool
  exists. The `go-linting` routing line and the Close With The Gate section
  repeat the no-shell rule at the point of use.
- **Contract Table.** A clause written as a class ("any other method", "any
  other value") takes its case from the member a library default treats
  unlike the rest — `HEAD` under a `GET` pattern, `t`/`1` under
  `strconv.ParseBool`, the bare path under a subtree pattern — with a
  `HEAD /healthz` row in the example table and the `GET`-serves-`HEAD`
  default in the list the contract overrides.
- **Declaration Budget.** Rule 2 adds one sentence: a caller that inspects or
  matches an error uses `errors.Is` on the wrapped sentinel, which needs no
  type.
- **Hook.** The routing gate's block message says the edit was not applied
  and the file is unchanged, and asks for the same edit again.
- **`go-http`.** The Routing bullet leads with "A `GET` pattern also serves
  `HEAD`" and names the test that catches it; Related Skills routes new
  handlers and servers to `go-code`.

## Run

- Finished: 2026-09-11 08:07 UTC
- Runner: `claude` 2.1.267
- Model: `claude-sonnet-5`, reasoning effort `medium`
- Seed: `1`; `-j 5`
- Corpus: `implement`, `-tasks catalog,feed,gateway`
- Arms: `no-skill`, `reference`, `baseline`
- Repetitions: 5 per fixture and arm, 45 sessions total
- `reference`: a `git worktree` of `ead8ce6`, plugin SHA-256
  `52832538f566efff400a1f2a5f1a1cf5e77877a8f2d60f08f0f86053c7d79a0b` — the
  `baseline` digest of the n=10 run
- `baseline`: the working tree with the edits above, plugin SHA-256
  `f816124fac7202590aa6849a796454920781ad3d2c63afca2007ecc4cc61a68f`
- Toolchain: Go 1.27.1 linux/amd64
- Raw report: [`2026-09-11-go-implement-newcode-workflow-sonnet-5-medium.json`](2026-09-11-go-implement-newcode-workflow-sonnet-5-medium.json)
  (SHA-256 `292304351ad5b27656cfcc369a8a9698c461bef031af7201c2c0de53936a8f89`)
- Session transcripts:
  [`2026-09-11-go-implement-newcode-workflow-sonnet-5-medium.traces.tar.gz`](2026-09-11-go-implement-newcode-workflow-sonnet-5-medium.traces.tar.gz)
  (SHA-256 `2a6825a99ad63e98193bb4ba07b9e31709b1e489505779529bbf428edf99fe03`),
  one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $0.6595 control, $3.6790 reference, $4.7495 baseline — **5.58x** and
  **7.20x**

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <worktree of ead8ce6> -arms no-skill,reference,baseline \
  -tasks catalog,feed,gateway -n 5 -j 5 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-11-go-implement-newcode-workflow-sonnet-5-medium.json
```

All 45 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. No shell in any arm. `go-code` fired in 13/15
reference and **15/15** baseline sessions; the reference misses were a
`catalog` session that loaded nothing and a `gateway` session that loaded
`go-http` alone.

## Results

| Fixture | Arm | Golden | Valid | Δlines (valid) | Δfuncs | Δtypes | $ / run |
|---|---|---|---|---|---|---|---:|
| `catalog` | no-skill | 4/5 | 4 | 23, 22, 26, 38 → 27.2 | 2 ×1, 0 ×3 | 1 ×1, 0 ×3 | 0.035 |
| `catalog` | reference | 4/5 | 4 | 21, 19, 19, 19 → 19.5 | 0 | 0 | 0.158 |
| `catalog` | baseline | **5/5** | 5 | 19, 19, 19, 19, 19 → **19.0** | 0 | 0 | 0.255 |
| `feed` | no-skill | 5/5 | 5 | 47, 48, 48, 48, 47 → 47.6 | 0 | 0 | 0.039 |
| `feed` | reference | 5/5 | 5 | 32, 32, 30, 30, 30 → 30.8 | 0 | 0 | 0.249 |
| `feed` | baseline | 5/5 | 5 | 30, 33, 35, 35, 30 → 32.6 | 0 | 0 | 0.245 |
| `gateway` | no-skill | 5/5 | 5 | 86, 86, 87, 99, 83 → 88.2 | 3 ×1, 0 ×4 | 0 | 0.058 |
| `gateway` | reference | 4/5 | 4 | 80, 77, 67, 84 → 77.0 | 2 ×2, 1 ×2 | 0 | 0.329 |
| `gateway` | baseline | 4/5 | 4 | 84, 88, 94, 73 → 84.8 | 2 ×2, 1 ×1, 0 ×1 | 0 | 0.450 |

| Comparison | `catalog` | `feed` | `gateway` |
|---|---|---|---|
| baseline − no-skill, lines (perm p) | −8.2 (0.008) | −15.0 (0.008) | −3.5 (0.51) |
| baseline − reference, lines (perm p) | −0.5 (0.44) | +1.8 (0.21) | +7.8 (0.29) |
| reference − no-skill, lines (perm p) | −7.8 (0.029) | −16.8 (0.008) | −11.2 (0.024) |
| golden, no-skill vs baseline (Fisher) | 4/5 vs 5/5 (1.00) | 5/5 vs 5/5 (1.00) | 5/5 vs 4/5 (1.00) |
| golden, reference vs baseline (Fisher) | 4/5 vs 5/5 (1.00) | — | 4/5 vs 4/5 (1.00) |

| Arm | golden | contract tests written · before the body | budget line | `checks:` line | says no shell | `go-linting` loads | hook blocks / session | stale edits | skills / session | turns | median report chars | $ / run |
|---|---|---|---|---|---|---|---:|---:|---:|---:|---:|---:|
| `no-skill` | 14/15 | 0 · 0 | 0/15 | 0/15 | 0/15 | 0/15 | 0 | 0 | 0 | 8.7 | 304 | 0.0440 |
| `reference` | 13/15 | 8 · 6 | 8/15 | 0/15 | 10/15 | 9/15 | 1.07 | 1 | 4.13 | 21.5 | 789 | 0.2453 |
| `baseline` | 14/15 | **15 · 10** | **15/15** | **15/15** | **15/15** | **5/15** | 1.00 | 1 | 4.80 | 27.4 | 837 | 0.3166 |

Skill loads over fifteen sessions — reference: `go-code` 13, `go-style-core`
13, `go-error-handling` 13, `go-linting` 9, `go-testing` 8, `go-http` 5,
`go-data-structures` 1. Baseline: `go-code` 15, `go-style-core` 15,
`go-error-handling` 15, `go-testing` 15, `go-linting` 5, `go-http` 5,
`go-security` 2.

## The instructions, followed and not

- **The report shape is followed in every session.** `checks:` line 15/15,
  budget line 15/15 (8/15 for the old wording, 12/30 at n=10), the no-shell
  statement 15/15. Median report length is level (837 against 789
  characters): the shape did not shorten the prose around it.
- **The Contract Table is written in 15/15 sessions, 10 before the body**
  (8 and 6 in the reference arm, 17 and 13 of 30 at n=10). Every baseline
  `gateway` test carries a `HEAD` case — **5/5**, after 0 of 11 across the
  two skilled arms at n=10 and 2/5 in this run's reference arm, which had the
  same `go-http` text.
- **`go-linting` loads fall from 9/15 to 5/15**, not to zero. The five are
  the same move as before — "Now run the tests", then the Skill call — and
  one of them read step 2 literally: "Now let's check for a shell tool to run
  the gate", followed by loading `go-linting`. A named tool in prose halves
  the rate; only a hook that sees the tool set would end it.
- **The hook message did not move the stale-edit count**: one in each arm.
- **Cost rose to 7.20x** from 5.58x, $0.317 against $0.245 a session, with
  turns at 27.4 against 21.5. The whole gap is the Contract Table being
  written every time — `go-testing` loaded 15/15 against 8/15 and a test
  file in every session — minus the four `go-linting` loads saved. Full
  compliance with the table costs about 30% more than the partial compliance
  of the reference tree.

## `catalog`: five sessions, nineteen lines each

Every baseline session wrote the same 19-line body with no added declaration
and passed the repeated-SKU clause; the reference arm failed it once, the
control once, both through the result map standing in for the set of SKUs
already asked for. The `resolveError` type of the n=10 run (2 of 10) did not
appear in 5 sessions; the rule-2 sentence is the text difference, and 0/5
against 2/10 is a direction, not a result.

## `feed`: level with the retune

32.6 against 30.8 lines (p = 0.21), 5/5 correct in every arm, no `null`. The
reference arm here is the Plain Code tree that produced 33.9 at n=10, and the
form held in both skilled arms: no package-level type in any of the ten
sessions.

## `gateway`: the case reaches the clause, the code does not always follow

| Arm | Golden | `HEAD` patterns registered | Manual `r.Method` | `HEAD` case in the model's own test |
|---|---|---|---|---|
| no-skill | 5/5 | 0/5 | 5/5 | — |
| reference | 4/5 | 4/5 | 0/5 | 2/5 |
| baseline | 4/5 | 1/5 | 3/5 | **5/5** |

The one baseline failure (`r2`) wrote the `HEAD` case, then registered
method-less fallback patterns — `mux.HandleFunc("/healthz", methodNotAllowed)`
beside `GET /healthz` — with a comment saying they would catch `HEAD`; the
`GET` pattern is more specific and wins, and without a shell the case was
read against a wrong model of `ServeMux`. Its reference-arm counterpart
(`r3`) failed the same three assertions with no `HEAD` case at all. The other
four baseline sessions passed with the control's `r.Method` form in three and
`HEAD` patterns in one; the reference arm's four passes all used `HEAD`
patterns. Aggregated over Sonnet 5 medium: release 1.7.0 0/19, the unaided
control 25/36, the pre-retune tree 20/35, the Plain Code tree 13/21, this
tree 4/5. At n=5 the arms do not separate (Fisher 1.00 everywhere); what
this run establishes is that the test now names the clause in every session,
and that a shell — or `-repair` — is what would let the case act on the code.

Lines are 84.8 against 77.0 for the reference tree (p = 0.29) and 88.2
unaided: the `r.Method` guard costs a few lines per handler, and one session
(94 lines) declared two helpers.

## What this changes

- **Kept**: the six-step workflow with the named shell check, the report
  shape, the class-clause rule and the `HEAD` row, the rule-2 sentence, the
  hook message, the `go-http` bullet. Every compliance measure moved the way
  the edit intended, and three of them reached 15/15.
- **Not established**: a correctness gain over the reference tree (14/15
  against 13/15) or the control (14/15); the `gateway` rate at n=5; a size
  change on any fixture against the reference tree.
- **The cost is the finding to weigh**: 7.20x against 5.58x, all of it the
  Contract Table written every time. A user who wants the test file pays
  for it here; one who does not should say so in the task.
- **README**: a subset run does not set the cell; the Sonnet 5 new-code cell
  stays on the n=5 full-corpus run and is qualified by this one.
- **Next**: `gateway` with a shell or `-repair`, so a `HEAD` case the model
  writes can fail in front of it; a hook that reads the session's tool set
  before a `go-linting` load, if the five remaining loads are worth a hook.
