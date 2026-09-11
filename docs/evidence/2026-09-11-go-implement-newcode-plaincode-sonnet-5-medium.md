# Plain Code Retune — three arms on the implementation corpus, Sonnet 5 medium (n=5)

The run that measures the `go-code` retune made against Anthropic's prompting
guides for Claude Sonnet 5 and Claude Opus 5, after its
[one-repetition smoke](2026-09-10-go-implement-newcode-plaincode-smoke-sonnet-5-medium.md).
`reference` is the tree the [2026-09-10 three-arm run](2026-09-10-go-implement-newcode-final-sonnet-5-medium.md)
measured as its `baseline`, byte for byte; `baseline` is that tree plus the
retune; `no-skill` is the unaided control on the same day, model, and effort.

The retune in the `baseline` arm:

- `go-code` Writing New Code gains a **Plain Code** section: the body reads as
  the documentation reads, the specification's nouns name the variables, a
  type or document built in one function is declared inside it, steps are
  inline, an error carrying context is wrapped with `%w`, a comment states
  only a constraint the code cannot show — with one positive example.
- The **Declaration Budget** counts package-level declarations against an
  expected zero for a body behind an existing signature and admits one for
  two named call sites, a caller that names it, or a distinct algorithm; the
  reason slot is gone and the report line names the call sites.
- Step 5 reports the outcome, the observed checks, the budget line and
  material gaps; without a shell every check is `unavailable (no shell)` in
  one line and `go-linting` is left unread; the report names the test file
  instead of walking its cases.
- `go-style-core` carries the owner rule for internal comments.

## Run

- Finished: 2026-09-11 05:57 UTC
- Runner: `claude` 2.1.267
- Model: `claude-sonnet-5`, reasoning effort `medium`
- Seed: `1`; `-j 5`
- Corpus: `implement`
- Fixtures: `catalog`, `feed`, `gateway`, `ledger`
- Arms: `no-skill`, `reference`, `baseline`
- Repetitions: 5 per fixture and arm, 60 sessions total
- `reference`: the working tree with `skills/go-code/SKILL.md` and
  `skills/go-style-core/SKILL.md` restored to their pre-retune state, plugin
  SHA-256 `819bfd31009cb4ff84ec0c3f2df536a8ca5beee0f2c9fd8b3d38328d0fca8239`,
  the digest of the 2026-09-10 three-arm run's `baseline` arm
- `baseline`: the working tree with the retune, plugin SHA-256
  `75a92af6301d1716114242e7b725ec7f24c498c23811987457a1c233faab1e62`
- Toolchain: Go 1.27.1 linux/amd64
- Raw report: [`2026-09-11-go-implement-newcode-plaincode-sonnet-5-medium.json`](2026-09-11-go-implement-newcode-plaincode-sonnet-5-medium.json)
  (SHA-256 `57ba1da75cac569201ce2c69b3c468ee480812e64a413c6f3ab77ccb5e55e35f`)
- Session transcripts:
  [`2026-09-11-go-implement-newcode-plaincode-sonnet-5-medium.traces.tar.gz`](2026-09-11-go-implement-newcode-plaincode-sonnet-5-medium.traces.tar.gz)
  (SHA-256 `51ecc7e5ed081e5950bed8f96d9e99c771e5852e27366b902355fdc18ab57ad2`),
  one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $0.8725 control, $4.8295 reference, $4.8560 baseline — **5.54x** and
  **5.57x**

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <working tree with the two skill files restored> \
  -arms no-skill,reference,baseline -n 5 -j 5 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-11-go-implement-newcode-plaincode-sonnet-5-medium.json
