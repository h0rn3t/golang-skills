# Prompt Routing Smoke — implementation corpus, three arms, Sonnet 5 medium (n=1)

A first look at the two routing edits made after the
[Opus 5 workflow run](2026-09-11-go-implement-newcode-workflow-opus-5-medium.md),
which left the router description as its open item: 5 of 18 skilled sessions
loaded no skill at all, every one on `feed` or `catalog`, whose prompt says
only "Implement the Go package in ./<dir>". One repetition per fixture and arm
supports no claim about size, correctness rate, or cost, and none is made
below. What this run reads is whether the new hook fires under the harness and
what the model does in the turn after it.

The edits in the `baseline` arm:

- **`go-code` description** also names implementing a Go package, function,
  or handler whose declarations and documentation already exist — write the
  bodies, fill in a stub, replace `panic("not implemented")` — even when the
  request names only the package.
- **`hooks/go-prompt-routing.sh`**, a `UserPromptSubmit` hook. When a prompt
  asks for Go work it adds one note to the model's context naming the router
  to load before the first edit: `go-code-refactor` for refactor, clean-up,
  or simplify wording, `go-code` for anything else. Once per skill per
  session; silent when the session already loaded the skill, when the prompt
  invokes a go-* skill by name, or when the prompt carries no work verb. It
  never blocks.

The `reference` arm is the committed tree at `f4f373f` (release 1.9.0), which
has neither edit, so the two are confounded in this run: a difference between
the skilled arms is the pair, not either one.

## Run

- Finished: 2026-09-11 10:28 UTC
- Runner: `claude` 2.1.267
- Model: `claude-sonnet-5`, reasoning effort `medium`
- Seed: `1`; `-j 4`
- Corpus: `implement`
- Fixtures: `catalog`, `feed`, `gateway`, `ledger`
- Arms: `no-skill`, `reference`, `baseline`
- Repetitions: 1 per fixture and arm, 12 sessions total
- `reference`: a `git worktree` of `f4f373f`, plugin SHA-256
  `ec904c52ac70fb646eaafe3f271548cf129279a8628fd53ab58ff36dc6542bf1`
- `baseline`: the working tree at `f4f373f` plus the uncommitted edits above,
  plugin SHA-256
  `8c7bc9f028080f74648d33156baf5c22512bb85e23150392407bd48c8bb0d5ad`
- Toolchain: Go 1.27.1 linux/amd64
- Raw report: [`2026-09-11-go-implement-prompt-routing-smoke-sonnet-5-medium.json`](2026-09-11-go-implement-prompt-routing-smoke-sonnet-5-medium.json)
  (SHA-256 `49b21e0ef32859a7d5ecca91bc56dbf9086a7e1e1069e99d483ca1bb943b339f`)
- Session transcripts:
  [`2026-09-11-go-implement-prompt-routing-smoke-sonnet-5-medium.traces.tar.gz`](2026-09-11-go-implement-prompt-routing-smoke-sonnet-5-medium.traces.tar.gz)
  (SHA-256 `fd9fc4fbb264802975cae86c983f0c755e253837877999743158ee8fcd776673`),
  one `traces/<arm>-<fixture>-r0.jsonl` per session
- Cost: $0.1700 control, $1.3106 reference, $1.2261 baseline — **7.71x** and
  **7.21x** the control

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <worktree of f4f373f> -arms no-skill,reference,baseline \
  -n 1 -j 4 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-11-go-implement-prompt-routing-smoke-sonnet-5-medium.json
