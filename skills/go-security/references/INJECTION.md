# Injection and Untrusted Input

> Sources: `os/exec`, `html/template`, `net/netip`, `net/url` package docs; OWASP Go-SCP
> Authority: normative for the stdlib defenses; advisory for the allowlist shapes
> Minimum Go: 1.24 for `os.Root`; everything else long-standing
> Last verified: 2026-10-07

## Contents

- [SQL](#sql)
- [Command execution](#command-execution)
- [HTML and templates](#html-and-templates)
- [File paths](#file-paths)
- [Outbound URLs (SSRF)](#outbound-urls-ssrf)
- [Redirects and headers](#redirects-and-headers)
- [Decoding untrusted structures](#decoding-untrusted-structures)

---

## SQL

[go-database](../../go-database/SKILL.md) owns the query form. The security
half: placeholders cover **values** only. Anything that is part of the SQL
grammar — table/column identifiers and sort direction — cannot be a value
placeholder and must come from a closed set. The numeric value of `LIMIT`
is a parameter (`LIMIT $2` in PostgreSQL); range-check it before the query:

```go
var sortCols = map[string]string{"created": "created_at", "name": "name"}

col, ok := sortCols[r.URL.Query().Get("sort")]
if !ok {
    col = "created_at"
}
limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
if err != nil || limit < 1 || limit > 100 {
    limit = 50 // range-checked before it reaches the query
}
q := "SELECT id, name FROM users WHERE org_id = $1 ORDER BY " + col + " LIMIT $2"
rows, err := db.QueryContext(ctx, q, orgID, limit) // col is ours; orgID and limit are placeholders
```

`gosec` G201/G202 flag `fmt.Sprintf` and `+` feeding a query function. A
`//nolint:gosec` there needs the allowlist visible in the same function.

Authorization sits on the same boundary. A row the caller reaches by ID is
selected with the caller's tenant in the `WHERE`, so a foreign ID is
`sql.ErrNoRows` rather than a loaded row that a later `if` may forget:

```go
// ✗ Bad — the row is loaded first; the ownership check lives in each handler, until one skips it
row := db.QueryRowContext(ctx, "SELECT org_id, body FROM notes WHERE id = $1", id)

// ✓ Good — ownership is part of the query; a foreign id is sql.ErrNoRows, mapped to 404
row := db.QueryRowContext(ctx, "SELECT body FROM notes WHERE id = $1 AND org_id = $2", id, caller.OrgID)
```

---

## Command execution

`exec.Command` passes each argument as one `argv` element; no shell is
involved, so metacharacters are inert. Injection only appears when a shell is
reintroduced (`sh -c`, `bash -c`, `cmd /C`) or when the **program name** comes
from input.

```go
// ✗ Bad — shell parses the string; input can add commands
exec.Command("sh", "-c", "convert "+name+" out.png")

// ✗ Bad — input chooses the binary
exec.Command(r.FormValue("tool"), "--version")

// ✓ Good — fixed program, input is data, "--" stops option parsing
cmd := exec.CommandContext(ctx, "gzip", "--keep", "--", name) //nolint:gosec // G204: fixed program, argv, "--" before input
cmd.Env = []string{"PATH=/usr/bin"} // do not inherit secrets from os.Environ()
```

`--` protects only against option parsing. A program with its own file-name
grammar — ImageMagick reads `msl:`/`ephemeral:` prefixes, `ffmpeg` reads
protocol prefixes — stays injectable: normalize to `./`+`filepath.Base(name)`
and allow-list the extension before the call.

Also:

- Set `cmd.Dir` explicitly; inherit nothing from the request.
- A path passed to a subprocess escapes `os.Root`: the child resolves the name
  itself, so in an attacker-writable directory a symlink can be swapped in
  between your check and its open.
  [go-defensive](../../go-defensive/SKILL.md#confine-filesystem-access) has
  the form that hands the child the open file instead.

---

## HTML and templates

The template *source* is code. `template.New("").Parse(text)` on text a client
wrote lets it call every exported method reachable from the data
(`{{.Store.DeleteAll}}`) and loop without bound; `html/template` escapes the
output, not the template. A user-editable message keeps the template fixed
and takes named placeholders:

```go
// ✗ Bad — template.Must(template.New("msg").Parse(user.Template)).Execute(w, page)
msg := strings.NewReplacer("{name}", customer.Name, "{order}", orderID).Replace(user.Template)
page.Execute(w, map[string]string{"Message": msg}) // a fixed html/template escapes {{.Message}}
```

The `html/template` escape hatches — `template.HTML`, `template.JS`, `template.URL`,
`template.HTMLAttr` — tell the engine "trust this". Wrapping input in one is
the vulnerability; they exist for content **you** produced (a sanitizer's
output, a constant snippet).

Set `Content-Type: text/html; charset=utf-8` yourself; sniffing turns a JSON
endpoint into an HTML one when a client saves it as `.html`. Add
`X-Content-Type-Options: nosniff` in middleware.

A file the client uploaded and later downloads is input on the way out: its
stored `Content-Type` is whatever the uploader sent, and served inline from
your origin an uploaded `text/html` runs with your cookies. Serve it as an
attachment with a type you derived, or from a separate origin that holds no
session:

```go
w.Header().Set("X-Content-Type-Options", "nosniff")
w.Header().Set("Content-Type", "application/octet-stream") // or http.DetectContentType(head); never the stored one
w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": name}))
```

---

## File paths

[go-defensive](../../go-defensive/SKILL.md#confine-filesystem-access) owns
the `os.Root` form, the archive loop, and `filepath.IsLocal`. The threat-model
half: a lexical check does not confine a path. No lexical check
(`filepath.Clean`, `strings.Contains(p, "..")`, `filepath.IsLocal`) sees a
symlink inside the directory that points outside it. `os.OpenRoot` resolves
every component inside the directory, so a `..` escape and an escaping
symlink return an error instead of a file.

One more sink:

- **Archive extraction** (`archive/zip`, `archive/tar`): the entry name, type,
  and size are attacker-controlled — `../` and absolute names, a symlink entry
  that a later entry writes through, and a 1 KB zip that expands to
  gigabytes. go-defensive's archive loop checks all three.
- **Deletes**: an ID from the path that resolves to the directory itself
  (`""`, `ws/..` from `ws%2F..`, `%2E%2E`) makes `RemoveAll` empty the whole
  tree, inside an `os.Root` too ([go-defensive](../../go-defensive/SKILL.md#deletes-name-what-they-remove)).

---

## Outbound URLs (SSRF)

A URL from a client that the server fetches is a request the attacker sends
from **inside** your network: cloud metadata endpoints (`169.254.169.254`),
`localhost` admin ports, internal services with no auth. `url.Parse` validates
syntax, not intent.

Where the set of legitimate hosts is known, an **allowlist of hostnames** is
the defense. It matches whole labels, and every URL the handler fetches goes
through it — a redirect target, and a second URL derived from the same record,
are the same input:

```go
func allowedHost(host, domain string) bool {
    return host == domain || strings.HasSuffix(host, "."+domain)
}

// partnerClient re-checks every redirect: an open redirect on an allowed host
// would otherwise reach any address.
func partnerClient(domain string) *http.Client {
    return &http.Client{
        Timeout: 10 * time.Second,
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            if len(via) >= 10 { // a custom CheckRedirect replaces the default 10-hop cap
                return fmt.Errorf("stopped after %d redirects", len(via))
            }
            if !allowedHost(req.URL.Hostname(), domain) {
                return fmt.Errorf("redirect to %s: host not allowlisted", req.URL.Host)
            }
            return nil
        },
    }
}
```

`strings.HasSuffix(host, domain)` without the dot accepts `evilpartner.example`
for `partner.example`; `strings.Contains` accepts it anywhere in the name.

### Arbitrary public destinations

When any public host is legitimate — webhooks, link previews — check the
address the client dials. `net.Dialer.Control` runs after DNS resolution and
before the connect, for every connection the client opens, so a rebinding
answer and a redirect to `127.0.0.1` fail the same check:

```go
// blockedPrefixes are the special-purpose ranges the netip.Addr methods below
// miss; the NAT64 and 6to4 prefixes embed an IPv4 address, metadata included.
var blockedPrefixes = []netip.Prefix{
    netip.MustParsePrefix("0.0.0.0/8"),
    netip.MustParsePrefix("100.64.0.0/10"), // CGNAT; 100.100.100.200 is Alibaba Cloud metadata
    netip.MustParsePrefix("192.0.0.0/24"),
    netip.MustParsePrefix("198.18.0.0/15"),
    netip.MustParsePrefix("64:ff9b::/96"),
    netip.MustParsePrefix("64:ff9b:1::/48"),
    netip.MustParsePrefix("2002::/16"),
}

func publicClient() *http.Client {
    dialer := &net.Dialer{
        Timeout: 5 * time.Second,
        Control: func(_, address string, _ syscall.RawConn) error {
            ap, err := netip.ParseAddrPort(address)
            if err != nil {
                return err
            }
            a := ap.Addr().Unmap().WithZone("") // Prefix.Contains never matches a zoned address
            if a.IsPrivate() || a.IsLoopback() || a.IsLinkLocalUnicast() || a.IsUnspecified() || a.IsMulticast() ||
                slices.ContainsFunc(blockedPrefixes, func(p netip.Prefix) bool { return p.Contains(a) }) {
                return fmt.Errorf("dial %s: blocked address range", address)
            }
            return nil
        },
    }
    return &http.Client{Transport: &http.Transport{DialContext: dialer.DialContext}, Timeout: 10 * time.Second}
}
```

Resolving and checking the hostname before the request instead leaves a gap:
the dial resolves again, and the attacker's DNS can answer differently the
second time. A new `http.Transport` sets no proxy; a proxy would make the
check see the proxy's address, not the target's.

---

## Redirects and headers

`net/http` rejects CR and LF in header values, so classic header injection is
closed. What remains:

- **Host header**: absolute URLs built from `r.Host` (password-reset links)
  follow whatever the client sent. Use a configured canonical host.

### Local redirects

Require a single leading slash and reject backslashes, spaces, and ASCII
control characters. Browsers interpret `/\host` and `/` + TAB + `/host` as
external destinations; checking only for a `//` prefix misses both.

```go
func localRedirect(next string) bool {
    return strings.HasPrefix(next, "/") && !strings.HasPrefix(next, "//") &&
        !strings.ContainsFunc(next, func(r rune) bool {
            return r == '\\' || r <= ' ' || r == '\x7f'
        })
}
```

Check the exact value passed to `http.Redirect`; do not decode or rebuild it
after validation. Percent-encoded spaces may remain encoded. Test rejected
browser-normalized destinations alongside valid local paths, queries, and
fragments. See the [WHATWG URL parser](https://url.spec.whatwg.org/#relative-slash-state).

---

## Decoding untrusted structures

Decoders are parsers running on attacker bytes; bound them.

| Input | Bound |
|---|---|
| JSON | Require one complete bounded document; use the [HTTP decoding rules](../../go-http/SKILL.md#handler-shape) or [JSON v2 example](../../go-http/references/JSON-V2.md#one-bounded-request-document) |
| XML | a size cap before `xml.Unmarshal`; `encoding/xml` expands no entity a DTD declares (`&b;` is a syntax error), so billion-laughs does not apply |
| Regex on input | RE2 is linear — Go's `regexp` is safe; a third-party PCRE engine is not |
| `strconv.Atoi` into a size | range-check before `make([]T, n)` |
| Compressed input (`Content-Encoding: gzip`, a `.gz` upload) | `MaxBytesReader` caps the compressed bytes only; read the `gzip.Reader` through `io.LimitReader(zr, limit+1)` and reject at `limit+1` bytes |
| Body of an outbound fetch | `io.LimitReader(resp.Body, limit)`: the SSRF check bounds where the client connects, not how much it reads |
| Multipart upload | Cap the body before `ParseMultipartForm`; `maxMemory` only sets the memory/disk threshold. Check file sizes and count; use `MultipartReader` with per-part limits when early rejection matters |
| Struct target | Decode into a request type holding only the client-writable fields, never into the storage model — a `Role`, `OrgID`, or `IsAdmin` field on it is set by whoever sends the body |

The response is the mirror image: `json.Marshal` writes every exported field,
so a storage record encoded as the response ships the `PasswordHash`,
`ResetToken`, or `OrgID` added to it later. Encode a response type that names
the public fields, as the [go-http handler](../../go-http/SKILL.md#handler-shape)
does, never the storage model.

`encoding/gob` and any format that instantiates types from the wire must
never see untrusted bytes — that is deserialization RCE in other languages and
a denial-of-service vector in Go.
