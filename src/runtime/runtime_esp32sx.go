//go:build esp32s3

package runtime

import (
	"device/esp"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

//type timeUnit int64

func putchar(c byte) {
	machine.Serial.WriteByte(c)
}

func getchar() byte {
	for machine.Serial.Buffered() == 0 {
		Gosched()
	}
	v, _ := machine.Serial.ReadByte()
	return v
}

func buffered() int {
	return machine.Serial.Buffered()
}

// Initialize .bss: zero-initialized global variables.
// The .data section has already been loaded by the ROM bootloader.
func clearbss() {
	ptr := unsafe.Pointer(&_sbss)
	for ptr != unsafe.Pointer(&_ebss) {
		*(*uint32)(ptr) = 0
		ptr = unsafe.Add(ptr, 4)
	}
}

func initTimer() {
	// Configure timer 0 in timer group 0, for timekeeping.
	//   EN:       Enable the timer.
	//   INCREASE: Count up every tick (as opposed to counting down).
	//   DIVIDER:  16-bit prescaler, set to 2 for dividing the APB clock by two
	//             (40MHz).
	// esp.TIMG0.T0CONFIG.Set(0 << esp.TIMG_T0CONFIG_T0_EN_Pos)
	esp.TIMG0.T0CONFIG.Set(esp.TIMG_TCONFIG_EN | esp.TIMG_TCONFIG_INCREASE | 2<<esp.TIMG_TCONFIG_DIVIDER_Pos)
	// esp.TIMG0.T0CONFIG.Set(1 << esp.TIMG_T0CONFIG_T0_DIVCNT_RST_Pos)
	// esp.TIMG0.T0CONFIG.Set(esp.TIMG_T0CONFIG_T0_EN)

	// Set the timer counter value to 0.
	esp.TIMG0.T0LOADLO.Set(0)
	esp.TIMG0.T0LOADHI.Set(0)
	esp.TIMG0.T0LOAD.Set(0) // value doesn't matter.
}

func ticks() timeUnit {
	// First, update the LO and HI register pair by writing any value to the
	// register. This allows reading the pair atomically.
	esp.TIMG0.T0UPDATE.Set(1)
	for esp.TIMG0.T0UPDATE.Get() != 0 {
		// Register is cleared when the update is complete.
	}
	// Then read the two 32-bit parts of the timer.
	return timeUnit(uint64(esp.TIMG0.T0LO.Get()) | uint64(esp.TIMG0.T0HI.Get())<<32)
}

func nanosecondsToTicks(ns int64) timeUnit {
	// Calculate the number of ticks from the number of nanoseconds. At a 80MHz
	// APB clock, that's 25 nanoseconds per tick with a timer prescaler of 2:
	// 25 = 1e9 / (80MHz / 2)
	return timeUnit(ns / 25)
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	// See nanosecondsToTicks.
	return int64(ticks) * 25
}

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

	// Map the TIMG0 T0 peripheral interrupt to a CPU interrupt line.
	esp.INTERRUPT_CORE0.SetTG_T0_INT_MAP(timerAlarmCPUInterrupt)

	// Register the interrupt handler and enable it once.
	timerAlarmInterrupt = interrupt.New(timerAlarmCPUInterrupt, timerAlarmHandler)
	timerAlarmInterrupt.Enable()
}

// sleepTicks spins until the given number of ticks have elapsed, using the
// TIMG0 alarm interrupt to avoid busy-waiting for the entire duration.
func sleepTicks(d timeUnit) {
	machine.FlushSerial()
	target := ticks() + d
	for ticks() < target {
		// Set the alarm to fire at the target tick count.
		interruptPending.Set(0)

		esp.TIMG0.T0ALARMLO.Set(uint32(target))
		esp.TIMG0.T0ALARMHI.Set(uint32(target >> 32))

		// Enable the alarm (auto-clears when alarm fires).
		esp.TIMG0.T0CONFIG.SetBits(esp.TIMG_TCONFIG_ALARM_EN)

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

func exit(code int) {
	abort()
}
