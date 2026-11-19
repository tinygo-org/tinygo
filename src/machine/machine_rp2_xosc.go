//go:build rp2040 || rp2350

package machine

import (
	"device/rp"
	"runtime/volatile"
	"unsafe"
)

// On some boards, the XOSC can take longer than usual to stabilize. On such
// boards, this is needed to avoid a hard fault on boot/reset. Refer to
// PICO_XOSC_STARTUP_DELAY_MULTIPLIER in the Pico SDK for additional details.
const XOSC_STARTUP_DELAY_MULTIPLIER = 64

type xoscType struct {
	ctrl     volatile.Register32
	status   volatile.Register32
	dormant  volatile.Register32
	startup  volatile.Register32
	reserved [3 - 3*rp2350ExtraReg]volatile.Register32
	count    volatile.Register32
}

var xosc = (*xoscType)(unsafe.Pointer(rp.XOSC))

// init initializes the crystal oscillator system.
//
// This function will block until the crystal oscillator has stabilised.
func (osc *xoscType) init() {
	xtalHz := xtalFreqHz()
	// Assumes 1-15 MHz input
	if xtalHz < 1*MHz || xtalHz > 15*MHz {
		panic("xosc frequency must be between 1MHz and 15MHz")
	}
	osc.ctrl.Set(rp.XOSC_CTRL_FREQ_RANGE_1_15MHZ)

	// Set xosc startup delay
	delay := ((xtalHz / 1000) + 128) / 256 * XOSC_STARTUP_DELAY_MULTIPLIER
	osc.startup.Set(delay)

	// Set the enable bit now that we have set freq range and startup delay
	osc.ctrl.SetBits(rp.XOSC_CTRL_ENABLE_ENABLE << rp.XOSC_CTRL_ENABLE_Pos)

	// Wait for xosc to be stable
	for !osc.status.HasBits(rp.XOSC_STATUS_STABLE) {
	}
}
