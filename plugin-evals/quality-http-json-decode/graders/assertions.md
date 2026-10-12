---
type: llm
weight: 3
---

PASS if the answer satisfies every item:

- Uses http.MaxBytesReader before decoding
- Requires a second decode to return io.EOF (or an equivalent full-body single-document check) before calling Create
- Rejects both a second value and malformed trailing data, while allowing trailing whitespace
- Checks the size limit across the whole accepted document including trailing whitespace
- Does not claim DisallowUnknownFields rejects trailing data or claim tests ran

FAIL if any item is missing or contradicted.
