package grpcsql

import (
	"database/sql/driver"
	"time"

	"github.com/CanonicalLtd/go-grpc-sql/internal/gateway"
	"github.com/CanonicalLtd/go-grpc-sql/internal/sql"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// RegisterService registers the gRPC SQL service implementation to the given
// gRPC server.
//
// The gRPC SQL service will expose the given driver over gRPC using the gRPC
// SQL protocol.
//
// The endpoint of the service is "/sql.Gateway".
func RegisterService(server *grpc.Server, driver driver.Driver, options ...ServiceOption) {
	o := defaultServiceOptions()

	for _, option := range options {
		option(o)
	}

	service := gateway.New(driver, o.config, o.logger)
	sql.RegisterGatewayServer(server, service)
}

// A ServiceOption can be used to tweak various aspects of a gRPC SQL service.
type ServiceOption func(options *serviceOptions)

// ServiceLogger sets the zap logger used by a gRPC SQL service.
func ServiceLogger(logger *zap.Logger) ServiceOption {
	return func(o *serviceOptions) {
		o.logger = logger
	}
}

type serviceOptions struct {
	config gateway.Config // Gateway configuration.
	logger *zap.Logger    // Zap logger to use.
}

// Create a serviceOptions object with sane defaults.
func defaultServiceOptions() *serviceOptions {
	return &serviceOptions{
		config: defaultServiceConfig(),
		logger: defaultLogger(),
	}
}

// Sane service defaults.
func defaultServiceConfig() gateway.Config {
	return gateway.Config{
		HeartbeatTimeout: 20 * time.Second,
	}
}
