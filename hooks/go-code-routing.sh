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
#                                     read of a sibling SKILL.md, and whether it
#                                     has read the idiom card
#                                     (go-style-core/references/CURRENT-GO.md)
#                                     whole.
#   PreToolUse  (Edit|Write|MultiEdit) before a .go edit, require a router
#                                     (go-code, go-code-refactor, or
#                                     go-code-review; with none, the one
#                                     go-prompt-routing.sh named, else
#                                     go-code), go-style-core, the owner skills
#                                     the edited text points at, and one whole
#                                     Read of the idiom card. Exit 2 blocks the
#                                     edit and names what is missing by exact
#                                     Skill names (golang-skills:go-code in a
#                                     plugin).
#
# loaded and reminded are kept apart: loaded holds only successful loads
# (Skill, a whole Read of SKILL.md, a slash command), reminded is only a log of
# what the gate has already named. A reminder is not a load, so a retry of the
# same edit without a load is blocked again. So that an unavailable skill does
# not cause endless retries: (1) a skill with no SKILL.md in this plugin copy
# is not required and is named in the message as missing; (2) every block
# gives a fallback, Read <root>/skills/<name>/SKILL.md, because Claude Code
# rejects a Skill call with an unknown name already in validateInput, and no
# hook sees that; (3) the third attempt in a row at the same edit without any
# new load is not blocked a third time but stops the session (continue: false)
# with an explanation for the user.
# GOLANG_SKILLS_ROUTING_GATE=off turns the gate off.
#
# The card is a gate item because no wording of go-code or go-style-core made
# Sonnet 5 medium read it: 0/24 sessions on 2026-09-18 under three wordings
# (a Resource Routing bullet, a numbered step of its own, the first clause of
# step 2), while Opus 5 reads it from the routing line in 8/8 and Haiku 4.5 in
# 4/5. A Read counts only whole — no offset, no limit shorter than the file —
# because the card's older rows apply at every go directive and a head over
# it misses what a Go 1.19 module still gets.
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
# "Also load" conditions; this gate only enforces the part it can see.
#
# A session without a router is blocked on its first .go edit too: it used to
# pass silently, and skipping the loads depended on the model's choice alone
# (Opus 5 loaded no skill in 5 of 18 sessions on 2026-09-11). A refactor prompt
# reaches go-code-refactor alone (go-prompt-routing.sh), and in the 2026-09-10
# and 2026-09-13 refactor runs such sessions loaded go-style-core in 2/20 and
# 6/20; a gate keyed on go-code alone stayed silent for them. State lives under
# CLAUDE_PLUGIN_DATA when the host provides it, else under TMPDIR.
set -u

# Root of this plugin copy: skills/ sits next to hooks/.
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." 2>/dev/null && pwd)"

# available <skill> — whether the skill has a SKILL.md in this plugin copy.
available() {
    [[ -f "$root/skills/$1/SKILL.md" ]]
}

# Names for a Skill call. A Claude Code plugin registers skills as
# <plugin>:<skill>; the host sets CLAUDE_PLUGIN_ROOT only for plugin hooks, so
# without it (manual wiring) the name stays bare. init_ns reads plugin.json
# once, and only where the gate prints something.
skill_ns=""
init_ns() {
    local manifest="${CLAUDE_PLUGIN_ROOT:-}/.claude-plugin/plugin.json"
    [[ -n "${CLAUDE_PLUGIN_ROOT:-}" && -f "$manifest" ]] || return 0
    skill_ns="$(python3 -c 'import json, sys; print(json.load(open(sys.argv[1])).get("name") or "")' \
        "$manifest" 2>/dev/null)" || skill_ns=""
}
# names <skill>... — `plugin:skill`, `plugin:skill2` for a message.
names() {
    local out="" s
    for s in "$@"; do out+="\`${skill_ns:+$skill_ns:}$s\`, "; done
    printf '%s' "${out%, }"
}

# emit_json stop|notice <text> — a PreToolUse JSON answer with exit 0. stop
# denies the edit and stops the session (continue: false), so the model does
# not repeat an edit the gate will not pass; notice only shows the text to the
# user.
emit_json() {
    python3 -c '
import json, sys
kind, text = sys.argv[1], sys.argv[2]
out = {"systemMessage": text}
if kind == "stop":
    out = {"continue": False, "stopReason": text,
           "hookSpecificOutput": {"hookEventName": "PreToolUse",
                                  "permissionDecision": "deny",
                                  "permissionDecisionReason": text}}
print(json.dumps(out))
' "$1" "$2"
}

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
    # Mirror the go-code HTTP and SQL rows: handlers and queries always load
    # go-error-handling too.
    if ("go-http" in out or "go-database" in out) and "go-error-handling" not in out:
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
' "$@" | while read -r -a owners; do
        # A skill missing from this copy is not named: the gate does not
        # require it.
        out=()
        for o in "${owners[@]}"; do available "$o" && out+=("$o"); done
        printf '%s\n' "${out[*]}"
    done
    exit 0
