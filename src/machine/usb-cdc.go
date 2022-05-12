//go:build baremetal && usb.cdc
// +build baremetal,usb.cdc

package machine

import "machine/usb"

var USB = &usb.CDC{Port: 0}

func init() {
	// Configure USB D+/D- pins.
	USBCDC_DM_PIN.Configure(PinConfig{Mode: PinCom})
	USBCDC_DP_PIN.Configure(PinConfig{Mode: PinCom})
}
