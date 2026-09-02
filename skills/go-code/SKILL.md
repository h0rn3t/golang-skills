---
name: go-code
description: Use when writing, fixing, or refactoring Go code without a single obvious topic, and whenever this skill is named as a modifier on another workflow (for example "/opsx:apply /go-code" or "/commit /go-code") — it routes the task to the go-* skills it actually needs, then closes with the verification gate. Passed as an argument to another command it is NOT a change name, a file path, or a topic — it means "run that workflow under the Go rules below". Does not carry rules of its own; every rule lives in the skill it routes to.
---

# Go Code Profile

Router. Owns no rules — it picks which `go-*` skills the task needs and
enforces the gate at the end.

## Resource Routing

This skill bundles no files. Everything it loads belongs to another skill:

- `../go-code-refactor/references/OVER-ENGINEERING.md` - Read on every invocation for the restraint ladder and cut tags.
- `../go-style-core/SKILL.md` - Read on every invocation for house style and the fallback rules.

## Invocation

- **Standalone**: `/go-code <task>` — write or fix the described Go code.
- **As a modifier**: `/opsx:apply /go-code`, `/opsx:apply add-auth /go-code`.
  The host workflow keeps control of its own steps, arguments, and state; this
  skill only adds the routing and the gate. Strip `/go-code` from the host
  command's arguments before parsing them — it is never a change name or a
  file path.

## Always Loaded

Two things load on every invocation, before any routing decision:

1. **The restraint rules.** Read
   [go-code-refactor/references/OVER-ENGINEERING.md](../go-code-refactor/references/OVER-ENGINEERING.md)
   — it owns the restraint ladder, the cut tags (`delete:`, `stdlib:`, `dep:`,
   `yagni:`, `shrink:`), and the Go hunt list; the module rungs belong to
   [go-packages](../go-packages/SKILL.md). Climb the ladder before writing a
   new type, layer, interface, option, or import, and stop at the first rung
   that holds: it applies to code being written, not only to code being
   audited.

   Climb it *after* reading the code the change touches and tracing the real
   flow, never instead. The ladder shortens the solution, not the reading — a
   small diff in the wrong place is a second bug.

   The ladder runs per entity in the diff, not once per task. "The feature is
   needed" does not carry over to the helpers, types, and methods written to
   implement it — each gets its own climb. Check the final diff, not the
   opening plan.
2. **The fallback style owner**, [go-style-core](../go-style-core/SKILL.md),
   which covers anything the routing table below does not — and owns the
   house-style rule: the repository's `.golangci.yml`, `CONTRIBUTING.md`, and
   neighboring code outrank every rule these skills carry. Read the neighbors
   before the first edit.

A deliberate shortcut with a known ceiling gets a `Kept:` marker naming the
ceiling and the upgrade path, so `check-debt.sh` can harvest it later.

Restraint never cuts input validation at trust boundaries, error handling that
prevents data loss, security controls, accessibility basics in anything
user-facing, or anything the user asked for explicitly. Those are the code that
has to exist.

> **Note**: Where the `ponytail` skill is installed, load it here too — it is
> the general-purpose source these Go-specific rules were derived from.

## Route Before The First Edit

Then load only the rows the task actually touches. The third column names the
skill a row almost always drags in — load it in the same pass, not after the
first draft exposes the gap:

