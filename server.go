package grpcsql

import (
	"github.com/CanonicalLtd/go-grpc-sql/internal/protocol"
	"google.golang.org/grpc"
)

// NewServer is a convenience for creating a gRPC server with a registered SQL gateway.
func NewServer() *grpc.Server {
	gateway := NewGateway()
	server := grpc.NewServer()
	protocol.RegisterSQLServer(server, gateway)
	return server
}
