package main

import (
	"machine"
	"machine/usb/hid/mouse"
	"time"
)

func main() {
	button := machine.BUTTON
	button.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	mouse := mouse.New()

	for {
		if !button.Get() {
			for j := 0; j < 5; j++ {
				for i := 0; i < 100; i++ {
					mouse.Move(1, 0)
					time.Sleep(1 * time.Millisecond)
				}

				for i := 0; i < 100; i++ {
					mouse.Move(0, 1)
					time.Sleep(1 * time.Millisecond)
				}

				for i := 0; i < 100; i++ {
					mouse.Move(-1, -1)
					time.Sleep(1 * time.Millisecond)
				}
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}
