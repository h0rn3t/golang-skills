# JSON v2 at API Boundaries

> Sources: https://go.dev/doc/jsonv2-migration; `go doc encoding/json/v2`; `go doc encoding/json.DefaultOptionsV1`
> Authority: advisory for API choice; package documentation for semantics
> Minimum Go: 1.27
> Last verified: 2026-10-01 against go1.27.1

Use `encoding/json/v2` for new code when its semantics fit the contract.
Preserve existing wire formats when refactoring. An import change can compile
while changing accepted inputs or output bytes.

## Choose the I/O API

| Need | v2 API | Condition |
|---|---|---|
| Encode to bytes | `json.Marshal(v)` | Returns the buffer and error before writing a response |
| Encode to a writer | `json.MarshalWrite(w, v)` | Adds no newline; unlike v1 `Encoder.Encode` |
| Decode one complete byte buffer | `json.Unmarshal(b, &v)` | Rejects a second value or trailing junk |
| Decode one complete reader | `json.UnmarshalRead(r, &v)` | Success requires EOF; bound untrusted readers |
| Process a sequence of documents | `json.UnmarshalDecode(dec, &v)` with `jsontext.Decoder` | Retain the decoder between values; do not substitute `UnmarshalRead` |

Writer APIs can leave partial output on error. Buffer with `Marshal` when the
handler must still be able to choose an error status. `DefaultOptionsV1()`
does not restore `Encoder.Encode`'s newline or single-`Decode` stream consumption.

## One bounded request document

With `import json "encoding/json/v2"`, a new endpoint can reject unknown
members, additional documents, and oversized bodies in one decode. This example
also rejects top-level `null`; the pointer distinguishes it from an object:

```go
r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
var req *struct {
    Name string `json:"name"`
}
if err := json.UnmarshalRead(r.Body, &req, json.RejectUnknownMembers(true)); err != nil || req == nil {
    http.Error(w, "invalid JSON body", http.StatusBadRequest)
    return
}
```

Validate required fields and domain rules before side effects; `{}` still
decodes successfully. A second EOF decode is unnecessary with `UnmarshalRead`.
Keep request deadlines: waiting for EOF is part of this API's contract.

Test valid input with trailing whitespace, exact and exceeded size limits,
empty input, `null`, unknown and wrong-case names, duplicate names, invalid
UTF-8, a second document, and trailing junk. For client response limits, use
the complete-buffer approach in [go-http](../SKILL.md#bounded-response-bodies).

## Defaults that can change the contract

| Area | v2 default | Compatibility option when required |
|---|---|---|
| Nil non-byte slices and maps | `[]` and `{}` | `FormatNilSliceAsNull(true)`, `FormatNilMapAsNull(true)` |
| Struct field matching | Case-sensitive | `MatchCaseInsensitiveNames(true)` with `jsonv1.MatchCaseSensitiveDelimiter(true)`: the first alone also ignores `_` and `-`, so `user_id` fills `UserID`, which v1 rejects |
| Unknown object members | Ignored | `RejectUnknownMembers(true)` to reject them |
| Duplicate names and invalid UTF-8 | Rejected | `jsontext.AllowDuplicateNames(true)`, `jsontext.AllowInvalidUTF8(true)`, only when the existing contract accepts them |
| Map order | Unspecified | `Deterministic(true)` for sorted keys |
| `omitempty` | Omits empty JSON values, not Go zero numbers/bools | Review tags; use `omitzero` when Go zero values are intended |
| `time.Duration` | No default representation: marshal and unmarshal fail with a `SemanticError` | `jsonv1.FormatDurationAsNano(true)` restores v1 nanoseconds; a new field takes a wire form from [go-defensive](../../go-defensive/references/TIME-ENUMS-TAGS.md#json-fields) |
| `string` option on a string or bool field | Runtime `SemanticError` (invalid use of `string` tag option) | `jsonv1.StringifyWithLegacySemantics(true)`, or drop the option |
| `<`, `>`, `&` in strings | Written as is; v1 writes `\u003c`, `\u003e`, `\u0026` | `jsontext.EscapeForHTML(true)` |
| `MarshalJSON` on a pointer receiver | Called on non-addressable values too, where v1 skipped it | `jsonv1.CallMethodsWithLegacySemantics(true)` |
| `[N]byte` arrays | Base64 string; v1 writes an array of numbers | `jsonv1.FormatByteArrayAsArray(true)` |

`go doc encoding/json` ("Migrating to v2") lists every difference and its option.

For staged migration, import `jsonv1 "encoding/json"` and
`json "encoding/json/v2"`:

```go
b, err := json.Marshal(v, jsonv1.DefaultOptionsV1())
```

This preserves v1 `Marshal` semantics. The equivalent option also works with
`Unmarshal`. Later options override earlier ones, so opt into one behavior at a
time, for example `json.FormatNilSliceAsNull(false)` after `DefaultOptionsV1()`.
Keep compatibility tests for nil/empty, field matching, tags, escaping, and
custom marshalers before changing defaults. Go 1.27's v1 backend can itself
change error wording; test error semantics unless wording is the contract.

Use the released `embed` tag, not the experimental `inline` spelling. Do not
copy experimental examples using removed `format`/`unknown` tag options,
`DiscardUnknownMembers`, or `SkipFunc`.

## JSON in golden tests

Compare decoded values when formatting and order are irrelevant. When exact
bytes are the contract, compare bytes and explicitly preserve newlines,
escaping, and nil/empty representation. `json.Deterministic(true)` sorts map
keys but does not promise identical bytes across toolchain versions, builds,
or custom marshalers that ignore the option; it is not canonical JSON.
