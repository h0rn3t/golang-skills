# Changelog

All notable changes to this repository are documented here.

## [Unreleased]

- The skills no longer mention gopls. `go-code-refactor/references/GOPLS.md`
  is deleted (73 reference files); `go-code` drops its "Go navigation: MCP
  first" section, the `navigation:` report line, and the navigation routing
  row; `go-code-refactor`, `go-code-review`, `go-interfaces`,
  `go-troubleshooting`, `go-linting`, and `go-error-handling` drop their gopls
  links and mentions. `CATALOG.md` uses `//go:fix inline` with
  `go fix -inline` where it named gopls code actions (inline, staged signature
  change, removing a middle man), per `go tool fix help inline`: the inliner
  keeps evaluation order and leaves alone a call that needs a function
  literal, such as a callee with `defer`. The prompt, subagent, and gate hooks
  print no gopls note, so a read-only Go question now gets no note at all. The
  READMEs drop the gopls section. `abrun -gopls` is unchanged.

## [1.25.2] - 2026-09-29

- The prompt hook recognizes a review. A prompt that asks for a review or
  audit gets a note naming `go-code-review` alone, to load before the first
  finding; the review corpus prompt ("what is wrong, and the fix") used to
  match the work verb `fix` and got "before the first edit, load `go-code`",
  a condition a review never reaches. The note lists no owners and no card:
  a variant that listed them made Opus 5.5 medium load up to nine skills at
  +37% session cost (n=6). Measured on the review corpus, n=1, against 1.25.1:
  Sonnet 5.5 medium and Opus 5.5 medium loaded `go-code-review` 6/6 in both
  arms, recall 0.93/0.92 and 0.95/0.97, must-fix 30/31 and 31/31 in both,
  cost equal. Opus 5.5 low skips the skill on a review whatever the note
  says (0/12 with the hook silent, 3/12 with the new note), with recall
  about 0.90 either way. `TestPromptRouting` covers a plain review prompt.
- Code comments in the hooks and the eval suite are in English.

## [1.25.1] - 2026-09-29

A content review of all 24 skills, every claim checked against go1.27.1
(go1.26.x where a claim is version-sensitive), `$GOROOT/api`, golangci-lint
2.13.2 with the bundled config, gopls v0.23.0, and runs of the examples and
scripts. The fixes below correct facts and examples. A reference (1.25.0)
versus baseline `abrun` run on all 17 fixtures, n=2, found no regression on
Sonnet 5.5 low or Opus 5.5 low: implement golden 14/14 against 14/14 on Sonnet
and 12/14 against 13/14 on Opus, the gap being `fetch`, which rerun at n=5 gave
1/5 in both arms (every failure the fixture's empty `resp.Status`); refactor
8/8 in every arm at the same line deltas; review recall 0.94 against 0.92 on
Sonnet and 0.90 against 0.91 on Opus, 61/62 must-fix defects in all four;
lint, `go fix` hunks, and session cost within noise. Opus 5.5 low loaded
`go-code-review` in about one review session in ten in either arm, so the
review corpus there barely exercises the skill. Raw reports are kept outside
the repository.

- The bundled `golangci.yml` enables what its comments claimed: `sloglint`
  `static-msg`, and `usetesting` `context-background`/`context-todo`, each
  pinned by a probe in `TestBundledLintConfig`. `depguard` no longer denies
  `google/uuid`, `zap`, or `logrus`, which go-packages and go-style-core keep
  where a package already uses them; the edit hook applied that deny list to
  every repository without a config. `pre-review.sh` lints with the baseline
  when the project has none, so go-code-review no longer counts categories as
  covered that nothing checked. It and `setup-lint.sh` count only
  golangci-lint exit 1 as findings; any other exit is `unavailable`, or exit 2
  in setup-lint, which checks for the binary before writing `.golangci.yml`.
- Examples no longer fail the pack's own lint config: a written file's `Close`
  joins into a named `err` through a deferred `errors.Join` and a file only
  read uses `os.ReadFile` (go-defensive owns the form), tests check `Close` in
  `t.Cleanup`, go-security's TLS example is a bare `tls.Config` (the server
  around it had no `ReadHeaderTimeout`), and the architecture fixture in
  go-code-refactor lints clean and reads the stored total it used to drop.
