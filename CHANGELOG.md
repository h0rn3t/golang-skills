# Changelog

All notable changes to this repository are documented here.

## [Unreleased]

### Added

- A shared JSON v2 boundary reference with bounded single-document decoding,
  compatibility options, nil collections, streaming/newline differences, and
  golden-test limits; routed from HTTP, package, defensive, and testing skills.
- An `errors.AsType` branch example that preserves the original error, with
  executable checks for second-branch matching and JSON input boundaries.
- `nolintlint` requiring named linters and explanations, and `usestdlibvars`
  for HTTP methods/status codes in the maintained lint configuration.

### Changed

- New-module guidance checks the actual `go` directive and local/CI toolchains
  after `go mod init`, without assuming its default or upgrading existing code.
- Modernization guidance distinguishes installed analyzer sets and calls out
  post-fix compilation, comment retention, and behavior-changing slice rewrites.

## [1.3.1] - 2026-09-09

### Fixed

- Go 1.27 conformance pass over the skills, verified against an installed
  go1.27.1 rather than from memory. Every inline `(Go 1.xx)` claim already
  matched `$(go env GOROOT)/api/go1.NN.txt`; the guidance around them did not.
  `go-troubleshooting` still taught count-comparison as the only way to find a
  goroutine leak, so it now leads with the `goroutineleak` profile and states
  what it cannot prove, and its timer row no longer points at a `go` directive
  that stopped mattering when 1.27 removed `asynctimerchan`. `go-testing`
  contradicted its own table by demoing `httptest.NewServer` in a `*testing.T`
  test; the example moves to `NewTestServer` and both files now carry the
  constraint that makes the swap safe — the in-memory server answers only
  `srv.Client()`, and its `srv.URL` of `http://example.com` sends a
  self-built client to the real example.com. `go-http` and `go-resilience`
  drop the manual read-to-EOF for connection reuse, which 1.27 does on `Close`
  within 256 KiB and 50 ms. `go-generics` reshapes the `maphash.Hasher`
  example: constrained to `any` rather than `comparable`, since a Hasher buys
  what the built-in map cannot do, and with the hash method that shows how the
  seam fits together.
- `plugin.json` and `marketplace.json` still advertised "each under 500 lines"
  after the cap dropped to 400; the longest `SKILL.md` is 362 lines.

### Added

- `go-linting` lists the modernizers 1.27 added — `atomictypes`,
  `slicesbackward`, `unsafefuncs` — plus `reflecttypefor`, and records that
  `waitgroup` became `waitgroupgo` and `fmtappendf` is gone, so a pinned
  command naming either now fails. Smaller 1.27 items reach the skill that owns
  them: goroutine labels in traceback headers and why the panicking goroutine
  usually lacks them (`go-troubleshooting`), `SystemCertPool` honoring
  `SSL_CERT_FILE` on Windows and macOS (`go-security`), `sql.ConvertAssign` and
  `driver.RowsColumnScanner` (`go-database`), require-block consolidation in
  `go mod tidy` (`go-packages`), and `go doc -ex` / `go doc pkg@version`
  (`go-documentation`).

## [1.3.0] - 2026-09-09

### Added

- Give the `go-code-refactor` concision gate a counter instead of an
  instruction: `verify-refactor.sh` gains `loc-baseline` and `loc-diff`, which
  record both production LOC counts and each file's digest before the first
  edit, recount afterwards, and return the verdict as the exit status. The
  count is token-aware, because a nonblank-line count reads a multiline string
  as code it is not; the convention is the harness's own and is verified
  against it on comments, raw strings, a missing trailing newline, nested
  directories, and added and deleted production files. The gate and the report
  template now name those commands. Screening on 5 fixtures x 2 arms with
  GPT-5.6-Luna medium: build and independent golden 5/5 in both arms, the
  counter runs in 4 of the 4 sessions that read the new text against none
  before, both LOC gates pass 4/5 against 3/5, and the means move from +0.2
  physical and +0.2 code to -0.4 and -1.2. Two of those four recorded the
  baseline after their first edit and every input has n=1, so this is one
  screening and not a reliability claim.
- Document how to install a pinned version from a release tag, in both READMEs.

- Add Go 1.27 guidance for promoted fields in struct literals, copying retained
  substrings only when justified by profiling, and merging/filtering maps in
  place. Add executable example checks and three implementation quality cases;
  record the JetBrains source review without claiming a measured model benefit.

