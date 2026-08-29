#!/usr/bin/env bash
set -uo pipefail

VERSION="1.0.0"
SCRIPT_NAME="$(basename "$0")"

usage() {
    cat <<EOF
$SCRIPT_NAME v$VERSION — Behavior-preservation harness for Go refactors

USAGE
    bash $SCRIPT_NAME [options] <mode> [path]

DESCRIPTION
    Captures a verifiable record of build, vet, test, race, and lint results
    so a refactor can be proven to change nothing observable.

    Modes:
      baseline   Record the state before any edit
      after      Record the state after a refactor step
      diff       Compare after against baseline (empty diff = behavior held)
      leaks      Run tests with the goroutine-leak profile (Go 1.26+)

    Results are written under .refactor-verify/ in the working directory.

    Exits 0 if all checks pass (or the diff is empty), 1 if a check failed
    or the diff is non-empty, 2 on usage or environment error.

OPTIONS
    -h, --help       Show this help message
    -v, --version    Show version
    --json           Output results as JSON
    --limit N        Max lines reported per section (0 = unlimited, default: 0)
    --out DIR        Result directory (default: .refactor-verify)

ARGUMENTS
    mode             baseline | after | diff | leaks
    path             Package pattern (default: ./...)

EXAMPLES
    bash $SCRIPT_NAME baseline ./...
    bash $SCRIPT_NAME after ./internal/...
    bash $SCRIPT_NAME diff
    bash $SCRIPT_NAME --json after ./...
    bash $SCRIPT_NAME leaks ./...
EOF
}

json_escape() {
    local s="$1"
    s="${s//\\/\\\\}"
    s="${s//\"/\\\"}"
    s="${s//$'\t'/\\t}"
    s="${s//$'\r'/}"
    s="${s//$'\n'/\\n}"
    printf '%s' "$s"
}

# Trim a blob to at most LIMIT lines. Echoes the (possibly trimmed) text and
# sets TRUNCATED for the caller.
TRUNCATED=false
apply_limit() {
    local text="$1"
    TRUNCATED=false
    if [[ "$LIMIT" -le 0 || -z "$text" ]]; then
        printf '%s' "$text"
        return
    fi
    local total
    total=$(printf '%s\n' "$text" | wc -l | tr -d ' ')
    if [[ "$total" -le "$LIMIT" ]]; then
        printf '%s' "$text"
        return
    fi
    TRUNCATED=true
    printf '%s' "$(printf '%s\n' "$text" | head -n "$LIMIT")"
}

JSON_OUTPUT=false
LIMIT=0
OUT_DIR=".refactor-verify"
MODE=""
TARGET=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)    usage; exit 0 ;;
        -v|--version) echo "$SCRIPT_NAME v$VERSION"; exit 0 ;;
        --json)       JSON_OUTPUT=true; shift ;;
        --limit)
            if [[ $# -lt 2 ]]; then
                echo "error: --limit requires a number" >&2
                exit 2
            fi
            LIMIT="$2"; shift 2 ;;
        --out)
            if [[ $# -lt 2 ]]; then
                echo "error: --out requires a directory" >&2
                exit 2
            fi
            OUT_DIR="$2"; shift 2 ;;
        -*)           echo "error: unknown option: $1" >&2; usage >&2; exit 2 ;;
        *)
            if [[ -z "$MODE" ]]; then MODE="$1"; else TARGET="$1"; fi
            shift ;;
    esac
done

if ! [[ "$LIMIT" =~ ^[0-9]+$ ]]; then
    echo "error: --limit must be a non-negative integer, got: $LIMIT" >&2
    exit 2
fi

case "$MODE" in
    baseline|after|diff|leaks) ;;
    "") echo "error: missing mode" >&2; usage >&2; exit 2 ;;
    *)  echo "error: unknown mode: $MODE" >&2; usage >&2; exit 2 ;;
esac

TARGET="${TARGET:-./...}"
GO_TEST_TIMEOUT="${GO_TEST_TIMEOUT:-5m}"

if ! command -v go &>/dev/null; then
    echo "error: go is not installed or not in PATH" >&2
    exit 2
fi

mkdir -p "$OUT_DIR" || { echo "error: cannot create $OUT_DIR" >&2; exit 2; }

