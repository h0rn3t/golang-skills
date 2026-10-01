#!/usr/bin/env bash
set -euo pipefail

VERSION="1.4.0"
SCRIPT_NAME="$(basename "$0")"

usage() {
    cat <<EOF
$SCRIPT_NAME v$VERSION — Run Go benchmarks with optional comparison

USAGE
    bash $SCRIPT_NAME [options] [package]

DESCRIPTION
    Wrapper around 'go test -bench' that runs benchmarks multiple times and
    optionally compares results against a saved baseline using benchstat.

    Results can be saved to a file for future comparison; --save writes only
    a run whose status is ok. With a baseline, the comparison uses
    'go tool benchstat' when go.mod declares the tool, else a benchstat on
    PATH; with neither, raw result lines are shown and the comparison is
    reported as skipped.

EXIT CODES
    0    Benchmarks ran successfully
    1    go test failed (compilation error, test failure, no benchmarks found)
    2    Usage error (missing arguments or option values, bad flags, file
         exists without --force, baseline missing or without Benchmark lines)

    With --json the object carries "status" (ok | no_benchmarks | error) and
    "exit_code", this script's exit code; "go_exit_code" is go test's own.
    "benchmarks_found" counts result lines: one per benchmark per --count run.
    Usage errors exit 2 before any benchmark runs and print no JSON.

OPTIONS
    -h, --help           Show this help message
    -v, --version        Show version
    -n, --count N        Number of benchmark iterations (default: 10)
    -b, --baseline FILE  Compare results against this baseline file
    -s, --save FILE      Save benchmark results to this file
    -f, --filter REGEX   Benchmark filter regex (default: ".")
    --json               Output metadata as JSON (human output goes to stderr)
    --benchmem           Include memory allocation stats (default: on)
    --no-benchmem        Disable memory allocation stats
    --force              Allow --save to overwrite existing files
    --limit N            Max benchmark result lines to include (default: 0 = all)

ARGUMENTS
    package              Go package to benchmark (default: ./...)

EXAMPLES
    bash $SCRIPT_NAME
    bash $SCRIPT_NAME -n 10 ./pkg/parser
    bash $SCRIPT_NAME --save baseline.txt ./...
    bash $SCRIPT_NAME --baseline baseline.txt --save current.txt ./...
    bash $SCRIPT_NAME --filter BenchmarkSort -n 3
    bash $SCRIPT_NAME --json --limit 5 ./...
    bash $SCRIPT_NAME --save results.txt --force ./...
EOF
}

json_escape() {
    local s="$1" i hex c rep
    s="${s//\\/\\\\}"
    s="${s//\"/\\\"}"
    s="${s//$'\t'/\\t}"
    s="${s//$'\r'/}"
    s="${s//$'\n'/\\n}"
    # Any other control character (an ANSI color in b.Log output, say) would
    # make the JSON invalid; bash strings cannot hold NUL.
    for ((i = 1; i < 32; i++)); do
        printf -v hex '%02x' "$i"
        printf -v c "\\x$hex"
        if [[ $s == *"$c"* ]]; then
            printf -v rep '\\u%04x' "$i"
            s="${s//"$c"/"$rep"}"
        fi
    done
    printf '%s' "$s"
}

# Exit 2 when an option that takes a value is last or given an empty one.
need_arg() {
    if [[ -z ${2-} ]]; then
        echo "error: $1 requires a value" >&2
        exit 2
    fi
}

# Pass go test output through with at most LIMIT benchmark result lines
# (0 = all); every other line passes unchanged.
limit_results() {
    awk -v limit="$LIMIT" '/^Benchmark/ { if (limit > 0 && ++seen > limit) next } { print; fflush() }'
}

# Print human-readable output: stdout in text mode, stderr in JSON mode.
log() {
    if $JSON_OUTPUT; then
        echo "$@" >&2
    else
        echo "$@"
    fi
}

