// TINYGO: The following is copied and modified from Go 1.19.3 official implementation.

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"fmt"

	"net/netip"
	"strconv"
	"time"
)

// TINYGO: Removed IPv6 stuff

// TCPAddr represents the address of a TCP end point.
type TCPAddr struct {
	IP   IP
	Port int
}

// Network returns the address's network name, "tcp".
func (a *TCPAddr) Network() string { return "tcp" }

func (a *TCPAddr) String() string {
	if a == nil {
		return "<nil>"
	}
	ip := []byte(a.IP)
	if a.Port == 0 {
		return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
	}
	return fmt.Sprintf("%d.%d.%d.%d:%d", ip[0], ip[1], ip[2], ip[3], a.Port)
}

// ResolveTCPAddr returns an address of TCP end point.
//
// The network must be a TCP network name.
//
// If the host in the address parameter is not a literal IP address or
// the port is not a literal port number, ResolveTCPAddr resolves the
// address to an address of TCP end point.
// Otherwise, it parses the address as a pair of literal IP address
// and port number.
// The address parameter can use a host name, but this is not
// recommended, because it will return at most one of the host name's
// IP addresses.
//
// See func Dial for a description of the network and address
// parameters.
func ResolveTCPAddr(network, address string) (*TCPAddr, error) {

	//	println("ResolveTCPAddr", address)
	switch network {
	case "tcp", "tcp4":
	default:
		return nil, fmt.Errorf("Network '%s' not supported", network)
	}

	// TINYGO: Use netdev resolver

	host, sport, err := SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(sport)
	if err != nil {
		return nil, fmt.Errorf("Error parsing port '%s' in address: %s",
			sport, err)
	}

	if host == "" {
		return &TCPAddr{Port: port}, nil
	}

	ip, err := currentdev.GetHostByName(host)
	if err != nil {
		return nil, fmt.Errorf("Lookup of host name '%s' failed: %s", host, err)
	}

	return &TCPAddr{IP: IP(ip[:]), Port: port}, nil
}

// TCPConn is an implementation of the Conn interface for TCP network
// connections.
type TCPConn struct {
	fd            uintptr
	laddr         *TCPAddr
	raddr         *TCPAddr
	readDeadline  time.Time
	writeDeadline time.Time
}

// DialTCP acts like Dial for TCP networks.
//
// The network must be a TCP network name; see func Dial for details.
//
// If laddr is nil, a local address is automatically chosen.
// If the IP field of raddr is nil or an unspecified IP address, the
// local system is assumed.
func DialTCP(network string, laddr, raddr *TCPAddr) (*TCPConn, error) {

	switch network {
	case "tcp", "tcp4":
	default:
		return nil, fmt.Errorf("Network '%s' not supported", network)
	}

	// TINYGO: Use netdev to create TCP socket and connect

	if raddr == nil {
		raddr = &TCPAddr{}
	}

	if raddr.IP.IsUnspecified() {
		return nil, fmt.Errorf("sorry, localhost isn't available on Tinygo")
	}

	fd, err := currentdev.Socket(AF_INET, SOCK_STREAM, IPPROTO_TCP)
	if err != nil {
		return nil, err
	}

	if err = currentdev.Connect(fd, raddr); err != nil {
		currentdev.Close(fd)
		return nil, err
	}

	return &TCPConn{
		fd:    fd,
		laddr: laddr,
		raddr: raddr,
	}, nil
}

// TCPAddrFromAddrPort returns addr as a TCPAddr. If addr.IsValid() is false,
// then the returned TCPAddr will contain a nil IP field, indicating an
// address family-agnostic unspecified address.
func TCPAddrFromAddrPort(addr netip.AddrPort) *TCPAddr {
	return &TCPAddr{
		IP: addr.Addr().AsSlice(),
		// Zone: addr.Addr().Zone(),
		Port: int(addr.Port()),
	}
}

// TINYGO: Use netdev for Conn methods: Read = Recv, Write = Send, etc.

func (c *TCPConn) Read(b []byte) (int, error) {
	var timeout time.Duration

	now := time.Now()

	if !c.readDeadline.IsZero() {
		if c.readDeadline.Before(now) {
			return 0, fmt.Errorf("Read deadline expired")
		} else {
			timeout = c.readDeadline.Sub(now)
		}
	}

	n, err := currentdev.Recv(c.fd, b, 0, timeout)
	// Turn the -1 socket error into 0 and let err speak for error
	if n < 0 {
		n = 0
	}
	return n, err
}

func (c *TCPConn) Write(b []byte) (int, error) {
	var timeout time.Duration

	now := time.Now()

	if !c.writeDeadline.IsZero() {
		if c.writeDeadline.Before(now) {
			return 0, fmt.Errorf("Write deadline expired")
		} else {
			timeout = c.writeDeadline.Sub(now)
		}
	}

	n, err := currentdev.Send(c.fd, b, 0, timeout)
	// Turn the -1 socket error into 0 and let err speak for error
	if n < 0 {
		n = 0
	}
	return n, err
}

func (c *TCPConn) Close() error {
	return currentdev.Close(c.fd)
}

func (c *TCPConn) LocalAddr() Addr {
	return c.laddr
}

func (c *TCPConn) RemoteAddr() Addr {
	return c.raddr
}

func (c *TCPConn) SetDeadline(t time.Time) error {
	c.readDeadline = t
	c.writeDeadline = t
	return nil
}

func (c *TCPConn) SetKeepAlive(keepalive bool) error {
	return currentdev.SetSockOpt(c.fd, SOL_SOCKET, SO_KEEPALIVE, keepalive)
}

func (c *TCPConn) SetKeepAlivePeriod(d time.Duration) error {
	// Units are 1/2 seconds
	return currentdev.SetSockOpt(c.fd, SOL_TCP, TCP_KEEPINTVL, 2*d.Seconds())
}

func (c *TCPConn) SetReadDeadline(t time.Time) error {
	c.readDeadline = t
	return nil
}

func (c *TCPConn) SetWriteDeadline(t time.Time) error {
	c.writeDeadline = t
	return nil
}

func (c *TCPConn) CloseWrite() error {
	return fmt.Errorf("CloseWrite not implemented")
}

type listener struct {
	fd    uintptr
	laddr *TCPAddr
}

func (l *listener) Accept() (Conn, error) {
	fd, err := currentdev.Accept(l.fd, &TCPAddr{})
	if err != nil {
		return nil, err
	}

	return &TCPConn{
		fd:    fd,
		laddr: l.laddr,
	}, nil
}

func (l *listener) Close() error {
	return currentdev.Close(l.fd)
}

func (l *listener) Addr() Addr {
	return l.laddr
}

func listenTCP(laddr *TCPAddr) (Listener, error) {
	fd, err := currentdev.Socket(AF_INET, SOCK_STREAM, IPPROTO_TCP)
	if err != nil {
		return nil, err
	}

	err = currentdev.Bind(fd, laddr)
	if err != nil {
		return nil, err
	}

	err = currentdev.Listen(fd, 5)
	if err != nil {
		return nil, err
	}

	return &listener{fd: fd, laddr: laddr}, nil
}
