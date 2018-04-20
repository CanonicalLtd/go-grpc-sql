package gateway

import (
	"sync"
	"time"
)

// Track information about an active gRPC SQL client connected to this gateway.
type client struct {
	mu        sync.RWMutex // Serialize access to the fields below here.
	heartbeat time.Time    // Time of last successful heartbeat from this client.
}

func newClient() *client {
	return &client{
		heartbeat: time.Now(),
	}
}

// Refresh the heartbeat timestamp of this client.
func (c *client) Heartbeat() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.heartbeat = time.Now()
}

// Return true if the last heartbeat timestamp is older than the current time
// minus the given timeout.
func (c *client) Dead(timeout time.Duration) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return time.Since(c.heartbeat) > timeout
}
