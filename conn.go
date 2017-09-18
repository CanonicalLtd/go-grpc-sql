package grpcsql

import (
	"database/sql/driver"

	"google.golang.org/grpc"
)

// Conn wraps a connection to a gRPC SQL gateway.
type Conn struct {
	grpcConn *grpc.ClientConn
}

// Prepare returns a prepared statement, bound to this connection.
func (c *Conn) Prepare(query string) (driver.Stmt, error) {
	stmt := &Stmt{}
	return stmt, nil
}

// Close invalidates and potentially stops any current
// prepared statements and transactions, marking this
// connection as no longer in use.
func (c *Conn) Close() error {
	return c.grpcConn.Close()
}

// Begin starts and returns a new transaction.
//
// Deprecated: Drivers should implement ConnBeginTx instead (or additionally).
func (c *Conn) Begin() (driver.Tx, error) {
	tx := &Tx{}
	return tx, nil
}
