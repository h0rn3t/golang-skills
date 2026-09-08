# Go Implementation Corpus Control — Sonnet 5 on the Claude CLI

Report supplied by the operator; counts, means, standard deviations, costs, and
the gateway permutation p-value were checked against the raw JSON. Welch
intervals are retained from the supplied report. CLI version, clean checkout,
and launch command are operator-reported.

## Run

- Finished: 2026-09-08 14:24 EEST
- Runner: `claude` 2.1.260
- Model: `claude-sonnet-5`
- Seed: `1`
- Corpus: `implement`
- Fixtures: `catalog`, `feed`, `gateway`, `ledger`
- Arms: `no-skill`, `baseline`
- Repetitions: 5 per fixture and arm, 40 sessions total
- Baseline plugin SHA-256: `2b6acb094bcd1cfbb2fba5fabbfcb57c3069b0ad4b3a3731bc596da54e561799`
- Plugin source: working tree at `8d7f571`, clean
- Raw report: [`2026-09-08-go-implement-control-sonnet-5.json`](2026-09-08-go-implement-control-sonnet-5.json)
- Raw report SHA-256: `6555a7da0b510bd64e1b3fa6b2ddc138c169f1a0eb168aab2462be2001a51507`
- Cost: $1.20 control, $2.78 baseline, $3.98 total

```bash
go run ./cmd/abrun -corpus implement -model claude-sonnet-5 \
  -arms no-skill,baseline -n 5 -j 4 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-08-go-implement-control-sonnet-5.json
```

All 40 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. Sessions ran under `--restricted` with the
plugin's PostToolUse gofmt/vet hook active in the baseline arm and the tool set
limited to `Skill,Read,Glob,Grep,Edit,Write`, so neither arm had a shell; the
control arm also loses the `Skill` tool.

A `go-*` skill fired in 20 of 20 baseline sessions: `go-code` in all 20,
`go-style-core` in 6, `go-http` in 5, `go-error-handling` in 5, `go-linting` in
1. Per fixture the owner was reached in `gateway` 5/5 (`go-http`) and `catalog`
4/5 (`go-error-handling`); `feed`'s owner `go-data-structures` and `ledger`'s
owner `go-defensive` were reached in 0/5 each, with `go-code` handling those
tasks alone.

## Results

Correctness is saturated in both arms. Every session built and passed the hidden
golden test.

| Fixture | Golden, no skill | Golden, skill |
|---|---:|---:|
| `catalog` | 5/5 | 5/5 |
| `feed` | 5/5 | 5/5 |
| `gateway` | 5/5 | 5/5 |
| `ledger` | 5/5 | 5/5 |
| All runs | 20/20 | 20/20 |

The unaided model takes none of the four traps. On `gateway` — the timeout fixture — all five control sessions set
`ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout` and `IdleTimeout` without a
skill loaded, which is what the golden test checks, and three of the five also
report capping `MaxHeaderBytes`, which it does not.

| Fixture | Lines, no skill | Lines, skill | Effect | 95% CI |
|---|---:|---:|---:|---|
| `catalog` | 27.0 ± 7.8 | 27.2 ± 8.4 | +0.2 | −13.0 … +13.4 |
| `feed` | 47.4 ± 1.9 | 45.4 ± 2.3 | −2.0 | −5.4 … +1.4 |
| `ledger` | 43.8 ± 1.2 | 41.4 ± 5.2 | −2.4 | −8.5 … +3.7 |
| `gateway` | 65.4 ± 1.5 | 73.6 ± 7.4 | **+8.2** | −0.6 … +17.0 |
| All runs | 45.9 | 46.9 | +1.0 | −11.1 … +13.1 |

Line values are mean delta with population standard deviation over five passing
runs per cell. The corpus-wide row averages four fixtures of different sizes and
is reported only because both arms have the same valid set here; the per-fixture
rows are the comparable numbers.

`gateway` has the largest observed increase. Its Welch interval includes zero
at n=5. An exact permutation test using the absolute difference in means gives
6/252 = 0.02381, two-sided, before adjustment for multiple comparisons. If
selected after inspecting all four fixtures, a four-test Bonferroni adjustment
gives 0.09524; treat this as an exploratory signal. The control wrote 64, 64, 65,
66, 68 lines; the skilled arm 66, 69, 70, 76, 87. Spread moved the same way:
1.5 against 7.4.

Where the extra lines went can only be attributed from the session outputs:
every run passed, and `-keep` retains a scratch tree only for a failure, so the
code itself was not preserved. All five baseline sessions reached `go-http` and
all five extracted at least one helper (`Δfuncs` +1 in four, +4 in one) against
`Δfuncs` 0 in every control session, and four of the five describe a defensive
copy of the caller's account slice in those words. Every control session sorts a
copy inline. Decision points barely move — the control is +7 branches in all
five runs, the skilled arm +7 in three and +8 in two — and both arms harden the
server. Structural counts across the corpus move the same small distance:
`Δtypes` 0.20 → 0.50, `Δfuncs` 0.20 → 0.60, `Δexported` 0.00 → 0.10,
`Δbranch` 5.35 → 5.50.

No session in either arm declared an interface, hit a pattern name, or wrote a
test file, so the corpus's over-engineering baits went untaken here as in every
previous run of it.

Cost is 2.3x: $0.0599 per control run against $0.1391 per baseline run.

## Interpretation

Correctness is saturated on these fixtures: the skill has no observed defect-rate
advantage on Sonnet 5 in this run. Code size and cost still provide useful
comparisons: mean growth is one line higher across the corpus, with more helper
functions and types, and recorded cost is 2.32 times higher.

On `gateway`, every baseline session loads `go-http`, adds functions, and passes
the golden tests. The control passes without adding functions. This supports
investigating helper overhead, but does not isolate a cause: the arms differ by
the complete plugin, including its hook, not one extraction or copying rule.
Generated source was not retained, so output descriptions cannot establish
which transformations caused the difference or whether extra code buys safety
outside the golden contract. The p-value is not an effect size; the observed
mean difference is +8.2 lines, with substantial uncertainty at n=5.

The routing gap is separate: `feed` never loads `go-data-structures` and `ledger`
never loads `go-defensive`, while both pass in all sessions. This run cannot
establish what that missing routing would cost on another model. It also does
not isolate the recent router compression or helper-rule changes; such a claim
requires reference and candidate arms with otherwise identical content.
