package grpcsql

import (
	"database/sql/driver"

	"github.com/CanonicalLtd/go-grpc-sql/internal/protocol"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Conn wraps a connection to a gRPC SQL gateway.
type Conn struct {
	grpcConn       *grpc.ClientConn
	grpcConnClient protocol.SQL_ConnClient
	grpcConnDoomed bool // Whether the connection should be considered broken.
}

// Prepare returns a prepared statement, bound to this connection.
func (c *Conn) Prepare(query string) (driver.Stmt, error) {
	response, err := c.exec(protocol.NewRequestPrepare(query))
	if err != nil {
		return nil, err
	}
	stmt := &Stmt{
		conn:     c,
		id:       response.Prepare().Id,
		numInput: int(response.Prepare().NumInput),
	}
	return stmt, nil
}

// Close invalidates and potentially stops any current
// prepared statements and transactions, marking this
// connection as no longer in use.
func (c *Conn) Close() error {
	if _, err := c.exec(protocol.NewRequestClose()); err != nil {
		return err
	}
	return c.grpcConn.Close()
}

// Begin starts and returns a new transaction.
//
// Deprecated: Drivers should implement ConnBeginTx instead (or additionally).
func (c *Conn) Begin() (driver.Tx, error) {
	response, err := c.exec(protocol.NewRequestBegin())
	if err != nil {
		return nil, err
	}
	tx := &Tx{
		conn: c,
		id:   response.Begin().Id,
	}
	return tx, nil
}

// Execute a request and waits for the response.
func (c *Conn) exec(request *protocol.Request) (*protocol.Response, error) {
	if c.grpcConnDoomed {
		// This means that we previously failed because of a connection
		// error, so we want to just fail again (since the sql package
		// retries ErrBadConn).
		return nil, driver.ErrBadConn
	}

	if err := c.grpcConnClient.Send(request); err != nil {
		return nil, c.errorf(err, "gRPC could not send %s request", request.Code)
	}

	response, err := c.grpcConnClient.Recv()
	if err != nil {
		return nil, c.errorf(err, "gRPC %s response error", request.Code)
	}
	return response, nil
}

// If the given error is due to the gRPC endpoint being unavailable, return
// ErrBadConn and mark the connection as doomed, otherwise return the original error.
func (c *Conn) errorf(err error, format string, v ...interface{}) error {
	if grpc.Code(err) == codes.Unavailable {
		c.grpcConnDoomed = true
		return driver.ErrBadConn
	}
	return errors.Wrapf(err, format, v...)
}
