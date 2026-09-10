package evals_test

import "testing"

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

// Перевірка реакції Go-коду на EOF від драйвера; SQL виконується лише в інтеграційних тестах.
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
`)
}
