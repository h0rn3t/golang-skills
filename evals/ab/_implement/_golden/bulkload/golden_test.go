package bulkload

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// goldenMaxParams is the PostgreSQL wire protocol's limit: the bind message
// counts its parameters in a 16-bit field.
const goldenMaxParams = 65535

var goldenErrDisk = errors.New("pq: could not extend file: No space left on device")

// goldenDB records what reaches the server: every statement with its
// arguments, and how each transaction ended.
type goldenDB struct {
	mu        sync.Mutex
	execs     []goldenExec
	commits   int
	rollbacks int
	outsideTx int
	failOn    int // fail the n-th statement, 1-based; 0 never
}

type goldenExec struct {
	query string
	args  []driver.Value
}

func (g *goldenDB) Connect(context.Context) (driver.Conn, error) { return &goldenConn{db: g}, nil }
func (g *goldenDB) Driver() driver.Driver                        { return goldenDriver{} }

type goldenDriver struct{}

func (goldenDriver) Open(string) (driver.Conn, error) { return nil, errors.New("use the connector") }

type goldenConn struct {
	db   *goldenDB
	inTx bool
}

func (c *goldenConn) Prepare(query string) (driver.Stmt, error) {
	return &goldenStmt{conn: c, query: query}, nil
}
func (c *goldenConn) Close() error { return nil }
func (c *goldenConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *goldenConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	c.inTx = true
	return goldenTx{c}, nil
}

func (c *goldenConn) ExecContext(_ context.Context, query string, named []driver.NamedValue) (driver.Result, error) {
	args := make([]driver.Value, len(named))
	for i, nv := range named {
		args[i] = nv.Value
	}
	return c.exec(query, args)
}

func (c *goldenConn) exec(query string, args []driver.Value) (driver.Result, error) {
	g := c.db
	g.mu.Lock()
	defer g.mu.Unlock()
	if len(args) > goldenMaxParams {
		return nil, fmt.Errorf("extended protocol limited to %d parameters", goldenMaxParams)
	}
	if !c.inTx {
		g.outsideTx++
	}
	g.execs = append(g.execs, goldenExec{query, args})
	if g.failOn == len(g.execs) {
		return nil, goldenErrDisk
	}
	return driver.RowsAffected(0), nil
}

type goldenStmt struct {
	conn  *goldenConn
	query string
}

func (s *goldenStmt) Close() error  { return nil }
func (s *goldenStmt) NumInput() int { return -1 }
func (s *goldenStmt) Exec(args []driver.Value) (driver.Result, error) {
	return s.conn.exec(s.query, args)
}
func (s *goldenStmt) Query([]driver.Value) (driver.Rows, error) {
	return nil, errors.New("bulkload sends no queries")
}

type goldenTx struct{ c *goldenConn }

func (t goldenTx) Commit() error {
	t.c.inTx = false
	t.c.db.mu.Lock()
	t.c.db.commits++
	t.c.db.mu.Unlock()
	return nil
}

func (t goldenTx) Rollback() error {
	t.c.inTx = false
	t.c.db.mu.Lock()
	t.c.db.rollbacks++
	t.c.db.mu.Unlock()
	return nil
}

func goldenOpen(t *testing.T, g *goldenDB) *sql.DB {
	t.Helper()
	db := sql.OpenDB(g)
	t.Cleanup(func() { db.Close() })
	return db
}

func goldenEvents(n int) []Event {
	at := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	events := make([]Event, n)
	for i := range events {
		events[i] = Event{
			ID:     int64(i + 1),
			Tenant: "t" + strconv.Itoa(i%7),
			Kind:   "charge",
			At:     at.Add(time.Duration(i) * time.Second),
			Amount: int64(i * 3),
			Note:   "n" + strconv.Itoa(i),
		}
	}
	return events
}

var (
	goldenColumns     = regexp.MustCompile(`(?is)insert\s+into\s+"?events"?\s*\(([^)]*)\)`)
	goldenPlaceholder = regexp.MustCompile(`\$(\d+)`)
)

