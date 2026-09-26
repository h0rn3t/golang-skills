# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Language
 
- Respond to the user in Ukrainian.
- Write reasoning summaries, planning notes, specs, review summaries, code comments, and user-facing messages in Ukrainian.
- Keep code identifiers, filenames, commands, package names, and commit messages in English.

## What this repository is

A Claude Code plugin and Agent Skills pack: 24 Go skills under `skills/go-*/`,
each a `SKILL.md` plus optional `references/`, `scripts/`, `assets/`. There is
no Go application. The only Go module is `evals/`, whose tests are the
structural suite for the Markdown, shell scripts, and hooks, plus two headless
runners (`evalrun`, `abrun`) that drive real model sessions. `source/` holds
upstream style-guide snapshots and is never edited. `docs/` holds the retained
maintenance policy; model-run reports and traces are not committed.

`CLAUDE.md` is listed in `.gitignore`, so this file stays local unless that
entry is removed.

## Commands

Everything the tests shell out to must be on PATH: `go` (1.27), `gofmt`, and
`golangci-lint` (v2.13.2 is pinned in CI; `TestScriptFunctional/SetupLintDryRun`
fails without the binary).

```bash
# Structural suite, as CI runs it. Execs the scripts, compiles every example
# block, drives the hooks through scripted sessions; about a minute with -race.
cd evals && go test -count=1 -race -shuffle=on ./...

# One test or subtest
cd evals && go test -count=1 -run 'TestSkillArchitecture' .
cd evals && go test -count=1 -run 'TestManifestCounts/README.md' .
cd evals && go test -count=1 -run 'TestRoutingGate|TestPromptRouting|TestVetHook|TestSubagentRouting' .

# abrun's own unit tests
cd evals && go test -count=1 ./cmd/abrun/

# The rest of the CI validate job
for d in skills/*/; do npx --yes agentskills-validate@1.0.1 "$d"; done
golangci-lint config verify --config skills/go-linting/assets/golangci.yml
bash -n hooks/*.sh
```

Model-driven runs cost tokens and are opt-in:

```bash
# Trigger + quality evals from evals/evals.json through `claude -p`
cd evals && go run ./cmd/evalrun -set validation -kind all -j 2 -out evals-results.json

# A/B a skill edit: reference arm from a second checkout, baseline = this tree
cd evals && go run ./cmd/abrun -corpus implement -reference-root ../golang-skills-before \
  -arms reference,baseline -n 5 -j 4 -effort medium -keep -out run.json
```

## Architecture

### Routing model

`go-code` is the router for any task that writes Go: load `go-style-core`
(always) and its `references/CURRENT-GO.md` idiom card before the first edit,
match the task against the "Route Before The First Edit" table to pick owner
skills and load them in one message, then close with the Verification Gate
that `go-linting` owns (`gofmt`, `go vet`, `go test -race`, `go fix -diff`,
`golangci-lint`, `govulncheck`). `go-code-refactor` is the second router, for
behavior-preserving changes. Skills link to siblings relatively
(`../go-x/SKILL.md#anchor`) so the same tree works under Codex, Cursor, or
copied into `~/.claude/skills/`.

`hooks/hooks.json` wires five scripts that make the routing enforceable in
Claude Code:

- `go-code-routing.sh`: PostToolUse on `Skill|Read` records loaded skills;
  PreToolUse on `Edit|Write` blocks (exit 2) a `.go` edit in a session that
  loaded `go-code` until `go-style-core` and the owners are loaded. Each name
  blocks once per session, so a retry passes. State lives under
  `${CLAUDE_PLUGIN_DATA:-$TMPDIR/golang-skills-hooks}/routing/<session>/`.
- `go-vet-on-edit.sh`: PostToolUse on `Edit|Write` of a `.go` file runs gofmt,
  vet, `go fix -diff`, the package tests, and golangci-lint.
  `GOLANG_SKILLS_EDIT_TESTS=off` / `GOLANG_SKILLS_EDIT_LINT=off` switch parts
  off. Never blocks.
- `go-prompt-routing.sh` (UserPromptSubmit) and `go-subagent-routing.sh`
  (SubagentStart) inject one note naming the router; silent otherwise.
- `go-restraint-ladder.sh`: SessionStart and SubagentStart in a Go directory
  print the ladder section of `OVER-ENGINEERING.md` plus the session's
  `lite`/`full`/`ultra` row from go-code's Intensity table, both read at run
  time; UserPromptSubmit records a level word in `<state>/intensity`.
  `GOLANG_SKILLS_LADDER=lite|full|ultra|off`.

`evals/hook_test.go` drives all five.

### Rule ownership

Every rule has one owner skill; other skills route to it in one or two lines
plus a link. `docs/RULE_OWNERSHIP.md` is the map and is updated before guidance
appears in a second skill. `TestRuleOwnershipMap` pins specific needles
("Reduce Nesting" only in go-style-core, `loggerFromCtx` only in the logging
reference); `TestRestraintLadder` and `TestDeleteFirstRule` pin that those two
rules each live in exactly one file. The map's "Scope Exceptions" section
records where this repository's own tooling deliberately departs from a
skill's advice (the CI workflow is the main case). Do not "fix" those without
updating the record.

### Invariants the eval tests enforce on skill files

Spread across `evals/eval_test.go`; an ordinary edit trips them:

