# Changelog

All notable changes to this repository are documented here.

## [Unreleased]

## [1.12.0] - 2026-09-11

### Added

- Three runs of the 1.12.0 edits against the committed 1.11.0 tree
  (`49b4e25`) on the implementation corpus, no shell in any session:
  - Opus 5 medium, `catalog`, `feed`, `gateway`, three arms, n=1:
    [`docs/evidence/2026-09-11-go-implement-load-cuts-smoke-opus-5-medium.md`](docs/evidence/2026-09-11-go-implement-load-cuts-smoke-opus-5-medium.md).
    Golden 2/3 unaided, 3/3 in both skilled arms. Every baseline session
    loaded `go-code` first and the owners in one message, where the reference
    `gateway` session loaded six owners one per turn: 16 API calls to 10,
    cache reads 562K to 294K tokens, $1.07 to $0.96. Cost 3.96x the control
    against 4.50x.
  - Sonnet 5 medium, the same fixtures and arms, n=1:
    [`docs/evidence/2026-09-11-go-implement-load-cuts-smoke-sonnet-5-medium.md`](docs/evidence/2026-09-11-go-implement-load-cuts-smoke-sonnet-5-medium.md).
    Golden 3/3, 3/3, 2/3: the baseline `feed` session rendered `kinds` as
    `null` for an account with no events, the case its own contract test
    carried; the same miss is in the 2026-09-10 control and the 2026-09-11
    Plain Code run at one session in five. `Skill` turns 2.3 against 3.3 a
    session; cost 6.50x against 7.09x.
  - Sonnet 5 medium, `feed`, reference and baseline, n=5:
    [`docs/evidence/2026-09-11-go-implement-load-cuts-feed-n5-sonnet-5-medium.md`](docs/evidence/2026-09-11-go-implement-load-cuts-feed-n5-sonnet-5-medium.md).
    Golden 5/5 in both arms, `null` in 0 of 10 sessions; lines 34.2 against
    32.8 inside a 30–39 spread, no helpers or types in either arm. Measured
    payloads: `go-style-core` 2867 to 2329–2340 tokens on load, `go-code`
    6601–6635 to 6543–6588, `go-error-handling` 3089 to 2988–3000,
    `go-linting` ~5160 to 5087–5119; cache writes 35.9K against 40.9K,
    $0.275 against $0.299 a session.

### Changed


- `go-code` step 3: the `Skill` calls for the selected owners go out in one
  message. In the 2026-09-11 Opus 5 medium traces
  (`docs/evidence/2026-09-11-go-implement-newcode-workflow-opus-5-medium.traces.tar.gz`)
  the model batched two owners in 12 of 61 Skill-bearing messages and loaded
  the rest one per turn: 3.4 loading turns per routed session, each one
  re-reading 20–40K tokens of context before the first edit. The routing
  gate's message for a blocked edit says the same.
- `go-testing`, `go-naming`, `go-documentation`: the `> **Validation**`
  callouts that told the model to run the new tests now, to run the naming
  script and then `go build`, or to run the docs script and "fix any gaps
  before proceeding" are each one sentence routing to the `go-linting` gate,
  once, at the end of the task. Anthropic's Claude Opus 5 guidance: the model
  verifies its own work unprompted, and a second "check now" instruction adds
  work and report length with no new evidence; the same delete `go-code` made
  in 1.11.0. The scripts stay listed in each skill's Resource Routing and, for
  `check-docs.sh`, in `go-code`'s Close With The Gate.
- `go-style-core` drops Declarations and Scope, Loops and Switches, and Naked
  Returns, about 600 tokens loaded on every routed session. Each rule there is
  Go the model knows: `:=` against `var`, if-init, map order, `break` inside
  `switch`, when a naked return reads. The decisions stay with the references
  Resource Routing already names (`SCOPE.md`, `SHADOWING.md`, `IOTA.md`,
  `INITIALIZATION.md`, `CONTROL-FLOW.md`, `SWITCH-PATTERNS.md`), which
  `go-code` routes to for declaration, enum, initialization, loop, and switch
  decisions. The prompt-audit rule applied: keep what only the author knows.
- `go-code`: the Resource Routing preamble on how a skill counts as loaded is
  four lines instead of seven, and the reason a contract test is written
  before the body is stated once, in step 4, rather than again at the end of
  Contract Table.
- Related Skills in 22 skills are one line per pointer, link and condition:
  17.1K to 11.0K characters across the pack, every link target kept (`go-code`
  and `go-http` were already in this form). `go-concurrency` drops its
  External Resources list of blog posts and talks, which no session can read.
  On the 2026-09-11 Opus 5 traces a routed session loads about 23K tokens of
  skill text; these three cuts remove about 3K across the pack and, measured
  on the `feed` n=5 run above, about 750 tokens per session (the estimate from
  character counts was 1.1K).
- README: a Loading cost paragraph under the per-model table with the
  measured numbers from the three runs; the table itself is unchanged, since
  the runs are n=1 or have no unaided arm.

## [1.11.0] - 2026-09-11

### Added

- `hooks/go-subagent-routing.sh`, a SubagentStart hook. A subagent starts
  with an empty context, so nothing the main session loaded reaches it; when
  the working directory holds Go, the hook adds the prompt hook's note naming
  `go-code`, or `go-code-refactor` for a refactor, to load before the first
  edit. It fires for every subagent, since each is a fresh context, skips the
  plugin's own `go-verify` agent, and never blocks. Motivation: Anthropic's
  Claude Opus 5 migration guidance records that the model delegates to
  subagents more readily than Opus 4.8; ponytail's plugin injects its ruleset
  into every subagent for the same reason. `TestSubagentRouting` covers the
  note in a Go directory, twice for two subagents, and silence elsewhere and
  for `go-verify`.

### Changed

- `go-code` drops its re-check choreography. "Inspect the final diff and
  complete authorized work. Reuse passing results for unchanged code; rerun
  affected checks after edits and honor host checkpoints" is gone from Close
  With The Gate, with that section's repeat of the no-shell rule; the
  reuse-and-rerun rule stays where `go-linting` states it. Anthropic's Claude
  Opus 5 migration guidance: the model verifies its own work unprompted, and
  instructions that tell it to verify now cause over-verification, so
  removing them is a delete, not a rewrite. What stays is the rule the
  2026-09-05 review made load-bearing: the report carries only observed
  results, each `pass`, `fail`, `unavailable`, or `skipped`.
- `go-code` no longer explains the routing gate and the vet hook to the
  model. The step 3 paragraph on how the PreToolUse gate infers owners, what
  it cannot see, and that a blocked edit is unapplied, and the Close With The
  Gate sentence on the PostToolUse hook's silence, are removed, about 300
  tokens on every load. The gate's own message now says which owners it
  recognizes, that the routing table decides the rest, and that its silence
  is not a result, so the text reaches the model only when the gate fires.