fi

input="$(cat)"
command -v python3 >/dev/null 2>&1 || exit 0

# One parse: event, session, tool, skill name, file path, and the owner skills
# inferred from the edited content. Fields are newline-separated; the hint
# list is the last line and may be empty.
parsed="$(printf '%s' "$input" | python3 -c "$owner_hints_py"'
import hashlib, json, sys
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
# A Read of the idiom card counts when it covers the whole file: no offset
# past the first line, no limit shorter than the card.
card = ""
if path.endswith("/go-style-core/references/CURRENT-GO.md"):
    card = "whole"
    try:
        if int(ti.get("offset") or 1) > 1:
            card = "partial"
        limit = ti.get("limit")
        if limit is not None:
            with open(path, encoding="utf-8", errors="replace") as f:
                if int(limit) < sum(1 for _ in f):
                    card = "partial"
    except (OSError, TypeError, ValueError):
        card = "partial"
# Edit key: a retry of the same edit after a block has the same key, while
# parallel edits of different files in one message have different keys.
key = hashlib.sha1((path + "\0" + text).encode("utf-8", "replace")).hexdigest()[:16]
for v in (d.get("hook_event_name") or "", d.get("session_id") or "",
          d.get("tool_name") or "", skill, path, " ".join(owners), card, key):
    print(v.replace("\n", " "))
')" || exit 0
[[ -n "$parsed" ]] || exit 0
{ read -r event; read -r session; read -r tool; read -r skill; read -r path; read -r hints; read -r card; read -r key; } <<< "$parsed"

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
        if [[ "$card" == "whole" ]]; then
            mkdir -p "$state" && : > "$state/card"
        fi
        ;;
    esac
    # Sessions end without notice; drop state older than two days.
    find "$(dirname "$state")" -mindepth 1 -maxdepth 1 -type d -mtime +2 -exec rm -rf {} + 2>/dev/null
    exit 0
    ;;
