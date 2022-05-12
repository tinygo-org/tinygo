//go:build usb.cdc
// +build usb.cdc

package machine

import "machine/usb"

var USB = &usb.CDC{}

func InitUSB(port int) {
	InitUSBIO()
	USB.Port = port
	USB.Configure(usb.CDCConfig{})
}
