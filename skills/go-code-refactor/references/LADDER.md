# Restraint Ladder

> Sources: github.com/DietrichGebert/ponytail; Go CodeReviewComments
> Authority: project policy for the ladder and the write rules
> Last verified: 2026-10-23 against go1.27.1

Normative whenever Go code is written, on any task. The session-start hook
prints the ladder section; the stdlib table it points to and the audit lane
live in [OVER-ENGINEERING.md](OVER-ENGINEERING.md).

## The Restraint Ladder

> **Normative**: The ladder runs *after* the problem is understood, not instead
> of it. Read the code the change touches and trace the real flow end to end
> first, then climb. It shortens the solution, never the reading — the smallest
> diff in the wrong place is a second bug, not laziness.

Before writing a new line — type, layer, interface, option, helper, import —
stop at the first rung that holds:

1. **Does this need to exist at all?** Speculative need → skip it and say so in
   one line (YAGNI). On a refactor this rung usually holds: deletion beats
   rewrite.
2. **Already in this codebase?** A helper, type, or pattern two files over →
   reuse it. Re-implementing what the repo already has is the most common slop;
   look before you write.
3. **Does the standard library ship it?** `go doc <pkg>` before assuming it is
   missing; the table in
   [Reach For What Go Ships](OVER-ENGINEERING.md#reach-for-what-go-ships) is the checklist for 3 and 4.
4. **Does a language or toolchain feature cover it?** `go:embed` over an asset
   loader, struct tags over a hand-written marshaler, a DB constraint over an
   app-side check, `t.Cleanup` over hand-rolled teardown, `testing/synctest`
   over sleep-based waits, a build tag over a runtime switch.
5. **A module already in `go.mod`?** Use it. Never add a new one for what a few
   lines do — [go-packages](../../go-packages/SKILL.md) owns the module rungs
   (stdlib → a module already in `go.mod` → a new module, `golang.org/x/...`
   included).
6. **Can it be one line?** One line.
7. **Only then**: the minimum that works, in the fewest files — a new file only
   when the existing one is unwieldy.

The ladder runs per entity, not once per task: each helper, type, and method
written to implement a needed feature gets its own climb — "the feature is
needed" does not carry over. Judge the final diff, not the opening plan.

Two rungs both work → take the higher one and move on. But if the higher rung
makes the **call site** read worse, take the lower one: the ladder minimizes
concepts, not clarity. [go-style-core](../../go-style-core/SKILL.md) owns the
order — Clarity > Simplicity > Concision — and it outranks every rung here.

Two options on the same rung, same size → take the one correct on edge cases.
Lazy means less code, not a flimsier algorithm.

**A bug fix is a root-cause fix.** A report names a symptom. Before the edit,
find every caller of the function you are about to touch and land the fix
where they all route through: one guard in the shared function is a smaller
diff than a guard in every caller, and patching only the path the ticket named
leaves the sibling callers broken.
[go-troubleshooting](../../go-troubleshooting/SKILL.md) owns the method while
the cause is still unknown.

**Never on the chopping block**: input validation at trust boundaries, error
handling that prevents data loss, security controls, accessibility basics in
anything user-facing, clarity at the call site, the tests that fail when the
logic breaks, and anything the user asked for explicitly. Lazy, not negligent
— a "simplification" that drops a bounds check is a bug.

## Ship, Then Question

Choose the smallest implementation that satisfies every explicit requirement.
Simplify implementation choices, not the requested deliverable: dry-run,
pagination, compatibility, and other requested behavior are not speculative.
Never stall on an answer you can default within that scope. Explain a routine
assumption briefly and continue; ask only when missing information changes
correctness, scope, or authorization, while doing independent work. Respect an
explicit choice on the first request — no re-arguing or requiring the user to
ask twice. `go-code-refactor` owns baseline and generated-code decisions.

For new code, lead with the code and the gate result, and name each deliberate
omission as `skipped: <X>, add when <Y>`; report length follows
[go-style-core](../../go-style-core/SKILL.md#how-much-to-say). A paragraph
defending a simplification is complexity smuggled back in as prose. A report, walkthrough, or per-step note the user
asked for is not prose debt; give it in full. A refactor reports through
`assets/refactor-report.md`, whose table is already the short form.

Lazy code without its check is unfinished. Non-trivial logic — a branch, a
loop, a parser, a money or security path — leaves a runnable check behind.
YAGNI applies to tests too:
[go-testing](../../go-testing/SKILL.md#what-earns-a-test) owns which tests
earn their place and their form. Tests that fail when the logic breaks are
never a cut.
