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

// UDPAddr represents the address of a UDP end point.
type UDPAddr struct {
	IP   IP
	Port int
}

// Network returns the address's network name, "udp".
func (a *UDPAddr) Network() string { return "udp" }

func (a *UDPAddr) String() string {
	if a == nil {
		return "<nil>"
	}

	// TINYGO: Work around not having internal/itoa

	ip := []byte(a.IP)
	if a.Port == 0 {
		return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
	}
	return fmt.Sprintf("%d.%d.%d.%d:%d", ip[0], ip[1], ip[2], ip[3], a.Port)
}

// ResolveUDPAddr returns an address of UDP end point.
//
// The network must be a UDP network name.
//
// If the host in the address parameter is not a literal IP address or
// the port is not a literal port number, ResolveUDPAddr resolves the
// address to an address of UDP end point.
// Otherwise, it parses the address as a pair of literal IP address
// and port number.
// The address parameter can use a host name, but this is not
// recommended, because it will return at most one of the host name's
// IP addresses.
//
// See func Dial for a description of the network and address
// parameters.
func ResolveUDPAddr(network, address string) (*UDPAddr, error) {

	switch network {
	case "udp", "udp4":
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
		return &UDPAddr{Port: port}, nil
	}

	ip, err := currentdev.GetHostByName(host)
	if err != nil {
		return nil, fmt.Errorf("Lookup of host name '%s' failed: %s", host, err)
	}

	return &UDPAddr{IP: ip[:4], Port: port}, nil
}

// UDPConn is the implementation of the Conn and PacketConn interfaces
// for UDP network connections.
type UDPConn struct {
	fd            uintptr
	laddr         *UDPAddr
	raddr         *UDPAddr
	readDeadline  time.Time
	writeDeadline time.Time
}

// Use IANA RFC 6335 port range 49152–65535 for ephemeral (dynamic) ports
var eport = int32(49151)

func ephemeralPort() int {
	// TODO: this is racy, if concurrent DialUDPs; use atomic?
	if eport == int32(65535) {
		eport = int32(49151)
	} else {
		eport++
	}
	return int(eport)
}

// UDPAddrFromAddrPort returns addr as a UDPAddr. If addr.IsValid() is false,
// then the returned UDPAddr will contain a nil IP field, indicating an
// address family-agnostic unspecified address.
func UDPAddrFromAddrPort(addr netip.AddrPort) *UDPAddr {
	return &UDPAddr{
		IP: addr.Addr().AsSlice(),
		// Zone: addr.Addr().Zone(),
		Port: int(addr.Port()),
	}
}

// DialUDP acts like Dial for UDP networks.
//
// The network must be a UDP network name; see func Dial for details.
//
// If laddr is nil, a local address is automatically chosen.
// If the IP field of raddr is nil or an unspecified IP address, the
// local system is assumed.
func DialUDP(network string, laddr, raddr *UDPAddr) (*UDPConn, error) {
	switch network {
	case "udp", "udp4":
	default:
		return nil, fmt.Errorf("Network '%s' not supported", network)
	}

	// TINYGO: Use netdev to create UDP socket and connect

	if laddr == nil {
		laddr = &UDPAddr{}
	}

	if raddr == nil {
		raddr = &UDPAddr{}
	}

	if raddr.IP.IsUnspecified() {
		return nil, fmt.Errorf("Sorry, localhost isn't available on Tinygo")
	}

	// If no port was given, grab an ephemeral port
	if laddr.Port == 0 {
		laddr.Port = ephemeralPort()
	}

	fd, err := currentdev.Socket(AF_INET, SOCK_DGRAM, IPPROTO_UDP)
	if err != nil {
		return nil, err
	}

	// Remote connect
	if err = currentdev.Connect(fd, raddr); err != nil {
		currentdev.Close(fd)
		return nil, err
	}

	// Local bind
	err = currentdev.Bind(fd, laddr)
	if err != nil {
		currentdev.Close(fd)
		return nil, err
	}

	return &UDPConn{
		fd:    fd,
		laddr: laddr,
		raddr: raddr,
	}, nil
}

// TINYGO: Use netdev for Conn methods: Read = Recv, Write = Send, etc.

func (c *UDPConn) Read(b []byte) (int, error) {
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

func (c *UDPConn) Write(b []byte) (int, error) {
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

func (c *UDPConn) Close() error {
	return currentdev.Close(c.fd)
}

func (c *UDPConn) LocalAddr() Addr {
	return c.laddr
}

func (c *UDPConn) RemoteAddr() Addr {
	return c.raddr
}

func (c *UDPConn) SetDeadline(t time.Time) error {
	c.readDeadline = t
	c.writeDeadline = t
	return nil
}

func (c *UDPConn) SetReadDeadline(t time.Time) error {
	c.readDeadline = t
	return nil
}

func (c *UDPConn) SetWriteDeadline(t time.Time) error {
	c.writeDeadline = t
	return nil
}
