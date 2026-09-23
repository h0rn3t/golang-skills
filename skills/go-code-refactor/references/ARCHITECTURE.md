# Architecture of an Existing Monolith

> Sources: [go.dev/doc/modules/layout](https://go.dev/doc/modules/layout); [`go help list`](https://go.dev/cmd/go/#hdr-List_packages_or_modules); [Executing transactions](https://go.dev/doc/database/execute-transactions); [Working with Errors in Go](https://go.dev/blog/go1.13-errors); [go.dev/blog/package-names](https://go.dev/blog/package-names); [Google Go interface guidance](https://google.github.io/styleguide/go/decisions.html#interfaces); Fowler, *Branch by Abstraction*; the background works listed under Sources and companion materials
> Authority: target shapes are advisory. The dependency rules below are project policy, not requirements of the Go language. For a multi-domain monolith, the preferred ownership model is **modules own the tree; layers live inside a module when they earn a package**.
> Minimum Go: 1.27 project baseline, retained from the original reference. Do not change a repository's `go.mod`, toolchain, or pinned linters as a side effect of an architecture refactor.
> Last verified: 2026-09-13. This date records the document revision, not successful execution against the target repository or verification of every external example.

The compiler restricts access to `internal/` packages and rejects import cycles. It does not enforce service boundaries, data ownership, or transaction semantics. Those need executable checks and behavior tests.

Use this reference when the problem concerns package dependencies or ownership, not merely a long function. Read repository conventions and measure the existing graph before choosing a target. **Keeping the current structure and repairing one boundary is a valid, often preferable result.**

An architecture move is High tier ([SKILL.md](../SKILL.md#risk-tiers)): propose it unless the user explicitly requested the move. Permission to refactor a function or improve readability is not permission to reorganize the application. Production code may grow, but the report must explain what concrete coupling, ownership violation, or behavior duplication that growth removes ([Concision Gate](../SKILL.md#concision-gate)). [go-packages](../../go-packages/SKILL.md#package-organization) owns placement of new packages.

Here, **business module** means an owned feature/domain package tree, not a Go module with a separate `go.mod`. Do not introduce multiple `go.mod` files merely to obtain business boundaries.

## Contents

- [When the smell is architectural](#when-the-smell-is-architectural)
- [Measure the import graph](#measure-the-import-graph)
- [Modules own the tree; layers live inside the module](#modules-own-the-tree-layers-live-inside-the-module)
- [The preferred layer names](#the-preferred-layer-names)
- [Cross-module contracts](#cross-module-contracts)
- [Transactions, errors, and authorization](#transactions-errors-and-authorization)
- [Real-world examples](#real-world-examples)
- [Target shapes](#target-shapes)
- [Choosing a target](#choosing-a-target)
- [Staging the move](#staging-the-move)
- [Enforcing the boundary](#enforcing-the-boundary)
- [Report contract](#report-contract)
- [Common mistakes](#common-mistakes)
- [Decision rule](#decision-rule)
- [Sources and companion materials](#sources-and-companion-materials)

## When the smell is architectural

| Smell | What it looks like in Go | Evidence to collect |
|---|---|---|
| God package | A global `models`, `types`, `common`, `util`, or `pkg` mixes unrelated owners | High fan-in/fan-out **and** no coherent ownership sentence; identify the unrelated consumers |
| Global layers outgrowing one domain | Global `handlers/services/repositories/models` mix independently changing domains | Feature changes repeatedly edit unrelated domain logic, force foreign contract changes, or require unrelated owners to coordinate |
| Layer leak | Services/models import transport or persistence drivers; handlers call repositories directly | Exact importing package, imported package, and the use that crosses the boundary |
| Cycle by proxy | Business types are moved into a generic third package to avoid a cycle | The third package has weak ownership and exists to preserve a bidirectional business dependency |
| Wiring by globals | Shared DB handles or `init()` registrations select implementations | Packages that read/mutate the global; [go-defensive](../../go-defensive/SKILL.md) owns global-state repair |
| Cross-feature data access | Billing reads or mutates orders-owned tables without an ownership decision | Actual queries, intended consistency, table owner, and affected use cases |
| Accidental transaction fragmentation | Moving methods into repositories changes one atomic action into independent writes | Before/after commit boundaries, rollback tests, partial-failure behavior |

The names `handlers`, `services`, `repositories`, and `models` are **not** the smell. Nor is a feature touching four layers: a new persisted field can legitimately require transport, validation, business, and storage changes. Count edits **outside the feature's owner**, not just edited directories.

A smell is a reason to investigate, not an automatic migration trigger. Propose a structural move only when a concrete consequence is established and a smaller repair is insufficient. Do not use a fixed number of smells, packages, files, or imports as a migration threshold.

## Measure the import graph

Run from the selected Go module root, with the same build configuration as the relevant production build. Record Go/tool versions, module path, `GOWORK`, `GOOS`, `GOARCH`, `CGO_ENABLED`, `GOFLAGS`, and explicit build tags. The commands below describe **one module and one build configuration**, not all workspace modules or all tagged files.

The commands are in
[ARCHITECTURE-CHECKS.md](ARCHITECTURE-CHECKS.md#measuring-the-graph-by-hand);
the checker automates the classification they only approximate.

Inspect package ownership and real call/data dependencies alongside the import graph. High fan-in can be appropriate for one well-owned value package. Replacing an import with an interface can leave the same business coupling in place. Do not optimize import count as a score. See [PACKAGE-SIZE.md](../../go-packages/references/PACKAGE-SIZE.md).

The before/after report compares package count, significant fan-in/fan-out, forbidden edges, foreign implementation access, and ownership/transaction violations. A package-splitting tool may help explore an acyclic partition; it does not decide the business owner or authorize a refactor. Verify the installed tool's capabilities before invoking it ([CATALOG.md](CATALOG.md#split-or-merge-a-package)).

## Modules own the tree; layers live inside the module

For several independently changing domains, prefer ownership at the first meaningful directory boundary. For one cohesive service, global layers remain valid.

```text
# B: one cohesive service          # C: several owned domains
internal/                         internal/
    handlers/                         order/
    services/                             handlers/
    repositories/                         services/
    models/                               repositories/
                                          models/
                                      billing/
                                          handlers/
                                          services/
                                          repositories/
                                          models/
```

These are alternatives under different constraints, not successive maturity levels. A module does not need all four packages. A small module can stay in one package with `model.go`, `service.go`, `repository.go`, and `handler.go`. Split when dependency isolation, size, or independent testing earns the package.

### Runtime flow is not import direction

```text
Runtime:
request -> handlers -> services -> repository implementation -> database

Typical compile-time dependencies:
handlers     -> services -> models
repositories ----------> models
app          -> handlers, services, repositories
cmd/server   -> app
```

A service declares the minimal behavior it consumes; a concrete repository satisfies that interface structurally. There is no required `repositories -> services` import. Constructor wiring in `app` is already a compile-time check of compatibility. A complete compilable example is in [the companion fixture](../testdata/architecture/README.md), rather than a partial method body in this reference.

**When an interface is justified:** if separate `services` and `repositories` packages must obey the inversion rule, that architectural seam justifies a consumer-side interface even with one implementation. Do not wait for an artificial second implementation. Elsewhere, use concrete types while the dependency is permitted and no useful seam exists. Do not generate an interface for every struct or mirror an entire repository API. See [Google Go interface guidance](https://google.github.io/styleguide/go/decisions.html#interfaces).

## The preferred layer names

Within an explicitly layered module, these are the default rules. Same-layer subpackages may depend on each other when ownership stays coherent and the compiler permits it. An exception must be narrow and recorded; it is not an implied allowance for the whole layer.

| Package | Owns | Allowed project dependencies | Forbidden by default |
|---|---|---|---|
| `models/` | Module-owned business data, invariants, domain errors | Its own model subpackages; separately approved value packages | Services, handlers, repositories, foreign business packages, `app`, transport/DB drivers |
| `services/` | Use cases, business orchestration, authorization decisions, transaction intent, consumer-side interfaces | Own services/models/root contract; another module's deliberate root API; approved platform packages | Handlers, repository implementations, `app`, transport/DB drivers |
| `repositories/` | Persistence, row mapping, outbound data access when repository semantics fit | Own repositories/models/root contract; approved platform packages; storage/client libraries | Handlers; service packages unless an exact justified exception is recorded; foreign business implementations |
| `handlers/` | Decode/encode, structural input checks, identity extraction, response mapping | Own handlers/services/models/root contract; approved platform packages; transport framework | Repositories, direct DB access, `app`; bypassing own use cases to orchestrate foreign modules |
| Layered module root, e.g. `order/` | Optional deliberately small cross-module contract | Standard library and justified pure external value libraries | Its implementation subpackages, foreign implementations, `app`, infrastructure/transport drivers; re-exported ORM/internal types |
| `internal/app/` | Dependency construction, registrations, startup/shutdown, wiring | Classified application implementations and infrastructure required for assembly | Business use cases, authorization policy, transaction decisions |
| `internal/platform/` | Explicitly approved shared technical capabilities | Technical libraries and approved platform packages | Business-module packages, business entities, `app` |
| `cmd/server/` | Process entry point | `app` and minimal process setup | Business rules and use-case coordination |

The root-contract row applies to a **layered** module. A **flat** module's root is its public implementation package and may contain I/O; do not pretend it is a driver-free contract. Its cross-module surface must still be deliberate. The checker distinguishes these cases explicitly.

`app` is not a home for "everything involving two modules." Creating services and passing dependencies is composition. Creating an order, charging a payment, and issuing an invoice is a business process owned by a service. Give a genuinely independent process its own business module only when ownership earns it. Business and platform packages must not import `app`; the executable entry point may.

### `models/` is allowed, but keep it local

Prefer `internal/order/models` and `internal/billing/models` once there are several owners. A global `internal/models` is acceptable in a genuinely single-domain service; its name alone does not imply missing ownership. Go's [package-naming guidance](https://go.dev/blog/package-names) is a useful warning about generic scope, not a reason to override a coherent house style.

One struct with `json`, `db`, or `gorm` tags is acceptable while transport, business, and persistence representations genuinely agree. Tags do not themselves create a driver import. Split at a real divergence: exposed/security-sensitive fields, nullability, validation, persistence representation, or a versioned wire contract. A GORM helper **type**, unlike a tag, introduces an actual dependency and needs separate justification.

Keep transport-only request/response types beside handlers and DB-only row types beside repositories. Do not require three DTOs and mapping functions for every CRUD operation. Do not make local models the automatic cross-module or public HTTP API.

## Cross-module contracts

Business modules must not import another module's handlers, services, repositories, or local models. A foreign service implementation is still an implementation, even if it exposes exported methods.

Prefer a small consumer-side interface in the calling module's services, wired by `app`. Use the provider's root contract when shared named types and a stable API are genuinely needed. Use asynchronous messages only when asynchronous ownership and delivery semantics are intentional.

**Define the data boundary, not just the import path.** A billing use case may need an order summary, not an entire persistence entity. Specify the minimal fields, lookup semantics, missing-object result, mutability expectations, and errors that the consumer can rely on. Do not leak framework contexts, ORM sessions, internal maps/slices that expose mutable ownership, or database rows merely to avoid a small mapping.

Go method signatures must agree on named types. A provider returning its own `OrderSummary` does not automatically satisfy a consumer interface returning an unrelated `billing.OrderSummary`. Choose deliberately: use a provider-owned root-contract type, or add a small adapter at the integration/wiring boundary to map types. That adapter may import both sides but contains no business policy. Do not invent a global `shared/types` package to hide the mismatch.

Import acyclicity does not establish semantic independence. Check synchronous call loops and ownership cycles even when interfaces hide them from the compiler. Reporting/read-model joins, foreign keys, and cross-module transactions are not universally forbidden; record the owner, purpose, consistency, and migration consequences of an intentional exception.

## Transactions, errors, and authorization

A package move must not change what commits together, which errors a caller
can match, or who decides that an actor may perform an operation.
[ARCHITECTURE-BEHAVIOR.md](ARCHITECTURE-BEHAVIOR.md) carries those three
contracts and the behavior checks each needs; read it before moving any code
that a use case's atomicity, error identity, or authorization passes through.

## Real-world examples

Community examples and their limitations are in [ARCHITECTURE-EXAMPLES.md](ARCHITECTURE-EXAMPLES.md). Read it when examples are needed to explain a choice, not before every refactor. They demonstrate naming and dependency techniques; they are not production-quality certifications or templates to copy wholesale. Inspect a specific revision before relying on a repository's current structure.

## Target shapes

| Shape | Tree / rule | Fits | Cost |
|---|---|---|---|
| **A. Flat package** | One package; introduce private subpackages when needed | One small domain, simple dependency graph | Lowest ceremony; ownership may eventually become too broad |
| **B. Classic layered service** | `internal/{handlers,services,repositories,models}`; `app` wires | One cohesive backend that benefits from explicit layers | Changes cross layer directories; unrelated domains can turn global layers into dumping grounds |
| **C. Module-first layered** | `internal/<module>/{handlers,services,repositories,models}`; local contracts | Several independently changing domains in one deploy | More packages and constructors; explicit module interfaces |
| **D. Hexagonal inside one module** | C plus justified local ports/adapters or command/query separation | Complex rules, difficult I/O-free tests, several adapters | More abstractions and mappings; unnecessary for simple CRUD |
| **E. Strong modular monolith** | C plus deliberate data ownership and stronger module contracts | Measurable coordination/data-ownership pressure or a planned extraction | More decisions around reads, writes, transactions, delivery, and migration |

A–E are **not** a mandatory progression. D and E add different constraints; they are not "better C" in every situation.

### A — start simple

Follow [Go's module-layout guidance](https://go.dev/doc/modules/layout): add packages when the code earns them. `internal/` controls import visibility. `cmd/` is a useful convention for entry points, not a language requirement or a privilege reserved for the second binary.

### B — familiar and valid

```text
cmd/server/main.go
internal/app/
internal/handlers/
internal/services/
internal/repositories/
internal/models/
```

Keep B while the domain and ownership remain coherent. An optional `app` package is useful when wiring warrants separation; it is not required ceremony for a small `main`.

### C — the default proposal when several domains justify a move

```text
cmd/server/main.go
internal/app/                     # constructors and lifecycle, not business processes
internal/order/
    api.go                        # optional stable root contract only
    handlers/
    services/
    repositories/
    models/
internal/billing/
    api.go
    handlers/
    services/
    repositories/
    models/
internal/platform/
    clock/
    ids/
    log/
    telemetry/
```

Create only the packages the module needs. `order/repositories` belongs to order persistence; it is not a home for all SQL. `platform` is an explicit infrastructure boundary, not a renamed `common`.

### D — hexagonal only where it pays

Use local dependency inversion and explicit adapters for a demonstrated need. Keep the preferred layer vocabulary where it remains clear. Separate transport/domain/persistence representations only when they differ. Multiple transports alone do not require a large framework or multiple mapping layers.

### E — stronger ownership, not the only modular-monolith definition

Modules own their persistence and expose deliberate APIs. Reads normally use the owning module's API or an agreed read model. Writes may be synchronous or asynchronous according to consistency requirements. Foreign keys and cross-module transactions require an explicit decision, not an automatic prohibition. Do not choose E merely because a module "might become a microservice."

## Choosing a target

Start with [House Style Wins](../../go-style-core/SKILL.md#house-style-wins), then consider the smallest effective change.

| Established evidence | First proposal |
|---|---|
| No ownership or dependency problem | Keep the tree; no architecture refactor |
| One forbidden edge, otherwise coherent ownership | Repair that seam locally |
| One small domain | A, or keep an already coherent B |
| Several independently changing domains mixed in global layers | C, only after explaining why a local repair is insufficient |
| One owned module has a specific adapter/testing problem | D for that module only |
| Existing module boundaries plus concrete ownership/data/extraction pressure | E with an explicit consistency and migration decision |

Not default proposals: global business `shared/common/pkg` packages, microservices to compensate for unpoliced package boundaries, DI containers when constructors suffice, generic repositories without a repeated need, or four packages and three DTOs for every CRUD operation.

## Staging the move

Every stage must remain buildable, testable, and reviewable. Run the existing refactor gates ([verify-refactor.sh](../scripts/verify-refactor.sh), `after`/`diff`) plus the boundary check from the first step. Passing tests is necessary, not proof of safe deployment when schemas or wire contracts change.

1. **State and encode the target rule before moving code.** Capture reviewed existing violations with exact edges, rule IDs, reason, owner, and removal condition. New and stale exceptions must fail the gate.
2. **Characterize observable behavior.** Cover the module surface, response/error contract, authorization, and affected atomicity ([SAFETY-NET.md](SAFETY-NET.md#the-three-tiers)). Do not freeze incidental internal function shapes.
3. **Introduce only the needed seam.** Declare the consumer interface, adapt existing code, migrate callers, switch, and delete the old path. Follow [Branch by Abstraction](https://martinfowler.com/bliki/BranchByAbstraction.html) and [STRUCTURAL.md](STRUCTURAL.md#breaking-an-import-cycle).
4. **Move one ownership concept at a time.** Use temporary type aliases/forwarders only when they preserve behavior and avoid cycles ([moving a type](STRUCTURAL.md#moving-a-type-alias-for-gradual-repair)). Track their removal; do not turn an alias into a permanent leaked contract.
5. **Retire the global dumping ground gradually.** For example, move order-owned data to `order/models`, migrate its callers, then remove the old alias. Do not move unrelated code in the same patch.
6. **Change data ownership/schema separately where possible.** First route calls through the owner; then make any explicit persistence migration. [go-database](../../go-database/SKILL.md) owns migration mechanics.

State what can be reverted at each step. Avoid combining a package relocation with unrelated query optimization, API changes, framework upgrades, or authorization changes. Code growth is justified by a real boundary or safety net, not by completing a diagram.

## Enforcing the boundary

Use [ARCHITECTURE-CHECKS.md](ARCHITECTURE-CHECKS.md) and the bundled checker/tests when installing or changing enforcement. Keep executable code out of this policy reference so there is one implementation to test, not an uncompiled Markdown copy.

| Requirement | Executable or review gate |
|---|---|
| No foreign implementation imports | `scripts/check-architecture.go`: `ownership` rule |
| Same-module layer directions | Same checker: `layer` rule; includes nested layer packages and B/C layouts |
| Business/platform code cannot import `app` | Same checker: `composition` rule |
| Platform cannot import business; shared imports are explicitly approved | Same checker: `platform` rule |
| Layered root contracts do not re-export local implementation dependencies | Same checker: `contract` imports, plus exported-signature review |
| No listed transport/DB drivers in services/models/contracts; no DB drivers in handlers | Same checker: `driver` rule; optional `depguard` duplicates file-level checks |
| Unknown internal ownership is not silently ignored | Same checker rejects unclassified packages and local targets |
| No newly waived violation or stale waiver | Exact rule/edge allowlist and stale-entry check; policy/config changes reviewed separately |
| Atomicity, errors, authorization, data ownership | Behavior/integration tests and query/contract review; **not** proven by import checking |

The checker covers production imports under the selected Go module's `internal/`, for one build configuration. It is not a workspace-wide architecture proof. Test-only imports, packages outside `internal/`, other Go modules, transitive library behavior, SQL table ownership, and runtime call cycles require their own declared checks. Run each supported build scope, and do not share a conditional exception baseline across unrelated scopes without aggregating observations.

The optional standalone `depguard` configuration declares golangci-lint schema version 2 and explicitly enables the linter. Its globs cover both `internal/services` and `internal/<module>/services`, including nested packages, and the equivalent models/handlers cases. It complements the checker; it is not a replacement for ownership/layer enforcement. Validate it with the repository's pinned golangci-lint version before adoption. See [configuration](https://golangci-lint.run/docs/configuration/file/) and [depguard settings](https://golangci-lint.run/docs/linters/configuration/#depguard).

**The model must not add a new allowlist entry, broaden a glob/allow rule, suppress a failure, change the selected layout, or disable a gate merely to make its refactor pass.** A new exception is a separate architecture decision. The supplied configurations start with empty `known` lists; approval metadata cannot be inferred from a failing test. Do not mark a check as passed when it was not run.

## Report contract

The compact report shape (observation, consequence, decision, why not smaller,
preserve, changes, verification, exceptions) and the required behavior per
situation live in [ARCHITECTURE-CHECKS.md](ARCHITECTURE-CHECKS.md#report-contract).
Fewer imports or more modules are never the success criterion.

## Common mistakes

The mistakes the checker and a review catch — blaming the layer names,
choosing C/D/E before considering no move, waiving the violation a refactor
introduced, losing atomicity when splitting repositories — are tabled with
their fixes in [ARCHITECTURE-CHECKS.md](ARCHITECTURE-CHECKS.md#common-mistakes).

## Decision rule

Use the names that make the codebase easy for its maintainers to navigate. Keep `handlers/`, `services/`, `repositories/`, and `models/` as the default layered vocabulary.

```text
small cohesive service:        layers may be the tree
multi-domain monolith:         modules own the tree; layers stay local
no demonstrated boundary bug: keep the current structure
```

Improve an established ownership or dependency problem with the smallest safe change, and prove the relevant behavior was preserved. Directory names are navigation; architectural quality comes from ownership, dependency direction, and explicit contracts.

## Sources and companion materials

Background sources: Go module layout and package naming; Google Go Style Guide; Ben Johnson, *Standard Package Layout* (2016); Kat Zien's GopherCon 2018 talk and JetBrains follow-up; Three Dots Labs on modular monoliths and Clean Architecture; Fowler's *Branch by Abstraction* and *Strangler Fig Application*; Feathers, *Working Effectively with Legacy Code*; Russ Cox's comment in `golang-standards/project-layout#117`; depguard and go-arch-lint. These works are background, not authorities for this project's exact folder spellings or policy preference order.

The links beside technical statements identify the directly relevant official/primary documentation. The checker rules and the report format are project policy, not attributed quotations from those sources.

Companions: [examples](ARCHITECTURE-EXAMPLES.md), [check installation and coverage](ARCHITECTURE-CHECKS.md), [checker](../scripts/check-architecture.go), [checker tests](../scripts/check-architecture_test.go), and [compilable fixture](../testdata/architecture/README.md).
