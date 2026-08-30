---
name: go-code-refactor
description: Use when refactoring, cleaning up, simplifying, restructuring, or modernizing existing Go code while keeping observable behavior identical — reducing nesting, splitting long functions, deleting dead code, renaming for clarity, or adopting newer Go APIs. Also use when a user hands over a Go file or package and calls it messy, hard to follow, too long, bloated, over-engineered, or outdated, even if they never say "refactor". Does not cover writing new Go code or the style rules themselves (see go-style-core and the rule owners).
allowed-tools: Bash(bash:*)
---

# Go Refactoring

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). Modernization
> targets the `go` directive in `go.mod`, not the installed toolchain.

Legacy Go usually fails on readability, not correctness. The entire value of
this work is one promise: **the code reads better and does exactly the same
thing.** Break it once and reviewers stop trusting the diff and re-read every
line — more expensive than the mess you started with. So behavior preservation
is the definition of the task, not a quality bar to aim at.

## Resource Routing

- `references/BEHAVIOR-TRAPS.md` - Read before touching concurrency, `defer`, error handling, slices, interfaces, or struct layout.
- `references/PLAYBOOK.md` - Read for the concrete transformations, ordered by payoff, with before/after Go.
- `references/MODERNIZATION.md` - Read before adopting a newer API; sorts Go 1.21–1.27 features into safe, conditional, and report-only.
- `references/OVER-ENGINEERING.md` - Read when the ask is "what can we delete": cut tags, the Go hunt list, and the ranked audit format.
- `scripts/verify-refactor.sh` - Run to capture a baseline, re-check after each step, and diff the two.
- `scripts/check-debt.sh` - Run to harvest `Kept:` markers into a ledger and flag the ones naming no upgrade path.
- `assets/refactor-report.md` - Use as the final report structure.

## Stop and Ask

Four situations end the turn with a question instead of an edit. Everything
else is a judgment call you make yourself.

1. **The baseline is red** — build broken or tests failing before you touched
   anything. You cannot sign a promise you cannot verify.
2. **The package has zero tests** — behavior preservation is an empirical
   claim with nothing to check it against. Offer one characterization test
   (no framework, no fixtures, smallest thing that fails if the logic breaks)
   and wait; the extra file is the user's call.
3. **The target is generated** — `// Code generated ... DO NOT EDIT.` reverts
   at the next generation. Point at the generator and ask what the real target is.
4. **Two readings lead to materially different work** — e.g. "clean up the
   package" where half the mess is an exported API you may not rename.

---

## What "Identical Behavior" Means

> **Normative**: Anything an outside observer could notice must not move.

- Exported signatures and names, struct tags, field order where it affects
  serialization, `unsafe`, or binary layout
- Error values and their **text**, `%w` chains, sentinels, exit codes, HTTP codes
- The sequence and count of side effects: I/O, logs, queries, lock acquisition
  order, channel sends, `defer` firing order
- Concurrency shape: goroutine counts, channel buffers, timeouts, ctx propagation
- Numeric types, overflow points, float associativity
- `nil` slice/map vs empty — `encoding/json` renders these differently, so
  "normalizing" one into the other is a wire-format change

Internal names, function boundaries, control-flow shape, and comments are fair
game. That is where the readability gain lives.

---

## Delete Before You Restructure

> **Normative**: Line count is the instrument; readability is the goal. The win
> comes from code that stops existing, not code that gets rearranged.

Before writing any new line — helper, wrapper, interface — stop at the first
rung that holds:

1. Does this need to exist at all? Deletion beats rewrite.
2. Does the standard library already do it? (`slices.Contains`, `strings.Builder`,
   `errors.Join` — see [go-data-structures](../go-data-structures/SKILL.md))
3. Does the repo already have it? A helper two files over beats a second one.
4. Can it be one line?
5. Only then, the minimum that works.

Delete only what is **provably** unreachable — "looks unused" is a finding, not
a licence. Apply the shorter form only where it reads as well; never golf.

**Never simplify away** input validation at trust boundaries, error handling
that prevents data loss, or security checks. A "simplification" that drops a
bounds check is a bug, not laziness. These stay even when the diff gets uglier.

---

## Workflow

### 1. Orient

Read enough of the package to know what idiomatic means *here*. A refactor that
imports a house style the repo does not use creates a second dialect — worse
than the original mess. A local `.golangci.yml` or `CONTRIBUTING.md` outranks
any external guide.

Flag two file classes before editing: **generated** files (exclude silently
when incidental, ask when they are the target) and **build-tagged** files for
another GOOS/GOARCH, which never compile here — run
`GOOS=<target> go build ./...` and say in the report that their tests did not run.

