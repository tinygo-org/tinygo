// blink program for the BBC micro:bit
package main

import (
	"machine"
	"time"
)

// The LED matrix in the micro:bit is a multiplexed display: https://en.wikipedia.org/wiki/Multiplexed_display
// Driver for easier control: https://github.com/tinygo-org/drivers/tree/master/microbitmatrix
func main() {
	ledrow := machine.LED_ROW_1
	ledrow.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ledcol := machine.LED_COL_1
	ledcol.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ledcol.Low()
	for {
		ledrow.Low()
		time.Sleep(time.Millisecond * 500)

		ledrow.High()
		time.Sleep(time.Millisecond * 500)
	}
}
