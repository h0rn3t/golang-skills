# Security Review Mode

> Sources: OWASP Go-SCP; OWASP API Security Top 10 (2023); Cloudflare security-audit skill (MIT) for the every-door and finding-gate method
> Authority: advisory
> Last verified: 2026-10-30

When asked to audit, trace **data flow**, not files: start at every input
(`r.Form`, `r.Body`, `r.PathValue`, headers and cookies, `os.Args`,
`os.Getenv`, queue messages, webhook bodies, rows from a shared table, model
output and tool-call arguments) and walk forward to the first sensitive sink.
Read the `gosec` output first, then check every [Quick Reference](../SKILL.md#quick-reference) row marked
`review` against every handler in scope: those rows are the findings the
review exists for.

A finding names the lower-trust caller and what it can already do, the input
it controls, the sink, the missing control, and **whose** resource the effect
reaches: *caller → input → sink → missing control → affected resource*. A
caller acting on its own data with its own rights crosses nothing.

- **Every door**: list each route to the same effect — the legacy route, the
  batch and export endpoints, the background job, the gRPC stream beside the
  unary method. The weakest one is the control; compare the checks, not only
  that each door has one.
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
`verified` means a test you ran shows the effect — two dummy tenants through
`httptest`, the forged input, the wrong answer — never a request to a
running deployment.

> **Validation**: add `go test -fuzz=^FuzzDecode$ -fuzztime=30s` on any
> hand-written parser at a boundary, with that parser's fuzz target in place
> of `FuzzDecode`. Report a skipped check as skipped.
