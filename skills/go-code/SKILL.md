---
name: go-code
description: Go coding profile — routes a task to the go-* skills it actually needs, then closes with the verification gate. Use when writing, fixing, or refactoring Go code without a single obvious topic, and whenever this skill is named as a modifier on another workflow (for example "/opsx:apply /go-code" or "/commit /go-code"). Passed as an argument to another command it is NOT a change name, a file path, or a topic — it means "run that workflow under the Go rules below". Does not carry rules of its own; every rule lives in the skill it routes to.
---

# Go Code Profile

Router. Owns no rules — it picks which `go-*` skills the task needs and
enforces the gate at the end.

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
2. **The fallback style owner**, [go-style-core](../go-style-core/SKILL.md),
   which covers anything the routing table below does not.

A deliberate shortcut with a known ceiling gets a `Kept:` marker naming the
ceiling and the upgrade path, so `check-debt.sh` can harvest it later.

Restraint never cuts input validation at trust boundaries, error handling that
prevents data loss, security controls, accessibility basics in anything
user-facing, or anything the user asked for explicitly. Those are the code that
has to exist.

> **Note**: Where the `ponytail` skill is installed, load it here too — it is
> the general-purpose source these Go-specific rules were derived from.

## Route Before The First Edit

Then load only the rows the task actually touches:

| Task touches | Skill |
|---|---|
| errors, wrapping, `errors.Is`/`errors.AsType` | [go-error-handling](../go-error-handling/SKILL.md) |
| goroutines, channels, mutexes, races | [go-concurrency](../go-concurrency/SKILL.md) |
| `context.Context`, timeouts, cancellation | [go-context](../go-context/SKILL.md) |
| tests, table-driven cases, `synctest` | [go-testing](../go-testing/SKILL.md) |
| new identifiers, new exported API | [go-naming](../go-naming/SKILL.md) + [go-documentation](../go-documentation/SKILL.md) |
| interfaces, embedding, test doubles | [go-interfaces](../go-interfaces/SKILL.md) |
| constructor with 3+ optional parameters | [go-functional-options](../go-functional-options/SKILL.md) |
| slices, maps, arrays, sets | [go-data-structures](../go-data-structures/SKILL.md) |
| `var`/`const` blocks, `iota`, composite literals | [go-declarations](../go-declarations/SKILL.md) |
| `if`/`for`/`switch` mechanics, statement scoping | [go-control-flow](../go-control-flow/SKILL.md) |
| type parameters, constraints, generic methods | [go-generics](../go-generics/SKILL.md) |
| `slog`, log levels, request-scoped fields | [go-logging](../go-logging/SKILL.md) |
| `defer` cleanup, boundary copies, mutable globals | [go-defensive](../go-defensive/SKILL.md) |
| hot paths, allocations, benchmarks | [go-performance](../go-performance/SKILL.md) |
| package layout, imports, dependencies | [go-packages](../go-packages/SKILL.md) |
| function ordering, signatures, `Printf` verbs | [go-functions](../go-functions/SKILL.md) |
| restructuring or deleting existing code | [go-code-refactor](../go-code-refactor/SKILL.md) |
| linter config, CI checks | [go-linting](../go-linting/SKILL.md) |

Load what the task needs and nothing else. Loading every row is noise, not
thoroughness — the rules that do not apply crowd out the ones that do.

If exactly one row matches, invoke that skill directly and skip this one; a
router in front of a single destination is overhead.

## Close With The Gate

Run the verification gate from [go-linting](../go-linting/SKILL.md) —
`go vet`, `go fix -diff`, `golangci-lint run`, `go test -race`, and
`govulncheck` before release. Fix what it reports; a task is not done because
the code compiles.

Topic-specific checks stack on top of the gate, not instead of it:

- Refactor → `bash skills/go-code-refactor/scripts/verify-refactor.sh` to prove
  behavior held, and `check-debt.sh` for the `Kept:` markers left behind.
- Errors → `bash skills/go-error-handling/scripts/check-errors.sh`.
- New exported API → `bash skills/go-documentation/scripts/check-docs.sh`.
- Before submitting → [go-code-review](../go-code-review/SKILL.md).

Under a host workflow with per-task checkpoints (such as an OpenSpec apply
loop), run the gate before marking each Go-touching task complete, not once at
the very end — a batched gate reports failures too late to attribute them.

> **Note**: If the environment provides verification subagents (`go-verify`,
> `go-db`, `go-concurrency`), delegate the gate to them; they run the same
> commands and return only the failures.

---

## Related Skills

- **Style fallback**: See [go-style-core](../go-style-core/SKILL.md) — it owns formatting, nesting, and every rule with no more specific owner
- **Review**: See [go-code-review](../go-code-review/SKILL.md) for the pre-submission checklist; it reviews finished code, this skill routes code being written
- **Refactoring**: See [go-code-refactor](../go-code-refactor/SKILL.md) when the task is restructuring existing code rather than adding to it
- **Gate definition**: See [go-linting](../go-linting/SKILL.md) — it owns the verification commands this skill defers to
- **Dependency ladder**: See [go-packages](../go-packages/SKILL.md) — it owns the normative stdlib-before-module rule this skill enforces on every write