- `go-code` step 5 and `go-testing`'s assertion policy: a task that does not
  ask for a dependency adds none. `cmp.Diff` is the default for structured
  values only in a module that already requires go-cmp; otherwise
  `reflect.DeepEqual`, `slices.Equal`, or `maps.Equal`, and `go.mod` is left
  alone. Motivation: in the 2026-09-11 Opus 5 prompt-routing run, `catalog`
  r0 spent five turns reading and editing `go.mod` and writing `go.sum` to
  add go-cmp for its contract test, following the old default, in the user's
  module.
- README: run the plugin at `--effort medium` on Opus 5, the condition every
  Opus 5 cell was measured at; a `low` control is the open item.

## [1.10.0] - 2026-09-11

### Added

- The three-arm run of the routing edits on the fixtures and model that
  recorded the gap: `catalog` and `feed`, Opus 5 medium, n=3, 18 sessions,
  `reference` the committed tree at `f4f373f`:
  [`docs/evidence/2026-09-11-go-implement-prompt-routing-opus-5-medium.md`](docs/evidence/2026-09-11-go-implement-prompt-routing-opus-5-medium.md).
  The reference arm reproduced the miss — two of three `catalog` sessions
  loaded no skill, edited the stub in prose and wrote no test — and the
  baseline arm loaded `go-code` in 6/6 sessions as its first tool call, at
  turn 2, straight after the hook's note; the four reference sessions that
  did route loaded it at turn 6 or 7. Every baseline session then followed
  the workflow (`checks:` line, budget line, contract test 6/6 against 4/6).
  Golden 6/6 in every arm; `catalog` 19.0 lines in every baseline session
  against 20.0 and 25.0, `feed` 34.7 against 35.0 and 49.7; cost 5.44x
  against 4.23x, the whole gap the two sessions that skipped the workflow.
  The hook and the description ride in one arm and are not separated.
- A three-arm smoke of the two routing edits below on the implementation
  corpus, Sonnet 5 medium, n=1, 12 sessions, `reference` the committed tree
  at `f4f373f`:
  [`docs/evidence/2026-09-11-go-implement-prompt-routing-smoke-sonnet-5-medium.md`](docs/evidence/2026-09-11-go-implement-prompt-routing-smoke-sonnet-5-medium.md).
  The `UserPromptSubmit` hook fired in 4/4 baseline sessions and in none of
  the other eight, shown by the `prompted` state it writes beside the routing
  gate's `loaded` list under `~/.claude/plugins/data/`; its note is not
  echoed in a `stream-json` trace. Every baseline session's first tool call
  is `Skill go-code`, before any file is read, where every reference session
  runs `Glob` first and loads `go-code` two or three turns later. Golden 4/4
  against 3/4 against 4/4 unaided, the one failure a reference `gateway`
  session rendering an empty list as `null`; turns 26.8 against 32.8; cost
  7.21x against 7.71x; `go-linting` loaded with no shell in 4/4 baseline
  sessions against 1/4. One session per cell: directions, not results. Sonnet
  5 routed 4/4 in both skilled arms, so the Opus 5 and Haiku 4.5 gaps the
  edits target are not measured here.
- `gateway` on Opus 5 medium with the routing edits, three arms, n=3, 9
  sessions, `reference` the committed tree at `f4f373f`:
  [`docs/evidence/2026-09-11-go-implement-prompt-routing-gateway-opus-5-medium.md`](docs/evidence/2026-09-11-go-implement-prompt-routing-gateway-opus-5-medium.md).
  Unaided 0/3, every session registering `GET` patterns only and saying the
  mux answers `HEAD` with 405; both skilled arms 3/3 with a `HEAD` case in
  every contract test, so Opus 5 on this fixture is 12/12 skilled against
  1/6 unaided across the day's two runs. The reference description routed
  3/3 here, at turn 7; the hook moved the load to the first tool call in 3/3
  and changed nothing else — lines 81.0 against 77.3, cost 3.35x against
  3.68x. The description's misses are on `catalog` and `feed`, not on a stub
  that imports `net/http`.
- The three-arm run of the routing edits on the refactor corpus, Haiku 4.5,
  n=3, 36 sessions, `reference` the committed tree at `f4f373f`:
  [`docs/evidence/2026-09-11-go-refactor-prompt-routing-haiku-4-5.md`](docs/evidence/2026-09-11-go-refactor-prompt-routing-haiku-4-5.md).
  The reference arm reached `go-code-refactor` in 4/12 sessions, at turn 8 to
  13, only on `pricing` and `store`; the baseline arm loaded it in 12/12 as
  the first tool call after the hook's note. Routed, the skill does not yet
  pay on this tier: golden 10/12 against 12/12 unaided, both failures
  `report` sessions that extracted a `formatAmount` helper and changed
  `%9d.%02d` to `%9s`, moving the aligned column — a cut the unrouted
  reference `report` session made too; `dispatch` −5.0 against −9.7 unaided;
  `pricing` and `store` level; three baseline sessions failed tests they
  wrote themselves. Cost 2.56x against 1.74x. The hook is kept as the
  mechanism that delivers the skill; what to change in the skill for this
  tier is the open item.
- `hooks/go-prompt-routing.sh`, a UserPromptSubmit hook. When a prompt asks
  for Go work — it names Go, a `.go` file or `go.mod`, or is sent from a
  directory holding Go and names a function, package, handler, or test — the
  hook adds one note to the model's context naming the router to load before
  the first edit: `go-code-refactor` for refactor, clean-up, or simplify
  wording, `go-code` for anything else. Once per skill per session; silent
  when the session already loaded the skill, when the prompt invokes a go-*
  skill by name, or when the prompt carries no work verb. It never blocks.
  The description of a skill reaches the model only when the host's matcher
  fires; this hook reads the prompt itself. Motivation: Opus 5 loaded no
  skill in 5 of 18 skilled sessions of the 2026-09-11 workflow run, all on
  `feed` and `catalog`, whose prompt says only "Implement the Go package in
  ./<dir>"; Haiku 4.5 reached `go-code-refactor` in 1 of 4 refactor sessions
  on 2026-09-10. Measured the same day in the four runs above: the hook fires
  in every skilled session, routes 6/6 against 4/6 on Opus 5 `catalog`/`feed`
  and 12/12 against 4/12 on Haiku 4.5, and moves the router load to the first
  tool call on every model.
- A validation trigger eval for `go-code` in the shape of that prompt: only
  the package and "write the bodies", no `new`, `function`, or `stub`.

### Changed

- `abrun` passes `--include-hook-events` to the Claude CLI, so hook
  lifecycle events — including the text a `UserPromptSubmit` hook adds to the
  model's context — appear in the trace as `system` messages with the
  `hook_started` and `hook_response` subtypes. Without the flag the host adds
  the note silently and a trace cannot show whether the hook fired; the smoke
  above had to read the hook's state directory instead.
