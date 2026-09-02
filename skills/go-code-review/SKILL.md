---
name: go-code-review
description: Use when reviewing Go code or checking code against community style standards. Also use proactively before submitting a Go PR or when reviewing any Go code changes, even if the user doesn't explicitly request a style review. Does not cover language-specific syntax — delegates to specialized skills.
allowed-tools: Bash(bash:*)
---

# Go Code Review Checklist

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). The HTTP and
> database rows below route to [go-http](../go-http/SKILL.md) and
> [go-database](../go-database/SKILL.md), which carry their own version notes.

## Resource Routing

- `assets/review-template.md` - Use when formatting review output with Must Fix, Should Fix, and Nits sections.
- `scripts/pre-review.sh` - Run before manual review to collect gofmt, go vet, and golangci-lint results.
- `../go-http/references/WEB-SERVER.md` - Read when reviewing an HTTP server that combines concurrency, context, logging, error handling, and shutdown behavior.

## Review Procedure

> Use `assets/review-template.md` when formatting the output of a code review to ensure consistent structure with Must Fix / Should Fix / Nits severity grouping.

1. Run the mechanical gate first — `bash scripts/pre-review.sh ./...` plus
   `go fix -diff ./...`. Never spend review attention on what a tool reports.
2. Read the diff file-by-file; for each file, check the categories below in order
3. Flag issues with specific line references and the rule name
4. After reviewing all files, re-read flagged items to verify they're genuine issues
5. Summarize findings grouped by severity (must-fix, should-fix, nit)

> **Validation**: Re-read the diff once more and delete every finding you cannot
> justify with a specific line reference. A review that reports a fabricated
> issue costs more trust than one that misses a real one.

**Report honestly**: name the checks you actually ran. If `golangci-lint` was
not installed or the tests were not run, say so — do not present a partial
review as a complete one.

---

## Formatting

- [ ] **gofmt**: Code is formatted with `gofmt` or `goimports` → [go-linting](../go-linting/SKILL.md)

---

## Documentation

- [ ] **Comment sentences**: Comments are full sentences starting with the name being described, ending with a period → [go-documentation](../go-documentation/SKILL.md)
- [ ] **Doc comments**: All exported names have doc comments; non-trivial unexported declarations too → [go-documentation](../go-documentation/SKILL.md)
- [ ] **Package comments**: Package comment appears adjacent to package clause with no blank line → [go-documentation](../go-documentation/SKILL.md)
- [ ] **Named result parameters**: Only used when they clarify meaning (e.g., multiple same-type returns), not just to enable naked returns → [go-documentation](../go-documentation/SKILL.md)

---

## Error Handling

- [ ] **Handle errors**: No discarded errors with `_`; handle, return, or (exceptionally) panic → [go-error-handling](../go-error-handling/SKILL.md)
- [ ] **Error strings**: Lowercase, no punctuation (unless starting with proper noun/acronym) → [go-error-handling](../go-error-handling/SKILL.md)
- [ ] **In-band errors**: No magic values (-1, "", nil); use multiple returns with error or ok bool → [go-error-handling](../go-error-handling/SKILL.md)
- [ ] **Indent error flow**: Handle errors first and return; keep normal path at minimal indentation → [go-error-handling](../go-error-handling/SKILL.md)

---

## Naming

- [ ] **MixedCaps**: Use `MixedCaps` or `mixedCaps`, never underscores; unexported is `maxLength` not `MAX_LENGTH` → [go-naming](../go-naming/SKILL.md)
- [ ] **Initialisms**: Keep consistent case: `URL`/`url`, `ID`/`id`, `HTTP`/`http` (e.g., `ServeHTTP`, `xmlHTTPRequest`) → [go-naming](../go-naming/SKILL.md)
- [ ] **Variable names**: Short names for limited scope (`i`, `r`, `c`); longer names for wider scope → [go-naming](../go-naming/SKILL.md)
- [ ] **Receiver names**: One or two letter abbreviation of type (`c` for `Client`); no `this`, `self`, `me`; consistent across methods → [go-naming](../go-naming/SKILL.md)
- [ ] **Package names**: No stuttering (use `chubby.File` not `chubby.ChubbyFile`); avoid `util`, `common`, `misc` → [go-packages](../go-packages/SKILL.md)
- [ ] **Avoid built-in names**: Don't shadow `error`, `string`, `len`, `cap`, `append`, `copy`, `new`, `make` → [go-declarations](../go-declarations/SKILL.md)

---

## Concurrency

