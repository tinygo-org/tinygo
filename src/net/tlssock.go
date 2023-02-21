// TINYGO: The following is copied and modified from Go 1.19.3 official implementation.

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TLS low level connection and record layer

package net

import (
	"fmt"
	"net/netdev"
	"strconv"
	"time"
)

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

	fd, err := dev.Socket(netdev.AF_INET, netdev.SOCK_STREAM, netdev.IPPROTO_TLS)
	if err != nil {
		return nil, err
	}

	saddr := netdev.NewSockAddr(host, netdev.Port(port), netdev.IP{})

	if err = dev.Connect(fd, saddr); err != nil {
		dev.Close(fd)
		return nil, err
	}

	return &TLSConn{
		fd: fd,
	}, nil
}

// A TLSConn represents a secured connection.
// It implements the net.Conn interface.
type TLSConn struct {
	fd            netdev.Sockfd
	readDeadline  time.Time
	writeDeadline time.Time
}

// Access to net.Conn methods.
// Cannot just embed net.Conn because that would
// export the struct field too.

// LocalAddr returns the local network address.
func (c *TLSConn) LocalAddr() Addr {
	// TODO
	return nil
}

// RemoteAddr returns the remote network address.
func (c *TLSConn) RemoteAddr() Addr {
	// TODO
	return nil
}

// SetDeadline sets the read and write deadlines associated with the connection.
// A zero value for t means Read and Write will not time out.
// After a Write has timed out, the TLS state is corrupt and all future writes will return the same error.
func (c *TLSConn) SetDeadline(t time.Time) error {
	c.readDeadline = t
	c.writeDeadline = t
	return nil
}

// SetReadDeadline sets the read deadline on the underlying connection.
// A zero value for t means Read will not time out.
func (c *TLSConn) SetReadDeadline(t time.Time) error {
	c.readDeadline = t
	return nil
}

// SetWriteDeadline sets the write deadline on the underlying connection.
// A zero value for t means Write will not time out.
// After a Write has timed out, the TLS state is corrupt and all future writes will return the same error.
func (c *TLSConn) SetWriteDeadline(t time.Time) error {
	c.writeDeadline = t
	return nil
}

func (c *TLSConn) Read(b []byte) (int, error) {
	var timeout time.Duration

	now := time.Now()

	if !c.readDeadline.IsZero() {
		if c.readDeadline.Before(now) {
			return 0, fmt.Errorf("Read deadline expired")
		} else {
			timeout = c.readDeadline.Sub(now)
		}
	}

	n, err := dev.Recv(c.fd, b, 0, timeout)
	// Turn the -1 socket error into 0 and let err speak for error
	if n < 0 {
		n = 0
	}
	return n, err
}

func (c *TLSConn) Write(b []byte) (int, error) {
	var timeout time.Duration

	now := time.Now()

	if !c.writeDeadline.IsZero() {
		if c.writeDeadline.Before(now) {
			return 0, fmt.Errorf("Write deadline expired")
		} else {
			timeout = c.writeDeadline.Sub(now)
		}
	}

	n, err := dev.Send(c.fd, b, 0, timeout)
	// Turn the -1 socket error into 0 and let err speak for error
	if n < 0 {
		n = 0
	}
	return n, err
}

func (c *TLSConn) Close() error {
	return dev.Close(c.fd)
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
