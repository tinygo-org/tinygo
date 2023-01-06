package main

import (
	"log"
	"machine/usb/joystick"
	"time"
)

var js = joystick.Port()

func main() {
	log.SetFlags(log.Lmicroseconds)
	ticker := time.NewTicker(10 * time.Millisecond)
	cnt := 0
	for range ticker.C {
		button := cnt%100 == 0
		js.SetButton(2, button)
		js.SetButton(3, !button)
		js.SetAxis(0, cnt%65535-32767)
		js.SendState()
	}
}
