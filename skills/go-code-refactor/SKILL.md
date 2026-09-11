---
name: go-code-refactor
description: Use when refactoring, cleaning up, simplifying, restructuring, or modernizing existing Go code while keeping observable behavior identical — reducing nesting, splitting long functions, deleting dead code, renaming for clarity, or adopting newer Go APIs. Also use when a user hands over a Go file or package and calls it messy, hard to follow, too long, bloated, over-engineered, or outdated, even if they never say "refactor". Does not cover writing new Go code or the style rules themselves (see go-style-core and the rule owners).
allowed-tools: Bash(bash:*)
---

# Go Refactoring

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). Modernization
> targets the `go` directive in `go.mod`, not the installed toolchain.

Improve readability while preserving observable behavior. Establish evidence
for that promise; compilation alone does not establish equivalent behavior.

## Resource Routing

Resolve resources from this installed skill directory; run scripts from the
target project using the resolved absolute script path.

- `references/BEHAVIOR-TRAPS.md` - Read its Pre-commit checklist before every refactor; read a section when a transform moves a `defer`, changes nil versus empty, alters goroutine or channel shape, or touches struct layout.
- `references/PLAYBOOK.md` - Read for the concrete transformations, ordered by payoff, with before/after Go.
- `references/POLICY-TABLES.md` - Read when repeated selection accesses fields of one shared policy record; includes a complete before/after example.
- `references/CATALOG.md` - Read when the move crosses a function, type, or package boundary: the smell that triggers each transform, the tool that performs it, and its risk tier.
- `references/SAFETY-NET.md` - Read when the blast radius has thin or no tests: coverage tiers, characterization tests, and seams for untested code.
- `references/MECHANICAL.md` - Read when the same edit recurs across many sites: `gofmt -r`, `eg`, `gopatch`, and `go/analysis` fixers instead of hand-editing each one.
- `references/STRUCTURAL.md` - Read before moving a type between packages, breaking an import cycle, or changing an exported API: type-alias gradual repair and the deprecation sequence.
- `references/MODERNIZATION.md` - Read when a hunk adopts a newer API or `go fix -diff` proposes one; sorts Go 1.21–1.27 features into safe, conditional, and report-only.
- `references/OVER-ENGINEERING.md` - Read when a step adds a helper, type, layer, option, or import (it owns the restraint ladder, the reach-for table, and the ship-then-question write rules), and when the ask is "what can we delete": cut tags, the Go hunt list, and the ranked audit format.
- `references/GOPLS.md` - Read before renaming, extracting, or inlining anything with more than one caller: semantic references and safe rename via gopls instead of grep.
- `scripts/verify-refactor.sh` - Run to capture baseline and final check results, and to count production LOC before and after; use focused checks between edits.
- `scripts/check-debt.sh` - Run to harvest `Kept:` markers into a ledger and flag the ones naming no upgrade path.
- `assets/refactor-report.md` - Use as the final report structure.

Every command below runs the scripts through `REFACTOR_SKILL_DIR`. Set it once,
before the first one, to the base directory the host printed for this skill
(`${CLAUDE_PLUGIN_ROOT}/skills/go-code-refactor` under the Claude Code plugin,
`~/.agents/skills/go-code-refactor` under Codex), and keep the working directory
in the target project. Unset, the path collapses to `/scripts/verify-refactor.sh`
and every call exits 127 (`No such file or directory`) — with no baseline
recorded, the gate below cannot pass.

```bash
export REFACTOR_SKILL_DIR="<base directory the host printed for this skill>"
bash "$REFACTOR_SKILL_DIR/scripts/verify-refactor.sh" --version    # must print a version; 127 means the path is wrong
```

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
a modernization that quietly becomes a migration dilute the guarantee — adopt
what makes the existing code read better, propose the rest. A structural
problem a readability pass cannot solve gets one sentence in the report.

---

## Risk Tiers

The tier sets what has to be true *before* the step, and pairs with the coverage
tiers in [SAFETY-NET.md](references/SAFETY-NET.md): low coverage on the blast
radius pushes every transform up a tier.

