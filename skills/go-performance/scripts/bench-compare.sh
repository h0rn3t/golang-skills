#!/usr/bin/env bash
set -euo pipefail

VERSION="1.5.1"
SCRIPT_NAME="$(basename "$0")"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

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

    --snapshot and --verdict decide keep or revert without reading a table.
    --snapshot DIR builds one package's test binary into DIR before the edit
    and copies its _test.go files there. --verdict DIR builds the current
    code with those frozen test files (go test -overlay), so a benchmark
    edited after the snapshot is not what gets compared; runs the two
    binaries one after the other, alternating which goes first, --count
    times in the package directory; and passes both outputs to
    bench-verdict.go. The verdict is keep for a significant improvement
    (sec/op, B/op or allocs/op by at least --threshold percent) with no
    significant regression (sec/op or allocs/op by the threshold);
    noisy when nothing changed by the threshold and a sec/op interval is
    wider than it; revert otherwise. A toolchain change since the snapshot is
    an error: re-snapshot on the new toolchain.

EXIT CODES
    0    Benchmarks ran successfully
    1    go test failed (compilation error, test failure, no benchmarks found);
         with --verdict also a verdict of revert or noisy
    2    Usage error (missing arguments or option values, bad flags, file
         exists without --force, baseline missing or without Benchmark lines);
         with --snapshot or --verdict also a pattern that is not exactly one
         package, a missing snapshot, a snapshot of another package or
         toolchain, --count below 6, or a helper that does not build

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
    --no-benchmem        Disable memory allocation stats (--verdict always
                         measures them)
    --force              Allow --save or --snapshot to overwrite existing files
    --limit N            Max benchmark result lines to include (default: 0 = all)
    --benchtime D        Passed to go test as -benchtime (e.g. 2s, 500x)
    --snapshot DIR       Freeze the package's test binary and test files in DIR
    --verdict DIR        Compare the current code against DIR's snapshot and
                         decide keep (exit 0), revert or noisy (exit 1)
    --threshold PCT      Smallest sec/op, B/op or allocs/op change that counts
                         (default: 5)
    --new-env KEY=VALUE  Environment for the changed side's runs only, e.g.
                         GOGC=400 against an unchanged snapshot (repeatable)

ARGUMENTS
    package              Go package to benchmark (default: ./..., or . with
                         --snapshot and --verdict, which take one package)

EXAMPLES
    bash $SCRIPT_NAME
    bash $SCRIPT_NAME -n 10 ./pkg/parser
    bash $SCRIPT_NAME --save baseline.txt ./...
    bash $SCRIPT_NAME --baseline baseline.txt --save current.txt ./...
    bash $SCRIPT_NAME --filter BenchmarkSort -n 3
    bash $SCRIPT_NAME --json --limit 5 ./...
    bash $SCRIPT_NAME --save results.txt --force ./...
    bash $SCRIPT_NAME --snapshot "\${TMPDIR:-/tmp}/snap" ./internal/codec
    bash $SCRIPT_NAME --verdict "\${TMPDIR:-/tmp}/snap" --filter Encode ./internal/codec
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
BENCHTIME=""
SNAPSHOT=""
VERDICT=""
THRESHOLD=5
NEW_ENV=()

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
        --benchtime)    need_arg "$1" "${2-}"; BENCHTIME="$2"; shift 2 ;;
        --snapshot)     need_arg "$1" "${2-}"; SNAPSHOT="$2"; shift 2 ;;
        --verdict)      need_arg "$1" "${2-}"; VERDICT="$2"; shift 2 ;;
        --threshold)    need_arg "$1" "${2-}"; THRESHOLD="$2"; shift 2 ;;
        --new-env)      need_arg "$1" "${2-}"; NEW_ENV+=("$2"); shift 2 ;;
        -*)             echo "error: unknown option: $1" >&2; usage >&2; exit 2 ;;
        *)              PACKAGE="$1"; shift ;;
    esac
done

if [[ -z $PACKAGE ]]; then
    if [[ -n "$SNAPSHOT" || -n "$VERDICT" ]]; then PACKAGE="."; else PACKAGE="./..."; fi
fi

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

# --- --snapshot and --verdict ----------------------------------------------