- `go-code`'s description also names implementing a Go package, function, or
  handler whose declarations and documentation already exist — write the
  bodies, fill in a stub, replace `panic("not implemented")` — even when the
  request names only the package. The old text said "writing, fixing, or
  refactoring Go code", which the Opus 5 sessions above did not match to
  "Implement the Go package". Measured only together with the hook, in the
  same arm; a run without the hook is the open item.

## [1.9.0] - 2026-09-11

### Added

- Two three-arm runs of the `go-code` workflow edits below on `catalog`,
  `feed` and `gateway`, `reference` the committed tree at `ead8ce6` and
  `baseline` the working tree. Sonnet 5 medium, n=5, 45 sessions:
  [`docs/evidence/2026-09-11-go-implement-newcode-workflow-sonnet-5-medium.md`](docs/evidence/2026-09-11-go-implement-newcode-workflow-sonnet-5-medium.md).
  Every compliance measure moved the way the edit intended: the `checks:`
  line, the budget line and the no-shell statement 15/15 (8/15, 0/15 and
  10/15 for the old wording), the Contract Table written 15/15 and 10 before
  the body (8 and 6), a `HEAD` case in the model's own `gateway` test 5/5
  after 0 of 11 at n=10, shell-less `go-linting` loads 5/15 against 9/15.
  `catalog` 5/5 at 19 lines in every session with no added declaration
  (−8.2 against the control, p = 0.008); `feed` level with the retune at
  32.6 against 30.8; `gateway` 4/5 in both skilled arms against 5/5 unaided,
  the one baseline failure a session that wrote the `HEAD` case and then
  trusted method-less fallback patterns to catch it. Cost **7.20x** against
  5.58x for the reference tree: the whole gap is the Contract Table written
  every time (`go-testing` 15/15 against 8/15). Opus 5 medium, n=3, 27
  sessions, the first measurement of the new-code section on this model:
  [`docs/evidence/2026-09-11-go-implement-newcode-workflow-opus-5-medium.md`](docs/evidence/2026-09-11-go-implement-newcode-workflow-opus-5-medium.md).
  9/9 correct in both skilled arms against 7/9 unaided, both control
  failures `gateway` answering `HEAD` with 200 at 110–147 lines; every
  `go-code` session followed the workflow in both trees, so the edits change
  this model less — reports shorter (median 1151 against 1802 characters),
  the last shell-less `go-linting` load gone, cost 3.52x against 3.23x. Five
  of eighteen skilled Opus sessions loaded no skill at all, all on `feed`
  and `catalog`. The README Opus 5 new-code cell moves to this run; the
  Sonnet 5 cell stays on the n=5 full-corpus run and is qualified by the
  subset.

- The n=10 three-arm run on `catalog`, `feed` and `gateway` that the n=5
  report asked for, 90 sessions at Sonnet 5 medium, `reference` the pre-retune
  tree and `baseline` the 1.8.0 tree with the reshaped example and the
  fewer-names line:
  [`docs/evidence/2026-09-11-go-implement-newcode-plaincode-n10-sonnet-5-medium.md`](docs/evidence/2026-09-11-go-implement-newcode-plaincode-n10-sonnet-5-medium.md).
  The `gateway` 5/5 of the n=5 run does not hold: 4/10 against 5/10 for the
  pre-retune tree and 10/10 unaided (Fisher p = 0.01), no baseline session
  registering a `HEAD` pattern, six of ten reports saying `ServeMux` answers
  405 on its own, the unaided control passing every session with an
  `r.Method` check neither skilled arm writes. `feed` holds with the reshaped
  example: 33.9 lines against 43.6 and 46.1 (p < 0.01 against each), 10/10
  correct in every arm, nine of ten sessions in both skilled arms using a
  function-local type and an anonymous document, so the gap is the inline
  form, not the local type. `catalog` ties the control at 9/10 while the
  pre-retune tree fails the repeated-SKU clause 4/10 (p = 0.30 between the
  skilled trees); a `resolveError` type is back in 2 of 10 baseline sessions.
  Contract tests 17 written and 13 before the body against 12 and 6;
  `go-linting` loads 17/30 against 11/30 with no shell in any session, and the
  baseline arm costs 20% more per session than the pre-retune tree (5.69x
  against 4.74x the control). The README Sonnet 5 new-code cell stays on the
  n=5 full-corpus run and is qualified by this subset.

### Changed

- `go-code`'s workflow has six steps and names what it asks for. Step 2
  checks the tool list for a shell tool by name (`Bash` in Claude Code) and
  decides the task's shape; step 3 names `go-testing` among the owners when
  a Contract Table will be written, and says a hook-blocked edit was not
  applied; step 4 is the Contract Table, for new code only; step 6 reports
  in a literal four-line shape — outcome, `checks:` line, budget line, gaps —
  with `checks: unavailable (no shell)` as the whole gate report when no
  shell tool exists. The `go-linting` routing line and Close With The Gate
  repeat the no-shell rule at the point of use. Measured above: Sonnet 5
  followed the shape in 15/15 sessions where the prose form reached 8/15,
  0/15 and 10/15; Opus 5 had followed the prose form already.
- The Contract Table takes a class clause's case — "any other method", "any
  other value" — from the member a library default treats unlike the rest:
  `HEAD` under a `GET` pattern, `t`/`1` under `strconv.ParseBool`, the bare
  path under a subtree pattern; the example table gains a `HEAD /healthz`
  row and the `GET`-serves-`HEAD` default joins the list the contract
  overrides. Measured above: a `HEAD` case in 5/5 Sonnet 5 and 3/3 Opus 5
  `gateway` tests, after 0 of 11 at n=10.
- Declaration Budget rule 2 says that a caller which inspects or matches an
  error uses `errors.Is` on the wrapped sentinel and needs no type;
  `catalog` declared none in 5/5 Sonnet 5 sessions after 2/10 at n=10.
- `go-http`'s Routing bullet leads with "A `GET` pattern also serves `HEAD`",
  names the `r.Method != http.MethodGet` form and the test that catches the
  default; its Related Skills route a handler or server written from a
  specification to `go-code`.
- The routing gate's block message says the edit was not applied and the
  file is unchanged, and asks for the same edit again; the n=10 run had four
  sessions retry against the content of an edit that never landed.
- The Plain Code example in `go-code` changes shape: a build manifest with an
  ordered list, the path of the largest file and a total, in place of the
  document with a sorted distinct list and a count map that shared its members
  with the `feed` fixture, so that fixture measures the rule rather than the
  copying of a near-solution. The example keeps the lesson — a function-local
  type, an anonymous document, `[]` for the empty case as code — and shows the
  fewer-names-not-fewer-states line as code, a `largest` size kept beside
  `doc.Largest`. Measured in the n=10 run above: `feed` at 33.9 lines against
  43.6 and 46.1 with six of ten sessions at 30 lines.

## [1.8.0] - 2026-09-11

### Added

