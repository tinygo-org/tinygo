//go:build usb.hid
// +build usb.hid

package machine

import "machine/usb"

var USB = &usb.HID{}

func InitUSB() {
	initUSB()
	USB.Configure(usb.HIDConfig{})
}
