package gateway

import (
	"context"
	"database/sql/driver"
	"io"
	"math"
	"sync"
	"time"

	"github.com/CanonicalLtd/go-grpc-sql/internal/cluster"
	"github.com/CanonicalLtd/go-grpc-sql/internal/sql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Wraps a driver.Conn and regulates access to it.
type conn struct {
	client *client                    // Client that opened this connection.
	id     sql.ConnID                 // ID of this connection.
	conn   driver.Conn                // Underlying driver.Conn.
	mu     sync.RWMutex               // Serialize access to the fields below here.
	stmts  map[sql.StmtID]driver.Stmt // All open statements, indexed by ID.
	txs    map[sql.TxID]driver.Tx     // All open transactions, indexed by ID.
	rows   map[sql.RowsID]driver.Rows // All open rows, indexed by ID.
}

func newConn(client *client, id sql.ConnID, driverConn driver.Conn) *conn {
	return &conn{
		client: client,
		id:     id,
		conn:   driverConn,
		stmts:  make(map[sql.StmtID]driver.Stmt),
		txs:    make(map[sql.TxID]driver.Tx),
		rows:   make(map[sql.RowsID]driver.Rows),
	}
}

func (c *conn) Prepare(ctx context.Context, text string) (*sql.Stmt, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.stmts) >= math.MaxUint32 {
		return nil, status.Error(codes.ResourceExhausted, "too many open statements")
	}

	stmt, err := c.conn.(driver.ConnPrepareContext).PrepareContext(ctx, text)
	if err != nil {
		return nil, driverErrorToStatus(err)
	}

	id := sql.StmtID(len(c.stmts))
	c.stmts[id] = stmt

	return sql.NewStmt(c.id, id, int32(stmt.NumInput())), nil
}

func (c *conn) Begin(ctx context.Context) (*sql.Tx, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.txs) >= math.MaxUint32 {
		return nil, status.Error(codes.ResourceExhausted, "too many open transactions")
	}

	tx, err := c.conn.(driver.ConnBeginTx).BeginTx(ctx, driver.TxOptions{})
	if err != nil {
		return nil, driverErrorToStatus(err)
	}

	id := sql.TxID(len(c.txs))
	c.txs[id] = tx

	return sql.NewTx(c.id, id), nil
}

func (c *conn) Exec(ctx context.Context, stmtID sql.StmtID, namedValues []driver.NamedValue) (*sql.Result, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stmt, err := c.lookupStmt(stmtID)
	if err != nil {
		return nil, err
	}

	result, err := stmt.(driver.StmtExecContext).ExecContext(ctx, namedValues)
	if err != nil {
		return nil, driverErrorToStatus(err)
	}

	return newSQLResult(result)
}

func (c *conn) ExecSQL(ctx context.Context, text string, namedValues []driver.NamedValue) (*sql.Result, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result, err := c.conn.(driver.ExecerContext).ExecContext(ctx, text, namedValues)
	if err != nil {
		return nil, driverErrorToStatus(err)
	}

	return newSQLResult(result)
}

func (c *conn) Query(ctx context.Context, stmtID sql.StmtID, namedValues []driver.NamedValue) (*sql.Rows, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.rows) >= math.MaxUint32 {
		return nil, status.Error(codes.ResourceExhausted, "too many open result sets")
	}

	stmt, err := c.lookupStmt(stmtID)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.(driver.StmtQueryContext).QueryContext(ctx, namedValues)
	if err != nil {
		return nil, driverErrorToStatus(err)
	}
	columns := rows.Columns()

	id := sql.RowsID(len(c.rows))
	c.rows[id] = rows

	return sql.NewRows(c.id, id, columns), nil
}

func (c *conn) Next(ctx context.Context, rowsID sql.RowsID) (*sql.Row, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rows, err := c.lookupRows(rowsID)
	if err != nil {
		return nil, err
	}

	columns := make([]driver.Value, len(rows.Columns()))
	err = rows.Next(columns)
	if err != nil {
		if err == io.EOF {
			// No need to check the error, since no conversion is
			// performed.
			row, _ := sql.NewRow(nil)
			return row, nil
		}
		return nil, driverErrorToStatus(err)
	}

	row, err := sql.NewRow(columns)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return row, nil
}

