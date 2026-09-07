# Go Refactor Skill Control — GPT-5.3-Codex-Spark (xhigh) on Codex CLI

## Run

- Finished: 2026-09-07 14:45 EEST
- Runner: `codex` 0.153.4
- Model: `gpt-5.3-codex-spark`, reasoning effort `xhigh`
- Seed: `1`
- Fixtures: `dispatch`, `pricing`, `report`, `store`
- Arms: `no-skill`, `baseline`
- Repetitions: 5 per fixture and arm, 40 sessions total
- Baseline plugin SHA-256: `6443a658f74cd465b98f0065cdb7c117bcde52c64af1acc1a5ef7f3a67369d3c`
- Plugin source: working tree at `5dcb55d`
- Raw report: [`2026-09-07-go-refactor-control-codex-spark-xhigh.json`](2026-09-07-go-refactor-control-codex-spark-xhigh.json)
- Raw report SHA-256: `9f66a61c4e89291313c292b061119062064a68baa3f63669bf1fcaf4a3189bdd`

```bash
go run ./cmd/abrun -runner codex -model gpt-5.3-codex-spark -effort xhigh \
  -arms no-skill,baseline -n 5 -j 4 -seed 1 \
  -out ../docs/evidence/2026-09-07-go-refactor-control-codex-spark-xhigh.json
```

All 40 sessions reached the model and changed the fixture, and none referenced
the repository checkout. Seven did not produce a measurable refactor; they are
counted in the failure table below rather than averaged in. A `go-*` skill was
read in 18 of 20 baseline sessions: `go-code-refactor` in 17, `go-style-core` in
11 and `go-code` in 2.

This is the first run on the `codex` runner, added for it, and one runner
difference changes what the `skill` column means. Codex has no skill tool.
Skills reach the model as a listing in the system prompt and a skill fires when
the model opens its `SKILL.md` through the shell, so `abrun` scores a skill from
the command that read it — and only from the command, never from its output,
since a loaded `SKILL.md` names other skills by path. Codex also keeps its own
tool set, so these sessions have a shell and could run `go test` on their own
work, which matches the [MiniMax M3 control](2026-09-07-go-refactor-control-minimax-m3.md)
and differs from the [Opus 5](2026-09-07-go-refactor-control-opus5.md) and
[MAI-Code-1.1-Flash](2026-09-07-go-refactor-control-mai-code-1.1-flash.md) files.
The plugin's PostToolUse hook and `go-verify` subagent do not apply here.

## Results

### Correctness

This is the second run of the refactor corpus in which correctness moves at all;
every previous one was tied at 20/20.

| Outcome | No skill | Skill |
|---|---:|---:|
| Left a Go file that does not parse | 3 | 1 |
| Changed the exported signature | 2 | 1 |
| **Usable refactor** | **15/20** | **18/20** |

The three unparseable control runs are all `pricing` (reps 1, 3, 4) and the one
skilled run is `report` rep 0: the session applied a partial patch and stopped,
leaving `expected declaration, found '}'` or `expected '(', found SeatCost`. The
signature failures are the same defect in both arms — `store.go` returning a
bare `string` where the exported API is `(string, bool)` — caught by the hidden
golden test in control reps 1 and 2 and skilled rep 4.

Fisher's exact on 15/20 against 18/20 gives a two-sided p of 0.41, so this is a
direction, not a result. It is worth recording because the failure mode is one
this corpus has never seen before: an "ultra-fast" model whose own system prompt
tells it to prefer mistakes over exploration and not to run tests will hand back
code that does not compile, in either arm.

### Code size

Line values are mean delta with sample standard deviation over the runs that
produced a usable refactor. Negative is less production code; the effect column
is baseline minus no-skill, so negative favors the skill.

| Fixture | No skill | Skill | Effect |
|---|---:|---:|---:|
| `dispatch` | +5.2 ± 7.5 (n=5) | +17.0 ± 15.9 (n=5) | +11.8 |
| `pricing` | −32.0 ± 8.5 (n=2) | −30.2 ± 11.8 (n=5) | not comparable |
| `report` | +30.6 ± 3.8 (n=5) | +31.8 ± 10.4 (n=4) | +1.1 |
| `store` | +9.0 ± 0.0 (n=3) | +4.2 ± 3.6 (n=4) | −4.8 |

The corpus-wide means the summary prints — +9.5 against +4.3 — must not be read
as an effect favoring the skill. The arms have different valid sets, and
`pricing` is the fixture with the large negative deltas: the control lost three
of its five `pricing` runs to unparseable output, so the fixture that pulls a
mean down contributes two runs to the control and five to the skilled arm. The
per-fixture rows are the honest comparison, and unweighted across the three
fixtures where both arms have runs they average to +2.7 lines against the skill.

| Structural additions across valid runs | No skill | Skill |
|---|---:|---:|
| Types | 4 | 6 |
| Interfaces | 0 | 0 |
| Functions | 35 | 43 |
| Pattern-name hits | 0 | 0 |

Neither arm wrote a test file in any run. Codex reports token counts rather than
dollars, so the `$/run` column is empty for both arms.

## Interpretation

The size effect does not reproduce, and unlike the MAI-Code-1.1-Flash run the
reason is not a saturated trap. On `report` the unaided model grows the package
by 30.6 lines — close to Opus 5's +33.4 and above MiniMax M3's +27.4 — so this
model takes the bait and the room to move it was there. The skill did not move
it: +31.8 with the skill, and the skilled arm added more functions and more
types than the control across the corpus. On `dispatch` it was clearly worse,
+17.0 against +5.2.

That is a genuine negative result for the wording on this model, and it is the
first one. Two candidate explanations are visible in the transcripts and neither
is settled by this run. The skill was read in 18 of 20 sessions, so it is not a
triggering failure. But `go-style-core` was pulled in alongside `go-code-refactor`
in 11 of them, where the other runners' sessions mostly loaded the refactor skill
alone — and this model's own system prompt tells it that every tool call is
expensive, that it should prefer mistakes to exploration, and that it should not
run tests. A model under that instruction reading two skill files may be
spending its budget on the reading rather than the refactor.

Correctness is the part of this run worth acting on. Both arms shipped code that
does not compile, in 4 of 40 sessions, and changed an exported signature in 3
more. `abrun` catches all seven, so nothing invalid reached the averages, but a
model that fails this way is a poor fit for a corpus whose whole score is a line
delta on a behavior-preserving refactor.

## What is missing

The implementation corpus was not run. All 40 sessions failed with
`403 Forbidden` on `wss://chatgpt.com/backend-api/codex/responses` — the
account's Codex quota was exhausted by the refactor corpus — and not one of them
edited a file, so the run produced no data and its report was discarded rather
than published as an empty table. Re-running it needs quota only:

```bash
go run ./cmd/abrun -corpus implement -runner codex \
  -model gpt-5.3-codex-spark -effort xhigh \
  -arms no-skill,baseline -n 5 -j 4 -seed 1 \
  -out ../docs/evidence/2026-09-07-go-implement-control-codex-spark-xhigh.json
```

That gap matters for the cross-model picture: `gateway` is the one fixture where
the skill has separated arms on every model that could run it, and this model's
habit of leaving code that does not compile is exactly what the implementation
corpus's golden test is built to catch.