### Fixed

- Keep Claude evaluation plugin resources inside each scratch project so
  `--restricted` permits reference reads without exposing golden fixtures.
  Apply the compact inline `go-http` layout with explicit HEAD-contract,
  empty-array, and whole-response-size rules; add executable response-limit
  checks and strengthen the gateway golden tests. The combined text has no
  established cost advantage; retain the routing and compression experiments.

- Correct Go snippet API calls and generics version notes; demonstrate Go 1.27
  inference in conversions, use stdlib `uuid`, `errors.AsType`, and `wg.Go`,
  and remove obsolete benchmark sink advice. Close files and transactions on
  failure paths, preserve argument evaluation and Git revision semantics, and
  keep raw request data out of logging examples. Add executable regressions
  for generic snippets and transaction finalization.
- Distinguish wrapping an existing error from introducing a sentinel or custom
  type; remove the duplicate selection table and clarify that `%v` omits the
  error chain without redacting its text. Tighten helper-extraction criteria,
  make shared HTTP structures and error mapping conditional on actual reuse,
  and apply the v1 requirement-derived acceptance checks in `go-code`.
  These instruction changes do not establish a measured model improvement.
- Make the HTTP JSON example reject trailing data before domain side effects.
- Preserve final response status and ResponseController operations in the
  logging middleware example; use standard `slog.NewMultiHandler` for fan-out.
- Run model-authored tests separately before hidden golden tests in `abrun`,
  exclude their failures from valid results, and retain successful source with
  `-keep`. Add executable example and runner regressions plus quality cases.

- Add `-repair` to `abrun`: after the session it measures production lines,
  replays the golden against a throwaway copy of the module, and if either the
  line gate or the golden failed it returns the exact numbers and the assertion
  text and grants one repair turn. The probe never puts the golden in the tree
  the model can read, and the failure reaches the model with file positions
  stripped, both pinned by tests. Records `repair_fired`, `pre_repair`,
  `pre_repair_golden` and the generated `repair_feedback`.
- Record stage 3 of the Pocock concision plan: the loop moved the line gate from
  8/15 to 15/15, the median from +0 to −3 and the worst case from +17 to +0
  (effect −3.80 lines, within-input randomization p = 0.01572), and repaired the
  `nil → null` wire regression the 5×5 run shipped as a behavior failure. Four
  of five criteria pass; the fifth does not, because two runs met the gate by
  deleting the doc comment that justifies the server's timeouts. Expanding to
  n=5 is blocked until the gate counts code separately from comments.
- Record stage 2 of the Pocock concision plan: exact numeric feedback against
  generic repair on the ten problem outputs of the 5×5 run. Feedback repaired
  10/10 against 1/10 (paired sign test p = 0.00391), added no golden failure
  where generic repair added one, and fixed the `nil → null` wire regression
  generic repair left in place. The finding that transfers is a definition, not
  a prompt: one fixture holds 138 physical lines and 101 non-blank non-comment
  lines, and generic repair reported the second number and stopped, so any
  feedback about lines has to name the counting convention.
- Record stage 1 of the Pocock concision plan: the candidate refactor prompt that
  permits an empty diff, measured against the current prompt on five fixed
  gateway inputs with the plugin held constant. Golden 5/5 on both arms, line
  gate 4/5 against 5/5, and the first genuine empty diff in the series. The
  line difference between prompts is not established (paired p = 0.6250), so no
  prompt or skill text changes on this evidence.

### Changed

- Separate a harness collision from a behavior regression in `abrun`. A golden
  overlay that fails to build because the model's production code declares one
  of its names is now `harness_failure`, prints as `HRN`, and stays out of every
  mean; an overlay that compiled and failed an assertion is `behavior_failure`.
  A broken public contract (`undefined: NewServer`) stays the model's. Rename
  every colliding golden helper (`serve` → `goldenServe`, `members`,
  `assertJSONEqual`, `fakeStore`, `errTransport`) and add
  `TestGoldenHelpersAreCollisionResistant` so a new one cannot reappear.
- Record `line_gate_pass`, `empty_diff`, `reported_counts`, `commands` and
  `trace_path` per run, and retain the raw session transcript as `trace.jsonl`
  under `-keep`, so a stated line count is checkable against the commands the
  session actually ran. Pin the production LOC definition in `metrics.Lines` and
  `evals/ab/README.md`. Closes stage 0 of the Pocock concision plan.
