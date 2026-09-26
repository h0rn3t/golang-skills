#!/usr/bin/env bash
# Restraint ladder hook, wired the way ponytail wires its ruleset. The ladder
# lives in go-code-refactor/references/OVER-ENGINEERING.md, and a model sees it
# only if it reads that file: go-code routes there only when a rung is in
# doubt, and the file was read in 0 of 20 skilled sessions of the 2026-09-10
# implementation control. This hook puts it in context without a Read.
#
#   SessionStart (startup, resume, clear, compact) and SubagentStart: in a
#     directory holding Go, print the ladder section and the session's level.
#     compact is in the matcher because compaction drops what startup printed;
#     a subagent starts with an empty context. The plugin's go-verify agent
#     runs checks only and is skipped.
#   UserPromptSubmit: a level word — `/go-code ultra <task>`, `ultra mode`,
#     `режим ultra` — records the level for the rest of the session and
#     prints that level's line. Any other prompt prints nothing.
#
# Both texts are read at run time — the ladder from OVER-ENGINEERING.md, the
# level line from the Intensity table in go-code/SKILL.md — so neither is a
# second copy. GOLANG_SKILLS_LADDER=lite|full|ultra sets the level a session
# starts at; off turns the hook off. Never blocks: exit 0 always.
set -u

input="$(cat)"
[[ "${GOLANG_SKILLS_LADDER:-}" == "off" ]] && exit 0
command -v python3 >/dev/null 2>&1 || exit 0

parsed="$(printf '%s' "$input" | python3 -c '
import json, os, re, sys
try:
    d = json.load(sys.stdin)
except Exception:
    sys.exit(0)

def has_go_files(root, depth=2):
    if not root or not os.path.isdir(root):
        return False
    base = root.rstrip(os.sep).count(os.sep)
    seen = 0
    for cur, dirs, files in os.walk(root):
        seen += 1
        if seen > 200:
            return False
        dirs[:] = [x for x in dirs if not x.startswith(".") and x not in ("vendor", "node_modules", "testdata")]
        if cur.count(os.sep) - base >= depth:
            dirs[:] = []
        if "go.mod" in files or any(f.endswith(".go") for f in files):
            return True
    return False

event = d.get("hook_event_name") or ""
word = ""
if event == "UserPromptSubmit":
    prompt = d.get("prompt")
    if not isinstance(prompt, str):
        sys.exit(0)
    # `/go-code-refactor ultra` sets nothing: a refactor runs at full.
    m = re.match(r"\s*/(?:[\w.-]+:)?go-code\s+(lite|full|ultra)\b", prompt) or \
        re.search(r"\b(lite|full|ultra)[ -]mode\b|\bрежим\s+(lite|full|ultra)\b", prompt, re.I)
    if not m:
        sys.exit(0)
    word = next(g for g in m.groups() if g).lower()
elif event in ("SessionStart", "SubagentStart"):
    if (d.get("agent_type") or "").rsplit(":", 1)[-1] == "go-verify":
        sys.exit(0)
    if not has_go_files(d.get("cwd") or ""):
        sys.exit(0)
else:
    sys.exit(0)
print(event)
print((d.get("session_id") or "default").replace("\n", " "))
print(word)
')" || exit 0
[[ -n "$parsed" ]] || exit 0
{ read -r event; read -r session; read -r word; } <<< "$parsed"

state="${CLAUDE_PLUGIN_DATA:-${TMPDIR:-/tmp}/golang-skills-hooks}/routing/${session:-default}"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." 2>/dev/null && pwd)" || exit 0
ladder_file="$root/skills/go-code-refactor/references/OVER-ENGINEERING.md"
levels_file="$root/skills/go-code/SKILL.md"

if [[ -n "$word" ]]; then
    mkdir -p "$state" && printf '%s\n' "$word" > "$state/intensity"
    level="$word"
else
    level="$(cat "$state/intensity" 2>/dev/null)"
fi
case "$level" in lite|full|ultra) ;; *) level="${GOLANG_SKILLS_LADDER:-full}" ;; esac
case "$level" in lite|full|ultra) ;; *) level=full ;; esac

row="$(awk -v l="$level" 'index($0, "| `" l "` | ") == 1 {
    sub(/^\| `[a-z]+` \| /, ""); sub(/ \|$/, ""); print; exit
}' "$levels_file" 2>/dev/null)"

if [[ "$event" == "UserPromptSubmit" ]]; then
    printf 'golang-skills: restraint level `%s` for the rest of this session (Intensity in %s; a behavior-preserving refactor runs at `full`). %s\n' \
        "$level" "$levels_file" "$row"
    exit 0
fi

section="$(awk '
    /^## The Restraint Ladder$/ { on = 1; print; next }
    on && /^## / { exit }
    on { print }
' "$ladder_file" 2>/dev/null)"
[[ -n "$section" ]] || exit 0

printf 'golang-skills: this project holds Go. Before writing Go, climb the restraint ladder below, read from %s; its links resolve from that file.\n' "$ladder_file"
printf 'Restraint level `%s`: %s Change it with `/go-code lite|full|ultra <task>` or `lite mode` / `ultra mode`; a behavior-preserving refactor runs at `full`.\n\n' \
    "$level" "$row"
printf '%s\n' "$section"
exit 0
