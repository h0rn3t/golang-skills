# Changelog

All notable changes to this repository are documented here.

## [Unreleased]

## [1.23.0] - 2026-09-27

- `go-code` ports ponytail's intensity levels. `/go-code lite <task>` or
  `lite mode` names the lazier rung in a `lazier:` line and builds the shape
  the request suggests; `/go-code ultra <task>` skips every part the request
  does not state and questions stated ones a higher rung covers
  (`Need <X>? <Y> covers it.`). No level word means `full`, the current
  behavior. No level touches the gate, the Contract Table, explicit
  requirements, or the never-cut list, and a behavior-preserving refactor
  always runs at `full`. The level holds for the rest of the session.
  `TestRuleOwnershipMap` pins `## Intensity` to `go-code`. The wording is
  unmeasured.
- New `hooks/go-restraint-ladder.sh`, wired like ponytail's ruleset: on
  `SessionStart` (startup, resume, clear, compact) and `SubagentStart` in a
  directory holding Go it prints the restraint ladder, read at run time from
  `OVER-ENGINEERING.md`, and the session's level row from `go-code`; on
  `UserPromptSubmit` a level word records the level for the session.
  `GOLANG_SKILLS_LADDER=lite|full|ultra` sets the starting level, `off` turns
  the hook off. `TestLadderHook` drives it. About 4 KB of context per session
  start and per subagent.

## [1.22.4] - 2026-09-26

- `go-code` finds references semantically before changing a used symbol —
  gopls MCP or the Claude Code `LSP` tool when either is already wired, grep
  otherwise — and routes to `go-code-refactor`'s `GOPLS.md`, whose ownership
  row now covers reference lookup. gopls diagnostics do not replace the gate.
- `GOPLS.md`: the `LSP` tool route no longer claims a `rename` operation or
  needs `ENABLE_LSP_TOOL=1`; rename goes through `go_rename_symbol` or the CLI.

## [1.22.3] - 2026-09-25

- `README.md` and `README.uk.md` gain an "Updating" section: the Claude Code
  plugin update commands, `npx skills update -g` for Codex, and a manual
  replace-not-overwrite refresh.
- **Behavior change for copies of `skills/go-linting/assets/golangci.yml`:**
  the edit hook lints with this file when a repository has none, so a finding
  on an idiom the skills teach became code the skills call slop. Fewer
  findings now, one linter per line (restore the old file from `v1.22.2`):
  - `revive` `exported` no longer reports under `internal/` or `cmd/`; code
    nothing imports needs no `// NewItem creates a new Item.`
  - `prealloc` is removed: it reported every `var out []T` filled by
    `append`, and its fix turns a v1 JSON `null` into `[]`.
  - `perfsprint` sets `string-format: false` and `strconcat: false`;
    `fmt.Sprintf("project/%s", p)` is no longer reported.
  - `errcheck` excludes `(*database/sql.Rows).Close`,
    `(*database/sql.Tx).Rollback`, and `(io.ReadCloser).Close`, the deferred
    closes in the go-database and go-http examples; `Close` on a written file
    still reports.
  - `gocyclo` reports from 30 instead of 15, above a flat chain of error
    checks.
  - `gosec` excludes G304, which fired on every `os.ReadFile(path)`; client
    paths are go-security's `os.Root` rule.
  - `revive` gains `var-naming`, `receiver-naming`, and `error-strings`, so
    `userId` is reported again (the explicit rule list had turned revive's
    defaults off).
- `check-docs.sh` 1.2.0 skips `package main`, methods of unexported types, and
  `Error`, `String`, `Unwrap`, `ServeHTTP`, and the JSON/text marshalers, as
  `revive` does. JSON shape and exit codes are unchanged.
- `check-interface-compliance.sh` 1.2.0 does not count an interface that a
  value is already assigned, returned, passed, or converted to, and its text
  output asks whether a consumer needs the interface instead of suggesting
  `var _ I = (*T)(nil)`. JSON keys and exit codes are unchanged.
- `bench-compare.sh` defaults to `--count 10`, the count go-performance asks
  for.
