#!/usr/bin/env bash
# UserPromptSubmit hook: when a prompt asks for Go work, remind the model which
# router to load before its first edit. Prose in a skill description reaches
# the model only if the host's skill matcher fires on the prompt; on
# 2026-09-11 Opus 5 loaded no skill in 5 of 18 sessions whose prompt said only
# "Implement the Go package in ./<dir>", and on 2026-09-10 Haiku 4.5 reached
# go-code-refactor in 1 of 4 refactor sessions. This hook sees the prompt
# itself, so it does not depend on the matcher.
#
#   over-engineering  -> go-code-refactor  (over-engineered, dead code, what can we delete)
#   review wording    -> go-code-review    (review, audit, ...)
#   refactor wording  -> go-code-refactor  (refactor, clean up, simplify, split, ...)
#   other Go work     -> go-code           (implement, write, add, fix, ...)
#   pasted Go trace   -> go-troubleshooting (panic: or fatal error: with a goroutine header, no work verb)
#   prose edits       -> no note           (translate into a human language, grammar)
#
# Go work is a prompt that names Go (the word Go, golang, a .go file, go.mod,
# goroutines, a go subcommand) or a prompt sent from a directory holding
# go.mod or *.go files within two levels. A read-only question about Go code
# gets no note: it starts no edit router. A pasted panic or goroutine trace is
# the exception, since its diagnosis is go-troubleshooting work. Text an IDE
# host adds to the prompt (<ide_opened_file>, <ide_selection>) is removed
# before any of these tests.
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
# go-code-routing.sh --hints: the edit table plus advisory import/signature
# and stub-contract signals; these extra hints do not widen the edit gate.
# On 2026-09-13 the gate blocked five edits in three sessions to name owners
# one at a time; a session that has the list before its first edit loads them
# without a block. Test files stay out of the scan.
#
# The idiom card is a section of go-style-core's SKILL.md, so the
# go-style-core load carries it. It used to be a reference this note named by
# path for a whole Read the gate required, since no skill wording made
# Sonnet 5 medium read it (0/24 on 2026-09-18).
set -u

# The plugin root: skills/ sits beside hooks/.
plugin_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# claude_note — the Claude note stays the router line. Another host has no
# Skill tool: only Claude Code sets CLAUDE_PROJECT_DIR, and Codex sets PLUGIN_ROOT.
claude_note() {
    [[ -n "${CLAUDE_PROJECT_DIR:-}" && -z "${PLUGIN_ROOT:-}" ]]
}

# read_alt <skill> — empty on Claude; otherwise the absolute SKILL.md, so the
# file is read whole. It is inserted before the parenthesis, so an empty
# result leaves the router line unchanged.
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

# One parse: session id and the skill to name, or nothing. The prompt text
# never reaches the shell; only the verdict is printed.
parsed="$(printf '%s' "$input" | "$helper" prompt-parse)" || exit 0
[[ -n "$parsed" ]] || exit 0
# A read loop, not mapfile: macOS ships bash 3.2, which has no mapfile.
files=()
{
    read -r session; read -r skill; read -r kind; read -r mode; read -r newcode; read -r speed
    while IFS= read -r f; do [[ -n "$f" ]] && files+=("$f"); done
} <<< "$parsed"
[[ -n "$skill" ]] || exit 0

state="${CLAUDE_PLUGIN_DATA:-${TMPDIR:-/tmp}/golang-skills-hooks}/routing/${session:-default}"
has() { # has <file> <skill>
    [[ -f "$1" ]] && grep -qx -- "$2" "$1"
}
# A load before the last compaction no longer counts (compacted holds the line
# count of loaded then; go-code-routing.sh reads it the same way): a slash
# command after it inserts the router again, and that insert is recorded.
since=0
[[ -f "$state/compacted" ]] && read -r since < "$state/compacted"
[[ "$since" =~ ^[0-9]+$ ]] || since=0
[[ -f "$state/loaded" ]] && tail -n "+$((since + 1))" "$state/loaded" | grep -qx -- "$skill" && exit 0
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
    ns="$("$helper" plugin-name "$manifest")" || ns=""
fi
q() { printf '%s%s' "${ns:+$ns:}" "$1"; }
owners=""
if [[ -n "${files[*]-}" ]]; then
    owners="$(bash "$(dirname "${BASH_SOURCE[0]}")/go-code-routing.sh" --hints "${files[@]}" 2>/dev/null)" || owners=""
fi
# A review writes nothing: the note names go-code-review and sets the boundary
# "before the first finding"; the review picks the owners behind its findings
# itself. An owner list in this note pulled Opus 5.5 medium up to 9 skills and
# +37% session cost (2026-09-29, n=6); without it the cost did not change.
# go-security is the one exception, for code that serves or calls HTTP or that
# the hints already tie to it: in the 2026-10-19 review corpus run (Sonnet 5.5
# medium, gatekeeper, vault, partner, n=3 per arm) go-code-review's own step 2
# loaded it in 0 of 18 sessions, on fixtures whose keys are mostly its defects.
if [[ "$skill" == "go-code-review" && "$mode" != "slash" ]]; then
    if [[ "$mode" == "mention" ]]; then
        printf 'golang-skills: this prompt names the `%s` skill; naming a skill does not load it.\n' "$(q "$skill")"
    else
        printf 'golang-skills: this prompt looks like %s.\n' "$kind"
    fi
    security=""
    if [[ " $owners " == *" go-http "* || " $owners " == *" go-security "* ]]; then
        security=" and the \`$(q go-security)\` skill (Skill tool, name \`$(q go-security)\`$(read_alt go-security)), since this code sits on a trust boundary"
    fi
    printf 'Before the first finding, load the `%s` skill (Skill tool, name `%s`%s)%s; %s.\n' "$(q "$skill")" "$(q "$skill")" "$(read_alt "$skill")" "$security" "$what"
    [[ "$mode" == "mention" ]] || printf 'If the task is not Go work, ignore this note.\n'
    exit 0
