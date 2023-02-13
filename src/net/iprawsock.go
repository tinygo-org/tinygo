// TINYGO: The following is copied and modified from Go 1.19.3 official implementation.

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

// BUG(mikio): On every POSIX platform, reads from the "ip4" network
// using the ReadFrom or ReadFromIP method might not return a complete
// IPv4 packet, including its header, even if there is space
// available. This can occur even in cases where Read or ReadMsgIP
// could return a complete packet. For this reason, it is recommended
// that you do not use these methods if it is important to receive a
// full packet.
//
// The Go 1 compatibility guidelines make it impossible for us to
// change the behavior of these methods; use Read or ReadMsgIP
// instead.

// BUG(mikio): On JS and Plan 9, methods and functions related
// to IPConn are not implemented.

// BUG(mikio): On Windows, the File method of IPConn is not
// implemented.

// IPAddr represents the address of an IP end point.
type IPAddr struct {
	IP   IP
	Zone string // IPv6 scoped addressing zone
}

// Network returns the address's network name, "ip".
func (a *IPAddr) Network() string { return "ip" }

func (a *IPAddr) String() string {
	if a == nil {
		return "<nil>"
	}
	ip := ipEmptyString(a.IP)
	if a.Zone != "" {
		return ip + "%" + a.Zone
	}
	return ip
}
