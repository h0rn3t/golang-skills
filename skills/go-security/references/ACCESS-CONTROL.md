# Access Control and Identity Binding

> Sources: `net/http`, `crypto/tls` package docs; `google.golang.org/grpc` v1.84 and `golang.org/x/oauth2` package docs; OWASP API Security Top 10 (2023); Cloudflare security-audit skill (MIT) for the every-door and finding-gate method
> Authority: normative for the stdlib forms; advisory for the gRPC and OAuth module forms
> Minimum Go: 1.22 for method patterns in `ServeMux`; everything else long-standing
> Last verified: 2026-10-19

Authentication says who is calling; everything here is the second question:
may **this** principal do **this** action to **this** resource, on every path
that reaches the effect.

## Contents

- [Authenticate the whole mux](#authenticate-the-whole-mux)
- [The tenant the check saw](#the-tenant-the-check-saw)
- [Every door to the same effect](#every-door-to-the-same-effect)
- [Caches and derived copies](#caches-and-derived-copies)
- [Model output and tool calls](#model-output-and-tool-calls)
- [Service identity](#service-identity)
- [OAuth and OIDC clients](#oauth-and-oidc-clients)

---

## Authenticate the whole mux

`ServeMux` has no route groups, so wrapping each handler leaves public the
next route someone registers. Wrap the mux once and list the public patterns
in the wrapper. `mux.Handler(r)` reports the pattern the request will match,
so the list names exact registered patterns, not path prefixes. The fragment
below takes `SessionStore`, `ErrNoSession`, and `withPrincipal` from the
service:

```go
func requireAuth(mux *http.ServeMux, sessions SessionStore, public ...string) http.Handler {
    open := make(map[string]bool, len(public)) // registered patterns: "GET /healthz"
    for _, p := range public {
        open[p] = true
    }
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if _, pattern := mux.Handler(r); open[pattern] {
            mux.ServeHTTP(w, r)
            return
        }
        c, err := r.Cookie("__Host-session")
        if err != nil {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        p, err := sessions.Lookup(r.Context(), c.Value)
        switch {
        case errors.Is(err, ErrNoSession):
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        case err != nil: // the store is down: deny, never continue as anonymous
            http.Error(w, "unavailable", http.StatusServiceUnavailable)
            return
        }
        mux.ServeHTTP(w, r.WithContext(withPrincipal(r.Context(), p)))
    })
}
```

The server's `Handler` is `requireAuth(mux, sessions, "GET /healthz")`: a
route registered later is private until someone lists it. Every branch of an auth check ends in a response and `return`, or in the next
handler with a principal. The shapes that fail open:

- `if err != nil { log…; }` with no `return`, and the request continues: the
  log line is the only trace.
- An empty allowlist read as "no restriction":
  `if len(allowed) > 0 && !allowed[role]` lets every role in when the list is
  empty — a typo in config disables the check.
- A missing environment variable or a feature flag that switches the check
  off. Refuse to start instead.
- A verifier error (the key set did not load, a certificate expired) that
  falls back to the unverified claims or to bearer-only.
- Authentication without authorization: the wrapper proves who; each handler
  that loads a resource still checks that this principal may do this action
  to it.

---

## The tenant the check saw

[INJECTION.md](INJECTION.md#sql) has the query form (`AND org_id = $2`). The
value in that parameter comes from the principal the auth wrapper put in the
context, never from anything the client sends:

- The request type holds no `OrgID`, `UserID`, or `Role`
  ([decoding row](INJECTION.md#decoding-untrusted-structures)).
- When the path names the tenant (`/orgs/{org}/…`), the middleware that
  checked membership in `r.PathValue("org")` and the handler that queries use
  that same value. A handler that reads `body.OrgID`, a query parameter, or a
  second path segment acts in a tenant nobody checked.
- An ID from the client picks a row inside the tenant; it never picks the
  tenant.
- A foreign ID answers exactly like a missing one: a `403` for a row that
  exists elsewhere confirms it exists.

---

## Every door to the same effect

The effective control is the weakest route to the effect. For each sensitive
operation, list:

- **Batch, bulk, export, import**: authorize every item after selection, not
  the request once. The finished export is a separate resource: its download
  link is a bearer token with an expiry and one owner.
- **The legacy route, the admin route, the method-override form**: the same
  check, not a weaker one written earlier.
- **Background jobs and queue consumers** run without a request principal.
  The tenant comes from the job record an authorized request wrote, and the
  job re-checks a membership that can be revoked between enqueue and run.
- **Restore, undelete, re-import**: authorize against today's ACLs, not the
  ones in the snapshot.
- **Check-then-act** on a quota, balance, coupon, or one-time invite: two
  concurrent requests both pass the read. One conditional
  `UPDATE … WHERE remaining >= $1` that reports the rows it changed, or a
  locked row ([go-database](../../go-database/SKILL.md#transactions) owns the
  form). `go test -race` sees no race between two transactions.

---

## Caches and derived copies

- An in-process map, an LRU, or a Redis key built from an ID alone serves one
  tenant's entry to another. Key on a struct, `struct{ org, id string }`, or
  join with a separator neither part can contain.
- A cached principal or authorization decision outlives revocation: bound its
  TTL by the revocation delay the product promises, or drop it on role change.
- Soft delete: every path — lookup, list, search, export, job — carries the
  `deleted_at IS NULL` predicate. One raw query that skips the view is the
  finding.
- A search index or analytics copy carries the tenant and filters on it at
  query time, not only at ingestion.

---

## Model output and tool calls

A model's output and a tool call's arguments are untrusted input, like a
request body, whoever wrote the prompt: a retrieved document or a fetched page
steers them.

- Decode tool arguments into a typed request struct that rejects unknown
  fields. Schema validation proves the shape and nothing else.
- The tool handler authorizes as the user the agent acts for, on the resource
  the arguments name. An agent that calls with a service credential the
  handler trusts is a confused deputy: it does for anyone what only some may.
- An approval binds the exact tool name and argument object the user saw. A
  retry or a later turn with changed arguments needs a new approval.
- Arguments that reach SQL, `os/exec`, a file path, or a URL get the defenses
  in [INJECTION.md](INJECTION.md); a model in between changes nothing.
- Retrieval and memory filter by the requesting user's tenant in the query.
  Content written to memory from a document is data in the next session, never
  an instruction.
- A per-request budget on tool calls and spend; the limiter is
  [go-resilience](../../go-resilience/SKILL.md)'s.

---

## Service identity

**mTLS.** With `ClientAuth: tls.RequireAndVerifyClientCert` and `ClientCAs`
set, the caller is `r.TLS.VerifiedChains[0][0]`. Map it to a principal by an
exact field (a URI SAN, a SPIFFE ID) through a fixed table, never by a
substring of the subject. Under `RequestClientCert` or `RequireAnyClientCert`,
`PeerCertificates` is whatever the client sent, unverified. A tenant in the
request body never overrides the certificate's.

When a proxy terminates TLS and forwards the identity in a header, the backend
trusts that header only if it accepts connections from that proxy alone and
the proxy overwrites the client's copy. Neither is visible in the service's
source: the finding is `plausible` and names that fact.

**gRPC.** `grpc.UnaryInterceptor` and `ChainUnaryInterceptor` cover unary
methods only; without `ChainStreamInterceptor` every streaming method is
open. Match the public allowlist on `info.FullMethod` exactly. A stream
authorized when it opens still checks each message that names a different
resource. `reflection.Register` lists every method: register it on an
internal server. The peer's verified chain is
`p.AuthInfo.(credentials.TLSInfo).State.VerifiedChains` after
`peer.FromContext(ctx)`.

**Queues and events.** A message claiming to be an admin event or a provider
callback is authenticated by a signature or by the broker identity of its
producer, never by a `type` field in the body. A tenant in the payload is not
isolation: the consumer takes it from the subscription the broker
authenticated, or checks the payload against it. Dead-letter queues and error
topics carry whole payloads: the redaction rules for logs apply there.
Duplicate delivery and replay are
[go-resilience](../../go-resilience/SKILL.md)'s idempotency.

---

## OAuth and OIDC clients

With `golang.org/x/oauth2`, each login attempt gets a fresh `state` and PKCE
verifier, kept server-side or in a `__Host-` cookie. The fragment below shows
the two calls on either side of the redirect:

```go
state, verifier := rand.Text(), oauth2.GenerateVerifier()
http.Redirect(w, r, conf.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier)), http.StatusFound)

// callback, after r.FormValue("state") equals the stored state, which is then deleted
tok, err := conf.Exchange(ctx, r.FormValue("code"), oauth2.VerifierOption(verifier))
```

- The callback rejects a `state` that is not the one stored for this browser,
  and uses it once.
- An ID token is verified against the provider's keys, `iss`, `aud` equal to
  your client ID, `exp`, and the `nonce` you sent; a token decoded without
  verification identifies nobody.
- With several providers, `state` also records which one was started, and
  the callback rejects a code from another.
- Linking a provider to an existing account needs a current session and the
  provider's verified subject; merging accounts on an unverified email hands
  one to whoever registers that address first.
- The post-login redirect is a [local redirect](INJECTION.md#local-redirects).
