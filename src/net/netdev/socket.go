// Netdev socket interface

package netdev

import (
	"fmt"
	"time"
)

type AddressFamily int

const (
	AF_UNSPEC = 0
	AF_INET   = 2
	// Currently not supporting AF_INET6 (IPv6)
)

func (f AddressFamily) String() string {
	switch f {
	case AF_INET:
		return "INET"
	}
	return "Unknown"
}

type SockType int

const (
	SOCK_STREAM = 1
	SOCK_DGRAM  = 2
)

func (t SockType) String() string {
	switch t {
	case SOCK_STREAM:
		return "STREAM"
	case SOCK_DGRAM:
		return "DATAGRAM"
	}
	return "Unknown"
}

type Protocol int

const (
	IPPROTO_TCP = 0x06
	IPPROTO_UDP = 0x11
	// Made up, not a real IP protocol number.  This is used to create
	// a TLS socket on the device, assuming the device supports mbed TLS.
	IPPROTO_TLS = 0xFE
)

func (p Protocol) String() string {
	switch p {
	case IPPROTO_TCP:
		return "TCP"
	case IPPROTO_UDP:
		return "UDP"
	case IPPROTO_TLS:
		return "TLS"
	}
	return "Unknown"
}

// sockaddr_in
type SockAddr struct {
	host string
	port Port
	ip   IP
}

func NewSockAddr(host string, port Port, ip IP) SockAddr {
	return SockAddr{host: host, port: port, ip: ip}
}

func (a SockAddr) String() string {
	if a.host == "" {
		return fmt.Sprintf("%s:%d", a.ip.String(), a.port)
	}
	return fmt.Sprintf("%s:%d", a.host, a.port)
}

func (a SockAddr) Port() uint16 {
	return uint16(a.port)
}

func (a SockAddr) IpUint32() uint32 {
	return uint32(a.ip[0])<<24 |
		uint32(a.ip[1])<<16 |
		uint32(a.ip[2])<<8 |
		uint32(a.ip[3])
}

func (a SockAddr) IpBytes() []byte {
	return a.ip[:]
}

func (a SockAddr) Host() string {
	return a.host
}

func (a SockAddr) Ip() IP {
	return a.ip
}

type SockFlags int

const (
	MSG_OOB       SockFlags = 1
	MSG_PEEK                = 2
	MSG_DONTROUTE           = 4
)

type SockOpt int

const (
	SO_KEEPALIVE  SockOpt = 9 // Value: bool
	TCP_KEEPINTVL         = 5 // Interval between keepalives, value (1/2 sec units)
)

type SockOptLevel int

const (
	SOL_SOCKET SockOptLevel = 1
	SOL_TCP                 = 6
)

type Sockfd int

// Berkely Sockets-like interface.  See man page for socket(2), etc.
//
// Multiple goroutines may invoke methods on a Socketer simultaneously.
type Socketer interface {
	Socket(family AddressFamily, sockType SockType, protocol Protocol) (Sockfd, error)
	Bind(sockfd Sockfd, myaddr SockAddr) error
	Connect(sockfd Sockfd, servaddr SockAddr) error
	Listen(sockfd Sockfd, backlog int) error
	Accept(sockfd Sockfd, peer SockAddr) (Sockfd, error)
	Send(sockfd Sockfd, buf []byte, flags SockFlags, timeout time.Duration) (int, error)
	Recv(sockfd Sockfd, buf []byte, flags SockFlags, timeout time.Duration) (int, error)
	Close(sockfd Sockfd) error
	SetSockOpt(sockfd Sockfd, level SockOptLevel, opt SockOpt, value any) error
}
