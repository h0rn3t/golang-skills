---
name: go-code
description: Use when writing, fixing, or refactoring Go code, whether or not the task has one obvious topic — it loads the go-* skills the task needs and runs the closing gate. Use it too when /go-code modifies another workflow (e.g. /opsx:apply /go-code); as a modifier it selects Go rules, not a change name or path.
---

# Go Code Profile

Route Go work to the relevant owners, then close with their verification gate.

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`).

## Resource Routing

Resolve sibling skills and references relative to this installed directory
(Claude Code prints it as the skill's base directory); run their scripts from
the target project. A skill counts as loaded only when its `SKILL.md` content
is in this context: in Claude Code call the `Skill` tool with the skill name;
in Codex or any host without a skill tool, read `../<name>/SKILL.md` next to
this file. Missing resources: report the gap, use available guidance, and
continue independent authorized work without inventing rules.

- `../go-style-core/SKILL.md` — Read once per task for house style, fallback
  rules, and communication guidance. Read its references only for a decision
  the task requires.
- `../go-linting/SKILL.md` — Read its Verification Gate at step 5 on every
  task that edits Go.
- `../go-code-refactor/references/OVER-ENGINEERING.md` — Read the detailed
  restraint ladder when deciding whether an added abstraction is justified;
  its replacement catalog when seeking a simpler existing API; its audit lane
  only when the requested deliverable is a complexity audit. Ordinary edits
  do not require this file.

## Workflow

1. **Resolve invocation.** `$go-code <task>` or `/go-code <task>` selects Go
   work. As a modifier, e.g. `/opsx:apply add-auth /go-code`, remove the modifier
   before the host parses its arguments. It is never a change name or path;
   the host retains workflow state, checkpoints, and delegation policy.
2. **Load `go-style-core`, then read the code.** Load `go-style-core` on every
   task. Then inspect repository instructions, `go.mod`, neighboring code,
   tests, and callers. Identify required inputs, outputs, environment
   constraints, failure behavior, and dependencies; no fixed written contract
   report is required.
3. **Load the owners before the first edit.** Match the task against
   [Route Before The First Edit](#route-before-the-first-edit) and load each
   matched owner plus every `Also load` entry whose condition holds. Select by
   the decisions and behavior being changed, not every syntax element present:
   a routine local variable or `if` does not trigger naming, documentation, or
   extra style references. The step ends when every selected skill's content is
   in context; do not make the first edit before that. Add owners when new
   evidence requires them; there is no numerical cap, and content already
   loaded this session is reused, not reread. Under the Claude Code plugin a
   PreToolUse hook blocks the first Go edit that precedes these loads and names
   the missing skills once; it is a reminder, not a substitute for this step.
   The hook infers owners from decision-bearing syntax only (a `_test.go`
   path, `%w` wrapping, goroutines, `context.With*`, SQL, `slog.`, exec and
   templates, `defer`, type parameters, `interface {`, `make`/`append`,
   `package main`, rate limiting, HTTP server and client calls); naming,
   documentation, functions, performance, refactoring, linting, and
   troubleshooting it cannot see, and its "Also load" conditions are this
   table's. Its silence is not a passing gate result.
4. **Implement the authorized scope.** For new functions, packages, or stub
   bodies, use [Writing New Code](#writing-new-code). For behavior-preserving
   restructuring, use the [delete-first priority](../go-code-refactor/SKILL.md#delete-before-you-restructure).
   Climb the restraint ladder in [OVER-ENGINEERING.md](../go-code-refactor/references/OVER-ENGINEERING.md#the-restraint-ladder)
   for each proposed helper, type, layer, option, or import; stop at the first
   sufficient rung. Preserve required behavior, validation, security controls,
   and meaningful tests.
   Resolve routine choices without stopping; ask only for missing information
   that changes correctness, scope, or authorization.
5. **Verify and report.** Follow the closing gate below and `go-style-core`'s
   communication guidance. Report outcomes, observed checks, and material gaps.

## Route Before The First Edit

The third column adds owners only under its stated condition. Prefer the
narrower owner when topics overlap; ordinary identifier creation alone does
not require `go-naming` or `go-documentation`.

| Task touches | Skill | Also load |
|---|---|---|
| errors, wrapping, `errors.Is`/`errors.AsType` | [go-error-handling](../go-error-handling/SKILL.md) | — |
| goroutines, channels, mutexes, races | [go-concurrency](../go-concurrency/SKILL.md) | [go-context](../go-context/SKILL.md) if anything is cancelled |
| `context.Context`, timeouts, cancellation | [go-context](../go-context/SKILL.md) | [go-concurrency](../go-concurrency/SKILL.md) if goroutines are started |
| tests, table-driven cases, `synctest` | [go-testing](../go-testing/SKILL.md) | — |
| API naming changes or a naming question | [go-naming](../go-naming/SKILL.md) | — |
| new or changed exported API, doc comments | [go-documentation](../go-documentation/SKILL.md) | [go-naming](../go-naming/SKILL.md) if API names change |
| interfaces, embedding, test doubles | [go-interfaces](../go-interfaces/SKILL.md) | — |
| function API design, constructor configuration, ordering, signatures, `Printf` helpers | [go-functions](../go-functions/SKILL.md) | — |
| slices, maps, arrays, sets | [go-data-structures](../go-data-structures/SKILL.md) | — |
| declaration, enum, or initialization decisions | [go-style-core references](../go-style-core/SKILL.md#resource-routing) | — |
| loop/switch mechanics or statement scoping decisions | [go-style-core references](../go-style-core/SKILL.md#resource-routing) | — |
| type parameters, constraints, generic methods | [go-generics](../go-generics/SKILL.md) | — |
| `slog`, log levels, request-scoped fields, metrics and trace correlation | [go-logging](../go-logging/SKILL.md) | [go-security](../go-security/SKILL.md) if a secret or PII could reach a log line |
| `defer` cleanup, boundary copies, mutable globals, nil/aliasing/overflow traps | [go-defensive](../go-defensive/SKILL.md) | — |
| hot paths, allocations, benchmarks | [go-performance](../go-performance/SKILL.md) | [go-troubleshooting](../go-troubleshooting/SKILL.md) if the cause of slowness is unknown |
| package layout, imports, dependencies | [go-packages](../go-packages/SKILL.md) | — |
| restructuring or deleting existing code | [go-code-refactor](../go-code-refactor/SKILL.md) | — |
| linter config, CI checks | [go-linting](../go-linting/SKILL.md) | — |
| HTTP handlers, routing, middleware, servers, clients | [go-http](../go-http/SKILL.md) | [go-error-handling](../go-error-handling/SKILL.md); [go-security](../go-security/SKILL.md) if input reaches a file, shell, URL, or template |
| SQL queries, transactions, repositories, migrations | [go-database](../go-database/SKILL.md) | [go-error-handling](../go-error-handling/SKILL.md); [go-security](../go-security/SKILL.md) if identifiers come from input |
| untrusted input, secrets, tokens, TLS, cookies | [go-security](../go-security/SKILL.md) | [go-defensive](../go-defensive/SKILL.md) |
| retries, idempotency, circuit breakers, overload, backpressure, fallback | [go-resilience](../go-resilience/SKILL.md) | the HTTP, SQL, context, or concurrency owner when its mechanics change |
| ticket, wrong result, environment regression, panic, hang, leak, flaky test; cause unknown | [go-troubleshooting](../go-troubleshooting/SKILL.md) | the owner of the mechanism once found |
| JSON and other wire formats, struct tags | [go-defensive](../go-defensive/SKILL.md) (tags) | [go-packages](../go-packages/SKILL.md) (`json/v2` on the ladder) |
| CLI entry point, flags, `main`/`run` | [go-packages](../go-packages/SKILL.md) | — |
| a list of findings from a review or audit | the rows the findings name | per area, not per task |

## Writing New Code

Start at the required entry point and its callers. Implement the documented
success and failure paths within the existing API; design a new API only when
the task calls for one. A stub's `panic("not implemented")` is missing behavior,
not a refactor contract: replace it and satisfy the acceptance tests. Use
baseline/after equivalence only for existing behavior the task must preserve.

Check that an existing API's defaults satisfy the contract before wrapping or
replacing it; prefer a small adapter for a semantic mismatch. Route the rest:

- Helper or inline: [go-code-refactor](../go-code-refactor/SKILL.md#delete-before-you-restructure) owns the helper rule.
- An interface for a consumer's substitution boundary: [go-interfaces](../go-interfaces/SKILL.md).
- Deriving the checks from the task's requirements: [go-testing](../go-testing/SKILL.md).
- Self-audit of the new code before closing: the audit lane in [OVER-ENGINEERING.md](../go-code-refactor/references/OVER-ENGINEERING.md#tags); no separate written audit is required.

## Close With The Gate

Use [go-linting](../go-linting/SKILL.md#verification-gate) to select checks for the requested
scope. Repository gates take precedence over defaults; do not union them.
Inspect the final diff and complete authorized work. Reuse passing results
for unchanged code; rerun affected checks after edits and honor host checkpoints.

Use bundled checks when they add evidence beyond that gate:

- Refactor: `../go-code-refactor/scripts/verify-refactor.sh` for baseline/after
  evidence; `../go-code-refactor/scripts/check-debt.sh` for deliberate `Kept:`
  markers, whose format is owned by `go-code-refactor`.
- Error handling: `../go-error-handling/scripts/check-errors.sh`.
- Exported API documentation: `../go-documentation/scripts/check-docs.sh`.
- Before submitting: [go-code-review](../go-code-review/SKILL.md).

Run routine checks inline. Claude Code's `go-verify` agent is optional. The
PostToolUse hook reports only gofmt and vet findings for the edited package and
is silent on success: its silence is not a passing result. Count only results
observed for the current diff and scope; report each check as `pass`, `fail`,
`unavailable (reason)`, or `skipped (reason)` per go-linting.

## Related Skills

- Every owner in [Route Before The First Edit](#route-before-the-first-edit).
- [go-style-core](../go-style-core/SKILL.md): house style, report length, and delegation.
- [go-linting](../go-linting/SKILL.md): the closing gate.
- [go-code-review](../go-code-review/SKILL.md): the checklist before submitting.
