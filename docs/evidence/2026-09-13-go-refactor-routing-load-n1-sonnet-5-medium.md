# The load step and the three-router gate — refactor corpus, three arms, Sonnet 5 medium (n=1)

Two routing edits landed in the working tree on 2026-09-13 without a
measurement. `go-code-refactor` Workflow step 1 and `go-code-review` Review
Procedure step 2 now say to load `go-style-core` and the owners of the
decisions the diff touches before the first edit, every `Skill` call in one
message, and both list `../go-style-core/SKILL.md` in Resource Routing;
`hooks/go-code-routing.sh` holds the first `.go` edit after any of the three
routers (`go-code`, `go-code-refactor`, `go-code-review`) instead of after
`go-code` alone. The motivation was the refactor runs on record: a refactor
prompt reaches `go-code-refactor` alone through the prompt hook, and the
sessions that loaded it also loaded `go-style-core` in 2/20
([2026-09-10, medium n=5](2026-09-10-go-refactor-control-sonnet-5-medium-n5.md))
and 6/20 ([2026-09-13, architecture n=5](2026-09-13-go-refactor-architecture-report-store-n5-sonnet-5-medium.json)),
an owner skill in 4 of those 44, while the old gate stayed silent by design.

This run is the smoke for the edits: one repetition per fixture and arm on the
refactor corpus, an unaided arm beside the before/after pair. What it can show
is what a session loads and when, whether the gate fires and lets the retry
through, the cost of the loads, and a correctness or lint regression. One
repetition supports no line claim.

| Arm | Tree |
|---|---|
| `no-skill` | no plugin |
| `reference` | `git worktree` of `831859b` (release 1.17.0) |
| `baseline` | the working tree: 1.17.0 plus the two routing edits, uncommitted at run time |

## Run

- Finished: 2026-09-13 21:12 UTC
- Runner: `claude` 2.1.267
- Model: `claude-sonnet-5`, reasoning effort `medium`
- Seed: `1`; `-j 3`
- Corpus and fixtures: `refactor` — `dispatch`, `pricing`, `report`, `store`
- Arms: `no-skill`, `reference`, `baseline`; 1 repetition per fixture and arm; 12 sessions, 0 CLI errors, 12/12 built and passed the golden test
- `reference` plugin SHA-256 `8c0be9f3d1e1c2556b99db2463fb48346f388007d373a4181d7e915a93b62472`
- `baseline` plugin SHA-256 `bac0f4a262eb06b78196a495ece96ff0cc2edaa84a25a74a88ec9832eeb84992`
- Toolchain: Go 1.27.1 linux/amd64; golangci-lint 2.13.2
- Report: [`2026-09-13-go-refactor-routing-load-n1-sonnet-5-medium.json`](2026-09-13-go-refactor-routing-load-n1-sonnet-5-medium.json)
  (SHA-256 `9e023327c85ada3f75a2839d8872dcea7725d0ffee159ca58a096106263216c5`);
  traces [`….traces.tar.gz`](2026-09-13-go-refactor-routing-load-n1-sonnet-5-medium.traces.tar.gz)
  (SHA-256 `8ac425286eff18a8c2951ae6e2e2b31f3a292454fca67594b6c785e70f6a35ca`),
  one `traces/<arm>-<fixture>-r0.jsonl` per session
- Cost: $0.162 no-skill, $0.745 reference, $1.393 baseline; $2.30 in all

```bash
go run ./cmd/abrun -corpus refactor -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <worktree of 831859b> -arms no-skill,reference,baseline \
  -tasks dispatch,pricing,report,store -n 1 -j 3 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-13-go-refactor-routing-load-n1-sonnet-5-medium.json
```

