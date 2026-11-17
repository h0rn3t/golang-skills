# Code Review: [PR Title]

Keep the prose short: the findings carry the information, and a paragraph
restating them is noise. No section is padded to look complete — an empty
severity is one line saying it is empty, and a severity the request filtered
out is one line counting its findings.

## Summary
Findings: N (Must Fix X, Should Fix Y, Nits Z) this round. Use
`Nothing to fix.` when N is 0 and no earlier blocker remains open or unverified.
Prior blockers: [finding ID or file:line, fixed | open | unverified | pre-existing, evidence].
Include this line when rechecking prior blockers, including ones on unchanged
code. State what evidence is missing for an unverified blocker.
[Brief description of the changes]
Needs a human: [changed CI workflows, auth, payment, release or installer
paths] — omit the line when the diff touches none; a review covers these but
does not stand in for a person's approval.
Net lines: +A / -B. Growth the change did not need is a finding below, not a footnote.
Verdict: approve | request changes. Request changes while a Must Fix the
change introduced or worsened stands, or a prior blocker is open or
unverified. A Must Fix with undetermined origin blocks until the base
comparison shows it is pre-existing or the defect is fixed.

## Findings

Every finding carries how it was established: `verified` names the run that
proved it, `plausible` admits it was only read. Mark a finding `pre-existing`
after its location when it was present on the base and the change did not
worsen it. Give the comparison evidence. It keeps its severity and
does not block the verdict. With inconclusive evidence use
`origin undetermined` and state what is missing.
Findings are numbered once across all severities, so the author can answer
"fix 2 and 5"; the count does not restart under a heading.

### Must Fix
1. [file:line] Description of critical issue
   Evidence: verified (how) | plausible (not proven)
   Fix: the concrete action

### Should Fix
2. [file:line] Description of recommended improvement
   Evidence: verified (how) | plausible (not proven)
   Fix: the concrete action
3. [file:line] delete: | reuse: | yagni: | stdlib: | dep: | shrink: what can stop existing
   Fix: the shorter form

### Nits
4. [file:line] Description of minor suggestion

## Automated Checks
Gate result per go-linting (PASS/FAIL/INCOMPLETE) with the checks actually run,
e.g. `gate: INCOMPLETE — gofmt pass, build pass, vet pass, test -race pass,
golangci-lint unavailable (not installed)`. A check that could not run is
`unavailable (reason)`, never clean.

## Not Reviewed
What stayed outside the review and why — tool unavailable, needs a human
decision, needs an external contract.
