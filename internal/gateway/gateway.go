package gateway

import (
	"context"
	"database/sql/driver"
	"io"
	"math"
	"sync"
	"time"

	"github.com/CanonicalLtd/go-grpc-sql/internal/cluster"
	"github.com/CanonicalLtd/go-grpc-sql/internal/sql"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Gateway implements the gRPC SQL gateway service interface.
type Gateway struct {
	driver  driver.Driver        // Underlying driver.
	config  Config               // Configuration parameters.
	logger  *zap.Logger          // Logger.
	stopCh  chan struct{}        // Stop internal goroutines.
	mu      sync.RWMutex         // Serialize access to the fields below here.
	clients map[string]*client   // All active clients, index by ID.
	conns   map[sql.ConnID]*conn // All open connections, indexed by ID.
}

// New creates a new Gateway.
func New(driver driver.Driver, config Config, logger *zap.Logger) *Gateway {
	gateway := &Gateway{
		driver:  driver,
		config:  config,
		logger:  logger.With(zap.Namespace("gateway")),
		stopCh:  make(chan struct{}),
		clients: make(map[string]*client),
		conns:   make(map[sql.ConnID]*conn),
	}
	go gateway.closeDeadConns()
	return gateway
}

// Stop the gateway.
func (g *Gateway) Stop() {
	close(g.stopCh)
}

// Leader returns the current leader of the cluster this gRPC SQL gateway is
// part of.
func (g *Gateway) Leader(ctx context.Context, empty *sql.Empty) (*sql.Server, error) {
	g.logger.Debug("leader")

	cluster, ok := g.driver.(cluster.DriverCluster)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "driver does not implement Leader()")
	}

	return sql.NewServer(cluster.Leader()), nil
}

// Register a new client, which can then start opening connections. Return
// the timeout after which all client connections will be killed if no
// heartbeat is received from the client.
//
// This API is idempotent and it's okay to call it again for the same client.
func (g *Gateway) Register(ctx context.Context, client *sql.Client) (*sql.Duration, error) {
	g.logger.Debug("register",
		zap.String("client", client.Id),
	)

	g.mu.Lock()
	defer g.mu.Unlock()

	if _, ok := g.clients[client.Id]; !ok {
		g.clients[client.Id] = newClient()
	}

	return sql.NewDuration(g.config.HeartbeatTimeout), nil
}

// Heartbeat is a bidirectional stream that should be kept open by clients in
// order to refresh their heartbeat and receive updates about the servers that
// are currently part of the cluster.
func (g *Gateway) Heartbeat(stream sql.Gateway_HeartbeatServer) error {
	for {
		if err := g.heartbeatRoundtrip(stream); err != nil {
			return err
		}
	}
}

// Perform a single heartbeat rountrip.
func (g *Gateway) heartbeatRoundtrip(stream sql.Gateway_HeartbeatServer) error {
	// Receive a heartbeat request from the stream.
	client, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	logger := g.logger.With(zap.String("client", client.Id))
	logger.Debug("heartbeat")

	// Take a read lock while we work with internal state, but release it
	// before the final stream.Send() call for responding to the heartbeat.
	g.mu.RLock()

	// Check that client is (still) registered.
	c, err := g.lookupClient(client.Id)
	if err != nil {
		g.mu.RUnlock()
		return err
	}

	// Get information about other servers in the cluster.
	servers, err := driverClusterServers(g.driver)
	if err != nil {
		g.mu.RUnlock()
		return err
	}

	// Update the heartbeat timestamp of this client.
	c.Heartbeat()

	g.mu.RUnlock()

	// Send the heartbeat response to the stream.
	cluster := sql.NewCluster(servers)
	if err := stream.Send(cluster); err != nil {
		return err
	}

	return nil
}

// Connect to a database exposed over this gRPC SQL gateway.
func (g *Gateway) Connect(ctx context.Context, database *sql.Database) (*sql.Conn, error) {
	g.logger.Info("connect",
		zap.String("client", database.ClientId),
		zap.String("name", database.Name),
	)

	g.mu.Lock()
	defer g.mu.Unlock()

	client, err := g.lookupClient(database.ClientId)
	if err != nil {
		return nil, err
	}

	if len(g.conns) >= math.MaxUint32 {
		return nil, status.Error(codes.ResourceExhausted, "too many open connections")
	}

	driverConn, err := g.driver.Open(database.Name)
	if err != nil {
		return nil, driverErrorToStatus(err)
	}

	id := sql.ConnID(len(g.conns))

	g.conns[id] = newConn(client, id, driverConn)

	return sql.NewConn(id), nil
}

