# Code Review: [PR Title]

Keep the prose short: the findings carry the information, and a paragraph
restating them is noise. No section is padded to look complete — an empty
severity is one line saying it is empty.

## Summary
[Brief description of the changes]
Include line counts only when relevant to the requested review.

## Findings

Each finding identifies an executed check or a static proof of the affected
path. Material unresolved hypotheses name missing evidence separately; they
are not mandatory fixes merely because they are conceivable.

### Must Fix
- [ ] [file:line] Description of critical issue
      Evidence: executed check or static proof (how)
      Fix: the concrete action

### Should Fix
- [ ] [file:line] Description of recommended improvement
      Evidence: executed check or static proof (how)
      Fix: the concrete action
- [ ] [file:line] delete: | yagni: | stdlib: | dep: | shrink: what can stop existing
      Fix: the shorter form

### Nits
- [ ] [file:line] Description of minor suggestion

## Automated Checks
List only checks selected through go-linting: exact command, scope, and result.

Report a check that could not run as `unavailable (reason)`, not as clean.

## Not Reviewed
What stayed outside the review and why — tool unavailable, needs a human
decision, needs an external contract.
