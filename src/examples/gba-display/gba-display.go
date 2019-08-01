package main

// Draw a red square on the GameBoy Advance screen.

import (
	"image/color"
	"machine"
)

var display = machine.Display

func main() {
	display.Configure()

	for x := int16(30); x < 50; x++ {
		for y := int16(80); y < 100; y++ {
			display.SetPixel(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	display.Display()
}
