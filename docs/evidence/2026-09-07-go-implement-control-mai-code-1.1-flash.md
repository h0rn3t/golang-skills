# Go Implementation Corpus Control — MAI-Code-1.1-Flash on GitHub Copilot CLI

## Run

- Finished: 2026-09-07 14:11 EEST
- Runner: `copilot` 1.0.83
- Model: `mai-code-1.1-flash`
- Seed: `1`
- Corpus: `implement`
- Fixtures: `catalog`, `feed`, `gateway`, `ledger`
- Arms: `no-skill`, `baseline`
- Repetitions: 5 per fixture and arm, 40 sessions total
- Baseline plugin SHA-256: `6443a658f74cd465b98f0065cdb7c117bcde52c64af1acc1a5ef7f3a67369d3c`
- Plugin source: working tree at `5dcb55d`
- Raw report: [`2026-09-07-go-implement-control-mai-code-1.1-flash.json`](2026-09-07-go-implement-control-mai-code-1.1-flash.json)
- Raw report SHA-256: `3474871d04bd8aa777513cf96d9b1b4809b950f8540c4cd9e880e2579a390c87`

```bash
go run ./cmd/abrun -corpus implement -runner copilot -model mai-code-1.1-flash \
  -arms no-skill,baseline -n 5 -j 4 -seed 1 \
  -out ../docs/evidence/2026-09-07-go-implement-control-mai-code-1.1-flash.json
```

All 40 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. A `go-*` skill fired in 19 of 20 baseline
sessions: `go-code` in 18, `go-linting` in 14, `go-code-review` in 4, `go-http`
and `go-style-core` in 2 each, `go-testing` and `go-error-handling` in 1 each.

The runner conditions are the ones described in the
[refactor control on the same model](2026-09-07-go-refactor-control-mai-code-1.1-flash.md):
per-arm `COPILOT_HOME`, sessions restricted to `skill,view,create,edit,grep,glob`
and therefore no shell, and skills without the plugin's hook or subagent.

## Results

This is the first run of this corpus in which correctness moves. Unlike the
refactor corpus, the golden test here is the specification and can be failed
outright, and this model fails it.

| Fixture | Golden, no skill | Golden, skill |
|---|---:|---:|
| `catalog` | 5/5 | 5/5 |
| `feed` | 4/5 | 5/5 |
| `gateway` | **1/5** | **3/5** |
| `ledger` | 5/5 | 5/5 |
| All runs | 15/20 | 18/20 |

Every `gateway` failure in both arms is the fixture's trap and nothing else:
`NewServer().ReadTimeout = 0` or `ReadHeaderTimeout = 0`, an edge server built
as `&http.Server{Addr: addr, Handler: h}` with no bound on a stalled connection.
The single `feed` failure in the control arm is a compile error
(`undefined: eventJSON`), not the nil-slice trap. All 20 baseline sessions built.

The mechanism is visible per run. Both baseline `gateway` sessions that loaded
`go-http` — the skill that owns the timeout rule — passed; of the three that
loaded only `go-code` and `go-linting`, one passed. Skill routing, not skill
presence, is what separates the arms here.

Fisher's exact test on `gateway` gives a two-sided p of 0.52, and 0.41 across
the corpus. With five repetitions per cell this is a direction, not a result.

| Fixture | Lines, no skill | Lines, skill | Effect |
|---|---:|---:|---:|
| `catalog` | 28.6 ± 8.4 (n=5) | 24.8 ± 2.2 (n=5) | −3.8 |
| `feed` | 49.5 ± 4.4 (n=4) | 54.6 ± 0.5 (n=5) | +5.1 |
| `ledger` | 47.4 ± 5.0 (n=5) | 54.0 ± 3.9 (n=5) | +6.6 |
| `gateway` | 103.0 (n=1) | 104.7 ± 9.1 (n=3) | not comparable |

Line values are mean delta with sample standard deviation over the runs that
passed. The corpus-wide means the summary prints — 45.4 against 54.5 — must not
be read as an effect: the arms have different valid sets, and `gateway` is both
the longest fixture at roughly 100 lines and the one where the control lost four
of five runs, so it contributes one run to the control mean and three to the
skilled one. Over the three fixtures where both arms pass everything the
difference is +3.2 lines with the skill (41.3 against 44.5), against Opus 5's
−52.8 on `gateway`.

Neither arm declared an interface or an unrequested exported symbol in any run,
and no pattern-name hit was recorded, so the corpus's over-engineering baits
went untaken here as they did in every previous run of it. Neither arm wrote a
test file.

Copilot bills in premium requests and AI credits rather than dollars, so the
`$/run` column is empty for both arms.

## Interpretation

The corpus's blocking problem is gone on this model. The
[first discovery run](2026-09-07-go-implement-discovery-minimax-m3.md) recorded
that `no-skill` passed the hidden test in 12 of 12 runs on every fixture — the
traps were live but no model fell into them, so there was nothing for the skill
to prevent. MAI-Code-1.1-Flash falls into the `gateway` trap in four of five
unaided runs, which is the first time the fixture has measured what it was built
to measure.

What it measures is a routing effect. `gateway` is owned by `go-http`, and the
two baseline sessions that reached `go-http` both set every timeout. The
remaining baseline sessions routed to `go-code` and `go-linting` and did no
better than chance. That makes the 3/5 against 1/5 a lower bound on what correct
routing is worth on this fixture and, at the same time, a finding about routing:
the model reached the topic owner in 2 of 5 sessions on a task whose entire
hidden requirement is an HTTP server's configuration.

The line result does not reproduce. On Opus 5 the skill cut `gateway` from 152.6
to 99.8 lines; here the control that finishes at all already writes about 103,
and across the fixtures both arms pass, the skilled arm writes slightly more.
This model does not over-engineer these fixtures — it under-implements them —
and the corpus scores a different failure mode than the one the line metric was
designed for.

Two limits are worth stating. Five repetitions per cell cannot separate 1/5 from
3/5, so nothing here is a defect-rate claim; it is a reason to re-run `gateway`
at a larger `n` on this model, which is now the only known configuration where
the corpus has room to move. And the arms differ in one way beyond the skills
whenever a session under-implements: the control's four failed `gateway` runs
are excluded from every structural mean, so those columns describe different
populations and are not an effect.
