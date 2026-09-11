# Plain Code Retune at n=10 — `catalog`, `feed`, `gateway`, three arms, Sonnet 5 medium

The run the [n=5 three-arm report](2026-09-11-go-implement-newcode-plaincode-sonnet-5-medium.md)
asked for: ten repetitions per fixture and arm on the three fixtures whose
readings were open — `gateway` at 5/5 against 2/5 (p = 0.17), `catalog` at
3/5 against 5/5 (p = 0.44), `feed` at −15.6 lines with the example-shape
caveat. `reference` is the pre-retune tree of the whole series; `baseline` is
the retune as released in 1.8.0 plus two edits made after the n=5 run: the
Plain Code line that fewer names never means fewer states, and the example
reshaped from a document sharing `feed`'s members to a build manifest with an
ordered list, a largest-file path and a total. `ledger`, level in every run,
is left out.

## Run

- Finished: 2026-09-11 06:36 UTC
- Runner: `claude` 2.1.267
- Model: `claude-sonnet-5`, reasoning effort `medium`
- Seed: `1`; `-j 5`
- Corpus: `implement`, `-tasks catalog,feed,gateway`
- Arms: `no-skill`, `reference`, `baseline`
- Repetitions: 10 per fixture and arm, 90 sessions total
- `reference`: plugin SHA-256
  `819bfd31009cb4ff84ec0c3f2df536a8ca5beee0f2c9fd8b3d38328d0fca8239`, the
  pre-retune tree measured in the 2026-09-10 three-arm run and as `reference`
  in the n=5 run
- `baseline`: the working tree after 1.8.0 with the reshaped example and the
  fewer-names line, plugin SHA-256
  `52832538f566efff400a1f2a5f1a1cf5e77877a8f2d60f08f0f86053c7d79a0b`
- Toolchain: Go 1.27.1 linux/amd64
- Raw report: [`2026-09-11-go-implement-newcode-plaincode-n10-sonnet-5-medium.json`](2026-09-11-go-implement-newcode-plaincode-n10-sonnet-5-medium.json)
  (SHA-256 `97cc8c87a247f74fce1f3287c330202efcd4846c3d1f6096a27801ac3c88a62d`)
- Session transcripts:
  [`2026-09-11-go-implement-newcode-plaincode-n10-sonnet-5-medium.traces.tar.gz`](2026-09-11-go-implement-newcode-plaincode-n10-sonnet-5-medium.traces.tar.gz)
  (SHA-256 `58d42b2eb7457d13dc5ce5095609f0540069acd7a10cf06961005a17692763c4`),
  one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $1.3475 control, $6.3828 reference, $7.6738 baseline — **4.74x** and
  **5.69x**

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <pre-retune tree> -arms no-skill,reference,baseline \
  -tasks catalog,feed,gateway -n 10 -j 5 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-11-go-implement-newcode-plaincode-n10-sonnet-5-medium.json
