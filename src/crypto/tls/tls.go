// TINYGO: The following is copied and modified from Go 1.21.4 official implementation.

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package tls partially implements TLS 1.2, as specified in RFC 5246,
// and TLS 1.3, as specified in RFC 8446.
package tls

// BUG(agl): The crypto/tls package only implements some countermeasures
// against Lucky13 attacks on CBC-mode encryption, and only on SHA1
// variants. See http://www.isg.rhul.ac.uk/tls/TLStiming.pdf and
// https://www.imperialviolet.org/2013/02/04/luckythirteen.html.

import (
	"context"
	"errors"
	"fmt"
	"net"
)

// Conn represents a secured connection. TINYGO: the actual TLS handshake and
// record layer are offloaded to the network device (see net.TLSConn), so Conn
// is a thin wrapper over an underlying net.Conn that exists to satisfy callers
// (e.g. google.golang.org/grpc/credentials) which expect the *tls.Conn API
// shape — ConnectionState/Handshake — that this package does not implement in
// software.
type Conn struct {
	net.Conn
}

// ConnectionState returns basic TLS details about the connection. TINYGO:
// empty; TLS is offloaded to the network device.
func (c *Conn) ConnectionState() ConnectionState {
	return ConnectionState{}
}

// Handshake runs the client or server handshake protocol if it has not yet been
// run. TINYGO: no-op; the handshake is performed by the network device.
func (c *Conn) Handshake() error {
	return c.HandshakeContext(context.Background())
}

// HandshakeContext is the context-aware variant of Handshake. TINYGO: no-op.
func (c *Conn) HandshakeContext(ctx context.Context) error {
	return nil
}

// NetConn returns the underlying connection that is wrapped by c.
func (c *Conn) NetConn() net.Conn {
	return c.Conn
}

// Client returns a new TLS client side connection
// using conn as the underlying transport.
// The config cannot be nil: users must set either ServerName or
// InsecureSkipVerify in the config.
func Client(conn net.Conn, config *Config) *Conn {
	return &Conn{Conn: conn}
}

// Server returns a new TLS server side connection
// using conn as the underlying transport.
// The configuration config must be non-nil and must include
// at least one certificate or else set GetCertificate.
func Server(conn net.Conn, config *Config) *Conn {
	return &Conn{Conn: conn}
}

// A listener implements a network listener (net.Listener) for TLS connections.
type listener struct {
	net.Listener
	config *Config
}

// NewListener creates a Listener which accepts connections from an inner
// Listener and wraps each connection with Server.
// The configuration config must be non-nil and must include
// at least one certificate or else set GetCertificate.
func NewListener(inner net.Listener, config *Config) net.Listener {
	l := new(listener)
	l.Listener = inner
	l.config = config
	return l
}

// DialWithDialer connects to the given network address using dialer.Dial and
// then initiates a TLS handshake, returning the resulting TLS connection. Any
// timeout or deadline given in the dialer apply to connection and TLS
// handshake as a whole.
//
// DialWithDialer interprets a nil configuration as equivalent to the zero
// configuration; see the documentation of Config for the defaults.
//
// DialWithDialer uses context.Background internally; to specify the context,
// use Dialer.DialContext with NetDialer set to the desired dialer.
func DialWithDialer(dialer *net.Dialer, network, addr string, config *Config) (*net.TLSConn, error) {
	switch network {
	case "tcp", "tcp4":
	default:
		return nil, fmt.Errorf("Network %s not supported", network)
	}

	return net.DialTLS(addr)
}

// Dial connects to the given network address using net.Dial
// and then initiates a TLS handshake, returning the resulting
// TLS connection.
// Dial interprets a nil configuration as equivalent to
// the zero configuration; see the documentation of Config
// for the defaults.
func Dial(network, addr string, config *Config) (*net.TLSConn, error) {
	return DialWithDialer(new(net.Dialer), network, addr, config)
}

// Dialer dials TLS connections given a configuration and a Dialer for the
// underlying connection.
type Dialer struct {
	// NetDialer is the optional dialer to use for the TLS connections'
	// underlying TCP connections.
	// A nil NetDialer is equivalent to the net.Dialer zero value.
	NetDialer *net.Dialer

	// Config is the TLS configuration to use for new connections.
	// A nil configuration is equivalent to the zero
	// configuration; see the documentation of Config for the
	// defaults.
	Config *Config
}

// DialContext connects to the given network address and initiates a TLS
// handshake, returning the resulting TLS connection.
//
// The provided Context must be non-nil. If the context expires before
// the connection is complete, an error is returned. Once successfully
// connected, any expiration of the context will not affect the
// connection.
//
// The returned Conn, if any, will always be of type *Conn.
func (d *Dialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	switch network {
	case "tcp", "tcp4":
	default:
		return nil, fmt.Errorf("Network %s not supported", network)
	}

	return net.DialTLS(addr)
}

// LoadX509KeyPair reads and parses a public/private key pair from a pair
// of files. The files must contain PEM encoded data. The certificate file
// may contain intermediate certificates following the leaf certificate to
// form a certificate chain. On successful return, Certificate.Leaf will
// be nil because the parsed form of the certificate is not retained.
func LoadX509KeyPair(certFile, keyFile string) (Certificate, error) {
	return Certificate{}, errors.New("tls:LoadX509KeyPair not implemented")
}
