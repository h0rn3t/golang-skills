#!/usr/bin/env bash
# UserPromptSubmit hook: when a prompt asks for Go work, remind the model which
# router to load before its first edit. Prose in a skill description reaches
# the model only if the host's skill matcher fires on the prompt; on
# 2026-09-11 Opus 5 loaded no skill in 5 of 18 sessions whose prompt said only
# "Implement the Go package in ./<dir>", and on 2026-09-10 Haiku 4.5 reached
# go-code-refactor in 1 of 4 refactor sessions. This hook sees the prompt
# itself, so it does not depend on the matcher.
#
#   refactor wording  -> go-code-refactor  (refactor, clean up, simplify, ...)
#   other Go work     -> go-code           (implement, write, add, fix, ...)
#
# Go work is a prompt that names Go (the word Go, golang, a .go file, go.mod,
# goroutines, a go subcommand) or a prompt sent from a directory holding
# go.mod or *.go files within two levels. A prompt with no work verb — a
# question, an explanation — is left alone, as is a prompt that already
# invokes a go-* skill by name.
#
# The reminder is printed once per skill per session on stdout, which the host
# adds to the model's context; it never blocks (exit 0 always). It stays silent
# when the session has already loaded the skill, using the same state directory
# go-code-routing.sh writes.
set -u

input="$(cat)"
command -v python3 >/dev/null 2>&1 || exit 0

# One parse: session id and the skill to name, or nothing. The prompt text
# never leaves python; only the verdict is printed.
parsed="$(printf '%s' "$input" | python3 -c '
import json, os, re, sys
try:
    d = json.load(sys.stdin)
except Exception:
    sys.exit(0)
event = d.get("hook_event_name") or "UserPromptSubmit"
if event != "UserPromptSubmit":
    sys.exit(0)
prompt = d.get("prompt") or ""
if not isinstance(prompt, str) or not prompt.strip():
    sys.exit(0)
# The user is already invoking a skill from this pack; the matcher has fired.
if re.search(r"(^|[\s/$:])go-(code|code-refactor|code-review)\b", prompt):
    sys.exit(0)

def names_go(text):
    return re.search(
        r"\bGo\b|\bgolang\b|\b[\w./-]+\.go\b|\bgo\.mod\b|\bgoroutines?\b|\bgofmt\b|\bgo (?:build|test|vet|run|mod|fix)\b",
        text) is not None

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

# A prompt that does not name Go counts as Go work only when the working
# directory holds Go and the prompt names a code element; "add a line to the
# README" in a Go repository must not fire.
code_noun = re.compile(
    r"\b(?:function\w?|func|method\w?|package\w?|handler\w?|endpoint\w?|struct\w?|interface\w?|type\w?|"
    r"test\w*|stub\w?|bug\w?|panic|бод[иі]\w*|функці\w*|метод\w*|пакет\w*|хендлер\w*|обробник\w*|"
    r"структур\w*|інтерфейс\w*|тест\w*|заглушк\w*|баг\w*)\b", re.I)
if not names_go(prompt):
    if not (has_go_files(d.get("cwd") or "") and code_noun.search(prompt)):
        sys.exit(0)

refactor = re.compile(
    r"\b(?:refactor\w*|clean(?:\s|-)?up|simplif\w*|restructur\w*|moderni[sz]\w*|tidy(?:\s|-)?up|reads? better|"
    r"messy|bloated|over-?engineer\w*|dead code|too long|hard to follow|"
    r"рефактор\w*|спрост\w*|упрост\w*|почист\w*|переструктур\w*|модерніз\w*|модерниз\w*)\b",
    re.I)
# Work verbs only: a noun such as "function" or "handler" appears in questions
# too, and a question gets no note.
work = re.compile(
    r"\b(?:implement\w*|write|add|create|build|make|fix\w*|bug\w*|stub\w*|not implemented|fill in|"
    r"реаліз\w*|реализ\w*|напиш\w*|напис\w*|дода\w*|добав\w*|виправ\w*|исправ\w*|створ\w*|созда\w*|зроби|сделай|баг\w*)\b",
    re.I)
if refactor.search(prompt):
    skill, kind = "go-code-refactor", "a behavior-preserving refactor"
elif work.search(prompt):
    skill, kind = "go-code", "Go code to write, implement, or fix"
else:
    sys.exit(0)
print((d.get("session_id") or "default").replace("\n", " "))
print(skill)
print(kind)
')" || exit 0
[[ -n "$parsed" ]] || exit 0
{ read -r session; read -r skill; read -r kind; } <<< "$parsed"
[[ -n "$skill" ]] || exit 0

state="${CLAUDE_PLUGIN_DATA:-${TMPDIR:-/tmp}/golang-skills-hooks}/routing/${session:-default}"
has() { # has <file> <skill>
    [[ -f "$1" ]] && grep -qx -- "$2" "$1"
}
has "$state/loaded" "$skill" && exit 0
has "$state/prompted" "$skill" && exit 0
mkdir -p "$state" && printf '%s\n' "$skill" >> "$state/prompted"

case "$skill" in
go-code-refactor)
    what="it records a baseline, deletes before it restructures, and verifies that behavior held" ;;
*)
    what="it loads the owner skills the task needs and closes with the verification gate" ;;
esac
printf 'golang-skills: this prompt looks like %s.\n' "$kind"
printf 'Before the first edit, load the `%s` skill (Skill tool, name `%s`); %s.\n' "$skill" "$skill" "$what"
printf 'If the task is not Go work, ignore this note.\n'
exit 0
