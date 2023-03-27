package net

import (
	"time"
)

// netdev is the current netdev, set by the application with useNetdev()
var netdev netdever

// (useNetdev is go:linkname'd from tinygo/drivers package)
func useNetdev(dev netdever) {
	netdev = dev
}

// Netdev is TinyGo's network device driver model.  Network drivers implement
// the netdever interface, providing a common network I/O interface to TinyGo's
// "net" package.  The interface is modeled after the BSD socket interface.
// net.Conn implementations (TCPConn, UDPConn, and TLSConn) use the netdev
// interface for device I/O access.
//
// A netdever is passed to the "net" package using net.useNetdev().
//
// Just like a net.Conn, multiple goroutines may invoke methods on a netdever
// simultaneously.
//
// NOTE: The netdever interface is mirrored in drivers/netdev.go.
// NOTE: If making changes to this interface, mirror the changes in
// NOTE: drivers/netdev.go, and visa-versa.

type netdever interface {

	// GetHostByName returns the IP address of either a hostname or IPv4
	// address in standard dot notation
	GetHostByName(name string) (IP, error)

	// Berkely Sockets-like interface, Go-ified.  See man page for socket(2), etc.
	Socket(domain int, stype int, protocol int) (int, error)
	Bind(sockfd int, ip IP, port int) error
	Connect(sockfd int, host string, ip IP, port int) error
	Listen(sockfd int, backlog int) error
	Accept(sockfd int, ip IP, port int) (int, error)
	Send(sockfd int, buf []byte, flags int, timeout time.Duration) (int, error)
	Recv(sockfd int, buf []byte, flags int, timeout time.Duration) (int, error)
	Close(sockfd int) error
	SetSockOpt(sockfd int, level int, opt int, value interface{}) error
}
