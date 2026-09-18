# The edit-hook record and the card in step 2 — Opus 5 medium, `feed` and `gateway` at n=2

The other half of the [Sonnet 5 pair of the day](2026-09-18-go-implement-card-step-roster-feed-gateway-n2-sonnet-5-medium.md):
the same working tree — `go-code` step 2 naming the idiom card, what the edit
hook runs stated once in `go-style-core` "The Edit Hook Record" with the
routers routing to it, the `go-code-refactor` Orient move, and the negation
rewrites — against release 1.20.1, on the model where the [2026-09-13
hook-vs-text run](2026-09-13-go-implement-hook-vs-text-feed-gateway-n1-opus-5-medium.md)
found the hook's cost was noise. The question here is the one the Opus 5
prompting guide raises: does a second description of the checks cost tokens
on this model, and does removing two of the three copies save any.

| Arm | Tree |
|---|---|
| `reference` | `git worktree` of `89d702a` (release 1.20.1): the hook's check list in `go-code` Resource Routing, `go-code-refactor` Resource Routing, and "Write Current Go" |
| `baseline` | the working tree (plugin SHA-256 `bfcfed52…`): one statement in `go-style-core`, three routes to it; the card as the first clause of `go-code` step 2 |

## Run

- Finished: 2026-09-18 16:59 UTC
- Runner: `claude` 2.1.267; model `claude-opus-5`, effort `medium`; seed `1`
- Corpus `implement`, fixtures `feed`, `gateway`, arms `reference`, `baseline`, 2 repetitions per fixture and arm, 8 sessions, `-j 4`, 0 CLI errors
- `reference` plugin SHA-256 `d478dd31988660fcb8934eba8e56e7e378221b321b278073006d4ecb4af19ca0`; `baseline` `bfcfed52ecd8273c265ff05738e65a40716a3caa9b16a78dde590c732cd14ef0`
- Toolchain Go 1.27.1 linux/amd64, golangci-lint 2.13.2 on PATH; the edit hook ran in every session, no session had a shell tool
- Report: [`2026-09-18-go-implement-hook-record-feed-gateway-n2-opus-5-medium.json`](2026-09-18-go-implement-hook-record-feed-gateway-n2-opus-5-medium.json) (SHA-256 `07f0e2805476702f940a5a996ac690c7a2ac8f74a703f8d11f12c063487c6979`); traces [`….traces.tar.gz`](2026-09-18-go-implement-hook-record-feed-gateway-n2-opus-5-medium.traces.tar.gz) (SHA-256 `bba09600fd03c49c9835c25d4324dd4b769d8f2dfd7329efc436f03aa03a3eb0`), one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $8.11

```bash
go run ./cmd/abrun -corpus implement -tasks feed,gateway -runner claude -model claude-opus-5 -effort medium \
  -reference-root <worktree of 89d702a> -arms reference,baseline \
  -n 2 -j 4 -seed 1 -timeout 10m -keep -verbose \
  -out ../docs/evidence/2026-09-18-go-implement-hook-record-feed-gateway-n2-opus-5-medium.json
```

The event columns are counted from the traces as in the Sonnet report.

## The reading

| Arm | Valid | Golden | Lint clean | Δlines | Skill calls/s | Skill turns/s | Card `Read` | Gate blocks | Hook findings | Assistant turns/s | $/session |
|---|---|---|---|---|---|---|---|---|---|---|---|
| `reference` | 4/4 | 4/4 | 4/4 | +52.8 | 5.50 | 2.00 | 4/4 | 0 | 15 | 11.0 | 1.037 |
| `baseline` | 4/4 | 4/4 | 4/4 | +53.2 | 5.00 | 2.25 | 4/4 | 0 | 14 | 10.2 | 0.989 |

