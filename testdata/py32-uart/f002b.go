//go:build py32f002bxx

package main

import "machine"

func configureUARTs() {
	_ = machine.DefaultUART.Configure(machine.UARTConfig{
		TX: machine.PA3,
		RX: machine.PA4,
	})
}
