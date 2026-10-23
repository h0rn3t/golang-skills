// Command helper is the parsing and receipt half of the golang-skills hooks.
// The scripts in hooks/ keep the policy and the messages; they call this
// program, built on first use by hooks/go-hook-helper.sh, for what a shell
// does badly: JSON payloads, the routing heuristics, the shell-command parser
// of the verification gate, and the check receipts of the edit hook. It uses
// the standard library only, so the build needs nothing but the Go toolchain
// the hooks already require.
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: helper <command> [args...]")
		os.Exit(2)
	}
	args := os.Args[2:]
	switch os.Args[1] {
	case "plugin-name":
		// plugin-name <plugin.json>: the plugin's name, or an empty line.
		if len(args) == 1 {
			fmt.Println(pluginName(args[0]))
		}
	case "context-json":
		// context-json <event>: stdin as hookSpecificOutput.additionalContext.
		if len(args) == 1 {
			text, _ := io.ReadAll(os.Stdin)
			fmt.Println(contextJSON(args[0], string(text)))
		}
	case "emit-json":
		// emit-json stop|notice <text>: a PreToolUse answer for exit 0.
		if len(args) == 2 {
			fmt.Println(gateJSON(args[0], args[1]))
		}
	case "ladder-parse":
		ladderParse()
	case "subagent-parse":
		subagentParse()
	case "prompt-parse":
		promptParse()
	case "gate-parse":
		gateParse()
	case "hints":
		fmt.Println(strings.Join(fileHints(args), " "))
	case "edit-file":
		editFile()
	case "session-key":
		sessionKey()
	case "verification":
		os.Exit(verification())
	case "receipt":
		os.Exit(receiptMain(args))
	default:
		fmt.Fprintln(os.Stderr, "helper: unknown command "+os.Args[1])
		os.Exit(2)
	}
}
