package driver_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration(t *testing.T) {
	driver, cleanup := newDriver(t)
	defer cleanup()

	sql.Register("grpc", driver)

	db, err := sql.Open("grpc", ":memory:")
	require.NoError(t, err)

	tx, err := db.Begin()
	require.NoError(t, err)

	result, err := tx.Exec("CREATE TABLE test (n INTEGER); INSERT INTO test(n) VALUES (1)")
	require.NoError(t, err)

	result, err = tx.Exec("INSERT INTO test(n) VALUES (2)")
	require.NoError(t, err)

	lastInsertedID, err := result.LastInsertId()
	require.NoError(t, err)

	assert.Equal(t, int64(2), lastInsertedID)

	rowsAffected, err := result.RowsAffected()
	require.NoError(t, err)

	assert.Equal(t, int64(1), rowsAffected)

	rows, err := tx.Query("SELECT n FROM test ORDER BY n")
	require.NoError(t, err)

	types, err := rows.ColumnTypes()
	require.NoError(t, err)
	require.Len(t, types, 1)

	name := types[0].DatabaseTypeName()
	assert.Equal(t, "INTEGER", name)

	numbers := []int{}
	for rows.Next() {
		var n int
		err := rows.Scan(&n)
		require.NoError(t, err)
		numbers = append(numbers, n)
	}
	require.NoError(t, rows.Err())

	assert.Equal(t, []int{1, 2}, numbers)

	require.NoError(t, rows.Close())

	require.NoError(t, tx.Rollback())

	require.NoError(t, db.Close())
}
