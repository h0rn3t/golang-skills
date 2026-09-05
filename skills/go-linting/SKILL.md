---
name: go-linting
description: Use when setting up linting for a Go project, configuring golangci-lint, or adding Go checks to a CI/CD pipeline. Also use when starting a new Go project and deciding which linters to enable, even if the user only asks about "code quality" or "static analysis" without mentioning specific linter names. Does not cover code review process (see go-code-review).
allowed-tools: Bash(bash:*)
---

# Go Linting

More important than any "blessed" linter set: **lint consistently across a
codebase**. This skill owns the repository's verification gate — the commands
that decide whether Go work is finished.

## Resource Routing

- `scripts/setup-lint.sh` - Run when generating a `.golangci.yml`, validating the first lint pass, or producing JSON metadata.
- `assets/golangci.yml` - Use as the v2 golangci-lint baseline for established projects.

## Verification Gate

> **Normative**: Verify the requested work with observed results. Complete the
> repository's required checks; use the defaults below when it has no gate.
> A question or documentation-only edit does not require a Go runtime gate.

```bash
gofmt -l .            # inspect output: exit 0 alone does not mean clean
go build ./...
go vet ./...          # includes stdversion, printf, lostcancel, waitgroup
go test -race ./...
go fix -diff ./...    # preview only; scope and findings rules below
golangci-lint run ./...
govulncheck ./...     # dependency CVEs; run before release, not every edit
```

Gate rules:

- Read `AGENTS.md`, `CLAUDE.md`, `CONTRIBUTING.md`, and CI where present. Use
  their gate at its required scope; do not automatically union it with this one.
  A specific request such as "check it builds" selects that check, not the full
  gate. Report the selected scope; build-only success is not full-gate success.
- Run from the target module or workspace, not the installed skill directory.
  For a local change without a prescribed gate, start with affected packages;
  include consumers for shared APIs and broaden for cross-package risk or release.
  Documentation-only work needs the applicable documentation checks.
- For modernization, identify existing packages from the task's actual diff,
  including staged and untracked files when relevant. Pass explicit quoted
  package arguments (for example `go fix -diff ./internal/store`). An empty
  package list means skip; do not fall back to the current package or `./...`.
  Separate pre-existing findings from new ones; do not rewrite unrelated code.
- Use `-race` for concurrency changes and wherever the repository requires it.
  Inspect a `make test` recipe before using it: a plain `go test` is not evidence
  of a race check. Retain its setup and use a race-enabled equivalent if needed.
- Run `govulncheck` before release, for dependency changes in the requested diff
  (including staged changes), or when requested. Otherwise mark it not applicable.
- Fix attributable failures, then rerun affected checks. Reuse passing results
  for unchanged code and configuration, including across checklist items. Repeat
  or broaden only for new edits, failures, unresolved concerns, or required gates.
- Report each selected check as `pass`, `fail`, or `unavailable (reason)`;
  explicitly omitted checks are `skipped (reason)`. Overall `PASS` requires all
  required checks to pass; `FAIL` means a finding; `INCOMPLETE` means required
  evidence is unavailable with no known finding. Report both when they coexist.
  Missing tools, unsupported toolchains, and infrastructure failures are not
  clean results. Attribute an environment gap from evidence, not error text alone.

---

## Modernization: `go fix`

Go 1.27 ships the modernizers as `go fix` analyzers. `go fix -diff ./...`
previews; `go fix ./...` applies. Use the package scope established above and
inspect the preview before applying changes.

`go tool fix help` lists the current set. The ones that change guidance:

| Analyzer | Rewrites to |
|---|---|
| `waitgroupgo` | `wg.Go(f)` instead of `Add(1)`/`go`/`Done` (Go 1.25+) |
| `errorsastype` | `errors.AsType[T]` instead of `errors.As` (Go 1.26+) |
| `newexpr` | `new(expr)` instead of a temp variable (Go 1.26+) |
| `testingcontext` | `t.Context()` instead of `context.WithCancel` in tests |
| `forvar` | Deletes `x := x` loop captures (dead since Go 1.22) |
| `rangeint`, `minmax`, `omitzero`, `any` | `for i := range n`, `min`/`max`, `omitzero` tags, `any` |
| `slicessort`, `slicescontains`, `mapsloop`, `stringsseq`, `stditerators` | `slices`/`maps`/iterator APIs instead of hand-written loops |
| `stringsbuilder`, `stringscut`, `stringscutprefix` | `strings.Builder`, `Cut`, `CutPrefix` |

Select a subset with `go fix -waitgroupgo ./...`, or exclude with
`-NAME=false`. Review the diff: these carry fixes, not just diagnostics, and a
few change allocation behavior.

---

## Setup Procedure

1. Create `.golangci.yml` with `scripts/setup-lint.sh` or copy `assets/golangci.yml`
2. `golangci-lint config verify --config .golangci.yml` — validate the schema first
3. `golangci-lint run ./...`
4. Fix category by category (formatting, vet, style); re-run until clean

---

## Minimum Recommended Linters

