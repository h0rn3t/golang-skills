---
name: go-code
description: Use when writing, fixing, or refactoring Go code, whether or not the task has one obvious topic — it loads the go-* skills the task needs and runs the closing gate. Also use when asked to implement a Go package, function, or handler whose declarations and documentation already exist — write the bodies, fill in a stub, replace panic("not implemented") — even if the request names only the package. Use it too when /go-code modifies another workflow (e.g. /opsx:apply /go-code); as a modifier it selects Go rules, not a change name or path.
---

# Go Code Profile

Route Go work to the relevant owners, then close with their verification gate.

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`).

## Resource Routing

Sibling skills resolve relative to this installed directory and their scripts
run from the target project. A skill is loaded when its `SKILL.md` is in
context: the `Skill` tool in Claude Code, a read of `../<name>/SKILL.md`
elsewhere. Report a missing resource and continue with the guidance at hand.

- `../go-style-core/SKILL.md` — Read once per task for house style, fallback
  rules, and communication guidance. Read its references only for a decision
  the task requires, except its `references/CURRENT-GO.md`: the idiom card is
  read whole before the first edit of a task that writes Go.
- `../go-linting/SKILL.md` — Read its Verification Gate at step 6, on a task
  that edits Go and only when a shell tool (`Bash` in Claude Code) is in your
  tool list. Without one nothing in it can run: leave the file unread; the
  edit hook's output is the check record
  ([go-style-core](../go-style-core/SKILL.md#the-edit-hook-record)).
- `references/NEW-CODE-EXAMPLES.md` — Read when the shape of a Contract Table
  case, a Plain Code body, or a budgeted helper is in doubt. Ordinary tasks do
  not require it.
- `../go-code-refactor/references/OVER-ENGINEERING.md` — Read the detailed
  restraint ladder when a [Declaration Budget](#declaration-budget) entry is
  in doubt; its replacement catalog when seeking a simpler existing API; its
  audit lane only when the requested deliverable is a complexity audit.

## Workflow

1. **Resolve invocation.** `$go-code <task>` or `/go-code <task>` selects Go
   work. A slash command inserts this file and loads nothing else: no sibling
   skill comes with it, so steps 2 and 3 make their `Skill` calls exactly as
   they do on any other invocation. As a modifier, e.g.
   `/opsx:apply add-auth /go-code`, remove the modifier before the host parses
   its arguments; it is never a change name or path, and the host keeps
   workflow state, checkpoints, and delegation policy.
2. **Load `go-style-core`, read the idiom card and the code, check for a
   shell.** Load `go-style-core` on every task and read its
   `references/CURRENT-GO.md` whole in the same message — one line per idiom
   with the older form, the form the module's `go` directive allows, and the
   trap; its older rows apply at every directive, so a `head` or a `grep`
   over it misses what a Go 1.19 module still gets. Then inspect repository
   instructions, `go.mod` (its `go` directive sets the idiom), neighboring
   code, tests, and callers. Before changing a symbol that other files use,
   find its references semantically — gopls MCP (`go_symbol_references`) or
   the `LSP` tool (`findReferences`), per
   [GOPLS.md](../go-code-refactor/references/GOPLS.md#three-ways-in); with
   neither wired, grep and let the step 6 build catch the rest. gopls
   diagnostics do not replace step 6. Do not install a server for the task.
   A shell tool (`Bash` in Claude Code) in your tool list means step 6 runs
   the checks; without one `go-linting` stays unread and the report says so
   in one line. A new function, package, or stub body makes step 4 apply; a
   fix or a restructuring skips it.
3. **Load the owners before the first edit.** Match the task against
   [Route Before The First Edit](#route-before-the-first-edit) and load each
   matched owner plus every `Also load` entry whose condition holds, with the
   `Skill` calls for all of them in one message: an owner loaded on a turn of
   its own re-reads the whole context. When step 4 applies, `go-testing` is
   one of them. Select by the decisions being changed: a routine local
   variable or `if` triggers no owner. No edit before every selected skill is
   in context; add owners when new evidence requires them.
4. **Write the Contract Table, for new code only.** Turn the documentation
   and the request into the test file the [Contract Table](#contract-table)
   describes before the first production edit: a case written after the body
   checks only what the body already does. Its cases are what step 6 runs or
   reads.
5. **Implement the authorized scope.** New code takes the
   [Plain Code](#plain-code) form, counts every package-level declaration
   it adds in the [Declaration Budget](#declaration-budget), and gets the
   [Delete Pass](#delete-pass) and [Reader Pass](#reader-pass) once its cases pass. Restructuring
   follows the [delete-first priority](../go-code-refactor/SKILL.md#delete-before-you-restructure)
   and climbs the restraint ladder in [OVER-ENGINEERING.md](../go-code-refactor/references/OVER-ENGINEERING.md#the-restraint-ladder)
   for each proposed helper, type, layer, option, or import. Preserve
   required behavior, validation, security controls, and meaningful tests. A
   task that does not ask for a dependency adds none: `go.mod` stays as it
   is, and a test compares with the standard library
   ([go-testing](../go-testing/SKILL.md#assertions-match-the-repository) owns
   the assertion rule). Resolve routine choices without stopping; ask only
   for missing information that changes correctness, scope, or authorization.
6. **Verify and report.** With a shell, run the Contract Table file with the
   closing gate below; without one, read each case against its code path.
   Report as [go-style-core](../go-style-core/SKILL.md#how-much-to-say) says,
   in this shape and order:

   ```text
   <what the change does, in one sentence>
   checks: gofmt pass · vet pass · test pass · lint skipped (no config)
   added package-level declarations: 1 — parseLimit: handleList, handleSearch
   <one sentence per material gap, or nothing>
   ```

   Without a shell the checks line carries the edit hook's record, check by
   check — `test pass (hook)`, `lint unavailable (no shell)` — as
   [go-style-core](../go-style-core/SKILL.md#the-edit-hook-record) defines
   it. The report names the test file and its result, not the cases one by
   one.

## Writing New Code

Start at the required entry point and its callers. Implement the documented
success and failure paths within the existing API; design a new API only when
the task calls for one. A stub's `panic("not implemented")` is missing
behavior, not a refactor contract: replace it and satisfy the acceptance
tests.

### Contract Table

Before the first production edit, turn the documentation and the request into
a table-driven test in the package — a new `<pkg>_contract_test.go`, or cases
added to the existing test file — one case per observable clause: each request
or input class, each method and status code, ordering, error text, and the
empty, nil, and invalid inputs; a clause that says what the code must *not*
do is a case too. The empty case is built from nil — a nil slice, a nil map,
a request with no body — not from an empty literal: `[]T{}` already has the
shape the case is meant to prove, and `slices.Clone` of nil is nil. A case
for a list that may come back empty compares the body text with `[]`
(`strings.TrimSpace(rec.Body.String())`), never a decoded value:
`json.Unmarshal` reads `null` into an empty slice, and the case passes on the
one body the contract forbids. A clause written as a class — "any other method", "any other value", "nothing else" —
takes its case from the member a library default treats unlike the rest,
because that member is where the class leaks. For "any method other than
`GET`" that member is `HEAD`, and the case is one `HEAD` request per `GET`
path in the contract, the health check included: a `GET` pattern answers
`HEAD` with 200 on its own. For `strconv.ParseBool` it is `t` and `1`; for a
subtree pattern, the bare path. The member the code plainly rejects (`POST`,
`maybe`) fails on its own and needs no case.

Where a case contradicts a standard-library default — a nil slice encoding as
`null`, a `GET` pattern answering `HEAD` with 200 — the contract wins, and the
default is overridden in code rather than explained in prose. A failing or
unrunnable case is reported as such, never deleted or weakened to pass; an
expectation that turns out wrong is corrected against the documentation, not
against the implementation. The file is plain code too: one case struct, one
table, one `t.Run` loop, and a failure message in the `Func(input) = got, want`
form; [go-testing](../go-testing/SKILL.md) owns the table-test form, and
[NEW-CODE-EXAMPLES.md](references/NEW-CODE-EXAMPLES.md) shows a worked table.

### Plain Code

The body reads as the documentation reads: reject invalid states early, then
do the main work in its natural order. Return directly when that makes the
result clear; do not introduce a result variable to enforce one final return.
This applies to the entry point, anything it calls, and the contract test.

- **Vocabulary from the specification.** The doc comment's nouns name the
  variables.
- **Smallest scope that works.** A document built in one function is an
  anonymous struct or a type declared inside that function; a step done once
  is inline unless it sits at another level of abstraction
  ([Declaration Budget](#declaration-budget) rule 4); an error carrying
  context is `fmt.Errorf("op %q: %w", key, err)`, matched by the caller with
  `errors.Is`.
- **Names that explain, states that stay.** An intermediate value gets a name
  when its expression nests deeper than one call or the name says what the
  expression does not; otherwise it is written where it is used. A fact the
  code tracks — what was already asked for, what was sent, what failed —
  keeps its own variable even when another one almost holds it.
- **A comment states what the code cannot show**: a constraint, a default
  deliberately overridden, the business or historical reason behind a choice
  that neither the code nor the signature shows — never a narration of the
  next line. [go-style-core](../go-style-core/SKILL.md#formatting)
  owns comment style and [the early return](../go-style-core/SKILL.md#reduce-nesting).
- **Use a standard-library operation when it states the intent more clearly.**
  `slices.SortFunc`, `slices.Sorted(maps.Keys(m))`, `min`, and `max` can
  replace mechanical loops; a loop that tracks several related facts may
  read better as a loop. Check ordering, ownership, and nil-versus-empty
  results. `cmp.Or` evaluates every argument, so use it only when that is
  acceptable. [Reach For What Go Ships](../go-code-refactor/references/OVER-ENGINEERING.md#reach-for-what-go-ships)
  lists replacements; [go-data-structures](../go-data-structures/SKILL.md)
  owns collection semantics.
- **New JSON is `encoding/json/v2`.** A package with no `encoding/json`
  import writes `import json "encoding/json/v2"` (Go 1.27): `json.Marshal(doc)`
  and `json.MarshalWrite(w, doc)` encode a nil slice as `[]` and a nil map
  as `{}`, so the body carries no `make([]T, 0, n)`, `[]T{}`, or
  `slices.AppendSeq` for the wire, and the empty case built from nil passes
  as written. A package already decoding with v1 keeps v1, where nil is
  `null` and `[]` is `make([]T, 0, n)`;
  [JSON-V2.md](../go-http/references/JSON-V2.md#choose-the-io-api) owns the API.
- **Neighbors set the register.** Where the package has code, match its
  naming and comment density ([House Style](../go-style-core/SKILL.md#house-style-wins)).
  Not its Go version: the current form at the module's `go` directive is
  written even beside an older neighbor
  ([Write Current Go](../go-style-core/SKILL.md#write-current-go)).

[NEW-CODE-EXAMPLES.md](references/NEW-CODE-EXAMPLES.md) carries a whole
documented JSON document in this form.

### Delete Pass

When the Contract Table passes, walk the diff once, top to bottom, and delete
what the test already proves or the code already shows. Each item below is a
line a reviewer sends back:

- A comment that narrates the next line or repeats what the code already
  shows. A comment giving a reason the code cannot show stays.
- A name used once that explains nothing, and the variable it fills: the
  value is written where it is used. A name that passes the Plain Code test
  stays.
- A field, header, or option set to what the library uses when it is
  absent: `MaxHeaderBytes: 1 << 20` is `http.DefaultMaxHeaderBytes`,
  `Content-Type: text/plain` before `w.Write([]byte("ok"))` is what the
  writer sniffs, and `false`, `0`, `nil`, or `""` in a keyed literal is the
  zero value. The line says nothing its absence does not.
- A failure branch, or a `fmt.Errorf` wrap, on a call that cannot fail for a
  value the function built itself — `json.Marshal` of its own document. The
  error is returned as it is (`return json.Marshal(doc)`). A write whose
  error has nowhere to go is discarded in the open, `_, _ = w.Write(body)`
  with its reason, never bare: `errcheck` in the gate reads a bare call as a
  finding, and the reason on that line is not a comment to delete.
- A doc comment on an unexported helper beyond one line;
  [go-documentation](../go-documentation/SKILL.md) owns the exported ones.

The pass runs behind the test, never ahead of it: a line a case needs stays,
and validation at a trust boundary, failure behavior, and security controls
are not deleted for a smaller diff. Rerun or reread the cases after it.

### Reader Pass

Read the changed path from its entry point to its result or side effect. Can a
reader find the main decision, error exits, and state the code tracks without
chasing wrappers that only rename an operation? Keep a one-use name when it
exposes a meaningful boundary or state; keep a helper when it hides details
that obscure the caller's decision. If either adds a jump without answering a
reader's question, inline it and run the cases again.

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
4. It is a step at another level of abstraction than the rest of the
   function — decoding a request body beside the business decision, building
   SQL beside the domain rule — even with one call site; the report names its
   caller and this rule. A helper whose name only restates two or three lines
   of its body is not such a step: those lines stay inline.

A function literal bound to a name — `writeJSON := func(w http.ResponseWriter,
v any) {...}` — that captures nothing from the function around it is a
package-level function written in the wrong place: it counts under the same
four rules, and when it earns its place it is a small unexported function of
the package with a one-line comment or none. A closure is for capturing state
— a mutex, a counter, the request being served — and a handler registered
once is written at its registration: it meets none of the four rules. In a
constructor that returns the `*http.Server`, a `type server struct` holding
what the routes read, one method per route, and a `handleHealthz` beside them
are four declarations for handlers registered once each; the values the
routes share are locals of the constructor that the handler literals capture.
The count is a record, not a score: a helper two call sites need is one
declaration named in the report, and hiding it inside a function changes the
number without changing the code.

A representation is a value, not a reason: a wire document, a formatted error,
and a sorted view take the [Plain Code](#plain-code) form. A helper that names
a step at its caller's own level of abstraction meets none of the four; write
the step inline.
The named helpers in skill examples (`validate`, `writeJSON`, `writeError`)
show the order of an operation, not a list to reproduce.

The budget is the restraint ladder applied per declaration; read the ladder in
[OVER-ENGINEERING.md](../go-code-refactor/references/OVER-ENGINEERING.md#the-restraint-ladder)
when a rung is in doubt. Never trade input validation at a trust boundary,
failure behavior, or a security control for a smaller count.

The report carries one line — `added package-level declarations: N` — and,
for each, its name and the call sites or caller that need it. A declaration
whose line cannot name them is removed before the report is written, not
explained in it.

## Close With The Gate

Select checks with [go-linting](../go-linting/SKILL.md#verification-gate) for
the requested scope; a repository gate replaces the defaults rather than
joining them. Bundled scripts add evidence the gate lacks:

- Refactor: `../go-code-refactor/scripts/verify-refactor.sh` for baseline/after
  evidence; `../go-code-refactor/scripts/check-debt.sh` for deliberate `Kept:`
  markers, whose format is owned by `go-code-refactor`.
- Error handling: `../go-error-handling/scripts/check-errors.sh`.
- Exported API documentation: `../go-documentation/scripts/check-docs.sh`.
- Before submitting: [go-code-review](../go-code-review/SKILL.md).

Run routine checks inline; Claude Code's `go-verify` agent is for checks the
user asks to delegate. The report carries only results observed for the
current diff and scope, each as `pass`, `fail`, `unavailable (reason)`, or
`skipped (reason)` per go-linting; a hook's silence is not one of them.

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

## Related Skills

- Every owner in [Route Before The First Edit](#route-before-the-first-edit).
- [go-style-core](../go-style-core/SKILL.md): house style, report length, and delegation.
- [go-linting](../go-linting/SKILL.md): the closing gate.
- [go-code-review](../go-code-review/SKILL.md): the checklist before submitting.
