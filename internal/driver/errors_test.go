package driver

import (
	"testing"

	"github.com/CanonicalLtd/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestNewSQLiteError(t *testing.T) {
	err := newError(1, 2, "hello")
	assert.Equal(t, sqlite3.ErrNo(1), err.Code)
	assert.Equal(t, sqlite3.ErrNoExtended(2), err.ExtendedCode)
	assert.EqualError(t, err, "hello")
}