- [ ] **Goroutine lifetimes**: Clear when/whether goroutines exit; document if not obvious → [go-concurrency](../go-concurrency/SKILL.md)
- [ ] **Synchronous functions**: Prefer sync over async; let callers add concurrency if needed → [go-concurrency](../go-concurrency/SKILL.md)
- [ ] **Contexts**: First parameter; not in structs; no custom Context types; pass even if you think you don't need to → [go-context](../go-context/SKILL.md)

---

## Interfaces

- [ ] **Interface location**: Define in consumer package, not implementor; return concrete types from producers → [go-interfaces](../go-interfaces/SKILL.md)
- [ ] **No premature interfaces**: Don't define before used; don't define "for mocking" on implementor side → [go-interfaces](../go-interfaces/SKILL.md)
- [ ] **Receiver type**: Use pointer if mutating, has sync fields, or is large; value for small immutable types; don't mix → [go-interfaces](../go-interfaces/SKILL.md)

---

## Data Structures

- [ ] **Empty slices**: Prefer `var t []string` (nil) over `t := []string{}` (non-nil zero-length) → [go-data-structures](../go-data-structures/SKILL.md)
- [ ] **Copying**: Be careful copying structs with pointer/slice fields; don't copy `*T` methods' receivers by value → [go-data-structures](../go-data-structures/SKILL.md)

---

## Security

- [ ] **Crypto rand**: Use `crypto/rand` for keys, not `math/rand` → [go-defensive](../go-defensive/SKILL.md)
- [ ] **Don't panic**: Use error returns for normal error handling; panic only for truly exceptional cases → [go-defensive](../go-defensive/SKILL.md)

---

## Declarations and Initialization

- [ ] **Group similar**: Related `var`/`const`/`type` in parenthesized blocks; separate unrelated → [go-declarations](../go-declarations/SKILL.md)
- [ ] **var vs :=**: Use `var` for intentional zero values; `:=` for explicit assignments → [go-declarations](../go-declarations/SKILL.md)
- [ ] **Reduce scope**: Move declarations close to usage; use if-init to limit variable scope → [go-declarations](../go-declarations/SKILL.md)
- [ ] **Struct init**: Always use field names; omit zero fields; `var` for zero structs → [go-declarations](../go-declarations/SKILL.md)
- [ ] **Use `any`**: Prefer `any` over `interface{}` in new code → [go-declarations](../go-declarations/SKILL.md)

---

## Functions

- [ ] **File ordering**: Types → constructors → exported methods → unexported → utilities → [go-functions](../go-functions/SKILL.md)
- [ ] **Signature formatting**: All args on own lines with trailing comma when wrapping → [go-functions](../go-functions/SKILL.md)
- [ ] **Naked parameters**: Add `/* name */` comments for ambiguous bool/int args, or use custom types → [go-functions](../go-functions/SKILL.md)
- [ ] **Printf naming**: Functions accepting format strings end in `f` for `go vet` → [go-functions](../go-functions/SKILL.md)

---

## Style

- [ ] **Line length**: No rigid limit, but avoid uncomfortably long lines; break by semantics, not arbitrary length → [go-style-core](../go-style-core/SKILL.md)
- [ ] **Naked returns**: Only in short functions; explicit returns in medium/large functions → [go-style-core](../go-style-core/SKILL.md)
- [ ] **Pass values**: Don't use pointers just to save bytes; pass `string` not `*string` for small fixed-size types → [go-performance](../go-performance/SKILL.md)
- [ ] **String concatenation**: `+` for simple; `fmt.Sprintf` for formatting; `strings.Builder` for loops → [go-performance](../go-performance/SKILL.md)

---

## Logging

- [ ] **Use slog**: New code uses `log/slog`, not `log` or `fmt.Println` for operational logging → [go-logging](../go-logging/SKILL.md)
- [ ] **Structured fields**: Log messages use static strings with key-value attributes, not fmt.Sprintf → [go-logging](../go-logging/SKILL.md)
- [ ] **Appropriate levels**: Debug for developer tracing, Info for notable events, Warn for recoverable issues, Error for failures → [go-logging](../go-logging/SKILL.md)
- [ ] **No secrets in logs**: PII, credentials, and tokens are never logged → [go-logging](../go-logging/SKILL.md)

---

## HTTP

- [ ] **Server timeouts**: `http.Server` with `ReadHeaderTimeout` set; no bare `http.ListenAndServe` → [go-http](../go-http/SKILL.md)
- [ ] **Bounded bodies**: `http.MaxBytesReader` before decoding; `r.Context()` passed downstream → [go-http](../go-http/SKILL.md)
- [ ] **Error mapping**: Sentinels map to status codes; 500 responses never carry `err.Error()` → [go-http](../go-http/SKILL.md)
- [ ] **Clients**: Per-dependency `*http.Client` with `Timeout`; `NewRequestWithContext`; body closed on every path → [go-http](../go-http/SKILL.md)