- `SKILL.md` at most 400 lines, largest fenced block at most 40 lines;
  references at most 300 lines, with a `## Contents` TOC when over 200.
- Frontmatter is `name`, `description`, and (script-backed skills only)
  `allowed-tools`. Descriptions are goldens in
  `TestFrontmatterDescriptionsInvariant`: changing one means changing the test
  in the same commit.
- Exactly one `## Resource Routing` section listing every file under
  `references/`, `scripts/`, `assets/` (an unlisted file fails as an orphan),
  and a `## Related Skills` section. Every relative link and `#anchor` must
  resolve (`TestCrossRefs`; the CI link check also covers `docs/`, `agents/`,
  and root Markdown).
- A `SKILL.md` body that names a Go version needs a `> Compatibility:` note
  pointing at `COMPATIBILITY.md`. Inline `(Go 1.NN+)` markers are checked
  against `$GOROOT/api/go1.NN.txt` by `TestVersionClaimsMatchToolchain`:
  verify with `go doc` or the api files, never from memory.
  `TestGoVersionBaseline` pins the 1.27 baseline and specific rows.
- Reference files open with `> Sources:`, `> Authority:`, `> Last verified:`
  within the first 12 lines.
- Self-contained Go example blocks are extracted and compiled (some run) by
  the `*_examples_test.go` files. A deliberate fragment starts mid-function or
  is introduced as one in the sentence before it.
- Counts stated in `README.md`, `README.uk.md`, `.claude-plugin/plugin.json`,
  and `.claude-plugin/marketplace.json` (skills, reference files, scripts,
  asset templates, trigger and quality evals) must equal what is on disk, and
  the two manifests must agree on the version (`TestManifestCounts`). Adding a
  reference or script means updating all four files; the Ukrainian README is a
  mirror, not optional.

`docs/SKILL_AUTHORING_TEMPLATE.md` is the authoring contract behind these,
including how to write for a capable model: cut what it already knows, put the
exception in the code sample rather than a caveat beside it, state a rule's
scope in the rule, name the check that fails when the rule is broken.

### Bundled scripts

The 11 `skills/*/scripts/*.sh` share one CLI contract: `--help`, `--json`,
exit 0 clean / 1 findings / 2 usage or environment error, `--limit` where
output is unbounded, `--force` on the three that write files. JSON shapes are
specified in `docs/SCRIPT_JSON_CONTRACTS.md` and asserted by
`TestScriptFunctional` against `evals/fixtures/<area>/` (each area has a clean
tree and a violations tree; `no_go_files` checks the empty-scan status
marker). Scripts resolve paths relative to their installed skill directory and
run against the target project.

### Evaluation harnesses

- `evals/cmd/evalrun` runs `evals.json` (`trigger_evals`, `quality_evals`)
  through `claude --plugin-dir` in an empty scratch dir with tools cut to
  `Skill` (plus read-only tools for quality), records which skills fired, and
  has a second model grade quality answers. `evals/internal/evalplugin` stages
  a copy of the plugin so restricted sessions can read references without
  seeing fixtures.
- `evals/cmd/abrun` measures the code a session writes, not its prose, across
  arms: `no-skill` (control), `baseline` (this tree), `reference`
  (`-reference-root` to another checkout; the arm for before/after claims),
  and one arm per `evals/ab/variants/*.md` spliced into
  `go-code-refactor/SKILL.md`. Three corpora, never averaged together:
  `evals/ab/<fixture>` refactor (score is lines removed with behavior held),
  `evals/ab/_implement` (stub bodies; the hidden golden test is the spec), and
  `evals/ab/_review` (seeded defects scored against
  `_golden/<fixture>/key.json` by source line). `_golden/` is copied in after
  the session and never reaches the model; the model's own tests are renamed
  `*_test.go.model` first. `-runner` supports claude, codex, copilot, and
  opencode with differences documented by the runner implementations; compare
  arms within one runner. `no-skill` vs `baseline` only shows a fixture has a trap;
  `reference` vs `baseline` is what judges a skill edit.
- Model-driven runs are exploratory and their reports, traces, and scratch
  trees stay outside this repository. If a result is published elsewhere,
  retain the raw JSON with it and compare arms within one runner.

## Conventions

- Changes are recorded under `## [Unreleased]` in `CHANGELOG.md`. A release is
  one `release: X.Y.Z` commit bumping `.claude-plugin/plugin.json`,
  `.claude-plugin/marketplace.json`, and the changelog heading, then an
  annotated `vX.Y.Z` tag. `docs/RELEASE_CHECKLIST.md` lists the validation set.
- Go baseline is 1.27 (supported: 1.26 and 1.27). `COMPATIBILITY.md` is the
  source of truth for every version-sensitive claim and says how to re-verify
  each against the installed toolchain. The language version is the module's
  `go` directive, not the installed Go. `go-release-watch.yml` reads the
  baseline sentence in `COMPATIBILITY.md` and `GOLANGCI_LINT_VERSION` in
  `validate-skills.yml` by regex, so keep both spellings intact.
- Skill wording is judged by measurement, not by reading: a claim that an edit
  helps needs a reference-vs-baseline `abrun` run with raw output retained
  outside the repository. The
  400-line cap is a ceiling, not a target; cut only on measured evidence.
- `source/` snapshots keep upstream licenses and provenance headers, and
  `THIRD_PARTY_NOTICES.md` must cover every file there. Project files are
  Apache-2.0.
