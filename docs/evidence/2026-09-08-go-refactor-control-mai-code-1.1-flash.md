# Go Refactor Skill Control — MAI-Code-1.1-Flash on GitHub Copilot CLI, after the 2026-09-08 extraction rule

## Run

- Finished: 2026-09-08 14:14 EEST
- Runner: `copilot` 1.0.83
- Model: `mai-code-1.1-flash`
- Seed: `1`
- Corpus: `refactor`
- Fixtures: `dispatch`, `pricing`, `report`, `store`
- Arms: `no-skill`, `baseline`
- Repetitions: 5 per fixture and arm, 40 sessions total
- Baseline plugin SHA-256: `2b6acb094bcd1cfbb2fba5fabbfcb57c3069b0ad4b3a3731bc596da54e561799`
- Plugin source: working tree at `dd4d196`, with the `go-code-refactor` extraction
  rule and the `go-code` rewrite later committed as `8d7f571`
- Compared against: [`2026-09-07-go-refactor-control-mai-code-1.1-flash.md`](2026-09-07-go-refactor-control-mai-code-1.1-flash.md),
  the same runner, model, seed and fixtures on plugin `5dcb55d`
- Raw report: [`2026-09-08-go-refactor-control-mai-code-1.1-flash.json`](2026-09-08-go-refactor-control-mai-code-1.1-flash.json)
- Raw report SHA-256: `061d25f81247e742303d75891b6439ff8133ed53c9cfd76a3e4c44ced375eb27`

```bash
go run ./cmd/abrun -corpus refactor -runner copilot -model mai-code-1.1-flash \
  -arms no-skill,baseline -n 5 -j 4 -seed 1 \
  -out ../docs/evidence/2026-09-08-go-refactor-control-mai-code-1.1-flash.json
```

All 40 sessions completed without a CLI error, built, changed the fixture,
stayed out of the repository checkout, and passed their hidden golden test, so
all 40 count as evidence. A `go-*` skill fired in 15 of 20 baseline sessions:
`go-code-refactor` in 15, `go-code` in 7, `go-linting` in 2, `go-code-review`
in 1.

The runner conditions are the ones described in the
[2026-09-07 run on the same model](2026-09-07-go-refactor-control-mai-code-1.1-flash.md):
per-arm `COPILOT_HOME`, sessions restricted to
`skill,view,create,edit,grep,glob` and therefore no shell, and skill trees
without the plugin's hook or subagent.

### One invalid predecessor, discarded

The first attempt at this pair of runs measured nothing, and the reason belongs
in the record. The `go-code` description rewritten on 2026-09-08 contained an
unquoted `: ` inside the YAML scalar, which makes the frontmatter unparseable.
The copilot CLI drops such a skill from its listing entirely and prints the
failure *after* the JSON array, and `abrun`'s arm pre-check only asserts that a
baseline arm loads at least one `go-*` skill — so 80 sessions ran against a
baseline arm with no router in it at all. The symptom was `go-code` firing 0 of
20 where the previous run had it in 18 of 20. The description was rewritten with
an em-dash, `copilot skill list --json` was checked to load 24 of 24 `go-*`
skills with no failure block, and both corpora were re-run from scratch. Only
the re-runs are published here.

## Results

Line values are mean delta with sample standard deviation over five runs.
Negative is less production code; the effect column is baseline minus no-skill,
so negative favors the skill.

| Fixture | No skill | Skill | Effect | 95% interval |
|---|---:|---:|---:|---|
| `dispatch` | +5.6 ± 7.4 | +5.2 ± 5.8 | −0.4 | −10.1 … +9.3 |
| `pricing` | −21.0 ± 11.5 | −28.0 ± 16.1 | −7.0 | −27.5 … +13.5 |
| `report` | +15.8 ± 3.3 | +12.6 ± 0.5 | −3.2 | −6.6 … +0.2 |
| `store` | +6.0 ± 9.4 | +9.4 ± 8.1 | +3.4 | −9.4 … +16.2 |
| All runs | +1.60 | −0.20 | **−1.80** | |

Every interval still includes zero. `report`, the trap fixture, is the one that
comes close: −3.2 lines with an interval of −6.6 to +0.2, and a skilled spread
of ±0.5 lines against the control's ±3.3 — four of its five skilled sessions
returned +13 lines and one +12.

| Structural additions across 20 runs | No skill | Skill | Change |
|---|---:|---:|---:|
| Types | 4 | 4 | no difference |
| Interfaces | 0 | 0 | no difference |
| Functions | 43 | 48 | +11.6% |
| Pattern-name hits | 0 | 0 | no difference |

The skill produced a test file in 1 of 20 runs against 0 of 20 without it, down
from 5 of 20 on 2026-09-07.

Copilot bills a session in premium requests and AI credits rather than dollars,
so the `$/run` column is empty for both arms and this file carries no cost
comparison.

## What changed against 2026-09-07

The comparison is across runs rather than a same-run `reference` arm, so it
carries the run-to-run noise of both, and the control arm is the measure of that
noise: identical code, identical seed, identical model name, and a corpus mean
that moved from −3.05 to +1.60 lines. That drift, 4.7 lines, is larger than the
corpus-wide effect being claimed. Read the per-fixture rows, not the corpus row.

| Metric | 2026-09-07 | 2026-09-08 |
|---|---:|---:|
| Corpus Δlines, skill vs control | +1.10 | −1.80 |
| `report` Δlines, skill vs control | −0.6 | −3.2 |
| `report` growth removed | 4% | 20% |
| Functions, control → skill | 37 → 47 (**+27.0%**) | 43 → 48 (+11.6%) |
| Types, control → skill | 1 → 0 | 4 → 4 |
| Test files written, skill arm | 5/20 | 1/20 |
| Build and golden, both arms | 20/20 | 20/20 |

The one result the 2026-09-07 file called out as running against the plugin —
the skilled arm adding 47 functions to the control's 37 — is the one that moved.
Helper growth is still positive at +11.6%, but the gap fell by more than half,
and it did so on the fixture the extraction rule addresses: on `report` the two
arms now add the same three functions per session and the skilled arm reaches
them in 12.6 lines instead of 15.8.

## Interpretation

The extraction rule moved the metric it was written for, in the right
direction, and nothing here is strong enough to call a result.

What supports the wording: the skilled arm no longer out-grows the control in
helper count on any fixture, `report` is tighter in both mean and spread, and
the corpus direction flipped from marginally against the skill to marginally for
it. What withholds it: no interval excludes zero, the control drifted 4.7 lines
between two runs with the same seed, and `store` moved 3.4 lines the wrong way
with an interval twice that wide.

The reason this corpus cannot do better on this model is the same one the
2026-09-07 file gave, and it has not changed: the traps are saturated. The
control still grows `report` by only 15.8 lines against Opus 5's +33.4 and
MiniMax M3's +27.4, still declares no interfaces in 20 unaided sessions, and
already deletes 21 lines from `pricing` on its own. A fixture earns its place
only when `no-skill` is measurably worse than `baseline`, and none of these four
clears that bar on this model. The honest use of this file is as the second
sample of a before/after pair on a model that cannot separate the arms, which
makes it evidence that the change did no harm and a weak signal that it helped
the specific behavior it names.

To make a claim about the extraction rule, run it as a `reference` arm against
`5dcb55d` in one invocation at `n=10` on a model whose control still takes the
bait — GPT-5.6-Luna on codex or Opus 5 on claude, both of which grow `report`
enough for the rule to have something to prevent.
