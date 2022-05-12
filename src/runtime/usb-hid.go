//go:build baremetal && usb.hid
// +build baremetal,usb.hid

package runtime

import (
	"machine"
	"machine/usb"
)

func initUSB() {
	// Configure HID interface.
	machine.USB.Configure(usb.HIDConfig{})
}
