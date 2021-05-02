// This is a echo console running on the device UART.
// Connect using default baudrate for this hardware, 8-N-1 with your terminal program.
package main

import (
	"machine"
	"time"
)

func main() {
	uart := machine.UART0
	uart.Write([]byte("Echo console enabled. Type something then press enter:\r\n"))

	input := make([]byte, 4096)
	i := 0
	for {
		if uart.Buffered() > 0 {
			data, _ := uart.ReadByte()

			switch data {
			case 13:
				// return key
				uart.Write([]byte("\r\n"))
				uart.Write([]byte("You typed: "))
				uart.Write(input[:i])
				uart.Write([]byte("\r\n"))
				i = 0
			default:
				// just echo the character
				uart.WriteByte(data)
				input[i] = data
				i++
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}
