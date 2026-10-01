#!/usr/bin/env bash
set -euo pipefail

VERSION="1.1.0"
SCRIPT_NAME="$(basename "$0")"

usage() {
    cat <<EOF
$SCRIPT_NAME v$VERSION — Generate a table-driven test scaffold for a Go function

USAGE
    bash $SCRIPT_NAME [options] <FuncName> <package>

DESCRIPTION
    Outputs a table-driven test file for the given function and package.
    By default writes to stdout; use --output to write to a file.

    Exits 0 on success, 2 on a usage error (a missing --output value, an
    extra argument, an invalid name) or an existing --output file without
    --force.

OPTIONS
    -h, --help           Show this help message
    -v, --version        Show version
    --output FILE        Write to FILE instead of stdout; a FILE that starts
                         with "-" takes a "./" prefix
    --force              Allow --output to overwrite an existing file
    --parallel           Include t.Parallel() in generated test
    --json               Output structured JSON metadata to stdout; without
                         --output the scaffold itself goes to stderr

ARGUMENTS
    FuncName             Name of the function to test (must be exported/uppercase)
    package              Go package name for the test file

EXAMPLES
    bash $SCRIPT_NAME ParseConfig config
    bash $SCRIPT_NAME --parallel ParseConfig config
    bash $SCRIPT_NAME --output config/parse_config_test.go ParseConfig config
    bash $SCRIPT_NAME --force --output config/parse_config_test.go ParseConfig config
    bash $SCRIPT_NAME --json --output config/parse_config_test.go ParseConfig config
    bash $SCRIPT_NAME ParseConfig config > config/parse_config_test.go
EOF
}

# json_escape escapes a string for a JSON string literal: backslash, quote, and
# every control character below 0x20 (a bash string cannot hold 0x00).
json_escape() {
    local s="$1" out="" c code i
    s="${s//\\/\\\\}"
    s="${s//\"/\\\"}"
    for (( i = 0; i < ${#s}; i++ )); do
        c="${s:i:1}"
        case "$c" in
            $'\t') out+='\t' ;;
            $'\n') out+='\n' ;;
            $'\r') out+='\r' ;;
            *)
                printf -v code '%d' "'$c"
                if (( code < 32 )); then
                    printf -v c '\\u%04x' "$code"
                fi
                out+="$c"
                ;;
        esac
    done
    printf '%s' "$out"
}

OUTPUT=""
PARALLEL=false
JSON_OUTPUT=false
FORCE=false
POSITIONAL=()

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)    usage; exit 0 ;;
        -v|--version) echo "$SCRIPT_NAME v$VERSION"; exit 0 ;;
        --output)
            if [[ $# -lt 2 || "$2" == -* ]]; then
                echo "error: --output requires a file path" >&2
                exit 2
            fi
            OUTPUT="$2"; shift 2 ;;
        --force)      FORCE=true; shift ;;
        --parallel)   PARALLEL=true; shift ;;
        --json)       JSON_OUTPUT=true; shift ;;
        -*)           echo "error: unknown option: $1" >&2; usage >&2; exit 2 ;;
        *)            POSITIONAL+=("$1"); shift ;;
    esac
done

if [[ ${#POSITIONAL[@]} -lt 2 ]]; then
    echo "error: FuncName and package are required" >&2
    usage >&2
    exit 2
fi

if [[ ${#POSITIONAL[@]} -gt 2 ]]; then
    echo "error: unexpected argument: ${POSITIONAL[2]}" >&2
    usage >&2
    exit 2
fi

FUNC="${POSITIONAL[0]}"
PKG="${POSITIONAL[1]}"

if [[ ! "$FUNC" =~ ^[A-Z][A-Za-z0-9_]*$ ]]; then
    echo "error: FuncName '$FUNC' must be an exported Go identifier" >&2
    exit 2
fi

if [[ ! "$PKG" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]]; then
    echo "error: package '$PKG' must be a valid Go package identifier" >&2
    exit 2
fi

case "$PKG" in
    _|break|default|func|interface|select|case|defer|go|map|struct|chan|else|goto|package|switch|const|fallthrough|if|range|type|continue|for|import|return|var)
        echo "error: package '$PKG' must not be a Go keyword or blank identifier" >&2
        exit 2
        ;;
esac

generate_test() {
    local parallel_top="" parallel_sub=""
    if $PARALLEL; then
        parallel_top=$'\tt.Parallel()\n'
        parallel_sub=$'\t\t\tt.Parallel()\n'
    fi

    cat <<EOF
package ${PKG}

import (
	"testing"
)

func Test${FUNC}(t *testing.T) {
${parallel_top}	tests := []struct {
		name string
		give string
		want string
	}{
		{name: "empty", give: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
${parallel_sub}			got := ${FUNC}(tt.give)
			if got != tt.want {
				t.Errorf("${FUNC}(%q) = %q, want %q", tt.give, got, tt.want)
			}
		})
	}
}
EOF
}

if [[ -n "$OUTPUT" ]]; then
    OUTPUT_DIR="$(dirname -- "$OUTPUT")"
    if [[ ! -d "$OUTPUT_DIR" ]]; then
        echo "error: directory '$OUTPUT_DIR' does not exist" >&2
        exit 2
    fi
    if [[ -f "$OUTPUT" ]] && ! $FORCE; then
        echo "error: '$OUTPUT' already exists (use --force to overwrite)" >&2
        exit 2
    fi
    generate_test > "$OUTPUT"
    if $JSON_OUTPUT; then
        FUNC_ESC="$(json_escape "$FUNC")"
        PKG_ESC="$(json_escape "$PKG")"
        OUTPUT_ESC="$(json_escape "$OUTPUT")"
        cat <<EOF
{"func":"$FUNC_ESC","package":"$PKG_ESC","output_file":"$OUTPUT_ESC","parallel":$PARALLEL,"written":true}
EOF
    else
        echo "Wrote test scaffold to $OUTPUT"
    fi
else
    if $JSON_OUTPUT; then
        generate_test >&2
        FUNC_ESC="$(json_escape "$FUNC")"
        PKG_ESC="$(json_escape "$PKG")"
        cat <<EOF
{"func":"$FUNC_ESC","package":"$PKG_ESC","output_file":"","parallel":$PARALLEL,"written":false}
EOF
    else
        generate_test
    fi
fi

exit 0
