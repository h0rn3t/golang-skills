# Data-Flow Tracing for Wrong Results

> Sources: local `../SKILL.md`; `../../go-http/SKILL.md`; `../../go-database/SKILL.md`; [Go diagnostics: tracing](https://go.dev/doc/diagnostics#tracing)
> Authority: project policy for evidence handling; advisory for tracing workflow
> Last verified: 2026-09-05

Follow one affected value through the actual execution path. Locate the first
boundary where a documented invariant or expected transformation fails; a
suspicious function name or final stack frame is only a starting point.

## Find the Path

Start with a route, error string, log field, job type, or symbol from the
artifact. Use `rg` for narrow discovery, then read definitions and call sites.
Use language-server definition/references/call hierarchy when available; search
and source inspection are a valid fallback. Resolve interface implementations,
router registration, and build/feature branches for the affected deployment.

Typical paths (skip layers the system does not have):

```text
HTTP request → router → auth/tenant middleware → decode/defaults
             → domain/service → repository/SQL or external call
             → scan/mapping/aggregation → response serialization

Event → consumer registration → decode/version → tenant/idempotency check
      → transaction/state change → acknowledgement or retry
```

Do not expand into the whole repository. At each relevant boundary record:

| Boundary | What to compare |
|---|---|
| Router/middleware | Matched method/path, authenticated actor, verified tenant, role, context propagation |
| Decode/defaults | Raw vs parsed values, missing vs empty vs zero, units, timezone, normalization |
| Domain branch | Effective flags, status transitions, filters, contract preconditions |
| SQL/external call | Actual statement/request, bound parameters, scope predicates, transaction/replica, response |
| Scan/mapping | Column order, nullable values, conversions, row count, errors including `Rows.Err()` |
| Aggregation | Map key identity, overwrite/dedup rules, joins, order, count, pagination |
| Serialization | DTO fields, JSON tags, omission, nil vs empty, precision, error/status mapping |

Carry the same correlation ID and identity through the chain. A row for
`tenant=B,id=7` is not evidence about `tenant=A,id=7`. Compare actual bound
parameters with the query; a correct-looking template does not establish that
its inputs or selected database were correct. Use scoped read-only queries on
available fixtures; do not alter real rows to make a reproduction convenient.

## Find the First Divergence

1. State the invariant: for example, every returned asset belongs to the
   authenticated tenant, or a report includes one item per `(tenant, local_id)`.
2. Locate the last boundary with correct observed data and the first with bad
   output. If distant, inspect an intermediate boundary to halve the search.
3. Inspect the transformation and its inputs. Check the upstream producer if
   the callee correctly follows its contract on an already-invalid value.
4. Compare the same boundaries for a nearby working case. Choose a fixture that
   retains the suspected condition: colliding IDs, NULL, a missing parameter,
   a time boundary, or the relevant transaction state.
5. Propose a discriminating check and, when fixing is authorized, put regression
   coverage at the responsible contract. Do not infer an executed result from
   reading the code.

### Example: Rows Disappear Without a SQL Error

Contract: the organization report includes both tenants. SQL and the post-Scan
capture contain `(A,7,10)` and `(B,7,20)`; the response contains only B's item.
The mapper uses `byID[row.LocalID] = item` before encoding its values.

The first loss is the assignment using `7` alone as identity: the second item
overwrites the first. A control with IDs 7 and 8 passes because no key collides.
This supports a mapping defect for the supplied case, not a `database/sql` Scan
failure. A compound `(tenant, local_id)` key or direct row output may satisfy the
contract; choose according to whether deduplication is required. Filtering away
one authorized tenant would violate this report's contract.

Overwrite follows input assignment order; later map iteration affects output
order, not which value survived an earlier key collision. A query without
`ORDER BY` need not return the same row on a repeated `LIMIT 1` request: test
tenant/content invariants rather than predicting a particular arbitrary row.

The regression retains both colliding rows and verifies content/count without
assuming map iteration order; include the distinct-ID control. In analysis-only
mode describe this check without creating tests or changing the mapper.

## Separate Similar-Looking Causes

- **Wrong scope:** find where tenant/actor identity is dropped or substituted;
  route SQL fixes to [go-database](../../go-database/SKILL.md) and authorization
  decisions to [go-security](../../go-security/SKILL.md). Do not bypass checks.
- **Invalid stored state:** distinguish a read-path bug from data violating a
  real invariant. Trace the writer or missing constraint; data repair is a
  separate action with its own scope, not an implicit debugging step.
- **Stale result:** establish that a cache/replica is on this request's path,
  then compare keys, tenant namespace, versions, and invalidation/replication
  timing. A ticket naming Redis is not evidence that Redis served the response.
- **Timing-dependent result:** static tracing locates shared ownership or
  ordering candidates; reproduce the relevant concurrent path and route runtime
  capture to [diagnostic tools](DIAGNOSTIC-TOOLS.md). Wrong output alone does not
  establish a data race, and a clean `-race` run does not prove correct ordering.

Stop expanding the path when evidence identifies the responsible mechanism or
a specific missing observation. Record the latter as the next check instead of
reporting a guessed file and line as a confirmed cause.
