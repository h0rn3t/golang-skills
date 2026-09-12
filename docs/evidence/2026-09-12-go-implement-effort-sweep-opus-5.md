# Effort Sweep — `catalog`, `feed`, `gateway`, no skills against baseline, Opus 5 low and high (n=5)

Every Opus 5 cell in the README was measured at `--effort medium`, and
Anthropic's Claude Opus 5 guidance says effort defaults do not carry over
from earlier models and that `low` and `medium` are unusually strong on this
one. This sweep runs the unaided control and the skilled `baseline` arm at
`low` and at `high` on the three admitted single-function fixtures, five
repetitions each, the same day and tree as the
[`gateway` closure pair](2026-09-12-go-implement-budget-closure-gateway-n5-opus-5-medium.md).
The medium reading it sits between is the
[2026-09-11 three-arm run at n=3](2026-09-11-go-implement-newcode-workflow-opus-5-medium.md).

`baseline` is the working tree at `2fab3d6` (release 1.12.0) plus the
2026-09-12 edits: the Declaration Budget closure rule, the `go-code`
reorder, the `go-error-handling` and `go-naming` cuts. There is no
`reference` arm; the question here is what the skills add at each effort,
not what the day's edits changed.

## Runs

- Finished: 2026-09-12 08:41 UTC (`low`), 08:53 UTC (`high`)
- Runner: `claude` 2.1.267
- Model: `claude-opus-5`; reasoning effort `low` in one run, `high` in the other
- Seed: `1`; `-j 4`
- Corpus: `implement`; fixtures: `catalog`, `feed`, `gateway`
- Arms: `no-skill`, `baseline`; 5 repetitions per fixture and arm, 30 sessions per run
- `baseline` plugin SHA-256 `ce495ac2088148f18754c0f99e60095443491440344f8551bcc4e4657416fbf6` in both runs
- Toolchain: Go 1.27.1 linux/amd64; golangci-lint 2.13.2 for the lint line
- `low`: [`2026-09-12-go-implement-effort-sweep-low-opus-5.json`](2026-09-12-go-implement-effort-sweep-low-opus-5.json)
  (SHA-256 `597768264f67bc965a499df3a2e0798a22950f8bf71b2f62f11bf97a3cb205cc`),
  transcripts [`2026-09-12-go-implement-effort-sweep-low-opus-5.traces.tar.gz`](2026-09-12-go-implement-effort-sweep-low-opus-5.traces.tar.gz)
  (SHA-256 `6d2b82067dfae16b7c2ecd7bd0f0a52016d42046ce191b84734758564b2d3d58`);
  cost $1.8165 control, $10.3531 baseline — **5.70x**
- `high`: [`2026-09-12-go-implement-effort-sweep-high-opus-5.json`](2026-09-12-go-implement-effort-sweep-high-opus-5.json)
  (SHA-256 `5a8e3cc8b6cc4db5838cfe91b995b1b22afd95407234180bd6d1eb5563b0bfeb`),
  transcripts [`2026-09-12-go-implement-effort-sweep-high-opus-5.traces.tar.gz`](2026-09-12-go-implement-effort-sweep-high-opus-5.traces.tar.gz)
  (SHA-256 `b13c1c8a163166cbc708dc5afb1c94babf1df2055e9d8bf1c9b657ad0b0e191b`);
  cost $3.8426 control, $14.8468 baseline — **3.86x**

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-opus-5 -effort low \
  -arms no-skill,baseline -tasks catalog,feed,gateway -n 5 -j 4 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-12-go-implement-effort-sweep-low-opus-5.json
go run ./cmd/abrun -corpus implement -runner claude -model claude-opus-5 -effort high \
  -arms no-skill,baseline -tasks catalog,feed,gateway -n 5 -j 4 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-12-go-implement-effort-sweep-high-opus-5.json
