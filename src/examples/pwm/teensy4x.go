//go:build teensy40 || teensy41

package main

import "machine"

var (
	pwm  = machine.PWM4_2 // FlexPWM4 submodule 2 drives pins 2 and 3
	pinA = machine.D2     // channel A
	pinB = machine.D3     // channel B
)
