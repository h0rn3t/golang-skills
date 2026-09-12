# Delete Pass — `gateway`, reference against baseline, Opus 5 medium (n=5, two trees)

The measurement behind the Delete Pass edits, on the model and fixture where
the 2026-09-12 traces showed the prose growing. The `reference` arm in both
runs is the committed tree at `b25ae01` (release 1.13.0). Two `baseline`
trees were measured the same morning, one after the other:

- **Tree A** (`782bf572…`): `go-code` Plain Code's comment rule becomes one
  comment per deliberately overridden default and nothing else; a new bullet
  names the standard-library calls that replace a loop and the
  `slices.Sorted(maps.Keys(m))` nil trap; a Delete Pass after the Contract
  Table deletes a comment that restates a case, a name used once, a blank
  line inside one operation, a failure branch or wrap on a call that cannot
  fail for the function's own value, and a doc comment on an unexported
  helper beyond one line; the Declaration Budget says an unexported helper
  carries a one-line comment or none. `abrun` gains `Δbcom`, comment lines
  inside function bodies.
- **Tree B**: Tree A plus two sentences written from the Tree A readings —
  the Delete Pass says a write whose error has nowhere to go is discarded in
  the open (`_, _ = w.Write(body)` with its reason), never bare; the budget
  says a handler registered once is written at its registration.

The [Sonnet 5 pair](2026-09-12-go-implement-delete-pass-feed-gateway-n5-sonnet-5-medium.md)
measured Tree A on `feed` and `gateway`.

## Runs

- Runner: `claude` 2.1.267; model `claude-opus-5`, reasoning effort `medium`;
  seed `1`; `-j 4`; corpus `implement`; fixture `gateway`; arms `reference`,
  `baseline`; 5 repetitions each, 10 sessions a run
- `reference`: a `git worktree` of `b25ae01`, plugin SHA-256
  `c5b603e94104ae6f7badca3c918dda9e258d6a163cb2eca5f80379d494065865`
- Toolchain: Go 1.27.1 linux/amd64; golangci-lint 2.13.2 for the lint line
- Tree A, finished 2026-09-12 10:44 UTC, `baseline` plugin SHA-256
  `782bf572c34526c87d27bab2e982785a0e55d3f4326f02c3497f9394c42e7d72`:
  [`…delete-pass-gateway-n5-opus-5-medium.json`](2026-09-12-go-implement-delete-pass-gateway-n5-opus-5-medium.json)
  (SHA-256 `288a091780932a42455dbfa7f1192fb9727414b57bd8d4bdfbff27f8bdbe021e`),
  transcripts [`….traces.tar.gz`](2026-09-12-go-implement-delete-pass-gateway-n5-opus-5-medium.traces.tar.gz)
  (SHA-256 `b6a7c25892687673ebbc8bfd08396e262ec987f3ce90877440f677eeee4843df`);
  cost $4.4869 reference, $4.7596 baseline
