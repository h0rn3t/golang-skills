---
type: llm
weight: 3
---

PASS if the answer satisfies every item:

- Uses maps.Copy(dst, updates) followed by maps.DeleteFunc with score < minimum
- Preserves dst identity and unrelated entries, with no clone-and-rebind implementation
- Keeps distinct updates unchanged and recognizes that filtering an aliased source changes the same map
- Handles nil/empty updates within the stated destination precondition without speculative allocation or error APIs
- Includes cases for overwrite-before-filter ordering and visible mutations through aliases without claiming execution

FAIL if any item is missing or contradicted.
