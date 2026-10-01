---
name: go-code-review
description: Use when reviewing Go code, checking community style, or preparing to submit a Go PR, including a final review of Go changes without an explicit style-review request.
allowed-tools: Bash(bash:*)
---

# Go Code Review Checklist

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`); the HTTP and
> database rows route to skills that carry their own version notes.

## Resource Routing

- `../go-style-core/SKILL.md` - Load on every review before the first finding (Review Procedure step 2); its convention files fix the report language.
- `assets/review-template.md` - Use when formatting review output with Must Fix, Should Fix, and Nits sections.
- `scripts/pre-review.sh` - Run before manual review to collect gofmt, go vet, and golangci-lint results.

## Review Procedure

1. **Settle the scope.** A diff (`git diff`, a PR, named files) is the scope.
   No diff → non-test code, riskiest packages first. Risk order picks files —
   Security → HTTP → Database → Concurrency → the rest; the section order below
   applies within a file. Ask only when the module has more than one binary.
   The flat checklist below is for a diff, not for a whole package.
2. **Load the owners before the first finding** (a read of
   `../<name>/SKILL.md` where there is no `Skill` tool):
   [go-style-core](../go-style-core/SKILL.md) on every review, then, for each
   checklist section the diff touches, the owner its rows' arrows name —
   `go-http` for a handler, `go-database` for a query, `go-concurrency` for a
   goroutine, `go-security` when input reaches a sink. Read the convention
   files [House Style Wins](../go-style-core/SKILL.md#house-style-wins) names;
   they fix the report language, error style, and test style, and outrank
   every rule here — except the idiom: an older form kept for consistency with
   the package is a finding ([Write Current Go](../go-style-core/SKILL.md#write-current-go)).
3. From the project, run `bash <installed-skill-dir>/scripts/pre-review.sh ./...` plus
   `go fix -diff <packages in the diff>`, writing the path itself into the
   command: `<installed-skill-dir>` is the base directory the host printed when
   it loaded this skill, or the directory this SKILL.md was read from (under
   Codex, `~/.codex/skills/go-code-review` or `~/.agents/skills/go-code-review`);
   exit 127 means the path is wrong. A project with no golangci-lint
   config is linted with the go-linting baseline where it is installed; the
   output's `config` says which configuration ran. Report what the tools find
   before the checklist; never spend review attention on what a tool reports.
   Fix only when the request asks for fixes, and then follow
   [go-code](../go-code/SKILL.md#workflow) steps 2–3 before the first edit:
   `go-style-core`, which carries the idiom card, and the owners the edit
   needs. The rows below carry no line for what `gofmt`, `go vet`, `revive`,
   `godot`, `staticcheck`, `gosec`, and `modernize` report — doc-comment
   form, error-string case, naming case, `interface{}`, `math/rand`, import
   order — so a category whose linter is `unavailable`, or not enabled in the
   config that ran, is reviewed by hand or listed under Not Reviewed.
4. **Subtract first**: before any style row, ask of each added block what could
   stop existing — the Less Code section below. Unneeded growth is a Should Fix.
5. Read the scope file-by-file; for each file, check the categories below in order
6. Report every finding, at every severity — one you could not prove is
   `plausible`, never dropped; filtering is the reader's pass, not yours —
   unless the request names a severity: then report that severity, plus one
   line counting the findings at the others
7. Report through `assets/review-template.md`, grouped by severity; findings
   carry the information, prose stays short ([go-style-core](../go-style-core/SKILL.md#how-much-to-say))

> **Validation**: Every finding names a file and line, and carries its
> `verified` / `plausible` marker — a guess dressed as `verified` costs trust.
> Name the checks you actually ran; a linter that was not installed or tests
> that did not run are `unavailable`, never presented as clean. Gate result per
> [go-linting](../go-linting/SKILL.md#verification-gate) (`PASS` / `FAIL` /
> `INCOMPLETE`) with the checks actually run.

## Less Code

- [ ] **Subtract first**: for each added block, what can stop existing? Name the cut tag and show the shorter form; the hunt list and the reach-for table live with the owner → [go-code-refactor](../go-code-refactor/SKILL.md#delete-before-you-restructure)
- [ ] **Shorter only where it reads as well**: never golf; validation at trust boundaries, data-loss error handling, and security checks are never "simplified" away, nor are the tests that fail when the logic breaks → [go-code-refactor](../go-code-refactor/references/OVER-ENGINEERING.md)

## Correctness

- [ ] **Does it do what it claims?** Trace each changed function from inputs to outputs on the happy path and at the edges — empty, nil, zero, max, concurrent. A wrong result or silent data loss is a Must Fix even when every style row passes
- [ ] **Invariants**: what the surrounding code assumes — ordering, non-nil, lock held, ctx alive — still holds after the change; name the assumption in the finding
- [ ] **Callers outside the diff**: for a changed exported signature, behavior, or interface method set, look up references and implementations and read the callers the diff did not touch; one it leaves broken is a Must Fix
- [ ] **Failure paths**: every error branch, timeout, and partial write leaves state a caller can recover from — read each with the failing call moved one line earlier

## Documentation

- [ ] **Package comments**: Package comment appears adjacent to package clause with no blank line → [go-documentation](../go-documentation/SKILL.md)

## Error Handling

- [ ] **Every error is handled**: returned, wrapped once, or logged once; a `_` on an error value is a finding unless the line says why → [go-error-handling](../go-error-handling/SKILL.md)
- [ ] **In-band errors**: No magic values (-1, "", nil); use multiple returns with error or ok bool → [go-error-handling](../go-error-handling/SKILL.md)
- [ ] **Indent error flow**: Handle errors first and return; keep normal path at minimal indentation → [go-style-core](../go-style-core/SKILL.md#reduce-nesting)

## Naming

- [ ] **Package names**: No stuttering (use `chubby.File` not `chubby.ChubbyFile`); avoid `util`, `common`, `misc` → [go-naming](../go-naming/SKILL.md)
- [ ] **Built-in names stay free**: `error`, `string`, `len`, `cap`, `append`, `copy`, `new`, `make` name only the builtins; a local of that name is a finding → [go-style-core](../go-style-core/SKILL.md)

## Concurrency

- [ ] **Goroutine lifetimes**: Clear when/whether goroutines exit; document if not obvious → [go-concurrency](../go-concurrency/SKILL.md)
- [ ] **Contexts**: the first parameter, passed through every call that can block or be cancelled; a `ctx` field in a struct or a custom Context type is a finding → [go-context](../go-context/SKILL.md)

## Interfaces

- [ ] **Interfaces where they are consumed**: an interface appears when a consumer substitutes implementations, declared in the consumer's package; one declared beside its only implementation "for mocking" is a finding → [go-interfaces](../go-interfaces/SKILL.md)
- [ ] **Receiver type**: Use pointer if mutating, has sync fields, or is large; value for small immutable types; don't mix → [go-interfaces](../go-interfaces/SKILL.md)

## Data Structures

- [ ] **Empty list on a v1 wire**: in a package on `encoding/json` v1, a list the contract writes as `[]` keeps `out := make([]T, 0, n)`, since v1 writes nil as `null`; a change to `var out []T` there is a finding → [go-data-structures](../go-data-structures/SKILL.md#declaring-empty-slices)
- [ ] **Copy depth**: check the ownership contract: `slices.Clone`/`maps.Clone` are shallow, and a type's `Clone` follows its documented contract. Flag a copy that violates the contract or an unrequested depth change presented as cleanup; a deeper copy that fixes an existing contract violation is a bug fix → [go-defensive](../go-defensive/references/BOUNDARY-COPYING.md#copy-depth-is-part-of-the-contract)
- [ ] **Copying values**: do not copy structs containing locks or other synchronization values after use; a value receiver on a type with `*T` methods is a finding → [go-data-structures](../go-data-structures/SKILL.md)

## Security

- [ ] **Trace untrusted input to its sink**: SQL, shell, template, file path, outbound URL, log line — each has a stdlib defense at the boundary → [go-security](../go-security/SKILL.md)
- [ ] **Secrets**: constant-time compare, memory-hard password hash, no credential in a log or error, `InsecureSkipVerify` only in tests → [go-security](../go-security/SKILL.md)
- [ ] **Errors over panics**: a failure the caller can act on returns an error; `panic` marks a programmer error the process cannot continue past → [go-defensive](../go-defensive/SKILL.md)

## Declarations and Initialization

- [ ] **Group similar**: Related `var`/`const`/`type` in parenthesized blocks; separate unrelated → [go-style-core](../go-style-core/SKILL.md)
- [ ] **var vs :=**: Use `var` for intentional zero values; `:=` for explicit assignments → [go-style-core](../go-style-core/SKILL.md)
- [ ] **Reduce scope**: Move declarations close to usage; use if-init to limit variable scope → [go-style-core](../go-style-core/SKILL.md)
- [ ] **Struct init**: Prefer keyed fields; preserve meaningful zero values and local exceptions → [go-style-core](../go-style-core/SKILL.md)

## Functions

- [ ] **File ordering**: Types → constructors → exported methods → unexported → utilities → [go-functions](../go-functions/SKILL.md)
- [ ] **Naked parameters**: Add `/* name */` comments for ambiguous bool/int args, or use custom types → [go-functions](../go-functions/SKILL.md)

## Style

- [ ] **Current Go**: Changed lines use the form available at the `go` directive; an older idiom kept because the neighbor uses it is Should Fix, and `go fix -diff` on the diff's packages reports nothing in changed lines; the forms are one line each on the [idiom card](../go-style-core/SKILL.md#current-go-idiom-card) → [go-style-core](../go-style-core/SKILL.md#write-current-go)
- [ ] **Pass values**: small fixed-size types (`string`, `time.Time`, a few ints) travel by value; a pointer parameter means mutation or identity → [go-functions](../go-functions/SKILL.md)

## Logging

- [ ] **Use slog**: a package with no logger uses `log/slog`, not `log` or `fmt.Println`, for operational logging; one already on `log`, zap, or logrus keeps it, and a migration is its own change → [go-logging](../go-logging/SKILL.md)
- [ ] **Structured fields**: Log messages use static strings with key-value attributes, not fmt.Sprintf → [go-logging](../go-logging/SKILL.md)

## HTTP

- [ ] **Server timeouts**: `http.Server` with `ReadHeaderTimeout` set; no bare `http.ListenAndServe` → [go-http](../go-http/SKILL.md)
- [ ] **Bounded bodies**: `http.MaxBytesReader` before decoding; `r.Context()` passed downstream → [go-http](../go-http/SKILL.md)
- [ ] **Error mapping**: Sentinels map to status codes; 500 responses never carry `err.Error()` → [go-http](../go-http/SKILL.md)
- [ ] **Clients**: Per-dependency `*http.Client` with `Timeout`; `NewRequestWithContext`; body closed on every path → [go-http](../go-http/SKILL.md)

## Database

- [ ] **Context on queries**: `QueryContext`/`ExecContext`/`BeginTx`, never the ctx-less forms → [go-database](../go-database/SKILL.md)
- [ ] **Rows lifecycle**: `defer rows.Close()` and `rows.Err()` checked after the loop → [go-database](../go-database/SKILL.md)
- [ ] **Transactions**: `defer tx.Rollback()` right after `BeginTx`; `Commit` error checked; only `tx` used inside → [go-database](../go-database/SKILL.md)
- [ ] **No query per row**: Batch with `ANY`/`IN` or a join; placeholders, never string-built SQL → [go-database](../go-database/SKILL.md)

## Generics

- [ ] **When to use**: Only when multiple types share identical logic and interfaces don't suffice → [go-generics](../go-generics/SKILL.md)
- [ ] **Type aliases**: Use definitions for new types; aliases only for package migration → [go-generics](../go-generics/SKILL.md)

## Testing

- [ ] **Examples**: Include runnable `Example` functions or tests demonstrating usage → [go-documentation](../go-documentation/SKILL.md)
- [ ] **Useful test failures**: Messages include what was wrong, inputs, got, and want; order is `got != want` → [go-testing](../go-testing/SKILL.md)
- [ ] **Real transports**: Prefer a test server over mocking HTTP: `httptest.NewTestServer(t, h)` (Go 1.27+) when every request goes through `srv.Client()`, which is the only client that reaches the in-memory server — until `srv.Start()`, `srv.URL` is empty and then `http://example.com`, so another client sends the request to the real host; code under test that builds its own client gets `srv.Start()` first, which listens on loopback and sets `srv.URL`. At a 1.26 directive, `httptest.NewServer(h)` plus `t.Cleanup(srv.Close)` → [go-testing](../go-testing/SKILL.md)
- [ ] **Test context**: Tests use `t.Context()`, not `context.Background()`; work inside `t.Cleanup` uses `context.WithoutCancel(t.Context())`, since `t.Context()` is canceled before cleanup runs → [go-testing](../go-testing/SKILL.md)
- [ ] **No sleep-based waits**: Timing tests use `synctest`, not `time.Sleep` → [go-testing](../go-testing/SKILL.md)

## Related Skills

- [go-http](../go-http/SKILL.md) and [WEB-SERVER.md](../go-http/references/WEB-SERVER.md): handlers, middleware, servers, clients.
