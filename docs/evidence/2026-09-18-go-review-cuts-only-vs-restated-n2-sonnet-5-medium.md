# The eleven cuts alone against the cuts with the seven restatements — Sonnet 5 medium, six review fixtures at n=2

The [afternoon's review run](2026-09-18-go-review-checklist-cut-n2-sonnet-5-medium.md)
measured one arm that carried two edits to `go-code-review`: eleven
checklist rows that restate what the bundled linters report were cut, and
seven rows were restated as the action to take ("Every error is handled",
"Built-in names stay free", "Interfaces where they are consumed", …). Recall
rose 0.72 → 0.80 over 152 key lines and the report's "Next" asked for the
split. This run is the split: a tree with the eleven cuts and the step-3
sentence but the seven rows in their 1.20.1 wording, against the 1.21.0
branch head that has both.

| Arm | Tree |
|---|---|
| `reference` (cuts only) | `git worktree` of `66f06cb` with the seven restated rows of `go-code-review/SKILL.md` returned to their 1.20.1 text — "Handle errors", "Avoid built-in names", "Contexts: First parameter; not in structs …", "No premature interfaces", "Copying: Be careful …", "Don't panic", "Pass values: Don't use pointers …" — 7 lines changed, nothing else; plugin SHA-256 `a2c6cecf2ea6daf76ccdf5a2d57a65477da122f4d2d153be64a54b2771f6129a` |
| `baseline` (cuts and restatements) | `git worktree` of `66f06cb` (the 1.21.0 branch head), plugin SHA-256 `bfcfed52ecd8273c265ff05738e65a40716a3caa9b16a78dde590c732cd14ef0`; `go-code-review` byte-identical to the afternoon's `baseline` |

## Run

- Finished: 2026-09-18 19:25 UTC
- Runner: `claude` 2.1.267; model `claude-sonnet-5`, effort `medium`; seed `2`
- Corpus `review`, fixtures `books`, `invoice`, `orders`, `partner`, `vault`, `worker`, arms `reference`, `baseline`, 2 repetitions per fixture and arm, 24 sessions, `-j 3`, 0 CLI errors; `skills=[go-code-review]` in 24/24
- Toolchain Go 1.27.1 linux/amd64, golangci-lint 2.13.2 on PATH; the review arm has no `Edit`, `Write` or shell tool, so `pre-review.sh` did not run and the model read the code
- Report: [`2026-09-18-go-review-cuts-only-vs-restated-n2-sonnet-5-medium.json`](2026-09-18-go-review-cuts-only-vs-restated-n2-sonnet-5-medium.json) (SHA-256 `3436f46947645119ad8a76c3bd9334129e9a827a0a1954aaa415cfb23512c81c`); traces [`….traces.tar.gz`](2026-09-18-go-review-cuts-only-vs-restated-n2-sonnet-5-medium.traces.tar.gz) (SHA-256 `d640c5a82b23f81b5e4f6eb720ef81cdddea546028e9e90f1404fc888c4462cd`), one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $6.20

```bash
go run ./cmd/abrun -corpus review -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <worktree of 66f06cb with the seven rows reverted> -arms reference,baseline \
  -n 2 -j 3 -seed 2 -timeout 10m -keep -verbose -out <json>
```

Scoring is the corpus's, as in the afternoon's report.

## The reading

| Arm | Sessions | Recall | Must | As-must | Read-only recall | Tool-reported recall | Baits hit | Unkeyed / citations | Assistant turns/s | $/session |
|---|---|---|---|---|---|---|---|---|---|---|
| `reference` (cuts only) | 12 | 117/152 (0.77) | 57/62 (0.92) | 50/57 | 102/136 (0.75) | 15/16 | 9/26 (0.35) | 19/142 (0.13) | 5.7 | 0.258 |
| `baseline` (cuts and restatements) | 12 | 116/152 (0.76) | 55/62 (0.89) | 49/55 | 103/136 (0.76) | 13/16 | 7/26 (0.27) | 13/134 (0.10) | 6.3 | 0.258 |
| afternoon, 1.20.1 (for scale) | 12 | 110/152 (0.72) | 51/62 (0.82) | 43/51 | 98/136 (0.72) | 12/16 | 8/26 (0.31) | 25/141 (0.18) | 6.0 | 0.262 |
| afternoon, cuts and restatements | 12 | 121/152 (0.80) | 54/62 (0.87) | 48/54 | 106/136 (0.78) | 15/16 | 5/26 (0.19) | 15/144 (0.10) | 5.5 | 0.259 |

Per fixture (both repetitions summed):

| Fixture | `reference` (cuts only) | `baseline` (cuts and restatements) |
|---|---|---|
| `books` | 26/28 (0.93) | 25/28 (0.89) |
| `invoice` | 13/14 (0.93) | 10/14 (0.71) |
| `orders` | 20/24 (0.83) | 21/24 (0.88) |
| `partner` | 21/34 (0.62) | 23/34 (0.68) |
| `vault` | 18/26 (0.69) | 16/26 (0.62) |
| `worker` | 19/26 (0.73) | 21/26 (0.81) |

Per session:

| Arm | Fixture | Rep | Found | Must | Read-only | Tool-reported | Baits | Unkeyed / citations | $ |
|---|---|---|---|---|---|---|---|---|---|
| `reference` | `books` | 0 | 13/14 | 5/5 | 13/14 | 0/0 | 1/3 | 2/13 | 0.253 |
| `reference` | `books` | 1 | 13/14 | 4/5 | 13/14 | 0/0 | 0/3 | 1/15 | 0.272 |
| `reference` | `invoice` | 0 | 7/7 | 1/1 | 5/5 | 2/2 | 1/1 | 3/9 | 0.152 |
| `reference` | `invoice` | 1 | 6/7 | 1/1 | 4/5 | 2/2 | 1/1 | 1/8 | 0.122 |
| `reference` | `orders` | 0 | 10/12 | 7/8 | 7/8 | 3/4 | 1/2 | 0/12 | 0.188 |
| `reference` | `orders` | 1 | 10/12 | 8/8 | 6/8 | 4/4 | 1/2 | 0/12 | 0.227 |
| `reference` | `partner` | 0 | 10/17 | 7/8 | 10/17 | 0/0 | 0/2 | 1/11 | 0.318 |
| `reference` | `partner` | 1 | 11/17 | 8/8 | 11/17 | 0/0 | 0/2 | 2/12 | 0.457 |
| `reference` | `vault` | 0 | 7/13 | 1/3 | 6/12 | 1/1 | 1/3 | 5/13 | 0.301 |
| `reference` | `vault` | 1 | 11/13 | 3/3 | 10/12 | 1/1 | 1/3 | 2/14 | 0.422 |
| `reference` | `worker` | 0 | 9/13 | 6/6 | 8/12 | 1/1 | 1/2 | 1/11 | 0.182 |
| `reference` | `worker` | 1 | 10/13 | 6/6 | 9/12 | 1/1 | 1/2 | 1/12 | 0.203 |
| `baseline` | `books` | 0 | 12/14 | 4/5 | 12/14 | 0/0 | 1/3 | 1/11 | 0.251 |
| `baseline` | `books` | 1 | 13/14 | 5/5 | 13/14 | 0/0 | 0/3 | 1/17 | 0.272 |
| `baseline` | `invoice` | 0 | 5/7 | 1/1 | 3/5 | 2/2 | 0/1 | 3/8 | 0.198 |
| `baseline` | `invoice` | 1 | 5/7 | 1/1 | 3/5 | 2/2 | 0/1 | 2/8 | 0.153 |
| `baseline` | `orders` | 0 | 10/12 | 8/8 | 6/8 | 4/4 | 2/2 | 1/14 | 0.183 |
| `baseline` | `orders` | 1 | 11/12 | 7/8 | 8/8 | 3/4 | 1/2 | 0/11 | 0.136 |
| `baseline` | `partner` | 0 | 11/17 | 8/8 | 11/17 | 0/0 | 0/2 | 2/11 | 0.360 |
| `baseline` | `partner` | 1 | 12/17 | 7/8 | 12/17 | 0/0 | 0/2 | 0/14 | 0.368 |
| `baseline` | `vault` | 0 | 8/13 | 1/3 | 8/12 | 0/1 | 1/3 | 2/10 | 0.382 |
| `baseline` | `vault` | 1 | 8/13 | 1/3 | 8/12 | 0/1 | 0/3 | 1/8 | 0.398 |
| `baseline` | `worker` | 0 | 10/13 | 6/6 | 9/12 | 1/1 | 1/2 | 0/11 | 0.222 |
| `baseline` | `worker` | 1 | 11/13 | 6/6 | 10/12 | 1/1 | 1/2 | 0/11 | 0.179 |

Defects the arms missed differently (sessions out of 2 that missed the line):

| Defect | `reference` (cuts only) missed | `baseline` (cuts and restatements) missed |
|---|---|---|
| `invoice/middle-man`, `vault/open-redirect-double-slash`, `vault/test-error-string` | 0 | 2 |
| `partner/base-without-host`, `vault/etag-map-order`, `worker/workers-unvalidated` | 2 | 0 |
| `books/test-enshrines-refund-fee`, `books/time-as-map-key`, `worker/sleep-ignores-ctx`, `invoice/total-duplicated`, `partner/breaker-probe-herd`, `vault/not-modified-without-etag`, `vault/pool-alias-after-put`, `vault/uploader-content-type-served` | 0 or 1 | one more than `reference` |
| `books/bytes-not-runes`, `orders/id-with-commit-error`, `partner/retry-after-nanoseconds`, `vault/partial-file-visible`, `vault/quota-check-then-act`, `vault/quota-never-released`, `worker/aliased-return` | one more than `baseline` | 0 or 1 |

**The two arms are one line apart on every column.** Recall 117 against 116
of 152, must 57 against 55 of 62, read-only 102 against 103 of 136, cost
$0.258 against $0.258 a session. The restatements' one visible column is the
sixteen linter-reported lines, 13/16 with them against 15/16 without — the
opposite of the afternoon (15/16 with the restatements against 12/16 for
1.20.1), which says that column moves with the seed, not the wording. Baits
9 against 7 and unkeyed citations 19 against 13 of about 140 lean toward
the restatements by two and six citations across twelve sessions.

**Per fixture the moves cancel.** `invoice` +3 for the cuts-only arm (the
`middle-man` type missed in 2/2 `baseline` sessions, found 2/2 without the
restatements — the "Interfaces where they are consumed" row is the one that
took over the "No premature interfaces" wording, and `invoice` is the
fixture built around a one-implementation interface and a middle man),
`vault` +2, `books` +1; `partner` −2, `worker` −2, `orders` −1. Twenty-one
defects were missed a different number of times by the two arms, ten one
way and eleven the other; at n=2 that is the corpus's coin.

**Against the afternoon, both arms sit where the cut arm sat.** The
cuts-only arm at 0.77 and the cuts-and-restatements arm at 0.76 are both
above the 1.20.1 arm's 0.72 and below the same tree's 0.80 of the afternoon.
The same bytes scored 0.80 at seed 1 and 0.76 at seed 2: that ±0.04 is the
run-to-run movement of this corpus on this model at n=2, and the
afternoon's +0.08 sits about two of those above its reference.

## What this establishes

- The afternoon's gain is the eleven cuts, not the seven restatements: the
  arm without the restatements reaches the same recall (0.77 against 0.76),
  the same must recall (0.92 against 0.89) and the same cost.
- The restatements are neutral on this corpus at n=2 — no column moves by
  more than the seed does — with one fixture-level reading against them
  (`invoice/middle-man` 0/2 with the consumer-side interface wording, 2/2
  with the 1.20.1 "No premature interfaces"), which is a reason to watch that
  row, not yet to revert it.
- The "cut what the model already knows" rule has its attribution: cutting
  moved recall, rewording did not.

## Next

- `invoice` at n≥5 on the two wordings of the interface row, the one place
  the restatement has a candidate effect.
- The review corpus's own noise floor is now on record (±0.04 recall for the
  same bytes across seeds); any future wording claim on it needs n≥4 or a
  per-defect count that survives it.
