package grpcsql

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/CanonicalLtd/go-grpc-sql/internal/protocol"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// DialOptions hold connection parameters for a gRPC SQL endpoint.
type DialOptions struct {
	Address      string
	Timeout      time.Duration
	CertFile     string
	KeyFile      string
	RootCertFile string
	SkipVerify   bool
}

// Convert the given options to DSN string suitable to be passed to sql.Open.
func (o *DialOptions) String() string {
	values := url.Values{}

	if o.Timeout != 0 {
		values.Add("timeout", o.Timeout.String())
	}
	if o.CertFile != "" {
		values.Add("certfile", o.CertFile)
	}
	if o.KeyFile != "" {
		values.Add("keyfile", o.KeyFile)
	}
	if o.RootCertFile != "" {
		values.Add("rootcertfile", o.RootCertFile)
	}
	if o.SkipVerify {
		values.Add("skipverify", "true")
	}

	name := o.Address
	if len(values) > 0 {
		name += fmt.Sprintf("?%s", values.Encode())
	}
	return name
}

// Generate a TLS configuration using the configured options.
func (o *DialOptions) tlsConfig() (*tls.Config, error) {
	config := &tls.Config{}

	if o.CertFile != "" {
		if err := tlsAddClientCertificate(config, o.CertFile, o.KeyFile); err != nil {
			return nil, err
		}
	}
	if o.RootCertFile != "" {
		if err := tlsAddRootCertificate(config, o.RootCertFile); err != nil {
			return nil, err
		}
	}
	if o.SkipVerify {
		config.InsecureSkipVerify = true
	}

	return config, nil
}

// Create a new connection to a gRPC SQL endpoint.
func dial(options *DialOptions) (*Conn, error) {
	tlsConfig, err := options.tlsConfig()
	if err != nil {
		return nil, err
	}
	grpcTarget := options.Address
	grpcOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	}
	grpcCtx := context.Background()
	if options.Timeout != 0 {
		var cancel func()
		grpcCtx, cancel = context.WithTimeout(grpcCtx, options.Timeout)
		defer cancel()

	}

	grpcConn, err := grpc.DialContext(grpcCtx, grpcTarget, grpcOptions...)
	if err != nil {
		return nil, fmt.Errorf("gRPC connection failed: %v", err)
	}

	grpcClient := protocol.NewSQLClient(grpcConn)
	grpcConnClient, err := grpcClient.Conn(grpcCtx)
	if err != nil {
		return nil, fmt.Errorf("gRPC conn method failed: %v", err)
	}

	conn := &Conn{
		grpcConn:       grpcConn,
		grpcConnClient: grpcConnClient,
	}

	return conn, nil

}

// Parse a datasource name and convert it to a structured DialOptions object.
func dialOptionsFromDSN(name string) (*DialOptions, error) {
	options := &DialOptions{
		Address: name,
	}

	pos := strings.IndexRune(options.Address, '?')
	if pos >= 1 {
		query, err := url.ParseQuery(options.Address[pos+1:])
		if err != nil {
			return nil, fmt.Errorf("failed to parse query params: %v", err)
		}
		if timeoutString := query.Get("timeout"); timeoutString != "" {
			options.Timeout, err = time.ParseDuration(timeoutString)
			if err != nil {
				return nil, fmt.Errorf("invalid timeout: %v", err)
			}
		}

		options.CertFile = query.Get("certfile")
		options.KeyFile = query.Get("keyfile")
		options.RootCertFile = query.Get("rootcertfile")
		options.SkipVerify = query.Get("skipverify") == "true"

		options.Address = options.Address[:pos]
	}

	return options, nil
}