- Examples that taught a wrong result show the right form in code: the
  go-database transfer uses `UPDATE … RETURNING` and rolls back on a missing
  account instead of committing one side; the go-resilience `retry` stops when
  the wait outlasts the deadline and errors on fewer than one attempt; the
  go-http handler writes nothing for a client that has gone, and the HEAD-as-405
  recipe also registers a method-less pattern so every refused method gets the
  path's own `Allow`; `context.AfterFunc` keeps its `stop`; derived contexts in
  a loop are cancelled per iteration; pooled buffers over 64 KiB are dropped;
  `for range time.Tick` is limited to `main`'s own loop; the SSRF dialer also
  refuses CGNAT (`100.64.0.0/10`, Alibaba Cloud metadata), NAT64, and 6to4;
  archive extraction makes parents with `root.MkdirAll` and refuses link
  entries; benchmarks keep results in a package-level sink and make them
  escape; go-testing's `:memory:` sqlite helper, which gave each pool
  connection an empty database, becomes a real-database harness under an
  `integration` tag (`INTEGRATION.md#real-databases`).
- JSON v2 gaps: `time.Duration` has no default v2 representation, so
  go-defensive scopes `time.Duration` to in-process values and shows integer-
  with-unit and `time.ParseDuration` wire forms, and JSON-V2.md lists it among
  the defaults; a `MarshalWrite` error after the headers is an encode error as
  well as a disconnect, so unvalidated values are marshalled first; nil→`null`
  claims are scoped to v1.
- Facts corrected: `t.TempDir` is Go 1.15 and `time.Since` 1.0 on the idiom
  card; `errorsastype` ships with the 1.27 toolchain; `go mod init` on 1.26.x
  after 1.26.0 writes its own version; `newexpr` rewrites only `&v` pointer
  helpers; a default `go fix` applies Tier 2 `hostport` hunks; `unsafefuncs`
  targets `unsafe.Add`; `go fix` `omitzero` only drops `omitempty`;
  `errors.Is` over `==` is a finding, not free; gopls CLI commands need `-w`
  and `#start-#end` spans; the eg template imports what it uses; `WriteTo`
  keeps `(int64, error)`; old-style doc headings still render as headings;
  a library embedding into a string needs `import _ "embed"`; cgroup-aware
  `GOMAXPROCS` needs a 1.25+ directive; function-type inference in
  composite-literal elements is ungated like conversions; troubleshooting
  commands anchor `-run`, loop `-shuffle` across invocations, skip unbuildable
  bisect steps with exit 125, and check vendor drift without writing `go.mod`.
- Ownership: go-defensive is the single owner of panic versus error (input and
  environment errors never cross a package boundary as a panic; programming
  errors may) and holds the `os.Root` and `crypto/rand` forms go-security
  carried; go-logging carries the `slog.LogValuer` redaction form and
  `slog.DiscardHandler` that other skills route to; go-functions owns
  signature wrapping and value-versus-pointer parameters; go-code routes JSON
  to go-http; request decoding stays in the handler, not a one-call-site
  helper under Declaration Budget rule 4; go-naming settles the `_` global
  prefix and plural layer package names. `docs/RULE_OWNERSHIP.md` records each.
