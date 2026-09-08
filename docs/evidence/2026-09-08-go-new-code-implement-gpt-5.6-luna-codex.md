# New-Code Edit A/B on the Implementation Corpus — GPT-5.6-Luna (medium) on Codex CLI

## Run

- Finished: 2026-09-08 16:01 EEST (started 15:40, 20.8 minutes wall clock)
- Runner: `codex` 0.153.4
- Model: `gpt-5.6-luna`, reasoning effort `medium`
- Seed: `1`
- Corpus: `implement`
- Fixtures: `catalog`, `feed`, `gateway`, `ledger`
- Arms: `no-skill`, `reference`, `baseline`
- Repetitions: 5 per fixture and arm, 60 sessions total
- `reference` plugin: `git archive HEAD` of `0aca2fe` into a scratch root —
  the skill tree *before* the new-code edit. SHA-256
  `2b6acb094bcd1cfbb2fba5fabbfcb57c3069b0ad4b3a3731bc596da54e561799`, which is
  byte-identical to the arm measured as `baseline` in the
  [2026-09-08 control](2026-09-08-go-implement-control-gpt-5.6-luna-medium.md).
- `baseline` plugin: working tree at `0aca2fe` plus the new-code edit —
  the `## Writing New Code` section in `skills/go-code/SKILL.md` and the
  accompanying `go-code-refactor` SKILL/PLAYBOOK/MODERNIZATION edits. SHA-256
  `4223809986c69102d4de921698352b6811cc3c1b01eddf978d2c9a77611e88bb`
- Raw report: [`2026-09-08-go-new-code-implement-gpt-5.6-luna-codex.json`](2026-09-08-go-new-code-implement-gpt-5.6-luna-codex.json)
- Raw report SHA-256: `10bbeb838b3f116fa71c1ad6e2c0995cc2688f19057d799d8bdd5d713bbbaad7`

```bash
go run ./cmd/abrun -corpus implement -runner codex \
  -model gpt-5.6-luna -effort medium \
  -reference-root <git archive HEAD checkout> \
  -arms no-skill,reference,baseline -n 5 -j 4 -seed 1 \
  -out ../docs/evidence/2026-09-08-go-new-code-implement-gpt-5.6-luna-codex.json
```

Unlike the two published control runs on this model, this is a three-arm run:
the `reference` arm isolates the new-code edit from the plugin-versus-nothing
question, because a two-arm `no-skill`/`baseline` run cannot say which of the
two the difference belongs to. All 60 sessions completed, built, changed the
fixture and stayed out of the repository checkout — 60/60 valid.

`±` below is the population standard deviation of the five sessions; each 95%
interval is the difference of means ± *t*(8)·SE from those same deviations, and
each *p* is the exact two-sided permutation test over all 252 splits of the ten
sessions (the corpus row uses 200 000 random relabelings of the 20 + 20).

## Results

### Correctness

| Fixture | Golden, no skill | Golden, reference | Golden, new-code edit |
|---|---:|---:|---:|
| `catalog` | 5/5 | 5/5 | 5/5 |
| `feed` | 5/5 | 5/5 | 5/5|
| `gateway` | 5/5 | 5/5 | 5/5 |
| `ledger` | 5/5 | 5/5 | 5/5 |
| All runs | **20/20** | **20/20** | **20/20** |

Saturated for the third time on this model, and this run adds a control arm's
worth of confirmation: unaided, it sets all four server timeouts, keeps the
error chain, copies the caller's slice and returns non-nil collections in every
one of 20 sessions. No correctness claim — for the edit or for the plugin — can
come out of this corpus on this model.

### Code size

| Fixture | No skill | Reference | New-code edit | Edit effect | 95% interval | p |
|---|---:|---:|---:|---:|---|---:|
| `catalog` | +22.4 ± 0.5 | +23.0 ± 2.0 | +23.4 ± 1.7 | +0.4 | −2.3 to +3.1 | 1.00 |
| `feed` | +47.0 ± 0.9 | +45.6 ± 1.6 | +47.4 ± 2.8 | +1.8 | −1.5 to +5.1 | 0.37 |
| `gateway` | +93.8 ± 5.7 | +93.2 ± 8.0 | +93.0 ± 11.3 | −0.2 | −14.5 to +14.1 | 1.00 |
| `ledger` | +43.0 ± 1.1 | +39.2 ± 3.5 | +40.0 ± 2.8 | +0.8 | −3.8 to +5.4 | 0.79 |
| All runs | +51.55 | +50.25 | +50.95 | +0.70 | | 0.94 |

The edit does not move code size. Three fixtures point slightly against it, one
slightly for it, every interval spans zero, and the corpus difference is under a
line — the same ceiling the two previous runs on this model hit, for the same
reason: the control arm already writes a correct, compact implementation.

Against no skill at all the whole plugin is again worth −0.60 lines across the
corpus (p = 0.95), with only `ledger` approaching a signal (−3.0, −6.1 to +0.1,
p = 0.095).

### Helper functions — the one thing that moved

> **Superseded by re-measurement.** The
> [`gateway` follow-up run](2026-09-08-go-new-code-gateway-gpt-5.6-luna-codex.md)
> re-ran this run's `baseline` arm from the byte-identical digest and got 4.40
> helpers per session where this run got 1.80. Pooled at n = 10 that tree writes
> 3.10 ± 2.26 helpers, so the 3.20 → 1.80 reduction below **does not hold**, and
> the p = 0.040 control comparison re-estimates to +1.80, p = 0.066 against the
> pooled control. The numbers in this section are what this run measured; read
> the follow-up for what they mean.

