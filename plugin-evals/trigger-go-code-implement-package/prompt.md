---
description: 'Names only the package and the bodies; the router description must fire without the words new, function, or stub'
tags: [trigger]
runs: 3
max_turns: 10
allowed_tools: [Read, Glob, Grep, Skill]
---

Implement the Go package in ./feed. Every exported declaration is already there with its documentation; write the bodies so the package does what the documentation says. Do not change the exported signatures.
