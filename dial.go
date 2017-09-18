package grpcsql

import (
	"crypto/tls"
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Create a new connection to a gRPC SQL endpoint.
func dial(options *dialOptions) (*Conn, error) {
	grpcTarget := options.Address
	grpcOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(options.TLSConfig)),
	}
	grpcCtx := context.Background()
	if options.Timeout != 0 {
		var cancel func()
		grpcCtx, cancel = context.WithTimeout(grpcCtx, options.Timeout)
		defer cancel()

	}

	grpcConn, err := grpc.DialContext(grpcCtx, grpcTarget, grpcOptions...)
	if err != nil {
		return nil, fmt.Errorf("gRPC connection failed: %v", err)
	}

	conn := &Conn{
		grpcConn: grpcConn,
	}

	return conn, nil

}

type dialOptions struct {
	Address   string
	Timeout   time.Duration
	TLSConfig *tls.Config
}