- Full Sonnet 5 controls at `-effort medium` on release 1.7.0, both corpora,
  n=5 per fixture and arm, 80 sessions and one plugin digest:
  [`docs/evidence/2026-09-10-go-refactor-control-sonnet-5-medium-n5.md`](docs/evidence/2026-09-10-go-refactor-control-sonnet-5-medium-n5.md)
  and
  [`docs/evidence/2026-09-10-go-implement-control-sonnet-5-medium.md`](docs/evidence/2026-09-10-go-implement-control-sonnet-5-medium.md).
  The refactor corpus turns the n=1 direction into a measurement and
  reproduces it: −9.7 lines against −9.6, correctness tied at 20/20 build and
  golden in both arms, new functions 1.45 → 0.58 per session, concision gate
  16/20 against 10/20. Three of four fixtures separate — `dispatch` −15.8 with
  arms that do not overlap (p = 0.00794, 0.032 with Bonferroni), `report` −9.3
  (p = 0.00794), `pricing` −9.0 (p = 0.024, exploratory after correction) —
  and `store` is a tie. The implementation corpus separates nothing: −1.78
  lines across the three evaluable fixtures, 13/15 correctness against 14/15,
  no fixture under p = 0.05. `gateway` remains unusable on this model at this
  effort, failing the documented HEAD/405 contract in 8 of 10 sessions across
  both arms — including 5/5 skilled sessions in which `go-http` loaded, which
  has carried the `ServeMux` HEAD rule since `f12c73b`; three of those also
  rendered the empty account list as `null`. Cost is 4.38x and 4.54x.
- A one-repetition smoke of the new-code edits below, three arms on the
  implementation corpus at Sonnet 5 medium, 12 sessions:
  [`docs/evidence/2026-09-10-go-implement-newcode-smoke-sonnet-5-medium.md`](docs/evidence/2026-09-10-go-implement-newcode-smoke-sonnet-5-medium.md).
  Not a measurement; it records which instructions the model followed. The
  Declaration Budget was reported in 4/4 sessions (`gateway` 2 functions
  against the reference arm's 4; `feed` rationalized two package-level types
  the reference did without). The `go-http` HEAD example was copied: the first
  skilled `gateway` session at this effort to pass the HEAD assertions, after
  0/10 across the two controls. The Contract Table was written before the first
  edit in 0/4 sessions and appeared post hoc in 4/4. The `go-data-structures`
  Copy caveat did not hold: the hook-forced load, the `make`+`copy` to
  `slices.Clone` rewrite and the `null` list reproduced step by step in the
  one `gateway` session. Cost 5.79x against the reference arm's 4.92x.
- The same edits after the hook fix, the split Copy row and the test-file
  form of the Contract Table, on `gateway` and `feed` at n=3, reference
  against baseline:
  [`docs/evidence/2026-09-10-go-implement-newcode-n3-sonnet-5-medium.md`](docs/evidence/2026-09-10-go-implement-newcode-n3-sonnet-5-medium.md).
  `gateway` 0/3 against 3/3 golden (Fisher p = 0.10 two-sided; the 1.7.0 tree
  is 0/14 on the HEAD assertions at this effort across four runs), every
  baseline session registering `HEAD` patterns beside `GET` and none returning
  `null`; `feed` 3/3 against 3/3 at level size. Contract tests written and
  passing in 6/6 baseline sessions, before the first production edit on
  `feed` 3/3 and after it on `gateway` 3/3. `go-data-structures` loads fell
  from 6/6 to 0/6; cost rose 43% per session, the test file and the
  `go-testing`/`go-defensive` loads its write triggers.
- The full implementation corpus at n=5, release 1.7.0 against the working
  tree, 40 sessions:
  [`docs/evidence/2026-09-10-go-implement-newcode-control-sonnet-5-medium.md`](docs/evidence/2026-09-10-go-implement-newcode-control-sonnet-5-medium.md).
  Golden 17/20 against 14/20; `gateway` 4/5 against 0/5 (Fisher p = 0.048),
  the one baseline miss a session that loaded `go-http` alone and never
  reached `go-code`; every session that did reach it passes the HEAD
  assertions, 7/7 across the three runs on this tree against 0/19 for 1.7.0.
  `catalog` and `ledger` tied; `feed` 4/5 against 5/5, the miss a `kinds`
  member rendered `null` through `slices.Sorted(maps.Keys(m))` that the
  session's own contract test caught and could not run. Size is level to
  slightly larger on the three fixtures both arms pass (+2.1 to +7.8 lines,
  no p below 0.27), and the growth is structure the Declaration Budget
  reasoned about rather than refused: an error type where `%w` serves, and
  package-level wire types. Contract tests written in 10/20 sessions, 8 pass,
  2 fail. Cost per session $0.2118 against $0.2043; hook blocks per session
  1.10 → 0.50.
- The three-arm run that sets the README cell, an hour later on the same
  model and effort: `no-skill`, the tree above as `reference`, and the tree
  with three refinements as `baseline`, n=5, 60 sessions:
  [`docs/evidence/2026-09-10-go-implement-newcode-final-sonnet-5-medium.md`](docs/evidence/2026-09-10-go-implement-newcode-final-sonnet-5-medium.md).
  It pulls the earlier readings back: `gateway` 4/5 unaided against 3/5 and
  2/5 for the skill trees, the `go-http` example copied in 2 of 10 sessions
  that loaded it against 8 of 9 in the two runs before. Aggregated over every
  Sonnet 5 medium session in the series the picture is: release 1.7.0 0/19
  on `gateway`, the unaided control 8/16, the new-code trees 12/19 — the
  regression 1.7.0 carries is established and gone (p = 0.00001), a benefit
  over no skill is not (p = 0.30). Correctness 16/20 against 18/20 unaided,
  size −3.8 and −4.2 lines on `catalog` and `feed` (p = 0.46, 0.07), level on
  `ledger`, cost 4.73x. Of the refinements, the narrowed reason 2 moved
  `catalog`'s error type the right way (3/5 → 1/4) and `feed`'s package-level
  wire types the wrong way (1/5 → 3/5), with the budget lines citing the
  reason verbatim against its own text; the hook change took the cost from
  5.49x to 4.73x.
- A one-repetition smoke of the Plain Code retune below, three arms on the
  implementation corpus at Sonnet 5 medium, 12 sessions:
  [`docs/evidence/2026-09-10-go-implement-newcode-plaincode-smoke-sonnet-5-medium.md`](docs/evidence/2026-09-10-go-implement-newcode-plaincode-smoke-sonnet-5-medium.md).
  Not a measurement; it records which instructions the model followed. The
  `feed` session wrote the Plain Code example's shape with the fixture's nouns
  at 30 lines against 41 and 44 in the other arms, which reads as shape
  transfer rather than a size result because the example and the fixture
  share a shape. The budget line replaced the reason sentence in 3 of 4
  reports, the contract test came before the body in 3 of 4 sessions against
  1 of 4, and the report shape held in 2 of 4. `gateway` failed both live
  traps — `GET` patterns only beside the `go-http` `HEAD` example, and
  `slices.Clone` serving `null` for a nil list — while its own pre-written
  contract test carried the `null` case it could not run; the reference
  session passed. Cost 7.37x against 5.81x, the test written first and the
  gate load without a shell the drivers.
