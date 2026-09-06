# Refactor Skill Review Fixes

## Goal

Make the staged `go-code-refactor` guidance safe and internally consistent,
and make `evals/cmd/abrun` produce results that can actually support a claim
about code-quality changes.

## Scope

The change covers the review findings in:

- `skills/go-code-refactor/SKILL.md`
- `skills/go-code-refactor/references/{CATALOG,MECHANICAL,SAFETY-NET,STRUCTURAL}.md`
- `evals/cmd/abrun`
- `evals/ab/README.md`
- focused architecture and runner tests under `evals`

It will not run paid live model evaluations. It will not redesign unrelated Go
skills or modify the four fixture contracts.

## Guidance Corrections

- Treat `deadcode` output as a candidate for investigation, never proof that an
  exported API is safe to delete.
- State that `gopls` rename protects source references and compilation but may
  still change dynamic behavior; low-risk rename requires an unexported symbol,
  no string/reflection contract, and focused verification.
- Split the "do not refactor" gate into true stop cases and cases that only
  change sequencing, such as writing characterization tests first.
- Require Replace Temp with Query expressions to be pure, stable, and cheap.
- Describe cross-package functions and mutable variables separately: functions
  may forward; mutable variables have no cross-package alias and need one owner
  plus accessors or an API migration.
- Qualify additive API compatibility for exported struct fields, concrete
  methods, and interface methods.
- Replace the invalid recursive `gofmt ./...` example and correct the coverage
  explanation for packages without tests and `-coverpkg`.

## Runner Design

Keep the existing command and report shape, adding only what closes a proven
measurement gap:

1. Propagate every non-zero Claude CLI exit even when partial stdout exists.
2. Walk the entire fixture tree when counting Go structure and hiding model
   tests, so moving code into a subpackage cannot improve the score invisibly.
3. Treat a golden failure as a failed live result and reject fixtures without a
   golden directory.
4. Add an optional `-reference-root` plugin-tree arm so a previous checkout can
   be compared with the current checkout. The existing `baseline` arm remains
   the current tree.
5. Shuffle jobs with a recorded deterministic seed to avoid arm-major ordering.
6. Validate positive repetitions, parallelism, timeout, selected arms, tasks,
   and required fixture/golden structure before launching model sessions.

Published result tables require the raw JSON report, explicit model selection,
and both reference and current arms. Existing uncommitted historical numbers
will be removed from the normative README narrative.

## Test Strategy

Tests are added before implementation and must initially expose the current
behavior:

- partial stdout plus non-zero subprocess exit returns an error;
- nested production files contribute to metrics;
- nested `_test.go` files are hidden;
- missing golden fixtures fail validation;
- reference-root and current arms use different plugin trees;
- seeded job construction is deterministic and not arm-major;
- golden failure is reported as failure;
- documentation regression checks pin the corrected safety statements and
  executable command forms.

After implementation:

```bash
gofmt -d <changed-go-files>
go test ./cmd/abrun -count=1
go test -race ./cmd/abrun -count=1
go test ./... -count=1
go test -race ./... -count=1
go vet ./...
git diff --check
```

Any full-suite failure caused by the locally installed lint binary is reported
as unavailable evidence rather than silently treated as clean.

## Success Criteria

- No reviewed guidance contradicts the identical-behavior contract.
- The documented commands execute with the stated scope.
- Failed or structurally hidden model output cannot enter A/B summaries as a
  successful simplification.
- The runner can compare a previous plugin tree with the current one without
  conflating that comparison with the no-skill discovery control.
- Focused tests, race tests, vet, formatting, and repository architecture checks
  are green, or any external toolchain blocker is isolated and reported.
