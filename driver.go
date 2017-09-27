package grpcsql

import (
	"context"
	"crypto/tls"
	"database/sql/driver"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/CanonicalLtd/go-grpc-sql/internal/protocol"
	"github.com/pkg/errors"
)

// Driver implements the database/sql/driver interface and executes the
// relevant statements over gRPC.
type Driver struct {
	targetsFunc TargetsFunc
	tlsConfig   *tls.Config
	timeout     time.Duration
}

// NewDriver creates a new gRPC SQL driver for connecting to the given backend
// gateways.
func NewDriver(targetsFunc TargetsFunc, tlsConfig *tls.Config, timeout time.Duration) *Driver {
	return &Driver{
		targetsFunc: targetsFunc,
		tlsConfig:   tlsConfig,
		timeout:     timeout,
	}
}

// TargetsFunc is a function that returns a list of gRPC targets that should be
// tried round-robin when creating a new connection in Driver.Conn.
type TargetsFunc func() ([]string, error)

// Open a new connection against a gRPC SQL server.
//
// The given data source name must be in the form:
//
// <host>[:<port>][?<options>]
//
// where valid options are:
//
// - certfile: cert file location
// - certkey: key file location
// - rootcertfile: location of the root certificate file
//
// Note that the connection will always use TLS, so valid TLS parameters for
// the target endpoint are required.
func (d *Driver) Open(name string) (driver.Conn, error) {
	targets, err := d.targetsFunc()
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get gRPC targets")
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("no gRPC target available")
	}

	var conn *Conn
	for _, target := range targets {
		conn, err = d.dial(target, name)
		if err != nil {
			continue
		}
	}
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Create a new connection to a gRPC endpoint.
func (d *Driver) dial(target string, name string) (*Conn, error) {
	grpcOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(d.tlsConfig)),
	}
	grpcCtx := context.Background()
	if d.timeout != 0 {
		var cancel func()
		grpcCtx, cancel = context.WithTimeout(grpcCtx, d.timeout)
		defer cancel()

	}

	grpcConn, err := grpc.DialContext(grpcCtx, target, grpcOptions...)
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
