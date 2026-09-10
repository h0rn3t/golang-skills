# Go Implementation Corpus Control — Sonnet 5 at medium effort on the Claude CLI (n=5)

The implementation half of the same-day pair with the
[refactor control](2026-09-10-go-refactor-control-sonnet-5-medium-n5.md): same
runner, model, effort, seed, plugin digest and day, only the corpus differs. Where
the refactor corpus separates the arms on three of four fixtures, this one
separates on none — and one fixture, `gateway`, remains unusable on this model
at this effort for a reason that belongs to the fixture, not to either arm.

## Run

- Finished: 2026-09-10 19:15 UTC
- Runner: `claude` 2.1.267
- Model: `claude-sonnet-5`, reasoning effort `medium`
- Seed: `1`
- Corpus: `implement`
- Fixtures: `catalog`, `feed`, `gateway`, `ledger`
- Arms: `no-skill`, `baseline`
- Repetitions: 5 per fixture and arm, 40 sessions total
- Baseline plugin SHA-256:
  `5028c91b61340fe99a22284cf0fe4ba32f48d8e7b1cf83baa0e412c88708a2ea`
  — byte-identical to the refactor run's arm
- Plugin source: checkout `d0a83f360fac7d8cd419a40007a2e14b2ef29660`
  (release 1.7.0), clean apart from this report's own untracked files under
  `docs/evidence`
- Toolchain: Go 1.27.1 linux/amd64
- Raw report: [`2026-09-10-go-implement-control-sonnet-5-medium.json`](2026-09-10-go-implement-control-sonnet-5-medium.json)
- Raw report SHA-256:
  `677b55a9a457bf91e9e547e954546ab382bda91492d6fcbdc60d4045bd7a2429`
- Session transcripts:
  [`2026-09-10-go-implement-control-sonnet-5-medium.traces.tar.gz`](2026-09-10-go-implement-control-sonnet-5-medium.traces.tar.gz)
  (SHA-256 `56b836023574cc1247e255bd50c6695804f3f476c1dec2a9a14a77c1445fac56`),
  one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $0.8724 control, $3.9635 baseline, $4.8359 total (**4.54x**)

```bash
go run ./cmd/abrun -corpus implement -arms no-skill,baseline -n 5 -j 4 -seed 1 \
  -model claude-sonnet-5 -effort medium -keep \
  -out ../docs/evidence/2026-09-10-go-implement-control-sonnet-5-medium.json
```

All 40 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. The tool set was `Skill,Read,Glob,Grep,Edit,Write`,
so neither arm had a shell to run the package with.

## Results

Here the golden test is the specification, so correctness is the gate and size
is the score only among the runs that pass it. Means cover valid runs.

| Fixture | No skill | Skill | Effect | perm p | valid | golden ns · bl |
|---|---:|---:|---:|---:|---|---|
| `catalog` | +25.2 ± 7.4 | +25.7 ± 7.3 | +0.4 | 0.97143 | 4v3 | 4/5 · 3/5 |
| `feed` | +46.2 ± 1.8 | +43.8 ± 1.0 | −2.4 | 0.11905 | 5v5 | 5/5 · 5/5 |
| `ledger` | +39.6 ± 4.3 | +34.6 ± 2.3 | −5.0 | 0.11905 | 5v5 | 5/5 · 5/5 |
| Three evaluable | +37.86 | +36.08 | **−1.78** | — | 14v13 | 14/15 · 13/15 |
| `gateway` | — | — | not evaluable | — | 2v0 | **2/5 · 0/5** |

| Arm | runs | valid | Δlines | Δtypes | Δiface | Δfuncs | Δbranch | build | golden | skill | $/run |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| `no-skill` | 20 | 16 | +43.31 | 0.12 | 0.00 | 0.12 | +5.44 | 100% | 80% | 0% | 0.0436 |
| `baseline` | 20 | 13 | +36.08 | 0.15 | 0.00 | 0.15 | +4.46 | 100% | 65% | 95% | 0.1982 |

