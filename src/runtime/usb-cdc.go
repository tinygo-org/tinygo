//go:build baremetal && usb.cdc
// +build baremetal,usb.cdc

package runtime

import (
	"machine"
	"machine/usb"
)

func initUSB() {
	// Configure CDC interface.
	machine.USB.Configure(usb.CDCConfig{})
}
