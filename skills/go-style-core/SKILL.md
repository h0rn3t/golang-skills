---
name: go-style-core
description: Use when working with Go formatting, line length, nesting, naked returns, semicolons, or core style principles. Also use when a style question isn't covered by a more specific skill, even if the user doesn't reference a specific style rule. Does not cover domain-specific patterns like error handling, naming, or testing (see specialized skills). Acts as fallback when no more specific style skill applies.
---

# Go Style Core Principles

## Resource Routing

- `references/PRINCIPLES.md` - Read when resolving conflicts between clarity, simplicity, concision, maintainability, and consistency.
- `references/FORMATTING.md` - Read when handling gofmt, line breaks, whitespace, comments, or semicolons.

## Style Principles (Priority Order)

When writing readable Go code, apply these principles in order of importance:

### Priority Order

1. **Clarity** — Can a reader understand the code without extra context?
2. **Simplicity** — Is this the simplest way to accomplish the goal?
3. **Concision** — Does every line earn its place?
4. **Maintainability** — Will this be easy to modify later?
5. **Consistency** — Does it match surrounding code and project conventions?

---

## House Style Wins

> **Owner**: this skill owns the consistency rule. Other skills route here
> instead of restating it.

Follow the host's instruction hierarchy. Within it, explicit user requirements
and repository instructions take precedence over these skill defaults. Read
`AGENTS.md`, `CLAUDE.md` where present, `.golangci.yml`, `CONTRIBUTING.md`, and
neighboring code before editing. Skills do not authorize extra work or require
renewed approval for work the user already authorized.

- Assertion style, error-wrapping style, logger, test layout, and the `_`
  global prefix follow the nearest existing code, not the guide.
- Introduce a convention the guide prefers only in new code with no neighbor
  to match, or as a whole-package migration the user asked for.
- A bug is not house style. Fix it within the authorized scope; report unrelated
  findings separately. A review-only request remains read-only.

---

## Formatting

Run `gofmt` — no exceptions. There is **no rigid line length limit**, but Uber suggests a soft limit of 99 characters. Break by semantics, not length — refactor rather than just wrap.

## Write Current Go

> **Normative**: Match the toolchain, not the codebase's oldest habits. Code
> written in a superseded idiom is a style defect even when it compiles.

Respect `go.mod`, build constraints, and supported CI toolchains; an installed
newer Go version does not authorize a version bump. A scoped `go fix -diff`
previews modernization. It flags
the patterns Go has since replaced — `x := x` loop captures, three-clause
counting loops, `sort.Slice`, `interface{}`, `wg.Add`/`Done` bookkeeping,
`errors.As`, hand-written `min`/`max`; [go-linting](../go-linting/SKILL.md)
lists the analyzers and owns the gate. `go fix` catches only what has a
modernizer; for the rest, reach for the feature that replaces the block before
writing it — [OVER-ENGINEERING.md](../go-code-refactor/references/OVER-ENGINEERING.md#reach-for-what-go-ships) is the checklist.

Keep modernization within the task. If consistency would require unrelated
rewrites, preserve the local idiom and report the opportunity separately.

---

## Reduce Nesting

> **Owner**: this skill owns nesting depth, early returns, and unnecessary
> `else`. Other skills route here instead of restating the rule.

Handle error cases and special conditions first. Return early or continue the loop to keep the "happy path" unindented.

```go
// Bad: Deeply nested
for _, v := range data {
    if v.F1 == 1 {
        v = process(v)
        if err := v.Call(); err == nil {
            v.Send()
        } else {
            return err
        }
    } else {
        log.Printf("Invalid v: %v", v)
    }
}

// Good: Flat structure with early returns
for _, v := range data {
    if v.F1 != 1 {
        log.Printf("Invalid v: %v", v)
        continue
    }

    v = process(v)
    if err := v.Call(); err != nil {
        return err
    }
    v.Send()
}
```

### Unnecessary Else

If a variable is set in both branches of an if, use default + override pattern.

```go
// Bad: Setting in both branches
var a int
if b {
    a = 100
} else {
    a = 10
}

// Good: Default + override
a := 10
if b {
    a = 100
}
```

---

## Naked Returns

A `return` statement without arguments returns the named return values. This is
known as a "naked" return.

```go
func split(sum int) (x, y int) {
    x = sum * 4 / 9
    y = sum - x
    return // returns x, y
}
```

### Guidelines for Naked Returns

- **OK in small functions**: Naked returns are fine in functions that are just a
  handful of lines
- **Be explicit in medium+ functions**: Once a function grows to medium size, be
  explicit with return values for clarity
- **Don't name results just for naked returns**: Clarity of documentation is
  always more important than saving a line or two

```go
// Good: Small function, naked return is clear
func minMax(a, b int) (min, max int) {
    if a < b {
        min, max = a, b
    } else {
        min, max = b, a
    }
    return
}

// Good: Larger function, explicit return
func processData(data []byte) (result []byte, err error) {
    result = make([]byte, 0, len(data))

    for _, b := range data {
        if b == 0 {
            return nil, errors.New("null byte in data")
        }
        result = append(result, transform(b))
    }

    return result, nil // explicit: clearer in longer functions
}
```

See **go-documentation** for guidance on Named Result Parameters.

---

## Semicolons

The lexer inserts a semicolon after any line ending in an identifier, literal,
`return`, `)`, or `}`, so an opening brace on its own line is a syntax error,
not a style choice — `gofmt` settles the rest. Explicit semicolons belong only
in `for` clauses.

---

## How Much To Say

> **Owner**: this skill owns how the agent talks while it works, how long its
> written output is, and when it delegates. Other skills route here.

Follow the host's communication requirements. Give a short initial update and
meaningful progress updates during longer work: findings, decisions, blockers,
or the next check. Avoid narrating every read. Close with the outcome, observed
verification results, and material limitations; never imply a skipped check ran.

Size anything written to disk — a report, a review, a design note — to what the
task needs. The `assets/` templates are the shape; filler sections, restated
summaries, and boilerplate are noise, and an empty section is one line.

Keep routine edits and checks inline. When the user or host authorizes parallel
work, delegate only bounded, independent tasks with clear ownership and useful
work remaining locally. Do not spawn a second agent merely to repeat a completed
check. A requested independent review is a separate task. Agent availability,
model choice, and delegation limits belong to the host, not to a Go style rule.

## Related Skills

- **Naming conventions**: See [go-naming](../go-naming/SKILL.md) when applying MixedCaps, choosing identifier names, or resolving naming debates
- **Error flow**: See [go-error-handling](../go-error-handling/SKILL.md) when choosing an error strategy, wrapping errors, or deciding log-vs-return
- **Statement mechanics**: See [go-control-flow](../go-control-flow/SKILL.md) when writing `if`-init, `range` loops, switch, or type switches
- **Documentation**: See [go-documentation](../go-documentation/SKILL.md) when writing doc comments, named return parameters, or package-level docs
- **Linting enforcement**: See [go-linting](../go-linting/SKILL.md) when automating style checks with golangci-lint or configuring CI
- **Code review**: See [go-code-review](../go-code-review/SKILL.md) when applying style principles during a systematic code review
- **Applying these rules to existing code**: See [go-code-refactor](../go-code-refactor/SKILL.md) when the fix is a multi-step behavior-preserving refactor rather than a single style call
- **Logging style**: See [go-logging](../go-logging/SKILL.md) when reviewing logging practices, choosing between log and slog, or structuring log output
