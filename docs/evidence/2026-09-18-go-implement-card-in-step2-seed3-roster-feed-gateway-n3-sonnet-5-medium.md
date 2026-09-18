# The card in step 2 against 1.20.1, a third run at n=3 — Sonnet 5 medium, and the three runs pooled

The [afternoon's two runs](2026-09-18-go-implement-card-step-roster-feed-gateway-n2-sonnet-5-medium.md)
left one figure open: over both, the working tree cost +8% a session against
release 1.20.1 ($0.384 against $0.354), with the sign the same way in each run
and the size inside the reference arm's own movement between runs ($0.371 to
$0.337). Two repetitions a fixture cannot tell a lean from a slope. This run
adds three more on a new seed, with the same trees and the same three
fixtures, so the question can be put to 21 sessions a side.

| Arm | Tree |
|---|---|
| `reference` | `git worktree` of `89d702a` (release 1.20.1), plugin SHA-256 `d478dd31988660fcb8934eba8e56e7e378221b321b278073006d4ecb4af19ca0` |
| `baseline` | `git worktree` of `66f06cb` (the 1.21.0 branch head), plugin SHA-256 `bfcfed52ecd8273c265ff05738e65a40716a3caa9b16a78dde590c732cd14ef0` — byte-identical to the afternoon's run 2 baseline |

## Run

- Finished: 2026-09-18 19:24 UTC
- Runner: `claude` 2.1.267; model `claude-sonnet-5`, effort `medium`; seed `3`
- Corpus `implement`, fixtures `roster`, `feed`, `gateway`, arms `reference`, `baseline`, 3 repetitions per fixture and arm, 18 sessions, `-j 4`, 0 CLI errors
- Report: [`2026-09-18-go-implement-card-in-step2-seed3-roster-feed-gateway-n3-sonnet-5-medium.json`](2026-09-18-go-implement-card-in-step2-seed3-roster-feed-gateway-n3-sonnet-5-medium.json) (SHA-256 `a368089f58829f84bf656a14f1f354d2031d9f4b3568994fddc6ce683de35339`); traces [`….traces.tar.gz`](2026-09-18-go-implement-card-in-step2-seed3-roster-feed-gateway-n3-sonnet-5-medium.traces.tar.gz) (SHA-256 `4202fe3f9c4e167bba04e258e251455b597f7fa35293f869fcbff021af2a6219`), one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $6.75

```bash
go run ./cmd/abrun -corpus implement -tasks roster,feed,gateway -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <worktree of 89d702a> -arms reference,baseline \
  -n 3 -j 4 -seed 3 -timeout 10m -keep -verbose -out <json>
```

Toolchain and event columns as in the afternoon's report; no session had a
shell tool, the edit hook ran after every `.go` edit.

## The reading

| Arm | Valid | Golden | Lint clean | Δlines | Skill calls/s | Skill turns/s | Card `Read` | Gate blocks | Hook findings | Assistant turns/s | $/session |
|---|---|---|---|---|---|---|---|---|---|---|---|
| `reference` | 9/9 | 9/9 | 6/9 | +48.9 | 4.78 | 3.11 | 0/9 | 6 | 56 | 16.9 | 0.386 |
| `baseline` | 9/9 | 8/9 | 6/9 | +45.7 | 4.67 | 2.67 | 1/9 | 5 | 39 | 15.0 | 0.364 |

