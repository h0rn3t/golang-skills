#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERSION="1.1.0"

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
    --limit N          Show at most N violations (0 = all, default: all)
    --include-tests    Also check TestImports and XTestImports
    --config FILE      architecture.json to use (default: <module-root>/architecture.json)

RULES
    ownership, layer, composition, platform, contract, driver, unclassified,
    and stale/duplicate/unexplained entries in the known list.
    references/ARCHITECTURE-CHECKS.md documents each one and the config file.

EXIT CODES
    0 clean; 1 violations or a bad known entry; 2 error (no go.mod in
    module-root, checked before the config; package load failure; missing
    or invalid config, where a missing one prints a minimal starter file;
    bad flag)
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
if ! { mkdir -p "$CACHE_ROOT" && [[ -w "$CACHE_ROOT" ]]; } 2>/dev/null; then
    CACHE_ROOT="${TMPDIR:-/tmp}/golang-skills-cache"
    mkdir -p "$CACHE_ROOT"
fi
CACHE_ROOT="$(cd "$CACHE_ROOT" && pwd)"

SRC="$SCRIPT_DIR/check-architecture.go"
# The helper parses with the toolchain that builds it. It is built with the
# toolchain the target project selects when that one resolves (else the local
# one), outside the project so its go.mod, go.work, and GOFLAGS cannot break a
# standard-library-only build, and for the host platform whatever GOOS/GOARCH
# the caller exports. The cache key names that toolchain and platform, so a Go
# upgrade rebuilds it. A helper that does not build or run is an environment
# error (2).
WANT_GO="$(go env GOVERSION 2>/dev/null)" || WANT_GO=""
HOSTOS="" HOSTARCH=""
if HOSTENV="$(cd "$CACHE_ROOT" && GOWORK=off GOFLAGS= GOTOOLCHAIN=local go env GOHOSTOS GOHOSTARCH)"; then
    { read -r HOSTOS; read -r HOSTARCH; } <<< "$HOSTENV" || true
fi
if [[ -z "$HOSTOS" || -z "$HOSTARCH" ]]; then
    echo "error: go env GOHOSTOS GOHOSTARCH failed" >&2
    exit 2
fi
BUILD_ENV=(env GOWORK=off GOFLAGS= "GOOS=$HOSTOS" "GOARCH=$HOSTARCH" "GOTOOLCHAIN=${WANT_GO:-local}")
if ! GOVERSION="$(cd "$CACHE_ROOT" && "${BUILD_ENV[@]}" go env GOVERSION 2>/dev/null)"; then
    BUILD_ENV=(env GOWORK=off GOFLAGS= "GOOS=$HOSTOS" "GOARCH=$HOSTARCH" GOTOOLCHAIN=local)
    if ! GOVERSION="$(cd "$CACHE_ROOT" && "${BUILD_ENV[@]}" go env GOVERSION)"; then
        echo "error: go env GOVERSION failed" >&2
        exit 2
    fi
fi
STAMP="$(cksum "$SRC" | awk '{print $1 "-" $2}')-$GOVERSION-${HOSTOS}_$HOSTARCH"
BIN="$CACHE_ROOT/check-architecture-$STAMP"

if [[ ! -x "$BIN" ]]; then
    if ! (cd "$CACHE_ROOT" && "${BUILD_ENV[@]}" GOCACHE="${GOCACHE:-$CACHE_ROOT/go-build}" go build -o "$BIN" "$SRC"); then
        echo "error: could not build $SRC" >&2
        exit 2
    fi
fi

rc=0
"$BIN" "$@" || rc=$?
if ((rc == 126 || rc == 127)); then
    echo "error: cannot run $BIN" >&2
    exit 2
fi
exit "$rc"
