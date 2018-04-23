package driver_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/CanonicalLtd/go-grpc-sql/internal/driver"
	"github.com/CanonicalLtd/go-grpc-sql/internal/gateway"
	"github.com/CanonicalLtd/go-grpc-sql/internal/sql"
	"github.com/CanonicalLtd/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
)

func TestDriver_Open(t *testing.T) {
	driver, cleanup := newDriver(t)
	defer cleanup()

	conn, err := driver.Open(":memory:")
	require.NoError(t, err)

	assert.NoError(t, conn.Close())
}

// Create a new driver with a client factory pointing to test gateway gRPC
// service.
func newDriver(t *testing.T) (*driver.Driver, func()) {
	t.Helper()

	logger := zaptest.NewLogger(t)
	gateway := newGateway(t, logger)
	server := newServer(t, gateway)

	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go server.Serve(listener)

	connector := func(context.Context) (sql.GatewayClient, error) {
		conn, err := grpc.Dial(listener.Addr().String(), grpc.WithInsecure())
		if err != nil {
			return nil, err
		}
		client := sql.NewGatewayClient(conn)
		return client, nil
	}

	config := driver.Config{
		ConnectTimeout: 100 * time.Millisecond,
		MethodTimeout:  100 * time.Millisecond,
	}

	driver := driver.New("0", connector, config, logger)

	cleanup := func() {
		server.Stop()
		gateway.Stop()
	}

	return driver, cleanup
}

// Create a new test gRPC SQL gateway identified by the given ID and using a
// SQLite driver.
func newGateway(t *testing.T, logger *zap.Logger) *gateway.Gateway {
	t.Helper()

	driver := &sqlite3.SQLiteDriver{}

	gateway := gateway.New(driver, logger)

	// Register a client with ID 0.
	_, err := gateway.Register(context.Background(), sql.NewClient("0"))
	require.NoError(t, err)

	return gateway
}

// Create a new gRPC server configured with the given gateway service.
func newServer(t *testing.T, gateway *gateway.Gateway) *grpc.Server {
	t.Helper()
	server := grpc.NewServer()
	sql.RegisterGatewayServer(server, gateway)
	return server
}
