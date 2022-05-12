//go:build usb.cdc
// +build usb.cdc

package machine

import "machine/usb"

var USB = &usb.CDC{}

func InitUSB() {
	initUSB()
	USB.Configure(usb.CDCConfig{})
}
