//go:build tinygo
// +build tinygo

package net

const (
	tinyGoInterfaceName = "tinygo-undefined"
	maxTransmissionUnit = 1500 // Ethernet?
)

// DE:AD:BE:EF:FE:FF
var defaultMAC = HardwareAddr{0xDE, 0xAD, 0xBE, 0xEF, 0xFE, 0xFF}

// If the ifindex is zero, interfaceTable returns mappings of all
// network interfaces. Otherwise it returns a mapping of a specific
// interface.
func interfaceTable(ifindex int) ([]Interface, error) {
	i, err := readInterface(0)
	if err != nil {
		return nil, err
	}
	return []Interface{*i}, nil
}

func readInterface(i int) (*Interface, error) {
	if i != 0 {
		return nil, errInvalidInterfaceIndex
	}
	ifc := &Interface{
		Index:        i + 1, // Offset the index by one to suit the contract
		Name:         tinyGoInterfaceName,
		MTU:          maxTransmissionUnit,
		HardwareAddr: defaultMAC,
		Flags:        0, // No flags since interface is not implemented.
	}
	return ifc, nil
}

func interfaceCount() (int, error) {
	return 1, nil
}

// If the ifi is nil, interfaceAddrTable returns addresses for all
// network interfaces. Otherwise it returns addresses for a specific
// interface.
func interfaceAddrTable(ifi *Interface) ([]Addr, error) {
	return nil, nil
}

// interfaceMulticastAddrTable returns addresses for a specific
// interface.
func interfaceMulticastAddrTable(ifi *Interface) ([]Addr, error) {
	return nil, nil
}
