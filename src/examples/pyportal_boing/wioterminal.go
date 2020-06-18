// +build wioterminal

package main

import (
	"machine"

	"examples/pyportal_boing/ili9341"
)

var (
	display = ili9341.NewSpi(
		&machine.SPI3,
		machine.LCD_DC,
		machine.NoPin, // machine.LCD_SS_PIN,
		machine.LCD_RESET,
	)

	backlight = machine.LCD_BACKLIGHT
)

func init() {
	machine.SPI3.Configure(machine.SPIConfig{
		SCK:       machine.LCD_SCK_PIN,
		MOSI:      machine.LCD_MOSI_PIN,
		MISO:      machine.LCD_MISO_PIN,
		Frequency: 40000000,
	})
	machine.SPI3.SetupDMA()

	machine.LCD_SS_PIN.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.LCD_SS_PIN.Low()
}
