//go:build baremetal && serial.usb
// +build baremetal,serial.usb

package machine

var Serial Serialer

func InitSerial() {
	Serial = USBCDC
}
