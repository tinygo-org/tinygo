// Netdev is TinyGo's network device driver model.  TinyGo's "net" package
// interfaces to the netdev driver directly to provide TCPConn, UDPConn,
// and TLSConn socket connections.

package netdev

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type HardwareAddr []byte

const hexDigit = "0123456789abcdef"

func (a HardwareAddr) String() string {
	if len(a) == 0 {
		return ""
	}
	buf := make([]byte, 0, len(a)*3-1)
	for i, b := range a {
		if i > 0 {
			buf = append(buf, ':')
		}
		buf = append(buf, hexDigit[b>>4])
		buf = append(buf, hexDigit[b&0xF])
	}
	return string(buf)
}

func ParseHardwareAddr(s string) HardwareAddr {
	parts := strings.Split(s, ":")
	if len(parts) != 6 {
		return nil
	}

	var mac []byte
	for _, part := range parts {
		b, err := strconv.ParseInt(part, 16, 64)
		if err != nil {
			return nil
		}
		mac = append(mac, byte(b))
	}

	return HardwareAddr(mac)
}

type Port uint16 // host byte-order

type IP [4]byte

func (ip IP) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

func ParseIP(s string) IP {
	var result IP
	octets := strings.Split(s, ".")
	if len(octets) != 4 {
		return IP{}
	}
	for i, octet := range octets {
		v, err := strconv.Atoi(octet)
		if err != nil {
			return IP{}
		}
		result[i] = byte(v)
	}
	return result
}

// NetConnect() errors
var ErrConnected = errors.New("Already connected")
var ErrConnectFailed = errors.New("Connect failed")
var ErrConnectTimeout = errors.New("Connect timed out")
var ErrMissingSSID = errors.New("Missing WiFi SSID")
var ErrStartingDHCPClient = errors.New("Error starting DHPC client")

// GethostByName() errors
var ErrHostUnknown = errors.New("Host unknown")

// Socketer errors
var ErrFamilyNotSupported = errors.New("Address family not supported")
var ErrProtocolNotSupported = errors.New("Socket protocol/type not supported")
var ErrNoMoreSockets = errors.New("No more sockets")
var ErrClosingSocket = errors.New("Error closing socket")
var ErrNotSupported = errors.New("Not supported")
var ErrRecvTimeout = errors.New("Recv timeout expired")

type Event int

// NetNotify network events
const (
	// The device's network connection is now UP
	EventNetUp Event = iota
	// The device's network connection is now DOWN
	EventNetDown
)

// Netdev drivers implement the Netdever interface.
//
// A Netdever is passed to the "net" package using netdev.Use().
//
// Just like a net.Conn, multiple goroutines may invoke methods on a Netdever
// simultaneously.
type Netdever interface {

	// NetConnect device to IP network
	NetConnect() error

	// NetDisconnect device from IP network
	NetDisconnect()

	// NetNotify to register callback for network events
	NetNotify(func(Event))

	// GetHostByName returns the IP address of either a hostname or IPv4
	// address in standard dot notation
	GetHostByName(name string) (IP, error)

	// GetHardwareAddr returns device MAC address
	GetHardwareAddr() (HardwareAddr, error)

	// GetIPAddr returns IP address assigned to device, either by DHCP or
	// statically
	GetIPAddr() (IP, error)

	// Socketer is a Berkely Sockets-like interface
	Socketer
}
