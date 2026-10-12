---
description: 'maps.Copy followed by maps.DeleteFunc implements the in-place contract without replacement maps or hand-written traversal.'
tags: [quality, cost]
runs: 3
max_turns: 12
allowed_tools: [Read, Glob, Grep, Skill]
---

Use go-data-structures for a Go 1.27 snippet-only implementation. Write func ApplyScores(dst, updates map[string]int, minimum int). Merge updates into dst, replacing matching values but keeping unrelated entries, then remove every entry below minimum from the resulting dst. Existing aliases of dst must see the mutation. Distinct updates must remain unchanged; dst and updates may also be the same map. The caller initializes dst whenever updates is nonempty; nil/nil and empty updates are valid. Show focused cases for overwrite-before-filter order and aliasing. Do not run commands.
