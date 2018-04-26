package grpcsql_test

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CanonicalLtd/go-grpc-sql"
	"github.com/CanonicalLtd/go-grpc-sql/internal/gateway"
	"github.com/CanonicalLtd/go-grpc-sql/internal/sql"
	"github.com/CanonicalLtd/go-sqlite3"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

func TestHandler(t *testing.T) {
	return
	gateway := newGateway(t)
	defer gateway.Stop()

	w := &testingWriter{t}
	l := grpclog.NewLoggerV2WithVerbosity(w, ioutil.Discard, ioutil.Discard, 10)
	grpclog.SetLoggerV2(l)

	server := grpc.NewServer()
	sql.RegisterGatewayServer(server, gateway)

	mux := http.NewServeMux()
	mux.Handle("/sql.Gateway/", grpcsql.NewHandler(t, server))
	ts := httptest.NewUnstartedServer(mux)
	ts.TLS = &tls.Config{
		NextProtos: []string{"h2"},
	}
	listener, err := net.Listen("tcp", "127.0.0.1:8000")
	require.NoError(t, err)
	ts.Listener = listener
	ts.StartTLS()

	certificate, err := x509.ParseCertificate(ts.TLS.Certificates[0].Certificate[0])
	certpool := x509.NewCertPool()
	certpool.AddCert(certificate)
	require.NoError(t, err)
	config := &tls.Config{
		RootCAs: certpool,
	}
	// dialer := func(address string, timeout time.Duration) (net.Conn, error) {
	// 	ci, err := net.DialTimeout("tcp", address, timeout)
	// 	t.Logf("dial: %v", err)
	// 	return ci, err
	// }
	ctx, cancel := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancel()
	conn, err := grpc.DialContext(
		ctx,
		ts.Listener.Addr().String(),
		grpc.WithBackoffMaxDelay(100*time.Millisecond),
		grpc.WithDefaultCallOptions(grpc.FailFast(false)),
		//grpc.WithDialer(dialer),
		grpc.WithTransportCredentials(credentials.NewTLS(config)),
	)
	require.NoError(t, err)
	go func() {
		for {
			t.Logf("CONN STATE %s", conn.GetState())
			time.Sleep(250 * time.Millisecond)
		}
	}()

	client := sql.NewGatewayClient(conn)
	go func() {
		time.Sleep(50 * time.Millisecond)
		t.Logf("STOP SERVER")
		ts.Config.Close()
		t.Logf("SERVER STOPPED")
		time.Sleep(2000 * time.Millisecond)
		ts := httptest.NewUnstartedServer(mux)
		ts.TLS = &tls.Config{
			NextProtos: []string{"h2"},
		}
		listener, err := net.Listen("tcp", "127.0.0.1:8000")
		require.NoError(t, err)
		ts.Listener = listener
		t.Logf("RESTART SERVER")
		ts.StartTLS()
	}()
	// go func() {
	// 	stream, err := client.Heartbeat(ctx, sql.NewTarget("my client"))
	// 	if err != nil {
	// 		t.Logf("CLIENT HEARTBEAT FAILED %v", err)
	// 		return
	// 	}
	// 	for i := 0; ; i++ {
	// 		if i == 10 {
	// 			err := stream.CloseSend()
	// 			t.Logf("CLIENT CLOSE SEND %v", err)
	// 			return
	// 		}
	// 		targets, err := stream.Recv()
	// 		if err == io.EOF {
	// 			break
	// 		}
	// 		if err != nil {
	// 			t.Logf("CLIENT HEARTBEAT RECV FAILED %v", err)
	// 			time.Sleep(250 * time.Millisecond)
	// 			continue
	// 		}
	// 		for _, target := range targets.Targets {
	// 			t.Logf("CLIENT HEARTBEAT TARGET %s", target.Value)
	// 		}
	// 	}
	// }()
	for i := 0; i < 10; i++ {
		_, err := client.Connect(ctx, &sql.Database{Name: ":memory:"})
		if err != nil {
			t.Fatalf("Connect error: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	conn.Close()
	ts.Config.Close()
}

func newGateway(t *testing.T) *gateway.Gateway {
	logger := zaptest.NewLogger(t)
	config := gateway.Config{
		HeartbeatTimeout: 200 * time.Millisecond,
	}
	server := gateway.New(&sqlite3.SQLiteDriver{}, config, logger)
	return server
}

// Implement io.Writer and forward what it receives to a
// testing logger.
type testingWriter struct {
	t testing.TB
}

// Write a single log entry. It's assumed that p is always a \n-terminated UTF
// string.
func (w *testingWriter) Write(p []byte) (n int, err error) {
	w.t.Logf(string(p))
	return len(p), nil
}