- The three-arm run that measures the Plain Code retune, n=5, 60 sessions,
  `reference` the tree the 2026-09-10 three-arm run measured as `baseline`:
  [`docs/evidence/2026-09-11-go-implement-newcode-plaincode-sonnet-5-medium.md`](docs/evidence/2026-09-11-go-implement-newcode-plaincode-sonnet-5-medium.md).
  Correctness ties at 17/20 in all three arms and hides two moves: `gateway`
  5/5 against 2/5 and 2/5, every baseline session registering the `HEAD`
  patterns where the reference arm copied them in two of five; `catalog` 3/5
  against 5/5 and 5/5, both failures using the result map as the set of SKUs
  already asked for. Size is smaller than both other arms on three fixtures
  and separates only on `feed`, −15.6 against the control (p = 0.01), where
  four of five baseline sessions wrote the Plain Code example's shape with
  the fixture's nouns at 30 lines. Reports carry `unavailable (no shell)` in
  12/20 against 3/20 and stop walking the contract cases (reports with five
  or more bullets 4 → 1), at the same length; `go-linting` still loads in
  13/20 shell-less sessions. Cost 5.57x against 5.54x.

### Changed

- Both READMEs read the Sonnet 5 refactoring cell as ✅ at medium effort,
  set by the n=5 pair above rather than qualified by the one-repetition runs,
  with the effort level named in the cell and the cost column carrying 4.4x at
  medium beside 3.3x at the CLI default and 5.6x at high. The new-code cell
  stays ➖ and now rests on the same-day three-arm run on the working tree:
  16/20 against 18/20 unaided, `gateway` 2/5 against 4/5, about four lines
  fewer on `catalog` and `feed`, 4.7x the cost, and the note that the 1.7.0
  tree's `gateway` regression is gone without a benefit over no skill being
  established.
- `go-code` Writing New Code gains two sections the refactor skill's results
  suggested and the implementation controls asked for. A **Contract Table**,
  written as a table-driven `_test.go` before the first production edit,
  turns each observable clause of the documentation into a case and is run
  with the closing gate where a shell exists and read against the code where
  it does not; a **Declaration Budget** charges every declaration added beyond the
  specification with one of three reasons and is reported as a count. Both
  replace routes the 2026-09-10 control shows were never followed: the
  restraint reference was read in 0 of 20 skilled sessions, the `go-http` HEAD
  rule was in context in 5 of 5 `gateway` sessions that still called HEAD an
  automatic 405, and `go-data-structures` was loaded in the 3 sessions that
  rewrote `make`+`copy` into `slices.Clone` and returned `null`.
  `docs/RULE_OWNERSHIP.md` records the new rule area and the ladder exception.
- Caveats moved from bullets into the examples that primed the defect. The
  `go-http` Routing example registers a `HEAD` pattern beside its `GET` pattern
  and says in the code that the GET pattern otherwise answers HEAD with 200
  (verified on go1.27.1: the pair registers without conflict, HEAD gets the
  405 and `Allow: GET`). The `go-data-structures` Copy row splits in two —
  `Clone` where a nil input may stay nil, and `make` + `copy` or
  `append([]T{}, s...)` where the copy must encode as `[]` or `{}` under
  `encoding/json` v1 — and the `go-defensive` boundary-copy example shows both
  forms as code. The first version of this edit kept the caveat in the same
  cell as `Clone`; the smoke above shows the positive half of the row applied
  and the caveat ignored, so the non-nil copy is now a row of its own.
- Three refinements after the n=5 control, measured in the three-arm follow-up
  rather than shipped on the reading alone. Reason 2 of the Declaration
  Budget covers an algorithm or a resource lifetime and says a wire document,
  a formatted error, or a sorted view is a representation the standard library
  already expresses; the budget names the declaration-free forms to take
  first (a wrapped error before an error type, an anonymous or function-local
  type before a package-level one, a closure before a single-caller helper),
  after 3/5 `catalog` and 4/5 `feed` baseline sessions charged exactly those
  declarations to it. The `go-data-structures` key-collection row splits the
  way the Copy row did — `slices.Collect`/`slices.Sorted` where an empty result
  may be nil, `slices.AppendSeq(make([]K, 0, len(m)), maps.Keys(m))` where it
  must encode as `[]` — after one `feed` session rendered `kinds` as `null`
  through the first form. The routing gate names `go-testing` alone for a
  `_test.go`: a test body's `defer` was pulling `go-defensive` into contract-test
  sessions, and `TestRoutingGate` now checks that a test body names no other
  owner.
- `go-code` Writing New Code is retuned against Anthropic's prompting guides
  for Claude Sonnet 5 and Claude Opus 5 and the transcripts of the 2026-09-10
  three-arm run. A **Plain Code** section states the form a body takes — the
  specification's vocabulary, a type or document declared inside the one
  function that builds it, steps inline, an error wrapped with `%w`, a comment
  only for a constraint the code cannot show — as one positive example rather
  than as caveats beside the budget, because both guides say the models copy
  the shape they are shown and apply a rule literally at the scope it names.
  The **Declaration Budget** counts package-level declarations against an
  expected zero for a body behind an existing signature, admits one for two
  named call sites, a caller that names it, or a distinct algorithm whose name
  says more than its body, and drops the reason slot the transcripts show
  being filled with the budget's own words: the `feed` sessions that kept
  package-level wire types cited "wire-format representations (reason 2)" for
  the types the reason was written to exclude, at 83 lines against the unaided
  arm's 53 for the same document built inside the function. Step 5 reports
  the outcome, the observed checks, the one budget line, and material gaps,
  and no longer walks the contract cases (a `feed` report listed all seven
  with check marks); the Opus 5 guide names stacked verification and re-check
  phrasing as the source of longer work and longer reports. `go-style-core`
  gains the owner rule for internal comments the section routes to, and the
  authoring template records the three prompting rules — scope in the rule, a
  reason slot is a template, verification stated once — with the guides as
  sources.
  Measured the next morning in the three-arm run above. After that run the
  Resource Routing line reads `go-linting` only when a shell can run its
  checks, step 5 reports every check as `unavailable (no shell)` in one line
  otherwise and names the test file instead of its cases, and Plain Code
  gains the line that fewer names never means fewer states — a value used
  once needs no name, a fact the code tracks keeps its own variable — after
  two `catalog` sessions used the result map as the set of SKUs already
  asked for. The last two of those are unmeasured; the first was in the
  measured tree in its earlier wording.

### Fixed

