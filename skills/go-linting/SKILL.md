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

- `scripts/setup-lint.sh` - Run from the target module as `bash <installed-skill-dir>/scripts/setup-lint.sh [local-prefix]` when generating a `.golangci.yml`, validating the first lint pass, or producing JSON metadata.
- `assets/golangci.yml` - Use as the v2 golangci-lint baseline for established projects, and as the gate's lint config where the repository has none.
- `../../hooks/go-check-receipt.sh` - Where the plugin hook runs, execute the matching Bash verification command it prints before crediting a receipt; absent on standalone installs.
- `../../hooks/go-check-receipt.py` - Internal receipt runner behind the Bash entry point; inspect only when debugging the hook or verifier.
- `../../hooks/go-verification-routing.sh` - The plugin's Bash verification prerequisite after a Go edit; inspect only when diagnosing routing or opt-out behavior.
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
golangci-lint run ./...   # no repository config: add --config <installed-skill-dir>/assets/golangci.yml
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
- `golangci-lint config path` exits non-zero when the repository has no lint
  config; lint with the bundled baseline then. Without it golangci-lint runs
  five default linters: none of the `noctx`, `bodyclose`, `rowserrcheck`,
  `sqlclosecheck`, `gosec`, or `sloglint` checks the skills name runs, and
  `errcheck` reports the deferred `rows.Close`, `tx.Rollback`, and
  `resp.Body.Close` the baseline excludes.
- Use `-race` for concurrency changes and wherever the repository requires it.
  Inspect a `make test` recipe before using it: a plain `go test` is not evidence
  of a race check. Retain its setup and use a race-enabled equivalent if needed.
- Run `govulncheck` before release, for dependency changes in the requested diff
  (including staged changes), or when requested. Otherwise mark it not applicable.
- Fix attributable failures in the code, not in the check
  ([Holding the Bar](#holding-the-bar)), then rerun affected checks. Reuse passing results
  for unchanged code and configuration, including across checklist items. Repeat
  or broaden only for new edits, failures, unresolved concerns, or required gates.
- Where the plugin edit hook runs, a receipt can satisfy only the selected check
  with the same command/flags, canonical target, config, toolchain and inputs.
  With a shell, first run the `bash .../go-check-receipt.sh --gate <receipt-dir> <package-dir>`
  command printed by the last edit hook. A redirection or a pipe that only displays
  output (`2>&1`, `| tail`, `| head`) is still that command; another verification command in the
  same Bash call is not. It verifies the package receipts together,
  credits only supported current checks and lists `required_direct`. It uses absolute paths and
  requires the expected cwd and argv; no plugin environment variable is needed.
  For example, to select `go vet .` in `/project/pkg`:
  `bash <plugin-dir>/hooks/go-check-receipt.sh <receipt-dir>/vet.json /project/pkg go vet .`.
  Require exit 0 **and `hook_credit=true`**, then credit only its reported scope.
  A batch exit 1 means no checks were credited: inspect its JSON and run the
  selected checks directly. The plugin permits direct verification only after
  loading this skill and attempting a current verifier; a subsequent Go edit
  requires a new attempt. `GOLANG_SKILLS_VERIFICATION_GATE=off` disables that
  prerequisite, without granting credit. This is a workflow guard, not a shell sandbox.
  `valid=true` from the state-only Python verifier is insufficient. Hook output
  with `reuse=unverified` is not a final pass. If the command is unavailable or
  rejects the receipt, run the selected check directly or report it unavailable.
  A later edit to any input, config, dependency or build environment invalidates it;
  completion order alone never selects a receipt. Without a shell, report the
  hook's observed status and scope as `observed (hook, reuse unverified)`;
  required gate checks lacking current matching verification are unavailable.
  Silence, skipped checks and unavailable tools are never passes.
  A disclaimer does not turn `lint pass (receipt, reuse=unverified)` into an
  observed-only status. Write `lint unavailable (receipt not verified)` or
  `lint observed (hook, reuse unverified)`, not `lint pass`.
  The plugin checks hook-sourced pass claims at Stop against the current
  verifier credits; correct unsupported labels rather than repeating them.
  File-only gofmt and package vet do not satisfy a wider repository gate;
  new-findings-only lint does not satisfy full lint. Hook tests use `-short`
  without `-race`. Run build, the required race check, broader fix/lint and
  applicable govulncheck whenever matching evidence is absent. Current Codex
  and opencode runners do not connect this plugin hook; use their normal gate.
- A test check passes on evidence that the tests ran, not on exit 0.
  `ok … [no tests to run]` (a `-run` pattern that matches nothing, or a file
  behind a `//go:build` tag the run did not pass), `[no test files]`, and a
  `TestMain` that returns before `m.Run`
  ([INTEGRATION.md](../go-testing/references/INTEGRATION.md#returning-before-mrun))
  all exit 0 having run nothing; an `Example` without `// Output:` is compiled
  and never run. The evidence is `--- PASS: TestX` under `-v` or its
  `"Action":"pass"` event under `-json`.
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

`go tool fix help` lists the current set; most rewrite to a form the
[idiom card](../go-style-core/SKILL.md#current-go-idiom-card) lists, and
[MODERNIZATION.md](../go-code-refactor/references/MODERNIZATION.md#start-with-go-fix)
names the rest, the Go 1.27 renames, and which hunks change behavior.

Select a subset with `go fix -waitgroupgo ./...`, or exclude with
`-NAME=false`. Review the diff: these carry fixes, not just diagnostics, and a
few change allocation behavior.
After applying fixes, build the affected packages and review retained comments:
modernizers can leave unused imports/variables or discard comments inside a
rewritten loop. Use the existing gate once on the final code, not a duplicate
verification cycle.


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

## Holding the Bar

> **Normative**: A failing check turns green by a change to the code, never by
> a change to the check.

The cheap road to green, and what it looks like in a diff:

| Move | In the diff |
|---|---|
| Silence a checker | a new `//nolint`, `//lint:ignore`, or `#nosec`; a `.golangci.yml` edit that disables a linter, adds an exclusion, or raises a threshold |
| Make a test easier | a new `t.Skip` or `testing.Short()` guard; a test file or test function deleted; an assertion, `// Output:` line, or table case removed from a test that stays; a `testdata/*.golden` file rewritten to match the new output |
| Drop a check | `-race` removed from the `Makefile`, CI, or a script |
| Leave the work unfinished | `panic("not implemented")` where the behavior should be |

Each move is legitimate only with a reason a reviewer can check, on the line
or in the report: the directive's reason ([Nolint Directives](#nolint-directives)),
the infrastructure a skipped test needs, the requirement that changed a golden
file. Without one the failure stays and is reported as `fail`. Tightening needs
no mention. `pre-review.sh --new-from-rev <base>` lists these moves in a diff
([go-code-review](../go-code-review/SKILL.md)).

---

## Related Skills

- [go-code-review](../go-code-review/SKILL.md): linter output alongside the manual checklist.
- [go-error-handling](../go-error-handling/SKILL.md): what to do with an `errcheck` finding.
- [go-testing](../go-testing/SKILL.md): linters and tests in one CI pipeline.
