# Secrets, Credentials, and Cryptography

> Sources: `crypto/*`, `crypto/tls`, `crypto/subtle`, `net/http` package docs; go.dev/blog/fips140; OWASP Password Storage Cheat Sheet
> Authority: normative for stdlib API choices; project policy for the argon2 default
> Minimum Go: 1.24 for `crypto/pbkdf2`, `crypto/hkdf`, `crypto/sha3`, `crypto/mlkem`, `rand.Text`
> Last verified: 2026-09-02

Rule zero: **do not invent cryptography**. Every primitive below is a stdlib
or Go-team-maintained call. The job is choosing the right one and handling the
key material around it.

## Contents

- [Where secrets live](#where-secrets-live)
- [Comparing secrets](#comparing-secrets)
- [Passwords](#passwords)
- [Tokens, IDs, and nonces](#tokens-ids-and-nonces)
- [Encryption at rest](#encryption-at-rest)
- [TLS](#tls)
- [Cookies and sessions](#cookies-and-sessions)
- [Keeping secrets out of output](#keeping-secrets-out-of-output)
- [Deprecated and replaced](#deprecated-and-replaced)

---

## Where secrets live

Environment variables or a mounted file, read once at startup, absent from
the binary and the repository. A flag *default* is a literal in the binary.

```go
func loadConfig() (Config, error) {
    key := os.Getenv("SIGNING_KEY")
    if key == "" {
        return Config{}, errors.New("SIGNING_KEY is required") // fail at boot, not on first use
    }
    ...
}
```

- Prefer a file (`/run/secrets/...`) to an env var where the platform offers
  one: env vars leak through `/proc/<pid>/environ`, crash dumps, and child
  processes (`exec.Command` inherits `os.Environ()` unless `cmd.Env` is set).
- Rotate without redeploy: hold the secret behind a small accessor that
  re-reads on `SIGHUP` or a timer only if the service actually needs it —
  otherwise the restart *is* the rotation.
- A `gitleaks`/`trufflehog` pre-commit hook is cheaper than a rotation.

---

## Comparing secrets

`==` on strings returns at the first differing byte; a remote attacker with
enough samples recovers the prefix. Compare anything secret in constant time:

```go
import "crypto/subtle"

func validToken(got, want []byte) bool {
    return subtle.ConstantTimeCompare(got, want) == 1 // also false when lengths differ
}
```

Applies to API keys, HMAC tags, session IDs, and password-reset tokens. Hash
both sides first (`sha256.Sum256`) when the reference value has variable
length and you want to hide even that.

---

## Passwords

A password hash must be **slow and memory-hard**; SHA-256 (with or without a
salt) is neither and falls to GPU brute force at billions per second.

| Option | Where | When |
|---|---|---|
| `argon2.IDKey` (argon2id) | `golang.org/x/crypto/argon2` | Default. `x/crypto` is Go-team maintained; the module is already indirect in most services |
| `bcrypt.GenerateFromPassword` | `golang.org/x/crypto/bcrypt` | Existing bcrypt stores; 72-byte input cap |
| `pbkdf2.Key` | `crypto/pbkdf2` (Go 1.24+) | Stdlib-only constraint or FIPS 140 mode; ≥600 000 iterations with SHA-256 |

```go
salt := make([]byte, 16)
if _, err := rand.Read(salt); err != nil { return nil, err }
// OWASP baseline: 19 MiB memory, 2 iterations, 1 thread, 32-byte tag
hash := argon2.IDKey([]byte(pw), salt, 2, 19*1024, 1, 32)
```

Store `salt`, the parameters, and `hash` together (the PHC string format
`$argon2id$v=19$m=19456,t=2,p=1$<salt>$<hash>`) so parameters can be raised
later and old hashes re-verified. Verify with `subtle.ConstantTimeCompare`.

---

## Tokens, IDs, and nonces

`crypto/rand` for anything an attacker must not predict; `math/rand/v2` for
everything else (jitter, sampling, shuffles). [go-defensive](../../go-defensive/SKILL.md)
owns the form; the choice table:

| Need | Call |
|---|---|
| URL-safe secret string (session ID, reset token) | `rand.Text()` (Go 1.24+, 128 bits, base32) |
| Raw key material | `rand.Read(buf)` — always check the error |
| UUID | stdlib `uuid` on the dependency ladder ([go-packages](../../go-packages/SKILL.md)) |
| Nonce for AES-GCM | `rand.Read` into a 12-byte buffer, **never reused** with the same key |

A token that must expire carries its expiry server-side (a store lookup) or
inside a signed payload (HMAC) — never trust a client-supplied expiry.

---

## Encryption at rest

Authenticated encryption only. `AES-GCM` (`cipher.NewGCM`) or
`chacha20poly1305` (`x/crypto`); unauthenticated modes (`NewCBCEncrypter`,
`NewCTR`, the deprecated `NewOFB`/`NewCFB*`) let an attacker flip bits
undetected.

```go
block, _ := aes.NewCipher(key) // 32 bytes → AES-256
aead, _ := cipher.NewGCM(block)
nonce := make([]byte, aead.NonceSize())
rand.Read(nonce)
ct := aead.Seal(nonce, nonce, plaintext, additionalData) // nonce prefixed for storage
```

Derive per-purpose keys from one master with `hkdf.Key` (`crypto/hkdf`, Go
1.24+) instead of reusing a key across purposes. GCM's 96-bit nonce limits a
single key to ~2³² messages before the birthday bound matters — rotate or use
`chacha20poly1305.NewX` (192-bit nonce) for high-volume streams.

---

## TLS

The `crypto/tls` defaults are safe: TLS 1.2 minimum, curated cipher suites,
post-quantum key exchange negotiated automatically. Configure only what you
must:

```go
srv := &http.Server{
    TLSConfig: &tls.Config{
        MinVersion: tls.VersionTLS13, // when every client supports it
    },
}
```

- `InsecureSkipVerify: true` outside a test with a local self-signed server is
  a finding, always. For a private CA, set `RootCAs`.
- `GODEBUG=tlsrsakex=1`-style knobs re-enable removed weak options; treat
  their presence in a Dockerfile as a finding.
- mTLS: `ClientAuth: tls.RequireAndVerifyClientCert` with `ClientCAs`; identity
  comes from `r.TLS.PeerCertificates[0]`, never from a header.
- FIPS 140-3: `GODEBUG=fips140=on` switches to the validated module; code
  changes are rarely needed.

---

## Cookies and sessions

```go
http.SetCookie(w, &http.Cookie{
    Name:     "__Host-session", // prefix: Secure, no Domain, Path=/ enforced by browsers
    Value:    sessionID,        // opaque random ID; data lives server-side
    Path:     "/",
    MaxAge:   int((8 * time.Hour).Seconds()),
    HttpOnly: true,             // no document.cookie
    Secure:   true,             // HTTPS only
    SameSite: http.SameSiteLaxMode, // Strict breaks top-level navigations from email links
})
```

- Regenerate the session ID on login and privilege change (fixation).
- Invalidate server-side on logout; clearing the cookie is cosmetic.
- CSRF for state-changing requests: `http.NewCrossOriginProtection` (Go 1.25+)
  in [go-http](../../go-http/SKILL.md); fall back to a per-session token in a
  hidden field when pre-1.25 browsers without `Sec-Fetch-Site` matter.

---

## Keeping secrets out of output

| Sink | Defense |
|---|---|
| `slog` | Implement `slog.LogValuer` on the secret type → `slog.StringValue("[REDACTED]")` ([go-logging](../../go-logging/SKILL.md)) |
| `fmt.Errorf` | Wrap the error, not the credential: `fmt.Errorf("auth for %s: %w", userID, err)` |
| `%v` / `%+v` on a config struct | `String()` method that omits secret fields; or a separate `Redacted()` view |
| Panics and crash dumps | `debug.SetCrashOutput` goes to a file with restricted permissions, not stdout |
| HTTP error body | Generic text; the detailed error is logged with a request ID the client can quote |
| `/debug/pprof`, `expvar` | Internal listener only — heap dumps contain live secrets |

---

## Deprecated and replaced

| Avoid | Use | Since |
|---|---|---|
| `math/rand` for anything secret | `crypto/rand` | always |
| `md5`, `sha1` for integrity or passwords | `sha256`, `sha3` (Go 1.24+) / a KDF | always |
| `cipher.NewOFB`, `NewCFB*` | AEAD (`NewGCM`) or `NewCTR` + HMAC | Go 1.24 deprecates |
| `rsa.EncryptPKCS1v15` for new designs | `rsa.EncryptOAEP` or a KEM (`crypto/mlkem`, `crypto/ecdh`) | Go 1.26 |
| `golang.org/x/crypto/{sha3,hkdf,pbkdf2}` | `crypto/{sha3,hkdf,pbkdf2}` | Go 1.24 |
| `crypto/elliptic` for ECDH | `crypto/ecdh` | Go 1.20 |
| `InsecureSkipVerify` + manual `VerifyPeerCertificate` | `RootCAs` / `ServerName` | always |