// goldenIDs reads the id of every row the statements carried, in order. It
// takes the column list from the statement and the rows as consecutive
// groups of that many arguments.
func goldenIDs(t *testing.T, execs []goldenExec) []int64 {
	t.Helper()
	var ids []int64
	for i, e := range execs {
		m := goldenColumns.FindStringSubmatch(e.query)
		if m == nil {
			t.Fatalf("statement %d is not an INSERT INTO events (columns): %.120q", i+1, e.query)
		}
		var cols []string
		for c := range strings.SplitSeq(m[1], ",") {
			cols = append(cols, strings.Trim(strings.ToLower(strings.TrimSpace(c)), `"`))
		}
		idCol := -1
		for j, c := range cols {
			if c == "id" {
				idCol = j
			}
		}
		if idCol < 0 || len(e.args)%len(cols) != 0 {
			t.Fatalf("statement %d: columns %v do not divide its %d arguments into rows", i+1, cols, len(e.args))
		}
		top := 0
		for _, pm := range goldenPlaceholder.FindAllStringSubmatch(e.query, -1) {
			n, _ := strconv.Atoi(pm[1])
			top = max(top, n)
		}
		if top != len(e.args) {
			t.Fatalf("statement %d uses placeholders up to $%d but sends %d arguments", i+1, top, len(e.args))
		}
		for r := 0; r < len(e.args); r += len(cols) {
			id, ok := e.args[r+idCol].(int64)
			if !ok {
				t.Fatalf("statement %d: id argument is %T, want int64", i+1, e.args[r+idCol])
			}
			ids = append(ids, id)
		}
	}
	return ids
}

// TestInsertWritesTheExport is the trap. One INSERT with every row is the
// fewest statements, and 50,000 rows of six columns is 300,000 parameters.
func TestInsertWritesTheExport(t *testing.T) {
	const n = 50_000
	g := &goldenDB{}
	if err := Insert(t.Context(), goldenOpen(t, g), goldenEvents(n)); err != nil {
		t.Fatalf("Insert(%d events) error = %v, want nil", n, err)
	}
	if len(g.execs) > 10 {
		t.Errorf("Insert(%d events) sent %d statements, want ten at most", n, len(g.execs))
	}
	if g.outsideTx > 0 || g.commits != 1 || g.rollbacks != 0 {
		t.Errorf("Insert(%d events): %d statements outside a transaction, %d commits, %d rollbacks; want all in one committed transaction", n, g.outsideTx, g.commits, g.rollbacks)
	}
	ids := goldenIDs(t, g.execs)
	if len(ids) != n {
		t.Fatalf("Insert(%d events) wrote %d rows, want %d", n, len(ids), n)
	}
	for i, id := range ids {
		if id != int64(i+1) {
			t.Fatalf("row %d has id %d, want %d: rows go in the order given", i+1, id, i+1)
		}
	}
}

func TestInsertRollsBackOnFailure(t *testing.T) {
	const n = 30_000
	g := &goldenDB{failOn: 2}
	err := Insert(t.Context(), goldenOpen(t, g), goldenEvents(n))
	if !errors.Is(err, goldenErrDisk) {
		t.Fatalf("Insert with the second statement failing: error = %v, want it to wrap %v", err, goldenErrDisk)
	}
	if g.commits != 0 || g.rollbacks != 1 {
		t.Errorf("Insert with the second statement failing: %d commits, %d rollbacks; want 0 and 1", g.commits, g.rollbacks)
	}
}

func TestInsertSmallExport(t *testing.T) {
	g := &goldenDB{}
	events := goldenEvents(3)
	if err := Insert(t.Context(), goldenOpen(t, g), events); err != nil {
		t.Fatalf("Insert(3 events) error = %v, want nil", err)
	}
	ids := goldenIDs(t, g.execs)
	if len(ids) != 3 || g.commits != 1 {
		t.Errorf("Insert(3 events) wrote ids %v with %d commits, want [1 2 3] committed once", ids, g.commits)
	}
}

func TestInsertNothing(t *testing.T) {
	g := &goldenDB{}
	if err := Insert(t.Context(), goldenOpen(t, g), nil); err != nil {
		t.Fatalf("Insert(nil) error = %v, want nil", err)
	}
	if len(g.execs) != 0 {
		t.Errorf("Insert(nil) sent %d statements, want none", len(g.execs))
	}
}