# ------------------------------------------------------------------ diff mode
if [[ "$MODE" == "diff" ]]; then
    if [[ ! -f "$OUT_DIR/baseline.summary" || ! -f "$OUT_DIR/after.summary" ]]; then
        echo "error: need both a baseline and an after run first" >&2
        exit 2
    fi
    DIFF_OUT="$(diff -u "$OUT_DIR/baseline.summary" "$OUT_DIR/after.summary" 2>&1)" && DIFF_RC=0 || DIFF_RC=1
    DISPLAY="$(apply_limit "$DIFF_OUT")"
    if $JSON_OUTPUT; then
        printf '{"mode":"diff","identical":%s,"diff":"%s","truncated":%s}\n' \
            "$( [[ $DIFF_RC -eq 0 ]] && echo true || echo false )" \
            "$(json_escape "$DISPLAY")" \
            "$($TRUNCATED && echo true || echo false)"
    elif [[ $DIFF_RC -eq 0 ]]; then
        echo "identical: same checks pass, same tests run"
    else
        echo "$DISPLAY"
        $TRUNCATED && echo "... (truncated at $LIMIT lines)"
        echo
        echo "Differences above. Any change in test counts or verdicts means"
        echo "behavior moved. Investigate before shipping the refactor."
    fi
    exit "$DIFF_RC"
fi

# ----------------------------------------------------------------- leaks mode
# The goroutineleak pprof profile is GA since Go 1.27; Go 1.26 gates it behind
# GOEXPERIMENT=goroutineleakprofile, and setting that on 1.27+ fails the build.
if [[ "$MODE" == "leaks" ]]; then
    GO_MINOR="$(go version 2>/dev/null | sed -E 's/.*go1\.([0-9]+).*/\1/')"
    if ! [[ "$GO_MINOR" =~ ^[0-9]+$ ]] || [[ "$GO_MINOR" -lt 26 ]]; then
        echo "error: leaks mode needs Go 1.26+ (got go1.${GO_MINOR:-?})" >&2
        exit 2
    fi
    LEAK_LOG="$OUT_DIR/leaks.log"
    if [[ "$GO_MINOR" -ge 27 ]]; then
        go test -count=1 -timeout "$GO_TEST_TIMEOUT" "$TARGET" >"$LEAK_LOG" 2>&1 && LEAK_RC=0 || LEAK_RC=1
    else
        GOEXPERIMENT=goroutineleakprofile go test -count=1 -timeout "$GO_TEST_TIMEOUT" "$TARGET" \
            >"$LEAK_LOG" 2>&1 && LEAK_RC=0 || LEAK_RC=1
    fi
    LEAK_OUT="$(apply_limit "$(cat "$LEAK_LOG")")"
    if $JSON_OUTPUT; then
        printf '{"mode":"leaks","go_minor":%s,"passed":%s,"output":"%s","truncated":%s}\n' \
            "$GO_MINOR" \
            "$( [[ $LEAK_RC -eq 0 ]] && echo true || echo false )" \
            "$(json_escape "$LEAK_OUT")" \
            "$($TRUNCATED && echo true || echo false)"
    else
        echo "=== goroutine leak run (go1.$GO_MINOR) ==="
        echo "$LEAK_OUT"
        $TRUNCATED && echo "... (truncated at $LIMIT lines)"
        if [[ "$GO_MINOR" -ge 27 ]]; then
            echo "Collect the profile in-process via runtime/pprof.Lookup(\"goroutineleak\")"
            echo "or the /debug/pprof/goroutineleak endpoint."
        fi
        echo "--- full log: $LEAK_LOG ---"
    fi
    exit "$LEAK_RC"
fi

# --------------------------------------------------------- baseline / after
SUMMARY="$OUT_DIR/$MODE.summary"
LOG="$OUT_DIR/$MODE.log"
: >"$SUMMARY"
: >"$LOG"
FAILED=0

record() {
    echo "$1: $2" >>"$SUMMARY"
}

GO_VERSION="$(go version 2>/dev/null | awk '{print $3}')"
GO_DIRECTIVE="$(awk '/^go [0-9]/ {print $2; exit}' go.mod 2>/dev/null)"

