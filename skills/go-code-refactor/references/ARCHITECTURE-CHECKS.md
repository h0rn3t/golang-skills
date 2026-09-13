# Enforcing the Architecture Boundary

> Sources: [`go help list`](https://go.dev/cmd/go/#hdr-List_packages_or_modules); [golangci-lint configuration](https://golangci-lint.run/docs/configuration/file/) and [depguard settings](https://golangci-lint.run/docs/linters/configuration/#depguard); [fe3dback/go-arch-lint](https://github.com/fe3dback/go-arch-lint)
> Authority: project policy — the rules, the exception format, and the evaluation scenarios; the tool mechanics are as documented upstream
> Minimum Go: 1.27 baseline; the checker needs only `go list` and the standard library
> Last verified: 2026-09-13

Companion to [ARCHITECTURE.md](ARCHITECTURE.md): how to install and read the
bundled checker, what each rule fails, what the checker cannot see, the
optional `depguard` duplicate, the report shape, and the scenarios the skill
is evaluated against. The checker is
[`scripts/check-architecture.go`](../scripts/check-architecture.go), run
through [`scripts/check-architecture.sh`](../scripts/check-architecture.sh);
its unit tests are
[`scripts/check-architecture_test.go`](../scripts/check-architecture_test.go)
and the [fixture](../testdata/architecture/README.md) is the module it passes.

## Contents

- [Running the checker](#running-the-checker)
- [Measuring the graph by hand](#measuring-the-graph-by-hand)
- [architecture.json](#architecturejson)
- [Rules](#rules)
- [What the checker does not cover](#what-the-checker-does-not-cover)
- [Optional depguard duplicate](#optional-depguard-duplicate)
- [Report contract and skill evaluation](#report-contract-and-skill-evaluation)
- [Common mistakes](#common-mistakes)

## Running the checker

```bash
bash "$REFACTOR_SKILL_DIR/scripts/check-architecture.sh" [--json] [--limit N] [--include-tests] [--config FILE] [module-root]
```

The wrapper builds the checker once into the skills cache and runs it from the
module root (default `.`). Exit 0 is a clean graph, 1 is at least one
violation or a stale, duplicate, or unexplained `known` entry, 2 is an error:
no `go.mod`, a package that does not load, a missing or invalid
`architecture.json`, an unknown flag, or `--limit` without a non-negative
integer. `--json` prints one object:

```json
{"module":"example.com/shop","layout":"modules","checked":["Imports"],"violations":[{"rule":"ownership","from":"internal/billing/services","to":"internal/order/repositories","message":"..."}],"total":1,"suppressed":0,"truncated":false}
```

`status: "no_internal_packages"` marks a module with nothing under
`internal/`; that is a successful empty check, exit 0. Package paths in output
are module-relative, so the same `known` list survives a module rename. The
checker runs `go list` with `GOWORK=off` and `-mod=readonly`: one module, one
build configuration, production imports only unless `--include-tests` adds
`TestImports` and `XTestImports`. Record the build tags, `GOOS`, and `GOARCH`
you ran under; a second configuration is a second run.

## Measuring the graph by hand

Run from the selected Go module root with the production build
configuration; the commands describe one module and one build configuration.
Failure to load packages must not produce a misleading partial graph.

```bash
# Bash; failure to load packages must not produce a misleading partial graph.
set -euo pipefail
export GOWORK=off # Deliberately select one Go module, not an implicit workspace.
module=$(go list -mod=readonly -m -f '{{.Path}}')
go version
go env GOWORK GOOS GOARCH CGO_ENABLED GOFLAGS

go list -mod=readonly -f '{{$p := .ImportPath}}{{range .Imports}}{{$p}} {{.}}{{"\n"}}{{end}}' ./... \
  | awk 'NF == 2' > edges.txt

# Internal fan-in: exact module-path prefix, not a regular-expression approximation.
awk -v prefix="$module/" 'index($2, prefix) == 1 { print $2 }' edges.txt \
  | sort | uniq -c | sort -rn | sed -n '1,10p'
awk '{ print $1 }' edges.txt | sort | uniq -c | sort -rn | sed -n '1,10p'

# A discovery aid only; the executable checker is the enforcement mechanism.
awk '$1 ~ /\/(services|models)(\/|$)/ && \
     $2 ~ /^(net\/http|database\/sql|github.com\/jackc\/pgx|gorm[.]io|github.com\/gofiber\/fiber)(\/|$)/' edges.txt
```

For automation, use `go list -json`, the selected module's actual `Path`/`Dir`, and an exact prefix such as `modulePath + "/internal/"`. Never split an arbitrary import path at its first `/internal/`: that segment can also occur in the Go module path itself. Handle package-load errors as failures. `.Imports`, `.TestImports`, and `.XTestImports` are distinct inputs; record which were checked. [Go command documentation](https://go.dev/cmd/go/#hdr-List_packages_or_modules).

## architecture.json

The file lives in the module root; `--config` points elsewhere. Unknown fields
are rejected so a typo cannot silently disable a rule.

| Field | Meaning | Default |
|---|---|---|
| `layout` | `"layers"` — shape B, one implicit module whose layers sit directly under `internal/`; `"modules"` — shape C/E, one directory per business module | required |
| `internal` | directory holding the checked packages | `"internal"` |
| `app` | the composition root under `internal/` | `"app"` |
| `platform_dir` | the shared-infrastructure directory under `internal/` | `"platform"` |
| `platform` | approved shared packages business code may import, module-relative and exact, e.g. `"internal/platform/clock"` | none approved |
| `modules` | layout `modules` only: `{"order": "layered", "search": "flat"}`; every directory under `internal/` that is not `app` or platform must appear | required for `modules` |
| `drivers.transport`, `drivers.database` | import-path prefixes counted as transport and database drivers | `net/http`, gin, echo, chi, fiber, grpc; `database/sql`, pgx, gorm, pq, sqlx, mongo-driver, ent |
| `known` | reviewed exceptions: `{"rule", "from", "to", "reason", "owner", "until"}` with module-relative paths; `until` is optional, the rest required | `[]` |

A layered module has a root contract package (`internal/order`) and layer
packages (`internal/order/handlers`, `…/services`, `…/repositories`,
`…/models`, nested subpackages included). A flat module is one package tree
whose root is its public implementation; it may contain I/O and is consumed
by another module's services through its root only.

## Rules

| Rule | Fails when | Fix |
|---|---|---|
| `ownership` | a business package imports another module's layer package, or a foreign root contract from anywhere but `services` or a flat module | consume the provider's root contract from the use case, or declare a consumer-side interface and wire it in `app` |
| `layer` | inside one module, a layer imports a layer the table in ARCHITECTURE.md forbids — handlers → repositories, services → handlers or repositories, repositories → services, models → anything but models — nested subpackages inherit their layer | move the call behind the layer that owns it; a service declares the store interface it needs |
| `composition` | business or platform code imports `app` | `app` imports them, never the reverse |
| `platform` | platform imports a business module, or business code imports a platform package that is not on the `platform` approved list | approve the exact package in `architecture.json` after review, or move the code |
| `contract` | a layered module's root imports one of its own implementation packages | keep the root to types and errors; implementation stays in the layers |
| `driver` | `services`, `models`, or a layered root imports a transport or database driver; `handlers` imports a database driver | handlers own transport, repositories own persistence |
| `unclassified` | a package under `internal/` is neither `app`, platform, a configured module, nor a layer; or business code imports a module-local package outside `internal/` | add the module or layer to the configuration, or move the package |
| `stale`, `duplicate`, `unexplained` | a `known` entry matches no current violation, appears twice, or lacks a reason or owner | remove or complete the entry; the list only shrinks |

An entry in `known` is a review decision with an owner and a removal
condition. **A refactor never adds one to make its own run pass**, never
broadens a driver list or the platform list, and never switches `layout`; each
of those is a separate architecture decision the report names.

## What the checker does not cover

Test-only imports (unless `--include-tests`), packages outside `internal/`,
other Go modules in a workspace, build configurations other than the one run,
transitive library behavior, SQL table ownership, runtime call cycles hidden
behind interfaces, atomicity, error identity, and authorization. Those last
three are behavior: [ARCHITECTURE-BEHAVIOR.md](ARCHITECTURE-BEHAVIOR.md) names
the tests each needs, and an import graph never proves them. Exported
signatures of a root contract that re-export an implementation type are a
review item; the checker sees imports, not signatures.

## Optional depguard duplicate

`depguard` can restate the `driver` rule per file inside `golangci-lint`; it
duplicates that one rule and none of the others. The configuration is schema
version 2 and enables the linter explicitly; both the global (`internal/services`)
and module-first (`internal/<module>/services`) spellings are covered, nested
packages included. Validate it with the repository's pinned `golangci-lint`
(`golangci-lint config verify --config <file>`) before adopting it.

```yaml
version: "2"
linters:
  enable:
    - depguard
  settings:
    depguard:
      rules:
        services-and-models-have-no-drivers:
          list-mode: lax
          files:
            - "**/internal/services/*.go"
            - "**/internal/services/**/*.go"
            - "**/internal/*/services/*.go"
            - "**/internal/*/services/**/*.go"
            - "**/internal/models/*.go"
            - "**/internal/models/**/*.go"
            - "**/internal/*/models/*.go"
            - "**/internal/*/models/**/*.go"
            - "!$test"
          deny:
            - pkg: net/http
              desc: transport belongs in handlers (ARCHITECTURE.md, layer rules)
            - pkg: github.com/gofiber/fiber
              desc: transport belongs in handlers
            - pkg: database/sql
              desc: persistence belongs in repositories
            - pkg: github.com/jackc/pgx
              desc: persistence belongs in repositories
            - pkg: gorm.io
              desc: persistence belongs in repositories
        handlers-have-no-database-drivers:
          list-mode: lax
          files:
            - "**/internal/handlers/*.go"
            - "**/internal/handlers/**/*.go"
            - "**/internal/*/handlers/*.go"
            - "**/internal/*/handlers/**/*.go"
            - "!$test"
          deny:
            - pkg: database/sql
              desc: persistence belongs in repositories
            - pkg: github.com/jackc/pgx
              desc: persistence belongs in repositories
            - pkg: gorm.io
              desc: persistence belongs in repositories
```

`go-arch-lint` can state the module and layer edges declaratively as
components with `mayDependOn` once the target tree exists; it is a separate
binary and a second gate step, and it has no exception ledger with owners.

## Report contract and skill evaluation

Use this compact report shape; cite repository paths/lines or measured edges where available:

```text
Observation: the concrete ownership/dependency issue and evidence.
Consequence: the work or behavior it currently makes unsafe or expensive.
Decision: keep the tree / repair a seam / propose a structural move.
Why not smaller: the simpler option considered and why it is insufficient.
Preserve: behavior, contracts, authorization, and transaction boundaries.
Changes: minimal package/API movement; code growth and what it buys.
Verification: commands, tool/build scope, results, and before/after findings.
Exceptions/limits: approved waivers, unexecuted checks, remaining uncertainty.
```

Do not turn "fewer imports" or "more modules" into success criteria. Evaluate the skill with both positive and negative cases:

| Fixture/scenario | Required behavior |
|---|---|
| Coherent one-domain B | No unsolicited directory migration |
| One handler-to-repository shortcut | Local boundary repair, not a new architecture |
| Unrelated domains mixed in global layers | Ownership-based proposal retaining preferred names |
| Same-module forbidden edge | Checker fails, including a nested subpackage |
| Foreign implementation vs foreign root contract | Reject the implementation; allow a legitimate contract from a use-case consumer |
| Business-to-app or platform-to-business import | Checker fails |
| Module path itself contains `/internal/` | Correct classification using the actual module prefix |
| Removed, duplicated, or unexplained exception | Checker fails rather than silently accepting it |
| Package move across an atomic use case | Atomicity and observable behavior preserved |
| Tool unavailable or build scope incomplete | Report the limitation without inventing a green gate |

Skill-level scenarios are acceptance criteria, not claims that the attached checker proves the model will choose the correct architecture. Compare behavior and justified changes, not one golden directory tree. Load examples and executable details only when they are needed; see [Agent Skills authoring practices](https://agentskills.io/skill-creation/best-practices).

## Common mistakes

| Mistake | Fix |
|---|---|
| Blaming `handlers/services/repositories/models` | Judge ownership, scope, and edges |
| Treating a four-layer feature edit as proof of bad architecture | Identify unrelated owners/contracts that had to change |
| Choosing C/D/E before considering no move | Try the smallest boundary repair first |
| "Concrete first" while forbidding service-to-repository imports | Use a minimal consumer interface at this real architectural seam |
| Drawing repository-to-service imports as mandatory | Let structural satisfaction and constructor wiring do the work |
| Putting multi-module business orchestration in `app` | Give the use case a business owner; reserve `app` for composition |
| Exposing ORM/local models through a root contract | Publish the small intended contract, not a re-export of the implementation |
| Creating DTOs and generic repositories for every layer | Split only at demonstrated representation or behavioral divergence |
| Losing atomicity when splitting repositories | Preserve the transaction contract and test partial failure |
| Assuming `%w` hides driver details | Deliberately choose error identity and abstraction boundaries |
| Making `platform/**` universally importable | Exact shared-package approval; no business dependencies from platform |
| Checking only cross-module edges | Check layers, composition, drivers, and unknown classification too |
| Waiving the violation introduced by the refactor | Fix it or report it; seek a separate policy decision |
| Claiming imports prove data ownership or authorization | Add query, contract, and behavior checks |
| Renaming an entire monolith in one patch | Stage one ownership concept; retain reversible checkpoints |
