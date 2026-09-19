---
name: go-security
description: Use when writing or reviewing Go code that touches untrusted input, secrets, or credentials — SQL, shell, or template injection, path traversal, SSRF, cookies and security headers, password hashing, TLS settings, token comparison, or what may appear in a log line. Also use when a handler accepts a file name, URL, or command argument from a client, or when asked to find "vulnerabilities", even if the user never says "security". Does not cover the os.Root and crypto/rand mechanics (see go-defensive) or the dependency CVE scan in the gate (see go-linting).
---

# Go Security

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). `os.Root` and
> `crypto/pbkdf2`, `crypto/hkdf`, `crypto/sha3` are Go 1.24+;
> `http.NewCrossOriginProtection` Go 1.25+; `crypto/subtle`, `html/template`,
> and `net/netip` are long-standing stdlib.

A vulnerability is a trust boundary nobody named. Before writing or reviewing,
answer three questions: **where does untrusted data enter**, **which sensitive
operation does it reach** (query, shell, file path, URL fetch, HTML, log), and
**what is the blast radius if the defense fails**. Every rule below is one of
those paths, with the standard-library defense that closes it.

## Resource Routing

- `references/INJECTION.md` - Read when untrusted data reaches SQL, `os/exec`, a template, a file path, an outbound URL, or a redirect, or is served back as a download.
- `references/SECRETS-AND-CRYPTO.md` - Read when handling passwords, tokens, API keys, TLS configuration, or choosing a hash or cipher.

## Triage: Follow the Data

```
Untrusted value arrives (request, env, file, DB row written by others)
├─ goes into SQL?            → placeholders only (go-database owns the form)
├─ picks a row by ID?        → the caller's tenant in the same WHERE, not a check after the fetch
├─ goes into a command?      → exec.Command(name, args...); never "sh -c"
├─ goes into HTML/JS?        → html/template; never text/template for HTML
├─ names a file?             → os.Root (go-defensive owns the form)
├─ comes back as a download? → Content-Disposition: attachment; nosniff; a Content-Type you chose
├─ becomes an outbound URL?  → hostname allowlist by whole label; block private ranges (SSRF)
├─ becomes a redirect?       → one leading "/", none of "//", "\", control characters
├─ is compared to a secret?  → crypto/subtle.ConstantTimeCompare
├─ is logged?                → redact; LogValuer (go-logging owns the form)
└─ is returned in an error?  → status + generic text; detail stays server-side
```

Stop at the branch that matches and apply that defense **at the boundary**,
once — not at every call site downstream, where it is forgotten.

---

## Quick Reference

| Threat | Defense | Caught by |
|---|---|---|
| SQL injection | `QueryContext(ctx, q, args...)` placeholders; identifiers from a `switch` over known names | `gosec` G201/G202 |
| Foreign row by ID | `WHERE id = $1 AND org_id = $2` with the caller's tenant as a parameter; a foreign ID is `sql.ErrNoRows` | review |
| Command injection | `exec.CommandContext(ctx, "gzip", "--keep", "--", name)`: argv, no shell, `--` before input | `gosec` G204 |
| XSS | `html/template` (contextual escaping) | `gosec` G203 (unsafe `template.HTML`) |
| Path traversal | `root.Open(name)` on an `os.Root` opened once at startup | `gosec` G304 |
| Upload served inline | `Content-Disposition: attachment`, `X-Content-Type-Options: nosniff`, a `Content-Type` you derived; or a separate origin | review |
| SSRF | Hostname allowlist by whole label: `host == d` or `strings.HasSuffix(host, "."+d)`; else resolve and reject `netip.Addr.IsPrivate()`/loopback | review |
| Open redirect | One leading `/`; reject `//`, `\`, and control characters; or an allowlist of hosts | review |
| Predictable tokens | `crypto/rand.Text()` / `rand.Read` | `gosec` G404 |
| Timing leak on compare | `subtle.ConstantTimeCompare(a, b) == 1` | review |
| Weak password hash | argon2id (or `crypto/pbkdf2` when stdlib-only) | review — G401 and the G501/G505 import blocklists catch MD5/SHA1 only, not `sha256.Sum256(password)` |
| `InsecureSkipVerify: true` | Never outside a test against a local self-signed server; `RootCAs` for a private CA | `gosec` G402 |
| `MinVersion` | Leave unset (1.2 is the default) unless the service is TLS 1.3-only | `gosec` G402 |
| `CurvePreferences` | No list is the default; a list that omits the ML-KEM hybrids is a finding | review |
| CSRF | `http.NewCrossOriginProtection().Handler(mux)` | review |
| Secret in log | `slog.LogValuer` returning `"[REDACTED]"` | review |
| Known CVE in deps | `govulncheck ./...` (gate) | gate |

`gosec` is in the baseline `.golangci.yml`; a finding it raises is a gate
failure, not advice. See [go-linting](../go-linting/SKILL.md).

---

## Injection

The defense is always the same shape: hand data to an API that knows it is
data. String assembly is what turns data into code.

```go
// ✗ Bad — the shell re-parses ref; "main; rm -rf /" is one argument to sh
out, err := exec.Command("sh", "-c", "git log "+ref).Output()