# gofmt: report unformatted files by name so the summary diff is meaningful.
GOFMT_DIR="${TARGET%%/...}"
GOFMT_DIR="${GOFMT_DIR:-.}"
GOFMT_STATUS="skip"
if command -v gofmt &>/dev/null; then
    UNFORMATTED="$(gofmt -l "$GOFMT_DIR" 2>/dev/null | grep -v '^vendor/' || true)"
    if [[ -z "$UNFORMATTED" ]]; then
        GOFMT_STATUS="pass"
    else
        GOFMT_STATUS="fail"
        FAILED=1
        echo "=== gofmt ===" >>"$LOG"
        echo "$UNFORMATTED" >>"$LOG"
    fi
fi
record gofmt "$GOFMT_STATUS"

run_step() {
    local name="$1"; shift
    echo "=== $name ===" >>"$LOG"
    if "$@" >>"$LOG" 2>&1; then
        record "$name" pass
        return 0
    fi
    record "$name" fail
    FAILED=1
    return 1
}

run_step build go build "$TARGET"
run_step vet go vet "$TARGET"

# Per-test verdicts go into the summary: if a test disappears, starts being
# skipped, or starts failing, the baseline/after diff catches it. Timings are
# stripped because they vary per run.
echo "=== go test ===" >>"$LOG"
TEST_RAW="$OUT_DIR/$MODE.test.raw"
if go test -count=1 -timeout "$GO_TEST_TIMEOUT" -v "$TARGET" >"$TEST_RAW" 2>&1; then
    record test pass
else
    record test fail
    FAILED=1
fi
cat "$TEST_RAW" >>"$LOG"
grep -E '^[[:space:]]*--- (PASS|FAIL|SKIP):' "$TEST_RAW" 2>/dev/null \
    | sed -E 's/ \([0-9.]+s\)//' | sort >>"$SUMMARY" || true

run_step race go test -count=1 -race -timeout "$GO_TEST_TIMEOUT" "$TARGET"

# Pending modernizations, informational: excluded from the strict summary
# because a refactor is expected to reduce them, not hold them equal.
FIX_PENDING="n/a"
if go fix -diff "$TARGET" >"$OUT_DIR/$MODE.fix.diff" 2>/dev/null; then
    FIX_PENDING="$(wc -l <"$OUT_DIR/$MODE.fix.diff" | tr -d ' ')"
fi

LINT_FINDINGS="n/a"
if command -v golangci-lint &>/dev/null; then
    golangci-lint run "$TARGET" >"$OUT_DIR/$MODE.lint.raw" 2>&1 || true
    cat "$OUT_DIR/$MODE.lint.raw" >>"$LOG"
    LINT_FINDINGS="$(grep -cE '\.go:[0-9]+:[0-9]+:' "$OUT_DIR/$MODE.lint.raw" 2>/dev/null || true)"
    LINT_FINDINGS="${LINT_FINDINGS:-0}"
fi

if $JSON_OUTPUT; then
    printf '{"mode":"%s","target":"%s","toolchain":"%s","go_directive":"%s","gofmt":"%s","summary_path":"%s","fix_pending_lines":"%s","lint_findings":"%s","passed":%s}\n' \
        "$MODE" "$(json_escape "$TARGET")" "$(json_escape "${GO_VERSION:-unknown}")" \
        "$(json_escape "${GO_DIRECTIVE:-none}")" "$GOFMT_STATUS" \
        "$(json_escape "$SUMMARY")" "$FIX_PENDING" "$LINT_FINDINGS" \
        "$( [[ $FAILED -eq 0 ]] && echo true || echo false )"
else
    echo "--- $MODE ($TARGET) ---"
    echo "toolchain: ${GO_VERSION:-unknown} / go.mod directive: ${GO_DIRECTIVE:-none}"
    apply_limit "$(cat "$SUMMARY")"
    echo
    $TRUNCATED && echo "... (truncated at $LIMIT lines)"
    echo "go fix pending: $FIX_PENDING diff lines (informational)"
    echo "lint findings: $LINT_FINDINGS (informational)"
    echo "--- full log: $LOG ---"
    if [[ "$MODE" == "baseline" && "$FAILED" -eq 1 ]]; then
        echo
        echo "Baseline is red. Stop and report it: later failures cannot be" >&2
        echo "attributed to the refactor if they were already failing." >&2
    fi
fi

exit "$FAILED"
