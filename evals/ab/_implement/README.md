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
| `feed` | go-data-structures | A nil slice marshals to `null`, not `[]`, so an account with no activity changes the shape of a JSON response the client parses strictly | Marshals the empty and fully-filtered documents and asserts both members render as `[]` |
| `ledger` | go-defensive | Keeping the caller's slice means the documented snapshot is not immutable, because the caller still owns the backing array | Mutates the input slice after `New`, then asserts the rendered report did not move |
| `catalog` | go-error-handling | Putting the SKU in the message with `%v` serves the operator and silently cuts the caller off from the reason | Asserts `errors.Is` reaches all three sentinels the Source reports and that the message still names the SKU |
| `gateway` | go-http | A zero timeout is no timeout, so an edge server built as `&http.Server{Addr: addr, Handler: h}` holds stalled and idle connections until it runs out | Asserts `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout` and `IdleTimeout` are all non-zero |

`ledger` and `gateway` also carry the bait, which is a separate thing from the
trap. `ledger` renders two formats and says a third has never been asked for,
which is the `report` fixture's temptation to grow a `Formatter` interface with
one implementation per format. `gateway` pins only `NewServer`, so how the
routing, filtering and encoding behind it are decomposed — a handler struct, a
repository interface, one mux with three closures — is entirely the model's
choice. `feed` and `catalog` still pin every declaration and therefore carry no
bait; that is the reason the numbers below have nothing to say about them.

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

`gateway` is admitted. On the [2026-09-07 claude run](../../../docs/evidence/2026-09-07-go-implement-gateway-ledger-claude.md)
the arms separate completely on how much code a passing implementation costs:
162.0 lines without the skill against 107.0 with it, control at 149/153/184
against 99/105/117, with functions down 52% and branches down 42%. Correctness
is tied at 6/6 golden, so the fixture measures conciseness, not whether the
package works.

`ledger` points the same way — 47.0 against 38.0 lines — without separating;
its interval includes zero at `n=3`. Its `Formatter` bait is inert: no session
in either arm grew an interface, so only the line count is doing work.

`feed` and `catalog` are not admitted. They pin every declaration, which is
what the run below found to be the problem.

## Why the first attempt measured nothing

The [first discovery run](../../../docs/evidence/2026-09-07-go-implement-discovery-minimax-m3.md)
found two blocking problems.

The traps are saturated. `no-skill` passed the hidden test in 12 of 12 runs on
every fixture: the unaided model reached for a non-nil slice, set all four
server timeouts, and kept the error chain without being told. The traps are
live — each golden test fails on the stub and passes against an idiomatic
reference — the model simply does not fall into them. Since the control was
perfect on a small, cheap model, a stronger one cannot do worse, so this is not
a matter of picking a different model.

Triggering ruled out that arm. A `go-*` skill loaded in only half the baseline
runs, and in none of the three `feed` runs, so half the skilled arm was a
control that happened to cost more. That turned out to be specific to the
runner and model rather than to the skill descriptions: seven trigger evals
written from the exact phrasings that missed were added to the `train` set in
`evals.json` and all seven pass against the `claude` CLI. An implementation run
therefore belongs on a runner where the skill reliably loads.

`ledger` and `gateway` were rebuilt after that run to pin a single entry point
instead of every declaration, which is what produced the effect recorded above.
`feed` and `catalog` were left alone and remain saturated: correctness has no
room there and neither does structure, so they measure nothing until they are
rebuilt the same way.

## Adding a fixture

A fixture earns its place only when `no-skill` is measurably worse than
`baseline`; name the trap in the table above before adding one. `go test
./cmd/abrun` checks that every directory here has a golden test and that no
name collides with a refactor fixture, since the golden directories are keyed
by fixture name.
