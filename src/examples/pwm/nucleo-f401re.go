//go:build nucleof401re

package main

import "machine"

// TIM3 is reserved by the runtime tick timer on STM32F401.
// Use TIM2 instead: PB10=D6 (TIM2_CH3/AF1), PB3=D3 (TIM2_CH2/AF1).
var (
	pwm  = &machine.TIM2
	pinA = machine.D6
	pinB = machine.D3
)
