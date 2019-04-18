// blink program for the BBC micro:bit that uses the entire LED matrix
package main

import (
	"machine"
	"time"
)

func main() {
	machine.InitLEDMatrix()

	for {
		machine.ClearLEDMatrix()
		time.Sleep(time.Millisecond * 500)

		machine.SetEntireLEDMatrixOn()
		time.Sleep(time.Millisecond * 500)
	}
}
