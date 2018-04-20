package gateway

import "time"

// Config holds various configuration parameters for Gateway.
type Config struct {
	HeartbeatTimeout time.Duration // Kill connections of clients sending no heartbeat within this time.
}
