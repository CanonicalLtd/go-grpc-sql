package grpcsql

import (
	"database/sql/driver"

	"github.com/CanonicalLtd/go-grpc-sql/internal/protocol"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Driver implements the database/sql/driver interface and executes the
// relevant statements over gRPC.
type Driver struct {
	dialer Dialer
}

// NewDriver creates a new gRPC SQL driver for creating connections to backend
// gateways.
func NewDriver(dialer Dialer) *Driver {
	return &Driver{
		dialer: dialer,
	}
}

// Dialer is a function that can create a gRPC connection.
type Dialer func() (conn *grpc.ClientConn, err error)

// Open a new connection against a gRPC SQL server.
//
// To establish the gRPC connection, the dialer passed to NewDriver() will
// used.
//
// The given data source name must be one that the driver attached to the
//remote Gateway can understand.
func (d *Driver) Open(name string) (driver.Conn, error) {
	conn, err := dial(d.dialer, name)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Create a new connection to a gRPC endpoint.
func dial(dialer Dialer, name string) (*Conn, error) {
	grpcConn, err := dialer()
	if err != nil {
		return nil, errors.Wrapf(err, "gRPC grpcConnection failed")
	}

	grpcClient := protocol.NewSQLClient(grpcConn)
	grpcConnClient, err := grpcClient.Conn(context.Background())
	if err != nil {
		return nil, errors.Wrapf(err, "gRPC conn method failed")
	}

	conn := &Conn{
		grpcConn:       grpcConn,
		grpcConnClient: grpcConnClient,
	}

	if _, err := conn.exec(protocol.NewRequestOpen(name)); err != nil {
		return nil, errors.Wrapf(err, "gRPC could not send open request")
	}

	return conn, nil
}
