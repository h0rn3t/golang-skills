---
name: go-database
description: Use when writing or reviewing Go SQL queries, transactions, repositories, pools, migrations, or row mapping with database/sql, pgx, or an ORM; also for slow queries and queries in loops.
---

# Go Database Access

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). Go examples use
> `database/sql`; `sql.Null[T]` requires Go 1.22+.

## Resource Routing

- `references/SQL-PATTERNS.md` - Read when writing the rows loop, a transaction helper, nullable columns, batch lookups, keyset pagination, or pool settings — the full code for each rule below.
- [references/POSTGRESQL.md](references/POSTGRESQL.md) - Read for PostgreSQL schema design, constraints, indexes, upserts, or migrations on live tables. Check the target PostgreSQL version; these rules do not apply to other databases.

## Stdlib First

> **Normative**: `database/sql` plus a driver covers most services. An ORM is
> a new module, the last rung of the dependency ladder in
> [go-packages](../go-packages/SKILL.md).

```
What does the repository already use?
├─ Nothing yet        → database/sql (+ pgx stdlib driver for Postgres)
├─ Needs typed, generated queries → sqlc over database/sql
├─ Postgres-only, hot path        → pgx native (batches, COPY, typed scans)
└─ An ORM already     → keep it; WithContext on every call, Select the columns,
                        Preload/Joins instead of a query per row
```

pgx native means one `*pgxpool.Pool` per process from `pgxpool.New(ctx, dsn)`
— a `*pgx.Conn` is not safe for concurrent use — and rows read with
`pgx.CollectRows(rows, pgx.RowToStructByName[T])`, which closes them and
returns their error.

---

## Every Query Carries a Context

> **Normative**: `QueryContext`, `QueryRowContext`, `ExecContext`, `BeginTx`.
> The ctx-less forms are unbounded and `noctx` in the lint gate flags them.

---

## `*sql.DB` Is a Pool

Open once in `run()`, `PingContext` to fail fast, inject it into repositories.
Never open per request, never store it in a global.

Configure the pool — the defaults are unlimited open connections and no
lifetime:

| Setting | Why |
|---|---|
| `SetMaxOpenConns` | Per process: the server's `max_connections` minus admin headroom, divided by every process that shares it at peak — replicas during a rolling deploy, workers, migration jobs |
| `SetMaxIdleConns` | Same as max open, or connections churn under load |
| `SetConnMaxLifetime` | Rotate through load balancers and credential changes |
| `SetConnMaxIdleTime` | Release idle connections back to the server |

An exhausted pool shows `db.Stats().WaitCount` climbing and every endpoint slow
at once. Fix it where connections are held: rows left open, or a transaction
kept across a network call. Raising `SetMaxOpenConns` past that budget only
moves the queue into the database. Instances that scale without a fixed count
share the budget through a pooler such as PgBouncer.

---

## Rows: Close and Check `Err`

```go
rows, err := db.QueryContext(ctx, q, customerID)
if err != nil {
    return nil, fmt.Errorf("list orders: %w", err)
}
defer rows.Close()

var orders []Order
for rows.Next() {
    var o Order
    if err := rows.Scan(&o.ID, &o.Total); err != nil {
        return nil, fmt.Errorf("scan order: %w", err)
    }
    orders = append(orders, o)
}
if err := rows.Err(); err != nil {
    return nil, fmt.Errorf("list orders: %w", err)
}
return orders, nil
```

- `rows.Err()` after the loop is not optional — a dropped connection ends the
  loop silently and looks like an empty result. `rowserrcheck` flags it.
- `defer rows.Close()` immediately after the error check; `sqlclosecheck`
  flags a missing close.
- For lookups, match `sql.ErrNoRows` with `errors.Is` and map it to the domain
  sentinel (`ErrNotFound`) at the repository boundary. For writes with
  `RETURNING`, interpret a missing row according to the write contract. Keep
  `sql.ErrNoRows` out of handlers — [go-error-handling](../go-error-handling/SKILL.md) owns wrapping.

---

## Transactions

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return fmt.Errorf("begin: %w", err)
}
defer tx.Rollback() // no-op after Commit; the ErrTxDone it returns is expected

