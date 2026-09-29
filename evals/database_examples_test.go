package evals_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestTransactionExampleFinalizes(t *testing.T) {
	code := exampleBlock(t, "skills/go-database/references/SQL-PATTERNS.md", "## Transaction helper")
	runExampleTest(t, `package example
import ("context"; "database/sql"; "database/sql/driver"; "errors"; "fmt"; "io"; "testing")
type AccountRepo struct { db *sql.DB }
var ErrNotFound = errors.New("not found") // the sentinel from "The repository boundary"
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
  {name: "rollback error keeps the callback error", callback: func(*sql.Tx) error { return errCallback }, rollbackErr: errDriver, wantErr: errCallback, rollbacks: 1},
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
   if got := db.Stats().InUse; got != 0 { t.Errorf("connections in use after withTx = %d, want 0", got) }
   if conn.commits != tt.commits || conn.rollbacks != tt.rollbacks {
    t.Errorf("commit/rollback calls = %d/%d, want %d/%d", conn.commits, conn.rollbacks, tt.commits, tt.rollbacks)
   }
  })
 }
}

func TestMissingAccountRollsBack(t *testing.T) {
 for _, missing := range []int64{0, 1, 2} {
  conn := &testConn{missingID: missing}
  db := sql.OpenDB(conn)
  repo := &AccountRepo{db: db}
  err := repo.Transfer(t.Context(), 1, 2, 100)
  wantErr := missing != 0
  if errors.Is(err, ErrNotFound) != wantErr {t.Errorf("Transfer(missing=%d): error=%v, want ErrNotFound=%t", missing, err, wantErr)}
  if errors.Is(err, sql.ErrNoRows) {t.Errorf("Transfer(missing=%d): sql.ErrNoRows leaked past the repository: %v", missing, err)}
  if wantErr && (conn.commits != 0 || conn.rollbacks != 1) {t.Errorf("missing=%d: commits=%d rollbacks=%d, want 0/1", missing, conn.commits, conn.rollbacks)}
  if !wantErr && (conn.commits != 1 || conn.rollbacks != 0) {t.Errorf("success: commits=%d rollbacks=%d, want 1/0", conn.commits, conn.rollbacks)}
  if got:=db.Stats().InUse;got!=0 {t.Errorf("Transfer: connections in use=%d, want 0",got)}
  if err:=db.Close();err!=nil {t.Fatal(err)}
 }
}
`+accountDriver)
}

// The go-database SKILL.md transfer is the form a reader copies: a missing
// account on either side rolls back and surfaces ErrNotFound, never
// sql.ErrNoRows, and nothing is committed.
func TestDatabaseExampleTransactionRollsBackMissingAccount(t *testing.T) {
	code := exampleBlock(t, "skills/go-database/SKILL.md", "## Transactions")
	runExampleTest(t, `package example
import ("context"; "database/sql"; "database/sql/driver"; "errors"; "fmt"; "io"; "testing")
var ErrNotFound = errors.New("not found")
func transfer(ctx context.Context, db *sql.DB, from, to, amount int64) error {
`+code+`
}
func TestMissingAccount(t *testing.T) {
 for _, missing := range []int64{0, 1, 2} {
  conn := &testConn{missingID: missing}
  db := sql.OpenDB(conn)
  err := transfer(t.Context(), db, 1, 2, 100)
  wantErr := missing != 0
  if errors.Is(err, ErrNotFound) != wantErr {t.Errorf("transfer(missing=%d): error=%v, want ErrNotFound=%t", missing, err, wantErr)}
  if errors.Is(err, sql.ErrNoRows) {t.Errorf("transfer(missing=%d): sql.ErrNoRows leaked: %v", missing, err)}
  if wantErr && (conn.commits != 0 || conn.rollbacks != 1) {t.Errorf("missing=%d: commits=%d rollbacks=%d, want 0/1", missing, conn.commits, conn.rollbacks)}
  if !wantErr && (conn.commits != 1 || conn.rollbacks != 0) {t.Errorf("success: commits=%d rollbacks=%d, want 1/0", conn.commits, conn.rollbacks)}
  if got:=db.Stats().InUse;got!=0 {t.Errorf("transfer: connections in use=%d, want 0",got)}
  if err:=db.Close();err!=nil {t.Fatal(err)}
 }
}
`+accountDriver)
}

// The real-database helper in go-testing's INTEGRATION.md skips without a DSN
// and closes the pool it opened; a fake "pgx" driver stands in for the one the
// module already imports. The file carries the integration build tag, so the
// module is tested with it.
func TestTestingExampleRealDatabaseHelper(t *testing.T) {
	code := exampleBlock(t, "skills/go-testing/references/INTEGRATION.md", "## Real Databases")
	dir := t.TempDir()
	for name, content := range map[string]string{
		"go.mod":     "module example\n\ngo 1.27\n",
		"db_test.go": code,
		"use_test.go": `//go:build integration

package store_test
import ("database/sql"; "database/sql/driver"; "errors"; "testing")
var closes int
type fakeDriver struct{}
type fakeConn struct{}
func (fakeDriver) Open(string) (driver.Conn, error) { return fakeConn{}, nil }
func (fakeConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unused") }
func (fakeConn) Close() error { closes++; return nil }
func (fakeConn) Begin() (driver.Tx, error) { return nil, errors.New("unused") }
func init() { sql.Register("pgx", fakeDriver{}) }
func TestOpenTestDB(t *testing.T) {
 t.Run("unset", func(t *testing.T) {
  t.Setenv("TEST_DATABASE_URL", "")
  defer func() { if !t.Skipped() { t.Error("openTestDB did not skip without TEST_DATABASE_URL") } }()
  openTestDB(t)
 })
 t.Run("set", func(t *testing.T) {
  t.Setenv("TEST_DATABASE_URL", "fake")
  if openTestDB(t) == nil { t.Fatal("openTestDB returned nil") }
 })
 if closes != 1 { t.Errorf("connections closed after the subtests = %d, want 1", closes) }
}
`,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"vet", "-tags", "integration", "./..."}, {"test", "-count=1", "-race", "-tags", "integration", "./..."}} {
		cmd := exec.CommandContext(t.Context(), "go", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("go %s: %v\n%s", args[0], err, out)
		}
	}
}

// accountDriver is a database/sql driver whose UPDATE ... RETURNING finds every
// account except missingID; both transaction examples run against it.
const accountDriver = `
// Checks how Go code reacts to EOF from the driver; SQL runs only in integration tests.
type testConn struct {
 beginErr, commitErr, rollbackErr error
 commits, rollbacks int
 missingID int64
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
func (c *testConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
 return driver.RowsAffected(1), nil
}
func (c *testConn) QueryContext(_ context.Context, _ string, args []driver.NamedValue) (driver.Rows, error) {
 id := args[1].Value.(int64)
 return &accountRows{id: id, available: id != c.missingID}, nil
}
type accountRows struct { id int64; available bool }
func (*accountRows) Columns() []string {return []string{"id"}}
func (*accountRows) Close() error {return nil}
func (r *accountRows) Next(dest []driver.Value) error {
 if !r.available {return io.EOF}
 r.available=false
 dest[0]=r.id
 return nil
}
func (c *testConn) Commit() error { c.commits++; return c.commitErr }
func (c *testConn) Rollback() error { c.rollbacks++; return c.rollbackErr }
`
