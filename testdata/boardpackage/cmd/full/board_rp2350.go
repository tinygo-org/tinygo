//go:build rp2350

package main

import (
	"machine"

	board "example.com/customhw/boards/custom2350"
)

func useBoardHardware() {
	board.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	board.LED.High()
	_, _ = board.UART0.Write([]byte("ready\n"))

	pwm0 := machine.NewPWM(0)
	_ = pwm0.Configure(machine.PWMConfig{})
	pwm11 := machine.NewPWM(11)
	_ = pwm11.Configure(machine.PWMConfig{})
}
