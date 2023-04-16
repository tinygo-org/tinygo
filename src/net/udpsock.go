// TINYGO: The following is copied and modified from Go 1.19.3 official implementation.

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"fmt"
	"internal/itoa"
	"io"
	"net/netip"
	"strconv"
	"syscall"
	"time"
)

// UDPAddr represents the address of a UDP end point.
type UDPAddr struct {
	IP   IP
	Port int
	Zone string // IPv6 scoped addressing zone
}

// AddrPort returns the UDPAddr a as a netip.AddrPort.
//
// If a.Port does not fit in a uint16, it's silently truncated.
//
// If a is nil, a zero value is returned.
func (a *UDPAddr) AddrPort() netip.AddrPort {
	if a == nil {
		return netip.AddrPort{}
	}
	na, _ := netip.AddrFromSlice(a.IP)
	na = na.WithZone(a.Zone)
	return netip.AddrPortFrom(na, uint16(a.Port))
}

// Network returns the address's network name, "udp".
func (a *UDPAddr) Network() string { return "udp" }

func (a *UDPAddr) String() string {
	if a == nil {
		return "<nil>"
	}
	ip := ipEmptyString(a.IP)
	if a.Zone != "" {
		return JoinHostPort(ip+"%"+a.Zone, itoa.Itoa(a.Port))
	}
	return JoinHostPort(ip, itoa.Itoa(a.Port))
}

func (a *UDPAddr) isWildcard() bool {
	if a == nil || a.IP == nil {
		return true
	}
	return a.IP.IsUnspecified()
}

func (a *UDPAddr) opAddr() Addr {
	if a == nil {
		return nil
	}
	return a
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

	ip, err := netdev.GetHostByName(host)
	if err != nil {
		return nil, fmt.Errorf("Lookup of host name '%s' failed: %s", host, err)
	}

	return &UDPAddr{IP: ip, Port: port}, nil
}

// UDPConn is the implementation of the Conn and PacketConn interfaces
// for UDP network connections.
type UDPConn struct {
	fd            int
	net           string
	laddr         *UDPAddr
	raddr         *UDPAddr
	readDeadline  time.Time
	writeDeadline time.Time
}

// Use IANA RFC 6335 port range 49152–65535 for ephemeral (dynamic) ports
var eport = int32(49151)

func ephemeralPort() int {
	if eport == int32(65535) {
		eport = int32(49151)
	} else {
		eport++
	}
	return int(eport)
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

	fd, err := netdev.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		return nil, err
	}

	// Local bind
	err = netdev.Bind(fd, laddr.IP, laddr.Port)
	if err != nil {
		netdev.Close(fd)
		return nil, err
	}

	// Remote connect
	if err = netdev.Connect(fd, "", raddr.IP, raddr.Port); err != nil {
		netdev.Close(fd)
		return nil, err
	}

	return &UDPConn{
		fd:    fd,
		net:   network,
		laddr: laddr,
		raddr: raddr,
	}, nil
}

// TINYGO: Use netdev for Conn methods: Read = Recv, Write = Send, etc.

func (c *UDPConn) Read(b []byte) (int, error) {
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

func (c *UDPConn) Write(b []byte) (int, error) {
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

func (c *UDPConn) Close() error {
	return netdev.Close(c.fd)
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
