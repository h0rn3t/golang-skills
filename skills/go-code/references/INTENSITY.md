# Intensity

> Sources: project policy ([go-code](../SKILL.md#workflow) step 1); ponytail (DietrichGebert) `Intensity`
> Authority: normative — how hard the restraint ladder pushes at each level
> Last verified: 2026-09-30

A level word right after the command — `/go-code lite <task>`,
`/go-code ultra <task>` — or `lite mode` / `ultra mode` in the prompt sets how
hard the [restraint ladder](../../go-code-refactor/references/OVER-ENGINEERING.md#the-restraint-ladder)
pushes. The word is never part of the task. The level holds for the rest of
the session, until another level word; `full` is the default and returns to
it. A behavior-preserving refactor runs at `full` whatever the word: the
[delete-first order](../../go-code-refactor/SKILL.md#delete-before-you-restructure)
already fixes its shape.

| Level | What changes |
|---|---|
| `lite` | Rungs 1–6 advise on what the request implies: build the shape the request suggests, and where a higher rung would hold, name it in one report line — `lazier: <X>` — for the user to pick. |
| `full` | The ladder and the [Declaration Budget](../SKILL.md#declaration-budget) as written. The default. |
| `ultra` | Rung 1 for every part the request does not state in words — an option, a config field, a hook, an export, a cache, a goroutine: skip it and report `skipped: <X>, add when <Y>`. Before adding a line, delete what the change leaves dead. A stated requirement a higher rung would cover still ships, with one line: `Need <X>? <Y> covers it.` |

No level changes the gate, the [Contract Table](../SKILL.md#contract-table),
an explicit requirement, or what the ladder never cuts: validation at trust
boundaries, error handling that prevents data loss, and security controls.