No session had a shell tool; the plugin's edit hook ran gofmt, vet,
`go fix -diff`, the package tests and golangci-lint after each edit in the
two skilled arms. The prompt hook named `go-code-refactor` in 8/8 skilled
sessions (the `prompted` file in each session's routing state), and every
skilled session loaded it as its first `Skill` call.

## What each session loaded

The column the edits target. "Before" is the set in context when the first
`.go` file was edited or written; "gate" is what the routing gate named, each
time followed by the load and a retry that landed (`edited=true` in every
session); "end" is the set at the end of the session. `Skill` messages counts
the assistant turns that carried a `Skill` call.

| Arm | Fixture | Before the first `.go` edit | Gate named | End | `Skill` messages |
|---|---|---|---|---|---|
| `reference` | `dispatch` | `go-code-refactor` | — | same | 1 |
| `reference` | `pricing` | `go-code-refactor` | — | same | 1 |
| `reference` | `report` | `go-code-refactor` | — | same | 1 |
| `reference` | `store` | `go-code-refactor`, `go-code` | `go-style-core`, `go-error-handling` (old gate, via `go-code`) | + both | 4 |
| `baseline` | `dispatch` | `go-code-refactor`, `go-style-core`, `go-error-handling` | `go-context` | + `go-context` | 5 |
| `baseline` | `pricing` | `go-code-refactor` | `go-style-core`, `go-testing` on the test file; `go-error-handling` on `pricing.go` | + all three | 5 |
| `baseline` | `report` | `go-code-refactor`, `go-style-core` | `go-error-handling` | + `go-error-handling` | 3 |
| `baseline` | `store` | `go-code-refactor`, `go-style-core`, `go-error-handling` | — | + `go-linting`, `go-testing`, `go-code-review` after the edit | 6 |

- `go-style-core` in context before the first edit: 3/4 baseline sessions
  from the text alone, 4/4 once the gate named it in `pricing`; 0/4
  reference sessions before the edit, 1/4 by the end (the `store` session
  loaded `go-code` on its own and met the old gate).
- An owner skill (`go-error-handling`, `go-context`, `go-testing`): 4/4
  baseline sessions by the end, 2/4 before the first edit; 1/4 reference.
- The gate fired five times in three baseline sessions and once in the
  reference arm. No fire repeated a name, no session stalled: each block was
  followed by the loads and the same edit against the unchanged file.
- "Every `Skill` call in one message" did not take. The baseline sessions
  carried their `Skill` calls on 3, 5, 5 and 6 separate turns; the reference
  sessions that loaded only `go-code-refactor` on 1. Each load on its own turn
  re-reads the context: cache-read tokens 345K–787K a session against
  122K–305K in the reference arm.
- The `pricing` baseline session took `SAFETY-NET.md`'s advice on a package
  with no tests and wrote a 167-line characterization test first
  (`pricing_test.go.model` in the kept tree; `model_tests=pass`), which is
  why its first `.go` write was a test file and why the gate named
  `go-testing`. The reference session on the same fixture wrote no test.
- The `store` baseline session loaded `go-linting`, `go-testing` and
  `go-code-review` after its edit, for the verification and report steps.
  None of them changed the code.

## Refactor corpus

| Arm | Δlines | Δfuncs | Line gate | Golden | Lint clean | $/run |
|---|---|---|---|---|---|---|
| `no-skill` | −2.8 | +1.50 | 2/4 | 4/4 | 2/4 | 0.041 |
| `reference` | −16.0 | +0.75 | 3/4 | 4/4 | 3/4 | 0.186 |
| `baseline` | −13.2 | +0.75 | 3/4 | 4/4 | 3/4 | 0.348 |

| Fixture | `no-skill` | `reference` | `baseline` |
|---|---|---|---|
| `dispatch` | +6, 3 funcs | −10 | −11 |
| `pricing` | −23, 1 type | −50, 1 type | −49, 1 type, characterization test |
| `report` | +12, 3 funcs | +13, 3 funcs | +13, 3 funcs |
| `store` | −6 | −17 | −6 |

Correctness and lint do not move: golden 4/4 in every arm, lint-clean 3/4 in
both skilled arms (the `dispatch` residue is the pre-existing `revive`
stutter finding, 6→1 in both), `go fix` pending 1→0 on `dispatch` in both.
`report` grew by the same three helpers in both skilled arms and failed the
line gate in both, as every skilled Sonnet session on this fixture has.
`dispatch` and `pricing` are level.

The 2.8-line corpus gap is the `store` session. The baseline session kept the
`if m.cache != nil` and `if m.remote != nil` guards and flattened only the
miss path (−6); the reference session deleted the guards as redundant — a
read from a nil map returns `ok == false` — and folded each lookup into one
`if v, ok := m.cache[k]; k != "" && ok` (−17). Both preserve behavior. This is
the shape the [morning's run](2026-09-13-go-arch-current-go-three-arm-n1-sonnet-5-medium.md)
had (−6 against −15), and the two `store` sessions that had `go-style-core`
in context before their first edit kept the guards while the two that did
not (this run's reference session loaded it after its first edit) deleted
them. Two of two is a pattern to watch on `store` at n=5, not a claim.

## Cost

| Arm | Fixture | Turns | Cache-create tokens | Cache-read tokens | Output tokens | $ |
|---|---|---|---|---|---|---|
| `reference` | `dispatch` | 7 | 19,177 | 121,702 | 3,327 | 0.135 |
| `reference` | `pricing` | 14 | 24,924 | 237,990 | 6,527 | 0.214 |
| `reference` | `report` | 10 | 20,426 | 202,784 | 4,136 | 0.165 |
| `reference` | `store` | 16 | 32,604 | 305,141 | 3,883 | 0.231 |
| `baseline` | `dispatch` | 21 | 43,374 | 489,244 | 5,466 | 0.327 |
| `baseline` | `pricing` | 30 | 56,175 | 787,086 | 15,019 | 0.533 |
| `baseline` | `report` | 16 | 28,222 | 345,237 | 5,777 | 0.241 |
| `baseline` | `store` | 19 | 47,793 | 354,794 | 2,912 | 0.292 |

$0.348 against $0.186 a session (+87%). Without `pricing`, whose baseline
session wrote and ran a characterization test, $0.287 against $0.177 (+62%).
The added cost is the loaded skills' payload (cache-create +50% to +130%)
and the extra turns each load took (cache-read ×2 to ×4), not longer code.

## What this run does and does not establish

- The edits change what a refactor session has in context: `go-style-core`
  4/4 against 1/4 and an owner skill 4/4 against 1/4 by the end of the
  session, 3/4 and 2/4 before the first edit from the text alone. The gate
  covers the rest and never deadlocked.
- The one-message load instruction is not followed by Sonnet 5 medium; the
  loads land one per turn, and the cost of the edits is +87% a session at
  n=1. Whether that price buys anything is not visible on this corpus: lines,
  golden and lint are within one-repetition noise or tied.
- No correctness or lint regression: golden 4/4 and lint-clean 3/4 in both
  skilled arms.
- The `go-code-review` path is unmeasured here: no corpus drives a review
  that goes on to edit, so its load step and the gate after it are covered by
  `TestRoutingGate` only.

## Next

An n=5 two-arm run (`reference`, `baseline`) on `store` and `pricing`
decides whether the guard-keeping reading on `store` is the loaded skill or
noise and sizes the test-writing cost on `pricing`. If the loads keep landing
one per turn, either the one-message wording goes or the gate's block message
becomes the place that enforces it.