```

All 12 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. The tool set was `Skill,Read,Glob,Grep,Edit,Write`,
so no session had a shell. `go-code` fired in 4/4 reference and 4/4 baseline
sessions; on Sonnet 5 the description was never the limiting factor, so the
Opus 5 gap this edit targets is not measured here.

## Results

Production-line delta against the shipped fixture, declared functions and
types added, and the hidden golden verdict. One session per cell.

| Fixture | No skill | Reference (1.9.0) | Baseline (edits) |
|---|---|---|---|
| `catalog` | pass, +26 | pass, +19 | pass, +19 |
| `feed` | pass, +44 | pass, +35 | pass, +30 |
| `gateway` | pass, +85, 0 funcs | **fail** `TestEmptyListIsJSONArray` (`null`), +83, +5 funcs, +1 type; its own contract test failed too | pass, +84, +2 funcs (`filterActive`, `writeJSON`) |
| `ledger` | pass, +43 | pass, +39 | pass, +36 |

| Arm | golden | first tool call | turn of the `go-code` load | `go-linting` loads (no shell) | skills / session | turns | median report chars | $ / run |
|---|---|---|---|---|---:|---:|---:|---:|
| `no-skill` | 4/4 | `Glob` or `Read` 4/4 | — | — | 0 | 8.5 | 248 | 0.0425 |
| `reference` | 3/4 | `Glob` 4/4 | 2nd–3rd, after the files were read | 1/4 | 4.25 | 32.8 | 772 | 0.3277 |
| `baseline` | 4/4 | **`Skill go-code` 4/4** | **1st, before any read** | 4/4 | 5.25 | 26.8 | 895 | 0.3065 |

Skill loads over four sessions — reference: `go-code` 4, `go-style-core` 4,
`go-testing` 4, `go-error-handling` 3, `go-http` 1, `go-linting` 1. Baseline:
`go-code` 4, `go-style-core` 4, `go-testing` 4, `go-linting` 4,
`go-error-handling` 3, `go-data-structures` 2, `go-http` 1.

## Did the hook fire

The hook's note does not appear in the `stream-json` transcript: the host adds
`UserPromptSubmit` stdout to the model's context without echoing it, so the
trace is silent about it (a run that wants it in the trace passes
`--include-hook-events`). The evidence is the state the hook writes. Under the
Claude CLI a plugin's `CLAUDE_PLUGIN_DATA` resolves to
`~/.claude/plugins/data/golang-skills-inline/`, and its `routing/<session>/`
directory — the same one `go-code-routing.sh` writes its `loaded` list to —
holds a `prompted` file naming `go-code` for **all four baseline sessions**
(`36143fd9`, `86135219`, `f832fe4d`, `7ed66dab`, matched to the traces by
session id) and for **none** of the eight other sessions. The hook ran once
per session, classified each of the four prompts as Go work to implement, and
named `go-code`.

A separate one-turn probe on Haiku 4.5 with the same plugin and prompt, run
by hand after the harness, had no `Skill` tool at all and still opened with
"I'll start by loading the go-code skill": the note reached the model.

## What the model did in the turn after

**The first tool call moved.** Every baseline session's first action is
`Skill go-code`, at line 2 or 3 of the trace, before a single file is read.
Every reference session's first action is `Glob` over the fixture, and
`go-code` is loaded two or three turns later, at lines 8–12, once the model
has seen the stub. On Sonnet 5 both arms reach the router in the end, so this
run cannot say whether the hook or the description closes the Opus 5 gap; it
says the note is read and acted on immediately.

**Turns fall, cost is level.** 26.8 turns against 32.8 and $0.307 against
$0.328 a session, with more skills loaded (5.25 against 4.25). Loading the
router first spares the exploratory turns the reference arm spends before it;
at n=1 the difference is a direction.

**`go-linting` loads with no shell: 4/4 against 1/4.** The rule that keeps
`go-linting` unread without a shell tool held in the reference arm here and in
the Opus 5 run (0/9); the baseline arm loaded it in every session. The four
transcripts show the same move as before — the workflow reaches step 6 and
the model loads the gate's owner to "select checks" — and nothing in the two
edits mentions `go-linting`. Whether an early `go-code` load changes how step 2
is read, or this is n=1 noise, is a question for the next run: the previous
measurement had the same tree at 5/15.

**Correctness: 4/4 against 3/4 against 4/4.** The one failure is the reference
`gateway` session, which rendered an empty account list as `null` and also
failed the contract test it wrote for that case — the shape the
[n=10 run](2026-09-11-go-implement-newcode-plaincode-n10-sonnet-5-medium.md)
recorded, a test that names the clause and cannot run. The baseline `gateway`
session passed every assertion including `HEAD`, with two helpers it named in
its budget line. One session per cell: no rate is claimed.

**Size.** `feed` 30 against 35 and 44, `ledger` 36 against 39 and 43, `catalog`
tied at 19, `gateway` level. Four single sessions, all in the direction the
skills have pointed since the Plain Code retune; nothing here separates the
two skilled arms.

## What this changes

- **The hook works under the harness**: it fires on the corpus prompt, once
  per session, and the model loads the router before reading anything. That
  is the mechanism the Opus 5 and Haiku 4.5 runs asked for; it is not yet
  measured on either model.
- **Nothing in the READMEs**: n=1 per cell is not a control, and Sonnet 5 was
  already routing 15/15 before the edit.
- **Next**: `feed` and `catalog` on Opus 5 medium, three arms, n≥3 — the
  fixtures and model where 5 of 18 sessions loaded nothing; the refactor corpus
  on Haiku 4.5 with `reference` against `baseline`, where the router reached 1
  of 4 sessions (both run the same day, see the
  [Opus 5](2026-09-11-go-implement-prompt-routing-opus-5-medium.md) and
  [Haiku 4.5](2026-09-11-go-refactor-prompt-routing-haiku-4-5.md) reports);
  `--include-hook-events` in the runner so the note is visible in the trace
  (added to `abrun` before those two runs); and the `go-linting` 4/4 watched
  at n≥5 before it is called a regression.
