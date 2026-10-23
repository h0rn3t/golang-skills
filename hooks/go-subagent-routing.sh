#!/usr/bin/env bash
# SessionStart and SubagentStart hook: both start with no record that a router
# was loaded (a subagent's context is empty; a compacted session has lost the
# prompt note). When the working directory holds Go, print one note naming
# go-code and go-style-core to load before the first edit, in the same
# Skill-tool sentence the UserPromptSubmit hook uses. SessionStart stdout
# reaches the context as it is; a subagent sees SubagentStart output only as
# JSON additionalContext and drops plain stdout, so that event prints the note
# wrapped in it.
#
# Fires for every subagent, since each one is a fresh context; skips the
# plugin's own go-verify agent, which runs checks and has no Skill tool, and
# the host's agents that write no Go: Explore searches, claude-code-guide
# answers questions about Claude Code, statusline-setup edits a setting. Plan
# still hears it, since a plan decides what gets written. It never blocks
# (SubagentStart cannot) and always exits 0.
set -u

# As in go-prompt-routing.sh: on Claude the router line stays as it is.
plugin_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
claude_note() {
    [[ -n "${CLAUDE_PROJECT_DIR:-}" && -z "${PLUGIN_ROOT:-}" ]]
}
read_alt() {
    if claude_note; then
        return 0
    fi
    printf ', or read %s/skills/%s/SKILL.md whole' "$plugin_root" "$1"
}

input="$(cat)"
# shellcheck source=go-hook-helper.sh
source "$(dirname "${BASH_SOURCE[0]}")/go-hook-helper.sh"
helper="$(hook_helper)" || exit 0

verdict="$(printf '%s' "$input" | "$helper" subagent-parse)" || exit 0
[[ "$verdict" == "SessionStart" || "$verdict" == "SubagentStart" ]] || exit 0

# Exact Skill names, as in go-prompt-routing.sh: <plugin>:<skill> in a plugin.
ns=""
manifest="${CLAUDE_PLUGIN_ROOT:-}/.claude-plugin/plugin.json"
if [[ -n "${CLAUDE_PLUGIN_ROOT:-}" && -f "$manifest" ]]; then
    ns="$("$helper" plugin-name "$manifest")" || ns=""
fi
code="${ns:+$ns:}go-code"
style="${ns:+$ns:}go-style-core"
refactor="${ns:+$ns:}go-code-refactor"
lint="${ns:+$ns:}go-linting"
testing="${ns:+$ns:}go-testing"
# One sentence, both skills named as Skill calls. A following line, or
# "if your task writes", was skipped and the first .go edit was blocked
# (2026-10-03, dhcore ap/operations.go). The prompt hook measured the same
# shape: both names in the router line, 21 of 21 first-message loads.
note="golang-skills: this project holds Go code. Before the first edit, load the \`$code\` skill (Skill tool, name \`$code\`$(read_alt go-code)) and the \`$style\` skill (Skill tool, name \`$style\`$(read_alt go-style-core)). For a behavior-preserving refactor, load \`$refactor\` instead of \`$code\`, and the \`$testing\` skill (Skill tool, name \`$testing\`$(read_alt go-testing)) too, since a refactor often edits tests. All of them in one message. Wait for successful Skill results; do not put Edit, Write or MultiEdit in the skill-loading tool message. If Bash is available, also load \`$lint\` (Skill tool, name \`$lint\`$(read_alt go-linting)) in that message. Reading, searching, or reviewing only needs no edit skill. If the task is not Go work, ignore this note."
if [[ "$verdict" == "SubagentStart" ]]; then
    printf '%s' "$note" | "$helper" context-json SubagentStart
else
    printf '%s\n' "$note"
fi
exit 0
