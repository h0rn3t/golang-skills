# Code Review: [PR Title]

## Summary
[Brief description of the changes]
Net lines: +A / -B. Growth the change did not need is a finding below, not a footnote.

## Findings

Every finding carries how it was established: `verified` names the run that
proved it, `plausible` admits it was only read.

### Must Fix
- [ ] [file:line] Description of critical issue
      Evidence: verified (how) | plausible (not proven)
      Fix: the concrete action

### Should Fix
- [ ] [file:line] Description of recommended improvement
      Evidence: verified (how) | plausible (not proven)
      Fix: the concrete action
- [ ] [file:line] delete: | yagni: | stdlib: | dep: | shrink: what can stop existing
      Fix: the shorter form

### Nits
- [ ] [file:line] Description of minor suggestion

## Automated Checks
- [ ] `gofmt -d .` — clean
- [ ] `go vet ./...` — clean
- [ ] `golangci-lint run` — clean

Report a check that could not run as `unavailable (reason)`, not as clean.

## Not Reviewed
What stayed outside the review and why — tool unavailable, needs a human
decision, needs an external contract.
