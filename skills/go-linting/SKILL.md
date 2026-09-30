---
name: go-linting
description: Use when configuring Go linting, golangci-lint, static analysis, or CI checks, or choosing linters for a new project. Reviewing code belongs to go-code-review.
allowed-tools: Bash(bash:*)
---

# Go Linting

This skill owns the repository's verification gate — the commands that decide
whether Go work is finished.

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). Analyzer and
> linter names are checked against go1.27.1 and golangci-lint 2.13.2.

## Resource Routing

- `scripts/setup-lint.sh` - Run when generating a `.golangci.yml`, validating the first lint pass, or producing JSON metadata.
- `assets/golangci.yml` - Use as the v2 golangci-lint baseline for established projects.
- `references/CONFIGURATION.md` - Read for any linter configuration, CI, or first-lint setup task.

## Verification Gate

> **Normative**: Verify the requested work with observed results. Complete the
> repository's required checks; use the defaults below when it has no gate.
> A question or documentation-only edit does not require a Go runtime gate.

```bash
gofmt -l .            # inspect output: exit 0 alone does not mean clean
go build ./...
go vet ./...          # includes stdversion, printf, lostcancel, waitgroup
go test -race ./...
go fix -diff ./...    # preview only
golangci-lint run ./...
govulncheck ./...     # dependency CVEs
```

Gate rules:

- Read the convention files [go-style-core](../go-style-core/SKILL.md#house-style-wins)
  names. Use their gate at its required scope; do not automatically union it with this one.
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

Since Go 1.26 the modernizers are `go fix` analyzers.
`go fix -diff <packages>` previews; `go fix <packages>` applies. Use the package
scope established above and inspect the preview before applying changes.

`go tool fix help` lists the current set; most rewrite to a form
[CURRENT-GO.md](../go-style-core/references/CURRENT-GO.md) lists. The ones it
does not show:

| Analyzer | Rewrites to |
|---|---|
| `newexpr` | A `&v` pointer helper (`func intPtr(v int) *int { return &v }`) and its calls become `new(v)` / `new(4)` (Go 1.26+); a temp-then-address form stays a hand edit |
| `slicesbackward` | `for i, v := range slices.Backward(s)` instead of a backward index loop (Go 1.23+) |
| `unsafefuncs` | `unsafe.Add(p, n)` instead of `unsafe.Pointer(uintptr(p) + n)` |
| `omitzero` | Deletes `omitempty` from a struct-typed field, where it has no effect; the `omitzero` tag (Go 1.24+) it offers instead omits a zero struct, a behavior change `go fix` does not apply |

`atomictypes`, `embedlit`, `errorsastype`, `slicesbackward`, and `unsafefuncs`
are new in Go 1.27; the same release renamed `waitgroup` to `waitgroupgo` and
dropped `fmtappendf`, so a pinned command naming either of those now fails.

Select a subset with `go fix -waitgroupgo ./...`, or exclude with
`-NAME=false`. Review the diff: these carry fixes, not just diagnostics, and a
few change allocation behavior.
After applying fixes, build the affected packages and review retained comments:
modernizers can leave unused imports/variables or discard comments inside a
rewritten loop. Use the existing gate once on the final code, not a duplicate
verification cycle.

The x/tools `modernize` suite, versioned apart from the toolchain, also
carries `appendclipped` and `slicesdelete`: they are not `go fix` analyzers,
and they change nilness or zero the old slice tail, so never classify them as
behavior-preserving swaps.
Check the installed tool's help before naming flags.

---

## Nolint Directives

```go
return rand.N(d) //nolint:gosec // G404: retry jitter, not a secret
```

`nolintlint` in the baseline rejects a bare `//nolint`, one without a reason,
and one on a line with no finding for that linter; place the comment on the
finding's line. `_ = f()` needs no directive: `errcheck` does not report an
explicit discard.

---

## Related Skills

- [go-code-review](../go-code-review/SKILL.md): linter output alongside the manual checklist.
- [go-error-handling](../go-error-handling/SKILL.md): what to do with an `errcheck` finding.
- [go-testing](../go-testing/SKILL.md): linters and tests in one CI pipeline.
