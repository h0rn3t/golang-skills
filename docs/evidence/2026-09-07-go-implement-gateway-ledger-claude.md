# Implementation Corpus — `gateway` and `ledger` on claude

The first implementation fixture to earn its place. `gateway` separates the arms
completely on how much code a passing implementation costs; `ledger` points the
same way without separating.

## Run

- Finished: 2026-09-07 12:27 EEST
- Corpus: `implement` (`evals/ab/_implement`)
- Runner: `claude`, Claude Code 2.1.261
- Model: **not pinned.** The run used the CLI's configured default because
  `-model` was not passed. A published table needs it pinned; this is a
  discovery run and the omission is recorded rather than papered over.
- Seed: `1`
- Fixtures: `gateway`, `ledger` — the two rebuilt to pin a single entry point
- Arms: `no-skill`, `baseline`
- Repetitions: 3 per fixture and arm, 12 sessions total
- Baseline plugin SHA-256: `6443a658f74cd465b98f0065cdb7c117bcde52c64af1acc1a5ef7f3a67369d3c`
- Raw report: [`2026-09-07-go-implement-gateway-ledger-claude.json`](2026-09-07-go-implement-gateway-ledger-claude.json)
- Raw report SHA-256: `5cdc5a6ebd806e5ea80c41723859fe24a16fe6dde5f0d074835326a46861f332`
- Total cost: $5.20 — $2.11 control, $3.09 skilled

```bash
go run ./cmd/abrun -corpus implement -runner claude \
  -tasks gateway,ledger -arms no-skill,baseline -n 3 -j 4 -seed 1 -keep \
  -out ../docs/evidence/2026-09-07-go-implement-gateway-ledger-claude.json
```

All 12 sessions completed, built, and passed their hidden specification. Every
session changed the fixture and none referenced the repository. A `go-*` skill
loaded in all 6 baseline sessions, and in every one it was `go-code`, the
router — no other skill was recorded.

## Result

Correctness is the gate and it is tied: 6/6 golden in both arms, on both
fixtures. The score is what a passing implementation cost.

| Fixture | Metric | No skill | Skill | Effect |
|---|---|---:|---:|---:|
| `gateway` | lines | 162.0 | 107.0 | **−55.0 (−34.0%)** |
| `gateway` | functions | 8.3 | 4.0 | −52% |
| `gateway` | branches | 13.7 | 8.0 | −42% |
| `gateway` | types | 1.0 | 0.7 | −0.3 |
| `ledger` | lines | 47.0 | 38.0 | −9.0 (−19.1%) |
| `ledger` | branches | 5.0 | 5.3 | +0.3 |
| `ledger` | types | 0.3 | 0.0 | −0.3 |

On `gateway` the two arms do not overlap: the control's best run is 149 lines
and the skilled arm's worst is 117, with control at 149/153/184 against
99/105/117. A Welch interval for the difference is approximately −94 to −16
lines at 95% (df ≈ 2.9), so the effect survives even at `n=3`. Functions and
branches fall with the line count rather than being traded for dispatch, which
is the shape of a leaner implementation rather than a differently-scaffolded
one.

On `ledger` the interval is −21 to +3 and includes zero. The direction matches
and the ranges nearly separate, at 41/47/53 against 36/37/41, but three runs
do not settle it.

## What the fixtures did and did not catch

The rebuild is what produced the signal. In the [previous discovery
run](2026-09-07-go-implement-discovery-minimax-m3.md) both fixtures pinned every
exported declaration, and `Δtypes`, `Δiface`, `Δfuncs` and `Δpattern` were zero
in all 24 sessions — there was nowhere to hang a scaffold, so nothing could
move. `gateway` now pins only `NewServer` and leaves routing, filtering and
encoding decomposition free; that is where the 4.3 functions and 5.7 branches
per run came from.

`ledger`'s bait did not land in either arm. It renders two formats and says a
third has never been asked for, which is the `report` fixture's temptation to
grow a `Formatter` interface with one implementation per format. No session in
either arm took it: `Δiface` was zero throughout and `Δbranch` sat flat at 5.
The task is small enough that the honest switch is the obvious choice with or
without a skill, so the bait is inert and only the line count is doing work.

## Limits

The model is not pinned, so this run cannot be compared line-for-line with a
later one. `n=3` is enough for `gateway` only because the arms separate
completely; it is not enough for `ledger`. Both fixtures tie on correctness, so
nothing here says the skill makes an implementation more likely to work — only
that the working implementation is smaller. And the claude runner restricts the
session to `Skill,Read,Glob,Grep,Edit,Write`, so none of these implementations
could compile or test itself before being measured.
