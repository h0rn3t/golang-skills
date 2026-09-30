# Catalog: Smell, Transform, Tool, Risk

> Sources: Fowler, *Refactoring* (2nd ed.); Feathers, *Working Effectively with Legacy Code*; `go tool fix help inline`; golang/go#20744
> Authority: advisory — the transform inventory; behavior rules stay in SKILL.md
> Minimum Go: 1.27 baseline
> Last verified: 2026-09-30

[PLAYBOOK.md](PLAYBOOK.md) owns the transforms that carry most refactors —
delete first, flatten, extract, rename, name the magic values. This file is the
long tail: moves that cross a function, a type, or a package boundary. Each
entry names the smell that triggers it, what the transform is in Go, the tool
that performs it where one exists, and its row in the risk table
([SKILL.md](../SKILL.md#risk-tiers)).

A transform in this file is a candidate, not a plan. The restraint ladder in
[OVER-ENGINEERING.md](OVER-ENGINEERING.md) runs first: several entries here add
a type or a layer, and adding one needs a reason the report can carry.

## Contents

- [Inline function](#inline-function)
- [Change function declaration](#change-function-declaration)
- [Introduce parameter object](#introduce-parameter-object)
- [Move function, field, or type](#move-function-field-or-type)
- [Split or merge a package](#split-or-merge-a-package)
- [Replace repeated switch with polymorphism](#replace-repeated-switch-with-polymorphism)
- [Hide delegate and remove middle man](#hide-delegate-and-remove-middle-man)
- [Sprout and wrap](#sprout-and-wrap)
- [Replace temp with query](#replace-temp-with-query)
- [Smell to transform](#smell-to-transform)

---

## Inline function

- **Smell**: a one-line forwarding function, or an indirection that has accreted
  no behavior of its own since it was introduced. Both are pure navigation cost.
- **In Go**: mark the function `//go:fix inline` and `go fix` substitutes its
  body at every call site, in this package and others. The inliner keeps
  argument evaluation order — an argument it cannot substitute safely is bound
  in a `var params = args` declaration instead of being duplicated — and it
  leaves alone a call it could only replace with a function literal (a callee
  body containing `defer`), a call through an interface method or a function
  value, and a call from the function's own test (`TestF` keeps calling `F`).
  Inline what it left by hand, then delete the function once nothing calls it.
- **Tool**: `go fix -inline -diff ./...` previews, `go fix -inline ./...`
  applies.
- **Risk**: low for what `go fix` inlines, which keeps behavior or leaves the
  call alone; medium for a call inlined by hand.

## Change function declaration

- **Smell**: a parameter list that outgrew what the function needs, or one
  parameter that stopped being relevant to its job.
- **In Go**: no single action changes a signature at every call site. Stage
  it: write the new variant beside the old one, make the old body one call to
  the new one and mark it `//go:fix inline` so `go fix` rewrites the call sites
  (an `eg` template in [MECHANICAL.md](MECHANICAL.md) covers what that body
  cannot express), verify, then delete the old function once nothing calls it.
  That keeps a wide signature change mechanical and reviewable rather than
  scattered by hand.
- **Tool**:

```bash
go fix -inline ./...         # the old function marked //go:fix inline
eg -t template.go -w ./...   # a migration the old body cannot express
```

- **Risk**: medium for one parameter and a handful of call sites; high when the
  signature moves across many callers with no one-shot action. An **exported**
  signature is an API change — findings list, not the diff (PLAYBOOK §3).

## Introduce parameter object

- **Smell**: data clumps — the same cluster of parameters recurring across
  several functions — or a long parameter list where several parameters are
  conceptually one unit.
- **In Go**: define a struct for the recurring group and take it instead of the
  loose values. No tool performs it: a manual struct plus the signature change
  above.
- **Risk**: medium. Whether the struct should instead configure construction is
  [go-functions](../../go-functions/SKILL.md)'s call, not this file's.

## Move function, field, or type

- **Smell**: feature envy — a function reads and writes another package's data
  more than its own — or a type whose responsibilities plainly belong elsewhere.
- **In Go**: no tool moves a symbol across a package boundary. Use the
  type-alias gradual repair sequence in [STRUCTURAL.md](STRUCTURAL.md): define
  the symbol in its new home, leave an alias or thin wrapper behind, migrate
  callers incrementally, delete the old name last. Moving declarations into
  another file *inside the same package* changes nothing a caller sees and is
  often enough on its own.
- **Risk**: high across packages; low for a same-package file split.

## Split or merge a package

- **Smell**: a god package, or divergent change — one package keeps changing for
  several unrelated reasons because it hosts several unrelated concerns.
- **In Go**: package boundaries encode API and ownership decisions no tool can
  see, so both directions are manual: move the declarations and break the
  resulting cycle with a consumer-side interface ([STRUCTURAL.md](STRUCTURAL.md))
  before reaching for anything larger.
- **Risk**: high. Target layout belongs to
  [go-packages](../../go-packages/SKILL.md).

## Replace repeated switch with polymorphism

- **Smell**: the **same** switch over a type tag or state value recurs at
  several call sites, so every new case has to be added in lock-step at each of
  them.
- **In Go**: an interface with one method per varying behavior, one implementing
  type per case, and callers invoking the method instead of switching.
- **Risk**: medium — and this is the entry most often reached for too early. One
  switch in one place is not this smell; PLAYBOOK §8 lists the interface added
  for a single implementation as an anti-pattern, and the ladder in
  [OVER-ENGINEERING.md](OVER-ENGINEERING.md) outranks this entry. The honest
  form of the trigger is three or more existing sites, today, that must change
  together. Anticipated future cases are not sites.

A table of data usually beats both shapes. When the cases differ only in
values — a rate, a prefix, a format string — the switch collapses into a map or
slice of structs with no interface and no new types at all, which is shorter
than either the repeated switch or the polymorphic hierarchy.
[Remove Duplication to the End](../SKILL.md#remove-duplication-to-the-end)
owns that fold and says when it is finished.

## Hide delegate and remove middle man

- **Smell**: message chains (`a.B().C().D()`) reach through several objects to
  get to the one that matters; a middle man is a method whose whole body is
  `return x.SameMethod(...)`.
- **In Go**: embedding is the usual mechanism for hiding a delegate — it
  promotes the delegate's methods onto the outer type, so callers stop chaining.
  Removing a middle man is the inverse: inline the pass-through at each call
  site, then delete it.
- **Tool**: manual for the embedding decision; `//go:fix inline` on the
  pass-through and `go fix -inline ./...` mechanize the removal once callers
  are ready.
- **Risk**: low to medium. Embedding has costs of its own —
  [go-interfaces](../../go-interfaces/SKILL.md) owns that call.

## Sprout and wrap

- **Smell**: new behavior is needed inside a function that has little or no test
  coverage. Editing it in place means no way to notice a break.
- **In Go**:
  - **Sprout** — write the new behavior as a new, fully tested function and call
    it from the one site that needs it. The old code is untouched.
  - **Wrap** — rename the original (`Save` → `saveInternal`), then add a method
    with the old name that calls it and adds the new behavior around it. Go has
    no method-wrapping mechanism, so this is a hand-written decorator.
- **Tool**: none needed; callers keep the old name, so only the original body
  moves, and the new code is written and tested like any other.
- **Risk**: low by construction — that is the point.
  [SAFETY-NET.md](SAFETY-NET.md) says when coverage makes this the right move
  instead of an in-place edit.

## Replace temp with query

- **Smell**: a local assigned once from an expression and read several times
  later, where the name reads like an input rather than a derived value.
- **In Go**: only replace it with a small unexported function when the expression
  is pure, stable, and cheap. Each read becomes a fresh evaluation, so time,
  randomness, I/O, counters, mutable state, allocation identity, and expensive
  computation must keep the temp's evaluate-once semantics.
- **Risk**: low only under that guard, and reversible at any time.

---

## Smell to transform

| Smell | Transform |
|---|---|
| Mixed abstraction levels | Extract meaningful operations (PLAYBOOK §2); length alone is not a reason |
| Deeply nested conditionals | Guard clauses (PLAYBOOK §1) |
| Long parameter list | Introduce parameter object, or [go-functions](../../go-functions/SKILL.md) for construction config |
| Data clumps | A struct for the recurring group |
| Primitive obsession | A defined type instead of a bare `string`/`int` |
| God package, divergent change | Split the package |
| Shotgun surgery — one change touches many packages | Move the concept into one package |
| Feature envy | Move the function to the package whose data it uses |
| The same switch repeated at several sites | Data table first, polymorphism second |
| Message chains | Hide the delegate |
| Middle man | Inline it away |
| Derived local read as an input | Replace temp with query |
| Untested code that must gain behavior | Sprout or wrap |

Divergent change and shotgun surgery look alike and pull opposite ways: one
package changing for many reasons splits apart; one reason rippling across many
packages consolidates.
