package grpcsql

import (
	"github.com/CanonicalLtd/go-grpc-sql/internal/protocol"
)

// Tx is a transaction.
type Tx struct {
	conn *Conn
	id   int64
}

// Commit the transaction.
func (tx *Tx) Commit() error {
	_, err := tx.conn.exec(protocol.NewRequestCommit(tx.id))
	return err
}

// Rollback the transaction.
func (tx *Tx) Rollback() error {
	_, err := tx.conn.exec(protocol.NewRequestRollback(tx.id))
	return err
}