fi
# A diagnosis edits nothing: one skill, named before the answer.
if [[ "$skill" == "go-troubleshooting" ]]; then
    printf 'golang-skills: this prompt looks like %s.\n' "$kind"
    printf 'Before answering, load the `%s` skill (Skill tool, name `%s`%s); it traces the fault back to where the bad value came from before any fix.\n' "$(q "$skill")" "$(q "$skill")" "$(read_alt "$skill")"
    exit 0
fi
# The router line names go-style-core the way it names the router: "load the
# X skill (Skill tool, name X)". Opus 5.5 low acts on that line and treats the
# next as optional. On 2026-09-30/10-01 (implement, -shell go):
#   - go-style-core on the next line, the card Read gone: loaded with the
#     router in 1 of 14 sessions, gate blocks 2.71 per session;
#   - two skills in the router line without that phrasing, or one list after
#     the classification line: 3 of 14 and 7 of 12 sessions explored through
#     Bash first and edited before any load;
#   - each skill on the next line with its own (Skill tool, name ...):
#     go-style-core first-message in 10 of 21;
#   - both skills in the router line, this form: 21 of 21 with go-style-core
#     and go-testing in the first message, 0.43 blocks per session, cost at
#     v1.27.1's level (n=3).
# A slash command has already inserted the router, so its line starts at
# go-style-core.
tt="\`$(q go-testing)\` (Skill tool, name \`$(q go-testing)\`$(read_alt go-testing))"
if [[ "$mode" == "slash" ]]; then
    line="Load \`$(q go-style-core)\` (Skill tool, name \`$(q go-style-core)\`$(read_alt go-style-core))"
    sep="; "
    owners_lead=", and the owners its code points at: "
else
    line=""
    sep="; "
    owners_lead="; load the owners its code points at: "
fi
if [[ -n "$owners" ]]; then
    list=""
    for o in $owners; do list+="\`$(q "$o")\` (Skill tool, name \`$(q "$o")\`$(read_alt "$o")), "; done
    line+="${owners_lead}${list%, }"
    sep="; "
elif [[ "$skill" == "go-code" && "$mode" != "slash" ]]; then
    # No Go file points at the owners. Haiku 5.5 then loaded only the skills
    # this note named, so the note asks for them by description. Plugin eval,
    # 2026-10-12, without -> with this line: Haiku go-generics 1/5 -> 5/5,
    # go-security 0/5 -> 5/5, but go-error-handling, named nowhere here,
    # 1/5 -> 0/5; Opus 5.5 unchanged within n=3. abrun implement on the three
    # fixtures without hints (Opus, medium, n=3): golden 9/9 both, skills per
    # run 4.33 -> 4.67, $/run +3%.
    line+="${sep}and the owners the task touches, picked by their skill descriptions (\`$(q go-generics)\` for type parameters, \`$(q go-database)\` for SQL, \`$(q go-security)\` when user input reaches a shell, file, or URL)"
fi
if [[ "$speed" == "speed" && " $owners " != *" go-performance "* ]]; then
    line+="${sep}\`$(q go-performance)\` (Skill tool, name \`$(q go-performance)\`$(read_alt go-performance)) too, since the request asks for speed"
    sep="; "
fi
if [[ "$skill" == "go-code" && "$newcode" == "new" ]]; then
    line+="${sep}$tt too, since new code starts with its contract test"
elif [[ "$skill" == "go-code-refactor" ]]; then
    line+="${sep}$tt too, since a refactor often edits tests"
else
    line+="${sep}$tt if you write or edit a test"
fi
line+='. All of them in one message, before the first edit.'
line+=' Wait for successful Skill results; do not put Edit, Write or MultiEdit in the skill-loading tool message.'
if [[ "$skill" != go-code-review ]]; then
    line+=" If Bash is available, also load \`$(q go-linting)\` (Skill tool, name \`$(q go-linting)\`$(read_alt go-linting)) in that message; after the final edit run the receipt verifier before other verification commands."
fi
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
printf 'Before the first edit, load the `%s` skill (Skill tool, name `%s`%s) and the `%s` skill (Skill tool, name `%s`%s)%s\n' \
    "$(q "$skill")" "$(q "$skill")" "$(read_alt "$skill")" "$(q go-style-core)" "$(q go-style-core)" "$(read_alt go-style-core)" "$line"
printf '`%s` %s.\n' "$(q "$skill")" "${what#it }"
[[ "$mode" == "mention" ]] || printf 'If the task is not Go work, ignore this note.\n'
exit 0
