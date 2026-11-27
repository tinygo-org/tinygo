//go:build waveshare_esp32c6_lcd_1_47

package main

import "machine"

var (
	pwm  = machine.PWM0
	pinA = machine.LCD_BL_PIN
	pinB = machine.GPIO18
)
