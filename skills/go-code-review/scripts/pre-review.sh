#!/usr/bin/env bash
set -euo pipefail

VERSION="1.0.0"
SCRIPT_NAME="$(basename "$0")"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

usage() {
    cat <<EOF
$SCRIPT_NAME v$VERSION — Run automated pre-review checks on Go code

USAGE
    bash $SCRIPT_NAME [options] [path]

DESCRIPTION
    Runs gofmt, go vet, and golangci-lint against the target path and
    reports any findings. Use before manual code review to catch
    mechanical issues early.

    golangci-lint runs with the project's configuration; without one it
    runs with the go-linting baseline (../../go-linting/assets/golangci.yml
    beside this skill) when that file is installed, and otherwise with
    golangci-lint's own defaults, which leave revive, godot, gosec, and
    modernize off. The output names which of the three ran (config:
    project, baseline, or defaults).

    With --new-from-rev REV, golangci-lint reports only issues new since the
    git revision REV, and gofmt checks only the .go files under the path that
    changed since REV (untracked files included); go vet still runs on the
    whole path.

    A missing golangci-lint, or one that exits with an error rather than
    findings (exit code other than 0 or 1), is reported as unavailable (the
    run is then INCOMPLETE, not clean); gofmt and go vet still run. Use
    --strict where the linter is guaranteed (CI) to make that an error.

    Exits 0 if no check failed, 1 if issues found, 2 on error.

OPTIONS
    -h, --help       Show this help message
    -v, --version    Show version
    --json           Output results as JSON
    --strict         Fail if golangci-lint is not installed or cannot run
    --force          Accepted and ignored (a missing linter is reported as unavailable by default)
    --limit N        Max items reported per section (0 = unlimited, default: 0)
    --new-from-rev REV
                     Report only lint issues and gofmt files new since REV

ARGUMENTS
    path             Package pattern to check (default: ./...)

EXAMPLES
    bash $SCRIPT_NAME
    bash $SCRIPT_NAME ./pkg/...
    bash $SCRIPT_NAME --json ./cmd/server/...
    bash $SCRIPT_NAME --strict ./...
    bash $SCRIPT_NAME --json --limit 10 ./...
    bash $SCRIPT_NAME --new-from-rev origin/main ./internal/cache/...
EOF
}

json_escape() {
    local s="$1" c r i
    s="${s//\\/\\\\}"
    s="${s//\"/\\\"}"
    s="${s//$'\t'/\\t}"
    s="${s//$'\r'/}"
    s="${s//$'\n'/\\n}"
    # JSON forbids the rest of U+0001-U+001F raw (a bash string holds no NUL);
    # colored tool output carries ESC, for one.
    if [[ "$s" == *[[:cntrl:]]* ]]; then
        for i in 1 2 3 4 5 6 7 8 11 12 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31; do
            printf -v c "\\$(printf '%03o' "$i")"
            printf -v r '\\u%04x' "$i"
            s="${s//"$c"/"$r"}"
        done
    fi
    printf '%s' "$s"
}

JSON_OUTPUT=false
STRICT=false
LIMIT=0
TARGET=""
NEW_FROM_REV=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)    usage; exit 0 ;;
        -v|--version) echo "$SCRIPT_NAME v$VERSION"; exit 0 ;;
        --json)       JSON_OUTPUT=true; shift ;;
        --strict)     STRICT=true; shift ;;
        --force)      shift ;;  # kept for compatibility: unavailable is the default
        --limit)
            if [[ $# -lt 2 ]]; then
                echo "error: --limit requires a number" >&2
                exit 2
            fi
            LIMIT="$2"
            shift 2
            ;;
        --new-from-rev)
            if [[ $# -lt 2 || -z "$2" ]]; then
                echo "error: --new-from-rev requires a git revision" >&2
                exit 2
            fi
            NEW_FROM_REV="$2"
            shift 2
            ;;
        --new-from-rev=*)
            NEW_FROM_REV="${1#--new-from-rev=}"
            if [[ -z "$NEW_FROM_REV" ]]; then
                echo "error: --new-from-rev requires a git revision" >&2
                exit 2
            fi
            shift
            ;;
        -*)           echo "error: unknown option: $1" >&2; usage >&2; exit 2 ;;
        *)            TARGET="$1"; shift ;;
    esac
