# Review corpus — first runs, no-skill against baseline, Sonnet 5 medium (n=2) and Opus 5 medium (n=1)

The first measurement of `go-code-review` on a Claude model. `evals/ab/_review`
is new the same afternoon: three packages with seeded defects, a hidden key
per fixture that names each defect by the source line it sits on, and the
model's final message as the review ([corpus README](../../evals/ab/_review/README.md)).
The claude arm has no `Edit` or `Write` tool. Each run is a smoke: it
establishes whether the fixtures have room for the skill to move, not what
the skill does.

## Runs

- Runner: `claude` 2.1.267; seed `1`; corpus `review`; fixtures `orders`,
  `worker`, `invoice`; arms `no-skill`, `baseline` (the working tree, plugin
  SHA-256 `f1e875abd3f1af9a1fb90a6bd71a8b28de053823dfa5b60cae97b9095c1c3b98`)
- Key as measured: `orders` 10 defects and 2 baits, `worker` 12 and 2,
  `invoice` 7 and 1 — 29 defects. Three entries were added to the key after
  these runs from citations both arms made (below); the JSON reports carry
  the scores against the key of 29.
- Toolchain: Go 1.27.1 linux/amd64
- Sonnet 5 medium, n=2, `-j 4`, finished 2026-09-12 13:41 UTC:
  [`2026-09-12-go-review-corpus-smoke-sonnet-5-medium.json`](2026-09-12-go-review-corpus-smoke-sonnet-5-medium.json)
  (SHA-256 `c8eedf421e8d2b8d37a68b1780ab3935b34bf386f54418615ff72bf6acbe2c31`),
  transcripts [`….traces.tar.gz`](2026-09-12-go-review-corpus-smoke-sonnet-5-medium.traces.tar.gz)
  (SHA-256 `7b887d064ef1d354d64fd29219875314771fff2156a1808042af10b29a94039d`);
  cost $0.3657 no-skill, $0.8601 baseline — $0.061 against $0.143 a session
- Opus 5 medium, n=1, `-j 3`, finished 2026-09-12 13:41 UTC:
  [`2026-09-12-go-review-corpus-smoke-opus-5-medium.json`](2026-09-12-go-review-corpus-smoke-opus-5-medium.json)
  (SHA-256 `0ec9d2fb2d5821ba149c6aba86bd8bb05c40ab79ee12217bc33cf18a65c90944`),
  transcripts [`….traces.tar.gz`](2026-09-12-go-review-corpus-smoke-opus-5-medium.traces.tar.gz)
  (SHA-256 `65efff155146db9d5501615b6619dfe6c84beb2aa1591d1b718dc4a2e5ed01ca`);
  cost $0.5548 no-skill, $1.0375 baseline — $0.185 against $0.346 a session

```bash
go run ./cmd/abrun -corpus review -runner claude -model claude-sonnet-5 -effort medium \
  -arms no-skill,baseline -n 2 -j 4 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-12-go-review-corpus-smoke-sonnet-5-medium.json
go run ./cmd/abrun -corpus review -runner claude -model claude-opus-5 -effort medium \
  -arms no-skill,baseline -n 1 -j 3 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-12-go-review-corpus-smoke-opus-5-medium.json
```

All 18 sessions completed, left the fixture unchanged, and ended with a
review. `go-code-review` loaded in every baseline session and in no
`no-skill` session.

## Results

Recall is defects found over defects seeded; `must` the same over the
must-severity defects; `as-must` how many of the found must defects the
review filed under Must Fix; `read-only` recall over the lines no bundled
linter reports; `baits` correct lines flagged over baits seeded; `unkeyed`
findings on neither a defect nor a bait over all findings.

| Model | Arm | sessions | recall | must | as-must | read-only | baits | unkeyed | `verified` marks / run | $ / run |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| Sonnet 5 | `no-skill` | 6 | 0.74 | 0.80 | 0.75 | 0.70 | 0.50 | 0.17 | 0.0 | 0.061 |
| Sonnet 5 | `baseline` | 6 | **0.90** | **0.93** | **0.93** | **0.86** | 0.60 | 0.11 | 6.8 | 0.143 |
| Opus 5 | `no-skill` | 3 | **1.00** | 1.00 | 0.80 | 1.00 | 0.40 | 0.21 | 0.0 | 0.185 |
| Opus 5 | `baseline` | 3 | 0.97 | 1.00 | 0.73 | 0.95 | 0.60 | 0.21 | 8.7 | 0.346 |

