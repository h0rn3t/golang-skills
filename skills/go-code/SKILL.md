---
name: go-code
description: Use when writing, fixing, or refactoring Go code, whether or not the task has one obvious topic — it loads the go-* skills the task needs and runs the closing gate. Use it too when /go-code modifies another workflow (e.g. /opsx:apply /go-code); as a modifier it selects Go rules, not a change name or path.
---

# Go Code Profile

Route Go work to the relevant owners, then close with their verification gate.

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`).

## Resource Routing

Resolve sibling skills and references relative to this installed directory
(Claude Code prints it as the skill's base directory); run their scripts from
the target project. A skill counts as loaded only when its `SKILL.md` content
is in this context: in Claude Code call the `Skill` tool with the skill name;
in Codex or any host without a skill tool, read `../<name>/SKILL.md` next to
this file. Missing resources: report the gap, use available guidance, and
continue independent authorized work without inventing rules.

- `../go-style-core/SKILL.md` — Read once per task for house style, fallback
  rules, and communication guidance. Read its references only for a decision
  the task requires.
- `../go-linting/SKILL.md` — Read its Verification Gate at step 6, on a task
  that edits Go and only when a shell tool (`Bash` in Claude Code) is in your
  tool list. Without one nothing in it can run: leave the file unread and
  report `checks: unavailable (no shell)` in one line.
- `../go-code-refactor/references/OVER-ENGINEERING.md` — Read the detailed
  restraint ladder when a [Declaration Budget](#declaration-budget) entry is
  in doubt; its replacement catalog when seeking a simpler existing API; its
  audit lane only when the requested deliverable is a complexity audit.
  Ordinary edits do not require this file.

## Workflow

1. **Resolve invocation.** `$go-code <task>` or `/go-code <task>` selects Go
   work. As a modifier, e.g. `/opsx:apply add-auth /go-code`, remove the modifier
   before the host parses its arguments. It is never a change name or path;
   the host retains workflow state, checkpoints, and delegation policy.
2. **Load `go-style-core`, read the code, and check for a shell.** Load
   `go-style-core` on every task. Then inspect repository instructions,
   `go.mod`, neighboring code, tests, and callers, and look at your tool
   list: a shell tool (`Bash` in Claude Code) means step 6 runs the checks; no
   shell tool means nothing in step 6 can run, `go-linting` stays unread, and
   the report says so in one line. Decide the shape of the task here as
   well: a new function, package, or stub body makes step 4 apply; a fix or a
   restructuring of existing code skips it.
3. **Load the owners before the first edit.** Match the task against
   [Route Before The First Edit](#route-before-the-first-edit) and load each
   matched owner plus every `Also load` entry whose condition holds. When
   step 4 applies, `go-testing` is one of them: the Contract Table is a test
   file. Select by the decisions and behavior being changed, not every syntax
   element present: a routine local variable or `if` does not trigger naming,
   documentation, or extra style references. The step ends when every selected
   skill's content is in context; do not make the first edit before that. Add
   owners when new evidence requires them; there is no numerical cap, and
   content already loaded this session is reused, not reread. Under the Claude
   Code plugin a PreToolUse hook blocks the first Go edit that precedes these
   loads and names the missing skills once. A blocked edit was not applied
   and the file is unchanged, so load what the hook names and retry the same
   edit; the hook is a reminder, not a substitute for this step. It infers
   owners from decision-bearing syntax only (a `_test.go` path, `%w`
   wrapping, goroutines, `context.With*`, SQL, `slog.`, exec and templates,
   `defer`, type parameters, `interface {`, `package main`, rate limiting,
   HTTP server and client calls); collections, naming, documentation,
   functions, performance, refactoring, linting, and troubleshooting it cannot
   see, and its "Also load" conditions are this table's. Its silence is not a
   passing gate result.
4. **Write the Contract Table, for new code only.** For a new function,
   package, or stub body, turn the documentation and the request into the
   test file the [Contract Table](#contract-table) describes, before the
   first production edit: a case written after the body checks only what the
   body already does. Its cases are what step 6 runs or reads.
5. **Implement the authorized scope.** For new functions, packages, or stub
   bodies, follow [Writing New Code](#writing-new-code): the body takes the
   [Plain Code](#plain-code) form, and every package-level declaration it
   adds is counted in the [Declaration Budget](#declaration-budget). For
   behavior-preserving restructuring, use the [delete-first priority](../go-code-refactor/SKILL.md#delete-before-you-restructure)
   and climb the restraint ladder in [OVER-ENGINEERING.md](../go-code-refactor/references/OVER-ENGINEERING.md#the-restraint-ladder)
   for each proposed helper, type, layer, option, or import. Preserve required
   behavior, validation, security controls, and meaningful tests.
   Resolve routine choices without stopping; ask only for missing information
   that changes correctness, scope, or authorization.
6. **Verify and report.** With a shell, run the Contract Table file with the
   closing gate below. Without one, read each case against its code path.
   Report as [go-style-core](../go-style-core/SKILL.md#how-much-to-say) says,
   in this shape and order: the outcome, one line of checks, the one budget
   line, and a sentence per material gap.

   ```text
   <what the change does, in one sentence>
   checks: gofmt pass · vet pass · test pass · lint skipped (no config)
   added package-level declarations: 1 — parseLimit: handleList, handleSearch
   <one sentence per material gap, or nothing>
   ```

   Without a shell the checks line is `checks: unavailable (no shell)` and
   nothing more is said about the gate. The test file is the record of the
   cases; the report names the file and its result, not the cases one by one.

## Route Before The First Edit

The third column adds owners only under its stated condition. Prefer the
narrower owner when topics overlap; ordinary identifier creation alone does
not require `go-naming` or `go-documentation`.

| Task touches | Skill | Also load |
|---|---|---|
| errors, wrapping, `errors.Is`/`errors.AsType` | [go-error-handling](../go-error-handling/SKILL.md) | — |
| goroutines, channels, mutexes, races | [go-concurrency](../go-concurrency/SKILL.md) | [go-context](../go-context/SKILL.md) if anything is cancelled |
| `context.Context`, timeouts, cancellation | [go-context](../go-context/SKILL.md) | [go-concurrency](../go-concurrency/SKILL.md) if goroutines are started |
| tests, table-driven cases, `synctest` | [go-testing](../go-testing/SKILL.md) | — |
| API naming changes or a naming question | [go-naming](../go-naming/SKILL.md) | — |
| new or changed exported API, doc comments | [go-documentation](../go-documentation/SKILL.md) | [go-naming](../go-naming/SKILL.md) if API names change |
| interfaces, embedding, test doubles | [go-interfaces](../go-interfaces/SKILL.md) | — |
| function API design, constructor configuration, ordering, signatures, `Printf` helpers | [go-functions](../go-functions/SKILL.md) | — |
| slices, maps, arrays, sets | [go-data-structures](../go-data-structures/SKILL.md) | — |
| declaration, enum, or initialization decisions | [go-style-core references](../go-style-core/SKILL.md#resource-routing) | — |
| loop/switch mechanics or statement scoping decisions | [go-style-core references](../go-style-core/SKILL.md#resource-routing) | — |
| type parameters, constraints, generic methods | [go-generics](../go-generics/SKILL.md) | — |
| `slog`, log levels, request-scoped fields, metrics and trace correlation | [go-logging](../go-logging/SKILL.md) | [go-security](../go-security/SKILL.md) if a secret or PII could reach a log line |
| `defer` cleanup, boundary copies, mutable globals, nil/aliasing/overflow traps | [go-defensive](../go-defensive/SKILL.md) | — |
| hot paths, allocations, benchmarks | [go-performance](../go-performance/SKILL.md) | [go-troubleshooting](../go-troubleshooting/SKILL.md) if the cause of slowness is unknown |
| package layout, imports, dependencies | [go-packages](../go-packages/SKILL.md) | — |
| restructuring or deleting existing code | [go-code-refactor](../go-code-refactor/SKILL.md) | — |
| linter config, CI checks | [go-linting](../go-linting/SKILL.md) | — |
| HTTP handlers, routing, middleware, servers, clients | [go-http](../go-http/SKILL.md) | [go-error-handling](../go-error-handling/SKILL.md); [go-security](../go-security/SKILL.md) if input reaches a file, shell, URL, or template |
| SQL queries, transactions, repositories, migrations | [go-database](../go-database/SKILL.md) | [go-error-handling](../go-error-handling/SKILL.md); [go-security](../go-security/SKILL.md) if identifiers come from input |
| untrusted input, secrets, tokens, TLS, cookies | [go-security](../go-security/SKILL.md) | [go-defensive](../go-defensive/SKILL.md) |
| retries, idempotency, circuit breakers, overload, backpressure, fallback | [go-resilience](../go-resilience/SKILL.md) | the HTTP, SQL, context, or concurrency owner when its mechanics change |
| ticket, wrong result, environment regression, panic, hang, leak, flaky test; cause unknown | [go-troubleshooting](../go-troubleshooting/SKILL.md) | the owner of the mechanism once found |
| JSON and other wire formats, struct tags | [go-defensive](../go-defensive/SKILL.md) (tags) | [go-packages](../go-packages/SKILL.md) (`json/v2` on the ladder) |
| CLI entry point, flags, `main`/`run` | [go-packages](../go-packages/SKILL.md) | — |
| a list of findings from a review or audit | the rows the findings name | per area, not per task |

## Writing New Code

Start at the required entry point and its callers. Implement the documented
success and failure paths within the existing API; design a new API only when
the task calls for one. A stub's `panic("not implemented")` is missing behavior,
not a refactor contract: replace it and satisfy the acceptance tests. Use
baseline/after equivalence only for existing behavior the task must preserve.

### Contract Table

Before the first production edit, turn the documentation and the request into
a table-driven test in the package — a new `<pkg>_contract_test.go`, or cases
added to the existing test file — one case per observable clause: each request
or input class, each method and status code, ordering, error text, and the
empty, nil, and invalid inputs. A case names the input and the result the
clause promises; a clause that says what the code must *not* do is a case too.
A clause written as a class — "any other method", "any other value", "nothing
else" — takes its case from the member a library default treats unlike the
rest, because that member is where the class leaks: `HEAD` under a `GET`
pattern, `t` and `1` under `strconv.ParseBool`, the bare path under a subtree
pattern. The member the code plainly rejects (`POST`, `maybe`) fails on its
own and needs no case. The file is the contract table, and it is written before the body because a
case derived from the code only checks what the code already does.

| Clause | Case | Expected |
|---|---|---|
| "an unknown id is a 404" | `GET /items/nope` | 404 |
| "returns the matching entries as a JSON array" | zero matches | `[]`, not `null` |
| "any other value is an error naming the parameter" | `limit=abc` | error text contains `limit` |
| "any method other than `GET` is a 405" | `HEAD /healthz` | 405, not the `GET` body |

Where a case contradicts what a standard-library default does — a nil slice
encoding as `null`, a `GET` pattern answering `HEAD` with 200,
`strconv.ParseBool` accepting `1` and `t`, a `ServeMux` subtree pattern
answering the bare path with a 301 — the contract wins, and
the default is overridden in code rather than explained in prose. With a
shell, the file runs as part of the closing gate; without one, every case is
read against its code path and the report says so. A failing or unrunnable
case is reported as such, never deleted or weakened to pass; an expectation
that turns out wrong is corrected against the documentation, not against the
implementation. The file is plain code too: one case struct, one table, one
`t.Run` loop, and a failure message in the `Func(input) = got, want` form;
[go-testing](../go-testing/SKILL.md) owns the table-test form.

### Plain Code

The body reads as the documentation reads: guard clauses first, the work
once, one return. This is the form for every function the task adds — the
entry point, anything it calls, and the contract test alike.

- **Vocabulary from the specification.** The doc comment's nouns name the
  variables; a value used once is written where it is used and has no name.
- **Smallest scope that works.** A document built in one function is an
  anonymous struct or a type declared inside that function; a step done once
  is inline; an error carrying context is `fmt.Errorf("op %q: %w", key, err)`,
  matched by the caller with `errors.Is`.
- **Fewer names, not fewer states.** A value used once needs no name; a fact
  the code tracks — what was already asked for, what was sent, what failed —
  keeps its own variable even when another one almost holds it.
- **A comment states a constraint the code cannot show** — a default
  deliberately overridden, the clause a branch serves — and nothing else.
  [go-style-core](../go-style-core/SKILL.md#formatting) owns comment style
  and [the early return](../go-style-core/SKILL.md#reduce-nesting).
- **Neighbors set the register.** Where the package has code, match its
  naming, comment density, and idiom ([House Style](../go-style-core/SKILL.md#house-style-wins)).

The function below is the whole implementation of a documented JSON document:
one function-local type, one anonymous document, the empty-input rule and the
one tracked fact as code. The same document as two package-level types, a
constructor for the entry, and a `writeJSON` helper is the growth the budget
below counts.

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

### Declaration Budget

The specification already declares what the task needs: a body behind an
existing signature adds no package-level declaration by default, and the count
of those it does add is what the report carries. Count every function, method,
type, interface, and variable added at package level in production code across
the whole diff; [go-testing](../go-testing/SKILL.md) owns test helpers. A
declaration is written when the code shows the need at the moment it is
written:

1. Two call sites exist in the diff, and the report names both.
2. A caller outside the function names it: it reads a field through
   `errors.AsType`, satisfies an interface a consumer declares
   ([go-interfaces](../go-interfaces/SKILL.md) owns the shape), or owns a
   resource whose lifetime outlives one call. A caller that *inspects* or
   *matches* an error reads no field: it uses `errors.Is` on the wrapped
   sentinel, and that needs no type.
3. It is a distinct algorithm — a parser, a scheduler — whose name at the
   call site says more than its body would, and the body left behind reads
   top to bottom without it.

A representation is a value, not a reason: a wire document, a formatted error,
and a sorted view take the [Plain Code](#plain-code) form. A helper that names
the steps of one call site meets none of the three; write the steps inline.
Skill examples show the *order* of an operation; their named helpers
(`validate`, `writeJSON`, `writeError`) are not a list to reproduce. Judge a
helper and its call sites together and count both.

The budget is the restraint ladder applied per declaration; read the ladder in
[OVER-ENGINEERING.md](../go-code-refactor/references/OVER-ENGINEERING.md#the-restraint-ladder)
when a rung is in doubt. Never trade input validation at a trust boundary,
failure behavior, or a security control for a smaller count.

The report carries one line — `added package-level declarations: N` — and,
for each, its name and the call sites or caller that need it. A declaration
whose line cannot name them is removed before the report is written, not
explained in it.

## Close With The Gate

Without a shell tool this section does not apply: the report carries
`checks: unavailable (no shell)` and `go-linting` stays unread.

Use [go-linting](../go-linting/SKILL.md#verification-gate) to select checks for the requested
scope. Repository gates take precedence over defaults; do not union them.
Inspect the final diff and complete authorized work. Reuse passing results
for unchanged code; rerun affected checks after edits and honor host checkpoints.

Use bundled checks when they add evidence beyond that gate:

- Refactor: `../go-code-refactor/scripts/verify-refactor.sh` for baseline/after
  evidence; `../go-code-refactor/scripts/check-debt.sh` for deliberate `Kept:`
  markers, whose format is owned by `go-code-refactor`.
- Error handling: `../go-error-handling/scripts/check-errors.sh`.
- Exported API documentation: `../go-documentation/scripts/check-docs.sh`.
- Before submitting: [go-code-review](../go-code-review/SKILL.md).

Run routine checks inline. Claude Code's `go-verify` agent is optional. The
PostToolUse hook reports only gofmt and vet findings for the edited package and
is silent on success: its silence is not a passing result. Count only results
observed for the current diff and scope; report each check as `pass`, `fail`,
`unavailable (reason)`, or `skipped (reason)` per go-linting.

## Related Skills

- Every owner in [Route Before The First Edit](#route-before-the-first-edit).
- [go-style-core](../go-style-core/SKILL.md): house style, report length, and delegation.
- [go-linting](../go-linting/SKILL.md): the closing gate.
- [go-code-review](../go-code-review/SKILL.md): the checklist before submitting.
