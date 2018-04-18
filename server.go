package grpcsql

import (
	"database/sql/driver"

	"github.com/CanonicalLtd/go-grpc-sql/internal/legacy"
	"google.golang.org/grpc"
)

// NewServer is a convenience for creating a gRPC server with a registered SQL
// gateway backed by the given driver.
func NewServer(driver driver.Driver, opt ...grpc.ServerOption) *grpc.Server {
	gateway := NewGateway(driver)
	server := grpc.NewServer(opt...)
	legacy.RegisterSQLServer(server, gateway)
	return server
}
