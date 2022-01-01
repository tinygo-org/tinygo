//go:build avr
// +build avr

package runtime

import (
	"device/avr"
	"machine"
	"runtime/interrupt"
	"unsafe"
)

const BOARD = "arduino"

// timeUnit in nanoseconds
type timeUnit int64

// Watchdog timer periods. These can be off by a large margin (hence the jump
// between 64ms and 125ms which is not an exact double), so don't rely on this
// for accurate time keeping.
const (
	WDT_PERIOD_16MS = iota
	WDT_PERIOD_32MS
	WDT_PERIOD_64MS
	WDT_PERIOD_125MS
	WDT_PERIOD_250MS
	WDT_PERIOD_500MS
	WDT_PERIOD_1S
	WDT_PERIOD_2S
)

const timerRecalibrateInterval = 6e7 // 1 minute

var nextTimerRecalibrate timeUnit

//go:extern _sbss
var _sbss [0]byte

//go:extern _ebss
var _ebss [0]byte

//export main
func main() {
	preinit()
	run()
	exit(0)
}

func preinit() {
	// Initialize .bss: zero-initialized global variables.
	ptr := unsafe.Pointer(&_sbss)
	for ptr != unsafe.Pointer(&_ebss) {
		*(*uint8)(ptr) = 0
		ptr = unsafe.Pointer(uintptr(ptr) + 1)
	}
}

func postinit() {
	// Enable interrupts after initialization.
	avr.Asm("sei")
}

func init() {
	initUART()
	machine.InitMonotonicTimer()
	nextTimerRecalibrate = ticks() + timerRecalibrateInterval
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns)
}

// Sleep this number of ticks of nanoseconds.
func sleepTicks(d timeUnit) {
	waitTill := ticks() + d
	// recalibrate if we have some time (>100ms) and it was a while when we did it last time.
	if d > 100000 {
		now := waitTill - d
		if nextTimerRecalibrate < now {
			nextTimerRecalibrate = now + timerRecalibrateInterval
			machine.AdjustMonotonicTimer()
		}
	}
	for {
		// wait for interrupt
		avr.Asm("sleep")
		if waitTill <= ticks() {
			// done waiting
			return
		}
		if hasScheduler {
			// The interrupt may have awoken a goroutine, so bail out early.
			return
		}
	}
}

// ticks return time since start in nanoseconds
func ticks() (ticks timeUnit) {
	state := interrupt.Disable()
	ticks = timeUnit(machine.Ticks)
	interrupt.Restore(state)
	return
}

func exit(code int) {
	abort()
}

func abort() {
	// Disable interrupts and go to sleep.
	// This can never be awoken except for reset, and is recogized as termination by simavr.
	avr.Asm("cli")
	for {
		avr.Asm("sleep")
	}
}