- Lower the `SKILL.md` cap from the Agent Skills spec ceiling of 500 lines to
  400 in `evals/eval_test.go`, `docs/SKILL_AUTHORING_TEMPLATE.md`, both READMEs,
  and the release-watch prompt. No skill changes: the largest is
  `go-code-refactor` at 328 lines, and `abrun` splices variant arms into that
  same file, so the working maximum is ~354. The cap bounds growth; shortening a
  skill still needs measured evidence, not a smaller number.

## [1.1.0] - 2026-09-08

### go-code-refactor

- Record the supplied Sonnet 5 refactor control: 40/40 valid, mean production
  difference −3.20 lines, new functions 34 → 20, types 6 → 10, cost 3.28x.
  `store` is an exploratory −8.2-line signal; `report` grows 2.8 more. Preserve
  raw JSON and add English/Ukrainian reports and README comparisons. This is
  plugin-versus-control evidence, not validation of an individual rule change
  ([report](docs/evidence/2026-09-08-go-refactor-control-sonnet-5.md)).

- Add «Remove Duplication to the End» to `SKILL.md` and a matching fold in
  `references/PLAYBOOK.md` §0: when branches differ only in the constants
  they carry, the step is done when each literal, each selection over the
  key, and each condition ladder appears once; the shape follows the final
  code, and a table is not to be serviced. Measured as the `selection-once`
  A/B variant under `gpt-5.6-luna`: on copilot, where baseline stopped at a
  `switch` in 4/10 `pricing` sessions, −11.6 lines vs baseline (95% CI
  −18.9 … −4.3, n=10) with `report` unchanged at −1; on opencode, where
  baseline had no gap, +0.1 (−7.7 … +7.9). On codex, measured after the
  port with the new tree as baseline: −12.4 lines vs no-skill (95% CI
  −18.0 … −6.8, n=10), data table in 9/10 `pricing` sessions, where the
  2026-09-07 tree had trailed the control by +11.4 (n=5); `report` at −1 in
  8/10 and, in a separate `report`-only check, 9/10 (−14.6 vs no-skill, CI
  −21.5 … −7.7); the outliers are helper extraction, a tail that predates
  the change and does not involve the new criterion.
- Add a percent summary of the measured effect to `README.md` and
  `README.uk.md`, and a block on the three-runner run and the change it led to.

### go-code

- Fix the `description` frontmatter, which an unquoted colon-space made
  unparseable. Copilot and any other YAML-parsing loader dropped the skill
  silently, so the router did not load at all.
- Re-run the implementation corpus on `gpt-5.6-luna` (`-effort medium`,
  `-runner codex`) after the rewrite, same prompt, seed and fixtures: 40/40
  valid, golden 20/20 in both arms as before, corpus size −0.95 lines with every
  interval including zero. Routing is what the rewrite moved — `go-http` reached
  in 4 of 5 `gateway` sessions against 2 of 5, `go-error-handling` 15 of 20
  against 9, `go-data-structures` 15 against 10, `go-linting` 16 against 8,
  `go-style-core` 17 against 11 — with no change in the code the sessions
  produced, because the control arm already passes every trap. Firing rate and
  measured outcome are separate results
  ([analysis](docs/evidence/2026-09-08-go-implement-control-gpt-5.6-luna-medium.md),
  [українською](docs/evidence/2026-09-08-go-implement-control-gpt-5.6-luna-medium.uk.md)).

### Evals harness

- Make the `abrun` arm pre-check reject a skill that fails to load. It compared
  counts against zero, so one unparseable skill in a 24-skill tree passed it and
  80 copilot sessions measured a baseline arm with no router; it now compares the
  set the CLI reports against the arm's own `skills/` directory and names what is
  missing. The copilot listing's stderr is captured and quoted into the error,
  because `copilot skill list` reports a skill it could not parse there while
  still exiting zero.
- Add a frontmatter guard that runs for every runner, claude included, before any
  CLI is invoked: a `SKILL.md` whose frontmatter carries an unquoted colon-space
  fails the run instead of silently shrinking the arm.

## [1.0.0] - 2026-09-07

### Skill descriptions

