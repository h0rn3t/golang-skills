# Go Implementation Corpus Control — MiMo v2.5 Pro on opencode

The first run in which three of the four traps are live at once.

## Run

- Finished: 2026-09-07 15:49 EEST
- Runner: `opencode` 1.18.29
- Model: `opencode-go/mimo-v2.5-pro`
- Seed: `1`
- Corpus: `implement`
- Fixtures: `catalog`, `feed`, `gateway`, `ledger`
- Arms: `no-skill`, `baseline`
- Repetitions: 5 per fixture and arm, 40 sessions total
- Baseline plugin SHA-256: `6443a658f74cd465b98f0065cdb7c117bcde52c64af1acc1a5ef7f3a67369d3c`
- Plugin source: working tree at `5dcb55d`
- Raw report: [`2026-09-07-go-implement-control-mimo-v2.5-pro.json`](2026-09-07-go-implement-control-mimo-v2.5-pro.json)
- Raw report SHA-256: `20233eff50c33f540b44defe29d3e2eab718f323570eeca3de0d04c99738de2f`

```bash
go run ./cmd/abrun -corpus implement -runner opencode \
  -model opencode-go/mimo-v2.5-pro \
  -arms no-skill,baseline -n 5 -j 4 -seed 1 \
  -out ../docs/evidence/2026-09-07-go-implement-control-mimo-v2.5-pro.json
```

Thirty-nine of 40 sessions completed, every one of them built, changed the
fixture and stayed out of the repository checkout. One control session (`ledger`
rep 3) died on opencode's `database is locked` — SQLite contention between the
four concurrent sessions sharing an arm home at `-j 4`, not a model failure —
and is excluded rather than counted as a miss. A `go-*` skill fired in 16 of 20
baseline sessions: `go-code` in 16, `go-style-core` and `go-error-handling` in 2
each, `go-http`, `go-security` and `go-data-structures` in 1 each.

## Results

### Correctness

| Fixture | Golden, no skill | Golden, skill | The trap that failed |
|---|---:|---:|---|
| `catalog` | 4/5 | 4/5 | the source was asked for `sku-404` twice |
| `feed` | 3/5 | 3/5 | `events` rendered as `null`, not `[]` |
| `gateway` | **1/5** | **3/5** | `ReadTimeout` / `ReadHeaderTimeout` = 0 |
| `ledger` | 4/4 | 5/5 | — |
| All runs | 12/19 | 15/20 | |

Every failure in this run is the fixture's own trap and nothing else — no
compile errors, no signature drift, no partial patches. That has not happened
before: the [first discovery run](2026-09-07-go-implement-discovery-minimax-m3.md)
recorded 12 of 12 unaided passes on every fixture, and the
[MAI-Code-1.1-Flash run](2026-09-07-go-implement-control-mai-code-1.1-flash.md)
brought `gateway` to life but left the other three saturated. Here `gateway`,
`feed` and `catalog` are all live.

The arms separate on exactly one of them. On `gateway` the skill takes the pass
rate from 1/5 to 3/5, reproducing the MAI-Code-1.1-Flash result at the same
numbers; on `feed` and `catalog` both arms miss the same trap the same number of
times, and on `ledger` neither misses. Fisher's exact gives a two-sided p of 0.52
on `gateway` and 0.50 across the corpus, so with five repetitions per cell this
is a direction, not a result.

Routing is thin everywhere. `go-code` fired in 16 of 20 baseline sessions and
mostly on its own; the topic owners barely appear. `go-http`, which owns the
timeout rule, was loaded in 1 of the 5 baseline `gateway` sessions — that one
passed, but so did two of the four that never reached it, so the correlation
here is much weaker than the 2-for-2 in the MAI-Code-1.1-Flash run.
`go-data-structures`, which owns the nil-slice rule `feed` turns on, was never
loaded in a `feed` session at all: its single appearance in the whole corpus is
on `catalog`.

### Code size

| Fixture | Lines, no skill | Lines, skill | Effect |
|---|---:|---:|---:|
| `catalog` | +23.5 ± 4.1 (n=4) | +27.5 ± 7.4 (n=4) | +4.0 |
| `feed` | +57.3 ± 6.8 (n=3) | +59.0 ± 6.2 (n=3) | +1.7 |
| `ledger` | +40.5 ± 7.0 (n=4) | +39.4 ± 3.3 (n=5) | −1.1 |
| `gateway` | +76.0 (n=1) | +86.0 ± 21.9 (n=3) | not comparable |

Line values are mean delta with sample standard deviation over the runs that
passed. The corpus-wide means the summary prints — 42.0 against 49.5 — are not
an effect: the arms have different valid sets, and `gateway` is both the longest
fixture and the one where the control lost four of five runs, so it contributes
one run to the control mean and three to the skilled one.

| Structural additions across valid runs | No skill | Skill |
|---|---:|---:|
| Types | 2 | 1 |
| Interfaces | 0 | 0 |
| Functions | 1 | 3 |
| Unrequested exported declarations | 0 | 1 |

Neither arm wrote a test file. The corpus's over-engineering baits went untaken
in both arms, as they have in every run of it: no interfaces, no pattern names.

The corpus cost $0.356: $0.131 for 20 control runs against $0.226 for 20 skilled.

## Interpretation

The corpus finally does what it was built to do, on three fixtures instead of
one. That makes this the reference configuration for it: `opencode-go/mimo-v2.5-pro`
is the only model on which `feed` and `catalog` have ever failed unaided, and
the `_implement/README.md` note that `feed` "needs a different task, not a bigger
`n`" is wrong as stated — it needed a different model.

What the run does not show is the skill closing those two. `feed` and `catalog`
are 3/5 and 4/5 in both arms, and the routing says why: not one of the ten
`feed` sessions loaded `go-data-structures`, the skill whose nil-slice rule is
the whole of that fixture's trap. The skill that would have prevented the defect
exists, is installed, is described for exactly this case, and never fired. That
is a triggering result rather than a content one, and it is the actionable
finding here — the same shape as `gateway`, where `go-http` reached 1 session in
5. The baseline arm mostly loaded `go-code` and stopped there.

`gateway` reproduces across two unrelated models now, at 1/5 → 3/5 on both
MAI-Code-1.1-Flash and MiMo v2.5 Pro. Neither is significant alone; together
they are the same direction from independent runners, models and providers,
which is the point at which raising `n` on `gateway` is worth the quota.
