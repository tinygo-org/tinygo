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

const tickMicros = 1024 * 16384

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
var _sbss unsafe.Pointer

//go:extern _ebss
var _ebss unsafe.Pointer

//go:export main
func main() {
	preinit()
	initAll()
	postinit()
	callMain()
	abort()
}

func preinit() {
	// Initialize .bss: zero-initialized global variables.
	ptr := uintptr(unsafe.Pointer(&_sbss))
	for ptr != uintptr(unsafe.Pointer(&_ebss)) {
		*(*uint8)(unsafe.Pointer(ptr)) = 0
		ptr += 1
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
	for {
		sleepWDT(WDT_PERIOD_2S)
	}
}
