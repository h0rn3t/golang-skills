# Language Version Checks

> Sources: https://go.dev/ref/mod#go-mod-file-go; https://go.dev/doc/toolchain; https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/stdversion
> Authority: normative
> Last verified: 2026-10-30

A language form newer than the directive fails to compile with `requires
go1.NN or later (-lang was set to go1.MM; check go.mod)` — except
function-type inference in a conversion or composite-literal element (Go 1.27)
and a constraint that names its own type parameter (Go 1.26), which compile at
an older directive on a newer toolchain and fail only on the older one
(the plugin repository's COMPATIBILITY.md); a library symbol
newer than it builds, and only `go vet` (`stdversion`) reports it:
`strings.CutLast requires go1.27 or later (module is go1.26)`. A file whose
`//go:build` line names a Go version is checked at that version instead,
above or below the directive and never below 1.21, by the compiler and
`stdversion` alike: in a `go 1.24` module, a file under `//go:build go1.21`
rejects `for range n`.
