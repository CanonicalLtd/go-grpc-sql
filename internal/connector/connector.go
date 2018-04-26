package connector

import (
	"context"
	"sync"

	"github.com/CanonicalLtd/go-grpc-sql/internal/cluster"
	"github.com/CanonicalLtd/go-grpc-sql/internal/sql"
	"github.com/Rican7/retry"
	"github.com/lxc/lxd/shared/logger"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Connector is in charge of creating a gRPC SQL client connected to the
// current leader of a gRPC SQL cluster and sending heartbeats to prevent
// connections created by that client from being killed by the server.
type Connector struct {
	id     string              // Client ID to use when registering against the gateway.
	store  cluster.ServerStore // Used to get and update current cluster servers.
	config Config              // Connection parameters.
	logger *zap.Logger         // Logger.
	stopCh chan struct{}       // Stop the connector on close.
	loopCh chan struct{}       // Used to force a new connect loop iteraction.
	cond   *sync.Cond          // Block connect request until the conn field below is not nil.
	server *server             // Healty connection to a gRPC SQL leader server.
}

// New creates a new connector that can be used by a gRPC SQL driver to create
// new clients connected to a leader gRPC SQL server.
func New(id string, store cluster.ServerStore, config Config, logger *zap.Logger) *Connector {
	connector := &Connector{
		id:     id,
		store:  store,
		config: config,
		logger: logger.With(zap.Namespace("connector")),
		stopCh: make(chan struct{}),
		loopCh: make(chan struct{}),
		cond:   sync.NewCond(&sync.Mutex{}),
	}
	go connector.connectLoop()
	return connector
}

// Stop the connector. This will eventually terminate any ongoing connection
// attempt or heartbeat process.
func (c *Connector) Stop() {
	select {
	case <-c.stopCh:
		return // Already stopped.
	default:
	}
	close(c.stopCh)
}

// Connect returns a gRPC SQL client connected to a leader server.
//
// If no connection can be stablished within Config.ConnectTimeout or if the
// connector is stopped, an error is returned.
func (c *Connector) Connect(ctx context.Context) (sql.GatewayClient, error) {
	// Wait for a leader server to be available.
	server, err := c.getServer(ctx)
	if err != nil {
		return nil, err
	}

	return server.Client(), nil
}

// Connect loop is in charge of establishing a connection to leader server,
// keep it alive by sending heartbeats, and switch to the next leader if the
// current leader fails.
func (c *Connector) connectLoop() {
	for {
		server := c.connect()
		if server == nil {
			// This means the connector was stopped.
			return
		}

		logger := c.logger.With(zap.String("target", server.Target()))
		logger.Info("starting heartbeat")

		// Make this channel buffered in case we abort this iteration
		// for a reason other than a heartbeat failure, and the
		// goroutine would be left running until the heartbeat
		// eventually errors out because we closed the gRPC
		// connection. At that point we don't want the goroutine to be
		// stuck at sending to the channel.
		ch := make(chan error, 1)
		go func() {
			ch <- server.Heartbeat(c.id)
		}()

		// Heartbeat until a failure happens, or we are stopped, or the
		// loopCh triggers.
		select {
		case <-ch:
		case <-c.loopCh:
		case <-c.stopCh:
			c.setServer(nil)
			server.Close()
			return
		}

		// We lost connectivity to the server or we we were interrupted
		// by Connect() because it detected that the server lost
		// leadership, let's connect again.
		c.setServer(nil)
		server.Close()
	}
}

// Connect finds the leader gRPC SQL gateway service and returns it.
//
// If the connector is stopped before a leader is found, nil is returned.
func (c *Connector) connect() *server {
	var server *server

	// The retry strategy should be configured to retry indefinitely, until
	// the connector is stopped.
	err := retry.Retry(func(attempt uint) error {
		// Regardless of whether we succeed or not, after each attempt
		// awake all waiting goroutines spawned by getServer(), to give
		// them a chance to terminate if their context is expired.
		defer func() {
			c.setServer(server)
		}()

		logger := c.logger.With(zap.Uint("attempt", attempt))

		select {
		case <-c.stopCh:
			// Stop retrying
			return nil
		default:
		}

		var err error
		server, err = c.connectAttemptAll(logger)
		if err != nil {
			logger.Info("connection failed", zap.String("err", err.Error()))
			return err
		}

		return nil
	}, c.config.RetryStrategies...)

	if err != nil {
		// The retry strategy should never give up until success or
		// connector shutdown.
		panic("connect retry aborted unexpectedly")
	}

	return server
}

// Make a single attempt to establish a connection to the leader gRPC SQL
// gateway, trying all available targets in the store.
func (c *Connector) connectAttemptAll(logger *zap.Logger) (*server, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.AttemptTimeout)
	defer cancel()

	targets, err := c.store.Get(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get gateway targets")
	}

	logger.Info("connecting to leader server", zap.Strings("targets", targets))

	// Make an attempt for each target until we find the leader.
	for _, target := range targets {
		logger := logger.With(zap.String("target", target))

		ctx, cancel := context.WithTimeout(context.Background(), c.config.AttemptTimeout)
		defer cancel()

		server, leader, err := c.connectAttemptOne(ctx, target)
		if err != nil {
			// This gateway is unavailable, try with the next target.
			logger.Info("server connection failed", zap.String("err", err.Error()))
			continue
		}
		if server != nil {
			// We found the leader
			logger.Info("connected")
			return server, nil
		}
		if leader == "" {
			// This gateway does not know who the current leader
			// is, try with the next target.
			continue
		}

		// If we get here, it means this server reported that another
		// server is the leader, let's close the connection to this
		// server and try with the suggested one.
		logger = logger.With(zap.String("leader", leader))
		server, leader, err = c.connectAttemptOne(ctx, leader)
		if err != nil {
			// The leader reported by the target gateway is
			// unavailable, try with the next target.
			logger.Info("leader server connection failed", zap.String("err", err.Error()))
			continue
		}
		if server == nil {
			// The leader reported by the target server does not consider itself
			// the leader, try with the next target.
			logger.Info("reported leader server is not the leader")
			continue
		}
		logger.Info("connected")
		return server, nil
	}

	return nil, errNoAvailableLeader
}

