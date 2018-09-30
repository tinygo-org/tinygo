// This is a echo console running on the device UART.
// Connect using default baudrate for this hardware, 8-N-1 with your terminal program.
package main

import (
	"machine"
	"time"
)

func main() {
	machine.UART0.Configure(machine.UARTConfig{})
	machine.UART0.Write([]byte("Echo console enabled. Type something then press enter:\r\n"))

	input := make([]byte, 64)
	i := 0
	for {
		if machine.UART0.Buffered() > 0 {
			data, _ := machine.UART0.ReadByte()

			// Remove high-order bit because 7-bit ascii
			data &^= 0x80

			switch data {
			case 13:
				// return key
				machine.UART0.Write([]byte("\r\n"))
				machine.UART0.Write([]byte("You typed: "))
				machine.UART0.Write(input[:i])
				machine.UART0.Write([]byte("\r\n"))
				i = 0
			default:
				// just echo the character
				machine.UART0.WriteByte(data)
				input[i] = data
				i++
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}