| Linter | Purpose |
|--------|---------|
| [errcheck](https://github.com/kisielk/errcheck) | Ensure errors are handled |
| [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) | Format code and manage imports |
| [revive](https://github.com/mgechev/revive) | Common style mistakes (modern replacement for the deprecated golint) |
| [govet](https://pkg.go.dev/cmd/vet) | Analyze code for common mistakes |
| [staticcheck](https://staticcheck.dev) | Various static analysis checks |

## Additional Recommended Linters

| Linter | Purpose | When to enable |
|--------|---------|----------------|
| [gosec](https://github.com/securego/gosec) | Security vulnerability detection | Always for services handling user input |
| [ineffassign](https://github.com/gordonklaus/ineffassign) | Detect ineffectual assignments | Always — catches dead code |
| [misspell](https://github.com/client9/misspell) | Correct common misspellings | Always |
| [gocyclo](https://github.com/fzipp/gocyclo) | Cyclomatic complexity threshold | When functions exceed ~15 complexity |
| [exhaustive](https://github.com/nishanths/exhaustive) | Ensure switch covers all enum values | When using iota enums |
| [bodyclose](https://github.com/timakin/bodyclose) | Detect unclosed HTTP response bodies | Always for HTTP client code |

## Linters That Enforce the Skills

The baseline config turns these on so the gate checks what the `go-*` skills
teach instead of leaving it to review attention:

| Linter | Enforces | Skill |
|--------|----------|-------|
| `depguard` | Deny list: `pkg/errors`, `logrus`, `zap`, `x/exp/slices`, `x/exp/maps`, `google/uuid` | [go-packages](../go-packages/SKILL.md) dependency ladder |
| `errname`, `errorlint` | `ErrFoo`/`FooError` names; `errors.Is`/`AsType` over `==` and type assertions (`errorf` check off — `%v` at boundaries is deliberate) | [go-error-handling](../go-error-handling/SKILL.md) |
| `sloglint` | Static message, key-value attrs, `snake_case` keys | [go-logging](../go-logging/SKILL.md) |
| `noctx` | Outbound HTTP/SQL calls carry a context | [go-context](../go-context/SKILL.md), [go-http](../go-http/SKILL.md) |
| `rowserrcheck`, `sqlclosecheck` | `rows.Err()` after the loop; rows and statements closed | [go-database](../go-database/SKILL.md) |
| `perfsprint`, `prealloc` | `strconv` over `fmt.Sprint`; capacity hints | [go-performance](../go-performance/SKILL.md) |
| `usetesting` | `t.Context`, `t.TempDir`, `t.Setenv` over hand-rolled forms | [go-testing](../go-testing/SKILL.md) |
| `godot` | Doc comments end in a period | [go-documentation](../go-documentation/SKILL.md) |
| `exhaustive` | `switch` covers every enum member (`default` counts) | [go-style-core](../go-style-core/SKILL.md) |
| `gosec` | String-built SQL, `sh -c`, `template.HTML` on input, weak hashes, `InsecureSkipVerify`, `math/rand` for secrets | [go-security](../go-security/SKILL.md) |

Opt-in, not in the baseline: `contextcheck` (context lost mid-chain; noisy on
deliberate breaks), `testifylint` (only in repositories that use testify),
`modernize` (same rewrites as `go fix`, useful when the gate runs only in
golangci-lint).

`govulncheck` is not a golangci-lint linter — install and run it separately:
`go install golang.org/x/vuln/cmd/govulncheck@latest`. It reports only
vulnerabilities on reachable call paths, so its findings are actionable.

---

## Example Configuration

`assets/golangci.yml` is the maintained example and the only copy —
`setup-lint.sh` emits it verbatim. It targets golangci-lint v2 (verified with
2.13.1 on 2026-09-01), keeps `goimports` under `formatters`, and enables the
core linters, the production additions, and the skill-enforcing set above.

```bash
# Pin the version this skill is verified against
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.1

golangci-lint run              # all linters
golangci-lint run ./pkg/...    # specific paths
```

---

## Nolint Directives

```go
//nolint:errcheck // fire-and-forget logging; error is not actionable
_ = logger.Sync()
```

- Use `//nolint:lintername` — never bare `//nolint`
- Place the comment on the same line as the finding
- Include a justification after `//`

---

## CI/CD Integration

Run the verification gate in CI, pinning every tool version so local and
release behavior do not drift. Use `golangci/golangci-lint-action` on GitHub
Actions.

```bash
#!/bin/sh
# .git/hooks/pre-commit — lint only changed code to keep the loop fast
golangci-lint run --new-from-rev=HEAD~1
```

---

## Quick Reference

| Task | Command |
|------|---------|
| Full gate | Run the selected commands above individually and inspect diagnostics as well as exit status |
| Preview modernizations | `go fix -diff ./...` |
| Apply modernizations | `go fix ./...` |
| List modernizers | `go tool fix help` |
| Validate config schema | `golangci-lint config verify --config .golangci.yml` |
| Lint changed code only | `golangci-lint run --new-from-rev=HEAD~1` |
| Dependency CVEs | `govulncheck ./...` |

---

## Related Skills

- **Style foundations**: See [go-style-core](../go-style-core/SKILL.md) when resolving style questions that linters enforce (formatting, nesting, naming)
- **Code review**: See [go-code-review](../go-code-review/SKILL.md) when combining linter output with a manual review checklist
- **Error handling**: See [go-error-handling](../go-error-handling/SKILL.md) when errcheck flags unhandled errors and you need to decide how to handle them
- **Testing**: See [go-testing](../go-testing/SKILL.md) when running linters alongside tests in CI pipelines
