//go:build baremetal && serial.usb

package machine

// Serial is implemented via USB (USB-CDC).
var Serial Serialer

func InitSerial() {
	Serial = USBCDC
}