// ✓ Good — argv, no shell; option parsing ends before ref, so "--output=x" stays a revision
out, err := exec.CommandContext(ctx, "git", "log", "--end-of-options", ref, "--").Output()
```

`--end-of-options` is git's spelling, because `--` already separates
revisions from paths there; every other program takes `--` before the first
argument built from input, as in the Quick Reference row.

- SQL: placeholders for values; identifiers (table, column, `ORDER BY`) come
  from a `switch` over known names, never from input. A row the caller may
  see only as its owner carries the caller's tenant in the same `WHERE`; a
  check after the fetch is the one the next handler forgets.
  [go-database](../go-database/SKILL.md) owns the query form.
- Templates: `html/template` escapes per context (attribute, URL, JS).
  `template.HTML(userInput)` opts out of that — treat it as a finding.
- Downloads: a stored `Content-Type` is what the uploader sent. Serve the file
  with `Content-Disposition: attachment`, `X-Content-Type-Options: nosniff`,
  and a type you derived, or from a separate origin that holds no session;
  an uploaded `text/html` served inline runs as your site.
- Headers: `net/http` neutralizes CR/LF in header values — a response writer
  turns them into spaces, the server rejects them in a request, `Transport`
  refuses to send them — so header injection is closed. A `Location` built
  from input still enables an open redirect: require one leading `/` and
  reject `//`, `\`, and control characters
  ([Local redirects](references/INJECTION.md#local-redirects)), or use an
  allowlist of hosts.

Full patterns, including the hostname allowlist and the `net/netip` range
check for SSRF, in [INJECTION.md](references/INJECTION.md).

---

## Secrets

- Read secrets from the environment or a mounted file at startup; never from
  a literal, a flag default, or a committed config. Fail fast when missing.
- Compare tokens and MACs with `subtle.ConstantTimeCompare`; `==` on a secret
  leaks its prefix length through timing.
- Verify a signed token the client presents (JWT, a signed cookie) with a
  fixed algorithm allowlist and your key; `alg`, `kid`, or `jku` inside the
  token never choose either, and `exp` is checked against the server clock.
- Hash passwords with a memory-hard KDF (argon2id from `golang.org/x/crypto`,
  the maintained Go-team module); when the dependency ladder forbids it,
  `crypto/pbkdf2` (Go 1.24+) with a high iteration count is the stdlib floor.
  Never a bare `sha256.Sum256(password)`.
- Keep secrets out of logs and errors: wrap the type in a `slog.LogValuer`
  ([go-logging](../go-logging/SKILL.md)) and never `fmt.Errorf("auth %s: %w",
  token, err)`.

Key derivation, TLS defaults, and cookie flags in
[SECRETS-AND-CRYPTO.md](references/SECRETS-AND-CRYPTO.md).

---

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

- Error responses carry a status and a generic message; the wrapped error with
  file paths, SQL, or hostnames goes to the log, not the client.
- `net/http/pprof` and `expvar` mount on a separate internal listener, never on
  the public mux — they leak heap contents and goroutine stacks.
- `X-Forwarded-For` is client-writable and append-only: read its **last**
  entry, and only when a known proxy set it; otherwise `r.RemoteAddr`.

---

## Review Mode

When asked to audit, trace **data flow**, not files: start at every input
(`r.Form`, `r.Body`, `os.Args`, `os.Getenv`, rows from a shared table) and walk
forward to the first sensitive sink. Report each finding as
*input → sink → missing defense → severity*. Severity by blast radius: remote
code execution and credential theft first, data exposure second, denial of
service third. An issue with no reachable input is reported as `not reachable`,
not dropped — a `sh -c` on a constant string is ugly, not a vulnerability, and
the reader decides whether it stays. Inside a
[go-code-review](../go-code-review/SKILL.md) pass, each finding goes into that
skill's template with its `verified`/`plausible` marker; blast radius orders
the Must Fix list, it does not replace the sections.

> **Validation**: `golangci-lint run --enable-only gosec ./...` for the
> mechanical findings, `govulncheck ./...` for dependencies, and `go test
> -fuzz=^FuzzDecode$ -fuzztime=30s` on any hand-written parser at a boundary,
> with that parser's fuzz target in place of `FuzzDecode`. Report a skipped
> check as skipped.

Restraint never cuts a security control: the restraint ladder in
[go-code-refactor](../go-code-refactor/SKILL.md) names them as the code that
has to exist. Prove a check unnecessary, or leave it and say why.

---

## Related Skills

- [go-defensive](../go-defensive/SKILL.md): `os.Root`, `crypto/rand`, boundary copies, `Must` at init.
- [go-database](../go-database/SKILL.md): placeholder queries, identifier allowlists.
- [go-http](../go-http/SKILL.md): body limits, timeouts, CSRF protection, error-to-status mapping.
- [go-logging](../go-logging/SKILL.md): `LogValuer` and what never reaches a log line.
- [go-linting](../go-linting/SKILL.md): `gosec` in the baseline, `govulncheck` in the gate.
- [go-code-review](../go-code-review/SKILL.md): the report template a security finding lands in during a review pass.
- [go-troubleshooting](../go-troubleshooting/SKILL.md): a crash or hang rather than a known vulnerability.
