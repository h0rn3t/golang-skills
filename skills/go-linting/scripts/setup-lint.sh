#!/usr/bin/env bash
set -euo pipefail

VERSION="1.0.0"
SCRIPT_NAME="$(basename "$0")"

usage() {
    cat <<EOF
$SCRIPT_NAME v$VERSION — Generate .golangci.yml and run initial lint

USAGE
    bash $SCRIPT_NAME [options] [local-prefix]

DESCRIPTION
    Creates a .golangci.yml from the bundled baseline (errcheck, revive,
    govet, staticcheck, the skill-enforcing linters, and the goimports
    formatter), verifies it against the golangci-lint schema, and runs
    golangci-lint. If local-prefix is provided, configures goimports to
    group local imports separately.

    Run it from the target module; it writes the config there and lints
    with exactly that file (--config).

    Exits 0 if lint passes, 1 if lint issues found, 2 on error (including
    an existing .golangci.{yml,yaml,toml,json} without --force, a generated
    config that fails schema verification, and a golangci-lint run that
    exits with an error rather than findings, such as no Go files).

OPTIONS
    -h, --help       Show this help message
    -v, --version    Show version
    --json           Output results as JSON
    --force          Overwrite an existing YAML config (.golangci.yaml if
                     present, else .golangci.yml); a .toml or .json config,
                     which golangci-lint reads first, must be removed first
    --dry-run        Print generated config to stdout without writing
                     (with --json: as the "config" string of a JSON object)
    --limit N        Max lint issue lines in JSON output (default: 50, 0 = unlimited)

ARGUMENTS
    local-prefix     Module path prefix for goimports grouping
                     (e.g., github.com/myorg/myrepo)

EXAMPLES
    bash $SCRIPT_NAME
    bash $SCRIPT_NAME github.com/myorg/myrepo
    bash $SCRIPT_NAME --force github.com/myorg/myrepo
    bash $SCRIPT_NAME --dry-run github.com/myorg/myrepo
    bash $SCRIPT_NAME --json
    bash $SCRIPT_NAME --json --limit 20
EOF
}

json_escape() {
    local s="$1" i c rep
    s="${s//\\/\\\\}"
    s="${s//\"/\\\"}"
    s="${s//$'\t'/\\t}"
    s="${s//$'\r'/\\r}"
    s="${s//$'\n'/\\n}"
    # The remaining control characters (a shell string holds no NUL).
    for (( i = 1; i < 32; i++ )); do
        case $i in 9|10|13) continue ;; esac
        printf -v c "\\x$(printf '%02x' "$i")"
        if [[ $s == *"$c"* ]]; then
            printf -v rep '\\u%04x' "$i"
            s="${s//$c/$rep}"
        fi
    done
    printf '%s' "$s"
}

JSON_OUTPUT=false
FORCE=false
DRY_RUN=false
LIMIT=50
LOCAL_PREFIX=""
PREFIX_SET=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)    usage; exit 0 ;;
        -v|--version) echo "$SCRIPT_NAME v$VERSION"; exit 0 ;;
        --json)       JSON_OUTPUT=true; shift ;;
        --force)      FORCE=true; shift ;;
        --dry-run)    DRY_RUN=true; shift ;;
        --limit)
            if [[ $# -lt 2 ]]; then
                echo "error: --limit requires a number" >&2
                exit 2
            fi
            LIMIT="$2"
            shift 2
            ;;
        -*)           echo "error: unknown option: $1" >&2; usage >&2; exit 2 ;;
        *)
            if $PREFIX_SET; then
                echo "error: unexpected argument: $1 (one local-prefix at most)" >&2
                usage >&2
                exit 2
            fi
            LOCAL_PREFIX="$1"
            PREFIX_SET=true
            shift
            ;;
    esac
done

if ! [[ "$LIMIT" =~ ^[0-9]+$ ]]; then
    echo "error: --limit must be a non-negative integer, got: $LIMIT" >&2
    exit 2
fi

ASSET="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/assets/golangci.yml"

