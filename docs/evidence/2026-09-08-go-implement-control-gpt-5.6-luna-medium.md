# Go Implementation Corpus Control — GPT-5.6-Luna (medium) on Codex CLI, re-run

## Run

- Finished: 2026-09-08 15:19 EEST (started 15:05, 13.4 minutes wall clock)
- Runner: `codex` 0.153.4
- Model: `gpt-5.6-luna`, reasoning effort `medium`
- Seed: `1`
- Corpus: `implement`
- Fixtures: `catalog`, `feed`, `gateway`, `ledger`
- Arms: `no-skill`, `baseline`
- Repetitions: 5 per fixture and arm, 40 sessions total
- Baseline plugin SHA-256: `2b6acb094bcd1cfbb2fba5fabbfcb57c3069b0ad4b3a3731bc596da54e561799`
- Plugin source: working tree at `8d7f571`, uncommitted documentation edits
  included. `skills/go-code-refactor` was edited at 15:14, after the arms were
  materialized and before the run ended; arms are copied once before the first
  session, so those edits are not in the measured arm and the digest above is
  what ran.
- Raw report: [`2026-09-08-go-implement-control-gpt-5.6-luna-medium.json`](2026-09-08-go-implement-control-gpt-5.6-luna-medium.json)
- Raw report SHA-256: `dcb5a8e70176f862854326f540b26186c926a74d3e7a4203dc9f04bef466d159`

```bash
go run ./cmd/abrun -corpus implement -runner codex \
  -model gpt-5.6-luna -effort medium \
  -arms no-skill,baseline -n 5 -j 4 -seed 1 \
  -out ../docs/evidence/2026-09-08-go-implement-control-gpt-5.6-luna-medium.json
```

This repeats the [2026-09-07 run](2026-09-07-go-implement-control-gpt-5.6-luna-medium.md)
with the same prompt, seed, fixtures, runner and model against the rewritten
`go-code` router and the extraction rule moved into `go-code-refactor`. All 40
sessions completed, built, changed the fixture and stayed out of the repository
checkout.

## Results

### Correctness

| Fixture | Golden, no skill | Golden, skill |
|---|---:|---:|
| `catalog` | 5/5 | 5/5 |
| `feed` | 5/5 | 5/5 |
| `gateway` | 5/5 | 5/5 |
| `ledger` | 5/5 | 5/5 |
| All runs | **20/20** | **20/20** |

Identical to the previous run, and the reason is the same: every trap is
saturated. Unaided, this model reaches for a non-nil slice, sets all four server
timeouts, keeps the error chain and copies the caller's slice — in every one of
20 sessions, without being told. There is no defect left for a skill to prevent,
so no correctness claim can come out of this corpus on this model.

### Code size

| Fixture | No skill | Skill | Effect | 95% interval | p (permutation) |
|---|---:|---:|---:|---|---:|
| `catalog` | +25.2 ± 5.5 | +22.6 ± 0.9 | −2.6 | −8.4 to +3.2 | 0.56 |
| `feed` | +47.4 ± 3.4 | +44.0 ± 3.1 | −3.4 | −8.1 to +1.3 | 0.18 |
| `gateway` | +90.8 ± 7.4 | +94.6 ± 16.1 | +3.8 | −14.4 to +22.0 | 0.71 |
| `ledger` | +41.4 ± 3.5 | +39.8 ± 3.1 | −1.6 | −6.4 to +3.2 | 0.53 |
| All runs | +51.20 | +50.25 | −0.95 | | 0.92 |

Three fixtures point the skill's way and `gateway` points against it; no
interval excludes zero and the corpus effect is under one line. The `gateway`
sign flip is one session: the skilled arm wrote 77, 88, 88, 101 and 119 lines,
and the 119-line session is also the only one in the whole run that declared a
type and six helper functions. Its own spread, ±16.1 against the control's
±7.4, is the number to read there rather than the mean.

