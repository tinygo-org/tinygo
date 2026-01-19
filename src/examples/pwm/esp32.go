//go:build esp32
// +build esp32

package main

import "machine"

var (
	pwm  = machine.PWM0  // Use high-speed timer 0
	pinA = machine.GPIO2 // Built-in LED on many ESP32 boards
	pinB = machine.GPIO4 // Another GPIO for testing
)