| Arm | Fixture | Rep | Golden | Lint | Δlines | Skills loaded (in load order) | Skill turns | Card `Read` | Gate blocks | Hook findings | Turns | $ |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `reference` | `feed` | 0 | pass | 0 | +39 | go-code, go-style-core, go-testing, go-data-structures, go-error-handling, go-defensive | 2 | no | 0 | 4 | 13 | 0.313 |
| `reference` | `feed` | 1 | pass | 0 | +41 | go-code, go-style-core, go-testing, go-error-handling | 4 | no | 1 | 4 | 13 | 0.280 |
| `reference` | `feed` | 2 | pass | 0 | +43 | go-code, go-style-core, go-testing, go-error-handling | 3 | no | 1 | 4 | 15 | 0.290 |
| `reference` | `gateway` | 0 | pass | 0 | +98 | go-code, go-style-core, go-http, go-error-handling, go-security, go-testing | 3 | no | 1 | 14 | 28 | 0.710 |
| `reference` | `gateway` | 1 | pass | 0 | +75 | go-code, go-style-core, go-http, go-error-handling, go-testing | 3 | no | 1 | 10 | 21 | 0.551 |
| `reference` | `gateway` | 2 | pass | 0 | +89 | go-code, go-style-core, go-http, go-error-handling, go-testing, go-linting | 4 | no | 0 | 7 | 20 | 0.470 |
| `reference` | `roster` | 0 | pass | 5 | +15 | go-code, go-style-core, go-testing, go-linting | 4 | no | 1 | 5 | 16 | 0.321 |
| `reference` | `roster` | 1 | pass | 5 | +17 | go-code, go-style-core, go-interfaces, go-testing | 2 | no | 0 | 5 | 12 | 0.255 |
| `reference` | `roster` | 2 | pass | 5 | +23 | go-code, go-style-core, go-testing, go-interfaces | 3 | no | 1 | 3 | 14 | 0.282 |
| `baseline` | `feed` | 0 | pass | 0 | +41 | go-code, go-style-core, go-error-handling, go-testing | 3 | no | 1 | 5 | 20 | 0.356 |
| `baseline` | `feed` | 1 | pass | 0 | +40 | go-code, go-style-core, go-testing, go-error-handling | 3 | no | 1 | 3 | 12 | 0.264 |
| `baseline` | `feed` | 2 | pass | 0 | +42 | go-code, go-style-core, go-testing, go-documentation, go-error-handling | 4 | no | 1 | 4 | 13 | 0.272 |
| `baseline` | `gateway` | 0 | **fail** | 0 | +74 | go-code, go-style-core, go-http, go-error-handling, go-testing, go-security | 2 | no | 0 | 2 | 13 | 0.375 |
| `baseline` | `gateway` | 1 | pass | 0 | +86 | go-code, go-style-core, go-http, go-error-handling, go-testing | 2 | no | 0 | 8 | 24 | 0.727 |
| `baseline` | `gateway` | 2 | pass | 0 | +78 | go-code, go-style-core, go-http, go-error-handling, go-testing, go-logging | 3 | no | 1 | 2 | 14 | 0.383 |
| `baseline` | `roster` | 0 | pass | 5 | +15 | go-code, go-style-core, go-interfaces, go-testing | 3 | no | 0 | 5 | 11 | 0.277 |
| `baseline` | `roster` | 1 | pass | 5 | +17 | go-code, go-style-core, go-interfaces, go-testing | 2 | whole | 0 | 5 | 14 | 0.342 |
| `baseline` | `roster` | 2 | pass | 5 | +18 | go-code, go-style-core, go-interfaces, go-testing | 2 | no | 1 | 5 | 14 | 0.282 |

**The sign turned.** $0.364 against $0.386 a session (−6%), 15.0 assistant
turns against 16.9, 39 hook findings against 56 — the columns that leaned the
other way in the afternoon lean this way here, by about the same amount.

**Pooled, the three runs are level.** Over 21 sessions a side (runs at seeds
1, 2 and 3; the seed-1 baseline carried the card as a numbered step, the
other two as the first clause of step 2):

| Arm | Sessions | $/session (mean ± SD) | `roster` | `feed` | `gateway` | Golden | Δlines |
|---|---|---|---|---|---|---|---|
| `reference` (1.20.1) | 21 | 0.368 ± 0.144 | 0.258 | 0.296 | 0.549 | 21/21 | +47.5 |
| `baseline` (1.21.0) | 21 | 0.376 ± 0.142 | 0.275 | 0.317 | 0.534 | 19/21 | +45.6 |

+2.2% a session, permutation p ≈ 0.86 over 20,000 shuffles; the per-run
figures were +3%, +15% and −6%. A session's cost on this corpus has a
standard deviation of about 40% of its mean — `gateway` runs from $0.37 to
$0.75 — and two repetitions a fixture cannot resolve an 8% difference against
that. The 1.21.0 text neither costs nor saves on this model at this effort.

**Golden 19/21 against 21/21, both misses `gateway`.** The afternoon's was the
nil-list `null`; this run's (`baseline` `gateway` r0) answers `HEAD` with 200,
the clause that has flipped every day it was measured. `gateway` in the
`reference` arm is 7/7 across the three runs and 5/7 in the `baseline`;
across the day's other Sonnet 5 runs on this fixture the 1.21.0 tree is 6/6
(the card-route runs' baseline arm), so the day's total for the tree is 11/13
against 7/7. Two misses on two different known clauses at n=13 do not point
at the text; they are the reason `gateway` stays in the corpus.

**The card: 1/21 against 0/21, the tree's one read unprompted.** This run's
`baseline` `roster` r1 read it whole before the first edit with nothing in
the host asking; the same tree read it once more in the card-route runs'
baseline arm. Two of 36 sessions is the rate the step-2 wording buys on
Sonnet 5 medium; the host route measured the same evening buys 15/15.

## What this establishes

- The +8% of the afternoon was noise: the third run reverses it and the
  pooled difference is +2% at p ≈ 0.86 over 42 sessions.
- `gateway` correctness on Sonnet 5 medium is 11/13 for the 1.21.0 tree
  today and 7/7 for 1.20.1, on two different clauses each recorded flipping
  before; not an effect of the text at this n, and not something to leave
  out of a release note either.
- The event counts that did move in the afternoon — `(hook)` on every
  claimed check, 8/11 against 1/11 — are unaffected by this run, which did
  not re-count them.

## Next

- `gateway` at n≥10 on the two trees in one run is what a correctness claim
  in either direction costs; the fixture's `HEAD` and nil-list clauses are
  the ones to count separately.
