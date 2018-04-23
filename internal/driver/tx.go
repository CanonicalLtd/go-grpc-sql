package driver

import (
	"context"
	"time"

	"github.com/CanonicalLtd/go-grpc-sql/internal/sql"
)

// Tx is a transaction.
type Tx struct {
	client    sql.GatewayClient // gRPC SQL gateway client to use.
	conn      *sql.Conn         // Info about the connection opened on the gateway.
	tx        *sql.Tx           // Info about the transaction on the gateway.
	timeout   time.Duration     // Taken from the driver's Config.MethodTimeout
	connector Connector         // Used to create a new gateway client, only to recover a commit.
}

func newTx(client sql.GatewayClient, conn *sql.Conn, tx *sql.Tx, timeout time.Duration, connector Connector) *Tx {
	return &Tx{
		client:    client,
		conn:      conn,
		tx:        tx,
		timeout:   timeout,
		connector: connector,
	}
}

// Commit the transaction.
func (tx *Tx) Commit() error {
	ctx, cancel := context.WithTimeout(context.Background(), tx.timeout)
	defer cancel()

	token, err := tx.client.Commit(ctx, tx.tx)
	if err != nil {
		errMethod := newErrorFromMethod(err)
		if token == nil || token.Value == 0 {
			return errMethod
		}
		// The commit failed but the gateway's driver returned a token
		// that can be used to try to recover it. Let's grab a client
		// connected to the next leader and try.
		client, err := tx.connector(ctx)
		if err != nil {
			// Just return the original error
			return errMethod
		}
		_, err = client.Recover(ctx, token)
		if err != nil {
			// TODO: Retry a few times in case the error is due to
			//       the leader being unstable.
			//
			// Just return the original error
			return errMethod
		}
		return nil
	}

	return nil
}

// Rollback the transaction.
func (tx *Tx) Rollback() error {
	ctx, cancel := context.WithTimeout(context.Background(), tx.timeout)
	defer cancel()

	_, err := tx.client.Rollback(ctx, tx.tx)
	if err != nil {
		return newErrorFromMethod(err)
	}

	return nil
}
