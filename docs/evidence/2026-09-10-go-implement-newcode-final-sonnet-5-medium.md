# New-Code Sections — three arms on the implementation corpus, Sonnet 5 medium (n=5)

The run that can set a README cell: the full implementation corpus at five
repetitions per cell with a `no-skill` arm beside two skill trees. `reference`
is the tree the [previous n=5 control](2026-09-10-go-implement-newcode-control-sonnet-5-medium.md)
measured an hour earlier as `baseline`; `baseline` is that tree plus three
refinements made on its reading. It is also the run that pulls the earlier
readings back toward the mean, and that is its most useful result.

The refinements in the `baseline` arm:

- Reason 2 of the `go-code` Declaration Budget covers an algorithm or a
  resource lifetime and names a wire document, a formatted error, or a sorted
  view as a representation the standard library already expresses; the budget
  lists the declaration-free forms to take first.
- The `go-data-structures` key-collection row splits like the Copy row:
  `slices.Collect`/`slices.Sorted` where an empty result may be nil,
  `slices.AppendSeq(make([]K, 0, len(m)), maps.Keys(m))` where it must encode
  as `[]`.
- The routing hook names `go-testing` alone for a `_test.go`.

## Run

- Finished: 2026-09-10 21:37 UTC
- Runner: `claude` 2.1.267
- Model: `claude-sonnet-5`, reasoning effort `medium`
- Seed: `1`
- Corpus: `implement`
- Fixtures: `catalog`, `feed`, `gateway`, `ledger`
- Arms: `no-skill`, `reference`, `baseline`
- Repetitions: 5 per fixture and arm, 60 sessions total
- `reference`: the working tree as measured in the previous n=5 control,
  plugin SHA-256
  `7e355813e433303a38e12699e7dd3d66809da2fba126b4a09cef32bd2f2d7f70`
- `baseline`: the working tree with the three refinements, plugin SHA-256
  `819bfd31009cb4ff84ec0c3f2df536a8ca5beee0f2c9fd8b3d38328d0fca8239`
- Toolchain: Go 1.27.1 linux/amd64
- Raw report: [`2026-09-10-go-implement-newcode-final-sonnet-5-medium.json`](2026-09-10-go-implement-newcode-final-sonnet-5-medium.json)
  (SHA-256 `964c886678917c3388648df6219dac90d337685a9bbbd13c0145f8039e63721f`)
- Session transcripts:
  [`2026-09-10-go-implement-newcode-final-sonnet-5-medium.traces.tar.gz`](2026-09-10-go-implement-newcode-final-sonnet-5-medium.traces.tar.gz)
  (SHA-256 `2070c2c8ceadaf22091a12ca64701dd41993b88fe6a76bf98134869c08ae9fbc`),
  one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $0.8583 control, $4.7109 reference, $4.0630 baseline — **5.49x** and
  **4.73x**

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <snapshot of the tree at 7e355813> -arms no-skill,reference,baseline \
  -n 5 -j 4 -seed 1 -keep \
  -out ../docs/evidence/2026-09-10-go-implement-newcode-final-sonnet-5-medium.json
