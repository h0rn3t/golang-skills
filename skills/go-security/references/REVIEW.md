# Security Review Mode

> Sources: OWASP Go-SCP, OWASP API Security Top 10 (2023), Cloudflare security-audit skill (MIT; every-door and finding-gate method), [Claude-BugHunter triage-validation](https://github.com/elementalsouls/Claude-BugHunter/blob/main/skills/triage-validation/SKILL.md) and [evidence-hygiene](https://github.com/elementalsouls/Claude-BugHunter/blob/main/skills/evidence-hygiene/SKILL.md) (controls and evidence handling)
> Authority: advisory
> Last verified: 2026-11-19

When asked to audit, trace **data flow**, not files: start at every input
(`r.Form`, `r.Body`, `r.PathValue`, headers and cookies, `os.Args`,
`os.Getenv`, queue messages, webhook bodies, rows from a shared table, model
output and tool-call arguments) and walk forward to the first sensitive sink.
Read the `gosec` output first, then check every [Quick Reference](../SKILL.md#quick-reference) row marked
`review` against every handler in scope: those rows are the findings the
review exists for. With no `gosec` run (no shell, no linter), every row is a
review row, and the report says `gosec` did not run.

A finding names the lower-trust caller and what it can already do, the input
it controls, the sink, the missing control, and **whose** resource the effect
reaches: *caller → input → sink → missing control → affected resource*. A
caller acting on its own data with its own rights crosses nothing.

- **Every door**: list each route to the same effect — the legacy route, the
  batch and export endpoints, the background job, the gRPC stream beside the
  unary method. The weakest one is the control; compare the checks, not only
  that each door has one.
- **Trace back from the effect**: for a sensitive read, write, or disclosure,
  start at the operation and follow its callers back to the entry points.
  At each hop, check where the principal, tenant, and resource ID come from
  and whether the operation uses the same values the authorization checked.
  Include deferred work and cached or exported copies that can return the
  same data. Check retrieval separately when it has its own access check.
- **Variants**: after a finding, search the package for the same shape and
  report each reachable copy; a fix in one caller leaves its siblings open.
- **Hardening, not a finding**: a missing layer behind a working one (no
  `SameSite` on a cookie a CSRF check already guards) is one hardening line,
  not Must Fix. With no reachable input it is `not reachable`, not dropped: a
  `sh -c` on a constant string is ugly, and the reader decides whether it stays.
- **Outside the repository**: whether the proxy overwrites `X-Forwarded-For`,
  the ingress authenticates, or the IdP pins an audience is never assumed
  either way. The finding is `plausible`, names the one fact an owner checks,
  and is ranked by what follows if that fact holds.
- **A comment that says why a path is safe** is a claim: verify it.

Rank by what the evidence shows, never above it:

| Rank | The finding demonstrates |
|---|---|
| critical | Code execution, a whole-store read, or any account taken over, without authentication |
| high | An explicit control fully defeated: auth bypass, another tenant's read or write, stored script run by other users, authenticated code execution, one request ending the shared process (an input-reachable panic in a goroutine) |
| medium | A real boundary crossed with a narrow blast radius or uncommon preconditions |
| low | Non-secret internals disclosed |

Fully defeats or only weakens decides high against medium. A handler panic
`net/http` recovers ends one connection, not the service. Inside a
[go-code-review](../../go-code-review/SKILL.md) pass, each finding goes into that
skill's template with its `verified`/`plausible` marker: a crossing you can
name is Must Fix, ordered by rank; low and hardening lines are Should Fix.
`verified` means a test you ran locally shows the effect, never a request to
a running deployment.

## Check an implementation

Before editing a trust boundary, identify the caller, protected operation or
data, and required control. In new code's Contract Table, cover allowed and
denied behavior for new or changed access paths and rejection for new or
changed input controls. Reuse cases that exercise the control through the
affected entry point.

Review the final diff and the paths it affects. Follow the applicable review
rules above, including alternate entry points and error paths. At the sensitive
operation, verify that the code uses the principal, tenant, and resource that
the authorization checked. With a shell, run applicable cases with the gate
and use the checks below to verify the effect. Fix confirmed security defects
introduced by the task and recheck. When tests cannot run, report what was checked in
source and which results remain unverified.

## Validate the effect

- For access control, compare an allowed case with a denied case for the
  same operation on the same dummy resource. Keep the input and resource
  fixed and vary the caller, membership, or permission under review. For
  HTTP, use `httptest` with the registered router and its middleware so the
  test exercises the same checks as the application.
- Read the order of parsing, authorization, and the business operation. A
  `400` can come from a parser before authorization runs. Use well-formed
  input when checking access. Check the protected data returned for reads
  and the protected state after a write attempt. A status code alone does
  not establish a finding. Mark a source-grounded candidate `plausible`
  when a named condition still needs verification.
- Report the observed outcome, the relevant source locations, and any
  control you could only inspect in source. Use dummy records in local tests
  and redact secrets and personal data from any evidence in the review.

> **Validation**: add `go test -fuzz=^FuzzDecode$ -fuzztime=30s` on any
> hand-written parser at a boundary, with that parser's fuzz target in place
> of `FuzzDecode`. Report a skipped check as skipped.
