# Review fixtures

`go run ./cmd/abrun -corpus review` hands the model a working package with
seeded defects and asks for a pull-request review. Nothing is edited — the
claude arm has no `Edit` or `Write` tool at all — and the model's final
message is the review. It is scored against a hidden key,
`_golden/<fixture>/key.json`, that names each defect by the source line it
sits on.

The other two corpora measure the code the model writes. This one measures
what it sees: until 2026-09-12 `go-code-review` had never been measured on a
Claude model, and every claim about the skill was a claim about its text.

## What a fixture is

A package that compiles, is gofmt-clean, passes `go vet` and its own tests,
and has nothing left for `go fix -diff` to propose — so the tools a reviewer
runs first do not hand over what the fixture wants found by reading. Into it,
eight to twelve defects are seeded, each owned by one skill, each on a line
the key can name. A key entry also says which linter in the bundled
`golangci.yml` reports the line, when one does; `orders`, `worker` and
`invoice` carry 4, 1 and 2 such lines, and the report separates that recall
from the recall on lines no tool reports.

The first runs ([report](../../../docs/evidence/2026-09-12-go-review-corpus-smoke.md))
found the corpus has room on Sonnet 5 — unaided recall 0.74 — and none on
Opus 5, which found every seeded defect unaided at n=1. A fixture that
measures the skill on Opus 5 has to carry defects an unaided Opus 5 review
misses; the `unkeyed` citations of an unaided run are the wrong place to look
for them, since those turned out to be real defects the key lacked. The
three fixtures built to that brief are below; their first run says the brief
does not hold either.

Each fixture also carries one or two **baits**: a correct line that a weak
review flags anyway — the open `_, _ = w.Write(body)` with its reason, the
deferred `tx.Rollback` guard, `slices.Clone` before a sort. A citation on a
bait counts against the review, as does a citation that lands on no key
entry at all, so a review that names every line does not score. Only the
first citation on a line is a finding; a later one on the same line is a
supporting reference that counts toward a defect it lands on and is never
unkeyed. A bait earns its place only when no review would mention the line
unless it thought something was wrong: `clone-then-sort` was cited in every
`invoice` review of the first runs, approvingly as often as not, and is the
first bait to replace.

The key is a set of substrings, not line numbers: `abrun` resolves each
`match` against the pristine fixture before the run and refuses to start
when one is missing or ambiguous. `TestReviewKeysResolve` does the same on
every push and checks that no two windows overlap within the tolerance.

## Fixtures

