# Changelog

All notable changes to this repository are documented here.

## [Unreleased]

### Added

- `go-code-refactor` gains `references/OVER-ENGINEERING.md`: the cut tags
  (`delete:`, `stdlib:`, `dep:`, `yagni:`, `shrink:`), a Go-specific hunt list
  (single-implementation interfaces, forwarding wrappers, `util` packages,
  hand-rolled stdlib, dependencies Go now ships, flexibility nobody uses), the
  ranked one-line audit output, and the prove-it-before-you-cut gate. Used when
  the ask is "what can we delete" instead of "make this read better".
  `go-code-review` routes to it for the bloat lane; correctness and security
  stay with the review checklist.
- `go-code-refactor` gains `scripts/check-debt.sh`: harvests the `Kept:`
  markers a refactor leaves behind into a ledger. Markers naming neither
  `Ceiling:` nor `Fix:` are tagged `no-trigger` and drive exit 1, so a
  deliberate shortcut cannot quietly become permanent. Supports `--json` and
  `--limit`; fixtures live in `evals/fixtures/debt/`.
- Pinned the marker convention: `Kept:` / `Ceiling:` / `Fix:` are fixed
  prefixes so the ledger can find them. The example in `SKILL.md` now uses one
  prefix per line.
- `TestManifestCounts` pins the advertised counts (skills, reference files,
  scripts, asset templates) in `README.md`, `README.uk.md`, `plugin.json`, and
  `marketplace.json` against what is on disk, and requires the plugin and
  marketplace versions to agree and to have a `CHANGELOG.md` section.

### Fixed

- `check-naming.sh` exited 0 for a nonexistent path: its `exit 2` ran inside a
  process substitution, so the caller continued with an empty file list and
  reported a clean scan. The target is now validated in the main shell.

### Changed

- Gave the nesting rule a single owner. "Reduce nesting / early returns /
  unnecessary else" now belongs to `go-style-core`; `go-control-flow` and
  `go-error-handling` route to it instead of restating it. This also breaks
  the three-way circular route (`go-style-core` → `go-error-handling` →
  `go-control-flow` → `go-style-core`) that used to send readers in a loop.
- Gave the `iota` enum rule a single owner. `go-defensive` no longer repeats
  the "start enums at one" block and routes to `go-declarations`, which owns
  the form and the zero-is-default exception.
- `go-declarations` hands map/set selection to `go-data-structures` and size
  hints to `go-performance` instead of implying it owns them.
- `go-defensive`: filled in the empty "Time, Struct Tags, and Embedding"
  heading with its routing line, and dropped the stale "enum zero values"
  promise from the `TIME-ENUMS-TAGS.md` routing entry.
- `docs/RULE_OWNERSHIP.md` gains rows for nesting and for iota enums.
- `go-style-core` and `go-code-review` now route back to `go-code-refactor`,
  the two consumers the refactor-workflow ownership row already named. The
  link was one-way: `go-code-refactor` referenced 16 of the other 20 skills
  and no skill referenced it.
- `TestRuleOwnershipMap` now pins both rules: each needle must appear in its
  owner document and nowhere else under `skills/`.

## [1.10.0] - 2026-08-29

### Added

- Added the `go-code-refactor` skill: behavior-preserving refactoring of
  existing Go. Owns the workflow (orient → baseline → audit → `go fix` →
  verifiable steps → verify → report), the four stop-and-ask cases, the
  observable-behavior contract, and the findings-not-fixes rule; routes every
  underlying style rule to its owner skill.
- Added `references/BEHAVIOR-TRAPS.md` — the Go rewrites that look equivalent
  and are not (nil vs empty, `defer`, typed nil, concurrency, slice aliasing,
  receivers, struct layout, evaluation order), with a pre-commit checklist.
- Added `references/PLAYBOOK.md` — transformations ordered by payoff, with the
  readability hierarchy and the anti-patterns of "cleanup".
- Added `references/MODERNIZATION.md` — Go 1.21–1.27 features sorted into safe
  swaps, conditional ones, and report-only, plus the toolchain shifts that
  break green tests on their own. Every entry verified against go1.27.0.
- Added `scripts/verify-refactor.sh` — baseline/after/diff/leaks harness with
  `--json`, `--limit`, `--out`, and 0/1/2 exit codes. The diff compares
  per-test verdicts, so a renamed, skipped, or vanished test is caught.
- Added `assets/refactor-report.md` — report structure that leads with
  deletions and keeps unfixed findings in their own section.
