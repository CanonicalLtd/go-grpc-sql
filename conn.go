package grpcsql

import (
	"database/sql/driver"
	"fmt"

	"github.com/CanonicalLtd/go-grpc-sql/internal/protocol"
	"google.golang.org/grpc"
)

// Conn wraps a connection to a gRPC SQL gateway.
type Conn struct {
	grpcConn       *grpc.ClientConn
	grpcConnClient protocol.SQL_ConnClient
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
	tx := &Tx{}
	return tx, nil
}

// Execute a request and waits for the response.
func (c *Conn) exec(request *protocol.Request) (*protocol.Response, error) {
	if err := c.grpcConnClient.Send(request); err != nil {
		return nil, fmt.Errorf("gRPC could not send %s request: %v", request.Code, err)
	}

	response, err := c.grpcConnClient.Recv()
	if err != nil {
		return nil, fmt.Errorf("gRPC %s response error: %v", request.Code, err)
	}
	return response, nil
}
