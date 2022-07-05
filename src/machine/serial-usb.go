//go:build baremetal && serial.usb
// +build baremetal,serial.usb

package machine

// Serial is implemented via USB (USB-CDC).
var Serial = USB

func InitSerial() {
}