// Prepare a new statement in a transaction..
func (g *Gateway) Prepare(ctx context.Context, SQL *sql.SQL) (*sql.Stmt, error) {
	g.logger.Debug("prepare",
		zap.Uint32("conn", uint32(SQL.ConnId)),
		zap.String("text", SQL.Text),
	)

	g.mu.RLock()
	defer g.mu.RUnlock()

	c, err := g.lookupConn(SQL)
	if err != nil {
		return nil, err
	}

	return c.Prepare(ctx, SQL.Text)
}

// Begin a new transaction on the given connection.
func (g *Gateway) Begin(ctx context.Context, conn *sql.Conn) (*sql.Tx, error) {
	g.logger.Debug("begin",
		zap.Uint32("conn", uint32(conn.Id)),
	)

	g.mu.RLock()
	defer g.mu.RUnlock()

	c, err := g.lookupConn(conn)
	if err != nil {
		return nil, err
	}

	return c.Begin(ctx)
}

// Exec executes a bound statement and returns its result.
func (g *Gateway) Exec(ctx context.Context, boundStmt *sql.BoundStmt) (*sql.Result, error) {
	g.logger.Debug("exec",
		zap.Uint32("conn", uint32(boundStmt.ConnId)),
		zap.Uint32("stmt", uint32(boundStmt.StmtId)),
	)

	g.mu.RLock()
	defer g.mu.RUnlock()

	c, err := g.lookupConn(boundStmt)
	if err != nil {
		return nil, err
	}

	namedValues, err := boundStmt.DriverNamedValues()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return c.Exec(ctx, boundStmt.StmtId, namedValues)
}

// ExecSQL prepares and executes a statement with parameters, and
// returns a Result object.
func (g *Gateway) ExecSQL(ctx context.Context, boundSQL *sql.BoundSQL) (*sql.Result, error) {
	g.logger.Debug("exec sql",
		zap.Uint32("conn", uint32(boundSQL.ConnId)),
		zap.String("text", boundSQL.Text),
	)

	g.mu.RLock()
	defer g.mu.RUnlock()

	c, err := g.lookupConn(boundSQL)
	if err != nil {
		return nil, err
	}

	namedValues, err := boundSQL.DriverNamedValues()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return c.ExecSQL(ctx, boundSQL.Text, namedValues)
}

// Query executes a bound statement and returns its rows.
func (g *Gateway) Query(ctx context.Context, boundStmt *sql.BoundStmt) (*sql.Rows, error) {
	g.logger.Debug("query",
		zap.Uint32("conn", uint32(boundStmt.ConnId)),
		zap.Uint32("stmt", uint32(boundStmt.StmtId)),
	)

	g.mu.RLock()
	defer g.mu.RUnlock()

	c, err := g.lookupConn(boundStmt)
	if err != nil {
		return nil, err
	}

	namedValues, err := boundStmt.DriverNamedValues()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return c.Query(ctx, boundStmt.StmtId, namedValues)
}

// Next returns the next row in a result set.
func (g *Gateway) Next(ctx context.Context, rows *sql.Rows) (*sql.Row, error) {
	g.logger.Debug("next",
		zap.Uint32("conn", uint32(rows.ConnId)),
		zap.Uint32("rows", uint32(rows.Id)),
	)

	g.mu.RLock()
	defer g.mu.RUnlock()

	c, err := g.lookupConn(rows)
	if err != nil {
		return nil, err
	}

	return c.Next(ctx, rows.Id)
}

// ColumnTypeScanType returns the type of certain column for the current row
// pointed by a Rows object.
func (g *Gateway) ColumnTypeScanType(ctx context.Context, column *sql.Column) (*sql.Type, error) {
	g.logger.Debug("column type scan type",
		zap.Uint32("conn", uint32(column.ConnId)),
		zap.Uint32("rows", uint32(column.RowsId)),
		zap.Int32("index", column.Index),
	)

	g.mu.RLock()
	defer g.mu.RUnlock()

	c, err := g.lookupConn(column)
	if err != nil {
		return nil, err
	}

	return c.ColumnTypeScanType(ctx, column.RowsId, int(column.Index))
}

// ColumnTypeDatabaseTypeName returns the name of the type of certain column in a Rows
// object.
func (g *Gateway) ColumnTypeDatabaseTypeName(ctx context.Context, column *sql.Column) (*sql.TypeName, error) {
	g.logger.Debug("column type database type name",
		zap.Uint32("conn", uint32(column.ConnId)),
		zap.Uint32("rows", uint32(column.RowsId)),
		zap.Int32("index", column.Index),
	)

	g.mu.RLock()
	defer g.mu.RUnlock()

	c, err := g.lookupConn(column)
	if err != nil {
		return nil, err
	}

	return c.ColumnTypeDatabaseTypeName(ctx, column.RowsId, int(column.Index))
}

