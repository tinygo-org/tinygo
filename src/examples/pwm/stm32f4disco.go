//go:build stm32f4disco
// +build stm32f4disco

package main

import "machine"

var (
	// These pins correspond to LEDs on the discovery
	// board
	pwm  = &machine.TIM4
	pinA = machine.PD12
	pinB = machine.PD13
)