PreToolUse)
    case "$path" in *.go) ;; *) exit 0 ;; esac
    [[ "${GOLANG_SKILLS_ROUTING_GATE:-}" == off ]] && exit 0
    routers=""
    for r in go-code go-code-refactor go-code-review; do
        has "$loaded" "$r" && routers+="$r "
    done
    # Without a router the gate requires the entry skill: the one
    # go-prompt-routing.sh named last (the prompted file), else go-code.
    entry=""
    if [[ -z "$routers" ]]; then
        entry="$(grep -x -E 'go-code|go-code-refactor|go-code-review' "$state/prompted" 2>/dev/null | tail -n 1)"
        entry="${entry:-go-code}"
    fi
    # Only loaded lifts a requirement: reminded logs reminders, not loads.
    missing="" absent=""
    for want in $entry go-style-core $hints; do
        has "$loaded" "$want" && continue
        if available "$want"; then missing+="$want "; else absent+="$want "; fi
    done
    card_path="$root/skills/go-style-core/references/CURRENT-GO.md"
    need_card=""
    [[ -f "$card_path" && ! -f "$state/card" ]] && need_card="$card_path"
    fresh_absent=""
    for a in $absent; do has "$state/absent" "$a" || fresh_absent+="$a "; done
    [[ -n "$missing" || -n "$need_card" || -n "$fresh_absent" ]] || exit 0
    mkdir -p "$state"
    [[ -z "$fresh_absent" ]] || printf '%s\n' $fresh_absent >> "$state/absent"
    init_ns
    absent_note=""
    [[ -z "$fresh_absent" ]] || absent_note="Not installed in this plugin copy, so not required: $(names $fresh_absent) (no SKILL.md under $root/skills). Reinstall the plugin to restore them."
    if [[ -z "$missing" && -z "$need_card" ]]; then
        # Nothing to load; the user hears about a missing skill once, and the
        # edit passes.
        emit_json notice "golang-skills routing gate: $absent_note"
        exit 0
    fi

    # An attempt at the same edit with no progress (a new load or a whole
    # Read of the card) since the last block is counted; progress resets the
    # counter.
    sig="$( (wc -l < "$loaded") 2>/dev/null | tr -d ' ')"
    sig="${sig:-0}:$([[ -f "$state/card" ]] && echo 1 || echo 0)"
    attempt=1
    prev="$(grep -m1 "^$key " "$state/blocked" 2>/dev/null)"
    if [[ -n "$prev" ]]; then
        read -r _ prev_sig prev_attempt <<< "$prev"
        [[ "$prev_sig" == "$sig" ]] && attempt=$((prev_attempt + 1))
    fi
    { grep -v "^$key " "$state/blocked" 2>/dev/null; printf '%s %s %s\n' "$key" "$sig" "$attempt"; } > "$state/blocked.$$" &&
        mv "$state/blocked.$$" "$state/blocked"
    fresh=""
    for m in $missing; do has "$reminded" "$m" || fresh+="$m "; done
    [[ -z "$fresh" ]] || printf '%s\n' $fresh >> "$reminded"
    [[ -z "$need_card" ]] || has "$reminded" current-go-card || printf 'current-go-card\n' >> "$reminded"

    list="$(names $missing)"
    what="$list"
    [[ -z "$need_card" ]] || what+="${what:+, }the idiom card"
    if (( attempt >= 3 )); then
        emit_json stop "golang-skills routing gate: the same Go edit was blocked twice and no skill load was recorded in between (still missing: $what). Stopping instead of blocking it a third time: the skills look unavailable in this session, because the Skill call fails or neither Skill nor Read reaches $root/skills. Check the install with \`claude plugin list\`, or set GOLANG_SKILLS_ROUTING_GATE=off to switch the gate off."
        exit 0
    fi
    {
        if (( attempt == 2 )); then
            printf 'golang-skills routing gate: this edit was blocked before, and no load has been\n'
            printf 'recorded since. Still missing: %s.\n' "$what"
            printf 'A reminder is not a load: a retry without the loads is blocked again, and a\n'
            printf 'third stalled retry of this edit stops the session.\n'
        elif [[ -n "$entry" && " $missing " == *" $entry "* ]]; then
            printf 'golang-skills routing gate: this session loaded no router skill, and a .go edit\n'
            printf 'needs one first. Missing: %s' "$list"
            [[ -z "$need_card" ]] || printf ', and the idiom card is unread'
            printf '.\n'
            if [[ "$entry" == go-code ]] && available go-code-refactor; then
                printf 'For a behavior-preserving refactor, load %s instead of %s.\n' "$(names go-code-refactor)" "$(names go-code)"
            fi
        elif [[ -n "$missing" ]]; then
            printf 'golang-skills routing gate: this session loaded %s but not: %s' "${routers% }" "$list"
            [[ -z "$need_card" ]] || printf ', and has not read the idiom card'
            printf '.\n'
        else
            printf 'golang-skills routing gate: this session loaded %s but has not read the idiom card.\n' "${routers% }"
        fi
        if [[ -n "$missing" && -n "$need_card" ]]; then
            printf 'This edit was not applied and the file is unchanged. Load them (one Skill call\n'
            printf 'per name) and Read the card whole, no offset or limit, all in one message, then\n'
            printf 'retry the same edit against the unchanged file. The card:\n%s\n' "$need_card"
        elif [[ -n "$missing" ]]; then
            printf 'This edit was not applied and the file is unchanged. Load them (one Skill call\n'
            printf 'per name, all in one message), then retry the same edit against the unchanged file.\n'
        else
            printf 'This edit was not applied and the file is unchanged. Read the card whole, no\n'
            printf 'offset or limit, then retry the same edit against the unchanged file. The card:\n%s\n' "$need_card"
        fi
        if [[ -n "$missing" ]]; then
            printf 'If the Skill tool is missing or answers "Unknown skill", Read\n'
            printf '%s/skills/<name>/SKILL.md whole for each name instead; the gate counts that Read.\n' "$root"
        fi
        [[ -z "$absent_note" ]] || printf '%s\n' "$absent_note"
        # gopls: the hook cannot see whether this particular chat has MCP, so
        # it does not block on gopls; it names the route once per session, in
        # the first block, when the model is already planning its loads.
        # 2026-09-28, Opus 5.5 low, fetch: 0 gopls calls in 2/2 sessions
        # with MCP or CLI available ("the files were small enough to read
        # directly").
        if [[ ! -f "$state/gopls-hint" ]]; then
            : > "$state/gopls-hint"
            printf 'gopls: if go_workspace and go_file_context are in your tool list (a gopls MCP\n'
            printf 'server, often mcp__gopls__*), call go_workspace once and go_file_context on\n'
            printf '%s in the same message as the loads; a small package is not an exception.\n' "$path"
            printf 'Without them but with a shell: command -v gopls, then gopls check or gopls references.\n'
        fi
        if (( attempt == 1 )) && [[ -n "$fresh" ]]; then
            cat <<'EOF'
The gate reads the edited text and recognizes some owners only: tests, error
wrapping, goroutines, context creation, SQL, slog, exec and templates, defer,
type parameters, interfaces, main, retries, HTTP. The routing table in
go-code/SKILL.md decides, including the owners the gate cannot see, and the
gate's silence is not a passing result.
EOF
        elif (( attempt == 1 )) && [[ -z "$missing" ]]; then
            printf 'Its older rows apply at every go directive, and the gate'"'"'s silence is not a\n'
            printf 'passing result.\n'
        fi
    } >&2
    exit 2
    ;;
esac
exit 0
