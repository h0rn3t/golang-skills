# Architecture fixture: a two-module layered monolith

A compilable `go 1.27` module (`example.com/shop`) in shape C of
[ARCHITECTURE.md](../../references/ARCHITECTURE.md): `internal/order` and
`internal/billing` are layered business modules, `internal/app` is the
composition root, `internal/platform/clock` is the one approved shared
package, `cmd/server` is the entry point. `architecture.json` is the
checker's configuration with an empty `known` list.

What it shows, in code rather than prose:

- **Runtime flow is not import direction.** `order/services` declares the
  `Store` it consumes; `order/repositories` satisfies it structurally and is
  imported only by `internal/app`.
- **Named types must agree across modules.** `billing/services.Orders` returns
  billing's own `OrderSummary`; `internal/app.orderReader` maps
  `order.Summary` to it. Neither module imports the other's implementation.
- **Atomicity is the repository's contract, not the service's mechanics.**
  `repositories.CreateWithAudit` runs both inserts on one `*sql.Tx`, rolls back
  on the second failure, and returns a commit error.
  `postgres_test.go` proves that control flow with a `database/sql/driver`
  double — begin, exec, exec, rollback — and nothing about PostgreSQL.
- **Errors are contracts.** `sql.ErrNoRows` becomes `order.ErrNotFound` in the
  repository; the handler maps it to 404; `internal/app` maps it to
  `billing.ErrOrderUnknown`.
- **Authentication is not authorization.** The handler reads `X-Actor`; the
  service refuses an empty actor.

Run from this directory:

```bash
go vet ./... && go test ./...
bash ../../scripts/check-architecture.sh .        # 0 violations
```

`evals/architecture_test.go` in the repository compiles this fixture, runs the
checker against it, and against copies with one forbidden edge added.
