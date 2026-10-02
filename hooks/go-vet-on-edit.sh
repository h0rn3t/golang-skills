#!/usr/bin/env bash
# PostToolUse hook: after Claude edits a .go file, run gofmt, go vet,
# go fix -diff, the package's own tests, and golangci-lint on its package and
# hand findings and explicit check receipts back. Exit 2 makes stderr visible
# to Claude; exit 0 emits PostToolUse additionalContext. Never blocks and never applies
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
case "$file" in /*) ;; *) file="$PWD/$file" ;; esac
dir="$(dirname "$file")"
pkg="./$(basename "$dir")"
findings=""
export GOLANG_SKILLS_HOOK_DEADLINE="$(( $(date +%s) + 135 ))"
receipt_helper="$(cd "$(dirname "$0")" && pwd)/go-check-receipt.py"
receipt_dir=""
if command -v python3 >/dev/null 2>&1 && [[ -f "$receipt_helper" ]]; then
    receipt_base="${CLAUDE_PLUGIN_DATA:-${TMPDIR:-/tmp}/golang-skills-hooks}/checks"
    if mkdir -p "$receipt_base"; then
        receipt_dir="$(mktemp -d "$receipt_base/run.XXXXXX")" || receipt_dir=""
        if [[ -n "$receipt_dir" ]]; then
            session="$(printf '%s' "$input" | python3 -c 'import json,sys; print(json.load(sys.stdin).get("session_id") or "default")')"
            python3 "$receipt_helper" begin "$receipt_dir" "$file" "$session" "$(dirname "$receipt_base")" || receipt_dir=""
        fi
    fi
fi
config=""
filter="all"
skip_reason="no module"
run_check() {
    local check="$1"
    shift
    if [[ -n "$receipt_dir" ]]; then
        python3 "$receipt_helper" run "$receipt_dir" "$check" "$file" "$config" "$filter" run "" "$@"
    else
        (cd -P "$dir" && "$@")
    fi
}
skip_check() {
    [[ -n "$receipt_dir" ]] || return 0
    python3 "$receipt_helper" run "$receipt_dir" "$1" "$file" "$config" "$filter" "$2" "$3" "${@:4}" >/dev/null
}
has_module() {
    if [[ -n "$receipt_dir" ]]; then
        python3 "$receipt_helper" module "$dir" >/dev/null 2>&1
    else
        (cd "$dir" && go list -m >/dev/null 2>&1)
    fi
}

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

fmt_status=0
unformatted="$(run_check gofmt gofmt -l "$file" 2>&1)" || fmt_status=$?
if [[ "$fmt_status" -ne 0 ]]; then
    findings+="$(section "gofmt: failed (exit $fmt_status):" "$unformatted")"$'\n'
elif [[ -n "$unformatted" ]]; then
    findings+="gofmt: $file is not gofmt-formatted (run gofmt -w)"$'\n'
fi

# go vet and go fix need a module; receipts distinguish skips from success.
if has_module; then
    if ! vet_out="$(run_check vet go vet . 2>&1)"; then
        findings+="$(section "go vet $pkg:" "$vet_out")"$'\n'
    fi
    # A package that does not type-check makes go fix restate vet's "vet: ..."
    # lines as "fix: ..."; skip it then. Otherwise judge go fix by what it
    # prints, not by its exit code: a pending rewrite may come back with 0 or 1.
    # Report only; -diff never writes.
    if ! printf '%s\n' "$vet_out" | grep -q '^vet: '; then
        fix_status=0
        fix_out="$(run_check fix go fix -diff . 2>&1)" || fix_status=$?
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
        elif [[ "$fix_status" -ne 0 ]]; then
            findings+="go fix -diff $pkg: failed (exit $fix_status)"$'\n'
        fi

        # The package type-checked, so its tests and the linter can run too.
        # Both are capped at 60 s: a hook that outlives the host's timeout
        # reports nothing at all.
        if [[ "${GOLANG_SKILLS_EDIT_TESTS:-on}" != "off" ]] && compgen -G "$dir/*_test.go" >/dev/null; then
            tt="${GOLANG_SKILLS_EDIT_TEST_TIMEOUT:-50}"
            [[ "$tt" =~ ^[0-9]+$ ]] || tt=50
            if ! test_out="$(run_check test timeout "$((tt + 10))" go test -short -count=1 -timeout "${tt}s" . 2>&1)"; then
                if printf '%s\n' "$test_out" | grep -q '^panic: test timed out after'; then
                    test_out="$(printf '%s\n' "$test_out" | awk '/^panic: test timed out after/{p=1} p && /^goroutine /{exit} p && NF')"
                    test_out+=$'\n'"(goroutine dump cut: the package's tests outlast the hook's ${tt}s; run go test yourself, or set GOLANG_SKILLS_EDIT_TESTS=off)"
                fi
                findings+="$(section "go test $pkg:" "$test_out")"$'\n'
            fi
        elif [[ "${GOLANG_SKILLS_EDIT_TESTS:-on}" == "off" ]]; then
            skip_check test skipped "disabled by GOLANG_SKILLS_EDIT_TESTS" go test -short -count=1 .
        else
            skip_check test skipped "no test files" go test -short -count=1 .
        fi
        if [[ "${GOLANG_SKILLS_EDIT_LINT:-on}" != "off" ]] && command -v golangci-lint >/dev/null 2>&1; then
            # --allow-parallel-runners: two sessions editing at once must not
            # fail each other's run on golangci-lint's global lock.
            lint_args=(run --allow-parallel-runners --path-mode=abs --output.text.print-issued-lines=false --show-stats=false)
            if (cd "$dir" && git rev-parse --verify -q HEAD >/dev/null 2>&1); then
                lint_args+=(--new-from-rev=HEAD)
                filter="new-since-HEAD"
            fi
            lint_args+=(.)
            # Обираємо config у логічному каталозі до cd -P. `config path`
            # пише шлях у stderr, відносно фізичного cwd; явно закріплюємо його,
            # щоб пошук у фізичному каталозі не обрав іншу конфігурацію.
            lint_config_status=0
            if [[ -n "$receipt_dir" ]]; then
                repo_config="$(python3 "$receipt_helper" config "$dir" 2>&1)" || lint_config_status=$?
            else
                repo_config="$(cd "$dir" && golangci-lint config path 2>&1)" || lint_config_status=$?
            fi
            if [[ "$lint_config_status" -eq 0 ]]; then
                case "$repo_config" in
                    /*) ;;
                    *) repo_config="$(cd "$dir" && pwd -P)/$repo_config" ;;
                esac
                lint_args+=(--config "$repo_config")
                config="$repo_config"
            elif [[ "$lint_config_status" -eq 1 ]] || { [[ "$lint_config_status" -eq 6 ]] && [[ "$repo_config" == *"No config file detected"* ]]; }; then
                bundled="${CLAUDE_PLUGIN_ROOT:-$(cd "$(dirname "$0")/.." && pwd)}/skills/go-linting/assets/golangci.yml"
                [[ -f "$bundled" ]] && lint_args+=(--config "$bundled")
                [[ -f "$bundled" ]] && config="$bundled"
                lint_config_status=0
            fi
            # Логічний PWD через symlink не збігається з фізичними шляхами
            # findings: --new-from-rev=HEAD тоді відфільтровує навіть нові дефекти.
            real="$(cd "$dir" && pwd -P)/$(basename "$file")"
            lint_status=0
            if [[ "$lint_config_status" -gt 1 ]]; then
                skip_check lint unavailable "lint config discovery failed (exit $lint_config_status)" golangci-lint "${lint_args[@]}"
                lint_status="$lint_config_status"
                lint_out="$repo_config"
            else
                lint_out="$(run_check lint timeout 60 golangci-lint "${lint_args[@]}" 2>&1)" || lint_status=$?
            fi
            if [[ -n "$lint_out" || "$lint_status" -ne 0 ]]; then
                # Зберігаємо scope findings для обох варіантів шляху файлу;
                # діагностика невдалого запуску не є finding іншого файлу.
                issues="$(printf '%s\n' "$lint_out" | grep -E '^/.*:[0-9]+:[0-9]+: ' || true)"
                mine="$(printf '%s\n' "$issues" | grep -F -e "$file:" -e "$real:" || true)"
                others="$(printf '%s\n' "$issues" | grep -E '^/.*:[0-9]+:[0-9]+: ' | grep -vF -e "$file:" -e "$real:" | wc -l | tr -d ' ')"
                if [[ -n "$mine" ]]; then
                    findings+="$(section "golangci-lint $pkg:" "$mine")"$'\n'
                fi
                if [[ "$others" -gt 0 ]]; then
                    findings+="golangci-lint $pkg: $others finding(s) in other files of the package; run golangci-lint run . to list them"$'\n'
                fi
                lint_diag="$(printf '%s\n' "$lint_out" | grep -vE '^/.*:[0-9]+:[0-9]+: ' || true)"
                if [[ "$lint_status" -ne 0 ]] && { [[ "$lint_status" -ne 1 ]] || [[ -n "$lint_diag" || -z "$issues" ]]; }; then
                    findings+="$(section "golangci-lint $pkg: failed (exit $lint_status):" "${lint_diag:-golangci-lint did not complete successfully}")"$'\n'
                fi
            fi
        elif [[ "${GOLANG_SKILLS_EDIT_LINT:-on}" == "off" ]]; then
            skip_check lint skipped "disabled by GOLANG_SKILLS_EDIT_LINT" golangci-lint run .
        else
            skip_check lint unavailable "golangci-lint not installed" golangci-lint run .
        fi
    else
        skip_reason="package does not type-check"
    fi
elif ! command -v go >/dev/null 2>&1; then
    skip_reason="Go not installed"
fi

if [[ -n "$receipt_dir" ]]; then
    for check in vet fix test lint; do
        if [[ ! -f "$receipt_dir/$check.json" ]]; then
            check_status=skipped
            [[ "$skip_reason" == "Go not installed" ]] && check_status=unavailable
            case "$check" in
                vet) skip_check vet "$check_status" "$skip_reason" go vet . ;;
                fix) skip_check fix "$check_status" "$skip_reason" go fix -diff . ;;
                test) skip_check test "$check_status" "$skip_reason" go test -short -count=1 . ;;
                lint) skip_check lint "$check_status" "$skip_reason" golangci-lint run . ;;
            esac
        fi
    done
else
    findings+="Edit-hook receipts unavailable (python3, helper or writable state missing); silence cannot count as pass (hook)."$'\n'
fi
if [[ -n "$findings" ]]; then
    if [[ -n "$receipt_dir" ]]; then
        # Exit 2 ignores structured stdout: put the same concrete next step
        # on stderr before findings, including verification for passing checks.
        receipt_context="$(python3 "$receipt_helper" context "$receipt_dir" 2>&1)" || receipt_context="Edit-hook receipts unavailable; do not credit pass (hook)."
        printf '%s\n' "$receipt_context" >&2
    fi
    printf '%s' "$findings" >&2
    exit 2
fi
python3 "$receipt_helper" report "$receipt_dir"
exit 0
