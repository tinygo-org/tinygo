// +build avr

package runtime

import (
	"device/avr"
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
	// Initialize UART at 115200 baud when running at 16MHz.
	*avr.UBRR0H = 0
	*avr.UBRR0L = 8
	*avr.UCSR0B = avr.UCSR0B_RXEN0 | avr.UCSR0B_TXEN0   // enable RX and TX
	*avr.UCSR0C = avr.UCSR0C_UCSZ01 | avr.UCSR0C_UCSZ00 // 8-bits data
}

func putchar(c byte) {
	for (*avr.UCSR0A & avr.UCSR0A_UDRE0) == 0 {
		// Wait until previous char has been sent.
	}
	*avr.UDR0 = avr.RegValue(c) // send char
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

// Sleep for a given period. The period is defined by the WDT peripheral, and is
// on most chips (at least) 3 bits wide, in powers of two from 16ms to 2s
// (0=16ms, 1=32ms, 2=64ms...). Note that the WDT is not very accurate: it can
// be off by a large margin depending on temperature and supply voltage.
//
// TODO: disable more peripherals etc. to reduce sleep current.
func sleepWDT(period uint8) {
	// Configure WDT
	avr.Asm("cli")
	avr.Asm("wdr")
	// Start timed sequence.
	*avr.WDTCSR |= avr.WDTCSR_WDCE | avr.WDTCSR_WDE
	// Enable WDT and set new timeout (0.5s)
	*avr.WDTCSR = avr.WDTCSR_WDIE | avr.RegValue(period)
	avr.Asm("sei")

	// Set sleep mode to idle and enable sleep mode.
	// Note: when using something other than idle, the UART won't work
	// correctly. This needs to be fixed, though, so we can truly sleep.
	*avr.SMCR = (0 << 1) | avr.SMCR_SE

	// go to sleep
	avr.Asm("sleep")

	// disable sleep
	*avr.SMCR = 0
}

func ticks() timeUnit {
	return currentTime
}

func abort() {
	avr.Asm("cli")
	for {
		sleepWDT(WDT_PERIOD_2S)
	}
}

// Align on a word boundary.
func align(ptr uintptr) uintptr {
	// No alignment necessary on the AVR.
	return ptr
}