- Added six trigger evals (including a negative control for new-code requests)
  and one quality eval for the new skill.

### Changed

- README, `plugin.json`, and `marketplace.json` now count 21 skills, 51
  references, 9 scripts, and 5 assets.
- `docs/RULE_OWNERSHIP.md` gains the refactor-workflow ownership row and lists
  `go-code-refactor` as a consumer of the `go-linting` verification gate.
- `docs/SCRIPT_JSON_CONTRACTS.md` documents the `verify-refactor.sh` shapes.
- `.gitignore` excludes the script's `.refactor-verify/` output.

## [1.9.0] - 2026-08-29

### Added

- Added `COMPATIBILITY.md` (previously referenced by the README but missing):
  Go 1.27 baseline, language and standard-library tables by version, and the
  commands that re-verify every claim against an installed toolchain.
- Added a canonical verification gate to `go-linting` (`gofmt`, `go vet`,
  `go test -race`, `go fix -diff`, `golangci-lint`, `govulncheck`), with
  route-only pointers from `go-code-review`, `go-style-core`, `go-testing`,
  `go-concurrency`, and `go-error-handling`.
- Added a `go fix` modernizer catalogue to `go-linting`.
- Added a stdlib-first dependency ladder to `go-packages`, covering the
  modules Go 1.27 absorbed (`uuid`, `encoding/json/v2`).
- Added Go 1.27 generic-method guidance and `maphash.ComparableHasher` to
  `go-generics`.
- Added `httptest.NewTestServer`, `testing/synctest`, `t.Context`, `t.Output`,
  and `t.ArtifactDir` guidance to `go-testing`.
- Added `errors.AsType[T]` guidance to `go-error-handling`.
- Added `slog.NewMultiHandler` and `slog.GroupAttrs` to `go-logging`.
- Added `os.Root` path confinement to `go-defensive`.
- Added `new(expr)` to `go-declarations`.
- Added `TestGoVersionBaseline` to the eval suite, pinning the Go 1.27
  guidance and failing on pre-Go-1.22 loop-variable captures anywhere in
  `skills/`.
- Added agent-facing authoring conventions to `docs/SKILL_AUTHORING_TEMPLATE.md`
  and two new ownership rows to `docs/RULE_OWNERSHIP.md`.

### Changed

- Raised the documented baseline from mixed 1.13–1.24 minimums to Go 1.27;
  every `> Compatibility:` note now routes to `COMPATIBILITY.md`.
- Rewrote `go-defensive/references/BOUNDARY-COPYING.md` around `slices.Clone`
  and `maps.Clone`, with a shallow-copy table and `url.URL`/`url.Values.Clone`.
- Modernized `references/WEB-SERVER.md`: `run()` pattern,
  `signal.NotifyContext`, `http.NewCrossOriginProtection`, full server
  timeouts, and a handled JSON encode error. Verified with `go vet`/`go build`.
- Replaced `sort.Slice` with `slices.SortFunc` and the three-clause counting
  loops with `for i := range n` in examples.
- Documented `for i := range n` and `iter.Seq` ranging in `go-control-flow`.
- Bumped CI to Go 1.27 and golangci-lint v2.13.1; `evals/go.mod` to `go 1.27`.

### Fixed

- Removed dead pre-Go-1.22 loop-variable captures (`item := item`, `i := i`,
  `tt := tt`) from concurrency and testing examples.
- Replaced `runtime.NumCPU()` with `runtime.GOMAXPROCS(0)` for worker sizing —
  `NumCPU` ignores the cgroup CPU limit and over-provisions in containers.
- Dropped stale "for older Go versions" fallbacks for APIs available on every
  supported release.

## [1.8.0] - 2026-06-20

### Added

- Added repository-level third-party notices for bundled `source/` snapshots.
- Added a Go compatibility policy for version-sensitive standard-library
  guidance, eval harness expectations, and golangci-lint config verification.
- Added a release checklist covering changelog, provenance, compatibility,
  validation, and tagging steps.

### Changed

- Expanded the validation workflow to run on pull requests, pushes to `main`,
  `v*` tags, and manual dispatch.
- Pinned skill validation to `agentskills-validate@1.0.1`.
- Added Go setup, eval tests, and golangci-lint config verification to CI.
- Clarified README provenance and license wording.

### Fixed

- Corrected the README project tree so `evals/` and `source/` are shown as
  top-level directories and `evals/fixtures/` is included.