done

TARGET="${TARGET:-./...}"

if ! [[ "$LIMIT" =~ ^[0-9]+$ ]]; then
    echo "error: --limit must be a non-negative integer, got: $LIMIT" >&2
    exit 2
fi

if ! command -v go &>/dev/null; then
    echo "error: go is not installed or not in PATH" >&2
    exit 2
fi

if ! command -v gofmt &>/dev/null; then
    echo "error: gofmt is not installed or not in PATH" >&2
    exit 2
fi

if [[ -n "$NEW_FROM_REV" ]]; then
    if ! command -v git &>/dev/null || ! git rev-parse --verify --quiet "${NEW_FROM_REV}^{commit}" >/dev/null 2>&1; then
        echo "error: --new-from-rev: not a git revision in this checkout: $NEW_FROM_REV" >&2
        exit 2
    fi
fi

GOFMT_STATUS="pass"
GOFMT_FINDINGS=()
GOFMT_DIR="${TARGET%%/...}"
GOFMT_DIR="${GOFMT_DIR:-.}"
if [[ -n "$NEW_FROM_REV" ]]; then
    # Only the .go files under the path that changed since REV, untracked
    # ones included; a deleted file has nothing left to format.
    CHANGED=()
    while IFS= read -r f; do
        [[ "$f" == *.go && -f "$f" ]] && CHANGED+=("$f")
    done < <({
        git diff --name-only --relative --diff-filter=d "$NEW_FROM_REV" -- "$GOFMT_DIR"
        git ls-files --others --exclude-standard -- "$GOFMT_DIR"
    } 2>/dev/null | sort -u)
    UNFORMATTED=""
    if [[ ${#CHANGED[@]} -gt 0 ]]; then
        UNFORMATTED=$(gofmt -l "${CHANGED[@]}" 2>&1 | grep -v -e '^vendor/' -e '/vendor/') || true
    fi
else
    UNFORMATTED=$(gofmt -l "$GOFMT_DIR" 2>&1 | grep -v -e '^vendor/' -e '/vendor/') || true
fi
if [[ -n "$UNFORMATTED" ]]; then
    GOFMT_STATUS="fail"
    while IFS= read -r f; do
        [[ -n "$f" ]] && GOFMT_FINDINGS+=("$f")
    done <<< "$UNFORMATTED"
fi

GOVET_STATUS="pass"
GOVET_OUTPUT=""
if ! GOVET_OUTPUT=$(go vet "$TARGET" 2>&1); then
    GOVET_STATUS="fail"
fi

LINT_STATUS="unavailable"
LINT_REASON="not installed"
LINT_OUTPUT=""
LINT_BASELINE=false
LINT_CONFIG=""
if command -v golangci-lint &>/dev/null; then
    LINT_ARGS=(run)
    # golangci-lint looks for a config in the working directory and from the
    # path argument upward; without one, the go-linting baseline enables the
    # linters the review checklist leaves to tools (revive, godot, gosec,
    # modernize).
    HAS_CONFIG=false
    for dir in . "$GOFMT_DIR"; do
        if [[ -d "$dir" ]] && (cd "$dir" && golangci-lint config path) >/dev/null 2>&1; then
            HAS_CONFIG=true
        fi
    done
    BASELINE="$SCRIPT_DIR/../../go-linting/assets/golangci.yml"
    if $HAS_CONFIG; then
        LINT_CONFIG="project"
    elif [[ -f "$BASELINE" ]]; then
        # Paths would otherwise be relative to the baseline's directory.
        LINT_ARGS+=(--config "$BASELINE" --path-mode=abs)
        LINT_BASELINE=true
        LINT_CONFIG="baseline"
    else
        # A single-skill install has no go-linting beside it.
        LINT_CONFIG="defaults"
    fi
    if [[ -n "$NEW_FROM_REV" ]]; then
        LINT_ARGS+=(--new-from-rev="$NEW_FROM_REV")
    fi
    # Exit 1 is findings; any other non-zero exit (no Go files, a go
    # directive newer than the linter, a broken config) is an environment
    # error, not a finding.
    LINT_EXIT=0
    LINT_OUTPUT=$(golangci-lint "${LINT_ARGS[@]}" "$TARGET" 2>&1) || LINT_EXIT=$?
    case "$LINT_EXIT" in
        0) LINT_STATUS="pass" ;;
        1) LINT_STATUS="fail" ;;
        *) LINT_REASON="golangci-lint exited $LINT_EXIT" ;;
    esac
