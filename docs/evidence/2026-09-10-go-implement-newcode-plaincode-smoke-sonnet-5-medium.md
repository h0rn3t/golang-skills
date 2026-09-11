# Plain Code Smoke — Sonnet 5 at medium effort, implementation corpus (n=1)

A first look at the `go-code` retune made after reading Anthropic's prompting
guides for Claude Sonnet 5 and Claude Opus 5, against the tree it changes and
against no skills at all. One repetition per fixture and arm reads *which
instructions the model followed*, not what their effect is; no claim about
size, correctness rate, or cost is made below.

The edits under test, all in the working tree:

- `go-code` Writing New Code gains a **Plain Code** section — the body reads
  as the documentation reads, the specification's nouns name the variables, a
  type or document built in one function is declared inside it, steps are
  inline, an error carrying context is wrapped with `%w`, a comment states
  only a constraint the code cannot show — as one positive example rather than
  as caveats beside the budget.
- The **Declaration Budget** counts package-level declarations against an
  expected zero for a body behind an existing signature and admits one for
  two named call sites, a caller that names it, or a distinct algorithm; the
  reason slot the three-arm run showed being filled with the budget's own
  words is gone. The report line is `added package-level declarations: N`
  with the call sites.
- Step 5 reports the outcome, the observed checks, the budget line, and
  material gaps, and does not walk the contract cases again.
- `go-style-core` carries the owner rule for internal comments.

## Run

- Finished: 2026-09-10 22:04 UTC
- Runner: `claude` 2.1.267
- Model: `claude-sonnet-5`, reasoning effort `medium`
- Seed: `1`
- Corpus: `implement`
- Fixtures: `catalog`, `feed`, `gateway`, `ledger`
- Arms: `no-skill`, `reference`, `baseline`
- Repetitions: 1 per fixture and arm, 12 sessions total
- `reference`: the working tree with `skills/go-code/SKILL.md` and
  `skills/go-style-core/SKILL.md` restored to their state before this retune,
  plugin SHA-256
  `819bfd31009cb4ff84ec0c3f2df536a8ca5beee0f2c9fd8b3d38328d0fca8239` — the
  same digest as the `baseline` arm of the
  [three-arm n=5 run](2026-09-10-go-implement-newcode-final-sonnet-5-medium.md),
  so that run's twenty sessions are this arm's history
- `baseline`: the working tree with the retune, plugin SHA-256
  `a97ddd064b15d86a3fa40928acc8d0e2de3799abe5398b6cc4851e364ec9ba3e`
- Toolchain: Go 1.27.1 linux/amd64
- Raw report: [`2026-09-10-go-implement-newcode-plaincode-smoke-sonnet-5-medium.json`](2026-09-10-go-implement-newcode-plaincode-smoke-sonnet-5-medium.json)
  (SHA-256 `a1ab3b17ecc1e527cc77d66ab6992087eeff3f4f9d4937fa3a28a876f9287083`)
- Session transcripts:
  [`2026-09-10-go-implement-newcode-plaincode-smoke-sonnet-5-medium.traces.tar.gz`](2026-09-10-go-implement-newcode-plaincode-smoke-sonnet-5-medium.traces.tar.gz)
  (SHA-256 `97f0fb3c4d71f44089bd8b6b374fe748a65f872d2081c583877b1e5a0328d2f6`),
  one `traces/<arm>-<fixture>-r0.jsonl` per session
- Cost: $0.1725 control, $1.0011 reference, $1.2706 baseline — **5.81x** and
  **7.37x** the control

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <working tree with the two skill files restored> \
  -arms no-skill,reference,baseline -n 1 -j 4 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-10-go-implement-newcode-plaincode-smoke-sonnet-5-medium.json
