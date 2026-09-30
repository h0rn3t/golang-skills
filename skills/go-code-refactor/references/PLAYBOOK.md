# Refactoring Playbook

> Sources: source/google-go-styleguide/guide.md, decisions.md; source/effective-go/effective_go.html
> Authority: advisory, except the readability hierarchy (normative in the Google guide)
> Minimum Go: 1.27 baseline
> Last verified: 2026-08-29

Transformations ordered roughly by payoff. Rules another skill owns are routed
there rather than restated — this file covers only how to *apply* them to code
that already exists.

## Contents

- [The readability hierarchy](#the-readability-hierarchy)
- [0. Delete first](#0-delete-first)
- [1. Flatten with early returns](#1-flatten-with-early-returns)
- [2. Extract meaningful operations](#2-extract-meaningful-operations)
- [3. Rename for the reader](#3-rename-for-the-reader)
- [4. Name the magic values](#4-name-the-magic-values)
- [5. Error handling in an existing codebase](#5-error-handling-in-an-existing-codebase)
- [6. Reduce what is in scope](#6-reduce-what-is-in-scope)
- [7. Comments that earn their place](#7-comments-that-earn-their-place)
- [8. Anti-patterns of "cleanup"](#8-anti-patterns-of-cleanup)
- [9. Test readability](#9-test-readability)

## The readability hierarchy

The Google style guide's order — clarity, simplicity, concision,
maintainability, consistency — is the tie-breaker for every judgment call below.
[go-style-core](../../go-style-core/SKILL.md) owns it.

### Signal-boost the unusual variant

Common idioms are read by pattern recognition. When code is *almost* a common
idiom but differs in a load-bearing way, boost the difference —
`if err := doSomething(); err == nil { // if NO error` — so the reader does not
glide past it. One of the few places a "what" comment earns its place.

## 0. Delete first

Nothing you write reads as well as code that is not there. Do this pass before
any restructuring: it shrinks the problem, and none of it needs a design
decision from you.

```go
// Dead branch — range over nil runs zero times, and nothing follows the loop
func notifyAll(items []Item) {
    if items == nil { // live once a statement such as flush() follows the loop
        return
    }
    for _, it := range items {
        notify(it)
    }
}

// Redundant else after return
if err != nil {
    return err
} else { // adds a level for nothing
    return save(x)
}

// A wrapper with one caller and no added meaning
func doWork(x int) int { return compute(x) }
```

Also usually deletable: unused parameters and struct fields; unreachable
returns after a panic or exhaustive switch; commented-out code; `err != nil`
handling for a function that cannot fail; `len(s) > 0` around a `range` that
nothing follows; `if b == true`; string conversions of strings.

The line to hold: **provably** unreachable. "Nothing calls this" needs `go vet`,
a linter with `unused`, or a repo-wide grep including tests, generated code,
and reflection-based dispatch — an exported symbol may have callers outside the
module. When you cannot prove it, it is a finding, not a deletion. Deletions
belong at the top of the report; reviewers approve them at a glance. Duplication
that differs only in values folds instead: [POLICY-TABLES.md](POLICY-TABLES.md).

## 1. Flatten with early returns

The highest-yield change in most Go code. Keep the happy path at the leftmost
indent and let failures exit;
[go-style-core](../../go-style-core/SKILL.md#reduce-nesting) owns the rule, and
[BEHAVIOR-TRAPS.md](BEHAVIOR-TRAPS.md#evaluation-order) the short-circuit trap.

## 2. Extract meaningful operations

Apply the [helper rule](../SKILL.md#delete-before-you-restructure): function
length alone does not justify extraction; mixed abstraction levels can.
Assembling SQL — clauses, placeholders, argument order — is string work beside
the overdue rule, so it is the one step named here, even with one call site.
The cutoff, the scan, and the error wrapping stay inline. Request decoding is
not such a step: [go-http](../../go-http/SKILL.md) keeps it in the handler.

```go
func (s *Service) Overdue(ctx context.Context, now time.Time, region string) ([]Invoice, error) {
    cutoff := now.AddDate(0, 0, -s.graceDays) // past due once the grace period ends
    query, args := overdueQuery(cutoff, region)
    rows, err := s.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("overdue invoices: %w", err)
    }
    defer rows.Close()
    var out []Invoice
    for rows.Next() {
        var inv Invoice
        if err := rows.Scan(&inv.ID, &inv.DueAt); err != nil {
            return nil, fmt.Errorf("scan invoice: %w", err)
        }
        out = append(out, inv)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("overdue invoices: %w", err)
    }
    return out, nil
}

// overdueQuery builds the statement; an empty region selects every region.
func overdueQuery(cutoff time.Time, region string) (string, []any) {
    query := "SELECT id, due_at FROM invoices WHERE paid_at IS NULL AND due_at < $1"
    args := []any{cutoff}
    if region != "" {
        query += " AND region = $2"
        args = append(args, region)
    }
    return query + " ORDER BY due_at", args
}
```

Extraction notes:

- Keep helpers unexported and near their caller; a new file per concern only
  when the file itself is unwieldy.
- A helper needing six parameters means the split is along the wrong seam —
  look for a struct that already groups them.
- **Preserve the original order of side effects exactly.** Extraction is safe;
  reordering is not.
- Extracted functions inherit the caller's context — pass `ctx`, never create
  a new one ([go-context](../../go-context/SKILL.md)).

## 3. Rename for the reader

[go-naming](../../go-naming/SKILL.md) owns the rules; what matters here is
scope. Renaming an unexported identifier is free when no reflection, template,
`go:linkname`, or string-based use names it. Renaming an **exported**
identifier or a **struct tag key** is an API or wire-format change — findings
list, not the diff.

## 4. Name the magic values

Prefer standard-library constants where they exist. For enum-like sets use a
defined type plus `iota` — **only if the numeric values stay identical**, since
they may be persisted or sent over the wire.

## 5. Error handling in an existing codebase

[go-error-handling](../../go-error-handling/SKILL.md) owns the strategy. The
refactor-specific constraint: **error message text is API** — callers grep
logs, tests assert on it, alerts match it.

| Change | Verdict |
|---|---|
| Scope `err` into the `if`; drop `else` after an error return | Free |
| `if err == X \|\| err == Y` → `errors.Is` | Findings list — `Is` also matches a wrapped `X` and an error whose `Is` method claims `X`; equivalent only when every producer returns `X` unwrapped and no error in the chain has an `Is` method |
| `errors.As` → `errors.AsType[T]` | Free (Go 1.26+) |
| Rewording an existing message | Findings list |
| `%v` ↔ `%w` | Findings list — changes what `errors.Is`/`AsType` see |
| `strings.Contains(err.Error(), ...)` → `errors.Is` | Findings list — matches a different error set |
| Log-and-return → handle once | Findings list — changes log output |
| Exported func returning a concrete error type | Findings list — API change |
| Bare `_ = f()` with no justifying comment | Findings list |

Messages **you** author follow the owner skill: lowercase, no trailing
punctuation, one clause of new context, `: %w` at the end.

## 6. Reduce what is in scope

Declare at first use; scope to the smallest block; eliminate accidental
shadowing. See [go-style-core](../../go-style-core/SKILL.md).

## 7. Comments that earn their place

Delete comments that restate the code; keep and add comments that explain
**why**. Adding a doc comment to an undocumented exported symbol is pure gain —
nothing changes at runtime. See [go-documentation](../../go-documentation/SKILL.md).
Leave `TODO`/`FIXME` in place; they are someone's open thread. Delete only
those describing work that demonstrably shipped.

## 8. Anti-patterns of "cleanup"

Things that feel like improvement and are not.

- **Interfaces with one implementation**, added "for testability". They move
  the definition away from the usage and force readers to chase.
- **Merging similar-looking code that means different things.** Two rhyming
  10-line blocks are cheaper than one 15-line function with a mode flag.
- **Splitting so far that following one request means opening eight
  functions.** The goal is one job per function, not one statement.
- **Clever over boring** — a bit-twiddling one-liner replacing an obvious loop
  is shorter and worse.
- **Reformatting the whole file**, so the meaningful diff drowns in whitespace.
  Run `gofmt`; stop there.
- **Blanket capacity hints** as drive-by optimization. A size hint earns its
  place when the size is known and the allocation was shown to matter.

## 9. Test readability

Test code has no production observers, so the bar is lower — but the constraint
holds: improving how a test **reports** is fair game; changing what it
**asserts** is not. Never edit a test to agree with refactored code.

Free improvements: `t.Helper()` in a helper, the function name and inputs in a
failure message, the `(-want +got)` direction key on a `cmp.Diff` message, one
`cmp.Diff` replacing a field-by-field ladder of `if`s. See
[go-testing](../../go-testing/SKILL.md).

Converting repetitive test functions to table-driven form is worth *proposing*
— but it is its own diff, not a rider on a production refactor.