| Task touches | Skill | Also load |
|---|---|---|
| errors, wrapping, `errors.Is`/`errors.AsType` | [go-error-handling](../go-error-handling/SKILL.md) | — |
| goroutines, channels, mutexes, races | [go-concurrency](../go-concurrency/SKILL.md) | [go-context](../go-context/SKILL.md) if anything is cancelled |
| `context.Context`, timeouts, cancellation | [go-context](../go-context/SKILL.md) | [go-concurrency](../go-concurrency/SKILL.md) if goroutines are started |
| tests, table-driven cases, `synctest` | [go-testing](../go-testing/SKILL.md) | — |
| new identifiers, new exported API | [go-naming](../go-naming/SKILL.md) | [go-documentation](../go-documentation/SKILL.md) |
| interfaces, embedding, test doubles | [go-interfaces](../go-interfaces/SKILL.md) | — |
| constructor with 3+ optional parameters | [go-functional-options](../go-functional-options/SKILL.md) | — |
| slices, maps, arrays, sets | [go-data-structures](../go-data-structures/SKILL.md) | — |
| `var`/`const` blocks, `iota`, composite literals | [go-declarations](../go-declarations/SKILL.md) | — |
| `if`/`for`/`switch` mechanics, statement scoping | [go-control-flow](../go-control-flow/SKILL.md) | — |
| type parameters, constraints, generic methods | [go-generics](../go-generics/SKILL.md) | — |
| `slog`, log levels, request-scoped fields | [go-logging](../go-logging/SKILL.md) | [go-security](../go-security/SKILL.md) if a secret or PII could reach a log line |
| `defer` cleanup, boundary copies, mutable globals | [go-defensive](../go-defensive/SKILL.md) | — |
| hot paths, allocations, benchmarks | [go-performance](../go-performance/SKILL.md) | [go-troubleshooting](../go-troubleshooting/SKILL.md) if the cause of slowness is unknown |
| package layout, imports, dependencies | [go-packages](../go-packages/SKILL.md) | — |
| function ordering, signatures, `Printf` verbs | [go-functions](../go-functions/SKILL.md) | — |
| restructuring or deleting existing code | [go-code-refactor](../go-code-refactor/SKILL.md) | — |
| linter config, CI checks | [go-linting](../go-linting/SKILL.md) | — |
| HTTP handlers, routing, middleware, servers, clients | [go-http](../go-http/SKILL.md) | [go-error-handling](../go-error-handling/SKILL.md); [go-security](../go-security/SKILL.md) if input reaches a file, shell, URL, or template |
| SQL queries, transactions, repositories, migrations | [go-database](../go-database/SKILL.md) | [go-error-handling](../go-error-handling/SKILL.md); [go-security](../go-security/SKILL.md) if identifiers come from input |
| untrusted input, secrets, tokens, TLS, cookies | [go-security](../go-security/SKILL.md) | [go-defensive](../go-defensive/SKILL.md) |
| panic, hang, leak, flaky test, cause unknown | [go-troubleshooting](../go-troubleshooting/SKILL.md) | the owner of the mechanism once found |
| JSON and other wire formats, struct tags | [go-defensive](../go-defensive/SKILL.md) (tags) | [go-packages](../go-packages/SKILL.md) (`json/v2` on the ladder) |
| CLI entry point, flags, `main`/`run` | [go-packages](../go-packages/SKILL.md) | — |
| a list of findings from a review or audit | the rows the findings name | per area, not per task |

A task that is a package of fixes has no single row. Group the findings by
area, take the row for each area, and work one area at a time — the "also
load" ceiling counts per area, not per task.

Load what the task needs and nothing else. Loading every row is noise, not
thoroughness — the rules that do not apply crowd out the ones that do; the
"also load" column is the ceiling, not a suggestion to keep adding.

If exactly one row matches and its third column is empty, invoke that skill
directly and skip this one; a router in front of a single destination is
overhead.

## Close With The Gate

Run the verification gate from [go-linting](../go-linting/SKILL.md) —
`go vet`, `go fix -diff`, `golangci-lint run`, `go test -race`, and
`govulncheck` before release. Fix what it reports; a task is not done because
the code compiles.

Where the repository defines its own gate (`CLAUDE.md`, `CONTRIBUTING.md`, CI
config), that gate wins and this list is only the default in its absence —
go-linting owns both rules, including how to report a tool that is not
installed and how far `go fix -diff` reaches outside the diff.

Topic-specific checks stack on top of the gate, not instead of it:

- Refactor → `bash skills/go-code-refactor/scripts/verify-refactor.sh` to prove
  behavior held, and `check-debt.sh` for the `Kept:` markers left behind.
- Errors → `bash skills/go-error-handling/scripts/check-errors.sh`.
- New exported API → `bash skills/go-documentation/scripts/check-docs.sh`.
- Before submitting → [go-code-review](../go-code-review/SKILL.md).

Under a host workflow with per-task checkpoints (such as an OpenSpec apply
loop), run the gate before marking each Go-touching task complete, not once at
the very end — a batched gate reports failures too late to attribute them.

> **Note**: This plugin ships a `go-verify` subagent that runs the gate and
> returns only the failures — delegate to it rather than pasting the full
> output into the main thread. Delegate audits to any `go-db` or
> `go-concurrency` agents the environment provides. The bundled PostToolUse
> hook already runs `gofmt` and `go vet` on every edited `.go` file; its
> report is the first gate step, not a substitute for the rest.

---

## Related Skills

- **Style fallback**: See [go-style-core](../go-style-core/SKILL.md) — it owns formatting, nesting, and every rule with no more specific owner
- **Review**: See [go-code-review](../go-code-review/SKILL.md) for the pre-submission checklist; it reviews finished code, this skill routes code being written
- **Refactoring**: See [go-code-refactor](../go-code-refactor/SKILL.md) when the task is restructuring existing code rather than adding to it
- **Gate definition**: See [go-linting](../go-linting/SKILL.md) — it owns the verification commands this skill defers to
- **Dependency ladder**: See [go-packages](../go-packages/SKILL.md) — it owns the normative stdlib-before-module rule this skill enforces on every write
