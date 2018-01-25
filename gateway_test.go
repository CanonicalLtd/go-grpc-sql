package grpcsql_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/CanonicalLtd/go-grpc-sql"
	"github.com/CanonicalLtd/go-grpc-sql/internal/protocol"
	"github.com/CanonicalLtd/go-sqlite3"
	"github.com/mpvl/subtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func TestGateway_ConnError(t *testing.T) {
	cases := []struct {
		title    string
		requests []*protocol.Request // Sequence of requests to submit
		err      string              // Error message
	}{
		{
			`invalid request code`,
			[]*protocol.Request{{Code: 666}},
			"invalid request code 666",
		},
		{
			`invalid request data`,
			[]*protocol.Request{{Code: protocol.RequestCode_OPEN, Data: []byte("x")}},
			"request parse error",
		},
		{
			`non-OPEN first request`,
			[]*protocol.Request{protocol.NewRequestPrepare("close")},
			"expected OPEN request, got PREPARE",
		},
		{
			`non-OPEN first request`,
			[]*protocol.Request{protocol.NewRequestOpen("/etc")},
			"could not open driver connection",
		},
	}

	for _, c := range cases {
		subtest.Run(t, c.title, func(t *testing.T) {
			client, cleanup := newGatewayClient()
			defer cleanup()

			var err error
			for i, request := range c.requests {
				err = client.Send(request)
				require.NoError(t, err)
				_, err = client.Recv()
				if i < len(c.requests)-1 {
					require.NoError(t, err)
				}
			}
			require.NotNil(t, err)
			assert.Contains(t, err.Error(), c.err)
		})
	}
}

// Create a new protocol.SQL_ConnClient stream connected to a Gateway backed by
// a SQLite driver.
func newGatewayClient() (protocol.SQL_ConnClient, func()) {
	server, address := newGatewayServer()

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %v", err))
	}

	client := protocol.NewSQLClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(2*time.Second))
	cleanup := func() {
		cancel()
		conn.Close()
		server.Stop()
	}

	connClient, err := client.Conn(ctx)
	if err != nil {
		panic(fmt.Errorf("gRPC conn method failed: %v", err))
	}

	return connClient, cleanup
}

// Create a new test gRPC server attached to a grpc-sql gateway backed by a
// SQLite driver.
//
// Return the newly created server and its network address.
func newGatewayServer() (*grpc.Server, string) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(fmt.Sprintf("failed to create listener: %v", err))
	}
	server := grpcsql.NewServer(&sqlite3.SQLiteDriver{})
	go server.Serve(listener)
	return server, listener.Addr().String()
}