func (c *conn) ColumnTypeScanType(ctx context.Context, rowsID sql.RowsID, index int) (*sql.Type, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rows, err := c.lookupRows(rowsID)
	if err != nil {
		return nil, err
	}

	typ := rows.(driver.RowsColumnTypeScanType).ColumnTypeScanType(index)

	return sql.NewType(typ), nil
}

func (c *conn) ColumnTypeDatabaseTypeName(ctx context.Context, rowsID sql.RowsID, index int) (*sql.TypeName, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rows, err := c.lookupRows(rowsID)
	if err != nil {
		return nil, err
	}

	name := rows.(driver.RowsColumnTypeDatabaseTypeName).ColumnTypeDatabaseTypeName(index)

	return sql.NewTypeName(name), nil
}

func (c *conn) Commit(ctx context.Context, txID sql.TxID) (*sql.Token, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	tx, err := c.lookupTx(txID)
	if err != nil {
		return nil, err
	}

	defer delete(c.txs, txID)

	token := sql.NewToken(0)

	if err := tx.Commit(); err != nil {
		if recovery, ok := tx.(cluster.TxToken); ok {
			token = sql.NewToken(recovery.Token())
		}
		return token, driverErrorToStatus(err)
	}

	return token, nil
}

func (c *conn) Rollback(ctx context.Context, txID sql.TxID) (*sql.Empty, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	tx, err := c.lookupTx(txID)
	if err != nil {
		return nil, err
	}

	defer delete(c.txs, txID)

	if err := tx.Rollback(); err != nil {
		return nil, driverErrorToStatus(err)
	}

	return sql.NewEmpty(), nil
}

func (c *conn) CloseStmt(ctx context.Context, stmtID sql.StmtID) (*sql.Empty, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	stmt, err := c.lookupStmt(stmtID)
	if err != nil {
		return nil, err
	}

	defer delete(c.stmts, stmtID)

	if err := stmt.Close(); err != nil {
		return nil, driverErrorToStatus(err)
	}

	return sql.NewEmpty(), nil
}

func (c *conn) CloseRows(ctx context.Context, rowsID sql.RowsID) (*sql.Empty, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	rows, err := c.lookupRows(rowsID)
	if err != nil {
		return nil, err
	}

	defer delete(c.rows, rowsID)

	if err := rows.Close(); err != nil {
		return nil, driverErrorToStatus(err)
	}

	return sql.NewEmpty(), nil
}

func (c *conn) Close() (*sql.Empty, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close all prepared statements. This is required by at least SQLite.
	for _, stmt := range c.stmts {
		stmt.Close() // Ignore errors
	}

	if err := c.conn.Close(); err != nil {
		return nil, driverErrorToStatus(err)
	}

	return sql.NewEmpty(), nil
}

// Return true if the client that opened this connection is dead.
// minus the given timeout.
func (c *conn) Dead(timeout time.Duration) bool {
	return c.client.Dead(timeout)
}

func (c *conn) lookupStmt(id sql.StmtID) (driver.Stmt, error) {
	stmt, ok := c.stmts[id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no prepared statement with ID %d", id)
	}

	return stmt, nil
}

func (c *conn) lookupTx(id sql.TxID) (driver.Tx, error) {
	tx, ok := c.txs[id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no transaction with ID %d", id)
	}

	return tx, nil
}

func (c *conn) lookupRows(id sql.RowsID) (driver.Rows, error) {
	rows, ok := c.rows[id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no result set with ID %d", id)
	}

	return rows, nil
}

// Wrapper around sql.NewResult that populates the LastInsertId and RowsAffected
// fields from the given driver.Result object.
func newSQLResult(result driver.Result) (*sql.Result, error) {
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return nil, driverErrorToStatus(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, driverErrorToStatus(err)
	}

	return sql.NewResult(lastInsertID, rowsAffected), nil
}
