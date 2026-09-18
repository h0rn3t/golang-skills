# The idiom card as a workflow step, and the edit-hook record — Sonnet 5 medium, two runs at n=2

The 1.20.1 `go-code` names the idiom card `go-style-core/references/CURRENT-GO.md`
only in a Resource Routing bullet; step 2 of its workflow ("Load
`go-style-core`, read the code, check for a shell") does not. The skill-authoring
guide's "Use workflows for complex tasks" says a mandatory action is a numbered
step, and the 2026-09-14 counter-run that reported the card read in 0/32 Sonnet
sessions left no report in this repository. These two runs put the read rate on
record and measure two wordings of the step against 1.20.1, on the same three
fixtures. The same edit also moved what the plugin's edit hook runs, and how a
no-shell report carries its output, into one place: `go-style-core` "The Edit
Hook Record", which `go-code` and `go-code-refactor` route to.

| Arm | Tree |
|---|---|
| `reference` | `git worktree` of `89d702a` (release 1.20.1) |
| `baseline`, run 1 | the working tree with the card as its own numbered step: step 2 "Load `go-style-core`", step 3 "Read the idiom card whole", step 4 "Read the code, check for a shell", owners at step 5, eight steps in all; plus the edit-hook record, the `go-code-refactor` Orient move, and the negation rewrites of the same change (plugin SHA-256 `868eac2c…`) |
| `baseline`, run 2 | the working tree with the card folded back into step 2 — "Load `go-style-core`, read the idiom card and the code, check for a shell", the Skill call and the card `Read` in the same message — six steps as in 1.20.1; everything else as in run 1 (plugin SHA-256 `bfcfed52…`) |

## Runs

Run 1 (card as its own step):

- Finished: 2026-09-18 16:44 UTC
- Runner: `claude` 2.1.267; model `claude-sonnet-5`, effort `medium`; seed `1`
- Corpus `implement`, fixtures `roster`, `feed`, `gateway`, arms `reference`, `baseline`, 2 repetitions per fixture and arm, 12 sessions, `-j 4`, 0 CLI errors
- `reference` plugin SHA-256 `d478dd31988660fcb8934eba8e56e7e378221b321b278073006d4ecb4af19ca0`; `baseline` `868eac2c01f7dfd350c56ada023dba7ac78e5dca37c709c17792ccec169b2466`
- Report: [`2026-09-18-go-implement-card-step-hook-record-roster-feed-gateway-n2-sonnet-5-medium.json`](2026-09-18-go-implement-card-step-hook-record-roster-feed-gateway-n2-sonnet-5-medium.json) (SHA-256 `e6fb7aaaf5cbabf9986b2f1390b4082529d25a1fb3335fb6f62656d41bd99290`); traces [`….traces.tar.gz`](2026-09-18-go-implement-card-step-hook-record-roster-feed-gateway-n2-sonnet-5-medium.traces.tar.gz) (SHA-256 `e6b4317a35c3795503d6e59289161b6294e4452a710c45bfe9175db0c93ffda6`), one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $4.52

Run 2 (card in step 2):

- Finished: 2026-09-18 16:52 UTC; seed `2`; otherwise as run 1
- `baseline` plugin SHA-256 `bfcfed52ecd8273c265ff05738e65a40716a3caa9b16a78dde590c732cd14ef0`
- Report: [`2026-09-18-go-implement-card-in-step2-hook-record-roster-feed-gateway-n2-sonnet-5-medium.json`](2026-09-18-go-implement-card-in-step2-hook-record-roster-feed-gateway-n2-sonnet-5-medium.json) (SHA-256 `6b985987ccc5f9ce494e7f52cad6111b71feb012cf9b1625beb97101de00f2b0`); traces [`….traces.tar.gz`](2026-09-18-go-implement-card-in-step2-hook-record-roster-feed-gateway-n2-sonnet-5-medium.traces.tar.gz) (SHA-256 `b481450486e7f530ec18b6f53f8dbc534af07d6bc0ebe5cb3d6cb2bb10909993`)
- Cost: $4.34

```bash
go run ./cmd/abrun -corpus implement -tasks roster,feed,gateway -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <worktree of 89d702a> -arms reference,baseline \
  -n 2 -j 4 -seed 1 -timeout 10m -keep -verbose -out <run 1 json>     # seed 2 for run 2
```

Toolchain Go 1.27.1 linux/amd64, golangci-lint 2.13.2 on PATH, so the edit
hook ran `gofmt`, `go vet`, `go fix -diff`, the package tests, and the linter
after every `.go` edit. No session had a shell tool. The event columns below
are counted from the traces: a Skill turn is one assistant message carrying
one or more `Skill` calls; a gate block is a `PreToolUse` hook response with
exit 2; a hook finding is a `PostToolUse:Edit|Write` response with exit 2; a
card read is a `Read` of `CURRENT-GO.md`.

## The reading

Run 1:

| Arm | Valid | Golden | Lint clean | Δlines | Skill calls/s | Skill turns/s | Card `Read` | Gate blocks | Hook findings | Assistant turns/s | $/session |
|---|---|---|---|---|---|---|---|---|---|---|---|
| `reference` | 6/6 | 6/6 | 4/6 | +48.7 | 4.33 | 2.17 | 0/6 | 1 | 26 | 12.5 | 0.371 |
| `baseline` | 6/6 | 6/6 | 4/6 | +44.5 | 4.50 | 2.67 | 0/6 | **6** | 28 | 14.7 | 0.381 |

Run 2:

| Arm | Valid | Golden | Lint clean | Δlines | Skill calls/s | Skill turns/s | Card `Read` | Gate blocks | Hook findings | Assistant turns/s | $/session |
|---|---|---|---|---|---|---|---|---|---|---|---|
| `reference` | 6/6 | 6/6 | 4/6 | +44.3 | 4.50 | 2.00 | 0/6 | 1 | 24 | 12.8 | 0.337 |
| `baseline` | 6/6 | 5/6 | 4/6 | +46.7 | 4.50 | 2.00 | 0/6 | 0 | 38 | 15.5 | 0.387 |

Per session, run 1:

| Arm | Fixture | Rep | Golden | Lint | Δlines | Skills loaded (in load order) | Skill turns | Card `Read` | Gate blocks | Hook findings | Turns | $ |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `reference` | `feed` | 0 | pass | 0 | +34 | go-code, go-style-core, go-testing, go-error-handling, go-data-structures | 2 | no | 0 | 2 | 8 | 0.250 |
| `reference` | `feed` | 1 | pass | 0 | +41 | go-code, go-style-core, go-testing, go-error-handling, go-data-structures | 2 | no | 0 | 2 | 9 | 0.351 |
| `reference` | `gateway` | 0 | pass | 0 | +82 | go-code, go-style-core, go-http, go-error-handling, go-testing | 3 | no | 0 | 6 | 18 | 0.572 |
| `reference` | `gateway` | 1 | pass | 0 | +94 | go-code, go-style-core, go-http, go-error-handling, go-testing | 2 | no | 0 | 10 | 21 | 0.602 |
| `reference` | `roster` | 0 | pass | 5 | +22 | go-code, go-style-core | 2 | no | 1 | 1 | 10 | 0.192 |
| `reference` | `roster` | 1 | pass | 5 | +19 | go-code, go-style-core, go-interfaces, go-testing | 2 | no | 0 | 5 | 9 | 0.262 |
| `baseline` | `feed` | 0 | pass | 0 | +40 | go-code, go-style-core, go-testing, go-error-handling | 3 | no | 2 | 2 | 14 | 0.292 |
| `baseline` | `feed` | 1 | pass | 0 | +34 | go-code, go-style-core, go-testing, go-error-handling | 3 | no | 2 | 4 | 16 | 0.306 |
| `baseline` | `gateway` | 0 | pass | 0 | +77 | go-code, go-http, go-error-handling, go-security, go-style-core, go-testing | 3 | no | 1 | 6 | 16 | 0.531 |
| `baseline` | `gateway` | 1 | pass | 0 | +84 | go-code, go-style-core, go-http, go-error-handling, go-testing | 2 | no | 0 | 7 | 18 | 0.577 |
| `baseline` | `roster` | 0 | pass | 5 | +15 | go-code, go-style-core, go-interfaces, go-testing | 3 | no | 1 | 5 | 11 | 0.279 |
| `baseline` | `roster` | 1 | pass | 5 | +17 | go-code, go-style-core, go-testing, go-interfaces | 2 | no | 0 | 4 | 13 | 0.304 |

Per session, run 2:

| Arm | Fixture | Rep | Golden | Lint | Δlines | Skills loaded (in load order) | Skill turns | Card `Read` | Gate blocks | Hook findings | Turns | $ |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `reference` | `feed` | 0 | pass | 0 | +41 | go-code, go-style-core, go-testing, go-error-handling | 2 | no | 1 | 3 | 12 | 0.268 |
| `reference` | `feed` | 1 | pass | 0 | +40 | go-code, go-style-core, go-testing, go-error-handling, go-data-structures | 2 | no | 0 | 3 | 13 | 0.321 |
| `reference` | `gateway` | 0 | pass | 0 | +77 | go-code, go-style-core, go-http, go-error-handling, go-testing | 2 | no | 0 | 3 | 14 | 0.443 |
| `reference` | `gateway` | 1 | pass | 0 | +67 | go-code, go-http, go-error-handling, go-testing, go-style-core | 2 | no | 0 | 7 | 17 | 0.491 |
| `reference` | `roster` | 0 | pass | 5 | +23 | go-code, go-style-core, go-testing, go-interfaces | 2 | no | 0 | 5 | 10 | 0.256 |
| `reference` | `roster` | 1 | pass | 5 | +18 | go-code, go-style-core, go-interfaces, go-testing | 2 | no | 0 | 3 | 11 | 0.240 |
| `baseline` | `feed` | 0 | pass | 0 | +40 | go-code, go-style-core, go-testing, go-error-handling | 2 | no | 0 | 5 | 15 | 0.348 |
| `baseline` | `feed` | 1 | pass | 0 | +39 | go-code, go-style-core, go-error-handling, go-data-structures, go-testing | 2 | no | 0 | 9 | 19 | 0.384 |
| `baseline` | `gateway` | 0 | pass | 0 | +74 | go-code, go-style-core, go-http, go-error-handling, go-testing, go-security | 2 | no | 0 | 8 | 17 | 0.493 |
| `baseline` | `gateway` | 1 | **fail** | 0 | +92 | go-code, go-style-core, go-http, go-error-handling, go-testing | 2 | no | 0 | 10 | 22 | 0.651 |
| `baseline` | `roster` | 0 | pass | 5 | +18 | go-code, go-style-core, go-testing, go-interfaces | 2 | no | 0 | 5 | 13 | 0.272 |
| `baseline` | `roster` | 1 | pass | 5 | +17 | go-code, go-style-core, go-data-structures | 2 | no | 0 | 1 | 7 | 0.172 |

**The card is read in 0/24 sessions**, whatever the wording: a Resource
Routing bullet plus a sentence in "Write Current Go" (1.20.1), a numbered step
of its own (run 1), or the first clause of step 2 with "in the same message"
(run 2). On this model at this effort the text route is exhausted; the
2026-09-14 figure now has a recorded counterpart. `roster` cannot show the cost
of that: both arms wrote `slices.Sort` and no older form in 4/4 sessions, as on
2026-09-13.

**The separate step cost routing, and folding it back recovered it.** In run
1 the baseline skipped steps 2, 3 and 5 as a block in both `feed` sessions —
`go-code`, glob, read the code, read `go.mod`, then `Write` the contract test,
which the gate refused for `go-style-core` and `go-testing`, and later a
second edit for `go-error-handling` — six gate blocks against one, 2.67 Skill
turns against 2.17. In run 2 the two arms have the same shape in every
session: `go-code` on one turn, `go-style-core` and the owners together on the
next, 2.00 Skill turns each, zero gate blocks in the baseline. The reference's
one block per run is a `roster` or `feed` session that edited before loading
`go-style-core`.

**The hook record changes the report.** Every session had no shell, so a
check line reading `gofmt pass` claims a run that did not happen. With the
1.20.1 text the reference wrote `(hook)` on every check it claimed in 1/11
sessions that printed a checks line (`gateway` r0 of run 2) and left the rest
bare or half-marked (`gofmt pass · vet pass · test pass (hook)`). With "The
Edit Hook Record" the baseline marked every claimed check `(hook)` in 8/11 and
wrote `lint unavailable (no shell)` where the linter had not printed (`feed`
r0 of run 1); the lines left are `feed` r0 (bare), `feed` r1 (half-marked)
and `roster` r1 (bare) of run 2. One baseline session in each run printed no
checks line at all.

**Correctness and size did not move; cost and turns leaned against the
baseline.** Golden 11/12 against 12/12: `gateway` r1 of run 2 answered
`GET /accounts` on a nil list with `null`, the fixture's known trap, in a
session that wrote the body before its contract table and spent nine edits on
the test file's lint findings. Lint-clean 8/12 in both arms, the five findings
in every `roster` session being the neighbor's `legacy.go`. Over the two runs
the baseline took 15.1 assistant turns a session against 12.7 and cost $0.384
against $0.354 (+8%); the hook-findings column, 66 against 50, says where the
turns went — more edit-fix rounds, mostly in test files — and at n=2 per
fixture the sign of the cost difference is within the run-to-run movement of
the reference itself ($0.371 to $0.337).

## What this establishes

- Sonnet 5 medium does not read the idiom card from any wording of
  `go-code` or `go-style-core` tried so far: 0/24 here, on top of the
  unrecorded 0/32. A read that must happen on this model has to come from
  the host — the prompt hook's note, the routing gate, or an injected
  payload — not from the skill text.
- A mandatory action as a step of its own is not free: splitting step 2
  into three made the model skip the load steps as a block in 2/6 sessions
  and let the gate do the routing. The bundled form ("Load `go-style-core`,
  read the idiom card and the code") keeps the 1.20.1 routing shape.
- The edit-hook record, stated once in `go-style-core`, is followed in form:
  `(hook)` on every claimed check in 8/11 baseline sessions against 1/11.
  Whether a bare `gofmt pass` from a shell-less session misleads a reader
  is the reason the rule exists; the rate is the measurement.
- Nothing here separates the hook record from the card wording or the
  negation rewrites: the baseline carried all three. The cost lean (+8%)
  and the extra hook findings are the open reading, on this model the Opus
  5 run of the same day is the other half.

## Next

- Count card reads on every routed run from now on (the trace grep is in
  the report's method paragraph); the number is the gate for any further
  text change to the card's routing.
- If the card is to be read on Sonnet 5, test the host routes one at a time:
  the prompt hook naming the card's installed path, then the routing gate
  requiring the `Read` once per session, each against this baseline at n≥3.
- The hook-record wording needs its own arm before it is credited: a
  reference of this tree minus "The Edit Hook Record" against this baseline,
  on the same three fixtures.
