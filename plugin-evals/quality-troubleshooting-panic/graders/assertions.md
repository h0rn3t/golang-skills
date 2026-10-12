---
type: llm
weight: 3
---

PASS if the answer satisfies every item:

- Identifies a nil pointer dereference inside (*Server).handle
- Says the faulting frame shows where the nil value was used, not where it came from, and directs tracing that value back to its origin (who built or passed it)
- Does not offer a nil check or recover() at the faulting line as the fix before the nil's origin is known
- Cautions that the printed argument words ({0x0, 0x0}) can be incomplete or stale, for example under optimization, and should be confirmed against the source rather than read as proof of which value was nil

FAIL if any item is missing or contradicted.