- Scripts: `verify-refactor.sh` 1.3.0 counts pending modernizations (it
  reported `"n/a"` exactly when some existed) and `loc-diff` exits 2 on a path
  other than the recorded baseline's; `check-errors.sh` 1.4.0 pairs a log with
  a return only in the same statement list; `check-interface-compliance.sh`
  1.3.0 reports go-interfaces' own Bad case (`returned_by`), treats
  composite-literal elements and sends as conversions, and exits 2 on a build
  failure; `check-naming.sh` 2.1.0 and `check-docs.sh` 1.3.0 skip what `./...`
  skips and report an unparsable file as `parse_error` with exit 2;
  `get-prefix` flags only parameterless `Get` methods; check-docs skips revive
  `exported`'s method set; `bench-compare.sh` 1.3.0 applies `--limit` to text
  output. `docs/SCRIPT_JSON_CONTRACTS.md` documents each shape.
- `TestIdiomCardDatesEachSymbol` requires every symbol a `(Go 1.NN)` marker
  dates in a CURRENT-GO.md cell to have arrived in that release (`;` separates
  releases), since the card withholds a row newer than the directive; the api
  index now covers `go1.txt` and resolves `t.`/`b.` methods through `testing`.

## [1.25.0] - 2026-09-28

- The routing gate holds the first `.go` edit even when no router is loaded
  and names the entry router: the one the prompt hook picked, else
  `go-code`. The initial load no longer rests on the model's own decision.
- A reminder no longer counts as a load: `loaded` records only successful
  loads, `reminded` only what the gate has named, and a retry without the
  loads is blocked again. So that an unavailable skill cannot cause endless
  retries, every block names the `Read <plugin>/skills/<name>/SKILL.md`
  fallback (Claude Code rejects an unknown Skill name before any hook
  fires), a skill missing from the plugin copy is reported instead of
  required, and a third retry of the same edit with nothing loaded in
  between stops the session with the reason (`continue: false`).
  `GOLANG_SKILLS_ROUTING_GATE=off` switches the gate off.
- A prompt that names a router (`use go-code`, `$go-code-refactor`,
  `/opsx:apply add-auth /go-code`) now selects it and gets the note, where it
  used to silence the note; a skill path such as `skills/go-code/SKILL.md`
  is not a mention. Russian work verbs and code nouns (`почини`,
  `обработчик`, `функция`) join the Ukrainian ones.
- Hook notes and gate messages name skills the way the plugin registers them,
  `golang-skills:go-code`; without `CLAUDE_PLUGIN_ROOT` they stay bare. The
  `name` field in every `SKILL.md` is unchanged.
- [`docs/PROJECT_INSTRUCTIONS.md`](docs/PROJECT_INSTRUCTIONS.md) is a short
  template for a work project's `CLAUDE.md` or `AGENTS.md`: router,
  `go-style-core` and `CURRENT-GO.md`, then only the owners the task needs,
  with what the hooks add in Claude Code and what clients without them rely on.
  Opus 5.5 low smoke sessions (implement `feed` n=1, `fetch` n=8) loaded the
  router, `go-style-core`, and the card before the first applied edit in 9/9,
  with one to four gate blocks each for an owner loaded late; the no-router
  block and the stop never fired. No reference-vs-baseline run measures the
  routing change.
- The first gate block of a session names the gopls route for the file being
  edited: `go_workspace` once and `go_file_context` when those tools are in the
  list, else `command -v gopls` and the CLI. The gate never blocks on gopls,
  since a hook cannot see whether MCP is in this chat. Opus 5.5 low, implement
  `fetch`, `-gopls mcp`, n=3 each: `go_workspace` and `go_file_context` before
  the first applied edit in 3/3 sessions against 0/3 (the reference called only
  `go_diagnostics`), $0.646 against $0.618 a session. Golden passed 0/3
  against 2/3, every failure the fixture's empty `resp.Status` (3/5 without
  the line); raw reports are kept outside the repository.
- `go-code`'s navigation section keeps the policy (which route, when to call
  each MCP tool, literal search and fallback) and drops the gopls mechanics
  that `go-code-refactor/references/GOPLS.md` owns: the CLI command list, the
  rename preview, and the `go_rename_symbol` step. GOPLS.md gains the
  `gopls rename -d` preview, the one line it did not already carry. A Sonnet
  5.5 low refactor A/B (`-gopls mcp`, 4 fixtures, n=2 each) showed no change,
  8/8 golden in both arms and $0.190 against $0.188 a session, but no session
  loaded `go-code` or read GOPLS.md, so it does not exercise the edit.