- Shorten 18 of 24 skill descriptions, from 9,978 to 6,388 characters (−36.0%).
  Trigger coverage, skill bodies, references, and permissions are unchanged;
  `go-code-refactor`, `go-testing`, `go-logging`, `go-http`, `go-security`, and
  `go-troubleshooting` keep their original wording. The exact-description test
  goldens in `evals/eval_test.go` move with them.
- Measure the change on the Codex host, where descriptions are the only input
  to the routing decision: all 105 trigger cases on both description sets under
  `gpt-5.6-luna` (medium), 210 sessions, no errors. The rendered skills catalog
  falls 13,930 → 10,338 characters, cases passed move 88/105 → 84/105 and
  expected-skill reads 98/110 → 94/110 — McNemar exact p = 0.42, with all 18
  negative controls holding case for case. Prompt size is established;
  behavioral equivalence is not, and three of the fourteen flips turn on a
  description that was never edited. Evidence, raw report, and harness are in
  `docs/evidence/2026-09-07-description-compression-codex-luna-medium.md`.

### Multi-runner evals

- Add `codex` and `copilot` runners to `evals/cmd/abrun` alongside `claude` and
  `opencode`, each with its own arm isolation: a private `HOME` per arm, and for
  Codex a `codex debug prompt-input` precondition check that the arm offers the
  skills it is supposed to and nothing the host installed. Codex has no skill
  tool, so a skill counts as fired when the model reads its `SKILL.md`, taken
  from the shell command and never from its output. Add `-effort` for the two
  runners whose CLI can set a reasoning level; it is rejected elsewhere rather
  than ignored.
- Add the `implement` corpus (`evals/ab/_implement`) with the `catalog`, `feed`,
  `gateway`, and `ledger` fixtures and their hidden golden tests. Where the
  refactor corpus scores structure removed from working code, this one scores
  whether documented-but-unimplemented declarations work at all.
- Publish control runs for Opus 5, GPT-5.6-Luna, and MiniMax M3 under
  `docs/evidence/`, including the runs where the arms do not separate because
  the model never falls into the fixture's trap.

### Refactoring depth

- Raise the `SKILL.md` ceiling from 225 to 500 lines, the limit the Agent Skills
  spec sets on a skill body. References still cap at 300 lines each.
- Add four `go-code-refactor` references: `CATALOG.md` (smell, transform, tool,
  and risk tier for moves that cross a function, type, or package boundary),
  `SAFETY-NET.md` (coverage tiers over the blast radius, characterization tests,
  seams), `MECHANICAL.md` (`gofmt -r`, `eg`, `gopatch`, `go/analysis` fixers for
  a recurring edit), and `STRUCTURAL.md` (type-alias gradual repair, import-cycle
  strategies, deprecate-before-delete). The pack now contains 65 references.
- Add two `go-code-refactor` sections: **When Not to Refactor**, which separates
  purposeless churn from safety prerequisites that only change the sequence,
  and **Risk Tiers**, which sets what must be true before a step and pairs with
  the coverage tiers.
- Record the new rule areas in `docs/RULE_OWNERSHIP.md` and attribute the topic
  selection to `samber/cc-skills-golang` (MIT) in `THIRD_PARTY_NOTICES.md`.

### Adopted from samber/cc-skills-golang

- Add a deprecated-API replacement table to
  `go-code-refactor/references/MODERNIZATION.md`, linked from `COMPATIBILITY.md`
  so it also ships with skill-only installations. Rows carry the deprecation
  version and risk conditions; crypto migrations require API and failure-path
  checks. Tier definitions and the Go 1.27 removed-`GODEBUG` checklist live in
  the same catalog.
- Add a safety-pitfall table to `go-defensive`, an observability definition of
  done to `go-logging` (advisory, scoped to a service in the stack the project
  already runs), benchmark discipline to `go-performance`, tool-directive and
  audit guidance to `go-packages`, a CI pipeline checklist to `go-linting`, and
  routing boundaries to `go-code`. Mid-refactor test scope in `GOPLS.md` is
  affected packages including consumers, matching the `go-linting` gate; CI
  flags and the `govulncheck` trigger contract have one owner in `go-linting`.
- Widen `go-logging`'s description to metrics and trace correlation, and
  `go-defensive`'s to the silent-correctness traps its new table covers, so
  both rule sets are reachable by their own triggers rather than only through
  the `go-code` router.
