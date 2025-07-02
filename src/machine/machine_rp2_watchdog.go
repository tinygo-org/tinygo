//go:build rp2040 || rp2350

package machine

import (
	"device/rp"
)

// Watchdog provides access to the hardware watchdog available
// in the RP2040.
var Watchdog = &watchdogImpl{}

const (
	// WatchdogMaxTimeout in milliseconds (approx 8.3s).
	//
	// Nominal 1us per watchdog tick, 24-bit counter,
	// but due to errata two ticks consumed per 1us.
	// See: Errata RP2040-E1
	WatchdogMaxTimeout = (rp.WATCHDOG_LOAD_LOAD_Msk / 1000) / 2
)

type watchdogImpl struct {
	// The value to reset the counter to on each Update
	loadValue uint32
}

// Configure the watchdog.
//
// This method should not be called after the watchdog is started and on
// some platforms attempting to reconfigure after starting the watchdog
// is explicitly forbidden / will not work.
func (wd *watchdogImpl) Configure(config WatchdogConfig) error {
	// x2 due to errata RP2040-E1
	wd.loadValue = config.TimeoutMillis * 1000 * 2
	if wd.loadValue > rp.WATCHDOG_LOAD_LOAD_Msk {
		wd.loadValue = rp.WATCHDOG_LOAD_LOAD_Msk
	}

	rp.WATCHDOG.CTRL.ClearBits(rp.WATCHDOG_CTRL_ENABLE)

	// Reset everything apart from ROSC and XOSC
	rp.PSM.WDSEL.Set(0x0001ffff &^ (rp.PSM_WDSEL_ROSC | rp.PSM_WDSEL_XOSC))

	// Pause watchdog during debug
	rp.WATCHDOG.CTRL.SetBits(rp.WATCHDOG_CTRL_PAUSE_DBG0 | rp.WATCHDOG_CTRL_PAUSE_DBG1 | rp.WATCHDOG_CTRL_PAUSE_JTAG)

	// Load initial counter
	rp.WATCHDOG.LOAD.Set(wd.loadValue)

	return nil
}

// Starts the watchdog.
func (wd *watchdogImpl) Start() error {
	rp.WATCHDOG.CTRL.SetBits(rp.WATCHDOG_CTRL_ENABLE)
	return nil
}

// Update the watchdog, indicating that the app is healthy.
func (wd *watchdogImpl) Update() {
	rp.WATCHDOG.LOAD.Set(wd.loadValue)
}
