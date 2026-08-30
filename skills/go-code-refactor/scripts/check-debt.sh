#!/usr/bin/env bash
set -euo pipefail

VERSION="1.0.0"
SCRIPT_NAME="$(basename "$0")"

usage() {
    cat <<EOF
$SCRIPT_NAME v$VERSION — Harvest deliberate-shortcut markers into a ledger

USAGE
    bash $SCRIPT_NAME [options] [path]

DESCRIPTION
    Collects every "Kept:" comment left behind by a refactor, so a deliberate
    shortcut cannot quietly become permanent. The convention is:

      // Kept: <what was left alone and why>.
      // Ceiling: <the limit this accepts>. Fix: <what to do instead, later>.

    A marker that names no Ceiling and no Fix is tagged no-trigger — those are
    the ones that rot. Test files are scanned too; vendor/ is not.

    Exits 0 when every marker names an upgrade path, 1 when at least one is
    no-trigger, 2 on usage or environment error.

OPTIONS
    -h, --help       Show this help message
    -v, --version    Show version
    --json           Output results as JSON
    --limit N        Show at most N markers (default: all)

ARGUMENTS
    path             Directory or Go file to scan (default: current directory)

EXAMPLES
    bash $SCRIPT_NAME
    bash $SCRIPT_NAME ./internal
    bash $SCRIPT_NAME --json ./...
    bash $SCRIPT_NAME --limit 20 .
EOF
}

JSON_OUTPUT=false
LIMIT=0
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
            LIMIT="$2"
            shift 2
            ;;
        -*)           echo "error: unknown option: $1" >&2; usage >&2; exit 2 ;;
        *)            TARGET="$1"; shift ;;
    esac
done

TARGET="${TARGET:-.}"

if ! [[ "$LIMIT" =~ ^[0-9]+$ ]]; then
    echo "error: --limit must be a non-negative integer, got: $LIMIT" >&2
    exit 2
fi

json_escape() {
    local s="$1"
    s="${s//\\/\\\\}"
    s="${s//\"/\\\"}"
    s="${s//$'\t'/\\t}"
    s="${s//$'\r'/}"
    s="${s//$'\n'/\\n}"
    printf '%s' "$s"
}

# Validate here, not inside find_go_files: an exit from the process
# substitution below would leave the caller running with an empty file list.
SEARCH_ROOT="$TARGET"
if [[ ! -e "$TARGET" ]]; then
    SEARCH_ROOT="${TARGET%%/...}"
    SEARCH_ROOT="${SEARCH_ROOT:-.}"
    if [[ ! -d "$SEARCH_ROOT" ]]; then
        echo "error: path not found: $TARGET" >&2
        exit 2
    fi
fi

# Resolve target to a list of .go files. Unlike the other checkers this keeps
# _test.go: a shortcut taken in a test is debt the same way.
find_go_files() {
    local t="$1"
    if [[ -f "$t" ]]; then
        echo "$t"
    else
        find "$t" -name '*.go' ! -path '*/vendor/*' ! -path '*/.git/*' 2>/dev/null
    fi
}

# One awk pass per file batch. A marker runs from the "Kept:" line through the
# comment lines directly under it, so Ceiling/Fix may sit on a later line.
AWK_HARVEST='
function emit(   note) {
    inm = 0
    note = mtext
    gsub(/[ \t]+/, " ", note)
    sub(/^ /, "", note)
    sub(/ $/, "", note)
    printf "%s\t%d\t%s\t%s\t%s\n", mfile, mline, \
        (mtext ~ /Ceiling:/) ? "true" : "false", \
        (mtext ~ /Fix:/) ? "true" : "false", note
}
FNR == 1 && inm { emit() }
{
    if (inm) {
        if ($0 ~ /^[ \t]*\/\//) {
            c = $0
            sub(/^[ \t]*\/\/[ \t]?/, "", c)
            mtext = mtext " " c
            next
        }
        emit()
    }
    if (match($0, /\/\/[ \t]*Kept:/)) {
        inm = 1
        mfile = FILENAME
        mline = FNR
        mtext = substr($0, RSTART + RLENGTH)
    }
}
END { if (inm) emit() }
'

FILES=()
while IFS= read -r f; do
    [[ -n "$f" ]] && FILES+=("$f")
done < <(find_go_files "$SEARCH_ROOT")

if [[ ${#FILES[@]} -eq 0 ]]; then
    if $JSON_OUTPUT; then
        echo '{"markers":[],"total":0,"no_trigger":0,"truncated":false,"status":"no_go_files"}'
    else
        echo "No Go files found in: $TARGET"
    fi
    exit 0
fi

MARKERS=()
while IFS= read -r record; do
    [[ -n "$record" ]] && MARKERS+=("$record")
done < <(awk "$AWK_HARVEST" "${FILES[@]}")

TOTAL=${#MARKERS[@]}
NO_TRIGGER=0
for m in ${MARKERS[@]+"${MARKERS[@]}"}; do
    IFS=$'\t' read -r _ _ ceiling upgrade _ <<< "$m"
    if [[ "$ceiling" == "false" && "$upgrade" == "false" ]]; then
        NO_TRIGGER=$((NO_TRIGGER + 1))
    fi
done

TRUNCATED=false
if [[ $LIMIT -gt 0 && $TOTAL -gt $LIMIT ]]; then
    MARKERS=("${MARKERS[@]:0:$LIMIT}")
    TRUNCATED=true
fi

if $JSON_OUTPUT; then
    echo "{"
    echo '  "markers": ['
    first=true
    for m in ${MARKERS[@]+"${MARKERS[@]}"}; do
        IFS=$'\t' read -r file line ceiling upgrade note <<< "$m"
        rule="tracked"
        if [[ "$ceiling" == "false" && "$upgrade" == "false" ]]; then
            rule="no-trigger"
        fi
        $first || echo ","
        first=false
        printf '    {"file":"%s","line":%s,"ceiling":%s,"upgrade":%s,"rule":"%s","note":"%s"}' \
            "$(json_escape "$file")" "$line" "$ceiling" "$upgrade" "$rule" "$(json_escape "$note")"
    done
    echo ""
    echo "  ],"
    printf '  "total": %d,\n' "$TOTAL"
    printf '  "no_trigger": %d,\n' "$NO_TRIGGER"
    printf '  "truncated": %s\n' "$TRUNCATED"
    echo "}"
else
    if [[ $TOTAL -eq 0 ]]; then
        echo "No Kept: markers found. Clean ledger."
        exit 0
    fi

    echo "Deliberate shortcuts:"
    echo ""
    for m in "${MARKERS[@]}"; do
        IFS=$'\t' read -r file line ceiling upgrade note <<< "$m"
        rule="tracked"
        if [[ "$ceiling" == "false" && "$upgrade" == "false" ]]; then
            rule="no-trigger"
        fi
        printf "  %s:%s  [%s] %s\n" "$file" "$line" "$rule" "$note"
    done
    if $TRUNCATED; then
        echo "  ... and $((TOTAL - LIMIT)) more (use --limit to adjust)"
    fi
    echo ""
    echo "Total: $TOTAL marker(s), $NO_TRIGGER with no upgrade path"
fi

if [[ $NO_TRIGGER -gt 0 ]]; then
    exit 1
fi
exit 0
