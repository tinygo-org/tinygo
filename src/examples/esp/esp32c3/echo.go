package main

import (
	"machine"
	"time"
)

func main() {

	uartConfig := &machine.UARTConfig{
		BaudRate: 115200,
		TX:       machine.Pin(2),
		RX:       machine.Pin(5),
	}
	uart := machine.UART1
	uart.Configure(uartConfig)

	uart.Write([]byte("Echo console enabled. Type something then press enter:\r\n"))
	input := make([]byte, 64)
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