// Both rows locked in id order, whichever way the money moves: opposite
// transfers that lock from-then-to deadlock each other.
const lock = `SELECT id FROM accounts WHERE id IN ($1, $2) ORDER BY id FOR UPDATE`
if _, err := tx.ExecContext(ctx, lock, from, to); err != nil {
    return fmt.Errorf("lock accounts: %w", err)
}

// RETURNING turns a missing account into sql.ErrNoRows; a bare ExecContext
// succeeds on zero rows and would commit the debit alone.
const move = `UPDATE accounts SET balance_cents = balance_cents + $1 WHERE id = $2 RETURNING id`
var id int64
err = tx.QueryRowContext(ctx, move, -amount, from).Scan(&id)
if errors.Is(err, sql.ErrNoRows) {
    return fmt.Errorf("debit %d: %w", from, ErrNotFound)
}
if err != nil {
    return fmt.Errorf("debit %d: %w", from, err)
}
err = tx.QueryRowContext(ctx, move, amount, to).Scan(&id)
if errors.Is(err, sql.ErrNoRows) {
    return fmt.Errorf("credit %d: %w", to, ErrNotFound)
}
if err != nil {
    return fmt.Errorf("credit %d: %w", to, err)
}
if err := tx.Commit(); err != nil {
    return fmt.Errorf("commit transfer: %w", err)
}
return nil
```

- Check the `Commit` error — it is where serialization failures surface.
- Everything inside uses `tx`, never `db`: a `db` call inside a transaction
  takes a second connection and deadlocks the pool at its limit.
- Keep transactions short: no network calls, no user waits, no logging that
  blocks. Lock order is part of the contract — same order everywhere.

---

## Parameters, Never Concatenation

> **Normative**: Every value goes through a placeholder (`$1`, `?`). A value
> formatted into SQL with `fmt.Sprintf` or `+` is an injection, full stop.

Identifiers that vary — a sort column, a table suffix — come from an allow-list
switch in Go, never from input. `gosec` in the lint gate flags string-built
queries; [go-security](../go-security/SKILL.md) owns the threat model.

---

## Queries in Loops

A query inside `for _, o := range orders` is the most common performance bug in
a service. Fix it with one query — `WHERE id = ANY($1)` (pgx) or an expanded
`IN (...)`, or a `JOIN` — then map in Go. Unbounded lists take a `LIMIT` and
keyset pagination (`WHERE id > $1 ORDER BY id LIMIT $2`); deep `OFFSET` scans
everything it skips.

Measure before restructuring: `EXPLAIN (ANALYZE, BUFFERS)` on the real
database decides whether an index or a rewrite is the fix.
[go-performance](../go-performance/SKILL.md) owns the benchmark discipline.

---

## Nullable Columns and Migrations

- Use `NOT NULL` when the domain requires a value, and defaults only when
  they have domain meaning. Preserve absent versus empty/zero values. Scan
  nullable columns into `sql.Null[T]` (Go 1.22+) or a pointer; scanning `NULL`
  into a `string` is a runtime error.
- Migrations are versioned, forward-only files shipped with the binary via
  `//go:embed` ([go-packages](../go-packages/SKILL.md)) and applied by a
  migration step at deploy — never from `init()`. Each one states its rollback
  or says it has none.

---

## Testing

Integration tests run against a real database (a container or a CI service),
not a mocked driver — a mock proves the code calls the mock. Unit-test only the
row-to-struct mapping. [go-testing](../go-testing/SKILL.md) owns the
integration harness — build tag, DSN from the environment, skip when unset —
in [`go-testing/references/INTEGRATION.md`](../go-testing/references/INTEGRATION.md#real-databases).

> **Validation**: `golangci-lint run` with `rowserrcheck`, `sqlclosecheck`,
> `noctx`, and `gosec` from the [go-linting](../go-linting/SKILL.md) baseline,
> then `go test -tags integration -race ./...`. Report a skipped
> integration run as skipped.

---

## Related Skills

- [go-resilience](../go-resilience/SKILL.md): idempotency, replay budgets, atomicity across remote side effects.
- [go-context](../go-context/SKILL.md): query timeouts and what may outlive the request.
- [go-http](../go-http/SKILL.md): the handler that calls the repository and maps its errors to status codes.
