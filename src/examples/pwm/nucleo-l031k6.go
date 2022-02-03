//go:build stm32l0
// +build stm32l0

package main

import "machine"

var (
	pwm  = &machine.TIM2
	pinA = machine.PA0
	pinB = machine.PB3
)