```

All 60 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. No shell in any arm. `go-code` fired in 19/20
reference and 17/20 baseline sessions; the sessions that skipped it loaded
`go-http` alone, all on `gateway`.

## Results

| Fixture | Arm | Golden | Valid | Δlines (valid) | Δfuncs | Δtypes | $ / run |
|---|---|---|---|---|---|---|---:|
| `catalog` | no-skill | 4/5 | 4 | 22, 29, 23, 38 → 28.0 | 0, 0, 0, 2 | 0, 0, 0, 1 | 0.035 |
| `catalog` | reference | 5/5 | 5 | 35, 38, 20, 20, 36 → 29.8 | 2, 2, 0, 0, 2 | 1, 1, 0, 0, 1 | 0.181 |
| `catalog` | baseline | 4/5 | 4 | 19, 37, 20, 21 → 24.2 | 0, 2, 0, 0 | 0, 1, 0, 0 | 0.118 |
| `feed` | no-skill | 5/5 | 5 | 50, 47, 46, 48, 50 → 48.2 | 0 | 0 | 0.042 |
| `feed` | reference | 5/5 | 5 | 47, 45, 43, 47, 49 → 46.2 | 0 | 0, 0, 0, 0, 2 | 0.212 |
| `feed` | baseline | 5/5 | 5 | 44, 50, 41, 42, 43 → 44.0 | 0 | 1, 2, 0, 0, 2 | 0.253 |
| `gateway` | no-skill | **4/5** | 4 | 85, 85, 87, 88 → 86.2 | 0 | 0 | 0.059 |
| `gateway` | reference | **3/5** | 3 | 103, 84, 84 → 90.3 | 5, 2, 2 | 0 | 0.351 |
| `gateway` | baseline | **2/5** | 2 | 94, 73 → 83.5 | 3, 2 | 0 | 0.244 |
| `ledger` | no-skill | 5/5 | 5 | 35, 41, 35, 37, 36 → 36.8 | 0 | 0 | 0.036 |
| `ledger` | reference | 5/5 | 3 | 35, 38, 37 → 36.7 | 0 | 0 | 0.198 |
| `ledger` | baseline | 5/5 | 4 | 37, 36, 39, 37 → 37.2 | 0 | 0 | 0.198 |

| Comparison | `catalog` | `feed` | `gateway` | `ledger` |
|---|---|---|---|---|
| baseline − no-skill, lines (perm p) | −3.8 (0.46) | −4.2 (0.07) | −2.8 (0.67) | +0.5 (0.89) |
| baseline − reference, lines (perm p) | −5.6 (0.44) | −2.2 (0.33) | −6.8 (0.60) | +0.6 (0.80) |
| golden, no-skill vs baseline (Fisher) | 4/5 vs 4/5 | 5/5 vs 5/5 | 4/5 vs 2/5 (0.52) | 5/5 vs 5/5 |

| Arm | golden | valid | contract tests written · pass · fail | skills / session | hook blocks / session | turns | $ / run | × control |
|---|---|---|---|---:|---:|---:|---:|---:|
| `no-skill` | 18/20 | 18 | — | 0 | 0 | 8.8 | 0.0429 | 1 |
| `reference` | 18/20 | 16 | 10 · 8 · 2 | 4.20 | 0.70 | 23.1 | 0.2355 | 5.49 |
| `baseline` | 16/20 | 15 | 9 · 8 · 1 | 3.75 | 0.65 | 20.4 | 0.2032 | 4.73 |

Correctness is level with the control in one arm and two sessions below it in
the other (Fisher p = 0.66); size is smaller than the control on `catalog` and
`feed` by about four lines (p = 0.46 and 0.07), level on `ledger`; cost is
4.7x to 5.5x.

## `gateway` across the runs

This is the fixture every claim in this series rests on, so here is every
skilled session at Sonnet 5 medium, by tree, with the unaided control beside
them. "Example copied" is a session whose code registers a `HEAD` pattern with
a 405 handler, the shape of the `go-http` Routing example.

| Tree | Sessions | Golden | HEAD assertions | Example copied | Manual `r.Method` |
|---|---|---|---|---|---|
| release 1.7.0 (`go-http` with the HEAD bullet) | 19 across five runs | **0** | 0/19 | 0/19 | 0/19 |
| new-code trees (example with a `HEAD` pattern) | 19 across four runs | **12** | 13/19 | 10/19 | 3/19 |
| unaided control | 16 across four runs | **8** | 8/16 | — | 8/16 |

1.7.0 against the new trees: Fisher p = 0.00001. 1.7.0 against the control:
p = 0.001. New trees against the control: **p = 0.30**. The regression the
1.7.0 tree carries on this fixture — method patterns pushed by the example,
the HEAD caveat left in a bullet, 0 of 19 — is established and gone. A benefit
over no skill is not established: the new trees pass at 63% where the control
passes at 50%, and the control's own rate moved from 1/5 and 2/5 in the two
earlier controls to 4/5 here, choosing `r.Method` checks over method patterns
more often than before.

The reading an hour earlier was 4/5 and 7/7 with the router; this run's two
new-tree arms are 3/5 and 2/5, and the example was copied in 2 of the 10
sessions that loaded `go-http` against 8 of 9 in the two runs before. Three
sessions with `go-code` and `go-http` both in context wrote `GET` patterns and
nothing for HEAD, as the 1.7.0 tree did. One of them had written a 141-line
contract test that passed its own run and never mentions HEAD: the model's
contract table missed the clause the fixture is built on. Five sessions per
cell cannot separate a rate near 60% from one near 100%; the aggregate above
is the number to carry, and it says the example moved the rate from zero to
about two thirds, not to one.

## The refinements

- **Error type on `catalog`**: 1 of 4 valid baseline sessions declared a
  `resolveError` type against 3 of 5 reference sessions and 1 of 4 unaided;
  lines 24.2 against 29.8 and 28.0. The direction the budget text asked for,
  at n=5.
- **Package-level wire types on `feed`**: 3 of 5 baseline sessions against 1 of
  5 reference — the opposite direction — and every budget line that declares
  them cites "wire-format representations (Declaration Budget reason 2)"
  verbatim, the reason the refined text now says does *not* apply. The model
  filled the budget's template with the budget's own vocabulary and inverted
  its sense. Types are the one structural count where this arm is worse than
  both others; the `feed` documents are still 44 lines against 48 unaided.
- **Key collection**: no `kinds: null` in 15 `feed` sessions; the one
  occurrence in the previous run is not enough to measure the row either way.
- **Hook**: `_test.go` writes named `go-testing` alone; `go-defensive` loads
  3 → 0, `go-linting` 12 → 9, cost 5.49x → 4.73x. Blocks per session 0.70 →
  0.65.
- **Contract tests**: 10/20 and 9/20 written; `feed` 5/5 in the baseline arm
  and before the first production edit in all five. Three `ledger` sessions
  across the two skill arms failed their own test on the width of the `TOTAL`
  line in the empty case — each expected a different width, none the one
  `%-20s %10d` produces — and are excluded from the means with the golden
  green. Without a shell the test file is a 100–190-line cost and an exclusion
  risk; with one it is the loop the 2026-09-09 repair experiment measured.

## What this changes

- **README**: the Sonnet 5 new-code cell stays ➖ and moves to this run.
  Correctness 16/20 against 18/20 unaided, `gateway` 2/5 against 4/5; size
  −3.8 and −4.2 lines on `catalog` and `feed`, level on `ledger`; cost 4.7x.
  The 1.7.0 tree's `gateway` regression is real and is gone; nothing above the
  unaided rate is established.
- **Kept**: the `go-http` example (0/19 → 10/19 copied, the only edit with a
  mechanism visible in every trace), the hook without the collections hint and
  with test files owned by `go-testing` (cost and the `Clone` chain), the
  Contract Table as a file (followed half the time where prose was followed
  never; pays only with a shell), the split Copy and key-collection rows.
- **Not established**: the Declaration Budget as a restraint. It is reported
  in every session and its reasons are used as a template; the `catalog`
  reading moved the right way and the `feed` reading the wrong way. The
  concision gate works because it is a number that must not grow; a count
  with reasons is not that.
- **Next**: one implementation run with a shell or `-repair` in both arms, so
  the contract tests the model writes can act; `gateway` at n≥10 per arm
  before any correctness claim against the control; a hook that enforces the
  test-before-body order is the one mechanism this series has not tried.
