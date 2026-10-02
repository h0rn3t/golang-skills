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
    skill_ns="$(python3 -c 'import json, sys; print(json.load(open(sys.argv[1])).get("name") or "")' \
        "$manifest" 2>/dev/null)" || skill_ns=""
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
    python3 -c '
import json, sys
kind, text = sys.argv[1], sys.argv[2]
out = {"systemMessage": text}
if kind == "stop":
    out = {"continue": False, "stopReason": text,
           "hookSpecificOutput": {"hookEventName": "PreToolUse",
                                  "permissionDecision": "deny",
                                  "permissionDecisionReason": text}}
print(json.dumps(out))
' "$1" "$2"
}

# owner_hints_py is the one owner table, as python: hints(path, text) names
# the owner skills the decision-bearing forms in text point at. The gate runs
# it over an edit; --hints runs it over whole files for go-prompt-routing.sh,
# Prompt hints additionally resolve imports/signatures/stub contracts in
# prompt_hints; those advisory additions never widen the edit gate.
owner_hints_py='
import re
# A test file has one owner. Its body is test plumbing — a defer, an
# http.Request, an errors.Is on a want — not the decisions the content hints
# below recognize; running them over a _test.go named go-defensive for every
# contract test in the 2026-09-10 runs.
# Heuristics, one decision-bearing pattern per owner. go-code/SKILL.md owns
# the routing decision; keep each regex narrow enough that ordinary syntax
# (fmt.Errorf with %v, r.Context(), http.StatusOK, "<-" inside a string, an
# empty interface{}, a sha256 content hash) does not name an owner.
OWNER_PATTERNS = [
    ("go-http", r"\bnet/http\b|\bhttp\.(Handle|HandleFunc|Server\b|Client\b|NewServeMux|NewRequest|ResponseWriter|Error|ListenAndServe|Redirect)"),
    ("go-error-handling", r"fmt\.Errorf\([^)]*%w|\berrors\.(Is|As|AsType|Join|New)\("),
    ("go-concurrency", r"\bgo\s+func\b|\bgo\s+[A-Za-z_]\w*\(|\bmake\(chan\b|\bchan\s|\bsync\.(Mutex|RWMutex|WaitGroup|Once|Map)\b"),
    ("go-context", r"\bcontext\.(Background|TODO|With[A-Za-z]+|AfterFunc)\b|^\s+\w+\s+context\.Context\s*(//.*)?$"),
    ("go-database", r"database/sql|\bsql\.(Open|DB|Tx|Rows|Null)\b|\bpgx(pool)?\."),
    ("go-logging", r"\bslog\."),
    ("go-security", r"os/exec|html/template|text/template|\bcrypto/(?!(?:sha256|sha3|sha512)\b)|\bexec\.Command|\bhttp\.(SetCookie|Cookie)\b|\bfilepath\.Join\("),
    ("go-defensive", r"\bdefer\s+func\b|\brecover\(\)"),
    ("go-generics", r"\[[A-Z][A-Za-z0-9]*\s+(any|comparable|~|[A-Za-z]+\.[A-Za-z]+)\b"),
    ("go-interfaces", r"\binterface\s*\{\s*[^\s}]"),
    ("go-packages", r"^package main\b|\bfunc main\("),
    ("go-resilience", r"x/time/rate|\bbackoff\b|\bRetry-After\b"),
]
def hints(path, text):
    if path.endswith("_test.go"):
        return ["go-testing"]
    out = [owner for owner, pat in OWNER_PATTERNS if re.search(pat, text, re.M)]
    # Mirror the go-code HTTP and SQL rows: handlers and queries always load
    # go-error-handling too.
    if ("go-http" in out or "go-database" in out) and "go-error-handling" not in out:
        out.append("go-error-handling")
    return out

