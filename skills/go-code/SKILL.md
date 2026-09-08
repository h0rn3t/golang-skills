---
name: go-code
description: Use when writing, fixing, or refactoring Go code, whether or not the task has one obvious topic — it loads the go-* skills the task needs and runs the closing gate. Use it too when /go-code modifies another workflow (e.g. /opsx:apply /go-code); as a modifier it selects Go rules, not a change name or path.
---

# Go Code Profile

Route Go work to the relevant owners, then close with their verification gate.

## Resource Routing

Resolve sibling skills and references relative to this installed directory;
run their scripts from the target project. Use the host's skill loader or read
`SKILL.md` directly. Missing resources: report the gap, use available guidance,
and continue independent authorized work without inventing rules.

- `../go-style-core/SKILL.md` — Read once per task for house style, fallback
  rules, and communication guidance. Read its references only for a decision
  the task requires.
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
2. **Read before routing.** Load `go-style-core`, inspect repository instructions,
   `go.mod`, neighboring code, tests, and callers. Identify required inputs,
   outputs, environment constraints, failure behavior, and dependencies. Use
   these to select rows below; no fixed written contract report is required.
3. **Load the relevant owners before editing.** Select by the decisions and
   behavior being changed, not every syntax element present. A routine local
   variable or `if` does not trigger naming, documentation, or extra style
   references. For mixed tasks, route each area; add owners when new evidence
   requires them. There is no numerical cap. Reuse guidance already read.
4. **Implement the authorized scope.** Apply the
   [delete-first priority](../go-code-refactor/SKILL.md#delete-before-you-restructure):
   after understanding the code, apply the ponytail ladder to each proposed
   helper, type, layer, option, or import: (1) omit speculative work,
   (2) reuse existing code, (3) use the standard library, (4) use language or
   platform features, (5) use an existing dependency, (6) use one clear line
   when sufficient, (7) otherwise write the minimum that works. Stop at the
   first sufficient rung; clarity and correctness outrank brevity. Judge the
   final code and call sites, not just declaration count. Preserve
   required behavior, validation, security controls, and meaningful tests.
   Resolve routine choices without stopping; ask only for missing information
   that changes correctness, scope, or authorization.
5. **Verify and report.** Follow the closing gate below and `go-style-core`'s
   communication guidance. Report outcomes, observed checks, and material gaps.

## Related Skills

### Route Before The First Edit

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


## Close With The Gate

Use [go-linting](../go-linting/SKILL.md) to select checks for the requested
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

Run routine checks inline. Claude Code's `go-verify` agent and PostToolUse hook
are optional: count only results actually observed for the current diff and
scope. Otherwise run the selected checks directly. Report skipped or unavailable
checks accurately; never infer success from an unobserved hook.
