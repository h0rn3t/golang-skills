# Navigate, Rename, and Extract with gopls

> Sources: golang.org/x/tools/gopls docs (`gopls help`, `gopls mcp -instructions`, the v0.23.0 MCP tool list); Claude Code LSP tool and `gopls-lsp` plugin docs
> Authority: advisory — the mechanics; behavior preservation rules stay in SKILL.md
> Minimum Go: gopls v0.20+ on PATH (`go install golang.org/x/tools/gopls@latest`)
> Last verified: 2026-09-27

Use `rg` for cheap textual discovery, then gopls to verify Go symbol meaning:
definitions, references, implementations, symbols, and type information.
Use `rg -n --column -g '*.go' 'Name'` to get a candidate position for CLI or
LSP queries, then read files at the locations gopls identifies. Skip `rg`
when the symbol's position is already known. An `rg` hit is not proof of a reference:
textual search can miss an interface implementation and can hit an unrelated
local with the same name. Avoid recursive file reads when gopls can locate
the relevant symbols.

## Three ways in

| Route | Addressing | Best for |
|---|---|---|
| gopls MCP server — `claude mcp add gopls -- gopls mcp` | Symbol names, file paths, fuzzy queries (`go_search`, `go_symbol_references`, `go_rename_symbol`, `go_diagnostics`, `go_package_api`, `go_file_context`) | Agent workflows: no cursor position needed |
| Native `LSP` tool (gopls wired as an LSP server, e.g. the `gopls-lsp@claude-plugins-official` plugin) | `line:character` (`findReferences`, `goToImplementation`, `workspaceSymbol`, `hover`, `documentSymbol`, call hierarchy); no rename, no code actions | After `rg` gives a location; diagnostics arrive after every edit for free |
| `gopls` CLI — `gopls workspace_symbol Name`, `gopls references file.go:12:6`, `gopls rename -w file.go:12:6 newName` | Name or `file:line:col` | Nothing else is wired; one-shot navigation and edits. Documented as experimental |

Which route is wired: an `LSP` tool in your tool list means the LSP route,
`go_*` tools mean MCP, `command -v gopls` succeeding means the CLI. If one
route is blocked or fails, try another available gopls route before treating
text hits as semantic references. The routes stack — the plugin supplies diagnostics and
positions, MCP supplies rename and name-based lookup — so use each for what it
does best rather than picking one.
Absent all three, fall back to `go build ./... && go vet ./...` after every
rename and accept that interface satisfaction breaks are found by the compiler,
not before the edit.

## Which tool answers which question

| Question | `LSP` tool | gopls MCP | gopls CLI |
|---|---|---|---|
| Where is `X` declared? | `workspaceSymbol` | `go_search` | `gopls workspace_symbol X` |
| Who uses this symbol? | `findReferences` | `go_symbol_references` | `gopls references file.go:line:col` |
| Who calls this function, and what does it call? | `incomingCalls` / `outgoingCalls` (static calls only) | `go_symbol_references` for uses | `gopls call_hierarchy file.go:line:col` |
| Which types implement this interface? | `goToImplementation` | — | `gopls implementation file.go:line:col` |
| What is this identifier's type and doc? | `hover` | `go_package_api` for a whole package | — |
| What does this file declare, and use from its package? | `documentSymbol` | `go_file_context` | `gopls symbols file.go` for declarations |
| Did the edit compile? | diagnostics pushed after `Edit`/`Write` | `go_diagnostics` | `gopls check file.go` |

`findReferences` and the CLI `references` command need a position. Use a
candidate's `file:line:column` from `rg --column`, or locate the declaration
with `workspaceSymbol`, `go_search`, or `gopls workspace_symbol`. CLI positions
use 1-based line and column numbers. Navigation tells
you where to read; it does not replace reading the code the change touches.

`gopls mcp -instructions` tells the agent to run `go_vulncheck` at session
start; the vulnerability scan belongs to the closing gate
([go-linting](../../go-linting/SKILL.md)), so run it there, once.

## Before touching a definition

1. **References, not grep.** `go_symbol_references` / `findReferences` on the
   symbol. The count is the blast radius; read every referencing file that
   needs a matching edit before the first change.
2. **Implementations both ways.** For a method: `goToImplementation` on the
   interface it might satisfy; for an interface: every type that implements it.
   Renaming `Close` on one type silently un-implements `io.Closer` — gopls's
   rename refuses that; a text replace does not.
3. **Exported symbol?** References outside the module are invisible to gopls
   too. Exported renames are a findings-list item, not a diff (PLAYBOOK §3).

## Applying the change

- **Rename**: gopls rename (`go_rename_symbol`, or `gopls rename -w` when
  only the `LSP` tool is wired) updates every reference in the workspace,
  including test files, doc comments that mention the identifier in backticks,
  and struct-literal field keys. It rejects a rename that would shadow or
  collide. Review the diff anyway — a reject is safe, an accept is merely
  consistent.
- **Extract / inline**: the `refactor.extract.*` and `refactor.inline.*` code
  actions preserve side-effect order by construction, but may drop comments
  and produce a six-parameter helper when the seam is wrong (PLAYBOOK §2).
  Extract, then read the signature; if it is ugly, the split is in the wrong
  place — undo rather than patch.
- **Fill struct / add tags / remove unused parameter**: `refactor.rewrite.*`
  actions; `removeUnusedParam` also updates every call site.
- **Generated files** (`// Code generated ... DO NOT EDIT`) receive no code
  actions. Trace and update their source inputs as described in `SKILL.md`.

## After every edit

`go_diagnostics` on each changed file (automatic with the native tool). Fix
compiler errors before moving to the next transformation; a half-applied
rename across two files is the state in which `verify-refactor.sh after` lies
to you — the package that failed to compile ran no tests at all. Re-test the
affected packages mid-step — the ones you changed plus the consumers of any
shared API you touched (`go test` on those paths, `-race` when concurrency
moved) — rather than the whole tree; the full suite belongs to the end of the
task, under the [go-linting](../../go-linting/SKILL.md) gate that owns this
scope rule.

## Gotchas

- The MCP server roots its workspace at the directory it was started in,
  not at the file you pass: in a repository whose `go.mod` sits in a
  subdirectory, `go_workspace` answers "not a Go workspace" and the other
  tools report "no package metadata". Use the `LSP` tool there, which finds
  the module per file; the fix, a server started from the module directory
  or a `go.work` at the root, is the user's setup, not the task's.

- References reflect the **build configuration of the queried file**: a query
  from `x_linux.go` does not see `x_windows.go`. Re-run under `GOOS=windows`
  when build-tagged files are in the package (SKILL.md, Orient).
- Call hierarchy shows **static** calls only; calls through function values or
  interface methods are invisible — corroborate with references.
- gopls reasons about the locally resolved build (`go.sum`, `replace`
  directives). It cannot tell you who imports your exported API from other
  modules.
