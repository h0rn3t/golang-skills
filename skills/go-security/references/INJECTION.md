# Injection and Untrusted Input

> Sources: `os/exec`, `html/template`, `net/netip`, `net/url` package docs; OWASP Go-SCP
> Authority: normative for the stdlib defenses; advisory for the allowlist shapes
> Minimum Go: 1.24 for `os.Root`; everything else long-standing
> Last verified: 2026-09-19

Every section is the same story: untrusted bytes reach an interpreter (SQL,
shell, HTML, the filesystem, a URL fetcher) as **code** instead of **data**.
The fix is an API that keeps the two apart, applied once at the boundary.

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
cmd := exec.CommandContext(ctx, "gzip", "--keep", "--", name)
cmd.Env = []string{"PATH=/usr/bin"} // do not inherit secrets from os.Environ()
```

`--` protects only against option parsing. A program with its own file-name
grammar — ImageMagick reads `msl:`/`ephemeral:` prefixes, `ffmpeg` reads
protocol prefixes — stays injectable: normalize to `./`+`filepath.Base(name)`
and allow-list the extension before the call.

Also:

- Set `cmd.Dir` explicitly; inherit nothing from the request.
- `os.Root` confines operations performed through that root, not a subprocess.
  Checking a path and then passing its name to another program leaves a
  symlink-swap race when the directory is attacker-writable. Open through the
  root and pass the open file as `cmd.Stdin` or an inherited descriptor when
  supported; retain it until the child exits. Programs requiring paths need
  a trusted staging directory or appropriate OS isolation. `cmd.Dir` alone
  does not confine filesystem access.
- Use `CommandContext` so a hung child dies with the request.

---

## HTML and templates

`html/template` is `text/template` plus contextual escaping: it knows whether
`{{.}}` sits in text, an attribute, a URL, or a script block and escapes for
that context. Rendering HTML with `text/template` or `fmt.Fprintf(w, "<p>%s</p>",
name)` is XSS.

The escape hatches — `template.HTML`, `template.JS`, `template.URL`,
`template.HTMLAttr` — tell the engine "trust this". Wrapping input in one is
the vulnerability; they exist for content **you** produced (a sanitizer's
output, a constant snippet).

```go
// ✗ Bad — sanitization bypassed
data := map[string]any{"Bio": template.HTML(user.Bio)}

// ✓ Good — escaped for the context it lands in
data := map[string]any{"Bio": user.Bio}
```

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

[go-defensive](../../go-defensive/SKILL.md) owns the `os.Root` form. The
threat-model half: `filepath.Clean`, `strings.Contains(p, "..")`, and
`filepath.Join(base, p)` all fail against symlinks that point outside `base`,
against `..` after a symlink, and against absolute paths on Windows.
`os.OpenRoot` resolves every component inside the directory, so those cases
return an error instead of a file.

Two more sinks the root does not cover:

- **Archive extraction** (`archive/zip`, `archive/tar`): entry names are
  attacker-controlled. Open the destination through `root.Create(hdr.Name)`
  and cap the total decompressed size — a 1 KB zip can expand to gigabytes.
- **Temp files**: `os.CreateTemp(dir, pattern)` — never build the name
  yourself; a predictable name in a shared `/tmp` is a symlink race.

---

## Outbound URLs (SSRF)

A URL from a client that the server fetches is a request the attacker sends
from **inside** your network: cloud metadata endpoints (`169.254.169.254`),
`localhost` admin ports, internal services with no auth. `url.Parse` validates
syntax, not intent.

```go
func safeTarget(ctx context.Context, raw string) (*url.URL, error) {
    u, err := url.Parse(raw)
    if err != nil || (u.Scheme != "https" && u.Scheme != "http") {
        return nil, fmt.Errorf("unsupported url %q", raw)
    }
    addrs, err := net.DefaultResolver.LookupNetIP(ctx, "ip", u.Hostname())
    if err != nil {
        return nil, err
    }
    for _, a := range addrs {
        a = a.Unmap()
        if a.IsPrivate() || a.IsLoopback() || a.IsLinkLocalUnicast() ||
            a.IsUnspecified() || a.IsMulticast() {
            return nil, fmt.Errorf("target %s resolves to a blocked range", u.Host)
        }
    }
    return u, nil
}
```

The resolve-then-connect gap (DNS rebinding) is real: pin the checked address
by setting `http.Transport.DialContext` to dial the vetted IP, or accept the
residual risk explicitly. Disable redirects (`CheckRedirect` returning
`http.ErrUseLastResponse`) or re-validate each hop; a public URL that 302s to
`127.0.0.1` defeats the check.

Where the set of legitimate hosts is known, an **allowlist of hostnames**
replaces all of this and is the better default. It matches whole labels, and
every URL the handler fetches goes through it — a second URL derived from the
same record is the same input:

```go
func allowedHost(host, domain string) bool {
    return host == domain || strings.HasSuffix(host, "."+domain) // "evilpartner.example" fails for "partner.example"
}
```

`strings.HasSuffix(host, domain)` without the dot accepts `evilpartner.example`
for `partner.example`; `strings.Contains` accepts it anywhere in the name.

---

## Redirects and headers

`net/http` rejects CR and LF in header values, so classic header injection is
closed. What remains:

- **Open redirect**: `http.Redirect(w, r, r.FormValue("next"), 302)` sends
  users to an attacker's site from your domain. Accept local paths using the
  check below, or use an explicit allowlist when external destinations are needed.
- **Host header**: absolute URLs built from `r.Host` (password-reset links)
  follow whatever the client sent. Use a configured canonical host.
- **`X-Forwarded-For`**: append-only and client-writable. Trust the *last*
  hop only when a known proxy set it; otherwise `r.RemoteAddr` is the truth.

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
| Request body | `http.MaxBytesReader(w, r.Body, limit)` before any decode |
| JSON | Require one complete bounded document; use the [HTTP decoding rules](../../go-http/SKILL.md#handler-shape) or [JSON v2 example](../../go-http/references/JSON-V2.md#one-bounded-request-document) |
| XML | a size cap before `xml.Unmarshal`; `encoding/xml` expands no entity a DTD declares (`&b;` is a syntax error), so billion-laughs does not apply |
| Regex on input | RE2 is linear — Go's `regexp` is safe; a third-party PCRE engine is not |
| `strconv.Atoi` into a size | range-check before `make([]T, n)` |
| Multipart upload | Cap the body before `ParseMultipartForm`; `maxMemory` only sets the memory/disk threshold. Check file sizes and count; use `MultipartReader` with per-part limits when early rejection matters |
| Struct target | Decode into a request type holding only the client-writable fields, never into the storage model — a `Role`, `OrgID`, or `IsAdmin` field on it is set by whoever sends the body |

`encoding/gob` and any format that instantiates types from the wire must
never see untrusted bytes — that is deserialization RCE in other languages and
a denial-of-service vector in Go.
