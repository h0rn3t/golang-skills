#!/usr/bin/env bash
# PostToolUse hook: after Claude edits a .go file, run gofmt, go vet,
# go fix -diff, the package's own tests, and golangci-lint on its package and
# hand the findings back. Exit 2 makes stderr visible to Claude; exit 0 stays
# silent. Never blocks — the edit has already happened — and never applies
# go fix: the rewrite is reported, not written.
#
# The tests and the linter are here because a session whose tool set has no
# shell never runs the verification gate itself: on 2026-09-12 every skilled
# Sonnet 5 `gateway` session left `w.Write` unchecked and one `feed` session
# in five shipped a `null` its own contract test would have caught, had
# anything run it. The hook is the host's process, so it runs either way.
#
#   go fix     -diff on the package; hunks in the edited file are printed and
#              the rest of the package is one count.
#   go test    -short -count=1 on the package, only when it type-checks and
#              has test files. GOLANG_SKILLS_EDIT_TESTS=off disables it for a
#              package whose tests need a service the hook cannot start;
#              GOLANG_SKILLS_EDIT_TEST_TIMEOUT sets its -timeout in seconds
#              (default 50). A timeout prints the running tests, not the
#              goroutine dump.
#   lint       golangci-lint on the package with the repository's own
#              configuration when one is found, the bundled
#              skills/go-linting/assets/golangci.yml otherwise. Findings in
#              the edited file are printed; the rest of the package is one
#              count, so a repository's older debt cannot flood the session.
#              In a git checkout with a HEAD only issues new since HEAD count:
#              on 2026-09-30 one edit of a large test file printed 40 lines of
#              findings older than the session, on every edit.
#              GOLANG_SKILLS_EDIT_LINT=off disables it.
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
        if [[ -n "$fix_out" ]] && printf '%s\n' "$fix_out" | grep -q '^--- '; then
            # Hunks are grouped under "--- <abs path> (old)"; match the path as
            # given and as the directory resolves (a symlinked TMPDIR).
            real="$(cd "$dir" && pwd -P)/$(basename "$file")"
            mine_fix="$(printf '%s\n' "$fix_out" | awk -v a="$file" -v b="$real" '/^--- /{cur=$2} cur==a || cur==b')"
            other_hunks="$(printf '%s\n' "$fix_out" | awk -v a="$file" -v b="$real" '/^--- /{cur=$2} cur!=a && cur!=b && /^@@ /{n++} END{print n+0}')"
            if [[ -n "$mine_fix" ]]; then
                findings+="$(section "go fix -diff $pkg:" "$mine_fix")"$'\n'
            fi
            if [[ "$other_hunks" -gt 0 ]]; then
                findings+="go fix -diff $pkg: $other_hunks hunk(s) in other files of the package; run go fix -diff . to list them"$'\n'
            fi
        elif [[ -n "$fix_out" ]]; then
            findings+="$(section "go fix -diff $pkg:" "$fix_out")"$'\n'
        fi

        # The package type-checked, so its tests and the linter can run too.
        # Both are capped at 60 s: a hook that outlives the host's timeout
        # reports nothing at all.
        if [[ "${GOLANG_SKILLS_EDIT_TESTS:-on}" != "off" ]] && compgen -G "$dir/*_test.go" >/dev/null; then
            tt="${GOLANG_SKILLS_EDIT_TEST_TIMEOUT:-50}"
            [[ "$tt" =~ ^[0-9]+$ ]] || tt=50
            if ! test_out="$(cd "$dir" && timeout "$((tt + 10))" go test -short -count=1 -timeout "${tt}s" . 2>&1)"; then
                if printf '%s\n' "$test_out" | grep -q '^panic: test timed out after'; then
                    test_out="$(printf '%s\n' "$test_out" | awk '/^panic: test timed out after/{p=1} p && /^goroutine /{exit} p && NF')"
                    test_out+=$'\n'"(goroutine dump cut: the package's tests outlast the hook's ${tt}s; run go test yourself, or set GOLANG_SKILLS_EDIT_TESTS=off)"
                fi
                findings+="$(section "go test $pkg:" "$test_out")"$'\n'
            fi
        fi
        if [[ "${GOLANG_SKILLS_EDIT_LINT:-on}" != "off" ]] && command -v golangci-lint >/dev/null 2>&1; then
            # --allow-parallel-runners: two sessions editing at once must not
            # fail each other's run on golangci-lint's global lock.
            lint_args=(run --allow-parallel-runners --path-mode=abs --output.text.print-issued-lines=false --show-stats=false)
            if (cd "$dir" && git rev-parse --verify -q HEAD >/dev/null 2>&1); then
                lint_args+=(--new-from-rev=HEAD)
            fi
            lint_args+=(.)
            # golangci-lint config path exits 0 and prints the file when the
            # repository has one; otherwise the bundled configuration applies.
            if ! (cd "$dir" && golangci-lint config path >/dev/null 2>&1); then
                bundled="${CLAUDE_PLUGIN_ROOT:-$(cd "$(dirname "$0")/.." && pwd)}/skills/go-linting/assets/golangci.yml"
                [[ -f "$bundled" ]] && lint_args+=(--config "$bundled")
            fi
            lint_out="$(cd "$dir" && timeout 60 golangci-lint "${lint_args[@]}" 2>/dev/null)" || true
            if [[ -n "$lint_out" ]]; then
                # Findings are "path:line:col: text (linter)"; keep the edited
                # file's and count the others.
                mine="$(printf '%s\n' "$lint_out" | grep -F "$file:" || true)"
                others="$(printf '%s\n' "$lint_out" | grep -E '^/.*:[0-9]+:[0-9]+: ' | grep -vF "$file:" | wc -l | tr -d ' ')"
                if [[ -n "$mine" ]]; then
                    findings+="$(section "golangci-lint $pkg:" "$mine")"$'\n'
                fi
                if [[ "$others" -gt 0 ]]; then
                    findings+="golangci-lint $pkg: $others finding(s) in other files of the package; run golangci-lint run . to list them"$'\n'
                fi
            fi
        fi
    fi
fi

if [[ -n "$findings" ]]; then
    printf '%s' "$findings" >&2
    exit 2
fi
exit 0
