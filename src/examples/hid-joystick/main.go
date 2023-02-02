package main

import (
	"log"
	"machine/usb/joystick"
	"math"
	"time"
)

var js = joystick.Port()

func main() {
	log.SetFlags(log.Lmicroseconds)
	ticker := time.NewTicker(10 * time.Millisecond)
	cnt := 0
	const f = 3.0
	for range ticker.C {
		t := float64(cnt) * 0.01
		x := 32767 * math.Sin(2*math.Pi*f*t)
		button := cnt%100 > 50
		js.SetButton(2, button)
		js.SetButton(3, !button)
		js.SetAxis(0, int(x))
		js.SendState()
		cnt++
	}
}
