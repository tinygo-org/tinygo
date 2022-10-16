package main

import (
	"machine"
	"machine/usb/hid/keyboard"
	"time"
)

func main() {
	button := machine.BUTTON
	button.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	kb := keyboard.Port()

	for {
		if !button.Get() {
			kb.Write([]byte("tinygo"))
			time.Sleep(200 * time.Millisecond)
		}
	}
}