Read the `go` directive in `go.mod`; it gates which modernization is legal. If
it lags the toolchain, mention the gap once — bumping it is the user's call and
carries its own behavior changes.

```bash
bash scripts/verify-refactor.sh baseline ./...
```

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

### 3. Let `go fix` do the mechanical work first

```bash
go fix -diff ./...   # read it
go fix ./...         # then apply
```

Running this first keeps mechanical changes attributable to the tool, so
everything left is a judgment call you have to justify. Read the diff rather
than trusting it. Then apply the Tier 1 hand edits from
`references/MODERNIZATION.md`; Tier 3 items go in the findings list.

### 4. Refactor in verifiable steps

One named transformation at a time, verified before the next begins. For
multi-package work: **one package per cycle** — audit, edit, verify. A
repo-wide diff produced in one sitting cannot be bisected and will not be
trusted. Shared helpers get their own cycle first.

`references/PLAYBOOK.md` has the transformations. The high-value ones: delete
dead code, extract until each function has one job, flatten with early returns,
name things after what they mean, name magic values, remove duplication that
has a name.

Avoid the failure mode where "cleaner" means "more layers". New interfaces,
indirection, or generics the call sites do not need make code harder to follow.
Add abstraction only when it removes more complexity than it introduces.

### 5. Verify

```bash
bash scripts/verify-refactor.sh after ./...
bash scripts/verify-refactor.sh diff
```

The diff must be empty. If a test fails, the refactor is wrong — revert that
step and redo it smaller. **Never adjust a test to match the new code**: a test
that had to change is proof that behavior changed.

One exception: if the toolchain was bumped as part of this work, the failure
may belong to Go rather than to you. `references/MODERNIZATION.md` lists the
releases that break green tests on their own. Attribute before rewriting.

Watch tests that assert on error strings or JSON output — they catch the
invisible breakages compilation misses.
[go-linting](../go-linting/SKILL.md) owns what the individual checks mean.

### 6. Report

Use `assets/refactor-report.md`. Lead with what was deleted — it is the part of
the diff that needed no design decision. Keep prose short and report skipped
checks as skipped; the table and the diff carry the information.

---

## Mark What You Deliberately Left Alone

When you keep something ugly because changing it would change behavior, say so
in the code, not only in the report. `Kept:` / `Ceiling:` / `Fix:` are fixed
prefixes, so the markers stay greppable:

```go
// Kept: defer stays inside the loop. Hoisting it into a helper would close
// files one iteration earlier, which is observable.
// Ceiling: descriptors accumulate for the worker's lifetime.
// Fix: close explicitly per iteration, in its own commit.
defer f.Close()
```

A marker naming no ceiling and no upgrade path rots into "later means never".
`bash scripts/check-debt.sh ./...` lists every marker and exits 1 on those.

---

## Scope

Deliver the refactor asked for, at the scope intended. Refactoring invites
drift — adding logging, metrics, error handling for cases that never existed,
"while I'm here" fixes. Each dilutes the guarantee. Modernization has its own
version: a new stdlib package appears and the diff quietly becomes a migration.
Adopt what makes the existing code read better; propose the rest.

If the code has a structural problem a readability pass cannot solve — the
wrong abstraction, a data model that forces the mess — say so in one sentence,
finish the refactor as asked, and let the user decide about the bigger change.

---

## Related Skills

- **Style rules being applied**: See [go-style-core](../go-style-core/SKILL.md) when deciding nesting, naked returns, or the clarity > simplicity > concision order
- **Verification gate**: See [go-linting](../go-linting/SKILL.md) for what `vet`, `go fix`, `-race`, and `golangci-lint` each prove
- **Renaming**: See [go-naming](../go-naming/SKILL.md) when improving identifier, receiver, or package names
- **Error-flow rewrites**: See [go-error-handling](../go-error-handling/SKILL.md) when restructuring wrapping, sentinels, or the handle-once pattern
- **Concurrency rewrites**: See [go-concurrency](../go-concurrency/SKILL.md) when goroutine lifetimes, channels, or locks are in the diff
- **Splitting packages**: See [go-packages](../go-packages/SKILL.md) when the refactor crosses package boundaries
- **Reviewing the result**: See [go-code-review](../go-code-review/SKILL.md) when checking the finished diff against the full checklist
- **Writing new code instead**: See [go-functions](../go-functions/SKILL.md) and [go-interfaces](../go-interfaces/SKILL.md) when the task is designing something new rather than reshaping what exists
