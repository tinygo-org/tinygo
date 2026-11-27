//go:build nxp && mk20d7

package runtime

import "machine"

func putchar(c byte) {
	machine.PutcharUART(machine.UART0, c)
}

func getchar() byte {
	// dummy, TODO
	return 0
}

func buffered() int {
	// dummy, TODO
	return 0
}
