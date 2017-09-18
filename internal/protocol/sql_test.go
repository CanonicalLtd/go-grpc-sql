package protocol_test

import (
	"database/sql/driver"
	"testing"
	"time"

	"github.com/mpvl/subtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CanonicalLtd/go-grpc-sql/internal/protocol"
)

// Marshal driver.Value slices.
func TestFromDriverValues(t *testing.T) {
	cases := []struct {
		title   string
		objects []driver.Value
	}{
		{`string`, []driver.Value{"hi"}},
		{`int64`, []driver.Value{int64(123)}},
		{`float64`, []driver.Value{float64(0.123)}},
		{`bool`, []driver.Value{true}},
		{`time`, []driver.Value{time.Unix(12345, 0)}},
		{`nil`, []driver.Value{nil}},
		{`multiple`, []driver.Value{"hi", int64(123), float64(0.123), nil, true}},
	}
	for _, c := range cases {
		subtest.Run(t, c.title, func(t *testing.T) {
			values, err := protocol.FromDriverValues(c.objects)
			require.NoError(t, err)

			objects, err := protocol.ToDriverValues(values)
			require.NoError(t, err)

			assert.Equal(t, c.objects, objects)
		})
	}
}

// Test failure modes when marshaling driver.Value objects.
func TestFromDriverValues_Error(t *testing.T) {
	cases := []struct {
		title   string
		objects []driver.Value
		err     string
	}{
		{
			`invalid argument type`,
			[]driver.Value{int32(123)},
			"cannot marshal object 0 (123): invalid type int32",
		},
	}
	for _, c := range cases {
		subtest.Run(t, c.title, func(t *testing.T) {
			_, err := protocol.FromDriverValues(c.objects)
			assert.EqualError(t, err, c.err)
		})
	}
}

// Test failure modes when unmarshalling driver.Value objects
func TestToDriverValues_Error(t *testing.T) {
	cases := []struct {
		title string
		value protocol.Value
		err   string
	}{
		{
			`invalid code`,
			protocol.Value{Code: 666},
			"cannot unmarshal value 0: invalid value type code 666",
		},
		{
			`garbage data`,
			protocol.Value{Code: protocol.ValueCode_INT64, Data: []byte("foo")},
			"cannot unmarshal value 0: proto: illegal wireType 6",
		},
	}
	for _, c := range cases {
		subtest.Run(t, c.title, func(t *testing.T) {
			_, err := protocol.ToDriverValues([]*protocol.Value{&c.value})
			assert.EqualError(t, err, c.err)
		})
	}

}
