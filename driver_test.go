package grpcsql_test

import (
	"database/sql"
	"testing"

	"github.com/CanonicalLtd/go-grpc-sql"
	"github.com/mpvl/subtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A global driver object is registered under the name "grpc".
func TestDriver_Registration(t *testing.T) {
	conn, err := sql.Open("grpc", "foo")
	assert.NoError(t, err)
	assert.NoError(t, conn.Close())
}

// Open a new gRPC connection.
/*func TestDriver_Open(t *testing.T) {
	address, cleanup := newServer()
	defer cleanup()

	driver := &grpcsql.Driver{}
	_, err := driver.Open(address)
	assert.NoError(t, err)
}*/

// Possible failure modes of Driver.Open().
func TestDriver_OpenError(t *testing.T) {
	cases := []struct {
		title string
		name  string
		err   string
	}{
		{
			"invalid escape",
			"1.2.3.4?%gh&%ij",
			"failed to parse query params",
		},
		{
			"timeout parse error",
			"1.2.3.4:1234?timeout=1woop",
			"invalid timeout",
		},
		{
			"connection timeout",
			"1.2.3.4:1234?timeout=1ns",
			"gRPC connection failed",
		},
		{
			"invalid client certificate",
			"1.2.3.4?certfile=/dev/null",
			"failed to load client certificate",
		},
		{
			"non-existing root certificate",
			"1.2.3.4?rootcertfile=/this/path/does/not/exists",
			"failed to read root certificate",
		},
		{
			"invalid root certificate",
			"1.2.3.4?rootcertfile=/dev/null",
			"failed to parse root certificate",
		},
	}
	for _, c := range cases {
		subtest.Run(t, c.title, func(t *testing.T) {
			driver := &grpcsql.Driver{}
			_, err := driver.Open(c.name)
			require.NotNil(t, err)
			assert.Contains(t, err.Error(), c.err)
		})
	}
}