| Over 20 runs | No skill | Reference | New-code edit |
|---|---:|---:|---:|
| Types | 1 | 2 | 3 |
| Interfaces | 0 | 0 | 0 |
| **Functions** | **4** | **16** | **9** |
| Unrequested exported declarations | 0 | 0 | 0 |
| Pattern-name hits | 0 | 0 | 0 |
| Mean branches per run | 6.75 | 6.40 | 6.95 |

Every one of those functions was written on `gateway`; the other three fixtures
added none in any arm. Per session there:

| `gateway` helpers | Sessions | Mean |
|---|---|---:|
| No skill | 0, 1, 3, 0, 0 | 0.80 |
| Reference | 5, 2, 3, 2, 4 | 3.20 |
| New-code edit | 0, 3, 0, 2, 4 | 1.80 |

Read as a pair, this is the only interval in the run that excludes zero, and it
is a finding *against the pre-edit skill tree*: `no-skill` → `reference` is
**+2.40** helpers, interval +0.7 to +4.1, p = 0.040. The skill tree as it stood
made this model write helper functions on `gateway` that it does not write
unaided, and it bought nothing in lines or branches for them.

The edit cuts that back toward the control — 3.20 → 1.80, −1.40, interval −3.4
to +0.6, p = 0.278 — which is the direction the section was written for: its
clause that a skill example's `validate`, `toDomain` or `writeError` methods are
"not a required list of helpers to create" targets exactly this. At n = 5 the
reduction is directional, not established; two of the five sessions still wrote
2 and 4 helpers, so nothing here says the phrasing removed the cause rather than
the current sample landing lower.

`Δiface` is zero in all 60 sessions and no arm wrote a test file, so the
`ledger`/`catalog` bait is untaken on all three sides. Codex reports token
counts rather than dollars, so `$/run` is empty for every arm.

### Reproducibility of the reference arm

Because the `reference` digest is byte-identical to the arm the 09-08 control
measured, this run also re-measures that arm a day-part later under the same
seed and prompt:

| Fixture | 09-08 control, skill arm | This run, `reference` |
|---|---:|---:|
| `catalog` | 22.6 | 23.0 |
| `feed` | 44.0 | 45.6 |
| `gateway` | 94.6 | 93.2 |
| `ledger` | 39.8 | 39.2 |
| All runs | **50.25** | **50.25** |
| Helper functions / 20 runs | 15 | 16 |

The corpus mean lands on the same number and no fixture moves more than 1.6
lines; helper count reproduces at 16 against 15. The `no-skill` arm is looser —
51.20 against 51.55 across the corpus, with `catalog` at 25.2 against 22.4 and
`gateway` at 90.8 against 93.8 — which is worth remembering before reading a
2–3 line control difference between any two runs as an effect.

### Routing

Sessions out of 20 that reached each skill:

| Skill | Reference | New-code edit |
|---|---:|---:|
| `go-code` | 20 | 20 |
| `go-style-core` | 19 | 14 |
| `go-linting` | 17 | 13 |
| `go-data-structures` | 15 | 12 |
| `go-error-handling` | 14 | 12 |
| `go-defensive` | 9 | 7 |
| `go-http` | 5 | 4 |
| `go-functions` | 3 | 0 |
| `go-code-review` | 2 | 3 |
| `go-testing` | 2 | 0 |
| `go-packages` | 1 | 1 |
| `go-security` | 1 | 0 |
| `go-interfaces` | 0 | 1 |
| Mean skills per session | 5.40 | 4.35 |

The router itself still fires in every session, and behind it the edit pulls
about one skill fewer per session. Per fixture the owning skill moves both ways:
`go-error-handling` 2/5 → 4/5 on `catalog`, `go-data-structures` 5/5 → 3/5 on
`feed`, `go-http` 5/5 → 4/5 on `gateway`, `go-defensive` 1/5 → 1/5 on `ledger`.
The [09-08 control](2026-09-08-go-implement-control-gpt-5.6-luna-medium.md)
already established on this model that firing rate and measured outcome are
separate results; this run is consistent with that in both directions — fewer
skills loaded, same lines, same 20/20.

## Interpretation

The edit is safe and unmeasured on size. Correctness stays at 20/20, the corpus
grows by 0.70 lines with p = 0.94, and no fixture separates — which on this
model is what any implementation-corpus result looks like, because the control
arm is already correct on all four traps.

The result worth keeping is the one the third arm made visible: on `gateway` the
pre-edit skill tree produced +2.40 helper functions per session against the
unaided model, interval excluding zero, with no line or branch benefit. That is
a plugin cost this corpus had never isolated, and it is the defect the new-code
section names. The section moves it 3.20 → 1.80 toward the control, directional
at n = 5.

Two follow-ups fall out of that, and neither is answered here. The helper-count
reduction needs n ≥ 10 on `gateway` alone to become a claim rather than a
direction — that fixture is where all 29 helper functions in this run were
written, so a `-tasks gateway` run at n = 10 costs 30 sessions and measures the
thing directly. And the `go-code` contract-intake section stays unmeasured for
the third run in a row, for the same saturation reason; it needs a model whose
control arm actually fails `gateway`.
