
// +build nrf

package runtime

// #include "runtime_nrf.h"
import "C"

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
