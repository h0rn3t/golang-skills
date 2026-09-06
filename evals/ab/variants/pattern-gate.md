## Patterns

> **Normative**: A design pattern — Strategy, Factory, Decorator, a registry, an
> interface introduced for substitution — is allowed on a refactor only when all
> three hold:

1. It removes duplication that exists in the code **today**, in three or more places.
2. The diff after it is shorter than the diff without it.
3. No call site reads worse.

Variants you expect to appear later are not an argument; that is rung 1 of the
restraint ladder. When a pattern passes, the report names the three duplicate
sites by `file:line` and the line delta it bought.
