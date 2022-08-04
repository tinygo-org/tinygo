//go:build rp2040
// +build rp2040

package machine

import (
	"runtime/volatile"
	"unsafe"

	"tinygo.org/x/device/rp"
)

type watchdogType struct {
	ctrl    volatile.Register32
	load    volatile.Register32
	reason  volatile.Register32
	scratch [8]volatile.Register32
	tick    volatile.Register32
}

var watchdog = (*watchdogType)(unsafe.Pointer(rp.WATCHDOG))

// startTick starts the watchdog tick.
// cycles needs to be a divider that when applied to the xosc input,
// produces a 1MHz clock. So if the xosc frequency is 12MHz,
// this will need to be 12.
func (wd *watchdogType) startTick(cycles uint32) {
	wd.tick.Set(cycles | rp.WATCHDOG_TICK_ENABLE)
}
