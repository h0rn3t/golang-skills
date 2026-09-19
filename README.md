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
multi-step tasks, 74 reference files load on demand via progressive disclosure,
11 bundled scripts automate common checks, and 5 asset templates ensure
consistent output. The Claude Code plugin also ships a `go-verify` subagent
that runs the verification gate, a PostToolUse hook that runs `gofmt`,
`go vet`, `go fix -diff`, the package's tests, and `golangci-lint` on every
edited `.go` file, a routing gate that holds the first
Go edit until the router (`go-code`, `go-code-refactor`, or `go-code-review`)
has loaded `go-style-core` and the owner skills and read the idiom card, a
prompt hook that names `go-code` or `go-code-refactor`, the owners, and the
card when a prompt asks for Go work, and a subagent hook that repeats that note to every subagent started
in a Go project.

## Skills Included

| Skill | Description |
|-------|-------------|
| **go-code** | Router for a mixed Go task: loads the go-* skills it needs, closes with the gate |
| **go-code-refactor** | Behavior-preserving refactor of existing Go: audit, delete, restructure, modernize, verify; a target architecture for a monolith |
| **go-code-review** | Systematic checklist for reviewing Go code and PR submissions |
| **go-concurrency** | Goroutine lifecycle, channels, mutexes, parallelization, thread-safety |
| **go-context** | Context.Context placement, cancellation, deadlines, request-scoped data |
| **go-data-structures** | Slices, maps, arrays — allocation with new vs make, append, copying |
| **go-database** | database/sql and ORMs — contexts, rows, transactions, N+1, pools; PostgreSQL constraints, indexes, and live migrations |
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

11 scripts automate common Go checks. All support `--help`, `--json` for
structured output, and meaningful exit codes (0 = clean, 1 = issues found,
2 = error). Analysis scripts support `--limit` to cap output size, and
destructive scripts require `--force` to overwrite existing files.

| Script | Skill | Purpose |
|--------|-------|---------|
| `verify-refactor.sh` | go-code-refactor | Record baseline/after check results and diff them to prove behavior held |
| `check-debt.sh` | go-code-refactor | Harvest `Kept:` shortcut markers into a ledger and flag those naming no upgrade path |
| `check-architecture.sh` | go-code-refactor | Check `internal/` imports against the layered-module policy — ownership, layer, composition, platform, contract, driver rules — with an exact `known` list that fails when stale |
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
| `hooks/go-vet-on-edit.sh` | PostToolUse hook: after every `Edit`/`Write` of a `.go` file it runs `gofmt -l`, `go vet`, `go fix -diff` (report only), the package's own tests (`go test -short -count=1`, once the package type-checks and has test files), and `golangci-lint` (the repository's configuration, else the bundled one; findings in the edited file are listed, the rest of the package is one count) on that package and hands the failures back to the agent. The hook is the host's process, so a session without a shell still sees what the gate would have said. `GOLANG_SKILLS_EDIT_TESTS=off` and `GOLANG_SKILLS_EDIT_LINT=off` switch the two off. Silent when clean; never blocks the edit |
| `hooks/go-prompt-routing.sh` | UserPromptSubmit hook: when a prompt asks for Go work — it names Go, a `.go` file or `go.mod`, or is sent from a directory holding Go and names a function, package, handler, or test — it adds one note to the model's context naming the router to load before the first edit: `go-code-refactor` for refactor, clean-up, or simplify wording, `go-code` for anything else — with `go-style-core` and, when the prompt names a package or file, the owner skills its code points at, from the same table the gate uses (`go-code-routing.sh --hints`), so the loads happen before the first edit rather than one gate block at a time; for `go-code` and `go-code-refactor` it also names the idiom card (`go-style-core/references/CURRENT-GO.md`) by its installed path for a whole `Read` in the same message, since no skill wording made Sonnet 5 read it. Once per skill per session; silent when the session already loaded it, when the prompt names a go-* skill in prose, or when there is no work verb. A slash invocation of a router (`/go-code`, `/golang-skills:go-code-refactor`) gets a note of its own: the host inserts the skill file and calls no tool, so this hook also records the router in `loaded`, which is what arms the edit gate. Never blocks |
| `hooks/go-code-routing.sh` | Routing gate for the `go-code`, `go-code-refactor`, and `go-code-review` routers. PostToolUse on `Skill` and `Read` records which go-* skills the session loaded; PreToolUse on `Edit`/`Write` of a `.go` file, in a session that loaded one of them, blocks the edit (exit 2) until `go-style-core` and the owners the edited content points at are loaded and the idiom card has been `Read` whole (no offset, no limit shorter than the file), naming the router, the missing skills, and the card's installed path. Each name is used once per session, so a retry always passes. Silent in sessions that loaded none of the three |
| `hooks/go-subagent-routing.sh` | SubagentStart hook: a subagent starts with an empty context, so when the working directory holds Go it adds one note naming `go-code`, or `go-code-refactor` for a refactor, to load before the first edit. Fires for every subagent; skips the plugin's own `go-verify` agent; never blocks |

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
project. Codex has no skill tool and no hooks: `go-code` tells it to read each
selected sibling `SKILL.md` directly before the first edit, and the selected
checks run directly when no hook output is available.

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

