//go:build esp32c3_supermini

package main

import "machine"

var (
	pwm  = machine.PWM0
	pinA = machine.GPIO5
	pinB = machine.GPIO6
)
