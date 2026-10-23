#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERSION="1.5.0"

for arg in "$@"; do
    case "$arg" in
        -h|--help)
            cat <<EOF
check-interface-compliance.sh v$VERSION - List exported interfaces implemented beside their declaration that nothing converts to or that an exported function returns

USAGE
    bash check-interface-compliance.sh [options] [path]

OPTIONS
    -h, --help       Show this help message
    -v, --version    Show version
    --json           Output results as JSON
    --include-test   Also scan _test.go files for interface definitions and implementations
    --limit N        Show at most N results (default: all)
EOF
            exit 0
            ;;
        -v|--version)
            echo "check-interface-compliance.sh v$VERSION"
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

SRC="$SCRIPT_DIR/check-interface-compliance.go"
# The helper parses with the toolchain that builds it. It is built with the
# toolchain the target project selects when that one resolves (else the local
# one), outside the project so its go.mod, go.work, and GOFLAGS cannot break a
# standard-library-only build, and for the host platform whatever GOOS/GOARCH
# the caller exports. The cache key names that toolchain and platform, so a Go
# upgrade rebuilds it. A helper that does not build or run is an environment
# error (2).
WANT_GO="$(go env GOVERSION 2>/dev/null)" || WANT_GO=""
{ read -r HOSTOS; read -r HOSTARCH; } < <(cd "$CACHE_ROOT" && GOWORK=off GOFLAGS= GOTOOLCHAIN=local go env GOHOSTOS GOHOSTARCH)
BUILD_ENV=(env GOWORK=off GOFLAGS= "GOOS=$HOSTOS" "GOARCH=$HOSTARCH" "GOTOOLCHAIN=${WANT_GO:-local}")
if ! GOVERSION="$(cd "$CACHE_ROOT" && "${BUILD_ENV[@]}" go env GOVERSION 2>/dev/null)"; then
    BUILD_ENV=(env GOWORK=off GOFLAGS= "GOOS=$HOSTOS" "GOARCH=$HOSTARCH" GOTOOLCHAIN=local)
    if ! GOVERSION="$(cd "$CACHE_ROOT" && "${BUILD_ENV[@]}" go env GOVERSION)"; then
        echo "error: go env GOVERSION failed" >&2
        exit 2
    fi
fi
STAMP="$(cksum "$SRC" | awk '{print $1 "-" $2}')-$GOVERSION-${HOSTOS}_$HOSTARCH"
BIN="$CACHE_ROOT/check-interface-compliance-$STAMP"

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
