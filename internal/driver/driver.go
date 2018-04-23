package driver

import (
	"context"
	"database/sql/driver"

	"github.com/CanonicalLtd/go-grpc-sql/internal/sql"
	"github.com/CanonicalLtd/go-sqlite3"
	"go.uber.org/zap"
)

// Driver implements the database/sql/driver interface and executes the
// relevant statements over gRPC.
type Driver struct {
	id        string      // Client ID of this driver.
	connector Connector   // Used to create a gateway client
	config    Config      // Configuration parameters
	logger    *zap.Logger // Logger
}

// Connector is function returning a GatewayClient pointing to the current
// leader of a cluster of gRPC SQL gateways.
type Connector func(context.Context) (sql.GatewayClient, error)

// New creates a new Driver instance.
func New(id string, connector Connector, config Config, logger *zap.Logger) *Driver {
	return &Driver{
		id:        id,
		connector: connector,
		config:    config,
		logger:    logger,
	}
}

// Open a new connection against a gRPC SQL gateway.
//
// To create the gRPC SQL gateway client, the factory passed to NewDriver()
// will used.
//
// The given database name must be one that the driver attached to the remote
// gateway can understand.
func (d *Driver) Open(name string) (driver.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.config.ConnectTimeout)
	defer cancel()

	client, err := d.connector(ctx)
	if err != nil {
		return nil, newError(sqlite3.ErrCantOpen, 0, err.Error())
	}

	conn, err := client.Connect(ctx, sql.NewDatabase(d.id, name))
	if err != nil {
		return nil, newErrorFromMethod(err)
	}

	return newConn(client, conn, d.config.MethodTimeout, d.connector), nil
}
