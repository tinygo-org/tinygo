//go:build stm32g0

package machine

import (
	"device/stm32"
	"unsafe"
)

// WindowWatchdog provides access to the Window Watchdog (WWDG) peripheral.
// Unlike IWDG, WWDG must be refreshed within a specific window - not too early
// and not too late. This provides protection against both runaway code and
// code that gets stuck in a loop refreshing the watchdog.
var WindowWatchdog = &windowWatchdogImpl{}

// WindowWatchdogConfig holds configuration for the window watchdog timer.
// The timeout (in microseconds) before the watchdog fires.
// The valid range depends on System frequency.
// At 64MHz: ~64µs to ~524ms
type WindowWatchdogConfig struct {
	TimeoutMicros uint32

	// The window value as a percentage of timeout (0-100).
	// Refresh must occur when counter is below this percentage of max.
	// Default (0) sets window to 100% (no window restriction).
	WindowPercent uint8
}

// WWDG prescaler values
const (
	wwdgPrescaler1   = 0 // CK Counter Clock (PCLK/4096) / 1
	wwdgPrescaler2   = 1 // CK Counter Clock (PCLK/4096) / 2
	wwdgPrescaler4   = 2 // CK Counter Clock (PCLK/4096) / 4
	wwdgPrescaler8   = 3 // CK Counter Clock (PCLK/4096) / 8
	wwdgPrescaler16  = 4 // CK Counter Clock (PCLK/4096) / 16
	wwdgPrescaler32  = 5 // CK Counter Clock (PCLK/4096) / 32
	wwdgPrescaler64  = 6 // CK Counter Clock (PCLK/4096) / 64
	wwdgPrescaler128 = 7 // CK Counter Clock (PCLK/4096) / 128
)

// WWDG counter limits
const (
	wwdgCounterMin = 0x40 // Minimum counter value (T6 must be set)
	wwdgCounterMax = 0x7F // Maximum counter value (7 bits)
	wwdgWindowMax  = 0x7F // Maximum window value
)

type windowWatchdogImpl struct {
	counter   uint8 // Configured counter reload value
	prescaler uint8 // Configured prescaler
}

// Configure the window watchdog.
//
// This method should not be called after the watchdog is started.
// The WWDG cannot be disabled once started, except by a system reset.
//
// Timeout formula: t_WWDG = (1/PCLK) × 4096 × 2^WDGTB × (T[5:0] + 1)
// Where T[5:0] = counter value - 0x40
// Refer RM0444 Rev 6 861/1384
func (wd *windowWatchdogImpl) Configure(config WindowWatchdogConfig) error {
	// Enable WWDG clock
	enableAltFuncClock(unsafe.Pointer(stm32.WWDG))

	// Calculate prescaler and counter value from timeout
	// Base tick = PCLK / 4096
	// With prescaler: tick = PCLK / (4096 * 2^prescaler)
	// Timeout = tick * (counter - 0x3F)

	pclk := CPUFrequency()              // Assuming PCLK = CPU frequency (no APB prescaler)
	baseTick := (4096 * 1000000) / pclk // Base tick in nanoseconds * 1000 for precision

	timeout := config.TimeoutMicros
	if timeout == 0 {
		timeout = 10000 // Default 10ms
	}

	// Find the best prescaler and counter-combination
	var bestPrescaler uint8
	var bestCounter uint8
	found := false

	for prescaler := uint8(0); prescaler <= 7; prescaler++ {
		// Tick duration in nanoseconds * 1000
		tickNs := baseTick << prescaler

		// Counter value needed (counter - 0x3F = timeout / tick)
		// Rearranged: counter = (timeout * 1000 / tickNs) + 0x3F
		counterVal := (uint32(timeout) * 1000000 / tickNs) + 0x3F

		if counterVal >= wwdgCounterMin && counterVal <= wwdgCounterMax {
			bestPrescaler = prescaler
			bestCounter = uint8(counterVal)
			found = true
			break
		}
	}

	if !found {
		// Use maximum timeout
		bestPrescaler = wwdgPrescaler128
		bestCounter = wwdgCounterMax
	}

	wd.prescaler = bestPrescaler
	wd.counter = bestCounter

	// Calculate window value
	windowVal := uint8(wwdgWindowMax)
	if config.WindowPercent > 0 && config.WindowPercent < 100 {
		// Window = 0x40 + ((counter - 0x40) * percent / 100)
		counterRange := uint16(bestCounter) - wwdgCounterMin
		windowOffset := (counterRange * uint16(config.WindowPercent)) / 100
		windowVal = uint8(wwdgCounterMin + windowOffset)
	}
	stm32.WWDG.CFR.Set((uint32(bestPrescaler) << stm32.WWDG_CFR_WDGTB_Pos) | uint32(windowVal))

	return nil
}

// Start enables the window watchdog.
// Once started, the WWDG cannot be disabled except by a system reset.
func (wd *windowWatchdogImpl) Start() error {
	stm32.WWDG.CR.Set(uint32(wd.counter) | (1 << 7))
	return nil
}

// Update refreshes the window watchdog counter.
// This must be called within the configured window to prevent a reset.
// Calling too early (counter > window) or too late (counter <= 0x3F) causes reset.
func (wd *windowWatchdogImpl) Update() {
	stm32.WWDG.CR.Set(uint32(wd.counter) | (1 << 7))
}

// GetCounter returns the current WWDG counter value.
// Useful for timing refresh operations within the window.
func (wd *windowWatchdogImpl) GetCounter() uint8 {
	return uint8(stm32.WWDG.CR.Get() & 0x7F)
}

// EnableEarlyWakeupInterrupt enables the Early Wakeup Interrupt (EWI).
// The EWI is triggered when the counter reaches 0x40, giving the application
// a chance to refresh the watchdog or perform cleanup before reset.
func (wd *windowWatchdogImpl) EnableEarlyWakeupInterrupt() {
	stm32.WWDG.CFR.SetBits(stm32.WWDG_CFR_EWI)
}

// ClearEarlyWakeupFlag clears the Early Wakeup Interrupt flag.
// Must be called in the interrupt handler.
func (wd *windowWatchdogImpl) ClearEarlyWakeupFlag() {
	stm32.WWDG.SR.Set(0) // Write 0 to clear EWIF
}

// IsEarlyWakeupFlagSet returns true if the Early Wakeup Interrupt flag is set.
func (wd *windowWatchdogImpl) IsEarlyWakeupFlagSet() bool {
	return stm32.WWDG.SR.Get()&1 != 0
}

// GetMaxTimeout returns the maximum timeout in microseconds for the current PCLK.
// Max timeout = (1/PCLK) × 4096 × 128 × 64
// At 64MHz: ~524ms = 524288µs
func (wd *windowWatchdogImpl) GetMaxTimeout() uint32 {
	pclk := uint64(CPUFrequency())
	return uint32((uint64(4096) * 128 * 64 * 1000000) / pclk)
}

// GetMinTimeout returns the minimum timeout in microseconds for the current PCLK.
// Min timeout = (1/PCLK) × 4096 × 1 × 1
// At 64MHz: ~64µs
func (wd *windowWatchdogImpl) GetMinTimeout() uint32 {
	pclk := uint64(CPUFrequency())
	return uint32((uint64(4096) * 1000000) / pclk)
}
