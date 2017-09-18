package grpcsql_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

// A global driver object is registered under the name "grpc".
func TestDriver_Registration(t *testing.T) {
	conn, err := sql.Open("grpc", "foo")
	assert.NoError(t, err)
	assert.NoError(t, conn.Close())
}
