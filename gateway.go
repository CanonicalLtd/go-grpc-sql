package grpcsql

import "github.com/CanonicalLtd/go-grpc-sql/internal/protocol"

// Gateway mapping gRPC requests to SQL queries.
type Gateway struct {
}

// NewGateway creates a new gRPC gateway executing requests against the given
// SQL database.
func NewGateway() *Gateway {
	return &Gateway{}
}

// Conn creates a new database connection using the underlying driver, and
// start accepting requests for it.
func (s *Gateway) Conn(stream protocol.SQL_ConnServer) error {
	return nil
}
