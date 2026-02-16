//go:build esp32c3

package machine

import "errors"

var (
	errUSBWrongSize            = errors.New("USB: invalid write size")
	errUSBCouldNotWriteAllData = errors.New("USB: could not write all data")
	errUSBBufferEmpty          = errors.New("USB: read buffer empty")
)

func initUSB() {}

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

func (usbdev *USB_DEVICE) ReadByte() (byte, error) {
	if usbdev.Bus.GetEP1_CONF_SERIAL_OUT_EP_DATA_AVAIL() != 0 {
		return byte(usbdev.Bus.GetEP1_RDWR_BYTE()), nil
	}
	return 0, nil
}

func (usbdev *USB_DEVICE) flush() {
	usbdev.Bus.SetEP1_CONF_WR_DONE(1)
	for usbdev.Bus.GetEP1_CONF_SERIAL_IN_EP_DATA_FREE() == 0 {
	}
}