fi
if [[ "$LINT_STATUS" == "unavailable" ]] && $STRICT; then
    echo "error: golangci-lint unavailable: $LINT_REASON (--strict)" >&2
    [[ -z "$LINT_OUTPUT" ]] || echo "$LINT_OUTPUT" >&2
    exit 2
fi

FAILED=0
[[ "$GOFMT_STATUS" == "fail" ]] && FAILED=1
[[ "$GOVET_STATUS" == "fail" ]] && FAILED=1
[[ "$LINT_STATUS" == "fail" ]] && FAILED=1

if $JSON_OUTPUT; then
    GOFMT_TRUNCATED=false
    GOFMT_DISPLAY=("${GOFMT_FINDINGS[@]+"${GOFMT_FINDINGS[@]}"}")
    if [[ $LIMIT -gt 0 && ${#GOFMT_DISPLAY[@]} -gt $LIMIT ]]; then
        GOFMT_DISPLAY=("${GOFMT_FINDINGS[@]:0:$LIMIT}")
        GOFMT_TRUNCATED=true
    fi

    GOFMT_JSON="["
    first=true
    for f in "${GOFMT_DISPLAY[@]+"${GOFMT_DISPLAY[@]}"}"; do
        $first || GOFMT_JSON+=","
        first=false
        GOFMT_JSON+="\"$(json_escape "$f")\""
    done
    GOFMT_JSON+="]"

    GOVET_TRUNCATED=false
    GOVET_DISPLAY="$GOVET_OUTPUT"
    if [[ $LIMIT -gt 0 && -n "$GOVET_OUTPUT" ]]; then
        GOVET_ARR=()
        while IFS= read -r line; do
            GOVET_ARR+=("$line")
        done <<< "$GOVET_OUTPUT"
        if [[ ${#GOVET_ARR[@]} -gt $LIMIT ]]; then
            GOVET_DISPLAY=""
            for (( i=0; i<LIMIT; i++ )); do
                [[ -n "$GOVET_DISPLAY" ]] && GOVET_DISPLAY+=$'\n'
                GOVET_DISPLAY+="${GOVET_ARR[$i]}"
            done
            GOVET_TRUNCATED=true
        fi
    fi
    GOVET_ESC="$(json_escape "$GOVET_DISPLAY")"

    LINT_TRUNCATED=false
    LINT_DISPLAY="$LINT_OUTPUT"
    if [[ $LIMIT -gt 0 && -n "$LINT_OUTPUT" ]]; then
        LINT_ARR=()
        while IFS= read -r line; do
            LINT_ARR+=("$line")
        done <<< "$LINT_OUTPUT"
        if [[ ${#LINT_ARR[@]} -gt $LIMIT ]]; then
            LINT_DISPLAY=""
            for (( i=0; i<LIMIT; i++ )); do
                [[ -n "$LINT_DISPLAY" ]] && LINT_DISPLAY+=$'\n'
                LINT_DISPLAY+="${LINT_ARR[$i]}"
            done
            LINT_TRUNCATED=true
        fi
    fi
    LINT_ESC="$(json_escape "$LINT_DISPLAY")"

    GOFMT_TRUNC=""
    $GOFMT_TRUNCATED && GOFMT_TRUNC=',"truncated":true'
    GOVET_TRUNC=""
    $GOVET_TRUNCATED && GOVET_TRUNC=',"truncated":true'
    LINT_TRUNC=""
    $LINT_TRUNCATED && LINT_TRUNC=',"truncated":true'

    cat <<EOF
{"gofmt":{"status":"$GOFMT_STATUS","files":$GOFMT_JSON$GOFMT_TRUNC},"govet":{"status":"$GOVET_STATUS","output":"$GOVET_ESC"$GOVET_TRUNC},"golangci_lint":{"status":"$LINT_STATUS","config":"$LINT_CONFIG","output":"$LINT_ESC"$LINT_TRUNC},"passed":$( [[ $FAILED -eq 0 ]] && echo true || echo false )}
EOF
else
    echo "=== gofmt ==="
    if [[ "$GOFMT_STATUS" == "fail" ]]; then
        echo "Unformatted files:"
        GOFMT_COUNT=0
        for f in "${GOFMT_FINDINGS[@]}"; do
            GOFMT_COUNT=$((GOFMT_COUNT + 1))
            if [[ $LIMIT -gt 0 && $GOFMT_COUNT -gt $LIMIT ]]; then
                echo "  ... ($(( ${#GOFMT_FINDINGS[@]} - LIMIT )) more items truncated)"
                break
            fi
            echo "  $f"
        done
    else
        echo "OK"
    fi

    echo ""
    echo "=== go vet ==="
    if [[ "$GOVET_STATUS" == "fail" ]]; then
        if [[ $LIMIT -gt 0 ]]; then
            GOVET_ARR=()
            while IFS= read -r line; do
                GOVET_ARR+=("$line")
            done <<< "$GOVET_OUTPUT"
            for (( i=0; i<${#GOVET_ARR[@]} && i<LIMIT; i++ )); do
                echo "${GOVET_ARR[$i]}"
            done
            if [[ ${#GOVET_ARR[@]} -gt $LIMIT ]]; then
                echo "... ($(( ${#GOVET_ARR[@]} - LIMIT )) more items truncated)"
            fi
        else
            echo "$GOVET_OUTPUT"
        fi
    else
        echo "OK"
    fi

    echo ""
    echo "=== golangci-lint ==="
    $LINT_BASELINE && echo "No project config: linted with the go-linting baseline"
    [[ "$LINT_CONFIG" == "defaults" ]] && echo "No project config and no go-linting baseline: linted with golangci-lint defaults; revive, godot, gosec, and modernize did not run"
    [[ -n "$NEW_FROM_REV" ]] && echo "Only issues new since $NEW_FROM_REV"
    [[ "$LINT_STATUS" == "unavailable" ]] && echo "Unavailable ($LINT_REASON)"
    if [[ "$LINT_STATUS" == "pass" ]]; then
        echo "OK"
    elif [[ -n "$LINT_OUTPUT" ]]; then
        if [[ $LIMIT -gt 0 ]]; then
            LINT_ARR=()
            while IFS= read -r line; do
                LINT_ARR+=("$line")
            done <<< "$LINT_OUTPUT"
            for (( i=0; i<${#LINT_ARR[@]} && i<LIMIT; i++ )); do
                echo "${LINT_ARR[$i]}"
            done
            if [[ ${#LINT_ARR[@]} -gt $LIMIT ]]; then
                echo "... ($(( ${#LINT_ARR[@]} - LIMIT )) more items truncated)"
            fi
        else
            echo "$LINT_OUTPUT"
        fi
    fi

    echo ""
    if [[ $FAILED -eq 1 ]]; then
        echo "Pre-review checks FAILED — report the findings above before the manual review."
    elif [[ "$LINT_STATUS" == "unavailable" ]]; then
        echo "Pre-review checks INCOMPLETE — golangci-lint unavailable; gofmt and go vet passed."
    else
        echo "All pre-review checks passed."
    fi
fi

exit $FAILED
