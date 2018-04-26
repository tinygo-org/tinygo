
// +build nrf

package runtime

// #include "runtime_nrf.h"
import "C"

func init() {
	C.uart_init(6) // pin_tx = 6, for NRF52840-DK
}

const Microsecond = 1

func putchar(c byte) {
	C.uart_send(C.uint8_t(c))
}

func Sleep(d Duration) {
	// TODO
}

func abort() {
	// TODO: wfi
	for {
	}
}
