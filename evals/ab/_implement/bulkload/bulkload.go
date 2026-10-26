// Package bulkload writes the nightly event export into PostgreSQL.
package bulkload

import (
	"context"
	"database/sql"
	"time"
)

// Event is one row of the export.
type Event struct {
	ID     int64
	Tenant string
	Kind   string
	At     time.Time
	Amount int64
	Note   string
}

// Insert writes events into the PostgreSQL table
// events (id, tenant, kind, at, amount, note), one row per event, in the
// order given. Placeholders are PostgreSQL's $1, $2, and so on.
//
// A night's export holds about 50,000 events. The previous job wrote one
// statement per event: the import took forty minutes and held a connection
// for all of it, so Insert sends the export in a few dozen statements at
// most. Either every event is written or none is: on any failure the rows
// already sent are rolled back and the error is returned. No events is a
// successful insert that sends nothing.
func Insert(ctx context.Context, db *sql.DB, events []Event) error {
	panic("not implemented")
}
