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
#                                     read of a sibling SKILL.md.
#   PreToolUse  (Edit|Write|MultiEdit) before a .go edit, require a router
#                                     (go-code, go-code-refactor, or
#                                     go-code-review; with none, the one
#                                     go-prompt-routing.sh named, else
#                                     go-code), go-style-core, and the owner
#                                     skills the edited text points at. Exit 2
#                                     blocks the edit and names what is
#                                     missing by exact Skill names
#                                     (golang-skills:go-code in a plugin).
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
# The idiom card used to be a gate item of its own, a whole Read of
# go-style-core/references/CURRENT-GO.md, because no wording made Sonnet 5
# medium read it (0/24 sessions on 2026-09-18 under three wordings). It is
# now a section of go-style-core's SKILL.md, so the go-style-core load carries
# it on every host, with this gate or without it.
#
# The owner hints below are heuristics: regular expressions over the edited
# text that recognize the decision-bearing forms of thirteen owners (error
# wrapping, goroutines, context creation or a Context field, SQL, slog,
# exec/templates/crypto, a deferred closure or recover, type parameters, a
# non-empty interface, main, retries, HTTP, tests).
# Routine syntax — a plain fmt.Errorf("%v"), r.Context(), a ctx parameter
# passed on, defer f.Close() / mu.Unlock() / cancel(), an http.StatusOK in a
# comment, make([]T, n), append — must not fire. The bare defer and
# context.Context hints of 1.7.0 fired on nearly every body: on 2026-09-30
# (abrun low, both 5.5 models) go-defensive was among the owners named in 16
# of 73 implement gate blocks. Collections have no hint on
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
#
# A subagent shares the session_id, so its records live under
# routing/<session>/agent-<agent_id>: its context holds only the skills it
# loaded itself, and its loads must not lift the main thread's requirement.
# A fork starts from its parent's context, skills included, and keeps the
# parent's record. prompted is the main thread's (UserPromptSubmit fires
# there only) and names the entry router for every agent of the session.
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
    skill_ns="$("$helper" plugin-name "$manifest")" || skill_ns=""
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
    "$helper" emit-json "$1" "$2"
}

