# Go Implementation Corpus Control — MAI-Code-1.1-Flash on GitHub Copilot CLI, after the 2026-09-08 `go-code` rewrite

## Run

- Finished: 2026-09-08 14:35 EEST
- Runner: `copilot` 1.0.83
- Model: `mai-code-1.1-flash`
- Seed: `1`
- Corpus: `implement`
- Fixtures: `catalog`, `feed`, `gateway`, `ledger`
- Arms: `no-skill`, `baseline`
- Repetitions: 5 per fixture and arm, 40 sessions total
- Baseline plugin SHA-256: `2b6acb094bcd1cfbb2fba5fabbfcb57c3069b0ad4b3a3731bc596da54e561799`
- Plugin source: working tree at `dd4d196`, with the `go-code` rewrite and the
  `go-code-refactor` extraction rule later committed as `8d7f571`
- Compared against: [`2026-09-07-go-implement-control-mai-code-1.1-flash.md`](2026-09-07-go-implement-control-mai-code-1.1-flash.md),
  the same runner, model, seed and fixtures on plugin `5dcb55d`
- Raw report: [`2026-09-08-go-implement-control-mai-code-1.1-flash.json`](2026-09-08-go-implement-control-mai-code-1.1-flash.json)
- Raw report SHA-256: `fbb9335d63fda22a0a433499394eee4e6c15b5fc2de3aebb720fb2cdd1ab0302`

```bash
go run ./cmd/abrun -corpus implement -runner copilot -model mai-code-1.1-flash \
  -arms no-skill,baseline -n 5 -j 4 -seed 1 \
  -out ../docs/evidence/2026-09-08-go-implement-control-mai-code-1.1-flash.json
```

All 40 sessions completed without a CLI error, built, changed the fixture, and
stayed out of the repository checkout. A `go-*` skill fired in all 20 baseline
sessions: `go-code` in **20**, `go-linting` in 10, `go-code-review` in 3,
`go-testing` in 3, `go-style-core` in 2, `go-code-refactor` in 1. `go-http`
fired in none.

The runner conditions are the ones described in the
[refactor control on the same model](2026-09-08-go-refactor-control-mai-code-1.1-flash.md):
per-arm `COPILOT_HOME`, sessions restricted to
`skill,view,create,edit,grep,glob` and therefore no shell, and skill trees
without the plugin's hook or subagent. That file also records the invalid first
attempt at this run — a `go-code` description whose broken YAML frontmatter kept
the skill from loading at all — and only the re-run is published here.

## Results

The headline of the 2026-09-07 run is gone, and it is the **control** that
changed, not the skill.

| Fixture | Golden, no skill | Golden, skill | 2026-09-07 |
|---|---:|---:|---|
| `catalog` | 5/5 | 5/5 | 5/5 → 5/5 |
| `feed` | 5/5 | 5/5 | 4/5 → 5/5 |
| `gateway` | **5/5** | **4/5** | **1/5 → 3/5** |
| `ledger` | 5/5 | 5/5 | 5/5 → 5/5 |
| All runs | 20/20 | 19/20 | 15/20 → 18/20 |

A day earlier this model failed the `gateway` golden test in four of five
unaided sessions, which made it the only known configuration where the
implementation corpus measured anything. Here the unaided arm passes all five.
The trap is dead on this model, and with it the 1/5 → 3/5 result.

The single failure is in the skilled arm: `gateway` rep 0, which loaded
`go-code` and `go-linting`, and left `ReadTimeout` at zero. Fisher's exact on
5/5 against 4/5 gives p = 1.0. One session is not a defect-rate claim in either
direction.

Line values are mean delta with sample standard deviation over the runs that
passed; the effect column is baseline minus no-skill.

| Fixture | Lines, no skill | Lines, skill | Effect | 95% interval |
|---|---:|---:|---:|---|
| `catalog` | +27.0 ± 4.3 (n=5) | +25.2 ± 1.8 (n=5) | −1.8 | −6.6 … +3.0 |
| `feed` | +55.0 ± 5.8 (n=5) | +53.0 ± 4.7 (n=5) | −2.0 | −9.7 … +5.7 |
| `gateway` | +99.8 ± 19.7 (n=5) | +99.5 ± 9.0 (n=4) | −0.3 | −23.2 … +22.6 |
| `ledger` | +49.4 ± 5.5 (n=5) | +53.6 ± 4.3 (n=5) | +4.2 | −3.0 … +11.4 |
| All runs | +57.80 (n=20) | +55.63 (n=19) | −2.17 | |

