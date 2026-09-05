> Sources: https://pkg.go.dev/database/sql; https://go.dev/doc/database; https://pkg.go.dev/github.com/jackc/pgx/v5
> Authority: advisory
> Minimum Go: 1.22 for `sql.Null[T]`
> Last verified: 2026-09-01 against go1.27.0

# SQL Patterns

The full code behind each rule in `SKILL.md`. Copy the shape, not the names.

## Opening the pool once

```go
func openDB(ctx context.Context, dsn string) (*sql.DB, error) {
    db, err := sql.Open("pgx", dsn) // does not connect yet
    if err != nil {
        return nil, fmt.Errorf("open db: %w", err)
    }
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(30 * time.Minute)
    db.SetConnMaxIdleTime(5 * time.Minute)

    pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    if err := db.PingContext(pingCtx); err != nil {
        db.Close()
        return nil, fmt.Errorf("ping db: %w", err)
    }
    return db, nil
}
```

`sql.Open` validates the DSN and returns; the first real connection happens on
`PingContext`. Close the pool from `run()` with `defer db.Close()`.

## The repository boundary

```go
var ErrNotFound = errors.New("not found")

type OrderRepo struct{ db *sql.DB }

func (r *OrderRepo) ByID(ctx context.Context, id int64) (Order, error) {
    const q = `SELECT id, customer_id, total_cents FROM orders WHERE id = $1`
    var o Order
    err := r.db.QueryRowContext(ctx, q, id).Scan(&o.ID, &o.CustomerID, &o.TotalCents)
    if errors.Is(err, sql.ErrNoRows) {
        return Order{}, fmt.Errorf("order %d: %w", id, ErrNotFound)
    }
    if err != nil {
        return Order{}, fmt.Errorf("order %d: %w", id, err)
    }
    return o, nil
}
```

`QueryRowContext(...).Scan` closes the row for you. Wrap `ErrNotFound` with
`%w` so the handler's `errors.Is(err, ErrNotFound)` still matches.

## Transaction helper

```go
func withTx(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin: %w", err)
    }
    if err := fn(tx); err != nil {
        if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
            return errors.Join(err, fmt.Errorf("rollback: %w", rbErr))
        }
        return err
    }
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit: %w", err)
    }
    return nil
}

func (r *AccountRepo) Transfer(ctx context.Context, from, to int64, cents int64) error {
    return withTx(ctx, r.db, func(tx *sql.Tx) error {
        // Lock in a fixed order so two transfers cannot deadlock each other.
        lo, hi := min(from, to), max(from, to)
        if _, err := tx.ExecContext(ctx, `SELECT id FROM accounts WHERE id IN ($1, $2) ORDER BY id FOR UPDATE`, lo, hi); err != nil {
            return fmt.Errorf("lock accounts: %w", err)
        }
        if _, err := tx.ExecContext(ctx, `UPDATE accounts SET balance_cents = balance_cents - $1 WHERE id = $2`, cents, from); err != nil {
            return fmt.Errorf("debit %d: %w", from, err)
        }
        if _, err := tx.ExecContext(ctx, `UPDATE accounts SET balance_cents = balance_cents + $1 WHERE id = $2`, cents, to); err != nil {
            return fmt.Errorf("credit %d: %w", to, err)
        }
        return nil
    })
}
```

Isolation level goes in `sql.TxOptions{Isolation: sql.LevelSerializable}` when
the invariant spans rows; then retry on the serialization-failure SQLSTATE
(`40001` on Postgres) by rerunning the complete transaction, including the
reads that informed its writes, a bounded number of times.
[go-resilience](../../go-resilience/SKILL.md) owns the time/attempt budget and
replay safety; keep nontransactional external side effects out of that retry.

## Batch lookup instead of a loop

```go
// Bad: one round trip per order
for i := range orders {
    orders[i].Customer, err = repo.CustomerByID(ctx, orders[i].CustomerID)
}

// Good: one round trip, map in Go (pgx driver: pass the slice, ANY expands it)
ids := make([]int64, 0, len(orders))
for _, o := range orders {
    ids = append(ids, o.CustomerID)
}
rows, err := db.QueryContext(ctx, `SELECT id, name FROM customers WHERE id = ANY($1)`, ids)
// ... rows loop into map[int64]Customer, then assign
```

With a driver that does not expand slices, build the placeholder list
(`$1, $2, ...`) from `len(ids)` — still parameters, never values in the string.

## Keyset pagination

```go
const q = `SELECT id, created_at, total_cents FROM orders
           WHERE customer_id = $1 AND id > $2
           ORDER BY id LIMIT $3`
rows, err := db.QueryContext(ctx, q, customerID, afterID, pageSize+1)
```

Fetch `pageSize+1` to know whether a next page exists without a `COUNT`.
`OFFSET n` reads and discards `n` rows on every page; it is fine for admin
screens and wrong for anything a client can walk.

## Nullable columns

```go
type Customer struct {
    ID    int64
    Email sql.Null[string] // Go 1.22+: Valid=false on NULL
    Phone *string          // pointer form; nil on NULL
}

err := row.Scan(&c.ID, &c.Email, &c.Phone)
if c.Email.Valid {
    send(c.Email.V)
}
```

Fix the schema first: `NOT NULL DEFAULT ''` removes the case entirely. Use
`sql.Null[T]` when the column is genuinely optional and the zero value is a
legal stored value that must stay distinguishable from absent.

## ORM rules when the repository already has one

| Rule | gorm form |
|---|---|
| Context on every call | `db.WithContext(ctx).Find(...)` — no bare `db.Find` |
| Select the columns | `.Select("id", "name")` or a projection struct; the default is `SELECT *` |
| No query per row | `.Preload("Customer")` or `.Joins("Customer")` instead of touching the association in a loop |
| Bounded results | `.Limit(n)` plus a keyset `Where`, not `.Offset` |
| Transactions | `db.Transaction(func(tx *gorm.DB) error { ... })`; every call inside uses `tx` |
| Not-found | `errors.Is(err, gorm.ErrRecordNotFound)` → the domain sentinel at the repository |

Generated SQL is still SQL: read it with the logger at Debug during
development, and `EXPLAIN` the ones that show up in the slow-query log.
