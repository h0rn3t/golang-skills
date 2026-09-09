# Implementation fixtures

`go run ./cmd/abrun -corpus implement` hands the model a package whose exported
declarations and documentation are already written and whose bodies all
`panic("not implemented")`. The golden test is the specification, so unlike the
refactor corpus it can be failed outright — which is the point. In the refactor
corpus both arms passed 20/20 golden on every published run, so correctness had
no room to move and the skill could only be scored on how much structure it
removed. Here it is scored on whether the package works.

The leading underscore in the directory name keeps `findTasks` from offering
these as refactor fixtures, so the two corpora can never be averaged together.

## What a fixture is

Each fixture pins its API in Go rather than in prose: real signatures, real doc
comments, empty bodies. The golden test therefore compiles by construction, and
a run cannot fail for having named something differently. What that buys in
reliability it gives up in scope — API design is not measured here, only the
implementation behind a fixed API.

The doc comments state the situation and the observable contract. They never
name the technique. `ledger` says a Ledger is a snapshot that nothing a caller
does can change; it does not say `slices.Clone`. `gateway` says the server is
published straight to the open internet with clients that stall and idle; it
does not say `ReadHeaderTimeout`. Naming the technique would hand the trap to
the control arm and the fixture would measure nothing.

## Fixtures

Each trap is a defect a Go reviewer would send back, owned by exactly one
skill, and checkable by a test that never reaches the model.

| Fixture | Owner | The trap | What the golden test does |
| --- | --- | --- | --- |
| `feed` | go-data-structures | A nil slice marshals to `null` and a nil map to `null`, so an account with no activity changes the JSON type of a member the client parses strictly | Renders the empty and fully-filtered documents and asserts `events` is `[]`, `kinds` is `[]` and `counts` is `{}` |
| `ledger` | go-defensive | Keeping the caller's slice means the documented snapshot is not immutable, because the caller still owns the backing array | Mutates the input slice after `New`, then asserts the rendered report did not move |
| `catalog` | go-error-handling | Putting the SKU in the message with `%v` serves the operator and silently cuts the caller off from the reason | Asserts `errors.Is` reaches the sentinel and the transport failure, and that the message still names the SKU |
| `gateway` | go-http | A zero timeout is no timeout, so an edge server built as `&http.Server{Addr: addr, Handler: h}` holds stalled and idle connections until it runs out | Asserts `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout` and `IdleTimeout` are all non-zero |

Every fixture pins one entry point and leaves the internals free. Two also
carry bait, which is a separate thing from the trap: `ledger` renders two
formats and says a third has never been asked for, which is the `report`
fixture's temptation to grow a `Formatter` interface per format, and `catalog`
must not pay for a repeated SKU twice, which tempts a cache where a `seen` map
does. Neither bait has ever been taken in either arm — `Δiface` is zero across
every run of this corpus — so the line count is what carries the results below.

Every fixture was checked in both directions before it was committed: the
golden test compiles against the stub and fails on its panic, and it passes
against an idiomatic reference implementation. A fixture that only fails proves
nothing, and one that passes on the stub would score a no-op run as a success.

## What gets measured

Correctness is the gate, not the score. The `golden` column says whether the
package satisfies the hidden specification at all; a run that fails it is
excluded from every mean, the same way it is in the refactor corpus. Among the
runs that pass, the score is how little code it took to get there:

| Metric | What it says about an implementation |
| --- | --- |
| `Δlines` | Positive by nature here. How much code the passing implementation cost. |
| `Δbranch` | Decision points: `if`, `for`, `range`, and each case that names a value. Separates leaning on what the standard library already decides from re-deciding it by hand. |
| `Δexp` | Exported declarations. Every one the specification needs is already in the stub, so growth is public API the task never asked for. |
| `Δtypes`, `Δiface`, `Δfuncs`, `Δpattern` | The scaffold. On a fixture that pins every declaration they have nowhere to move; on one that pins a single entry point they are the whole question. |

`Δbranch` is not a score on its own and must be read against `Δtypes` and
`Δiface`. An interface with one implementation per format does not remove the
decisions, it replaces them with dispatch, so the scaffolded solution shows
*fewer* branches and more types. Flat with one switch is few types and some
branches; a hand-rolled mess is few types and many branches; a premature
abstraction is many types and few branches. Only the pair tells them apart.

The `skill` column asks only whether any `go-*` skill was reached, because each
fixture routes to a different owner rather than to one shared skill the way the
refactor corpus routes to `go-code-refactor`.

The same guards apply: a session that changed no file, or whose transcript
mentions the fixture corpus in the repository, is excluded and counted
separately.

## Status

