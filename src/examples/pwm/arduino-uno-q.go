//go:build arduino_uno_q

package main

import "machine"

var (
	pwm  = &machine.TIM3
	pinA = machine.D3 // PB0 = TIM3_CH3
	pinB = machine.D6 // PB1 = TIM3_CH4
)
