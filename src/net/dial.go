// TINYGO: The following is copied and modified from Go 1.19.3 official implementation.

// TINYGO: Omit DualStack support
// TINYGO: Omit Fast Fallback support
// TINYGO: Don't allow alternate resolver
// TINYGO: Omit DialTimeout

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"context"
	"fmt"
	"time"
)

// defaultTCPKeepAlive is a default constant value for TCPKeepAlive times
// See golang.org/issue/31510
const (
	defaultTCPKeepAlive = 15 * time.Second
)

// A Dialer contains options for connecting to an address.
//
// The zero value for each field is equivalent to dialing
// without that option. Dialing with the zero value of Dialer
// is therefore equivalent to just calling the Dial function.
//
// It is safe to call Dialer's methods concurrently.
type Dialer struct {
	// Timeout is the maximum amount of time a dial will wait for
	// a connect to complete. If Deadline is also set, it may fail
	// earlier.
	//
	// The default is no timeout.
	//
	// When using TCP and dialing a host name with multiple IP
	// addresses, the timeout may be divided between them.
	//
	// With or without a timeout, the operating system may impose
	// its own earlier timeout. For instance, TCP timeouts are
	// often around 3 minutes.
	Timeout time.Duration

	// Deadline is the absolute point in time after which dials
	// will fail. If Timeout is set, it may fail earlier.
	// Zero means no deadline, or dependent on the operating system
	// as with the Timeout option.
	Deadline time.Time

	// LocalAddr is the local address to use when dialing an
	// address. The address must be of a compatible type for the
	// network being dialed.
	// If nil, a local address is automatically chosen.
	LocalAddr Addr

	// KeepAlive specifies the interval between keep-alive
	// probes for an active network connection.
	// If zero, keep-alive probes are sent with a default value
	// (currently 15 seconds), if supported by the protocol and operating
	// system. Network protocols or operating systems that do
	// not support keep-alives ignore this field.
	// If negative, keep-alive probes are disabled.
	KeepAlive time.Duration
}

// Dial connects to the address on the named network.
//
// See Go "net" package Dial() for more information.
//
// Note: Tinygo Dial supports a subset of networks supported by Go Dial,
// specifically: "tcp", "tcp4", "udp", and "udp4".  IP and unix networks are
// not supported.
func Dial(network, address string) (Conn, error) {
	var d Dialer
	return d.Dial(network, address)
}

// DialTimeout acts like Dial but takes a timeout.
//
// The timeout includes name resolution, if required.
// When using TCP, and the host in the address parameter resolves to
// multiple IP addresses, the timeout is spread over each consecutive
// dial, such that each is given an appropriate fraction of the time
// to connect.
//
// See func Dial for a description of the network and address
// parameters.
func DialTimeout(network, address string, timeout time.Duration) (Conn, error) {
	d := Dialer{Timeout: timeout}
	return d.Dial(network, address)
}

// Dial connects to the address on the named network.
//
// See func Dial for a description of the network and address
// parameters.
//
// Dial uses context.Background internally; to specify the context, use
// DialContext.
func (d *Dialer) Dial(network, address string) (Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

// DialContext connects to the address on the named network using
// the provided context.
//
// The provided Context must be non-nil. If the context expires before
// the connection is complete, an error is returned. Once successfully
// connected, any expiration of the context will not affect the
// connection.
//
// When using TCP, and the host in the address parameter resolves to multiple
// network addresses, any dial timeout (from d.Timeout or ctx) is spread
// over each consecutive dial, such that each is given an appropriate
// fraction of the time to connect.
// For example, if a host has 4 IP addresses and the timeout is 1 minute,
// the connect to each single address will be given 15 seconds to complete
// before trying the next one.
//
// See func Dial for a description of the network and address
// parameters.
func (d *Dialer) DialContext(ctx context.Context, network, address string) (Conn, error) {

	// TINYGO: Ignoring context

	switch network {
	case "tcp", "tcp4":
		raddr, err := ResolveTCPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return DialTCP(network, nil, raddr)
	case "udp", "udp4":
		raddr, err := ResolveUDPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return DialUDP(network, nil, raddr)
	}

	return nil, fmt.Errorf("Network %s not supported", network)
}

// Listen announces on the local network address.
//
// See Go "net" package Listen() for more information.
//
// Note: Tinygo Listen supports a subset of networks supported by Go Listen,
// specifically: "tcp", "tcp4".  "tcp6" and unix networks are not supported.
func Listen(network, address string) (Listener, error) {

	//	println("Listen", address)
	switch network {
	case "tcp", "tcp4":
	default:
		return nil, fmt.Errorf("Network %s not supported", network)
	}

	laddr, err := ResolveTCPAddr(network, address)
	if err != nil {
		return nil, err
	}

	return listenTCP(laddr)
}
