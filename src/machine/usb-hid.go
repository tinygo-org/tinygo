//go:build baremetal && usb.hid
// +build baremetal,usb.hid

package machine

import "machine/usb"

var USB = &usb.HID{Port: 0}

func init() {
	// Configure USB D+/D- pins.
	USBCDC_DM_PIN.Configure(PinConfig{Mode: PinCom})
	USBCDC_DP_PIN.Configure(PinConfig{Mode: PinCom})
}