```

All 60 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. No shell in any arm. `go-code` fired in 17/20
reference and 18/20 baseline sessions. `perm p` is the exact two-sided
permutation test over the valid runs' line deltas; Fisher is the exact
two-sided test on pass counts.

## Results

| Fixture | Arm | Golden | Valid | Δlines (valid) | Δfuncs | Δtypes | $ / run |
|---|---|---|---|---|---|---|---:|
| `catalog` | no-skill | 5/5 | 5 | 23, 22, 20, 38, 39 → 28.4 | 0, 0, 0, 2, 2 | 0, 0, 0, 1, 1 | 0.036 |
| `catalog` | reference | 5/5 | 5 | 20, 21, 19, 38, 21 → 23.8 | 0, 0, 0, 2, 0 | 0, 0, 0, 1, 0 | 0.167 |
| `catalog` | baseline | **3/5** | 3 | 20, 20, 19 → 19.7 | 0 | 0 | 0.148 |
| `feed` | no-skill | 5/5 | 5 | 47, 43, 47, 47, 48 → 46.4 | 0 | 0 | 0.047 |
| `feed` | reference | 5/5 | 5 | 45, 47, 45, 40, 45 → 44.4 | 0 | 0, 2, 0, 0, 0 | 0.283 |
| `feed` | baseline | 4/5 | 4 | 33, 30, 30, 30 → **30.8** | 0 | 0 | 0.244 |
| `gateway` | no-skill | 2/5 | 2 | 82, 94 → 88.0 | 1, 0, 0, 1, 1 | 0 | 0.053 |
| `gateway` | reference | 2/5 | 2 | 90, 80 → 85.0 | 1, 5, 1, 1, 1 | 0 | 0.327 |
| `gateway` | baseline | **5/5** | 4 | 71, 82, 82, 84 → 79.8 | 0, 2, 2, 2, 2 | 0 | 0.321 |
| `ledger` | no-skill | 5/5 | 5 | 35, 39, 35, 39, 36 → 36.8 | 0 | 0 | 0.039 |
| `ledger` | reference | 5/5 | 4 | 40, 35, 37, 35 → 36.8 | 0 | 0 | 0.189 |
| `ledger` | baseline | 5/5 | 4 | 35, 41, 35, 31 → 35.5 | 0 | 0, 1, 0, 0, 0 | 0.259 |

| Comparison | `catalog` | `feed` | `gateway` | `ledger` |
|---|---|---|---|---|
| baseline − no-skill, lines (perm p) | −8.7 (0.16) | −15.6 (0.01) | −8.2 (0.47) | −1.3 (0.61) |
| baseline − reference, lines (perm p) | −4.1 (0.57) | −13.6 (0.01) | −5.2 (0.53) | −1.2 (0.77) |
| reference − no-skill, lines (perm p) | −4.6 (0.26) | −2.0 (0.30) | −3.0 (0.67) | 0.0 (1.00) |
| golden, no-skill vs baseline (Fisher) | 5/5 vs 3/5 (0.44) | 5/5 vs 4/5 (1.00) | 2/5 vs 5/5 (0.17) | 5/5 vs 5/5 (1.00) |

| Arm | golden | valid | contract tests written · before the body · pass · fail | budget line | `unavailable (no shell)` | reports with ≥5 bullets | skills / session | hook blocks / session | turns | $ / run |
|---|---|---|---|---|---|---|---:|---:|---:|---:|
| `no-skill` | 17/20 | 17 | — | — | — | 0/20 | 0 | 0 | 9.0 | 0.0436 |
| `reference` | 17/20 | 16 | 10 · 8 · 9 · 1 | 9/20 | 3/20 | 4/20 | 4.00 | 0.75 | 22.8 | 0.2415 |
| `baseline` | 17/20 | 15 | 12 · 9 · 10 · 2 | 8/20 | 12/20 | 1/20 | 4.30 | 0.80 | 23.4 | 0.2428 |

Skill loads over twenty sessions — reference: `go-code` 17, `go-style-core`
17, `go-error-handling` 13, `go-linting` 10, `go-testing` 10, `go-http` 5,
`go-defensive` 3, `go-data-structures` 3, `go-logging` 1, `go-documentation`
1. Baseline: `go-code` 18, `go-style-core` 18, `go-error-handling` 15,
`go-linting` 13, `go-testing` 12, `go-http` 5, `go-security` 3,
`go-documentation` 1, `go-data-structures` 1.

Correctness is tied at 17/20 in all three arms, and the ties hide two moves
in opposite directions: `gateway` 5/5 against 2/5 and 2/5, `catalog` 3/5
against 5/5 and 5/5. Size is smaller than both other arms on three fixtures;
only `feed` separates (p = 0.01 against each, 0.04 after correcting for four
comparisons), and `feed` is the fixture whose shape the retune's example
shares. Cost is level with the reference tree, 5.6x the control.

## `gateway`: 5/5, and every session registered the `HEAD` patterns

| Arm | Golden | `HEAD` patterns registered | Manual `r.Method` | Contract test before the body |
|---|---|---|---|---|
| no-skill | 2/5 | 0/5 | 2/5 (both passed) | — |
| reference | 2/5 | 2/5 (both passed) | 0/5 | 1/5 |
| baseline | 5/5 | **5/5** | 0/5 | 0/5 |

The three unaided failures and the three reference failures are the same
assertion, `HEAD /healthz = 200, want 405`. Both skilled arms had the
`go-http` Routing example with its `HEAD` pattern in context in all five
sessions; the reference arm copied it in two, the baseline arm in five. The
Contract Table is not the mechanism: no baseline `gateway` session wrote its
test before the body and two wrote none. The one text difference that touches
this shape is the budget. The reference budget asked for a reason per added
declaration, and a `HEAD` fix is a helper plus three registrations — the
shape the reference sessions explained away with "ServeMux handles 405 for
other methods" while the fixture's clause says every non-GET method is a 405.
The retuned budget admits a declaration with two named call sites without an
argument, and four of five baseline sessions charged `writeJSON` and a
`methodNotAllowed` closure or function to it by naming the routes. That is a
reading of the traces, not a demonstrated cause. Across every Sonnet 5 medium
session in the series the trees now stand: release 1.7.0 0/19, the unaided
control 10/21, the 2026-09-10 new-code tree 15/25, this tree 5/6 counting the
smoke's one failure. Five sessions per cell cannot separate a rate near 60%
from one near 100% (Fisher p = 0.17 against the control); `gateway` at n≥10
per arm is still the run before a correctness claim.

## `feed`: the example's shape, four times out of five

Four baseline sessions wrote the same 30-line `Render`: a function-local
`event` type, an anonymous document initialized with `[]event{}`,
`[]string{}` and `map[string]int{}`, the loop appending into the document,
`slices.Sort` on the kinds. It is the Plain Code section's example with the
fixture's nouns, and the section's example describes a JSON document with the
same members — an ordered list, a sorted distinct list, a count map, each
non-nil when empty. The reference arm shaped the same document at 44.4 lines
with separate variables and, in one session, two package-level types; the
control at 46.4. Read the −13.6 and −15.6 as **shape transfer**: the section
works the way the prompting guides say an example works, and on this fixture
that is close to handing over the solution. The one baseline failure is the
fifth session, which collected `doc.Kinds = slices.Sorted(maps.Keys(doc.Counts))`
and rendered `kinds` as `null` for an account with no activity — the line the
example's own comment names as the one to avoid, so the example was copied
in shape and not in that line. `go-data-structures` was loaded in 1 of 20
baseline sessions; its split key-collection row was not in context here.

## `catalog`: no structure, and two sessions with one state too few

No baseline session declared a type or a function: 19.7 lines against 23.8
(one `resolveError` type in the reference arm) and 28.4 (two in the control).
Two of five failed the hidden clause "a repeated SKU costs one round trip",
the same way:

```go
if _, done := names[sku]; done {
    continue
}
```

The result map stands in for the set of SKUs already asked for, so a SKU the
source does not know is asked for again. The three passing sessions kept a
`seen` map beside the result. This bug appears about once per arm per run
across the series; here it is 2 against 0 and 0 (Fisher p = 0.44 against the
control). One reading is that "a value used once is written where it is used
and has no name" and "smallest scope that works" were applied to a *state*
the specification names, not to a value: the section as measured did not say
that fewer names never means fewer states. That line is added after this run
and is unmeasured.

## `ledger`: level, with the recurring own-test exclusion

Level on size (35.5 against 36.8 and 36.8) and 5/5 in every arm. One session
each in the reference and baseline arms is excluded for failing its own
contract test on the width of the `TOTAL` line in the empty case, the third
and fourth readings of that class; the hidden golden passed both. Baseline
`ledger` r4 declared a `balance` type and charged it as "an unexported
representation type", the word the retuned budget says is not a reason — the
count line carried a rationale anyway.

## The instructions, followed and not

- **Budget line: 8 of 20**, against 9 of 20 for the old form. Where N > 0 it
  names the declarations and their call sites (`gateway` r0 and r4); one
  session wrote "Added one package-level declaration: none".
- **`unavailable (no shell)`: 12 of 20**, against 3. Reports with five or
  more bullets fell from 4 to 1 of 20, and no baseline report walks its
  contract cases with check marks; median report length is 923 characters
  against 828, so the reports are differently shaped, not shorter.
- **Contract tests: 12 written, 9 before the body**, against 10 and 8. Ten
  pass; the two failures are the `ledger` width and a `gateway` session that
  expected `Allow: GET` where `ServeMux` writes `GET, HEAD`.
- **`go-linting` loaded in 13 of 20 sessions** against 10, with no shell to
  run anything in any of them: the Resource Routing line that exempts a
  shell-less session did not reduce the loads. Cost is level, 5.57x against
  5.54x.

## What this changes

- **README**: the Sonnet 5 new-code cell moves to this run and reads as a
  mixed signal. Against no skills: −8.5 lines per task, three of four
  fixtures smaller, `feed` −15.6 at p = 0.01 with the example-shape caveat;
  correctness tied at 17/20 with `gateway` 5/5 against 2/5 and `catalog` 3/5
  against 5/5; cost 5.6x.
- **Kept**: the Plain Code section and its example — the shape transfers,
  and on the fixture that is not the example's shape (`gateway`) the sessions
  are smaller and pass; the count-with-call-sites budget line; step 5's
  one-line no-shell report.
- **Added after the run, unmeasured**: a Plain Code line that fewer names
  never means fewer states, after the two `catalog` sessions.
- **Not established**: a benefit over no skills on correctness (tied), the
  `gateway` rate (p = 0.17), the `catalog` regression (p = 0.44). The
  `feed` size result measures example copying until the example or the
  fixture changes shape.
- **Next**: `catalog` and `gateway` at n≥10 per arm; a shell or `-repair` so
  the contract tests the model writes can act on the `null` and the
  width cases they already catch.
