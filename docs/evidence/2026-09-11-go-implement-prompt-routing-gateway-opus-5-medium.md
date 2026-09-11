# Prompt Routing — `gateway`, three arms, Opus 5 medium (n=3)

The one fixture where routing changes the golden on Opus 5: in the
[workflow run](2026-09-11-go-implement-newcode-workflow-opus-5-medium.md)
both unaided failures were `gateway` sessions answering `HEAD` with 200, and
the [`catalog`/`feed` run](2026-09-11-go-implement-prompt-routing-opus-5-medium.md)
of the same day showed the reference description missing 2 of 6 sessions.
If a `gateway` session skipped the router, it would fail; this run asks
whether that happens and whether the hook prevents it. `reference` is the
committed tree at `f4f373f`; `baseline` adds the `UserPromptSubmit` hook and
the `go-code` description edit.

## Run

- Finished: 2026-09-11 11:51 UTC
- Runner: `claude` 2.1.267, with `--include-hook-events`
- Model: `claude-opus-5`, reasoning effort `medium`
- Seed: `1`; `-j 3`
- Corpus: `implement`, `-tasks gateway`
- Arms: `no-skill`, `reference`, `baseline`
- Repetitions: 3 per arm, 9 sessions total
- `reference`: a `git worktree` of `f4f373f`, plugin SHA-256
  `ec904c52ac70fb646eaafe3f271548cf129279a8628fd53ab58ff36dc6542bf1`
- `baseline`: the working tree at `f4f373f` plus the uncommitted routing
  edits, plugin SHA-256
  `8c7bc9f028080f74648d33156baf5c22512bb85e23150392407bd48c8bb0d5ad`
- Toolchain: Go 1.27.1 linux/amd64
- Raw report: [`2026-09-11-go-implement-prompt-routing-gateway-opus-5-medium.json`](2026-09-11-go-implement-prompt-routing-gateway-opus-5-medium.json)
  (SHA-256 `68d1ee2f5430a29a730a1c859da40ce9d787cd7a67dda56f42147af73a83d7eb`)
- Session transcripts:
  [`2026-09-11-go-implement-prompt-routing-gateway-opus-5-medium.traces.tar.gz`](2026-09-11-go-implement-prompt-routing-gateway-opus-5-medium.traces.tar.gz)
  (SHA-256 `9305f873e3bcc60d49149fc41aebac687654128328fffc546ed430a8731e8f64`),
  one `traces/<arm>-gateway-r<rep>.jsonl` per session
- Cost: $0.9000 control, $3.3106 reference, $3.0150 baseline — **3.68x** and
  **3.35x** the control

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-opus-5 -effort medium \
  -reference-root <worktree of f4f373f> -arms no-skill,reference,baseline \
  -tasks gateway -n 3 -j 3 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-11-go-implement-prompt-routing-gateway-opus-5-medium.json
```

All 9 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. No shell in any arm.

## Results

| Arm | Golden | Δlines | Δfuncs | `HEAD` handled by | empty list as `[]` literal | own test | `HEAD` case in it | prompt hook | `go-code` loaded | turn | $ / run |
|---|---|---|---|---|---|---|---|---|---|---|---:|
| `no-skill` | **0/3** | 96, 123, 120 → 113.0 | 2, 4, 3 | nothing: `GET` patterns only, 3/3 | 3/3 | 2/3 | 0/3 | — | — | — | 0.300 |
| `reference` | 3/3 | 84, 76, 72 → 77.3 | 0, 0, 0 | `HEAD` patterns 3/3 | 2/3 | 3/3 | 3/3 | — | 3/3 | 7th, after `Glob` and reads | 1.104 |
| `baseline` | 3/3 | 85, 80, 78 → 81.0 | 0, 1, 0 | `HEAD` patterns 1/3; method-less patterns with an `r.Method` guard 2/3 | 3/3 | 3/3 | 3/3 | 3/3 | 3/3 | **2nd, first tool call** | 1.005 |

Both skilled arms loaded `go-code`, `go-style-core`, `go-error-handling`,
`go-http`, `go-testing` and `go-data-structures` in every session,
`go-defensive` in 3/3 reference and 2/3 baseline sessions, `go-linting` in
none. Every skilled session reported the `checks:` line and the budget line;
turns 25.7 against 28.7, median report 1562 against 1525 characters, the
control 17.7 turns and 2088 characters.

## Reading

**Unaided Opus 5 fails `gateway` 3/3, the same way each time.** Every control
session registered `GET /healthz`, `GET /accounts`, `GET /accounts/{id}` and
nothing else, and every report says the mux "answers 405 with `Allow`" for
other methods — true for `POST`, false for `HEAD`, which a `GET` pattern
serves. Two of three wrote a test file; neither has a `HEAD` case. With the
workflow run's control, unaided Opus 5 on this fixture is now 1/6.

**Both skilled arms pass 3/3, with a `HEAD` case in every contract test.**
The reference arm registers `HEAD` patterns beside `GET`, the form the
`go-http` example shows; two baseline sessions instead register method-less
patterns and check `r.Method` inside the handler, the control's form on
Sonnet 5, and pass too. Skilled Opus 5 on `gateway` today is 12/12 across the
two runs against 1/6 unaided; correctness here is the `go-http` rule, and the
router is what puts it in context.

**The reference description routed 3/3 on this fixture.** The misses of the
same day were on `catalog` (2/3 unrouted) and, in the workflow run, on `feed`
and `catalog`; on `gateway` the model went to the router at turn 7 in every
reference session, after `Glob` and two reads. Whatever the matcher keys on,
`gateway`'s stub — `net/http` imports, `*http.Server` in the signature — gets
there and `catalog`'s does not. So on this fixture the hook had no miss to
prevent; what it changed is the turn: `Skill go-code` is the first call in
3/3 baseline sessions, before a file is read, against the seventh.

**Size and cost are level.** 81.0 against 77.3 lines, both far under the
control's 113 and its two to four helpers; $1.005 against $1.104 a session.
n=3, directions only.

## What this changes

- **Nothing about the hook's verdict**: it fires 3/3, moves the load to the
  first turn, and costs nothing here. Its value on Opus 5 is on the fixtures
  where the description misses, which this run confirms is not `gateway`.
- **`gateway` on Opus 5 medium is now 12/12 skilled against 1/6 unaided**
  across the day's two runs, the strongest correctness separation in the
  implementation corpus on any model. The [README](../../README.md) Opus 5
  new-code cell stays on the workflow run, which covers three fixtures; this
  run is one fixture and is cited beside it.
- **Next**: a run with the hook removed from `baseline`, on `catalog` at n≥5,
  to say whether the description alone catches the misses the hook now
  covers; and `gateway` with a shell, where the model's own `HEAD` case would
  fail in front of it and the two skilled forms could be compared on what they
  cost to get right.