- Record the new rule areas in `docs/RULE_OWNERSHIP.md` and attribute the topic
  selection to Samuel Berthe (`samber`) in `THIRD_PARTY_NOTICES.md`. Add six
  trigger evals (modernization priority, benchmark evidence, tool directives,
  CI shape, observability routing, safety pitfalls) and three quality evals
  covering the full-slice-expression trap, `slog` context handling, and Actions
  SHA pinning. `TestManifestCounts` now pins both eval counts to the READMEs.
- Add a **Scope Exceptions** section to `docs/RULE_OWNERSHIP.md` recording why
  `.github/workflows/validate-skills.yml` does not follow the new `go-linting`
  CI checklist: no version matrix, `govulncheck`, or tidy-drift check for a
  zero-dependency pack, and unpinned action tags, a missing `permissions:`
  block, and tests without `-race -shuffle=on` accepted as known gaps. The
  workflow carries a header comment pointing at that record.

### Evals

- Add `evals/cmd/abrun`: runs one refactoring prompt against fixtures under the
  current plugin, an optional complete reference checkout, and wording variants.
  It reports recursive structural deltas only for builds that pass an external
  golden test the model never sees. The `no-skill` control decides whether a
  fixture measures anything; the reference arm enables before/after comparison.
- Add `evals/ab` with four fixtures, their golden characterization tests, and
  an evidence contract requiring raw JSON, exact model, seed, and both compared
  plugin roots before a result is published. Record the first compliant Opus 5
  control: 40/40 valid runs, 4.15 fewer lines per run overall, and 49.7% less
  growth on the `report` over-engineering trap with the skill.

### Toolchain

- Bump the pinned golangci-lint from v2.13.1 to v2.13.2 in
  `.github/workflows/validate-skills.yml`, `skills/go-linting/SKILL.md`, and
  `docs/RELEASE_CHECKLIST.md`. The release carries dependency bumps and a cache
  fix only — no new linters and no config schema change, so
  `assets/golangci.yml` is unchanged and `golangci-lint config verify` passes
  against it on 2.13.2.

## [0.9.0] - 2026-09-06

### Skill consolidation

- Reduce the pack from 27 to 24 skills: merge `go-functional-options` into
  `go-functions`, and `go-control-flow` plus `go-declarations` into
  `go-style-core`. Update explicit invocations to those owner names; manually
  copied installations must remove the retired directories.
- Keep the style entrypoint short with conditional syntax references; combine
  overlapping initialization/literal/struct guides into one reference. The
  pack now contains 61 references.
- Narrow function API activation, preserve config-struct and functional-options
  choices, and update routing, ownership, and eval expectations. Add focused
  translation, shadowing, config-convention, and iterator-stop scenarios.
  Probe results and limitations are in `docs/SKILL_CONSOLIDATION_REVIEW.md`.

### Cross-model compatibility

- Align shared Go instructions with GPT-6 Astra guidance while retaining the
  Claude plugin: preserve explicit requirements, avoid redundant approval and
  verification loops, resolve installed resources, and honor host delegation.
- Scope `go-verify` to requested checks; distinguish incomplete verification
  from findings, preserve race coverage, and account for staged dependencies.
- Add six quality scenarios and document the GPT-6 application probes and the
  remaining Claude-only automated runner in `docs/CROSS_MODEL_REVIEW.md`.

### Added

- `go-resilience`: replay safety and retry budgets, idempotency across replicas,
  bounded load admission, circuit recovery and graceful degradation. Includes
  two routed references and six quality plus six trigger scenarios. Behavioral
  evidence and limitations are in `docs/GO_RESILIENCE_REVIEW.md`.

- `go-troubleshooting` references for ticket investigation and data-flow tracing,
  with deployed-version/config checks, working-case comparison, testable
  hypotheses, and confirmed/probable/unresolved findings.
- Six troubleshooting quality scenarios and six trigger cases, covering tenant
  scope, stage drift, incomplete evidence, mapping loss, scoped regression proof,
  live hang capture, and ticket-text tasks that must not trigger debugging.
  GPT-6/Opus probe evidence and limits are in `docs/GO_TROUBLESHOOTING_REVIEW.md`.

