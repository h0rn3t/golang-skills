#!/usr/bin/env bash
# SessionStart and SubagentStart hook: both start with no record that a router
# was loaded (a subagent's context is empty; a compacted session has lost the
# prompt note). When the working directory holds Go, print one note naming
# go-code and go-style-core to load before the first edit, in the same
# Skill-tool sentence the UserPromptSubmit hook uses. The host adds stdout
# to the context.
#
# Fires for every subagent, since each one is a fresh context; skips the
# plugin's own go-verify agent, which runs checks and has no Skill tool, and
# the host's agents that write no Go: Explore searches, claude-code-guide
# answers questions about Claude Code, statusline-setup edits a setting. Plan
# still hears it, since a plan decides what gets written. It never blocks
# (SubagentStart cannot) and always exits 0.
set -u

input="$(cat)"
command -v python3 >/dev/null 2>&1 || exit 0

verdict="$(printf '%s' "$input" | python3 -c '
import json, os, sys
try:
    d = json.load(sys.stdin)
except Exception:
    sys.exit(0)
event = d.get("hook_event_name") or "SubagentStart"
if event not in ("SubagentStart", "SessionStart"):
    sys.exit(0)
# A plugin install names the agent "golang-skills:go-verify". SessionStart
# has no agent; the filter is only for a subagent that writes no Go.
if event == "SubagentStart":
    agent = (d.get("agent_type") or "").rsplit(":", 1)[-1].lower()
    if agent in ("go-verify", "explore", "claude-code-guide", "statusline-setup"):
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

if has_go_files(d.get("cwd") or ""):
    print("go")
')" || exit 0
[[ "$verdict" == "go" ]] || exit 0

# Exact Skill names, as in go-prompt-routing.sh: <plugin>:<skill> in a plugin.
ns=""
manifest="${CLAUDE_PLUGIN_ROOT:-}/.claude-plugin/plugin.json"
if [[ -n "${CLAUDE_PLUGIN_ROOT:-}" && -f "$manifest" ]]; then
    ns="$(python3 -c 'import json, sys; print(json.load(open(sys.argv[1])).get("name") or "")' "$manifest" 2>/dev/null)" || ns=""
fi
code="${ns:+$ns:}go-code"
style="${ns:+$ns:}go-style-core"
refactor="${ns:+$ns:}go-code-refactor"
lint="${ns:+$ns:}go-linting"
# One sentence, both skills named as Skill calls. A following line, or
# "if your task writes", was skipped and the first .go edit was blocked
# (2026-10-03, dhcore ap/operations.go). The prompt hook measured the same
# shape: both names in the router line, 21 of 21 first-message loads.
printf '%s\n' "golang-skills: this project holds Go code. Before the first edit, load the \`$code\` skill (Skill tool, name \`$code\`) and the \`$style\` skill (Skill tool, name \`$style\`). For a behavior-preserving refactor, load \`$refactor\` instead of \`$code\`. All of them in one message. Wait for successful Skill results; do not put Edit, Write or MultiEdit in the skill-loading tool message. If Bash is available, also load \`$lint\` (Skill tool, name \`$lint\`) in that message. Reading, searching, or reviewing only needs no edit skill. If the task is not Go work, ignore this note."
exit 0
