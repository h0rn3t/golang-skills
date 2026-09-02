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

- `references/INJECTION.md` - Read when untrusted data reaches SQL, `os/exec`, a template, a file path, an outbound URL, or a response header.
- `references/SECRETS-AND-CRYPTO.md` - Read when handling passwords, tokens, API keys, TLS configuration, or choosing a hash or cipher.

## Triage: Follow the Data

```
Untrusted value arrives (request, env, file, DB row written by others)
├─ goes into SQL?          → placeholders only (go-database owns the form)
├─ goes into a command?    → exec.Command(name, args...); never "sh -c"
├─ goes into HTML/JS?      → html/template; never text/template for HTML
├─ names a file?           → os.Root (go-defensive owns the form)
├─ becomes an outbound URL?→ allowlist host; block private ranges (SSRF)
├─ is compared to a secret?→ crypto/subtle.ConstantTimeCompare
├─ is logged?              → redact; LogValuer (go-logging owns the form)
└─ is returned in an error?→ status + generic text; detail stays server-side
```

Stop at the branch that matches and apply that defense **at the boundary**,
once — not at every call site downstream, where it is forgotten.

---

## Quick Reference

| Threat | Defense | Caught by |
|---|---|---|
| SQL injection | `QueryContext(ctx, q, args...)` placeholders | `gosec` G201/G202 |
| Command injection | `exec.Command("git", "log", ref)` — args, no shell | `gosec` G204 |
| XSS | `html/template` (contextual escaping) | `gosec` G203 (unsafe `template.HTML`) |
| Path traversal | `os.OpenRoot(dir).Open(name)` | `gosec` G304 |
| SSRF | Resolve host, reject `netip.Addr.IsPrivate()`/loopback, allowlist | review |
| Predictable tokens | `crypto/rand.Text()` / `rand.Read` | `gosec` G404 |
| Timing leak on compare | `subtle.ConstantTimeCompare(a, b) == 1` | review |
| Weak password hash | argon2id (or `crypto/pbkdf2` when stdlib-only) | `gosec` G401/G501 |
| Plain TLS | `MinVersion: tls.VersionTLS12`; never `InsecureSkipVerify` | `gosec` G402 |
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

// ✓ Good — ref is one argv element; the shell never sees it
out, err := exec.CommandContext(ctx, "git", "log", "--", ref).Output()
```

- SQL: placeholders for values; identifiers (table, column, `ORDER BY`) come
  from a `switch` over known names, never from input.
  [go-database](../go-database/SKILL.md) owns the query form.
- Templates: `html/template` escapes per context (attribute, URL, JS).
  `template.HTML(userInput)` opts out of that — treat it as a finding.
- Headers: `net/http` rejects CR/LF in header values, so header injection is
  closed; a `Location` built from input still enables open redirects —
  accept relative paths or an allowlist of hosts.

Full patterns, including SSRF checks with `net/netip`, in
[INJECTION.md](references/INJECTION.md).

---

## Secrets

- Read secrets from the environment or a mounted file at startup; never from
  a literal, a flag default, or a committed config. Fail fast when missing.
- Compare tokens and MACs with `subtle.ConstantTimeCompare`; `==` on a secret
  leaks its prefix length through timing.
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
    Name: "session", Value: id,
    HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode,
    Path: "/", MaxAge: 3600,
})
```

- Error responses carry a status and a generic message; the wrapped error with
  file paths, SQL, or hostnames goes to the log, not the client.
- `net/http/pprof` and `expvar` mount on a separate internal listener, never on
  the public mux — they leak heap contents and goroutine stacks.
- Trust `X-Forwarded-For` only from a known proxy; otherwise use `r.RemoteAddr`.

---

## Review Mode

When asked to audit, trace **data flow**, not files: start at every input
(`r.Form`, `r.Body`, `os.Args`, `os.Getenv`, rows from a shared table) and walk
forward to the first sensitive sink. Report each finding as
*input → sink → missing defense → severity*. Severity by blast radius: remote
code execution and credential theft first, data exposure second, denial of
service third. Skip theoretical issues with no reachable input — a `sh -c` on
a constant string is ugly, not a vulnerability.

> **Validation**: `golangci-lint run --enable-only gosec ./...` for the
> mechanical findings, `govulncheck ./...` for dependencies, and `go test
> -fuzz=FuzzParse -fuzztime=30s` on any hand-written parser at a boundary. Report
> a skipped check as skipped.

Restraint never cuts a security control: the restraint ladder in
[go-code-refactor](../go-code-refactor/SKILL.md) names them as the code that
has to exist. Prove a check unnecessary, or leave it and say why.

---

## Related Skills

- **Boundary mechanics**: See [go-defensive](../go-defensive/SKILL.md) for the `os.Root` and `crypto/rand` forms, boundary copies, and `Must` at init
- **Queries**: See [go-database](../go-database/SKILL.md) for placeholder queries and identifier allowlists
- **Handlers and servers**: See [go-http](../go-http/SKILL.md) for body limits, timeouts, CSRF protection, and error-to-status mapping
- **Redaction**: See [go-logging](../go-logging/SKILL.md) for `LogValuer` and what never goes in a log line
- **Gate**: See [go-linting](../go-linting/SKILL.md) for `gosec` in the baseline config and `govulncheck` in the gate
- **Root cause of an incident**: See [go-troubleshooting](../go-troubleshooting/SKILL.md) when the symptom is a crash or hang rather than a known vulnerability
