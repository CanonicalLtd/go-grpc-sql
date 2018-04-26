package connector_test

import (
	"context"
	"database/sql/driver"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/CanonicalLtd/go-grpc-sql/internal/cluster"
	"github.com/CanonicalLtd/go-grpc-sql/internal/connector"
	"github.com/CanonicalLtd/go-grpc-sql/internal/gateway"
	"github.com/CanonicalLtd/go-grpc-sql/internal/sql"
	"github.com/CanonicalLtd/go-sqlite3"
	"github.com/Rican7/retry/backoff"
	"github.com/Rican7/retry/strategy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
)

// Successful connection.
func TestConnector_Connect(t *testing.T) {
	cluster, cleanup := newCluster(t, 3)
	defer cleanup()

	connector, cleanup := newConnectorForCluster(t, cluster)
	defer cleanup()

	ctx := context.Background()

	client, err := connector.Connect(ctx)
	require.NoError(t, err)

	conn, err := client.Connect(ctx, sql.NewDatabase("0", ":memory:"))
	require.NoError(t, err)
	require.NotNil(t, conn)
}

// Connection failed because the server store is empty.
func TestConnector_Connect_Error_EmptyServerStore(t *testing.T) {
	store := cluster.NewInmemServerStore()

	connector, cleanup := newConnectorWithStore(t, store)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	_, err := connector.Connect(ctx)
	require.EqualError(t, err, "no available gRPC SQL leader server found")
}

// Connection failed because the connector was stopped.
func TestConnector_Connect_Error_AfterStop(t *testing.T) {
	cluster, cleanup := newCluster(t, 3)
	defer cleanup()

	connector, cleanup := newConnectorForCluster(t, cluster)
	defer cleanup()

	connector.Stop()

	ctx := context.Background()

	_, err := connector.Connect(ctx)
	require.EqualError(t, err, "connector was stopped")
}

// If an election is in progress, the connector will retry until a leader gets elected.
func TestConnector_Connect_ElectionInProgress(t *testing.T) {
	cluster, cleanup := newCluster(t, 3)
	defer cleanup()

	cluster.DriverLeaders()

	connector, cleanup := newConnectorForCluster(t, cluster)
	defer cleanup()

	go func() {
		// Simulate server 1 winning the election after 10ms
		time.Sleep(10 * time.Millisecond)
		cluster.DriverLeaders(1)
	}()

	ctx := context.Background()

	_, err := connector.Connect(ctx)
	require.NoError(t, err)
}

// If a server reports that it knows about the leader, the hint will be taken
// and an attempt will be made to connect to it.
func TestConnector_Connect_ServerKnowsAboutLeader(t *testing.T) {
	cluster, cleanup := newCluster(t, 3)
	defer cleanup()

	// Server 0 will be contacted first, which will report that server 1 is
	// the leader.
	cluster.DriverLeaders(1)

	connector, cleanup := newConnectorForCluster(t, cluster)
	defer cleanup()

	ctx := context.Background()

	_, err := connector.Connect(ctx)
	require.NoError(t, err)
}

// If a server reports that it knows about the leader, the hint will be taken
// and an attempt will be made to connect to it. If that leader has died, the
// next target will be tried.
func TestConnector_Connect_ServerKnowsAboutDeadLeader(t *testing.T) {
	cluster, cleanup := newCluster(t, 3)
	defer cleanup()

	// Server 0 will be contacted first, which will report that server 1 is
	// the leader. However server 1 has crashed, and after a bit server 0
	// gets elected.
	cluster.DriverLeaders(1)
	cluster.Stop(1)

	connector, cleanup := newConnectorForCluster(t, cluster)
	defer cleanup()

	go func() {
		// Simulate server 0 becoming the new leader after server 1
		// crashed.
		time.Sleep(10 * time.Millisecond)
		cluster.DriverLeaders(0)
	}()

	ctx := context.Background()

	_, err := connector.Connect(ctx)
	require.NoError(t, err)
}

// If a server reports that it knows about the leader, the hint will be taken
// and an attempt will be made to connect to it. If that leader is not actually
// the leader the next target will be tried.
func TestConnector_Connect_ServerKnowsAboutStaleLeader(t *testing.T) {
	cluster, cleanup := newCluster(t, 3)
	defer cleanup()

	// Server 0 will be contacted first, which will report that server 1 is
	// the leader. However server 1 things that 2 is the leader, and server
	// 2 is actually the leader.
	cluster.DriverLeaders(1, 2, 2)

	connector, cleanup := newConnectorForCluster(t, cluster)
	defer cleanup()

	go func() {
		// Simulate server 0 becoming the new leader after server 1
		// crashed.
		time.Sleep(10 * time.Millisecond)
		cluster.DriverLeaders(0)
	}()

	ctx := context.Background()

	_, err := connector.Connect(ctx)
	require.NoError(t, err)
}

