#!/usr/bin/env bash
# PostToolUse hook: after Claude edits a .go file, run gofmt and go vet on its
# package and hand the findings back. Exit 2 makes stderr visible to Claude;
# exit 0 stays silent. Never blocks — the edit has already happened.
set -u

input="$(cat)"
file="$(printf '%s' "$input" | sed -n 's/.*"file_path"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1)"

case "$file" in
    *.go) ;;
    *) exit 0 ;;
esac
[[ -f "$file" ]] || exit 0
command -v go >/dev/null 2>&1 || exit 0

dir="$(dirname "$file")"
findings=""

if unformatted="$(gofmt -l "$file" 2>&1)" && [[ -n "$unformatted" ]]; then
    findings+="gofmt: $file is not gofmt-formatted (run gofmt -w)"$'\n'
fi

# go vet needs a module; skip silently outside one (scratch files, examples).
if (cd "$dir" && go list -m >/dev/null 2>&1); then
    if ! vet_out="$(cd "$dir" && go vet . 2>&1)"; then
        findings+="go vet ./$(basename "$dir"):"$'\n'"$(printf '%s\n' "$vet_out" | head -n 40)"$'\n'
    fi
fi

if [[ -n "$findings" ]]; then
    printf '%s' "$findings" >&2
    exit 2
fi
exit 0
