# Agent Skills For Go

**English** | [Українська](README.uk.md)

AI [Agent Skills](https://agentskills.io/) for writing idiomatic,
production-quality Go 1.27 code. The pack contains **24 modular skills**,
**74 reference files**, **11 bundled scripts**, and **5 asset templates**.
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
restraint ladder at session start and into each subagent; `/go-code ultra
<task>` or `lite mode` changes its level for the session, and
`GOLANG_SKILLS_LADDER=off` turns it off.

## Installation

### npx skills

```bash
npx skills add h0rn3t/golang-skills --all
```

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

### Codex

Codex reads the skills that `npx skills` installs under `~/.agents/skills/`.
Update them, then start a new Codex session:

```bash
npx skills update -g
```

Re-running `npx skills add h0rn3t/golang-skills --all -g` also works and picks
up skills added since the last install.

### Manual installation

Pull the checkout and replace the skill directories rather than copying over
them, so files removed upstream do not linger:

```bash
git pull
rm -rf ~/.claude/skills/go-* && cp -R skills/go-* ~/.claude/skills/
```

For Codex, use `~/.agents/skills/` as the target directory.

## Validation

`evals/` contains the structural Go test suite, fixtures, golden tests, and
the optional `evalrun`/`abrun` harnesses. It contains **108 trigger evals** and
**62 quality evals**. Model-driven eval output is local scratch data and is not
stored in this repository.

Run the repository checks with:

```bash
for skill_dir in skills/*/; do
  npx --yes agentskills-validate@1.0.1 "$skill_dir"
done

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