- `go-security`: the trust-boundary threat model the repository had no owner
  for — follow untrusted data to its sink (SQL, `os/exec`, `html/template`,
  file path, outbound URL, log line, error response) and apply the stdlib
  defense once at the boundary. Covers SSRF checks with `net/netip`,
  constant-time comparison, argon2id/`crypto/pbkdf2` for passwords, AEAD-only
  encryption, TLS and cookie settings, redaction, and a data-flow review mode.
  `os.Root` and `crypto/rand` stay owned by `go-defensive`; `gosec` in the
  baseline lint config is now listed as the enforcing linter. `go-code-review`,
  `go-http`, `go-defensive`, `go-linting`, and the `go-code` router route to it.
- `go-troubleshooting`: root-cause method for the cases where the cause is
  unknown — reproduce, capture, read, hypothesize, confirm, fix once, pin with a
  test. `references/DIAGNOSTIC-TOOLS.md` is the command reference for
  `GOTRACEBACK`/`GODEBUG`, goroutine dumps, `pprof` capture and reading,
  `go tool trace` and the Go 1.25 `FlightRecorder`, the race detector, Delve,
  and the test flags that turn a flake into a reproduction rate.
  `references/SYMPTOM-CATALOG.md` maps each runtime panic message, hang shape,
  leak signature, wrong-result pattern, and CI-only failure to its mechanisms,
  the command that confirms each, and the skill that owns the fix.
  `go-performance` and `go-concurrency` route to it for the "why" before the
  "how".
- `go-code-refactor/references/GOPLS.md`: semantic references and safe rename
  through gopls (MCP server, native LSP tool, or CLI) instead of grep — the
  rename that would un-implement an interface is refused, and diagnostics run
  after every edit before the next transformation. Step 4 of the workflow now
  sends renames and extractions there.
- `go-code` routing table gains an "also load" column naming the skill a row
  almost always drags in (concurrency ↔ context, HTTP and SQL → error handling
  and security, performance → troubleshooting when the cause is unknown), so
  the pair loads in one pass instead of after the first draft exposes the gap.
- `.github/workflows/go-release-watch.yml`: a monthly job that compares the
  latest Go release and golangci-lint release against the versions this
  repository pins (`COMPATIBILITY.md`, `validate-skills.yml`, the `go-linting`
  baseline) and, only when one is behind, asks Claude Code to open a PR
  updating `COMPATIBILITY.md`, `MODERNIZATION.md`, the `go fix` table, and
  `TestGoVersionBaseline`. Nothing runs, and no tokens are spent, while the
  versions match.
- Trigger evals for `go-security` (path traversal, command injection, password
  storage, a Ukrainian SSRF prompt) and `go-troubleshooting` (RSS growth,
  a pasted race report, a Ukrainian hang prompt, a pasted panic trace, and a
  known-cause negative control that must route to `go-performance` instead).

- `go-http`: handler shape, Go 1.22 `ServeMux` method patterns, bounded request
  bodies, error-to-status mapping, middleware, `http.Server` timeouts and
  graceful shutdown, and client rules (per-dependency client with a timeout,
  `NewRequestWithContext`, body closed on every path). `WEB-SERVER.md` moved
  here from `go-code-review`, which now routes to it.
- `go-database`: `database/sql` first, context on every query, `*sql.DB` as a
  pool with limits, the rows loop with `rows.Err()`, transactions with a
  deferred `Rollback` and a checked `Commit`, placeholders over string-built
  SQL, queries-in-loops and keyset pagination, `sql.Null[T]`, embedded
  migrations, and ORM rules for repositories that already have one.
  `references/SQL-PATTERNS.md` carries the full code.
- `go-code` routes HTTP, SQL, wire-format, and CLI tasks — the most common
  service work had no row before.
- The `go-linting` baseline `.golangci.yml` now enforces what the skills teach:
  `depguard` (deny `pkg/errors`, `logrus`, `zap`, `x/exp/slices`, `x/exp/maps`,
  `google/uuid`), `errname`, `errorlint` (with `errorf` off — `%v` at a
  boundary is deliberate), `exhaustive`, `godot`, `noctx`, `perfsprint`,
  `prealloc`, `rowserrcheck`, `sloglint` (`snake_case` keys), `sqlclosecheck`,
  `usetesting`. Verified with golangci-lint 2.13.1.
- `agents/go-verify.md`: a bundled subagent that runs the gate and returns only
  failures. `hooks/go-vet-on-edit.sh`: a PostToolUse hook that runs `gofmt -l`
  and `go vet` on the package of every edited `.go` file and hands findings
  back via exit 2. Both install with the Claude Code plugin.
