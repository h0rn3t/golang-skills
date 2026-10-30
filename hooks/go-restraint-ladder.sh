#!/usr/bin/env bash
# Restraint ladder hook, wired the way ponytail wires its ruleset. The ladder
# lives in go-code-refactor/references/LADDER.md, and a model sees it
# only if it reads that file: go-code requires the ladder on every task, yet
# the file was read in 0 of 20 skilled sessions of the 2026-09-10
# implementation control. This hook puts it in context without a Read.
#
#   SessionStart (startup, resume, clear, compact) and SubagentStart: in a
#     directory holding Go, print LADDER.md from the ladder heading to its
#     end, and the session's level. The whole body, not the ladder section
#     alone: given only the section, Haiku 5.5 read the file again in 261 of
#     678 Go sessions (2026-10-23..29), and Sonnet and Opus never saw
#     "Ship, Then Question".
#     compact is in the matcher because compaction drops what startup printed;
#     a subagent starts with an empty context. The plugin's go-verify agent
#     runs checks only and is skipped, as are the host's agents that write no
#     Go (Explore, claude-code-guide, statusline-setup) — 5 KB each they
#     would never use. Plan keeps the ladder: it shapes what a plan proposes.
#     A subagent sees SubagentStart output only as JSON additionalContext, so
#     that event prints the same text wrapped in it.
#   UserPromptSubmit: a level word — `/go-code ultra <task>`, `ultra mode`,
#     `режим ultra` — records the level for the rest of the session and
#     prints that level's line. Any other prompt prints nothing.
#
# Both texts are read at run time — the ladder from LADDER.md, the
# level line from the table in go-code/references/INTENSITY.md — so neither is a
# second copy. GOLANG_SKILLS_LADDER=lite|full|ultra sets the level a session
# starts at; off turns the hook off. Never blocks: exit 0 always.
set -u

input="$(cat)"
[[ "${GOLANG_SKILLS_LADDER:-}" == "off" ]] && exit 0
# shellcheck source=go-hook-helper.sh
source "$(dirname "${BASH_SOURCE[0]}")/go-hook-helper.sh"
helper="$(hook_helper)" || exit 0

parsed="$(printf '%s' "$input" | "$helper" ladder-parse)" || exit 0
[[ -n "$parsed" ]] || exit 0
{ read -r event; read -r session; read -r word; } <<< "$parsed"

state="${CLAUDE_PLUGIN_DATA:-${TMPDIR:-/tmp}/golang-skills-hooks}/routing/${session:-default}"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." 2>/dev/null && pwd)" || exit 0
ladder_file="$root/skills/go-code-refactor/references/LADDER.md"
levels_file="$root/skills/go-code/references/INTENSITY.md"

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
    printf 'golang-skills: restraint level `%s` for the rest of this session (levels in %s; a behavior-preserving refactor runs at `full`). %s\n' \
        "$level" "$levels_file" "$row"
    exit 0
fi

section="$(awk '/^## The Restraint Ladder$/ { on = 1 } on' "$ladder_file" 2>/dev/null)"
[[ -n "$section" ]] || exit 0

note="$(printf 'golang-skills: this project holds Go. Before writing Go, climb the restraint ladder below. It is %s from `## The Restraint Ladder` to its end, already in context: do not Read that file; its links resolve from it.\n' "$ladder_file"
printf 'Restraint level `%s`: %s Change it with `/go-code lite|full|ultra <task>` or `lite mode` / `ultra mode`; a behavior-preserving refactor runs at `full`.\n\n' \
    "$level" "$row"
printf '%s\n' "$section")"
if [[ "$event" == "SubagentStart" ]]; then
    printf '%s' "$note" | "$helper" context-json SubagentStart
else
    printf '%s\n' "$note"
fi
exit 0
