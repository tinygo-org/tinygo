package main

import (
	"machine"
	"time"
)

func main() {
	serial := machine.UART0
	serial.Configure(machine.UARTConfig{BaudRate: 9600})
	for {
		serial.Write([]byte("hello tinyGo!\n"))
		time.Sleep(time.Millisecond * 1000)
	}
}
