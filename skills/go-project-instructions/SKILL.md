---
name: go-project-instructions
description: Use when asked to audit, create, update, or reconcile AGENTS.md or CLAUDE.md for a Go repository, including stale build or test commands, changed module or workspace layout, and conflicting agent instructions. Ordinary Go work that only reads those files belongs to the skill for that work.
---

# Go Project Instructions

Keep the instructions agents load tied to durable project rules and the
commands developers and CI actually use.

## Resource Routing

- [Codex project instructions](https://learn.chatgpt.com/docs/agent-configuration/agents-md) for directory scope, overrides, and instruction discovery.
- [Claude Code project instructions](https://code.claude.com/docs/en/memory) for discovery, `@path` imports, and scoped rules.

## Workflow

1. **Set the scope and mode.** Start at the target repository root or the
   directory the user names. Audit, review, and dry-run requests are
   read-only. Update or create requests authorize the requested files and
   scopes. When updating both guides, use their existing locations and
   create a missing counterpart where the host can discover it. When
   updating one guide, read the other for context and report conflicts
   outside scope.
2. **Find the instructions in effect.** Read existing root guides and the
   nested guides for the requested modules or directories. Check for
   `AGENTS.override.md`, `.claude/CLAUDE.md`, path-scoped rules, local
   overrides, and imports where they affect that scope. Resolve relative
   imports from the file containing them. Preserve existing import and
   symlink arrangements. Before editing through a shared arrangement,
   check that its target and affected guides are within the authorized
   scope. Inspect local changes before applying a patch.
3. **Trace each checkable claim.** Inventory the relevant `go.mod` files and
   tracked `go.work` membership. Record each module's directory and `go`
   directive separately from a `toolchain` directive and CI's tested Go
   versions. Inspect workflow steps, their working directories, and the
   Makefile, Taskfile, or scripts those steps call. Capture actual build and
   test commands, tags, service prerequisites, integration opt-ins, and
   linter or generator configuration where relevant. Verify architecture
   and package paths against the source tree and durable design docs.
4. **Resolve drift with evidence.** For each stale claim, missing instruction,
   or conflict, record the guide path and the supporting repository path
   and line or target. Use executable configuration for factual claims
   about how the project builds and tests. Preserve explicit project policy
   unless the user authorizes changing it. If policy and CI disagree or the
   evidence is incomplete, report the gap and keep the disputed policy
   intact. Keep mandatory checks, optional checks, and CI-only steps distinct.
   Also record a rule stated in two guides (root and nested, or `AGENTS.md`
   and `CLAUDE.md`) as drift even when the copies agree, and record a guide
   that leaves the agent no command to check its own change. A guide whose
   `git log` stops at its creation commit while `go.mod`, CI, or the build
   recipes changed after it is stale until each claim is re-traced.
5. **Apply the smallest authorized edit.** Keep the existing language,
   structure, and project-specific policies. Put shared rules at their
   existing source of truth and module-specific commands at the scope they
   serve. Keep host-specific guidance and applicable nested exceptions.
   When creating both root files, write shared rules in `AGENTS.md` and
   import it with `@AGENTS.md` in `CLAUDE.md`. A `CLAUDE.md` that only tells
   the agent in words to read `AGENTS.md` gets the import instead: prose
   loads the file only if the agent decides to open it. For a new guide, include the
   project layout, commands with working directories and prerequisites, and
   durable local constraints that an agent needs. Keep secrets, local
   machine paths, and temporary observations out of persistent instructions.
   Hold every guide to these rules:
   - **Always-loaded text is a budget.** The root guide and its imports
     carry only rules that apply to every task. A procedure one workflow
     needs (release, migrations, profiling) goes in a maintained doc,
     nested guide, or path-scoped rule that loads with that work, linked in
     one line from the root that says what the target holds and when to
     read it. A guide over 200 lines is a finding to split, not a target;
     an `@path` import counts toward it, since imports load with the guide.
   - **Rules are concrete.** "Run `make check` from `services/api` before
     committing" replaces "test your changes"; "handlers live in
     `internal/http/`" replaces "keep files organized". A rule the agent
     cannot act on or check is reworded or dropped.
   - **The agent verifies its own change.** List the build, lint, and test
     commands it runs before calling a change done, through the entry point
     CI already uses (`make check`, a Taskfile target, a script), with tool
     versions as CI pins them.
   - **Enforced rules name their check.** A rule a linter, hook, or CI job
     enforces gets one line naming that check, not a restatement. A rule
     that must never break and has no check stays as written; report that
     a check is missing instead of strengthening the wording.
   - **New rules need evidence.** Add a rule from configuration, explicit
     policy, an agent mistake the user reports as recurring, or a review
     finding the agent should have known from the repository, not from
     general Go advice the agent already follows. Project-independent
     agent behavior ("think before coding", "keep changes surgical") stays
     out of the project guide; it belongs in user-level instructions such
     as `~/.claude/CLAUDE.md` or in a plugin.
6. **Verify and report.** Check the resulting Markdown, relative links and
   imports, command targets, and working directories. Inspect the diff and
   newly created guides for scope and preservation of local edits. Run
   documented commands only when required by the repository or requested
   by the user, with their prerequisites and side effects understood.
   Report inspected commands separately from observed execution results.
   For an audit, give findings and suggested wording with evidence
   locations. For an update, give the changed files, the line count of the
   always-loaded text before and after, verification results, and
   unresolved conflicts. If the guides already match the evidence,
   leave them as they are.

## Example

A root guide says to run `go test ./...` from the root. The root has no
`go.mod` or `go.work`, and the CI job sets `working-directory: services/api`
and runs `make check`. The API Makefile's `check` target runs tests with
`-race`. Document `(cd services/api && make check)` for that module and cite
the workflow step and the Makefile target. Keep a separate command for
other modules when their checks differ.

## Related Skills

- [go-style-core](../go-style-core/SKILL.md#house-style-wins) for project rules and local conventions.
- [go-linting](../go-linting/SKILL.md#verification-gate) when the request also changes the validation gate itself.
