package main

import (
	"machine"
	"time"
)

var uartGps, uartConsole *machine.UART

// serialGps() function handles communication with GPS Module UART
func serialGps(uart *machine.UART) string {
	input := make([]byte, 300)

	i := 0

	for {

		if i == 300 {
			println("Serial Buffer overrun")
			i = 0
		}
		if uart.Buffered() > 0 {

			data, _ := uart.ReadByte() // read a character

			switch data {
			case '\n': // We received a Frame
			case '\r':
				cmd := string(input[:i])
				i = 0
				println(cmd)
			//	uartConsole.Write([]byte(cmd))
			//	uartConsole.Write([]byte("\r\n"))
			default: // pressed any other key
				machine.LED_RED.Set(!machine.LED_RED.Get())
				input[i] = data
				i++
			}
		}

		//time.Sleep(3 * time.Millisecond)
	}

}

//----------------------------------------------------------------------------------------------//

func main() {

	// Turn RED led for 1sec
	machine.LED_RED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.LED_RED.Set(true)
	time.Sleep(1000 * time.Millisecond)
	machine.LED_RED.Set(false)

	// Enable UART0 (Console)
	uartConsole := &machine.UART0
	uartConsole.Configure(machine.UARTConfig{TX: machine.UART_TX_PIN, RX: machine.UART_TX_PIN, BaudRate: 9600})
	println("*** LGT92 demo 3 : Enable GPS and read sentences")

	// Enable UART1 (GPS)
	uartGps = &machine.UART1
	uartGps.Configure(machine.UARTConfig{TX: machine.UART1_TX_PIN, RX: machine.UART1_TX_PIN, BaudRate: 9600})

	// Power On GPS
	machine.GPS_STANDBY_PIN.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.GPS_RESET_PIN.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.GPS_POWER_PIN.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.GPS_STANDBY_PIN.Set(false) // GPS Pin Standby_L (PB3)
	machine.GPS_RESET_PIN.Set(false)   // GPS Pin Reset OFF (PB4)
	machine.GPS_POWER_PIN.Set(true)    // GPS Pin Power ON (PB5)

	// Run a Go routine for handling GPS UART messages
	go serialGps(uartGps)

	for {
		time.Sleep(1 * time.Second)
		machine.LED_RED.Set(!machine.LED_RED.Get())
	}

}