The spread-narrowing the previous run reported survives on the other three
fixtures — ±0.9 against ±5.5 on `catalog`, ±3.1 against ±3.5 on `ledger`,
±3.1 against ±3.4 on `feed`.

| Structural additions across 20 runs | No skill | Skill |
|---|---:|---:|
| Types | 0 | 2 |
| Interfaces | 0 | 0 |
| Functions | 7 | 15 |
| Unrequested exported declarations | 0 | 0 |
| Pattern-name hits | 0 | 0 |
| Mean branches per run | 6.75 | 6.65 |

Neither arm wrote a test file, and no session in either arm declared an
interface — the `ledger` and `catalog` bait is still untaken on both sides.
Codex reports token counts rather than dollars, so `$/run` is empty for both
arms.

### Against the previous run

| Fixture | No skill, 09-07 → 09-08 | Skill, 09-07 → 09-08 |
|---|---:|---:|
| `catalog` | 27.0 → 25.2 | 23.0 → 22.6 |
| `feed` | 47.2 → 47.4 | 45.4 → 44.0 |
| `gateway` | 93.8 → 90.8 | 93.6 → 94.6 |
| `ledger` | 42.0 → 41.4 | 38.4 → 39.8 |
| All runs | 52.50 → 51.20 | 50.10 → 50.25 |

The control moved by −1.3 lines across the corpus and no fixture by more than
3. On this model the control arm is stable enough between days that a re-run
reproduces it rather than replacing it, which is not something to assume for
another model — a control is a property of a served model version and is worth
re-measuring rather than inheriting from an earlier file.

### Routing

| Skill | Baseline sessions, 09-07 | Baseline sessions, 09-08 |
|---|---:|---:|
| `go-code` | 20/20 | 20/20 |
| `go-style-core` | 11 | 17 |
| `go-linting` | 8 | 16 |
| `go-error-handling` | 9 | 15 |
| `go-data-structures` | 10 | 15 |
| `go-defensive` | 7 | 7 |
| `go-http` | 2 | 4 |
| `go-interfaces` | 1 | 2 |
| `go-security` | 0 | 2 |
| `go-functions` | 0 | 2 |
| `go-testing` | 0 | 1 |
| `go-code-review` | 0 | 1 |

This is the only thing the re-run moved, and it moved in the direction the
`go-code` rewrite intended: the router still fires in every baseline session and
now pulls more rule owners behind it. Read per fixture, each trap's owning skill
is reached more often than before — `go-error-handling` in 5 of 5 `catalog`
sessions, `go-data-structures` in 4 of 5 `feed`, `go-http` in 4 of 5 `gateway`
against 2 of 5 on 09-07, and `go-defensive` in 2 of 5 `ledger`.

## Interpretation

The re-run confirms the previous conclusion rather than changing it: on this
model the implementation corpus measures nothing about the plugin, because the
control arm is already correct on all four traps and the size difference is
under a line.

What it adds is a clean separation of two measurements that are easy to conflate.
Routing improved substantially — four owner skills reached in half again as many
sessions, and the `gateway` owner doubled — and the measured outcome did not
move at all. Firing rate is necessary for this corpus to show anything and it is
not sufficient; the model has to get the thing wrong first. Any future claim that
a description or router edit helped needs a fixture the served model actually
fails, not a better firing rate.

Two open items follow from that. The corpus has no recorded live trap on any
model currently in the set, so admitting a correctness claim here needs either a
weaker model or harder fixtures. And the `go-code` contract-intake section stays
unmeasured: this run cannot see it for the same saturation reason, so it needs a
model where `gateway` genuinely fails, at n ≥ 10 per arm.

The contrast with the [refactor corpus on the same model](2026-09-07-go-refactor-control-gpt-5.6-luna-medium.md)
is unchanged and is still the sharpest statement of where the plugin pays: same
model, same runner, same skills, and there the skill removed all of the growth on
`report` and cut helper functions 71%. Removing structure from code that already
works is the payoff; adding the right code to an empty body is not, because this
model already writes it.