def prompt_hints(path, text):
    # Prompt-only lexical hints: do not widen PreToolUse decisions with stub
    # contracts or an import. Comments and literals are not executable code.
    if path.endswith("_test.go"):
        return ["go-testing"]
    tokens = list(re.finditer(r"//[^\n]*|/\*.*?\*/|\"(?:\\.|[^\"\\])*\"|`[^`]*`|\x27(?:\\.|[^\x27\\])*\x27", text, re.S))
    comments = []
    code = list(text)
    literal_code = list(text)
    literals = {}
    for token in tokens:
        value = token.group()
        is_comment = value.startswith(("//", "/*"))
        if is_comment:
            comments.append(value)
        else:
            literals[token.start()] = value
        for i in range(token.start(), token.end()):
            if code[i] != "\n":
                code[i] = " "
            if is_comment and literal_code[i] != "\n":
                literal_code[i] = " "
    code = "".join(code)
    lexical_code = code
    literal_code = "".join(literal_code)
    # Locate import spans in literal/comment-free code, then recover only
    # their string tokens. This handles grouped, aliased and raw imports.
    imports = {}
    dot = []
    extra_imports = []
    canonical = {"context": "context", "sync": "sync", "errors": "errors",
                 "log/slog": "slog", "net/http": "http", "database/sql": "sql",
                 "os/exec": "exec", "fmt": "fmt", "html/template": "template",
                 "text/template": "template"}
    for start in re.finditer(r"\bimport\b", code):
        tail = code[start.end():]
        grouped = re.match(r"\s*\(", tail)
        end = code.find(")", start.end()) + 1 if grouped else code.find("\n", start.end())
        if end <= start.end():
            end = len(code)
        for pos, value in literals.items():
            if not start.end() <= pos < end:
                continue
            if value[0] not in (chr(34), "`"):
                continue
            package = value[1:-1]
            prefix = re.split(r"[;\n(]", code[start.end():pos])[-1]
            alias_match = re.search(r"([A-Za-z_]\w*|\.)\s*$", prefix)
            alias = alias_match.group(1) if alias_match else package.rsplit("/", 1)[-1]
            if alias == ".":
                if package in canonical:
                    dot.append(canonical[package])
            elif alias != "_":
                extra_owner = ""
                if package.startswith("crypto/") and package not in ("crypto/sha256", "crypto/sha3", "crypto/sha512"):
                    extra_owner = "go-security"
                if package == "x/time/rate" or package == "golang.org/x/time/rate":
                    extra_owner = "go-resilience"
                imports[alias] = canonical.get(package, alias if extra_owner else "foreign")
                if extra_owner:
                    extra_imports.append((imports[alias], extra_owner))
    code = re.sub(r"\b([A-Za-z_]\w*)\s*\.",
                  lambda m: imports.get(m.group(1), m.group(1)) + ".", code)
    # Import paths themselves are blanked, so existing decision patterns
    # now match usage rather than an unused import or a prose example.
    out = hints(path, code)
    for owner, pattern in [
        ("go-http", r"\bhttp\.(?:Request|Response|Transport|RoundTripper|Handler)\b"),
        ("go-database", r"\bsql\.(?:Conn|Stmt|Result|NamedArg|Null\w*)\b"),
        ("go-security", r"\b(?:template\.\w+|exec\.Cmd)\b")]:
        if re.search(pattern, code):
            out.append(owner)
    for alias, owner in extra_imports:
        if re.search(r"\b" + re.escape(alias) + r"\s*\.\s*\w+", code):
            out.append(owner)
    if re.search(r"\bcontext\.Context\b", code) or ("context" in dot and re.search(r"(?<![\w.])Context\b", code)):
        out.append("go-context")
    dot_patterns = [("sync", "go-concurrency", r"(?:Mutex|RWMutex|WaitGroup|Once|Map)\b"),
                    ("errors", "go-error-handling", r"(?:Is|As|AsType|Join|New)\s*\("),
                    ("slog", "go-logging", r"(?:Logger|Handler|Attr|LogAttrs|Info|Error)\b"),
                    ("http", "go-http", r"(?:ResponseWriter|Request|Server|Client|NewServeMux)\b"),
                    ("sql", "go-database", r"(?:DB|Tx|Rows|Open)\b"),
                    ("exec", "go-security", r"(?:Command|CommandContext|Cmd)\b")]
    for package, owner, pattern in dot_patterns:
        if package in dot and re.search(r"(?<![\w.])" + pattern, code):
            out.append(owner)
    # A %w literal matters only as an argument of an actual Errorf call.
    fmt_names = [alias for alias, name in imports.items() if name == "fmt"]
    if "fmt" not in imports:
        fmt_names.append("fmt")
    for alias in fmt_names:
        for call in re.finditer(r"\b" + re.escape(alias) + r"\.Errorf\s*\(", lexical_code):
            for pos, value in literals.items():
                if pos >= call.end() and not lexical_code[call.end():pos].strip() and "%w" in value:
                    out.append("go-error-handling")
    stub = False
    for call in re.finditer(r"\bpanic\s*\(", lexical_code):
        for pos, value in literals.items():
            if pos >= call.end() and not lexical_code[call.end():pos].strip() and value in (chr(34) + "not implemented" + chr(34), "`not implemented`"):
                stub = True
    contract = " ".join(comments)
    if stub and re.search(r"\bbounded\b.{0,50}\bworkers?\b|\bat most\b.{0,70}\bin flight\b|\bowns\b.{0,40}\bgoroutines\b", contract, re.I):
        out.append("go-concurrency")
    if stub and re.search(r"\b(?:honor(?:s|ing)?|respect(?:s|ing)?|retr(?:ies|ied|ying)|retry(?!-))\b.{0,100}\b(?:Retry-After|backoff)\b", contract, re.I | re.S):
        out.append("go-resilience")
    if ("go-http" in out or "go-database" in out) and "go-error-handling" not in out:
        out.append("go-error-handling")
    return list(dict.fromkeys(out))
