# Changelog

All notable changes to this repository are documented here.

## [Unreleased]

- `README.md` and `README.uk.md` gain an "Updating" section: the Claude Code
  plugin update commands, `npx skills update -g` for Codex, and a manual
  replace-not-overwrite refresh.

## [1.22.2] - 2026-09-24

- `evalrun`'s quality judge returns its verdict through `claude -p
  --json-schema` and reads `structured_output`; the "JSON only" prompt line
  and the brace-slicing `extractJSON` are gone.
- `go-code-review` drops the "fast pass, then deep pass" Depth note; the
  Review Procedure's risk and section order is the one reading order.
- `OVER-ENGINEERING.md` no longer caps a new-code report's omission notes at
  three lines; report length follows `go-style-core` "How Much To Say".
- `ARCHITECTURE.md` and `ARCHITECTURE-CHECKS.md` drop revision-handoff and
  skill-evaluation wording; the section is "Report contract", and its table
  states the required behavior per situation.
- `TABLE-DRIVEN-TESTS.md` removes `tt := tt` only under a `go` directive of
  1.22 or later and runs `go fix -forvar` on the packages in scope, matching
  `BEHAVIOR-TRAPS.md` and the gate's scope rule.

## [1.22.1] - 2026-09-23

- `go-code` now uses a reader pass after its delete pass, permits direct
  returns, and selects standard-library collection operations only when they
  clarify the call site and preserve ordering, ownership, and empty results.
  `go-code-refactor` applies the same reader-path check alongside LOC; the
  new-code examples show a useful one-use name and a justified one-call helper.
- `go-code`'s Plain Code and Delete Pass no longer cap body comments at one
  per overridden default. A comment may give a constraint, an overridden
  default, or the reason behind a choice; narration of the next line is still
  deleted.
- `go-code`'s `NEW-CODE-EXAMPLES.md` sets `largest` and `doc.Largest` on two
  lines instead of one parallel assignment of two unrelated facts.
- `go-code`'s Declaration Budget gains a fourth rule: a step at another level
  of abstraction than its caller (request decoding beside the business
  decision) may be extracted with one call site; a helper that only restates
  two or three lines stays inline. `go-code-refactor` cites the four rules,
  which also settles its conflict with `PLAYBOOK.md` §2.
- `go-code`'s Plain Code names an intermediate value when its expression
  nests deeper than one call or the name says what the expression does not,
  and allows a comment giving the business or historical reason for a choice.
  The Delete Pass no longer removes blank lines inside a function, and keeps
  a single-use name that explains.

## [1.22.0] - 2026-09-23

- Removed archived model-run reports, generated traces, and obsolete
  maintenance notes from the repository. Structural tests and golden fixtures
  remain under `evals/`; model-run output is local scratch data.
- Corrected examples that contradicted their own skill. `go-documentation`'s
  `doc-template.go` no longer documents a sentinel it never returns, a no-op
  `Close`, an unused constant, or concurrency safety it lacks.
  `go-http`'s `WEB-SERVER.md` no longer returns a one-implementation interface
  from its constructor, drops the doc line about a `Shutdown` method that does
  not exist, and discards the response write with its reason as
  `go-http/SKILL.md` requires. `go-error-handling`'s `ERROR-TYPES.md` drops the
  `switch err { case ErrDuplicate: }` form marked Good. `go-naming`'s
  `REPETITION.md` wraps with `%w` without a `failed to` prefix.
  `go-generics`'s `CONSTRAINTS.md` no longer shows `slices.Contains` or a
  single-interface type parameter as Good. `go-data-structures`'s `SLICES.md`
  uses `slices.Clone` and `bytes.Clone`. `go-testing`'s `INTEGRATION.md` shows
  `TestMain` returning rather than a `runMain` exit-code helper.
- The dependency ladder is stated one way: `go-database` and
  `OVER-ENGINEERING.md` now match the three rungs `go-packages` owns.
- `check-errors.sh` 1.3.0 reports log-and-return when the logger is a struct
  field (`s.logger.Error`) and when the error is returned wrapped
  (`return fmt.Errorf("...: %w", err)`); before, only `return err` after a
  `log`, `logger`, or `slog` call counted.
- The bundled `golangci.yml` enables `iface` (the `opaque` check only),
  `nilnil`, `unparam`, and `revive`'s `early-return`, `indent-error-flow`, and
  `superfluous-else`, each enforcing a rule a skill already states. On a
  deliberately over-built sample file the config reports 5 findings where it
  reported 2; on five existing codebases (this repository's `evals/`, `fiber`,
  `excelize`, two MCP servers) it adds 0 to 4.4% to the findings. `revive`'s
  `unused-parameter` and `iface`'s `unused` and `identical` were measured and
  left off as noise. `abrun`'s lint counts from before this change are not
  comparable with counts after it.

## [1.21.2] - 2026-09-19

- Current plugin release with 24 Go 1.27 skills, routing hooks, the shared
  verification gate, and the structural evaluation suite.