| Tier | Transforms | Required before the step |
|---|---|---|
| **Low** | gopls inline; rename of an unexported symbol with no reflection or string-based contract; extract variable or constant; `gofmt -s`; organize imports; guard clauses | Baseline plus the focused check after it |
| **Medium** | Extract function or method, inline across packages, adding or removing one parameter, introducing generics, a bulk rewrite ([MECHANICAL.md](references/MECHANICAL.md)) | Tests that provably reach the touched lines |
| **High** | Signature change across many callers, cross-package moves, package split or merge, breaking an import cycle, any exported API change | Full net, and it is a findings-list item unless the user asked for it |

gopls inline preserves behavior or refuses. Rename is compilation-aware, but
rename may introduce dynamic errors through reflection, templates,
serialization conventions, or indirect interface assertions; extract may drop
comments. A refusal is a semantic hazard — investigate it, never hand-edit
around it ([GOPLS.md](references/GOPLS.md#gotchas)).

---

## What "Identical Behavior" Means

> **Normative**: Anything an outside observer could notice must not move.

Exported names and signatures; struct tags and serialization order; error
text, `%w` chains, sentinels, exit and HTTP codes; the order and count of side
effects (I/O, logs, queries, locks, sends, `defer`); concurrency shape; numeric
types and overflow points; nil versus empty collections on the wire.
[BEHAVIOR-TRAPS.md](references/BEHAVIOR-TRAPS.md) has the mechanism behind each.
Internal names, function boundaries, control-flow shape, and comments are fair
game — that is where the readability gain lives.

---

## Delete Before You Restructure

> **Normative**: Line count is the instrument; readability is the goal. The win
> comes from code that stops existing, not code that gets rearranged.

Work in that order — delete, then shorten, then restructure — and read the net
line count after each step. Growth is a new declaration, layer, indirection,
file, or dependency, and it needs a reason in the report; a guard clause or a
named constant is a name, not growth.

Extract a helper when it removes repeated logic or hides a meaningful operation.
Keep a short, single-use sequence inline when the helper only renames its steps.
Count the helper and its call sites when comparing complexity.

Before writing any new line — helper, wrapper, interface — climb the restraint
ladder in `references/OVER-ENGINEERING.md` and stop at the first rung that
holds. On a refactor the top rung usually holds: deletion beats rewrite, and
rung 2 (the helper two files over) beats a second copy of it. Climb only after
reading the code the change touches — the ladder shortens the diff, never the
reading.

Delete only what is **provably** unreachable — "looks unused" is a finding, not
a licence. Apply the shorter form only where it reads as well; never golf.

**Never simplify away** input validation at trust boundaries, error handling
that prevents data loss, or security checks. A "simplification" that drops a
bounds check is a bug, not laziness. These stay even when the diff gets uglier.

---

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

---

## Concision Gate

Keep a transformation only when the final code is at least as clear and
behavior is preserved. Measure both production LOC counts; growth in either
needs the reason [Delete Before You Restructure](#delete-before-you-restructure)
requires:

- **Physical LOC:** every line in the scoped non-test `*.go` files, including
  blank lines and comments; include new files and account for deleted files.
- **Code LOC:** lines containing Go tokens other than comments. A line with a
  trailing comment counts once; every line of a multiline literal counts.

Record both starting counts before the first edit, and compare after `gofmt`,
with the counter this skill ships:

```bash
bash "$REFACTOR_SKILL_DIR/scripts/verify-refactor.sh" loc-baseline ./internal/gateway
bash "$REFACTOR_SKILL_DIR/scripts/verify-refactor.sh" loc-diff ./internal/gateway
```

`loc-diff` exits 0 when neither count grew and 1 when one did. Exit 1 is a
signal, not a verdict: name the declaration, layer, or dependency that grew and
why; unjustified growth is a finding. Do not estimate the numbers, do not
substitute a nonblank-line count, and do not report counts the counter did not
print. If it cannot run, report that. The `baseline`, `after` and `diff` modes
compare check records and say nothing about size.

Keep documentation that explains a decision or contract. Removing comments
or blank lines cannot compensate for added code; do not compress statements
onto one line to meet the limit. Preserve validation, failure behavior,
security controls, and useful abstraction boundaries. If a transformation
fails, revise it or undo only your own edits. An empty diff is successful when
no qualifying improvement exists. Report starting and final physical/code
counts, their deltas, and the checks supporting behavior preservation.

---

## Workflow

### 1. Orient

Before rewriting, read [go-style-core](../go-style-core/SKILL.md) for the shared
style and control-flow rules, including snippet-only refactors, and the
convention files its House Style Wins names; they take precedence over the
guide's defaults.

Flag two file classes before editing: **generated** files (exclude silently
when incidental, ask when they are the target) and **build-tagged** files for
another GOOS/GOARCH, which never compile here — run
`GOOS=<target> go build ./...` and say in the report that their tests did not run.

Read the `go` directive in `go.mod`; it gates which modernization is legal. If
it lags the toolchain, mention the gap once — bumping it is the user's call and
carries its own behavior changes.

```bash
bash "$REFACTOR_SKILL_DIR/scripts/verify-refactor.sh" baseline ./...
bash "$REFACTOR_SKILL_DIR/scripts/verify-refactor.sh" loc-baseline ./...
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
off-by-ones. **Do not fix them.** A diff that mixes "reads better" with
"behaves differently" cannot be reviewed as a refactor. Collect them and hand
them back. Report every severity; the user triages faster than you can filter.

### 3. Scope mechanical modernization

When modernization helps the requested refactor, preview it for the affected
packages, for example `go fix -diff ./internal/cache`. Use `./...` only for a
module-wide scope. Apply `go fix` only when every proposed edit is in scope;
for a function-only task, apply relevant hunks manually if the preview also
changes neighboring functions. A broader verification gate does not widen the
edit scope. If no relevant modernization helps, proceed to the hand edits.

Keep mechanical edits attributable and respect `go.mod` and the risk tiers in
`references/MODERNIZATION.md`; Tier 3 items go in the findings list.

### 4. Refactor in verifiable steps

Use small, attributable transformations with focused checks between meaningful
steps. For dependent packages, work in dependency order; shared helpers go first.
Reuse unchanged passing results and finish with the required repository gate.
For independent packages, follow the host's delegation policy and
[go-style-core](../go-style-core/SKILL.md#how-much-to-say).

`references/PLAYBOOK.md` has the transformations. The high-value ones: delete
dead code, flatten with early returns, name things after what they mean,
name magic values, remove duplication that
has a name, fold branches that differ only in values into one selection
(see [Remove Duplication to the End](#remove-duplication-to-the-end)). Renames and extractions go through gopls (`references/GOPLS.md`):
find references semantically first, inspect reflection and string-based uses,
then let rename refuse compilation hazards such as a directly observed broken
interface implementation — grep cannot see those semantic references.

Three cases leave the playbook. A move that crosses a function, type, or package
boundary is in `references/CATALOG.md`, with its tool and tier. The same edit
recurring across many sites is a generated rewrite, not thirty hand-edits
(`references/MECHANICAL.md`). A type changing packages is a staged alias
migration, never one atomic commit (`references/STRUCTURAL.md`).

### 5. Verify

```bash
bash "$REFACTOR_SKILL_DIR/scripts/verify-refactor.sh" after ./...
bash "$REFACTOR_SKILL_DIR/scripts/verify-refactor.sh" diff
bash "$REFACTOR_SKILL_DIR/scripts/verify-refactor.sh" loc-diff ./...
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

Use `assets/refactor-report.md`. Lead with what was deleted and the net line
count — the part of the diff that needed no design decision. Keep prose short
([go-style-core](../go-style-core/SKILL.md#how-much-to-say) owns the length);
report skipped checks as skipped; the table and the diff carry the information.

---

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
`bash "$REFACTOR_SKILL_DIR/scripts/check-debt.sh" ./...` lists every marker and
exits 1 on those.

---

## Related Skills

- [go-style-core](../go-style-core/SKILL.md): nesting and the clarity > simplicity > concision order.
- [go-naming](../go-naming/SKILL.md): renames.
- [go-error-handling](../go-error-handling/SKILL.md): wrapping, sentinels, handle-once rewrites.
- [go-concurrency](../go-concurrency/SKILL.md): goroutine lifetimes, channels, locks in the diff.
- [go-packages](../go-packages/SKILL.md): a refactor that crosses package boundaries.
- [go-code-review](../go-code-review/SKILL.md): the finished diff against the checklist.