### Sonnet 5, per defect (found / sessions)

| Defect | `no-skill` | `baseline` |
|---|---:|---:|
| `orders/rows-lifecycle` — no `Close`, no `Err` | 0/2 | **2/2** |
| `invoice/single-impl-interface` | 0/2 | **2/2** |
| `invoice/middle-man` | 0/2 | **2/2** |
| `worker/test-background` | 0/2 | **2/2** |
| `orders/bare-listen`, `orders/sentinel-eq`, `worker/ctx-in-struct`, `worker/wrap-with-v` | 1/2 | 2/2 |
| `worker/aliased-return` | 0/2 | 1/2 |
| `orders/500-leaks-error` | 1/2 | 1/2 |
| `worker/lock-across-call`, `worker/unbuffered-errs` | 2/2 | 1/2 |
| `invoice/total-duplicated` | 2/2 | **0/2** |
| the other 16 | 2/2 | 2/2 |

Opus 5 found 29/29 unaided and 28/29 with the skill, missing
`invoice/middle-man` once.

## Reading

- **The corpus has room on Sonnet 5 and none on Opus 5.** Unaided Sonnet 5
  found 0.74 of the seeded defects, Opus 5 all of them in every session. A
  wording comparison on Opus 5 against these three fixtures would measure
  noise; the fixture that would carry it needs defects Opus 5 misses, and
  the `unkeyed` column says where to look for them — what an unaided review
  gets wrong, not what it gets right.
- **What the skill moved on Sonnet 5, at n=2.** Recall 0.74 to 0.90, must
  recall 0.80 to 0.93, and the found must defects filed under Must Fix 0.75
  to 0.93. Four defects went from 0/2 to 2/2: the `rows` lifecycle, the
  one-implementation interface, the middle-man type, and `context.Background`
  in a test — one each from the Database, Interfaces, Less Code and Testing
  sections of the checklist, which is the checklist doing what a checklist
  does. Two went the other way, `total-duplicated` 2/2 to 0/2 and
  `unbuffered-errs` 2/2 to 1/2; n=2 cannot separate that from the draw.
- **Baits went up, not down**, 0.50 to 0.60 on Sonnet 5 and 0.40 to 0.60 on
  Opus 5. `clone-then-sort` was cited in every `invoice` review of every arm;
  the citations read as remarks on the clone, some approving, and the scorer
  cannot tell an approving citation from a finding. A bait is evidence only
  when it is a line a review would not mention unless it thought something
  was wrong; `clone-then-sort` is not that line and is the first bait to
  replace.
- **`unkeyed` is what the key missed, not what the review invented.** The
  unkeyed citations both arms made on both models are real: `New(ctx, 0)`
  leaves `Run` blocked on a semaphore with no capacity; `Total float64` for
  money; `return id, tx.Commit()` hands the caller an id beside a commit
  error. All three are in the key now (`worker/workers-unvalidated`,
  `orders/float-money`, `orders/id-with-commit-error`), which takes the
  corpus to 32 defects; the ones not added — no `LIMIT` on `Search`, no
  validation in `Create`, `%-30s` misaligning a long description — are
  judgment calls a fixture should not grade.
- **The template is followed.** Every baseline review carried `verified` and
  `plausible` markers (6.8 and 8.7 `verified` a session), no unaided review
  did, and the Sonnet 5 skilled reviews put must defects under Must Fix where
  the unaided ones spread them over Should Fix and Nit.
- **Cost 2.3x on Sonnet 5, 1.9x on Opus 5**, a review being cheap to begin
  with: $0.06 to $0.14 and $0.19 to $0.35 a session.

## What this changes

- `go-code-review` has a measurement. Its Sonnet 5 reading is a smoke at
  n=2: the direction is clean on 29 defects and the per-defect table names
  what moved, and an n=5 on the same three fixtures is the next run before
  any wording edit is judged against it.
- On Opus 5 the corpus is saturated. The next fixture is built from what the
  unaided Opus 5 review gets wrong, or the skill is not measurable there.
- Scoring change after these runs: only the first citation on a line is a
  finding; a later one is a supporting reference that counts toward the
  defect it lands on and is never `unkeyed`. On this run's outputs the
  supporting references were `store.go:34` ("Get wraps the sentinel there")
  and the like.