| Fixture | Owners | Defects | Baits | The shape |
| --- | --- | ---: | ---: | --- |
| `orders` | go-http, go-context, go-error-handling, go-security, go-database, go-defensive | 12 | 2 | An HTTP order API over `database/sql`: a bare `ListenAndServe`, a 500 that carries `err.Error()`, `==` on a wrapped sentinel, an `Authorization` header in a log line, an unbounded body, a request context handed to a goroutine that outlives the request, string-built SQL, a `rows` loop with no `Close` and no `Err`, an insert that runs on `s.db` inside a transaction, a monetary amount in `float64`, and a commit error returned beside the new id |
| `worker` | go-concurrency, go-context, go-defensive, go-data-structures, go-error-handling, go-testing | 13 | 2 | A bounded worker pool: a context in a struct, an unbuffered result channel that deadlocks any batch larger than `Workers` and leaks senders on early return, an append without the mutex the reader takes, an aliased return, `copy` into a zero-length slice, a lock held across a callback, an uncancellable sleep, `%v` where the chain is needed, `math/rand` for a bearer token, a `New` that accepts zero workers and then blocks forever, and a test with `context.Background()` and a sleep |
| `invoice` | go-interfaces, go-code-refactor, go-documentation, go-code-review | 7 | 1 | A text renderer: a one-implementation interface, a middle-man type, an options struct nothing reads, a `formatCents` that renders a credit as `$-1.-50`, a hand-rolled sum next to `Total`, a doc comment that does not start with the name, and dead code |
| `books` | go-database, go-defensive, go-code-review, go-data-structures, go-testing | 14 | 3 | A ledger over `database/sql`: an `Equal` that compares four of five fields, a `Clone` that shares a map, a memo limit counted in bytes, a fee that rounds refunds toward zero and a test that asserts the wrong value, `strings.Split("")`, a `time.Time` map key, `AddDate` over a month end, `Truncate` for a local midnight, NULL scanned into a `string`, two same-typed columns scanned in the other order, an inclusive keyset cursor, and a destination balance read unlocked and written back absolute, and a self-transfer that mints money |
| `partner` | go-security, go-http, go-resilience, go-defensive, go-error-handling, go-context, go-testing | 17 | 2 | An HTTP client with retries: a `bytes.Reader` drained by the first attempt, an `Idempotency-Key` minted per attempt inside the helper, a redacting `String` on the pointer receiver while the value is logged, a host allowlist by `HasSuffix` without the dot and a `Label` path that skips it, a padded decoder for an unpadded encoder, a `Flush` error assigned to a local, `Retry-After` in nanoseconds, a `defer cancel()` per poll, a token bucket that never refills under load, a half-open breaker that lets every caller through, a test that passes for the wrong reason, and the four the first reviews added: a base URL without a host, a local throttle that burns a retry, a ticker that panics on zero, and an unvalidated limiter |
| `vault` | go-security, go-concurrency, go-http, go-data-structures, go-code-review, go-testing | 13 | 3 | A tenant file store: `HasPrefix` without the separator, a quota charged after an unlocked check and never released by `Delete`, an ETag hashed in map order, a pooled buffer returned after `Put`, the uploader's `Content-Type` served inline, `//evil.example` past a leading-slash check, the first `X-Forwarded-For` hop as the rate-limit key, a `WriteTimeout` that cuts large downloads, a test that compares an error's message, and from the first reviews a half-written file a concurrent read can see, a labels map stored by reference, and a 304 without its ETag |

`invoice` is the Less Code half of the checklist; the one Must Fix in it is
a correctness defect with no tool behind it, placed among structure that is
merely unnecessary, to see whether a review that subtracts still reads.

## Fixtures for a saturated model

`books`, `partner` and `vault` were added on 2026-09-12 after the first runs
found Opus 5 finding every seeded defect of the first three fixtures unaided.
Each defect in them is drawn from a class a thorough single pass over the
file was expected not to resolve:

- **Evidence in two places.** The column order of a `SELECT` against the
  order of its `Scan`; a nullable column in `schema.sql` against a `string`
  target; `RawURLEncoding` against `URLEncoding`; a `String` on the pointer
  receiver against a value in a log call; the host check `Status` makes
  against a `Label` that does not.
- **A sequence of calls.** A token bucket whose refill truncates to whole
  seconds and then moves its clock; a breaker whose half-open state nothing
  closes; a `bytes.Reader` at EOF from the attempt before; a `sync.Pool`
  `Put` deferred before the caller has used the bytes.
- **An enumeration one short.** `Equal` over four of five fields, `Clone`
  over one of two reference fields, a `Delete` that removes the file and the
  index entry but not the bytes it charged.
- **Standard-library semantics that read as correct.** `AddDate` over a
  month end, `Truncate(24h)` outside UTC, `time.Time` as a map key,
  `strings.Split("")`, `len` for characters, `HasSuffix` and `HasPrefix`
  without the separator, `time.Duration(secs)`.
- **A passing test that disagrees with the contract.** `Fee(-30, 250)`
  asserted as 0 under a name that explains it away; a host-rejection test
  that passes because the receipt fails to decode first.

