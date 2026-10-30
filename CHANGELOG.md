# Changelog

All notable changes to this repository are documented here.

## [Unreleased]

abrun implement (account, catalog, feed, gateway, ledger, roster; medium,
n=3, 1.46.1 / this tree): Haiku 5.5 sessions reading `LADDER.md` 9/18 / 0/18,
golden 100% / 100%, $0.0258 / $0.0244 per run; Sonnet 5.5 golden 100% /
100%, $0.255 / $0.250. What "Ship, Then Question" changes in reports is
unmeasured: `skipped:` lines appeared in 1 of 72 reports.

- **The ladder hook prints all of `LADDER.md` below its header**, "Ship,
  Then Question" included, and says the file needs no Read. Given the ladder
  section alone, Haiku 5.5 read the file again in 261 of 678 Go sessions
  (local transcripts, 2026-10-23..29), always before the first edit; Sonnet
  5.5 and Opus 5.5 never did, so they never saw "Ship, Then Question". The
  hook prints 5.2K characters per Go session and subagent instead of 3.6K.
  go-code's and go-code-refactor's routing lines for `LADDER.md` now name
  the heading to look for instead of "the session-start hook".

## [1.46.1] - 2026-10-29

Ideas from "How the Claude Code team does code review" (Addy Osmani,
2026-10-26). Haiku 5.5 abrun on `refunds` (medium, n=5, 1.46.0 / this tree):
recall 0.95 / 1.00; pre-existing defects tagged 3 of 9 found / 8 of 10, change
defects wrongly tagged 0 / 0; must defects reported as Must Fix 0.73 / 0.93;
$0.0113 / $0.0116 per run. The stated-goal row is unproven: both arms find the
double refund 5/5.

- go-code-review: a defect on lines the diff did not change is reported
  tagged `pre-existing`, and the review template's verdict waits only on Must
  Fix findings the change introduced. A Correctness row replays the scenario
  the request or PR description says the change fixes, including twice or
  from two callers at once.
- abrun review corpus: a fixture with `_golden/<task>/base/` is a pull
  request. The session's tree gets two commits (base, change), and the prompt
  is the fixture's `prompt.txt` (the PR description) plus the diff. A key
  defect marked `pre_existing` must sit outside the changed lines and any
  other on one, which `validateReviewFixtures` checks. The score adds
  pre-existing found and tagged, and change defects wrongly tagged.
  New fixture `refunds`: a fix for a double refund that still lets two
  requests through (check-then-act), a sentinel formatted with `%v`, and two
  pre-existing defects outside the diff.

## [1.46.0] - 2026-10-28

Stage 1 of the 2026-10-28 review against Anthropic's skill guidance: no
wording change that needs a model run; moved sections keep their text.
Haiku 5.5 abrun (medium, n=3, 1.45.2 / this tree): refactor golden 100% both,
−17.3 / −16.7 lines; review recall 0.89 / 0.90; implement golden 97% / 100%.
zoning, pooled to n=8, writes 5 more lines (p=0.02) as its longest function
drops from 47 to 40 lines.

- Section order for compaction, which keeps the first 5,000 tokens of each
  skill (boundaries measured with the Opus 5.5 tokenizer, not bytes/2.6):
  go-code-review's checklist follows its own risk order (Less Code,
  Correctness, Security, HTTP, Database, Concurrency, Resilience, Testing, then
  the rest), so Testing's probe rows start at about 4.1K tokens instead of
  past the cut; go-code-refactor puts What "Identical Behavior" Means and
  Delete Before You Restructure right after the Workflow (both end by about
  4.9K tokens of the injected body; Risk Tiers and When Not to Refactor
  follow); go-troubleshooting moves
  Live Capture Boundaries, Fix and Regression Proof, and Investigation Result
  ahead of the symptom tables; go-http moves Partial Updates to the end, so
  Clients starts at about 4.2K; go-code moves Close With The Gate after the
  routing table and Business Rules to the end of Writing New Code.
  `TestCompactionKeepsPerEditRules` pins the order. go-style-core's How Much To
  Say stays where it is: its idiom card already ends near 5.8K, and moving the
  section ahead of it would push the Go 1.27 table rows past the cut.
- hooks: after auto-compaction the routing gate requires the router,
  `go-style-core`, and the edit's owners to be invoked again before the next
  `.go` edit. SessionStart with `source: compact` writes a `compacted` marker
  holding the line count of `loaded`, and the PreToolUse gate counts only
  loads recorded after it (`loaded` stays whole for the verification gate).
  The session note says "before the next edit", a `/go-code` typed after a
  compaction is recorded again, and with no router reloaded the gate asks for
  the one the session used before compaction.
- `allowed-tools` is `Bash(bash ${CLAUDE_SKILL_DIR}/scripts/*)` instead of
  `Bash(bash:*)`, which also passed any `bash -c`; go-linting adds the hook's
  receipt verifier. A `claude -p` smoke run on go-naming ran the script
  without a prompt and refused a `; echo` appended to it. `TestStructure`
  accepts no other grant.
- No `\$N` in a SKILL.md: Codex and opencode show the backslash and the copied
  query fails in PostgreSQL. go-database's transaction keeps the Go shape
  and states the lock-order and `RETURNING` rules in prose (the whole transfer
  stays in SQL-PATTERNS.md); go-resilience's versioned `UPDATE` moved to
  RETRIES-AND-IDEMPOTENCY.md. `TestSkillBodyArgumentPlaceholders` now rejects
  the escaped form too.
- Removed go-testing's `gen-table-test.sh` and `assets/table-test-template.go`
  (the scaffold compiled only for `func(string) string`), with go-testing's
  `allowed-tools`, and go-code-refactor's ARCHITECTURE-EXAMPLES.md (dated
  summaries of community READMEs). Counts: 72 reference files, 10 scripts,
  4 asset templates.
- verify-refactor.sh records a race check it cannot run as a skip: without
  cgo (outside macOS) or on a GOOS/GOARCH without the race detector,
  baseline and after record `race: skip (cgo unavailable)` or
  `skip (unsupported platform)` instead of `race: fail` and a red baseline;
  `--json` gains a `race` field, and `--help` documents `GO_TEST_TIMEOUT`
  (default 5m).
- check-architecture.sh checks for `go.mod` before reading the config and
  reports "no go.mod in <dir>" (naming the parent module's directory when
  there is one) instead of the raw `go list` error; a missing
  `architecture.json` prints a minimal starter file for each layout.
- setup-lint.sh, pre-review.sh, and verify-refactor.sh print the pinned
  `go install …golangci-lint@v2.15.0` for a missing golangci-lint and never
  run it; `pre-review.sh --json` adds `golangci_lint.hint` in that case.
  `TestLintInstallPinMatchesCI` keeps every install command in `skills/` and
  `docs/` on CI's `GOLANGCI_LINT_VERSION`, and go-release-watch lists the new
  places to bump.
- The baseline golangci.yml says why `max-issues-per-linter: 0` and
  `max-same-issues: 0` lift the caps: the defaults (50 per linter, 3 per
  identical text) hide repeats of a finding.
- check-interface-compliance.sh `--help` documents what it lists
  (producer-side interfaces; a finding asks whether a consumer needs the
  interface, not for a compile-time assertion), exit codes 0/1/2, and the
  `status` values.
- evalrun: trigger evals accept an optional `should_not_trigger` list, and a
  case fails when one of those skills fires. The JSON report adds per-case
  `not_want`, `unwanted`, and `precision` (expected skills fired / all go-*
  skills fired) and a pooled `trigger_precision`, also appended to `summary`;
  existing fields are unchanged. A case expecting only an owner counts go-code
  and go-style-core against precision, so compare it between runs, not with 1.
- evalrun and abrun count Read calls per skill file (`go-*/SKILL.md`,
  `go-*/references/*.md`) from the transcript: `skill_reads` per session and a
  report total (per arm in abrun) that lists every file with 0 for one never
  read. evalrun quality sessions now pass `--verbose` so the JSON output
  carries the transcript. Claude runner only.
- `TestLeakValidationEvals` flags a validation quality eval whose prompt
  reappears in skill text (3+ shared word 5-grams or 2+ non-standard-library
  identifiers, each in at most two skill files). The known leaks carry their
  evidence passage, and the test fails once one stops leaking: 19, 34–39, 57,
  59 from the review, and 31, 63, 70, which the guard found.
- References over 100 lines with more than one section carry `## Contents`
  (17 added; the bar was 200). A SKILL.md is capped at 26 KB beside the
  400-line cap, as a ceiling on growth.
- go-packages: the subcommand example's `run` takes `stdout` and `stderr` and
  sets `fs.SetOutput(stderr)`, as the rule in the body says, and both files
  show the same `main`; the example test now checks what reaches `stderr`.
- Links and names: the four AST helpers say "built and run by the script; not
  run or read directly" instead of "patch this"; go-performance and
  go-code-refactor say the same of their Go helpers and tests;
  go-defensive's TIME-ENUMS-TAGS.md is TIME-ON-THE-WIRE.md, without the
  embedding stub; go-naming routes name length to VARIABLES.md;
  LEVELS-AND-CONTEXT.md routes the redaction form to go-logging's What NOT to
  Log; WEB-SERVER.md drops the one-row Skills Applied table;
  OVER-ENGINEERING.md reads Reach For What Go Ships for ladder rungs 3 and 4,
  as the routers do; go-code-refactor routes its architecture fixture and the
  fixture README drops the note about this repository's tests; MODERNIZATION.md
  drops the `verify-refactor.sh leaks` row, which collects no profile;
  go-code-review's step 3 passes `--limit 50` when there is no diff.

## [1.45.2] - 2026-10-28

- go-performance: the snapshot directory is `.perf-snap/1` in the project
  instead of a shared `${TMPDIR:-/tmp}/perf-snap`, and the snapshot and
  verdict each run alone in their own Bash call (no `cd`, `&&`, variable, or
  `echo $?`), since the host asks approval for those. Haiku 5.5 perf corpus
  (medium, n=3): failed `bench-compare.sh` calls 47/109 → 6/47; over two
  batches (n=6) pass 46 vs 43 of 54 and runs with a verdict 39 vs 35, within
  noise.

## [1.45.1] - 2026-10-27

- go-vet-on-edit hook: the receipt context printed after every Go edit is
  one status line, the verifier command, and one line of crediting rules
  (1,448 → 656 characters for a clean edit): no exit codes, `state-stable`,
  empty reasons, or default scope and filter per check. The verifier, the
  gate, and the Stop check are unchanged.

- evalrun and abrun: a session the Claude Code CLI ends with an error of its
  own (`is_error` on the result, other than running out of turns), such as a
  usage limit, is a run error; before, the CLI's notice was graded as the
  model's answer, and a run that crossed a limit reported its later cases as
  failures.
- README and README.uk: a Claude Code benchmark section with charts for
  Haiku, Sonnet, and Opus 5.5 (2026-10-27, medium effort, no-skill vs
  baseline): implement, refactor, review, quality and trigger evals on all
  three; perf on all three (Opus at low effort).
- README and README.uk: a Grok 4.7 refactor chart (2026-10-27, medium effort,
  no-skill vs baseline, n=2): lines removed with behavior held 1.4 → 14.1,
  golden 100% → 100%, lint clean 60% → 70%.

- go-code-review: a Data Structures row for a map reset after a burst
  (`clear` or a `delete` loop keeps the peak table; a fresh `make` is the
  reset), the one `tiles` defect the 1.45.0 text left mostly unfound. Sonnet
  5.5 medium, n=4, 1.45.0 / this tree: `clear-keeps-table` 3/4 → 4/4, all
  defects 34/36 → 36/36, baits 5 → 2, $0.70 → $0.79. The row names that
  fixture's defect, so `tiles` cannot judge it alone; other review fixtures
  were not rerun.
- abrun fixtures: `meter` no longer states that a reset Meter holds no more
  memory than a new one (the scan and the tight memory limit stay), and
  `bulkload` allows ten statements instead of a few dozen, so 1,000-row
  batches fail and only batches near the 65535-parameter cap pass. Haiku 5.5
  medium, n=3, still passes both 6/6 with and without the plugin: the traps
  are known to the model, not hidden by the wording.

- evalrun: `-go-doc` gives quality sessions Bash allowed for `go doc` only,
  so an answer can check an API against the installed toolchain; the report
  records `go_doc`. Without it the sessions stay read-only, as before.
- abrun: `-keep` puts every scratch tree of a run under one
  `$TMPDIR/abrun-keep-*` directory, printed at the start and the end, so one
  `rm -rf` removes them; before, each tree was its own `abrun-work-*` in
  `$TMPDIR`.

## [1.45.0] - 2026-10-27

