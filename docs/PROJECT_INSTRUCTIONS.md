# Project Instructions Template

A short block for a work project's instruction file, so the first Go edit
starts from the router and `go-style-core`, which carries the idiom card,
instead of depending on the model to pick them. Paste the block for the client the team
uses; a team on several clients keeps both, one per file.

## Claude Code with the plugin

In the project's `CLAUDE.md`:

```markdown
## Go work

Before the first edit of a `.go` file, in one message:

1. Load the router with the Skill tool: `golang-skills:go-code` to write or
   fix Go, `golang-skills:go-code-refactor` for a behavior-preserving refactor.
2. Load `golang-skills:go-style-core`; it carries the idiom card.
3. Load only the owner skills the router's "Route Before The First Edit" table
   names for this task, never the whole pack.

Close with the router's verification gate.
```

The plugin registers each skill as `golang-skills:<name>`, so the block uses
those names; the `name` field in each `SKILL.md` stays bare (`go-code`) for
the other clients. The plugin's hooks enforce the same order:

- `hooks/go-prompt-routing.sh` names the router for a prompt that asks for Go
  work, including a prompt that only names a router (`use go-code`,
  `/opsx:apply add-auth /go-code`), with the exact Skill names: the router
  and `go-style-core` in one line, then `go-testing` and the owners the
  target's code points at; a review prompt gets `go-code-review` alone.
- `hooks/go-code-routing.sh` holds every `.go` edit until a router,
  `go-style-core` (which carries the idiom card), and the owners the edit's
  content points at are recorded. A session with no router is told which one to
  load. A retry without the loads is blocked again; a third retry of the same
  edit with nothing loaded in between stops the session with the reason
  instead of a third block. A skill missing from the plugin copy is reported,
  not required. `GOLANG_SKILLS_ROUTING_GATE=off` switches the gate off.
- `hooks/go-subagent-routing.sh` repeats that note at session start in a Go
  directory and to each subagent that may write Go: `go-code` and
  `go-style-core` in one message, before the first edit. `go-verify`,
  Explore, claude-code-guide, and statusline-setup hear nothing.

The hooks catch a skipped load after the fact; the instruction gets the loads
into the first message, which saves the blocked turn.

Claude Code with the skills copied into `~/.claude/skills/` instead of the
plugin runs none of these hooks and registers bare names: use the block above
with `go-code`, `go-code-refactor`, and `go-style-core`.

## Clients without these hooks

Codex, opencode, GitHub Copilot, Cursor, and any other client that reads the
skills from a directory. In `AGENTS.md`, or the file the client reads for project
instructions:

```markdown
## Go work

The golang-skills pack is installed in `<skills>` (for example
`~/.agents/skills`). Before the first edit of a `.go` file, read whole:

1. The router: `<skills>/go-code/SKILL.md` to write or fix Go, or
   `<skills>/go-code-refactor/SKILL.md` for a behavior-preserving refactor.
2. `<skills>/go-style-core/SKILL.md`, which carries the idiom card.
3. Only the owner skills the router's routing table names for this task,
   never the whole pack.

No hook checks this here: name the skills you read before the first edit.
```

Replace `<skills>` with the install directory from the
[README](../README.md#installation) table. Nothing blocks an edit in these
clients, so the instruction is the whole mechanism, and the last line makes a
skipped load visible in the reply rather than only in the transcript. Codex
also takes an explicit `$go-code <task>`.
