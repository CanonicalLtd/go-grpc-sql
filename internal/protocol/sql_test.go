package protocol_test

import (
	"database/sql/driver"
	"reflect"
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

// Marshal reflect.Type into ValueCode.
func TestToValueCode(t *testing.T) {
	cases := []struct {
		title string
		typ   reflect.Type
		code  protocol.ValueCode
	}{
		{`string`, reflect.TypeOf(""), protocol.ValueCode_STRING},
		{`int64`, reflect.TypeOf(int64(0)), protocol.ValueCode_INT64},
		{`float64`, reflect.TypeOf(float64(0)), protocol.ValueCode_FLOAT64},
		{`bool`, reflect.TypeOf(false), protocol.ValueCode_BOOL},
		{`bytes`, reflect.TypeOf(byte(0)), protocol.ValueCode_BYTES},
		{`time`, reflect.TypeOf(time.Time{}), protocol.ValueCode_TIME},
		{`nil`, reflect.TypeOf(nil), protocol.ValueCode_NULL},
	}
	for _, c := range cases {
		subtest.Run(t, c.title, func(t *testing.T) {
			code := protocol.ToValueCode(c.typ)
			assert.Equal(t, c.code, code)
		})
	}
}

// Unmarshal ValueCode into reflect.Type.
func TestFromValueCode(t *testing.T) {
	cases := []struct {
		title string
		typ   reflect.Type
		code  protocol.ValueCode
	}{
		{`string`, reflect.TypeOf(""), protocol.ValueCode_STRING},
		{`int64`, reflect.TypeOf(int64(0)), protocol.ValueCode_INT64},
		{`float64`, reflect.TypeOf(float64(0)), protocol.ValueCode_FLOAT64},
		{`bool`, reflect.TypeOf(false), protocol.ValueCode_BOOL},
		{`bytes`, reflect.TypeOf(byte(0)), protocol.ValueCode_BYTES},
		{`time`, reflect.TypeOf(time.Time{}), protocol.ValueCode_TIME},
		{`nil`, reflect.TypeOf(nil), protocol.ValueCode_NULL},
	}
	for _, c := range cases {
		subtest.Run(t, c.title, func(t *testing.T) {
			typ := protocol.FromValueCode(c.code)
			assert.Equal(t, c.typ, typ)
		})
	}
}
