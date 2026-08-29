//go:build py32f003xx || py32f030xx

package main

import "machine"

func configureUARTs() {
	_ = machine.DefaultUART.Configure(machine.UARTConfig{
		TX: machine.PA2,
		RX: machine.PA3,
	})
	_ = machine.UART2.Configure(machine.UARTConfig{
		TX: machine.PA0,
		RX: machine.PA1,
	})
}
