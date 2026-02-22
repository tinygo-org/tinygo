//go:build esp32s3_wroom1

package main

import "machine"

var (
	pwm  = machine.PWM0
	pinA = machine.GPIO17
	pinB = machine.GPIO18
)
