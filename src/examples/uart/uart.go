// This reads from UART1 and outputs to default serial, usually UART0 or USB.
// Example of how to work with UARTs other than the default.
package main

import (
	"machine"
	"time"
)

var (
	uart = machine.UART1
	tx   = machine.UART1_TX_PIN
	rx   = machine.UART1_RX_PIN
)

func main() {
	uart.Configure(machine.UARTConfig{TX: tx, RX: rx})
	for {
		if uart.Buffered() > 0 {
			data, _ := uart.ReadByte()
			print(string(data))
		}
		time.Sleep(10 * time.Millisecond)
	}
}
