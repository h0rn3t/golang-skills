# A refactor's helper meets a Declaration Budget rule or is not added — Sonnet 5 `medium` on `report`, n=5

`report` is the refactor fixture that has been bimodal on Sonnet 5 medium:
a flat `Render` at ±1 line, or three to four write helpers at +12 to +19
([2026-09-18](2026-09-18-go-refactor-orient-move-n3-sonnet-5-medium.md):
4 helpers in 3/3 sessions of one run, 1/3 of another; n=5 reversed the sign,
p = 0.75). The `go-code-refactor` sentence under test replaces three
principle-shaped lines in "Delete Before You Restructure" ("Extract a helper
when it removes repeated logic or hides a meaningful operation. Keep a short,
single-use sequence inline when the helper only renames its steps.") with a
rule that names the count and the shape:

> A helper the refactor adds meets one of the three Declaration Budget rules
> — two call sites in the final code, a caller outside the function that
> names it, or a distinct algorithm — or it is not added: a `writeHeader`,
> `writeRow`, and `writeTotal` that `Render` calls once each rename the steps
> of one call site, and the report names the rule each kept helper meets.

The working tree also carries the same day's `go-code` and `go-http` edits
(`encoding/json/v2` for new JSON, the Delete Pass default-value bullet); none
of them touches this fixture, which has no JSON and no server.

| Arm | Tree |
|---|---|
| `reference` | `git worktree` of `904793d` (release 1.21.1); plugin SHA-256 `1747042b38be8026315c9495e2f2e3ee875eed2de929c0717b9de79942bbe26e` |
| `baseline` | the working tree with the sentence above; plugin SHA-256 `3aac7f9eff0d22f7d433670a0f97bbcd7f8e10dd0049bbc87ffbf2cae35367db` |

## Run

- Finished: 2026-09-19 09:06 UTC
- Runner: `claude` 2.1.267; model `claude-sonnet-5`, effort `medium`; seed `9`
- Corpus `refactor`, fixture `report` alone, arms `reference`, `baseline`, 5 repetitions per arm, 10 sessions, `-j 4`, 0 CLI errors
- Toolchain Go 1.27.1 linux/amd64, golangci-lint 2.13.2 on PATH; the edit hook ran in every session, no session had a shell tool
- Cost: $3.81
- Report: [`2026-09-19-go-refactor-helper-rule-report-n5-sonnet-5-medium.json`](2026-09-19-go-refactor-helper-rule-report-n5-sonnet-5-medium.json) (SHA-256 `dd08599efdaa56649101447e5925a25388793f8404e45b52ebf0c1454ca5696f`); traces [`….traces.tar.gz`](2026-09-19-go-refactor-helper-rule-report-n5-sonnet-5-medium.traces.tar.gz) (SHA-256 `176f9d706dc349106e7180fefa59d2b0a9f3df5f951dbd4dfc0f59959a489c5e`), one `traces/<arm>-report-r<rep>.jsonl` per session

```bash
go run ./cmd/abrun -corpus refactor -tasks report -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root ../../golang-skills-1.21.1 -arms reference,baseline -n 5 -j 4 -seed 9 -timeout 10m -keep -verbose \
  -out ../docs/evidence/2026-09-19-go-refactor-helper-rule-report-n5-sonnet-5-medium.json
```

Each session's `report.go` is read for the helpers it declares and how many
places call each; the p-values are exact permutation tests over the ten
sessions (252 splits).

## The reading

| Arm | Valid | Golden | Lint | Δlines | Δfuncs | Sessions adding a helper | One-call-site helpers | Line gate | Counts reported | $/session |
|---|---|---|---|---|---|---|---|---|---|---|
| `reference` | 5/5 | 5/5 | 6 → 0 in 5/5 | +5.8 | 1.40 | 3/5 | 6 (in 2 sessions) | 2/5 | 0/5 | 0.396 |
| `baseline` | 5/5 | 5/5 | 6 → 0 in 5/5 | +2.0 | 0.20 | 1/5 | 0 | 1/5 | 2/5 | 0.366 |

| Arm | Rep | Δlines | Δfuncs | Helpers (call sites) | Golden | Own test | $ |
|---|---|---|---|---|---|---|---|
| `reference` | 0 | 0 | 0 | — | pass | none | 0.350 |
| `reference` | 1 | +4 | 1 | `formatLine` (2) | pass | none | 0.443 |
| `reference` | 2 | +13 | 3 | `writeHeader` (1), `writeRow` (1), `writeTotal` (1) | pass | none | 0.368 |
| `reference` | 3 | 0 | 0 | — | pass | none | 0.351 |
| `reference` | 4 | +12 | 3 | `writeHeader` (1), `writeRow` (1), `writeTotal` (1) | pass | yes | 0.468 |
| `baseline` | 0 | +2 | 0 | — | pass | none | 0.288 |
| `baseline` | 1 | −1 | 0 | — | pass | none | 0.405 |
| `baseline` | 2 | +7 | 1 | `formatAmount` (4) | pass | none | 0.435 |
| `baseline` | 3 | +1 | 0 | — | pass | none | 0.379 |
| `baseline` | 4 | +1 | 0 | — | pass | none | 0.324 |

**The one-call-site shape did not appear in the baseline arm.** The reference
wrote the `writeHeader`/`writeRow`/`writeTotal` trio in 2/5 sessions, each
helper called once from `Render` — the shape the sentence names, at +12 to
+13 lines; a third reference session's `formatLine` has two call sites, the
row and the total, and passes rule 1. The baseline's one helper,
`formatAmount(cents, width)`, folds the `"%d.%02d"` amount formatting that the
fixture repeats four times into one function with four call sites; its report
says so ("to remove the four duplicated amount-formatting call sites"), which
is rule 1 stated in the code's own terms. Δfuncs 0.20 against 1.40
(p = 0.29), Δlines +2.0 against +5.8 (p = 0.28): the direction is the one
the sentence asks for, and ten sessions do not make it a claim.

**The remaining growth is spacing, not structure.** The three baseline
sessions at +1 and +2 added the `errors` import (the hook's `perfsprint`
finding turned the static `fmt.Errorf` into `errors.New`) and a blank line
between the two guard clauses and the loop body; every session in both arms
flattened the nested `if` into guards in the original order and swapped
`b.WriteString(fmt.Sprintf(…))` for `fmt.Fprintf(&b, …)`, which is the
6 → 0 lint column. Two reference sessions came out at exactly 0 lines by
writing no blank line; the line gate (2/5 against 1/5) reads that spacing,
not the helpers.

**Nothing else moved.** Golden 10/10; skills loaded `go-code-refactor`,
`go-style-core`, `go-error-handling` in 10/10 (reference r4 also `go-testing`
and `go-linting`, and wrote the run's only test); cost $0.366 against $0.396
a session; two baseline sessions stated a line count in their report against
none in the reference, both without a shell.

## What this establishes

- On the fixture whose bimodality motivated it, the sentence took the
  one-call-site helper from 2/5 sessions to 0/5 on this seed, and the
  helpers that remain on either side (`formatLine`, `formatAmount`) are ones
  two or four call sites need. That is the shape the rule was written for.
- Neither mean clears p = 0.05 at n=5, so the corpus README's line-count
  numbers do not change on this run; a claim needs the four-fixture corpus at
  n=3 or `report` at n=10.
- The `Render` flatten and the `Fprintf` swap are identical in both arms,
  so the sentence cost nothing on the rest of the refactor.

## Next

- The four refactor fixtures at n=3 against the same reference, reading
  `Δfuncs` beside `Δlines`; `dispatch` is the other fixture where the
  unaided arm has extracted `put`/`del`/`eventKey`.
- If `formatAmount`-style helpers with several call sites show up as growth
  on `report`, that is the rule working, not a regression to fix.