The two arm rows average different fixture sets — `gateway` is knocked out of
one arm entirely and nearly out of the other — so the comparable number is the
«three evaluable» row, not the arm row. On those three fixtures the skill is
worth **−1.78 lines** with correctness at 13/15 against 14/15: no separation in
either direction. No fixture reaches p < 0.05, and `catalog` is a dead tie.

The over-engineering bait is untouched again: `Δiface` and `Δpattern` are 0.00
across all 40 sessions, and new types and functions are level between the arms
(0.12/0.12 against 0.15/0.15). Unlike the 2026-09-08 control, the skilled arm
does not write *more* code here; it writes marginally less, by less than the
run can resolve.

## `gateway` is still not evaluable on this model at this effort

Eight of ten sessions fail the same three assertions, in both arms:

```
golden_test.go:179: HEAD /healthz      = 200, want 405
golden_test.go:179: HEAD /accounts     = 200, want 405
golden_test.go:179: HEAD /accounts/a-1 = 200, want 405
```

The fixture documents «A known path asked for with any method other than GET is
a 405» and the golden test checks it, but `net/http.ServeMux` deliberately
routes HEAD into the handler registered for `GET /path`. `go-http` has named
this since `f12c73b` — "a pattern without a method matches every method; `GET`
also matches `HEAD`… reject it explicitly on that endpoint or register its
matching HEAD pattern with a 405 handler" — and it fired in **5/5** `gateway`
sessions of the skilled arm, which still failed 5/5. The control failed 3/5.
The idiom the model reaches for on this fixture is not moved by the skill text
that describes it.

Three skilled `gateway` sessions also returned `null` rather than `[]` for the
empty account list — the nil-slice trap `feed` is built around, on a fixture
where `go-data-structures` also loaded 5/5. No control session made that
mistake. At three of five it is a single-run-scale observation, not a measured
regression, but it is the one place in this run where the skilled arm looks
worse than the control on correctness, and it deserves a targeted re-run rather
than a footnote.

This reproduces the [2026-09-09 finding](2026-09-09-go-controls-sonnet-5-medium.ru.md)
(control 1/5, skill 0/5) on a newer tree. The corpus owner's decision is still
open: either HEAD on a known path stops being a 405 and the doc comment matches
`ServeMux`, or the fixture keeps the trap and is scored knowing that the control
arm cannot pass it either. Until then `gateway` measures neither the skill nor
the model on Sonnet 5 medium.

## `catalog`

One failure in each arm, and the same one both times:
`the source was asked for "sku-404" 2 times, want 1`. Deduplication is done
through the results map, a SKU that resolves to an error never lands there, and
the repeat goes back to the source. It is an ordinary bug appearing at the same
rate on both sides — 4/5 against 3/5 says nothing about a difference in
correctness.

## Skill routing

`go-code` fired in 19/20 baseline sessions, `go-data-structures` in 19,
`go-style-core` in 19, `go-error-handling` in 17, `go-http` in 5 (5/5 on
`gateway`) and `go-linting` in 4. One `catalog` session loaded nothing and still
passed the golden test. Owner routing is far better than the
[2026-09-08 control](2026-09-08-go-implement-control-sonnet-5.md), where
`go-data-structures` and `go-defensive` reached 0/5 of their own fixtures — and
it buys no measurable change in what the sessions produced, which is the same
result the `gpt-5.6-luna` pair recorded on 2026-09-08. Firing rate and outcome
are separate results.

## What this changes

- The Sonnet 5 new-code verdict stands: no established benefit. The
  2026-09-08 reading — same correctness, *more* code and helpers — softens to a
  −1.78-line difference that no fixture supports, so «benefit not established»
  is now backed by a null result rather than by a slightly adverse one.
- Correctness on the three evaluable fixtures is 13/15 against 14/15: tied
  within this sample.
- Cost is 4.54x, the highest this corpus has recorded among the runs that
  report USD at all (2.32x on 2026-09-08, 2.36x on 2026-09-09).
- `gateway` needs the HEAD/405 decision before it can carry weight on this
  model, and the three `null` empty-list failures in the skilled arm need their
  own run.
