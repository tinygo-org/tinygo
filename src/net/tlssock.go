// TINYGO: The following is copied and modified from Go 1.19.3 official implementation.

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TLS low level connection and record layer

package net

import (
	"internal/itoa"
	"io"
	"strconv"
	"time"
)

// TLSAddr represents the address of a TLS end point.
type TLSAddr struct {
	Host string
	Port int
}

func (a *TLSAddr) Network() string { return "tls" }

func (a *TLSAddr) String() string {
	if a == nil {
		return "<nil>"
	}
	return JoinHostPort(a.Host, itoa.Itoa(a.Port))
}

// A TLSConn represents a secured connection.
// It implements the net.Conn interface.
type TLSConn struct {
	fd            int
	net           string
	laddr         *TLSAddr
	raddr         *TLSAddr
	readDeadline  time.Time
	writeDeadline time.Time
}

func DialTLS(addr string) (*TLSConn, error) {

	host, sport, err := SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(sport)
	if err != nil {
		return nil, err
	}

	if port == 0 {
		port = 443
	}

	fd, err := netdev.Socket(_AF_INET, _SOCK_STREAM, _IPPROTO_TLS)
	if err != nil {
		return nil, err
	}

	if err = netdev.Connect(fd, host, IP{}, port); err != nil {
		netdev.Close(fd)
		return nil, err
	}

	return &TLSConn{
		fd:    fd,
		net:   "tls",
		raddr: &TLSAddr{host, port},
	}, nil
}

func (c *TLSConn) Read(b []byte) (int, error) {
	n, err := netdev.Recv(c.fd, b, 0, c.readDeadline)
	// Turn the -1 socket error into 0 and let err speak for error
	if n < 0 {
		n = 0
	}
	if err != nil && err != io.EOF {
		err = &OpError{Op: "read", Net: c.net, Source: c.laddr, Addr: c.raddr, Err: err}
	}
	return n, err
}

func (c *TLSConn) Write(b []byte) (int, error) {
	n, err := netdev.Send(c.fd, b, 0, c.writeDeadline)
	// Turn the -1 socket error into 0 and let err speak for error
	if n < 0 {
		n = 0
	}
	if err != nil {
		err = &OpError{Op: "write", Net: c.net, Source: c.laddr, Addr: c.raddr, Err: err}
	}
	return n, err
}

func (c *TLSConn) Close() error {
	return netdev.Close(c.fd)
}

func (c *TLSConn) LocalAddr() Addr {
	return c.laddr
}

func (c *TLSConn) RemoteAddr() Addr {
	return c.raddr
}

func (c *TLSConn) SetDeadline(t time.Time) error {
	c.readDeadline = t
	c.writeDeadline = t
	return nil
}

func (c *TLSConn) SetReadDeadline(t time.Time) error {
	c.readDeadline = t
	return nil
}

func (c *TLSConn) SetWriteDeadline(t time.Time) error {
	c.writeDeadline = t
	return nil
}

// Handshake runs the client or server handshake
// protocol if it has not yet been run.
//
// Most uses of this package need not call Handshake explicitly: the
// first Read or Write will call it automatically.
//
// For control over canceling or setting a timeout on a handshake, use
// HandshakeContext or the Dialer's DialContext method instead.
func (c *TLSConn) Handshake() error {
	panic("TLSConn.Handshake() not implemented")
	return nil
}
