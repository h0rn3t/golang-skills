# Script JSON Contracts

Shell scripts keep their current command-line UX and emit stable JSON when
called with `--json`. Exit codes are part of the contract: `0` means no
findings or successful generation, `1` means findings or tool-reported issues,
and `2` means usage/environment errors or, for `check-naming.sh` and
`check-docs.sh`, a file that does not parse.

## Findings Scripts

`go-naming/scripts/check-naming.sh`:

```json
{"violations":[{"file":"path","line":1,"rule":"rule-id","message":"text"}],"total":1,"truncated":false}
```

`go-documentation/scripts/check-docs.sh`:

```json
{"missing":[{"file":"path","line":1,"kind":"type","name":"Name"}],"total":1,"truncated":false}
```

`go-error-handling/scripts/check-errors.sh`:

```json
{"findings":[{"file":"path","line":1,"rule":"rule-id","message":"text"}],"total":1,"truncated":false}
```

`go-code-refactor/scripts/check-debt.sh` — exit 1 counts only `no-trigger`
markers, since a marker naming a ceiling and a fix is tracked debt, not a
finding:

```json
{"markers":[{"file":"path","line":1,"ceiling":true,"upgrade":true,"rule":"tracked","note":"text"}],"total":1,"no_trigger":0,"truncated":false}
```

`go-interfaces/scripts/check-interface-compliance.sh`:

```json
{"interfaces":[{"name":"Reader","file":"path","line":1},{"name":"Store","file":"path","line":9}],"missing":[{"name":"Reader","file":"path","line":1},{"name":"Store","file":"path","line":9,"returned_by":"NewStore"}],"count_interfaces":2,"count_missing":2,"truncated":false}
```

A `missing` entry is an exported interface implemented in its own package that
nothing there converts to (assignment, return, argument, composite-literal
element, send, or conversion). An entry with `returned_by` is one that the
named exported function of the same package returns while no function or
method there takes it as a parameter. Its `file` and `line` are the
interface's. The script exits 1 when `missing` is non-empty, and 2 when its
helper does not build.

`go-code-refactor/scripts/check-architecture.sh` — exit 1 on any violation,
including `stale`, `duplicate`, and `unexplained` entries in the `known` list
of `architecture.json`; paths are module-relative:

```json
{"module":"example.com/shop","layout":"modules","checked":["Imports"],"violations":[{"rule":"ownership","from":"internal/billing/services","to":"internal/order/repositories","message":"text"}],"total":1,"suppressed":0,"truncated":false}
```

A module with nothing under `internal/` is a successful empty check:

```json
{"module":"example.com/shop","layout":"modules","checked":["Imports"],"violations":[],"total":0,"suppressed":0,"truncated":false,"status":"no_internal_packages"}
```

No-Go-file targets are successful empty scans and include a status marker:

```json
{"violations":[],"total":0,"truncated":false,"status":"no_go_files"}
{"missing":[],"total":0,"truncated":false,"status":"no_go_files"}
{"findings":[],"total":0,"truncated":false,"status":"no_go_files"}
{"markers":[],"total":0,"no_trigger":0,"truncated":false,"status":"no_go_files"}
{"interfaces":[],"missing":[],"count_interfaces":0,"count_missing":0,"truncated":false,"status":"no_go_files"}
```

A file that does not parse does not stop `check-naming.sh` or `check-docs.sh`:
the other files are still checked and their findings reported, the JSON adds
`"status":"parse_error"` and a `parse_errors` list, and the run exits `2`:

```json
{"violations":[{"file":"path","line":1,"rule":"rule-id","message":"text"}],"total":1,"truncated":false,"status":"parse_error","parse_errors":[{"file":"path","message":"path:3:14: expected ')', found '{'"}]}
{"missing":[{"file":"path","line":1,"kind":"type","name":"Name"}],"total":1,"truncated":false,"status":"parse_error","parse_errors":[{"file":"path","message":"path:3:14: expected ')', found '{'"}]}
```

## Tool Scripts

`go-code-review/scripts/pre-review.sh`:

```json
{"gofmt":{"status":"pass","files":[]},"govet":{"status":"pass","output":""},"golangci_lint":{"status":"pass","output":""},"passed":true}
```

`golangci_lint.status` is `pass`, `fail` (golangci-lint exit 1), or
`unavailable` (not installed, or any other non-zero exit; `output` carries its
message). Without a project golangci-lint config the script lints with
`go-linting/assets/golangci.yml`. `--strict` turns `unavailable` into exit 2.

`go-linting/scripts/setup-lint.sh`:

```json
{"config_path":".golangci.yml","local_prefix":"","created":true,"lint_issues":false,"lint_output":""}
```

