#!/usr/bin/env bash
# The hooks' Go helper (hooks/helper), built on first use. Sourced, it defines
# hook_helper, which prints the path of the built binary or fails when Go is
# missing or the build fails; run, it executes the helper with its arguments:
#
#   bash go-hook-helper.sh <command> [args...]
#
# The binary is cached outside the plugin, keyed by a checksum of the helper
# sources, so an update rebuilds it and an unchanged copy reuses it. It is
# built outside the project, so the project's go.work and GOFLAGS cannot break
# a standard-library-only build. The first build compiles part of the
# standard library and can take seconds; the SessionStart hooks pay it.

hook_helper() {
    command -v go >/dev/null 2>&1 || return 1
    local src cache stamp bin
    src="$(cd "$(dirname "${BASH_SOURCE[0]}")/helper" 2>/dev/null && pwd)" || return 1
    cache="${XDG_CACHE_HOME:-${HOME:-${TMPDIR:-/tmp}}/.cache}/golang-skills"
    if ! { mkdir -p "$cache" && [[ -w "$cache" ]]; } 2>/dev/null; then
        cache="${TMPDIR:-/tmp}/golang-skills-cache"
        mkdir -p "$cache" 2>/dev/null || return 1
    fi
    stamp="$(cat "$src"/go.mod "$src"/*.go 2>/dev/null | cksum | awk '{print $1 "-" $2}')"
    bin="$cache/hook-helper-$stamp"
    if [[ ! -x "$bin" ]]; then
        # A temporary name and a rename: two hooks starting at once never run
        # a half-written binary.
        (cd "$src" && GOWORK=off GOFLAGS= GOTOOLCHAIN=local go build -o "$bin.$$" .) >/dev/null 2>&1 &&
            mv -f "$bin.$$" "$bin" || { rm -f "$bin.$$"; return 1; }
    fi
    printf '%s\n' "$bin"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
    bin="$(hook_helper)" || { printf 'golang-skills: the hook helper does not build (Go missing or build failed)\n' >&2; exit 2; }
    GOLANG_SKILLS_HOOK_ROOT="${GOLANG_SKILLS_HOOK_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}" exec "$bin" "$@"
fi
