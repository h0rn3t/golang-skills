#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERSION="1.0.0"

for arg in "$@"; do
    case "$arg" in
        -h|--help)
            cat <<HELPTEXT
check-architecture.sh v$VERSION - Check internal/ imports against the layered-module policy (references/ARCHITECTURE.md)

USAGE
    bash check-architecture.sh [options] [module-root]

OPTIONS
    -h, --help         Show this help message
    -v, --version      Show version
    --json             Output results as JSON
    --limit N          Show at most N violations (default: all)
    --include-tests    Also check TestImports and XTestImports
    --config FILE      architecture.json to use (default: <module-root>/architecture.json)

RULES
    ownership, layer, composition, platform, contract, driver, unclassified,
    and stale/duplicate/unexplained entries in the known list.
    references/ARCHITECTURE-CHECKS.md documents each one and the config file.

EXIT CODES
    0 clean; 1 violations or a bad known entry; 2 error (no go.mod, package
    load failure, missing or invalid config, bad flag)
HELPTEXT
            exit 0
            ;;
        -v|--version)
            echo "check-architecture.sh v$VERSION"
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

SRC="$SCRIPT_DIR/check-architecture.go"
STAMP="$(cksum "$SRC" | awk '{print $1 "-" $2}')"
BIN="$CACHE_ROOT/check-architecture-$STAMP"

if [[ ! -x "$BIN" ]]; then
    GOCACHE="${GOCACHE:-$CACHE_ROOT/go-build}" go build -o "$BIN" "$SRC"
fi

exec "$BIN" "$@"
