// +build output.uart

package runtime

import "machine"

// The "uart" output writes all output over a serial (UART) or USB-CDC
// connection. Most, but not all, microcontrollers have an UART available.

func initOutput() {
	machine.UART0.Configure(machine.UARTConfig{})
}

func putchar(c byte) {
	machine.UART0.WriteByte(c)
}
