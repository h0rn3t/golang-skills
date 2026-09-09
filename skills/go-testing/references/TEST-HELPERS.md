# Test Helpers, Assertions, and Comparisons

Detailed reference for writing test helpers, choosing assertion style, and
choosing between t.Error and t.Fatal.
Sources: Google Go Style Guide, Uber Go Style Guide.

---

## Test Helper Pattern

Test helpers must call `t.Helper()` first so failures point to the caller.
Use `t.Fatal` for setup failures, and `t.Cleanup` for teardown.

```go
func mustLoadTestData(t *testing.T, filename string) []byte {
    t.Helper()
    data, err := os.ReadFile(filename)
    if err != nil {
        t.Fatalf("Setup failed: could not read %s: %v", filename, err)
    }
    return data
}

func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("Could not open database: %v", err)
    }
    t.Cleanup(func() { db.Close() })
    return db
}
```

**Key rules:**
- Call `t.Helper()` as the first statement to attribute failures to the caller
- Use `t.Fatal` for setup failures (don't return errors from helpers)
- Use `t.Cleanup()` for teardown instead of defer — it runs even if the test
  calls `t.FailNow`

---

## Assertion Style

Follow [the assertion policy](../SKILL.md#assertions-match-the-repository):
`testify/assert` and `testify/require` are allowed in existing and new projects.
Honor the user's choice and repository conventions; without either, use
standard comparisons and `cmp.Diff`. Do not rewrite working tests just to
switch assertion libraries.

With testify, use `require` when the next operation depends on success and
`assert` for independent checks. `require` calls `FailNow`, so it belongs only
in the test goroutine. Use semantic helpers such as `ErrorIs` for wrapped
errors and enable `testifylint`.

```go
// Imports: github.com/stretchr/testify/assert and .../require
got, err := GetPost(id)
require.NoError(t, err, "GetPost(%q)", id)
require.NotNil(t, got, "GetPost(%q)", id)
assert.Equal(t, "blogPost", got.Type, "GetPost(%q).Type", id)
assert.Equal(t, 2, got.Comments, "GetPost(%q).Comments", id)
```

For a project using standard comparisons and `cmp`:

```go
want := BlogPost{
    Type:     "blogPost",
    Comments: 2,
    Body:     "Hello, world!",
}
if diff := cmp.Diff(want, got); diff != "" {
    t.Errorf("GetPost() mismatch (-want +got):\n%s", diff)
}
```

### Domain-Specific Comparisons

For domain-specific comparisons, return values or errors instead of calling
`t.Error`:

```go
func postLength(p BlogPost) int { return len(p.Body) }

func TestBlogPost(t *testing.T) {
    post := BlogPost{Body: "Hello"}
    if got, want := postLength(post), 5; got != want {
        t.Errorf("postLength(post) = %v, want %v", got, want)
    }
}
```

---

## Comparisons and Diffs

Prefer `cmp.Equal` and `cmp.Diff` for complex types. Always include the
direction key `(-want +got)` in diff messages.

```go
// Struct comparison
want := &Doc{Type: "blogPost", Authors: []string{"isaac", "albert"}}
if diff := cmp.Diff(want, got); diff != "" {
    t.Errorf("AddPost() mismatch (-want +got):\n%s", diff)
}

// Protocol buffers
if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
    t.Errorf("Foo() mismatch (-want +got):\n%s", diff)
}
```

**Avoid unstable comparisons** — don't compare JSON/serialized output that may
change. Compare semantically instead.

---

## t.Error vs t.Fatal: Detailed Guidance

Use `t.Error` to keep tests going and report all failures in a single run:

```go
// Good: Report all mismatches
if diff := cmp.Diff(wantMean, gotMean); diff != "" {
    t.Errorf("Mean mismatch (-want +got):\n%s", diff)
}
if diff := cmp.Diff(wantVariance, gotVariance); diff != "" {
    t.Errorf("Variance mismatch (-want +got):\n%s", diff)
}
```

Use `t.Fatal` when subsequent checks would be meaningless:

```go
gotEncoded := Encode(input)
if gotEncoded != wantEncoded {
    t.Fatalf("Encode(%q) = %q, want %q", input, gotEncoded, wantEncoded)
}
gotDecoded, err := Decode(gotEncoded)
if err != nil {
    t.Fatalf("Decode(%q) error: %v", gotEncoded, err)
}
```

### Don't Call t.Fatal from Goroutines

> **Normative**: Never call `t.Fatal`, `t.Fatalf`, or `t.FailNow` from a
> goroutine other than the test goroutine. Use `t.Error` instead and let the
> goroutine return naturally.
