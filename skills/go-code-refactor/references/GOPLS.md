# Navigate, Rename, and Extract with gopls

> Sources: golang.org/x/tools/gopls docs (`gopls help`, `gopls mcp -instructions`, the v0.23.0 MCP tool list); Claude Code LSP tool and `gopls-lsp` plugin docs
> Authority: advisory — the mechanics; behavior preservation rules stay in SKILL.md
> Minimum Go: gopls v0.20+ on PATH (`go install golang.org/x/tools/gopls@latest`)
> Last verified: 2026-09-27

When MCP tools are available, start unknown Go symbol searches with `go_search`;
it accepts a name without a file position. Use `rg` directly for literal text.
For position-based LSP/CLI operations, `rg -n --column -g '*.go' 'Name'` can
find a candidate, which gopls must verify. An `rg` hit is not a semantic
reference: it can miss implementations or hit an unrelated local. Read only
the files and declarations needed after navigation.

## Three ways in

| Route | Addressing | Best for |
|---|---|---|
| gopls MCP server — `claude mcp add gopls -- gopls mcp` | Symbol names, file paths, fuzzy queries (`go_search`, `go_symbol_references`, `go_rename_symbol`, `go_diagnostics`, `go_package_api`, `go_file_context`) | Agent workflows: no cursor position needed |
| Native `LSP` tool (gopls wired as an LSP server, e.g. the `gopls-lsp@claude-plugins-official` plugin) | `line:character` (`findReferences`, `goToImplementation`, `workspaceSymbol`, `hover`, `documentSymbol`, call hierarchy); no rename, no code actions | When a position is known or found; diagnostics arrive after every edit for free |
| `gopls` CLI — `gopls workspace_symbol Name`, `gopls references file.go:12:6`, `gopls rename -w file.go:12:6 newName` | Name or `file:line:col` | Nothing else is wired; one-shot navigation and edits. Documented as experimental |

Which route is wired: an `LSP` tool in your tool list means the LSP route,
`go_workspace` and `go_search` tools mean MCP, `command -v gopls` succeeding means the CLI. If one
route is blocked or fails, try another available gopls route before treating
text hits as semantic references. Use each route for the operation it exposes.
Do not infer MCP availability from `command -v gopls`, or install or restart
tooling merely to satisfy navigation instructions.
Absent all three, report uses, implementations, and callers as unverified;
do not perform a semantic rename from text hits alone. Build and vet can catch
some errors, but do not establish the complete set of relationships.

## Which tool answers which question

| Question | `LSP` tool | gopls MCP | gopls CLI |
|---|---|---|---|
| Where is `X` declared? | `workspaceSymbol` | `go_search` | `gopls workspace_symbol X` / `gopls definition file.go:line:col` |
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

The MCP server does not automatically deliver its workflow instructions on
connect; `gopls mcp -instructions` displays them. Its session-start
`go_vulncheck` instruction does not supersede this repository's
[verification policy](../../go-linting/SKILL.md): scan for dependency or
security work, or when the gate requires it.

## MCP call examples

Use the installed tool names and schemas; the host may add a server prefix.
Paths below stand for real absolute paths. In particular, the current
`go_symbol_references` schema uses `symbol`, not `name`.

```text
go_search({"query":"UserService"})
go_file_context({"file":"/repo/internal/users/service.go"})
go_symbol_references({"file":"/repo/internal/users/service.go","symbol":"UserService.Create"})
go_package_api({"packagePaths":["example.com/project/internal/users"]})
```

`go_file_context` identifies same-package file dependencies; read only the
relevant declarations. `go_package_api` accepts multiple package paths. There
is no dedicated MCP implementation search: use native LSP
`goToImplementation` or CLI `gopls implementation`, or state the limit.
References are not a full implementation list.

## Before touching a definition

1. **References, not grep.** `go_symbol_references` / `findReferences` on the
   symbol. The count is the blast radius; read every referencing file that
   needs a matching edit before the first change.
2. **Implementations both ways.** When native LSP or CLI is available, query
   implementations for a method's interface or for an interface itself.
   Renaming `Close` on one type silently un-implements `io.Closer` — gopls's
   rename refuses that; a text replace does not.
3. **Exported symbol?** References outside the module are invisible to gopls
   too. Exported renames are a findings-list item, not a diff (PLAYBOOK §3).

## Applying the change

- **Rename**: `go_rename_symbol` returns edits (currently a unified diff):
  review, apply, and diagnose them; investigate a refusal rather than work
  around it. When only CLI is available, preview with
  `gopls rename -d file.go:line:col NewName`, then write with `-w`. A successful rename updates workspace references,
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

## After a coherent edit batch

Call `go_diagnostics` once with all changed Go paths. It checks parse and build
errors across the workspace; `files` selects active files for extra analysis,
so per-file calls duplicate work. Native LSP diagnostics may arrive after
each edit automatically. Fix compiler errors before the next transformation;
a half-applied
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
