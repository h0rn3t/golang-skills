# Plugin evals

Cases for `claude plugin eval` (Claude Code 2.1.269+). They live here, not in
`evals/`, because `evals/` is the Go module with the structural suite; pass
`--eval-dir plugin-evals` on every run.

The cases are drawn from the `validation` set of `evals/evals.json`, which
`evalrun` keeps running; neither tool reads the other's format. Each run
starts in an empty directory, so a case that needs code pastes it into the
prompt (`trigger-refactor-split-function`) rather than relying on
`--scaffold`, which runs outside the sandbox.

- `trigger-*`: a `tool_used: Skill` grader per expected skill, `arm: both`,
  so the no-plugin arm scores 0 and Δ is the trigger rate. Keep
  `max_turns: 10`: at 3, Opus 5.5 spent its turns on `Glob` before loading
  go-code (a false miss), and the saving was 6–12%, since a run's cost is
  the cache write of the skills it loads, not its later turns.
- `negative-*`: no `go-*` skill may fire (`min: 0, max: 0`).
- `quality-*`: an `llm` grader over the eval's assertions (weight 3), a cheap
  `regex` check, and the skill-fired indicator.

```bash
# One case, one run, no baseline arm
claude plugin eval . --eval-dir plugin-evals --case trigger-context-todo --runs 1 --ablation none --judge-model haiku

# By tag: trigger, negative, quality
claude plugin eval . --eval-dir plugin-evals --tag trigger -j 4 --judge-model haiku --no-publish

# Full suite with a budget
claude plugin eval . --eval-dir plugin-evals -j 4 --threshold 0.8 --max-cost-usd 10 --judge-model haiku --no-publish
```

Runs make real model calls. Output goes to `plugin-evals/results/`, which is
ignored. The judge model is pinned because its default can change with the
CLI. `--json` records `costUsd` and `turns` per run; which skills loaded is
only in the run's trace, kept with `--keep-temp`.