```

All 90 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. No shell in any arm. `go-code` fired in 25/30
reference and 26/30 baseline sessions; the misses loaded `go-http` alone on
`gateway`.

## Results

| Fixture | Arm | Golden | Valid | Δlines (valid) | Δfuncs | Δtypes | $ / run |
|---|---|---|---|---|---|---|---:|
| `catalog` | no-skill | 9/10 | 9 | 38, 23, 22, 23, 39, 20, 35, 36, 38 → 30.4 | 2 ×6, 0 ×4 | 1 ×6, 0 ×4 | 0.039 |
| `catalog` | reference | **6/10** | 6 | 19, 20, 20, 20, 19, 20 → 19.7 | 0 | 0 | 0.164 |
| `catalog` | baseline | 9/10 | 9 | 21, 20, 19, 40, 19, 19, 38, 19, 19 → 23.8 | 2 ×2, 0 ×8 | 1 ×2, 0 ×8 | 0.158 |
| `feed` | no-skill | 10/10 | 10 | 47, 47, 45, 44, 44, 45, 48, 48, 46, 47 → 46.1 | 0 | 2 ×1, 0 ×9 | 0.039 |
| `feed` | reference | 10/10 | 10 | 44, 47, 42, 45, 44, 39, 45, 46, 40, 44 → 43.6 | 0 | 2 ×1, 0 ×9 | 0.209 |
| `feed` | baseline | 10/10 | 10 | 30, 30, 35, 47, 35, 30, 42, 30, 30, 30 → **33.9** | 0 | 2 ×1, 0 ×9 | 0.239 |
| `gateway` | no-skill | **10/10** | 10 | 87, 82, 83, 92, 83, 83, 92, 89, 87, 87 → 86.5 | 1 ×3, 0 ×7 | 0 | 0.057 |
| `gateway` | reference | 5/10 | 5 | 74, 82, 73, 78, 72 → 75.8 | 2 ×4, 1 ×5, 0 ×1 | 0 | 0.265 |
| `gateway` | baseline | **4/10** | 3 | 85, 82, 78 → 81.7 | 1 ×7, 0 ×3 | 0 | 0.370 |

| Comparison | `catalog` | `feed` | `gateway` |
|---|---|---|---|
| baseline − no-skill, lines (perm p) | −6.7 (0.09) | −12.2 (< 0.01) | −4.8 (0.08) |
| baseline − reference, lines (perm p) | +4.1 (0.44) | −9.7 (< 0.01) | +5.9 (0.14) |
| reference − no-skill, lines (perm p) | −10.8 (< 0.01) | −2.5 (0.02) | −10.7 (< 0.01) |
| golden, no-skill vs baseline (Fisher) | 9/10 vs 9/10 (1.00) | 10/10 vs 10/10 (1.00) | 10/10 vs 4/10 (**0.01**) |
| golden, reference vs baseline (Fisher) | 6/10 vs 9/10 (0.30) | — | 5/10 vs 4/10 (1.00) |

| Arm | golden | valid | contract tests written · before the body · pass · fail | budget line | reports with ≥5 bullets | median report chars | skills / session | hook blocks / session | turns | $ / run |
|---|---|---|---|---|---|---:|---:|---:|---:|---:|
| `no-skill` | 29/30 | 29 | — | — | 0/30 | 298 | 0 | 0 | 8.7 | 0.0449 |
| `reference` | 21/30 | 21 | 12 · 6 · 12 · 0 | 16/30 | 3/30 | 752 | 3.77 | 0.73 | 21.0 | 0.2128 |
| `baseline` | 23/30 | 22 | 17 · 13 · 16 · 1 | 12/30 | 0/30 | 725 | 4.20 | 1.27 | 24.1 | 0.2558 |

Skill loads over thirty sessions — reference: `go-code` 25, `go-style-core`
25, `go-error-handling` 25, `go-testing` 12, `go-linting` 11, `go-http` 10,
`go-data-structures` 2, `go-security` 2, `go-defensive` 1. Baseline:
`go-code` 26, `go-style-core` 26, `go-error-handling` 26, `go-linting` 17,
`go-testing` 17, `go-http` 10, `go-security` 3, `go-code-review` 1.

## `gateway`: the n=5 reading did not hold

| Arm | Golden | `HEAD` patterns registered | Manual `r.Method` | Report says `ServeMux` answers 405 on its own | `HEAD` case in the model's own test |
|---|---|---|---|---|---|
| no-skill | 10/10 | 0/10 | 10/10 | — | — |
| reference | 5/10 | 6/10 | 0/10 | 2/10 | 0 of 4 tests |
| baseline | 4/10 | **0/10** | 4/10 | 6/10 | 0 of 7 tests |

Every failure but one is `HEAD /healthz = 200, want 405`; the other is a
reference session serving `null` for a nil list through `slices.Clone`. The
unaided control passed all ten by checking `r.Method` inside plain path
handlers, the shape it has always used; its rate in this series has been 1/5,
2/5, 4/5, 2/5 and now 10/10. The baseline arm registered no `HEAD` pattern in
ten sessions, after five of five an hour earlier on a tree that differed by
the example's shape and one bullet; the four that passed did so the control's
way. The `go-http` example was in context in all twenty skilled sessions and
was copied in six, all in the reference arm. Across the series the copy rate
under `go-code` is 13 of 30 skilled sessions, and the five-of-five was the
upper tail of that rate, not a mechanism — the n=5 report said as much and
this run is the check. The hidden clause is inferential: the documentation
lists three `GET` routes and says any other method is a 405; it never names
`HEAD`, and none of the eleven contract tests the two skilled arms wrote has a
`HEAD` case, so the Contract Table did not reach the clause either. Aggregated
over every Sonnet 5 medium session: release 1.7.0 0/19, the unaided control
20/31, the pre-retune tree 20/35, the Plain Code tree 9/16. On this fixture
the skilled trees sit at the control's long-run rate or below it, and this
run's control is above them (Fisher p = 0.01 against `baseline`). What moves
this clause is `go-http`'s Routing guidance, not `go-code`'s text: the model
either copies the `HEAD` pattern or trusts the method pattern, and the
`r.Method` check the control writes is the form neither skilled arm used.

## `catalog`: the fewer-names line, and the error type back in two

The repeated-SKU clause failed 1 of 10 in the baseline arm, 4 of 10 in the
reference arm, 1 of 10 unaided — every failure the same code, the result map
standing in for the set of SKUs already asked for. The line "fewer names, not
fewer states" is the text difference between the two skilled trees on this
clause; 1/10 against 4/10 is Fisher p = 0.30, a direction, not a result. The
reference arm's six passing sessions are the smallest code in the run, 19.7
lines with no added declaration; the baseline arm's nine average 23.8 because
two declared a `resolveError` type with `Error` and `Unwrap` (38 and 40
lines), which the retuned budget did not stop — 0 of 5 at n=5, 2 of 10 here,
6 of 10 unaided.

## `feed`: the form transfers without the fixture's shape

Baseline 33.9 lines against 43.6 and 46.1, p < 0.01 against each; six of ten
sessions at exactly 30 lines, all ten correct in every arm, no `null`. The
example no longer shares a member with the fixture, and the reading held. The
mechanism is visible in the shapes: nine of ten sessions in *both* skilled arms
used a function-local type and an anonymous document, so the 10-line gap is
not the local type. It is the inline form — the document literal initialized
once, the loop appending into the document's fields, one `json.Marshal`, no
intermediate slice, map or `kinds` variable — which is what the Plain Code
section's bullets and example show and the reference tree did not. One
baseline session declared two package-level wire types at 47 lines; one
reference and one unaided session did the same.

## The instructions, followed and not

- **Contract tests: 17 written, 13 before the body**, against 12 and 6 in the
  reference arm; 16 pass, the one failure expects `[]` where the encoder wrote
  `[]\n`. No test in either arm carries a `HEAD` case.
- **Budget line in 12 of 30 reports** against 16 of 30 for the old wording;
  reports with five or more bullets 0 against 3; median length level (725
  against 752 characters).
- **`go-linting` loaded in 17 of 30 sessions** against 11, with no shell in
  any: the routing line that exempts a shell-less session is not followed.
  With `go-testing` at 17 against 12 and hook blocks at 1.27 against 0.73,
  the baseline arm costs **20% more** per session than the reference tree
  (5.69x against 4.74x the control).

## What this changes

- **README**: a subset run does not set the cell; the Sonnet 5 new-code cell
  stays on the n=5 full-corpus run and is qualified by this one. The
  `gateway` 5/5 does not stand: 4/10 against 10/10 unaided (p = 0.01) and
  5/10 for the pre-retune tree. `feed` stands at n=10 with an example that no
  longer shares its shape. `catalog` ties the control at 9/10.
- **`gateway` is `go-http`'s finding**, not `go-code`'s: the `HEAD` example
  is copied in 13 of 30 skilled sessions and the unaided model's `r.Method`
  form passes; the skilled arms never write that form. The routing guidance
  is where a change would act.
- **The fewer-names line moved `catalog` the right way** (4/10 → 1/10 against
  the pre-retune tree) at p = 0.30; the error type returned in 2 of 10.
- **The shell-less `go-linting` exemption does not work as prose**; with the
  contract tests it is the 20% cost gap. Either a hook that sees the tool set,
  or accept the cost, or drop the line.
