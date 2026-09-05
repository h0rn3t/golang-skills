---
name: go-style-core
description: Use when resolving Go style or language-mechanics questions about formatting, nesting, declarations, initialization, variable scope, shadowing, loops, switches, or enum zero values. Provides the baseline for go-code and the fallback for style questions without a specialized owner. Function API design, naming, error strategy, and testing belong to their specialized skills.
---

# Go Style and Language Mechanics

Apply the baseline below; read detailed syntax guidance only for the decision
the task requires. An ordinary function edit does not require every reference.

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). Respect the target
> module's language version; references name version-sensitive features inline.

## Resource Routing

- `references/PRINCIPLES.md` - Read when resolving a tradeoff between clarity, simplicity, concision, maintainability, and consistency.
- `references/FORMATTING.md` - Read for line breaks, whitespace, comments, and semicolon mechanics.
- `references/SCOPE.md` - Read for `var` vs `:=`, grouping declarations, if-init, and reassignment across scopes.
- `references/SHADOWING.md` - Read when an inner declaration hides an outer variable or a predeclared identifier.
- `references/IOTA.md` - Read when designing enum defaults, bitmasks, or grouped constants.
- `references/INITIALIZATION.md` - Read for struct/map initialization, keyed literals, zero values, and pointers to optional values.
- `references/CONTROL-FLOW.md` - Read when choosing loop/range forms, writing iterators, or preserving iteration behavior.
- `references/SWITCH-PATTERNS.md` - Read for expression switches, fallthrough, and labeled breaks; route interface semantics to go-interfaces.
- `references/BLANK-IDENTIFIER.md` - Read for intentional discards and side-effect imports; route interface assertions to go-interfaces.

## Style Principles

Resolve readability tradeoffs in this order: clarity, simplicity, concision,
maintainability, consistency. Use the least mechanism that delivers the user's
requirements. These defaults operate within the precedence below.

## House Style Wins

Follow the host's instruction hierarchy. Within it, explicit user requirements
and repository instructions take precedence over these skill defaults. Read
`AGENTS.md`, `CLAUDE.md` where present, `.golangci.yml`, `CONTRIBUTING.md`, and
neighboring code before editing. Skills do not authorize extra work or require
renewed approval for work the user already authorized.

- Assertion style, error-wrapping style, logger, test layout, and the `_`
  global prefix follow the nearest existing code.
- Introduce a convention the guide prefers only in new code with no neighbor
  to match, or as a whole-package migration the user asked for.
- A bug is not house style. Fix it within the authorized scope; report unrelated
  findings separately. A review-only request remains read-only.

## Formatting

Use `gofmt` for Go source. This guide imposes no rigid line-length limit;
break by meaning and readability, while respecting repository requirements.

## Write Current Go

Respect `go.mod`, build constraints, and supported CI toolchains; an installed
newer Go version does not authorize a version bump. A scoped `go fix -diff`
previews modernization. [go-linting](../go-linting/SKILL.md) owns the analyzers
and verification gate; [OVER-ENGINEERING.md](../go-code-refactor/references/OVER-ENGINEERING.md#reach-for-what-go-ships)
lists standard-library replacements beyond the automated modernizers.

Keep modernization within the task. If consistency would require unrelated
rewrites, preserve the local idiom and report the opportunity separately.

## Reduce Nesting

Handle errors and special conditions first. Return early or continue the loop
so the success path stays unindented. Preserve the order of validation, side
effects, and returned errors when flattening existing code.
Negate the original predicate exactly: for floating-point values, `!(x > 0)`
also rejects NaN, while `x <= 0` does not. Preserve short-circuit evaluation.

### Unnecessary Else

Omit `else` after a branch that exits. For two branches assigning one value,
use default plus override when the default is safe to evaluate unconditionally;
keep the branches when evaluation has side effects or is expensive.

## Declarations and Scope

- Use `:=` for local explicit values; `var` for intentional zero values,
  top-level variables, or a type that differs from the expression.
- Keep declarations near use. Use if-init when the value is confined to the
  conditional; keep it outside when needed afterward or to avoid nesting.
- Check whether `:=` reassigns in the same scope or shadows an outer variable.
- Choose enum zero values deliberately: valid useful default or invalid/unset.
- Preserve nil/empty and explicit-zero distinctions required by the API.
  Prefer keyed struct literals; leave literal layout and examples to the reference.

## Loops and Switches

Preserve iteration order, value semantics, and exit targets when changing a
loop. Map order is unspecified; string range yields byte offsets and runes.
An iterator must stop when `yield` returns false. A `break` inside a switch
exits that switch; use a label when the intended target is the enclosing loop.
Read the relevant reference before adopting a version-sensitive loop form.

## Naked Returns

Use explicit return values once a function is too long to see its named
results easily. A naked return is acceptable in a handful of clear lines;
do not name results solely to omit them at `return`. Named-result documentation
belongs to [go-documentation](../go-documentation/SKILL.md).

## How Much To Say

This skill owns narration, report length, and delegation guidance for the pack.
Follow the host's communication requirements. Give a short initial update and
meaningful progress updates during longer work: findings, decisions, blockers,
or the next check. Avoid narrating every read. Close with the outcome, observed
verification results, and material limitations; never imply a skipped check ran.

Size reports, reviews, and design notes to the task. Use applicable `assets/`
templates without filler sections or repeated summaries.

Keep routine edits and checks inline. When the user or host authorizes parallel
work, delegate only bounded, independent tasks with clear ownership and useful
work remaining locally. Do not spawn a second agent merely to repeat a completed
check. A requested independent review is a separate task. Agent availability,
model choice, and delegation limits belong to the host, not to a Go style rule.

## Related Skills

- **Function APIs**: [go-functions](../go-functions/SKILL.md) for signatures, constructors, config structs, and functional options.
- **Naming**: [go-naming](../go-naming/SKILL.md) for identifiers and receiver names.
- **Errors**: [go-error-handling](../go-error-handling/SKILL.md) for error strategy, wrapping, and log-vs-return.
- **Interfaces**: [go-interfaces](../go-interfaces/SKILL.md) for type assertions, type switches, and compile-time checks.
- **Collections**: [go-data-structures](../go-data-structures/SKILL.md) for choosing and owning slices/maps; [go-performance](../go-performance/SKILL.md) for capacity hints.
- **Documentation**: [go-documentation](../go-documentation/SKILL.md) for exported API comments and examples.
- **Verification**: [go-linting](../go-linting/SKILL.md) for the shared gate and CI configuration.
- **Review and refactoring**: [go-code-review](../go-code-review/SKILL.md) for a systematic review; [go-code-refactor](../go-code-refactor/SKILL.md) for behavior-preserving restructuring.
