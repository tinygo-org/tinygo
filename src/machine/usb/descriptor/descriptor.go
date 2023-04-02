package descriptor

import (
	"runtime/volatile"
)

const (
	TypeDevice                = 0x1
	TypeConfiguration         = 0x2
	TypeString                = 0x3
	TypeInterface             = 0x4
	TypeEndpoint              = 0x5
	TypeDeviceQualifier       = 0x6
	TypeInterfaceAssociation  = 0xb
	TypeClassHID              = 0x21
	TypeHIDReport             = 0x22
	TypeClassSpecific         = 0x24
	TypeClassSpecificEndpoint = 0x25
)

// DeviceDescBank is the USB device endpoint .
type DeviceDescBank struct {
	ADDR      volatile.Register32
	PCKSIZE   volatile.Register32
	EXTREG    volatile.Register16
	STATUS_BK volatile.Register8
	_reserved [5]volatile.Register8
}

type Device struct {
	DeviceDescBank [2]DeviceDescBank
}

type Descriptor struct {
	Device        []byte
	Configuration []byte
	HID           map[uint16][]byte
}

func (d *Descriptor) Configure(idVendor, idProduct uint16) {
	dev := DeviceType{d.Device}
	dev.VendorID(idVendor)
	dev.ProductID(idProduct)

	conf := ConfigurationType{d.Configuration}
	conf.TotalLength(uint16(len(d.Configuration)))
}

func appendSlices[T any](slices [][]T) []T {
	var size, pos int

	for _, s := range slices {
		size += len(s)
	}

	result := make([]T, size)

	for _, s := range slices {
		pos += copy(result[pos:], s)
	}

	return result
}
