//go:build stm32f7

package main

import "machine"

var (
	pwm  = &machine.TIM1
	pinA = machine.PA8
	pinB = machine.PA9
)
