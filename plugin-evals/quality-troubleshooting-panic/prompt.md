---
description: 'A pasted nil-dereference panic is traced to where the nil came from, with the printed argument words treated as a lead, not a fix at the faulting line.'
tags: [quality]
runs: 3
max_turns: 12
allowed_tools: [Read, Glob, Grep, Skill]
---

panic: runtime error: invalid memory address or nil pointer dereference — goroutine 42 [running]: main.(*Server).handle(0xc000010000, {0x0, 0x0}) — what does this mean?
