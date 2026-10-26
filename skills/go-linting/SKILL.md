---
name: go-linting
description: Use when configuring Go linting, golangci-lint, static analysis, or CI checks, or choosing linters for a new project. Reviewing code belongs to go-code-review.
allowed-tools: Bash(bash:*)
---

# Go Linting

This skill owns the repository's verification gate — the commands that decide
whether Go work is finished.

> Compatibility: Baseline Go 1.27 (see the plugin repository's COMPATIBILITY.md). Analyzer and
> linter names are checked against go1.27.2 and golangci-lint 2.15.0.

## Resource Routing

- `scripts/setup-lint.sh` - Run from the target module as `bash <installed-skill-dir>/scripts/setup-lint.sh [local-prefix]` when generating a `.golangci.yml`, validating the first lint pass, or producing JSON metadata.
- `assets/golangci.yml` - Use as the v2 golangci-lint baseline for established projects, and as the gate's lint config where the repository has none.
- `../../hooks/go-check-receipt.sh` - Where the plugin hook runs, execute the matching Bash verification command it prints before crediting a receipt; absent on standalone installs.
- `../../hooks/helper/receipt.go` - Internal receipt runner behind the Bash entry point; inspect only when debugging the hook or verifier.
- `../../hooks/go-verification-routing.sh` - The plugin's Bash verification prerequisite after a Go edit; inspect only when diagnosing routing or opt-out behavior.
- `references/CONFIGURATION.md` - Read for any linter configuration, CI, or first-lint setup task.
- `references/RECEIPTS.md` - Read before crediting an edit-hook receipt or writing a `pass (hook)` or `pass (receipt)` label.

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
- The default gate always runs `-race`. A repository gate without it gets one
  `-race` run of the changed packages for a concurrency change: inspect a
  `make test` recipe first, since a plain `go test` is not evidence of a race
  check, and keep its setup in the race-enabled equivalent.
- A check's result is its own exit status, not the pipeline's: `go test ./... | tail`
  exits 0 on a failure, and `| head` kills `go test` with SIGPIPE before later
  packages run. Run gate commands unpiped or under `set -o pipefail`.
- Run `govulncheck` before release, for dependency changes in the requested diff
  (including staged changes), or when requested. Otherwise report it
  `skipped (no dependency change)`.
- Fix attributable failures in the code, not in the check
  ([Holding the Bar](#holding-the-bar)), then rerun affected checks. Reuse passing results
  for unchanged code and configuration, including across checklist items. Repeat
  or broaden only for new edits, failures, unresolved concerns, or required gates.
  A check that fails the same way after two fixes aimed at it has a cause not
  yet found: stop patching, take [go-troubleshooting](../go-troubleshooting/SKILL.md)'s
  loop from the first diagnostic, and if it does not close, report the check
  `fail` with both attempts.
- Where the plugin edit hook runs, first run the `go-check-receipt.sh --gate`
  command the last edit hook printed, as printed and alone in its Bash call,
  before any direct check and again after every later Go edit: the plugin
  blocks a direct check until then. Credit only what it reports
  with exit 0 **and `hook_credit=true`**, at its reported scope; run every
  other selected check directly. Hook tests use `-short` without `-race`.
  Without a shell, a hook result is `observed (hook, reuse unverified)`, never
  `pass`. [RECEIPTS.md](references/RECEIPTS.md) holds the rest of the receipt
  rules; read it before writing a `pass (hook)` or `pass (receipt)` label.
- A test check passes on evidence that the tests ran, not on exit 0.
  `ok … [no tests to run]` (a `-run` pattern that matches nothing, or a file
  behind a `//go:build` tag the run did not pass), `[no test files]`, and a
  `TestMain` that returns before `m.Run`
  ([INTEGRATION.md](../go-testing/references/INTEGRATION.md#returning-before-mrun)),
  and a package whose tests all called `t.Skip` all exit 0 having run nothing; an `Example` without `// Output:` is compiled
  and never run. The evidence is `--- PASS: TestX` under `-v` or its
  `"Action":"pass"` event under `-json`.
- Report each selected check as `pass`, `fail`, or `unavailable (reason)`;
  explicitly omitted checks are `skipped (reason)`. Overall `PASS` requires all
  required checks to pass; `FAIL` means a finding; `INCOMPLETE` means required
  evidence is unavailable with no known finding. Report both when they coexist.
  Missing tools, unsupported toolchains, and infrastructure failures are not
  clean results. Attribute an environment gap from evidence, not error text alone.
  A missing tool is reported, not installed: unless the task asks, the gate
  never runs `go install` or `go get`, edits `go.mod`, or switches the
  toolchain to make a check runnable.

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
Fixes from two analyzers that overlap are not shown at all: `go fix -diff`
exits 1 with only `fix: N of M fixes skipped (e.g. due to conflicts)`, and on
go1.27.1 a plain `go fix` then reports them applied without writing a file.
Preview and apply one analyzer at a time (`go fix -diff -rangeint ./pkg`), and
confirm a write with `git diff --stat`.
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
| Quiet the compiler | `_` on an error a changed signature now returns; `return nil, nil` or a zero value added for `missing return`; `n, _ := strconv.Atoi(s)`: the build passes and the behavior is invented |

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