COUNT=10
BASELINE=""
SAVE=""
FILTER="."
PACKAGE=""
JSON_OUTPUT=false
BENCHMEM=true
FORCE=false
LIMIT=0

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)      usage; exit 0 ;;
        -v|--version)   echo "$SCRIPT_NAME v$VERSION"; exit 0 ;;
        -n|--count)     need_arg "$1" "${2-}"; COUNT="$2"; shift 2 ;;
        -b|--baseline)  need_arg "$1" "${2-}"; BASELINE="$2"; shift 2 ;;
        -s|--save)      need_arg "$1" "${2-}"; SAVE="$2"; shift 2 ;;
        -f|--filter)    need_arg "$1" "${2-}"; FILTER="$2"; shift 2 ;;
        --json)         JSON_OUTPUT=true; shift ;;
        --benchmem)     BENCHMEM=true; shift ;;
        --no-benchmem)  BENCHMEM=false; shift ;;
        --force)        FORCE=true; shift ;;
        --limit)        need_arg "$1" "${2-}"; LIMIT="$2"; shift 2 ;;
        -*)             echo "error: unknown option: $1" >&2; usage >&2; exit 2 ;;
        *)              PACKAGE="$1"; shift ;;
    esac
done

PACKAGE="${PACKAGE:-./...}"

if ! command -v go &>/dev/null; then
    echo "error: 'go' command not found in PATH" >&2
    exit 2
fi

if ! [[ "$COUNT" =~ ^[1-9][0-9]*$ ]]; then
    echo "error: --count must be a positive integer, got: $COUNT" >&2
    exit 2
fi

if ! [[ "$LIMIT" =~ ^[0-9]+$ ]]; then
    echo "error: --limit must be a non-negative integer, got: $LIMIT" >&2
    exit 2
fi

if [[ -n "$BASELINE" && ! -f "$BASELINE" ]]; then
    echo "error: baseline file not found: $BASELINE" >&2
    exit 2
fi

# A baseline without result lines (a failed run saved by an earlier version)
# would let the comparison "pass" with nothing compared.
if [[ -n "$BASELINE" ]] && ! grep -qE '^Benchmark' "$BASELINE"; then
    echo "error: baseline has no Benchmark result lines: $BASELINE" >&2
    exit 2
fi

if [[ -n "$SAVE" && -f "$SAVE" ]] && ! $FORCE; then
    echo "error: save target already exists: $SAVE (use --force to overwrite)" >&2
    exit 2
fi

# Prefer the version go.mod pins (go get -tool), then a benchstat on PATH.
BENCHSTAT=()
if [[ -n "$BASELINE" ]]; then
    if go tool -n benchstat &>/dev/null; then
        BENCHSTAT=(go tool benchstat)
    elif command -v benchstat &>/dev/null; then
        BENCHSTAT=(benchstat)
    fi
fi

BENCH_ARGS=(-bench "$FILTER" -count "$COUNT" -run '^$')
if $BENCHMEM; then
    BENCH_ARGS+=(-benchmem)
fi

# The X's must be trailing: BSD mktemp leaves any other template literal.
TMPFILE=$(mktemp "${TMPDIR:-/tmp}/bench.XXXXXX")
trap 'rm -f "$TMPFILE"' EXIT

log "Running benchmarks: go test ${BENCH_ARGS[*]} $PACKAGE"
log "Iterations: $COUNT"
log ""

GO_EXIT=0
if $JSON_OUTPUT; then
    go test "${BENCH_ARGS[@]}" "$PACKAGE" 2>&1 | tee "$TMPFILE" | limit_results >&2 || GO_EXIT=$?
else
    go test "${BENCH_ARGS[@]}" "$PACKAGE" 2>&1 | tee "$TMPFILE" | limit_results || GO_EXIT=$?
fi

BENCH_COUNT=$(grep -cE '^Benchmark' "$TMPFILE" || true)

