#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERSION="2.0.0"

for arg in "$@"; do
    case "$arg" in
        -h|--help)
            cat <<EOF
check-naming.sh v$VERSION — Check Go code for common naming anti-patterns

USAGE
    bash check-naming.sh [options] [path]

DESCRIPTION
    Parses Go source files (go/ast) and reports:
      - SCREAMING_SNAKE_CASE constants (should be MixedCaps)
      - Get-prefixed getter methods (should omit Get)
      - Packages named util/helper/common/misc
      - Receivers named "this" or "self"

    Exits 0 if no violations found, 1 if violations found, 2 on error.

OPTIONS
    -h, --help       Show this help message
    -v, --version    Show version
    --json           Output results as JSON
    --limit N        Show at most N results (default: all)

ARGUMENTS
    path             Directory, ./... pattern, or Go file (default: current directory)
EOF
            exit 0
            ;;
        -v|--version)
            echo "check-naming.sh v$VERSION"
            exit 0
            ;;
    esac
done

if ! command -v go >/dev/null 2>&1; then
    echo "error: go is not installed or not in PATH" >&2
    exit 2
fi

CACHE_ROOT="${XDG_CACHE_HOME:-${HOME:-${TMPDIR:-/tmp}}/.cache}/golang-skills"
if ! mkdir -p "$CACHE_ROOT"; then
    CACHE_ROOT="${TMPDIR:-/tmp}/golang-skills-cache"
    mkdir -p "$CACHE_ROOT"
fi

SRC="$SCRIPT_DIR/check-naming-ast.go"
STAMP="$(cksum "$SRC" | awk '{print $1 "-" $2}')"
BIN="$CACHE_ROOT/check-naming-ast-$STAMP"

if [[ ! -x "$BIN" ]]; then
    GOCACHE="${GOCACHE:-$CACHE_ROOT/go-build}" go build -o "$BIN" "$SRC"
fi

exec "$BIN" "$@"
