# Go Implementation Corpus Control — GPT-5.6-Luna (medium) on Codex CLI

## Run

- Finished: 2026-09-07 16:52 EEST
- Runner: `codex` 0.153.4
- Model: `gpt-5.6-luna`, reasoning effort `medium` (the model's own default)
- Seed: `1`
- Corpus: `implement`
- Fixtures: `catalog`, `feed`, `gateway`, `ledger`
- Arms: `no-skill`, `baseline`
- Repetitions: 5 per fixture and arm, 40 sessions total
- Baseline plugin SHA-256: `6443a658f74cd465b98f0065cdb7c117bcde52c64af1acc1a5ef7f3a67369d3c`
- Plugin source: working tree at `5dcb55d`
- Raw report: [`2026-09-07-go-implement-control-gpt-5.6-luna-medium.json`](2026-09-07-go-implement-control-gpt-5.6-luna-medium.json)
- Raw report SHA-256: `ff8df7a67ff44722e19ffb669497dc30e72fe19b47d9bee57b41d62fac48daec`

```bash
go run ./cmd/abrun -corpus implement -runner codex \
  -model gpt-5.6-luna -effort medium \
  -arms no-skill,baseline -n 5 -j 4 -seed 1 \
  -out ../docs/evidence/2026-09-07-go-implement-control-gpt-5.6-luna-medium.json
```

All 40 sessions completed, built, changed the fixture and stayed out of the
repository checkout. Skill routing is the best this corpus has recorded:
`go-code` fired in all 20 baseline sessions, `go-style-core` in 11,
`go-data-structures` in 10, `go-error-handling` in 9, `go-linting` in 8,
`go-defensive` in 7, `go-http` in 2 and `go-interfaces` in 1.

## Results

### Correctness

| Fixture | Golden, no skill | Golden, skill |
|---|---:|---:|
| `catalog` | 5/5 | 5/5 |
| `feed` | 5/5 | 5/5 |
| `gateway` | 5/5 | 5/5 |
| `ledger` | 5/5 | 5/5 |
| All runs | **20/20** | **20/20** |

Every trap is saturated. Unaided, this model reaches for a non-nil slice, sets
all four server timeouts, keeps the error chain and copies the caller's slice —
without being told, in every one of 20 sessions. That puts it with Opus 5 and
the pinned-fixture claude run: no model measured in this corpus has an unaided
arm that fails a trap.

### Code size

| Fixture | No skill | Skill | Effect | 95% interval |
|---|---:|---:|---:|---|
| `catalog` | +27.0 ± 5.0 | +23.0 ± 1.7 | −4.0 | −9.5 to +1.5 |
| `feed` | +47.2 ± 3.6 | +45.4 ± 1.1 | −1.8 | −5.7 to +2.1 |
| `gateway` | +93.8 ± 8.1 | +93.6 ± 9.9 | −0.2 | −13.4 to +13.0 |
| `ledger` | +42.0 ± 4.8 | +38.4 ± 5.2 | −3.6 | −11.0 to +3.8 |
| All runs | +52.50 | +50.10 | −2.40 | |

Every fixture points the same way and not one of them separates the arms; every
interval includes zero. The skilled arm is consistently tighter — ±1.7 against
±5.0 on `catalog`, ±1.1 against ±3.6 on `feed` — which is the same
spread-narrowing the [Opus 5 `gateway` run](2026-09-07-go-implement-gateway-opus5.md)
reported, without the size effect that accompanied it there.

| Structural additions across 20 runs | No skill | Skill |
|---|---:|---:|
| Types | 2 | 2 |
| Interfaces | 0 | 0 |
| Functions | 6 | 13 |
| Unrequested exported declarations | 0 | 0 |
| Pattern-name hits | 0 | 0 |

Neither arm wrote a test file. Codex reports token counts rather than dollars,
so the `$/run` column is empty for both arms.

## Interpretation

This run measures nothing about the skill, and it is worth publishing for
exactly that reason: it is the cleanest demonstration of the corpus's own
admission rule. A fixture earns its place only when `no-skill` is measurably
worse than `baseline`, and on this model none of the four is. The traps are live
— each golden test fails against the stub and passes against an idiomatic
reference — but a model that never falls into them leaves the skill nothing to
prevent, and the line metric then has only the spread to show.

The result is a useful contrast in two directions.

Against the same model on the [refactor corpus](2026-09-07-go-refactor-control-gpt-5.6-luna-medium.md),
where the skill removed all of the growth on `report` and cut helper functions
71%: the same model, the same runner, the same skills, opposite outcomes.
Removing structure from code that already works is where this skill pays; adding
the right code to an empty body is not, because this model already writes it.

Against the routing story from the other implementation runs, where the arms
failed to separate on `feed` because `go-data-structures` never fired in a
single `feed` session. Here it fired in 10 of 20 sessions, `go-defensive` in 7
and `go-error-handling` in 9 — the routing those runs were missing — and the arms still
did not separate, because there was no defect left to prevent. Good routing is
necessary for the corpus to show an effect and it is not sufficient; the model
has to be one that gets it wrong first.