// Commit a transaction.
func (g *Gateway) Commit(ctx context.Context, tx *sql.Tx) (*sql.Token, error) {
	g.logger.Debug("commit",
		zap.Uint32("conn", uint32(tx.ConnId)),
		zap.Uint32("tx", uint32(tx.Id)),
	)

	g.mu.RLock()
	defer g.mu.RUnlock()

	c, err := g.lookupConn(tx)
	if err != nil {
		return nil, err
	}

	return c.Commit(ctx, tx.Id)
}

// Rollback a transaction.
func (g *Gateway) Rollback(ctx context.Context, tx *sql.Tx) (*sql.Empty, error) {
	g.logger.Debug("rollback",
		zap.Uint32("conn", uint32(tx.ConnId)),
		zap.Uint32("tx", uint32(tx.Id)),
	)

	g.mu.RLock()
	defer g.mu.RUnlock()

	c, err := g.lookupConn(tx)
	if err != nil {
		return nil, err
	}

	return c.Rollback(ctx, tx.Id)
}

// Recover makes an attempt to recover a possibly lost transaction.
func (g *Gateway) Recover(ctx context.Context, token *sql.Token) (*sql.Empty, error) {
	g.logger.Debug("recover",
		zap.Uint64("token", token.Value),
	)

	cluster, ok := g.driver.(cluster.DriverCluster)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "driver does not implement Recover()")
	}

	if err := cluster.Recover(token.Value); err != nil {
		return nil, driverErrorToStatus(err)
	}

	return sql.NewEmpty(), nil
}

// CloseStmt closes a prepared statement.
func (g *Gateway) CloseStmt(ctx context.Context, stmt *sql.Stmt) (*sql.Empty, error) {
	g.logger.Debug("close stmt",
		zap.Uint32("conn", uint32(stmt.ConnId)),
		zap.Uint32("tx", uint32(stmt.Id)),
	)

	g.mu.RLock()
	defer g.mu.RUnlock()

	c, err := g.lookupConn(stmt)
	if err != nil {
		return nil, err
	}

	return c.CloseStmt(ctx, stmt.Id)
}

// CloseRows closes a Rows result set.
func (g *Gateway) CloseRows(ctx context.Context, rows *sql.Rows) (*sql.Empty, error) {
	g.logger.Debug("close rows",
		zap.Uint32("conn", uint32(rows.ConnId)),
		zap.Uint32("tx", uint32(rows.Id)),
	)

	g.mu.RLock()
	defer g.mu.RUnlock()

	c, err := g.lookupConn(rows)
	if err != nil {
		return nil, err
	}

	return c.CloseRows(ctx, rows.Id)
}

// Close closes a connection.
func (g *Gateway) Close(ctx context.Context, conn *sql.Conn) (*sql.Empty, error) {
	g.logger.Debug("close",
		zap.Uint32("conn", uint32(conn.Id)),
	)

	g.mu.Lock()
	defer g.mu.Unlock()

	c, err := g.lookupConn(conn)
	if err != nil {
		return nil, err
	}

	defer delete(g.conns, conn.Id)

	return c.Close()
}

// Lookup a registered client matching the given ID.
//
// An error is returned is no matching client is found.
func (g *Gateway) lookupClient(id string) (*client, error) {
	client, ok := g.clients[id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no client with ID %d", id)
	}

	return client, nil
}

// Lookup a registered connection matching the ID returned by the GetConnId()
// method of an object implementing sql.WithConn (e.g. sql.Conn, sql.Stmt,
// sql.Tx and sql.Rows).
//
// An error is returned is no matching connection is found.
func (g *Gateway) lookupConn(withConn sql.WithConn) (*conn, error) {
	id := withConn.GetConnId()
	conn, ok := g.conns[id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no connection with ID %d", id)
	}

	return conn, nil
}

// Close all connections that for which no heartbeat has been received.
func (g *Gateway) closeDeadConns() {
	for {
		select {
		case <-g.stopCh:
			return
		case <-time.After(g.config.HeartbeatTimeout):
			g.mu.Lock()

			// Look for dead connections.
			for id, conn := range g.conns {
				if conn.Dead(g.config.HeartbeatTimeout) {
					g.logger.Info("timeout", zap.Uint32("conn", uint32(id)))
					conn.Close()
					delete(g.conns, id)
				}
			}

			// Look for dead clients.
			for id, client := range g.clients {
				if client.Dead(g.config.HeartbeatTimeout) {
					g.logger.Info("timeout", zap.String("client", id))
					delete(g.clients, id)
				}
			}

			g.mu.Unlock()
		}
	}
}
