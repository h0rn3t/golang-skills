---
name: go-security
description: Use when writing or reviewing Go code that touches untrusted input, secrets, or credentials — SQL, shell, or template injection, path traversal, SSRF, cookies and security headers, password hashing, TLS settings, token comparison, or what may appear in a log line. Also use when a handler accepts a file name, URL, or command argument from a client, or when asked to find "vulnerabilities", even if the user never says "security". Does not cover the os.Root and crypto/rand mechanics (see go-defensive) or the dependency CVE scan in the gate (see go-linting).
---

# Go Security

> Compatibility: Baseline Go 1.27 (see the plugin repository's COMPATIBILITY.md). `os.Root` and
> `crypto/pbkdf2`, `crypto/hkdf`, `crypto/sha3` are Go 1.24+;
> `http.NewCrossOriginProtection` Go 1.25+; `crypto/subtle`, `html/template`,
> and `net/netip` are long-standing stdlib.

A vulnerability is a trust boundary nobody named. Before writing or reviewing,
answer three questions: **where does untrusted data enter**, **which sensitive
operation does it reach** (query, shell, file path, URL fetch, HTML, log), and
**what is the blast radius if the defense fails**. Every rule below is one of
those paths, with the standard-library defense that closes it.

## Resource Routing

- `references/INJECTION.md` - Read when untrusted data reaches SQL, `os/exec`, a template or a template's source, a file path, an outbound URL, a redirect, or a decompressor, or is served back as a download.
- `references/SECRETS-AND-CRYPTO.md` - Read when handling passwords, a login or reset flow, tokens, API keys, webhook signatures, TLS configuration, or choosing a hash or cipher.
- `references/ACCESS-CONTROL.md` - Read when code decides who may act: auth middleware and route registration, the tenant a query uses, batch, export, import, or background jobs, a cache, model tool calls, mTLS, gRPC, or queue identity, or an OAuth login.

## Triage: Follow the Data

```
Untrusted value arrives (request, env, file, DB row written by others)
├─ goes into SQL?            → placeholders only (go-database owns the form)
├─ picks a row by ID?        → the caller's tenant in the same WHERE, not a check after the fetch
├─ names a tenant?           → only the one the auth middleware checked; never a body field
├─ lists many items?         → authorize each item, not the request
├─ is a cache key?           → the tenant in the key beside the ID
├─ came from a model?        → a request body: validate and authorize in the tool handler
├─ goes into a command?      → exec.Command(name, args...); never "sh -c"
├─ goes into HTML/JS?        → html/template; never text/template for HTML
├─ is a template's source?   → never Parse it; fixed placeholders the user fills
├─ names a file?             → os.Root (go-defensive owns the form)
├─ is compressed?            → cap the decompressed bytes; MaxBytesReader sees only the compressed ones
├─ comes back as a download? → Content-Disposition: attachment; nosniff; a Content-Type you chose
├─ becomes an outbound URL?  → hostname allowlist by whole label; block private ranges (SSRF)
├─ becomes a redirect?       → one leading "/", none of "//", "\", control characters
├─ is compared to a secret?  → crypto/subtle.ConstantTimeCompare
├─ carries a signature?      → HMAC over the raw bytes and timestamp, hmac.Equal, a tolerance window
├─ is an Origin for CORS?    → exact-origin allowlist; never reflect it, never "null"
├─ is a login attempt?       → unknown user and wrong password cost one hash and get one answer
├─ is logged?                → a secret type that redacts in every sink (go-logging owns the form)
└─ is returned in an error?  → status + generic text; detail stays server-side
```

Stop at the branch that matches and apply that defense **at the boundary**,
once — not at every call site downstream, where it is forgotten.

## Quick Reference

| Threat | Defense | Caught by |
|---|---|---|
| SQL injection | `QueryContext(ctx, q, args...)` placeholders; identifiers from a `switch` over known names | `gosec` G201/G202 |
| Foreign row by ID | `AND org_id =` the caller's tenant, a parameter beside the ID ([INJECTION.md](references/INJECTION.md#sql)); a foreign ID is `sql.ErrNoRows` | review |
| Tenant from the client | The tenant the middleware checked is the one the query uses; no `OrgID` in the request type ([ACCESS-CONTROL.md](references/ACCESS-CONTROL.md#the-tenant-the-check-saw)) | review |
| Route outside the auth wrapper | Wrap the whole mux once and list the public patterns in the wrapper ([ACCESS-CONTROL.md](references/ACCESS-CONTROL.md#authenticate-the-whole-mux)) | review |
| Auth that fails open | Every branch answers and returns: a store or verifier error is `503`, never anonymous; an empty allowlist denies | review |
| Batch, export, import, background job | Authorize each item as the requesting principal when the effect happens ([ACCESS-CONTROL.md](references/ACCESS-CONTROL.md#every-door-to-the-same-effect)) | review |
| Cache shared across tenants | The tenant in the key: `struct{ org, id string }`, not the ID alone | review |
| Quota, balance, or one-time code checked then spent | One conditional `UPDATE … WHERE` or a locked row ([go-database](../go-database/SKILL.md#transactions)) | review — `-race` sees no race between two transactions |
| Model output, tool-call arguments | Untrusted input; the tool handler authorizes as the user the agent acts for ([ACCESS-CONTROL.md](references/ACCESS-CONTROL.md#model-output-and-tool-calls)) | review |
| Command injection | `exec.CommandContext(ctx, "gzip", "--keep", "--", name)`: argv, no shell, `--` before input | `gosec` G204, which fires on this safe form too: suppress it on that line with the reason, as in [Injection](#injection) |
| XSS | `html/template` (contextual escaping) | `gosec` G203 (unsafe `template.HTML`) |
| Template written by a client | Never `Parse` client text: it calls any exported method on the data; fixed placeholders ([INJECTION.md](references/INJECTION.md#html-and-templates)) | review |
| Path traversal | `root.Open(name)` on an `os.Root` opened once at startup ([go-defensive](../go-defensive/SKILL.md#confine-filesystem-access) owns the form) | review — `gosec` G304 fires on every variable path, so the bundled config excludes it |
| Upload served inline | `Content-Disposition: attachment`, `X-Content-Type-Options: nosniff`, a `Content-Type` you derived; or a separate origin | review |
| SSRF | Hostname allowlist by whole label: `host == d` or `strings.HasSuffix(host, "."+d)`, re-checked on every redirect in `CheckRedirect`; else a `net.Dialer.Control` that rejects, on the dialed address, `netip.Addr.IsPrivate()`/loopback and the CGNAT and NAT64 prefixes those methods miss ([INJECTION.md](references/INJECTION.md#arbitrary-public-destinations)) | review |
| Decompression bomb | `io.LimitReader` on the `gzip.Reader`, not only `MaxBytesReader` on the body; the same cap on a fetched response body | review |
| Open redirect | One leading `/`; reject `//`, `\`, and control characters; or an allowlist of hosts | review |
| Predictable tokens | `crypto/rand.Text()` / `rand.Read` | `gosec` G404 |
| Timing leak on compare | `subtle.ConstantTimeCompare(a, b) == 1` | review |
| Webhook signature | HMAC over the raw body bytes and the signed timestamp, `hmac.Equal`, a tolerance window, decode only after ([SECRETS-AND-CRYPTO.md](references/SECRETS-AND-CRYPTO.md#inbound-webhook-signatures)) | review |
| Account enumeration at login | A dummy hash for an unknown user, one error for both failures, a semaphore on concurrent hashes ([SECRETS-AND-CRYPTO.md](references/SECRETS-AND-CRYPTO.md#passwords)) | review |
| Weak password hash | argon2id (or `crypto/pbkdf2` when stdlib-only) | review — G401 and the G501/G505 import blocklists catch MD5/SHA1 only, not `sha256.Sum256(password)` |
| `InsecureSkipVerify: true` | Never outside a test against a local self-signed server; `RootCAs` for a private CA | `gosec` G402 |
| `MinVersion` | Leave unset (1.2 is the default) unless the service is TLS 1.3-only | `gosec` G402 |
| `CurvePreferences` | No list: the default has `X25519MLKEM768`, `SecP256r1MLKEM768`, and `SecP384r1MLKEM1024`. A list that omits any is a finding, fixed by deleting it, not reordering it (order is ignored); only opt-in `MLKEM1024` needs a list ([SECRETS-AND-CRYPTO.md](references/SECRETS-AND-CRYPTO.md#tls)) | review |
| CSRF | `http.NewCrossOriginProtection().Handler(mux)` | review |
| CORS with credentials | Exact-origin allowlist, `Vary: Origin`; the same list in `AddTrustedOrigin` ([HTTP Surface](#http-surface)) | review |
| Secret in a log, `fmt`, or JSON | The secret type in [go-logging](../go-logging/SKILL.md#what-not-to-log), which covers an exported struct field; a struct with the secret in an unexported field prints itself without it (same section) | review |
| Known CVE in deps | `govulncheck ./...` (gate) | gate |

`gosec` is in the baseline `.golangci.yml`; a finding it raises is a gate
failure, not advice. See [go-linting](../go-linting/SKILL.md). No tool covers
a row whose last column says `review`: a clean `gosec` run says nothing about
those rows.

## Injection

The defense is always the same shape: hand data to an API that knows it is
data. String assembly is what turns data into code.

```go
// ✗ Bad — the shell re-parses ref; "main; rm -rf /" is one argument to sh
out, err := exec.Command("sh", "-c", "git log "+ref).Output()

// ✓ Good — argv, no shell; option parsing ends before ref, so "--output=x" stays a revision
out, err := exec.CommandContext(ctx, "git", "log", "--end-of-options", ref, "--").Output() //nolint:gosec // G204: fixed program, argv, options end before ref
```

`--end-of-options` is git's spelling, because `--` already separates
revisions from paths there; every other program takes `--` before the first
argument built from input, as in the Quick Reference row.

## Secrets

- Read secrets from the environment or a mounted file at startup; never from
  a literal, a flag default, or a committed config. Fail fast when missing.
- Verify a signed token the client presents (JWT, a signed cookie) with a
  fixed algorithm allowlist and keys you already hold: `kid` only chooses
  among those keys, and `alg`, `jku`, or `x5u` inside the token never choose
  either. Reject a token without `exp`, check `exp` against the server clock,
  and check `iss` and `aud` against fixed values.
- Keep secrets out of errors: never `fmt.Errorf("auth %s: %w", token, err)`.

Key derivation, TLS defaults, and cookie flags in
[SECRETS-AND-CRYPTO.md](references/SECRETS-AND-CRYPTO.md).

## HTTP Surface

[go-http](../go-http/SKILL.md) owns the server construction; the security
items it carries are `ReadHeaderTimeout`, `MaxBytesReader` on every decoded
body, `MaxHeaderValueCount`, and `http.NewCrossOriginProtection`. Add here:

```go
http.SetCookie(w, &http.Cookie{
    Name: "__Host-session", Value: id, // the prefix: browsers enforce Secure, Path=/, no Domain
    HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode,
    Path: "/", MaxAge: 3600,
})
```

CORS relaxes the same-origin policy, so the allowlist is the whole defense.
Reflecting `Origin` with `Allow-Credentials` lets every site read every
response with the user's cookies, and `"null"` is the origin of a sandboxed
iframe any site can create. A credentialed cross-origin `POST` that CORS
allows is still rejected by `CrossOriginProtection` unless the same origin is
trusted there:

```go
func withCORS(next http.Handler, origins ...string) (http.Handler, error) {
    cop := http.NewCrossOriginProtection()
    allowed := make(map[string]struct{}, len(origins)) // exact "https://app.example.com"; no suffix match
    for _, o := range origins {
        if err := cop.AddTrustedOrigin(o); err != nil {
            return nil, err
        }
        allowed[o] = struct{}{}
    }
    return cop.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Add("Vary", "Origin") // a cache must not serve one origin's answer to another
        origin := r.Header.Get("Origin")
        if _, ok := allowed[origin]; !ok {
            next.ServeHTTP(w, r) // no CORS headers: the browser withholds the response
            return
        }
        w.Header().Set("Access-Control-Allow-Origin", origin)
        w.Header().Set("Access-Control-Allow-Credentials", "true")
        if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST") // every method and request header the mux serves
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next.ServeHTTP(w, r)
    })), nil
}
```

- `net/http/pprof` and `expvar` register on `http.DefaultServeMux` in `init`,
  so one blank import anywhere in the binary publishes them on every server
  whose `Handler` is nil. Serve the public listener from your own mux and
  mount the debug handlers on a separate internal listener — they leak heap
  contents and goroutine stacks.
- `X-Forwarded-For` is client-writable and append-only: join every
  `r.Header.Values("X-Forwarded-For")` line and take the rightmost entry that
  is not one of your own proxies; `r.Header.Get` returns only the first line,
  which the client wrote when a proxy adds its own. With no proxy in front,
  `r.RemoteAddr`.

## Review Mode

When asked to audit, trace **data flow**, not files: start at every input
(`r.Form`, `r.Body`, `r.PathValue`, headers and cookies, `os.Args`,
`os.Getenv`, queue messages, webhook bodies, rows from a shared table, model
output and tool-call arguments) and walk forward to the first sensitive sink.
Read the `gosec` output first, then check every Quick Reference row marked
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
[go-code-review](../go-code-review/SKILL.md) pass, each finding goes into that
skill's template with its `verified`/`plausible` marker: a crossing you can
name is Must Fix, ordered by rank; low and hardening lines are Should Fix.
`verified` means a test you ran shows the effect — two dummy tenants through
`httptest`, the forged input, the wrong answer — never a request to a
running deployment.

> **Validation**: add `go test -fuzz=^FuzzDecode$ -fuzztime=30s` on any
> hand-written parser at a boundary, with that parser's fuzz target in place
> of `FuzzDecode`. Report a skipped check as skipped.

## Related Skills

- [go-database](../go-database/SKILL.md): placeholder queries, identifier allowlists.
- [go-logging](../go-logging/SKILL.md): `LogValuer` and what never reaches a log line.
- [go-resilience](../go-resilience/SKILL.md): rate limits, duplicate delivery, and idempotency keys.
- [go-troubleshooting](../go-troubleshooting/SKILL.md): a crash or hang rather than a known vulnerability.
