package grpcsql

import (
	"context"
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
func NewGateway(driver driver.Driver) *Gateway {
	return &Gateway{
		driver: driver,
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
			conn = &gatewayConn{driver: s.driver, ctx: stream.Context()}
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
	ctx        context.Context
	driver     driver.Driver
	driverConn driver.Conn
}

// Handle a single gRPC request for this connection.
func (c *gatewayConn) Handle(request *protocol.Request) (*protocol.Response, error) {
	var message proto.Message

	switch request.Code {
	case protocol.RequestCode_OPEN:
		message = &protocol.RequestOpen{}
	case protocol.RequestCode_PREPARE:
		message = &protocol.RequestPrepare{}
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
	_, err := c.driverConn.Prepare(request.Query)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
