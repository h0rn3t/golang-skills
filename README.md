# Agent Skills For Go

**English** | [Українська](README.uk.md)

AI [Agent Skills](https://agentskills.io/) for writing idiomatic,
production-quality **Go 1.27** code. 26 modular skills teach AI coding
assistants Go best practices derived from:

- [Google Go Style Guide](https://google.github.io/styleguide/go/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Go Wiki CodeReviewComments](https://github.com/golang/go/wiki/CodeReviewComments)

Skills are tuned following
[agentskills.io best practices](https://agentskills.io/skill-creation/best-practices):
content the agent already knows is omitted, procedural decision trees guide
multi-step tasks, 58 reference files load on demand via progressive disclosure,
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
| **go-control-flow** | Idiomatic conditionals, loops, switch/break behavior, guard clauses |
| **go-data-structures** | Slices, maps, arrays — allocation with new vs make, append, copying |
| **go-database** | database/sql and ORMs — contexts on queries, rows lifecycle, transactions, N+1, pool settings |
| **go-declarations** | Variable/const/type declarations, var vs :=, iota enums, shadowing |
| **go-defensive** | API boundary hardening, defer cleanup, Must functions, time handling |
| **go-documentation** | Doc comments, package docs, godoc formatting, runnable examples |
| **go-error-handling** | Error strategy decisions, wrapping (%v vs %w), sentinels, logging patterns |
| **go-functional-options** | Functional options pattern for constructors with optional config |
| **go-functions** | Function ordering, signature formatting, Printf verbs, Stringer interface |
| **go-generics** | When to use generics, constraints, common pitfalls, type aliases |
| **go-http** | net/http handlers, ServeMux routing, middleware, server timeouts, shutdown, clients |
| **go-interfaces** | Interface design, abstractions, embedding, "accept interfaces return structs" |
| **go-linting** | Linters, golangci-lint setup, nolint directives, CI/CD integration |
| **go-logging** | Structured logging with slog, log levels, request-scoped context, migration |
| **go-naming** | Naming decision flow for packages, types, functions, variables, receivers |
| **go-packages** | Package organization, imports, package size, CLI/flag patterns |
| **go-performance** | String optimization, capacity hints, benchmarking, strconv over fmt |
| **go-security** | Trust-boundary threat model — injection, SSRF, secrets, constant-time compare, password hashing, TLS, cookies, redaction |
| **go-style-core** | Formatting, nesting reduction, style principles, fallback style guide |
| **go-testing** | Table-driven tests, subtests, test helpers, assertions, test organization |
| **go-troubleshooting** | Root-cause method for panics, hangs, leaks, flaky tests — pprof, traces, goroutine dumps, symptom catalog |

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
# all 26 skills
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
3. **Progressive disclosure**: Core rules load immediately; 58 reference files
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

`evals/evals.json` holds 78 trigger evals (does the right skill fire for this
prompt?) and 27 quality evals (does the answer satisfy each assertion?). The Go
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
| `new(expr)` for non-composite pointers | go-declarations |
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
│       ├── SKILL.md      # Core rules (< 225 lines each)
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
