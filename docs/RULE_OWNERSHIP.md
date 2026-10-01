# Rule Ownership Map

Each rule should have one canonical owner. Other skills should route to that
owner with a short pointer instead of repeating a full explanation.

| Rule area | Canonical owner | Route from | Source basis |
|---|---|---|---|
| Interface placement and shape | `go-interfaces` | `go-code-review`, `go-style-core`, `go-defensive` | Go CodeReviewComments `Interfaces`; Effective Go interface names |
| Compile-time interface assertions | `go-interfaces` | `go-defensive`, `go-style-core` | Uber `Verify Interface Compliance`; Effective Go blank identifier |
| Embedding in public structs; a generic method cannot satisfy an interface (the generic-method form itself is the generics row's) | `go-interfaces` | `go-defensive`, `go-generics` | Uber `Avoid Embedding Types in Public Structs`; Effective Go embedding; Go 1.27 generic methods |
| Generics: when a type parameter earns its place over an interface or a concrete type, constraints, generic type aliases versus definitions, generic methods (Go 1.27), function-type inference | `go-generics` | `go-code`, `go-interfaces`, `go-code-review` | Go specification `Type parameters`; Go blog `When To Use Generics`; Go 1.24 generic type aliases; Go 1.27 generic methods |
| Context parameter placement and values | `go-context` | `go-concurrency`, `go-code-review`, `go-logging` | Go CodeReviewComments `Contexts`; Google documentation conventions |
| Goroutine lifetime and synchronization | `go-concurrency` | `go-context`, `go-code-review` | Go CodeReviewComments `Goroutine Lifetimes`; Uber goroutine guidance |
| Error matching, wrapping, and ownership | `go-error-handling` | `go-code-review`, `go-logging`, `go-defensive` | Uber `Errors`; Go CodeReviewComments `Handle Errors` |
| Log levels and structured logging; the `slog.LogValuer` redaction form and `slog.DiscardHandler` | `go-logging` | `go-error-handling`, `go-code-review`, `go-context` | Google logging best practices; `log/slog` docs |
| Documentation comments and examples | `go-documentation` | all skills that add exported APIs | Google doc comments; Go CodeReviewComments `Doc Comments` |
| Naming, initialisms, receivers, packages, sentinel and error-type names (`ErrX`, `XError`) | `go-naming` | `go-packages`, `go-interfaces`, `go-functions`, `go-error-handling` | Effective Go naming; Go CodeReviewComments naming sections; Google naming decisions; `errname` |
| Pointers to interfaces and value versus pointer parameters | `go-functions` | `go-interfaces`, `go-code-review`, `go-performance` | Uber `Pointers to Interfaces`; Go CodeReviewComments `Pass Values` |
| Receiver type: pointer versus value receivers, one receiver kind per type, receivers of types that hold sync primitives | `go-interfaces` | `go-naming`, `go-concurrency`, `go-functions` | Go CodeReviewComments `Receiver Type`; Effective Go `Pointers vs. Values` |
| Declarations, literals, initialization | `go-style-core` | `go-data-structures`, `go-code-review` | Google declarations decisions; Uber initialization guidance |
| Statement scope, shadowing, loop/range mechanics, switch exits, and blank identifiers | `go-style-core` | `go-code`, `go-data-structures`, `go-naming` | Go specification; Effective Go control statements |
| Nesting depth, early returns, unnecessary else | `go-style-core` | `go-error-handling`, `go-code-refactor` | Uber `Reduce Nesting` / `Unnecessary Else`; Effective Go `if` |
| iota enums and zero-value validity | `go-style-core` | `go-defensive` | Google constant decisions; Uber `Start Enums at One` |
| Data structure selection | `go-data-structures` | `go-performance` | Go CodeReviewComments slices/maps; Google style decisions |
| Nil versus empty collections: nil slices and maps by default, what `slices.Clone`, `maps.Clone`, and `slices.Collect` return for nil, and when a non-nil empty value is required (JSON v1 `[]`, a later map write); what each JSON encoder writes for nil stays with the JSON v2 row | `go-data-structures` | `go-defensive`, `go-style-core`, `go-http` | Go CodeReviewComments `Declaring Empty Slices`; `slices` and `maps` docs |
| Copy depth and aliasing at API boundaries; shallow `slices.Clone`/`maps.Clone` versus type-specific `Clone` contracts | `go-defensive` | `go-data-structures`, `go-code-review`, `go-code`, `go-code-refactor` | Go specification assignments; `slices`, `maps`, and `net/url` docs |
| Function signatures (including wrapping), constructor configuration, functional options vs config structs | `go-functions` | `go-code`, `go-interfaces`, `go-style-core` | Uber functional options; Google option struct guidance |
| Lint setup and static analysis | `go-linting` | `go-code-review`, `go-style-core` | Uber linting; golangci-lint v2 config schema |
| Benchmarks, profiling, hot-path changes | `go-performance` | `go-data-structures` | Uber performance guidance; Go testing benchmark docs |
| Table tests, helpers, integration tests | `go-testing` | `go-code-review`, `go-documentation` | Google testing best practices; Uber test tables |
| Package structure, imports, main/run pattern | `go-packages` | `go-naming` | Go CodeReviewComments package names/imports; Uber exit-in-main guidance |
| Dependency selection and the stdlib-first ladder | `go-packages` | `go-logging`, `go-performance` | `COMPATIBILITY.md`; Go 1.27 standard library |
| Verification gate (`gofmt`/`vet`/`test -race`/`go fix`/lint/`govulncheck`) | `go-linting` | `go-code-review`, `go-style-core`, `go-testing`, `go-error-handling`, `go-code-refactor` | `go tool vet help`; `go tool fix help`; golangci-lint v2 |
| Edit hook record: what the plugin's PostToolUse hook runs after a `.go` edit (`gofmt`, `go vet`, `go fix -diff`, the package tests, `golangci-lint`), its output as the check record where no shell tool exists, `pass (hook)` / `unavailable (no shell)` in the report, and what its silence means | `go-style-core` | `go-code`, `go-code-refactor` | `hooks/go-vet-on-edit.sh`; `evals/hook_test.go` |
| Behavior-preserving refactor workflow and modernization tiers | `go-code-refactor` | `go-style-core`, `go-code-review` | Google readability hierarchy; `go tool fix help`; verified `api/go1.2*.txt` deltas |
| Restraint ladder, reach-for table (replacements the idiom card in `go-style-core/SKILL.md` does not already list), ship-then-question write rules, over-engineering audit, cut tags, and the `Kept:` shortcut ledger | `go-code-refactor` | `go-code`, `go-code-review`, `go-style-core` | Go CodeReviewComments `Interfaces`; Uber `Avoid Embedding Types`; stdlib replacements in `COMPATIBILITY.md` |
| Delete first: line count as the instrument, readability as the goal; delete, then shorten, then restructure | `go-code-refactor` | `go-code`, `go-code-review` | ponytail (DietrichGebert); Google `Least mechanism`, with `Concision` third in its hierarchy — readability stays the goal |
| New code from a specification: the plain-code form of a body (the specification's vocabulary, function-local types and documents, steps inline, errors wrapped), the contract table of observable clauses walked before closing, the reader pass, and the declaration budget counted per package-level declaration added beyond the specification | `go-code` | `go-code-review`, `go-http`, `go-code-refactor` | project policy; `evals/ab/_implement/_golden`; `evals/eval_test.go` |
| Restraint intensity (`lite` / `full` / `ultra`): the level word, its hold for the session, what each level changes on top of the restraint ladder, and what no level touches; `hooks/go-restraint-ladder.sh` prints the ladder and the level row from their owner files at session start and into each subagent | `go-code` | none: a behavior-preserving refactor runs at `full` | ponytail (DietrichGebert) `Intensity` and its SessionStart/SubagentStart injection; `evals/hook_test.go` |
| Internal code comments: a constraint the code cannot show, at the neighbors' density | `go-style-core` | `go-code`, `go-code-review`, `go-documentation` | Google style guide `Comments`; Go CodeReviewComments `Comment Sentences` |
| House style: repository conventions outrank the guide; current-Go idioms at the `go` directive outrank an older neighbor (Write Current Go — conventions stay, idioms do not; the write-time idiom card, a section of the same SKILL.md) | `go-style-core` | `go-code`, `go-code-refactor`, `go-code-review`, `go-packages`, `go-testing`, `go-naming`, `go-logging` | Google `Consistency` principle; Effective Go; project policy 2026-09-13, adopting the stance of JetBrains go-modern-guidelines `use-modern-go`, its 54 guidelines cross-checked at `155dc7c` |
| HTTP handler shape, `ServeMux` routing, server timeouts, shutdown, clients, error-to-status mapping | `go-http` | `go-code`, `go-code-review`, `go-database` | `net/http` docs; Go 1.22 routing enhancements; Go 1.25 `CrossOriginProtection` |
| JSON v2 I/O, wire compatibility, defaults, and exact-byte test conditions | `go-http` | `go-packages`, `go-defensive`, `go-testing`, `go-code`, `go-data-structures` | Go 1.27 `encoding/json/v2` docs and migration guide |
| SQL access: context on queries, rows lifecycle, transactions, placeholders, N+1, pool settings; PostgreSQL constraints, indexes, live migrations | `go-database` | `go-code`, `go-code-review`, `go-http` | `database/sql` docs; go.dev/doc/database; PostgreSQL manual; pg-aiguide topic selection |
| Linters that enforce skill rules (`depguard`, `sloglint`, `errorlint`, `rowserrcheck`, ...) | `go-linting` | every skill whose rule has a linter | golangci-lint v2 linter catalogue |
| Trust-boundary threat model: injection, object-level authorization in the query, SSRF and hostname allowlists, open redirects, uploads served back, secrets, constant-time compare, password hashing, signed-token verification, TLS and cookie settings, redaction, `gosec` findings | `go-security` | `go-code`, `go-code-review`, `go-http`, `go-database`, `go-logging`, `go-defensive` | OWASP Go-SCP; `crypto/*`, `html/template`, `net/netip` docs |
| `os.Root` and `crypto/rand` mechanics (the form, not the threat): the root opened once and held on the server, archive extraction through the root (`root.MkdirAll`, no link entries, size cap), handing a subprocess the open file, `filepath.IsLocal` for names never opened, the `rand.Read` error contract | `go-defensive` | `go-security` | `os`, `path/filepath`, `archive/tar`, `crypto/rand` docs (Go 1.24, 1.25) |
| Panic versus error return: input and environment errors never cross a package boundary as a panic, programming errors (API misuse, `MustX` at init on build-time input, an unreachable case) may; package-internal `recover`; `recover` in goroutines a server starts, since `net/http` recovers only its handler goroutine | `go-defensive` | `go-error-handling`, `go-code-review` | Effective Go `Panic`/`Recover`; Uber `Don't Panic`; `net/http` `Server` docs |
| Checked `Close`: a written file joins `Close` into a named `err` with a deferred `errors.Join`; a file only read uses `os.ReadFile` / `root.ReadFile` | `go-defensive` | `go-error-handling`, `go-code-review` | errcheck; `os.File.Close` docs |
| Replay eligibility, retry/attempt budgets, idempotency across delivery, overload admission, circuit recovery, and fallback contracts | `go-resilience` | `go-code`, `go-http`, `go-context`, `go-concurrency`, `go-database`, `go-troubleshooting` | RFC 9110/6585; `net/http`, `x/time/rate`; provider idempotency contract; Google SRE; Azure circuit breaker guidance |
| Ticket investigation, deployed version/config comparison, data-flow tracing, root-cause method, diagnostic capture (`pprof`, traces, goroutine dumps, `GODEBUG`, `dlv`), symptom-to-mechanism catalog | `go-troubleshooting` | `go-code`, `go-performance`, `go-concurrency` | go.dev/doc/diagnostics; `runtime`, `runtime/pprof`, `runtime/trace` docs |
| Deciding not to refactor or changing its sequence: no change coming, untested critical path, minimal-change request, no stated purpose | `go-code-refactor` | `go-code`, `go-code-review` | Fowler `Refactoring` (2nd ed.); project policy |
| Refactoring catalog: smell, transform, tool, and risk tier per move | `go-code-refactor` | `go-code-review`, `go-packages` | Fowler `Refactoring` (2nd ed.) |
| Safety-net sizing before a refactor: coverage tiers over the blast radius, characterization tests, seams | `go-code-refactor` | `go-linting` | Feathers `Working Effectively with Legacy Code`; `go tool cover` docs |
| Bulk mechanical rewrites: `gofmt -r`, `eg`, `gopatch`, `go/analysis` fixers | `go-code-refactor` | `go-linting`, `go-code` | golang.org/x/tools/cmd/eg; uber-go/gopatch; `go/analysis` docs |
| Cross-package moves during a refactor: type-alias gradual repair, import-cycle strategies, deprecate-before-delete | `go-code-refactor` | `go-packages` | Go 1.9 type alias proposal; Go Modules Reference |
| Architecture of an existing monolith: the smells and their evidence, the import graph, "modules own the tree; layers live inside a module when they earn a package", the preferred layer names (`handlers`, `services`, `repositories`, `models`) and their allowed imports, cross-module contracts, the behavior contracts a move carries (atomicity, error identity, authorization), the target shapes (flat, classic layered, module-first layered, hexagonal inside a module, strong modular monolith), the staged move, the `check-architecture.sh` checker (ownership, layer, composition, platform, contract, driver, unclassified rules; exact `known` list with stale check), the report contract | `go-code-refactor` | `go-packages`, `go-code` | go.dev/doc/modules/layout; Go blog `Package names`; Ben Johnson `Standard Package Layout`; Kat Zien (GopherCon 2018); Three Dots Labs; Fowler `Branch by Abstraction`, `Strangler Fig`; depguard v2; go-arch-lint |
| User scope and house style precedence, narration, report length, host-controlled delegation | `go-style-core` | `go-code`, `go-code-review`, `go-code-refactor`, `go-troubleshooting` | Project policy; [OpenAI GPT-6 Astra prompting guidance](https://developers.openai.com/api/docs/guides/latest-model/gpt-6-astra.md#prompting-best-practices); Anthropic Opus guidance remains host-specific |
| Deprecated API replacement targets, versions, and risk conditions | `go-code-refactor` | `go-packages` | `go-code-refactor/references/MODERNIZATION.md`; `$GOROOT/api/go1.*.txt` |
| Boundary safety pitfalls: typed nil, `append` aliasing, narrowing conversions, float compare, nil channel, division by zero | `go-defensive` | `go-code`, `go-code-review`, `go-troubleshooting` | Go specification; `gosec` G115; Effective Go |
| Production observability checklist (metrics shape, trace correlation, done criteria) | `go-logging` | none | `log/slog` docs; Prometheus histogram guidance |
| Benchmark discipline (file layout, serial runs, benchstat evidence, perf commits) | `go-performance` | none | Go testing benchmark docs; `benchstat` |
| Tool directives and dependency audit (`go get -tool`, tidy check) | `go-packages` | `go-linting` | `go help get`; `go help tool` |
| CI pipeline shape (version matrix, test flags, pinned actions, least-privilege permissions) | `go-linting` | `go-testing` | GitHub Actions docs; `go help testflag` |
| Review procedure: the scope (a diff first, then risk order), owner loads before the first finding, tool output before the checklist, the checklist order, severity labels (Must Fix / Should Fix / Nits), `verified` / `plausible` markers, and the review report template | `go-code-review` | `go-code`, `go-code-refactor`, `go-style-core`, `go-linting`, `go-security`, `go-logging` | Go CodeReviewComments; Google code review guide; project policy |
| Owner selection before the first edit: the routing table, the load before the first edit, the prompt note that names the owners, the gate that holds every `.go` edit until a router (the entry router it names when none is loaded), `go-style-core` (which carries the idiom card), and the owners are loaded — a reminder is not a load, and a stalled retry stops the session instead of a third block — and the project instruction template for clients the hooks do not reach | `go-code` | `go-code-refactor`, `go-code-review` | Project policy; `hooks/go-code-routing.sh`, `hooks/go-prompt-routing.sh`; `docs/PROJECT_INSTRUCTIONS.md`; `evals/hook_test.go` |

## Maintenance Rules

- Add new rule areas here before duplicating guidance in another skill.
- In non-owner skills, keep route text to one or two lines plus a link.
- If sources conflict, record the chosen repository policy in the owner
  reference and link to it from route-only skills.

## Scope Exceptions

A rule area is written for the code its owner skill teaches about. Where this
repository's own tooling does not follow one, record the exception here so a
reader of the workflow finds the reason instead of an apparent oversight.

### `go-code` carries a per-entity form of the restraint ladder

The ladder's owner stays `OVER-ENGINEERING.md`, and `TestRestraintLadder`
still forbids a second copy of its rung text. `go-code` restates the ladder as
the Declaration Budget for new code because the routed file was read in 0 of
20 skilled sessions of the 2026-09-10 implementation control, while the one
large new-code result on record — Opus 5 on `gateway`, 2026-09-07, −34.6%
lines — came from a `go-code` that carried the restraint rule inline.
Both measurements predate the current targets: re-verify on the 5.5 targets
(and on GPT 6.1 and Grok 4.7 once a runner exists) before relying on this
exception or removing it.

### CI pipeline shape does not apply to this repository's workflows

The checklist in `go-linting` addresses a Go **service**. This repository is a
skills pack: its only committed module, `evals/`, has no `require` block, and
its tests read Markdown and shell out to the bundled scripts. Decision on
2026-09-07 — `.github/workflows/validate-skills.yml` stays as it is.

Not applicable, by the pack's own shape:

- **No version matrix.** The pack states one baseline (`COMPATIBILITY.md`), and
  CI pins `go-version: '1.27.x'` to match it. A second entry would test the
  toolchain, not the pack.
- **No `govulncheck`, no `go mod tidy` drift check.** A zero-dependency module
  has no reachable third-party CVE and no tidy drift to catch. Both become
  applicable the first time `evals/go.mod` gains a `require`.

Deliberate divergence, with the residual risk stated rather than argued away:

- **Actions carry `@vN` tags, not full commit SHAs.** The `validate` job reads
  public source only, but the opt-in `evals` job passes
  `secrets.ANTHROPIC_API_KEY`, so a compromised tag in that job is a real
  exposure. Accepted for now against the cost of re-pinning on every bump.
- **No `permissions:` block in `validate-skills.yml`**, so the repository
  default applies. `go-release-watch.yml` sets one at workflow level.
- ~~Tests ran `go test -count=1 ./...` without `-race -shuffle=on`~~ — closed
  on 2026-09-10; the workflow now runs `go test -count=1 -race -shuffle=on ./...`
  because `eval_test.go` has dozens of `t.Parallel()` subtests that write
  files and exec scripts.

`evals/evals.json` quality eval 45 grades a model on flagging exactly this
shape in a *service* pipeline. That eval is correct as written; it does not
describe this repository.
