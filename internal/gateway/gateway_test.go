package gateway_test

import (
	"context"
	"database/sql/driver"
	"reflect"
	"testing"
	"time"

	"github.com/CanonicalLtd/go-grpc-sql/internal/gateway"
	"github.com/CanonicalLtd/go-grpc-sql/internal/sql"
	"github.com/CanonicalLtd/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/status"
)

// Create a new connection.
func TestGateway_Connect(t *testing.T) {
	gateway := newGateway(t)
	defer gateway.Stop()

	conn, err := gateway.Connect(context.Background(), newDatabase())
	require.NoError(t, err)

	assert.Equal(t, sql.ConnID(0), conn.Id)
}

// Begin a new transaction.
func TestGateway_Begin(t *testing.T) {
	gateway := newGateway(t)
	defer gateway.Stop()

	// Create a new connection
	conn, err := gateway.Connect(context.Background(), newDatabase())
	require.NoError(t, err)

	// Start a new transaction
	tx, err := gateway.Begin(context.Background(), conn)
	require.NoError(t, err)

	assert.Equal(t, sql.TxID(0), tx.Id)
}

// Prepare a new statement.
func TestGateway_Prepare(t *testing.T) {
	gateway := newGateway(t)
	defer gateway.Stop()

	// Create a new connection
	conn, err := gateway.Connect(context.Background(), newDatabase())
	require.NoError(t, err)

	// Prepare a new CREATE TABLE statement
	stmt, err := gateway.Prepare(context.Background(), newCreateTable(conn))
	require.NoError(t, err)

	assert.Equal(t, sql.StmtID(0), stmt.Id)
}

// Execute a prepared statement to create a table.
func TestGateway_Exec_CreateTable(t *testing.T) {
	gateway := newGateway(t)
	defer gateway.Stop()

	// Create a new connection
	conn, err := gateway.Connect(context.Background(), newDatabase())
	require.NoError(t, err)

	// Prepare a new CREATE TABLE statement
	stmt, err := gateway.Prepare(context.Background(), newCreateTable(conn))
	require.NoError(t, err)

	// Bound the statement
	boundStmt, err := sql.NewBoundStmt(stmt, nil)
	require.NoError(t, err)

	// Execute the statement.
	result, err := gateway.Exec(context.Background(), boundStmt)
	require.NoError(t, err)

	assert.Equal(t, int64(0), result.LastInsertId)
	assert.Equal(t, int64(0), result.RowsAffected)
}

// Execute a prepared statement to insert values into a table.
func TestGateway_Exec_Insert(t *testing.T) {
	gateway := newGateway(t)
	defer gateway.Stop()

	// Create a new connection
	conn, err := gateway.Connect(context.Background(), newDatabase())
	require.NoError(t, err)

	// Prepare a new CREATE TABLE statement
	stmt, err := gateway.Prepare(context.Background(), newCreateTable(conn))
	require.NoError(t, err)
	assert.Equal(t, int32(0), stmt.NumInput)

	// Bound the statement
	boundStmt, err := sql.NewBoundStmt(stmt, nil)
	require.NoError(t, err)

	// Execute the statement
	_, err = gateway.Exec(context.Background(), boundStmt)
	require.NoError(t, err)

	// Prepare a new INSERT statement
	stmt, err = gateway.Prepare(context.Background(), newInsert(conn))
	require.NoError(t, err)
	assert.Equal(t, int32(1), stmt.NumInput)

	// Bound the statement
	boundStmt, err = sql.NewBoundStmt(stmt, newNamedValues(int64(1)))
	require.NoError(t, err)

	// Execute the statement
	result, err := gateway.Exec(context.Background(), boundStmt)
	require.NoError(t, err)

	assert.Equal(t, int64(1), result.LastInsertId)
	assert.Equal(t, int64(1), result.RowsAffected)
}

