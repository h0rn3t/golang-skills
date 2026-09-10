# Script JSON Contracts

Shell scripts keep their current command-line UX and emit stable JSON when
called with `--json`. Exit codes are part of the contract: `0` means no
findings or successful generation, `1` means findings or tool-reported issues,
and `2` means usage/environment errors.

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
{"interfaces":[{"name":"Reader","file":"path","line":1}],"missing":[{"name":"Reader","file":"path","line":1}],"count_interfaces":1,"count_missing":1,"truncated":false}
```

No-Go-file targets are successful empty scans and include a status marker:

```json
{"violations":[],"total":0,"truncated":false,"status":"no_go_files"}
{"missing":[],"total":0,"truncated":false,"status":"no_go_files"}
{"findings":[],"total":0,"truncated":false,"status":"no_go_files"}
{"markers":[],"total":0,"no_trigger":0,"truncated":false,"status":"no_go_files"}
{"interfaces":[],"missing":[],"count_interfaces":0,"count_missing":0,"truncated":false,"status":"no_go_files"}
```

## Tool Scripts

`go-code-review/scripts/pre-review.sh`:

```json
{"gofmt":{"status":"pass","files":[]},"govet":{"status":"pass","output":""},"golangci_lint":{"status":"skip","output":""},"passed":true}
```

`go-linting/scripts/setup-lint.sh`:

```json
{"config_path":".golangci.yml","local_prefix":"","created":true,"lint_issues":false,"lint_output":""}
```

`go-performance/scripts/bench-compare.sh`:

```json
{"count":1,"package":"./...","filter":".","benchmarks_found":1,"baseline":"","save":"","status":"ok","exit_code":0,"go_exit_code":0,"output":"Benchmark..."}
```

`status` is `ok`, `no_benchmarks` (go test passed but no `Benchmark` line
matched the filter; the script exits 1), or `error` (go test failed; exit 1).
`exit_code` is the script's own exit code, the one the process returns;
`go_exit_code` is go test's, which is 0 in the `no_benchmarks` case.

`go-testing/scripts/gen-table-test.sh`:

```json
{"func":"ParseConfig","package":"config","output_file":"","parallel":false,"written":false}
```

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

`fix_pending_lines` is `"n/a"` when its check cannot run. Lint metadata:

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
