# Go Refactor Skill Control — Sonnet 5 on the Claude CLI

Operator-supplied run. Counts, means, population standard deviations, costs,
and exact permutation p-values were recomputed from the JSON. Welch intervals
are retained from the supplied report; CLI version, clean checkout, and command
are operator-reported. No new model sessions were launched for this report.

## Run

- Finished: 2026-09-08 14:40 EEST
- Runner: `claude` 2.1.260
- Model: `claude-sonnet-5`
- Seed: `1`
- Corpus: `refactor`
- Fixtures: `dispatch`, `pricing`, `report`, `store`
- Arms: `no-skill`, `baseline`
- Repetitions: 5 per fixture and arm, 40 sessions total
- Baseline plugin SHA-256: `2b6acb094bcd1cfbb2fba5fabbfcb57c3069b0ad4b3a3731bc596da54e561799`
- Plugin source: working tree at `8d7f571`, clean
- Raw report: [`2026-09-08-go-refactor-control-sonnet-5.json`](2026-09-08-go-refactor-control-sonnet-5.json)
- Raw report SHA-256: `8826095b8753728826d0c0db19b8d65fd93360991f1b2c2d0cae923f82708d35`
- Cost: $1.27 control, $4.17 baseline, $5.44 total

```bash
go run ./cmd/abrun -arms no-skill,baseline -n 5 -j 4 -seed 1 \
  -model claude-sonnet-5 -keep -verbose \
  -out ../docs/evidence/2026-09-08-go-refactor-control-sonnet-5.json
```

All 40 sessions completed without a CLI error, built, and passed their hidden
golden tests. `go-code-refactor` fired in all 20 baseline sessions and `go-code`
in one. Sessions ran under `--restricted` with the tool set limited to
`Skill,Read,Glob,Grep,Edit,Write`, so no session had a shell; most final
messages say so and ask the operator to run `go build`/`go vet`.

This run is the companion to the
[implementation control on the same model and day](2026-09-08-go-implement-control-sonnet-5.md),
which found that corpus saturated.

## Results

Line values are mean delta with population standard deviation over five runs.
Negative is less production code; the effect column is baseline minus no-skill,
so negative favors the skill. `p` is an exact two-sided permutation test over
the ten runs of that fixture.

| Fixture | No skill | Skill | Effect | 95% CI | p |
|---|---:|---:|---:|---|---:|
| `dispatch` | 0.0 ± 7.7 | −4.4 ± 5.2 | −4.4 | −15.1 … +6.3 | 0.38 |
| `pricing` | −38.8 ± 4.0 | −41.8 ± 8.9 | −3.0 | −14.2 … +8.2 | 0.53 |
| `report` | +17.6 ± 4.5 | +20.4 ± 6.2 | +2.8 | −6.0 … +11.6 | 0.52 |
| `store` | −0.6 ± 4.5 | −8.8 ± 4.1 | **−8.2** | −15.2 … −1.2 | **0.04** |
| All runs | −5.45 | −8.65 | **−3.20** | | |

| Structural additions across 20 runs | No skill | Skill | Change |
|---|---:|---:|---:|
| Types | 6 | 10 | +4 |
| Interfaces | 0 | 0 | no difference |
| Functions | 34 | 20 | −41.2% |
| Pattern-name hits | 0 | 0 | no difference |
| Branches | −120 | −136 | −16 |

The skill produced a test file in 9 of 20 runs against none in the control.
`abrun` hides model-authored tests before applying the golden test, so this is
evidence of test creation only, not that those tests compile or assert the right
thing.

Cost is 3.3x: $0.0636 per control run against $0.2085 per baseline run.

## Which fixture separates, and why

`store` has the largest observed reduction. The following transformation
descriptions come from final session messages, not retained generated source.
The control extracted a helper in all five runs — `lookup` in four,
`lookupCache`/`lookupRemote`/`cacheSet` in one — and finished at −0.6 lines,
because the nesting moved into a function instead of leaving. All five skilled
runs collapsed the three duplicate miss branches into one guard clause, and
three of them additionally deleted the `if m.cache != nil` / `if m.remote != nil`
guards around map *reads* as dead checks, keeping the guard on the write; only
one added a helper. Per-fixture function growth is 7 against 1, and the arms
barely overlap on lines: −7 … +6 against −15 … −3.

`report`, the fixture that carried the Opus 5 result, does not reproduce, and
the control starts with less growth than in the Opus run. It grows `report` by 17.6 lines here against Opus 5's
+33.4, declares no type in any of its five runs and no interface, so the
over-engineering bait is mostly untaken and there is little for the skill to
suppress. The skilled arm ends 2.8 lines higher, and two of its runs fold the
csv/text choice into a single `format`/`lineFormat` value — the
«Remove Duplication to the End» shape, each selection over the same key made
once — which costs a type and a few lines on a package this small.

`dispatch` moves toward the skill for the first time on a Claude model (−4.4
against Opus 5's +2.0) with the same folding visible: two skilled runs replaced
the three near-identical `if/else if` branches with a lookup table keyed by kind
(`kindConfig` in one, `kinds` in the other), each declaring one type. `pricing`
is −3.0 with both arms already removing about 40 lines, and its interval spans
zero.

## Interpretation

The observed corpus mean favors the plugin by 3.20 production lines per run;
new functions total 34 versus 20, while new types total 6 versus 10. The direction
matches the Opus 5 control, but separate model runs do not establish the same
mechanism or isolate the effect of a particular skill edit. `report` does not
reproduce the Opus benefit: its observed difference is +2.8 lines here.

The `store` difference is -8.2 lines, with unadjusted exact permutation
p = 10/252 = 0.03968. Four fixtures were examined without a preregistered
`store` hypothesis; a four-test Bonferroni adjustment gives p = 0.15873.
The reported Welch interval is also unadjusted. Treat this as an exploratory
signal worth a focused comparison, not a confirmed general improvement.

Session messages suggest that some extra types represent lookup tables or
format values. Aggregate counts cannot identify all four net extra types or
prove that the duplication rule caused them. Likewise, the 3.28x recorded cost
cannot be attributed mostly to skill reading: token usage, thinking, cache
usage, and complete traces are absent from the supplied JSON.

Together with the implementation control, both corpora have 40/40 golden passes
per arm (80 sessions total). Refactor has a favorable size direction, especially
on `store`; implement does not, and `gateway` grows by 8.2 lines with the plugin.
Neither corpus demonstrates a correctness advantage on this model. The same
baseline digest is recorded in both, but no reference arm isolates the recent
router, helper-rule, or duplication-rule edits.