// Execute a query using a prepared statement.
func TestGateway_Query(t *testing.T) {
	gateway := newGateway(t)
	defer gateway.Stop()

	// Create a new connection
	conn, err := gateway.Connect(context.Background(), newDatabase())
	require.NoError(t, err)

	// Prepare a new CREATE TABLE statement
	stmt, err := gateway.Prepare(context.Background(), newCreateTable(conn))
	require.NoError(t, err)

	// Bound the statement
	boundStmt, err := sql.NewBoundStmt(stmt, nil)
	require.NoError(t, err)

	_, err = gateway.Exec(context.Background(), boundStmt)
	require.NoError(t, err)

	// Prepare a new INSERT statement
	stmt, err = gateway.Prepare(context.Background(), newInsert(conn))
	require.NoError(t, err)

	// Bound the statement
	boundStmt, err = sql.NewBoundStmt(stmt, newNamedValues(int64(1)))
	require.NoError(t, err)

	_, err = gateway.Exec(context.Background(), boundStmt)
	require.NoError(t, err)

	// Prepare a new SELECT statement
	stmt, err = gateway.Prepare(context.Background(), newQuery(conn))
	require.NoError(t, err)

	// Bound the statement
	boundStmt, err = sql.NewBoundStmt(stmt, newNamedValues(int64(1)))
	require.NoError(t, err)

	// Execute the query statement
	rows, err := gateway.Query(context.Background(), boundStmt)
	require.NoError(t, err)

	assert.Equal(t, sql.RowsID(0), rows.Id)
	assert.Equal(t, []string{"n"}, rows.Columns)

	// Fetch the first row.
	row, err := gateway.Next(context.Background(), rows)
	require.NoError(t, err)

	columns, err := row.DriverColumns()
	require.NoError(t, err)

	assert.Equal(t, []driver.Value{int64(1)}, columns)

	// Check the type of the first column of the current row.
	column := sql.NewColumn(rows, 0)
	typ, err := gateway.ColumnTypeScanType(context.Background(), column)
	require.NoError(t, err)

	assert.Equal(t, reflect.TypeOf(int64(0)), typ.DriverType())

	// Fetch the second row, which should be empty and signal the end of
	// the result set.
	row, err = gateway.Next(context.Background(), rows)
	require.NoError(t, err)

	columns, err = row.DriverColumns()
	require.NoError(t, err)

	assert.Nil(t, columns)

	// Check the name of the type of the first column.
	typeName, err := gateway.ColumnTypeDatabaseTypeName(context.Background(), column)
	require.NoError(t, err)
	assert.Equal(t, "INT", typeName.Value)
}

// Execute a transaction then commit.
func TestGateway_Transaction_Commit(t *testing.T) {
	gateway := newGateway(t)
	defer gateway.Stop()

	// Create a new connection
	conn, err := gateway.Connect(context.Background(), newDatabase())
	require.NoError(t, err)

	// Start a new transaction.
	tx, err := gateway.Begin(context.Background(), conn)
	require.NoError(t, err)

	// Execute a CREATE TABLE statement and an INSERT one in one go.
	result, err := gateway.ExecSQL(context.Background(), newCreateTableAndInsert(conn))
	require.NoError(t, err)

	assert.Equal(t, int64(1), result.LastInsertId)
	assert.Equal(t, int64(1), result.RowsAffected)

	// Commit the transaction.
	token, err := gateway.Commit(context.Background(), tx)
	require.NoError(t, err)

	assert.Equal(t, uint64(0), token.Value)

	// Prepare a new SELECT statement
	stmt, err := gateway.Prepare(context.Background(), newQuery(conn))
	require.NoError(t, err)

	// Bound the statement
	boundStmt, err := sql.NewBoundStmt(stmt, newNamedValues(int64(1)))
	require.NoError(t, err)

	// Execute the query statement
	rows, err := gateway.Query(context.Background(), boundStmt)
	require.NoError(t, err)

	// Fetch the first row.
	row, err := gateway.Next(context.Background(), rows)
	require.NoError(t, err)

	columns, err := row.DriverColumns()
	require.NoError(t, err)

	assert.Equal(t, []driver.Value{int64(1)}, columns)

	// Fetch the second row, which should be empty and signal the end of
	// the result set.
	row, err = gateway.Next(context.Background(), rows)
	require.NoError(t, err)
}

