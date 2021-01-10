package main

import (
	"machine"
	"time"
)

func hw_init() {

	// Led GPIO init
	machine.LED_RED.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// Switch off LEDS
	machine.LED_RED.Low()

	// Enable UART
	uartConsole := &machine.UART0
	uartConsole.Configure(machine.UARTConfig{TX: machine.UART_TX_PIN, RX: machine.UART_TX_PIN, BaudRate: 9600})

}

//----------------------------------------------------------------------------------------------//

func main() {

	println("*** LGT92 demo 2 : Led is toggling every second")

	hw_init()

	for {
		time.Sleep(1 * time.Second)
		machine.LED_RED.Set(!machine.LED_RED.Get())
	}

}
