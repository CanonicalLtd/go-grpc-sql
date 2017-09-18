package grpcsql

import (
	"database/sql/driver"
	"fmt"
	"io"

	"github.com/CanonicalLtd/go-grpc-sql/internal/protocol"
	"github.com/golang/protobuf/proto"
)

// Gateway mapping gRPC requests to SQL queries.
type Gateway struct {
	driver driver.Driver // Underlying SQL driver.
}

// NewGateway creates a new gRPC gateway executing requests against the given
// SQL driver.
func NewGateway(drv driver.Driver) *Gateway {
	return &Gateway{
		driver: drv,
	}
}

// Conn creates a new database connection using the underlying driver, and
// start accepting requests for it.
func (s *Gateway) Conn(stream protocol.SQL_ConnServer) error {
	var conn *gatewayConn
	for {
		request, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to receive request: %v", err)
		}

		if conn == nil {
			conn = &gatewayConn{
				driver: s.driver,
				stmts:  make(map[int64]driver.Stmt),
			}
			defer conn.Close()
		}

		response, err := conn.Handle(request)
		if err != nil {
			return fmt.Errorf("failed to handle %s request: %v", request.Code, err)
		}

		if err := stream.Send(response); err != nil {
			return fmt.Errorf("failed to send response: %v", err)
		}
	}
}

// Track a single driver connection
type gatewayConn struct {
	driver     driver.Driver
	driverConn driver.Conn
	stmts      map[int64]driver.Stmt
	serial     int64
}

// Handle a single gRPC request for this connection.
func (c *gatewayConn) Handle(request *protocol.Request) (*protocol.Response, error) {
	var message proto.Message

	switch request.Code {
	case protocol.RequestCode_OPEN:
		message = &protocol.RequestOpen{}
	case protocol.RequestCode_PREPARE:
		message = &protocol.RequestPrepare{}
	case protocol.RequestCode_EXEC:
		message = &protocol.RequestExec{}
	case protocol.RequestCode_STMT_CLOSE:
		message = &protocol.RequestStmtClose{}
	case protocol.RequestCode_CLOSE:
		message = &protocol.RequestClose{}
	default:
		return nil, fmt.Errorf("invalid request code %d", request.Code)
	}

	if err := proto.Unmarshal(request.Data, message); err != nil {
		return nil, fmt.Errorf("request parse error: %v", err)
	}

	// The very first request must be an OPEN one.
	if c.driverConn == nil && request.Code != protocol.RequestCode_OPEN {
		return nil, fmt.Errorf("expected OPEN request, got %s", request.Code)
	}

	switch r := message.(type) {
	case *protocol.RequestOpen:
		return c.handleOpen(r)
	case *protocol.RequestPrepare:
		return c.handlePrepare(r)
	case *protocol.RequestExec:
		return c.handleExec(r)
	case *protocol.RequestStmtClose:
		return c.handleStmtClose(r)
	case *protocol.RequestClose:
		return c.handleClose(r)
	default:
		panic("unhandled request payload type")
	}
}

// Close the underlying driver connection, if any.
func (c *gatewayConn) Close() error {
	if c.driverConn != nil {
		return c.driverConn.Close()
	}
	return nil
}

// Handle a request of type OPEN.
func (c *gatewayConn) handleOpen(request *protocol.RequestOpen) (*protocol.Response, error) {
	driverConn, err := c.driver.Open(request.Name)
	if err != nil {
		return nil, fmt.Errorf("could not open driver connection: %v", err)
	}

	c.driverConn = driverConn

	response := protocol.NewResponseOpen()
	return response, nil
}

// Handle a request of type PREPARE.
func (c *gatewayConn) handlePrepare(request *protocol.RequestPrepare) (*protocol.Response, error) {
	driverStmt, err := c.driverConn.Prepare(request.Query)
	if err != nil {
		return nil, err
	}
	c.serial++
	c.stmts[c.serial] = driverStmt
	return protocol.NewResponsePrepare(c.serial, driverStmt.NumInput()), nil
}

// Handle a request of type EXEC.
func (c *gatewayConn) handleExec(request *protocol.RequestExec) (*protocol.Response, error) {
	driverStmt, ok := c.stmts[request.Id]
	if !ok {
		return nil, fmt.Errorf("no prepared statement with ID %d", request.Id)
	}

	args, err := protocol.ToDriverValues(request.Args)
	if err != nil {
		return nil, err
	}

	result, err := driverStmt.Exec(args)
	if err != nil {
		return nil, err
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	response := protocol.NewResponseExec(lastInsertID, rowsAffected)

	return response, nil
}

// Handle a request of type STMT_CLOSE.
func (c *gatewayConn) handleStmtClose(request *protocol.RequestStmtClose) (*protocol.Response, error) {
	driverStmt, ok := c.stmts[request.Id]
	if !ok {
		return nil, fmt.Errorf("no prepared statement with ID %d", request.Id)
	}

	if err := driverStmt.Close(); err != nil {
		return nil, err
	}

	response := protocol.NewResponseStmtClose()
	return response, nil
}

// Handle a request of type CLOSE.
func (c *gatewayConn) handleClose(request *protocol.RequestClose) (*protocol.Response, error) {
	if err := c.driverConn.Close(); err != nil {
		return nil, err
	}

	response := protocol.NewResponseClose()
	return response, nil
}
