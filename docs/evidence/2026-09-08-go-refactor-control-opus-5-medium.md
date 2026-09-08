# Go Refactor Skill Control — Opus 5 at medium effort on the Claude CLI

Run launched from this checkout. Means, population standard deviations, exact
permutation p-values, Welch intervals, and costs were computed from the JSON
report; the t critical values behind the intervals come from a small table, so
treat the interval bounds as approximate to the tenth of a line.

## Run

- Finished: 2026-09-08 18:34 UTC
- Runner: `claude` 2.1.260
- Model: `claude-opus-5`, reasoning effort `medium`
- Seed: `1`
- Corpus: `refactor`
- Fixtures: `dispatch`, `pricing`, `report`, `store`
- Arms: `no-skill`, `baseline`
- Repetitions: 5 per fixture and arm, 40 sessions total
- Baseline plugin SHA-256: `eefdafe8c65819a574f7f804e0553f07b13ee80d38003ddf4da4b06f32d330bc`
- Plugin source: working tree at `ad11c3d`; the plugin subdirectories were clean,
  the only modified files were `evals/cmd/abrun/main.go` and its test, which the
  arm does not contain
- Raw report: [`2026-09-08-go-refactor-control-opus-5-medium.json`](2026-09-08-go-refactor-control-opus-5-medium.json)
- Raw report SHA-256: `691b4fb312c0b698f1111b6d4c76c0f46732d61606effda8262734ac5198a6e1`
- Cost: $3.21 control, $9.44 baseline, $12.65 total

```bash
go run ./cmd/abrun -arms no-skill,baseline -n 5 -j 4 -seed 1 \
  -model claude-opus-5 -effort medium -verbose \
  -out ../docs/evidence/2026-09-08-go-refactor-control-opus-5-medium.json
```

This is the first run to set a reasoning effort on the `claude` runner. `abrun`
rejected `-effort` outside codex and copilot until the commit that carries this
report; the claude CLI has `--effort <level>`, and `claudeSession` now passes it
through, so the `effort` field in the JSON is a condition the CLI was actually
asked for rather than a label.

All 40 sessions completed without a CLI error, built, and passed their hidden
golden tests. One baseline run (`report` rep 3) failed the test it wrote itself
and is excluded from every number below, leaving 19 valid baseline runs against
20 control runs. `go-code-refactor` fired in 19 of 20 baseline sessions; the
twentieth (`store` rep 0) loaded `go-code` alone. Sessions ran under
`--restricted` with the tool set limited to `Skill,Read,Glob,Grep,Edit,Write`,
so no session had a shell — most final messages say so and ask the operator to
run `go test`.

## Results

Line values are mean delta with population standard deviation over the valid
runs. Negative is less production code; the effect column is baseline minus
no-skill, so negative favors the skill. `p` is an exact two-sided permutation
test over that fixture's runs.

| Fixture | No skill | Skill | Effect | 95% CI | p |
|---|---:|---:|---:|---|---:|
| `dispatch` | −9.2 ± 2.7 | −5.8 ± 2.2 | +3.4 | −0.8 … +7.6 | 0.10 |
| `pricing` | −29.6 ± 1.5 | −38.8 ± 7.5 | **−9.2** | −19.8 … +1.4 | 0.06 |
| `report` | +22.4 ± 8.0 | +17.2 ± 2.2 | −5.1 | −16.7 … +6.4 | 0.40 |
| `store` | −12.0 ± 3.5 | −11.2 ± 4.4 | +0.8 | −5.9 … +7.5 | 0.81 |
| All runs | −7.10 (n=20) | −11.05 (n=19) | **−3.95** | | |

| Structural additions across the valid runs | No skill (20) | Skill (19) | Change |
|---|---:|---:|---:|
| Types | 12 | 9 | −3 |
| Interfaces | 1 | 0 | −1 |
| Functions | 26 | 10 | −61.5% |
| Pattern-name hits | 0 | 0 | no difference |
| Exported symbols | 0 | 0 | no difference |
| Branches | −130 | −147 | −17 |

The skill produced a test file in 9 of 19 valid runs against 1 of 20 in the
control. `abrun` hides model-authored tests before applying the golden test, so
this is evidence of test creation only — and the one excluded run shows the
other side of it: a test written without a shell to run it can simply be wrong.

Cost is 2.94x: $0.1605 per control run against $0.4722 per baseline run.

## Which fixture separates, and why

`pricing` carries the corpus difference at −9.2 lines, and it is the only
fixture whose permutation p is near the conventional line (14/252 = 0.0556,
unadjusted). Both arms already remove a lot — the control averages −29.6 — but
the skilled arm goes further and with far more spread (±7.5 against ±1.5).
Final session messages describe the same transformation in the skilled runs:
the three parallel name ladders (`Price`, `SeatCost`, `Discount`) collapse into
one ordered `planSpec` table with a single lookup, and the 4×3 discount
threshold ladder becomes one method over the record.

`report` grows in both arms (+22.4 control, +17.2 skilled). The fixture's bait
is a `csv` flag consulted three times, and folding it into one `layout` value
costs a type and some lines on a package this small, so a reduction was never
the likely outcome here; the skilled arm is 5.1 lines lower with a much tighter
spread, but the interval spans zero.

`store` and `dispatch` do not separate. `dispatch` moves against the skill
(+3.4, interval −0.8 … +7.6): the control removes more lines than the skilled
arm on this model, which reverses the sonnet-5 control's −4.4. `store` is flat
at +0.8, where the sonnet-5 control found its one significant result (−8.2).
Neither fixture is currently earning its place against this model at this
effort, and that is worth a focused re-run before either is used as evidence.

## Interpretation

The corpus mean favors the plugin by 3.95 production lines per run, and the
structural direction is the familiar one: 26 new functions against 10, 12 new
types against 9. The direction agrees with both earlier controls, but the
fixture that produces it has changed — `pricing` here, `store` on sonnet 5,
`report` on the earlier Opus run. Four fixtures were examined without a
preregistered hypothesis, so a four-test Bonferroni adjustment puts `pricing` at
p = 0.22; the Welch intervals are unadjusted and all four span zero. Treat the
corpus mean as a direction and `pricing` as an exploratory signal, not as a
confirmed effect.

Two limits are structural rather than statistical. This is a two-arm run, so it
measures having the plugin against not having it and isolates no particular
wording; a before/after claim needs `-reference-root`. And the 2.94x cost cannot
be attributed mostly to skill reading — token usage, thinking, and cache figures
are absent from the JSON.
