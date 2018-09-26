// This program runs on an Arduino that has the following four devices connected:
// - Button connected to D2
// - Rotary analog dial connected to A0
// - RGB LED connected to D3, D5, and D6 used as PWM pins
// - BlinkM I2C RGB LED
//
// Pushing the button switches which color is selected.
// Rotating the dial changes the value for the currently selected color.
// Changing the color value updates the color displayed on both the
// PWM-controlled RGB LED and the I2C-controlled BlinkM.
package main

import (
	"machine"
	"time"
)

const (
	buttonPin = 2
	redPin    = 3
	greenPin  = 5
	bluePin   = 6

	red   = 0
	green = 1
	blue  = 2
)

func main() {
	machine.InitADC()
	machine.InitPWM()
	machine.I2C0.Configure(machine.I2CConfig{})

	// Init BlinkM
	machine.I2C0.WriteTo(0x09, []byte("o"))

	button := machine.GPIO{buttonPin}
	button.Configure(machine.GPIOConfig{Mode: machine.GPIO_INPUT})

	dial := machine.ADC{machine.ADC0}
	dial.Configure()

	redLED := machine.PWM{redPin}
	redLED.Configure()

	greenLED := machine.PWM{greenPin}
	greenLED.Configure()

	blueLED := machine.PWM{bluePin}
	blueLED.Configure()

	selectedColor := red
	colors := []uint16{0, 0, 0}

	for {
		// If we pushed the button, switch active color.
		if !button.Get() {
			if selectedColor == blue {
				selectedColor = red
			} else {
				selectedColor++
			}
		}

		// Change the intensity for the currently selected color based on the dial setting.
		colors[selectedColor] = (dial.Get())

		// Update the RGB LED.
		redLED.Set(colors[red])
		greenLED.Set(colors[green])
		blueLED.Set(colors[blue])

		// Update the BlinkM.
		machine.I2C0.WriteTo(0x09, []byte("n"))
		machine.I2C0.WriteTo(0x09, []byte{byte(colors[red] >> 8), byte(colors[green] >> 8), byte(colors[blue] >> 8)})

		time.Sleep(time.Millisecond * 100)
	}
}
