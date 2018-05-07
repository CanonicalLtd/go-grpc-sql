package grpcsql

import (
	"crypto/rand"
	sqldriver "database/sql/driver"
	"fmt"
	"time"

	"github.com/CanonicalLtd/go-grpc-sql/internal/cluster"
	"github.com/CanonicalLtd/go-grpc-sql/internal/connector"
	"github.com/CanonicalLtd/go-grpc-sql/internal/driver"
	"github.com/Rican7/retry/backoff"
	"github.com/Rican7/retry/jitter"
	"github.com/Rican7/retry/strategy"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// NewDriver creates a new gRPC SQL driver for creating connections to backend
// gateways.
func NewDriver(store cluster.ServerStore, options ...DriverOption) sqldriver.Driver {
	o := defaultDriverOptions()

	for _, option := range options {
		option(o)
	}

	connector := connector.New(o.id, store, o.connectorConfig, o.logger)
	config := defaultDriverConfig()

	return driver.New(o.id, connector.Connect, config, o.logger)
}

// DriverCluster is an optional interface that may be implemented by a standard
// driver.Driver that is cluster-aware (such as dqlite).
type DriverCluster cluster.DriverCluster

// A DriverOption can be used to tweak various aspects of a gRPC SQL driver.
type DriverOption func(options *driverOptions)

// DriverLogger sets the zap logger used by a gRPC SQL driver.
func DriverLogger(logger *zap.Logger) DriverOption {
	return func(o *driverOptions) {
		o.logger = logger
	}
}

type driverOptions struct {
	id              string           // gRPC SQL client identifier.
	connectorConfig connector.Config // Connector configuration.
	logger          *zap.Logger      // Zap logger to use.
}

// Create a driverOptions object with sane defaults.
func defaultDriverOptions() *driverOptions {
	return &driverOptions{
		id:              defaultClientID(),
		connectorConfig: defaultConnectorConfig(),
		logger:          defaultLogger(),
	}
}

// Generate a random UUID as client identifier.
func defaultClientID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// Sane connector defaults.
func defaultConnectorConfig() connector.Config {
	return connector.Config{
		AttemptTimeout: time.Second,
		DialOptions: []grpc.DialOption{
			grpc.WithInsecure(),
		},
		RetryStrategies: []strategy.Strategy{defaultRetryStrategy()},
	}
}

// Return a retry strategy with jittered exponential backoff, capped at 45
// seconds.
func defaultRetryStrategy() strategy.Strategy {
	backoff := backoff.BinaryExponential(5 * time.Millisecond)
	jitter := jitter.Equal(nil)
	cap := 45 * time.Second

	return func(attempt uint) bool {
		if attempt > 0 {
			duration := jitter(backoff(attempt))
			if duration > cap {
				duration = cap
			}
			time.Sleep(duration)
		}

		return true
	}
}

// Sane driver defaults.
func defaultDriverConfig() driver.Config {
	return driver.Config{
		ConnectTimeout: 1 * time.Second,
		MethodTimeout:  1 * time.Second,
	}
}
