package usb

import (
	"machine/usb/descriptor"
)

// Controller abstracts the USB interactions to allow for testing without hardware.
type Controller interface {
	Enable()
	ConfigureUSBEndpoint(desc descriptor.Descriptor, epSettings []EndpointConfig, setup []SetupConfig)
	SendUSBInPacket(ep uint32, data []byte) bool
	AckUsbOutTransfer(ep uint32)
	SendZlp()
	IsInitEndpointComplete() bool
	SetStallEPIn(ep uint32)
	SetStallEPOut(ep uint32)
	ClearStallEPIn(ep uint32)
	ClearStallEPOut(ep uint32)
	ReceiveUSBControlPacket() ([7]byte, error)
}

var DefaultController Controller
