#!/usr/bin/env bash
# SubagentStart hook: a subagent starts with an empty context, so nothing the
# main session loaded reaches it. When the working directory holds Go, print
# one note naming the router to load before the first edit — the note the
# UserPromptSubmit hook gives the main session. The host adds stdout to the
# subagent's context.
#
# Fires for every subagent, since each one is a fresh context; skips the
# plugin's own go-verify agent, which runs checks and has no Skill tool. It
# never blocks (SubagentStart cannot) and always exits 0.
set -u

input="$(cat)"
command -v python3 >/dev/null 2>&1 || exit 0

verdict="$(printf '%s' "$input" | python3 -c '
import json, os, sys
try:
    d = json.load(sys.stdin)
except Exception:
    sys.exit(0)
if (d.get("hook_event_name") or "SubagentStart") != "SubagentStart":
    sys.exit(0)
# A plugin install names the agent "golang-skills:go-verify".
agent = (d.get("agent_type") or "").rsplit(":", 1)[-1]
if agent == "go-verify":
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

cat <<'NOTE'
golang-skills: this project holds Go code. If your task writes, fixes, or refactors Go, load the `go-code` skill (Skill tool, name `go-code`) before the first edit, or `go-code-refactor` for a behavior-preserving refactor; it loads the owner skills the task needs and closes with the verification gate. Reading, searching, or reviewing only needs no skill. If the task is not Go work, ignore this note.
NOTE
exit 0
