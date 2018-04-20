package gateway

import (
	"database/sql/driver"

	"github.com/CanonicalLtd/go-grpc-sql/internal/cluster"
	"github.com/CanonicalLtd/go-grpc-sql/internal/sql"
	"github.com/CanonicalLtd/go-sqlite3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// DriverFailed is a custom gRPC error code returned when the
	// underlying driver returns an error.
	DriverFailed codes.Code = 1000
)

// Convert an error returned by the driver to a gRPC status.Status error
// object.
func driverErrorToStatus(err error) error {
	st := status.Newf(DriverFailed, "SQL driver failed: %v", err)

	code := int64(-1)
	extendedCode := int64(-1)
	message := err.Error()

	// For now only errors from the SQLite driver are supported.
	switch err := err.(type) {
	case sqlite3.Error:
		code = int64(err.Code)
		extendedCode = int64(err.ExtendedCode)
	}

	ds, err := st.WithDetails(sql.NewError(code, extendedCode, message))
	if err != nil {
		return st.Err()
	}

	return ds.Err()
}

// Convenience to invoke DriverCluster.Servers(), if implemented, and possibly
// convert the returned error into an appropriate gRPC status.Status.
func driverClusterServers(driver driver.Driver) ([]*sql.Server, error) {
	driverCluster, ok := driver.(cluster.DriverCluster)
	if !ok {
		// Driver does not implement cluster.DriverCluster, just return
		// an nil slice.
		return nil, nil
	}

	addresses, err := driverCluster.Servers()
	if err != nil {
		// If the driver uses hashicorp/raft as cluster engine backend,
		// there are two possible failure modes:
		//
		// - The call to Raft.GetConfiguration() to get the current
		//   list of servers fails. Currently in hashicorp/raft v1 the
		//   only possible error that can be returned by this API is
		//   raft.ErrShutdown, which we can safely translate to
		//   codes.Unavailable.
		//
		// - A call to Raft.State() reveals that the server is not
		//   anymore the leader and in that case the driver must return
		//   an error implementing cluster.Error indicating that the
		//   server is not the leader.
		//
		code := codes.Unavailable
		if err, ok := err.(cluster.Error); ok && err.NotLeader() {
			code = codes.FailedPrecondition
		}
		return nil, status.Errorf(code, "failed to get current servers: %v", err)
	}

	servers := make([]*sql.Server, len(addresses))

	for i, address := range addresses {
		servers[i] = sql.NewServer(address)
	}

	return servers, nil
}
