package driver

import (
	"context"
	"database/sql/driver"
	"time"

	"github.com/CanonicalLtd/go-grpc-sql/internal/sql"
)

// Conn drives a connection to a gRPC SQL gateway.
type Conn struct {
	client    sql.GatewayClient // gRPC SQL gateway client to use.
	conn      *sql.Conn         // Info about the connection opened on the gateway.
	timeout   time.Duration     // Taken from the driver's Config.MethodTimeout
	connector Connector         // Used to create a new gateway client, only to recover a commit.
}

func newConn(client sql.GatewayClient, conn *sql.Conn, timeout time.Duration, connector Connector) *Conn {
	return &Conn{
		client:    client,
		conn:      conn,
		timeout:   timeout,
		connector: connector,
	}
}

// PrepareContext returns a prepared statement, bound to this connection.
// context is for the preparation of the statement,
// it must not store the context within the statement itself.
func (c *Conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	stmt, err := c.client.Prepare(ctx, sql.NewSQL(c.conn, query))
	if err != nil {
		return nil, newErrorFromMethod(err)
	}
	return newStmt(c.client, c.conn, stmt), nil
}

// Prepare returns a prepared statement, bound to this connection.
func (c *Conn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}

// Close invalidates and potentially stops any current
// prepared statements and transactions, marking this
// connection as no longer in use.
func (c *Conn) Close() error {
	_, err := c.client.Close(context.Background(), c.conn)
	if err != nil {
		return newErrorFromMethod(err)
	}
	return nil
}

// BeginTx starts and returns a new transaction.  If the context is canceled by
// the user the sql package will call Tx.Rollback before discarding and closing
// the connection.
//
// This must check opts.Isolation to determine if there is a set isolation
// level. If the driver does not support a non-default level and one is set or
// if there is a non-default isolation level that is not supported, an error
// must be returned.
//
// This must also check opts.ReadOnly to determine if the read-only value is
// true to either set the read-only transaction property if supported or return
// an error if it is not supported.
func (c *Conn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	tx, err := c.client.Begin(ctx, c.conn)
	if err != nil {
		return nil, newErrorFromMethod(err)
	}
	return newTx(c.client, c.conn, tx, c.timeout, c.connector), nil
}

// Begin starts and returns a new transaction.
//
// Deprecated: Drivers should implement ConnBeginTx instead (or additionally).
func (c *Conn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

// ExecContext is an optional interface that may be implemented by a Conn.
func (c *Conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	boundSQL, err := sql.NewBoundSQL(c.conn, query, args)
	if err != nil {
		return nil, newErrorMisuse(err)
	}

	result, err := c.client.ExecSQL(ctx, boundSQL)
	if err != nil {
		return nil, newErrorFromMethod(err)
	}
	return newResult(result.LastInsertId, result.RowsAffected), nil
}

// Exec is an optional interface that may be implemented by a Conn.
func (c *Conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return c.ExecContext(context.Background(), query, valuesToNamedValues(args))
}
