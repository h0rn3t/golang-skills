# Eleven checklist rows cut and seven restated — Sonnet 5 medium, six review fixtures at n=2

`go-code-review` step 3 says "never spend review attention on what a tool
reports", and then the checklist carried eleven rows that restate what the
bundled linters report or what the model writes unprompted: comment
sentences and doc comments (`revive exported`, `godot`), error-string case
(`staticcheck` ST1005), MixedCaps, initialisms, receiver names, `any` over
`interface{}` (`modernize`), `crypto/rand` (`gosec` G404), import groups, blank
and dot imports. The same edit restated seven rows as the action to take
instead of the thing to avoid ("Every error is handled", "Built-in names stay
free", "Errors over panics", …), and added one sentence to step 3 naming what
the checklist no longer carries. This run measures the two together against
release 1.20.1 on the review corpus, on the model where the corpus has room
(Opus 5 medium found 34/35 seeded defects unaided on 2026-09-12).

| Arm | Tree |
|---|---|
| `reference` | `git worktree` of `89d702a` (release 1.20.1): 226-line checklist, 14 `Never`/`Do not`/`Avoid` rules |
| `baseline` | the working tree (plugin SHA-256 `11eec293…`): `go-code-review` at 221 lines with the eleven rows gone, seven restated, step 3 extended; the rest of the day's edits ride along but only `go-code-review` loaded (`skills=[go-code-review]` in 24/24 sessions) |

## Run

- Finished: 2026-09-18 17:06 UTC
- Runner: `claude` 2.1.267; model `claude-sonnet-5`, effort `medium`; seed `1`
- Corpus `review`, fixtures `books`, `invoice`, `orders`, `partner`, `vault`, `worker`, arms `reference`, `baseline`, 2 repetitions per fixture and arm, 24 sessions, `-j 2`, 0 CLI errors
- `reference` plugin SHA-256 `d478dd31988660fcb8934eba8e56e7e378221b321b278073006d4ecb4af19ca0`; `baseline` `11eec2930c89cb5ccffc6beca704b3c89c64c03a64acefab572cfb1017a62c3c`
- Toolchain Go 1.27.1 linux/amd64, golangci-lint 2.13.2 on PATH; the review arm has no `Edit`, `Write` or shell tool, so `pre-review.sh` did not run and the model read the code
- Report: [`2026-09-18-go-review-checklist-cut-n2-sonnet-5-medium.json`](2026-09-18-go-review-checklist-cut-n2-sonnet-5-medium.json) (SHA-256 `1d6d446f8001bd456f209d7dd16732b06b7dd3cce752c28ab82d4dc7b01badf3`); traces [`….traces.tar.gz`](2026-09-18-go-review-checklist-cut-n2-sonnet-5-medium.traces.tar.gz) (SHA-256 `20653b4d300847ccf9c08caafd86b171006a440ac233f99a0c8934f826cafaf8`), one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $6.25

```bash
go run ./cmd/abrun -corpus review -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <worktree of 89d702a> -arms reference,baseline \
  -n 2 -j 2 -seed 1 -timeout 10m -keep -verbose \
  -out ../docs/evidence/2026-09-18-go-review-checklist-cut-n2-sonnet-5-medium.json
```

Scoring is the corpus's: recall over the hidden key by source line, must
defects, the split between lines a bundled linter reports and lines no tool
reports, bait lines flagged, and citations landing on no key entry.

## The reading

| Arm | Sessions | Recall | Must | Read-only recall | Tool-reported recall | Baits hit | Unkeyed / citations | Assistant turns/s | $/session |
|---|---|---|---|---|---|---|---|---|---|
| `reference` | 12 | 110/152 (0.72) | 51/62 (0.82) | 98/136 (0.72) | 12/16 (0.75) | 8/26 (0.31) | 25/141 (0.18) | 6.0 | 0.262 |
| `baseline` | 12 | **121/152 (0.80)** | 54/62 (0.87) | 106/136 (0.78) | 15/16 (0.94) | **5/26 (0.19)** | **15/144 (0.10)** | 5.5 | 0.259 |

Per fixture (both repetitions summed):

| Fixture | `reference` | `baseline` |
|---|---|---|
| `books` | 24/28 (0.86) | 24/28 (0.86) |
| `invoice` | 11/14 (0.79) | 13/14 (0.93) |
| `orders` | 17/24 (0.71) | 21/24 (0.88) |
| `partner` | 19/34 (0.56) | 18/34 (0.53) |
| `vault` | 17/26 (0.65) | 22/26 (0.85) |
| `worker` | 22/26 (0.85) | 23/26 (0.88) |

Per session:

| Arm | Fixture | Rep | Found | Must | Read-only | Tool-reported | Baits | Unkeyed / citations | $ |
|---|---|---|---|---|---|---|---|---|---|
| `reference` | `books` | 0 | 12/14 | 5/5 | 12/14 | 0/0 | 1/3 | 5/16 | 0.229 |
| `reference` | `books` | 1 | 12/14 | 5/5 | 12/14 | 0/0 | 1/3 | 2/13 | 0.249 |
| `reference` | `invoice` | 0 | 6/7 | 1/1 | 4/5 | 2/2 | 1/1 | 3/10 | 0.181 |
| `reference` | `invoice` | 1 | 5/7 | 1/1 | 3/5 | 2/2 | 0/1 | 2/8 | 0.145 |
| `reference` | `orders` | 0 | 9/12 | 6/8 | 5/8 | 4/4 | 2/2 | 1/12 | 0.212 |
| `reference` | `orders` | 1 | 8/12 | 7/8 | 5/8 | 3/4 | 0/2 | 0/8 | 0.173 |
| `reference` | `partner` | 0 | 10/17 | 6/8 | 10/17 | 0/0 | 0/2 | 3/12 | 0.418 |
| `reference` | `partner` | 1 | 9/17 | 5/8 | 9/17 | 0/0 | 0/2 | 2/11 | 0.367 |
| `reference` | `vault` | 0 | 8/13 | 2/3 | 8/12 | 0/1 | 0/3 | 4/13 | 0.399 |
| `reference` | `vault` | 1 | 9/13 | 2/3 | 9/12 | 0/1 | 1/3 | 2/11 | 0.362 |
| `reference` | `worker` | 0 | 11/13 | 6/6 | 10/12 | 1/1 | 1/2 | 0/14 | 0.247 |
| `reference` | `worker` | 1 | 11/13 | 5/6 | 11/12 | 0/1 | 1/2 | 1/13 | 0.158 |
| `baseline` | `books` | 0 | 13/14 | 5/5 | 13/14 | 0/0 | 1/3 | 0/14 | 0.240 |
| `baseline` | `books` | 1 | 11/14 | 4/5 | 11/14 | 0/0 | 0/3 | 3/14 | 0.248 |
| `baseline` | `invoice` | 0 | 7/7 | 1/1 | 5/5 | 2/2 | 0/1 | 0/9 | 0.149 |
| `baseline` | `invoice` | 1 | 6/7 | 1/1 | 4/5 | 2/2 | 0/1 | 1/8 | 0.154 |
| `baseline` | `orders` | 0 | 10/12 | 7/8 | 7/8 | 3/4 | 0/2 | 1/14 | 0.149 |
| `baseline` | `orders` | 1 | 11/12 | 8/8 | 7/8 | 4/4 | 0/2 | 1/15 | 0.195 |
| `baseline` | `partner` | 0 | 9/17 | 6/8 | 9/17 | 0/0 | 0/2 | 3/11 | 0.377 |
| `baseline` | `partner` | 1 | 9/17 | 6/8 | 9/17 | 0/0 | 1/2 | 1/11 | 0.369 |
| `baseline` | `vault` | 0 | 10/13 | 2/3 | 9/12 | 1/1 | 1/3 | 2/12 | 0.434 |
| `baseline` | `vault` | 1 | 12/13 | 3/3 | 11/12 | 1/1 | 0/3 | 3/13 | 0.380 |
| `baseline` | `worker` | 0 | 10/13 | 5/6 | 9/12 | 1/1 | 1/2 | 0/10 | 0.215 |
| `baseline` | `worker` | 1 | 13/13 | 6/6 | 12/12 | 1/1 | 1/2 | 0/13 | 0.198 |

Defects the arms missed differently (count of sessions out of 2 that missed the line):

| Defect | Owner | `reference` missed | `baseline` missed |
|---|---|---|---|
| `aliased-return` | go-defensive | 2 | 0 |
| `bytes-not-runes` | go-code-review | 2 | 0 |
| `middle-man` | go-code-refactor | 2 | 0 |
| `open-redirect-double-slash` | go-security | 2 | 0 |
| `quota-never-released` | go-code-review | 2 | 0 |
| `500-leaks-error`, `base-without-host`, `flush-error-lost`, `idempotency-key-per-attempt`, `math-rand-token`, `rows-lifecycle`, `unbounded-body`, `workers-unvalidated` | various | 1 each | 0 |
| `id-with-commit-error`, `uploader-content-type-served` | go-database, go-http | 2 | 1 |
| `breaker-probe-herd`, `lock-across-call`, `retry-after-nanoseconds`, `self-transfer-mints`, `test-background`, `wrap-with-v` | various | 0 | 1 each |
| `receipt-alphabet-mismatch`, `test-enshrines-refund-fee`, `throttle-burns-attempt` | various | 1 | 2 |
| `error-trailing-colon`, `limiter-unvalidated`, `test-any-error`, `test-error-string`, `ticker-panics-on-zero` | various | 2 | 2 |

**Recall rose on every column the corpus scores, at level cost.** 121
against 110 of 152 key lines, must 54 against 51 of 62, read-only 106
against 98 of 136, and the sixteen lines a bundled linter reports — the
category the cut rows named — 15 against 12. The review that says less about
naming case and doc-comment form found more `math/rand` tokens and `rows`
lifecycles, not fewer. Three fixtures moved (`invoice` +2, `orders` +4,
`vault` +5), two held (`books`, `worker` +1), `partner` gave one back
(18 against 19 of 34). Cost $0.259 against $0.262 a session, 5.5 assistant
turns against 6.0.

**Precision moved with it.** Bait lines — correct code a weak review flags —
5 against 8 of 26, and citations on no key line 15 against 25 of about 140.
Fewer rows to walk, fewer lines cited for walking them.

**Where the reading is thin.** Two repetitions a fixture; `partner`, the
fixture with the most room (17 defects, recall near 0.55 in both arms), did
not move. Six defects were missed once by the baseline and never by the
reference, and five the other way round, so a good share of the per-defect
table is the coin the corpus flips at n=2. The arm carries the eleven cuts
and the seven restatements together; nothing here says which of the two did
the moving, or whether the moving is anything but the shorter file.

## What this establishes

- Cutting the rows the linters and the model already cover did not lower
  recall on the lines they covered — it rose, 12 to 15 of 16 — and recall on
  the lines no tool reports rose with it, 0.72 to 0.78, over 152 key lines.
- The Sonnet 5 review corpus keeps its room (0.80), and `partner` is where
  it is.
- The "cut what the model already knows" rule of the authoring template has
  its first review-corpus reading, in the direction the template predicts.

## Next

- `partner` at n≥4 in both arms is the fixture that would confirm or undo
  the per-fixture spread; the other five are near their ceiling for this
  model.
- Split the arm once: the eleven cuts alone against this baseline, to
  attribute the gain between cut and restatement.
