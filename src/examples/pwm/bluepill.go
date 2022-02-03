//go:build bluepill
// +build bluepill

package main

import "machine"

var (
	pwm  = &machine.TIM2
	pinA = machine.PA0
	pinB = machine.PA1
)
