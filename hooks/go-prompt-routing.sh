#!/usr/bin/env bash
# UserPromptSubmit hook: when a prompt asks for Go work, remind the model which
# router to load before its first edit. Prose in a skill description reaches
# the model only if the host's skill matcher fires on the prompt; on
# 2026-09-11 Opus 5 loaded no skill in 5 of 18 sessions whose prompt said only
# "Implement the Go package in ./<dir>", and on 2026-09-10 Haiku 4.5 reached
# go-code-refactor in 1 of 4 refactor sessions. This hook sees the prompt
# itself, so it does not depend on the matcher.
#
#   review wording    -> go-code-review    (review, audit, ...)
#   refactor wording  -> go-code-refactor  (refactor, clean up, simplify, ...)
#   other Go work     -> go-code           (implement, write, add, fix, ...)
#
# Go work is a prompt that names Go (the word Go, golang, a .go file, go.mod,
# goroutines, a go subcommand) or a prompt sent from a directory holding
# go.mod or *.go files within two levels. A read-only question about Go code
# gets navigation guidance without loading an edit router.
#
# A prompt naming a router in its text (`use go-code`, `$go-code-refactor`,
# `/opsx:apply add-auth /go-code`) selects that router instead of silencing
# the reminder: a mention loads nothing, and such a prompt used to leave the
# loads to the model's choice. A path such as skills/go-code/SKILL.md does not
# count as a mention. The selected router is written to prompted, from which
# go-code-routing.sh takes the entry skill for a session without a router.
#
# Skill names in the note are exact for Skill: in a Claude Code plugin they are
# <plugin>:<skill> (golang-skills:go-code), without CLAUDE_PLUGIN_ROOT bare.
#
# A slash command is the exception. The host expands `/go-code` and
# `/golang-skills:go-code-refactor` itself: it inserts the router's SKILL.md
# and calls no tool, so PostToolUse never fires, go-code-routing.sh records no
# load, and its edit gate — which needs a router in `loaded` — stays silent for
# the rest of the session. Three such sessions on 2026-09-16..18 loaded no
# owner skill at all and left no `loaded` file behind, while every session that
# reached a router through the Skill tool loaded three to eight. A probe on
# CLI 2.1.267 shows why: a slash invocation raises UserPromptSubmit and nothing
# else. So for a slash command this hook records the router itself — which arms
# the gate — and names the owners the inserted file cannot load by itself.
#
# The reminder is printed once per skill per session on stdout, which the host
# adds to the model's context; it never blocks (exit 0 always). It stays silent
# when the session has already loaded the skill, using the same state directory
# go-code-routing.sh writes.
#
# With the router it names go-style-core and, when the prompt points at Go
# files (./dispatch, internal/x/y.go, or a bare word naming a directory of Go
# files under cwd), the owner skills their code points at, through
# go-code-routing.sh --hints: the same table the gate applies to each edit.
# On 2026-09-13 the gate blocked five edits in three sessions to name owners
# one at a time; a session that has the list before its first edit loads them
# without a block. Test files stay out of the scan.
#
# For go-code and go-code-refactor the note also names the idiom card by its
# installed path, since the gate requires one whole Read of it before the
# first .go edit and no skill wording made Sonnet 5 medium read it (0/24 on
# 2026-09-18): named here, the Read can land in the same message as the
# go-style-core load instead of costing a gate block.
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
# A slash command raises this event and no other: the host expands it, inserts
# the router SKILL.md, and calls no tool, so nothing else records the load.
slash = re.match(r"\s*/(?:[A-Za-z0-9_.-]+:)?(go-code-refactor|go-code-review|go-code)\b", prompt)
# A router named in the text is selected: a word at a boundary (space, $,
# quote, bracket), possibly with / or a namespace, and not part of a path or
# of a longer name.
mention = None if slash else re.search(
    r"(?:^|(?<=[\s`\x27\"(\[$]))/?(?:[A-Za-z0-9_.-]+:)?(go-code-refactor|go-code-review|go-code)(?![\w/-]|\.\w)",
    prompt)

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

# The Go files the prompt points at, for the owner hints.
def go_files(p):
    try:
        names = sorted(os.listdir(p))
    except OSError:
        return []
    return [os.path.join(p, n) for n in names if n.endswith(".go") and not n.endswith("_test.go")]

def target_files():
    files, seen = [], set()
    cwd = d.get("cwd") or ""
    if not os.path.isdir(cwd):
        return files
    for tok in re.findall(r"(?<![\w./-])(?:\.{1,2}/)?[\w.-]+(?:/[\w.-]+)*", prompt):
        tok = tok.rstrip(".,;:!?")
        if not tok or tok in seen or tok in (".", ".."):
            continue
        seen.add(tok)
        pathlike = "/" in tok or tok.endswith(".go")
        p = os.path.normpath(os.path.join(cwd, tok))
        if os.path.isdir(p):
            files.extend(go_files(p))
        elif pathlike and os.path.isfile(p) and p.endswith(".go") and not p.endswith("_test.go"):
            files.append(p)
    return files

