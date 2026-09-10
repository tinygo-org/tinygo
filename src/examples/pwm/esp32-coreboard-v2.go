//go:build esp32_coreboard_v2

package main

import "machine"

var (
	pwm  = machine.PWM0
	pinA = machine.GPIO18
	pinB = machine.GPIO19
)
