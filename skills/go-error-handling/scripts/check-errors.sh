#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERSION="1.6.0"

for arg in "$@"; do
    case "$arg" in
        -h|--help)
            cat <<EOF
check-errors.sh v$VERSION - Check Go code for common error handling anti-patterns

USAGE
    bash check-errors.sh [options] [path]

    Reports text matching on err.Error() (==, !=, switch, and strings.Contains,
    HasPrefix, HasSuffix, EqualFold, Index) and an error that is both logged
    and returned. Skips _test.go and generated files and, as go ./... does,
    vendor and testdata directories, names that begin with "." or "_",
    nested modules, and directories the go.mod ignore directive names.

    Exits 0 if nothing is found, 1 on findings, 2 on a usage error or when a
    file does not parse. The other files are still checked; --json then adds
    "status":"parse_error" and a "parse_errors" list.

OPTIONS
    -h, --help       Show this help message
    -v, --version    Show version
    --json           Output results as JSON
    --bare-return    Also flag bare 'return err' for review (off by default: the
                     skill allows a bare return when annotation adds nothing)
    --no-bare-return Accepted for compatibility; the check is already off
    --limit N        Show at most N results (default: all)
EOF
            exit 0
            ;;
        -v|--version)
            echo "check-errors.sh v$VERSION"
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

SRC="$SCRIPT_DIR/check-errors-ast.go"
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
BIN="$CACHE_ROOT/check-errors-ast-$STAMP"

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