- `evals/cmd/evalrun`: a headless runner for the trigger and quality evals
  through `claude -p --plugin-dir --restricted`, each prompt in a scratch
  directory with the tool set cut to `Skill`, with a model-graded checklist for
  quality evals and a JSON report. The `Validate Skills` workflow gains an opt-in
  `evals` job behind the `run_evals` dispatch input.
- Trigger evals for `go-http` and `go-database` (including a Ukrainian prompt
  and a CSV negative control) and quality evals 18 (HTTP handler) and 19
  (repository with a transaction).
- `go-style-core` owns the house-style rule: the repository's `.golangci.yml`,
  `CONTRIBUTING.md`, and neighboring code outrank the guide. `go-code`,
  `go-code-refactor`, `go-testing`, and `go-naming` route to it.
- Gaps filled: `errors.Join` (go-error-handling); `context.WithCancelCause`,
  `context.Cause`, `context.AfterFunc`, `context.WithoutCancel` (go-context);
  `errgroup.SetLimit` and a goroutine-or-not decision tree (go-concurrency);
  writing `iter.Seq` producers (go-control-flow); `t.Parallel()` in the testing
  quick reference; `//go:build` and `//go:embed` (go-packages).

- `go-code-refactor/references/OVER-ENGINEERING.md` now owns the full restraint
  ladder: the seven rungs (does it need to exist → already in this codebase →
  stdlib → language/toolchain feature → module already in `go.mod` → one line →
  the minimum that works), the rule that it runs *after* the code and flow are
  read rather than instead, and the never-on-the-chopping-block list (trust
  boundaries, data-loss handling, security, accessibility). `go-code` and
  `go-code-refactor` route to it instead of carrying their own shorter,
  differently ordered ladders.
- `TestRestraintLadder` pins that ladder: seven rungs present and in order, the
  read-first rule and the never-cut list intact, the rung text living in
  exactly one file (a second copy anywhere under `skills/` fails), and both
  `go-code` and `go-code-refactor` routing to the owner. `quality_evals` gains
  id 17, the behavioral half: a prompt dangling a one-implementation interface,
  a factory, and a hand-rolled `contains` helper, asserting the ladder skips
  them, reaches for `slices.Contains`, and still keeps the trust-boundary
  check.

- `go-code`: a routing skill for Go tasks that span several topics, and the
  modifier form other workflows can carry (`/opsx:apply /go-code`) — when
  passed as an argument it is explicitly not a change name or a file path.
  Loads the restraint rules on every invocation (the
  `go-code-refactor/references/OVER-ENGINEERING.md` cut tags and the normative
  stdlib-before-dependency ladder in `go-packages`), applying them to code
  being written rather than only to code being audited, then routes by topic
  and closes with the `go-linting` gate. Owns no rules of its own; every rule
  stays with its existing owner. README, `plugin.json`, and `marketplace.json`
  now count 22 skills.

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

- Replace `go-http`'s blanket ban on 4xx retries with routing to `go-resilience`:
  documented 429 handling depends on replay safety, Retry-After and budgets.
  Link context, concurrency, database and troubleshooting guidance to the owner.

- Troubleshooting no longer treats a blocked stack as proof of a leak, a panic
  site as its root cause, or a closed-channel panic as necessarily a data race.
  Prefer a nonterminating admin dump to SIGQUIT, explain capture overhead, and
  distinguish investigation-only scope from authorized correction/verification.

- `go-code` had no `## Resource Routing` section and no golden description, so
  `TestSkillArchitecture` and `TestFrontmatterDescriptionsInvariant` failed on
  `main`. Its description now starts with "Use when", as `TestStructure`
  requires.
- `README.uk.md` said 51 reference files in "Як це працює" while the rest of
  the repository said 52.
- `check-naming.sh` exited 0 for a nonexistent path: its `exit 2` ran inside a
  process substitution, so the caller continued with an empty file list and
  reported a clean scan. The target is now validated in the main shell.

### Changed

