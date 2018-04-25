package cluster

// DriverCluster is an optional interface that may be implemented by a standard
// driver.Driver that is cluster-aware (such as dqlite).
type DriverCluster interface {
	// Return the address of the current cluster leader, if any. If not
	// empty, the address string must a be valid gRPC target name, that
	// clients can use to connect to the driver using the go-grpc-sql
	// protocol.
	//
	// See https://github.com/grpc/grpc/blob/master/doc/naming.md.
	Leader() string

	// If this driver is the current leader of the cluster, return the
	// addresses of all other servers. Each address must be a valid gRPC
	// target name, that clients can use to connect to that driver using
	// go-grpc-sql, in case the current leader is deposed and a new one is
	// elected.
	//
	// If this driver is not the current leader of the cluster, an error
	// implementing the Error interface below and returning true in
	// NotLeader() must be returned.
	Servers() ([]string, error)

	// Attempt to recover the transaction associated with the given token.
	Recover(uint64) error
}

// TxToken is an optional interface that can be implemented by a driver.Tx in
// case a failed commit or rollback can be recovered by contacting another
// server and presenting a transaction-specific token.
type TxToken interface {
	Token() uint64 // Token to present to the other server.
}

// An Error represents a cluster-related error.
type Error interface {
	error
	NotLeader() bool // Is the error due to the server not being the leader?
}
