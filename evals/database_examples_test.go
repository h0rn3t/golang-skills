package evals_test

import (
	"strings"
	"testing"
)

func TestTransactionExampleFinalizes(t *testing.T) {
	code, _, _ := strings.Cut(exampleBlock(t, "skills/go-database/references/SQL-PATTERNS.md", "## Transaction helper"), "\nfunc (r *AccountRepo)")
	runExampleTest(t, `package example
import ("context"; "database/sql"; "database/sql/driver"; "errors"; "fmt"; "testing")
`+code+`
var errCallback = errors.New("callback")
var errDriver = errors.New("driver")

func TestFinalization(t *testing.T) {
 for _, tt := range []struct {
  name string
  callback func(*sql.Tx) error
  beginErr, commitErr, rollbackErr error
  wantErr error
  wantPanic bool
  commits, rollbacks int
 }{
  {name: "commit", callback: func(*sql.Tx) error { return nil }, commits: 1},
  {name: "callback error", callback: func(*sql.Tx) error { return errCallback }, wantErr: errCallback, rollbacks: 1},
  {name: "panic", callback: func(*sql.Tx) error { panic("callback") }, wantPanic: true, rollbacks: 1},
  {name: "begin error", beginErr: errDriver, wantErr: errDriver},
  {name: "commit error", callback: func(*sql.Tx) error { return nil }, commitErr: errDriver, wantErr: errDriver, commits: 1},
  {name: "rollback error", callback: func(*sql.Tx) error { return errCallback }, rollbackErr: errDriver, wantErr: errCallback, rollbacks: 1},
 } {
  t.Run(tt.name, func(t *testing.T) {
   conn := &testConn{beginErr: tt.beginErr, commitErr: tt.commitErr, rollbackErr: tt.rollbackErr}
   db := sql.OpenDB(conn)
   defer db.Close()
   var gotErr error
   var panicked bool
   func() {
    defer func() { panicked = recover() != nil }()
    gotErr = withTx(t.Context(), db, tt.callback)
   }()
   if panicked != tt.wantPanic { t.Errorf("withTx panic = %t, want %t", panicked, tt.wantPanic) }
   if !errors.Is(gotErr, tt.wantErr) { t.Errorf("withTx error = %v, want %v", gotErr, tt.wantErr) }
   if tt.rollbackErr != nil && !errors.Is(gotErr, tt.rollbackErr) { t.Errorf("withTx error = %v, missing rollback error", gotErr) }
   if got := db.Stats().InUse; got != 0 { t.Errorf("connections in use after withTx = %d, want 0", got) }
   if conn.commits != tt.commits || conn.rollbacks != tt.rollbacks {
    t.Errorf("commit/rollback calls = %d/%d, want %d/%d", conn.commits, conn.rollbacks, tt.commits, tt.rollbacks)
   }
  })
 }
}

// The driver boundary exercises database/sql ownership without a live database.
type testConn struct {
 beginErr, commitErr, rollbackErr error
 commits, rollbacks int
}
func (c *testConn) Connect(context.Context) (driver.Conn, error) { return c, nil }
func (c *testConn) Driver() driver.Driver { return c }
func (c *testConn) Open(string) (driver.Conn, error) { return c, nil }
func (c *testConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unused") }
func (c *testConn) Close() error { return nil }
func (c *testConn) Begin() (driver.Tx, error) {
 if c.beginErr != nil { return nil, c.beginErr }
 return c, nil
}
func (c *testConn) Commit() error { c.commits++; return c.commitErr }
func (c *testConn) Rollback() error { c.rollbacks++; return c.rollbackErr }
`)
}
