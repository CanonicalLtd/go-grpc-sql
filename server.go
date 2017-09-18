package grpcsql

import (
	"database/sql/driver"

	"github.com/CanonicalLtd/go-grpc-sql/internal/protocol"
	"google.golang.org/grpc"
)

// NewServer is a convenience for creating a gRPC server with a registered SQL gateway.
func NewServer(driver driver.Driver) *grpc.Server {
	gateway := NewGateway(driver)
	server := grpc.NewServer()
	protocol.RegisterSQLServer(server, gateway)
	return server
}
