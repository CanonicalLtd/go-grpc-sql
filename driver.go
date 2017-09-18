package grpcsql

import (
	"database/sql"
	"database/sql/driver"
)

func init() {
	sql.Register("grpc", &Driver{})
}

// Driver implements the database/sql/driver interface and executes the
// relevant statements over gRPC.
type Driver struct {
}

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
	dialOptions, err := dialOptionsFromDSN(name)
	if err != nil {
		return nil, err
	}

	conn, err := dial(dialOptions)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
