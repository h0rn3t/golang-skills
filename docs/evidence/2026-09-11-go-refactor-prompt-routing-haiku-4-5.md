# Prompt Routing — refactor corpus, three arms, Haiku 4.5 (n=3)

The run the [Haiku 4.5 control](2026-09-10-go-refactor-control-haiku-4-5.md)
asked for. That control reached `go-code-refactor` in 1 of 4 skilled sessions
and said a wording comparison on this tier measures the router until that
changes. The `baseline` arm carries the `UserPromptSubmit` hook that names
the router from the prompt text, plus the `go-code` description edit, which
this corpus does not exercise; `reference` is the committed tree at
`f4f373f`, whose routing is the control's.

## Run

- Finished: 2026-09-11 11:20 UTC
- Runner: `claude` 2.1.267, with `--include-hook-events`
- Model: `claude-haiku-4-5-20251001`, no effort flag (the CLI default, as in
  the control)
- Seed: `1`; `-j 4`
- Corpus: `refactor`
- Fixtures: `dispatch`, `pricing`, `report`, `store`
- Arms: `no-skill`, `reference`, `baseline`
- Repetitions: 3 per fixture and arm, 36 sessions total
- `reference`: a `git worktree` of `f4f373f`, plugin SHA-256
  `ec904c52ac70fb646eaafe3f271548cf129279a8628fd53ab58ff36dc6542bf1`
- `baseline`: the working tree at `f4f373f` plus the uncommitted routing
  edits, plugin SHA-256
  `8c7bc9f028080f74648d33156baf5c22512bb85e23150392407bd48c8bb0d5ad`
- Toolchain: Go 1.27.1 linux/amd64
- Raw report: [`2026-09-11-go-refactor-prompt-routing-haiku-4-5.json`](2026-09-11-go-refactor-prompt-routing-haiku-4-5.json)
  (SHA-256 `897e8c867cc4240faf88cd6dcda2e8fca7ad65b0e03d479cb22915d615582c7a`)
- Session transcripts:
  [`2026-09-11-go-refactor-prompt-routing-haiku-4-5.traces.tar.gz`](2026-09-11-go-refactor-prompt-routing-haiku-4-5.traces.tar.gz)
  (SHA-256 `6611ca88333931309af1a97c35b8972fffc54b29668176356307896c22115f7e`),
  one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $0.4400 control, $0.7660 reference, $1.1264 baseline — **1.74x** and
  **2.56x** the control

```bash
go run ./cmd/abrun -corpus refactor -runner claude -model claude-haiku-4-5-20251001 \
  -reference-root <worktree of f4f373f> -arms no-skill,reference,baseline \
  -n 3 -j 4 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-11-go-refactor-prompt-routing-haiku-4-5.json
```

All 36 sessions completed without a CLI error, changed the fixture, and
stayed out of the repository checkout. No shell in any arm.

## Routing

| Arm | prompt hook fired | `go-code-refactor` loaded | first tool call | turn of the load |
|---|---|---|---|---|
| `no-skill` | — | — | `Glob` 12/12 | — |
| `reference` | — | **4/12** (`pricing` 3/3, `store` 1/3, `dispatch` 0/3, `report` 0/3) | `Glob` 12/12 | 8th–13th |
| `baseline` | **12/12**, naming `go-code-refactor` | **12/12** | **`Skill go-code-refactor` 12/12** | 2nd–3rd |

The control's reading reproduces and sharpens: the description alone reaches
this tier in a third of sessions and only on two fixtures, after the model
has read the code. With the hook the router is the first action in every
session, before any file is read. The hook classified all twelve prompts as
a refactor; the note is in the trace as a `hook_response` event.

## Results

Production-line delta against the shipped fixture over valid runs — build,
hidden golden, and the session's own tests when it wrote any. The `report`
row for `baseline` is empty because no session there was valid; its raw
deltas are given in the text.

