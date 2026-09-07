# Go Refactor Skill Control — MiMo v2.5 Pro on opencode

## Run

- Finished: 2026-09-07 15:37 EEST
- Runner: `opencode` 1.18.29
- Model: `opencode-go/mimo-v2.5-pro`
- Seed: `1`
- Fixtures: `dispatch`, `pricing`, `report`, `store`
- Arms: `no-skill`, `baseline`
- Repetitions: 5 per fixture and arm, 40 sessions total
- Baseline plugin SHA-256: `6443a658f74cd465b98f0065cdb7c117bcde52c64af1acc1a5ef7f3a67369d3c`
- Plugin source: working tree at `5dcb55d`
- Raw report: [`2026-09-07-go-refactor-control-mimo-v2.5-pro.json`](2026-09-07-go-refactor-control-mimo-v2.5-pro.json)
- Raw report SHA-256: `9cb146ec78fb66d115c9caeae6936f0ae63470cab5a82ef25d5887c3255022ea`

```bash
go run ./cmd/abrun -runner opencode -model opencode-go/mimo-v2.5-pro \
  -arms no-skill,baseline -n 5 -j 4 -seed 1 \
  -out ../docs/evidence/2026-09-07-go-refactor-control-mimo-v2.5-pro.json
```

Thirty-nine of 40 sessions completed; every one of them changed the fixture and
none referenced the repository checkout. One baseline session (`report` rep 4)
died on `database is locked` — opencode's own SQLite contention between the four
concurrent sessions sharing an arm home at `-j 4`, not a model failure — and is
excluded from the arm rather than counted against it. The `go-code-refactor`
skill fired in 18 of 20 baseline sessions and `go-style-core` in 1; one session
loaded nothing.

The runner conditions are the [MiniMax M3 control's](2026-09-07-go-refactor-control-minimax-m3.md):
opencode keeps its own tool set, so these sessions have a shell and can run
`go test` on their own work, and the plugin's PostToolUse hook and `go-verify`
subagent do not apply.

## Results

Line values are mean delta with sample standard deviation over the valid runs.
Negative is less production code; the effect column is baseline minus no-skill,
so negative favors the skill.

| Fixture | No skill | Skill | Effect |
|---|---:|---:|---:|
| `dispatch` | −6.0 ± 6.0 (n=5) | −3.0 ± 3.5 (n=5) | +3.0 |
| `pricing` | −24.2 ± 13.9 (n=4) | −24.8 ± 12.7 (n=5) | −0.6 |
| `report` | +16.4 ± 7.7 (n=5) | +5.2 ± 6.6 (n=4) | **−11.1** |
| `store` | −4.0 ± 3.5 (n=5) | −5.0 ± 3.1 (n=5) | −1.0 |
| All runs | −3.42 | −7.53 | −4.11 |

On `report`, the trap fixture, the skill cut mean growth by 68.0% and the median
from 19 lines to 4.5. Its Welch interval for the mean difference at 95% is −22.2
to −0.1 lines, the only fixture here whose interval excludes zero; `dispatch`
(−4.2 to +10.2), `store` (−5.8 to +3.8), `pricing` (−21.3 to +20.2) and the
corpus-wide interval (−15.2 to +7.0) all include it. With `n=5` per cell and
four fixtures tested at once these are descriptive intervals, not a
preregistered result, and `report`'s barely clears the line.

| Structural additions across valid runs | No skill | Skill | Change |
|---|---:|---:|---:|
| Types | 6 | 3 | −50.0% |
| Interfaces | 0 | 0 | no difference |
| Functions | 20 | 8 | −60.0% |
| Pattern-name hits | 0 | 0 | no difference |

The skill produced a test file in 2 of 19 runs against 0 of 19 without it.
`abrun` hides model-authored tests before applying the independent golden test,
so this is evidence of test creation, not evidence that those tests assert the
right behavior.

Loading the skill roughly doubled session cost: $0.183 for 20 no-skill runs
against $0.391 for 20 baseline runs, $0.575 for the corpus.

## Interpretation

This is the cleanest replication of the effect so far. The trap is live — the
unaided model grows `report` by 16.4 lines — and the skill removes most of it,
which is the pattern Opus 5 and MiniMax M3 showed and which
[MAI-Code-1.1-Flash](2026-09-07-go-refactor-control-mai-code-1.1-flash.md), whose
trap is saturated, broke.

Both structural mechanisms reproduce at once, which no earlier run managed.
Opus 5 cut new types by 60% but function growth by 37.5%; MiniMax M3 cut
functions by 48.4% and added a type. Here types fall 50% and functions 60%, so
on this model the skill suppresses both the premature abstraction and the
helper sprawl rather than trading one for the other.

`dispatch` is again the fixture that goes the other way, +3.0 lines with the
skill, as it did on every published run: Opus 5 +2.0, MiniMax M3 +0.8,
MAI-Code-1.1-Flash +5.8. Four models pointing the same direction on one fixture
is worth a look at that fixture rather than another repetition.

Correctness is effectively tied — 19/20 golden against 20/20, the single control
failure being `pricing` rep 2 — so this run is no evidence about behavior
preservation. It compares the complete current skill with no skill; it does not
measure the improvement from a previous revision of the skill.

## A harness note

The `database is locked` failure is the first of its kind in this corpus and it
is a property of `-j 4` under opencode, where every arm's sessions share one
HOME and therefore one SQLite file. It cost one session here and one in the
[implementation run](2026-09-07-go-implement-control-mimo-v2.5-pro.md) on the
same model. Lower `-j`, or a per-run home, would remove it; at two sessions in
80 it did not change any conclusion above.
