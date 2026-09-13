# GPT-5.6-Luna medium control benchmarks

**Date:** 2026-09-13
**Runner:** Codex
**Model:** `gpt-5.6-luna`
**Reasoning effort:** `medium`
**Plugin tree:** `d1b71249acd8a06b089970e02cb14ed9ebdcf9f1` plus documentation-only local changes
**Arms:** isolated `no-skill` control and current-tree `baseline`; no wording variants
**Repetitions:** 5 per fixture and arm; deterministic job-order seed 1; four concurrent sessions

Raw results: [refactor JSON](2026-09-13-go-refactor-control-gpt-5-6-luna-medium-n5.json) and [implement JSON](2026-09-13-go-implement-control-gpt-5-6-luna-medium-n5.json). Both commands used `-keep`; each result records its retained scratch tree and trace path.

The harness first builds the model output, then runs an independent hidden golden test. A golden failure is excluded from size/function means; it remains in the pass-rate column. Codex reports token counts rather than USD, so no dollar-cost comparison is available.

## Refactoring existing code

| Arm | Valid / attempts | Golden | Mean production-line delta | Mean new functions | `go-code-refactor` read |
| --- | ---: | ---: | ---: | ---: | ---: |
| No skill | 20 / 20 | 20 / 20 | -4.2 | 1.80 | 0 / 20 |
| Baseline skills | 20 / 20 | 20 / 20 | -11.0 | 0.05 | 20 / 20 |

On this four-fixture corpus, the skilled arm removed 6.8 more physical production lines and added 1.75 fewer functions per valid session, with identical 20/20 golden outcomes. The effect is fixture-dependent: baseline was smaller on `dispatch` and `report`, tied on `store`, and slightly larger on `pricing`. This is evidence that the skills help this tested refactoring corpus, not a guarantee for every refactor.

## Writing new code

| Arm | Valid / attempts | Golden | Mean production-line delta | Mean new functions | `go-code` read |
| --- | ---: | ---: | ---: | ---: | ---: |
| No skill | 27 / 30 | 27 / 30 | +68.4 | 0.93 | 0 / 30 |
| Baseline skills | 26 / 30 | 26 / 30 | +56.4 | 0.15 | 30 / 30 |

The skilled outputs were 12.0 lines and 0.78 functions smaller among valid sessions, but their independent golden rate was lower: 26/30 versus 27/30. The difference is concentrated in `fetch`, which passed 1/5 with skills and 2/5 without; both arms missed `TestGetOrderGivesUpWithinAttempts`. Therefore this run does **not** establish a new-code benefit, despite the smaller valid outputs.

## Reproduction

```bash
cd evals
go run ./cmd/abrun -runner codex -model gpt-5.6-luna -effort medium \
  -corpus refactor -arms no-skill,baseline -n 5 -j 4 -keep \
  -out ../docs/evidence/2026-09-13-go-refactor-control-gpt-5-6-luna-medium-n5.json
go run ./cmd/abrun -runner codex -model gpt-5.6-luna -effort medium \
  -corpus implement -arms no-skill,baseline -n 5 -j 4 -keep \
  -out ../docs/evidence/2026-09-13-go-implement-control-gpt-5-6-luna-medium-n5.json
```