### Pinning a version

Every release is a git tag (`v1.21.1`). None of the installers above takes a
version argument — `npx skills add` and `/plugin marketplace add
h0rn3t/golang-skills` both follow the default branch, so they always give you
the newest release. To pin one, install from a tagged checkout:

```bash
git clone --branch v1.21.1 --depth 1 https://github.com/h0rn3t/golang-skills.git
cd golang-skills

# manual install from this checkout
cp -R skills/go-* ~/.claude/skills/

# or register the checkout itself as the marketplace, in Claude Code
/plugin marketplace add /absolute/path/to/golang-skills
/plugin install golang-skills@golang-skills
```

The tag stays put, so the skills stay put: nothing moves until you clone a
different one. Useful commands:

```bash
# every released version
git ls-remote --tags https://github.com/h0rn3t/golang-skills.git

# which version a checkout or install is
grep '"version"' .claude-plugin/plugin.json

# move an installed plugin to the newest release
/plugin marketplace update golang-skills
```

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
3. **Progressive disclosure**: Core rules load immediately; 74 reference files
   load on demand when specific situations arise
4. **Automation**: 11 bundled scripts handle repetitive checks so the agent
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

`evals/evals.json` holds 108 trigger evals (does the right skill fire for this
prompt?) and 62 quality evals (does the answer satisfy each assertion?). The Go
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

## Do the skills help this model?

**The clearest benefits are in refactoring — with GPT-5.6-Luna, Opus 5,
MiniMax M3, and now Sonnet 5 at medium reasoning effort. There is no convincing
benefit for writing new code yet. Reviewing code has its first reading, on
2026-09-12: the skill moves Sonnet 5's recall on seeded defects and has no room
to move Opus 5's; on 2026-09-18 cutting the checklist rows the linters already
report moved that recall further, 0.72 to 0.80. DeepSeek V4.1-Flash joined the
set on 2026-09-19 as the cheap carrier: the routers fire on it, refactoring and
reviewing have room, writing new code has none at 7/7 golden in both arms.**

✅ **Helps on tested tasks** · 🟡 **Mixed / weak signal** ·
➖ **Benefit not established** · — **Not tested**.
These are practical interpretations of the tests, not guarantees for every project.

