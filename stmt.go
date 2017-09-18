package grpcsql

import "database/sql/driver"

// Stmt is a prepared statement. It is bound to a Conn and not
// used by multiple goroutines concurrently.
type Stmt struct {
}

// Close closes the statement.
func (s *Stmt) Close() error {
	return nil
}

// NumInput returns the number of placeholder parameters.
func (s *Stmt) NumInput() int {
	return 0
}

// Exec executes a query that doesn't return rows, such
func (s *Stmt) Exec(args []driver.Value) (driver.Result, error) {
	result := &Result{}
	return result, nil
}

// Query executes a query that may return rows, such as a
func (s *Stmt) Query(args []driver.Value) (driver.Rows, error) {
	rows := &Rows{}
	return rows, nil
}