// After a heartbeat failure the current connection gets closed.
func TestConnector_HeartbeatFailure(t *testing.T) {
	cluster, cleanup := newCluster(t, 3)
	defer cleanup()

	connector, cleanup := newConnectorForCluster(t, cluster)
	defer cleanup()

	ctx := context.Background()

	client, err := connector.Connect(ctx)
	require.NoError(t, err)

	cluster.Stop(0)

	_, err = client.Connect(ctx, sql.NewDatabase("0", ":memory:"))
	requireErrorDisconnect(t, err, cluster.Addresses()[0])
}

// After a heartbeat failure, if the connector has stopped, no more connections
// can be opened.
func TestConnector_HeartbeatFailure_AfterStop(t *testing.T) {
	cluster, cleanup := newCluster(t, 3)
	defer cleanup()

	connector, cleanup := newConnectorForCluster(t, cluster)
	defer cleanup()

	ctx := context.Background()

	_, err := connector.Connect(ctx)
	require.NoError(t, err)

	cluster.Stop(0)
	connector.Stop()

	_, err = connector.Connect(ctx)
	require.EqualError(t, err, "connector was stopped")
}

// After a connected leader server dies and the heartbeat fails, a new
// connection against the new leader is established.
func TestConnector_Failover(t *testing.T) {
	cluster, cleanup := newCluster(t, 3)
	defer cleanup()

	connector, cleanup := newConnectorForCluster(t, cluster)
	defer cleanup()

	ctx := context.Background()

	client, err := connector.Connect(ctx)
	require.NoError(t, err)

	cluster.Stop(0)
	cluster.DriverLeaders(1)

	// The first client does not work anymore.
	_, err = client.Connect(ctx, sql.NewDatabase("0", ":memory:"))
	requireErrorDisconnect(t, err, cluster.Addresses()[0])

	// A new client can be acquired.
	client, err = connector.Connect(ctx)
	require.NoError(t, err)

	conn, err := client.Connect(ctx, sql.NewDatabase("0", ":memory:"))
	require.NoError(t, err)
	require.NotNil(t, conn)
}

// A connected leader server dies after a few heartbeats, however it manages to
// update the connector's server store with a new list of addresses, one of
// which becomes the new leader.
func TestConnector_Failover_AfterServersUpdate(t *testing.T) {
	cluster, cleanup := newCluster(t, 3)
	defer cleanup()

	// The store initially knows only about the first 2 servers.
	store := newStore(t, cluster.Addresses()[:2])

	connector, cleanup := newConnectorWithStore(t, store)
	defer cleanup()

	// Simulate server 0 crashing after a few heartbeats have been
	// made. The new leader will be server 2 which was not
	// initially known by the client
	time.Sleep(100 * time.Millisecond)
	cluster.Stop(0)
	cluster.DriverLeaders(2)

	ctx := context.Background()

	_, err := connector.Connect(ctx)
	require.NoError(t, err)

	// The store now knows about all servers.
	addresses, err := store.Get(ctx)
	require.NoError(t, err)

	assert.Equal(t, cluster.Addresses(), addresses)
}

// Return a new Connector whose store is an in-memory store set with all the given
// addresses in the given cluster.
func newConnectorForCluster(t *testing.T, cluster *Cluster) (*connector.Connector, func()) {
	t.Helper()

	store := newStore(t, cluster.Addresses())

	return newConnectorWithStore(t, store)
}

func newConnectorWithStore(t *testing.T, store cluster.ServerStore) (*connector.Connector, func()) {
	t.Helper()

	logger := zaptest.NewLogger(t)
	config := connector.Config{
		AttemptTimeout: 100 * time.Millisecond,
		DialOptions: []grpc.DialOption{
			grpc.WithInsecure(),
		},
		RetryStrategies: []strategy.Strategy{
			strategy.Backoff(backoff.BinaryExponential(time.Millisecond)),
		},
	}

	connector := connector.New("0", store, config, logger)

	cleanup := func() {
		connector.Stop()
		logger.Sync()
	}
	return connector, cleanup
}

// Create a new in-memory server store populated with the given targets.
func newStore(t *testing.T, targets []string) cluster.ServerStore {
	t.Helper()

	store := cluster.NewInmemServerStore()
	require.NoError(t, store.Set(context.Background(), targets))

	return store
}

// A cluster of gRPC SQL servers.
type Cluster struct {
	addresses []string                  // Addresses all all servers in the cluster.
	drivers   map[string]*sqlite3Driver // Driver by network address.
	servers   map[string]*grpc.Server   // gRPC servers by network address.
}

