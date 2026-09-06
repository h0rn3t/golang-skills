# Rule Ownership Map

Each rule should have one canonical owner. Other skills should route to that
owner with a short pointer instead of repeating a full explanation.

| Rule area | Canonical owner | Route from | Source basis |
|---|---|---|---|
| Interface placement and shape | `go-interfaces` | `go-code-review`, `go-style-core`, `go-defensive` | Go CodeReviewComments `Interfaces`; Effective Go interface names |
| Compile-time interface assertions | `go-interfaces` | `go-defensive`, `go-style-core` | Uber `Verify Interface Compliance`; Effective Go blank identifier |
| Context parameter placement and values | `go-context` | `go-concurrency`, `go-code-review`, `go-logging`, `go-testing` | Go CodeReviewComments `Contexts`; Google documentation conventions |
| Goroutine lifetime and synchronization | `go-concurrency` | `go-context`, `go-code-review`, `go-testing` | Go CodeReviewComments `Goroutine Lifetimes`; Uber goroutine guidance |
| Error matching, wrapping, and ownership | `go-error-handling` | `go-code-review`, `go-logging`, `go-defensive` | Uber `Errors`; Go CodeReviewComments `Handle Errors` |
| Log levels and structured logging | `go-logging` | `go-error-handling`, `go-code-review`, `go-context` | Google logging best practices; `log/slog` docs |
| Documentation comments and examples | `go-documentation` | all skills that add exported APIs | Google doc comments; Go CodeReviewComments `Doc Comments` |
| Naming, initialisms, receivers, packages | `go-naming` | `go-packages`, `go-interfaces`, `go-functions` | Effective Go naming; Go CodeReviewComments naming sections; Google naming decisions |
| Pointers to interfaces | `go-functions` | `go-interfaces`, `go-code-review` | Uber `Pointers to Interfaces`; Go CodeReviewComments `Pass Values` |
| Declarations, literals, initialization | `go-style-core` | `go-data-structures`, `go-code-review` | Google declarations decisions; Uber initialization guidance |
| Statement scope, shadowing, loop/range mechanics, switch exits, and blank identifiers | `go-style-core` | `go-code`, `go-data-structures`, `go-naming` | Go specification; Effective Go control statements |
| Nesting depth, early returns, unnecessary else | `go-style-core` | `go-error-handling`, `go-code-refactor` | Uber `Reduce Nesting` / `Unnecessary Else`; Effective Go `if` |
| iota enums and zero-value validity | `go-style-core` | `go-defensive` | Google constant decisions; Uber `Start Enums at One` |
| Data structure selection | `go-data-structures` | `go-generics`, `go-performance` | Go CodeReviewComments slices/maps; Google style decisions |
| Function signatures, constructor configuration, functional options vs config structs | `go-functions` | `go-code`, `go-interfaces` | Uber functional options; Google option struct guidance |
| Lint setup and static analysis | `go-linting` | `go-code-review`, `go-style-core` | Uber linting; golangci-lint v2 config schema |
| Benchmarks, profiling, hot-path changes | `go-performance` | `go-data-structures`, `go-functions` | Uber performance guidance; Go testing benchmark docs |
| Table tests, helpers, integration tests | `go-testing` | `go-code-review`, `go-documentation` | Google testing best practices; Uber test tables |
| Package structure, imports, main/run pattern | `go-packages` | `go-code-review`, `go-naming` | Go CodeReviewComments package names/imports; Uber exit-in-main guidance |
| Dependency selection and the stdlib-first ladder | `go-packages` | `go-logging`, `go-error-handling`, `go-performance` | `COMPATIBILITY.md`; Go 1.27 standard library |
| Verification gate (`gofmt`/`vet`/`test -race`/`go fix`/lint) | `go-linting` | `go-code-review`, `go-style-core`, `go-testing`, `go-concurrency`, `go-error-handling`, `go-code-refactor` | `go tool vet help`; `go tool fix help`; golangci-lint v2 |
| Behavior-preserving refactor workflow and modernization tiers | `go-code-refactor` | `go-style-core`, `go-code-review` | Google readability hierarchy; `go tool fix help`; verified `api/go1.2*.txt` deltas |
| Restraint ladder, reach-for table, ship-then-question write rules, over-engineering audit, cut tags, and the `Kept:` shortcut ledger | `go-code-refactor` | `go-code`, `go-code-review`, `go-style-core` | Go CodeReviewComments `Interfaces`; Uber `Avoid Embedding Types`; stdlib replacements in `COMPATIBILITY.md` |
| Delete first: line count as the instrument, readability as the goal; delete, then shorten, then restructure | `go-code-refactor` | `go-code`, `go-code-review` | ponytail (DietrichGebert); Google `Least mechanism`, with `Concision` third in its hierarchy — readability stays the goal |
| House style: repository conventions outrank the guide | `go-style-core` | `go-code`, `go-code-refactor`, `go-testing`, `go-naming`, `go-http`, `go-database` | Google `Consistency` principle; Effective Go |
| HTTP handler shape, `ServeMux` routing, server timeouts, shutdown, clients, error-to-status mapping | `go-http` | `go-code`, `go-code-review`, `go-database` | `net/http` docs; Go 1.22 routing enhancements; Go 1.25 `CrossOriginProtection` |
| SQL access: context on queries, rows lifecycle, transactions, placeholders, N+1, pool settings | `go-database` | `go-code`, `go-code-review`, `go-http` | `database/sql` docs; go.dev/doc/database |
| Linters that enforce skill rules (`depguard`, `sloglint`, `errorlint`, `rowserrcheck`, ...) | `go-linting` | every skill whose rule has a linter | golangci-lint v2 linter catalogue |
| Trust-boundary threat model: injection, SSRF, secrets, constant-time compare, password hashing, TLS and cookie settings, redaction, `gosec` findings | `go-security` | `go-code`, `go-code-review`, `go-http`, `go-database`, `go-logging`, `go-defensive` | OWASP Go-SCP; `crypto/*`, `html/template`, `net/netip` docs |
| `os.Root` and `crypto/rand` mechanics (the form, not the threat) | `go-defensive` | `go-security` | `os`, `crypto/rand` docs (Go 1.24) |
| Replay eligibility, retry/attempt budgets, idempotency across delivery, overload admission, circuit recovery, and fallback contracts | `go-resilience` | `go-code`, `go-http`, `go-context`, `go-concurrency`, `go-database`, `go-troubleshooting` | RFC 9110/6585; `net/http`, `x/time/rate`; provider idempotency contract; Google SRE; Azure circuit breaker guidance |
| Ticket investigation, deployed version/config comparison, data-flow tracing, root-cause method, diagnostic capture (`pprof`, traces, goroutine dumps, `GODEBUG`, `dlv`), symptom-to-mechanism catalog | `go-troubleshooting` | `go-code`, `go-performance`, `go-concurrency`, `go-testing` | go.dev/doc/diagnostics; `runtime`, `runtime/pprof`, `runtime/trace` docs |
| Semantic rename and extract via gopls during a refactor | `go-code-refactor` | `go-naming`, `go-code` | golang.org/x/tools/gopls docs |
| Deciding not to refactor or changing its sequence: no change coming, untested critical path, minimal-change request, no stated purpose | `go-code-refactor` | `go-code`, `go-code-review` | Fowler `Refactoring` (2nd ed.); project policy |
| Refactoring catalog: smell, transform, tool, and risk tier per move | `go-code-refactor` | `go-code-review`, `go-packages` | Fowler `Refactoring` (2nd ed.); gopls code actions |
| Safety-net sizing before a refactor: coverage tiers over the blast radius, characterization tests, seams | `go-code-refactor` | `go-testing`, `go-linting` | Feathers `Working Effectively with Legacy Code`; `go tool cover` docs |
| Bulk mechanical rewrites: `gofmt -r`, `eg`, `gopatch`, `go/analysis` fixers | `go-code-refactor` | `go-linting`, `go-code` | golang.org/x/tools/cmd/eg; uber-go/gopatch; `go/analysis` docs |
| Cross-package moves during a refactor: type-alias gradual repair, import-cycle strategies, deprecate-before-delete | `go-code-refactor` | `go-packages`, `go-interfaces` | Go 1.9 type alias proposal; Go Modules Reference |
| User scope and house style precedence, narration, report length, host-controlled delegation | `go-style-core` | `go-code`, `go-code-review`, `go-code-refactor`, `go-troubleshooting` | Project policy; [OpenAI GPT-6 Astra prompting guidance](https://developers.openai.com/api/docs/guides/latest-model/gpt-6-astra.md#prompting-best-practices); Anthropic Opus guidance remains host-specific |

## Maintenance Rules

- Add new rule areas here before duplicating guidance in another skill.
- In non-owner skills, keep route text to one or two lines plus a link.
- If sources conflict, record the chosen repository policy in the owner
  reference and link to it from route-only skills.
