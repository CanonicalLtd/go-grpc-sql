package grpcsql

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

// Add the given client certificate to the given tls.Config object.
func tlsAddClientCertificate(config *tls.Config, certFile string, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("failed to load client certificate: %v", err)
	}

	config.Certificates = []tls.Certificate{cert}
	return nil
}

// Add the given root CA certificate to the given tls.Config object.
func tlsAddRootCertificate(config *tls.Config, rootCertFile string) error {
	cert, err := ioutil.ReadFile(rootCertFile)
	if err != nil {
		return fmt.Errorf("failed to read root certificate: %v", err)
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(cert)
	if !ok {
		return fmt.Errorf("failed to parse root certificate: %v", err)
	}

	config.RootCAs = roots
	return nil

}
