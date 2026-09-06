# The Safety Net, Sized to the Blast Radius

> Sources: Feathers, *Working Effectively with Legacy Code*; `go help testflag`; `go tool cover` docs; golang.org/x/tools/cmd/deadcode
> Authority: project policy for the tiers; advisory for the seam mechanics
> Minimum Go: 1.27 baseline
> Last verified: 2026-09-06

How much caution a refactor needs is not a fixed policy. It is set by how well
tested **the code you are about to touch** already is — not by the project's
overall coverage number. A repository at 90% can have the one function in the
diff at 0%; a repository at 30% can have it fully pinned. Measure the blast
radius, then pick a tier.

The gate that decides *done* is the repository's, and
[go-linting](../../go-linting/SKILL.md) owns it. This file decides only how much
net has to exist *before* the first edit.

## Contents

- [The three tiers](#the-three-tiers)
- [Measuring the blast radius](#measuring-the-blast-radius)
- [Two things the coverage number will not tell you](#two-things-the-coverage-number-will-not-tell-you)
- [Seams](#seams)

---

## The three tiers

| Tier | Coverage of the touched lines | Strategy | Transforms allowed |
|---|---|---|---|
| **High** | ≳80% | Refactor, verify after each step | Anything in [CATALOG.md](CATALOG.md) and [MECHANICAL.md](MECHANICAL.md) |
| **Medium** | ~40–80% | Harden the touched paths first, confirm green, then refactor | Same as high, once targeted tests provably reach the touched lines |
| **Low / zero** | <40%, or the touched lines specifically | Characterize first, then refactor | Inline; unexported rename with no dynamic-name contract; sprout or wrap for new behavior; no extract and no cross-package move until a net exists |

**High** — the net catches many regressions within one test run. Keep steps
small enough to attribute a failure and review the behavior claim; coverage
reduces feedback cost, not transformation risk to zero.

**Medium** — add targeted tests over the paths the change touches, confirm they
pass against the *current* code, and only then edit. The step that is easy to
skip: re-run with `-coverprofile` afterwards and check the new tests actually
execute the lines you are about to change. A test that imports the right package
but never reaches the branch looks like a net from the outside and catches
nothing.

**Low / zero** — write characterization tests first, recording what the code
does today, warts included. This is deliberately not a correctness test; you are
not asserting the behavior is right, only pinning it so the refactor can be
checked against it. Before writing any of them, run `deadcode -test ./...`:
some apparently untested code may be unreachable from the programs and tests in
the current module. Its output is a candidate for investigation, never proof
that deletion is safe. Exported API may have consumers the analysis cannot see,
and even unexported results still need the build-tag, generated-code,
`go:linkname`, and dynamic-use checks from the main deletion rule.

> **Normative**: A characterization test written after the edit proves nothing.
> It has to be green against the unchanged code first, or it is pinning the bug
> you just introduced.

## Measuring the blast radius

References first, coverage second — [GOPLS.md](GOPLS.md) owns finding every
caller semantically. Then measure only those packages:

```bash
go test -covermode=atomic -coverpkg=./... -coverprofile=cover.out ./...
go tool cover -func=cover.out    # per-function, ranked — read the functions in the diff
go tool cover -html=cover.out    # which branches are actually green
```

[go-testing](../../go-testing/SKILL.md) owns how to write the tests themselves;
`scripts/verify-refactor.sh` captures the before/after comparison.

## Two things the coverage number will not tell you

- **Go measures statements, not branches.** A line inside an `if` that ran once
  counts as covered even when the `else` never executed and the `switch` only
  ever hit one case. Treat the percentage as a floor on what is exercised, and
  read the branches in the code you are about to touch.
- **`go test -cover ./...` reports packages with no tests as 0%.** The separate
  `-coverpkg=./...` flag expands the set of packages instrumented in each test
  binary, so tests in callers can contribute coverage for dependencies. It does
  not turn reachability or an overall percentage into branch coverage.

## Seams

A seam is a place where behavior can be changed without editing that exact spot.
It is how a fake gets into a characterization test without first performing the
refactor the test exists to protect.

Two kinds matter in Go:

- An **object seam** — an interface, or a function-typed field or parameter,
  injected at construction. This is the one that matters, because Go satisfies
  interfaces implicitly: introducing one at the point of use touches only the
  consumer, never the package it depends on.
- A **build-tag seam** — swapping a whole implementation at build time via
  `//go:build`. Rare, and mostly for platform-specific substitution.

The enabling move for untested code with no seam: extract the smallest possible
interface, often one method, at the exact point where the code reaches for
something external — a database handle, the filesystem, a clock — and inject it
through the constructor instead of building it inline.

```go
// Before — the concrete field gives a test nowhere to substitute a small fake.
type Report struct {
    db *sql.DB
}

func newReport(db *sql.DB) *Report {
    return &Report{db: db}
}

// After — one method, declared where it is consumed. *sql.DB already
// satisfies it, so the package that provides it needs no change at all.
type rowQuerier interface {
    QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type Report struct {
    db rowQuerier
}

func newReport(db rowQuerier) *Report {
    return &Report{db: db}
}
```

Two cautions before this becomes a habit. It is a seam for code that has no
net — not a licence to split every struct into an interface plus an
implementation; PLAYBOOK §8 and the ladder in
[OVER-ENGINEERING.md](OVER-ENGINEERING.md) still apply, and
[go-interfaces](../../go-interfaces/SKILL.md) owns where an interface belongs.
Changing an **exported** constructor's signature to accept the seam is an API
change and stays outside a behavior-preserving refactor. Prefer an internal seam
or an additive constructor when the public signature must remain available.