# The owner table lives in the helper (hooks/helper/hints.go): the gate runs
# it over an edit; --hints runs it over whole files for go-prompt-routing.sh,
# adding advisory import, signature, and stub-contract signals that never
# widen the edit gate.
# shellcheck source=go-hook-helper.sh
source "$(dirname "${BASH_SOURCE[0]}")/go-hook-helper.sh"
if ! helper="$(hook_helper)"; then
    [[ "${1:-}" == "--hints" ]] && exit 0
    # The gate cannot read its input without the helper. Say so once per
    # session instead of passing every edit in silence; the note lives in the
    # session's own state, so the next session hears it again.
    input="$(cat)"
    [[ "$input" =~ \"hook_event_name\"[[:space:]]*:[[:space:]]*\"PreToolUse\" ]] || exit 0
    session="default"
    [[ "$input" =~ \"session_id\"[[:space:]]*:[[:space:]]*\"([A-Za-z0-9_-]+)\" ]] && session="${BASH_REMATCH[1]}"
    note="${CLAUDE_PLUGIN_DATA:-${TMPDIR:-/tmp}/golang-skills-hooks}/routing/$session/helper-unavailable"
    if [[ ! -e "$note" ]] && mkdir -p "$(dirname "$note")" && touch "$note"; then
        printf '{"systemMessage": "golang-skills: the routing gate is off for this session: its Go helper does not build (go missing from PATH, or the build failed)."}\n'
    fi
    exit 0
fi

if [[ "${1:-}" == "--hints" ]]; then
    shift
    "$helper" hints "$@" | while read -r -a owners; do
        # A skill missing from this copy is not named: the gate does not
        # require it.
        out=()
        for o in "${owners[@]}"; do available "$o" && out+=("$o"); done
        printf '%s\n' "${out[*]}"
    done
    exit 0
fi

input="$(cat)"

# One parse: event, session, agent, tool, skill name, file path, the owner
# skills inferred from the edited content, and the edit key. Fields are
# newline-separated; the agent and the hint list may be empty.
parsed="$(printf '%s' "$input" | "$helper" gate-parse)" || exit 0
[[ -n "$parsed" ]] || exit 0
{ read -r event; read -r session; read -r agent; read -r tool; read -r skill; read -r path; read -r hints; read -r key; } <<< "$parsed"

routing="${CLAUDE_PLUGIN_DATA:-${TMPDIR:-/tmp}/golang-skills-hooks}/routing"
session_state="$routing/${session:-default}"
state="$session_state${agent:+/agent-$agent}"
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
    # Sessions end without notice; drop state older than two days, with the
    # edit hook's receipt runs and the verification state beside it.
    find "$routing" "${routing%/routing}/checks" "${routing%/routing}/verification" \
        -mindepth 1 -maxdepth 1 -type d -mtime +2 -exec rm -rf {} + 2>/dev/null
    exit 0
    ;;
PreToolUse)
    # apply_patch is Codex: the Edit|Write matcher catches it, but tool_name
    # stays apply_patch and the patch sits in command. A path in file_path is
    # not turned into a Skill requirement here.
    [[ "$tool" == "apply_patch" ]] && exit 0
    case "$path" in *.go) ;; *) exit 0 ;; esac
    [[ "${GOLANG_SKILLS_ROUTING_GATE:-}" == off ]] && exit 0
    # The gate runs on Claude Code only: it exports CLAUDE_PROJECT_DIR and sends
    # Skill|Read. Codex sets PLUGIN_ROOT (and CLAUDE_PLUGIN_ROOT). A host with
    # Edit but without those events would otherwise block every edit, and the
    # third attempt would stop the session.
    [[ -n "${CLAUDE_PROJECT_DIR:-}" && -z "${PLUGIN_ROOT:-}" ]] || exit 0
    routers=""
    for r in go-code go-code-refactor go-code-review; do
        has "$loaded" "$r" && routers+="$r "
    done
    # Without a router the gate requires the entry skill: the one
    # go-prompt-routing.sh named last (the prompted file), else go-code.
    entry=""
    if [[ -z "$routers" ]]; then
        entry="$(grep -x -E 'go-code|go-code-refactor|go-code-review' "$session_state/prompted" 2>/dev/null | tail -n 1)"
        entry="${entry:-go-code}"
    fi
    # Only loaded lifts a requirement: reminded logs reminders, not loads.
    missing="" absent=""
    for want in $entry go-style-core $hints; do
        has "$loaded" "$want" && continue
        if available "$want"; then missing+="$want "; else absent+="$want "; fi
    done
    fresh_absent=""
    for a in $absent; do has "$state/absent" "$a" || fresh_absent+="$a "; done
    [[ -n "$missing" || -n "$fresh_absent" ]] || exit 0
    mkdir -p "$state"
    [[ -z "$fresh_absent" ]] || printf '%s\n' $fresh_absent >> "$state/absent"
    init_ns
    absent_note=""
    [[ -z "$fresh_absent" ]] || absent_note="Not installed in this plugin copy, so not required: $(names $fresh_absent) (no SKILL.md under $root/skills). Reinstall the plugin to restore them."
    if [[ -z "$missing" ]]; then
        # Nothing to load; the user hears about a missing skill once, and the
        # edit passes.
        emit_json notice "golang-skills routing gate: $absent_note"
        exit 0
    fi

    # An attempt at the same edit with no progress (a new load) since the
    # last block is counted; progress resets the counter.
    sig="$( (wc -l < "$loaded") 2>/dev/null | tr -d ' ')"
    sig="${sig:-0}"
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

    list="$(names $missing)"
    if (( attempt >= 3 )); then
        emit_json stop "golang-skills routing gate: the same Go edit was blocked twice and no skill load was recorded in between (still missing: $list). Stopping instead of blocking it a third time: the skills look unavailable in this session, because the Skill call fails or neither Skill nor Read reaches $root/skills. Check the install with \`claude plugin list\`, or set GOLANG_SKILLS_ROUTING_GATE=off to switch the gate off."
        exit 0
    fi
    {
        if (( attempt == 2 )); then
            printf 'golang-skills routing gate: this edit was blocked before, and no load has been\n'
            printf 'recorded since. Still missing: %s.\n' "$list"
            printf 'A reminder is not a load: a retry without the loads is blocked again, and a\n'
            printf 'third stalled retry of this edit stops the session.\n'
        elif [[ -n "$entry" && " $missing " == *" $entry "* ]]; then
            printf 'golang-skills routing gate: this session loaded no router skill, and a .go edit\n'
            printf 'needs one first. Missing: %s.\n' "$list"
            if [[ "$entry" == go-code ]] && available go-code-refactor; then
                printf 'For a behavior-preserving refactor, load %s instead of %s.\n' "$(names go-code-refactor)" "$(names go-code)"
            fi
        else
            printf 'golang-skills routing gate: this session loaded %s but not: %s.\n' "${routers% }" "$list"
        fi
        printf 'This edit was not applied and the file is unchanged. Load them (one Skill call\n'
        printf 'per name, all in one message), then retry the same edit against the unchanged file.\n'
        printf 'If the Skill tool is missing or answers "Unknown skill", Read\n'
        printf '%s/skills/<name>/SKILL.md whole for each name instead; the gate counts that Read.\n' "$root"
        [[ -z "$absent_note" ]] || printf '%s\n' "$absent_note"
        if (( attempt == 1 )) && [[ -n "$fresh" ]]; then
            cat <<'EOF'
The gate reads the edited text and recognizes some owners only: tests, error
wrapping, goroutines, context creation, SQL, slog, exec and templates,
deferred closures, type parameters, interfaces, main, retries, HTTP. The routing table in
go-code/SKILL.md decides, including the owners the gate cannot see, and the
gate's silence is not a passing result.
EOF
        fi
    } >&2
    exit 2
    ;;
esac
exit 0