| Fixture | No skill | Reference | Baseline |
|---|---|---|---|
| `dispatch` | −8, −11, −10 → −9.7 | 0, +2, −10 → −2.7 (two sessions +2 funcs) | −2, −9, −4 → −5.0 |
| `pricing` | −42, −38, −39 → −39.7 | −41, −37 → −39 (one excluded: own test failed) | −42 (two excluded: own tests failed) |
| `report` | +17, +19, +23 → +19.7, +3 to +4 funcs | +25, +13 → +19, +3 to +4 funcs (one excluded: golden failed) | none valid; raw +2, +2, +4 with 0–1 funcs; **golden failed 2/3**, own test failed 1/3 |
| `store` | −3, −6, −2 → −3.7 | −3, −3, −3 | −3, −3, −3 |

| Arm | valid | golden | own test failed | Δfuncs (valid) | line gate | counts reported | turns | median report chars | $ / run |
|---|---|---|---|---:|---|---|---:|---:|---:|
| `no-skill` | 12/12 | **12/12** | 0 | 0.92 | 9/12 | 1/12 | 17.3 | 682 | 0.0367 |
| `reference` | 10/12 | 11/12 | 1 | 1.10 | 8/12 | 5/12 | 24.0 | 828 | 0.0638 |
| `baseline` | 7/12 | 10/12 | 3 | 0.00 | 9/12 | 11/12 | 35.6 | 1252 | 0.0939 |

## What the router changed on this tier

**`report`: the skill stops the helpers and breaks the width.** Unaided,
Haiku grows the package by 17–23 lines and three or four helpers in every
session, and every session passes the golden. With `go-code-refactor` loaded
the diff is +2 to +4 lines — early returns, one `formatAmount` helper — and
two of three sessions fail `TestGoldenRender`: `%9d.%02d` became `%9s` around
the helper's `"%d.%02d"` string, which pads the whole amount to nine columns
instead of the dollars, and the aligned text output moves. The third session
kept the width and instead failed the test it wrote itself. The reference
session that routed nowhere on `report` made the same `formatAmount` cut and
failed the same way, so the extraction is Haiku's own move; what the skill
removed was the three other helpers around it, and what it did not supply is
the check that would have caught the width. BEHAVIOR-TRAPS names formatting
width; a session without a shell has only the file to read it against.

**`pricing`: level, with the model's own tests wrong.** All three arms fold
the plan literals into one table and take about forty lines out. Two baseline
sessions and one reference session then wrote a test that fails against
their own passing code — the shape the Sonnet 5 `report` exclusions had — and
are excluded on that rule. The skill routed 3/3 here in the reference arm
too, so this fixture never measured routing.

**`dispatch` and `store`: the control is as good or better.** On `dispatch`
the unaided sessions take 9.7 lines out and the routed sessions 5.0, with the
reference arm's two unrouted sessions adding two helpers each; on `store` the
three arms write the same three-line deletion. The Sonnet 5 result on
`dispatch` (−15.8, arms not overlapping) does not appear on this tier at n=3.

**Cost 2.56x against 1.74x**, turns 35.6 against 24.0: the routed session
runs the audit, the concision gate and the report, and this tier takes twice
the turns of the control to do it. Counts are reported in 11/12 sessions
against 5/12 and 1/12, the instruction followed.

## What this changes

- **Routing on Haiku 4.5 is closed**: 12/12 against 4/12 and 1/4, as the
  first action. The hook is what the control asked for; kept.
- **The routed skill does not yet pay on this tier.** Correctness 10/12
  against 12/12 unaided (Fisher p = 0.48 at this n), size level or worse on
  three fixtures, and the one fixture where the skill changes the shape of the
  diff is the one where it breaks behavior. This is a statement about
  `go-code-refactor` under Haiku 4.5 without a shell, measurable for the first
  time because the router now reaches every session; it is not a reason to
  remove the hook, which only delivers the skill.
- **README**: nothing moves. n=3 per cell, and the Haiku row stays "routing
  1/4" until a run with a shell or a rule for the width trap says otherwise.
- **Next**: the `report` width failure needs a rule the model can apply
  without running anything — a `Printf` verb whose width applies to a
  different operand after an extraction is a behavior change, and
  BEHAVIOR-TRAPS should show that pair; the same corpus with a shell so the
  golden-shaped failures fail in front of the model; and n≥5 on `dispatch`
  before reading its direction.