| Arm | Fixture | Rep | Golden | Lint | Δlines | Skills loaded (in load order) | Skill turns | Card `Read` | Hook findings | Turns | $ | Checks line |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `reference` | `feed` | 0 | pass | 0 | +32 | go-code, go-style-core, go-testing, go-data-structures, go-error-handling, go-defensive | 2 | yes | 3 | 10 | 0.996 | every check `(hook)` |
| `reference` | `feed` | 1 | pass | 0 | +40 | go-code, go-style-core, go-testing, go-data-structures, go-error-handling, go-defensive | 2 | yes | 2 | 9 | 0.807 | every check `(hook)` |
| `reference` | `gateway` | 0 | pass | 0 | +71 | go-code, go-style-core, go-http, go-error-handling, go-testing | 2 | yes | 5 | 12 | 1.178 | every check `unavailable (no shell)` |
| `reference` | `gateway` | 1 | pass | 0 | +68 | go-code, go-style-core, go-http, go-error-handling, go-testing | 2 | yes | 5 | 13 | 1.167 | every check `(hook)` |
| `baseline` | `feed` | 0 | pass | 0 | +37 | go-code, go-style-core, go-testing, go-data-structures, go-error-handling | 2 | yes | 3 | 8 | 0.733 | `(hook)` ×3, `lint unavailable (no shell)` |
| `baseline` | `feed` | 1 | pass | 0 | +32 | go-code, go-style-core, go-data-structures, go-error-handling, go-testing | 2 | yes | 2 | 8 | 0.824 | every check `(hook)` |
| `baseline` | `gateway` | 0 | pass | 0 | +77 | go-code, go-style-core, go-http, go-error-handling, go-testing | 2 | yes | 5 | 13 | 1.218 | every check `(hook)` |
| `baseline` | `gateway` | 1 | pass | 0 | +67 | go-code, go-style-core, go-http, go-error-handling, go-testing | 3 | yes | 4 | 12 | 1.182 | every check `(hook)` |

**Opus 5 reads the card from the routing line already: 4/4 in both arms.**
The three models now sit apart on one `Read`: Opus 5 medium 8/8 today,
Haiku 4.5 5/10 with the 1.20.1 text and 4/5 with step 2, Sonnet 5 medium
0/24. The step-2 wording had nothing to add on this model.

**Removing two copies of the check list cost nothing.** Golden 4/4 and
lint-clean 4/4 in both arms, `go fix` with nothing left to propose in 8/8,
Δlines +53.2 against +52.8. The baseline took 10.2 assistant turns a session
against 11.0 and cost $0.989 against $1.037 (−5%), with one fewer edit-hook
finding; at two sessions a fixture the sign is not a result, the absence of
a cost in the other direction is. The 2026-09-13 reading — the hook is not a
cost on Opus 5 — holds for its description too.

**The report form was already Opus 5's.** The reference marked every claimed
check `(hook)` in 3/4 sessions from the 1.20.1 sentence alone; the fourth
(`gateway` r0) wrote `unavailable (no shell)` on all four checks although the
hook had printed five findings in that session, the over-cautious reading the
record now rules out. The baseline marked 4/4, and wrote `lint unavailable (no
shell)` in the one session where the linter had not printed.

**The skill set narrowed by one.** Both reference `feed` sessions loaded
`go-defensive` beside the four owners the baseline loaded; neither baseline
`feed` session did. Two sessions each, and nothing in the edited text names
`go-defensive`, so a note, not a reading.

## What this establishes

- On Opus 5 medium the change is neutral on every measured axis and costs
  no tokens: the guide's warning about a second verification description
  is not visible here, and the dedup did not remove anything the model used.
- The card is read on this model without a step; Sonnet 5 is the model the
  card's routing has to be designed for, and Haiku 4.5 the one that shows a
  step's effect.
- The `(hook)` marking rule is followed 4/4; the one reference session that
  called a printed check `unavailable` is the case the new sentence about the
  hook's silence exists for.

## Next

- Nothing on this model until a fixture Opus 5 medium fails is found; the
  three-model comparison of the card read is the reusable result.
