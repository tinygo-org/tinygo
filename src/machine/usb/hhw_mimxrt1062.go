// +build mimxrt1062

package usb

// Implementation of USB host controller driver (hcd) for NXP iMXRT1062.

import (
	"device/nxp"
	"runtime/interrupt"
)

// hcdInterruptPriority defines the priority for all USB host interrupts.
const hcdInterruptPriority = 3

// hcd implements USB host controller driver (hcd) interface.
type hhw struct {
	*hcd // USB host controller driver

	bus *nxp.USB_Type       // USB core register
	phy *nxp.USBPHY_Type    // USB PHY register
	irq interrupt.Interrupt // USB IRQ, only a single interrupt on iMXRT1062
}

// allocHHW returns a reference to the USB hardware abstraction for the given
// host controller driver. Should be called only one time and during host
// controller initialization.
func allocHHW(port, instance int, hc *hcd) *hhw {
	switch port {
	case 0:
		hhwInstance[instance].hcd = hc
		hhwInstance[instance].bus = nxp.USB1
		hhwInstance[instance].phy = nxp.USBPHY1

	case 1:
		hhwInstance[instance].hcd = hc
		hhwInstance[instance].bus = nxp.USB2
		hhwInstance[instance].phy = nxp.USBPHY2
	}

	return &hhwInstance[instance]
}

// init configures the USB port for host mode operation by initializing all
// endpoint and transfer descriptor data structures, initializing core registers
// and interrupts, resetting the USB PHY, and enabling power on the bust.
func (h *hhw) init() status {

	return statusOK
}

// enable causes the USB core to enter (or exit) the normal run state and
// enables/disables all interrupts on the receiver's USB port.
func (h *hhw) enable(enable bool) {
	if enable {
		h.irq.Enable() // Enable USB interrupts
	} else {
		h.irq.Disable() // Disable USB interrupts
	}
}
