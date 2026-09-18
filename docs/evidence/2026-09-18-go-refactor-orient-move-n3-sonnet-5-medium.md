# The Orient move and the negation rewrites on the refactor corpus — Sonnet 5 medium, n=3 plus `report` at n=5

The 1.21.0 change to `go-code-refactor` touched no rule: the
`REFACTOR_SKILL_DIR` paragraph and its `bash` block moved from Resource
Routing to the top of Workflow step 1 (Orient), where the first command runs;
the `loc-diff` sentence, the bugs-found sentence, and step 5's edit-hook line
were restated as the action to take, and step 5 routes to `go-style-core`
"The Edit Hook Record" instead of carrying its own copy. The same tree
carries the `go-style-core` edits of the day (the hook record, the "How Much
To Say" rewording), which every refactor session loads. Until this run the
change had passed the structural suite only. Two runs, both against release
1.20.1 on the model where the refactor corpus discriminates (2026-09-10,
n=5): the four fixtures at n=3, then `report` alone at n=5 after the first
run's `report` cells sat 14 lines apart.

| Arm | Tree |
|---|---|
| `reference` | `git worktree` of `89d702a` (release 1.20.1), plugin SHA-256 `d478dd31988660fcb8934eba8e56e7e378221b321b278073006d4ecb4af19ca0` |
| `baseline` | `git worktree` of `66f06cb` (the 1.21.0 branch head), plugin SHA-256 `bfcfed52ecd8273c265ff05738e65a40716a3caa9b16a78dde590c732cd14ef0` — the same bytes as the day's `baseline` in the implement and review runs |

## Runs

Run 1 (four fixtures):

- Finished: 2026-09-18 19:10 UTC
- Runner: `claude` 2.1.267; model `claude-sonnet-5`, effort `medium`; seed `1`
- Corpus `refactor`, fixtures `dispatch`, `pricing`, `report`, `store`, arms `reference`, `baseline`, 3 repetitions per fixture and arm, 24 sessions, `-j 4`, 0 CLI errors
- Report: [`2026-09-18-go-refactor-orient-move-n3-sonnet-5-medium.json`](2026-09-18-go-refactor-orient-move-n3-sonnet-5-medium.json) (SHA-256 `cd29bf6af2223ce034d4700b3595f632139acf9116644b65d555444504a6eda5`); traces [`….traces.tar.gz`](2026-09-18-go-refactor-orient-move-n3-sonnet-5-medium.traces.tar.gz) (SHA-256 `e1bf5b27f89dabf0a4ae2b33266cd8a1280211b86951aab018dcbe12f2700bde`), one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $6.72 ($3.53 reference, $3.19 baseline)

Run 2 (`report` alone):

- Finished: 2026-09-18 19:20 UTC; seed `2`; 5 repetitions per arm, 10 sessions, `-j 2`, 0 CLI errors; otherwise as run 1
- Report: [`2026-09-18-go-refactor-orient-move-report-n5-sonnet-5-medium.json`](2026-09-18-go-refactor-orient-move-report-n5-sonnet-5-medium.json) (SHA-256 `bc335f6dbb15c3ab9f2966d7e75251c7e951968bdd36a9ea22e1ce3fe0f16f8d`); traces [`….traces.tar.gz`](2026-09-18-go-refactor-orient-move-report-n5-sonnet-5-medium.traces.tar.gz) (SHA-256 `0afd75583d4809113fcedd83be5b0aec3de0efd4115e8f797216abbeed3a28e0`)
- Cost: $3.11

```bash
go run ./cmd/abrun -corpus refactor -tasks dispatch,pricing,report,store -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <worktree of 89d702a> -arms reference,baseline \
  -n 3 -j 4 -seed 1 -timeout 10m -keep -verbose -out <run 1 json>
go run ./cmd/abrun -corpus refactor -tasks report ... -n 5 -j 2 -seed 2 ... -out <run 2 json>   # run 2
```

Toolchain Go 1.27.1 linux/amd64, golangci-lint 2.13.2 on PATH, so the edit
hook ran after every `.go` edit; no session had a shell tool, so
`verify-refactor.sh` and `loc-diff` never ran and the `REFACTOR_SKILL_DIR`
block was never executed in either arm. What the move can change here is
what a shell-less session does instead: re-load the skill, read the scripts,
load `go-linting` or `go-code-review` to find a way to verify — the
behaviors the 2026-09-13 runs attributed to the old paragraph's absence. The
event columns are counted from the traces as in the day's implement reports;
"refactor loads" is the number of `Skill` calls for `go-code-refactor`,
"post-edit loads" the `Skill` calls after the first edit the host applied,
"script reads" the `Read`s under a skill's `scripts/`.

## The reading

Run 1:

| Arm | Valid | Golden | Lint clean | Δlines | Δfuncs | Line gate | Skill turns/s | Gate blocks | Hook findings | Refactor loads | Post-edit loads | Script reads | Card `Read` | Counts claimed | Turns/s | $/session |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `reference` | 12/12 | 12/12 | 9/12 | −15.5 | +0.25 | 10/12 | 2.08 | 7 | 14 | 12 | 3 | 2 | 1/12 | 6/12 | 10.3 | 0.294 |
| `baseline` | 12/12 | 12/12 | 9/12 | −10.8 | +1.17 | 9/12 | 1.83 | 5 | 17 | 12 | 3 | 1 | 0/12 | 7/12 | 9.5 | 0.266 |

| Fixture | `reference` Δlines (r0, r1, r2) | `baseline` Δlines (r0, r1, r2) | `reference` Δfuncs | `baseline` Δfuncs |
|---|---|---|---|---|
| `dispatch` | −10, −10, −11 (−10.3) | −12, −10, −4 (−8.7) | 0, 0, 0 | 0, 0, 0 |
| `pricing` | −51, −44, −50 (−48.3) | −49, −48, −40 (−45.7) | 0, 0, 0 | +1, +1, 0 |
| `report` | +1, −1, +13 (+4.3) | +17, +18, +19 (+18.0) | 0, 0, +3 | +4, +4, +4 |
| `store` | −11, −6, −6 (−7.7) | −6, −9, −6 (−7.0) | 0, 0, 0 | 0, 0, 0 |

Run 2, `report` at n=5:

| Arm | Golden | Lint clean | Δlines (r0…r4) | Mean | Δfuncs | Skill turns/s | Gate blocks | Hook findings | Refactor loads | Counts claimed | Turns/s | $/session |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `reference` | 5/5 | 5/5 | −1, +19, +12, +10, +5 | +9.0 | 0, +4, +3, +3, +1 | 2.00 | 5 | 11 | 5 | 1/5 | 12.0 | 0.305 |
| `baseline` | 5/5 | 5/5 | +5, +12, +13, +7, −1 | +7.2 | +1, +3, +3, +1, 0 | 2.60 | 5 | 14 | 5 | 1/5 | 13.2 | 0.316 |

Per session, run 1:

| Arm | Fixture | Rep | Golden | Lint | Δlines | Skills loaded (in load order) | Skill turns | Gate blocks | Hook findings | Turns | $ |
|---|---|---|---|---|---|---|---|---|---|---|---|
| `reference` | `dispatch` | 0 | pass | 1 | −10 | go-code-refactor, go-style-core, go-error-handling, go-context, go-interfaces | 1 | 0 | 1 | 8 | 0.301 |
| `reference` | `dispatch` | 1 | pass | 1 | −10 | go-code-refactor, go-style-core, go-error-handling, go-context | 2 | 1 | 1 | 10 | 0.249 |
| `reference` | `dispatch` | 2 | pass | 1 | −11 | go-code-refactor, go-style-core, go-error-handling, go-context, go-interfaces | 1 | 0 | 1 | 8 | 0.225 |
| `reference` | `pricing` | 0 | pass | 0 | −51 | go-code-refactor, go-style-core, go-error-handling, go-testing | 4 | 0 | 2 | 15 | 0.421 |
| `reference` | `pricing` | 1 | pass | 0 | −44 | go-code-refactor, go-style-core, go-error-handling | 2 | 1 | 1 | 13 | 0.385 |
| `reference` | `pricing` | 2 | pass | 0 | −50 | go-code-refactor, go-style-core, go-error-handling | 2 | 1 | 0 | 12 | 0.406 |
| `reference` | `report` | 0 | pass | 0 | +1 | go-code-refactor, go-style-core, go-error-handling | 2 | 1 | 2 | 12 | 0.334 |
| `reference` | `report` | 1 | pass | 0 | −1 | go-code-refactor, go-style-core, go-error-handling | 2 | 1 | 4 | 15 | 0.383 |
| `reference` | `report` | 2 | pass | 0 | +13 | go-code-refactor, go-style-core, go-error-handling | 3 | 1 | 2 | 11 | 0.247 |
| `reference` | `store` | 0 | pass | 0 | −11 | go-code-refactor, go-style-core, go-error-handling | 2 | 1 | 0 | 8 | 0.207 |
| `reference` | `store` | 1 | pass | 0 | −6 | go-code-refactor, go-style-core, go-error-handling | 2 | 0 | 0 | 6 | 0.174 |
| `reference` | `store` | 2 | pass | 0 | −6 | go-code-refactor, go-style-core, go-error-handling | 2 | 0 | 0 | 6 | 0.203 |
| `baseline` | `dispatch` | 0 | pass | 1 | −12 | go-code-refactor, go-style-core, go-error-handling, go-context, go-interfaces | 2 | 0 | 1 | 11 | 0.259 |
| `baseline` | `dispatch` | 1 | pass | 1 | −10 | go-code-refactor, go-style-core, go-error-handling, go-context, go-interfaces | 1 | 0 | 1 | 6 | 0.211 |
| `baseline` | `dispatch` | 2 | pass | 1 | −4 | go-code-refactor, go-style-core, go-error-handling, go-context | 2 | 1 | 1 | 9 | 0.234 |
| `baseline` | `pricing` | 0 | pass | 0 | −49 | go-code-refactor, go-style-core, go-error-handling | 2 | 0 | 2 | 13 | 0.348 |
| `baseline` | `pricing` | 1 | pass | 0 | −48 | go-code-refactor, go-style-core, go-error-handling | 2 | 0 | 1 | 9 | 0.285 |
| `baseline` | `pricing` | 2 | pass | 0 | −40 | go-code-refactor, go-style-core, go-error-handling | 1 | 0 | 0 | 5 | 0.193 |
| `baseline` | `report` | 0 | pass | 0 | +17 | go-code-refactor, go-style-core, go-error-handling | 2 | 1 | 2 | 12 | 0.285 |
| `baseline` | `report` | 1 | pass | 0 | +18 | go-code-refactor, go-style-core, go-error-handling | 2 | 1 | 5 | 14 | 0.332 |
| `baseline` | `report` | 2 | pass | 0 | +19 | go-code-refactor, go-style-core, go-testing, go-error-handling | 3 | 1 | 4 | 14 | 0.473 |
| `baseline` | `store` | 0 | pass | 0 | −6 | go-code-refactor, go-style-core, go-error-handling | 1 | 0 | 0 | 7 | 0.187 |
| `baseline` | `store` | 1 | pass | 0 | −9 | go-code-refactor, go-style-core, go-error-handling | 2 | 1 | 0 | 7 | 0.202 |
| `baseline` | `store` | 2 | pass | 0 | −6 | go-code-refactor, go-style-core, go-error-handling | 2 | 0 | 0 | 7 | 0.180 |

**Correctness and lint did not move.** Golden 12/12 and 5/5 in both arms of
both runs; lint-clean 9/12 and 9/12, the three findings in each arm being
the one pre-existing `revive` stutter on `DispatchAll` in every `dispatch`
session; `go fix` pending 0 in every session. `store` kept its `cache` and
`remote` nil guards in 5/6 sessions; the one that dropped the `remote` guard
(`reference` r0, −11) is the `store` session whose first edit the gate
refused for `go-style-core`, so the skill entered its context after the
refactor was already shaped — the 2026-09-13 pattern (guards kept when
`go-style-core` is in context before the first edit) holds at 8/8 with one
session that loaded it late and deleted.

**The shell-less behaviors the Orient move guards stayed absent.** One load
of `go-code-refactor` per session in 34/34 sessions of both arms (the
2026-09-13 reference arm loaded it six times in four sessions before the
paragraph existed at all). Script reads 2 against 1 — a `reference` session
read `verify-refactor.sh` in `report` r0 and a wrong path for the card, a
`baseline` session read `check-debt.sh`. Loads after the first applied edit
3 against 3 in run 1, 5 against 6 in run 2, all owners the edit introduced
(`go-error-handling` for a `%w`, `go-testing` for a test the session chose to
write). The one load the paragraph forbids by name is `baseline` `report`
r3 of run 2: `go-linting` and `go-testing` after the edit, then a 44-line
test file — 1 session in 17 against 0 in 17.

**Cost leaned toward the baseline in run 1 and level in run 2.** $0.266
against $0.294 a session (−10%) with 9.5 assistant turns against 10.3 and 5
gate blocks against 7 in run 1; $0.316 against $0.305 (+4%) in run 2. The
gate blocks are the `report` `%w` in every `report` session of both arms (a
form the source does not carry, so the prompt note cannot name it) and
`go-style-core` or `go-error-handling` missing at the first edit in three
`reference` and two `baseline` sessions elsewhere. Neither run's dollar gap
is outside the corpus's run-to-run movement.

**Lines: the `report` cells of run 1 were the fixture's coin.** Run 1's
corpus gap, −10.8 against −15.5, is `report`: the `baseline` extracted four
helpers (`writeHeader`, `writeRow`, `writeTotal`, a cents formatter) in 3/3
sessions at +17 to +19 lines while the `reference` did so in 1/3. Run 2 at
n=5 on that fixture reverses the lean, +7.2 against +9.0 (exact permutation
p = 0.75), with three or four helpers in 2/5 `baseline` sessions against 3/5
`reference`; over both runs `report` is +11.3 against +7.3 (n=8 each,
p = 0.30). The other three fixtures sit within a session of each other:
`pricing` −45.7 against −48.3 with the same table rewrite in every session,
`dispatch` −8.7 against −10.3 with one `baseline` session that met the gate
first and folded less, `store` −7.0 against −7.7.

**Two of the restatements have a reading here, one does not.** The `loc-diff`
sentence ("report only the two counts `loc-diff` printed") did not move the
rate at which a shell-less report states a line count anyway: 7/12 against
6/12 in run 1, 1/5 against 1/5 in run 2, with `abrun`'s `counts reported`
column as the measure. The `(hook)` marker of the hook record appeared in
no refactor report of either arm — the refactor report template has no
checks line of that shape, and the sessions wrote prose ("the edit hook's
`go vet`/`golangci-lint` passed after each edit") in 8/12 and 8/12. The
bugs-found sentence had nothing to act on: no session reported a bug it
declined to fix.

## What this establishes

- The `go-code-refactor` half of 1.21.0 is a no-regression change on the
  refactor corpus at n=3 plus n=5 on its widest fixture: golden 17/17 in
  both arms, lint and `go fix` identical, one skill load per session in both
  arms, cost −10% and +4% in the two runs.
- `report` remains bimodal on Sonnet 5 medium — a flat `Render` at ±1 or
  three to four write helpers at +12 to +19 — and three sessions cannot read
  it; the 2026-09-10 control's `+5.5 ± 6.1` was the same fact.
- The negation rewrites are neutral where they can be measured: the
  line-count claim rate is level, and no report wrote `(hook)` because the
  refactor template has no checks line. The hook record's effect is the
  implement corpus's reading (8/11 against 1/11 the same day).
- Sonnet 5 medium read the idiom card in 1/17 refactor sessions with the
  1.20.1 text and 0/17 with this one; the card route measured on the
  implement corpus the same evening applies to refactor prompts as well and
  is unmeasured here.

## Next

- A refactor-corpus run of the card route (`go-prompt-routing.sh` naming the
  card, the gate requiring it) against this baseline, n≥3, before the route
  is credited for refactor sessions.
- `report` at n≥8 in one run is what a line claim on that fixture costs; the
  fixture is worth keeping precisely because it is the one Sonnet 5 medium
  has not settled.
