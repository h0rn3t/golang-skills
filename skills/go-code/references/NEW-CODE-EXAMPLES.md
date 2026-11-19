# New Code Examples

> Sources: project policy ([go-code](../SKILL.md#writing-new-code)); `go doc encoding/json/v2`
> Authority: advisory
> Last verified: 2026-09-19 against go1.27.1

Use these examples alongside the [parent skill](../SKILL.md).

## Contents

- [A Contract Table](#a-contract-table)
- [A Plain Code Body](#a-plain-code-body)
- [A Name Used Once](#a-name-used-once)
- [A Helper That Earns Its Line](#a-helper-that-earns-its-line)
- [A Business Rule Chain](#a-business-rule-chain)

## A Contract Table

One case per observable clause, the class clause taken from the member a
library default treats unlike the rest:

| Clause | Case | Expected |
|---|---|---|
| "an unknown id is a 404" | `GET /items/nope` | 404 |
| "returns the matching entries as a JSON array" | zero matches | `[]`, not `null` |
| "any other value is an error naming the parameter" | `limit=abc` | error text contains `limit` |
| "any method other than `GET` is a 405" | `HEAD /healthz` | 405, not the `GET` body |
| "largest names the largest listed file" | one file `empty.txt` of size 0 | `largest: "empty.txt"`, `total: 0` |

For the method clause, a `GET /healthz` pattern on
`ServeMux` answers `HEAD` with 200 and the `GET` body, so the class "any other
method" is tested at `HEAD`, not at `POST`. The same reasoning picks `t` and
`1` under `strconv.ParseBool` and the bare path under a subtree pattern.

As a test file the table is one case struct, one table, one `t.Run` loop, and
a failure message in the `Func(input) = got, want` form;
[go-testing](../../go-testing/SKILL.md) owns the table-test form.

## A Plain Code Body

The function below is the whole implementation of a documented JSON document,
with `import json "encoding/json/v2"`: a function-local type projects the
wire fields from `File` into an anonymous document. The empty-input rule and
the one tracked fact (`largest`, the size
behind `doc.Largest`) are visible in the code. v2 writes the nil
`Files` of a build with no files as `[]`, so that clause is a case in the
contract test rather than a line in the body. Under `encoding/json` v1 the
same clause costs `Files: []entry{}` in the literal plus the one comment that
marks the overridden default. The same document as two package-level types, a
constructor for the entry, and a `writeJSON` helper is the growth the
[Declaration Budget](../SKILL.md#declaration-budget) counts.

Its loop collects entries, totals sizes, and tracks the largest file in one
pass. Replacing it with several collection calls would make those related
state changes harder to follow.

```go
func Manifest(build string, files []File) ([]byte, error) {
	if build == "" {
		return nil, errors.New("manifest: build is empty")
	}
	type entry struct {
		Path string `json:"path"`
		Size int64  `json:"size"`
	}
	doc := struct {
		Build   string  `json:"build"`
		Files   []entry `json:"files"`
		Largest string  `json:"largest"`
		Total   int64   `json:"total"`
	}{Build: build}
	var largest int64
	for _, f := range files {
		if f.Path == "" {
			continue
		}
		doc.Files = append(doc.Files, entry{Path: f.Path, Size: f.Size})
		doc.Total += f.Size
		if doc.Largest == "" || f.Size > largest {
			largest = f.Size
			doc.Largest = f.Path
		}
	}
	return json.Marshal(doc)
}
```

## A Name Used Once

With `time` imported, `expiresAt` names the boundary the comparison checks.
Inlining it saves a line but makes that boundary less visible:

```go
func expired(now, issuedAt time.Time, ttl time.Duration) bool {
	expiresAt := issuedAt.Add(ttl)
	return !now.Before(expiresAt)
}
```

## A Helper That Earns Its Line

Two handlers write the same JSON document, so the step has two call sites in
the diff and the report names both:

```text
added package-level declarations: 1 — writeJSON: handleList, handleGet
```

Written as `writeJSON := func(w http.ResponseWriter, v any) {...}` inside the
constructor it captures nothing, serves the same two call sites, and is one
more thing a reader has to scroll past to find the routes. It is the same
declaration; the budget counts it in either position, and the package-level
form is the one a reviewer accepts.

A single caller can also justify a helper when it separates substantial work
at another level of abstraction. A repository method may call
`listQuery(filter)` once when that function assembles the SQL text, the
allow-listed sort column, and the placeholder arguments, and the method keeps
the domain rule. Request decoding is not such a step: bound, decode, and
validate stay in the handler, as [go-http](../../go-http/SKILL.md#handler-shape)
writes them. A helper that only forwards one call adds a name to follow
without hiding meaningful detail.

## A Business Rule Chain

In this entry-point fragment, `planZoneChanges` validates every snapshot and
computes changes in period order. Binding clears an uncovered period, while
unbinding leaves it as it is. A bad snapshot leaves the store untouched.

```go
	interval, ok, err := b.processingInterval()
	if err != nil || !ok {
		return false, err
	}
	snapshots, err := b.loadSnapshots(interval)
	if err != nil {
		return false, err
	}
	changes, err := planZoneChanges(interval, snapshots, data)
	if err != nil {
		return false, err
	}
	for _, change := range changes {
		if err := b.setZone(change); err != nil {
			return false, err
		}
	}
	return true, nil
```
