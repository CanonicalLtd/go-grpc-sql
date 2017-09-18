package grpcsql_test

import (
	"github.com/CanonicalLtd/go-grpc-sql"
)

func newGateway() *grpcsql.Gateway {
	return grpcsql.NewGateway()
}
