package grpcsql_test

import (
	"database/sql/driver"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CanonicalLtd/go-grpc-sql"
	"github.com/mattn/go-sqlite3"
	"github.com/mpvl/subtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Open a new gRPC connection.
func TestDriver_Open(t *testing.T) {
	driver, server := newDriver()
	defer server.Close()

	conn, err := driver.Open(":memory:")
	require.NoError(t, err)
	defer conn.Close()
}

// If one of the targets fails, the next is tried.
func TestDriver_OpenRoundRobin(t *testing.T) {
	server := newGatewayServer()
	targetsFunc := func() ([]string, error) {
		return []string{
			"1.2.3.4",
			server.Listener.Addr().String(),
		}, nil
	}
	driver := grpcsql.NewDriver(targetsFunc, tlsConfig, 2*time.Second)
	defer server.Close()

	conn, err := driver.Open(":memory:")
	require.NoError(t, err)
	defer conn.Close()
}

// Create a transaction and commit it.
func TestDriver_TxCommit(t *testing.T) {
	driver, server := newDriver()
	defer server.Close()

	conn, err := driver.Open(":memory:")
	require.NoError(t, err)
	defer conn.Close()

	tx, err := conn.Begin()
	require.NoError(t, err)
	assert.NoError(t, tx.Commit())
}

// Error happening upon Conn.Begin.
func TestDriver_BeginError(t *testing.T) {
	driver, server := newDriver()
	defer server.Close()

	conn, err := driver.Open(":memory:")
	require.NoError(t, err)
	defer conn.Close()

	_, err = conn.Begin()
	require.NoError(t, err)

	// Trying to run a second BEGIN will fail.
	_, err = conn.Begin()
	require.NotNil(t, err)
	sqliteErr, ok := err.(sqlite3.Error)
	require.True(t, ok)
	assert.Equal(t, sqlite3.ErrNo(1), sqliteErr.Code)
	assert.Equal(t, "cannot start a transaction within a transaction", sqliteErr.Error())
}

// Open a new gRPC connection.
func TestDriver_BadConn(t *testing.T) {
	drv, server := newDriver()
	conn, err := drv.Open(":memory:")
	assert.NoError(t, err)
	defer conn.Close()

	// Shutdown the server to interrupt the gRPC connection.
	server.CloseClientConnections()

	stmt, err := conn.Prepare("SELECT * FROM sqlite_master")
	assert.Nil(t, stmt)
	assert.Equal(t, driver.ErrBadConn, err)
}

// Possible failure modes of Driver.Open().
func TestDriver_OpenError(t *testing.T) {
	cases := []struct {
		title       string
		targetsFunc grpcsql.TargetsFunc
		err         string
	}{
		{
			"invalid escape",
			func() ([]string, error) {
				return []string{"1.2.3.4"}, nil
			},
			"gRPC conn method failed",
		},
	}
	for _, c := range cases {
		subtest.Run(t, c.title, func(t *testing.T) {
			driver := grpcsql.NewDriver(c.targetsFunc, tlsConfig, time.Millisecond)
			_, err := driver.Open(":memory:")
			require.NotNil(t, err)
			assert.Contains(t, err.Error(), c.err)
		})
	}
}

// Return a new Driver instance configured to connect to a test gRPC server.
func newDriver() (*grpcsql.Driver, *httptest.Server) {
	server := newGatewayServer()
	targetsFunc := func() ([]string, error) {
		return []string{server.Listener.Addr().String()}, nil
	}
	driver := grpcsql.NewDriver(targetsFunc, tlsConfig, 2*time.Second)
	return driver, server
}
