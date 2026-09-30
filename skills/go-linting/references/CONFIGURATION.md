# Linter Configuration and CI

> Sources: golangci-lint v2 configuration schema and linter catalogue; source/uber-go-style/style.md (Linting); GitHub Actions docs; `go help testflag`
> Authority: project policy for the baseline config and the CI pipeline shape; tool behavior follows golangci-lint 2.13.2 and go1.27.1
> Last verified: 2026-09-30

## Setup Procedure

1. Create `.golangci.yml` with `scripts/setup-lint.sh` or copy `assets/golangci.yml`
2. `golangci-lint config verify --config .golangci.yml` — validate the schema first
3. `golangci-lint run ./...`
4. Fix category by category (formatting, vet, style); re-run until clean

## Baseline Linters

`assets/golangci.yml` enables `errcheck`, `govet`, `revive`, and `staticcheck`
as the minimum, adds `bodyclose`, `gocyclo`, `gosec`, `ineffassign`, and
`misspell` for production code, and turns on the skill-enforcing set below; the
comment beside each entry names the rule it enforces. `goimports` runs as a
formatter under `formatters`, not as a linter.

## Linters That Enforce the Skills

The baseline config turns these on so the gate checks what the `go-*` skills
teach instead of leaving it to review attention; the comments in
`assets/golangci.yml` name the skill rule behind each linter and setting.

Opt-in, not in the baseline: `contextcheck` (context lost mid-chain; noisy on
deliberate breaks) and `testifylint` (only in repositories that use testify).
Left off after a noise check on five codebases: `revive`'s `unused-parameter`,
which fires on the `w, r` a handler type fixes, and `iface`'s `unused` and
`identical`, which fire on exported interfaces a library offers its callers.
Left off because it contradicts a skill: `prealloc`, which reports every
`var out []T` filled by `append` in a loop, hot path or not, and whose fix
turns a nil result (`null` under `encoding/json` v1) into an empty one —
[go-performance](../../go-performance/SKILL.md) measures first, and
[go-data-structures](../../go-data-structures/SKILL.md) owns nil against empty.

`govulncheck` is not a golangci-lint linter. Track it as a tool dependency so
local runs and CI share one pin — `go get -tool golang.org/x/vuln/cmd/govulncheck@vX.Y.Z`,
then `go tool govulncheck ./...` ([go-packages](../../go-packages/SKILL.md#adding-and-auditing-dependencies)
owns the directive).

## Example Configuration

`assets/golangci.yml` is the maintained example and the only copy —
`setup-lint.sh` emits it verbatim. It targets golangci-lint v2.

```bash
# Pin the version this skill is verified against
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2
```

## CI/CD Integration

Run the verification gate in CI, pinning every tool version so local and
release behavior do not drift. Use `golangci/golangci-lint-action` on GitHub
Actions.

A minimal Go pipeline runs test and lint on every PR:

- **Matrix** covers the version in the `go` directive plus `stable`, with
  `fail-fast: false` so one version never cancels the rest. While the directive
  names the current release those two resolve to the same toolchain — the
  second entry earns its minutes when the next minor ships. Older directives
  add their supported minors.
- **Test flags**: `go test -race -shuffle=on ./...`, plus `-count=1` for
  suites that touch real services so caching cannot hide flakes.
- **Hygiene**: the tidy check [go-packages](../../go-packages/SKILL.md#adding-and-auditing-dependencies)
  owns.
- **Vulnerability scan**: the gate's `govulncheck` triggers plus a scheduled
  run, since a new advisory lands against code that did not change.
- **Pinning**: pin each GitHub Action to a full commit SHA. A `@vN` tag moves,
  so it is a compatibility marker, not a supply-chain control — GitHub's
  hardening guidance treats the SHA as the only immutable reference. Pin tool
  versions to the ones local runs use, `govulncheck` included: its
  vulnerability database is fetched at run time, so a pinned binary still
  reports today's advisories.
- **Permissions**: least-privilege `permissions:` on each job; only release
  jobs get `contents: write`.
