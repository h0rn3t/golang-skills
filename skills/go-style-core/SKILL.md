---
name: go-style-core
description: Use when resolving Go style or language mechanics such as formatting, nesting, declarations, initialization, scope, shadowing, loops, switches, or enum zero values. API design, naming, errors, and tests have specialized skills.
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
- `references/INITIALIZATION.md` - Read for struct/map initialization, keyed literals (including embedded fields in Go 1.27), zero values, and pointers to optional values.
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
`AGENTS.md`, `CLAUDE.md`, `CONTRIBUTING.md`, `.golangci.yml`, the CI
configuration, and neighboring code before editing. Skills do not authorize
extra work or require renewed approval for work the user already authorized.

- Assertion style, error-wrapping style, logger, test layout, and the `_`
  global prefix follow the nearest existing code.
- Introduce a *convention* the guide prefers only in new code with no neighbor
  to match, or as a whole-package migration the user asked for. An older Go
  idiom in the neighbor is not a convention:
  [Write Current Go](#write-current-go) outranks it.
- A bug is not house style. Fix it within the authorized scope; report unrelated
  findings separately. A review-only request remains read-only.

## Formatting

Use `gofmt` for Go source. This guide imposes no rigid line-length limit;
break by meaning and readability, while respecting repository requirements.

A comment states what the code cannot show: a constraint, a default
deliberately overridden, the clause a branch serves. Code that reads as its
documentation reads carries none; a comment that narrates the next line or
argues that a change is correct is removed before the diff closes. Match the
neighboring code's comment density. Doc comments on exported API belong to
[go-documentation](../go-documentation/SKILL.md).

## Write Current Go

> **Normative**: The module's `go` directive, not the neighboring code, sets
> the idiom. Every line you write or edit uses the current language form and
> standard-library API available at that version, even when the rest of the
> file keeps the older form. Consistency with an older neighbor is not a
> reason to write the older form.

```go
// At go 1.27, the line you write:
return slices.ContainsFunc(items, func(it Item) bool { return it.ID == id })

// The neighbor two functions up stays as it is — not rewritten, not copied:
for i := 0; i < len(items); i++ {
    if items[i].ID == id {
        return true
    }
}
```

Write the older form only when one of three things is true, and name which in
the report: the current form does not compile at the `go` directive (`go vet`'s
`stdversion` reports library symbols; `COMPATIBILITY.md` lists the language
features); it changes observable behavior (the tiers in
[MODERNIZATION.md](../go-code-refactor/references/MODERNIZATION.md) say which
swaps do); or it does not fit the code at hand. An installed newer toolchain
does not authorize a version bump; respect build constraints and the CI
toolchains.

The rule covers idioms — language features and standard-library APIs — not
dependencies: the logger, assertion library, router, or ORM the package already
uses stays ([House Style Wins](#house-style-wins);
[go-packages](../go-packages/SKILL.md) owns replacing one). Scope stays too:
untouched neighbors are not rewritten for consistency, and the report names
the older forms left in place as an opportunity. A scoped `go fix -diff`
previews what the mechanical modernizers would change, and the plugin's edit
hook prints it after each edit where installed.
[go-linting](../go-linting/SKILL.md) owns the analyzers and verification gate;
[OVER-ENGINEERING.md](../go-code-refactor/references/OVER-ENGINEERING.md#reach-for-what-go-ships)
lists the standard-library replacements beyond the automated modernizers.

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

- [go-functions](../go-functions/SKILL.md): signatures, constructors, config structs, functional options.
- [go-naming](../go-naming/SKILL.md): identifiers and receiver names.
- [go-error-handling](../go-error-handling/SKILL.md): error strategy, wrapping, log-vs-return.
- [go-interfaces](../go-interfaces/SKILL.md): type assertions, type switches, compile-time checks.
- [go-data-structures](../go-data-structures/SKILL.md): slices and maps; [go-performance](../go-performance/SKILL.md): capacity hints.
- [go-documentation](../go-documentation/SKILL.md): exported API comments and examples.
- [go-linting](../go-linting/SKILL.md): the shared gate and CI configuration.
- [go-code-review](../go-code-review/SKILL.md): systematic review; [go-code-refactor](../go-code-refactor/SKILL.md): behavior-preserving restructuring.