# A slash command already names its router; the wording tests below are for the
# prompts that do not.
if slash:
    print((d.get("session_id") or "default").replace("\n", " "))
    print(slash.group(1))
    print("")
    print("slash")
    for p in target_files()[:40]:
        print(p)
    sys.exit(0)
if mention:
    print((d.get("session_id") or "default").replace("\n", " "))
    print(mention.group(1))
    print("")
    print("mention")
    for p in target_files()[:40]:
        print(p)
    sys.exit(0)

# A prompt that does not name Go counts as Go work only when the working
# directory holds Go and the prompt names a code element; "add a line to the
# README" in a Go repository must not fire.
code_noun = re.compile(
    r"\b(?:function\w?|func|method\w?|package\w?|handler\w?|endpoint\w?|struct\w?|interface\w?|type\w?|"
    r"test\w*|stub\w?|bug\w?|panic|бод[иі]\w*|функці\w*|метод\w*|пакет\w*|хендлер\w*|обробник\w*|"
    r"структур\w*|інтерфейс\w*|тест\w*|заглушк\w*|баг\w*|функци\w*|обработчик\w*|эндпоинт\w*)\b", re.I)
if not names_go(prompt):
    if not (has_go_files(d.get("cwd") or "") and code_noun.search(prompt)):
        sys.exit(0)

refactor = re.compile(
    r"\b(?:refactor\w*|clean(?:\s|-)?up|simplif\w*|restructur\w*|moderni[sz]\w*|tidy(?:\s|-)?up|reads? better|"
    r"messy|bloated|over-?engineer\w*|dead code|too long|hard to follow|monolith\w*|modulari[sz]\w*|"
    r"рефактор\w*|спрост\w*|упрост\w*|почист\w*|переструктур\w*|модерніз\w*|модерниз\w*|монол[иі]т\w*|модуляриз\w*)\b",
    re.I)
# Work verbs select an edit router; questions receive only navigation guidance.
work = re.compile(
    r"\b(?:implement\w*|write|add|create|build|make|fix\w*|bug\w*|stub\w*|not implemented|fill in|"
    r"реаліз\w*|реализ\w*|напиш\w*|напис\w*|дода\w*|добав\w*|виправ\w*|исправ\w*|створ\w*|созда\w*|зроби|сделай|баг\w*|"
    r"почин\w*|поправ\w*|допиш\w*|заполн\w*)\b",
    re.I)
navigate = re.search(
    r"\b(?:explain|find|locate|where|who calls|how does|show|trace|understand|"
    r"поясн\w*|объясн\w*|знайд\w*|найд\w*|покаж\w*|де\b|где\b|хто виклика\w*|кто вызыва\w*|"
    r"як працю\w*|как работа\w*)\b", prompt, re.I)
question = navigate and (prompt.strip().endswith("?") or re.match(
    r"\s*(?:explain|find|locate|where|who|how|show|trace|understand|"
    r"поясн\w*|объясн\w*|знайд\w*|найд\w*|покаж\w*|де\b|где\b|хто\b|кто\b|як\b|как\b)", prompt, re.I))
# Review is checked before the work verbs: the review corpus prompt says
# "what is wrong, and the fix", and `fix` used to send it to go-code with the
# condition "before the first edit", which a review never reaches.
review = re.compile(r"\b(?:review\w*|audit\w*|рев[ьи]ю\w*|ревью\w*)\b", re.I)
if question:
    skill, kind = "", "read-only Go code navigation"
elif review.search(prompt):
    skill, kind = "go-code-review", "a Go code review"
elif refactor.search(prompt):
    skill, kind = "go-code-refactor", "a behavior-preserving refactor"
elif work.search(prompt):
    skill, kind = "go-code", "Go code to write, implement, or fix"
else:
    if not navigate:
        sys.exit(0)
    skill, kind = "", "read-only Go code navigation"
print((d.get("session_id") or "default").replace("\n", " "))
print(skill)
print(kind)
print("navigate" if not skill else "note")
for p in target_files()[:40]:
    print(p)
')" || exit 0
[[ -n "$parsed" ]] || exit 0
{ read -r session; read -r skill; read -r kind; read -r mode; mapfile -t files; } <<< "$parsed"
[[ -n "$skill" || "$mode" == "navigate" ]] || exit 0

state="${CLAUDE_PLUGIN_DATA:-${TMPDIR:-/tmp}/golang-skills-hooks}/routing/${session:-default}"
has() { # has <file> <skill>
    [[ -f "$1" ]] && grep -qx -- "$2" "$1"
}
navigation_note() {
    [[ -f "$state/gopls-navigation" ]] && return
    mkdir -p "$state" && : > "$state/gopls-navigation"
    printf '%s\n' 'When gopls MCP tools are available: use go_workspace once, go_search for unknown Go symbols, go_file_context after reading a relevant Go file, go_package_api for package APIs, and go_symbol_references before changing an existing symbol. After a coherent edit batch, use go_diagnostics. Use rg for literal text, not to reconstruct Go symbol relationships.'
}
if [[ "$mode" == "navigate" ]]; then
    navigation_note
    exit 0
