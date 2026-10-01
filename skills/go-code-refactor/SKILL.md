---
name: go-code-refactor
description: Use when refactoring, cleaning up, simplifying, restructuring, or modernizing existing Go code while keeping observable behavior identical — reducing nesting, splitting long functions, deleting dead code, renaming for clarity, or adopting newer Go APIs. Also use when a user hands over a Go file or package and calls it messy, hard to follow, too long, bloated, over-engineered, or outdated, even if they never say "refactor", and when an existing Go monolith needs an architecture proposal — layers, modules, package boundaries, a god package to split — before or instead of a rewrite. Does not cover writing new Go code or the style rules themselves (see go-style-core and the rule owners).
allowed-tools: Bash(bash:*)
---

# Go Refactoring

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`).

Improve readability while preserving observable behavior. Establish evidence
for that promise; compilation alone does not establish equivalent behavior.

## Resource Routing

Resolve resources from this installed skill directory; run scripts from the
target project using the resolved absolute script path.

- `../go-style-core/SKILL.md` - Load on every refactor before the first edit; it carries the idiom card.
- `references/BEHAVIOR-TRAPS.md` - Its Pre-commit checklist before every refactor; a section when a transform moves a `defer`, nil versus empty, goroutine or channel shape, or struct layout.
- `references/PLAYBOOK.md` - The concrete transformations, ordered by payoff.
- `references/POLICY-TABLES.md` - When repeated selection reads fields of one shared policy record.
- `references/CATALOG.md` - When a move crosses a function, type, or package boundary: the tool and the risk tier.
- `references/SAFETY-NET.md` - When the blast radius has thin or no tests.
- `references/MECHANICAL.md` - When the same edit recurs across many sites.
- `references/ARCHITECTURE.md` - Only when the smell is package-scale (a god package, shared global layers, drivers in services, a cycle) or the ask names architecture, layers, modules, or a monolith.
- `references/ARCHITECTURE-BEHAVIOR.md` - Before moving code a use case's atomicity, error identity, or authorization passes through.
- `references/ARCHITECTURE-CHECKS.md` - When installing or reading the architecture checker.
- `references/ARCHITECTURE-EXAMPLES.md` - Only when an example is needed to explain an architecture choice.
- `scripts/check-architecture.sh` - Run from the module root to check `internal/` imports against `architecture.json`; exit 1 on a violation or a stale `known` entry, 2 when a package fails to load or the policy is missing.
- `scripts/check-architecture.go` - The checker the wrapper builds; `scripts/check-architecture_test.go` holds its unit tests.
- `references/STRUCTURAL.md` - Before moving a type between packages, breaking an import cycle, or changing an exported API.
- `references/MODERNIZATION.md` - When a hunk adopts a newer API or `go fix -diff` proposes one: safe, conditional, report-only.
- `references/OVER-ENGINEERING.md` - When a step adds a helper, type, layer, option, or import, and when the ask is "what can we delete".
- `scripts/verify-refactor.sh` - Baseline and final check results, and production LOC before and after.
- `scripts/check-debt.sh` - Harvest `Kept:` markers and flag the ones naming no ceiling and no fix.
- `assets/refactor-report.md` - The final report structure.

The commands below need a shell tool; Workflow step 1 says how to write their
path.

## When Not to Refactor

> **Normative**: A refactor is an investment repaid by a future change. With no
> change coming to spend it on, it is churn carrying a nonzero chance of
> breaking something that works.

Two cases can make the honest deliverable a sentence rather than a diff. Say
which one applies and stop only when the requested refactor has no concrete
readability cost or future change to repay it.

| Case | What to do instead |
|---|---|
| The code works and nothing planned will touch it again | Name it and stop; a stable, rarely-read package earns nothing from being restructured for its own sake |
| No purpose behind "refactor this" — no upcoming feature, no bug class, no smell a review flagged | Name the purpose you inferred and refactor to *that*; if there is none, say so in the report |

Two other cases change the sequence, not cancel the work:

- **Critical path with no tests** — add characterization coverage first
  ([SAFETY-NET.md](references/SAFETY-NET.md)); continue once the net can support
  the behavior promise, otherwise report exactly what remains unproved.
- **Minimal change under time pressure** — make the minimal safe change the
  user requested and propose the larger refactor separately.

None of these is a licence to skip authorized work that has a concrete purpose.
Deliver the refactor asked for at the scope intended: "while I'm here" fixes and
a modernization that quietly becomes a migration dilute the guarantee — every
line the refactor touches takes the current form
([Write Current Go](../go-style-core/SKILL.md#write-current-go)); the untouched
rest is proposed, not rewritten. A structural problem a readability pass
cannot solve gets one sentence in the report.

## Risk Tiers

The tier sets what has to be true *before* the step, and pairs with the coverage
tiers in [SAFETY-NET.md](references/SAFETY-NET.md): low coverage on the blast
radius pushes every transform up a tier except those its Low / zero row allows.

| Tier | Transforms | Required before the step |
|---|---|---|
| **Low** | `go fix` inline of a `//go:fix inline` function; rename of an unexported symbol with no reflection or string-based contract; extract variable or constant; `gofmt -s`; organize imports; guard clauses | Baseline plus the focused check after it |
| **Medium** | Extract function or method, inline across packages, adding or removing one parameter, introducing generics, a bulk rewrite ([MECHANICAL.md](references/MECHANICAL.md)) | Tests that provably reach the touched lines |
| **High** | Signature change across many callers, cross-package moves, package split or merge, breaking an import cycle, a module or layer boundary move (`references/ARCHITECTURE.md`), any exported API change | Full net, and it is a findings-list item unless the user asked for it |

