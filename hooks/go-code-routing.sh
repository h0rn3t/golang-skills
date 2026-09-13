#!/usr/bin/env bash
# Routing gate for the routers go-code, go-code-refactor, and go-code-review.
# One script serves two hook events and one helper mode:
#
#   --hints <file>...                 print the owner skills the files point
#                                     at, from the same table the gate uses;
#                                     go-prompt-routing.sh names them before
#                                     the first edit so a session loads them
#                                     in one go instead of one gate block per
#                                     owner (five blocks in three sessions on
#                                     2026-09-13).
#   PostToolUse (Skill|Read)          record which go-* skills this session has
#                                     loaded, via the Skill tool or a direct
#                                     read of a sibling SKILL.md.
#   PreToolUse  (Edit|Write|MultiEdit) before an edit of a .go file in a session
#                                     that loaded a router, require go-style-core
#                                     plus the owner skills the edit's content
#                                     points at. Exit 2 blocks the edit and names
#                                     the missing skills. Each skill is named at
#                                     most once per session, so a retry always
#                                     passes: the gate reminds, it cannot deadlock.
#
# The owner hints below are heuristics: regular expressions over the edited
# text that recognize the decision-bearing forms of thirteen owners (error
# wrapping, goroutines, context creation, SQL, slog, exec/templates, defer,
# type parameters, interfaces, main, retries, HTTP, tests).
# Routine syntax — a plain fmt.Errorf("%v"), r.Context(), an http.StatusOK in a
# comment, make([]T, n), append — must not fire. Collections have no hint on
# purpose: make/append appear in nearly every Go body, and the 2026-09-10
# implementation control plus its follow-up smoke show the forced
# go-data-structures load itself starting a make+copy -> slices.Clone rewrite
# that returned null for a nil list. The routing table in
# skills/go-code/SKILL.md ("Route Before The First Edit") is authoritative: it
# covers owners no regex can see (collections, naming, documentation,
# functions, performance, refactor, linting, troubleshooting) and its
# "Also load" conditions; this gate only reminds.
#
# Sessions that loaded no router are never touched. A refactor prompt reaches
# go-code-refactor alone (go-prompt-routing.sh), and in the 2026-09-10 and
# 2026-09-13 refactor runs such sessions loaded go-style-core in 2/20 and
# 6/20; a gate keyed on go-code alone stayed silent for them. State lives under
# CLAUDE_PLUGIN_DATA when the host provides it, else under TMPDIR.
set -u

# owner_hints_py is the one owner table, as python: hints(path, text) names
# the owner skills the decision-bearing forms in text point at. The gate runs
# it over an edit; --hints runs it over whole files for go-prompt-routing.sh,
# so the note before the first edit and the gate after it name the same
# owners.
owner_hints_py='
import re
# A test file has one owner. Its body is test plumbing — a defer, an
# http.Request, an errors.Is on a want — not the decisions the content hints
# below recognize; running them over a _test.go named go-defensive for every
# contract test in the 2026-09-10 runs.
# Heuristics, one decision-bearing pattern per owner. go-code/SKILL.md owns
# the routing decision; keep each regex narrow enough that ordinary syntax
# (fmt.Errorf with %v, r.Context(), http.StatusOK, "<-" inside a string)
# does not name an owner.
OWNER_PATTERNS = [
    ("go-http", r"\bnet/http\b|\bhttp\.(Handle|HandleFunc|Server\b|Client\b|NewServeMux|NewRequest|ResponseWriter|Error|ListenAndServe|Redirect|StatusCode)"),
    ("go-error-handling", r"fmt\.Errorf\([^)]*%w|\berrors\.(Is|As|AsType|Join|New)\("),
    ("go-concurrency", r"\bgo\s+func\b|\bgo\s+[A-Za-z_]\w*\(|\bmake\(chan\b|\bchan\s|\bsync\.(Mutex|RWMutex|WaitGroup|Once|Map)\b"),
    ("go-context", r"\bcontext\.(Background|TODO|With[A-Za-z]+|Context)\b"),
    ("go-database", r"database/sql|\bsql\.(Open|DB|Tx|Rows|Null)\b|\bpgx(pool)?\."),
    ("go-logging", r"\bslog\."),
    ("go-security", r"os/exec|html/template|text/template|\bcrypto/|\bexec\.Command|\bhttp\.(SetCookie|Cookie)\b|\bfilepath\.Join\("),
    ("go-defensive", r"\bdefer\s|\bunsafe\."),
    ("go-generics", r"\[[A-Z][A-Za-z0-9]*\s+(any|comparable|~|[A-Za-z]+\.[A-Za-z]+)\b"),
    ("go-interfaces", r"\binterface\s*\{"),
    ("go-packages", r"^package main\b|\bfunc main\("),
    ("go-resilience", r"x/time/rate|\bbackoff\b|\bRetry-After\b"),
]
def hints(path, text):
    if path.endswith("_test.go"):
        return ["go-testing"]
    out = [owner for owner, pat in OWNER_PATTERNS if re.search(pat, text, re.M)]
    # Mirror the go-code HTTP row: handlers always load go-error-handling too.
    if "go-http" in out and "go-error-handling" not in out:
        out.append("go-error-handling")
    return out
