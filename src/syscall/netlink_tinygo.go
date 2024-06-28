//go:build tinygo

package syscall

// This is (mostly) a copy of src/syscall/netlink_linux.go.  netlink_linux.go
// imports "sync", but that causes an import cycle:
//
// package tinygo.org/x/drivers/examples/net/tcpclient
//        imports bytes
//        imports io
//        imports sync
//        imports internal/task
//        imports runtime/interrupt
//        imports device/arm
//        imports syscall
//        imports sync: import cycle not allowed
//
// So until we can solve the import cycle, stub out things.

// NetlinkMessage represents a netlink message.
type NetlinkMessage struct {
	Header NlMsghdr
	Data   []byte
}

// NetlinkRouteAttr represents a netlink route attribute.
type NetlinkRouteAttr struct {
	Attr  RtAttr
	Value []byte
}

// NetlinkRIB returns routing information base, as known as RIB, which
// consists of network facility information, states and parameters.
func NetlinkRIB(proto, family int) ([]byte, error) {
	println("syscall.NetlinkRIB not implemented", proto, family)
	return []byte{}, EOPNOTSUPP
}

// ParseNetlinkMessage parses b as an array of netlink messages and
// returns the slice containing the NetlinkMessage structures.
func ParseNetlinkMessage(b []byte) ([]NetlinkMessage, error) {
	println("syscall.ParseNetlinkMessage not implemented", b)
	return []NetlinkMessage{}, EOPNOTSUPP
}

// ParseNetlinkRouteAttr parses m's payload as an array of netlink
// route attributes and returns the slice containing the
// NetlinkRouteAttr structures.
func ParseNetlinkRouteAttr(m *NetlinkMessage) ([]NetlinkRouteAttr, error) {
	println("syscall.ParseNetlinkRouteAttr not implemented", m)
	return []NetlinkRouteAttr{}, EOPNOTSUPP
}
