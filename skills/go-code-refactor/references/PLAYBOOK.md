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
- [2. Extract until one job per function](#2-extract-until-one-job-per-function)
- [3. Rename for the reader](#3-rename-for-the-reader)
- [4. Name the magic values](#4-name-the-magic-values)
- [5. Error handling in an existing codebase](#5-error-handling-in-an-existing-codebase)
- [6. Reduce what is in scope](#6-reduce-what-is-in-scope)
- [7. Comments that earn their place](#7-comments-that-earn-their-place)
- [8. Anti-patterns of "cleanup"](#8-anti-patterns-of-cleanup)
- [9. Test readability](#9-test-readability)

---

## The readability hierarchy

The Google style guide's order — clarity, simplicity, concision,
maintainability, consistency — is the tie-breaker for every judgment call below.
[go-style-core](../../go-style-core/SKILL.md) owns it.

Concision is *third*: when a shorter form costs clarity, the hierarchy already
made the call. The inverse matters too — repetitive code is a concision failure
precisely because it hides the one difference between near-identical blocks, so
factoring it out serves clarity, not just line count.

Two corollaries the guide states outright. **Least mechanism**: prefer the most
standard tool — core language construct first (channel, slice, map, loop,
struct), then stdlib, then anything heavier. **Complexity must be deliberate**:
where genuinely required it stays *and* gets a comment explaining why;
unexplained complexity in a simple-purpose function is a signal to simplify.

### Signal-boost the unusual variant

Common idioms are read by pattern recognition. When code is *almost* a common
idiom but differs in a load-bearing way, boost the difference —
`if err := doSomething(); err == nil { // if NO error` — so the reader does not
glide past it. One of the few places a "what" comment earns its place.

---

## 0. Delete first

Nothing you write reads as well as code that is not there. Do this pass before
any restructuring: it shrinks the problem, and none of it needs a design
decision from you.

```go
// Dead branch — the type system already guarantees it
if items == nil { // range over nil is fine; this check does nothing
    return
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
handling for a function that cannot fail; `len(s) > 0` before a `range`;
`if b == true`; string conversions of strings.

The line to hold: **provably** unreachable. "Nothing calls this" needs `go vet`,
a linter with `unused`, or a repo-wide grep including tests, generated code,
and reflection-based dispatch — an exported symbol may have callers outside the
module. When you cannot prove it, it is a finding, not a deletion. Deletions
belong at the top of the report; reviewers approve them at a glance.

---

## 1. Flatten with early returns

The highest-yield change in most Go code. Keep the happy path at the leftmost
indent and let failures exit.

```go
// before
func process(u *User) error {
    if u != nil {
        if u.Active {
            if err := validate(u); err == nil {
                return save(u)
            } else {
                return err
            }
        } else {
            return ErrInactive
        }
    }
    return ErrNilUser
}

// after
func process(u *User) error {
    if u == nil {
        return ErrNilUser
    }
    if !u.Active {
        return ErrInactive
    }
    if err := validate(u); err != nil {
        return err
    }
    return save(u)
}
```

Rules that fall out: no `else` after a `return`; handle the error case first;
scope `err` into the `if` when it is not used later
([go-control-flow](../../go-control-flow/SKILL.md)).

**Careful**: do not hoist a condition's subexpressions into variables above the
`if` — that defeats short-circuiting and can panic where the original did not.

---

## 2. Extract until one job per function

Long functions are the dominant readability failure in Go services. The test is
not line count but **mixed abstraction levels**: one function that both parses
an HTTP body and builds SQL makes the reader context-switch mid-scroll.

```go
// after: a 150-line handler becomes a thin orchestrator
func (s *Server) HandleOrder(w http.ResponseWriter, r *http.Request) {
    req, err := decodeOrderRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err)
        return
    }
    if err := validateOrder(req); err != nil {
        writeError(w, http.StatusUnprocessableEntity, err)
        return
    }
    order := buildOrder(req, s.pricing)
    if err := s.orders.Save(r.Context(), order); err != nil {
        writeError(w, http.StatusInternalServerError, err)
        return
    }
    writeJSON(w, http.StatusOK, orderResponse(order))
}
```

A well-named call is documentation the compiler checks. Extraction notes:

- Keep helpers unexported and near their caller; a new file per concern only
  when the file itself is unwieldy.
- A helper needing six parameters means the split is along the wrong seam —
  look for a struct that already groups them.
- **Preserve the original order of side effects exactly.** Extraction is safe;
  reordering is not.
- Extracted functions inherit the caller's context — pass `ctx`, never create
  a new one ([go-context](../../go-context/SKILL.md)).

---

## 3. Rename for the reader

[go-naming](../../go-naming/SKILL.md) owns the rules; what matters here is
scope. Renaming unexported identifiers is free. Renaming an **exported**
identifier or a **struct tag key** is an API or wire-format change — findings
list, not the diff.

Both directions are worth fixing: names too short for a wide scope (`d`, `tmp`
living 40 lines) and names too long for a narrow one (`elementIndex` as a
three-line loop variable). Go scales name length *with* scope.

---

## 4. Name the magic values

`if resp.StatusCode == 429 { time.Sleep(3 * time.Second) }` becomes
`if resp.StatusCode == http.StatusTooManyRequests { time.Sleep(retryAfter) }`.

Prefer standard-library constants where they exist. For enum-like sets use a
defined type plus `iota` — **only if the numeric values stay identical**, since
they may be persisted or sent over the wire. `0`, `1`, `""`, `-1` in obvious
positions need no name; the goal is removing questions, not maximizing
constants.

---

## 5. Error handling in an existing codebase

[go-error-handling](../../go-error-handling/SKILL.md) owns the strategy. The
refactor-specific constraint: **error message text is API** — callers grep
logs, tests assert on it, alerts match it.

| Change | Verdict |
|---|---|
| Scope `err` into the `if`; drop `else` after an error return | Free |
| `if err == X \|\| err == Y` → `errors.Is` | Free |
| `errors.As` → `errors.AsType[T]` | Free (Go 1.26+) |
| Rewording an existing message | Findings list |
| `%v` ↔ `%w` | Findings list — changes what `errors.Is`/`AsType` see |
| `strings.Contains(err.Error(), ...)` → `errors.Is` | Findings list — matches a different error set |
| Log-and-return → handle once | Findings list — changes log output |
| Exported func returning a concrete error type | Findings list — API change |
| Bare `_ = f()` with no justifying comment | Findings list |

Messages **you** author follow the owner skill: lowercase, no trailing
punctuation, one clause of new context, `: %w` at the end.

---

## 6. Reduce what is in scope

Readers hold live variables in their head, and each one costs. Declare at first
use rather than at the top of the function; scope to the smallest block
(`if v, err := f(); err != nil`); inline variables used once unless the name
does explanatory work; eliminate accidental shadowing — two `err`s at different
depths is a reliable source of confusion; group related package-level
declarations into one block. See
[go-declarations](../../go-declarations/SKILL.md).

---

## 7. Comments that earn their place

Delete comments that restate the code (`counter++ // increment counter`); they
add scroll and go stale. Keep and add comments that explain **why** — the
non-obvious constraint, the workaround for an upstream bug, the reason an
ordering matters:

```go
// The vendor API rejects bursts above 10 rps, so we pace even on retries.
time.Sleep(rateLimitInterval)
```

Adding a doc comment to an undocumented exported symbol is pure gain — nothing
changes at runtime. See [go-documentation](../../go-documentation/SKILL.md).
Leave `TODO`/`FIXME` in place; they are someone's open thread. Delete only
those describing work that demonstrably shipped.

---

## 8. Anti-patterns of "cleanup"

Things that feel like improvement and are not. The common thread: each one
*adds* something in the name of cleanliness.

- **Interfaces with one implementation**, added "for testability". They move
  the definition away from the usage and force readers to chase.
- **Generics where a concrete type worked.** Below roughly three call sites,
  type parameters cost more than the duplication they remove.
- **Merging similar-looking code that means different things.** Two rhyming
  10-line blocks are cheaper than one 15-line function with a mode flag.
- **Splitting so far that following one request means opening eight
  functions.** The goal is one job per function, not one statement.
- **A `util`, `helpers`, or `common` package.** `util.Process(x)` says nothing,
  grab-bags grow forever, and the names collide on import.
- **A new dependency** for what the standard library covers
  ([go-packages](../../go-packages/SKILL.md)), or **clever over boring** — a
  bit-twiddling one-liner replacing an obvious loop is shorter and worse.
- **Reformatting the whole file**, so the meaningful diff drowns in whitespace.
  Run `gofmt`; stop there.
- **Blanket capacity hints** as drive-by optimization. A size hint earns its
  place when the size is known and the allocation was shown to matter.
- **Fixing bugs mid-refactor**, or **editing generated files** — the first
  breaks reviewability, the second reverts at the next generation.

The one that looks like restraint but is not: **removing a check because it
seemed redundant.** Validation at a trust boundary, error handling that
prevents data loss, and security checks stay, even when they are the ugliest
lines in the file. Prove it unnecessary, or leave it and say why.

---

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