// Connect to the given gRPC SQL target gateway and check if it's the leader.
//
// Return values:
//
// - Any failure is hit:                     -> nil, "", err
// - Target not leader and no leader known:  -> nil, "", nil
// - Target not leader and leader known:     -> nil, leader, nil
// - Target is the leader:                   -> server, "", nil
//
func (c *Connector) connectAttemptOne(ctx context.Context, target string) (*server, string, error) {
	// Establish the gRPC connection.
	conn, err := grpc.DialContext(ctx, target, c.config.DialOptions...)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to establish gRPC connection")
	}

	server := newServer(target, conn, c.store, c.logger)

	// Get the current leader address.
	leader, err := server.Leader(ctx)
	if err != nil {
		if err != errNotClustered {
			server.Close()
			return nil, "", errors.Wrap(err, "failed to get current leader")
		}
		// For servers whose backend driver does not implement
		// DriverCluster, we assume this is the leader.
		leader = target
	}

	switch leader {
	case "":
		// Currently this server does not know about any leader.
		server.Close()
		return nil, "", nil
	case target:
		// This server is the leader, register ourselves and return.
		if err := server.Register(ctx, c.id); err != nil {
			logger.Info("registration failed", zap.String("err", err.Error()))
			server.Close()
			server = nil
			return nil, "", err
		}

		return server, "", nil
	default:
		// This gateway knows who the current leader is.
		server.Close()
		return nil, leader, nil
	}
}

// Set the current leader server (or nil) and awake anything waiting on
// getServer.
func (c *Connector) setServer(server *server) {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()
	c.server = server
	c.cond.Broadcast()
}

// Get a leader server. If the context is done before a leader server is
// available, return an errNoAvailableLeader.
func (c *Connector) getServer(ctx context.Context) (*server, error) {
	// Make this channel buffered, so the goroutine below will not block on
	// sending to it terminates after this method returns early, because
	// the context was done or the stopCh was cllosed.
	ch := make(chan *server, 1)
	go func() {
		// Loop until we find a connected server which is still the
		// leader.
		for {
			server, err := c.waitServer(ctx)
			if err != nil {
				if err == errStaleLeader {
					continue
				}
				return
			}
			ch <- server
			return
		}
	}()

	select {
	case <-c.stopCh:
		return nil, errStop
	case <-ctx.Done():
		return nil, errNoAvailableLeader
	case server := <-ch:
		return server, nil
	}

}

// Wait for the connect loop to connect to a leader server.
func (c *Connector) waitServer(ctx context.Context) (*server, error) {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	// As long as ctx has a deadline, this loop is guaranteed to finish
	// because the connect loop will call c.cond.Broadcast() after each
	// connection attempt (either successful or not), and a connection
	// attempts can't last more than c.config.AttemptTimeout times the
	// number of available targets in c.store.
	for c.server == nil {
		select {
		case <-ctx.Done():
			// Abort
			return nil, ctx.Err()
		default:
		}
		c.cond.Wait()
	}

	// Check that this server is actually still the leader. Since
	// lost leadership is detected via heartbeats (which by default
	// happen every 4 seconds), this connect attempt might be
	// performed when the server has actually lost leadership but
	// the connect loop didn't notice it yet. For example when the
	// sql package tries to open a new driver connection after the
	// go-grpc-sql driver has returned ErrBadConn due to lost
	// leadership, we don't want to return a stale leader.
	leader, err := c.server.Leader(ctx)
	if err != errNotClustered && (err != nil || leader != c.server.Target()) {
		// Force the connect loop to start a new iteration if
		// it's heartbeating.
		c.logger.Info("detected stale leader", zap.String("target", c.server.Target()))
		c.server = nil
		select {
		case c.loopCh <- struct{}{}:
		default:
		}
		return nil, errStaleLeader
	}

	return c.server, nil
}
