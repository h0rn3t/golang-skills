> Authority: advisory
> Last verified: 2026-09-08 against PostgreSQL 17 documentation; version gates noted below.
> Topic selection informed by [pg-aiguide](https://github.com/timescale/pg-aiguide/tree/acf42427fed507b7bfe98c4039fbacf0c4a69b65/skills); independently written and checked against the official sources linked below.

# PostgreSQL Schemas and Migrations

Use for PostgreSQL tasks only. Establish the server version, existing schema,
query patterns, and migration runner's transaction behavior before choosing DDL.
Preserve the repository's driver and migration tooling.

## Constraints express domain rules

- `CHECK (amount > 0)` accepts `NULL`; add `NOT NULL` only if absence is invalid.
- `UNIQUE` normally permits multiple nulls. Choose `NULLS NOT DISTINCT` (PG15+)
  only when the domain treats nulls as equal for uniqueness.
- A primary key or unique constraint already creates an index. PostgreSQL does
  not automatically index referencing FK columns: inspect existing indexes and
  parent delete/update and join workloads before adding one.
- Select FK delete/update actions explicitly from ownership semantics; cascading
  deletion is not a general default.

Source: [Constraints](https://www.postgresql.org/docs/17/ddl-constraints.html).

## Indexes and conflict handling

Choose indexes for actual filters, joins, and ordering; account for write cost.
Match a partial unique index's predicate in the conflict target, for example:

```sql
CREATE UNIQUE INDEX accounts_live_email_uidx
    ON accounts (email) WHERE deleted_at IS NULL;

INSERT INTO accounts (email) VALUES ($1)
ON CONFLICT (email) WHERE deleted_at IS NULL DO NOTHING;
```

This assumes `deleted_at` defaults to NULL and other required fields have
defaults. The index permits reusing a deleted account's email; null emails
remain distinct. Use concurrent creation for an existing busy table as below.
Choose `DO NOTHING` versus `DO UPDATE` from the operation's contract. With
`RETURNING`, skipped writes return no row; account for that in Go instead of
blindly applying the usual not-found mapping.

Source: [INSERT / ON CONFLICT](https://www.postgresql.org/docs/17/sql-insert.html).

## Evolve live tables in stages

- Metadata-only DDL still takes locks. Bound lock waits with `lock_timeout`;
  keep DDL transactions short and assess long-running blockers.
- For CHECK/FK on existing data, consider `ADD CONSTRAINT ... NOT VALID`, commit,
  then `VALIDATE CONSTRAINT` separately. New writes are checked immediately;
  `NOT VALID` skips the initial scan, not lock acquisition.
- For `SET NOT NULL` on PG12+, a validated `CHECK (col IS NOT NULL)` can avoid
  the scan. Keep that check until `SET NOT NULL` completes.

Source: [ALTER TABLE](https://www.postgresql.org/docs/17/sql-altertable.html).

- On busy ordinary tables, `CREATE INDEX CONCURRENTLY` permits writes but must
  run outside a transaction block. Use the runner's per-migration setting or a
  separately tracked step. Partitioned tables require a different build plan.
- After a failed concurrent build, inspect `pg_index.indisvalid`; an invalid
  unique index can still enforce uniqueness. Resolve the cause and plan cleanup
  or rebuild before retrying. `IF NOT EXISTS` does not establish index validity.

Source: [CREATE INDEX](https://www.postgresql.org/docs/17/sql-createindex.html).

For rolling deployments, add compatible schema first, coordinate writers,
backfill if needed, switch readers, and remove old fields only after old code
has retired. Backfill in resumable batches with short commits; avoid overwriting
new writes. Verify duplicates, nulls, and application compatibility before
enforcing the final constraints. Test on representative data and record recovery
steps for partially completed migrations.

## Defaults are not historical data

PG11+ can add a column with a nonvolatile default without rewriting existing
rows. `now()` is STABLE; `clock_timestamp()` is VOLATILE. Do not prescribe a
backfill merely because the default calls `now()`. Existing rows get the value
evaluated for the migration, not their historical creation times. `SET DEFAULT`
on an existing column does not fill its nulls.

Sources: [ALTER TABLE](https://www.postgresql.org/docs/17/sql-altertable.html),
[function volatility](https://www.postgresql.org/docs/17/xfunc-volatility.html),
[time functions](https://www.postgresql.org/docs/17/functions-datetime.html).
