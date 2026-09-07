# Go Refactor Skill Control — MAI-Code-1.1-Flash on GitHub Copilot CLI

## Run

- Finished: 2026-09-07 13:45 EEST
- Runner: `copilot` 1.0.83
- Model: `mai-code-1.1-flash`
- Seed: `1`
- Fixtures: `dispatch`, `pricing`, `report`, `store`
- Arms: `no-skill`, `baseline`
- Repetitions: 5 per fixture and arm, 40 sessions total
- Baseline plugin SHA-256: `6443a658f74cd465b98f0065cdb7c117bcde52c64af1acc1a5ef7f3a67369d3c`
- Plugin source: working tree at `5dcb55d`
- Raw report: [`2026-09-07-go-refactor-control-mai-code-1.1-flash.json`](2026-09-07-go-refactor-control-mai-code-1.1-flash.json)
- Raw report SHA-256: `c6b77d671d86230ca77c20fe01ef69bef50cd0963f59cf49858a0cbb12d1512f`

```bash
go run ./cmd/abrun -runner copilot -model mai-code-1.1-flash \
  -arms no-skill,baseline -n 5 -j 4 -seed 1 \
  -out ../docs/evidence/2026-09-07-go-refactor-control-mai-code-1.1-flash.json
```

All 40 sessions completed, built, and passed their hidden golden tests. Every
session changed the fixture and none referenced the repository checkout, so all
40 count as evidence. A `go-*` skill fired in 19 of 20 baseline sessions:
`go-code-refactor` in 18, `go-code-review` in 2 and `go-code` in 1; the session
that loaded nothing was `store` rep 2.

This is the first run on the `copilot` runner, added for it. Copilot discovers
personal skills from `COPILOT_HOME`, so each arm gets its own and the control
arm's is empty — without that it would load the operator's globally installed
`go-*` skills and stop being a control. The session is restricted to
`skill,view,create,edit,grep,glob`, which is the claude runner's
`Skill,Read,Glob,Grep,Edit,Write` under copilot's names, so these sessions have
no shell and cannot run `go test` on their own work; that matches the
[Opus 5 control](2026-09-07-go-refactor-control-opus5.md) and differs from the
[MiniMax M3 control](2026-09-07-go-refactor-control-minimax-m3.md). Two
conditions still differ from the Opus 5 file: the repository's PostToolUse
gofmt/vet hook and the `go-verify` subagent apply only under claude, where a
copilot arm home carries skills alone. Within this file both arms are identical
apart from the skills.

## Results

Line values are mean delta with sample standard deviation over five runs.
Negative is less production code; the effect column is baseline minus no-skill,
so negative favors the skill.

| Fixture | No skill | Skill | Effect |
|---|---:|---:|---:|
| `dispatch` | −0.6 ± 6.8 | +5.2 ± 8.2 | +5.8 |
| `pricing` | −35.4 ± 4.2 | −38.0 ± 4.5 | −2.6 |
| `report` | +16.2 ± 2.6 | +15.6 ± 5.9 | −0.6 |
| `store` | +7.6 ± 7.1 | +9.4 ± 7.7 | +1.8 |
| All runs | −3.05 | −1.95 | +1.10 |

No fixture separates the arms. Welch intervals for the mean difference at 95%
are approximately −5.2 to +16.8 lines on `dispatch`, −9.0 to +3.8 on `pricing`,
−7.3 to +6.1 on `report` and −9.0 to +12.6 on `store`; every one includes zero,
and so does the corpus-wide interval of −14.7 to +16.9. On `report` — the trap
fixture that carries the effect in both published runs — the median moved from
16 lines to 14.

| Structural additions across 20 runs | No skill | Skill | Change |
|---|---:|---:|---:|
| Types | 1 | 0 | −1 |
| Interfaces | 0 | 0 | no difference |
| Functions | 37 | 47 | +27.0% |
| Pattern-name hits | 0 | 0 | no difference |

The skill produced a test file in 5 of 20 runs against 0 of 20 without it.
`abrun` hides model-authored tests before applying the independent golden test,
so this is evidence of test creation, not evidence that those tests compile or
assert the right behavior.

Copilot bills a session in premium requests and AI credits rather than dollars,
so the `$/run` column is empty for both arms and this file carries no cost
comparison.

## Interpretation

The effect does not reproduce on this model. Opus 5 and MiniMax M3 both cut
mean growth on `report` by roughly 16 lines; here the difference is −0.6, and
the corpus-wide direction is marginally against the skill at +1.1 lines per run.

The reason is visible in the control column rather than the skill column: the
trap is saturated. Without any skill, MAI-Code-1.1-Flash grew `report` by 16.2
lines on average, against +33.4 under Opus 5 and +27.4 under MiniMax M3. Roughly
half to a third as much structure was there to remove, and it declared one type
across 20 unaided runs and no interfaces at all. This model does not take the
`report` bait, so the fixture cannot measure whether the skill would stop it.
The same holds for `pricing`, the hardest bait in the corpus, where both arms
already delete about 35 lines. Correctness is tied at 20/20 build and golden
passes in both arms, so the run is no evidence about behavior preservation
either.

The one place the arms clearly diverge is helper count, and it runs the wrong
way: the skilled arm added 47 functions across 20 runs against the control's 37.
That is the mirror image of the MiniMax M3 control, where the skill suppressed
function growth by 48.4%. Combined with the extra test files, the picture is a
skilled arm that does more work per session without that work showing up as a
shorter diff.

Read this as a fixture-sensitivity result, not a skill result. The corpus rule
is that a fixture earns its place only when `no-skill` is measurably worse than
`baseline`; on this model none of the four clears that bar, so nothing here
supports or contradicts a claim about the wording. It says nothing about
MAI-Code-1.1-Flash on tasks with more room to over-engineer, which is what the
implementation corpus was built to supply.
