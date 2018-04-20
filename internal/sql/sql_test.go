package sql

import (
	"database/sql/driver"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Convert a driver.NamedValue to an rpc.Value and back.
func TestFromAndToNamedValue(t *testing.T) {
	cases := []struct {
		title            string
		driverNamedValue driver.NamedValue
	}{
		{"string", driver.NamedValue{Ordinal: 1, Value: "hi"}},
		{"int64", driver.NamedValue{Ordinal: 1, Value: int64(123)}},
		{"float64", driver.NamedValue{Ordinal: 1, Value: float64(0.123)}},
		{"bool", driver.NamedValue{Ordinal: 1, Value: true}},
		{"time", driver.NamedValue{Ordinal: 1, Value: time.Unix(12345, 0)}},
		{"nil", driver.NamedValue{Ordinal: 1, Value: nil}},
	}
	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			namedValue, err := fromDriverNamedValue(c.driverNamedValue)
			require.NoError(t, err)

			driverNamedValue, err := toDriverNamedValue(namedValue)
			require.NoError(t, err)

			assert.Equal(t, c.driverNamedValue, driverNamedValue)
		})
	}
}
