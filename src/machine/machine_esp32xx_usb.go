//go:build esp32c3 || esp32c6 || esp32s3

package machine

import (
	"device/esp"
	"errors"
)

// USB Serial/JTAG Controller
// See esp32-c3_technical_reference_manual_en.pdf
// pg. 736
type USB_DEVICE struct {
	Bus *esp.USB_DEVICE_Type
}

var (
	_USBCDC = &USB_DEVICE{
		Bus: esp.USB_DEVICE,
	}

	USBCDC Serialer = _USBCDC
)

var (
	errUSBWrongSize            = errors.New("USB: invalid write size")
	errUSBCouldNotWriteAllData = errors.New("USB: could not write all data")
	errUSBBufferEmpty          = errors.New("USB: read buffer empty")
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

func initUSB() {}

func (usbdev *USB_DEVICE) Configure(config UARTConfig) error {
	return nil
}

func (usbdev *USB_DEVICE) WriteByte(c byte) error {
	if usbdev.Bus.GetEP1_CONF_SERIAL_IN_EP_DATA_FREE() == 0 {
		return errUSBCouldNotWriteAllData
	}

	usbdev.Bus.SetEP1_RDWR_BYTE(uint32(c))
	usbdev.flush()

	return nil
}

func (usbdev *USB_DEVICE) Write(data []byte) (n int, err error) {
	if len(data) == 0 || len(data) > 64 {
		return 0, errUSBWrongSize
	}

	for i, c := range data {
		if usbdev.Bus.GetEP1_CONF_SERIAL_IN_EP_DATA_FREE() == 0 {
			if i > 0 {
				usbdev.flush()
			}

			return i, errUSBCouldNotWriteAllData
		}
		usbdev.Bus.SetEP1_RDWR_BYTE(uint32(c))
	}

	usbdev.flush()
	return len(data), nil
}

func (usbdev *USB_DEVICE) Buffered() int {
	return int(usbdev.Bus.GetEP1_CONF_SERIAL_OUT_EP_DATA_AVAIL())
}

func (usbdev *USB_DEVICE) ReadByte() (byte, error) {
	if usbdev.Bus.GetEP1_CONF_SERIAL_OUT_EP_DATA_AVAIL() != 0 {
		return byte(usbdev.Bus.GetEP1_RDWR_BYTE()), nil
	}

	return 0, nil
}

func (usbdev *USB_DEVICE) DTR() bool {
	return false
}

func (usbdev *USB_DEVICE) RTS() bool {
	return false
}

func (usbdev *USB_DEVICE) flush() {
	usbdev.Bus.SetEP1_CONF_WR_DONE(1)
	for usbdev.Bus.GetEP1_CONF_SERIAL_IN_EP_DATA_FREE() == 0 {
	}
}
