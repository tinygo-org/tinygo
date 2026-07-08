//go:build esp32

package runtime

import (
	"device/esp"
	"runtime/interrupt"
	"runtime/volatile"
)

// CPU interrupt number used for the TIMG0 timer alarm.
const timerAlarmCPUInterrupt = 9

var interruptPending volatile.Register8

func signalInterrupt() {
	interruptPending.Set(1)
}

var timerAlarmInterrupt interrupt.Interrupt

// timerAlarmHandler clears the timer interrupt at the peripheral level
// and disables INT_ENA to prevent level-triggered re-assertion.
func timerAlarmHandler(interrupt.Interrupt) {
	esp.TIMG0.INT_ENA_TIMERS.ClearBits(1)
	esp.TIMG0.INT_CLR_TIMERS.Set(1)
}

// initTimerInterrupt routes the TIMG0 timer 0 alarm interrupt to a CPU
// interrupt and registers a handler that clears the alarm flag.
func initTimerInterrupt() {
	// Clear any stale timer interrupt before enabling.
	esp.TIMG0.INT_CLR_TIMERS.Set(1)

	// Map the TIMG0 T0 peripheral interrupt to a CPU interrupt line
	// via the DPORT interrupt matrix.
	esp.DPORT.PRO_TG_T0_LEVEL_INT_MAP.Set(timerAlarmCPUInterrupt)

	// Register the interrupt handler and enable it once.
	timerAlarmInterrupt = interrupt.New(timerAlarmCPUInterrupt, timerAlarmHandler)
	timerAlarmInterrupt.Enable()
}

// sleepTicks spins until the given number of ticks have elapsed, using the
// TIMG0 alarm interrupt to avoid busy-waiting for the entire duration.
func sleepTicks(d timeUnit) {
	target := ticks() + d
	for ticks() < target {
		// Set the alarm to fire at the target tick count.
		interruptPending.Set(0)

		esp.TIMG0.T0ALARMLO.Set(uint32(target))
		esp.TIMG0.T0ALARMHI.Set(uint32(target >> 32))

		// Enable the alarm (auto-clears when alarm fires).
		esp.TIMG0.T0CONFIG.SetBits(esp.TIMG_T0CONFIG_ALARM_EN)

		// Re-enable the timer interrupt (handler disables INT_ENA).
		esp.TIMG0.INT_CLR_TIMERS.Set(1)
		esp.TIMG0.INT_ENA_TIMERS.SetBits(1)

		// Wait for any interrupt (timer alarm or other) or timeout.
		for interruptPending.Get() == 0 {
			if ticks() >= target {
				return
			}
		}
	}
}
