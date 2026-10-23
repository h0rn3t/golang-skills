#!/usr/bin/env bash
# Explicit shell-profile entry point; cwd and argv select the check to credit.
set -eu
here="$(dirname "${BASH_SOURCE[0]}")"
if [[ "${1:-}" == --gate ]]; then
    [[ $# -eq 3 ]] || { printf 'Usage: bash %s --gate <receipt-dir> <package-dir>\n' "$0" >&2; exit 2; }
    exec bash "$here/go-hook-helper.sh" receipt verify-gate "$2" "$3"
fi
if (( $# < 3 )); then
    printf 'Usage: bash %s <receipt.json> <expected-cwd> <command> [args...]\n' "$0" >&2
    exit 2
fi
exec bash "$here/go-hook-helper.sh" receipt verify "$@"
