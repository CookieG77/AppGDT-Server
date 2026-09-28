package server

import (
	"crypto/tls"
	"fmt"
)

// LoadTLSConfig loads the certificate and its private key, so that a missing
// or invalid file stops the startup with a clear error.
// TLS 1.2 is the oldest accepted version; Go chooses secure cipher suites.
func LoadTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("loading TLS certificate failed: %w", err)
	}
	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
	}, nil
}