fi
has "$state/loaded" "$skill" && exit 0
navigation_note
if [[ "$mode" == "slash" ]]; then
    # The host inserted the skill file and called no tool, so record the load
    # here: go-code-routing.sh gates an edit only for a session whose `loaded`
    # names a router.
    mkdir -p "$state" && printf '%s\n' "$skill" >> "$state/loaded"
else
    has "$state/prompted" "$skill" && exit 0
    mkdir -p "$state" && printf '%s\n' "$skill" >> "$state/prompted"
fi

case "$skill" in
go-code-refactor)
    what="it records a baseline, deletes before it restructures, and verifies that behavior held" ;;
go-code-review)
    what="it routes each finding to the owner skill behind it" ;;
*)
    what="it loads the owner skills the task needs and closes with the verification gate" ;;
esac
# Exact names for Skill: the plugin registers skills as <plugin>:<skill>, and
# the host sets CLAUDE_PLUGIN_ROOT only for plugin hooks.
ns=""
manifest="${CLAUDE_PLUGIN_ROOT:-}/.claude-plugin/plugin.json"
if [[ -n "${CLAUDE_PLUGIN_ROOT:-}" && -f "$manifest" ]]; then
    ns="$(python3 -c 'import json, sys; print(json.load(open(sys.argv[1])).get("name") or "")' "$manifest" 2>/dev/null)" || ns=""
fi
q() { printf '%s%s' "${ns:+$ns:}" "$1"; }
# A review writes nothing: the note names go-code-review alone and sets the
# boundary "before the first finding"; the review picks the owners behind its
# findings itself. An owner list in this note pulled Opus 5.5 medium up to 9
# skills and +37% session cost (2026-09-29, n=6); without it the cost did not
# change.
if [[ "$skill" == "go-code-review" && "$mode" != "slash" ]]; then
    if [[ "$mode" == "mention" ]]; then
        printf 'golang-skills: this prompt names the `%s` skill; naming a skill does not load it.\n' "$(q "$skill")"
    else
        printf 'golang-skills: this prompt looks like %s.\n' "$kind"
    fi
    printf 'Before the first finding, load the `%s` skill (Skill tool, name `%s`); %s.\n' "$(q "$skill")" "$(q "$skill")" "$what"
    [[ "$mode" == "mention" ]] || printf 'If the task is not Go work, ignore this note.\n'
    exit 0
fi
owners=""
if (( ${#files[@]} > 0 )); then
    owners="$(bash "$(dirname "${BASH_SOURCE[0]}")/go-code-routing.sh" --hints "${files[@]}" 2>/dev/null)" || owners=""
fi
if [[ "$mode" == "slash" ]]; then
    line="Load \`$(q go-style-core)\`"
else
    line="Load \`$(q go-style-core)\` with it"
fi
if [[ -n "$owners" ]]; then
    list=""
    for o in $owners; do list+="\`$(q "$o")\`, "; done
    line+=", and the owners its code points at: ${list%, }"
fi
# The idiom card, at the path this plugin copy carries it; the edit gate
# requires one whole Read of it, and names it again if this note is not
# followed.
card=""
if [[ "$skill" != "go-code-review" ]]; then
    card="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." 2>/dev/null && pwd)/skills/go-style-core/references/CURRENT-GO.md"
    [[ -f "$card" ]] || card=""
fi
# One sentence carries the loads and the card: named in a sentence of its own
# ("in the same message as the go-style-core load"), the card pulled
# go-style-core off the owners' turn and cost 3.22 Skill turns a session
# against 2.44 (2026-09-18, Sonnet 5 medium, n=3).
line+="; \`$(q go-testing)\` if you write or edit a test"
[[ -z "$card" ]] || line+="; and Read the idiom card whole (no offset or limit): $card"
line+='. All of them in one message, before the first edit.'
if [[ "$mode" == "slash" ]]; then
    printf 'golang-skills: the `/%s` command inserted that skill file and loaded nothing else.\n' "$skill"
    printf '%s\n' "$line"
    exit 0
fi

if [[ "$mode" == "mention" ]]; then
    printf 'golang-skills: this prompt names the `%s` skill; naming a skill does not load it.\n' "$(q "$skill")"
else
    printf 'golang-skills: this prompt looks like %s.\n' "$kind"
fi
printf 'Before the first edit, load the `%s` skill (Skill tool, name `%s`); %s.\n' "$(q "$skill")" "$(q "$skill")" "$what"
printf '%s\n' "$line"
[[ "$mode" == "mention" ]] || printf 'If the task is not Go work, ignore this note.\n'
exit 0
