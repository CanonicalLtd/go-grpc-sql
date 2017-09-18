package grpcsql

// Tx is a transaction.
type Tx struct {
}

// Commit the transaction.
func (tx *Tx) Commit() error {
	return nil
}

// Rollback the transaction.
func (tx *Tx) Rollback() error {
	return nil
}