It exits 2 with no JSON when golangci-lint is missing (checked before the
config is written) or `golangci-lint run` exits with anything but 0 or 1, for
example 5 for no Go files; a config written by then stays.

`go-performance/scripts/bench-compare.sh`:

```json
{"count":1,"package":"./...","filter":".","benchmarks_found":1,"baseline":"","save":"","status":"ok","exit_code":0,"go_exit_code":0,"output":"Benchmark..."}
```

`status` is `ok`, `no_benchmarks` (go test passed but no `Benchmark` line
matched the filter; the script exits 1), or `error` (go test failed; exit 1).
`exit_code` is the script's own exit code, the one the process returns;
`go_exit_code` is go test's, which is 0 in the `no_benchmarks` case.
`benchmarks_found` counts `Benchmark` result lines in the go test output: one
per benchmark per `--count` run, so two benchmarks at `-n 3` give 6.
`--limit N` keeps the first N of those lines in `output` and in the human
output, and sets `truncated` when it cut any.

`go-testing/scripts/gen-table-test.sh`:

```json
{"func":"ParseConfig","package":"config","output_file":"","parallel":false,"written":false}
```

With `--json` and no `--output`, stdout carries only this object and the
scaffold goes to stderr; `written` is `false` and `output_file` is empty.

`go-code-refactor/scripts/verify-refactor.sh` — one shape per mode.
`baseline` and `after`:

```json
{"mode":"after","target":"./...","toolchain":"go1.27.1","go_directive":"1.27","gofmt":"pass","summary_path":".refactor-verify/after.summary","fix_pending_lines":"0","lint_findings":"0","lint_status":"pass","lint_exit_code":0,"lint_log_path":".refactor-verify/after.lint.raw","passed":true}
```

`diff` (exit 1 when `identical` is false) and `leaks`:

```json
{"mode":"diff","identical":true,"diff":"","truncated":false}
{"mode":"leaks","go_minor":27,"passed":false,"tests_passed":true,"leaks_checked":false,"reason":"No in-process leak profile was collected or inspected","output":"ok\tscratch\t0.2s","truncated":false}
```

`leaks` exits 3 when tests pass but no leak profile was checked, 1 when tests
fail, and 2 for usage/environment errors. `truncated` reports whether `--limit`
actually shortened the `diff` or `output` field.

`loc-baseline` and `loc-diff` print indented JSON. Each count object is
`{"root":"/abs/dir","physical":9,"code":8,"files":2,"test_files":0,"scan_errors":0,"detail":[{"path":"sum.go","physical":8,"code":7,"sha256":"…"}]}`:

```json
{"mode":"loc-baseline","counts":{…},"record":"/abs/.refactor-verify/loc.baseline.json","convention":"physical: …\ncode: …"}
{"mode":"loc-diff","root":"/abs/dir","before":{…},"after":{…},"delta":{"physical":2,"code":0,"files":0},"added":null,"removed":null,"gate_pass":false,"convention":"physical: …\ncode: …"}
```

`loc-diff` exits 0 when `gate_pass` is true and 1 when a count grew. It exits 2,
with an error on stderr and no JSON, when no record exists or the record's
`root` is not this run's directory. `added` and `removed` list production files
by relative path, or are `null` when none changed.

`fix_pending_lines` is the line count of the `go fix -diff` output, which exits
1 when that diff is non-empty; it is `"n/a"` when a package fails to load or
the tool fails (anything on stderr beyond package headers and
skipped-alternative-fix notices). Lint metadata:

- `lint_status`: `pass` after exit 0, `fail` after exit 1 with parsed findings,
  otherwise `unavailable`; failures without usable diagnostics are not clean.
- `lint_exit_code`: actual exit code, or `null` if the executable is absent.
- `lint_findings`: count of parsed text diagnostics, or `"n/a"` when unavailable.
- `lint_log_path`: captured output for inspection, or empty if lint did not run.

Lint and modernization remain informational and excluded from the summaries
that `diff` compares. In `baseline`/`after`, `passed` covers the core checks,
not lint; callers must inspect `lint_status` and honor their repository gate.

## Migration Note

Keep shell wrappers as the public interface. The findings scripts —
documentation, interface compliance, naming, and error flow — are Go AST
helpers behind a wrapper that builds them once into the user cache. The
remaining regex-based scripts (`check-debt.sh`, `pre-review.sh`,
`verify-refactor.sh`, `bench-compare.sh`, `setup-lint.sh`, `gen-table-test.sh`)
orchestrate tools or match fixed markers, where regex is the right tool.

A single `go/analysis` multichecker across skills was considered and rejected:
each skill directory must stay installable on its own, so a shared module
outside `skills/go-*/` would break single-skill installs.