```

All 12 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. The tool set was `Skill,Read,Glob,Grep,Edit,Write`,
so no session had a shell; the routing hook was active in both skilled arms,
and `go-code` fired in 8 of 8 skilled sessions.

## Results

Production-line delta against the shipped fixture, declared functions and
types added, and the hidden golden verdict. One session per cell.

| Fixture | No skill | Reference | Baseline (retune) |
|---|---|---|---|
| `catalog` | pass, +20 | pass, +19 | pass, +19; contract test 103 lines, written after the body |
| `feed` | pass, +44 | pass, +41 | pass, **+30**; the Plain Code shape |
| `gateway` | pass, +83 (`r.Method` checks) | pass, +76, +1 func (`HEAD` patterns, `make`+`copy`) | **fail** `HEAD /healthz = 200, want 405` and `GET /accounts = "null", want []`, +64, +1 func (`writeJSON`) |
| `ledger` | pass, +43 | pass, +36 | pass, +34 |

| Arm | golden | valid | contract tests written · before the body · pass · fail | skills / session | hook blocks / session | turns | $ / run |
|---|---|---|---|---:|---:|---:|---:|
| `no-skill` | 4/4 | 4 | — | 0 | 0 | 8.75 | 0.0431 |
| `reference` | 4/4 | 4 | 3 · 1 · 3 · 0 | 4.25 | 0.50 | 25.75 | 0.2503 |
| `baseline` | 3/4 | 3 | 4 · 3 · 3 · 1 | 4.75 | 1.50 | 30.0 | 0.3176 |

Skills reached by the baseline arm: `go-code` 4/4, `go-style-core` 4/4,
`go-testing` 4/4, `go-linting` 3/4, `go-error-handling` 3/4, `go-http` 1/4
(on `gateway`). The reference arm: `go-code` 4/4, `go-style-core` 4/4,
`go-error-handling` 4/4, `go-testing` 3/4, `go-http` 1/4, `go-documentation`
1/4, `go-linting` 0/4. No session in either arm read a reference file. No
session in any arm added a package-level type.

## What the model followed

**Plain Code on `feed`: the example's shape, at 30 lines.** The baseline
session's `Render` is the section's example with the fixture's nouns: a
function-local `event` type, one anonymous document initialized as
`[]event{}`, `[]string{}`, `map[string]int{}` in a keyed literal, the loop
appending into the document, `slices.Sort` on the kinds, one `json.Marshal`.
The reference session shaped the same document with the same local type but
three separate variables and a `sort.Strings` collect loop, 41 lines; the
unaided session 44. Read this as shape transfer, not as a finding about the
fixture: the example describes a JSON document with the same members as
`feed` — an ordered list, a sorted distinct list, a count map, each non-nil
when empty — so on this fixture the section is close to a worked solution.
The same-shape risk is recorded in the changelog and is what the n≥5 run has
to read around, or the example has to move to a different shape first.

**Budget line: 3 of 4 baseline reports.** `feed` and `ledger` report `added
package-level declarations: 0`, `gateway` reports 1 and names the two call
sites of `writeJSON`, which is what the rewritten rule admits; `catalog` omits
the line. The reference arm carried its `declarations beyond the
specification` line in 1 of 4 reports. No baseline report used a reason
sentence; the count and the call sites replaced it.

**Contract test before the first production edit: 3 of 4** (`feed`,
`gateway`, `ledger`; `catalog` wrote it after), against 1 of 4 in the
reference arm and 0 of 4 in the first smoke of the test-file form. The test
that was written first on `gateway` carries the case its implementation then
failed — see below.

**Report shape: 2 of 4.** The `gateway` and `ledger` baseline reports follow
step 5 — outcome, the checks observed with `pass` / `unavailable (no shell
tool)`, the budget line, a sentence on what was not run. `feed` and `catalog`
still list every contract case in bullets, `catalog` with check marks: the
instruction that the report does not walk the cases again was applied where
the report was short and ignored where the model had a list ready.

## What the model did not follow, and what failed

**`gateway` baseline: the two live traps, both taken.** With `go-http` loaded,
the session registered `GET` patterns only and wrote the sentence every
failing session of the series writes — "405 from ServeMux's built-in method
mismatch handling" — beside an example whose `HEAD` pattern was not copied;
and it wrote `sorted := slices.Clone(accounts)` and served `sorted` on the
unfiltered path, `null` for a nil list. Its own contract test, written before
the body, has the case:

```text
gateway_contract_test.go:104: GET /accounts body = "null", want "[]"
```

Without a shell the case could not run, the report says every case was
traced by hand and matched, and the harness records a wrong implementation
beside a right test — the third reading of that shape in this series. The
reference session on the same fixture registered the three `HEAD` patterns,
copied with `make`+`copy`, and passed at 76 lines. One session each: across
every Sonnet 5 medium session on the trees carrying the `go-http` `HEAD`
example this run moves the count from 12/19 to 13/21. The retune's text
touches neither trap; the Plain Code example builds a document, not a
handler, and `slices.Clone` is `go-data-structures`' row.

**Cost: 7.37x against 5.81x, +27% per session.** Turns 30.0 against 25.75,
hook blocks 1.50 against 0.50: writing the test file first costs one block
for `go-testing` and the production edit a second, and step 5's "run the
Contract Table file with the closing gate" pulled `go-linting` into 3 of 4
baseline sessions against 0 of 4 reference, in sessions that had no shell to
run it. The contract tests themselves are 73–173 lines of output tokens per
session.

## What this changes

- Nothing in the READMEs: n=1 per cell is not a control.
- The Plain Code section transfers its shape; whether the shape is *smaller*
  on code that is not the example's document — `gateway` at 64 lines against
  76 is one failing session — is the n≥5 question, and `feed` cannot answer
  it while the example and the fixture share a shape. Either move the example
  to a shape no fixture has before the run, or read `feed` in that run as an
  example-copying rate rather than a size result.
- The budget line replaced the reason sentence in every report that carried
  it; whether the count restrains, rather than describes, needs the n≥5 run
  with `Δfuncs` and `Δtypes` as the readings and `gateway` at n≥10 before any
  correctness claim.
- The gate load without a shell is cost with no evidence: step 5 could name
  the gate only when a shell exists and otherwise say so in one line.
