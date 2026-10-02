#!/usr/bin/env bash
# Explicit shell-profile entry point; cwd and argv select the check to credit.
set -eu
if [[ "${1:-}" == --gate ]]; then
    [[ $# -eq 3 ]] || { printf 'Usage: bash %s --gate <receipt-dir> <package-dir>\n' "$0" >&2; exit 2; }
    exec python3 "$(dirname "${BASH_SOURCE[0]}")/go-check-receipt.py" verify-gate "$2" "$3"
fi
if (( $# < 3 )); then
    printf 'Usage: bash %s <receipt.json> <expected-cwd> <command> [args...]\n' "$0" >&2
    exit 2
fi
if ! command -v python3 >/dev/null 2>&1; then
    printf 'Receipt verification unavailable: python3 not installed\n' >&2
    exit 2
fi
exec python3 "$(dirname "${BASH_SOURCE[0]}")/go-check-receipt.py" verify "$@"
