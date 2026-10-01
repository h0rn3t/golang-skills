#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERSION="2.2.0"

for arg in "$@"; do
    case "$arg" in
        -h|--help)
            cat <<EOF
check-naming.sh v$VERSION — Check Go code for common naming anti-patterns

USAGE
    bash check-naming.sh [options] [path]

DESCRIPTION
    Parses Go source files and reports naming violations from the Go style
    guides:
      - SCREAMING_SNAKE_CASE constants (should be MixedCaps)
      - Get-prefixed methods with no parameters (simple accessors; omit Get)
      - Packages named util, utils, helper, helpers, common, misc, shared,
        base, or lib
      - Receivers named "this" or "self"

    Skips _test.go files, even one named as the path; generated files (a
    "Code generated ... DO NOT EDIT." line), as the go-linting gate does; and,
    as go ./... does, vendor and testdata directories and directories or
    files whose names begin with "." or "_".

    Exits 0 if no violations found, 1 if violations found, 2 on a usage
    error, when the helper does not build, or when a file does not parse. The other files are still checked; --json
    then adds "status":"parse_error" and a "parse_errors" list.

OPTIONS
    -h, --help       Show this help message
    -v, --version    Show version
    --json           Output results as JSON
    --limit N        Show at most N results (default: all); --limit=N too

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
CACHE_ROOT="$(cd "$CACHE_ROOT" && pwd)"

SRC="$SCRIPT_DIR/check-naming-ast.go"
# The helper parses with the toolchain that builds it. It is built with the
# toolchain the target project selects when that one resolves (else the local
# one), outside the project so its go.mod, go.work, and GOFLAGS cannot break a
# standard-library-only build, and the cache key names that toolchain, so a Go
# upgrade rebuilds it. A helper that does not build is an environment error (2).
WANT_GO="$(go env GOVERSION 2>/dev/null)" || WANT_GO=""
BUILD_ENV=(env GOWORK=off GOFLAGS= "GOTOOLCHAIN=${WANT_GO:-local}")
if ! GOVERSION="$(cd "$CACHE_ROOT" && "${BUILD_ENV[@]}" go env GOVERSION 2>/dev/null)"; then
    BUILD_ENV=(env GOWORK=off GOFLAGS= GOTOOLCHAIN=local)
    if ! GOVERSION="$(cd "$CACHE_ROOT" && "${BUILD_ENV[@]}" go env GOVERSION)"; then
        echo "error: go env GOVERSION failed" >&2
        exit 2
    fi
fi
STAMP="$(cksum "$SRC" | awk '{print $1 "-" $2}')-$GOVERSION"
BIN="$CACHE_ROOT/check-naming-ast-$STAMP"

if [[ ! -x "$BIN" ]]; then
    if ! (cd "$CACHE_ROOT" && "${BUILD_ENV[@]}" GOCACHE="${GOCACHE:-$CACHE_ROOT/go-build}" go build -o "$BIN" "$SRC"); then
        echo "error: could not build $SRC" >&2
        exit 2
    fi
fi

exec "$BIN" "$@"