| Fixture | Status | Lines, no skill → skill | Evidence |
| --- | --- | --- | --- |
| `gateway` | **admitted** | 152.6 → 99.8 (−34.6%), ranges disjoint, 95% CI −96 to −10 | [Opus 5, n=5](../../../docs/evidence/2026-09-07-go-implement-gateway-opus5.md) |
| `ledger` | directional | 47.0 → 38.0 (−19.1%), CI includes zero | [claude, n=3](../../../docs/evidence/2026-09-07-go-implement-gateway-ledger-claude.md) |
| `catalog` | directional | 26.0 → 20.3 (−21.8%), CI includes zero | [Opus 5, n=3](../../../docs/evidence/2026-09-07-go-implement-feed-catalog-opus5.md) |
| `feed` | not admitted | 50.0 → 49.0; the control wrote exactly 50 lines in all three runs | [Opus 5, n=3](../../../docs/evidence/2026-09-07-go-implement-feed-catalog-opus5.md) |

On `gateway` the skilled arm also wrote the *same* amount every time —
population standard deviation 4.4 lines against the control's 30.9 — which is
the more useful of the two properties when the question is what a change costs
to review. Correctness is tied at 5/5 golden in both arms there, and at 6/6 on
the other fixtures, so none of the runs above is evidence about defect rates.

No published run in this corpus is. On every model measured so far the control
arm passes the hidden test in every session, which is what the corpus was
designed to make visible and also what keeps it from producing a correctness
claim. Read every `n=5` correctness cell here as a dated observation of a served
model version, not a property of the fixtures.

Routing is the one thing that has moved. The
[2026-09-08 GPT-5.6-Luna re-run](../../../docs/evidence/2026-09-08-go-implement-control-gpt-5.6-luna-medium.md)
reached `go-http` in 4 of 5 `gateway` sessions against 2 of 5 the day before,
`go-error-handling` in 5 of 5 `catalog` and `go-data-structures` in 4 of 5
`feed`, and produced exactly the same result as before: golden 20/20 in both
arms and a corpus size difference under one line. The firing rate, the routing to
the owning skill, and the measured outcome are three separate results; on a model
whose control arm is already correct, improving the first two changes nothing in
the third.

Only `gateway` responded to the single-entry-point rebuild, and the reason is
the fixture design rule this corpus learned the hard way: **room to
over-engineer comes from a task with many small independent placement
decisions, not from an unpinned API.** `gateway` is five routes, three status
codes, an ordering rule and a filter with an error path. `feed` and `catalog`
are each one data transformation, and a specification precise enough for a
golden test to check mechanically is also precise enough to leave a single
sensible shape. That reasoning covers the *line* metric only. Whether `feed`
and `catalog` can fail unaided on some model — which would make the room they
lack room to over-engineer rather than room to get the thing wrong — currently
has no published run behind it.

## Why the first attempt measured nothing

The [first discovery run](../../../docs/evidence/2026-09-07-go-implement-discovery-minimax-m3.md)
found two blocking problems.

The traps are saturated. `no-skill` passed the hidden test in 12 of 12 runs on
every fixture: the unaided model reached for a non-nil slice, set all four
server timeouts, and kept the error chain without being told. The traps are
live — each golden test fails on the stub and passes against an idiomatic
reference — the model simply does not fall into them. Since the control was
perfect on a small, cheap model, a stronger one cannot do worse, so this was
read at the time as not a matter of picking a different model. Treat that
reading as provisional: saturation is a property of a served model *version*,
so it has to be re-measured on the model in front of you rather than inherited
from an earlier file. Every model measured since has saturated too, which is why
the corpus still has no correctness result.

Triggering ruled out that arm. A `go-*` skill loaded in only half the baseline
runs, and in none of the three `feed` runs, so half the skilled arm was a
control that happened to cost more. That turned out to be specific to the
runner and model rather than to the skill descriptions: seven trigger evals
written from the exact phrasings that missed were added to the `train` set in
`evals.json` and all seven pass against the `claude` CLI. An implementation run
therefore belongs on a runner where the skill reliably loads.

All four fixtures were rebuilt after that run to pin a single entry point
instead of every declaration. That is what produced the `gateway` result; on
`feed` and `catalog` it changed nothing, for the reason recorded above.

## Adding a fixture

The status table above is this corpus's half of the
[fixture selection](../README.md#fixture-selection) a screening or decision run
draws from: `gateway` is the only admitted row, and on a model whose control
saturates there is nothing here to spend sessions on at all.

A fixture earns its place only when `no-skill` is measurably worse than
`baseline`; name the trap in the table above before adding one. `go test
./cmd/abrun` checks that every directory here has a golden test and that no
name collides with a refactor fixture, since the golden directories are keyed
by fixture name.
