package driver_test

import (
	"database/sql/driver"
	"testing"

	"github.com/CanonicalLtd/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Create a transaction and commit it.
func TestDriver_TxCommit(t *testing.T) {
	driver, cleanup := newDriver(t)
	defer cleanup()

	conn, err := driver.Open(":memory:")
	require.NoError(t, err)
	defer conn.Close()

	tx, err := conn.Begin()
	require.NoError(t, err)
	assert.NoError(t, tx.Commit())
}

// Error happening upon Conn.Begin.
func TestDriver_BeginError(t *testing.T) {
	driver, cleanup := newDriver(t)
	defer cleanup()

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

// The gateway is shutdown.
func TestDriver_BadConn(t *testing.T) {
	drv, cleanup := newDriver(t)

	conn, err := drv.Open(":memory:")
	assert.NoError(t, err)
	defer conn.Close()

	// Shutdown the server to interrupt the gRPC connection.
	cleanup()

	stmt, err := conn.Prepare("SELECT * FROM sqlite_master")
	assert.Nil(t, stmt)
	assert.Equal(t, driver.ErrBadConn, err)
}