---

## Database

- [ ] **Context on queries**: `QueryContext`/`ExecContext`/`BeginTx`, never the ctx-less forms → [go-database](../go-database/SKILL.md)
- [ ] **Rows lifecycle**: `defer rows.Close()` and `rows.Err()` checked after the loop → [go-database](../go-database/SKILL.md)
- [ ] **Transactions**: `defer tx.Rollback()` right after `BeginTx`; `Commit` error checked; only `tx` used inside → [go-database](../go-database/SKILL.md)
- [ ] **No query per row**: Batch with `ANY`/`IN` or a join; placeholders, never string-built SQL → [go-database](../go-database/SKILL.md)

---

## Imports

- [ ] **Import groups**: Standard library first, then blank line, then external packages → [go-packages](../go-packages/SKILL.md)
- [ ] **Import renaming**: Avoid unless collision; rename local/project-specific import on collision → [go-packages](../go-packages/SKILL.md)
- [ ] **Import blank**: `import _ "pkg"` only in main package or tests → [go-packages](../go-packages/SKILL.md)
- [ ] **Import dot**: Only for circular dependency workarounds in tests → [go-packages](../go-packages/SKILL.md)

---

## Generics

- [ ] **When to use**: Only when multiple types share identical logic and interfaces don't suffice → [go-generics](../go-generics/SKILL.md)
- [ ] **Type aliases**: Use definitions for new types; aliases only for package migration → [go-generics](../go-generics/SKILL.md)

---

## Testing

- [ ] **Examples**: Include runnable `Example` functions or tests demonstrating usage → [go-documentation](../go-documentation/SKILL.md)
- [ ] **Useful test failures**: Messages include what was wrong, inputs, got, and want; order is `got != want` → [go-testing](../go-testing/SKILL.md)
- [ ] **TestMain**: Use only when all tests need common setup with teardown; prefer scoped helpers first → [go-testing](../go-testing/SKILL.md)
- [ ] **Real transports**: Prefer `httptest.NewTestServer(t, h)` + real client over mocking HTTP → [go-testing](../go-testing/SKILL.md)
- [ ] **Test context**: Tests use `t.Context()`, not `context.Background()` → [go-testing](../go-testing/SKILL.md)
- [ ] **No sleep-based waits**: Timing tests use `synctest`, not `time.Sleep` → [go-testing](../go-testing/SKILL.md)

---

## Automated Checks

Run automated pre-review checks:

```bash
bash scripts/pre-review.sh ./...         # gofmt + go vet + golangci-lint
bash scripts/pre-review.sh --json ./...  # structured JSON output
go fix -diff ./...                       # pending modernizations
go test -race ./...                      # required if the diff touches goroutines
```

Fix everything these report before starting the checklist above — a human
review of machine-detectable defects is wasted attention.
[go-linting](../go-linting/SKILL.md) owns the full verification gate and linter
configuration.

---

## Related Skills

- **Style foundations**: See [go-style-core](../go-style-core/SKILL.md) when resolving formatting debates or applying the clarity > simplicity > concision priority
- **Linting setup**: See [go-linting](../go-linting/SKILL.md) when configuring golangci-lint or adding automated checks to CI
- **Error strategy**: See [go-error-handling](../go-error-handling/SKILL.md) when reviewing error wrapping, sentinel errors, or the handle-once pattern
- **Naming conventions**: See [go-naming](../go-naming/SKILL.md) when evaluating identifier names, receiver names, or package-symbol stuttering
- **Testing patterns**: See [go-testing](../go-testing/SKILL.md) when reviewing test code for table-driven structure, failure messages, or helper usage
- **Concurrency safety**: See [go-concurrency](../go-concurrency/SKILL.md) when reviewing goroutine lifetimes, channel usage, or mutex placement
- **Logging practices**: See [go-logging](../go-logging/SKILL.md) when reviewing log usage, structured logging, or slog configuration
- **Acting on the findings**: See [go-code-refactor](../go-code-refactor/SKILL.md) when the review turns into restructuring existing code and behavior must be proven unchanged
- **Reviewing for bloat instead of defects**: See [go-code-refactor](../go-code-refactor/SKILL.md) and its `references/OVER-ENGINEERING.md` when the ask is what to delete — single-implementation interfaces, hand-rolled stdlib, dependencies Go now ships
- **HTTP servers and clients**: See [go-http](../go-http/SKILL.md) and its `references/WEB-SERVER.md` when the diff is a handler, middleware, server setup, or client call
- **SQL access**: See [go-database](../go-database/SKILL.md) when the diff is a repository, query, transaction, or migration
