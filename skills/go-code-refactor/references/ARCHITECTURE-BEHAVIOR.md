# Behavior Contracts Across Layers

> Sources: [Executing transactions](https://go.dev/doc/database/execute-transactions); [Working with Errors in Go](https://go.dev/blog/go1.13-errors); `database/sql` package documentation
> Authority: project policy for the rules; the `database/sql` mechanics are normative Go documentation
> Minimum Go: 1.27 baseline
> Last verified: 2026-09-13

Companion to [ARCHITECTURE.md](ARCHITECTURE.md): the three contracts a package
move must carry across unchanged — what commits together, which errors a caller
can match, and who decides that an actor may act. The compilable
[fixture](../testdata/architecture/README.md) shows each one in code:
`CreateWithAudit` on one `*sql.Tx` with a driver-double rollback test,
`sql.ErrNoRows` mapped to the module's `ErrNotFound`, and an actor check in the
service rather than the handler.


### Preserve atomicity before changing packages

The service owns **which changes must succeed together**; repositories own the storage mechanics. A refactor must not silently change isolation, lock behavior, commit/rollback boundaries, or externally visible partial-failure outcomes.

For a small atomic action, a single repository method such as `CreateWithAudit` can remain the simplest seam. Its contract specifies atomicity without exposing `*sql.Tx`, `*gorm.DB`, or a driver transaction to the service. If several repository operations must share a transaction, use a narrowly scoped transaction runner/callback and transaction-bound stores, or the project's existing equivalent. Introduce a universal Unit of Work only for a demonstrated need, not because packages were separated.

Inside a database transaction, execute the atomic statements through the same transaction; calling a normal pool/DB method can run outside it. Preserve cancellation and return commit failures rather than claiming success. The companion includes a complete `database/sql` example and a driver-double rollback test. That test proves the example's control flow, not production PostgreSQL isolation or failure recovery. See [Executing transactions](https://go.dev/doc/database/execute-transactions).

Required behavior checks for an affected use case: the success path commits, failure of a later write rolls back earlier changes, errors/cancellation propagate according to the existing contract, and retry/idempotency behavior is not accidentally changed. Add real-database integration tests when the change depends on actual isolation, constraints, or locking.

Avoid holding a DB transaction open across remote calls by accident. A DB write followed by publishing a message is not automatically atomic. When event loss or duplicates matter, choose an explicit delivery/idempotency design, such as a transactional outbox with an idempotent consumer where appropriate. Do not prescribe events or an outbox for every CRUD module. Distributed workflow mechanics belong in the relevant specialized reference.

### Errors are contracts

Define expected errors at the consuming business boundary: missing object, conflict, invalid transition, and unavailable dependency as needed. A repository translates expected storage errors to that contract. A service applies use-case semantics. A handler maps results to the existing HTTP/gRPC/message contract.

Wrapping an implementation error with `%w` exposes it to `errors.Is`/`errors.As`; it can leak a driver dependency even if the caller never imports the repository package. Preserve cancellation identity where required, retain operational diagnostics through the project's error/logging conventions, and do not return raw database details to a client. See [Working with Errors in Go](https://go.dev/blog/go1.13-errors).

### Authentication is not use-case authorization

The handler authenticates or consumes middleware identity and checks the request's structural validity. The service enforces the actor's permission to perform the business operation and the operation's invariants. Transport-level access gates can remain as defense in depth, but a CLI, worker, or second handler must not bypass essential use-case authorization. Preserve the existing identity/tenant propagation contract; do not add an unrelated authorization framework during a package move.