- Tree B: see [Rerun](#rerun-tree-b)

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-opus-5 -effort medium \
  -reference-root <worktree of b25ae01> -arms reference,baseline \
  -tasks gateway -n 5 -j 4 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-12-go-implement-delete-pass-gateway-n5-opus-5-medium.json
```

Every session completed without a CLI error, changed the fixture, stayed out
of the repository checkout, and passed the hidden golden test. The tool set
had no shell. `go-code` fired first in every session, the owners in two
messages (one session in three).

## Tree A

| Arm | rep | Golden | Δlines | Δfuncs | Δbcom | blank lines in body | helpers | writes | lint |
|---|---|---|---:|---:|---:|---:|---|---|---:|
| `reference` | 0 | 1/1 | 82 | 2 | 5 | 5 | `getOnly`, `writeJSON` | `_, _ =` with reason | 0 |
| `reference` | 1 | 1/1 | 88 | 2 | 5 | 9 | `methodNotAllowed`, `writeJSON` | `_, _ =` with reason | 0 |
| `reference` | 2 | 1/1 | 83 | 2 | 3 | 9 | `methodNotAllowed`, `writeJSON` | bare | 2 |
| `reference` | 3 | 1/1 | 82 | 2 | 4 | 10 | `methodNotAllowed`, `writeJSON` | `if err :=` branches | 0 |
| `reference` | 4 | 1/1, own test failed | 75 | 2 | 4 | 8 | `methodNotAllowed`, `writeJSON` | bare | 2 |
| `baseline` | 0 | 1/1 | 70 | 1 | 5 | 7 | `writeJSON` | `_, _ =` / `_ =` with reason | 0 |
| `baseline` | 1 | 1/1 | 77 | 2 | 0 | 5 | `getOnly`, `writeJSON` | bare | 2 |
| `baseline` | 2 | 1/1 | 75 | 2 | 5 | 8 | `headNotAllowed`, `writeJSON` | `//nolint:errcheck` with reason | 2 (`gosec` G104) |
| `baseline` | 3 | 1/1 | 82 | 2 | 1 | 5 | `getOnly`, `writeJSON` | bare | 2 |
| `baseline` | 4 | 1/1 | 68 | 2 | 2 | 5 | `methodNotAllowed`, `writeJSON` | bare | 2 |

| Arm | golden | valid | Δlines | Δfuncs | Δclos | Δbcom | lint after, mean | lint clean | `Skill` msgs | API calls | cache writes | cache reads | output tokens | thinking tokens | median report chars | $ / run |
|---|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| `reference` | 5/5 | 4 | 88, 82, 83, 82 → 83.8 | 2.00 | 0.00 | 4.2 | 0.50 | 3/4 | 2.0 | 8.8 | 45.6K | 290K | 11851 | 5965 | 1723 | 0.898 |
| `baseline` | 5/5 | 5 | 75, 68, 70, 82, 77 → 74.4 | 1.80 | 0.00 | 2.6 | 1.60 | 1/5 | 2.2 | 9.0 | 47.8K | 285K | 13207 | 6894 | 1845 | 0.952 |

Lines −9.3 (exact permutation p = 0.032). The reference session excluded from
the means passed the golden test and failed its own contract test, which
expected `active=` with an empty value to be a 400 where the body served it
as no filter; the golden test has no case for that clause.

### Reading, Tree A

- **The pass takes lines out.** −9.3 against the 1.13.0 tree, more than the
  +7.0 the closure rule had cost the day's earlier pair, with golden 5/5 in
  both arms and cost level. Blank lines in the body fell from 8.2 to 6.0 a
  session and body comments from 4.2 to 2.6; the two baseline sessions still
  at 5 comments carry one per overridden default — the `HEAD` registration,
  the `[]`-not-`null` slice, the timeouts, the discard reason — which is
  what the rule allows.
- **It also took the discard with the comment.** Three baseline sessions
  wrote `w.Write(body)` bare where four reference sessions wrote
  `_, _ = w.Write(body) // reason` or an `if err :=` branch; a fourth wrote
  `//nolint:errcheck` with a reason, which silences `errcheck` and leaves
  `gosec` G104 on the same line. Lint findings went from 0.50 a session to
  1.60, clean sessions from 3/4 to 1/5. The comment bullet — "a comment that
  restates a case, or narrates the next line" — was read as covering the
  reason on the discard, and the discard went with it. Tree B adds the
  sentence that names the idiom and keeps it.
- **Helpers stayed where the closure rule put them**: package-level, 1–2 a
  session, each named with its call sites in the budget line; `Δclos` 0 in
  all ten sessions.

## Rerun (Tree B)

- Finished: 2026-09-12 10:56 UTC; `baseline` plugin SHA-256
  `b13127521f8ff29b4b4dab616bba29c3703eb81f24ebe4470b0ad4a23b7145dd`
- Raw report: [`2026-09-12-go-implement-delete-pass-discard-gateway-n5-opus-5-medium.json`](2026-09-12-go-implement-delete-pass-discard-gateway-n5-opus-5-medium.json)
  (SHA-256 `1a6640f6d744f410145ba36b5531825d8123dcf397b167fc387137efc2c74abd`)
- Session transcripts: [`2026-09-12-go-implement-delete-pass-discard-gateway-n5-opus-5-medium.traces.tar.gz`](2026-09-12-go-implement-delete-pass-discard-gateway-n5-opus-5-medium.traces.tar.gz)
  (SHA-256 `dc143c2d24f0ef2c8031f81986bd2685fbe66a6bd116a045b7a2cef1f68a020d`)
- Cost: $5.1294 reference, $5.0809 baseline — $1.03 against $1.02 a session

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-opus-5 -effort medium \
  -reference-root <worktree of b25ae01> -arms reference,baseline \
  -tasks gateway -n 5 -j 4 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-12-go-implement-delete-pass-discard-gateway-n5-opus-5-medium.json
```

| Arm | rep | Golden | Δlines | Δfuncs | Δbcom | blank lines in body | helpers | writes | lint |
|---|---|---|---:|---:|---:|---:|---|---|---:|
| `reference` | 0 | 1/1 | 84 | 2 | 6 | 9 | `writeJSON`, `methodNotAllowed` | `_, _ =` with reason | 0 |
| `reference` | 1 | 1/1 | 86 | 2 | 8 | 5 | `methodNotAllowed`, `writeJSON` | `_, _ =` with reason | 0 |
| `reference` | 2 | 1/1 | 86 | 1 | 9 | 8 | `writeJSON` | bare, with reason | 2 |
| `reference` | 3 | 1/1 | 94 | 2 | 8 | 5 | `getOnly`, `writeJSON` | `_, _ =` | 0 |
| `reference` | 4 | 1/1 | 89 | 2 | 6 | 8 | `getOnly`, `writeJSON` | bare, with reason | 2 |
| `baseline` | 0 | 1/1 | 76 | 1 | 4 | 8 | `writeJSON` | `_, _ =` with reason | 0 |
| `baseline` | 1 | 1/1 | 81 | 2 | 4 | 10 | `getOnly`, `writeJSON` | `_, _ =` with reason | 0 |
| `baseline` | 2 | 1/1 | 72 | 2 | 5 | 8 | `methodNotAllowed`, `writeJSON` | `_, _ =` with reason | 0 |
| `baseline` | 3 | 1/1 | 75 | 2 | 7 | 4 | `getOnly`, `writeJSON` | `_, _ =` with reason | 0 |
| `baseline` | 4 | 1/1 | 75 | 2 | 6 | 8 | `methodNotAllowed`, `writeJSON` | `_, _ =` with reason | 0 |

| Arm | golden | valid | Δlines | Δfuncs | Δclos | Δbcom | lint after, mean | lint clean | `Skill` msgs | API calls | cache writes | cache reads | output tokens | thinking tokens | median report chars | $ / run |
|---|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| `reference` | 5/5 | 5 | 86, 89, 84, 86, 94 → 87.8 | 1.80 | 0.00 | 7.4 | 0.80 | 3/5 | 2.0 | 10.2 | 48.0K | 366K | 14466 | 7570 | 1505 | 1.026 |
| `baseline` | 5/5 | 5 | 72, 75, 76, 75, 81 → 75.8 | 1.80 | 0.00 | 5.2 | 0.00 | 5/5 | 2.4 | 9.8 | 49.5K | 340K | 14024 | 7817 | 1471 | 1.016 |

Lines −12.0 (exact permutation p = 0.008); the arms do not overlap (72–81
against 84–94). Every baseline session reported the budget line with both
call sites for every helper.

### Reading, Tree B

- **The discard came back and the lines stayed down.** All five baseline
  sessions wrote `_, _ = w.Write(body)` (or `io.WriteString`) with a reason
  on the line, and all five are lint-clean; two reference sessions wrote the
  same call bare. −12.0 lines against the 1.13.0 tree, golden 5/5 in both
  arms, cost $1.02 against $1.03 a session, `Skill` messages 2.4 against
  2.0. Read with Tree A, the sentence that names the idiom to keep cost no
  lines and removed the one regression the pass had introduced.
- **Where the twelve lines went.** In the shortest baseline session against
  a typical reference one: the `byID` map built for one lookup is gone
  (`slices.BinarySearchFunc` on the sorted slice at the one call site); the
  `json.Marshal` failure branch on a value the server built itself is gone
  (`body, _ := json.Marshal(v)` with its reason, four lines to one); the
  unfiltered and filtered list paths are one loop; the copy-and-sort
  preamble lost its two-line explanation; and each helper carries a one-line
  comment where the reference's carried two. Blank lines in the body did
  not move (7.6 against 7.0), and the pass is not the paragraph rule.
- **Body comments 5.2 against 7.4.** Two of each session's comments are the
  discard reasons the rule keeps, so the comparable narration is about 3
  against 5 a session. The 12–14 of the high-effort sweep were not
  re-measured; this pair is at `medium`.
- **Helpers are unchanged in number and kind** — `writeJSON` plus one
  `HEAD`/method helper, package-level, each with two or three named call
  sites — so the closure rule of 1.13.0 and the pass compose: the helpers
  the budget counts stay counted, and the lines around them go.

## What this changes

- The Delete Pass ships as measured on Tree B: −12.0 lines on `gateway`
  against 1.13.0 at n=5 with correctness, cost, and lint held or improved.
  Tree A is recorded as the version that regressed lint and why.
- `Δbcom` and the lint line are what separated the two trees; `Δlines`
  alone read them as the same edit.
- The README's Opus 5 new-code cell carries this pair as a qualifier; the
  cell itself stays on the three-arm n=3 run, which has a control.
