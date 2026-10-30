# Adding and Auditing Dependencies

> Sources: https://go.dev/ref/mod; https://go.dev/doc/toolchain; `go help mod`
> Authority: advisory
> Minimum Go: `tool` directive 1.24; `go mod tidy -diff` 1.23
> Last verified: 2026-10-30

The one-module-per-bump rule and checksum mismatches stay in
[SKILL.md](../SKILL.md#adding-and-auditing-dependencies).

- **After `go mod init`**, inspect the generated `go` directive against the
  local and CI toolchains: a toolchain writes its own patch level — 1.27.1
  writes `go 1.27.1`, which makes a 1.27.0 CI toolchain download 1.27.1, and
  1.26.4 writes `go 1.26.4` (only 1.26.0 wrote `go 1.25.0`). For a new module
  targeting Go 1.27, set `go mod edit -go=1.27.0`; do not infer the directive
  from the installed version or bump an existing module as an incidental
  cleanup.
  `go test ./...` includes `stdversion` in Go 1.27 and rejects standard-library
  APIs newer than the effective file version. `go fix` also gates replacements
  by supported version, so an empty preview alone does not prove a 1.27 target.
- **Before adding a module**, check the [Dependency Ladder](../SKILL.md#dependency-ladder), license compatibility,
  and maintenance status. Approval policy:
  [go-style-core](../../go-style-core/SKILL.md#house-style-wins).
- **`replace` and `exclude` apply only in the main module.** A library's
  directives are ignored by every module that imports it, and a `replace` to
  a local path never reaches a tag: a library ships its fix as a version
  bump of the dependency.
- **Pin executable tools with `go get -tool <package>@<version>`** (Go 1.24+),
  then run them via `go tool <name>`. Tool dependencies share the module graph.
  For golangci-lint, prefer a version-pinned release binary; if using `go tool`,
  isolate it in a dedicated module or modfile to avoid dependency conflicts
  ([upstream guidance](https://golangci-lint.run/docs/welcome/install/local/)).
  The legacy `tools.go` blank-import file is only for modules below Go 1.24.
- **Tidy before committing** dependency changes: check with `go mod tidy -diff`
  (Go 1.23+), which prints what tidy would change and exits non-zero without
  touching the module files or inspecting unrelated Git changes; then apply
  `go mod tidy`. On `go 1.27+` modules tidy also merges duplicate `require`
  blocks down to two — direct and indirect — so the first run after the bump
  can produce a large diff that changes no dependency.
  `go mod tidy && git diff --exit-code` is the clean-checkout form for CI.
- **Scan before releasing**: `govulncheck ./...`, pinned through the `tool`
  directive above. The gate lives in [go-linting](../../go-linting/SKILL.md);
  finding triage routes to [go-security](../../go-security/SKILL.md).
