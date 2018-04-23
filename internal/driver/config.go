package driver

import (
	"time"
)

// Config holds various configuration parameters for Driver.
type Config struct {
	// How long to wait before giving up creating a connection.
	ConnectTimeout time.Duration

	// How log to wait before cancelling a gRPC method invokation, when
	// stdlib sql APIs are used that do not support passing a context.
	//
	// This is currently used only for Tx.Commit() and Tx.Rollback().
	MethodTimeout time.Duration
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() Config {
	return Config{
		ConnectTimeout: 5 * time.Second,
	}
}
