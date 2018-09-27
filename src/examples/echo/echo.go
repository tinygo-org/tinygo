// This is a echo console running on the device UART.
// Connect using 57600 baud, 8-N-1 with your terminal program.
package main

import (
	"machine"
	"time"
)

func main() {
	// Set baudrate to 56k for an Arduino Uno to be error-free.
	machine.UART0.Configure(machine.UARTConfig{Baudrate: 57600})
	machine.UART0.Write([]byte("Echo console enabled. Type something then press enter:\r\n"))

	input := make([]byte, 64)
	i := 0
	for {
		data, err := machine.UART0.ReadByte()
		if err != nil {
			machine.UART0.Write([]byte("input error!\r\n"))
		} else {
			// Remove high-order bit because 7-bit ascii
			data &^= 0x80

			switch data {
			case 0:
				break
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
