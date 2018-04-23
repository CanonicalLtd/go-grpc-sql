package driver

import (
	"context"
	"database/sql/driver"

	"github.com/CanonicalLtd/go-grpc-sql/internal/sql"
)

// Stmt is a prepared statement. It is bound to a Conn and not
// used by multiple goroutines concurrently.
type Stmt struct {
	client sql.GatewayClient // gRPC SQL gateway client to use.
	conn   *sql.Conn         // Info about the connection opened on the gateway.
	stmt   *sql.Stmt         // Info about the prepared statement on the gateway.
}

func newStmt(client sql.GatewayClient, conn *sql.Conn, stmt *sql.Stmt) *Stmt {
	return &Stmt{
		client: client,
		conn:   conn,
		stmt:   stmt,
	}
}

// Close closes the statement.
//
// As of Go 1.1, a Stmt will not be closed if it's in use
// by any queries.
func (s *Stmt) Close() error {
	_, err := s.client.CloseStmt(context.Background(), s.stmt)
	if err != nil {
		return newErrorFromMethod(err)
	}
	return nil
}

// NumInput returns the number of placeholder parameters.
//
// If NumInput returns >= 0, the sql package will sanity check
// argument counts from callers and return errors to the caller
// before the statement's Exec or Query methods are called.
//
// NumInput may also return -1, if the driver doesn't know
// its number of placeholders. In that case, the sql package
// will not sanity check Exec or Query argument counts.
func (s *Stmt) NumInput() int {
	return int(s.stmt.NumInput)
}

// ExecContext executes a query that doesn't return rows, such
// as an INSERT or UPDATE.
//
// ExecContext must honor the context timeout and return when it is canceled.
func (s *Stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	boundStmt, err := sql.NewBoundStmt(s.stmt, args)
	if err != nil {
		return nil, newErrorMisuse(err)
	}

	result, err := s.client.Exec(ctx, boundStmt)
	if err != nil {
		return nil, newErrorFromMethod(err)
	}
	return newResult(result.LastInsertId, result.RowsAffected), nil
}

// Exec executes a query that doesn't return rows, such
// as an INSERT or UPDATE.
//
// Deprecated: Drivers should implement StmtExecContext instead (or additionally).
func (s *Stmt) Exec(args []driver.Value) (driver.Result, error) {
	return s.ExecContext(context.Background(), valuesToNamedValues(args))
}

// QueryContext executes a query that may return rows, such as a
// SELECT.
//
// QueryContext must honor the context timeout and return when it is canceled.
func (s *Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	boundStmt, err := sql.NewBoundStmt(s.stmt, args)
	if err != nil {
		return nil, newErrorMisuse(err)
	}

	rows, err := s.client.Query(ctx, boundStmt)
	if err != nil {
		return nil, newErrorFromMethod(err)
	}
	return newRows(s.client, s.conn, rows), nil
}

// Query executes a query that may return rows, such as a
// SELECT.
//
// Deprecated: Drivers should implement StmtQueryContext instead (or additionally).
func (s *Stmt) Query(args []driver.Value) (driver.Rows, error) {
	return s.QueryContext(context.Background(), valuesToNamedValues(args))
}
