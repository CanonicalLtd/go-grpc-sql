package grpcsql_test

import (
	"database/sql/driver"
	"testing"

	"google.golang.org/grpc"

	"github.com/CanonicalLtd/go-grpc-sql"
	"github.com/mpvl/subtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Open a new gRPC connection.
func TestDriver_Open(t *testing.T) {
	driver, cleanup := newDriver()
	defer cleanup()

	conn, err := driver.Open(":memory:")
	require.NoError(t, err)
	defer conn.Close()
}

// Create a transaction and commit it.
func TestDriver_TxCommit(t *testing.T) {
	driver, cleanup := newDriver()
	defer cleanup()

	conn, err := driver.Open(":memory:")
	require.NoError(t, err)
	defer conn.Close()

	tx, err := conn.Begin()
	require.NoError(t, err)
	assert.NoError(t, tx.Commit())
}

// Open a new gRPC connection.
func TestDriver_BadConn(t *testing.T) {
	drv, cleanup := newDriver()

	conn, err := drv.Open(":memory:")
	assert.NoError(t, err)
	defer conn.Close()

	// Shutdown the server to interrupt the gRPC connection.
	cleanup()

	stmt, err := conn.Prepare("SELECT * FROM sqlite_master")
	assert.Nil(t, stmt)
	assert.Equal(t, driver.ErrBadConn, err)
}

// Possible failure modes of Driver.Open().
func TestDriver_OpenError(t *testing.T) {
	cases := []struct {
		title  string
		dialer grpcsql.Dialer
		err    string
	}{
		{
			"gRPC connection failed",
			func() (*grpc.ClientConn, error) {
				return grpc.Dial("1.2.3.4", grpc.WithInsecure())
			},
			"gRPC conn method failed",
		},
	}
	for _, c := range cases {
		subtest.Run(t, c.title, func(t *testing.T) {
			driver := grpcsql.NewDriver(c.dialer)
			_, err := driver.Open(":memory:")
			require.NotNil(t, err)
			assert.Contains(t, err.Error(), c.err)
		})
	}
}

// Return a new Driver instance configured to connect to a test gRPC server.
func newDriver() (*grpcsql.Driver, func()) {
	server, address := newGatewayServer()
	dialer := func() (*grpc.ClientConn, error) {
		return grpc.Dial(address, grpc.WithInsecure())
	}
	driver := grpcsql.NewDriver(dialer)
	cleanup := func() { server.Stop() }
	return driver, cleanup
}