# resolve_package sets PKG_PATH and PKG_DIR for the one package PACKAGE names.
resolve_package() {
    local listed errfile
    errfile=$(mktemp "${TMPDIR:-/tmp}/bench-list.XXXXXX")
    if ! listed=$(go list -f '{{.ImportPath}}{{"\t"}}{{.Dir}}' "$PACKAGE" 2>"$errfile"); then
        echo "error: go list $PACKAGE failed:" >&2
        cat "$errfile" >&2
        rm -f "$errfile"
        exit 2
    fi
    rm -f "$errfile"
    if [[ -z $listed || $listed == *$'\n'* ]]; then
        echo "error: --snapshot and --verdict take exactly one package; $PACKAGE matches $(printf '%s' "$listed" | grep -c . || true)" >&2
        exit 2
    fi
    PKG_PATH=${listed%%$'\t'*}
    PKG_DIR=${listed#*$'\t'}
}

meta_value() {
    sed -n "s/^$1=//p" "$SNAP_DIR/meta"
}

# build_helper sets HELPER to bench-verdict.go built once into the user cache.
# It is built with the toolchain the target project selects when that one
# resolves (else the local one), outside the project so its go.mod, go.work,
# and GOFLAGS cannot break a standard-library-only build, and keyed by the
# source checksum and that toolchain.
build_helper() {
    local cache src want build_env goversion
    cache="${XDG_CACHE_HOME:-${HOME:-${TMPDIR:-/tmp}}/.cache}/golang-skills"
    if ! mkdir -p "$cache" 2>/dev/null; then
        cache="${TMPDIR:-/tmp}/golang-skills-cache"
        mkdir -p "$cache"
    fi
    cache="$(cd "$cache" && pwd)"
    src="$SCRIPT_DIR/bench-verdict.go"
    want="$(go env GOVERSION 2>/dev/null)" || want=""
    build_env=(env GOWORK=off GOFLAGS= "GOTOOLCHAIN=${want:-local}")
    if ! goversion="$(cd "$cache" && "${build_env[@]}" go env GOVERSION 2>/dev/null)"; then
        build_env=(env GOWORK=off GOFLAGS= GOTOOLCHAIN=local)
        if ! goversion="$(cd "$cache" && "${build_env[@]}" go env GOVERSION)"; then
            echo "error: go env GOVERSION failed" >&2
            exit 2
        fi
    fi
    HELPER="$cache/bench-verdict-$(cksum "$src" | awk '{print $1 "-" $2}')-$goversion"
    if [[ ! -x "$HELPER" ]]; then
        if ! (cd "$cache" && "${build_env[@]}" GOCACHE="${GOCACHE:-$cache/go-build}" go build -o "$HELPER" "$src"); then
            echo "error: could not build $src" >&2
            exit 2
        fi
    fi
}

# json_strings prints its arguments as a JSON array of strings.
json_strings() {
    local sep="" v
    printf '['
    for v in "$@"; do
        printf '%s"%s"' "$sep" "$(json_escape "$v")"
        sep=","
    done
    printf ']'
}

snapshot_mode() {
    SNAP_DIR=$SNAPSHOT
    resolve_package
    if [[ -e "$SNAP_DIR/base.test" ]] && ! $FORCE; then
        echo "error: snapshot already exists: $SNAP_DIR (use --force to replace it)" >&2
        exit 2
    fi
    mkdir -p "$SNAP_DIR"
    SNAP_DIR="$(cd "$SNAP_DIR" && pwd)"
    rm -rf "$SNAP_DIR/tests" "$SNAP_DIR/base.test" "$SNAP_DIR/new.test" "$SNAP_DIR/meta" \
        "$SNAP_DIR/base.txt" "$SNAP_DIR/new.txt" "$SNAP_DIR/overlay.json"
    mkdir "$SNAP_DIR/tests"
    local f build_out go_exit=0 status=ok final=0 benches=()
    for f in "$PKG_DIR"/*_test.go; do
        if [[ -e $f ]]; then cp "$f" "$SNAP_DIR/tests/"; fi
    done
    log "Building the snapshot: go test -c -o $SNAP_DIR/base.test $PACKAGE"
    build_out=$(go test -c -o "$SNAP_DIR/base.test" "$PACKAGE" 2>&1) || go_exit=$?
    if [[ $go_exit -ne 0 ]]; then
        status=error
    elif [[ ! -x "$SNAP_DIR/base.test" ]]; then
        status=no_benchmarks # go test -c writes no binary for a package without tests
    else
        while IFS= read -r f; do
            if [[ $f == Benchmark* ]]; then benches+=("$f"); fi
        done < <(cd "$PKG_DIR" && "$SNAP_DIR/base.test" -test.list "$FILTER" 2>/dev/null || true)
        if [[ ${#benches[@]} -eq 0 ]]; then status=no_benchmarks; fi
    fi
    if [[ $status == ok ]]; then
        printf 'import_path=%s\ngo_version=%s\n' "$PKG_PATH" "$(go env GOVERSION)" >"$SNAP_DIR/meta"
        log "Snapshot of $PKG_PATH in $SNAP_DIR: ${#benches[@]} benchmarks match $FILTER"
    else
        rm -f "$SNAP_DIR/base.test" # only a snapshot with benchmarks can be compared against
        final=1
        log "error: snapshot not taken: status $status"
        if [[ -n $build_out ]]; then log "$build_out"; fi
    fi
    if $JSON_OUTPUT; then
        printf '{"mode":"snapshot","package":"%s",' "$(json_escape "$PACKAGE")"
        printf '"import_path":"%s","snapshot":"%s",' "$(json_escape "$PKG_PATH")" "$(json_escape "$SNAP_DIR")"
        printf '"go_version":"%s","filter":"%s",' "$(json_escape "$(go env GOVERSION)")" "$(json_escape "$FILTER")"
        printf '"benchmarks":%s,' "$(json_strings ${benches[@]+"${benches[@]}"})"
        printf '"status":"%s","exit_code":%d,"go_exit_code":%d,' "$status" "$final" "$go_exit"
        printf '"output":"%s"}\n' "$(json_escape "$build_out")"
    fi
    exit $final
}

# run_side appends one run of the base or new binary to its result file.
run_side() {
    local env_args=()
    if [[ $1 == new ]]; then env_args=(${NEW_ENV[@]+"${NEW_ENV[@]}"}); fi
    (cd "$PKG_DIR" && env ${env_args[@]+"${env_args[@]}"} "$SNAP_DIR/$1.test" "${RUN_ARGS[@]}") >>"$SNAP_DIR/$1.txt" 2>&1
}

verdict_mode() {
    SNAP_DIR=$VERDICT
    if [[ ! -x "$SNAP_DIR/base.test" || ! -f "$SNAP_DIR/meta" ]]; then
        echo "error: no snapshot in $SNAP_DIR; run --snapshot $SNAP_DIR first" >&2
        exit 2
    fi
    SNAP_DIR="$(cd "$SNAP_DIR" && pwd)"
    if [[ $COUNT -lt 6 ]]; then
        echo "error: --verdict needs --count 6 or more for a 95% interval, got: $COUNT" >&2
        exit 2
    fi
    resolve_package
    local snap_path snap_go cur_go
    snap_path=$(meta_value import_path)
    if [[ $snap_path != "$PKG_PATH" ]]; then
        echo "error: the snapshot in $SNAP_DIR is of $snap_path, not $PKG_PATH" >&2
        exit 2
    fi
    snap_go=$(meta_value go_version)
    cur_go=$(go env GOVERSION)
    if [[ $snap_go != "$cur_go" ]]; then
        echo "error: the snapshot was built by $snap_go and the toolchain is now $cur_go; a comparison across toolchains measures the toolchain, so re-snapshot first" >&2
        exit 2
    fi
    build_helper

    # The new binary is built with the snapshot's test files in place of the
    # package's current ones, and without any test file added since.
    local f sep=""
    {
        printf '{"Replace":{'
        for f in "$SNAP_DIR"/tests/*_test.go; do
            if [[ ! -e $f ]]; then continue; fi
            printf '%s"%s":"%s"' "$sep" "$(json_escape "$PKG_DIR/$(basename "$f")")" "$(json_escape "$f")"
            sep=","
        done
        for f in "$PKG_DIR"/*_test.go; do
            if [[ ! -e $f || -e "$SNAP_DIR/tests/$(basename "$f")" ]]; then continue; fi
            printf '%s"%s":""' "$sep" "$(json_escape "$f")"
            sep=","
        done
        printf '}}\n'
    } >"$SNAP_DIR/overlay.json"

    local out="" go_exit=0 status=ok final=0 comparison="" helper_exit=0 result="" i side order
    rm -f "$SNAP_DIR/new.test"
    log "Building the current code against the snapshot's test files"
    out=$(go test -c -overlay "$SNAP_DIR/overlay.json" -o "$SNAP_DIR/new.test" "$PACKAGE" 2>&1) || go_exit=$?
    if [[ $go_exit -ne 0 ]]; then
        status=error
        out="the current code does not build with the snapshot's test files (did an API the benchmarks use change?):"$'\n'"$out"
    fi

    RUN_ARGS=(-test.run '^$' -test.bench "$FILTER" -test.count 1 -test.benchmem)
    if [[ -n $BENCHTIME ]]; then RUN_ARGS+=(-test.benchtime "$BENCHTIME"); fi
    : >"$SNAP_DIR/base.txt"
    : >"$SNAP_DIR/new.txt"
    if [[ $status == ok ]]; then
        log "Running base and new $COUNT times each, alternating which goes first, in $PKG_DIR"
        for ((i = 1; i <= COUNT; i++)); do
            if ((i % 2)); then order=(base new); else order=(new base); fi
            for side in "${order[@]}"; do
                if ! run_side "$side"; then
                    status=error
                    out="the $side binary failed:"$'\n'"$(tail -n 20 "$SNAP_DIR/$side.txt")"
                    break 2
                fi
            done
        done
    fi
    if [[ $status == ok ]] && { ! grep -qE '^Benchmark' "$SNAP_DIR/base.txt" || ! grep -qE '^Benchmark' "$SNAP_DIR/new.txt"; }; then
        status=no_benchmarks
    fi

    if [[ $status == ok ]]; then
        local helper_args=(--threshold "$THRESHOLD")
        if $JSON_OUTPUT; then helper_args+=(--json); fi
        comparison=$("$HELPER" "${helper_args[@]}" "$SNAP_DIR/base.txt" "$SNAP_DIR/new.txt" 2>&1) || helper_exit=$?
        case $helper_exit in
            0 | 1)
                final=$helper_exit
                if $JSON_OUTPUT; then
                    result=$(printf '%s' "$comparison" | grep -oE '"verdict":"[a-z]+"' | head -n 1 | cut -d'"' -f4)
                else
                    result=$(printf '%s\n' "$comparison" | sed -n 's/^verdict: \([a-z]*\).*/\1/p')
                fi
                ;;
            *)
                status=error
                out=$comparison
                comparison=""
                ;;
        esac
    fi
    if [[ $status != ok ]]; then
        final=1
        if [[ $status == error && $helper_exit -gt 1 ]]; then final=2; fi
    fi

    if $JSON_OUTPUT; then
        if [[ -n $out ]]; then log "$out"; fi
        printf '{"mode":"verdict","package":"%s",' "$(json_escape "$PACKAGE")"
        printf '"import_path":"%s","snapshot":"%s",' "$(json_escape "$PKG_PATH")" "$(json_escape "$SNAP_DIR")"
        printf '"go_version":"%s","count":%d,"filter":"%s",' "$(json_escape "$cur_go")" "$COUNT" "$(json_escape "$FILTER")"
        printf '"threshold_pct":%s,"benchtime":"%s",' "$THRESHOLD" "$(json_escape "$BENCHTIME")"
        printf '"new_env":%s,' "$(json_strings ${NEW_ENV[@]+"${NEW_ENV[@]}"})"
        printf '"status":"%s","verdict":"%s","exit_code":%d,"go_exit_code":%d,' "$status" "$result" "$final" "$go_exit"
        printf '"comparison":%s,' "${comparison:-null}"
        printf '"output":"%s"}\n' "$(json_escape "$out")"
    else
        if [[ -n $out ]]; then log "$out"; fi
        if [[ -n $comparison ]]; then
            log ""
            log "$comparison"
            log "Raw runs: $SNAP_DIR/base.txt, $SNAP_DIR/new.txt"
        elif [[ $status == no_benchmarks ]]; then
            log "error: no benchmarks found matching filter: $FILTER"
        fi
    fi
    exit $final
}

