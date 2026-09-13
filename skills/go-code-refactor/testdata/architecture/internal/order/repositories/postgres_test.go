package repositories

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"example.com/shop/internal/order/models"
)

// recorder is a database/sql/driver double: it records Begin, Exec, Commit and
// Rollback and fails the Exec whose query contains failOn. It proves the
// control flow of CreateWithAudit, not PostgreSQL isolation.
type recorder struct {
	ops    []string
	failOn string
}

func (r *recorder) Open(string) (driver.Conn, error) { return &conn{r: r}, nil }

type conn struct{ r *recorder }

func (c *conn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("not used") }
func (c *conn) Close() error                        { return nil }
func (c *conn) Begin() (driver.Tx, error) {
	c.r.ops = append(c.r.ops, "begin")
	return &tx{r: c.r}, nil
}

func (c *conn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	c.r.ops = append(c.r.ops, "exec")
	if c.r.failOn != "" && strings.Contains(query, c.r.failOn) {
		return nil, errors.New("boom")
	}
	return driver.RowsAffected(1), nil
}

type tx struct{ r *recorder }

func (t *tx) Commit() error   { t.r.ops = append(t.r.ops, "commit"); return nil }
func (t *tx) Rollback() error { t.r.ops = append(t.r.ops, "rollback"); return nil }

var registered int

func open(t *testing.T, failOn string) (*OrderStore, *recorder) {
	t.Helper()
	rec := &recorder{failOn: failOn}
	registered++
	name := "recorder" + string(rune('a'+registered))
	sql.Register(name, rec)
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewOrderStore(db), rec
}

var sample = models.Order{ID: "o-1", Customer: "c-1", Lines: []models.Line{{SKU: "a", Quantity: 2, PriceCents: 250}}}

func TestCreateWithAuditCommitsBothWrites(t *testing.T) {
	store, rec := open(t, "")
	if err := store.CreateWithAudit(t.Context(), sample, "placed"); err != nil {
		t.Fatalf("CreateWithAudit: %v", err)
	}
	if got := strings.Join(rec.ops, " "); got != "begin exec exec commit" {
		t.Fatalf("ops = %q, want begin exec exec commit", got)
	}
}

func TestCreateWithAuditRollsBackWhenTheAuditWriteFails(t *testing.T) {
	store, rec := open(t, "order_audit")
	err := store.CreateWithAudit(t.Context(), sample, "placed")
	if err == nil || !strings.Contains(err.Error(), "insert audit") {
		t.Fatalf("error = %v, want the audit insert failure", err)
	}
	if got := strings.Join(rec.ops, " "); got != "begin exec exec rollback" {
		t.Fatalf("ops = %q, want begin exec exec rollback: the order insert must not survive", got)
	}
}
