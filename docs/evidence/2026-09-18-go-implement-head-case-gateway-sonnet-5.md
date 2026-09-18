# The `HEAD` case as a named case, not an example of a principle — Sonnet 5 on `gateway`, `medium` at n=5 and `low` at n=3

Every Sonnet 5 golden miss on `gateway` today was one of two clauses: `HEAD`
on a `GET` path answered 200 instead of 405, or a nil account list encoded as
`null`. Twenty `gateway` sessions from the evening's runs, read against the
model's own test file, gave the mechanism: every session whose contract test
carried a `HEAD` case passed the `HEAD` clause, because the edit hook runs the
package tests and the model fixes what it printed; every session without one
shipped `GET` patterns and nothing for `HEAD`. At `medium` the case was
missing in 2/13 tests; at `low` 3/4 sessions wrote no test at all and the
fourth wrote one without `HEAD`. One session registered `HEAD` for the two
`/accounts` routes and not for `/healthz` — the "beside each `GET` route"
rule applied to the resource routes and not carried to the health check,
which is how the Sonnet 5 prompting guide describes the model: it "does not
silently generalize an instruction from one item to another".

Two edits, both stating the scope in the rule and the exception in the code:

- `go-code` "Contract Table": the sentence that derived the `HEAD` case from
  a principle ("the member a library default treats unlike the rest … `HEAD`
  under a `GET` pattern, `t` and `1` under `strconv.ParseBool`") now names
  the case — "one `HEAD` request per `GET` path in the contract, the health
  check included" — and gives the reason after it.
- `go-http` "Routing": the example registers `HEAD` for both `GET` routes
  through one local `methodNotAllowed`, with the comment "One per GET
  pattern, the index included"; the bullet says "every `GET` pattern gets a
  `HEAD` pattern — the health check and the index as much as the resource
  routes, so the file registers as many `HEAD` patterns as `GET` patterns —
  or every handler opens with the `r.Method != http.MethodGet` check; the
  test file sends `HEAD` to each `GET` path".

| Arm | Tree |
|---|---|
| `reference` | `git worktree` of `66f06cb` with the day's two hook changes (the card route) copied in, so the arms differ in the two skill files only; plugin SHA-256 `f77b3f47783f032ee71e1f6d7a761039f29af3af628f48407f874d56d949e7c1` |
| `baseline` | the working tree: the same plus the two edits above; plugin SHA-256 `b3d1240d7660af4731d2e8069840baee39278c3a74235f057708e66af6d3ad00` |

## Runs

- Runner: `claude` 2.1.267; model `claude-sonnet-5`; corpus `implement`, fixture `gateway` alone; seed `6`; `-keep -verbose`; 0 CLI errors in both
- `medium`: 5 repetitions per arm, 10 sessions, `-j 4`, finished 2026-09-18 20:43 UTC, $6.42. Report: [`2026-09-18-go-implement-head-case-gateway-n5-sonnet-5-medium.json`](2026-09-18-go-implement-head-case-gateway-n5-sonnet-5-medium.json) (SHA-256 `d2dd57fe6b6636ee077140dd2a2125b8d48e0b09da5efeaec9618a8af599c561`); traces [`….traces.tar.gz`](2026-09-18-go-implement-head-case-gateway-n5-sonnet-5-medium.traces.tar.gz) (SHA-256 `e8c132aa74a811640621d7c3411c35346393e8737cff30ef3bba01438604d672`)
- `low`: 3 repetitions per arm, 6 sessions, `-j 3`, finished 20:38 UTC, $2.57. Report: [`2026-09-18-go-implement-head-case-gateway-n3-sonnet-5-low.json`](2026-09-18-go-implement-head-case-gateway-n3-sonnet-5-low.json) (SHA-256 `ef1e86bb966cdb0a4380feb7353f9c6fd4aa5aa0b6539512d540ba0652addb86`); traces [`….traces.tar.gz`](2026-09-18-go-implement-head-case-gateway-n3-sonnet-5-low.traces.tar.gz) (SHA-256 `54237f4d0638c908c2093bde23f2845e5f040eca53c112dc363f0edfae45651a`)

```bash
go run ./cmd/abrun -corpus implement -tasks gateway -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <worktree of 66f06cb + hooks> -arms reference,baseline -n 5 -j 4 -seed 6 -timeout 10m -keep -verbose -out <json>
# -effort low -n 3 -j 3 for the second run
```

Toolchain Go 1.27.1 linux/amd64, golangci-lint 2.13.2 on PATH; the edit hook
ran after every `.go` edit; no session had a shell tool. Besides the corpus
columns, each session's own `*_test.go` (renamed `.model` by `abrun`) and
`gateway.go` are read for: a `HEAD` case, a `HEAD` case naming `/healthz`, a
nil-accounts case, a body-text comparison with `[]`, `HEAD` handling in the
code, and the number of `"GET /` and `"HEAD /` patterns registered.

## The reading

`medium`, n=5:

| Arm | Golden | Own test | `HEAD` case | `HEAD /healthz` case | Body text `[]` | `HEAD` in code | `HEAD` patterns / `GET` patterns | Hook findings | Turns/s | $/session |
|---|---|---|---|---|---|---|---|---|---|---|
| `reference` | 4/5 | 5/5 | 5/5 | 5/5 | 2/5 | 5/5 | 3/3 in 5/5 | 41 | 18.0 | 0.633 |
| `baseline` | 5/5 | 5/5 | 5/5 | 4/5 | 4/5 | 5/5 | 3/3 in 5/5 | 33 | 19.6 | 0.651 |

`low`, n=3:

| Arm | Golden | Own test | `HEAD` case | `HEAD /healthz` case | Body text `[]` | `HEAD` in code | `HEAD` patterns / `GET` patterns | Hook findings | Turns/s | $/session |
|---|---|---|---|---|---|---|---|---|---|---|
| `reference` | 0/3 | 2/3 | 0/3 | 0/3 | 1/3 | 0/3 | 0/3 in 3/3 | 6 | 12.7 | 0.357 |
| `baseline` | 2/3 | 3/3 | 3/3 | 1/3 | 2/3 | 3/3 | 3/3 in 3/3 | 22 | 18.0 | 0.500 |

Per session:

| Effort | Arm | Rep | Golden | Failing clause | Own test | `HEAD` case | `HEAD /healthz` | Body `[]` | `HEAD` patterns | $ |
|---|---|---|---|---|---|---|---|---|---|---|
| `medium` | `reference` | 0 | **fail** | nil list → `null` | yes | yes | yes | no | 3 | 0.654 |
| `medium` | `reference` | 1 | pass | | yes | yes | yes | no | 3 | 0.661 |
| `medium` | `reference` | 2 | pass | | yes | yes | yes | yes | 3 | 0.648 |
| `medium` | `reference` | 3 | pass | | yes | yes | yes | no | 3 | 0.625 |
| `medium` | `reference` | 4 | pass | | yes | yes | yes | yes | 3 | 0.577 |
| `medium` | `baseline` | 0 | pass | | yes | yes | yes | yes | 3 | 0.433 |
| `medium` | `baseline` | 1 | pass | | yes | yes | no | no | 3 | 0.738 |
| `medium` | `baseline` | 2 | pass | | yes | yes | yes | yes | 3 | 0.470 |
| `medium` | `baseline` | 3 | pass | | yes | yes | yes | yes | 3 | 0.797 |
| `medium` | `baseline` | 4 | pass | | yes | yes | yes | yes | 3 | 0.816 |
| `low` | `reference` | 0 | **fail** | `HEAD` → 200 ×3, `null` | no | no | no | no | 0 | 0.271 |
| `low` | `reference` | 1 | **fail** | `HEAD` → 200 ×3 | yes | no | no | no | 0 | 0.419 |
| `low` | `reference` | 2 | **fail** | `HEAD` → 200 ×3 | yes | no | no | yes | 0 | 0.382 |
| `low` | `baseline` | 0 | pass | | yes | yes | no | yes | 3 | 0.423 |
| `low` | `baseline` | 1 | pass | | yes | yes | yes | yes | 3 | 0.693 |
| `low` | `baseline` | 2 | **fail** | nil list → `null` | yes | yes | no | no | 3 | 0.384 |

**On the `HEAD` clause the edit is 8/8 against 5/8, and the five are all
`medium`.** With the named case every session at both efforts wrote a `HEAD`
case and registered three `HEAD` patterns beside its three `GET` patterns,
`/healthz` included, even in the four sessions whose test did not name
`/healthz` — the `go-http` count rule did what the test did not. Without it,
`medium` on this seed wrote the case and the patterns in 5/5 (the day's
earlier `medium` rate was 11/13), and `low` in 0/3: three sessions of `GET`
patterns with nothing for `HEAD`, one of them with no test at all.

**Golden 5/5 against 4/5 at `medium`, 2/3 against 0/3 at `low`; every
remaining miss is `null`.** The `HEAD` clause did not fail once in the
`baseline` arm at either effort. The `reference` `medium` miss and the
`baseline` `low` miss are the other clause: a nil account list encoded as
`null`, in sessions whose own test decoded the body into a slice
(`json.Unmarshal` reads `null` as an empty slice) and so passed on the body
the contract forbids. Body-text comparison with `[]` appeared in 6/8
`baseline` tests against 3/8 — the paragraph that got the `HEAD` sentence is
the one that says the empty case is built from nil, so the neighborhood may
have moved it; the body-text rule itself was added after this run and is
measured separately.

**Cost: level at `medium`, +40% at `low`, which is the work.** $0.651
against $0.633 at `medium`. At `low` the `reference` sessions cost $0.357
because they wrote 67 lines and no `HEAD` handling in 12.7 turns; the
`baseline` sessions wrote 84 lines, a test with `HEAD` cases, three more
patterns, and met the edit hook 22 times against 6 fixing what those tests
printed. `low` is the level the guide describes as scoped to the ask; the
edit makes `HEAD` part of the ask.

## What this establishes

- Naming the case ("one `HEAD` request per `GET` path, the health check
  included") and making the code rule a count ("as many `HEAD` patterns as
  `GET` patterns") gets the `HEAD` clause right in 8/8 Sonnet 5 sessions
  against 5/8, with the whole difference at `low`, where the principle-shaped
  sentence produced 0/3. This is the Sonnet 5 guide's literalism read as an
  authoring rule: the pack's own template already says "state a rule's scope
  in the rule"; the `HEAD` sentence had not.
- At `medium` on this seed the old wording was already 5/5 on `HEAD`, so the
  `medium` reading is no regression at level cost; the day's earlier
  `medium` misses (2/13) are what the edit is for there.
- The other `gateway` clause, nil → `null`, is the same shape one level down:
  the model's own test decodes the body and cannot see `null`. That is the
  next edit, in the same paragraph.

## Next

- The body-text rule for the empty-list case, measured the same way.
- `gateway` at `medium` n≥10 on the two wordings is what a golden claim at
  that effort costs; the `HEAD` column at n=5 is already at its ceiling.