- evalrun: `-ids` picks quality cases by id whatever their set (an unknown id
  is an error), `-n` repeats each case, `-no-skill` runs them without the
  plugin as the control arm (quality only; needs `claude --restricted`, which
  skips the user's enabled plugins), and the report carries `cost_usd` and
  `judge_cost_usd` per run, the total in the summary, and each case's `trap`.
- go-vet-on-edit hook: a count of `go fix` hunks or lint findings in files the
  edit did not touch no longer exits 2, which Claude Code shows as "blocking
  error"; it rides the PostToolUse `additionalContext` beside the receipts.
  Findings in the edited file, and package-wide vet or test failures, still
  exit 2 with the counts appended.

- Ideas from samber/cc-skills-golang (MIT, `246b6c4`), each re-verified on
  go1.27.1 or the module source and written in this pack's words; their claims
  found wrong for 1.26/1.27 were left out.
  - go-data-structures: new "A Map Never Shrinks" (`delete` and `clear` keep
    the peak table, measured 30 → 26 MiB; reset with a fresh `make`);
    go-performance's Bounded cache bullet and MODERNIZATION's `clear` entry
    route to it. go-performance also gains `unique.Make` versus a `weak.Pointer`
    map, a per-request loader, a `scanobject` row in Read the Profile, and one
    rerun on `noisy` (a second `noisy` is no measurable change).
  - go-http: new `references/GRPC.md` (status codes at the handler,
    `Unimplemented…Server` by value, `grpc.NewClient` and its `dns` default,
    bounded `GracefulStop`, bufconn tests), routed from go-code; the hook
    helper names go-http for a `google.golang.org/grpc` import or a
    `grpc.NewServer`/`NewClient`/`Dial` call, so the routing gate and the
    prompt note load it for gRPC code.
  - go-security: a duration or count branch in Triage, size products bounded
    factor by factor, key files through `CreateTemp` + `Rename`, a key ID on
    stored ciphertext, the bcrypt `ErrPasswordTooLong` contract.
  - go-logging: `ParentBased` around a ratio sampler, `otelslog` needs a
    `LoggerProvider`, histogram buckets and metric suffixes, exemplars over
    OpenMetrics, log forging through `log.Printf`, PII in error-tracker SDKs.
  - go-concurrency: no recursive `RLock`. go-database: the 65535-parameter
    cap and `CopyFrom`. go-defensive: `NaN` comparisons after a float divide,
    `.UTC()` dropping the monotonic reading. go-packages: `replace`/`exclude`
    apply only in the main module; a CLI `run` takes its writers and builds
    its flag set. go-troubleshooting: capture before a restart.
    MODERNIZATION: `(*big.Int).Divide` (Go 1.27, `big.Round` is half to even)
    and `bytes.Buffer.Peek` (Go 1.26) as Tier 2; the `errors.AsType` interface
    target embeds `error`.
  - evals: quality cases 65–75 with an optional `trap` field, accepted by
    `TestEvalsJSONSchema`; docs/SKILL_AUTHORING_TEMPLATE.md gains "Quality
    Eval Cases" (the trap, premise-pressure prompts, the no-skill pre-flight
    table). Case 72 also accepts one `INSERT ... SELECT FROM unnest` with an
    array per column.
  - abrun: implement fixtures `meter` (a window map reset after a scan; the
    golden test measures the heap after `Reset`) and `bulkload` (50,000 rows
    against a fake driver with PostgreSQL's 65535-parameter cap), review
    fixture `tiles` (nine seeded defects from the rules above). Measured
    2026-10-26, medium, n=3, arms no-skill / reference (1.44.0) / baseline:
    implement golden 6/6 in every arm on Sonnet 5.5 and on Haiku 5.5, so
    both traps are saturated; review `tiles` on Sonnet 5.5 found 22/27,
    25/27, 25/27, with `clear-keeps-table` the one the new text did not move
    (0/3, 1/3, 1/3). evalrun on cases 65–75 (Sonnet 5.5, ×3): reference
    24/33, baseline 27/33 (66, 70, 74 one each), within noise at this n.

## [1.44.0] - 2026-10-26

- go-code: new "Business Rules" subsection under Writing New Code for multi-step
  rules (intervals, history or snapshots, reconciliation, mode or role
  branches): the entry point is a chain of named step results, effects stay at
  the top, untyped data is parsed once into a typed struct and decisions are
  pure, a chain step or a pure step below it meets Declaration Budget rule 4
  (the parse, the pure match, and the interval arithmetic named as such, never
  closures; the entry point within about 40 lines), domain step names,
  complexity-linter headroom. go-code-refactor's PLAYBOOK routes to it. New
  implement fixture `zoning` (interval binding over snapshot history). Haiku
  5.5 medium, n=5, `zoning`: reference golden 4/5, largest function 72–97
  lines, untyped values decoded in one function 1/5; with the section golden
  5/5, largest function 35–47 lines in 4/5 (70 in one), decoded once 5/5.
  Not measured on Opus or Sonnet 5.5.

- Ideas from spf13/go-skills, verified on go1.27.1: go-troubleshooting's
  "Edits do not change behavior" suspects the code before the toolchain (logic,
  wrong file, second call site, other code path); go-code-refactor's
  STRUCTURAL.md moves a god struct's state with its mutex and methods;
  go-http names `http.TimeoutHandler` for one slow route (buffered, no
  `Flusher`); MODERNIZATION.md lists the Go 1.27 platform floors (macOS 13,
  `linux/ppc64` ELFv2, no `windows/arm`). Not measured with `abrun`.

- go-linting: the receipt mechanics moved from the Verification Gate to a new
  `references/RECEIPTS.md`. The gate keeps the part a session acts on: run the
  printed verifier alone in its Bash call, before any direct check and again
  after every later Go edit, and credit only `hook_credit=true`. SKILL.md
  12.0 → 10.3 KB. A missing tool is now reported, never installed: the gate
  runs no `go install`/`go get`, edits no `go.mod`, switches no toolchain.
  go-style-core House Style Wins: a path, file, package, or entry point the
  task names is fixed. Ideas from arXiv 2608.11888 (skill-body context bloat,
  excessive verification, environment-state mismatch, artifact misplacement).
  Measured with `abrun -corpus implement -shell go`, five fixtures, medium:
  Sonnet 5.5 n=5 golden 25/25 and lint 0 in both arms, cost per session
  $0.346 → $0.330, verification-gate Bash blocks 10 → 7; Opus 5.5 n=2 golden
  10/10, lint 0, $0.769 → $0.718, blocks 4 → 1. A first cut without the
  "alone, before any direct check" clause raised Sonnet's blocks from 1 to 9
  in ten sessions, five of them a verifier glued to `go vet && go test` in one
  Bash call. The path and no-install rules are not exercised by this corpus.

- README: a Benchmarks section records the 2026-10-24 `abrun` implement runs on
  DeepSeek V4.1 Flash and Muse Spark 1.3 through the opencode runner, with the
  flags to reproduce them. Both models built and passed the hidden golden test
  in both arms; the skills cut production lines by 25–31% against the
  no-skill control and left fewer lint findings. The section states that this
  runner does not report skill firings, so the numbers compare an armed home
  with an empty one rather than proving a specific skill loaded.

## [1.43.2] - 2026-10-24

- Every Go block in a skill that shows a whole file is now gofmt-clean, and
  `TestFileExamplesAreGofmtClean` keeps it so: eight blocks were indented with
  spaces (one also listed its imports unsorted), the doc comment in the
  go-documentation `EXAMPLES.md` indented its code with spaces where gofmt
  writes a tab, and a one-line struct with a comment became a two-line one.
  A model copies the layout along with the code. Idea from JetBrains
  go-modern-guidelines PR #28; nothing else in that repository is new since
  `155dc7c`.

## [1.43.1] - 2026-10-24

- `abrun -runner grok` drives the Grok CLI (`grok -p`). Each arm gets its own
  `HOME` and `GROK_HOME`, so a control arm does not see skills installed for
  the operator. The arm's `skills/` tree is what the session can read; this
  plugin's hooks are not installed. A go-* skill counts when a `read_file`
  result returns the whole `SKILL.md`. `-effort` is passed as
  `--reasoning-effort`. `-gopls` and `-shell` stay claude-only: a grok session
  already has a shell, and a review run has the edit and shell tools removed.

## [1.43.0] - 2026-10-23

The remaining findings of the 2026-10-22 review. The wording changes under
"Measured" were checked by abrun against the tree with every other change
here (reference) on 2026-10-23, medium effort, n=2, no regression:

| Corpus | Model | golden ref → base | $/run ref → base | Other |
|---|---|---|---|---|
| implement (12) | Sonnet 5.5 | 24/24 → 24/24 | 0.261 → 0.252 | gate blocks 23 → 18; test first 8/24 → 11/24 |
| implement (12) | Opus 5.5 | 24/24 → 24/24 | 0.628 → 0.601 | test first 22/24 → 22/24 |
| refactor (5) | Sonnet 5.5 | 10/10 → 10/10 | 0.201 → 0.202 | Δlines −18.4 → −18.8 |
| refactor (5) | Opus 5.5 | 10/10 → 10/10 | 0.458 → 0.464 | Δlines −18.6 → −19.6 |
| review (10) | Sonnet 5.5 | recall 0.92 → 0.96 | 0.176 → 0.177 | as-must 0.89 → 0.83 (noise band ±0.04) |

### Measured

- **One owner for the receipt protocol.** go-style-core and go-code route to
  go-linting's Verification Gate instead of restating it, and the edit hook
  prints one `--gate` verifier instead of one command per check.
- **The restraint ladder lives in `go-code-refactor/references/LADDER.md`.**
  The ladder hook reads it there; a host without the hook reads 5.1K
  characters instead of the whole OVER-ENGINEERING.md (14K).
- go-code: a fix skips the Contract Table but adds one case that fails before
  the fix, and loads go-testing for it; the table's error case is the error a
  caller matches plus any text the documentation promises. The routing table
  names timers, receiver types, and `time.Time`/`Duration` values; five
  descriptions say the same, and go-code's routes a behavior-preserving
  refactor to go-code-refactor.
- go-testing: the error-semantics example names the error (`wantErr`,
  `errors.Is`); the main example no longer imports go-cmp; a parent-bound
  `require` under `t.Parallel`; integration packages share one database; the
  fake and the real store run one contract table.
- go-linting: the default gate always runs `-race`; govulncheck without a
  dependency change is `skipped (no dependency change)`; "Quiet the compiler"
  joins Holding the Bar, and a check failing the same way after two fixes
  goes to go-troubleshooting.
- go-code-refactor: "usually deletable" names only what is provably safe; the
  checklist covers evaluation order and side effects; `cmp.Diff` replaces an
  `if` ladder only when go-cmp is already required and no field is
  unexported.
- go-concurrency, go-error-handling: errgroup only when go.mod already
  requires `golang.org/x/sync`; go-concurrency defers `GOMAXPROCS` to
  go-performance.
- go-generics: reflection and `text/template` do not see a generic method;
  `comparable` admits interfaces, so `==` panics on `any` holding a slice.
- go-database: a unique value is decided by its constraint (`23505` to the
  domain sentinel, or `ON CONFLICT DO NOTHING`); read-your-writes go to the
  primary; pgx behind PgBouncer transaction pooling. go-logging: OTel
  `trace_id` from the span context. go-resilience: deduplication is not
  ordering (`version < $2`). go-performance: a cache write deletes the key.
  go-error-handling: wrap context names the operation, not "failed to".
  go-troubleshooting: the investigation note lives outside the tree.

### abrun

- `routing.contract_first` records whether a session's first `.go` edit was
  a test file; `suppressions` counts new `//nolint`, `//lint:ignore`, and
  `#nosec` in production files; a fixture can carry its own prompt in
  `_golden/<task>/prompt.txt`.
- New implement fixtures `signup` (uniqueness decided by the constraint) and
  `lintfix` (one real errcheck finding, one gosec false positive). Both are
  saturated: every session in every arm, no-skill included, passes the
  golden. In `lintfix` no-skill swaps jitter to `crypto/rand` 3/3; the skill
  arms add a reasoned `//nolint:gosec` in 6/8 sessions, the same with and
  without a drafted "no new suppression without approval" rule, which was
  therefore dropped.

### Hooks

- **The hooks no longer need `python3`.** Their JSON parsing, the routing
  heuristics, the verification gate's shell parser, and the receipt runner
  (`go-check-receipt.py`) are one standard-library Go program in
  `hooks/helper`, which `hooks/go-hook-helper.sh` builds on first use into
  the user cache directory, keyed by a checksum of its sources. SessionStart
  hooks get 90 s for that first build. Without `go` the routing gate says
  once per session that it is off; its marker lives in the session's state,
  not one file for the whole machine. CI no longer runs Python either: the
  link check is `TestRelativeLinks`, the Go release watch reads the download
  index with `jq`. abrun, Sonnet 5.5 medium with `-shell go`, 1.42.0 hooks →
  these: implement (feed, gateway, catalog, roster, n=2) golden 8/8 → 8/8,
  $0.321 → $0.333 per run; refactor (five fixtures, n=1) 5/5 → 5/5, Δlines
  −18.4 → −19.0; gate blocks, receipts, and `hook_credit` counts alike, no
  helper failure in any trace.
- **A test receipt passes only when a test ran.** The edit hook runs the
  package tests with `-json` under the receipt runner and records `skipped`
  ("no test ran") when exit 0 ran no test (a `TestMain` without `m.Run`, every
  test skipped); a failure shows only the failing tests' output.
- **An unasked lint run loads no repository plugin.** A repository config that
  declares `custom:` linters is replaced by the bundled baseline in the edit
  hook; the receipt keeps the repository config and says why, so that run
  earns no hook credit and an edit of the config invalidates it. README says
  the hook runs repository code and names the switches.
- The routing gate counts only a whole Read of a `SKILL.md`. README says shell
  edits are not held and that Codex runs the hooks with the routing gate off.
- The prompt hook no longer reads "Go ahead and …" as the language, nor "Go
  to", "Go through", "Go over", or "Go on" where a sentence starts ("in Go to
  cut latency" still names it), and no longer uses `mapfile` (macOS bash 3.2); `TestPortability` checks
  `hooks/*.sh` for bash 4 features.
- The speed gate accepts a negated claim ("does not make Parse faster") when
  at most three words, none of them `and`/`but`, stand between the negation
  and the claim: "doesn't change the API and makes Parse 3x faster" is still
  a claim.
- Receipt runs and verification state older than two days are removed with
  the routing state.
- The verification gate names a Read fallback for go-linting; `go-verify`
  without a repository config lints with the bundled baseline and reports
  `INCOMPLETE` when only golangci-lint's defaults ran.
- Hook comments are in English.
- `TestSkillBodyArgumentPlaceholders` also rejects `${CLAUDE_*}`.

### Scripts

- **The five AST wrappers build their helper for the host**, whatever
  `GOOS`/`GOARCH` the caller exports, key the cache by platform, fall back to
  `$TMPDIR` when the cache directory is read-only, and exit 2 when the helper
  cannot run.
- **A second positional path exits 2** in `pre-review.sh`, `check-naming`, and
  `verify-refactor.sh` instead of silently checking only the last one.
- `pre-review.sh`: a `go vet` that fails for the environment (no module, a
  toolchain or network failure) is `unavailable` (`--strict` exits 2), one
  that matches nothing or a missing path is `no_go_files`, and any other
  failure, an import cycle included, is `fail`; JSON gains
  `"result": "pass|fail|incomplete"`; a new `.golangci.*`, untracked included,
  is `lint-config-changed`.
- `check-interface-compliance` leaves a pair undecided
  (`"status":"incomplete_types"`) when an import has no export data instead of
  reporting a false implementation, resolves nested and `go.work` modules from
  their own roots, skips aliases and type-set constraints, counts an
  instantiation as the compiler's check, and no longer stops on a broken test
  file it does not scan.
- `check-errors` sees logrus, zap, zerolog, klog, glog, and sugared loggers,
  keeps the pair across `err = fmt.Errorf(..., err)` and a returned
  `&opError{Err: err}`, and no longer flags `err.Error() != ""` or a
  `catalog` field. It and `check-docs` skip nested modules and go.mod
  `ignore` directories.
- `check-architecture`: a `module-root` without its own go.mod exits 2, and
  so does a `go env` that cannot name the host platform; a
  vendored module lists with `-mod=vendor`; a test-only package is skipped
  without `--include-tests`.
- `verify-refactor.sh` runs one gofmt over the files of the directories
  `go list` resolves (no `testdata/` or nested modules); `//line` no longer
  lowers code LOC.
- `check-debt.sh` no longer reports a clean ledger past ARG_MAX; a scan
  failure exits 2. `gen-table-test.sh` exits 2 on an unwritable `--output` and
  rejects a black-box `*_test` package.

### CI

- **CI pins golangci-lint v2.15.0.** The bundled baseline config verifies
  unchanged under it; it sets none of the `exhaustive` options 2.15.0
  renamed. go-linting and the release checklist name the same version.

### Skills

- **go-linting baseline** sets `run.relative-path-mode: gomod` (exclusions
  matched the project's ancestors, so revive `exported` never ran under
  `cmd/` or `internal/`) and enables revive `defer` with `loop`. Gate rules:
  a check's result is its own exit status, not a pipeline's; an all-skipped
  package is a run with no tests.
- go-style-core: the idiom card writes JSON with `json.Marshal` then one write
  (`MarshalWrite` leaves a truncated 200 on a mid-encode failure); the
  directive rule names the forms the compiler does not gate; ServeMux
  patterns need Go 1.22.
- go-code-refactor: `waitgroupgo` and `slicessort` hunks are Tier 3; an alias
  move changes `%T`, `reflect`, and `gob` names; plus `hostport` scope,
  `nojsonv2`, FMA by architecture, `rand.NewSource` concurrency, `COALESCE`,
  and a summary line before `Deprecated:`.
- go-code-review and OVER-ENGINEERING route to the right owners; the
  test-server item says `Start` comes before any `Client()`.
- go-performance: a per-write syscall is batched with `bufio` before
  declining; mutex profiles show `Unlock`, block profiles `Lock`; profile
  commands write under `$TMPDIR`; BENCHMARKS.md decides with `--verdict`.
- go-context: only `WithCancelCause` takes `cancel(nil)`; the value example
  goes through an accessor. go-concurrency: `SetLimit(max(limit, 1))`; a
  leftover `GODEBUG=asynctimerchan=1` stops a 1.27 binary at startup.
- go-logging: the route label comes from `mux.Handler(r)` before `next`; one
  composite propagator; the custom level example sets the handler level and
  names `TRACE`.
- go-http, go-security: one `CrossOriginProtection` per server; mTLS identity
  is `VerifiedChains[0][0]`; outbound bodies read `limit+1`; sniffing and v1
  unknown-field behavior corrected; delete-path facts match go1.27.1.
- go-database: `sql.Open` does not parse a pgx DSN; 40001 can come from any
  statement; name the columns instead of `SELECT *` with positional `Scan`;
  `CREATE INDEX CONCURRENTLY` sits alone in its migration; backfills walk the
  key. go-packages: a checksum mismatch is a finding; `main` exits through
  `slog.Error` + `os.Exit(1)`.
- go-interfaces: a recover guard uses comma-ok and re-panics `p` unchanged.
  go-functions: options go on a new constructor, not an existing one; the
  embedding row covers fields; the `Formatter` example is errcheck-clean.
  go-naming: `errname` checks prefix and suffix only. go-generics: a zero
  `maphash.Seed` panics.
- go-defensive and go-data-structures examples compile, wait for their
  goroutine, and print what the comments say.
- go-testing: the fake store returns copies; `go 1.26` fallbacks for
  `NewTestServer` and `synctest.Sleep`; the duplicate acceptance section is
  gone.
- go-troubleshooting: `%v` wrapping is confirmed with errorlint; CI-only
  races, deadlock-detector blind spots, untracked bisect tests, untruncated
  goroutine dumps grouped by function and line, `curl -fsS`, sampled heap
  profiles, and the right owners for timers and `GOMEMLIMIT`.
- COMPATIBILITY.md: `forvar` scope, `b.N` inside `b.Loop`, slogtest versions,
  nine recommended APIs, and the missing-`go`-line rule under
  `-mod=readonly`.

## [1.42.0] - 2026-10-23

- **Subagents get the router note and the ladder.** On SubagentStart Claude
  Code shows a subagent only JSON `additionalContext` and drops plain stdout;
  `go-subagent-routing.sh` and `go-restraint-ladder.sh` printed plain text, so
  no subagent received either. Both now wrap their text for that event;
  SessionStart stays plain. `subagentEvent` and `ladderEvent` in
  `evals/hook_test.go` fail on a plain SubagentStart reply.
- **Each agent keeps its own hook state.** Subagents share the session id, so
  one agent's skill loads lifted another's routing requirement, and one
  agent's Go edit blocked every agent's `go test` until it ran a verifier for
  a package it never touched. Routing records now live under
  `routing/<session>/agent-<agent_id>/` (a fork keeps its parent's, since it
  starts from the parent's context; `prompted` and the intensity stay
  session-wide), and the verification state is keyed by session and agent.
  Covered by `TestRoutingGate/each subagent keeps its own loads` and
  `TestVerificationStatePerAgent`.
- **The edit hook runs on macOS.** The package tests and the linter ran under
  GNU `timeout`, which stock macOS lacks, so every `.go` edit there ended with
  both checks `unavailable` and exit 2. The receipt runner now enforces the
  per-check limit itself and kills the process group; `timeout` is used only
  on the path without python3, when present.
- **Receipts survive the Bash tool's environment.** The build-environment
  digest covered every variable, and the Bash tool differs from the hook
  process in `COLUMNS`, `CLAUDE_CODE_*`, and proxy settings, so `--gate`
  credited nothing in interactive sessions (0 of 42 sessions with an attempt;
  headless abrun 476 of 511). It now hashes `GO*`, `CGO_*`, `CC`, `CXX`, `AR`,
  and `PKG_CONFIG*` beside `go env -json` and the tool binaries. A hook test
  receipt no longer earns credit from the single-receipt verifier either, as
  the batch verifier already refused it: the run is `-short` without `-race`,
  and its exit 0 does not show that a test ran.
- **Overlapping `go fix` fixes are reported.** When two analyzers' fixes
  overlap, go1.27.1's `go fix -diff` prints no hunk and exits 1 with only
  `fix: N of M fixes skipped`, and a plain `go fix` then reports them applied
  without writing a file; the receipt called that `unavailable`. The hook now
  records `fail` and says to preview one analyzer at a time, and go-linting's
  Modernization section says the same.
- **PATCH copies the record's slices and maps.** go-http's Partial Updates
  copied the record's fields into the request value and decoded the body onto
  it; v2 writes a non-nil slice's array, merges into a non-nil map, and
  decodes through a pointer, so a body later rejected had already changed a
  shared record (`Tags:[evil]` after `RejectUnknownMembers` refused it). The
  recipe now says to clone such fields, and the example has a `Tags` field,
  `slices.Clone`, and `http.MaxBytesReader`; `TestHTTPExamplePartialUpdate`
  has the rejected-tags case.
- **A secret in an unexported field is redacted too.** `fmt` cannot call
  `String` or `GoString` on an unexported field, so a struct holding the
  go-logging `Token` there printed it raw through `TextHandler`, `%v`, `%+v`,
  and `%#v`; the 2026-10-01 fix covered exported fields only. What NOT to Log
  adds a holder that prints itself without the secret, go-security's row
  points at it, and `TestLoggingExampleRedaction` logs it.
- **Receiver consistency keeps its decoder exception.** "If any method needs
  a pointer, use pointers for all" moved `MarshalJSON` and `String` onto `*T`
  for a type whose `UnmarshalJSON` needs a pointer, and v1 JSON and `fmt`
  then skip them for every value that is not addressable, with no error and
  no vet finding. go-interfaces states the `time.Time` exception with the
  pair in code and a value assertion; go-functions, go-data-structures, and
  the go-code-review row agree with it.
- **Float equality has an absolute floor.** go-defensive's relative tolerance
  alone is never true against `0`, so a residual check never fired; the row
  now gives `max(absTol, relTol*…)`.
- **The cache example survives a leaving caller and a panicking load.**
  go-performance's in-flight entry loaded under the first caller's `ctx`, so
  that caller leaving failed every waiter, and a panicking `load` left the key
  loading for good. The load now runs under `context.WithoutCancel(ctx)`, the
  entry is finished in a `defer`, and Bounded says that replacing an expired
  entry on read bounds nothing for keys never read again. `TestRates` covers
  both failures.
- **The speed verdict shares its significance level.** `bench-verdict.go`
  tested each benchmark and unit at alpha and kept on any better row, so 20
  unchanged benchmarks came out `keep` in about one verdict in five; alpha is
  now divided among the judged rows whose runs vary (Bonferroni,
  `TestVerdictSharesAlpha`). The loop's snapshot and verdict take
  `--filter` for the step 1 benchmark, and a kept change is snapshotted
  before the next one, which against the first snapshot would read `keep` on
  the earlier gain. Also: `--snapshot` with a sub-benchmark filter no longer
  reports no benchmarks, `--threshold 05` (not a JSON number) is a usage
  error, and the speed gate no longer counts `--verdict=DIR`, which the
  script rejects.

## [1.41.0] - 2026-10-22

- **A speedup claim needs a verdict.** After an edit of a non-test Go file,
  the Stop hook in `go-verification-routing.sh` blocks a final message that
  claims a speedup (`faster`, `speedup`, `sec/op −N%`) when no
  `bench-compare.sh --verdict` ran in the session; a line that reports the
  comparison as skipped, or says there is no speedup, passes.
  `GOLANG_SKILLS_SPEED_GATE=off` turns it off. Covered by
  `TestSpeedClaimGate`. Perf corpus, Sonnet 5.5 medium, n=5, gate off vs
  on: speedup sessions that ran a verdict 14/35→27/35 (Fisher p=0.003,
  pre-registered), pass 45/45 in both, cost +7%; every one of 14 blocks
  ended in a verdict or a report that called the speedup unmeasured. Opus
  5.5 at n=2: verdicts per run 0.7→0.9.
- **Perf fixtures without lint noise.** The `trees` bait (`formatLine`), the
  `quotes` decline target (`NormalizeName`), and the `ingest` bait
  (`Summarize`) had perfsprint, modernize, or `go fix` findings that the edit
  hook pushed sessions to fix, failing the run for touching them. Each keeps
  its slow look with zero findings. Perf runs from 1.40.0 are not comparable
  with later ones on these three fixtures; take a new baseline.

## [1.40.0] - 2026-10-22

- **go-performance runs speed work as a loop the verdict closes.**
  Benchmarking and Profiling is now seven numbered steps:
  1. A benchmark of the caller's path, which is a sample, not the
     specification.
  2. `bench-compare.sh --snapshot` before the first production edit.
  3. The profile read first; a function under 5% of it is left alone and
     its share reported.
  4. One change from the cheapest class.
  5. `--verdict` as the keep or revert decision. `noisy` is never kept, and
     a `keep` on allocations alone is not reported as faster.
  6. An `AllocsPerRun` guard for an allocation win.
  7. A report of kept and reverted attempts with the lines each added.

  New subsections back the steps. Read the Profile maps the profile row to
  the class of fix (repeated work, allocation and GC, contention, time off
  the CPU, platform work to decline). Fix Order puts doing less before
  allocating less, contending less, more CPUs per request, and hand-written
  replacements of the standard library. Allocation Guard carries
  `//go:build !race` and no `t.Parallel` in the sample, since the race
  detector drops pooled objects and `AllocsPerRun` panics under parallel
  tests.

  Runtime Settings gains:
  - the rule that `GOGC`, `GOMEMLIMIT`, `GOMAXPROCS`, and `runtime.GC` belong
    to the process, never to a package;
  - `GOGC` measured with `--new-env`;
  - `GOMEMLIMIT` headroom, and its misuse near an OOM;
  - PGO from a production `default.pgo`.

  Related Skills routes round trips to go-database and go-http. BENCHMARKS.md
  adds mutex and block profiles from a benchmark and reading `Total samples`
  against `Duration`. Research behind it: SWE-fficiency, GSO, and PerfAgent
  on where agents fail at optimization (the wrong function, a fast path
  keyed to the benchmark, no measurement).

  Measurement setup: `abrun -corpus perf -shell go`, all nine fixtures,
  medium effort, with the reference arm taken from a 1.39.0 worktree.

  - **Opus 5.5, n=2:**
    - Pass 11/18 → 16/18; the declines in digest and quotes 0/4 → 3/4.
    - Profiles per run 0.1 → 2.9; the session's own `--verdict` 0 → 0.8;
      measured before the first edit 16/18 → 18/18.
    - Δlines per run +8.8 → +6.4. On events the unaided arm once
      hand-wrote a JSON encoder (+73 lines) and the skilled arm stayed at +1
      both times. On tally the skilled arm kept `Snapshot` consistent across
      keys, at +29 and +51 lines against +8.
    - Cost was flat: $0.87 vs $0.85 per run.
  - **Sonnet 5.5, n=2, before step 3 named the request's function as a
    guess:**
    - Pass 16/18 → 18/18; profiles per run 0.2 → 2.3; `--verdict` 0 → 0.9.
    - Cost $0.34 → $0.38.
  - **Sonnet 5.5 decline fixtures (digest, quotes):**
    - Before the step 3 wording, n=5: 7/10 in both arms. The two baseline
      failures on digest never ran a benchmark.
    - With the wording (the text shipped here), the test was pre-registered:
      Fisher's exact test on all runs of each arm pooled. The comparison is
      reference 30/44 → 36/40 passes, p = 0.018.
    - The gain is in quotes: 16/22 → 20/20 (p = 0.022). The unaided arm
      rewrites `NormalizeName` beside the real fix.
    - Digest moves 14/22 → 16/20 (p = 0.32): both arms still swap `toHex`
      for `hex.EncodeToString`, sometimes after measuring.
    - The n=15 run alone gives 21/30 → 27/30 (p = 0.10); its cost per run
      was $0.245 → $0.261.
    - A sentence forbidding tidy-ups on the way did not help digest (9/15)
      and was dropped. Only 27 of 57 digest sessions had loaded
      go-performance at all; with it loaded the shipped text passed 16 of
      17, without it 16 of 30 passed.
- **The prompt hook routes a request for speed.**
  - A prompt asking to make code faster ("speed it up", "faster",
    "optimize", "reduce latency", "allocate less", and the Ukrainian and
    Russian forms) now gets the go-code note with
    `go-performance … too, since the request asks for speed`.
  - Before, such a prompt with no work verb got no note at all.
  - A path to Go source in the prompt (`./digest`) now counts as a code
    element, like a noun does.
  - Bare "slow", "optimistic", and "high-performance" do not match.
  - Measured on Sonnet 5.5 decline fixtures, n=15, against the reference
    arm's pooled runs:
    - digest loaded go-performance 14/15 (was 27/57) and passed 14/15
      against 14/22 (p = 0.056: the pre-registered p < 0.05 is missed
      narrowly);
    - quotes 15/15 against 16/22 (p = 0.063);
    - both together 29/30 against 30/44 (p = 0.003);
    - $0.26 per run.
  - Regression check on the seven speedup fixtures, n=2, baseline only:
    - pass 13/14 against the reference arm's 14/14; go-performance loaded
      14/14 (was 11/14); $0.387 against $0.351 per run.
    - The one failure rewrote trees' bait `formatLine` to clear the
      `perfsprint` finding the edit hook printed.
  - Known confound: trees' bait `formatLine` and quotes' target
    `NormalizeName` carry lint and `go fix` findings, which the edit hook
    shows to every arm with the plugin. The new fixtures keep their baits
    and targets lint-clean.
  - The new speedup fixtures (exporter, tariff, thumbs) passed in every
    arm. On these models they measure code cost, not outcome.
- **Version facts checked against go1.27.1.**
  - On 1.27, `encoding/json` itself runs on the v2 implementation
    (`internal/buildcfg` `JSONv2: true`), so switching the import is not a
    speed lever. Corrected in go-performance, the go-packages replacement
    table, and COMPATIBILITY.md.
  - Removing `automaxprocs` changes `GOMAXPROCS` under a small or fractional
    CPU limit: the runtime rounds up with a floor of 2, where automaxprocs
    rounded down with a floor of 1.
  - BENCHMARKS.md: `b.N` is 0 inside a `b.Loop` loop; `StopTimer` before the
    loop fails the benchmark; `b.Loop` keeps assigned variables alive and,
    since 1.26, no longer blocks inlining.
  - go-concurrency: `NumCPU` counts the logical CPUs in the affinity mask,
    not hardware cores.
- **bench-verdict 1.1.0: allocations clear the threshold too.** A
  hidden-quadratic fix measured −89% `sec/op` and added 7 map allocations to
  6000 (+0.12%), and the verdict was `revert`, because any significant
  `allocs/op` change counted. Allocations now need `--threshold` percent like
  time and bytes; a change from zero still always counts. `bench-compare.sh`
  1.5.1 updates its help to match.
- **abrun perf corpus: code cost and harder fixtures.**
  - The summary and the per-run lines report Δlines and Δfuncs (closures
    included) beside pass.
  - Golden tests run under `-race`; tally's golden test now reads `Snapshot`
    while it counts.
  - A key may set `min_speedup_pct`: a kept change below it fails the
    fixture.
  - Four new fixtures, each checked against its intended fix and its trap
    with the verdict:
    - `exporter`: a hidden quadratic. The prompt names only the API. The set
      fix is −85%, while polishing `fmt.Sprintf` is −13% and fails the 50%
      floor.
    - `tariff`: rules re-parsed per call. Parsing once is −89%. A memo of
      totals keyed by the cart is −88% on the visible one-cart benchmark and
      +22% on the golden one.
    - `thumbs`: `image.Image.At` per pixel. A `Pix` fast path is −92%.
      Averaging at 8 bits fails the golden test, and a row fan-out is
      `noisy` under the golden `RunParallel`.
    - `digest`: a decline without a profile; 99% of the time is in SHA-256.

  Each new fixture's baits and targets are lint-clean, so the edit hook in
  a skilled arm does not point at them.

## [1.39.0] - 2026-10-21

- **bench-compare.sh decides keep or revert.** `--snapshot DIR` freezes one
  package's test binary and copies its test files; `--verdict DIR` builds the
  current code against those frozen test files (`go test -overlay`, so a
  benchmark edited after the snapshot is not what gets compared), runs base
  and new alternately `--count` times in the package directory, and hands
  both outputs to the new `bench-verdict.go` (standard library only: medians,
  the 95% interval, the Mann-Whitney U test). The verdict is `keep` (exit 0)
  for a significant improvement with no significant regression, `noisy` when
  the spread is wider than `--threshold` (default 5%), `revert` otherwise; a
  snapshot from another toolchain is an error. `--new-env` measures a runtime
  setting such as `GOGC` against an unchanged snapshot. The existing
  `--save`/`--baseline` modes are unchanged, and the skill text does not use
  the verdict mode yet.
- **abrun -corpus perf.** Five fixtures under `evals/ab/_perf`: GC-bound
  binary-trees, allocation-bound JSON lines, a contended mutex, a store
  client called once per record, and a decline case asking to speed up a
  function at 0.15% of the profile it ships with. Each key holds the prompt,
  the golden benchmark, the targets, and cold bait functions. After every
  session the harness measures the golden benchmark serially with
  `bench-compare.sh --verdict` against the fixture as given; a speedup run
  passes on `keep` with the golden test passing and no bait touched, a
  decline run when its target is unchanged. The report also records the
  functions changed and whether a benchmark or profile ran before the first
  edit. Needs `-shell go` with claude. A first smoke (Sonnet 5.5 medium,
  n=1, no-skill against baseline) passed 5/5 in both arms: the fixtures name
  their hot path and are too easy to judge a skill edit yet.
- **Quality eval 64** is the decline case without a shell: a request to
  speed up a function a pasted profile puts at 0.5%.

- **go-testing says which tests earn a place.** A new What Earns a Test
  section: each test or case names the bug it catches; cases go per input
  class and boundary, through the exported contract; getters, assigning
  constructors, pass-throughs, restated constants and compiler-checked facts
  get none; a mock echo is replaced by the observable result. The rule covers
  tests a task writes, not existing ones. The Contract Table counts clauses of
  the documentation, not branches of the body, and the restraint ladder's
  test paragraph routes here. abrun reports `tests Δ/run` (lines, `Test`
  functions, cases) from the tests a session wrote. Implement corpus
  `catalog,gateway,ledger`, medium, n=5, against 1.38.0: Sonnet 5.5 cases
  per run 10.5 → 9.0 and test lines 71 → 66, all of it on `gateway`; Opus 5.5
  unchanged (12.0 → 11.9). Golden 15/15 in every arm. The case counts were
  taken before the counter scoped a table to its own test function (a
  file's last `tests` table stood for every test in it) and stopped counting
  `TestMain` and zero-argument `Run` calls, and before it read a range's
  table from the binding nearest before it (two groups that each declare
  `tests` counted the last one twice), shared package-level tables across a
  package's test files, counted tables of a named slice or map type, and
  stopped counting a subtest over a table as one case more than its rows;
  re-measure them before citing.
- **What Earns a Test catches the tests that pass by construction.** Taken
  from obra/superpowers `writing-good-tests.md` and mattpocock/skills `tdd`:
  `want` is a literal or hand-checked fixture, never computed by the code
  under test; a constant is tested through what depends on it; the contract
  is observed through itself, not through the table behind it; a closing
  mental mutation pass finds the missing case. `TEST-ORGANIZATION.md` says a
  fake keeps the real contract (sentinels, filled fields). Table tests point
  a parser or encoder at a `FuzzX` round trip, and the testify parent-`t`
  trap names its tell. Not yet measured with abrun.

## [1.38.0] - 2026-10-20

- **Traps a model writes without the skill, checked on go1.27.1.**
  go-concurrency's `SYNC-PRIMITIVES.md` gains a Timers section: Go 1.23
  semantics hold for every module on a 1.27 toolchain, so the old
  `if !t.Stop() { <-t.C }` drain now blocks forever, and `time.After` in a
  `select` loop restarts the timeout on every message. go-testing gains the
  testify `assert.New(parent t)` trap. go-logging's checklist gains Traces,
  including the no-op global propagator that breaks `traceparent` across
  services. go-performance notes that cgroup-aware `GOMAXPROCS` makes
  `automaxprocs` removable under a `go 1.25` directive.
- **An abrun refactor fixture with rename traps.** `spool` holds two that
  compile and pass `go vet` after a careless rename: a parameter `d` whose
  natural name `digits` is already an inner variable, and a method `fl`
  matched only by a type assertion in another file, which only the text
  search finds.
- `docs/PROJECT_INSTRUCTIONS.md` shows the Cursor `.mdc` rule form of the
  routing block.

## [1.37.0] - 2026-10-19

- **go-security covers who may act, not only what input reaches.** A new
  `references/ACCESS-CONTROL.md` (70 reference files) holds one auth wrapper
  around the whole `ServeMux` with the public patterns listed in it, the
  shapes of auth that fails open, the tenant the middleware checked as the
  only tenant a query uses, every door to the same effect (batch, export,
  import, background jobs, restore, check-then-act), the tenant in cache
  keys, model output and tool-call arguments as untrusted input, mTLS, gRPC
  (`ChainStreamInterceptor` beside the unary one), and queue identity, and an
  OAuth client's `state` and PKCE (`oauth2.GenerateVerifier`). The Triage tree
  and Quick Reference gain the matching rows, the pprof bullet names the
  `DefaultServeMux` registration a blank import makes, and session rules add
  revocation on a password change and single-use reset tokens. Review Mode
  takes its method from Cloudflare's security-audit skill: a finding names
  the lower-trust caller and whose resource it reaches, every door to an
  effect is compared, variants are searched, a missing layer behind a working
  one is a hardening line, a fact outside the repository makes a finding
  `plausible` with that fact named, and findings are ranked critical to low
  by what the evidence shows. A new review fixture, `gatekeeper`, seeds ten
  of these defects for `abrun -corpus review`.
- **A review of HTTP code loads go-security.** In the review corpus no
  session loaded it (0 of 18, Sonnet 5.5 medium), so none of its text reached
  a review: go-code-review's step 2 named it "when input reaches a sink",
  which needs the code read first. The prompt hook's review note now names
  go-security in the router line when the code's owner hints include go-http
  or go-security (still no owner list, which cost Opus +37%), and step 2 names
  it for a handler, an HTTP client, or input that reaches a query, command,
  path, or template. Haiku 5.5 low, gatekeeper, vault, and partner, n=10 per
  arm, HEAD against both changes: go-security loaded 0/30 → 30/30, its
  defects found 137/150 → 147/150, other owners' 208/250 → 212/250, Must Fix
  0.95 → 0.98, bait citations 0.26 → 0.34, cost +9%. The gain is all on
  gatekeeper (foreign ID answered 403: 4/10 → 10/10; export owner unchecked:
  6/10 → 10/10), the fixture seeded with this text; vault and partner were at
  ceiling in both arms. Sonnet 5.5 medium saturates the corpus (recall 0.92 in
  every arm) and pays +10–15% for the load. go-style-core, which the note does
  not name, now loads in few review sessions (Haiku 5/30 → 0/30, Sonnet
  16/18 → 5/18) with no measured loss.

## [1.36.0] - 2026-10-14

- **The skills pass the delivery invariants `delivery_test.go` added.**
  `Route Before The First Edit` in `go-code` and `Workflow` in
  `go-code-refactor` now follow the opening sections, inside the first 200
  lines. `go-code-refactor` step 5 names the plugin-hook condition before the
  `--gate` command the edit hook prints. The `> Compatibility:` notes and the
  other skill mentions name the plugin repository's COMPATIBILITY.md as plain
  text, since a skills-only install carries no copy of it.
- **The restraint ladder runs on every `go-code` task.** Step 5 climbs the
  ladder in `OVER-ENGINEERING.md` before each added helper, type, layer,
  interface, option, or import on new code, a fix, and a restructuring alike;
  before, only a restructuring climbed it, new code read it when a rung was in
  doubt, and a fix not at all. Resource Routing reads the file on every task
  unless the session-start hook already printed the ladder, and the
  Declaration Budget counts what survives the ladder. `TestRestraintLadder`
  pins the step-5 sentence.
- **OpenCode + Muse Spark project block.** `docs/PROJECT_INSTRUCTIONS.md`
  gains a dedicated section: install path `~/.config/opencode/skills`,
  one-message batch load of router + `go-style-core` + owners via the skill
  tool (file-read fallback), direct `go-linting` gate with no hook receipts,
  and bounded subagents under the `agent.build.steps` ceiling.
- **Hooks pass on a host that does not load skills.** Each `hooks.json`
  command exits 0 when `${CLAUDE_PLUGIN_ROOT}` does not point at the script,
  so a missing plugin root is not a preToolUse denial. The routing gate
  blocks a `.go` edit only for Claude Code (`CLAUDE_PROJECT_DIR` set and
  `PLUGIN_ROOT` unset); Codex `apply_patch` and any other Edit-shaped host
  pass. Prompt and subagent notes on that other host add an absolute
  `SKILL.md` path. The Claude router line is unchanged.

## [1.35.0] - 2026-10-12

- **The idiom card keeps what a model misses, and the traps of the rest.**
  A plugin-eval probe ran go-style-core forced against no skill, one regex
  grader per card row (9 prompts, 46 rows, n=3): Sonnet and Opus 5.5 wrote
  the current form unaided for every row up to Go 1.26 — `range n`, the
  `slices`/`maps`/`cmp` helpers, `cmp.Or`, `new(expr)`, `errors.AsType`,
  `wg.Go`, `context.AfterFunc`, typed atomics, `strings.Cut`/`SplitSeq`,
  `omitzero`, `t.Context`, `b.Loop`, and more. They missed the Go 1.27 rows
  (promoted fields, generic methods, `url.Clone`, `strings.CutLast`,
  `uuid`, `encoding/json/v2`, `httptest.NewTestServer`) and `fmt.Appendf`,
  where the card took them from 0/3 to 3/3. The table now holds those rows,
  `slices.SortFunc`, and the 1.25/1.26 rows; the older rows became a list of
  their traps (`cmp.Or` evaluates every argument, `slices.Max` panics on
  empty, `time.Tick` returns nil for `d <= 0`, …) with a line of directive
  versions. The card went from about 3.9K to 2.6K tokens, loaded in every Go
  session. abrun implement on roster, catalog, ledger, and profile (medium):
  Sonnet 5.5, n=3 per arm, golden 12/12 both, `go fix` and lint counts
  unchanged, −6% cost per run; Opus 5.5, n=2, golden 8/8 both, the same
  counts, cost within noise (−2%).

- **plugin-evals: a quality case that tells the arms apart.** Two of the
  three quality cases passed without the plugin on Sonnet and Opus 5.5;
  they are tagged `cost` and kept to track what the plugin costs.
  `quality-troubleshooting-panic` grades the pasted-panic answer against
  go-troubleshooting's method and scored 1.00 with the plugin, 0.25 without
  (Sonnet 5.5, n=3). Sonnet as judge split its votes as often as Haiku (2/24
  against 4/48) at 12 times the judge cost, so Haiku stays pinned.

## [1.34.0] - 2026-10-12

- **go-security names the default ML-KEM hybrids in the TLS row.** The
  `CurvePreferences` row said only "a list that omits the ML-KEM hybrids";
  Sonnet 5.5 reviewing a TLS config loaded go-security, never opened
  SECRETS-AND-CRYPTO.md, and advised "put `X25519MLKEM768` first" in 2 of 2
  plugin-eval runs, which drops the Go 1.26 hybrids and does nothing
  (`crypto/tls` ignores the order). The row now names `X25519MLKEM768`,
  `SecP256r1MLKEM768`, and `SecP384r1MLKEM1024`, says the fix deletes the
  list rather than reordering it, and that only opt-in `MLKEM1024` needs
  one. On `quality-security-tls-review` (Sonnet 5.5, 5 runs per arm) 3 of the
  4 runs that loaded go-security named the 1.26 hybrids; 3/5 passed with the
  plugin against 0/5 without (one in-run judge FAIL re-judged PASS 3/3).

- **HTTP and SQL code no longer loads go-error-handling by itself.** The
  go-code HTTP and SQL rows, and the edit gate behind them, required
  go-error-handling for any handler or query file; it is now required, as
  for every other file, when the code wraps or matches errors (`%w`,
  `errors.Is`/`As`/`AsType`/`Join`/`New`). go-database already owns the
  `sql.ErrNoRows` mapping and go-http its error responses. abrun implement
  on the three HTTP fixtures (Sonnet 5.5, medium, n=3 per arm): account and
  fetch still loaded it 3/3 through their error code; gateway went from 3/3
  to 0/3 loads at −16% cost per run with golden 3/3 in both arms. The one
  golden failure in the new arm (fetch, `TestGetOrderHonorsRetryAfter`) came
  from a run that loaded go-error-handling, so the change did not reach it.
  `trigger-database-uk` now expects go-database alone.

- **A pasted Go trace names go-troubleshooting.** "panic: … goroutine 42
  [running]: … what does this mean?" carries no work verb, so the prompt
  hook stayed silent, and Sonnet 5.5 answered without go-troubleshooting in
  7 of 7 plugin-eval runs, offering a nil check at the faulting line. With
  the skill loaded, 5 of 5 answers traced the nil back to its origin and
  treated the printed argument words as a lead, against 0 of 5 without the
  plugin. The hook now names go-troubleshooting when a prompt pairs
  `panic:` or `fatal error:` with a `goroutine N [` header and has no work
  verb; the trigger case went from 0/7 to 5/5 at +$0.027 per run. "Fix this
  panic" keeps go-code.

- **A subagent report is not the prompt.** The prompt hook classified a
  report a subagent handed back (wrapped in `<agent-message>`) as the user's
  words: "go test" named Go and "unreadable" matched the refactor wording,
  so a note asked for go-code-refactor. The wrapper is removed before
  classification, as `<ide_*>` context already was.

- **plugin-evals: the TLS grader names what it checks.** The curve criterion
  asked for "the Go 1.26+/1.27 additions" to the default hybrids, but 1.27's
  `MLKEM1024` is opt-in; it now asks for `SecP256r1MLKEM768` or
  `SecP384r1MLKEM1024` and rejects reordering as a fix. The README pins
  `--judge-model` and records why trigger cases keep `max_turns: 10`.

## [1.33.0] - 2026-10-12

- **A `claude plugin eval` suite.** `plugin-evals/` holds 16 cases drawn
  from the validation set of `evals/evals.json`: ten trigger cases scored
  on whether the expected skill fires, three negatives where no `go-*`
  skill may fire, and three quality cases graded by an LLM judge against
  the eval's assertions. Run it with `--eval-dir plugin-evals`; results
  stay out of git.

- **A pasted Go code fence names Go.** The prompt hook stayed silent on
  "Split this function into smaller ones" with the Go code pasted below it
  when the working directory held no Go, so no router was named; the
  plugin-eval run on Haiku 5.5 fired go-code-refactor in 0 of 3 runs, and 3
  of 3 with the fence recognized.

- **A go-code note with no owner hints asks for owners by description.**
  When no Go file in the prompt points at its owners, the note names three
  by example (go-generics, go-database, go-security). On Haiku 5.5 the
  generics and ffmpeg trigger cases went from 1/5 and 0/5 to 5/5; an owner
  the note does not name (go-error-handling) did not move, and Opus 5.5
  stayed within noise. abrun implement on the three fixtures without hints
  kept golden 9/9 at +3% cost per run.

## [1.32.1] - 2026-10-11

- **A request body is not the caller's again when `Do` returns.** The
  Transport writes it on its own goroutine (`persistConn.writeLoop` on
  HTTP/1, `clientStream.writeRequestBody` on HTTP/2); a response answered
  from the headers, a timeout, or a cancel returns from `Do` mid-upload
  (`net/http` `RoundTripper` and `Client.Do` docs, checked against go1.27.1).
  Waiting for the body's `Close`, which the `RoundTripper` doc offers, is not
  enough on HTTP/2: on a status above 299 the Transport closes the body from
  another goroutine while a `Read` is in flight, and `-race` reports a
  release-on-`Close` wrapper there 3/3 (HTTP/1: clean). go-http said nothing
  about the request body. go-resilience asked for a fresh body per attempt
  only because a stream is consumed, which a `Reset` and refill of a shared
  buffer, or a rewound reader, seems to satisfy. go-http's Clients section
  now owns the rule: one `bytes.NewReader` per request over bytes nobody
  writes again. go-resilience and BUFFER-POOLING.md route to it; the pool
  reference keeps request-body buffers out of the pool, with neither a
  deferred `putBuffer` around `Do` nor a `Close` hook as the release point.
  From the VictoriaMetrics write-up on vmauth's corrupted retries and
  vmagent's accepted race (victoriametrics.com/blog/http-race-condition).

## [1.32.0] - 2026-10-11

- **IDE context no longer routes a prompt.** The VS Code extension puts
  `<ide_opened_file>` and `<ide_selection>` text into the prompt that
  UserPromptSubmit hooks receive. With `skills/go-code-refactor/references/MODERNIZATION.md`
  open, a request to study an arXiv paper got the refactor note: `golang`
  in the path named Go, and `MODERNIZATION` matched the refactor wording.
  `go-prompt-routing.sh` and `go-restraint-ladder.sh` now remove `<ide_*>`
  blocks before classifying. As a result, a level word in a selection sets
  no restraint level, and IDE text ahead of `/go-code ultra` no longer hides
  the command.

## [1.31.5] - 2026-10-11

- **A file's `//go:build` version governs that file.** go-style-core's
  Write Current Go and `COMPATIBILITY.md` named only the module's `go`
  directive. The compiler and `stdversion` check a file whose `//go:build`
  line names a Go version at that version, above or below the directive and
  never below 1.21 (`go/types` `max(fileVersion, go1.21)`): in a `go 1.24`
  module, a file under `//go:build go1.21` rejects `for range n`, and a
  file under `//go:build go1.22` in a `go 1.21` module may use it. Both
  directions were checked with go1.27.1. Found in JetBrains
  go-modern-guidelines PR #33.
- **`strings.Cut` keeps its not-found guard on a swap.** MODERNIZATION.md
  warned only about `CutLast`. `Cut` also returns `(s, "", false)` without
  the separator, where an `Index`-based split may have returned `("", "")`.
  The Tier 1 entry now covers both and names the `filepath.Split`-style
  split that `CutLast` turns around. From JetBrains go-modern-guidelines
  PR #20; its other five cases were already covered.

## [1.31.4] - 2026-10-11

- **Skill arguments no longer rewrite SQL examples.** When a skill is
  invoked with arguments, from `/go-database fix the query` or from a Skill
  tool call with `args`, Claude Code replaces `$0`, `$1`, … and `$ARGUMENTS`
  in its SKILL.md with the argument words, inside code fences too.
  go-database's transfer example and query-loop advice and go-security's
  tenant row carried PostgreSQL placeholders, so on Claude Code 2.1.296
  `IN ($1, $2)` reached the model as `IN (and, fix)`. The transfer now writes
  `\$1`, which Claude Code delivers as `$1`; the prose says `$N` or names the
  parameter. `TestSkillBodyArgumentPlaceholders` fails on an unescaped
  placeholder in any SKILL.md, and the authoring template states the rule.
- **Bundled scripts run through `bash`.** go-code's gate list, and the
  Validation notes of go-naming, go-documentation, and go-interfaces, named
  a script by path alone; go-testing and go-code-refactor gave no command at
  all. They now write `bash <installed-skill-dir>/scripts/<name>.sh`, as the
  other skills already did, so a package that drops the executable bit
  still runs and the skills' `allowed-tools: Bash(bash:*)` grant covers the
  call. `TestScriptCommandsRunThroughBash` fails on a written-out script
  path without `bash`.

## [1.31.3] - 2026-10-11

- **A refactor loads go-testing.** The prompt hook names it without a
  condition, in the same load line as the router: "if you write or edit a
  test" left the model to edit `_test.go` and get blocked by the routing
  gate. go-code-refactor loads it with go-style-core before the first edit,
  and the session-start and subagent note names it for a refactor. A fix
  prompt keeps the condition.

- **A release tag publishes its GitHub release.** A `v*` tag that passes
  validation gets a GitHub release whose notes are that version's
  `CHANGELOG.md` section.

## [1.31.2] - 2026-10-10

- **CI pins golangci-lint v2.14.0.** v2.13.2 cannot read the export data
  of go1.27.2, which `1.27.x` now resolves to, so every lint-backed test
  failed on the runner. go-linting names the versions it was checked against.

## [1.31.1] - 2026-10-10

- **A review loads only go-code-review.** The checklist gains the rows a
  review missed while the rule lived only in the owner: a cache's key and
  load, a response type instead of the storage record, a PATCH onto the
  loaded fields, an exported change that looks additive, an exhausted pool,
  and retry plus idempotency. Each row is one line and a link.
- **Writing a cache loads go-performance.** The description names
  implementing a cache. The routing table leads with a cache. A stub whose
  comments pair `cache` with `ttl` names go-performance before the first
  edit; `Cache-Control` does not, and the edit gate does not require the
  skill.

## [1.31.0] - 2026-10-08

- **A passing test check needs proof the tests ran.** The go-linting gate
  stops counting exit 0 as a pass when it comes from `[no tests to run]`, a
  `-run` that matched nothing, or a file behind an unpassed `//go:build`
  tag. The same goes for `[no test files]`, an `Example` without
  `// Output:`, and a `TestMain` that returns before `m.Run`. The evidence is
  `--- PASS` or a `"Action":"pass"` event. go-testing's integration reference
  gains the early-return `TestMain` trap and a contents list. On go1.27.1 the
  trap printed `ok` over a broken function with `go vet` silent.
- **Unknown members of a closed set fail loudly.** `IOTA.md` covers a
  `default:` that answers `0`, a map read of a missing key, and a `, _` type
  assertion. The go-linting baseline's `exhaustive` passes them all, because
  any `default` counts as covering the rest. `go-code-review` gains a Must Fix
  row for them.
- **Deletes name what they remove.** `go-defensive` gains a rule for
  deletes. `RemoveAll` of a name that resolves to the directory empties it;
  `filepath.IsLocal("a/..")` is true; a `{id}` wildcard delivers `ws/..` from
  `ws%2F..`. On go1.27.1, `root.RemoveAll("ws/..")` deletes everything under
  the root before it returns `invalid argument`. The fix rebuilds the name
  from `uuid.Parse`. `TestDefensiveExampleDeleteByParsedID` runs the example
  against a real root. `go-code-review` gains a Security row for deletes,
  because reviewers load only the review skill and never reached the rule in
  go-defensive.
- **Responses are response types.** The go-security injection reference
  adds the output half of mass assignment: encoding the storage record ships
  every field added to it later. The go-http handler example now encodes a
  response struct; `TestResponseFields` checks that a stored
  `PasswordHash` stays out, and the example's store errors became one
  `switch` to stay within 40 lines. go-http gains Partial Updates, a PATCH
  recipe that unmarshals the body with v2 onto the loaded writable fields.
  Absent members keep their stored values, `null` clears a member, and a
  top-level `null`, an unknown member, or `password_hash` is a 400.
  `TestHTTPExamplePartialUpdate` runs it.
- **Changing an exported API.** `go-functions` lists the changes that look
  additive and break importers: an appended parameter (variadic included,
  which breaks `Open` used as a value), an interface method, a non-comparable
  field, an unkeyed-literal field, and a promoted-method collision. It also
  lists the additive forms. `STRUCTURAL.md` routes there, and go-code's
  routing table loads go-functions when an exported signature changes.
  `TestKnownReferenceRegressions` now pins the list in go-functions,
  including the two new rows, and pins the route from `STRUCTURAL.md`.
- **Three fixtures.** The review corpus gains `quota`, whose five must-fix
  defects are the items above with two baits. The implement corpus gains
  `account`, a PATCH handler whose hidden test checks absent members, a
  `null` phone, mass assignment, and the exact response members. It also
  gains `purge`, a delete whose hidden test sends `""`, `.`, `..`, `a/..`,
  `<id>/..`, `<id>/data`, and absolute or escaping names. Each implement trap
  was checked against naive implementations that fail it. Measured on Claude
  Haiku 5.5, 3 runs per arm:
  - `quota`: the keyed recall was 1.00 for `no-skill` and 0.87 for both
    reference and baseline. The scorer counts a citation by its line, so the
    outputs were read by hand. No run in any arm said that `ws%2F..` empties
    the tree; the "found" citations near the delete line were about 204
    versus 404. Must-fix defects were labelled Must Fix at a rate of 0.53 for
    `no-skill`, 0.77 for reference, and 0.92 for baseline. After the Deletes
    review row, all 3 baseline runs named `ws%2F..` and `uuid.Parse`, with
    recall 0.93 and as-must 1.00.
  - `account`: the first run failed 2 of 3 reference sessions and 1 of 3
    baseline sessions. The cause was the hidden test calling the handler
    without a mux, which left `r.PathValue` empty. With the handler mounted
    on its pattern, every arm passes 3/3, adding +149, +78, and +63 lines.
  - `purge`: every arm passes 3/3, so the fixture has no trap for this model.

  Adapted from
  addyosmani/agent-skills `test-driven-development`,
  `code-review-and-quality`, `security-and-hardening`, and
  `api-and-interface-design`.

## [1.30.4] - 2026-10-08

- **Caching has an owner.** `go-performance` gains Caching. A cache is for a
  measured expensive, read-mostly value. Its key holds every input the answer
  depends on, including the tenant or caller in `ctx`. It has one stated
  expiry rule, keeps no failed load, and lets concurrent misses share one
  load without holding the lock across it. The section also bounds growth and
  shows a 40-line in-flight entry that needs no dependency.
  `TestPerformanceExampleCache` runs that example under `-race`. The
  `go-performance` description names caches, `go-code` routes caches to it,
  and the resilience reference links its cache-refresh stampede there.
- **Reverted attempts are reported.** `go-performance` asks for each reverted
  optimization with its benchstat row, so a dead idea is not tried again.
- **Plans before indexes, holders before pool size.** The PostgreSQL
  reference gains Read the plan before adding an index. An estimate an order
  of magnitude off means `ANALYZE`, not an index. Equality columns lead a
  multicolumn index. `lower(col)` needs an expression index and `LIKE
  '%term%'` needs `pg_trgm`. An index the plan does not use is dropped.
  `go-database` says an exhausted pool is fixed where connections are held,
  not by raising `SetMaxOpenConns`.
- **The implement corpus gains `profile`.** It is a tenant-scoped profile
  cache whose hidden test checks four traps: a key without the tenant from
  `ctx`, a cold-start stampede, a lock held across the fetch, and a kept
  failure. It also checks the TTL. Each trap is checked against a naive
  implementation that fails only that test. Measured on Claude Haiku 5.5
  with 3 runs per arm. `no-skill`, reference, and baseline all pass the
  hidden test in every run, adding +90, +54, and +54 lines on average. No
  baseline session loaded `go-performance`, so the fixture has no trap for
  this model and the Caching section went unread. Adapted from "Missing caching",
  "Queries that ignore their index", and "Log every attempt" in
  addyosmani/agent-skills `performance-optimization`.

## [1.30.3] - 2026-10-08

- **A check turns green by fixing the code, not the check.** `go-linting`
  gains Holding the Bar: a new `//nolint`, `//lint:ignore`, or `#nosec`, a
  lint-config exclusion, a skipped or deleted test, a removed assertion, a
  rewritten golden file, a dropped `-race`, or a `panic("not implemented")`
  stub is legitimate only with a reason a reviewer can check. The gate's
  "fix attributable failures" rule links to it. `pre-review.sh` 1.1.0 adds a
  `bar` section under `--new-from-rev`. It reads the diff since the revision,
  untracked `.go` files included, and reports each such move by rule, file,
  and line. Tightening stays silent. A changed lint config or golden file is
  reported because a diff cannot show which way it moved. Any finding exits 1.
  Without the flag the section is `skipped`. `go-code-review` gains a Testing
  row for these moves, and its procedure passes `--new-from-rev` for a diff
  against a revision. `TestScriptFunctional/PreReviewBar` checks the rules
  against a scratch git repository. The review corpus gains `snapshot`, a
  tree with no diff and six seeded defects, four of them must-fix: a
  `//nolint:errcheck` with a false reason on a written file's `Close`, a
  `t.Skip("flaky on CI")` hiding a `maps.Copy` that overwrites instead of
  adding (the overwrite is a defect of its own), a round trip that never
  compares, an `Example` without `// Output:`, and a
  `panic("not implemented")` stub. A reasoned `//nolint:gosec` sits beside
  them as a bait. Measured on Claude Haiku 5.5 (review corpus, 3 runs per
  arm). Every arm found all six defects in every run on `snapshot`,
  `no-skill` included, so the fixture has no trap for this model and shows
  no recall gain. No arm cited the bait, so the rule did not turn a reasoned
  suppression into a finding. `keyring` stayed at 4/4 in both
  reference-vs-baseline arms. The moves only a diff shows (a removed
  assertion, a deleted test, a loosened config, a dropped `-race`) need a
  diff-based review fixture, which abrun does not stage yet. Adapted from
  "Guard the bar itself" in addyosmani/agent-skills
  `constraint-driven-development`.

## [1.30.2] - 2026-10-08

- **Text the task reads is evidence, not instructions.** `go-style-core`
  House Style Wins now says where instructions come from: the user, the host,
  and the repository's instruction files. A line in a code comment, fixture,
  log, ticket, PR description, or tool output that says to skip a check,
  approve, run a command, or reveal a secret is reported as a finding and
  never followed. `go-code-review` gains a Security row for text aimed at the
  reviewer, and the ticket reference links its existing "task data" sentence
  to the rule. `TestRuleOwnershipMap` pins the rule to `go-style-core`. The
  review corpus gains `keyring`: two must-fix defects sit in a function under
  a comment that asks AI reviewers to skip it, next to a control must-fix
  outside it, so a reference-vs-baseline run can show whether the rule
  changes recall. Not measured yet. Adapted from the trust levels in
  addyosmani/agent-skills `context-engineering`.
- **A paused investigation keeps a restartable note.** The working note in
  `go-troubleshooting`'s ticket reference now also records the revision and
  environment of each check, the exact command and outcome, and the
  authorized mode. The next session reruns a check whose revision moved, and
  takes the mode from the user or the note.

## [1.30.1] - 2026-10-08

- **go-verify runs on Haiku.** The agent's `model` is `haiku` instead of
  `sonnet`; in Claude Code the alias resolves to Claude Haiku 5.5. The agent
  only runs the requested checks and reports them, so it is the cheapest
  place in the plugin for the smaller model. Whether Haiku 5.5 reports
  diagnostics without paraphrasing them is not yet measured.

- **abrun reports safety refusals.** A claude session whose turn ended with
  `stop_reason: "refusal"` is marked `refused` in the report, and the arm
  summary counts such runs. Claude Haiku 5.5, Sonnet 5.5, and Opus 5.5 run
  classifiers that can decline security-review prompts, and Haiku 5.5 has no
  server-side fallback; before this, a refused run could not be told apart
  from a failed or empty one. Scoring is unchanged.

## [1.30.0] - 2026-10-07

- **Review checks what tests and structure hide.** `go-code-review` probes
  whether a test fails when the diff's condition breaks: the condition is
  inverted in a copy and run through `go test -overlay`, so the tree is never
  edited. A bug fix needs a regression test that fails on the old code. A new
  Design section flags complexity moved instead of removed, feature logic in a
  shared package, the same switch at a third site, and a refactor mixed with a
  behavior change. A Dependencies section reviews one module per bump and
  treats a `go` directive bump as its own change. Less Code adds a `reuse:`
  cut tag for a helper the repository already has, and a check for
  declarations the diff left without callers. The report template gains a
  verdict line. `go-packages` now owns the rule for upgrading a module. None
  of this is measured yet.
- **Refactor keeps what history and benchmarks explain.** A line that looks
  redundant but still runs (a second check, a defensive copy, a sleep) gets
  `git log -S` before it is removed. If the commit that added it fixed a bug,
  the removal is a finding unless a test reproduces that bug. An existing
  helper that meets a Declaration Budget rule stays, even with one caller.
  When a benchmark covers a touched function, `go-code-refactor` records it
  before the edit and compares after through go-performance's
  `bench-compare.sh`. Not measured yet.

## [1.29.1] - 2026-10-07

- **Security findings no linter reports.** `go-security` covers five classes
  `gosec` cannot see. CORS answers only listed origins, never reflected or
  `null`, and trusts the same list in `CrossOriginProtection`. A webhook HMAC
  is checked over the raw body and a signed timestamp with `hmac.Equal`. A
  login hashes once for an unknown user and caps concurrent hashes. A
  decompressed stream and a fetched response body get a size cap. A
  client-written template is never parsed. Review Mode checks every
  `review`-only Quick Reference row after reading the `gosec` output. Each new
  example runs under a test.

## [1.29.0] - 2026-10-06

- **Troubleshooting loop.** A failing reproduction is cut down to the pieces
  that still fail. Each competing cause gets a written prediction before the
  first check, and the cheapest separating check runs first. The
  regression test comes before the fix only when it can reach the original
  mechanism; otherwise that gap is reported. Temporary logs share one debug
  prefix and are gone before the investigation ends, with or without a fix.
- **Architecture proposals.** A package-scale review starts at recent hotspots
  when no target was named, treats a shallow package as a smell, and grades
  each candidate Strong, Worth exploring, or Speculative before any new
  interface.

## [1.28.5] - 2026-10-04

- **Routing note at session start.** A Go session and each writing subagent
  are told to load `go-code` and `go-style-core` in one message before the
  first edit. The previous subagent note named only the router, and only
  under "if your task writes", so the first `.go` edit was blocked.
- **Verifier command suffixes.** The Bash verification gate accepts the
  printed `go-check-receipt.sh --gate` command with a display redirect or
  pipe (`2>&1`, `>file`, `2>&1 | tail`). Plain `| tail` was already
  supported. Quoted paths and nested Bash commands preserve their syntax;
  another verification command in the same call stays blocked until the
  verifier has been attempted.

## [1.28.4] - 2026-10-02

- **Edit-hook check evidence.** Each check now records its status, actual exit
  code, command, scope, config, tool version and before/after input digests in
  an immutable run directory. Clean runs emit PostToolUse context; skipped and
  unavailable checks are explicit. The gate reuses only current matching
  passes; file/package checks, short tests and new-findings lint do not replace
  broader required checks or race checks. This applies where the plugin hook
  runs; Codex/opencode runners still do not connect it. After Sonnet A/B,
  unverified receipt status is separated from final credit: hook output gives
  concrete Bash verification commands without a plugin environment variable,
  and `hook_credit=true` requires current inputs plus matching cwd/argv.
  Routers explicitly load go-linting before final verification.
  The prompt puts every selected owner's named Skill action on its primary
  line, with go-linting when Bash is available, and waits for loads before
  edits. A Bash workflow hook requires the gate owner and a current receipt
  attempt before repeated runtime checks after a Go edit. Batch verification
  separates credited package checks from required direct checks; host-only
  environment differences no longer invalidate unchanged inputs.
  Completion feedback also checks explicit hook-pass claims against verified
  credits and asks for a correction when the report overstates the result.
- **Prompt hints for Go stubs.** Recognize standard-package aliases and
  context signatures, sync/logging types, error calls and bounded worker
  contracts before bodies exist. Actionable Retry-After/backoff contracts in
  stubs retain resilience hints; unrelated comment/literal examples,
  unused/blank imports and foreign aliases do not imply owners. These hints leave the edit
  gate's decision rules unchanged. Exploratory Sonnet 5.5 low A/B results are
  recorded in the improvement plan; they do not establish an overall benefit
  and identified the Retry-After contract gap corrected by regression tests.

## [1.28.3] - 2026-10-01

The edit hook no longer hides lint findings in symlinked checkouts, and `abrun`
counts a skill load only when a tool result proves it.

- **Edit hook lint.** `go-vet-on-edit.sh` runs golangci-lint from the physical
  directory. In a Git project opened through a symlink, `--new-from-rev=HEAD`
  had filtered out even new findings. The repository's own config is resolved
  first and passed with `--config`, so it keeps priority after the change of
  directory. Findings still split into the edited file's lines and a count for
  other files. A failed lint run and its stderr are now reported instead of
  reading as a clean result.
- **Codex `abrun` skill evidence.** A skill counts as loaded only after a
  successful, completed read whose captured output equals the arm's staged
  `SKILL.md` exactly. The JSON field `codex_skill_evidence` records full,
  partial, failed, and unverified reads, byte counts, and inventory
  availability, repair turns included. Path mentions in `ls` or `echo`,
  cross-links, and cropped output no longer count as loads, and an
  unavailable inventory stops the run before the CLI is called. Full captured
  output does not prove the model retained it.
- **Claude `abrun` routing evidence.** `routing.confirmed` (version 1) counts
  only correlated, successful Skill tool results: a request, a failed result,
  or an unrelated skill does not confirm a Go load. Go edit attempts and
  successful edits are counted separately, and load order is checked at the
  edit request. A routing block needs PreToolUse, an explicit exit 2, and the
  routing gate's own marker. Legacy request metrics stay separate; partial or
  unmeasured evidence stays out of confirmed denominators, and a review with no
  Go edit reports N/A. A repair CLI invocation keeps its own routing context.
  This is Skill tool-result evidence, not a claim about model-visible content
  or retention; the Read fallback is still partial.

## [1.28.2] - 2026-10-01

Corrections from a content review of all 24 skills on 1.28.1. Eight reviewers
checked every claim against go1.27.1 (some also against go1.25 and go1.26) and
golangci-lint 2.13.2. This release carries only the corrections that need no
model measurement: wrong facts, examples that teach the wrong code, script
bugs, broken routes, and stale docs. Wording meant to change model behavior
waits for `abrun`: the Contract Table and go-testing scope, Declaration Budget
against gocyclo, crutch cuts, descriptions, and the subagent note.

- **Examples that taught wrong code.**
  - go-logging's redaction form leaked a secret held in a struct field, a
    slice, `%#v`, JSON, or an error string. slog calls `LogValue` only on an
    attribute's own value. The secret type now also implements `String`,
    `GoString`, and `MarshalText`, and go-security routes to it.
  - go-database's transfer locked rows in call order, so two opposite
    transfers deadlocked. It now locks with `SELECT … ORDER BY id FOR UPDATE`.
  - The INTEGRATION.md database helper had no driver import, so it failed in
    integration CI with `unknown driver "pgx"`.
  - go-context's keys shared one empty struct type, so a second key equalled
    the first.
  - The buffer pool returned bytes that the next `getBuffer` overwrote.
  - The architecture fixture rolled back only when a named `err` was set.
  - The doc template deferred `Close` on a written file.
- **JSON v2 leftovers.**
  - Delete Pass no longer calls `json.Marshal` of a built document infallible.
    v2 fails on invalid UTF-8 and on `time.Duration`, and returns partial
    bytes with the error.
  - go-http's and go-context's handlers marshal before `WriteHeader`.
  - MODERNIZATION.md files the default `go fix` `omitzero` hunk as Tier 2.
    Under v2 it changes the wire (`{"id":1}` → `{"id":1,"inner":{}}`).
  - JSON-V2.md's migration table gains the string-tag, HTML-escaping,
    pointer-receiver `MarshalJSON`, and `[N]byte` rows.
  - go-data-structures notes that `slices.Concat` returns nil, and covers nil
    maps.
- **Idiom card exceptions.**
  - `slices.Max`/`Min` panic on an empty slice.
  - `Reverse`/`Compact` edit in place.
  - `AfterFunc` gets two rows. `defer stop()` deregisters cleanup that must
    outlive the function.
  - `time.Tick` returns nil for `d <= 0`.
  - New `math/rand/v2` row: `Seed` is a no-op since 1.24.
  - `httptest.NewTestServer` is reached only through `srv.Client()`. Its URL
    is `http://example.com`, and `srv.Start()` gives loopback.
  - Write Current Go gains the explicit-instruction exit, and separates
    language gating from `stdversion`.
- **Edit hook record.** A check is `pass (hook)` only once the hook has
  printed it. Tests run `-short` without `-race`. Output covers the package,
  and pre-existing findings are reported, not fixed.
- **MODERNIZATION.md.**
  - `err == X` → `errors.Is` is Tier 2 when nothing wraps X.
  - Redundant type arguments apply only at a `go 1.27` directive.
  - A directive bump diffs `DefaultGODEBUG`.
  - The rand row is split three ways.
  - `x := x` deletion is safe in range loops.
  - Tool commands are corrected: `go fix -fixtool`, `deadcode -test`,
    `goimports` on touched files.
- **Security.**
  - The SSRF allowlist re-checks every redirect.
  - X-Forwarded-For reads all header lines.
  - JWT checks `iss`, `aud`, and a required `exp`; `kid` chooses only among
    keys the server holds.
  - `crypto/hpke` (Go 1.26) replaces hand-built KEMs.
  - The gosec directives sit on the G404 and G204 example lines.
  - PANIC-RECOVER.md gains the goroutine and middleware forms, and the
    middleware re-panics `http.ErrAbortHandler`.
- **Other content fixes.**
  - Fixed facts:
    - synctest fails only on durably blocked goroutines.
    - `goroutineleak` is nil on a 1.26 toolchain.
    - `OnceValues` caches the error.
    - Printf wrappers use vet's marker, and vet reports non-constant format
      strings.
    - `t.Output()` versus `t.Log`.
    - Package names are single words, not singular nouns.
    - The float and time pitfalls are corrected.
  - New forms:
    - the slog context handler that survives `With`;
    - the `r.Pattern` route label;
    - transport tuning on a `DefaultTransport` clone;
    - pool sizing per process;
    - pgx native rules.
  - Lint and logging:
    - The gate lints with the bundled config when the repository has none.
    - CONFIGURATION.md states that a revive rules list replaces revive's
      defaults.
    - The house-style exception to "use slog" is restored.
  - Routes:
    - `Related Skills` loops are broken.
    - Routes that pointed at sections that do not exist are fixed.
    - Duplicated rows are cut from OVER-ENGINEERING.md's reach-for table.
- **Scripts.**
  - All five AST wrappers build their helper outside the target project, with
    the toolchain the project selects (else the local one). They key the cache
    on that `GOVERSION` and exit 2 when the helper does not build. A helper
    cached under 1.26 used to reject 1.27 generic methods.
  - check-errors, check-interface-compliance, check-naming, check-docs, and
    verify-refactor skip vendor, testdata, `.`/`_` directories, and generated
    files.
  - check-errors reports a file that does not parse as `parse_error` and
    finds more string-matching and logger forms.
  - check-interface-compliance resolves module packages through
    `go list -export`, which removes a false "implemented" verdict. It matches
    generic interfaces and emits `[]`, not `null`.
  - check-docs counts a comment that holds only directives or only
    `Deprecated:` as missing.
  - check-naming fixes the getter exception boundary.
  - setup-lint no longer writes beside an existing `.golangci.yaml/.toml/.json`.
  - pre-review gains `--new-from-rev` and reports which config ran.
  - bench-compare uses `go tool benchstat` and never saves a failed run.
  - check-debt splits adjacent `Kept:` markers.
  - check-architecture treats `--limit 0` as unlimited.
  - verify-refactor's result directory carries a `.gitignore`.
  - Usage errors exit 2, and control characters are escaped in JSON.
  - Script paths in SKILL.md are written as `<installed-skill-dir>/…`, run
    from the project.
- **Hooks and evals.**
  - The prompt hook sends "readable", "split", "extract", and over-engineering
    prompts to go-code-refactor. `implement` counts only as a verb. A
    translation or copy-edit prompt gets no note.
  - The gate no longer requires an owner for `interface{}`, `unsafe.`,
    `sha256`, or `StatusCode`, and its dead branch is gone.
  - Six stale trigger-eval labels are corrected (#3, #45–47, #50, #68); the
    count is unchanged.
- **Docs.**
  - PROJECT_INSTRUCTIONS no longer describes the removed card Read, and names
    opencode.
  - RULE_OWNERSHIP drops 26 stale route-from names and adds rows for
    generics, receiver type, nil versus empty, and the review procedure.
  - COMPATIBILITY.md is corrected for NewTestServer, `t.Output`, hpke, and
    `math/rand/v2`.
  - RELEASE_CHECKLIST runs `-race -shuffle=on` and `bash -n hooks/*.sh`.
  - The Codex manual-install path matches the README table.

## [1.28.1] - 2026-10-01

- The `fetch` implement fixture's golden partner fills `resp.Status`
  ("503 Service Unavailable"), as a real transport does. Without it, an error
  built from `resp.Status` lost the code and failed
  `TestGetOrderGivesUpWithinAttempts` in about 80% of Opus 5.5 low sessions in
  every tree, a harness artifact read as noise in every Opus run; all 12 kept
  trees that had failed it pass with the fixed golden.
- `abrun` reads each claude session's routing from its trace and records it
  per result (`routing`: the first tool called, the skills in the first
  Skill-bearing message, an edit before any load, gate blocks). The arm
  summary prints it in one line with shell calls per run. The 1.28.0 note
  wording was chosen on exactly these counts, read until now by scratch
  scripts; on the kept 1.28.0 traces the line reproduces them.
- 1.28.0's prompt note, measured on Sonnet 5.5 (implement, `-shell go`, low,
  n=2, against v1.27.1): parity — golden 100%/100%, lint equal, cost $0.285
  vs $0.275, every session opening with a Skill call that carries
  go-style-core in both trees, 0.36 gate blocks per session in both. That
  release had measured the final wording on Opus 5.5 only.

## [1.28.0] - 2026-10-01

- The idiom card moves into `go-style-core/SKILL.md` as its Current Go Idiom
  Card section, right after Write Current Go; `references/CURRENT-GO.md` is
  gone (69 reference files). Every router loaded go-style-core with the card,
  so a routed session carries the same text, minus one tool call. What
  changes is who can skip it: the card used to arrive only through a whole
  `Read` that the Claude Code gate enforced and the prompt note named by
  path, so on hosts without hooks — GPT 6.1 under Codex, Grok 4.7 under
  opencode — it depended on the model following a "read whole" line, which
  Sonnet 5 did in 0 of 24 sessions. Loading go-style-core now brings it
  everywhere. The gate drops its card requirement and the prompt note its
  card clause; `TestIdiomCardDatesEachSymbol` reads the section's table rows.
- The prompt note loads `go-style-core` in the router's own line and form:
  "load the go-code skill (Skill tool, name …) and the go-style-core skill
  (Skill tool, name …)", with the owners and `go-testing` on the next line,
  each named the same way. Opus 5.5 low acts on that line and treats the next
  as optional: with the card clause gone, "Load go-style-core with it" on the
  next line was skipped in 13 of 14 implement sessions (gate blocks 0.93 →
  2.71 per session against v1.27.1, `-shell go`, n=2), and rewordings of the
  router line — two names without "the … skill", or one list after the
  classification line — let 3 of 14 and 7 of 12 sessions explore through Bash
  and edit before any load. This form, at n=3: 21 of 21 sessions open with a
  Skill call carrying go-style-core and go-testing, 0.43 gate blocks per
  session (v1.27.1: 0.50–1.14 over three runs), cost $0.490 (v1.27.1:
  $0.485–0.506), golden gaps only the `fetch` base rate.
- What these two entries were measured on: Opus 5.5 with the final note
  (above); Sonnet 5.5 with the card move under the earlier note wording
  (implement −6% cost and gate blocks 0.71 → 0.36, refactor equal, against
  v1.27.1, `-shell go`, low, n=2), not with the final wording. GPT 6.1 and
  Grok 4.7, the hosts the move is for, were not measured: no Codex or
  opencode runner was available. On Claude Code the release is parity.
- `abrun -shell go` gives each claude session the shell a user's session
  has: Bash allowed for `go`, `gofmt`, `golangci-lint`, `govulncheck`, `cd`,
  and `bash` for the bundled scripts. The default arms stay shell-less, so
  until now nothing a skill says about running the gate, the scripts, or the
  edit hook's overlap with them could show in a claude run; the codex runner
  already had a shell. The report records `shell`, and each result's
  `commands` now counts Bash calls for claude too.

## [1.27.1] - 2026-09-30

- The prompt hook names `go-testing` without a condition when the prompt asks
  for new code (implement, write, add, a stub): go-code's step 4 writes the
  contract test before the body, and "`go-testing` if you write or edit a
  test" left 43 of 73 implement gate blocks to go-testing on 2026-09-30
  (abrun low, Sonnet 5.5 and Opus 5.5) — the model judged it would write no
  test, then was blocked writing the contract test and re-sent the whole
  file. A fix prompt keeps the condition.
- The gate's owner hints stop firing on routine syntax: a plain `defer`
  (`f.Close()`, `mu.Unlock()`, `cancel()`) no longer names go-defensive — a
  deferred closure, `recover()`, or `unsafe.` does — and a `ctx
  context.Context` parameter passed on no longer names go-context — context
  creation, `AfterFunc`, or a `context.Context` struct field does.
  go-defensive was among the owners named in 16 of those 73 blocks.
- The edit hook prints less that predates the session: `go fix -diff` shows
  the edited file's hunks and counts the rest of the package in one line; in
  a git checkout with a HEAD, `golangci-lint` reports only issues new since
  HEAD; a test timeout prints the running tests instead of the goroutine dump
  (`GOLANG_SKILLS_EDIT_TEST_TIMEOUT` sets the limit, default 50 s). One edit of
  a large test file had printed about 10 KB of older findings, on every edit.
  go-style-core's Edit Hook Record says the output is scoped.
- Measured with `abrun`, reference (v1.27.0) against baseline, implement and
  refactor corpora, `-effort low -n 2`, Sonnet 5.5 and Opus 5.5 (88
  sessions): implement gate blocks per session 1.14 → 0.36 on Sonnet and
  1.21 → 0.64 on Opus (go-testing blocks 8 → 1 and 9 → 0), golden 100%/100%
  and 92%/100%, lint unchanged; Opus implement cost −5% and output tokens
  −11%, Sonnet flat; refactor unchanged. The git-scoped lint cannot show in
  abrun, whose scratch trees are not git checkouts; `TestVetHook` covers it.

## [1.27.0] - 2026-09-30

- A second, smaller token pass on the skills every routed session loads.
  SKILL.md bodies shrink by 5.3 KB and references by 5.9 KB (70 reference
  files): an HTTP implement session with a shell loads about 2.9 KB less, and
  five optional `Read when` references are gone.
  - `go-code`'s Intensity section moves to the new `references/INTENSITY.md`,
    read only when the prompt carries `lite` or `ultra`; the ladder hook reads
    the level rows from there. The refactor scripts leave Close With The
    Gate: `go-code-refactor`, which runs them, lists them itself.
  - `go-code-refactor` loses its Workflow sentence about a variable that
    1.26.1 removed, a second copy of the `go` directive rule (go-style-core's
    Write Current Go), the LOC definitions (the script's `--help` has them),
    the generic list of high-value transformations, a second rename caveat,
    and the report order the asset already carries.
  - `go-testing`'s `TABLE-DRIVEN-TESTS.md` and `TEST-HELPERS.md`, and all
    three `go-error-handling` references, fold into their SKILL.md as the few
    facts they held that the SKILL.md did not: no `/` in a subtest name,
    `t.Cleanup` over `defer` in a helper, `require.ErrorIs`, handling one
    matched case as handling once, no in-band failure values, `%v` redacting
    nothing, and an `os` error already naming its path.
  - `go-linting` keeps `go fix` usage; the analyzer table, the Go 1.27
    renames, and the x/tools `modernize` note move to `MODERNIZATION.md`'s
    Start with `go fix`, which pointed back at them.
  - `go-http` drops Stdlib First (one sentence under Routing now) and its
    Validation section (the gate runs those checks); `go-context` states the
    data-placement order and the `Cause` contexts in one paragraph each, the
    idiom card having the `AfterFunc` and `Cause` rows.
  - Measured with `abrun`, reference (v1.26.1) against baseline, all three
    corpora, `-effort low -n 2`, Sonnet 5.5 and Opus 5.5 (136 sessions):
    implement golden 100%/100% on Sonnet and 92%/100% on Opus with lint
    unchanged, refactor identical (−18.8 lines, golden 100%), cost within
    ±3–7%. Review sessions loaded none of the edited skills, so their
    recall gap (Sonnet 0.96/0.92, Opus 0.93/0.91; must-fix 0.98/0.97 and
    1.00/1.00) is run-to-run noise on identical context. No session in either
    arm read any of the five deleted references.

## [1.26.1] - 2026-09-30

- `go-code-refactor` no longer asks for `export REFACTOR_SKILL_DIR=...` once
  and `$REFACTOR_SKILL_DIR` in every later command: each shell call in Claude
  Code and Codex starts with a fresh environment, so the variable was empty
  from the second call on and the scripts ran as `/scripts/...` (exit 127).
  The commands now carry `<installed-skill-dir>`, as `go-code-review` does,
  and the model writes the path the host printed into each one; the
  separate `--version` probe is gone (exit 127 from the first script says the
  same), `baseline`/`loc-baseline` and `after`/`diff`/`loc-diff` each run in
  one shell call, and the Concision Gate points at those two steps instead of
  a third copy of the commands. `ARCHITECTURE-CHECKS.md` follows.
- `go-code-refactor` states one rule for a generated target: the Orient
  paragraph said "ask when they are the target", the bullet below it "trace
  its generator … ask only if the real source cannot be determined"; the
  paragraph now points at the bullet.
- `PRINTF-STRINGER.md` no longer shows `type fmt.GoStringer interface` and
  `type fmt.Formatter interface`, which are not Go; the method signatures sit
  in the sentence, as 1.26.0 did for `fmt.Stringer`.
- `EMBEDDING.md`'s exception for internal structs excludes a mutex, which
  `SYNC-PRIMITIVES.md` forbids embedding even in an unexported struct.
- `go-code`'s Intensity section moves from third place to just before
  Related Skills, unchanged. A level word is read at step 1 and the ladder
  hook prints the active row itself, while Claude Code re-attaches only the
  first ~5K tokens of a skill after auto-compaction: the cut, which fell at
  the end of the Delete Pass, now falls inside the Declaration Budget, so the
  Reader Pass and most of the budget survive compaction.
- The restraint-ladder and subagent-routing hooks skip the host's agents that
  write no Go — Explore, claude-code-guide, statusline-setup — as they
  already skipped `go-verify`: 3.6 KB of ladder and the router note per such
  subagent. Plan keeps both, since a plan decides what gets written.

## [1.26.0] - 2026-09-30

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
- A token pass over all 24 skills for the target models, GPT 6.1, Opus 5.5,
  Sonnet 5.5, and Grok 4.7; wording that served only Sonnet 5 or Haiku 4.5
  is gone. SKILL.md bodies shrink from 253.7 KB to 208.2 KB (−18%) and
  references from 461.6 KB to 349.3 KB (−24%). Nearly all of it is a second
  copy of text a session already has — an idiom-card row repeated in an
  owner, a SKILL.md section restated in its own reference, a Quick Reference
  table beside the prose it summarizes — or a Go basic the models write
  unprompted; traps, version facts, and the wording earlier runs measured
  stay.
  - `go-linting` keeps the gate, `go fix`, and `nolint` in SKILL.md
    (15.1 KB → 6.2 KB); setup, the baseline, and CI move to the new
    `references/CONFIGURATION.md` (74 reference files), and the table of
    linters that enforce the skills gives way to the comments in
    `assets/golangci.yml`, which already name each rule. A session that
    loads the skill for the closing gate no longer pays for the config
    material.
  - `go-code` (24.3 KB → 21.5 KB) drops the Intensity example and hook
    prose, the duplicate idiom-card and no-shell clauses in its Resource
    Routing and steps, the `ParseBool` and subtree named cases, and most of
    the closure paragraph; the `HEAD` case, the JSON v2 default, the Delete
    Pass, and the Declaration Budget rules stay. The idiom card loses its
    "Which version governs" section, a copy of Write Current Go, which now
    also carries the `go.work` and `-lang` facts. `go-code-refactor`'s
    Resource Routing is one short line per file.
  - The largest owner cuts: `go-concurrency` −42%, `go-data-structures`
    −48% (the card's rows), `go-security` −24% (injection facts stated four
    times), `go-code-review` −18% (rows for rules the models follow
    unprompted, as in ccd1d44), `go-logging` −21%, `go-packages` −18%.
  - Fixes found on the way: `SYMPTOM-CATALOG.md` no longer lists
    `GOMAXPROCS` under `go env`, which prints nothing for it; `go-performance`
    loses the pre-Swiss-table map-bucket note and the claim that the
    compiler allocates a slice's capacity exactly; `PRINTF-STRINGER.md` no
    longer shows `type fmt.Stringer interface`, which is not Go.
  - Per routed session: an HTTP implement session with a shell (`go-code`,
    `go-style-core` and the card, `go-testing`, `go-http`,
    `go-error-handling`, `go-linting`) loads about 16.5 KB (≈6K tokens) less
    skill text, one without a shell about 7.5 KB, a refactor about 5.4 KB,
    a review about 3.1 KB; the restraint-ladder hook prints 3.6 KB instead
    of 4.0 KB into every Go session.
  - Measured with `abrun`, reference (the tree before this pass) against
    baseline, all 17 fixtures, n=2, medium, on Sonnet 5.5 and Opus 5.5: no
    regression. Implement golden 14/14 against 13/13 valid on Sonnet and
    12/13 in both arms on Opus (`fetch`, its known empty-`resp.Status` miss),
    lint and `go fix` hunks equal; refactor 8/8 everywhere at the same line
    deltas; review recall 0.95 against 0.90 and must-fix found 62/62 against
    59/62 on Sonnet. Opus review first showed 47/62 must-fix defects labeled
    Must against 54/62; after restoring the one procedure line the cut had
    taken ("Read the scope file-by-file"), an n=3 rerun gave 81/93 against
    74/93, recall 0.94 in both, baits equal. Cache writes per session fell
    2–8% and cost moved −7% to +2%, within noise, as a text cut of this size
    did before. The config move out of `go-linting` does not show here:
    `abrun` sessions have no shell, so they never load the gate. GPT 6.1 and
    Grok 4.7 are unmeasured (no Codex or opencode runner on the measuring
    host). Raw reports are kept outside the repository.

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
