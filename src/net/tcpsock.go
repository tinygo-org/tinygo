// TINYGO: The following is copied and modified from Go 1.19.3 official implementation.

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"fmt"
	"internal/itoa"
	"net/netip"
	"strconv"
	"time"
)

// TCPAddr represents the address of a TCP end point.
type TCPAddr struct {
	IP   IP
	Port int
	Zone string // IPv6 scoped addressing zone
}

// AddrPort returns the TCPAddr a as a netip.AddrPort.
//
// If a.Port does not fit in a uint16, it's silently truncated.
//
// If a is nil, a zero value is returned.
func (a *TCPAddr) AddrPort() netip.AddrPort {
	if a == nil {
		return netip.AddrPort{}
	}
	na, _ := netip.AddrFromSlice(a.IP)
	na = na.WithZone(a.Zone)
	return netip.AddrPortFrom(na, uint16(a.Port))
}

// Network returns the address's network name, "tcp".
func (a *TCPAddr) Network() string { return "tcp" }

func (a *TCPAddr) String() string {
	if a == nil {
		return "<nil>"
	}
	ip := ipEmptyString(a.IP)
	if a.Zone != "" {
		return JoinHostPort(ip+"%"+a.Zone, itoa.Itoa(a.Port))
	}
	return JoinHostPort(ip, itoa.Itoa(a.Port))
}

func (a *TCPAddr) isWildcard() bool {
	if a == nil || a.IP == nil {
		return true
	}
	return a.IP.IsUnspecified()
}

func (a *TCPAddr) opAddr() Addr {
	if a == nil {
		return nil
	}
	return a
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

	ip, err := dev.GetHostByName(host)
	if err != nil {
		return nil, fmt.Errorf("Lookup of host name '%s' failed: %s", host, err)
	}

	return &TCPAddr{IP: IP(ip[:]), Port: port}, nil
}

// TCPConn is an implementation of the Conn interface for TCP network
// connections.
type TCPConn struct {
	fd            netdev.Sockfd
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
		return nil, fmt.Errorf("Sorry, localhost isn't available on Tinygo")
	}

	fd, err := dev.Socket(netdev.AF_INET, netdev.SOCK_STREAM, netdev.IPPROTO_TCP)
	if err != nil {
		return nil, err
	}

	var ip netdev.IP
	copy(ip[:], raddr.IP)
	addr := netdev.NewSockAddr("", netdev.Port(raddr.Port), ip)

	if err = dev.Connect(fd, addr); err != nil {
		dev.Close(fd)
		return nil, err
	}

	return &TCPConn{
		fd:    fd,
		laddr: laddr,
		raddr: raddr,
	}, nil
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

	n, err := dev.Recv(c.fd, b, 0, timeout)
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

	n, err := dev.Send(c.fd, b, 0, timeout)
	// Turn the -1 socket error into 0 and let err speak for error
	if n < 0 {
		n = 0
	}
	return n, err
}

func (c *TCPConn) Close() error {
	return dev.Close(c.fd)
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
	return dev.SetSockOpt(c.fd, netdev.SOL_SOCKET, netdev.SO_KEEPALIVE, keepalive)
}

func (c *TCPConn) SetKeepAlivePeriod(d time.Duration) error {
	// Units are 1/2 seconds
	return dev.SetSockOpt(c.fd, netdev.SOL_TCP, netdev.TCP_KEEPINTVL, 2*d.Seconds())
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
	fd    netdev.Sockfd
	laddr *TCPAddr
}

func (l *listener) Accept() (Conn, error) {
	fd, err := dev.Accept(l.fd, netdev.SockAddr{})
	if err != nil {
		return nil, err
	}

	return &TCPConn{
		fd:    fd,
		laddr: l.laddr,
	}, nil
}

func (l *listener) Close() error {
	return dev.Close(l.fd)
}

func (l *listener) Addr() Addr {
	return l.laddr
}

func listenTCP(laddr *TCPAddr) (Listener, error) {
	fd, err := dev.Socket(netdev.AF_INET, netdev.SOCK_STREAM, netdev.IPPROTO_TCP)
	if err != nil {
		return nil, err
	}

	var ip netdev.IP
	copy(ip[:], laddr.IP)
	addr := netdev.NewSockAddr("", netdev.Port(laddr.Port), ip)

	err = dev.Bind(fd, addr)
	if err != nil {
		return nil, err
	}

	err = dev.Listen(fd, 5)
	if err != nil {
		return nil, err
	}

	return &listener{fd: fd, laddr: laddr}, nil
}
