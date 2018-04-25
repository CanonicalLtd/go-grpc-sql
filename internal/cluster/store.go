package cluster

import (
	"context"
)

// ServerStore is used by a gRPC SQL driver to get an initial list of candidate
// gRPC target names that it can dial in order to find a leader gRPC SQL server
// to use.
//
// Once connected, the driver periodically updates the targets in the store by
// querying the leader server about changes in the cluster (such as servers
// being added or removed).
type ServerStore interface {
	// Get return the list of known gRPC SQL server addresses.
	//
	// Each address name must follow the sysntax defined by the gRPC dial
	// protocol.
	//
	// See https://github.com/grpc/grpc/blob/master/doc/naming.md.
	Get(context.Context) ([]string, error)

	// Set updates the list of known gRPC SQL server addresses.
	Set(context.Context, []string) error
}

// InmemServerStore keeps the list of target gRPC SQL servers in memory.
type InmemServerStore struct {
	targets []string
}

// NewInmemServerStore creates ServerStore which stores its data in-memory.
func NewInmemServerStore() ServerStore {
	return &InmemServerStore{
		targets: make([]string, 0),
	}
}

// Get the current targets.
func (i *InmemServerStore) Get(ctx context.Context) ([]string, error) {
	return i.targets, nil
}

// Set the targets.
func (i *InmemServerStore) Set(ctx context.Context, targets []string) error {
	i.targets = targets
	return nil
}