No fixture separates the arms and every interval includes zero. `gateway` is
worth one more look because it is the fixture Opus 5 separated completely: the
per-session line counts are 73/87/108/108/123 without the skill against
89/98/100/111 with it — the same mean, the spread cut by half. On Opus 5 the
same fixture went from 152.6 lines to 99.8 with the spread seven times tighter;
here both arms already write about 100 lines, which is what the skilled arm cost
on Opus 5.

| Structural additions across 20 runs | No skill | Skill | Change |
|---|---:|---:|---:|
| Types | 4 | 10 | +6 |
| Interfaces | 0 | 0 | no difference |
| Functions | 12 | 19 | +7 |
| Pattern-name hits | 0 | 0 | no difference |
| Unrequested exported symbols | 0 | 0 | no difference |

Neither arm wrote a test file. The corpus's over-engineering baits went untaken
in both arms, as in every previous run of it. The extra types are `feed`, where
the skilled arm declares 1.8 per session against the control's 0.8 — a JSON
wrapper struct, which is one of the fixture's legitimate solutions rather than a
bait.

Copilot bills in premium requests and AI credits rather than dollars, so the
`$/run` column is empty for both arms.

## Routing

`go-code` now fires in 20 of 20 baseline sessions, against 18 of 20 before the
rewrite, and it is the only skill that fired in every session. The routing
result underneath it did not improve.

| Skill | 2026-09-07 | 2026-09-08 |
|---|---:|---:|
| `go-code` | 18/20 | **20/20** |
| `go-linting` | 14/20 | 10/20 |
| `go-http` | 2/20 | **0/20** |
| `go-code-review` | 4/20 | 3/20 |
| `go-testing` | 1/20 | 3/20 |
| `go-style-core` | 2/20 | 2/20 |

`gateway` is owned by `go-http`, and its entire hidden requirement is an HTTP
server's timeout configuration. The router reached that owner in 2 of 5
`gateway` sessions before the rewrite and in 0 of 5 after it; both of the
earlier sessions that reached it set every timeout, and the one session that
failed here is again one that did not. At 5 sessions per cell 2/5 against 0/5 is
p ≈ 0.44, so this is not a demonstrated regression — but the rewrite made the
router fire more often while, on the one fixture with a topic owner that
matters, routing to that owner did not happen at all. That is the thing to
measure next, and it is measurable cheaply: the firing rate of `go-http` on
`gateway` needs no golden test to read.

## Interpretation

Three separate statements, and only the first is solid.

**The 2026-09-07 implementation result does not reproduce, and the corpus is
back to having no live trap.** The control arm went from 15/20 to 20/20 on
identical fixtures, prompt, seed, runner and model name. Nothing in the
repository changed the control arm — it loads no skills — so the change is
outside this repository: `mai-code-1.1-flash` as served through the copilot CLI
does not write the same code it wrote a day earlier. Any published claim resting
on that control needs the date attached, and the `gateway` 20% → 60% figure in
the README is now a historical measurement of a model version, not a current
property of the plugin.

**The skill's own numbers are unchanged and tied.** 19/20 against 20/20 on
correctness, −2.17 lines corpus-wide with every interval crossing zero, no
interface and no pattern name in either arm. Where the model already avoids a
defect the skill has nothing to add, which is the same reading the GPT-5.6-Luna
implementation control produced at 20/20 in both arms.

**The `go-code` rewrite fires better and routes no better.** 20/20 firing is the
best this corpus has recorded for the router, and it did not convert into
reaching `go-http` on the fixture that needs it. The rewrite removed the
"without one clear topic" hedge from the description, which plausibly explains
the firing rate; the routing table it hands the model is what decides the second
step, and this run gives no evidence that step improved.

What the corpus needs is a model whose unaided arm still falls into a trap. As
of this run there is none: Opus 5, GPT-5.6-Luna and now MAI-Code-1.1-Flash all
pass every fixture unaided. Until one is found, this corpus measures the size
and shape of a working implementation, not defect prevention.