- Factual corrections checked against go1.27.1 and golangci-lint 2.13.2:
  go-generics says the compiler gates generic methods by the `go` directive
  (self-referential constraints and conversion inference stay ungated);
  `crypto/rand.Read` is not error-checked; `errors.Is(err, fs.ErrNotExist)`
  replaces `os.IsNotExist`; `go vet` finds printf wrappers without the `f`
  suffix and `-printf.funcs` checks only names ending in `f`; `errcheck` does
  not report `_ =`; the go-linting `nolint` example is one nolintlint accepts;
  a recovering middleware logs `debug.Stack()`, not `%+v`; `*Context` calls
  add no request fields under `JSONHandler`; the TLS 1.3 `MinVersion` example
  is marked as the TLS 1.3-only case; `encoding/xml` expands no DTD entities;
  the go-testing Resource Routing lines name the files that hold each topic;
  `CATALOG.md` links the duplication fold to its real owner.
- `TestBundledLintConfig` runs the bundled config over `evals/fixtures/lint`
  and pins which findings it reports and which it does not.
- Positive examples follow the pack's own rules, since a model copies the
  code rather than the caveat beside it. `TestPositiveExamplesCarryNoSlop`
  fails on a section banner, a `// go-<skill>:` tag, or a `failed to` /
  `could not` / `couldn't` error text in any Go block not marked Bad,
  Before, or fragment. Rewritten:
  - `WEB-SERVER.md`: one `package main` with `run()`, a concrete store, the
    handler at its registration; no banners, rule tags, one-implementation
    interface, or no-op `CrossOriginProtection` on a GET-only server.
  - go-http: the default routing block has no capture-free closure; the
    `HEAD` 405 contract is a package function in its own block; the create
    handler maps errors inline and decodes into `var req struct`.
  - `PLAYBOOK.md` §2 extracts one step, decoding, and says why.
  - go-error-handling: `<operation> <key>: %w` texts; `%v` only for the
    named opaque case; a flat `switch` in `ERROR-FLOW.md`; `case err != nil`.
  - go-logging: event logs instead of narration; a local `LevelVar` in
    `run()` instead of `init`; `type loggerKey struct{}`; no SQL arguments
    in a debug log.
  - go-defensive: `IsExpired(now, expiry)` instead of a `Checker` with an
    injected clock; no generic `Must[T]`; `rand.Text()` inline; one recover
    example.
  - go-concurrency: errgroup with `SetLimit` and one `return g.Wait()`; no
    `processInBackground`, no channel semaphore, no narrating comments.
  - go-testing: the table-test template and `gen-table-test.sh` carry no
    `TODO` or commented-out code; failure messages name the call and input.
  - go-packages: subcommands as `run(args) error` with
    `flag.ContinueOnError`, a checked `Parse`, and a usage error on no input.
  - go-database: `withTx` rolls back once through its `defer`; the lock
    query needs no `min`/`max`; `db.Close` is not discarded.
  - go-security: SSRF puts the hostname allowlist first and checks arbitrary
    destinations in `net.Dialer.Control` on the dialed address.
  - go-resilience: the backoff shift is capped, so a long budget cannot
    overflow to a negative wait.
  - Smaller fixes in go-documentation, go-interfaces, go-functions,
    go-naming, go-style-core, go-context, and go-performance.
- `abrun` records readability, which the counts could not see on a model
  whose golden tests saturate: `max_func_lines`, `p90_func_lines`,
  `max_nesting`, and three slop proxies — `echo_docs` (a doc comment whose
  first sentence only restates the name), `one_call_helpers` (an unexported
  function of at most three statements used once), and `log_and_return`.
  Each run also keeps its production `source`. The fields are new keys; a
  report written before them loads, and its summary prints a dash instead of
  a zero.
- `abrun -judge` asks a blind pairwise judge (`-judge-model`, default
  `claude-fable-5-1`) which of two arms' diffs reads better
  (`-judge-pair`, default `reference,baseline`). Each pair is judged in both
  orders; a preference counts only when both agree, the rest are ties with a
  positional-disagreement count, and a failed call is `skipped (reason)`.
- `abrun -runner opencode` accepts `-effort`, passed as `--variant` after
  checking that the model declares that variant (opencode accepts an unknown
  one silently); arm homes get the operator's model catalog, so a model newer
  than the binary's bundled catalog resolves; a failed session reports the
  transcript's error event instead of a bare `exit status 1`.

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