- The routing gate named `go-data-structures` for `make([]`, `make(map` and
  `append(`, which appear in nearly every Go body — routine syntax by the
  gate's own rule. The forced load is where the 2026-09-10 control's three
  `null` lists started (the `make`+`copy` to `slices.Clone` rewrite follows
  it in the traces) and the smoke reproduced the chain in its one `gateway`
  session. The hint is gone; the `go-code` table still routes a task about
  collections to its owner, and `TestRoutingGate` now checks that `make` and
  `append` leave the gate silent.

## [1.7.0] - 2026-09-10

### Fixed

- `go-security` said the default TLS curve list carries standalone `MLKEM1024`
  and told reviewers to delete any explicit `CurvePreferences`. `MLKEM1024` is
  opt-in in Go 1.27 (`defaultCurveEnabled` returns false for it), so the advice
  deleted deliberate post-quantum hardening; the finding is now a list that
  omits the hybrids. The same reference claimed an environment `GODEBUG` pin
  of a removed TLS setting "breaks the build" — it makes the binary refuse to
  start; only a `go.mod` or `//go:debug` pin fails the build. ML-DSA signature
  schemes are advertised by default; the certificate chain is the opt-in.
- The SSRF example `safeTarget` in `go-security/references/INJECTION.md` did
  not compile (`ctx` undefined); it now takes a `context.Context` and is
  compile-tested with blocked and public literal addresses.
- `encoding/json/v2` ignores unknown tag options without error: a copied
  `inline` or `unknown` tag nests or drops data silently. `go-defensive`
  Struct Field Tags now says so and asks for a byte-level test.
- `go-generics` and `COMPATIBILITY.md` implied the `go` directive gates
  language features. It does not: generic methods, function-type inference,
  `new(expr)`, and self-referential constraints compile on a 1.27 toolchain
  with `go 1.26` in `go.mod` and fail only on a real older toolchain. Both now
  say to verify on the CI toolchain; `go-generics` gains a self-referential
  constraints section (Go 1.26+).
- `go-packages` said flags default to `snake_case` while its own reference
  called `output_dir` bad and required hyphens; the reference now matches the
  skill. `go mod init` guidance names both toolchain behaviors (1.27.x writes
  its patch level, 1.26.x writes `go 1.25.0`) and the CI consequence.
- Error naming was routed in a circle between `go-error-handling` and
  `go-naming` with the rule stated nowhere; `go-naming` now owns `ErrX`
  sentinels and `XError` types.
- `go-http` told clients to `defer resp.Body.Close()` "on every response, error
  or not", which dereferences nil on error; the close now follows the error
  check. Handler Shape names the trigger for v1 versus v2 request decoding.
- `go-code-review` and `pre-review.sh` reported a missing golangci-lint as
  `skipped`, the word the gate reserves for a deliberate omission; both now say
  `unavailable`, and the pre-review summary reads `INCOMPLETE` in that case.
  The review's Automated Checks listed its own partial gate with `-race`
  conditional on goroutines; it now reports the go-linting gate result.
- The go-code-refactor Concision Gate made any production-LOC growth a hard
  failure while three other passages accepted growth with a reason. `loc-diff`
  exit 1 is now a signal that the report must justify; `loc-baseline` and
  `loc-diff` are in Workflow steps 1 and 5, where the gate expected them.
- `go-code` carried its own paraphrase of the seven-rung restraint ladder that
  1.5.0 had removed, plus a 42-line essay of rules owned by other skills; both
  are routes now. Every task now lists `go-linting` in Resource Routing.
- Non-compiling examples fixed: the `slogtest` file in LOGGING-PATTERNS.md, the
  "Bug! This compiles" channel example that did not compile, a
  declaration-plus-statement block in GOROUTINE-PATTERNS.md, the WEB-SERVER
  example's missing `User` and `NewDBStore`, the table-test asset's package
  mismatch, and `SafeCounter` in go-data-structures.
- `go-documentation/references/FORMATTING.md` taught pre-Go 1.19 Godoc syntax
  (headings without `#`, lists as verbatim blocks, no `[Name]` links); it now
  teaches the syntax `gofmt` rewrites comments into.
- `BEHAVIOR-TRAPS.md` linked a heading anchor that did not exist; CI and
  `TestCrossRefs` now check anchors, not just paths.
- `bench-compare.sh --json` reported `"exit_code":0` for a package without
  benchmarks while the process exited 1; the JSON now carries `status` and the
  script's own exit code (v1.2.0).
- `go-linting` presented `appendclipped`/`slicesdelete` as `go fix` flags; they
  belong to gopls's modernize suite. The modernizers are `go fix` analyzers
  since 1.26, not 1.27; the pre-commit snippet lints only the staged change
  (`--new-from-rev=HEAD`); `govulncheck` is pinned through the `tool` directive.

### Added

- `abrun` records the modernizations `go fix -diff` still proposes for the
  fixture package: `fix_hunks_before` on the fixture as shipped, `fix_hunks`
  after the last turn, with hunks in `*_test.go` skipped so the count follows
  the same production-only rule as `lines`. It is the one reading of "reaches
  for what the toolchain already ships" that costs nothing and needs no rubric
  — the analyzers decide. `go fix -diff` exits non-zero exactly when the diff
  is not empty, so an empty diff from a package that does not type-check is
  recorded as unmeasured (`fix_hunks_unmeasured`) rather than as a clean zero,
  and the summary averages the pair only over the valid runs it could read at
  both ends.
- First A/B control on Haiku, closing the "check on Haiku, Sonnet and Opus"
  gap in the skill-authoring guidance:
  [`docs/evidence/2026-09-10-go-refactor-control-haiku-4-5.md`](docs/evidence/2026-09-10-go-refactor-control-haiku-4-5.md).
  n=1 per fixture and arm supports no structural claim, and the finding is
  categorical instead: `go-code-refactor` reached 1 of 4 baseline sessions
  against 19/20 on Opus 5 and 20/20 on codex, with the arm loading correctly
  in every one. On this tier a wording comparison measures the router, and the
  1.6.0 routing gate does not fire in a session that never loaded `go-code`.
  Its same-conditions pair on Sonnet 5 at medium effort —
  [`docs/evidence/2026-09-10-go-refactor-control-sonnet-5-medium.md`](docs/evidence/2026-09-10-go-refactor-control-sonnet-5-medium.md),
  same seed, tree, digest and day — reached 4/4 and settles that the Haiku
  number belongs to the tier and not to the arm. 8/8 valid there, correctness
  tied at 4/4, gate 4/4 against 2/4, −15.8 lines against −6.2 with `report`
  rewritten to its original size while the control grew 14; one repetition, so
  a direction to spend n=5 on rather than an effect. The same run at
  `-effort high` —
  [`docs/evidence/2026-09-10-go-refactor-control-sonnet-5-high.md`](docs/evidence/2026-09-10-go-refactor-control-sonnet-5-high.md)
  — closes that gap to −1.8: the unaided control gains 3.3 lines of concision
  and the skilled arm loses 4.6, and on `report` both arms grow by 12 lines
  where medium's skilled arm held the original size. Routing goes the other
  way, 4/4 with eleven skill loads against five, which is firing rate and
  measured outcome separating again. Effort is a condition a control has to
  hold fixed and name.

