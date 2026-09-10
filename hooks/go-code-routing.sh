#!/usr/bin/env bash
# Routing gate for the go-code router. One script serves two hook events:
#
#   PostToolUse (Skill|Read)          record which go-* skills this session has
#                                     loaded, via the Skill tool or a direct
#                                     read of a sibling SKILL.md.
#   PreToolUse  (Edit|Write|MultiEdit) before an edit of a .go file in a session
#                                     that loaded go-code, require go-style-core
#                                     plus the owner skills the edit's content
#                                     points at. Exit 2 blocks the edit and names
#                                     the missing skills. Each skill is named at
#                                     most once per session, so a retry always
#                                     passes: the gate reminds, it cannot deadlock.
#
# Sessions that never loaded go-code are never touched. State lives under
# CLAUDE_PLUGIN_DATA when the host provides it, else under TMPDIR.
set -u

input="$(cat)"
command -v python3 >/dev/null 2>&1 || exit 0

# One parse: event, session, tool, skill name, file path, and the owner skills
# inferred from the edited content. Fields are newline-separated; the hint
# list is the last line and may be empty.
parsed="$(printf '%s' "$input" | python3 -c '
import json, re, sys
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
hints = []
if path.endswith("_test.go"):
    hints.append("go-testing")
for owner, pat in [
    ("go-http", r"\bnet/http\b|\bhttp\."),
    ("go-error-handling", r"fmt\.Errorf\(|\berrors\."),
    ("go-concurrency", r"\bgo func\b|\bchan\b|\bsync\.|<-"),
    ("go-context", r"\bcontext\."),
    ("go-database", r"database/sql|\bsql\.|\bpgx\b"),
    ("go-logging", r"\bslog\."),
    ("go-security", r"os/exec|html/template|\bcrypto/"),
]:
    if re.search(pat, text):
        hints.append(owner)
for v in (d.get("hook_event_name") or "", d.get("session_id") or "",
          d.get("tool_name") or "", skill, path, " ".join(hints)):
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
    has "$loaded" go-code || exit 0
    missing=""
    for want in go-style-core $hints; do
        has "$loaded" "$want" && continue
        has "$reminded" "$want" && continue
        missing+="$want "
    done
    [[ -n "$missing" ]] || exit 0
    mkdir -p "$state" && printf '%s\n' $missing >> "$reminded"
    cat >&2 <<EOF
go-code routing gate: this session loaded go-code but not: ${missing% }
Load them before editing Go files, then retry this edit. In Claude Code that is
one Skill call per name. This reminder names each skill once per session.
EOF
    exit 2
    ;;
esac
exit 0
