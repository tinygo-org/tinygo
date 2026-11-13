package main

import (
	"machine"
	"time"
)

func main() {
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	pad := machine.TouchPad{Pin: machine.GPIO4}
	if err := pad.Configure(); err != nil {
		println(err.Error())
		return
	}

	// Measure the untouched value first, so do not touch the pad at startup.
	var sum uint32
	for i := 0; i < 16; i++ {
		sum += uint32(pad.Get())
		time.Sleep(10 * time.Millisecond)
	}
	threshold := sum / 16 * 80 / 100

	for {
		val := uint32(pad.Get())
		touched := val != 0 && val < threshold
		led.Set(touched)
		println(val, touched)
		time.Sleep(time.Millisecond * 100)
	}
}
