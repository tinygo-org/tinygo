//go:build stm32

package machine

import "device/stm32"

var (
	Watchdog = &watchdogImpl{}
)

const (
	// WatchdogMaxTimeout in milliseconds (32.768s)
	//
	// Timeout is based on 12-bit counter with /256 divider on
	// 32.768kHz clock.  See 21.3.3 of RM0090 for table.
	WatchdogMaxTimeout = ((0xfff + 1) * 256 * 1024) / 32768
)

const (
	// Enable access to PR, RLR and WINR registers (0x5555)
	iwdgKeyEnable = 0x5555
	// Reset the watchdog value (0xAAAA)
	iwdgKeyReset = 0xaaaa
	// Start the watchdog (0xCCCC)
	iwdgKeyStart = 0xcccc
	// Divide by 256
	iwdgDiv256 = 6
)

type watchdogImpl struct {
}

// Configure the watchdog.
//
// This method should not be called after the watchdog is started and on
// some platforms attempting to reconfigure after starting the watchdog
// is explicitly forbidden / will not work.
func (wd *watchdogImpl) Configure(config WatchdogConfig) error {

	// Enable configuration of IWDG
	stm32.IWDG.KR.Set(iwdgKeyEnable)

	// Unconditionally divide by /256 since we don't really need accuracy
	// less than 8ms
	stm32.IWDG.PR.Set(iwdgDiv256)

	timeout := config.TimeoutMillis
	if timeout > WatchdogMaxTimeout {
		timeout = WatchdogMaxTimeout
	}

	// Set reload value based on /256 divider
	stm32.IWDG.RLR.Set(((config.TimeoutMillis*32768 + (256 * 1024) - 1) / (256 * 1024)) - 1)
	return nil
}

// Starts the watchdog.
func (wd *watchdogImpl) Start() error {
	stm32.IWDG.KR.Set(iwdgKeyStart)
	return nil
}

// Update the watchdog, indicating that `source` is healthy.
func (wd *watchdogImpl) Update() {
	stm32.IWDG.KR.Set(iwdgKeyReset)
}
