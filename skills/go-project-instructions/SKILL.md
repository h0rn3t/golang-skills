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
5. **Apply the smallest authorized edit.** Keep the existing language,
   structure, and project-specific policies. Put shared rules at their
   existing source of truth and module-specific commands at the scope they
   serve. Keep host-specific guidance and applicable nested exceptions.
   When creating both root files, write shared rules in `AGENTS.md` and
   import it with `@AGENTS.md` in `CLAUDE.md`. For a new guide, include the
   project layout, commands with working directories and prerequisites, and
   durable local constraints that an agent needs. Link to maintained
   project docs for long procedures. Keep secrets, local machine paths, and
   temporary observations out of persistent instructions.
6. **Verify and report.** Check the resulting Markdown, relative links and
   imports, command targets, and working directories. Inspect the diff and
   newly created guides for scope and preservation of local edits. Run
   documented commands only when required by the repository or requested
   by the user, with their prerequisites and side effects understood.
   Report inspected commands separately from observed execution results.
   For an audit, give findings and suggested wording with evidence
   locations. For an update, give the changed files, verification results,
   and unresolved conflicts. If the guides already match the evidence,
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