### Changed

- One verification gate, one order, one vocabulary. `go-linting` owns the
  command order (gofmt, build, vet, test -race, fix -diff, lint, govulncheck)
  and the states `pass` / `fail` / `unavailable (reason)` / `skipped (reason)`;
  `go-verify`, `go-code`, `go-code-review`, and `go-code-refactor` follow it
  instead of restating it. The convention-file list (`AGENTS.md`, `CLAUDE.md`,
  `CONTRIBUTING.md`, `.golangci.yml`, CI, neighboring code) lives once in
  `go-style-core` House Style Wins.
- Rules returned to their owners: typed nil to `go-defensive`; nesting and
  `if`-init scope and log levels out of `go-error-handling`'s references;
  `util`/`common` package names to `go-naming`; embedding in public structs and
  generic-method interface satisfaction to `go-interfaces` (new ownership row);
  the handler exception to handle-once stated once in `go-error-handling` and
  routed from `go-logging`, `go-http`, and `go-context`; pprof capture out of
  BENCHMARKS.md into `go-troubleshooting`; `govulncheck` added to the gate row.
- Eleven Quick Reference tables that restated their skill's headings are gone
  (about 130 lines); the two facts that lived only there — the `t.Parallel()`
  default and the `Enabled()` guard — moved into sections. `go-security` keeps
  its table for the gosec IDs, with the false G401/G501 claim for
  `sha256.Sum256(password)` corrected.
- Two-default rules now name the trigger: constructor return type (interface
  only when pre-existing and consumer- or stdlib-owned), receiver type (the
  RECEIVER-TYPE list), preallocation (known `len` at the call site), request
  decoding (new endpoint → v2), logger placement (context for handler chains,
  parameter for libraries).
- `go-resilience` gains its first code: a compile-tested `retry` with capped
  jitter, a `Retry-After` floor, and cancellation-aware waiting.
- `go-functions` routes `ctx` placement to `go-context` and adds the
  `iter.Seq[T]` versus `[]T` return decision; `go-packages` gains `cmd/` and
  `internal/` layout guidance; `go-concurrency` and the symptom catalog name
  the GA `goroutineleak` profile.
- Every reference now starts with the provenance header (`Sources`,
  `Authority`, `Last verified`); 43 lacked it. `TestLongReferenceTOCs` enforces
  it, and `TestStructure` derives the `> Compatibility:` requirement from any
  inline `Go 1.NN` claim instead of a hardcoded list, which added the note to
  `go-code`, `go-linting`, `go-packages`, `go-performance`, `go-documentation`,
  and `go-functions`.
- Compile tests now cover every skill that ships Go: `skill_examples_test.go`
  adds 19 example tests (SSRF, TLS, retry, slogtest, the web server, goroutine
  patterns, struct tags, `os.Root`, self-referential constraints, interface
  assertions, test helpers, both assets, the `run` pattern, the rows loop, byte
  reuse). Eleven skills previously had none.
- `check-errors.sh` (v1.2.0) no longer flags a bare `return err` by default —
  the skill allows it when annotation adds nothing; `--bare-return` opts in.
- `hooks/go-vet-on-edit.sh` parses the payload as JSON, reports `go fix -diff`
  for the edited package (report only, never applied), and has a behavioral
  test, `TestVetHook`. `hooks/go-code-routing.sh` hints fire on
  decision-bearing syntax only (a `%w` wrap, `context.With*`, `go func`), not
  on every `fmt.Errorf` or `http.` token, and cover six more owners; the
  routing table in `go-code` stays authoritative.
- The bundled `golangci.yml` enables `modernize`; `setup-lint.sh` verifies the
  generated config before the first run. CI runs the eval suite with
  `-race -shuffle=on` and checks Markdown anchors.
- `docs/SKILL_AUTHORING_TEMPLATE.md` requires the provenance header on every
  reference and asks for short, paired examples that the `exampleBlock` tests
  compile.

## [1.6.1] - 2026-09-10

### Fixed

- `go-code-routing.sh` never learned that `go-code` was loaded under the
  plugin, because the Skill tool names a plugin skill `golang-skills:go-code`
  and the hook accepted only bare `go-*` names; the gate was silent in every
  plugin session. It now strips the plugin prefix. `TestRoutingGate` sends the
  prefixed name, and `docs/evidence/2026-09-10-routing-gate-sonnet-5.ru.md`
  records four Sonnet 5 sessions on 1.6.0 versus v1.5.0 plus the live session
  in which the corrected hook blocked once and the model recovered.

## [1.6.0] - 2026-09-10

### Added

- `hooks/go-code-routing.sh`, a routing gate for the Claude Code plugin. The
  `go-code` router asked the model to load `go-style-core` and the owner skills
  before the first edit, and the 2026-09-08 Sonnet 5 sessions in
  `docs/evidence` show that prose alone did not make it happen: 0 of 4 loaded
  `go-error-handling` before editing. The hook records loaded skills from
  `Skill` calls and direct `SKILL.md` reads, then blocks the first `.go` edit
  in a session that loaded `go-code` until `go-style-core` and the owners the
  edited content points at are loaded, naming them once per session so a retry
  always passes. `TestRoutingGate` drives one session through every branch;
  `TestHookScriptsSyntax` checks `hooks.json` against the scripts it names.

### Changed

- `go-code` states what loading a skill means on each host: the `Skill` tool
  in Claude Code, a read of the sibling `SKILL.md` in Codex. Step 2 loads
  `go-style-core` before reading the code; step 3 ends only when every
  selected owner is in context, before the first edit. The routing table moves
  out of Related Skills into its own section, Route Before The First Edit,
  which step 3 links to.

## [1.5.0] - 2026-09-10

### Changed

- Removed duplicated guidance inside four skills; no rule changed owner or
  meaning. `go-error-handling` stated the `%w` versus `%v` choice three times
  (an intro section, the Error Types table, and Error Wrapping); the intro now
  points at the two sections that own it. `go-naming` dropped the Quick
  Reference table, which restated every section heading above it, and merged
  two `go-style-core` pointers into one. `go-resilience` and `go-packages` no
  longer repeat the review-only, missing-skill, and approval rules that
  `go-style-core` owns under House Style Wins; `go-packages` links there
  instead.

## [1.4.0] - 2026-09-10

### Added

- Local-redirect guidance in `go-security` with a check that also rejects
  backslashes, spaces, and ASCII control characters. A `//`-prefix test alone
  accepts `/\host` and `/` + TAB + `/host`, which browsers resolve to an
  external origin, so the previous advice was an open redirect.
