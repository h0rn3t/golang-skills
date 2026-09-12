# New Code Examples

> Sources: project policy ([go-code](../SKILL.md#writing-new-code)); docs/evidence/2026-09-11-go-implement-newcode-workflow-opus-5-medium.md
> Authority: advisory
> Last verified: 2026-09-12

Read when the shape of a Contract Table case or a Plain Code body is in
doubt. The rules live in the parent skill; this file only shows them applied
once.

## A Contract Table

One case per observable clause, the class clause taken from the member a
library default treats unlike the rest:

| Clause | Case | Expected |
|---|---|---|
| "an unknown id is a 404" | `GET /items/nope` | 404 |
| "returns the matching entries as a JSON array" | zero matches | `[]`, not `null` |
| "any other value is an error naming the parameter" | `limit=abc` | error text contains `limit` |
| "any method other than `GET` is a 405" | `HEAD /healthz` | 405, not the `GET` body |

The last row is the one a default leaks through: a `GET /healthz` pattern on
`ServeMux` answers `HEAD` with 200 and the `GET` body, so the class "any other
method" is tested at `HEAD`, not at `POST`. The same reasoning picks `t` and
`1` under `strconv.ParseBool` and the bare path under a subtree pattern.

As a test file the table is one case struct, one table, one `t.Run` loop, and
a failure message in the `Func(input) = got, want` form;
[go-testing](../../go-testing/SKILL.md) owns the table-test form.

## A Plain Code Body

The function below is the whole implementation of a documented JSON document:
one function-local type, one anonymous document, the empty-input rule and the
one tracked fact as code. The same document as two package-level types, a
constructor for the entry, and a `writeJSON` helper is the growth the
[Declaration Budget](../SKILL.md#declaration-budget) counts.

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
	}{Build: build, Files: []entry{}} // [] for a build with no files, never null
	var largest int64 // the size behind doc.Largest: a tracked fact, not a value used once
	for _, f := range files {
		if f.Path == "" {
			continue
		}
		doc.Files = append(doc.Files, entry{Path: f.Path, Size: f.Size})
		doc.Total += f.Size
		if f.Size > largest {
			doc.Largest, largest = f.Path, f.Size
		}
	}
	return json.Marshal(doc)
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
