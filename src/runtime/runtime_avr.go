// +build avr

package runtime

import (
	"device/avr"
	"machine"
	"unsafe"
)

const BOARD = "arduino"

type timeUnit uint32

var currentTime timeUnit

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

//go:extern _sbss
var _sbss [0]byte

//go:extern _ebss
var _ebss [0]byte

//export main
func main() {
	preinit()
	run()
	abort()
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
}

func initUART() {
	machine.UART0.Configure(machine.UARTConfig{})
}

func putchar(c byte) {
	machine.UART0.WriteByte(c)
}

const asyncScheduler = false

const tickNanos = 1024 * 16384 // roughly 16ms in nanoseconds

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * tickNanos
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / tickNanos)
}

// Sleep this number of ticks of 16ms.
//
// TODO: not very accurate. Improve accuracy by calibrating on startup and every
// once in a while.
func sleepTicks(d timeUnit) {
	currentTime += d
	for d != 0 {
		sleepWDT(WDT_PERIOD_16MS)
		d -= 1
	}
}

func ticks() timeUnit {
	return currentTime
}

func abort() {
	avr.Asm("cli")
	for {
		// Sleeping with interrupts disabled will mean the MCU sleeps forever.
		// To be extra sure, put it in a loop.
		// This mechanism is used by simavr to detect a program exit, the
		// emulator will stop when it detects a sleep instruction with
		// interrupts disabled.
		avr.Asm("sleep")
	}
}
