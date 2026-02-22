//go:build xiao_esp32s3

package main

import "machine"

var (
	pwm  = machine.PWM0
	pinA = machine.D0
	pinB = machine.D1
)
