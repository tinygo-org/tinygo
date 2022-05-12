//go:build usb.hid
// +build usb.hid

package machine

import "machine/usb"

var USB = &usb.HID{}

func InitUSB(port int) {
	InitUSBIO()
	USB.Port = port
	USB.Configure(usb.HIDConfig{})
}
