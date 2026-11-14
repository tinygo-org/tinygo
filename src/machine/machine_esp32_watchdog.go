//go:build esp32

package machine

import (
	"device/esp"
)

// Watchdog uses the timer group 1 watchdog (MWDT1), which is free after boot.
// See ESP32 TRM v5.9 chapter 11 Watchdog Timers.
var Watchdog = &watchdogImpl{}

const (
	// WatchdogMaxTimeout in milliseconds (about 24.8 days).
	WatchdogMaxTimeout = 0xffffffff / 2

	wdtWriteKey = 0x50D83AA1

	// 80 MHz APB clock / 40000 gives a 0.5 ms tick, the same as ESP-IDF.
	wdtPrescaler  = 40000
	wdtTicksPerMs = 2

	// Stage action 3 resets the main system, length 7 is a 3.2 µs pulse.
	// See ESP32 TRM v5.9 register 10.10 TIMGn_WDTCONFIG0_REG.
	wdtStageResetSystem = 3
	wdtResetLength      = 7
)

type watchdogImpl struct{}

// Configure the watchdog.
//
// This method should not be called after the watchdog is started.
func (wd *watchdogImpl) Configure(config WatchdogConfig) error {
	ms := config.TimeoutMillis
	if ms > WatchdogMaxTimeout {
		ms = WatchdogMaxTimeout
	}

	esp.TIMG1.WDTWPROTECT.Set(wdtWriteKey)
	esp.TIMG1.WDTCONFIG0.Set(wdtResetLength<<esp.TIMG_WDTCONFIG0_WDT_SYS_RESET_LENGTH_Pos |
		wdtResetLength<<esp.TIMG_WDTCONFIG0_WDT_CPU_RESET_LENGTH_Pos)
	esp.TIMG1.WDTCONFIG1.Set(wdtPrescaler << esp.TIMG_WDTCONFIG1_WDT_CLK_PRESCALE_Pos)
	esp.TIMG1.WDTCONFIG2.Set(ms * wdtTicksPerMs)
	esp.TIMG1.WDTWPROTECT.Set(0)
	return nil
}

// Starts the watchdog.
func (wd *watchdogImpl) Start() error {
	esp.TIMG1.WDTWPROTECT.Set(wdtWriteKey)
	esp.TIMG1.WDTFEED.Set(1)
	esp.TIMG1.SetWDTCONFIG0_WDT_STG0(wdtStageResetSystem)
	esp.TIMG1.SetWDTCONFIG0_WDT_EN(1)
	esp.TIMG1.WDTWPROTECT.Set(0)
	return nil
}

// Update the watchdog, indicating that the app is healthy.
func (wd *watchdogImpl) Update() {
	esp.TIMG1.WDTWPROTECT.Set(wdtWriteKey)
	esp.TIMG1.WDTFEED.Set(1)
	esp.TIMG1.WDTWPROTECT.Set(0)
}
