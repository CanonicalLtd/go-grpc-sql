package grpcsql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/CanonicalLtd/go-grpc-sql/internal/cluster"
	_ "github.com/CanonicalLtd/go-sqlite3" // For DefaultServerStore
	"github.com/pkg/errors"
)

// ServerStore is used by a gRPC SQL driver to get an initial list of candidate
// gRPC target names that it can dial in order to find a leader gRPC SQL server
// to use.
//
// Once connected, the driver periodically updates the targets in the store by
// querying the leader server about changes in the cluster (such as servers
// being added or removed).
type ServerStore cluster.ServerStore

// DefaultServerStore creates a new DatabaseServerStore using the given
// filename to open a SQLite database, with default names for the schema, table
// and column parameters.
//
// It also creates the table if it doesn't exist yet.
func DefaultServerStore(filename string) (ServerStore, error) {
	// Open the database.
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database")
	}

	// Multiple connections don't work with in-memory databases.
	if filename == ":memory:" {
		db.SetMaxOpenConns(1)
	}

	// Create the servers table if it does not exist yet.
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS servers (target TEXT, UNIQUE(target))")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create servers table")
	}

	store := NewDatabaseServerStore(db, "main", "servers", "target")

	return store, nil
}

// DatabaseServerStore persists the list of target gRPC SQL servers in a SQL
// table.
type DatabaseServerStore struct {
	db     *sql.DB // Database handle to use.
	schema string  // Name of the schema holding the targets table.
	table  string  // Name of the targets table.
	column string  // Column name in the targets table holding the server address.
}

// NewDatabaseServerStore creates a new DatabaseServerStore.
func NewDatabaseServerStore(db *sql.DB, schema, table, column string) ServerStore {
	return &DatabaseServerStore{
		db:     db,
		schema: schema,
		table:  table,
		column: column,
	}
}

// Get the current targets.
func (d *DatabaseServerStore) Get(ctx context.Context) ([]string, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	query := fmt.Sprintf("SELECT %s FROM %s.%s", d.column, d.schema, d.table)
	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query targets table")
	}
	defer rows.Close()

	targets := make([]string, 0)
	for rows.Next() {
		var target string
		err := rows.Scan(&target)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch target row")
		}
		targets = append(targets, target)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "result set failure")
	}

	return targets, nil
}

// Set the targets.
func (d *DatabaseServerStore) Set(ctx context.Context, targets []string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	query := fmt.Sprintf("DELETE FROM %s.%s", d.schema, d.table)
	if _, err := tx.ExecContext(ctx, query); err != nil {
		tx.Rollback()
		return errors.Wrap(err, "failed to delete existing target rows")
	}

	query = fmt.Sprintf("INSERT INTO %s.%s(%s) VALUES (?)", d.schema, d.table, d.column)
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "failed to prepare insert statement")
	}
	defer stmt.Close()

	for _, target := range targets {
		if _, err := stmt.ExecContext(ctx, target); err != nil {
			tx.Rollback()
			return errors.Wrapf(err, "failed to insert target %s", target)
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

// NewInmemServerStore creates ServerStore which stores its data in-memory.
var NewInmemServerStore = cluster.NewInmemServerStore
