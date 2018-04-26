package connector

import (
	"time"

	"github.com/Rican7/retry/strategy"
	"google.golang.org/grpc"
)

// Config holds various configuration parameters for Connector.
type Config struct {
	AttemptTimeout  time.Duration       // Timeout for each individual connection attempt.
	DialOptions     []grpc.DialOption   // Options to pass to grpc.DialContext.
	RetryStrategies []strategy.Strategy // Strategies used for retrying to connect to a leader.
}
