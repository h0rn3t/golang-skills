---
description: 'A bounded single-document decoder validates EOF before any domain side effect and describes checks without claiming execution.'
tags: [quality, cost]
runs: 3
max_turns: 12
allowed_tools: [Read, Glob, Grep, Skill]
---

Use go-http for a snippet-only exercise. Write the JSON decoding and validation portion of a POST handler that accepts exactly one object {"name":string}, requires a nonempty name, and calls store.Create only after the whole request body is accepted. Limit the body to 1 MiB. Include checks for a second JSON value, trailing junk, allowed trailing whitespace, and an over-limit body. Do not run commands.
