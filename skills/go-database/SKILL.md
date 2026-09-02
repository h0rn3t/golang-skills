---
name: go-database
description: Use when writing or reviewing Go code that talks to a SQL database — database/sql or pgx queries, transactions, repositories, connection pools, migrations, or an ORM such as gorm. Also use when a handler is slow because of its queries, when a query sits inside a loop, or when mapping rows to structs, even if the user never says "database". Does not cover the HTTP handler around the query (see go-http) or context placement (see go-context).
---

# Go Database Access

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). Everything below is
> `database/sql`; `sql.Null[T]` requires Go 1.22+.

## Resource Routing

- `references/SQL-PATTERNS.md` - Read when writing the rows loop, a transaction helper, nullable columns, batch lookups, keyset pagination, or pool settings — the full code for each rule below.

## Stdlib First

> **Normative**: `database/sql` plus a driver covers most services. An ORM is
> rung four of the dependency ladder in [go-packages](../go-packages/SKILL.md).

```
What does the repository already use?
├─ Nothing yet        → database/sql (+ pgx stdlib driver for Postgres)
├─ Needs typed, generated queries → sqlc over database/sql
├─ Postgres-only, hot path        → pgx native (batches, COPY, typed scans)
└─ An ORM already     → keep it; WithContext on every call, Select the columns,
                        Preload/Joins instead of a query per row
```

---

## Every Query Carries a Context

> **Normative**: `QueryContext`, `QueryRowContext`, `ExecContext`, `BeginTx`.
> The ctx-less forms are unbounded and `noctx` in the lint gate flags them.

The request context bounds the query to the client's patience; a background
job gets its own `context.WithTimeout`. [go-context](../go-context/SKILL.md)
owns placement.

---

## `*sql.DB` Is a Pool

Open once in `run()`, `PingContext` to fail fast, inject it into repositories.
Never open per request, never store it in a global.

Configure the pool — the defaults are unlimited open connections and no
lifetime:

| Setting | Why |
|---|---|
| `SetMaxOpenConns` | Below the server's `max_connections` minus headroom for other clients |
| `SetMaxIdleConns` | Same as max open, or connections churn under load |
| `SetConnMaxLifetime` | Rotate through load balancers and credential changes |
| `SetConnMaxIdleTime` | Release idle connections back to the server |

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
- `sql.ErrNoRows` is matched with `errors.Is` and mapped to the domain
  sentinel (`ErrNotFound`) at the repository boundary. It never crosses into a
  handler — [go-error-handling](../go-error-handling/SKILL.md) owns wrapping.

---

## Transactions

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return fmt.Errorf("begin: %w", err)
}
defer tx.Rollback() // no-op after Commit; the ErrTxDone it returns is expected

if _, err := tx.ExecContext(ctx, debit, from, amount); err != nil {
    return fmt.Errorf("debit %s: %w", from, err)
}
if _, err := tx.ExecContext(ctx, credit, to, amount); err != nil {
    return fmt.Errorf("credit %s: %w", to, err)
}
if err := tx.Commit(); err != nil {
    return fmt.Errorf("commit transfer: %w", err)
}
return nil
```

- `defer tx.Rollback()` right after `BeginTx`, `Commit` last, and check the
  `Commit` error — it is where serialization failures surface.
- Everything inside uses `tx`, never `db`: a `db` call inside a transaction
  takes a second connection and deadlocks the pool at its limit.
- Keep transactions short: no network calls, no user waits, no logging that
  blocks. Lock order is part of the contract — same order everywhere.
- A `withTx(ctx, db, func(tx *sql.Tx) error)` helper removes the boilerplate;
  it is in `references/SQL-PATTERNS.md`.

---

## Parameters, Never Concatenation

> **Normative**: Every value goes through a placeholder (`$1`, `?`). A value
> formatted into SQL with `fmt.Sprintf` or `+` is an injection, full stop.

Identifiers that vary — a sort column, a table suffix — come from an allow-list
switch in Go, never from input. `gosec` (G201/G202) in the lint gate flags
string-built queries.

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

- Prefer `NOT NULL` with a default in the schema. When a column is nullable,
  scan into `sql.Null[T]` (Go 1.22+) or a pointer; scanning `NULL` into a
  `string` is a runtime error.
- Migrations are versioned, forward-only files shipped with the binary via
  `//go:embed` ([go-packages](../go-packages/SKILL.md)) and applied by a
  migration step at deploy — never from `init()`. Each one states its rollback
  or says it has none.

---

## Testing

Integration tests run against a real database (a container or a CI service),
not a mocked driver — a mock proves the code calls the mock. Unit-test only the
row-to-struct mapping. [go-testing](../go-testing/SKILL.md) owns the
integration harness in
[`go-testing/references/INTEGRATION.md`](../go-testing/references/INTEGRATION.md).

> **Validation**: `golangci-lint run` with `rowserrcheck`, `sqlclosecheck`,
> `noctx`, and `gosec` from the [go-linting](../go-linting/SKILL.md) baseline,
> then `go test -race ./...` with the integration tag. Report a skipped
> integration run as skipped.

---

## Quick Reference

| Do | Don't |
|----|-------|
| `QueryContext(ctx, ...)` | `Query(...)` |
| `defer rows.Close()` + `rows.Err()` after the loop | Trust an empty loop |
| `errors.Is(err, sql.ErrNoRows)` → `ErrNotFound` | Leak `sql.ErrNoRows` to handlers |
| `defer tx.Rollback()`, check `Commit` | Rollback only on the error path |
| `WHERE id = ANY($1)` | A query per element |
| `$1` placeholders | `fmt.Sprintf` into SQL |
| Pool limits set in `run()` | Default unlimited pool |

---

## Related Skills

- **Context**: See [go-context](../go-context/SKILL.md) for timeouts on queries and what may outlive the request
- **Errors**: See [go-error-handling](../go-error-handling/SKILL.md) for wrapping driver errors and mapping `sql.ErrNoRows` to a sentinel
- **HTTP**: See [go-http](../go-http/SKILL.md) for the handler that calls the repository and maps its errors to status codes
- **Testing**: See [go-testing](../go-testing/SKILL.md) for integration tests against a real database
- **Packages**: See [go-packages](../go-packages/SKILL.md) before adding an ORM or a second driver, and for `//go:embed` migrations
- **Performance**: See [go-performance](../go-performance/SKILL.md) for measuring a slow query before rewriting it
