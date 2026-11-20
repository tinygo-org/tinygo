//go:build rp2040

package main

import (
	"machine"

	board "example.com/customhw/boards/custom2040"
)

func useBoardHardware() {
	board.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	board.LED.High()
	_, _ = board.UART0.Write([]byte("ready\n"))

	pwm := machine.NewPWM(0)
	_ = pwm.Configure(machine.PWMConfig{})
	ch, _ := pwm.Channel(machine.GPIO0)
	pwm.Set(ch, pwm.Top()/2)
}