func newCluster(t *testing.T, n int) (*Cluster, func()) {
	t.Helper()

	gateways := make([]*gateway.Gateway, n)

	addresses := make([]string, n)
	servers := make(map[string]*grpc.Server)
	drivers := make(map[string]*sqlite3Driver)

	for i := 0; i < n; i++ {
		driver := newSQLite3Driver()
		gateway := newGateway(t, i, driver)

		server := newServer(gateway)

		listener, err := net.Listen("tcp", ":0")
		require.NoError(t, err)

		go server.Serve(listener)

		gateways[i] = gateway

		address := listener.Addr().String()
		addresses[i] = address
		drivers[address] = driver
		servers[address] = server
	}

	// Initially all drivers know about each other and the leader is the
	// first driver.
	for _, driver := range drivers {
		driver.SetServers(addresses)
		driver.SetLeader(addresses[0])
	}

	cluster := &Cluster{
		addresses: addresses,
		drivers:   drivers,
		servers:   servers,
	}

	cleanup := func() {
		for _, server := range servers {
			server.Stop()
		}
		for _, gateway := range gateways {
			gateway.Stop()
		}
	}

	return cluster, cleanup
}

// Update what server each driver thinks the leader is.
//
// If a single value is passed, all drivers will agree about that server being
// the leader.
//
// If no value is passed, all drivers will have no idea about the leader.
//
// If multiple values are passed, drivers will have different opinions about
// the leader.
func (c *Cluster) DriverLeaders(servers ...int) {
	leaders := make([]string, len(c.drivers))
	if len(servers) > 0 {
		if len(servers) == 1 {
			for i := range leaders {
				leaders[i] = c.addresses[servers[0]]
			}
		} else {
			for i := range leaders {
				leaders[i] = c.addresses[servers[i]]
			}
		}
	}
	for i, leader := range leaders {
		driver := c.drivers[c.addresses[i]]
		driver.SetLeader(leader)
	}
}

// Update the servers that the given driver knowns about.
func (c *Cluster) DriverServers(i int, servers []string) {
	driver := c.drivers[c.addresses[i]]
	driver.SetServers(servers)
}

// Return all server addresses.
func (c *Cluster) Addresses() []string {
	return c.addresses
}

// Stop the gRPC server with the given index.
func (c *Cluster) Stop(i int) {
	c.servers[c.addresses[i]].Stop()
}

// Create a new gRPC SQL gateway.
func newGateway(t *testing.T, i int, driver driver.Driver) *gateway.Gateway {
	logger := zaptest.NewLogger(t).With(zap.Namespace(fmt.Sprintf("server %d", i)))
	config := gateway.Config{
		HeartbeatTimeout: 200 * time.Millisecond,
	}
	return gateway.New(driver, config, logger)
}

// Create a new gRPC server configured with the given gateway service.
func newServer(gateway *gateway.Gateway) *grpc.Server {
	server := grpc.NewServer()
	sql.RegisterGatewayServer(server, gateway)
	return server
}

// A sqlite3.SQLiteDriver extended with the cluster.DriverCluster interface.
type sqlite3Driver struct {
	sqlite3.SQLiteDriver
	mu      sync.RWMutex
	leader  string
	servers []string
}

func newSQLite3Driver() *sqlite3Driver {
	return &sqlite3Driver{}
}

func (d *sqlite3Driver) Leader() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.leader
}

func (d *sqlite3Driver) Servers() ([]string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.servers, nil
}

func (d *sqlite3Driver) Recover(uint64) error {
	return fmt.Errorf("not implemented")
}

func (d *sqlite3Driver) SetLeader(leader string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.leader = leader
}

func (d *sqlite3Driver) SetServers(servers []string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.servers = servers
}

// Convenience to assert that an error return by a gRPC SQL client is due to a
// transport disconnection from the server with the given address.
func requireErrorDisconnect(t *testing.T, err error, address string) {
	t.Helper()

	require.Error(t, err)

	// Possible error messages, depending on the timing of the gRPC client
	// connection state changes.
	messages := []string{
		msgTransportIsClosing,
		msgClientIsClosing,
		msgTransientFailure,
		fmt.Sprintf(msgConnectionRefused, address),
	}
	match := false
	for _, message := range messages {
		if err.Error() == message {
			match = true
			break
		}
	}
	if !match {
		t.Fatal("unexpected error:", err)
	}
}

const (
	msgTransportIsClosing = "rpc error: code = Unavailable desc = transport is closing"
	msgClientIsClosing    = "rpc error: code = Canceled desc = grpc: the client connection is closing"
	msgTransientFailure   = "rpc error: code = Unavailable desc = all SubConns are in TransientFailure" +
		", latest connection error: <nil>"
	msgConnectionRefused = "rpc error: code = Unavailable desc = all SubConns are in TransientFailure" +
		", latest connection error: connection error: desc = " +
		"\"transport: Error while dialing dial tcp %s: connect: connection refused\""
)
