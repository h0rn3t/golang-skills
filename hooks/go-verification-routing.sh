#!/usr/bin/env bash
# PreToolUse on Bash: complete the gate-owner load and receipt attempt before
# repeating runtime checks for the last Go edit. Stop: hold back a report whose
# pass (hook) or speedup claims lack evidence. The gate itself is
# hooks/helper/verification.go. Never executes the command.
set -u
export GOLANG_SKILLS_VERIFICATION_ROOT="${CLAUDE_PLUGIN_ROOT:-$(cd "$(dirname "$0")/.." && pwd)}"
# shellcheck source=go-hook-helper.sh
source "$(dirname "${BASH_SOURCE[0]}")/go-hook-helper.sh"
helper="$(hook_helper)" || exit 0
exec "$helper" verification
