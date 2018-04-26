package grpcsql_test

import (
	"database/sql"
	"fmt"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/CanonicalLtd/go-grpc-sql"
	"github.com/CanonicalLtd/go-grpc-sql/cluster"
	"github.com/CanonicalLtd/go-sqlite3"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func TestIntegration(t *testing.T) {
	listener := newListener(t)
	server := newServer(t)
	go server.Serve(listener)
	defer server.Stop()

	name := registerDriver(t, listener.Addr().String())

	db, err := sql.Open(name, ":memory:")
	require.NoError(t, err)

	ctx, cancel := newContext()
	defer cancel()

	_, err = db.BeginTx(ctx, &sql.TxOptions{})
	require.NoError(t, err)
}

// Create a new net.Listener.
func newListener(t *testing.T) net.Listener {
	t.Helper()

	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	return listener
}

// Return a new test gRPC server configured with gRPC SQL service.
func newServer(t *testing.T) *grpc.Server {
	server := grpc.NewServer()
	driver := &sqlite3.SQLiteDriver{}
	logger := zaptest.NewLogger(t)

	grpcsql.RegisterService(server, driver, grpcsql.ServiceLogger(logger))

	return server
}

// Regisgter a new test gRPC SQL driver in the stdlib sql package. Returns the
// name used for registration
func registerDriver(t *testing.T, target string) string {
	t.Helper()

	store := newStore(t)
	logger := zaptest.NewLogger(t)

	driver := grpcsql.NewDriver(store, grpcsql.DriverLogger(logger))
	name := fmt.Sprintf("grpcsql-%d", atomic.AddInt32(&driverCount, 1))

	sql.Register(name, driver)

	// Register the server address.
	require.NoError(t, store.Set(context.Background(), []string{target}))

	return name
}

// Create a new context.Context with a timeout of 100 milliseconds.
func newContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 100*time.Millisecond)
}

// Return a new test store.
func newStore(t *testing.T) cluster.ServerStore {
	t.Helper()

	store, err := cluster.DefaultServerStore(":memory:")
	require.NoError(t, err)

	return store
}

// For generating unique driver registration names.
var driverCount int32
