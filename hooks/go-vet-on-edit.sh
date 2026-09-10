#!/usr/bin/env bash
# PostToolUse hook: after Claude edits a .go file, run gofmt, go vet, and
# go fix -diff on its package and hand the findings back. Exit 2 makes stderr
# visible to Claude; exit 0 stays silent. Never blocks — the edit has already
# happened — and never applies go fix: the rewrite is reported, not written.
set -u

input="$(cat)"

# tool_input.file_path, parsed as JSON like go-code-routing.sh does, so an
# escaped quote or a "file_path" string inside the edited content cannot
# redirect the hook. Without python3 fall back to the regex extraction.
file=""
if command -v python3 >/dev/null 2>&1; then
    file="$(printf '%s' "$input" | python3 -c '
import json, sys
try:
    d = json.load(sys.stdin)
except Exception:
    sys.exit(0)
path = (d.get("tool_input") or {}).get("file_path") or ""
if isinstance(path, str):
    sys.stdout.write(path.replace("\n", ""))
' 2>/dev/null)" || file=""
else
    file="$(printf '%s' "$input" | sed -n 's/.*"file_path"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1)"
fi

case "$file" in
    *.go) ;;
    *) exit 0 ;;
esac
[[ -f "$file" ]] || exit 0
command -v go >/dev/null 2>&1 || exit 0

dir="$(dirname "$file")"
pkg="./$(basename "$dir")"
findings=""

# section <heading> <text> — the heading, the first 40 lines of text, and a
# count of what was cut, so a long diff cannot flood the transcript.
section() {
    local total
    total="$(printf '%s\n' "$2" | wc -l | tr -d ' ')"
    printf '%s\n' "$1"
    printf '%s\n' "$2" | head -n 40
    if [[ "$total" -gt 40 ]]; then
        printf '... %d more lines; run the command above for the full output\n' "$((total - 40))"
    fi
}

if unformatted="$(gofmt -l "$file" 2>&1)" && [[ -n "$unformatted" ]]; then
    findings+="gofmt: $file is not gofmt-formatted (run gofmt -w)"$'\n'
fi

# go vet and go fix need a module; skip both silently outside one (scratch
# files, examples).
if (cd "$dir" && go list -m >/dev/null 2>&1); then
    if ! vet_out="$(cd "$dir" && go vet . 2>&1)"; then
        findings+="$(section "go vet $pkg:" "$vet_out")"$'\n'
    fi
    # A package that does not type-check makes go fix restate vet's "vet: ..."
    # lines as "fix: ..."; skip it then. Otherwise judge go fix by what it
    # prints, not by its exit code: a pending rewrite may come back with 0 or 1.
    # Report only; -diff never writes.
    if ! printf '%s\n' "$vet_out" | grep -q '^vet: '; then
        fix_out="$(cd "$dir" && go fix -diff . 2>&1)" || true
        if [[ -n "$fix_out" ]]; then
            findings+="$(section "go fix -diff $pkg:" "$fix_out")"$'\n'
        fi
    fi
fi

if [[ -n "$findings" ]]; then
    printf '%s' "$findings" >&2
    exit 2
fi
exit 0
