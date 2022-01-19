//go:build (sam && atsamd51) || (sam && atsame5x)
// +build sam,atsamd51 sam,atsame5x

package usb

// Implementation of USB host controller driver (hcd) for Microchip SAMD51.

import (
	"device/sam"
	"runtime/interrupt"
)

// hhwInterruptPriority defines the priority for all USB host interrupts.
const hhwInterruptPriority = 3

// hhw implements USB host controller hardware abstraction interface.
type hhw struct {
	*hcd // USB host controller driver

	bus *sam.USB_HOST_Type  // USB core registers
	irq interrupt.Interrupt // USB IRQ, only a single interrupt on SAMx51

	speed Speed
}

// allocHHW returns a reference to the USB hardware abstraction for the given
// host controller driver. Should be called only one time and during host
// controller initialization.
func allocHHW(port, instance int, speed Speed, hc *hcd) *hhw {
	switch port {
	case 0:
		hhwInstance[instance].hcd = hc
		hhwInstance[instance].bus = sam.USB_HOST
	}

	// Port defaults to full-speed (12 Mbit/sec) on SAMx51
	if 0 == speed {
		speed = HighSpeed
	}
	hhwInstance[instance].speed = speed

	return &hhwInstance[instance]
}

// init configures the USB port for host mode operation by initializing all
// endpoint and transfer descriptor data structures, initializing core registers
// and interrupts, resetting the USB PHY, and enabling power on the bus.
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
