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

	ip, err := dev.GetHostByName(host)
	if err != nil {
		return nil, fmt.Errorf("Lookup of host name '%s' failed: %s", host, err)
	}

	return &UDPAddr{IP: ip[:4], Port: port}, nil
}

// UDPConn is the implementation of the Conn and PacketConn interfaces
// for UDP network connections.
type UDPConn struct {
	fd            netdev.Sockfd
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

	fd, err := dev.Socket(netdev.AF_INET, netdev.SOCK_DGRAM, netdev.IPPROTO_UDP)
	if err != nil {
		return nil, err
	}

	var ip netdev.IP

	copy(ip[:], laddr.IP)
	local := netdev.NewSockAddr("", netdev.Port(laddr.Port), ip)

	copy(ip[:], raddr.IP)
	remote := netdev.NewSockAddr("", netdev.Port(raddr.Port), ip)

	// Remote connect
	if err = dev.Connect(fd, remote); err != nil {
		dev.Close(fd)
		return nil, err
	}

	// Local bind
	err = dev.Bind(fd, local)
	if err != nil {
		dev.Close(fd)
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

	n, err := dev.Recv(c.fd, b, 0, timeout)
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

	n, err := dev.Send(c.fd, b, 0, timeout)
	// Turn the -1 socket error into 0 and let err speak for error
	if n < 0 {
		n = 0
	}
	return n, err
}

func (c *UDPConn) Close() error {
	return dev.Close(c.fd)
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
