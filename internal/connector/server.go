package connector

import (
	"context"
	"time"

	"github.com/CanonicalLtd/go-grpc-sql/internal/cluster"
	"github.com/CanonicalLtd/go-grpc-sql/internal/sql"
	"github.com/lxc/lxd/shared/logger"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Hold information about a single connection to a gRPC SQL server.
type server struct {
	target  string              // gRPC target used to create the connection
	conn    *grpc.ClientConn    // gRPC connection.
	store   cluster.ServerStore // Update this store upon successful heartbeats.
	client  sql.GatewayClient   // gRPC SQL client instanced backed the above conn.
	logger  *zap.Logger         // Logger.
	timeout time.Duration       // Heartbeat timeout reported at registration.
}

func newServer(target string, conn *grpc.ClientConn, store cluster.ServerStore, logger *zap.Logger) *server {
	return &server{
		target: target,
		conn:   conn,
		store:  store,
		client: sql.NewGatewayClient(conn),
		logger: logger.With(zap.String("target", target)),
	}
}

// Close the connection to the server.
func (s *server) Close() {
	s.conn.Close()
}

// Return the gRPC target used to create the server connection.
func (s *server) Target() string {
	return s.target
}

// Return the gRPC SQL client connected to this server
func (s *server) Client() sql.GatewayClient {
	return s.client
}

// Get the address of the current leader as known by this server.
func (s *server) Leader(ctx context.Context) (string, error) {
	leader, err := s.client.Leader(ctx, sql.NewEmpty())
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			return "", errNotClustered
		}
		return "", err
	}
	return leader.Address, nil
}

// Register the client against this server.
func (s *server) Register(ctx context.Context, clientID string) error {
	client := sql.NewClient(clientID)

	duration, err := s.client.Register(ctx, client)
	if err != nil {
		logger.Warn("registration failed", zap.String("err", err.Error()))
		return err
	}

	s.timeout = time.Duration(duration.Value)

	return nil
}

// Send heartbeats to the server until we hit an error.
func (s *server) Heartbeat(clientID string) error {
	logger := s.logger.With(zap.String("client", clientID))
	// First attempt to open an heartbeat stream against the
	// server.
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout/10)
	defer cancel()

	stream, err := s.client.Heartbeat(ctx)
	if err != nil {
		// If we can't get a heartbeat stream, most likely the server
		// is down or unreachable. Let's bail out.
		logger.Info("failed to open heartbeat stream", zap.String("err", err.Error()))
		return err
	}

	// Execute one heartbeat roundtrip every timeout/5, to stay on the
	// safe side in case of network delays or busy server. Since the
	// default timeout is 20 seconds, the interval will be 4 seconds by
	// default.
	interval := s.timeout / 5

	err = s.heartbeatStream(stream, clientID, interval)

	logger.Info("heartbeat stream interrupted", zap.String("err", err.Error()))

	return err
}

// Send heartbeats to the server at regular intervals until an error occurrs.
func (s *server) heartbeatStream(stream sql.Gateway_HeartbeatClient, id string, interval time.Duration) error {
	for {
		if err := s.heartbeatStreamRoundtrip(stream, id); err != nil {
			return err
		}

		time.Sleep(interval)
	}
}

// Handle a single heartbeat roundtrip.
func (s *server) heartbeatStreamRoundtrip(stream sql.Gateway_HeartbeatClient, id string) error {
	if err := stream.Send(sql.NewClient(id)); err != nil {
		return errors.Wrap(err, "failed to send heartbeat")
	}

	cluster, err := stream.Recv()
	if err != nil {
		return errors.Wrap(err, "failed to receive heartbeat")
	}

	// The Servers field is set to nil if the driver does not
	// implement DriverCluster, so don't update the store in that
	// case.
	if cluster.Servers == nil {
		return nil
	}

	// Update the store with the current servers.
	targets := make([]string, len(cluster.Servers))
	for i, server := range cluster.Servers {
		targets[i] = server.Address
	}

	ctx, cancel := context.WithTimeout(context.Background(), serverStoreSetTimeout)
	defer cancel()

	if err := s.store.Set(ctx, targets); err != nil {
		return errors.Wrap(err, "failed to update servers store")
	}

	return nil
}

// Hard coded timeout for updating the server store.
var serverStoreSetTimeout = 250 * time.Millisecond
