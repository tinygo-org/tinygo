//go:build avr && attiny
// +build avr,attiny

package runtime

import (
	"device/avr"
)

func initUART() {
}

func putchar(c byte) {
	// UART is not supported.
}

func sleepWDT(period uint8) {
	// TODO: use the watchdog timer instead of a busy loop.
	for i := 0x45; i != 0; i-- {
		for i := 0xff; i != 0; i-- {
			avr.Asm("nop")
		}
	}
}
