//go:build esp32s3

package machine

import "errors"

func initUSB() {
	USBCDC.Configure(UARTConfig{BaudRate: 115200})
}

func (usbdev *USB_DEVICE) WriteByte(c byte) error {
	timeout := 10000
	for usbdev.Bus.GetEP1_CONF_SERIAL_IN_EP_DATA_FREE() == 0 && timeout > 0 {
		timeout--
	}
	if timeout == 0 {
		return nil
	}
	usbdev.Bus.SetEP1_RDWR_BYTE(uint32(c))
	usbdev.Bus.SetEP1_CONF_WR_DONE(1)
	return nil
}

func (usbdev *USB_DEVICE) Write(data []byte) (n int, err error) {
	for _, c := range data {
		err = usbdev.WriteByte(c)
		if err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func (usbdev *USB_DEVICE) ReadByte() (byte, error) {
	return 0, errors.New("ReadByte not implemented")
}

func (usbdev *USB_DEVICE) flush() {
	usbdev.Bus.SetEP1_CONF_WR_DONE(1)
}