```

All 60 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. The tool set had no shell. `go-code` fired
first in 30/30 baseline sessions.

## Results

Lines are over valid runs (golden, build, and the model's own tests passing);
p is an exact permutation test on the mean, Fisher's exact test for golden.

### `low`

| Fixture | golden no-skill | golden baseline | Fisher p | Δlines no-skill | Δlines baseline | diff | perm p | lint after, mean · clean | $ / run |
|---|---|---|---:|---|---|---:|---:|---|---|
| `catalog` | 5/5 | 5/5 | 1.00 | 23, 21, 22, 23, 23 → 22.4 | 20, 20, 22, 22, 22 → 21.2 | −1.2 | 0.175 | 0.00 · 5/5 / 0.00 · 5/5 | 0.064 / 0.554 |
| `feed` | 5/5 | 5/5 | 1.00 | 47, 49, 45, 46, 46 → 46.6 | 51, 43, 52, 50, 51 → 49.4 | +2.8 | 0.183 | 0.00 · 5/5 / 0.00 · 5/5 | 0.093 / 0.604 |
| `gateway` | **1/5** | 5/5 | 0.05 | 131 (the one pass; 94, 89, 96, 124 failed) | 86, 97, 81, 77, 77 → 83.6 | −47.4 | 0.167 | 4.00 · 0/1 / 1.20 · 2/5 | 0.207 / 0.913 |

### `high`

| Fixture | golden no-skill | golden baseline | Fisher p | Δlines no-skill | Δlines baseline | diff | perm p | lint after, mean · clean | $ / run |
|---|---|---|---:|---|---|---:|---:|---|---|
| `catalog` | 5/5 | 5/5 | 1.00 | 24, 24, 31, 25, 27 → 26.2 | 23, 22, 21, 22, 22 → 22.0 | **−4.2** | **0.008** | 0.00 · 5/5 / 0.00 · 5/5 | 0.150 / 0.633 |
| `feed` | 5/5 | 5/5 | 1.00 | 47, 50, 50, 48, 53 → 49.6 | 45, 35, 43, 47, 46 → 43.2 | **−6.4** | **0.016** | 0.00 · 5/5 / 0.00 · 5/5 | 0.229 / 0.773 |
| `gateway` | **1/5** | 5/5 | 0.05 | 159 (the one pass; 137, 130, 124, 121 failed) | 88, 100, 89, 94, 85 → 91.2 | −67.8 | 0.167 | 3.00 · 0/1 / 0.40 · 4/5 | 0.390 / 1.564 |

### Arms across efforts

| Effort | Arm | golden | Δlines, valid | Δfuncs | Δclos | lint after, mean | lint clean | `Skill` msgs | API calls | cache writes | cache reads | output tokens | thinking tokens | median report chars | $ / run |
|---|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| low | `no-skill` | 11/15 | 43.3 | 0.36 | 0.00 | 0.36 | 10/11 | 0.0 | 7.5 | 3.9K | 53K | 2197 | 330 | 700 | 0.121 |
| low | `baseline` | 15/15 | 51.4 | 0.47 | 0.00 | 0.40 | 12/15 | 2.0 | 10.0 | 37.5K | 301K | 6539 | 1499 | 1030 | 0.690 |
| medium, 2026-09-11, n=3 | `no-skill` | 7/9 | — | — | — | — | — | 0.0 | — | — | — | — | — | 1578 | 0.205 |
| medium, 2026-09-11, n=3 | `baseline` | 9/9 | — | — | — | — | — | — | — | — | — | — | — | 1151 | 0.723 |
| high | `no-skill` | 11/15 | 48.9 | 0.64 | 0.00 | 0.27 | 10/11 | 0.0 | 9.3 | 7.5K | 73K | 5767 | 2230 | 1980 | 0.256 |
| high | `baseline` | 15/15 | 52.1 | 0.67 | 0.00 | 0.13 | 14/15 | 2.1 | 9.1 | 47.2K | 314K | 14380 | 9193 | 1522 | 0.990 |

Every unaided `gateway` failure at both efforts is the same clause: `HEAD`
answered with 200 on all three routes. Every baseline `gateway` session
wrote two package-level helpers and reported `added package-level
declarations: 2` with both call sites, at both efforts; `Δclos` is zero in
all 60 sessions. One baseline `catalog` session at `high` omitted the budget
line; the other 29 carried it.

## Reading

- **Correctness does not move with effort; the skills do.** Unaided,
  `gateway` passes 1/5 at `low`, 1/3 at `medium`, 1/5 at `high` — the `HEAD`
  default leaks through a `GET` pattern however long the model thinks. With
  the skills it is 5/5, 3/3, 5/5. `catalog` and `feed` are 5/5 everywhere.
  Fisher p = 0.05 per run on `gateway`, the same 1/5 against 5/5 twice.
- **The size effect is an effort effect.** At `high` the unaided model
  writes more — `catalog` 26.2, `feed` 49.6, `gateway` 121–159 lines with
  three to seven helpers — and the skilled arm is smaller on every fixture,
  `catalog` −4.2 (p = 0.008) and `feed` −6.4 (p = 0.016) surviving a
  Bonferroni correction for three comparisons. At `low` the unaided model is
  already compact — `catalog` 22.4, `feed` 46.6, 700 characters of report —
  and the skills buy nothing in lines (−1.2 and +2.8, neither near
  significance). The guidance's "low is unusually strong" holds for the
  shape of the code; it does not hold for the class clause.
- **Cost.** The skilled arm costs $0.69 a session at `low` and $0.99 at
  `high`; the control $0.12 and $0.26. The multiplier is therefore worst at
  `low` (5.7x) and best at `high` (3.9x), because the skill text is a fixed
  cache-write cost of about 37–47K tokens a session while the model's own
  spend scales with effort. `Skill` messages 2.0 and 2.1 a session: the
  batching edit holds at both efforts.
- **Lint.** The unaided `gateway` sessions carry two to four findings at
  both efforts (`errcheck` on `w.Write`, and at `low` an unused parameter or
  a `Sprintf` where `strconv` serves). The skilled arm is clean in 4/5 at
  `high` and 2/5 at `low`; the write goes unchecked more often when the
  model thinks less, and without a shell nothing in the session says so.
- **What the sweep does not say.** It has no `medium` control of its own at
  n=5 and no `reference` arm, so it cannot rank the three efforts against
  one another on one day's model, and it says nothing about the 2026-09-12
  edits in isolation; the closure pair does that.

## What this changes

- The README's Opus 5 new-code cell gains its effort qualifiers: at `low`
  the skills are a correctness effect only, at 5.7x; at `high` they are a
  correctness and a size effect, at 3.9x. `medium` stays the recommended
  setting for the plugin: it is where every cell was measured and it sits
  between the two on cost.
- "A `low` control has not been run" comes out of the README.
