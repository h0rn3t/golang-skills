# Agent Skills For Go

**English** | [Українська](README.uk.md)

AI [Agent Skills](https://agentskills.io/) for writing idiomatic,
production-quality **Go 1.27** code. 24 modular skills teach AI coding
assistants Go best practices derived from:

- [Google Go Style Guide](https://google.github.io/styleguide/go/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Go Wiki CodeReviewComments](https://github.com/golang/go/wiki/CodeReviewComments)

Skills are tuned following
[agentskills.io best practices](https://agentskills.io/skill-creation/best-practices):
content the agent already knows is omitted, procedural decision trees guide
multi-step tasks, 65 reference files load on demand via progressive disclosure,
10 bundled scripts automate common checks, and 5 asset templates ensure
consistent output. The Claude Code plugin also ships a `go-verify` subagent
that runs the verification gate and a PostToolUse hook that runs `gofmt` and
`go vet` on every edited `.go` file.

## Skills Included

| Skill | Description |
|-------|-------------|
| **go-code** | Router for a mixed Go task: loads the go-* skills it needs, closes with the gate |
| **go-code-refactor** | Behavior-preserving refactor of existing Go: audit, delete, restructure, modernize, verify |
| **go-code-review** | Systematic checklist for reviewing Go code and PR submissions |
| **go-concurrency** | Goroutine lifecycle, channels, mutexes, parallelization, thread-safety |
| **go-context** | Context.Context placement, cancellation, deadlines, request-scoped data |
| **go-data-structures** | Slices, maps, arrays — allocation with new vs make, append, copying |
| **go-database** | database/sql and ORMs — contexts on queries, rows lifecycle, transactions, N+1, pool settings |
| **go-defensive** | API boundary hardening, defer cleanup, Must functions, time handling |
| **go-documentation** | Doc comments, package docs, godoc formatting, runnable examples |
| **go-error-handling** | Error strategy decisions, wrapping (%v vs %w), sentinels, logging patterns |
| **go-functions** | Function APIs, ordering, signatures, Printf helpers, constructor config structs and functional options |
| **go-generics** | When to use generics, constraints, common pitfalls, type aliases |
| **go-http** | net/http handlers, ServeMux routing, middleware, server timeouts, shutdown, clients |
| **go-interfaces** | Interface design, abstractions, embedding, "accept interfaces return structs" |
| **go-linting** | Linters, golangci-lint setup, nolint directives, CI/CD integration |
| **go-logging** | Structured logging with slog, log levels, request-scoped context, migration |
| **go-naming** | Naming decision flow for packages, types, functions, variables, receivers |
| **go-packages** | Package organization, imports, package size, CLI/flag patterns |
| **go-performance** | String optimization, capacity hints, benchmarking, strconv over fmt |
| **go-resilience** | Retry budgets, idempotency, circuit breakers, bulkheads, rate limits, backpressure, graceful degradation |
| **go-security** | Trust-boundary threat model — injection, SSRF, secrets, constant-time compare, password hashing, TLS, cookies, redaction |
| **go-style-core** | Style baseline, formatting, nesting, declarations, initialization, scope, loops, switches, and enums |
| **go-testing** | Table-driven tests, subtests, test helpers, assertions, test organization |
| **go-troubleshooting** | Ticket and root-cause investigation, deployed version/config checks, data-flow tracing, panics, hangs, leaks, flaky tests |

### Consolidated Skill Names

Update explicit invocations and copied skill installations using this mapping:

| Previous name | Current name |
|---|---|
| `go-functional-options` | `go-functions` |
| `go-control-flow` | `go-style-core` |
| `go-declarations` | `go-style-core` |

The old skill directories are removed. For manual installations, replace them
with the current owner directories, including their references; remove stale
copies of the three retired directories to avoid duplicate discovery. The
always-loaded style entrypoint stays short, with syntax details read on demand.

## Bundled Scripts

10 scripts automate common Go checks. All support `--help`, `--json` for
structured output, and meaningful exit codes (0 = clean, 1 = issues found,
2 = error). Analysis scripts support `--limit` to cap output size, and
destructive scripts require `--force` to overwrite existing files.

| Script | Skill | Purpose |
|--------|-------|---------|
| `verify-refactor.sh` | go-code-refactor | Record baseline/after check results and diff them to prove behavior held |
| `check-debt.sh` | go-code-refactor | Harvest `Kept:` shortcut markers into a ledger and flag those naming no upgrade path |
| `pre-review.sh` | go-code-review | Run gofmt + go vet + golangci-lint before review |
| `check-naming.sh` | go-naming | Detect SCREAMING_SNAKE, Get-prefixed getters, bad package names |
| `check-docs.sh` | go-documentation | Find exported symbols missing doc comments |
| `check-errors.sh` | go-error-handling | Catch bare returns, string comparison on errors, log-and-return |
| `check-interface-compliance.sh` | go-interfaces | Find interfaces missing compile-time verification |
| `bench-compare.sh` | go-performance | Run benchmarks with optional benchstat comparison |
| `setup-lint.sh` | go-linting | Generate .golangci.yml with recommended linters |
| `gen-table-test.sh` | go-testing | Scaffold a table-driven test file |

## Bundled Agent and Hook

Installed with the Claude Code plugin (Option 2 below); `npx skills` copies
skills only.

| File | What it does |
|------|--------------|
| `agents/go-verify.md` | Opt-in Claude agent for requested checks: "check it builds" selects build; "run the gate" selects the full gate. Reports findings and unavailable checks, with `INCOMPLETE` when required evidence is missing. Routine checks stay inline unless the user or host requests delegation |
| `hooks/go-vet-on-edit.sh` | PostToolUse hook: after every `Edit`/`Write` of a `.go` file it runs `gofmt -l` and `go vet` on that package and hands the findings back to the agent. Silent when clean; never blocks the edit |

The shared instructions support Claude and GPT-6 without a model-specific
fork. `go-style-core` owns user-scope precedence, progress updates, report
length, and host-controlled delegation. `go-linting` requires observed results
and scales verification to the work. These rules incorporate
[OpenAI's GPT-6 Astra prompting guidance](https://developers.openai.com/api/docs/guides/latest-model/gpt-6-astra.md#prompting-best-practices).
See the [review and validation limits](docs/CROSS_MODEL_REVIEW.md).

### Codex

Install the skill directories under `.agents/skills/` in a project or
`~/.agents/skills/` for your user, as described in the
[official skills documentation](https://learn.chatgpt.com/docs/build-skills).
Invoke `$go-code <task>` or an individual skill such as `$go-error-handling`.
Install the whole pack for the router's sibling references. With a partial
installation, it reports missing guidance and continues using available skills.
Resolve scripts relative to the installed skill and run them against the target
project. Codex does not need this repository's Claude agent or PostToolUse hook;
the selected checks run directly when no hook output is available.

## Installation

### Option 1: npx skills (Recommended)

The easiest way to install across **any** AI coding agent. Supports Cursor,
Codex, OpenCode, Cline, GitHub Copilot, Windsurf, Roo Code, and [25+ more
agents](https://github.com/vercel-labs/skills#supported-agents).

```bash
# all 24 skills
npx skills add h0rn3t/golang-skills --all

# or pick individual skills
npx skills add h0rn3t/golang-skills go-error-handling go-testing
```

Run it from your project root — skills land in the current agent's directory
(e.g. `.cursor/rules/`, `.github/copilot/skills/`).

### Option 2: Claude Code (plugin)

```bash
# Add the marketplace (one time)
/plugin marketplace add h0rn3t/golang-skills

# Install the skills
/plugin install golang-skills@golang-skills
```

Verify with `/plugin` — `golang-skills` should be listed as enabled. Skills
activate automatically once you touch Go code.

Update and remove:

```bash
/plugin marketplace update golang-skills
/plugin uninstall golang-skills@golang-skills
```

### Option 3: Manual install (Claude Code / Agent Skills)

Use this when you want only a few skills, or have no marketplace access.

```bash
git clone https://github.com/h0rn3t/golang-skills.git
cd golang-skills

# all skills, for your user
cp -R skills/go-* ~/.claude/skills/

# or scoped to one project
cp -R skills/go-* /path/to/project/.claude/skills/

# or a single skill
cp -R skills/go-error-handling ~/.claude/skills/
```

Each skill is a self-contained directory: `SKILL.md` plus `references/`,
`scripts/`, `assets/`. Copy the whole directory — relative links inside
`SKILL.md` break otherwise.

Make the scripts executable:

```bash
chmod +x ~/.claude/skills/go-*/scripts/*.sh
```

To uninstall: `rm -rf ~/.claude/skills/go-*`.

### Option 4: Cursor (Native Remote Rule)

1. Open **Cursor Settings** (Cmd+Shift+J on Mac, Ctrl+Shift+J on Windows/Linux)
2. Navigate to **Rules** → **Add Rule** → **Remote Rule (Github)**
3. Enter: `https://github.com/h0rn3t/golang-skills`

### Prerequisites

The skills themselves are Markdown and need nothing. The bundled scripts shell
out to the standard Go toolchain:

| Tool | Used by | Install |
| --- | --- | --- |
| Go 1.26+ (1.27 targeted) | `gofmt`, `go vet`, `go test`, `go fix` | [go.dev/dl](https://go.dev/dl/) |
| `golangci-lint` | `pre-review.sh`, `setup-lint.sh` | `brew install golangci-lint` |
| `govulncheck` | vulnerability gate | `go install golang.org/x/vuln/cmd/govulncheck@latest` |
| `benchstat` (optional) | `bench-compare.sh` comparison | `go install golang.org/x/perf/cmd/benchstat@latest` |

## How It Works

These skills follow the [Agent Skills open standard](https://agentskills.io/),
which works across multiple AI coding tools. When you're writing Go code:

1. **Automatic activation**: The AI agent loads relevant skills based on context
   (e.g., `go-naming` when you're writing a new function)
2. **Procedural guidance**: Decision trees and step-by-step procedures for
   multi-step tasks like code review and error strategy selection
3. **Progressive disclosure**: Core rules load immediately; 65 reference files
   load on demand when specific situations arise
4. **Automation**: 10 bundled scripts handle repetitive checks so the agent
   focuses on higher-level guidance
5. **Conditional cross-references**: Skills link to each other with "when"
   conditions to avoid unnecessary context loading
6. **Rule ownership**: `docs/RULE_OWNERSHIP.md` keeps duplicated guidance out
   of non-owner skills
7. **Verification gate**: `go-linting` owns one checkable definition of "done"
   — `gofmt`, `go vet`, `go test -race`, `go fix -diff`, `golangci-lint`,
   `govulncheck` — that the other skills route to instead of inventing their own
8. **Lint-enforced rules**: the `go-linting` baseline `.golangci.yml` enables
   the linters that check what the skills teach — `depguard` for the dependency
   ladder, `sloglint`, `errorlint`, `errname`, `noctx`, `rowserrcheck`,
   `perfsprint`, `usetesting`, `godot` — so the gate catches drift instead of a
   reviewer

## Running the Evals

`evals/evals.json` holds 105 trigger evals (does the right skill fire for this
prompt?) and 44 quality evals (does the answer satisfy each assertion?). The Go
tests in `evals/` validate their schema on every push; running them against a
model is opt-in because it costs tokens:

```bash
cd evals
go run ./cmd/evalrun -set validation -kind all -j 2 -out evals-results.json
```

The runner loads the checked-out skills with `claude --plugin-dir`, runs each
prompt in an empty scratch directory with the tool set cut down to `Skill`
(plus read-only tools for quality evals), records which `go-*` skills the model
invokes for each trigger prompt, runs each quality prompt to completion, and
has a second model grade the answer against the eval's assertions. A failing
trigger eval means the model read the skill descriptions and chose not to load
the expected one — the signal to tune that description. In CI, trigger the `Validate Skills` workflow manually
with **run_evals** checked; it needs an `ANTHROPIC_API_KEY` secret.

The automated runner is Claude-only; `-model` does not switch providers.
Quality cases 22–27 also cover scope, unavailable checks, evidence reuse,
requirements, characterization tests, and skill-only installations. Run these
prompts in fresh GPT-6/Codex sessions using the checked-out skills and record
the answers separately; the [review](docs/CROSS_MODEL_REVIEW.md) distinguishes
these application probes from a full cross-model benchmark.

## Measured Effect Across Models

Trigger and quality evals ask whether the right skill fires and whether the
answer reads well. Neither says what the skills do to the code the model
actually writes. `evals/ab` answers that: it hands the same fixture to the same
model twice, once with no plugin loaded and once with the whole skill tree, then
measures the resulting Go — not the prose — against a golden test the model
never sees. Every table below is computed from the raw JSON reports in
[`docs/evidence/`](docs/evidence), seed `1`, 5 repetitions per fixture per arm
in the per-model runs and 10 in the runner and wording runs.

### Summary in percent

One row per thing a skill is supposed to do to the code. "Control" is the arm
with no plugin loaded; every percentage is skill relative to control on the
same model, runner and fixtures.

| What the skills do | Effect | Basis |
| --- | --- | --- |
| **Refactor: stop the model adding structure** on the trap fixture `report` | Growth cut by **50–60%** on Opus 5 and MiniMax M3 and by **94%** on GPT-5.6-Luna at n=5; **84–106%** on GPT-5.6-Luna across codex, copilot and opencode at n=10 (above 100% = the package ends smaller than it was handed) | 4 models, 4 runners; [per-model](docs/evidence/2026-09-07-go-refactor-control-gpt-5.6-luna-medium.md), [multirunner](docs/evidence/2026-09-07-go-multirunner-gpt-5.6-luna-medium.uk.md), [codex n=10](docs/evidence/2026-09-08-selection-once-luna-codex.uk.md) |
| **Refactor: stop helper sprawl** (new functions across 20 sessions) | **−38% to −71%** per model at n=5; **−84% to −88%** on GPT-5.6-Luna on each of three runners | 3 of 4 models; MAI-Code-1.1-Flash +27%, the one model that never took the bait |
| **Refactor: finish removing duplication** on `pricing`, lines removed vs control | Before the 2026-09-08 change the skill removed **32–40% less** than control on all three runners; after it, **36% more** on codex, **11% more** on copilot, **±0%** on opencode where control already got there | GPT-5.6-Luna, n=5 before / n=10 after; [copilot](docs/evidence/2026-09-07-selection-once-luna-copilot.uk.md), [codex](docs/evidence/2026-09-08-selection-once-luna-codex.uk.md), [opencode](docs/evidence/2026-09-07-selection-once-luna-opencode.uk.md) |
| **Refactor: keep behavior** | Build and hidden golden test **100%** in both arms on every model and runner | 160 + 240 + 230 sessions; the corpus measures size, not defects |
| **Implement: catch the hidden defect** (packages passing the golden spec) | **75% → 90%** on MAI-Code-1.1-Flash; `gateway` alone **20% → 60%**; **100% → 100%** on GPT-5.6-Luna and Opus 5, which never fall in | n=5 per cell, Fisher p ≈ 0.5 — a direction, not a demonstration |
| **Implement: size of a working implementation** where correctness is tied | **−35% lines, −44% functions** on Opus 5 `gateway`, spread seven times tighter; **±0%** on GPT-5.6-Luna, no interval excludes zero | [Opus 5](docs/evidence/2026-09-07-go-implement-gateway-opus5.md), [multirunner](docs/evidence/2026-09-07-go-multirunner-gpt-5.6-luna-medium.uk.md) |

Read it as two findings and one repair. The skills reliably prevent growth —
that is the effect with intervals excluding zero on every runner. They did not,
until 2026-09-08, make the model delete more than it would alone, and on
duplication that wants a data table they made it delete less; the
«Remove Duplication to the End» section in `go-code-refactor` closes that gap on
the two runners where it existed. On new code the skills help the models that
fall into the trap and are neutral on the ones that do not.

### Refactor corpus: does the skill remove structure?

Four working packages, each with an honest refactor that removes structure and a
tempting one that adds it. `report` is the trap fixture — two output formats that
invite a `Formatter` interface no caller needs. Values are the mean line delta
with the skill minus the mean without it, so **negative favors the skill**.

| Model | Runner | `dispatch` | `pricing` | `report` | `store` | Corpus |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| [GPT-5.6-Luna (medium)](docs/evidence/2026-09-07-go-refactor-control-gpt-5.6-luna-medium.md) | codex | **−5.8** | +2.6 | **−17.6** | −3.6 | **−6.10** |
| [MiniMax M3](docs/evidence/2026-09-07-go-refactor-control-minimax-m3.md) | opencode | +0.8 | −6.8 | **−16.4** | −9.4 | **−7.95** |
| [Opus 5](docs/evidence/2026-09-07-go-refactor-control-opus5.md) | claude | +2.0 | −0.4 | **−16.6** | −1.6 | **−4.15** |
| [MAI-Code-1.1-Flash](docs/evidence/2026-09-07-go-refactor-control-mai-code-1.1-flash.md) | copilot | +5.8 | −2.6 | −0.6 | +1.8 | +1.10 |

The one model where nothing moved is the one that never took the bait. Read the
control column first — it is how much there was to remove:

| Model | `report` without the skill | With it | Removed |
| --- | ---: | ---: | ---: |
| Opus 5 | +33.4 | +16.8 | 50% |
| MiniMax M3 | +27.4 | +11.0 | 60% |
| GPT-5.6-Luna | +18.8 | **+1.2** | **94%** |
| MAI-Code-1.1-Flash | +16.2 | +15.6 | 4% |

MAI-Code-1.1-Flash grows the package half as much as Opus 5 does unaided and
declares one new type across 20 sessions; there is no over-engineering there for
the skill to prevent. On GPT-5.6-Luna the skill does not merely shrink the
growth — four of its five `report` sessions returned a package *smaller* than
the one they were handed, with the hidden golden test still green.

The mechanism differs by model, and the pair of counts is what tells premature
abstraction apart from helper sprawl:

| Model | New types | New functions |
| --- | --- | --- |
| Opus 5 | 15 → 6 (**−60%**) | 32 → 20 (−38%) |
| GPT-5.6-Luna | 1 → 1 | 31 → 9 (**−71%**) |
| MiniMax M3 | 5 → 6 | 31 → 16 (−48%) |
| MAI-Code-1.1-Flash | 1 → 0 | 37 → 47 (+27%) |

Opus 5's failure mode is reaching for a type; everyone else's is reaching for a
helper. Across all 160 valid sessions exactly one interface was declared — by
Opus 5's control arm, none by any skilled arm — and no arm anywhere produced a
pattern-flavored name. Correctness was tied on every model: 20/20 build and
golden passes in both arms. The refactor corpus is
evidence about code size, not about defect rates.

#### The same model on three runners, and what it changed in the skill

Running GPT-5.6-Luna through codex, opencode and copilot on the same fixtures
([report](docs/evidence/2026-09-07-go-multirunner-gpt-5.6-luna-medium.uk.md),
n=5) kept the two effects above — `report` growth removed on every runner,
helper sprawl down 84–88% — and broke the per-fixture means: `pricing` and
`store` changed sign between runners, and the corpus mean is not a number to
publish at n=5. It also surfaced the one result against the plugin: on
`pricing`, a package whose duplication wants one data table, the skill stopped
the model at a `switch` and removed fewer lines than control on all three
runners, with all three intervals excluding zero.

Two wordings were tried as variant arms at n=10. A permission to add a table
moved nothing. A completion criterion — each literal once, each selection over
the same key once, each condition ladder once, table not to be serviced —
became the «Remove Duplication to the End» section of `go-code-refactor` on
2026-09-08:

| Runner | `pricing`, skill vs control, before | After | `report` after |
| --- | ---: | ---: | ---: |
| [codex](docs/evidence/2026-09-08-selection-once-luna-codex.uk.md) | +11.4 (+3.6 … +19.2) | **−12.4 (−18.0 … −6.8)**, table in 9/10 | −1 in 17/20, three helper-extraction outliers |
| [copilot](docs/evidence/2026-09-07-selection-once-luna-copilot.uk.md) | +13.8 (+4.1 … +23.5) | **−11.6 vs old skill (−18.9 … −4.3)**, table in 7/10 | −1 in 9/10 |
| [opencode](docs/evidence/2026-09-07-selection-once-luna-opencode.uk.md) | +16.8 (+5.5 … +28.1) | +0.1 vs old skill (−7.7 … +7.9) | −1 in 9/10 |

The opencode row is not a failure of the text: three later n=10 samples showed
the old skill never trailed control there (+1.3, −0.4), so there was nothing to
repair. The codex comparison with the old skill is across runs; the copilot one
is the same-run variant arm. `report` stayed under protection on all three
(−10.8 to −16.5 vs control, intervals excluding zero).

### Implementation corpus: does the skill make the code work?

Documented but unimplemented packages, where the golden test *is* the
specification and can be failed outright. Each fixture hides one defect a Go
reviewer would send back — a nil slice that marshals to `null`, a server with no
timeouts, an error chain cut with `%v`, a snapshot that still aliases the
caller's slice — and the doc comments never name the technique.

| Model | Runner | `catalog` | `feed` | `gateway` | `ledger` | All |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| [GPT-5.6-Luna (medium)](docs/evidence/2026-09-07-go-implement-control-gpt-5.6-luna-medium.md) | codex | 5/5 → 5/5 | 5/5 → 5/5 | 5/5 → 5/5 | 5/5 → 5/5 | 20/20 → 20/20 |
| Opus 5 ([gateway](docs/evidence/2026-09-07-go-implement-gateway-opus5.md), [feed/catalog](docs/evidence/2026-09-07-go-implement-feed-catalog-opus5.md)) | claude | 3/3 → 3/3 | 3/3 → 3/3 | 5/5 → 5/5 | — | 11/11 → 11/11 |
| [MAI-Code-1.1-Flash](docs/evidence/2026-09-07-go-implement-control-mai-code-1.1-flash.md) | copilot | 5/5 → 5/5 | 4/5 → 5/5 | **1/5 → 3/5** | 5/5 → 5/5 | 15/20 → 18/20 |

Where the model already avoids the defect, the skill has nothing to add and the
score is what a working implementation costs instead: on Opus 5 `gateway` went
from 152.6 lines to 99.8 with correctness tied at 5/5, functions down 44% and the
skilled arm's spread seven times tighter.

Where the model falls in, the score is whether the package works at all.
`gateway` — an edge server built as `&http.Server{Addr: addr, Handler: h}`, whose
zero timeouts hold stalled connections until it runs out — goes 1/5 to 3/5 on
MAI-Code-1.1-Flash. Every failure in both arms is that one defect: `ReadTimeout`
or `ReadHeaderTimeout` left at zero.

At `n=5` per cell Fisher's exact gives p ≈ 0.5, so that cell is a direction
rather than a demonstrated result, and `gateway` is the fixture worth a larger
`n`.

### What this does and does not establish

It compares the complete current skill tree against no skill at all, which is
what decides whether a fixture contains a trap the plugin can catch. With one
exception it is not a before/after measurement of a wording change; that claim
needs a variant or `reference` arm, and the 2026-09-08 change above is the one
that has it. Percentages in the summary are ratios of means at n=5 or n=10 and
carry the intervals of the tables they come from; a single cell without an
interval is a direction. The runners differ in ways that matter across files —
tool sets, whether the session has a shell, whether the plugin's hook and
subagent apply — and those differences are recorded in
[`evals/ab/README.md`](evals/ab/README.md). Every table here has its raw JSON
report and a SHA-256 beside it; the repository carries no result without one.

## Go 1.27

Skills target Go 1.27 and say so where it matters. Notable guidance that
changed with recent releases:

| Guidance | Skill |
|---|---|
| Generic methods; `maphash.ComparableHasher` | go-generics |
| `errors.AsType[T]` over `errors.As` | go-error-handling |
| `httptest.NewTestServer`, `synctest`, `t.Context` | go-testing |
| Stdlib `uuid` and `encoding/json/v2` on the dependency ladder | go-packages |
| `go fix` modernizers as part of the gate | go-linting, go-style-core |
| `slices.Clone`/`maps.Clone` and `os.Root` at boundaries | go-defensive |
| `slog.NewMultiHandler`, `slog.GroupAttrs` | go-logging |
| `new(expr)` for non-composite pointers | go-style-core |
| `ServeMux` method patterns, `http.NewCrossOriginProtection`, `MaxHeaderValueCount` | go-http |
| `sql.Null[T]` for nullable columns | go-database |
| `crypto/pbkdf2`, `crypto/hkdf`, `crypto/sha3` in stdlib; `rsa.EncryptPKCS1v15` deprecated for new designs | go-security |
| `runtime/trace.FlightRecorder`, `debug.SetCrashOutput` | go-troubleshooting |

Version-sensitive claims are tracked in [COMPATIBILITY.md](COMPATIBILITY.md)
and pinned by `TestGoVersionBaseline` in `evals/eval_test.go`.

## Project Structure

```
.
├── skills/
│   └── go-*/
│       ├── SKILL.md      # Core rules (< 500 lines each)
│       ├── references/   # Detailed guidance, loaded on demand
│       ├── scripts/      # Automation scripts and helpers
│       └── assets/       # Output templates (5 skills)
├── agents/               # go-verify subagent (Claude Code plugin)
├── hooks/                # PostToolUse gofmt/vet hook (Claude Code plugin)
├── evals/
│   ├── evals.json        # Trigger and quality eval definitions
│   ├── cmd/evalrun/      # Opt-in headless eval runner (claude -p)
│   ├── files/            # Sample Go files for quality evals
│   └── fixtures/         # Test fixtures for script/eval coverage
├── docs/                 # Repository maintenance notes
├── .github/workflows/    # CI validation
└── source/               # Original style guide sources
```

## Provenance and Compatibility

Bundled upstream source snapshots live under `source/`. Each source file keeps
its own inline provenance header, and [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)
summarizes the source path, upstream project, URL, license, and copyright at the
repository level.

Go-version-sensitive guidance is tracked in [COMPATIBILITY.md](COMPATIBILITY.md),
which also documents how to re-verify every claim against an installed
toolchain. When a skill recommends a standard-library API tied to a specific Go
release, the guidance names the minimum version — and names a fallback only
when the API is newer than the oldest supported release (currently 1.26).

## License

Project-authored skill files, scripts, assets, docs, and evals are licensed
under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.
Bundled upstream snapshots under `source/` retain their upstream licenses; see
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
