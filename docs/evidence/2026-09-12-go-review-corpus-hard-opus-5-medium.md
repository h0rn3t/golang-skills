# Review corpus — three fixtures built for a saturated model, Opus 5 medium, no-skill against baseline (n=2)

The [first review-corpus runs](2026-09-12-go-review-corpus-smoke.md) found
Opus 5 finding all 29 seeded defects of `orders`, `worker` and `invoice`
unaided, and said the next fixture had to carry defects an unaided Opus 5
review misses. `books`, `partner` and `vault` were written to that brief the
same evening ([corpus README](../../evals/ab/_review/README.md#fixtures-for-a-saturated-model)):
every defect in them is from a class a thorough single pass over a file was
expected not to resolve — evidence split across two places, a bug that
appears only when a sequence of calls is simulated, an enumeration one field
short, standard-library semantics that read as correct, a passing test that
contradicts the contract. This run is the brief's test, and the brief fails:
unaided Opus 5 finds 0.97 of them.

## Runs

- Runner: `claude` 2.1.267; seed `1`; corpus `review`; fixtures `books`,
  `partner`, `vault`; arms `no-skill`, `baseline` (the working tree, plugin
  SHA-256 `503d79cc6e8e3a4a8ab192a83743937032727d818fd369f07dbd470139a20d04`)
- Opus 5 medium, n=2, `-j 4`, finished 2026-09-12 20:20 UTC, 12 sessions,
  no errors, no fixture edited, `go-code-review` loaded in 6/6 baseline and
  0/6 `no-skill` sessions
- Key as run: `books` 13 defects, `partner` 12, `vault` 10 — 35, with 8 baits.
  Nine entries were added afterwards from what the reviews cited and two
  anchors widened (below); the report was then scored again with
  `abrun -rescore`, so the JSON carries the scores against the key of 44 and
  the field `rescored` says so. The sessions and their outputs are the
  originals.
- Report [`2026-09-12-go-review-corpus-hard-opus-5-medium.json`](2026-09-12-go-review-corpus-hard-opus-5-medium.json)
  (SHA-256 `44193aea6293cc494a704df757644a4e152aa26672602e19c3ea9642c8ec6abd`),
  transcripts [`….traces.tar.gz`](2026-09-12-go-review-corpus-hard-opus-5-medium.traces.tar.gz)
  (SHA-256 `6e1054760131bbca016459000b1ef1f19bc660e012b4b9e3227121f58601a09f`)
- Cost $2.04 `no-skill`, $3.34 `baseline` — $0.340 against $0.557 a session,
  1.64x
- Toolchain: Go 1.27.1 linux/amd64; every fixture gofmt-, vet- and
  `go fix`-clean with its tests passing (`TestReviewFixturesAreToolClean`)

```bash
go run ./cmd/abrun -corpus review -runner claude -model claude-opus-5 -effort medium \
  -arms no-skill,baseline -tasks books,partner,vault -n 2 -j 4 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-12-go-review-corpus-hard-opus-5-medium.json
go run ./cmd/abrun -rescore ../docs/evidence/2026-09-12-go-review-corpus-hard-opus-5-medium.json \
  -out ../docs/evidence/2026-09-12-go-review-corpus-hard-opus-5-medium.json
```

## Results

Against the key of 44. Columns as in the [corpus README](../../evals/ab/_review/README.md#what-gets-measured).

| Arm | sessions | recall | must | as-must | read-only | baits | unkeyed | citations / run | `verified` / run | $ / run |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| `no-skill` | 6 | **0.97** | 1.00 | 0.84 | 0.97 | 0.38 | 0.27 | 26.0 | 0.0 | 0.340 |
| `baseline` | 6 | **0.97** | 1.00 | 0.78 | 0.97 | 0.38 | 0.35 | 26.8 | 10.7 | 0.557 |

Against the key of 35 as run, recall was 0.97 unaided and 0.94 with the skill,
and unkeyed 0.40 and 0.50; the difference is the nine entries below.

### Per defect (found / sessions), the ones not at 2/2 in both arms

| Defect | `no-skill` | `baseline` | What happened |
|---|---:|---:|---|
| `partner/test-any-error` — the host-rejection test passes because the receipt fails to decode first | **0/2** | **0/2** | Every review found the base64 mismatch and the suffix check; none asked why `TestStatusRejectsForeignHost` passes. Three reviews wrote a `client_test.go` finding with no line — "coverage gaps line up with the bugs" — and one that it "only exercises an obviously unrelated host" |
| `partner/limiter-unvalidated` — `NewLimiter(0, 0)` refuses everything | 1/2 | 2/2 | Added from the reviews; the one unaided session that missed it did not mention `NewLimiter` |
| `vault/partial-file-visible` — an upload written into its final name | 2/2 | 1/2 | Added from the reviews |

The other 41 are 2/2 in both arms. Two of them were 1/2 in `baseline` against
the key as run — `partner/breaker-probe-herd` and `vault/pool-alias-after-put` —
because the review cited the function header (`guard.go:59`, `server.go:257`)
while the anchor sat on the defective statement six and seven lines below.
Those anchors, and eight more on short functions, now match on the `func` line
with a `span` down to the statement; the corpus README records the rule.

### Baits (hit / sessions)

| Bait | `no-skill` | `baseline` |
|---|---:|---:|
| `books/omitzero-tag`, `vault/omitzero-tag` — `json:",omitzero"` on a `time.Time` | 2/2 each | 2/2 each |
| `partner/math-rand-jitter` — `math/rand/v2` for backoff jitter (gosec G404) | 1/2 | 1/2 |
| `books/fetch-plus-one` — `limit + 1` for the has-more flag | 1/2 | 0/2 |
| `books/deferred-rollback` | 0/2 | 1/2 |
| `partner/time-after-in-select`, `vault/without-cancel-audit`, `vault/open-discard` | 0/2 | 0/2 |

`omitzero` was flagged in every session of both arms — as an unknown option,
or as `omitempty` misspelled — which is the one bait that measured what it
was built to measure, a fact about Go 1.24 the model does not hold. The skill
text does not carry it either (`go-linting` names the `omitzero` modernizer in
a table; nothing says what the option does), so the baseline arm could not
have corrected it.

## What the reviews cited that the key did not

Unaided and skilled reviews alike anchored 19 to 31 findings a session on
fixtures of 10 to 13 defects. Clustering the citations no key entry covered:

- **Nine were defects the fixtures carried without meaning to**, all cited in
  at least two sessions, all now in the key: a transfer from an account to
  itself mints money (`books`, cited in 4/4 sessions — the absolute
  destination write applied to the row just debited); a base URL with no host
  makes `HasSuffix(host, "")` true and disables the receipt guard (`partner`,
  4/4); a labels map stored by reference and handed back by `List` (`vault`,
  4/4); an upload copied straight into its final name so a concurrent `Open`
  serves half a file (`vault`, 3/4); a refusal by the shop's own limiter
  burning one of the three partner attempts (`partner`, 3/4);
  `time.NewTicker(0)` panicking (`partner`, 3/4); `NewLimiter` accepting zero
  (`partner`, 2/4 and 3/4 for the wall clock seeded past the injectable
  `now`); a 304 without its `ETag` (`vault`, 2/4); an error message ending in
  a dangling colon (`partner`, 4/4).
- **Supporting references the scorer counts as anchors.** A finding that
  names its evidence on a later line — "`memo` is nullable (`schema.sql:17`,
  and `Transfer` writes `NULLIF` at `store.go:125`)" — puts a first `.go`
  citation on that line, and the scorer reads it as a new finding. Three
  sessions carried that one; the rule that only the first citation *on a
  line* is an anchor does not reach it.
- **Judgment the fixture does not grade**: a token compared by map lookup
  rather than in constant time (4/4 `vault` sessions), a nil logger not
  guarded, sentinels not grouped in one `var` block, the destination account's
  `frozen` flag not checked when the doc says only the source is, error
  messages that carry the wrapped prefix, no `t.Parallel()`, no `ctx` on the
  store's methods.
- **Two doc comments the reviews read as defects**: `SettledAt` and `Totals`
  are filled in outside the package, which the fixture now says on the same
  lines, and `Submit`'s "the receipt Status and Label take" read as three
  return values; both reworded without moving a line.

## Reading

- **The brief fails.** Defects whose evidence is in two files, that need a
  sequence of calls simulated, that omit one member of an enumeration, or
  that rest on a standard-library semantic are found by an unaided Opus 5
  medium review of a 200-to-330-line package at the same rate as the plainly
  local ones: 0.97, against 1.00 on the first three fixtures. The model reads
  every file, traces every path it names, and cites 26 lines a session on a
  package with 13 defects; recall on this model is not where a fixture finds
  room, whatever the defect class.
- **One class held: a test that passes for the wrong reason.** Both bugs
  behind `TestStatusRejectsForeignHost` were found in every session, and no
  session asked what the test was actually testing. The reviews treat tests
  as coverage to be extended, not as claims to be checked; the two test
  defects they did catch (2/2 in both arms) were local or reached from the
  code side — `vault/test-error-string` is a message compared on its own
  line, and `books/test-enshrines-refund-fee` was noticed by every review
  after it had found the rounding defect the row asserts.
- **The skill does not move recall and slightly worsens filing.** 0.97 in
  both arms; must defects filed under Must Fix 0.84 unaided against 0.78
  with the skill (one baseline `books` session filed one of five must defects
  as Must Fix); unkeyed 0.27 against 0.35; 10.7 `verified` and 10.5
  `plausible` markers a session against none. At n=2 none of these is an
  effect; the direction is the checklist adding citations, not findings, at
  1.64x the cost — the same statement the first smoke made on 29 defects.
- **What the corpus can measure on Opus 5 is precision and filing.** Baits
  0.38 in both arms, `omitzero` flagged in every session, 19 to 31 anchored
  findings a session on 10 to 13 defects, and must defects filed below Must
  Fix in a fifth of the cases. A wording change to `go-code-review` that
  says what to leave out — a stale fact, a judgment call, a style row on a
  correctness review — has room in those columns; one that adds checks does
  not.
- **The key improves with every run.** Nine of 44 entries came from the
  reviews; the first three fixtures gained three of 32 the same way. The
  `unkeyed` column is a precision measure only once that pass has been made,
  and `abrun -rescore` exists so the pass costs no sessions.

## What this changes

- The review corpus has six fixtures and 76 defects. On Sonnet 5 the first
  three had room (0.74 unaided); on Opus 5 none of the six does, and the
  next measurement of `go-code-review` on Opus 5 is of baits, unkeyed and
  as-must at n=5, not of recall.
- Where recall room on Opus 5 might still be found, going by this run: tests
  that pass for the wrong reason as a fixture's main load; a diff-scoped
  prompt, where the defect is in the interaction of a changed line with
  unchanged code the review is not invited to read; and package sizes past
  what one session reads in full. The first is cheap to build and this run
  says it holds; the other two need `abrun` changes before a fixture.
