package main

import (
	"machine"
	"time"
)

func main() {
	machine.UART0.Configure(machine.UARTConfig{})
	machine.UART0.Write([]byte("UART echo enabled...\r\n"))

	for {
		input := make([]byte, 10)
		n, err := machine.UART0.Read(input)
		if err != nil {
			machine.UART0.Write([]byte("input error!\r\n"))
		} else {
			if n > 0 {
				machine.UART0.Write([]byte("You typed:\r\n"))
				for _, v := range input {
					// Remove high-order bit because 7-bit ascii
					machine.UART0.WriteByte(v &^ (1 << 7))
				}

				machine.UART0.Write([]byte("\r\n"))
			}
		}
		time.Sleep(time.Second)
	}
}
