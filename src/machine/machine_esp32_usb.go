//go:build esp32s3 || esp32c3

package machine

import (
	"device/esp"
)

type USB_DEVICE struct {
	Bus *esp.USB_DEVICE_Type
}

var (
	_USBCDC = &USB_DEVICE{
		Bus: esp.USB_DEVICE,
	}

	USBCDC Serialer = _USBCDC
)

type Serialer interface {
	WriteByte(c byte) error
	Write(data []byte) (n int, err error)
	Configure(config UARTConfig) error
	Buffered() int
	ReadByte() (byte, error)
	DTR() bool
	RTS() bool
}

func (usbdev *USB_DEVICE) Configure(config UARTConfig) error {
	return nil
}

func (usbdev *USB_DEVICE) Buffered() int {
	return int(usbdev.Bus.GetEP1_CONF_SERIAL_OUT_EP_DATA_AVAIL())
}

func (usbdev *USB_DEVICE) DTR() bool {
	return false
}

func (usbdev *USB_DEVICE) RTS() bool {
	return false
}
