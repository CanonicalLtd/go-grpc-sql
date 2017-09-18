package protocol_test

import (
	"testing"
	"time"

	"github.com/mpvl/subtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CanonicalLtd/go-grpc-sql/internal/protocol"
)

// Marshal statements with their arguments.
func TestNewStatement(t *testing.T) {
	cases := []struct {
		title string
		args  []interface{}
	}{
		{`with string arg`, []interface{}{"hi"}},
		{`with integer arg`, []interface{}{int64(123)}},
		{`with float arg`, []interface{}{float64(0.123)}},
		{`with bool arg`, []interface{}{true}},
		{`with time arg`, []interface{}{time.Unix(12345, 0)}},
		{`with NULL arg`, []interface{}{nil}},
		{`with multiple args`, []interface{}{"hi", int64(123), float64(0.123), nil, true}},
	}
	for _, c := range cases {
		subtest.Run(t, c.title, func(t *testing.T) {
			stmt, err := protocol.NewStatement("SQL TEXT", c.args)
			require.NoError(t, err)

			assert.Equal(t, "SQL TEXT", stmt.Text)

			args, err := stmt.UnmarshalArgs()
			require.NoError(t, err)

			assert.Equal(t, c.args, args)
		})
	}
}

// Test failure modes when creating a new Statement object.
func TestNewStatement_Error(t *testing.T) {
	cases := []struct {
		title string
		args  []interface{}
		err   string
	}{
		{
			`invalid argument type`,
			[]interface{}{int32(123)},
			"cannot marshal object 0 (123): invalid type int32",
		},
	}
	for _, c := range cases {
		subtest.Run(t, c.title, func(t *testing.T) {
			_, err := protocol.NewStatement("SQL TEXT", c.args)
			assert.EqualError(t, err, c.err)
		})
	}
}

// Test failure modes when unmarshalling statement arguments.
func TestStatement_UnmarshalArgsError(t *testing.T) {
	cases := []struct {
		title string
		arg   protocol.Value
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
			stmt := &protocol.Statement{Args: []*protocol.Value{&c.arg}}
			_, err := stmt.UnmarshalArgs()
			assert.EqualError(t, err, c.err)
		})
	}

}