'

if [[ "${1:-}" == "--hints" ]]; then
    shift
    command -v python3 >/dev/null 2>&1 || exit 0
    python3 -c "$owner_hints_py"'
import sys
seen = []
for p in sys.argv[1:]:
    try:
        text = open(p, encoding="utf-8", errors="replace").read()
    except OSError:
        continue
    for h in prompt_hints(p, text):
        if h not in seen:
            seen.append(h)
print(" ".join(seen))
' "$@" | while read -r -a owners; do
        # A skill missing from this copy is not named: the gate does not
        # require it.
        out=()
        for o in "${owners[@]}"; do available "$o" && out+=("$o"); done
        printf '%s\n' "${out[*]}"
    done
    exit 0
fi

input="$(cat)"
command -v python3 >/dev/null 2>&1 || exit 0

# One parse: event, session, tool, skill name, file path, and the owner skills
# inferred from the edited content. Fields are newline-separated; the hint
# list is the last line and may be empty.
parsed="$(printf '%s' "$input" | python3 -c "$owner_hints_py"'
import hashlib, json, sys
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
owners = hints(path, text)
# Edit key: a retry of the same edit after a block has the same key, while
# parallel edits of different files in one message have different keys.
key = hashlib.sha1((path + "\0" + text).encode("utf-8", "replace")).hexdigest()[:16]
for v in (d.get("hook_event_name") or "", d.get("session_id") or "",
          d.get("tool_name") or "", skill, path, " ".join(owners), key):
    print(v.replace("\n", " "))
')" || exit 0
[[ -n "$parsed" ]] || exit 0
{ read -r event; read -r session; read -r tool; read -r skill; read -r path; read -r hints; read -r key; } <<< "$parsed"

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
    [[ "${GOLANG_SKILLS_ROUTING_GATE:-}" == off ]] && exit 0
    routers=""
    for r in go-code go-code-refactor go-code-review; do
        has "$loaded" "$r" && routers+="$r "
    done
    # Without a router the gate requires the entry skill: the one
    # go-prompt-routing.sh named last (the prompted file), else go-code.
    entry=""
    if [[ -z "$routers" ]]; then
        entry="$(grep -x -E 'go-code|go-code-refactor|go-code-review' "$state/prompted" 2>/dev/null | tail -n 1)"
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