- Aligned the procedural skills with Anthropic's prompting guide for Claude
  Opus 5. `go-style-core` now owns "How Much To Say" — narration cadence,
  written-output length, and when a subagent is justified — and `go-code`,
  `go-code-review`, `go-code-refactor`, and `go-troubleshooting` route to it,
  so a skill invoked without the router still carries the rule. The
  verification gate is stated once (`go-linting`) and run once, at the end;
  the `go-error-handling` and `go-testing` validation notes no longer
  re-invoke it. `agents/go-verify.md` triggers only on an explicit request,
  never as a post-edit self-check — its old description invited exactly the
  redundant verification the guide says to remove. `go-code-review` gains a
  Correctness section ahead of the style rows, and `go-security` review mode
  reports unreachable issues as `not reachable` instead of skipping them,
  since the model follows "report less" literally.
- Less code and readability are now the stated priority of the three skills
  that write, reshape, or judge Go. `go-code-refactor` owns the rule — "Delete
  Before You Restructure": line count is the instrument, readability the goal,
  and the order of work is delete, then shorten, then restructure; a step that
  adds net lines needs a reason in the report, whose summary now ends with
  `net: -<N> lines`. `go-code` states the priority before any routing decision
  and reads the final diff's net line count alongside the per-entity ladder
  climb. `go-code-review` opens its checklist with a "Less Code" section and
  subtracts before it styles — unneeded growth is a Should Fix, not a nit; the
  review template carries a net-lines line and a cut-tag example. The `gofmt`
  checklist row is gone: `pre-review.sh` already runs it, and a human finding
  for a tool's job was the checklist's own bit of bloat. `docs/RULE_OWNERSHIP.md`
  gains the row, `TestDeleteFirstRule` pins the owner and both routes, and a
  quality eval checks that a review leads with what can stop existing.
- `go-code-refactor/references/OVER-ENGINEERING.md` carries the rest of what
  `ponytail` says about writing less code and the Go skills did not: a bug fix
  lands once, in the function every caller routes through; two options on one
  rung go to the one correct on edge cases; a request bigger than its need
  ships the rung that holds and questions the rest in the same reply, never
  stalling on an answer that has a default; a write is reported code first
  with at most three lines after it (`skipped: <X>, add when <Y>`); the user's
  insistence ends the argument; non-trivial logic leaves one runnable check,
  trivial one-liners none, and that one check is never a cut; the minimum that
  works lives in the fewest files. The `stdlib:` tag now covers language and
  toolchain features (`go:embed`, struct tags, `synctest`). Ponytail's
  intensity levels are deliberately not carried: the Go skills always run at
  `full`, and the audit lane is the `ultra`. `go-code` points to the write
  rules next to the ladder; `TestRestraintLadder` pins them in the owner.
- `OVER-ENGINEERING.md` gains "Reach For What Go Ships": rungs 3 and 4 as a
  four-part table (language; collections and strings; errors, concurrency,
  context; I/O, HTTP, tests, logging) — instead of this hand-written block,
  this language or stdlib feature, with the minimum Go version and the owner
  skill, every version checked against `api/go1.NN.txt`. The hunt list's
  stdlib table now points there; `go-style-core` "Write Current Go" routes to
  it for what `go fix` has no modernizer for; `MODERNIZATION.md` names it as
  the checklist for new code, keeps the swap caveats for existing code, and
  dates `errors.Join` to Go 1.20. `COMPATIBILITY.md` gains the 1.22, 1.23,
  and 1.24 APIs the table recommends and the `tool` directive. The file now
  carries a provenance header and moves the per-entity ladder rule in from
  `go-code`, which keeps pointers. `TestRestraintLadder` pins the table and
  the `go-style-core` route; quality eval 21 checks that a write under a
  `go 1.24` directive reaches for the stdlib, respects the directive, and
  declines a `util` package.
- `check-naming.sh` is now a wrapper around `check-naming-ast.go`, matching the
  other findings scripts: a SCREAMING_SNAKE word in a string or a comment is no
  longer a violation, and the JSON contract is unchanged.
- `setup-lint.sh` emits `assets/golangci.yml` verbatim instead of carrying a
  second copy of the config.
- `go-testing`: "No assertion libraries" is now project policy — none in a
  repository without one; match `testify` and enable `testifylint` where it is
  already used.
- `go-naming`: the `_` prefix on unexported globals is labelled as Uber-only
  (Google style omits it) and follows the repository.
- `go-concurrency`: "Default to channels" replaced by a decision tree whose
  first question is whether a goroutine is needed at all.
- `docs/SCRIPT_JSON_CONTRACTS.md` records why a cross-skill `go/analysis`
  multichecker was rejected: each skill directory must stay installable alone.
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
