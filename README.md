# Agent Skills For Go

**English** | [Українська](README.uk.md)

AI [Agent Skills](https://agentskills.io/) for writing idiomatic,
production-quality Go 1.27 code. The pack contains **24 modular skills**,
**69 reference files**, **11 bundled scripts**, and **5 asset templates**.
The Claude Code plugin also ships a `go-verify` agent and hooks for routing and
post-edit verification.

The guidance is derived from the [Google Go Style Guide](https://google.github.io/styleguide/go/),
[Effective Go](https://go.dev/doc/effective_go), the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md),
and [Go Wiki CodeReviewComments](https://github.com/golang/go/wiki/CodeReviewComments).

## Skills Included

The pack covers routing, refactoring, review, HTTP, databases, security,
resilience, troubleshooting, style, naming, errors, testing, concurrency,
generics, logging, documentation, performance, and package organization.
The current skill names are the directories under `skills/go-*/`.

## Bundled Scripts

**11 scripts automate** common checks. They support `--help`, structured
`--json` output, and documented exit codes; analysis scripts also support
`--limit`, and file-writing scripts require `--force`.

The plugin includes `agents/go-verify.md`, routing hooks under `hooks/`, and
the manifests under `.claude-plugin/`. In a Go project the hooks print the
restraint ladder at session start and into each subagent except `go-verify`
and the host's Explore, claude-code-guide, and statusline-setup; `/go-code ultra
<task>` or `lite mode` changes its level for the session, and
`GOLANG_SKILLS_LADDER=lite|ultra` sets the level a session starts at, and `off`
turns it off.

## Installation

### npx skills

```bash
npx skills add h0rn3t/golang-skills --all
```

### bun

`bunx` runs the same CLI without installing it, or install it once globally
(`bun add -g skills` puts it on `PATH` as `skills`):

```bash
# run without installing
bunx skills add h0rn3t/golang-skills --all

# or install the CLI once and reuse it
bun add -g skills
skills add h0rn3t/golang-skills --all
```

Updates work the same way: `bunx skills update` (add `-g` for the global
scope).

### Codex, Copilot, and Cursor

The same CLI installs into any agent it detects; name the targets with `-a`
when several are installed. All three read the project's `.agents/skills/`
directory, and `-g` installs for the user instead:

```bash
bunx skills add h0rn3t/golang-skills --all -a codex -a github-copilot -a cursor -g
```

| Agent | `--agent` | Global path |
| --- | --- | --- |
| Codex | `codex` | `~/.codex/skills/` |
| GitHub Copilot | `github-copilot` | `~/.copilot/skills/` |
| Cursor | `cursor` | `~/.cursor/skills/` |

Start a new session in the agent after installing.

### Claude Code plugin

```text
/plugin marketplace add h0rn3t/golang-skills
/plugin install golang-skills@golang-skills
```

### Manual installation

Copy the complete skill directories, including their `references/`, `scripts/`,
and `assets/` subdirectories:

```bash
cp -R skills/go-* ~/.claude/skills/
```

## Updating

### Claude Code

Refresh the marketplace snapshot, update the plugin, then restart Claude Code
to load the new version:

```bash
claude plugin marketplace update golang-skills
claude plugin update golang-skills@golang-skills
claude plugin list   # shows the installed version
```

### Codex, Copilot, and Cursor

Update the global install, then start a new session in the agent:

```bash
bunx skills update -g
```

Re-running `bunx skills add h0rn3t/golang-skills --all -g` also works and picks
up skills added since the last install.

### Manual installation

Pull the checkout and replace the skill directories rather than copying over
them, so files removed upstream do not linger:

```bash
git pull
rm -rf ~/.claude/skills/go-* && cp -R skills/go-* ~/.claude/skills/
```

For Codex, use `~/.codex/skills/` (the global path in the table above) or the project's `.agents/skills/` as the target directory.

## Project Instructions

A skill loads when the host's matcher or the model picks it. For a work
project where every Go edit must start from the router, paste the short block
from [`docs/PROJECT_INSTRUCTIONS.md`](docs/PROJECT_INSTRUCTIONS.md) into its
`CLAUDE.md` or `AGENTS.md`: router first, then `go-style-core` with its
idiom card, then only the owner skills the task needs. Under the Claude
Code plugin the routing gate also holds every `.go` edit until those loads
are recorded (`GOLANG_SKILLS_ROUTING_GATE=off` turns it off); Codex, Copilot,
and Cursor run no hooks, so the instruction is all there is.

## Validation

`evals/` contains the structural Go test suite, fixtures, golden tests, and
the optional `evalrun`/`abrun` harnesses. It contains **108 trigger evals** and
**62 quality evals**. Model-driven eval output is local scratch data and is not
stored in this repository. The suite also checks every `SKILL.md` against
the Agent Skills specification (`TestStructure`), so the checks
need Go and no Node.js.

Run the repository checks with:

```bash
(cd evals && go test -count=1 -race -shuffle=on ./...)
bash -n hooks/*.sh
golangci-lint config verify --config skills/go-linting/assets/golangci.yml
```

The release validation set is documented in
[`docs/RELEASE_CHECKLIST.md`](docs/RELEASE_CHECKLIST.md).

## Project Structure

```text
skills/       Runtime skill directories.
hooks/        Claude Code routing and post-edit hooks.
agents/       Optional plugin agents.
evals/        Structural tests, fixtures, golden tests, and runners.
source/       Upstream style-guide snapshots with provenance and licenses.
docs/         Authoring, ownership, script-contract, and release policy.
```

`COMPATIBILITY.md` is the source of truth for version-sensitive Go guidance.
`docs/RULE_OWNERSHIP.md` records which skill owns each rule, and
`docs/SCRIPT_JSON_CONTRACTS.md` defines the bundled script output contracts.

## Go 1.27

The skills target Go 1.27 and support Go 1.26 and 1.27. Verify version-sensitive
claims against the installed toolchain and update
[`COMPATIBILITY.md`](COMPATIBILITY.md) when the baseline changes.

## License

Project-authored files are Apache-2.0. Upstream snapshots and derived guidance
are listed in [`THIRD_PARTY_NOTICES.md`](THIRD_PARTY_NOTICES.md).
