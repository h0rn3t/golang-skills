# Catalog: Smell, Transform, Tool, Risk

> Sources: Fowler, *Refactoring* (2nd ed.); Feathers, *Working Effectively with Legacy Code*; golang.org/x/tools/gopls docs; golang/go#20744, golang/go#65552
> Authority: advisory — the transform inventory; behavior rules stay in SKILL.md
> Minimum Go: 1.27 baseline; gopls v0.20+
> Last verified: 2026-09-06

[PLAYBOOK.md](PLAYBOOK.md) owns the transforms that carry most refactors —
delete first, flatten, extract, rename, name the magic values. This file is the
long tail: moves that cross a function, a type, or a package boundary. Each
entry names the smell that triggers it, what the transform is in Go, the tool
that performs it, and its row in the risk table
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
- **In Go**: gopls substitutes rather than splices — an argument with a side
  effect is hoisted into a `var` temporary instead of being duplicated at every
  use of the parameter, implicit conversions at the call boundary become
  explicit, and a callee body containing `defer` is wrapped in an immediately
  invoked function literal so the deferred call still fires at the same point.
  It cannot inline through a dynamic dispatch (interface method, function value)
  or inline a generic function.
- **Tool**: `gopls codeaction -exec -kind=refactor.inline.call file.go:#offset`
- **Risk**: low. With rename, one of the two gopls actions that is behavior
  preserving by construction — it refuses rather than emit a wrong inline.

## Change function declaration

- **Smell**: a parameter list that outgrew what the function needs, or one
  parameter that stopped being relevant to its job.
- **In Go**: gopls covers only the narrow cases — drop a parameter nobody passes
  meaningfully, or swap two adjacent ones. Adding a parameter across every call
  site is not one action. Stage it instead: write the new variant beside the old
  one, migrate call sites with an `eg` template
  ([MECHANICAL.md](MECHANICAL.md)), verify, then delete the old function once
  nothing calls it. That keeps a wide signature change mechanical and reviewable
  rather than scattered by hand.
- **Tool**:

```bash
gopls codeaction -exec -kind=refactor.rewrite.removeUnusedParam file.go:#offset
gopls codeaction -exec -kind=refactor.rewrite.moveParamLeft     file.go:#offset
eg -t template.go -w ./...   # staged migration for an added parameter
```

- **Risk**: medium for one parameter and a handful of call sites; high when the
  signature moves across many callers with no one-shot action. An **exported**
  signature is an API change — findings list, not the diff (PLAYBOOK §3).

## Introduce parameter object

- **Smell**: data clumps — the same cluster of parameters recurring across
  several functions — or a long parameter list where several parameters are
  conceptually one unit.
- **In Go**: define a struct for the recurring group and take it instead of the
  loose values. gopls has no action for this yet (golang/go#65552 tracks one),
  so it is a manual struct plus the signature change above.
- **Risk**: medium. Whether the struct should instead configure construction is
  [go-functions](../../go-functions/SKILL.md)'s call, not this file's.

## Move function, field, or type

- **Smell**: feature envy — a function reads and writes another package's data
  more than its own — or a type whose responsibilities plainly belong elsewhere.
- **In Go**: no gopls action moves a symbol across a package boundary. Use the
  type-alias gradual repair sequence in [STRUCTURAL.md](STRUCTURAL.md): define
  the symbol in its new home, leave an alias or thin wrapper behind, migrate
  callers incrementally, delete the old name last. Splitting a large file into
  another file *inside the same package* is a one-shot action and is often
  enough on its own.
- **Tool**:

```bash
gopls codeaction -exec -kind=refactor.extract.toNewFile file.go:#start,#end
```

- **Risk**: high across packages; low for a same-package file split.

## Split or merge a package

- **Smell**: a god package, or divergent change — one package keeps changing for
  several unrelated reasons because it hosts several unrelated concerns.
- **In Go**: gopls has an experimental action that partitions top-level
  declarations into acyclic components. Read its output as a draft: package
  boundaries encode API and ownership decisions no tool can see. Merging has no
  action at all — move the declarations and break the resulting cycle with a
  consumer-side interface ([STRUCTURAL.md](STRUCTURAL.md)) before reaching for
  anything larger.
- **Tool**: `gopls codeaction -exec -kind=source.splitPackage file.go`
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

## Hide delegate and remove middle man

- **Smell**: message chains (`a.B().C().D()`) reach through several objects to
  get to the one that matters; a middle man is a method whose whole body is
  `return x.SameMethod(...)`.
- **In Go**: embedding is the usual mechanism for hiding a delegate — it
  promotes the delegate's methods onto the outer type, so callers stop chaining.
  Removing a middle man is the inverse: inline the pass-through at each call
  site, then delete it.
- **Tool**: manual for the embedding decision; `refactor.inline.call` mechanizes
  the removal once callers are ready.
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
- **Tool**: `gopls rename` for the rename half; the new code is written and
  tested like any other.
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
- **Risk**: low only under that guard. Reversible with
  `refactor.inline.call` at any time.

---

## Smell to transform

| Smell | Transform |
|---|---|
| Long function | Extract until one job (PLAYBOOK §2) |
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
