# Skill Authoring Template

Use this structure for every Go skill unless a section is genuinely empty.
Keep core `SKILL.md` files decision-oriented; move long examples, tables, and
edge-case catalogs into `references/`, reusable commands into `scripts/`, and
copyable files into `assets/`.

Before duplicating a rule in more than one skill, check
[`RULE_OWNERSHIP.md`](RULE_OWNERSHIP.md) and route non-owner skills to the
canonical owner.

## Frontmatter

Use trigger-critical fields plus runtime-required tool grants:

```yaml
---
name: go-example
description: Use when ...
allowed-tools: Bash(bash:*) # only when the skill bundles scripts that need it
---
```

Keep license, compatibility, provenance, source authority, and validation
policy in the body or references. Keep `allowed-tools` in frontmatter for
script-backed skills when the target runtime uses it to pre-authorize bundled
commands.

## Body Structure

```md
# Skill Title

One short paragraph stating the core job.

## Resource Routing

- `references/FILE.md` - Read when ...
- `scripts/tool.sh` - Run when ...
- `assets/file.ext` - Use when ...

## Domain-Specific Guidance

Use headings that match the skill's domain. Keep them decision-oriented:
short rules, routing trees, checklists, and small examples.

## Related Skills

- `go-other-skill` - Route when ...
```

## Writing for a Capable Agent

The reader already knows Go. Optimize for the decisions it gets wrong, not for
the syntax it gets right.

- **Cut what the model already knows.** No "goroutines are lightweight
  threads". Every paragraph should change an output.
- **Make decisions explicit.** State the outcome and conditions for a rule.
  Reserve absolute language for correctness constraints; identify defaults that
  yield to the user's request, repository conventions, or supported toolchain.
- **Make claims checkable.** Attach the command that proves it — `go vet ./...`,
  `go fix -diff ./...`, `golangci-lint run` — instead of asserting a rule the
  reader must take on faith. The verification gate lives in `go-linting`; route
  to it rather than restating it.
- **Say what "done" means.** A skill that produces work should end with the
  check that fails when the work is wrong.
- **Calibrate verification.** State required evidence through `go-linting`.
  Do not assume any model has checked its work implicitly. Reuse observed
  passing results for unchanged code; repeat or expand checks only when an edit,
  failure, unresolved concern, or repository requirement justifies it.
- **Preserve the requested result.** Simplicity guides implementation, not
  which explicit requirements get delivered. Missing tests or a routine choice
  do not automatically require renewed approval; continue authorized work.
- **Keep runtime assumptions local.** Resolve scripts from the installed skill
  directory, execute against the target project, and handle absent siblings.
  Claude hooks, `Skill`, and agent model names are not portable prerequisites.
- **Say how much to say once.** Narration cadence, report length, and
  delegation live in `go-style-core` "How Much To Say"; a procedural skill
  routes there instead of restating it.
- **Respect report scope.** Give evidence and severity for findings at the
  depth requested; a skill default must not override an explicit severity filter.
- **Require honest reporting.** Where a skill tells the agent to run something,
  it also says: report a skipped or failing step as skipped or failing.
- **Name the Go version inline** for anything newer than 1.21, and route the
  skill's `> Compatibility:` note to `COMPATIBILITY.md`.

These defaults incorporate [OpenAI's GPT-6 Astra prompting guidance](https://developers.openai.com/api/docs/guides/latest-model/gpt-6-astra.md#prompting-best-practices)
on instruction priority, follow-through, communication, delegation, and bounded
verification. They are shared rules for GPT-6 and Claude, not claims about either
model's automatic behavior. Test behavior in each host: a schema check alone
does not establish model quality. See [the cross-model review](CROSS_MODEL_REVIEW.md).

## Required Conformance

- Keep `SKILL.md` frontmatter to `name`, `description`, and runtime-required
  `allowed-tools` grants for script-backed skills.
- Include exactly one `## Resource Routing` section, and list every bundled
  file under `references/`, `scripts/`, and `assets`.
- Include a `## Related Skills` section for handoffs to owner skills.
- Keep core files at or below 225 lines; use references for long examples,
  edge cases, and source-sensitive details.
- Include validation guidance near the relevant command or rule when a
  deterministic check exists.

## Reference Headers

Start reference files with compact provenance when source authority matters:

```md
> Sources: source/path.md; official docs URL
> Authority: normative | advisory | historical | project policy
> Minimum Go: 1.xx, if version-sensitive
> Last verified: YYYY-MM-DD
```

Use the authority labels this way:

- `normative`: canonical Go or project-required guidance.
- `advisory`: style guidance that may vary by codebase.
- `historical`: useful context that may not reflect modern Go.
- `project policy`: this repository's chosen rule where sources differ.

## Long References

Any reference over 200 lines must include a `## Contents` section near the top.
Prefer splitting only when a TOC still leaves unrelated subtopics hard to route.
