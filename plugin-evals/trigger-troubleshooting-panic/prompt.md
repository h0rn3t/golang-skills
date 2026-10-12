---
description: 'Pasted stack trace; reading a panic'
tags: [trigger]
runs: 3
max_turns: 10
allowed_tools: [Read, Glob, Grep, Skill]
---

panic: runtime error: invalid memory address or nil pointer dereference — goroutine 42 [running]: main.(*Server).handle(0xc000010000, {0x0, 0x0}) — what does this mean?