`go fix` inline keeps behavior or leaves the call alone. A rename that
compiles is not proven safe: rename may introduce dynamic errors through
reflection, templates, serialization conventions, or indirect interface
assertions.

## What "Identical Behavior" Means

> **Normative**: Anything an outside observer could notice must not move.

Exported names and signatures; struct tags and serialization order; error
text, `%w` chains, sentinels, exit and HTTP codes; the order and count of side
effects (I/O, logs, queries, locks, sends, `defer`); concurrency shape; numeric
types and overflow points; nil versus empty collections on the wire.
[BEHAVIOR-TRAPS.md](references/BEHAVIOR-TRAPS.md) has the mechanism behind each.
Internal names, function boundaries, control-flow shape, and comments are fair
game — that is where the readability gain lives.

## Delete Before You Restructure

> **Normative**: Line count is the instrument; readability is the goal. The win
> comes from code that stops existing, not code that gets rearranged.

Work in that order — delete, then shorten, then restructure — and read the net
line count after each step. Growth is a new declaration, layer, indirection,
file, or dependency, and it needs a reason in the report; a guard clause or a
named constant is a name, not growth.

A helper the refactor adds meets one of the four
[Declaration Budget](../go-code/SKILL.md#declaration-budget) rules — two call
sites in the final code, a caller outside the function that names it, a
distinct algorithm, or a step at another level of abstraction than its
caller — or it is not added. A few-line `writeHeader`, `writeRow`, and
`writeTotal` that `Render` calls once each are not added: they only rename its
steps at its own level. The report names the rule each kept helper meets.
Count the helper and its call sites when comparing complexity.

Before writing any new line — helper, wrapper, interface — climb the restraint
ladder in `references/OVER-ENGINEERING.md` and stop at the first rung that
holds; on a refactor the top one usually does.

Delete only what is **provably** unreachable — "looks unused" is a finding, not
a licence. Apply the shorter form only where it reads as well; never golf.

**Never simplify away** input validation at trust boundaries, error handling
that prevents data loss, or security checks. A "simplification" that drops a
bounds check is a bug, not laziness. These stay even when the diff gets uglier.

## Remove Duplication to the End

When branches repeat one policy or computation and differ only in its values,
separate **selection** (which case applies) from **computation** (what is done
with the selected values). Within that shared operation, aim to represent each
policy fact, selection, and condition ladder once. Replacing literals with
constants or changing `if` to `switch` alone does not remove repeated logic.

Equal literals and matching switch keys do not establish a shared policy: a
retry limit of 3 and a grace period of 3 days stay independent facts. Combine
values only when they belong to one record and the final code reads better;
stop when another fold adds more coupling than it removes, and name a material
remaining duplication in the report.

Pick the shape by the final code, call sites included: an exported accessor
that already performs the selection, before a new unexported helper; a
`switch` when the cases carry logic; a `map` or slice literal indexed by the
key when they carry only values. A table exists to delete the branches, not
to be serviced — no search helper, no method, no loop to rebuild a list that
was already a literal. If the lookup needs those, the `switch` was shorter.
Map iteration order is not source order, so an ordered literal stays a
literal. Error texts and the point where an unknown key fails do not move.
`references/POLICY-TABLES.md` shows the fold.

## Architecture at Package Scale

> **Normative**: Modules own the tree; layers live inside a module when they
> earn a package. `handlers/`, `services/`, `repositories/`, `models/` may be
> the tree of one cohesive service; several independently changing domains
> move the same names under `internal/<module>/`. Services and models import
> no transport or database driver; cross-module code stops at a module's root
> contract; keeping the tree and repairing one boundary is a valid result.

When the smell is the import graph, `references/ARCHITECTURE.md` owns the
call: measure the graph, name the shape, propose the smallest repair — a
target plus a staged plan, applied only as far as authorized. A move is High
tier and a proposal unless the user asked for it; its first commit
encodes the rule in `architecture.json` and runs `scripts/check-architecture.sh`,
whose `known` list a refactor never extends to make its own run pass.

## Concision Gate

Keep a transformation only when behavior is preserved and a reader can trace
decisions, errors, and side effects from the entry point without chasing
trivial wrappers. Measure both production LOC counts, physical and code (the
script's `--help` defines them); growth in either needs the reason
[Delete Before You Restructure](#delete-before-you-restructure) requires.
Record both starting counts before the first edit, and compare after `gofmt` on
the same path, with the counter this skill ships: `loc-baseline` in step 1 of
the [Workflow](#workflow), `loc-diff` in step 5.

`loc-diff` exits 0 when neither count grew, 1 when one did, and 2 when the
recorded path differs. Exit 1 is a signal, not a verdict: name the declaration,
layer, or dependency that grew and why; unjustified growth is a finding. The
report carries only the two counts `loc-diff` printed, physical and code; when
it could not run, that sentence stands where the numbers would. The `baseline`,
`after` and `diff` modes compare check records and say nothing about size.

Keep documentation that explains a decision or contract. Removing comments
or blank lines cannot compensate for added code; do not compress statements
onto one line to meet the limit. An empty diff is successful when no
qualifying improvement exists.

## Workflow

### 1. Orient

Scripts run from the target project by full path: `<installed-skill-dir>`
below is the base directory the host printed when it loaded this skill
(`~/.agents/skills/go-code-refactor` under Codex). Write the path itself into
each command — every shell call starts with a fresh environment, so a
variable exported in one call is empty in the next. Exit 127 from the first
script means the path is wrong.

With no shell tool, run nothing and read no script: the
[edit hook's output](../go-style-core/SKILL.md#the-edit-hook-record) is the
check record.

Load the skills before the first edit (a read of `../<name>/SKILL.md` where
there is no `Skill` tool): [go-style-core](../go-style-core/SKILL.md) on every
refactor, then the owner of each decision the diff moves, from
[Related Skills](#related-skills), go-code's
[Route Before The First Edit](../go-code/SKILL.md#route-before-the-first-edit),
or the host's routing note when it names them. No edit before every selected
skill is in context.

Flag two file classes before editing: **generated** files (exclude silently
when incidental; a generated target changes through its generator, below) and **build-tagged** files for
another GOOS/GOARCH, which never compile here — run
`GOOS=<target> go build ./...` and say in the report that their tests did not run.

Record the baseline in one shell call; the second line runs even when a red
baseline exits 1:

```bash
bash "<installed-skill-dir>/scripts/verify-refactor.sh" baseline ./...
bash "<installed-skill-dir>/scripts/verify-refactor.sh" loc-baseline ./...
```

If characterization tests are needed, run them against unchanged production
code, then capture a new baseline with those tests included. Retain the earlier
results for known failures. Use existing authorization and continue work that
does not depend on an answer:

- **If the baseline is red** — record the failing command and isolate
  pre-existing or environmental failures. Continue inspection and independently
  verifiable changes; do not claim behavior preservation without evidence.
- **If the package has zero tests** — add a small characterization test when
  the authorized refactor needs it; [SAFETY-NET.md](references/SAFETY-NET.md)
  sizes it. Missing tests alone do not require another approval. Honor an
  explicit prohibition on new tests and report the limitation.
- **If the target is generated** — trace its generator and source inputs;
  update those and regenerate when within scope. Do not hand-edit generated
  output; ask only if the real source cannot be determined.
- **If two readings change the contract or scope** — ask a focused question
  only if the request and repository do not resolve it; continue independent
  work meanwhile.

### 2. Audit before rewriting

Name what makes the code hard to follow *before* proposing fixes. The naming is
what produces a real transformation; jumping to edits produces cosmetic churn —
renamed variables, shuffled lines, same confusion. Record location, what is
hard to read, and the intended transformation.

When the ask is a cut list rather than a rewrite — "what can we delete", a repo
handed over as bloated — the audit *is* the deliverable: use the tags and
ranked format in `references/OVER-ENGINEERING.md` and stop there.

You will notice actual bugs while auditing — races, ignored errors, leaks,
off-by-ones. **Record each one in the findings list and leave the code as it
is**: a diff that mixes "reads better" with "behaves differently" cannot be
reviewed as a refactor. Report every severity; the user triages faster than
you can filter.

### 3. Scope mechanical modernization

When modernization helps the requested refactor, preview it for the affected
packages, for example `go fix -diff ./internal/cache`. Use `./...` only for a
module-wide scope. Apply `go fix` only when every proposed edit is in scope;
for a function-only task, apply relevant hunks manually if the preview also
changes neighboring functions. A neighbor's older form is not a reason to keep
it in the touched lines, and an untouched neighbor is not a reason to widen
the scope. A broader verification gate does not widen the edit scope. If no
relevant modernization helps, proceed to the hand edits.

Keep mechanical edits attributable and respect `go.mod` and the risk tiers in
`references/MODERNIZATION.md`; Tier 3 items go in the findings list.

### 4. Refactor in verifiable steps

Use small, attributable transformations with focused checks between meaningful
steps. For dependent packages, work in dependency order; shared helpers go first.
Reuse unchanged passing results and finish with the required repository gate.
For independent packages, follow the host's delegation policy and
[go-style-core](../go-style-core/SKILL.md#how-much-to-say).

`references/PLAYBOOK.md` has the transformations, ordered by payoff. Before a
rename or an extraction, search its uses by text: the build catches the static
ones, not reflection, templates, string lookups, or a type matched only
through an assertion or a type switch.

Three cases leave the playbook. A move that crosses a function, type, or package
boundary is in `references/CATALOG.md`, with its tool and tier. The same edit
recurring across many sites is a generated rewrite, not thirty hand-edits
(`references/MECHANICAL.md`). A type changing packages is a staged alias
migration, never one atomic commit (`references/STRUCTURAL.md`).

### 5. Verify

In one shell call:

```bash
bash "<installed-skill-dir>/scripts/verify-refactor.sh" after ./...
bash "<installed-skill-dir>/scripts/verify-refactor.sh" diff
bash "<installed-skill-dir>/scripts/verify-refactor.sh" loc-diff ./...
```

The diff compares recorded check results, not program behavior. An empty diff
can include the same failures or skipped checks; a nonempty diff can include
new passing tests or resolved failures. Inspect each difference and the actual
baseline/after statuses. Attribute new failures before changing code; undo only
your own failing transformation, preserving pre-existing user changes.

Keep assertions about observable behavior unchanged. Mechanical updates to
references after an authorized rename and new characterization tests are
allowed; weakening expectations to make the refactor pass is not. If a
toolchain change was authorized, `references/MODERNIZATION.md` lists failures
that may need attribution to that change. Passing checks support only the
behavior they exercise; report what remains unverified.

Watch tests that assert on error strings or JSON output — they catch the
invisible breakages compilation misses.
[go-linting](../go-linting/SKILL.md) owns what the individual checks mean.

### 6. Report

Use `assets/refactor-report.md`; report skipped checks as skipped.

## Mark What You Deliberately Left Alone

When you keep something ugly because changing it would change behavior, say so
in the code, not only in the report. `Kept:` / `Ceiling:` / `Fix:` are fixed
prefixes, so the markers stay greppable:

```go
// Kept: defer stays inside the loop; hoisting it closes files one iteration earlier.
// Ceiling: descriptors accumulate for the worker's lifetime.
// Fix: close explicitly per iteration, in its own commit.
```

A marker naming no ceiling and no upgrade path rots into "later means never".
`bash "<installed-skill-dir>/scripts/check-debt.sh" ./...` lists every marker and
exits 1 on those.

## Related Skills

- [go-style-core](../go-style-core/SKILL.md): nesting and the clarity > simplicity > concision order.
- [go-naming](../go-naming/SKILL.md): renames.
- [go-error-handling](../go-error-handling/SKILL.md): wrapping, sentinels, handle-once rewrites.
- [go-concurrency](../go-concurrency/SKILL.md): goroutine lifetimes, channels, locks in the diff.
- [go-packages](../go-packages/SKILL.md): a refactor that crosses package boundaries; where a new package goes once `references/ARCHITECTURE.md` has named the target.
- [go-code-review](../go-code-review/SKILL.md): the finished diff against the checklist.