// Execute a transaction then rollback.
func TestGateway_Transaction_Rollback(t *testing.T) {
	gateway := newGateway(t)
	defer gateway.Stop()

	// Create a new connection
	conn, err := gateway.Connect(context.Background(), newDatabase())
	require.NoError(t, err)

	// Start a new transaction.
	tx, err := gateway.Begin(context.Background(), conn)
	require.NoError(t, err)

	// Prepare a new CREATE TABLE statement
	stmt, err := gateway.Prepare(context.Background(), newCreateTable(conn))
	require.NoError(t, err)

	// Bound the statement
	boundStmt, err := sql.NewBoundStmt(stmt, nil)
	require.NoError(t, err)

	// Execute the statement
	_, err = gateway.Exec(context.Background(), boundStmt)
	require.NoError(t, err)

	// Rollback the ransaction.
	_, err = gateway.Rollback(context.Background(), tx)
	require.NoError(t, err)

	// Prepare a new SELECT statement, it should fail since the table does
	// not exists.
	stmt, err = gateway.Prepare(context.Background(), newQuery(conn))

	require.EqualError(t, err, "rpc error: code = Code(1000) desc = SQL driver failed: no such table: test")
	assert.Nil(t, stmt)

	status, ok := status.FromError(err)
	require.True(t, ok)

	details := status.Details()
	assert.Len(t, details, 1)

	failure := details[0].(*sql.Error)
	assert.Equal(t, int64(1), failure.Code)
	assert.Equal(t, int64(1), failure.ExtendedCode)
	assert.Equal(t, "no such table: test", failure.Message)
}

// Close a prepared statement.
func TestGateway_CloseStmt(t *testing.T) {
	gateway := newGateway(t)
	defer gateway.Stop()

	// Create a new connection
	conn, err := gateway.Connect(context.Background(), newDatabase())
	require.NoError(t, err)

	// Prepare a new SELECT statement.
	stmt, err := gateway.Prepare(context.Background(), sql.NewSQL(conn, "SELECT 1"))
	require.NoError(t, err)

	// Close the statement.
	_, err = gateway.CloseStmt(context.Background(), stmt)
	require.NoError(t, err)

	// After the statement is closed, it can't be used anymore.
	boundStmt, err := sql.NewBoundStmt(stmt, []driver.NamedValue{})
	rows, err := gateway.Query(context.Background(), boundStmt)
	assert.EqualError(t, err, "rpc error: code = NotFound desc = no prepared statement with ID 0")
	assert.Nil(t, rows)
}

// Close a Rows result set.
func TestGateway_CloseRows(t *testing.T) {
	gateway := newGateway(t)
	defer gateway.Stop()

	// Create a new connection
	conn, err := gateway.Connect(context.Background(), newDatabase())
	require.NoError(t, err)

	// Prepare a new SELECT statement.
	stmt, err := gateway.Prepare(context.Background(), sql.NewSQL(conn, "SELECT 1"))
	require.NoError(t, err)

	// Execute the statement.
	boundStmt, err := sql.NewBoundStmt(stmt, []driver.NamedValue{})
	rows, err := gateway.Query(context.Background(), boundStmt)
	require.NoError(t, err)

	// Close the result set.
	_, err = gateway.CloseRows(context.Background(), rows)
	require.NoError(t, err)

	// After the Rows result set is closed, it can't be used anymore.
	row, err := gateway.Next(context.Background(), rows)
	assert.EqualError(t, err, "rpc error: code = NotFound desc = no result set with ID 0")
	assert.Nil(t, row)
}

// Close a connection.
func TestGateway_Close(t *testing.T) {
	gateway := newGateway(t)
	defer gateway.Stop()

	// Create a new connection
	conn, err := gateway.Connect(context.Background(), newDatabase())
	require.NoError(t, err)

	// Close the connection
	_, err = gateway.Close(context.Background(), conn)
	require.NoError(t, err)

	// After the connection is closed, it can't be used anymore.
	stmt, err := gateway.Prepare(context.Background(), newQuery(conn))
	assert.EqualError(t, err, "rpc error: code = NotFound desc = no connection with ID 0")
	assert.Nil(t, stmt)
}

