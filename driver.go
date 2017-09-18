package grpcsql

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
	"time"
)

func init() {
	sql.Register("grpc", &Driver{})
}

// Driver implements the database/sql/driver interface and executes the
// relevant statements over gRPC.
type Driver struct {
}

// Open a new connection against a gRPC SQL server.
//
// The given data source name must be in the form:
//
// <host>[:<port>][?<options>]
//
// where valid options are:
//
// - certfile: cert file location
// - certkey: key file location
// - rootcertfile: location of the root certificate file
//
// Note that the connection will always use TLS, so valid TLS parameters for
// the target endpoint are required.
func (d *Driver) Open(name string) (driver.Conn, error) {
	dialOptions, err := parseName(name)
	if err != nil {
		return nil, err
	}

	conn, err := dial(dialOptions)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Convert a datasource name to an actual gRPC target string and dial options.
func parseName(name string) (*dialOptions, error) {
	dialOptions := &dialOptions{
		Address: name,
	}

	pos := strings.IndexRune(name, '?')
	if pos >= 1 {
		query, err := url.ParseQuery(name[pos+1:])
		if err != nil {
			return nil, fmt.Errorf("failed to parse query params: %v", err)
		}
		if key := query.Get("timeout"); key != "" {
			timeout, err := time.ParseDuration(key)
			if err != nil {
				return nil, fmt.Errorf("invalid timeout: %v", err)
			}
			dialOptions.Timeout = timeout
		}

		tlsConfig, err := newTLSConfig(
			query.Get("certfile"), query.Get("keyfile"), query.Get("rootcertfile"))
		if err != nil {
			return nil, err
		}
		dialOptions.TLSConfig = tlsConfig

		dialOptions.Address = name[:pos]
	}

	return dialOptions, nil
}

// Generate a tls.Config object using the given parameters.
func newTLSConfig(certFile string, keyFile string, rootCertFile string) (*tls.Config, error) {
	config := &tls.Config{}

	// Client certificate.
	if certFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %v", err)
		}
		config.Certificates = []tls.Certificate{cert}
	}

	// Certificate authority.
	if rootCertFile != "" {
		cert, err := ioutil.ReadFile(rootCertFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read root certificate: %v", err)
		}

		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM(cert)
		if !ok {
			return nil, fmt.Errorf("failed to parse root certificate: %v", err)
		}

	}

	return config, nil
}
