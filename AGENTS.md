# Repository Guidelines

## Project Structure & Module Organization

This repository packages Go-focused Agent Skills; it is not a Go application and has no root `go.mod`.

- `skills/go-*/`: runtime skills. Keep each `SKILL.md` with its optional `references/`, `scripts/`, and `assets/`.
- `evals/`: the only Go module; contains structural tests, fixtures, golden cases, and the `evalrun`/`abrun` runners.
- `hooks/` and `agents/`: Claude Code routing, post-edit hooks, and the optional `go-verify` agent.
- `docs/`: authoring, ownership, script-contract, and release policy. `source/` contains licensed upstream snapshots and is not edited.

## Build, Test, and Development Commands

Run these from the repository root:

```bash
(cd evals && go test -count=1 -race -shuffle=on ./...)
for skill_dir in skills/*/; do npx --yes agentskills-validate@1.0.1 "$skill_dir"; done
bash -n hooks/*.sh
golangci-lint config verify --config skills/go-linting/assets/golangci.yml
```

The first command is the CI structural suite; the others validate skill metadata, shell syntax, and the pinned lint configuration. Model-driven runs are opt-in and costly: `cd evals && go run ./cmd/evalrun -set validation -kind all -j 2 -out evals-results.json`.

## Coding Style & Naming Conventions

Write decision-oriented Markdown. Keep each `SKILL.md` at 400 lines or fewer and references at 300 or fewer. Route every bundled resource in `## Resource Routing`, use relative links, and assign each rule one owner in `docs/RULE_OWNERSHIP.md`. Go examples must be formatted and compilable; target Go 1.27 with 1.26 and 1.27 supported. Use `gofmt` for Go and portable Bash for hooks and scripts.

## Testing Guidelines

Put structural tests in `evals/*_test.go` with descriptive `Test...` names. Update fixtures and golden cases when changing scripts, hooks, routing, manifest counts, or example behavior. Run a focused subtest while iterating, then the full race-enabled suite before opening a PR. Keep model-run reports and traces outside the repository.

## Commit & Pull Request Guidelines

Use the history’s scoped conventional style, such as `feat(go-http): ...`, `fix(hooks): ...`, `docs: ...`, or `release: 1.21.2`. Keep commits focused. PRs should explain intent, list affected files, record validation commands and baseline failures, update `CHANGELOG.md` under `## [Unreleased]` for user-visible changes, and update README/manifests when resource counts or versions change. For releases, follow [`docs/RELEASE_CHECKLIST.md`](docs/RELEASE_CHECKLIST.md); do not bump versions for ordinary PRs.