- `verify-refactor.sh baseline`/`after` report `lint_status`, `lint_exit_code`,
  and `lint_log_path` (v1.2.0). A lint run that exits non-zero without parsable
  diagnostics is `unavailable`, not clean, so an absent or misconfigured
  `golangci-lint` can no longer read as a passing gate. `passed` still covers
  only the core checks; callers must inspect `lint_status` and honor their own
  repository gate. `docs/SCRIPT_JSON_CONTRACTS.md` records the shape, and its
  `leaks` example now matches what the mode has emitted since 1.3.2.
- `TestLocalRedirectExample`, `TestContextHandlerFailureBoundaries`,
  `TestRefactorLintStatus`, `TestRefactorTruncation`, and
  `TestRefactorApplyLimitCounting`, which reproduce every defect corrected here
  against the shipped Markdown and script rather than against a copy of them.
  The truncation test now pins the exact boundary — `--limit` equal to the line
  count does not truncate — and the counting test drives `apply_limit` directly,
  because no mode passes it text ending in a newline.

### Changed

- Clone guidance names the actual contract: `slices.Clone` and `maps.Clone`
  preserve nilness, including a non-nil empty input; `slices.Collect` and
  `slices.Sorted` return nil for an empty iterator. The nil-to-`null` mapping
  is JSON v1's — v2 defaults encode nil non-byte slices as `[]` and
  `FormatNilSliceAsNull(true)` restores `null` — so `make`+`copy` is required
  only when the selected encoder or a caller contract needs it
  (`go-data-structures`, `go-defensive`, `go-http`).
- The HTTP error table separates an incoming `r.Context()` cancellation, which
  writes nothing, from a downstream `context.Canceled` while the request is
  still alive, which is a 500. Treating the two alike turned an unwritten
  response into an implicit 200.
- `go-code-refactor` defines `REFACTOR_SKILL_DIR` in its resource list, before
  the first command that uses it, and shows a `--version` call that proves the
  path resolved.

### Fixed

- The cancellation example in `go-context` returned `err.Error()` to the client
  and skipped the response whenever any operation reported `context.Canceled`,
  including a child context's. It now checks `r.Context().Err()`, logs the
  detail server-side, and returns a generic status; the encode error is no
  longer discarded.
- `verify-refactor.sh` reported `truncated: false` for a `diff` or `leaks`
  result that `--limit` had actually shortened: `apply_limit` set the flag
  inside a command substitution, so the subshell's value never reached the
  caller.
- The concision gate's first command used `$REFACTOR_SKILL_DIR` 42 lines before
  the only sentence that told you to set it. Unset, the path collapsed to
  `/scripts/verify-refactor.sh` and the baseline was never recorded, which is
  the failure the gate exists to prevent.
- `apply_limit` counted one line too many for text ending in a newline, so a
  complete blob could be reported as truncated. No mode reaches it today —
  all three strip the trailing newline through command substitution — but the
  count is now correct for any caller.

## [1.3.2] - 2026-09-10

### Added

- A shared JSON v2 boundary reference with bounded single-document decoding,
  compatibility options, nil collections, streaming/newline differences, and
  golden-test limits; routed from HTTP, package, defensive, and testing skills.
- An `errors.AsType` branch example that preserves the original error, with
  executable checks for second-branch matching and JSON input boundaries.
- `nolintlint` requiring named linters and explanations, and `usestdlibvars`
  for HTTP methods/status codes in the maintained lint configuration.

- `TestVersionClaimsMatchToolchain` and `TestAnalyzerToolAttribution`, which
  resolve two kinds of claim no build or link check can see: an inline
  `(Go 1.NN)` marker against `$GOROOT/api`, and an analyzer name against the
  tool that actually registers it. Both reproduce a real defect fixed below.
- Two `go-security` quality evals, covering the trust-boundary review (SQL
  identifier, path traversal, token comparison, credential in a log line) and
  the TLS config review. It was the only skill with no quality eval.

### Changed

- New-module guidance checks the actual `go` directive and local/CI toolchains
  after `go mod init`, without assuming its default or upgrading existing code.
- Modernization guidance distinguishes installed analyzer sets and calls out
  post-fix compilation, comment retention, and behavior-changing slice rewrites.
- The plugin manifests described all ten scripts as taking `--json`, `--limit`,
  and `--force`. Only `--json` is universal; the manifests now match the README,
  which was already accurate.
- **Breaking for callers of `verify-refactor.sh leaks`**: the mode no longer
  reports success for a passing test run, because it collects no leak profile.
  It exits 3 with `"leaks_checked": false`, keeping 0 for verified success, 1
  for failed tests, and 2 for a usage or environment error.
- Performance guidance states the direction of each optimization instead of
  fixed speedup factors, which were unattributed to any toolchain or workload.
  Clone guidance is conditional on the nil, capacity, and JSON contract rather
  than forbidding `make`+`copy` outright.

### Fixed

- Skill-review corrections preserve nil/empty collection contracts, distinguish
  `errors.Join` error-tree changes and NaN-sensitive modernization, and make
  the transfer example roll back when either account is absent.
- Security guidance no longer uses `Decoder.More` as an EOF check or treats
  path prevalidation as subprocess confinement. It distinguishes FIPS mode
  from module selection, parameterizes numeric SQL limits, and demonstrates
  checked AES setup with automatic GCM nonces.
- The transfer example mapped `sql.ErrNoRows` to no sentinel, contradicting the
  skill's own rule against leaking driver errors past a repository; it now
  returns `ErrNotFound`, and the regression test asserts that `sql.ErrNoRows`
  does not escape.
- Conversion benchmarks keep buffer size bounded. Executable regressions cover
  collection wire format, AES key errors, trailing JSON delimiters,
  missing-row rollback, and leak verification status.

- `MODERNIZATION.md` credited `waitgroupgo` to `go vet`. The vet analyzer is
  `waitgroup`; `waitgroupgo` belongs to `go fix`, and `go vet -waitgroupgo`
  fails with "flag provided but not defined". The Go 1.27 rename applied to the
  fix tool only.
- `OVER-ENGINEERING.md` dated `slices.Compact`, `Reverse`, `Max`/`Min` to Go
  1.22 along with `Concat`. Only `Concat` is 1.22; the rest are 1.21. This is
  the one inline claim the 1.3.1 conformance pass missed, so that release's
  "every inline claim matched" note was not quite true.
- `crypto/mldsa` and `http.Server.DisableClientPriority` sat in the
  `COMPATIBILITY.md` Go 1.27 table while no skill recommended them, against
  that file's own scope rule. Post-quantum signatures and the
  `CurvePreferences` downgrade trap now live in `go-security`, and the HTTP/2
  priority switch in `go-http`.
- Ukrainian comments in the shipped `golangci.yml` baseline, the only
  non-English text under `skills/`.
- Version markers missing from the `go fix` table rows for `testingcontext`,
  `rangeint`, `omitzero`, `stringsseq`, and `stditerators`, and from
  `maps.Keys` in `go-performance`.
- An empty `skills/go-code-review/references/` directory, invisible to git but
  present in every working tree that had it.

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
