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
func (d *Driver) Open(name string) (driver.Conn, error) {
	conn := &Conn{}
	return conn, nil
}
