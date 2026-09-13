# The no-shell paragraph and the owner note — refactor corpus, two arms, Sonnet 5 medium (n=1)

The [morning's run](2026-09-13-go-refactor-routing-load-n1-sonnet-5-medium.md)
showed the load step in `go-code-refactor` doing what it was written for and
costing +87% a session, of which only about $0.04 was the skills it meant to
load. The rest was behavior around them: without a shell tool the model
re-invoked the 9K-token skill "to run the baseline" (2/4 sessions, $0.05
each), loaded `go-linting`, `go-testing` and `go-code-review` after its edit
"to verify" ($0.06), met the routing gate five times, each time a wasted
edit and two extra turns, and once wrote a 167-line characterization test.
Three edits followed, all in the working tree:

- `go-code-refactor`'s `REFACTOR_SKILL_DIR` paragraph and step 5 say what to
  do without a shell tool: run nothing, do not read the scripts, reload the
  skill, or load `go-linting` or `go-code-review` to find a way to verify;
  the edit hook's output is the check record and a check no hook ran is
  `unavailable (no shell)`.
- The "every `Skill` call in one message" clause is gone from
  `go-code-refactor` and `go-code-review`; Sonnet 5 medium loaded one per
  turn regardless.
- `go-prompt-routing.sh` names `go-style-core` and the owners the target's
  code points at — the `.go` files the prompt names, scanned through a new
  `go-code-routing.sh --hints` mode with the gate's own table — so a session
  can load everything before its first edit instead of meeting the gate once
  per owner.

The nine horizontal rules in `go-code-refactor/SKILL.md` went to make room
under the 400-line cap (389 lines after). This run is the smoke for the three
edits: `reference` against `baseline`, one repetition per fixture; the
unaided control is the morning's. One repetition supports no line claim, and
a dollar figure at n=1 is noise (see the reference arm below); what it can
show is which loads, re-loads, and gate blocks happened.

| Arm | Tree |
|---|---|
| `reference` | `git worktree` of `831859b` (release 1.17.0) |
| `baseline` | the working tree: 1.17.0 plus the load step, the three-router gate, and the three edits above, uncommitted at run time |

## Run

- Finished: 2026-09-13 21:53 UTC
- Runner: `claude` 2.1.267
- Model: `claude-sonnet-5`, reasoning effort `medium`
- Seed: `1`; `-j 3`
- Corpus and fixtures: `refactor` — `dispatch`, `pricing`, `report`, `store`
- Arms: `reference`, `baseline`; 1 repetition per fixture and arm; 8 sessions, 0 CLI errors, 8/8 built and passed the golden test
- `reference` plugin SHA-256 `8c0be9f3d1e1c2556b99db2463fb48346f388007d373a4181d7e915a93b62472` (the morning's reference arm, byte-identical)
- `baseline` plugin SHA-256 `16fbb490b6a76db665f70be491a106f2d94316119b5865860fc95abd8340882e`
- Toolchain: Go 1.27.1 linux/amd64; golangci-lint 2.13.2
- Report: [`2026-09-13-go-refactor-routing-cost-n1-sonnet-5-medium.json`](2026-09-13-go-refactor-routing-cost-n1-sonnet-5-medium.json)
  (SHA-256 `95d2614b3751e4898af70a09978d33400885dd8badb64c0d8a47e50ebed6fd69`);
  traces [`….traces.tar.gz`](2026-09-13-go-refactor-routing-cost-n1-sonnet-5-medium.traces.tar.gz)
  (SHA-256 `fb8015795ba305c6b2a28afc2739d6d6b85f7d44a161f7b32c0c313a6278d6af`),
  one `traces/<arm>-<fixture>-r0.jsonl` per session
- Cost: $0.911 reference, $0.823 baseline; $1.73 in all

```bash
go run ./cmd/abrun -corpus refactor -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <worktree of 831859b> -arms reference,baseline \
  -tasks dispatch,pricing,report,store -n 1 -j 3 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-13-go-refactor-routing-cost-n1-sonnet-5-medium.json
```

No session had a shell tool; the edit hook ran gofmt, vet, `go fix -diff`,
the package tests and golangci-lint after each edit in both arms. The prompt
hook named `go-code-refactor` in 8/8 sessions; in the baseline arm its note
also named `go-style-core` and, for `dispatch`, `go-error-handling`,
`go-context` and `go-interfaces`, for `pricing` and `store`
`go-error-handling`, for `report` nothing (its code carries none of the
decision-bearing forms).

## What each session loaded

"Before" is the set in context at the first `.go` edit; "gate" is what the
routing gate named, each fire followed by the loads and a retry that landed.

| Arm | Fixture | Before the first `.go` edit | Gate named | Refactor skill loads | $ |
|---|---|---|---|---|---|
| `reference` | `dispatch` | `go-code-refactor`, `go-code` | `go-style-core`, `go-error-handling`, `go-context` | 1 | 0.242 |
| `reference` | `pricing` | `go-code-refactor` ×2 | — | 2 | 0.303 |
| `reference` | `report` | `go-code-refactor` | — | 1 | 0.157 |
| `reference` | `store` | `go-code-refactor` ×2, `go-style-core` | — | 2 | 0.209 |
| `baseline` | `dispatch` | `go-code-refactor`, `go-style-core`, `go-error-handling`, `go-context`, `go-interfaces` | — | 1 | 0.170 |
| `baseline` | `pricing` | `go-code-refactor`, `go-style-core`, `go-error-handling` | — | 1 | 0.227 |
| `baseline` | `report` | `go-code-refactor` | `go-style-core`; then `go-error-handling` | 1 | 0.279 |
| `baseline` | `store` | `go-code-refactor`, `go-style-core`, `go-error-handling` | — | 1 | 0.147 |

- **Re-loads are gone.** `go-code-refactor` was loaded once in each baseline
  session; the reference arm loaded it six times in four sessions, the same
  "run the baseline" re-invocation the morning's baseline arm showed. The
  morning's re-loads were the 1.17.0 text's base rate, not the load step's.
- **No post-edit loads.** No baseline session loaded `go-linting`,
  `go-testing` or `go-code-review` after its edit.
- **The note got the loads before the edit in 3/4 sessions.** `dispatch`,
  `pricing` and `store` loaded exactly what the note named, before reading
  the code or right after, and met no gate. `report`, the fixture whose note
  named `go-style-core` only, loaded nothing past the router until the gate
  blocked its first edit, then met the gate again for the `%w` its own edit
  introduced: two blocks, the arm's only ones, and the arm's most expensive
  session at $0.279.
- **The note is a superset.** `dispatch` loaded `go-interfaces` (2.2K
  tokens) for the `interface{` in its source; the morning's gate, which reads
  the edit, never named it. The price is one small payload per false hit.
- Loads still arrive one per turn (`Skill` calls on 3, 3, 3 and 5 turns).

## Refactor corpus

| Arm | Δlines | Δfuncs | Line gate | Golden | Lint clean | Gate blocks | Turns | $/run |
|---|---|---|---|---|---|---|---|---|
| `reference` | −13.2 | +0.75 | 3/4 | 4/4 | 3/4 | 1 | 59 | 0.228 |
| `baseline` | −11.2 | +0.75 | 3/4 | 4/4 | 3/4 | 2 | 58 | 0.206 |
| morning `baseline` (load step alone) | −13.2 | +0.75 | 3/4 | 4/4 | 3/4 | 5 | 86 | 0.348 |
| morning `reference` | −16.0 | +0.75 | 3/4 | 4/4 | 3/4 | 1 | 47 | 0.186 |

| Fixture | `reference` | `baseline` |
|---|---|---|
| `dispatch` | −10 | −11 |
| `pricing` | −50 | −41 |
| `report` | +13, 3 funcs | +13, 3 funcs |
| `store` | −6 | −6 |

Correctness and lint do not move: golden 4/4 and lint-clean 3/4 in both arms
(the `dispatch` residue is the pre-existing `revive` stutter finding, 6→1 in
both), `report` grew by the same three helpers in both and failed the line
gate in both. The 2-line corpus gap is `pricing`, −50 against −41: both
sessions replaced the branch chains with tables; the reference session
folded the plan facts into one record (81 lines), the baseline session
kept one table per file, `seatCosts` and `discountTiers`, plus a `switch`
on the term (90 lines).
Not a reading of the edits at n=1.

`store` is −6 in both arms this time. The reference session loaded
`go-style-core` before its first edit and kept the two nil-map guards, as
the morning's baseline session had; the count is now 3/3 sessions with
`go-style-core` in context before the first `store` edit keeping the guards
and 2/2 without deleting them. Still a pattern for n=5, not a claim.

## Cost

| Arm | Turns | Cache-create tokens | Cache-read tokens | Output tokens | $ (4 sessions) |
|---|---|---|---|---|---|
| morning `reference` | 47 | 97,131 | 867,617 | 17,873 | 0.745 |
| morning `baseline` | 86 | 175,564 | 1,976,361 | 29,174 | 1.393 |
| `reference` | 59 | 121,487 | 1,064,895 | 20,789 | 0.911 |
| `baseline` | 58 | 113,934 | 907,259 | 18,173 | 0.823 |

At Sonnet 5 rates (cache write $4/M for the 1-hour cache, cache read $0.20/M,
output $10/M) each row's estimate lands within a cent of the bill. The
baseline arm is −41% against the morning's baseline and −10% against the
reference arm in the same run: fewer cache writes (no re-loads), half the
cache reads (58 turns against 86), and no test file. The reference arm
itself rose from $0.745 to $0.911 between the two runs — two re-loads and a
`go-code` load in `dispatch` this time, none in the morning — which is the
size of n=1 noise on this corpus and why the dollar gap between the arms is
not the result. The structural counts are: refactor-skill loads 4 against 6,
post-edit loads 0 against 0 (the morning's baseline had 3), gate blocks 2
against 1 (the morning's baseline had 5), sessions with every named skill in
context before the first edit 3/4 against 0/4.

## What this run does and does not establish

- The no-shell paragraph removes the re-load and the post-edit loads that
  the morning attributed to the load step: 4 loads in 4 sessions against 6
  in the 1.17.0 arm, 0 post-edit loads.
- The prompt note puts `go-style-core` and the owners in context before the
  first edit in 3/4 sessions; the gate covers the fourth. Loads still land
  one per turn, so the note buys the absence of blocks, not fewer turns per
  load.
- Cost is back at the reference arm's level within n=1 noise; whether it is
  below it needs n=5.
- No correctness or lint regression; lines within one-repetition noise.
- `go-code-review`'s no-shell path and load step remain covered by tests
  only; no corpus drives a review that edits.

## Next

n=5 two-arm on `store` and `pricing`: `store` for the guard pattern (3/3
against 2/2), `pricing` for the test-writing rate and the shape of the fold, one shared
record or one table per file.
A note that names owners the edit will introduce — `report`'s `%w` — is not
possible from the source; the gate is the right place for that case.
