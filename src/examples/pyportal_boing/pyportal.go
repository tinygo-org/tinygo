// +build pyportal

package main

import (
	"machine"

	"examples/pyportal_boing/ili9341"
)

var (
	display = ili9341.NewParallel(
		machine.LCD_DATA0,
		machine.TFT_WR,
		machine.TFT_DC,
		machine.TFT_CS,
		machine.TFT_RESET,
		machine.TFT_RD,
	)

	backlight = machine.TFT_BACKLIGHT
)
