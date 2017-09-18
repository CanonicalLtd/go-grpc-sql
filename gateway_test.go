package grpcsql_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CanonicalLtd/go-grpc-sql"
	"github.com/CanonicalLtd/go-grpc-sql/internal/protocol"
	"github.com/mattn/go-sqlite3"
	"github.com/mpvl/subtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
		{
			`invalid prepared query`,
			[]*protocol.Request{
				protocol.NewRequestOpen(":memory:"),
				protocol.NewRequestPrepare("foo"),
			},
			"failed to handle PREPARE request",
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

// Create a new protocol.SQL_ConnClient stream connected to a Gateway backed by a
// SQLite driver.
func newGatewayClient() (protocol.SQL_ConnClient, func()) {
	driver := &sqlite3.SQLiteDriver{}
	server := httptest.NewUnstartedServer(grpcsql.NewServer(driver))
	server.TLS = &tls.Config{NextProtos: []string{"h2"}}
	server.StartTLS()

	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	options := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(2*time.Second))
	conn, err := grpc.Dial(server.Listener.Addr().String(), options...)

	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %v", err))
	}

	client := protocol.NewSQLClient(conn)
	cleanup := func() {
		cancel()
		conn.Close()
		server.Close()
	}

	connClient, err := client.Conn(ctx)
	if err != nil {
		panic(fmt.Errorf("gRPC conn method failed: %v", err))
	}

	return connClient, cleanup
}
