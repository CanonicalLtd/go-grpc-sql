package connector

import "fmt"

var (
	errNoAvailableLeader = fmt.Errorf("no available gRPC SQL leader server found")
	errStop              = fmt.Errorf("connector was stopped")
	errStaleLeader       = fmt.Errorf("server has lost leadership")
	errNotClustered      = fmt.Errorf("server is not clustered")
)