- `abrun -gopls mcp|cli` gives each claude session a gopls route: the gopls MCP
  server from abrun's own `--mcp-config` (with `--strict-mcp-config`), or Bash
  allowed for `gopls` alone. Without it a session has no shell and no MCP
  server, so `navigation: grep-only` is the only honest report. Each result
  records the route and its `gopls_calls`.

## [1.24.2] - 2026-09-27

- Go navigation checks whether gopls MCP tools are present in the current chat,
  gives a concrete LSP/CLI fallback, and reports the route used. Text search
  no longer stands in for semantic references, callers, implementations, or
  renames when gopls is unavailable.
- The READMEs document installing the skills with bun: `bunx skills add
  h0rn3t/golang-skills --all`, or a one-time `bun add -g skills` and then
  `skills add`. Updating is `bunx skills update` with the same scope flags as
  the npx route.
- The READMEs document installing into Codex, GitHub Copilot, and Cursor with
  the same CLI: `bunx skills add h0rn3t/golang-skills --all -a codex -a
  github-copilot -a cursor -g`, with the `--agent` name and global path for
  each. The Updating section covers all three instead of Codex alone, and its
  examples use `bunx` too. Both README.md and README.uk.md.

## [1.24.1] - 2026-09-27

- Go navigation now starts with available gopls MCP tools for unknown symbols,
  file context, package APIs, and change impact. The prompt hook also gives
  read-only Go questions a once-per-agent navigation hint; `rg` remains the
  direct route for literal text. gopls diagnostics run after coherent edit
  batches, and the reference documents MCP limits and call examples.

- Copy-depth guidance now distinguishes a bug fix that satisfies an existing
  ownership contract from a style-only change, and conditions `Clone`
  modernization on preserving observable aliasing.

## [1.24.0] - 2026-09-27

- Go navigation now uses `rg` for cheap textual discovery, available gopls
  MCP, LSP, or CLI tools for semantic relationships, and targeted file reads
  for context. The `go-gopls-first.sh` hook and its test were removed because
  they blocked `rg` discovery before gopls navigation.

- The Agent Skills frontmatter check moved from `npx agentskills-validate@1.0.1`
  into the Go suite: `agentSkillsSpecErrors` in `evals/eval_test.go` ports
  its checks (allowed fields, name shape, length, and directory match,
  description and compatibility length) and runs from `TestStructure` on
  every skill. It parses no YAML, so it accepts only this repository's shape,
  one plain scalar per line, and rejects a value YAML would misread: a leading
  indicator, `: `, ` #`, or a bare number or keyword. `TestAgentSkillsSpecErrors`
  pins one case per check. The validate job no longer sets up Node.js; the
  opt-in `evals` job still does, to install Claude Code.

- `go-code-refactor/references/GOPLS.md` now covers navigation, not only
  rename and extract: a question-to-tool table for the `LSP` tool and the
  gopls MCP server (references, callers, implementations, declarations,
  hover, file outline, diagnostics), how to tell which route is wired, and
  that the routes stack rather than compete. `go-troubleshooting` (callers of
  the diverging function), `go-code-review` (callers outside the diff), and
  `go-interfaces` (implementations before a method-set change) route to it in
  one line each. A gotcha records that the MCP server roots at its start
  directory, so a repository with `go.mod` in a subdirectory gets "not a Go
  workspace" and falls back to the `LSP` tool. The READMEs gain a gopls section: the binary, the
  `gopls-lsp@claude-plugins-official` plugin, and the MCP server for Claude
  Code and Codex; golang-skills declares no dependency on any of them. The
  wording is unmeasured: abrun runs with `--restricted` and without the
  `LSP` tool, so no arm can call gopls yet.

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