TRUNCATED=false
if [[ $LIMIT -gt 0 && $BENCH_COUNT -gt $LIMIT ]]; then
    TRUNCATED=true
fi

if $TRUNCATED; then
    log ""
    log "Note: $BENCH_COUNT benchmark result lines found, showing the first $LIMIT (--limit $LIMIT)"
fi

# STATUS names the outcome so a JSON consumer cannot mistake go test's exit 0
# on a package without benchmarks for success; exit_code below is FINAL_EXIT,
# the code this script exits with, never go test's.
FINAL_EXIT=0
STATUS="ok"
if [[ $GO_EXIT -ne 0 ]]; then
    FINAL_EXIT=1
    STATUS="error"
elif [[ $BENCH_COUNT -eq 0 ]]; then
    FINAL_EXIT=1
    STATUS="no_benchmarks"
fi

# Only a successful run becomes a baseline; SAVED stays empty otherwise.
SAVED=""
if [[ -n "$SAVE" ]]; then
    log ""
    if [[ $STATUS == ok ]]; then
        cp "$TMPFILE" "$SAVE"
        SAVED="$SAVE"
        log "Results saved to: $SAVE"
    else
        log "Results not saved to $SAVE: status $STATUS"
    fi
fi

if [[ -n "$BASELINE" ]]; then
    log ""
    log "=== Comparison with baseline: $BASELINE ==="
    log ""
    if [[ ${#BENCHSTAT[@]} -gt 0 ]]; then
        if $JSON_OUTPUT; then
            "${BENCHSTAT[@]}" "$BASELINE" "$TMPFILE" >&2 || true
        else
            "${BENCHSTAT[@]}" "$BASELINE" "$TMPFILE" || true
        fi
    else
        log "note: benchstat not found; statistical comparison skipped. Pin it in the project:"
        log "  go get -tool golang.org/x/perf/cmd/benchstat@latest"
        log ""
        log "--- Baseline ---"
        if $JSON_OUTPUT; then
            grep -E '^Benchmark' "$BASELINE" | limit_results >&2 || true
        else
            grep -E '^Benchmark' "$BASELINE" | limit_results || true
        fi
        log ""
        log "--- Current ---"
        if $JSON_OUTPUT; then
            grep -E '^Benchmark' "$TMPFILE" | limit_results >&2 || true
        else
            grep -E '^Benchmark' "$TMPFILE" | limit_results || true
        fi
    fi
fi

if ! $JSON_OUTPUT; then
    if [[ $STATUS == error ]]; then
        log ""
        log "error: go test exited with code $GO_EXIT"
    elif [[ $STATUS == no_benchmarks ]]; then
        log ""
        log "error: no benchmarks found matching filter: $FILTER"
    fi
fi

if $JSON_OUTPUT; then
    BENCH_OUTPUT=$(limit_results < "$TMPFILE")

    escaped_package=$(json_escape "$PACKAGE")
    escaped_filter=$(json_escape "$FILTER")
    escaped_baseline=$(json_escape "$BASELINE")
    escaped_save=$(json_escape "$SAVED")
    escaped_output=$(json_escape "$BENCH_OUTPUT")

    printf '{"count":%d,' "$COUNT"
    printf '"package":"%s",' "$escaped_package"
    printf '"filter":"%s",' "$escaped_filter"
    printf '"benchmarks_found":%d,' "$BENCH_COUNT"
    printf '"baseline":"%s",' "$escaped_baseline"
    printf '"save":"%s",' "$escaped_save"
    printf '"status":"%s",' "$STATUS"
    printf '"exit_code":%d,' "$FINAL_EXIT"
    printf '"go_exit_code":%d,' "$GO_EXIT"
    printf '"output":"%s"' "$escaped_output"
    if $TRUNCATED; then
        printf ',"truncated":true'
    fi
    printf '}\n'
fi

exit $FINAL_EXIT
