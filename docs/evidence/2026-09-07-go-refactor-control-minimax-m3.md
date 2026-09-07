# Go Refactor Skill Control — MiniMax M3 on opencode

## Run

- Finished: 2026-09-07 11:14 EEST
- Runner: `opencode` 1.18.29
- Model: `opencode-go/minimax-m3`
- Seed: `1`
- Fixtures: `dispatch`, `pricing`, `report`, `store`
- Arms: `no-skill`, `baseline`
- Repetitions: 5 per fixture and arm, 40 sessions total
- Baseline plugin SHA-256: `6443a658f74cd465b98f0065cdb7c117bcde52c64af1acc1a5ef7f3a67369d3c`
- Plugin source: working tree at `d90b355` with the release-0.9.1 edits applied
- Raw report: [`2026-09-07-go-refactor-control-minimax-m3.json`](2026-09-07-go-refactor-control-minimax-m3.json)
- Raw report SHA-256: `628b582521023b8671e08ad3a74994f443c9f39c1b7a4466939cac755a43d706`

```bash
go run ./cmd/abrun -runner opencode -model opencode-go/minimax-m3 \
  -arms no-skill,baseline -n 5 -j 4 -seed 1 -keep \
  -out ../docs/evidence/2026-09-07-go-refactor-control-minimax-m3.json
```

All 40 sessions completed, built, and passed their hidden golden tests. Every
session changed the fixture and none referenced the repository checkout, so all
40 count as evidence. The `go-code-refactor` skill fired in 19 of 20 baseline
sessions; `go-naming` and `go-style-core` were the only other `go-*` skills
recorded, and the one baseline session that loaded nothing was `store` rep 3.

Two conditions differ from the [Opus 5 control](2026-09-07-go-refactor-control-opus5.md)
and both matter for any comparison across the two files. The opencode sessions
keep opencode's own tool set, so they have a shell and can run `go test` on
their own work, where the claude sessions are restricted to
`Skill,Read,Glob,Grep,Edit,Write`. The repository's PostToolUse gofmt/vet hook
applies only under claude. Within this file both arms are identical apart from
the skills, which is what the effect column measures.

## Results

Line values are mean delta with sample standard deviation over five runs.
Negative is less production code; the effect column is baseline minus no-skill,
so negative favors the skill.

| Fixture | No skill | Skill | Effect |
|---|---:|---:|---:|
| `dispatch` | −7.2 ± 4.0 | −6.4 ± 3.7 | +0.8 |
| `pricing` | −35.6 ± 9.0 | −42.4 ± 8.4 | −6.8 |
| `report` | +27.4 ± 9.2 | +11.0 ± 8.7 | **−16.4** |
| `store` | −2.2 ± 4.5 | −11.6 ± 7.3 | **−9.4** |
| All runs | −4.40 | −12.35 | **−7.95** |

On `report`, the skill reduced mean code growth by 59.9% and the median from
28 to 13 lines. Welch intervals for the mean difference at 95% are
approximately −29.5 to −3.3 lines on `report` and −18.2 to −0.6 on `store`;
`pricing` (−19.5 to +5.9) and `dispatch` (−4.9 to +6.5) both include zero. With
`n=5` per cell and four fixtures tested at once these are descriptive
intervals, not a preregistered result.

| Structural additions across 20 runs | No skill | Skill | Change |
|---|---:|---:|---:|
| Types | 5 | 6 | +1 |
| Interfaces | 0 | 0 | no difference |
| Functions | 31 | 16 | −48.4% |
| Pattern-name hits | 0 | 0 | no difference |

The skill also produced a test file in 11 of 20 runs against 1 of 20 without
it. `abrun` hides model-authored tests before applying the independent golden
test, so this is evidence of test creation, not evidence that those tests
compile or assert the right behavior.

Loading the skill roughly tripled session cost: $0.547 for 20 no-skill runs
against $1.391 for 20 baseline runs, $1.94 for the corpus.

## Interpretation

The result replicates the direction of the Opus 5 control on a much smaller
model, and on the trap fixture it replicates the size too: −16.4 lines here
against −16.6 there. The two arms differ from each other by more overall
(−7.95 against −4.15 lines per run), and the benefit spreads to `store`, which
was inside the noise under Opus 5. `dispatch` is again marginally worse with
the skill, and `report` remains the fixture that carries the effect.

Type growth does not reproduce: Opus 5 cut new types by 60%, while MiniMax M3
added one more with the skill than without. What does reproduce is helper-count
suppression, at −48.4% functions. This model's failure mode on `report` is
extra functions rather than extra types, and the skill is what holds them down.

Correctness is tied, at 20/20 build and golden passes in both arms, so the run
provides no evidence that the skill improves behavior preservation. It compares
the complete current skill with no skill; it does not measure the improvement
from the previous revision of the skill, and it says nothing about MiniMax M3
under the claude runner's restricted tool set.
