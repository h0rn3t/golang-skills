# Codex and Cursor plugins

These installs package the existing `skills/` tree. Installing with `npx skills`
still installs skills alone, without plugin hooks or rules.

## Codex

From a release that contains the native manifests:

```bash
codex plugin marketplace add h0rn3t/golang-skills
codex plugin add golang-skills@golang-skills
```

For local development, replace the first command's repository argument with
the absolute path of this checkout. Open a new Codex session, then use `/hooks`
to review and trust the plugin hooks. New or changed definitions require a new
review; installation alone does not trust them. Clients without these plugin
commands can use the skills-only install described in the README.

The package uses the supported `.codex-plugin/plugin.json` compatibility
manifest and `.agents/plugins/marketplace.json` catalog. Keeping separate
Codex and Cursor manifests lets each select its own components. OpenAI also
supports a portable root `plugin.json`, but the portable components alone do
not include Cursor rules.

Codex loads `hooks/codex.json`, which registers session and prompt routing
notes and a `PostToolUse` handler for `apply_patch`. That handler reads
`tool_input.command`, resolves add/update paths against the event's `cwd`,
checks rename destinations, and ignores deletions and non-Go files. It runs
the existing file checker for each surviving Go file and combines its receipt
contexts into one hook response. A finding in one file does not skip the rest.
The outer hook has a 600-second timeout; checks can remain incomplete when a
large patch exceeds it. Shell edits are outside this handler's coverage.

The checker inspects gofmt output, runs `go vet`, previews `go fix -diff`, runs
package tests, and invokes golangci-lint. It reports skipped or unavailable
checks in receipts. The Claude skill-load gate and `go-verify` agent are not
registered by the Codex manifest. Routing notes ask for skills, but do not
prove that the model read them. Use the project's `AGENTS.md` block from
[Project Instructions](PROJECT_INSTRUCTIONS.md#clients-without-these-hooks)
when routing is a project requirement.

The helper needs bash and Go 1.26 or newer. It compiles into the user cache on
first use. `GOLANG_SKILLS_EDIT_TESTS=off` and `GOLANG_SKILLS_EDIT_LINT=off`
disable package tests and lint respectively, as in the Claude integration.
Only receipts accepted by `go-check-receipt.sh` may be credited as hook passes.

To update a Git-backed install, refresh the marketplace and reinstall:

```bash
codex plugin marketplace upgrade golang-skills
codex plugin add golang-skills@golang-skills
```

Start a new session and review changed hooks again.

## Cursor

Open **Customize**, choose **From GitHub Repository**, and enter
`h0rn3t/golang-skills`. Install `golang-skills` from that marketplace at the
project or user scope. The catalog is `.cursor-plugin/marketplace.json`; its
plugin manifest selects `skills/` and `rules/`.

For local development, copy the checkout to
`~/.cursor/plugins/local/golang-skills`, then reload Cursor. Local plugin
imports must be allowed by the host's organization settings.

The bundled `rules/go-skills.mdc` has `alwaysApply: true`. It routes Go work
to the installed skills without asking the user to paste a project rule or
hard-code an installation directory. It asks for `go-code-review` on reviews,
or `go-code`/`go-code-refactor` plus `go-style-core` before Go edits. Non-Go
work is outside the rule's scope.

The manifest explicitly supplies an empty Cursor hooks configuration and an
empty agents list. Cursor otherwise discovers `hooks/hooks.json` and
`agents/` automatically; those resources implement the Claude integration.
Cursor's native hook events and payloads need their own adapter before that
integration can be enabled. The Cursor rule is guidance, not an edit gate.

Manage updates through the installed marketplace in Customize. After updating,
check the installed rule and skills in a new chat.

## Host verification

Structural tests validate the manifests, catalogs, rules, hook paths, and
adapter behavior. They do not establish that a host discovered or delivered
the components. Verify each host separately when changing its packaging:

1. **Codex:** use a temporary `CODEX_HOME`, add this checkout as a marketplace,
   install the plugin, and inspect the installed version. In an interactive
   session, complete hook trust. Start a new chat and change a Go file through
   `apply_patch`; inspect persisted hook execution and receipts for that file.
   Confirm session routing context and post-edit feedback reached the model.
   Remove temporary credentials after the run; keep sanitized evidence outside
   the repository. Do not treat hooks skipped for lack of trust as a pass.
2. **Cursor:** install locally, reload, and inspect Customize for this plugin's
   skills and the always-applied Go rule. In a fresh chat, ask for a Go edit
   and check that the model receives the rule and loads its selected skills.
   Confirm that Claude hooks and the Claude `go-verify` agent were not loaded.
   A parsed manifest without a running Cursor session is structural evidence.

Format references: [OpenAI plugin packaging](https://developers.openai.com/plugins/build/plugins),
[Codex hooks](https://learn.chatgpt.com/docs/hooks),
[Cursor plugin reference](https://cursor.com/docs/reference/plugins), and
[Cursor rule import](https://cursor.com/docs/rules).