generate_config() {
    if [[ ! -f "$ASSET" ]]; then
        echo "error: baseline config not found at $ASSET" >&2
        exit 2
    fi
    cat "$ASSET"

    if [[ -n "$LOCAL_PREFIX" ]]; then
        cat <<YAML
  settings:
    goimports:
      local-prefixes:
        - ${LOCAL_PREFIX}
YAML
    fi
}

CONFIG_PATH=".golangci.yml"

if $DRY_RUN; then
    if $JSON_OUTPUT; then
        CONFIG_TEXT="$(generate_config)"
        printf '{"config_path":"%s","local_prefix":"%s","created":false,"dry_run":true,"config":"%s"}\n' \
            "$(json_escape "$CONFIG_PATH")" "$(json_escape "$LOCAL_PREFIX")" "$(json_escape "$CONFIG_TEXT")"
    else
        generate_config
    fi
    exit 0
fi

# golangci-lint reads .golangci.json, .toml, .yaml, then .yml, so a file
# written beside another spelling would not be the config later runs use.
EXISTING=""
for f in .golangci.yml .golangci.yaml .golangci.toml .golangci.json; do
    if [[ -f "$f" ]]; then
        EXISTING="${EXISTING:+$EXISTING, }$f"
    fi
done
if [[ -n "$EXISTING" ]]; then
    if ! $FORCE; then
        echo "error: $EXISTING already exists (use --force to overwrite)" >&2
        exit 2
    fi
    for f in .golangci.toml .golangci.json; do
        if [[ -f "$f" ]]; then
            echo "error: $f would shadow the generated YAML config; remove it before --force" >&2
            exit 2
        fi
    done
    if [[ -f .golangci.yaml ]]; then
        CONFIG_PATH=".golangci.yaml"
    fi
fi

if ! command -v golangci-lint &>/dev/null; then
    echo "error: golangci-lint is not installed" >&2
    exit 2
fi

generate_config > "$CONFIG_PATH"

LINT_OUTPUT=""
LINT_EXIT=0
if ! VERIFY_OUTPUT=$(golangci-lint config verify --config "$CONFIG_PATH" 2>&1); then
    echo "error: $CONFIG_PATH failed schema verification:" >&2
    echo "$VERIFY_OUTPUT" >&2
    exit 2
fi

LINT_OUTPUT=$(golangci-lint run --config "$CONFIG_PATH" ./... 2>&1) || LINT_EXIT=$?
# Exit 1 is findings; any other non-zero exit (5: no Go files, 3: a go
# directive newer than the linter) is an environment error.
if [[ $LINT_EXIT -ne 0 && $LINT_EXIT -ne 1 ]]; then
    echo "error: $CONFIG_PATH was written, but golangci-lint run exited $LINT_EXIT:" >&2
    echo "$LINT_OUTPUT" >&2
    exit 2
fi

if $JSON_OUTPUT; then
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
    CONFIG_ESC="$(json_escape "$CONFIG_PATH")"
    PREFIX_ESC="$(json_escape "$LOCAL_PREFIX")"
    CREATED=true
    HAS_ISSUES=$( [[ $LINT_EXIT -ne 0 ]] && echo true || echo false )
    TRUNC_FIELD=""
    $LINT_TRUNCATED && TRUNC_FIELD=',"truncated":true'
    cat <<EOF
{"config_path":"$CONFIG_ESC","local_prefix":"$PREFIX_ESC","created":$CREATED,"lint_issues":$HAS_ISSUES,"lint_output":"$LINT_ESC"$TRUNC_FIELD}
EOF
else
    echo "Created $CONFIG_PATH"
    if [[ $LINT_EXIT -ne 0 ]]; then
        echo ""
        echo "$LINT_OUTPUT"
        echo ""
        echo "Lint issues found — fix them category by category (formatting first, then vet, then style)."
    else
        echo "golangci-lint: all clean."
    fi
fi

if [[ $LINT_EXIT -ne 0 ]]; then
    exit 1
fi
exit 0