| Model / tool | Refactoring existing code | Writing new code | Reviewing code | Cost with skills | Practical takeaway |
|---|---|---|---|---|---|
| **GPT-5.6-Luna / Codex** | [✅ 20/20 golden; −6.8 lines and −1.75 functions vs control](docs/evidence/2026-09-13-gpt-5-6-luna-medium-controls.md) | [➖ Smaller, but 26/30 golden vs 27/30 control](docs/evidence/2026-09-13-gpt-5-6-luna-medium-controls.md) | — Not tested | Not measured in USD | **Use for refactoring; do not claim a new-code gain.** |
| **Opus 5 / Claude** | [✅ Helps: less unnecessary structure, especially on `report`](docs/evidence/2026-09-07-go-refactor-control-opus5.md) | [🟡 At medium effort: 9/9 correct in both skilled arms against 7/9 unaided, both unaided failures `gateway` answering `HEAD` with 200 at 110–147 lines; −6.7, −10.3 and −68.7 lines per task against no skills, none significant at n = 3 on three tasks](docs/evidence/2026-09-11-go-implement-newcode-workflow-opus-5-medium.md); [the 2026-09-12 Declaration Budget rule moves the two `gateway` helpers out of closures in 5/5 sessions at +7 lines, golden 5/5 in both skilled arms](docs/evidence/2026-09-12-go-implement-budget-closure-gateway-n5-opus-5-medium.md); [across efforts at n = 5 the correctness effect holds — unaided `gateway` 1/5 at low and 1/5 at high against 5/5 with skills — while the size effect appears only at high: `catalog` −4.2 (p = 0.008), `feed` −6.4 (p = 0.016), against −1.2 and +2.8 at low](docs/evidence/2026-09-12-go-implement-effort-sweep-opus-5.md); [the Delete Pass takes `gateway` from 87.8 to 75.8 lines against 1.13.0 (−12.0, p = 0.008), golden 5/5 in both arms, every baseline session lint-clean against 3/5](docs/evidence/2026-09-12-go-implement-delete-pass-gateway-n5-opus-5-medium.md); [the edit hook that runs tests and lint after every edit changes nothing on a fixture the model already gets right — golden 5/5 and lint-clean 5/5 in both arms, lines level — at +8% cost](docs/evidence/2026-09-12-go-implement-edit-hook-gateway-n5-opus-5-medium.md); [split into three arms at n = 1 — 1.14.0, the hook alone, the hook with the three skill sentences — every arm is 2/2 golden and lint-clean on `feed` and `gateway` at $1.03, $0.94 and $0.86 a session; the arms cannot be told apart on a model with nothing to fix](docs/evidence/2026-09-13-go-implement-hook-vs-text-feed-gateway-n1-opus-5-medium.md); [the edit hook's check list stated once in `go-style-core` instead of three times, against 1.20.1 at n = 2 on `feed` and `gateway`: golden 4/4 and lint-clean 4/4 in both arms, $0.989 against $1.037 a session, the idiom card read 8/8 from the routing line alone](docs/evidence/2026-09-18-go-implement-hook-record-feed-gateway-n2-opus-5-medium.md) | [🟡 First reading, n = 1: unaided Opus 5 finds all 29 seeded defects in three fixtures and the skill 28, so the review corpus has no room on this model until a fixture carries defects it misses; the skilled reviews carry `verified`/`plausible` markers (8.7 a session against 0) at 1.9× the cost](docs/evidence/2026-09-12-go-review-corpus-smoke.md); [three fixtures built to carry defects it misses — evidence split across two files, a bug only a simulated call sequence shows, an enumeration one field short, a passing test that contradicts the contract — give unaided recall 0.97 over 44 defects at n = 2 in both arms; the one class that held is a test that passes for the wrong reason (0/4); the room on this model is precision and filing: baits 0.38, unkeyed 0.27 against 0.35 with the skill, must defects filed as Must Fix 0.84 against 0.78, at 1.6× the cost](docs/evidence/2026-09-12-go-review-corpus-hard-opus-5-medium.md) | New code ≈ **3.5×** at medium (n = 3), ≈ 5.7× at low and ≈ 3.9× at high (n = 5); refactoring unavailable | **Worth using for refactoring; new code correct in every skilled session, too few repetitions to claim a size effect.** |
| **MiniMax M3 / OpenCode** | [✅ Helps: less code and fewer unnecessary helpers](docs/evidence/2026-09-07-go-refactor-control-minimax-m3.md) | [➖ Benefit unproven: fewer passing sessions](docs/evidence/2026-09-07-go-implement-discovery-minimax-m3.md) | — Not tested | Refactoring ≈ **2.5×**, new code ≈ **1.9×** | **Better refactoring, not a cost saving.** |
| **DeepSeek V4.1-Flash / OpenCode** | [✅ −18.2 lines against −4.8 unaided, golden 4/4 in both arms, `go-code-refactor` loaded in 4/4 sessions; one repetition per fixture and the whole gap in `report`, where the unaided session added five helpers to a refactoring task](docs/evidence/2026-09-19-go-three-corpora-control-n1-deepseek-v4.1-flash.md) | [➖ No room on the corpus: the hidden golden test passes 7/7 in both arms, so nothing here is left for a wording to fix; +56.0 lines against +74.0, almost all of it `fetch` (+138 against +220), `go-code` loaded in 5/7 sessions and silent on `catalog` and `ledger`](docs/evidence/2026-09-19-go-three-corpora-control-n1-deepseek-v4.1-flash.md) | [🟡 First reading, n = 1: recall 69/76 against 65/76, must-severity defects 31/31 against 29/31 and 27 against 22 of them filed as Must Fix, off-key citations 27 → 13, evidence markers 47 against 5 — but six correct bait lines flagged against four; two of the six defects only the skill found are tests that pass for the wrong reason](docs/evidence/2026-09-19-go-three-corpora-control-n1-deepseek-v4.1-flash.md) | Refactoring ≈ **3.6×**, new code ≈ **3.0×**, review ≈ **1.6×**, on a $0.005–0.016 session | **The cheap carrier: the routers fire here, where `deepseek-v4-flash` never loaded `go-code` at all, and a 34-session sweep of all three corpora costs $0.34. Read the numbers as a firing smoke at n = 1, not as an effect.** |
| **Sonnet 5 / Claude** | [✅ Helps at medium effort: −9.7 lines per task, three of four tasks improve, two of them still after correcting for four comparisons, correctness tied at 20/20 (n = 5)](docs/evidence/2026-09-10-go-refactor-control-sonnet-5-medium-n5.md); how much room is left depends on the reasoning effort — at n = 1 the same corpus gap is [−1.8 at high](docs/evidence/2026-09-10-go-refactor-control-sonnet-5-high.md), because the unaided control improves faster than the skilled arm; [the architecture reference and the current-Go rule against 1.15.0 at n = 1: −11.5 against −14.2 lines, golden 4/4 and lint-clean 3/4 in both arms, +3% cost, `ARCHITECTURE.md` read in 4/4 sessions on single-package tasks](docs/evidence/2026-09-13-go-arch-current-go-three-arm-n1-sonnet-5-medium.md); [at n = 5 on `report` and `store` with the three-file reference: −3.0 lines in the skill's favor (p = 0.50), golden and lint-clean 10/10 in both arms, −17% cost](docs/evidence/2026-09-13-go-arch-current-go-n5-report-store-roster-n3-sonnet-5-medium.md); [the load step and the three-router gate against 1.17.0 at n = 1: `go-style-core` in context before the first edit in 3/4 sessions and named by the gate in the fourth, against 0/4; an owner skill 4/4 against 1/4; golden 4/4 and lint-clean 3/4 in both arms; lines −13.2 against −16.0 with the gap in one `store` session; +87% cost, the loads landing one per turn](docs/evidence/2026-09-13-go-refactor-routing-load-n1-sonnet-5-medium.md); [with the no-shell paragraph and the prompt note that names the owners, the same night: $0.206 against $0.228 a session in the same run, one load of the refactor skill per session against six in four, two gate blocks against five, golden 4/4 and lint-clean 3/4 in both arms](docs/evidence/2026-09-13-go-refactor-routing-cost-n1-sonnet-5-medium.md); [the 1.21.0 refactor text against 1.20.1 at n = 3 plus `report` at n = 5: golden 17/17 and lint identical in both arms, one skill load per session in 34/34, cost −10% and +4%, lines −10.8 against −15.5 with the gap in `report`, which at n = 5 reversed to +7.2 against +9.0 (p = 0.75)](docs/evidence/2026-09-18-go-refactor-orient-move-n3-sonnet-5-medium.md) | [🟡 Mixed at medium effort: −8.5 lines per task against no skills, three of four tasks smaller, `feed` −15.6 (p = 0.01, still after correcting for four comparisons) but the skill's Plain Code example shares that task's shape; correctness tied at 17/20, with `gateway` 5/5 against 2/5 (`HEAD` patterns in every skilled session) and `catalog` 3/5 against 5/5 on the repeated-SKU clause (n = 5)](docs/evidence/2026-09-11-go-implement-newcode-plaincode-sonnet-5-medium.md); [at n = 10 on three tasks the `gateway` reading reversed — 4/10 against 10/10 unaided, no skilled session registering `HEAD` — while `feed` held at −12.2 with an example that no longer shares its shape and `catalog` tied at 9/10](docs/evidence/2026-09-11-go-implement-newcode-plaincode-n10-sonnet-5-medium.md); [the workflow retune on the same three tasks at n = 5 has every skilled session report the budget line and write a `HEAD` case, `catalog` 5/5 at 19 lines, `gateway` 4/5 against 5/5 unaided, at 7.2× the cost](docs/evidence/2026-09-11-go-implement-newcode-workflow-sonnet-5-medium.md); [the Delete Pass against 1.13.0 at n = 5: `feed` 41.8 to 38.2 lines with the hand-rolled key loop gone, golden 4/5 against 5/5 on the fixture's recorded `null` miss; `gateway` 4/5 against 1/5 on the `HEAD` clause whose rate has flipped every day, so not read as the edit's effect](docs/evidence/2026-09-12-go-implement-delete-pass-feed-gateway-n5-sonnet-5-medium.md); [the edit hook that runs tests and lint after every edit, with the `go-http` discard bullet and the nil-built empty case, against 1.14.0 at n = 5: golden 7/10 to 9/10, every baseline session lint-clean against 3/5 on `feed`, lines level, +5% cost](docs/evidence/2026-09-12-go-implement-edit-hook-feed-gateway-n5-sonnet-5-medium.md); [the normative current-Go rule against 1.15.0 at n = 1: golden 3/3 against 2/3, lint-clean 3/3 against 2/2, no older idiom in either skilled arm while the unaided arm wrote `sort.Slice` in 2/3, so the rule had nothing left to move on these tasks; `MODERNIZATION.md` now read in 3/3 sessions](docs/evidence/2026-09-13-go-arch-current-go-three-arm-n1-sonnet-5-medium.md); [on the `roster` fixture with a pre-1.21 neighbor, n = 3: the unaided arm copied `sort.Strings` 3/3, both skilled arms 0/6 and nobody rewrote the neighbor, so the pack separates from no skill and the rule does not separate from 1.15.0](docs/evidence/2026-09-13-go-arch-current-go-n5-report-store-roster-n3-sonnet-5-medium.md); [the idiom card named in `go-code`'s workflow step against 1.20.1 at n = 2 on three tasks, twice: the card is read in 0/24 sessions under any wording; a step of its own made the model skip the load steps in 2/6 sessions (six gate blocks against one), folded into step 2 the routing shape is 1.20.1's; golden 11/12 against 12/12, shell-less reports marking every claimed check `(hook)` 8/11 against 1/11, +8% cost within the reference's own movement](docs/evidence/2026-09-18-go-implement-card-step-roster-feed-gateway-n2-sonnet-5-medium.md); [a third run at n = 3 turns the cost sign: pooled over 21 sessions a side +2%, p ≈ 0.86, golden 19/21 against 21/21 on two `gateway` clauses recorded flipping before](docs/evidence/2026-09-18-go-implement-card-in-step2-seed3-roster-feed-gateway-n3-sonnet-5-medium.md); [the card through the host route — the prompt note naming its path, the gate requiring one whole `Read` — against the same tree at n = 3 plus n = 2: read 15/15 against 1/15, 14 from the note and 1 from the gate, golden 14/15 against 15/15 on the `HEAD` clause, +15% cost of which about half is the card's own tokens](docs/evidence/2026-09-18-go-implement-card-hook-route-roster-feed-gateway-n3-sonnet-5-medium.md); [the same route at low effort, n = 2: card read 6/6 against 0/6 from the note alone, +10% cost, the gate met for `go-style-core` in 1/6 against 6/6 sessions, `gateway` 0/4 in both arms](docs/evidence/2026-09-18-go-implement-card-hook-route-roster-feed-gateway-n2-sonnet-5-low.md); [the `HEAD` case named with its scope ("one `HEAD` request per `GET` path, the health check included") and the routing rule made a count, on `gateway` at n = 5 medium plus n = 3 low: `HEAD` in the code 8/8 against 5/8, golden 5/5 against 4/5 at medium and 2/3 against 0/3 at low, cost level at medium](docs/evidence/2026-09-18-go-implement-head-case-gateway-sonnet-5.md); [the empty-list case read as body text against `[]`: 7/8 against 5/8 tests compare the body, the two remaining `null` misses in low sessions that wrote no empty-list case](docs/evidence/2026-09-18-go-implement-null-body-gateway-sonnet-5.md) | [🟡 First reading, n = 2: recall over 29 seeded defects 0.74 unaided to 0.90 with the skill, must-severity defects filed under Must Fix 0.75 to 0.93, four defects (the `rows` lifecycle, a one-implementation interface, a middle-man type, `context.Background` in a test) from 0/2 to 2/2; correct bait lines flagged 0.50 to 0.60; 2.3× the cost of a $0.06 review](docs/evidence/2026-09-12-go-review-corpus-smoke.md); [eleven checklist rows the linters already report cut and seven restated, against 1.20.1 on all six fixtures at n = 2: recall 0.72 → 0.80 over 152 key lines, must 0.82 → 0.87, the linter-reported lines 12/16 → 15/16, bait lines flagged 8 → 5, off-key citations 25 → 15, at level cost](docs/evidence/2026-09-18-go-review-checklist-cut-n2-sonnet-5-medium.md); [the eleven cuts alone against the cuts with the seven restatements at n = 2: recall 0.77 against 0.76, must 0.92 against 0.89, cost equal — the gain is the cuts; the same bytes scored 0.80 and 0.76 on two seeds, the corpus's noise floor](docs/evidence/2026-09-18-go-review-cuts-only-vs-restated-n2-sonnet-5-medium.md) | Refactoring ≈ **4.4×** at medium (n = 5); ≈ 3.3× at the CLI default (n = 5) and ≈ 5.6× at high (n = 1). New code ≈ **5.6×** at medium (n = 5), ≈ 7.2× on the three-task workflow run | **Worth using for refactoring at medium effort, at 4.4× the price; fix the effort level before comparing two runs.** |
| **Haiku 4.5 / Claude** | [🟡 The router is the limit: by description alone `go-code-refactor` reached 1 of 4 sessions at n = 1 and 4 of 12 at n = 3; the prompt hook makes it the first action in 12/12, at 2.56× the unaided cost; wording effects on this tier are unmeasured](docs/evidence/2026-09-11-go-refactor-prompt-routing-haiku-4-5.md) | [✅ On `roster` (n = 5): unaided sessions copied `sort.Strings` from the pre-1.21 neighbor 5/5, one without the import so it did not build; the 1.16.0 text left it in 2/5; the idiom card 0/5, with `slices.Sort` 5/5 and `slices.ContainsFunc` 4/5, golden 5/5 in both skilled arms — one card session also modernized the neighbor it was told to leave](docs/evidence/2026-09-13-go-implement-roster-current-go-card-n5-haiku-4-5.md); [the card named in `go-code` step 2 against 1.20.1 at n = 5: read 4/5 against 3/5, older form 0/5 against 1/5, cost level, and no session that read the card wrote an older form (12/12 over two runs); as a step of its own it cost +18% and two gate blocks](docs/evidence/2026-09-18-go-implement-card-step-roster-n5-haiku-4-5.md) | — Not tested | New code ≈ **4.7×** on `roster` (n = 5); refactoring ≈ 2.6× with the hook (n = 3) | **The idiom card moves Haiku off stale forms; use it with the prompt hook, which is what gets the router loaded at all.** |

Uses the latest available control run for each model + tool + work type by
JSON `finished` (September 7–13, 2026, local time), with one exception: a newer
run with fewer repetitions does not displace one with more. The GPT-5.6-Luna
refactoring and new-code cells are set by the September 13 medium-effort
control at n = 5 per fixture and arm. The Sonnet
refactoring cell is set by the September 10 medium-effort control at n = 5 per
fixture and arm on release 1.7.0; its new-code cell by the September 11
three-arm run at n = 5 on the working tree after 1.7.0, whose `baseline` arm
carries the new-code edits and the Plain Code retune the changelog describes;
the one-repetition runs qualify these and do not set them. Reasoning effort belongs to a
control's identity: the same
model and plugin tree on the same day produced a −9.7 corpus difference at
medium (n = 5) and −1.8 at high (n = 1), so a row read across effort levels is
a row read wrong. A newer subset run does not cover the full corpus. In runs
containing variants, this compares **baseline against no skills**, not the best
variant. Historical results and detailed numbers remain in the linked reports.

**Reasoning effort on Opus 5.** Run the plugin at `--effort medium` on Opus 5:
every Opus 5 cell above was measured there, and Anthropic's Claude Opus 5
migration guidance names `low` and `medium` as the primary cost lever, with
`high` the API default. The [2026-09-12 sweep](docs/evidence/2026-09-12-go-implement-effort-sweep-opus-5.md)
ran the unaided control and the skilled arm at `low` and `high` on three
new-code tasks at n = 5: the `HEAD` correctness effect is the same at every
effort, the size effect exists only at `high`, and the cost multiplier is
worst at `low` (5.7×) because the skill text is a fixed charge against a
cheaper session. `medium` stays the recommended setting, measured for every
cell and between the two on cost. The Sonnet 5 pair above, −9.7 lines at
medium and −1.8 at high on one tree in one day, is why a comparison holds
only within one effort level.

**Loading cost.** On the 2026-09-11 Opus 5 medium traces a routed
implementation session loads about 23K tokens of skill text over three to four
`Skill` turns; written to the cache once and re-read on every later call, that
text is about 37% of the session's cost. Release 1.12.0 loads the owners in one
message after `go-code`, drops three `go-style-core` sections the model already
knows and the verification callouts in three skills, and compacts every Related
Skills section. Measured on `feed`, Sonnet 5 medium, n=5, reference against
baseline: golden 5/5 in both arms, `go-style-core` 2867 → 2335 tokens on load,
cache writes 40.9K → 35.9K, $0.299 → $0.275 a session
([report](docs/evidence/2026-09-11-go-implement-load-cuts-feed-n5-sonnet-5-medium.md)).
On Opus 5 the reference `gateway` session loaded six owners one per turn where
1.12.0 loads them in one message: 16 API calls to 10
([smoke, n=1](docs/evidence/2026-09-11-go-implement-load-cuts-smoke-opus-5-medium.md);
[Sonnet 5 smoke](docs/evidence/2026-09-11-go-implement-load-cuts-smoke-sonnet-5-medium.md)).
The table above is unchanged: these runs are n=1 or have no unaided arm.

The [Sonnet 5 HTTP experiments](docs/evidence/2026-09-08-sonnet-http-compact.md)
also exposed blocked reference reads in the Claude evaluation setup. The
September 9 update fixes that setup and adopts a compact inline HTTP guide
with additional correctness rules. Its combined cost effect is not yet measured.

The GPT-5.6-Luna cells come from the 2026-09-13 medium control at n = 5
([report](docs/evidence/2026-09-13-gpt-5-6-luna-medium-controls.md)):
refactoring 20/20 golden in both arms, −6.8 production lines and −1.75
functions per valid session against no skills; new code 12.0 lines and 0.78
functions smaller among valid sessions, but golden 26/30 against 27/30, the
gap concentrated in `fetch`. That run does not establish a new-code benefit.
An earlier three-arm new-code run on 2026-09-08 separated the plugin from a
wording update (20/20 hidden-test passes in all three arms, +0.70 lines,
p = 0.94) and left a `gateway` helper-count spread of 0 to 7 per session,
unsettled at n = 5. The September 13 control supersedes those cells.

“Helps” means observed reductions in unnecessary code or helpers without
failures in the available checks; readability was not separately assessed by
blind review. Samples are small: 3–10 repetitions per task. “Benefit not
established” does not mean “always harmful.” Cost is the ratio of recorded USD
for the whole run, not output-token savings. In the Sonnet 5 medium refactor
run, `dispatch` and `report` survive a Bonferroni correction for the corpus's
four comparisons; `pricing` remains exploratory after it and `store` is a tie.

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
│       ├── SKILL.md      # Core rules (<= 400 lines each)
│       ├── references/   # Detailed guidance, loaded on demand
│       ├── scripts/      # Automation scripts and helpers
│       └── assets/       # Output templates (5 skills)
├── agents/               # go-verify subagent (Claude Code plugin)
├── hooks/                # gofmt/vet/go fix, go-code routing, prompt and subagent routing hooks (Claude Code plugin)
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