// Connections for which no heartbeat is received within the timeout are
// automatically closed.
func TestGateway_Close_AfterTimeout(t *testing.T) {
	config := gateway.Config{
		HeartbeatTimeout: 200 * time.Millisecond,
	}

	gateway := newGatewayWithConfig(t, config)
	defer gateway.Stop()

	// Create a new connection
	conn, err := gateway.Connect(context.Background(), newDatabase())
	require.NoError(t, err)

	// Start a new transaction.
	tx, err := gateway.Begin(context.Background(), conn)
	require.NoError(t, err)

	// Prepare a new CREATE TABLE statement
	stmt, err := gateway.Prepare(context.Background(), newCreateTable(conn))
	require.NoError(t, err)

	// Bound the statement
	boundStmt, err := sql.NewBoundStmt(stmt, nil)
	require.NoError(t, err)

	// Execute the statement
	_, err = gateway.Exec(context.Background(), boundStmt)
	require.NoError(t, err)

	// Prepare a new SELECT statement
	stmt, err = gateway.Prepare(context.Background(), newQuery(conn))
	require.NoError(t, err)

	// Bound the statement
	boundStmt, err = sql.NewBoundStmt(stmt, newNamedValues(int64(1)))
	require.NoError(t, err)

	// Execute the query statement
	rows, err := gateway.Query(context.Background(), boundStmt)
	require.NoError(t, err)

	// Wait a bit and see that the connection all its resources are gone.
	time.Sleep(400 * time.Millisecond)

	_, err = gateway.Close(context.Background(), conn)
	require.EqualError(t, err, "rpc error: code = NotFound desc = no connection with ID 0")

	_, err = gateway.Next(context.Background(), rows)
	require.EqualError(t, err, "rpc error: code = NotFound desc = no connection with ID 0")

	_, err = gateway.CloseStmt(context.Background(), stmt)
	require.EqualError(t, err, "rpc error: code = NotFound desc = no connection with ID 0")

	_, err = gateway.Commit(context.Background(), tx)
	require.EqualError(t, err, "rpc error: code = NotFound desc = no connection with ID 0")
}

// Create a new test gRPC SQL gateway using a SQLite driver as backend.
func newGateway(t *testing.T) *gateway.Gateway {
	t.Helper()

	config := gateway.Config{
		HeartbeatTimeout: 20 * time.Second,
	}

	return newGatewayWithConfig(t, config)
}

func newGatewayWithConfig(t *testing.T, config gateway.Config) *gateway.Gateway {
	t.Helper()

	logger := zaptest.NewLogger(t)
	driver := &sqlite3.SQLiteDriver{}
	gateway := gateway.New(driver, config, logger)

	// Register a client with ID 0.
	_, err := gateway.Register(context.Background(), sql.NewClient("0"))
	require.NoError(t, err)

	return gateway
}

// Create an sql.Database object set with a in-memory database name.
func newDatabase() *sql.Database {
	return sql.NewDatabase("0", ":memory:")
}

// Create an sql.SQL object for creating a test table.
// Prepare a new CREATE TABLE statement
func newCreateTable(conn *sql.Conn) *sql.SQL {
	return sql.NewSQL(conn, "CREATE TABLE test (n INT, UNIQUE(n))")
}

// Create an sql.SQL object for inserting a value into the test table.
func newInsert(conn *sql.Conn) *sql.SQL {
	return sql.NewSQL(conn, "INSERT INTO test VALUES(?)")
}

// Create an sql.BoundSQL object for creating a test table and inserting a
// value into it.
func newCreateTableAndInsert(conn *sql.Conn) *sql.BoundSQL {
	boundSQL, _ := sql.NewBoundSQL(
		conn,
		"CREATE TABLE test (n INT); INSERT INTO test VALUES(?)",
		newNamedValues(int64(1)),
	)
	return boundSQL
}

// Create an sql.SQL object for querying into the test table.
func newQuery(conn *sql.Conn) *sql.SQL {
	return sql.NewSQL(conn, "SELECT n FROM test WHERE n=?")
}

// Create a driver.NamedValue slice from the given driver.Value slice.
func newNamedValues(values ...driver.Value) []driver.NamedValue {
	namedValues := make([]driver.NamedValue, len(values))
	for i, value := range values {
		namedValues[i].Ordinal = i + 1
		namedValues[i].Value = value
	}
	return namedValues
}
