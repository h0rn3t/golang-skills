---
name: go-packages
description: Use when creating or splitting Go packages, organizing imports or dependencies, or structuring a new Go project. Individual identifier naming belongs to go-naming.
---

# Go Packages and Imports

> Compatibility: Baseline Go 1.27 (see the plugin repository's COMPATIBILITY.md). The `tool`
> directive requires Go 1.24+; stdlib `uuid` and `encoding/json/v2` Go 1.27+.

## Resource Routing

- `references/IMPORTS.md` - Read when grouping imports, using blank imports, dot imports, or import aliases.
- `references/PACKAGE-SIZE.md` - Read when splitting packages, avoiding `init`, structuring `main`, or designing CLI flags/subcommands.
- `references/DEPENDENCIES.md` - Read when running `go mod init` or setting the `go` directive, adding a module, using `replace` or `exclude`, pinning a tool, tidying, or scanning before a release.

## Dependency Ladder

> **Normative**: Before adding a module to `go.mod`, check the stdlib. Go 1.27
> absorbed several of the most-added dependencies.

Stop at the first rung that works:

1. **Standard library** — check `go doc <pkg>` before assuming it is missing
2. **A module already in `go.mod`** — reuse a suitable supported API
3. **A new module**, including `golang.org/x/...` — when the above do not meet
   the contract or would require materially more implementation and maintenance

Judge suitability by required semantics, supported versions, and maintenance,
not line count alone. Existing project conventions take precedence; this ladder
does not require replacing a working dependency in neighboring code. A
dependency is a convention; a standard-library idiom is not
([Write Current Go](../go-style-core/SKILL.md#write-current-go)).

Commonly added modules the standard library now covers:

| Was | Use instead | Since |
|---|---|---|
| `github.com/google/uuid` | `uuid` — `New`/`NewV4`/`NewV7`, `Parse`/`MustParse`, `Nil`/`Max`, `Compare`, text marshaling | 1.27 |
| A third-party JSON library kept for speed | `encoding/json`, which runs on the v2 implementation (unmarshal faster, marshal at parity); `encoding/json/v2` + `encoding/json/jsontext` for v2 semantics or streaming | 1.27 |
| `github.com/sirupsen/logrus`, `go.uber.org/zap` | `log/slog` (see [go-logging](../go-logging/SKILL.md)) | 1.21 |
| `github.com/pkg/errors` | `fmt.Errorf` with `%w`, `errors.Is`, `errors.AsType` | 1.13 / 1.26 |
| `golang.org/x/exp/slices`, `.../maps` | `slices`, `maps` | 1.21 |
| A CSPRNG string helper | `crypto/rand.Text` | 1.24 |

`encoding/json/v2` is not a drop-in replacement for `encoding/json`: it changes
`omitempty`, case-matching, and error semantics. Keep v1 for existing wire
formats; reach for v2 for new code or when you need `jsontext` streaming. Both
ship in the toolchain — no build tag or `GOEXPERIMENT` needed on Go 1.27.
For I/O forms and compatibility options, read
[JSON v2 at API boundaries](../go-http/references/JSON-V2.md).

`github.com/google/uuid` stays justified for the algorithms stdlib `uuid` does
not have (v1, v3, v5, custom sources).

---

## Adding and Auditing Dependencies

- **Upgrade one module per change** — `go get <module>@<version>`, not
  `go get -u ./...` — so a failure names its cause and the revert is one line;
  modules that release together count as one. A minor bump needs its release
  notes read, a major one its migration notes, and a `v0` minor may break.
  Minimal version selection raises the indirect requirements the new version
  declares: the `// indirect` and `go.sum` lines are part of the diff, and
  `go mod graph | grep <module>@` names who pulled a surprise version. Tests
  run before and after; `govulncheck` covers the new versions.
- **A checksum mismatch is a finding.** `SECURITY ERROR` from the `go` command
  means the bytes differ from `go.sum` or the checksum database: report the
  module, version, and both hashes. Never edit `go.sum` or set `GOSUMDB`,
  `GONOSUMDB`, or `GOPRIVATE` to get past it; `go clean -modcache` fixes only
  what `go mod verify` reports as `dir has been modified`.

`go mod init` and the `go` directive, the pre-add check, `replace` and
`exclude`, tool pinning, tidy, and the release scan:
[DEPENDENCIES.md](references/DEPENDENCIES.md).

---

## Package Organization

### Layout

- `cmd/<binary>` holds each `main` package; everything it imports lives
  elsewhere.
- `internal/` holds packages that are not API. Use it for anything you do not
  want to support; do not wrap the whole module in `internal/pkg`. A layer tree
  (`internal/handlers`, `internal/services`, `internal/repositories`,
  `internal/models`) fits one cohesive service; once several domains change
  independently, the same layer names move under `internal/<module>/`.
- Restructuring an existing module — a god package, a global layer tree
  several domains share, a monolith heading for modules — is
  [go-code-refactor's ARCHITECTURE.md](../go-code-refactor/references/ARCHITECTURE.md):
  the import-graph audit, the target shapes, the staged move, and the
  boundary test. This skill decides where a *new* package goes.
- A package named `util`, `helper`, or `common` is a finding:
  [go-naming](../go-naming/references/IDENTIFIERS.md#package-names) owns the rule.

### Package Size

**Don't split** just because a file is long, to create single-type packages, or
if it would create circular dependencies.

---

## Imports

| Rule | Guidance |
|------|----------|
| Blank imports (`import _`) | Only in `main` packages or tests |
| `import _ "embed"` | Any package, in the file whose `//go:embed` fills a `string` or `[]byte` — the compiler requires it |

---

## Avoid init()

Avoid `init()` where possible. When unavoidable, it must be:

1. Completely deterministic
2. Independent of other `init()` ordering
3. Free of environment state and I/O

---

## Exit in Main

Call `os.Exit` or `log.Fatal*` **only in `main()`**. All other functions should
return errors.

**Why**: Non-obvious control flow, untestable, `defer` statements skipped.

**Best practice**: Use the `run()` pattern — extract logic into a `run` that
returns an error, called from `main()` with a single exit point:

```go
func main() {
    if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

A CLI's `run` takes `args []string, stdout, stderr io.Writer` and builds its
flag set or command tree inside: a package-level `FlagSet` or cobra root keeps
flags and state from one test to the next.

---

## Command-Line Flags

> **Advisory**: Define flags only in `package main`.

- Follow the project's flag naming convention; this pack defaults to
  `snake_case` (`--output_dir`) when none exists.
- Libraries should accept configuration as parameters, not read flags directly —
  this keeps them testable and reusable
- Prefer the standard `flag` package, which already accepts `-v` and `--v` as
  the same flag; use `pflag` only when POSIX/GNU conventions are required —
  grouped single-letter flags (`-xvf`) or short/long pairs (`-v`/`--verbose`)

---

## Build Constraints and Embedded Files

- Platform-specific code lives in `_linux.go` / `_windows.go` suffix files or
  behind `//go:build linux`; a runtime `if runtime.GOOS == ...` switch is the
  last resort. Every constrained file needs a fallback so the package still
  compiles under `go vet ./...` on any GOOS.
- Static assets (templates, SQL, schemas) ship via `//go:embed` in the package
  that uses them — never by reading a relative path at runtime, which breaks
  as soon as the binary runs from another directory.

```go
//go:embed schema/*.sql
var schemaFS embed.FS
```

---

## Related Skills

- [go-naming](../go-naming/SKILL.md): package names, stuttering, exported symbols.
- [go-error-handling](../go-error-handling/SKILL.md): `%w` versus `%v` at package boundaries.
- [go-linting](../go-linting/SKILL.md): goimports local prefixes, import grouping.
- [go-defensive](../go-defensive/SKILL.md): replacing `init()`, mutable globals.