The baits in these fixtures are forms the plugin's own skills prescribe, so
a stale-knowledge review flags them and a skilled one must not: `omitzero`
on a `time.Time`, `time.After` in a cancellation-aware select (Go 1.23
collects the timer), `math/rand/v2` for jitter (the go-resilience example
itself), `context.WithoutCancel` around an audit write, `limit + 1` for a
has-more flag. Two of them are lines the bundled gate reports — gosec G404
on the jitter and G705 on the JSON write — which is the tool being wrong,
and what a review that parrots the tool will repeat.

The [first run](../../../docs/evidence/2026-09-12-go-review-corpus-hard-opus-5-medium.md)
of these three, Opus 5 medium at n=2 in both arms, did not find the room it
was built for: unaided recall **0.97 over 44 defects**, must recall 1.00, in
both arms, at 26 citations a session. The one class that held is the test
that passes for the wrong reason — `partner/test-any-error` 0/4: every review
found the two bugs behind it and none asked why the test still passed. Nine
of the key's 44 entries were added from what those reviews cited (a
self-transfer that mints money, a base URL with no host, a half-written file
a concurrent read sees, a labels map stored by reference, and five more),
which is where the room went. On Opus 5 the corpus measures precision and
filing, not recall: baits 0.38, unkeyed 0.27 unaided and 0.35 with the skill,
must defects filed under Must Fix 0.84 and 0.78 — and those are the columns a
wording change to `go-code-review` has to move there.

The [2026-09-18 run](../../../docs/evidence/2026-09-18-go-review-checklist-cut-n2-sonnet-5-medium.md),
Sonnet 5 medium on all six fixtures at n=2, measured `go-code-review` with
eleven linter-covered checklist rows cut and seven restated against release
1.20.1: recall **0.72 → 0.80 over 152 key lines**, must 0.82 → 0.87, the
sixteen linter-reported lines 12 → 15, baits 8 → 5 of 26, unkeyed 25 → 15,
cost level. `invoice`, `orders` and `vault` moved, `books` and `worker`
held, `partner` gave one line back and stays the fixture with the room
(≈0.55 in both arms).

## What gets measured

| Column | What it says |
| --- | --- |
| `recall` | Defects found over defects seeded, mean over valid runs. A citation within two lines of the defect's window counts. |
| `must` | The same over the must-severity defects — data loss, a security hole, a deadlock, a wrong result. |
| `as-must` | Of the must defects found, how many the review filed under Must Fix; a real defect filed as a nit is a finding the author will skip. |
| `read-only` | Recall over the defects no bundled linter reports: what reading found. |
| `baits` | Baits cited over baits seeded. |
| `unkeyed` | Citations on neither a defect nor a bait, over all distinct citations. |
| `skill` | Valid runs in which `go-code-review` loaded. |

Under the table, one row per defect gives found/valid per arm, so a report
can say which defect the skill text made the difference on rather than that
recall moved.

A run is valid when the session ended, the fixture still builds and was not
edited, and there is a final message to score. The refactor and implement
columns (`Δlines`, golden) are not printed for this corpus; they would be
zero by construction.

## Adding a fixture

Write the package clean first and run it through gofmt, `go vet`, `go fix
-diff` and its tests. Then seed defects one at a time, each with one owner
skill and one line, and write the key entry before the next defect. Run
`golangci-lint run --config ../../../skills/go-linting/assets/golangci.yml`
over it and record in `tool` what the bundled gate reports. Add at least one
bait that is not within two lines of any defect. Then run
`go test ./cmd/abrun -run 'Review'`.

Anchor a defect where a review will cite it. The first Opus 5 runs cited
short functions by their header line (`guard.go:59` for a defect on line 65)
and a range for a block, so a defect inside a function of a dozen lines is
matched on the `func` line with a `span` down to the defective statement,
and a bait is kept at least three lines from any line such a range could
cover. When a key is amended after a run — the unkeyed citations of the
first runs on every fixture so far were partly real defects the key lacked —
`go run ./cmd/abrun -rescore <report.json> -out <report.json>` scores the
saved reviews again against the current key, without a session, and marks
the report `rescored`.
