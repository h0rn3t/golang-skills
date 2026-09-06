# Go Refactor Skill Control — Opus 5

## Run

- Finished: 2026-09-07 00:45 EEST
- Claude Code: 2.1.261
- Model: `claude-opus-5`
- Seed: `20260907`
- Fixtures: `dispatch`, `pricing`, `report`, `store`
- Arms: `no-skill`, `baseline`
- Repetitions: 5 per fixture and arm, 40 sessions total
- Baseline plugin SHA-256: `42fb1bc30c5314e8c4cc7d4a2e20628260ee44bf68d91826a0412b632e898ee7`
- Raw report: [`2026-09-07-go-refactor-control-opus5.json`](2026-09-07-go-refactor-control-opus5.json)
- Raw report SHA-256: `f1a92a958e8e6bf0a216146f67042b8ea3b6a29e76bb765c4e49bab3945a49b2`

```bash
go run ./cmd/abrun -arms no-skill,baseline -n 5 -j 4 \
  -model claude-opus-5 -seed 20260907 \
  -out ../docs/evidence/2026-09-07-go-refactor-control-opus5.json
```

All 40 sessions completed, built, and passed their hidden golden tests. The
`go-code-refactor` skill fired in all 20 baseline sessions and no other `go-*`
skill was recorded.

## Results

Line values are mean delta with sample standard deviation over five runs.
Negative is less production code; the effect column is baseline minus no-skill,
so negative favors the skill.

| Fixture | No skill | Skill | Effect |
|---|---:|---:|---:|
| `dispatch` | −10.4 ± 1.5 | −8.4 ± 3.2 | +2.0 |
| `pricing` | −28.8 ± 2.5 | −29.2 ± 3.4 | −0.4 |
| `report` | +33.4 ± 11.3 | +16.8 ± 7.9 | **−16.6** |
| `store` | −9.6 ± 5.1 | −11.2 ± 3.5 | −1.6 |
| All runs | −3.85 | −8.00 | **−4.15** |

On `report`, the skill reduced mean code growth by 49.7% and the median from
34 to 16 lines. A Welch interval for the mean difference is approximately
−31.1 to −2.1 lines at 95%, but `n=5` and this was not a preregistered single
hypothesis, so treat that interval as descriptive evidence rather than a final
benchmark.

| Structural additions across 20 runs | No skill | Skill | Change |
|---|---:|---:|---:|
| Types | 15 | 6 | −60.0% |
| Interfaces | 1 | 0 | −1 |
| Functions | 32 | 20 | −37.5% |
| Pattern-name hits | 0 | 0 | no difference |

The skill also produced one test file in 15 of 20 runs; no-skill produced none.
`abrun` hides model-authored tests before applying the independent golden test,
so this is evidence of test creation, not evidence that those tests compile or
assert the right behavior.

## Interpretation

The measurable benefit is concentrated in the ambiguous `report` fixture: the
skill strongly suppresses helper/type/interface growth there. `pricing` and
`store` show small gains; `dispatch` is two lines worse with the skill. This
supports “less over-engineering on a known trap,” not a universal percentage
improvement for all refactors.

Correctness is tied in this corpus: both arms achieved 20/20 build and golden
passes. The run therefore provides no evidence that the skill improves behavior
preservation. It also compares the complete current skill with no skill; it does
not measure the improvement from the previous revision of the skill.
