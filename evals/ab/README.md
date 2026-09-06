# A/B fixtures for skill wording

`go run ./cmd/abrun` runs one refactoring prompt against these fixtures under
two or more versions of `go-code-refactor/SKILL.md` and reports what changed in
the code the model wrote — not in the prose it produced.

## Arms

`no-skill` loads no plugin — the discovery control. `baseline` is the complete
skill tree currently checked out. Supplying `-reference-root` adds a
`reference` arm from another complete plugin checkout, which is the arm to use
for a before/after comparison. Every `variants/*.md` file becomes one more arm,
spliced into the current `SKILL.md` in front of the `## Workflow` heading.
`-arms no-skill,baseline` runs a subset.

A fixture earns its place only when `no-skill` is measurably worse than
`baseline` — that gap is the trap the fixture exists to catch, and it is what
any wording change has room to move. Name the trap in the table below before
adding one.

| Variant | What it tests |
| --- | --- |
| `pattern-sentence.md` | "Apply a suitable design pattern, but only if it improves the solution." — an instruction whose condition the model itself judges |
| `pattern-gate.md` | The same intent as a gate with externally checkable conditions: three existing duplicate sites, a shorter diff, no call site that reads worse |

## Fixtures

Each is a package where the honest refactor **removes** structure and the
tempting one **adds** it. Fixture files carry no hints — the model reads them,
so the intent is documented here instead.

| Fixture | Honest refactor | The temptation |
| --- | --- | --- |
| `dispatch` | One switch plus a shared put helper | A handler registry or per-kind Strategy for three branches that never leave the file |
| `report` | Early returns plus one row builder | A `Formatter` interface with `TextFormatter`/`CSVFormatter` for two formats no caller extends |
| `store` | Flatten the nesting, delete the duplicate miss branch | A `Repository` interface plus adapters over two plain maps |
| `pricing` | One plan table the three lookups share | Five plan types behind a `PricingStrategy` interface and a registry — the hardest bait: the plan list is duplicated across two files and a `TODO` promises a sixth plan next quarter |

## What gets measured

`_golden/<fixture>/golden_test.go` is copied in **after** the model finishes and
pins the observable behavior: exported signatures, error texts, key formats,
column widths. It never reaches the model, so it cannot be adjusted to match a
changed implementation — a failing golden test means behavior moved.

Any `_test.go` the model wrote is renamed out of the build first. A refactor is
supposed to leave characterization tests behind, and the helper types in them
collide with the golden file's by name; without this step the harness scores its
own collision as a behavior break. It also keeps the model's tests from being
the reason the golden run passes.

Per run: recursive line delta, declared types, interfaces, functions,
pattern-flavored identifiers, whether the package still builds, whether the
golden test passes, and which `go-*` skills fired. The summary averages
structural deltas only over runs that both build and pass the hidden golden
test; session errors and invalid runs remain visible in separate counts.

A wording that helps shows up as fewer lines with the golden test still green.
A wording that licenses growth shows up as `Δtypes`, `Δiface` and `Δpattern`
rising — that is the number the "only if it improves the solution" phrasing was
added to move.

## Evidence contract

The repository carries no result table without its raw JSON report. A published
comparison records the exact model, seed, prompt, repetitions, both complete
plugin roots, and the revisions or content hashes they came from. `abrun`
records a SHA-256 digest of every materialized plugin arm in the JSON report;
record the source revisions beside it when publishing the file.

Use `no-skill` versus `baseline` only to decide whether a fixture contains a
trap the plugin can catch. It does not measure whether a skill edit improved the
previous skill. For that claim, compare `reference` versus `baseline` with the
same fixtures and save the report.

Runs with a CLI/session error, build failure, or golden failure are not evidence
of a shorter behavior-preserving refactor. Keep them in the report for diagnosis
and exclude their structural deltas from the arm mean.

## Current validated result

The 2026-09-07 Opus 5 control completed 40/40 valid runs. Across the corpus the
skill produced 4.15 fewer production lines per run and 60% fewer new types. On
the only strong over-engineering trap, `report`, it reduced mean growth from
+33.4 to +16.8 lines (49.7%). Correctness was tied at 20/20 build and golden
passes in both arms. See the [full analysis and raw report](../../docs/evidence/2026-09-07-go-refactor-control-opus5.md).

## Running it

```bash
# Discover whether the current plugin beats no skill on these fixtures.
go run ./cmd/abrun -arms no-skill,baseline -n 3 -j 4 \
  -model claude-opus-4-1-20250805 -seed 1 -out control.json

# Compare a previous complete checkout with the current plugin tree.
go run ./cmd/abrun -reference-root ../golang-skills-before \
  -arms reference,baseline -n 5 -j 4 \
  -model claude-opus-4-1-20250805 -seed 1 -out before-after.json

go run ./cmd/abrun -tasks report -n 1 -verbose     # one fixture, one pass
```

Needs the `claude` CLI and the `go` toolchain on PATH. Each run is a full
headless session, so cost scales with fixtures x arms x `-n`.