'

if [[ "${1:-}" == "--hints" ]]; then
    shift
    command -v python3 >/dev/null 2>&1 || exit 0
    python3 -c "$owner_hints_py"'
import sys
seen = []
for p in sys.argv[1:]:
    try:
        text = open(p, encoding="utf-8", errors="replace").read()
    except OSError:
        continue
    for h in hints(p, text):
        if h not in seen:
            seen.append(h)
print(" ".join(seen))
' "$@"
    exit 0
fi

input="$(cat)"
command -v python3 >/dev/null 2>&1 || exit 0

# One parse: event, session, tool, skill name, file path, and the owner skills
# inferred from the edited content. Fields are newline-separated; the hint
# list is the last line and may be empty.
parsed="$(printf '%s' "$input" | python3 -c "$owner_hints_py"'
import json, sys
try:
    d = json.load(sys.stdin)
except Exception:
    sys.exit(0)
ti = d.get("tool_input") or {}
skill = ti.get("skill") or ti.get("name") or ""
path = ti.get("file_path") or ""
text = ti.get("new_string") or ti.get("content") or ""
for e in ti.get("edits") or []:
    text += "\n" + (e.get("new_string") or "")
owners = hints(path, text)
for v in (d.get("hook_event_name") or "", d.get("session_id") or "",
          d.get("tool_name") or "", skill, path, " ".join(owners)):
    print(v.replace("\n", " "))
')" || exit 0
[[ -n "$parsed" ]] || exit 0
{ read -r event; read -r session; read -r tool; read -r skill; read -r path; read -r hints; } <<< "$parsed"

state="${CLAUDE_PLUGIN_DATA:-${TMPDIR:-/tmp}/golang-skills-hooks}/routing/${session:-default}"
loaded="$state/loaded"
reminded="$state/reminded"

record() { # record <skill> — a plugin install names skills "golang-skills:go-code"
    local name="${1##*:}"
    [[ "$name" =~ ^go-[a-z-]+$ ]] || return 0
    mkdir -p "$state" && printf '%s\n' "$name" >> "$loaded"
}

has() { # has <file> <skill>
    [[ -f "$1" ]] && grep -qx -- "$2" "$1"
}

case "$event" in
PostToolUse)
    case "$tool" in
    Skill) record "$skill" ;;
    Read)
        case "$path" in
        */go-*/SKILL.md) record "$(basename "$(dirname "$path")")" ;;
        esac
        ;;
    esac
    # Sessions end without notice; drop state older than two days.
    find "$(dirname "$state")" -mindepth 1 -maxdepth 1 -type d -mtime +2 -exec rm -rf {} + 2>/dev/null
    exit 0
    ;;
PreToolUse)
    case "$path" in *.go) ;; *) exit 0 ;; esac
    routers=""
    for r in go-code go-code-refactor go-code-review; do
        has "$loaded" "$r" && routers+="$r "
    done
    [[ -n "$routers" ]] || exit 0
    missing=""
    for want in go-style-core $hints; do
        has "$loaded" "$want" && continue
        has "$reminded" "$want" && continue
        missing+="$want "
    done
    [[ -n "$missing" ]] || exit 0
    mkdir -p "$state" && printf '%s\n' $missing >> "$reminded"
    cat >&2 <<EOF
golang-skills routing gate: this session loaded ${routers% } but not: ${missing% }
This edit was not applied and the file is unchanged. Load them (in Claude Code,
one Skill call per name, all in one message), then retry the same edit
against the unchanged file.
The gate reads the edited text and recognizes some owners only: tests, error
wrapping, goroutines, context creation, SQL, slog, exec and templates, defer,
type parameters, interfaces, main, retries, HTTP. The routing table in
go-code/SKILL.md decides, including the owners the gate cannot see; each skill
is named once per session, and the gate's silence is not a passing result.
EOF
    exit 2
    ;;
esac
exit 0
