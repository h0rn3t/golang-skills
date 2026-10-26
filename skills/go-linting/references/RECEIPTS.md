# Hook Receipts

> Sources: project policy ([go-linting](../SKILL.md#verification-gate)); `hooks/go-check-receipt.sh`; `hooks/helper/receipt.go`; `hooks/go-verification-routing.sh`
> Authority: normative — when an edit-hook receipt may stand in for a gate check, and how the report labels it
> Last verified: 2026-10-25

Where the plugin's edit hook runs, it leaves one receipt per check. A receipt
can satisfy only the selected check with the same command/flags, canonical
target, config, toolchain and inputs.

## The Verifier

Where the plugin hook runs, with a shell, first run the
`bash .../go-check-receipt.sh --gate <receipt-dir> <package-dir>` command
printed by the last edit hook. A redirection or a pipe that only displays
output (`2>&1`, `| tail`, `| head`) is still that command; another
verification command in the same Bash call is not. It verifies the package
receipts together, credits only supported current checks and lists
`required_direct`. It uses absolute paths and requires the expected cwd and
argv; no plugin environment variable is needed. For example, to select
`go vet .` in `/project/pkg`:
`bash <plugin-dir>/hooks/go-check-receipt.sh <receipt-dir>/vet.json /project/pkg go vet .`.

Require exit 0 **and `hook_credit=true`**, then credit only its reported
scope. A batch exit 1 means no checks were credited: inspect its JSON and run
the selected checks directly. `valid=true` from the state-only verifier is
insufficient. If the command is unavailable or rejects the receipt, run the
selected check directly or report it unavailable.

The plugin permits direct verification only after loading go-linting and
attempting a current verifier; a subsequent Go edit requires a new attempt.
`GOLANG_SKILLS_VERIFICATION_GATE=off` disables that prerequisite, without
granting credit. This is a workflow guard, not a shell sandbox.

## What A Receipt Never Covers

- A later edit to any input, config, dependency or build environment
  invalidates it; completion order alone never selects a receipt.
- File-only gofmt and package vet do not satisfy a wider repository gate;
  new-findings-only lint does not satisfy full lint.
- Hook tests use `-short` without `-race`. Run build, the required race check,
  broader fix/lint and applicable govulncheck whenever matching evidence is
  absent.
- A host that runs no plugin hooks (opencode, a copied skill) leaves no hook
  evidence; use the normal gate there.

## Report Labels

Hook output with `reuse=unverified` is not a final pass. Without a shell,
report the hook's observed status and scope as
`observed (hook, reuse unverified)`; required gate checks lacking current
matching verification are unavailable. Silence, skipped checks and
unavailable tools are never passes.

A disclaimer does not turn `lint pass (receipt, reuse=unverified)` into an
observed-only status. Write `lint unavailable (receipt not verified)` or
`lint observed (hook, reuse unverified)`, not `lint pass`. The plugin checks
hook-sourced pass claims at Stop against the current verifier credits;
correct unsupported labels rather than repeating them.
