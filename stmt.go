package grpcsql

import (
	"database/sql/driver"

	"github.com/CanonicalLtd/go-grpc-sql/internal/protocol"
)

// Stmt is a prepared statement. It is bound to a Conn and not
// used by multiple goroutines concurrently.
type Stmt struct {
	conn     *Conn
	id       int64
	numInput int
}

// Close closes the statement.
func (s *Stmt) Close() error {
	_, err := s.conn.exec(protocol.NewRequestStmtClose(s.id))
	return err
}

// NumInput returns the number of placeholder parameters.
func (s *Stmt) NumInput() int {
	return s.numInput
}

// Exec executes a query that doesn't return rows, such
func (s *Stmt) Exec(args []driver.Value) (driver.Result, error) {
	values, err := protocol.FromDriverValues(args)
	if err != nil {
		return nil, err
	}

	response, err := s.conn.exec(protocol.NewRequestExec(s.id, values))
	if err != nil {
		return nil, err
	}

	result := &Result{
		lastInsertID: response.Exec().LastInsertId,
		rowsAffected: response.Exec().RowsAffected,
	}
	return result, nil
}

// Query executes a query that may return rows, such as a
func (s *Stmt) Query(args []driver.Value) (driver.Rows, error) {
	values, err := protocol.FromDriverValues(args)
	if err != nil {
		return nil, err
	}

	response, err := s.conn.exec(protocol.NewRequestQuery(s.id, values))
	if err != nil {
		return nil, err
	}

	rows := &Rows{
		conn:    s.conn,
		id:      response.Query().Id,
		columns: response.Query().Columns,
	}

	return rows, nil
}
