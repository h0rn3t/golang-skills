# Implementation Corpus — `feed` and `catalog` on Opus 5

Both fixtures were rebuilt to pin a single entry point, the change that made
[`gateway`](2026-09-07-go-implement-gateway-opus5.md) produce a result. It did
not work here, and the reason is worth recording: freedom over the internals is
not the same thing as room to over-engineer.

## Run

- Finished: 2026-09-07 12:48 EEST
- Corpus: `implement` (`evals/ab/_implement`)
- Runner: `claude`, Claude Code 2.1.261
- Model: `claude-opus-5`
- Seed: `20260907`
- Fixtures: `feed`, `catalog`
- Arms: `no-skill`, `baseline`
- Repetitions: 3 per fixture and arm, 12 sessions total
- Baseline plugin SHA-256: `6443a658f74cd465b98f0065cdb7c117bcde52c64af1acc1a5ef7f3a67369d3c`
- Raw report: [`2026-09-07-go-implement-feed-catalog-opus5.json`](2026-09-07-go-implement-feed-catalog-opus5.json)
- Raw report SHA-256: `3a9ceaa83daf606d6ca559a97a4a682d10e905dc94fe7a18b5624e656f0612fa`
- Total cost: $3.27 — $0.86 control, $2.41 skilled

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-opus-5 \
  -tasks feed,catalog -arms no-skill,baseline -n 3 -j 4 -seed 20260907 -keep \
  -out ../docs/evidence/2026-09-07-go-implement-feed-catalog-opus5.json
```

All 12 sessions completed, built and passed the hidden specification. `go-code`
loaded in all 6 baseline sessions and no other skill was recorded.

## Results

| Fixture | Arm | Lines | Per-run | Functions | Branches | Types |
|---|---|---:|---|---:|---:|---:|
| `catalog` | no skill | 26.0 | 24, 24, 30 | 0 | 4.33 | 0 |
| `catalog` | skill | 20.3 | 19, 20, 22 | 0 | 4.00 | 0 |
| `feed` | no skill | 50.0 | 50, 50, 50 | 0 | 4.00 | 2.00 |
| `feed` | skill | 49.0 | 48, 48, 51 | 0 | 4.00 | 2.00 |

`catalog` is −5.7 lines (−21.8%) with the ranges separating, 24/24/30 against
19/20/22, but its Welch interval is −12.6 to +1.3 and includes zero. It is
directional at `n=3` and unresolved.

`feed` is −1.0 lines and measures nothing. The control wrote **exactly 50
lines in all three runs**, with the same two types and the same four branches
in every session of both arms. There is no variance for a skill to act on.

## Why the rebuild did not help these two

`gateway` and these two got the same treatment — one pinned entry point,
internals free — and only `gateway` responded. The difference is not how much
was pinned but what the task is.

`gateway` is a routing job: five routes, three status codes, an ordering rule
and a filter with an error path. Each of those is a small independent decision
about where code goes, and the number of ways to arrange them is large. That is
what the skilled arm spent its 53 fewer lines on.

`feed` and `catalog` are each a single data transformation with a precisely
specified output. Pinning one entry point gave the model freedom to invent its
own internal types — and it did, the same two types in all six `feed` sessions
— but freedom without alternatives is not room. Every competent implementation
of "filter, format, sort, count, marshal" is the same shape, and the golden
test has to specify the output exactly for it to be checkable at all, which
removes the last of the choices.

The lesson for the next fixture: room to over-engineer comes from a task with
many small independent placement decisions, not from an unpinned API. A spec
precise enough to test mechanically and a transformation simple enough to have
one obvious shape are the same spec.

## Status

`catalog` stays unadmitted but is worth resolving at `n=5`, where its
separating ranges would either hold or collapse. `feed` is not admitted and
should not be re-run as it stands; it needs a task with placement decisions in
it, not a bigger `n`.