if [[ -n "$SNAPSHOT" || -n "$VERDICT" ]]; then
    if [[ -n "$SNAPSHOT" && -n "$VERDICT" ]] || [[ -n "$SAVE" || -n "$BASELINE" ]]; then
        echo "error: --snapshot, --verdict, and --save/--baseline are separate modes; give one" >&2
        exit 2
    fi
    if ! [[ "$THRESHOLD" =~ ^[0-9]+(\.[0-9]+)?$ ]] || [[ "$THRESHOLD" =~ ^0+(\.0+)?$ ]]; then
        echo "error: --threshold must be a positive number, got: $THRESHOLD" >&2
        exit 2
    fi
    for kv in ${NEW_ENV[@]+"${NEW_ENV[@]}"}; do
        if ! [[ $kv =~ ^[A-Za-z_][A-Za-z0-9_]*= ]]; then
            echo "error: --new-env takes KEY=VALUE, got: $kv" >&2
            exit 2
        fi
    done
    if [[ -n "$SNAPSHOT" && ${#NEW_ENV[@]} -gt 0 ]]; then
        echo "error: --new-env applies to --verdict runs only" >&2
        exit 2
    fi
    if [[ -n "$SNAPSHOT" ]]; then snapshot_mode; else verdict_mode; fi
fi

if [[ ${#NEW_ENV[@]} -gt 0 ]]; then
    echo "error: --new-env applies to --verdict runs only" >&2
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
if [[ -n $BENCHTIME ]]; then
    BENCH_ARGS+=(-benchtime "$BENCHTIME")
fi
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
